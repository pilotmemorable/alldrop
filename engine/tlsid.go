package engine

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Every installation owns a self-signed key pair. Its device ID is derived
// from the certificate, so peers cannot impersonate each other and all
// app-to-app traffic is encrypted with TLS.

func certID(der []byte) string {
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:12])
}

func loadOrCreateCert(dir string) (tls.Certificate, string, error) {
	certPath := filepath.Join(dir, "identity.crt")
	keyPath := filepath.Join(dir, "identity.key")
	if c, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil && len(c.Certificate) > 0 {
		return c, certID(c.Certificate[0]), nil
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, "", err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	tpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "AllDrop device"},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().AddDate(30, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"alldrop.local"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, "", err
	}
	kb, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return tls.Certificate{}, "", err
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb}), 0o600); err != nil {
		return tls.Certificate{}, "", err
	}
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644); err != nil {
		return tls.Certificate{}, "", err
	}
	c, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, "", err
	}
	return c, certID(der), nil
}

func (a *App) serverTLSConfig() *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{a.tlsCert},
		ClientAuth:   tls.RequestClientCert,
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"http/1.1"},
	}
}

var errPeerIdentity = errors.New("peer identity mismatch")

// peerTransport returns a pooled HTTP/1.1 TLS transport that only talks to
// the device whose certificate hashes to expectID ("" accepts any device;
// the caller must then verify resp.TLS itself).
func (a *App) peerTransport(expectID string) *http.Transport {
	a.mu.Lock()
	defer a.mu.Unlock()
	if t, ok := a.transports[expectID]; ok {
		return t
	}
	cfg := &tls.Config{
		Certificates:       []tls.Certificate{a.tlsCert},
		InsecureSkipVerify: true, // self-signed; pinned below instead
		MinVersion:         tls.VersionTLS12,
		NextProtos:         []string{"http/1.1"},
		VerifyConnection: func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return errPeerIdentity
			}
			if expectID != "" && certID(cs.PeerCertificates[0].Raw) != expectID {
				return errPeerIdentity
			}
			return nil
		},
	}
	t := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 4 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:       cfg,
		TLSHandshakeTimeout:   6 * time.Second,
		MaxIdleConnsPerHost:   6,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    true,
		ForceAttemptHTTP2:     false,
		TLSNextProto:          map[string]func(string, *tls.Conn) http.RoundTripper{},
		WriteBufferSize:       256 << 10,
		ReadBufferSize:        256 << 10,
		ResponseHeaderTimeout: 0,
	}
	a.transports[expectID] = t
	return t
}

func (a *App) peerClient(expectID string, timeout time.Duration) *http.Client {
	return &http.Client{Transport: a.peerTransport(expectID), Timeout: timeout}
}

// senderID returns the TLS identity of the device that made the request.
func senderID(r *http.Request) string {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return ""
	}
	return certID(r.TLS.PeerCertificates[0].Raw)
}
