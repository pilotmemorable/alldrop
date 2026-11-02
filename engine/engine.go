// Package engine is the AllDrop transfer engine shared by every platform:
// the Windows/macOS/Linux desktop program, the Android app (as a child
// process) and the iOS app (as a gomobile framework).
package engine

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultPort is the web port; the encrypted app-to-app port is DefaultPort+1.
const DefaultPort = 45454

const defaultHTTPPort = DefaultPort

// Version of the engine.
const Version = appVersion

var (
	mobileMode bool // no browser, no desktop helpers, events go to OnEvent
	noBrowser  bool
	ownerToken string
	logger     *log.Logger
)

func logf(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf(format, args...)
	}
}

// Options configures Start.
type Options struct {
	Port         int
	DataDir      string
	DownloadDir  string
	Name         string   // used for a fresh installation only
	HintIPs      []string // LAN addresses when the OS hides interfaces (Android 11+)
	Platform     string   // "android", "ios", "ipad"; empty = this OS
	NoBrowser    bool
	Mobile       bool
	OwnerToken   string // required from the local UI when set (mobile)
	PeriodicScan bool   // scan the subnet regularly (iOS cannot use broadcast)
	OnEvent      func(parts ...string)
	OnQuit       func()
	Log          io.Writer
}

// DefaultDataDir is where settings and the device key live on desktops.
func DefaultDataDir() string { return defaultDataDir() }

// IdentityID returns the device id of the installation in dataDir.
func IdentityID(dataDir string) (string, error) {
	_, id, err := loadOrCreateCert(dataDir)
	return id, err
}

// Start runs the engine until Close is called.
func Start(o Options) (*App, error) {
	if o.Port == 0 {
		o.Port = DefaultPort
	}
	if o.DataDir == "" {
		o.DataDir = defaultDataDir()
	}
	if err := os.MkdirAll(o.DataDir, 0o700); err != nil {
		return nil, err
	}
	mobileMode = o.Mobile
	noBrowser = o.NoBrowser || o.Mobile
	ownerToken = o.OwnerToken
	if o.Log != nil {
		logger = log.New(o.Log, "", log.LstdFlags)
	} else if lf, err := os.OpenFile(filepath.Join(o.DataDir, "alldrop.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644); err == nil {
		logger = log.New(io.MultiWriter(lf, stderrWriter{}), "", log.LstdFlags)
	} else {
		logger = log.New(stderrWriter{}, "", log.LstdFlags)
	}

	a := newApp()
	a.opts = o
	a.dataDir = o.DataDir
	a.inboxDir = filepath.Join(o.DataDir, "inbox")
	a.cfgPath = filepath.Join(o.DataDir, "config.json")
	a.platform = platformName()
	if o.Platform != "" {
		a.platform = o.Platform
	}
	for _, h := range o.HintIPs {
		if h = strings.TrimSpace(h); h != "" {
			a.hintIPs = append(a.hintIPs, h)
		}
	}
	cert, id, err := loadOrCreateCert(o.DataDir)
	if err != nil {
		return nil, err
	}
	a.tlsCert = cert
	_, statErr := os.Stat(a.cfgPath)
	a.cfg = loadConfig(a.cfgPath)
	a.cfg.ID = id
	if o.Name != "" && os.IsNotExist(statErr) {
		a.cfg.Name = cleanName(o.Name, a.cfg.Name)
	}
	if o.DownloadDir != "" {
		a.cfg.DownloadDir = o.DownloadDir
	}
	a.saveLocked()

	httpLn, tlsLn, err := listenPair(o.Port)
	if err != nil {
		return nil, err
	}
	a.httpLn, a.tlsLn = httpLn, tlsLn
	a.httpPort = httpLn.Addr().(*net.TCPAddr).Port
	a.tlsPort = tlsLn.Addr().(*net.TCPAddr).Port
	a.dirOK = dirWritable(a.cfg.DownloadDir)
	os.RemoveAll(a.inboxDir)
	os.MkdirAll(a.inboxDir, 0o755)
	logf("AllDrop %s (%s) id=%s web=%d tls=%d downloads=%s ok=%v", appVersion, a.platform, a.cfg.ID, a.httpPort, a.tlsPort, a.cfg.DownloadDir, a.dirOK)

	a.uiSrv = &http.Server{Handler: a.uiHandler(), ReadHeaderTimeout: 20 * time.Second, IdleTimeout: 2 * time.Minute}
	a.peerSrv = &http.Server{
		Handler:           a.peerHandler(),
		ReadHeaderTimeout: 20 * time.Second,
		IdleTimeout:       2 * time.Minute,
		TLSConfig:         a.serverTLSConfig(),
		TLSNextProto:      map[string]func(*http.Server, *tls.Conn, http.Handler){},
		ErrorLog:          log.New(io.Discard, "", 0),
	}
	go func() { logf("web server: %v", a.uiSrv.Serve(httpLn)) }()
	go func() { logf("peer server: %v", a.peerSrv.ServeTLS(tlsLn, "", "")) }()
	go a.housekeeping()
	a.startDiscovery()
	a.emit("PORT", fmt.Sprint(a.httpPort))
	return a, nil
}

// Port is the local web port of the running engine.
func (a *App) Port() int { return a.httpPort }

// Stage queues files from this device for sending (share sheets, "Send to").
func (a *App) Stage(paths []string, temp bool) int { return a.stagePaths(paths, temp) }

// Close stops the engine and tells other devices it is gone.
func (a *App) Close() {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return
	}
	a.closed = true
	close(a.done)
	a.mu.Unlock()
	if a.udp != nil {
		a.announce("bye")
		a.udp.Close()
	}
	if a.uiSrv != nil {
		a.uiSrv.Close()
	}
	if a.peerSrv != nil {
		a.peerSrv.Close()
	}
	a.mu.Lock()
	for _, t := range a.transports {
		t.CloseIdleConnections()
	}
	a.mu.Unlock()
	logf("engine stopped")
}

// sleep waits d or returns false when the engine is closing.
func (a *App) sleep(d time.Duration) bool {
	select {
	case <-a.done:
		return false
	case <-time.After(d):
		return true
	}
}

func (a *App) quit() {
	if a.opts.OnQuit != nil {
		a.opts.OnQuit()
		return
	}
	a.Close()
}

// FindRunning returns the web port of an engine with this id already
// running on this machine, or 0.
func FindRunning(id string, base int) int {
	cl := &http.Client{Timeout: 700 * time.Millisecond, Transport: &http.Transport{Proxy: nil}}
	for k := 0; k < 10; k++ {
		p := base + 2*k
		resp, err := cl.Get(fmt.Sprintf("http://127.0.0.1:%d/ping", p))
		if err != nil {
			continue
		}
		var out struct{ App, ID string }
		_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&out)
		resp.Body.Close()
		if out.App == "alldrop" && out.ID == id {
			return p
		}
	}
	return 0
}

// listenPair opens the web port and the encrypted peer port (web+1).
func listenPair(base int) (net.Listener, net.Listener, error) {
	for k := 0; k < 10; k++ {
		p := base + 2*k
		l1, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err != nil {
			continue
		}
		l2, err := net.Listen("tcp", fmt.Sprintf(":%d", p+1))
		if err != nil {
			l1.Close()
			continue
		}
		return l1, l2, nil
	}
	l1, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, err
	}
	l2, err := net.Listen("tcp", ":0")
	if err != nil {
		l1.Close()
		return nil, nil, err
	}
	return l1, l2, nil
}

// stderrWriter ignores errors (Windows GUI builds have no console).
type stderrWriter struct{}

func (stderrWriter) Write(p []byte) (int, error) {
	os.Stderr.Write(p)
	return len(p), nil
}
