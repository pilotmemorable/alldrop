package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const discoveryPort = 45453

type hello struct {
	App  string `json:"app"`
	V    int    `json:"v"`
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
	Port int    `json:"port"`
	TLS  int    `json:"tls"`
	T    string `json:"t"`
}

type peerInfo struct {
	App  string `json:"app"`
	V    int    `json:"v"`
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
	Port int    `json:"port"`
	TLS  int    `json:"tls"`
}

func (a *App) helloPacket(t string) []byte {
	a.mu.Lock()
	h := hello{App: "alldrop", V: 1, ID: a.cfg.ID, Name: a.cfg.Name, OS: a.platform, Port: a.httpPort, TLS: a.tlsPort, T: t}
	a.mu.Unlock()
	b, _ := json.Marshal(h)
	return b
}

func (a *App) startDiscovery() {
	lc := net.ListenConfig{Control: reuseControl}
	pc, err := lc.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", discoveryPort))
	if err != nil {
		logf("discovery: cannot bind udp %d: %v (falling back to random port)", discoveryPort, err)
		pc, err = net.ListenPacket("udp4", ":0")
		if err != nil {
			logf("discovery disabled: %v", err)
			return
		}
	}
	a.udp = pc
	go a.readLoop(pc)
	go a.announceLoop()
	go a.probeLoop()
	go func() {
		if !a.sleep(2500 * time.Millisecond) {
			return
		}
		a.mu.Lock()
		n := len(a.peers)
		a.mu.Unlock()
		if n == 0 || a.opts.PeriodicScan {
			a.scan()
		}
		for a.opts.PeriodicScan && a.sleep(40*time.Second) {
			a.scan()
		}
	}()
}

func (a *App) broadcastTargets() []string {
	t := map[string]bool{"255.255.255.255": true}
	for _, la := range localIPv4(a.hintIPs) {
		if la.Bcast != "" {
			t[la.Bcast] = true
		}
	}
	out := make([]string, 0, len(t))
	for k := range t {
		out = append(out, k)
	}
	return out
}

func (a *App) announce(kind string) {
	if a.udp == nil {
		return
	}
	pkt := a.helloPacket(kind)
	for _, ip := range a.broadcastTargets() {
		addr := &net.UDPAddr{IP: net.ParseIP(ip), Port: discoveryPort}
		_, _ = a.udp.WriteTo(pkt, addr)
	}
}

func (a *App) announceLoop() {
	for i := 0; ; i++ {
		a.announce("hello")
		d := 3 * time.Second
		if i < 3 {
			d = time.Second
		}
		if !a.sleep(d) {
			return
		}
	}
}

func (a *App) readLoop(pc net.PacketConn) {
	buf := make([]byte, 2048)
	lastReply := map[string]time.Time{}
	for {
		n, src, err := pc.ReadFrom(buf)
		if err != nil {
			if !a.sleep(time.Second) {
				return
			}
			logf("discovery read: %v", err)
			continue
		}
		var h hello
		if json.Unmarshal(buf[:n], &h) != nil || h.App != "alldrop" || h.ID == "" {
			continue
		}
		a.mu.Lock()
		self := h.ID == a.cfg.ID
		a.mu.Unlock()
		if self {
			continue
		}
		ua, ok := src.(*net.UDPAddr)
		if !ok || h.TLS <= 0 || h.TLS > 65535 {
			continue
		}
		addr := net.JoinHostPort(ua.IP.String(), strconv.Itoa(h.TLS))
		if h.T == "bye" {
			a.mu.Lock()
			if p := a.peers[h.ID]; p != nil && p.Addr == addr && !p.Manual {
				delete(a.peers, h.ID)
			}
			a.mu.Unlock()
			continue
		}
		a.mu.Lock()
		p := a.peers[h.ID]
		known := p != nil && p.Addr == addr
		if known {
			p.LastSeen = time.Now()
			p.Name = cleanName(h.Name, p.Name)
			p.OS = cleanOS(h.OS)
		}
		a.mu.Unlock()
		if !known {
			go a.verifyAndAdd(addr, h.ID, false, false)
		}
		if h.T == "hello" && time.Since(lastReply[src.String()]) > time.Second {
			lastReply[src.String()] = time.Now()
			_, _ = pc.WriteTo(a.helloPacket("reply"), src)
		}
	}
}

func cleanName(n, fallback string) string {
	n = strings.TrimSpace(n)
	if n == "" {
		return fallback
	}
	if len([]rune(n)) > 48 {
		n = string([]rune(n)[:48])
	}
	return n
}

func cleanOS(o string) string {
	switch o {
	case "windows", "mac", "linux", "android", "ios", "ipad", "browser":
		return o
	}
	return "linux"
}

// fetchInfo talks TLS to addr and returns the identity it proved.
func (a *App) fetchInfo(addr, expectID string, timeout time.Duration) (*peerInfo, error) {
	cl := a.peerClient(expectID, timeout)
	resp, err := cl.Get("https://" + addr + "/peer/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return nil, fmt.Errorf("bad response %d", resp.StatusCode)
	}
	var info peerInfo
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8192)).Decode(&info); err != nil {
		return nil, err
	}
	proven := certID(resp.TLS.PeerCertificates[0].Raw)
	if info.App != "alldrop" || info.ID != proven {
		return nil, errPeerIdentity
	}
	return &info, nil
}

// verifyAndAdd confirms over TLS that addr really is the device it claims,
// then records it as a peer.
func (a *App) verifyAndAdd(addr, expectID string, probe, manual bool) bool {
	key := addr + "|" + expectID
	a.mu.Lock()
	if a.probing[key] {
		a.mu.Unlock()
		return false
	}
	a.probing[key] = true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.probing, key)
		a.mu.Unlock()
	}()
	timeout := 3 * time.Second
	if probe && !manual {
		timeout = 1500 * time.Millisecond
	}
	info, err := a.fetchInfo(addr, expectID, timeout)
	if err != nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if info.ID == a.cfg.ID {
		return false
	}
	p := a.peers[info.ID]
	if p == nil {
		p = &Peer{ID: info.ID}
		a.peers[info.ID] = p
		logf("peer found: %s (%s) at %s", info.Name, info.OS, addr)
		if probe {
			go a.sayHelloTo(addr, info.ID)
		}
	}
	p.Addr = addr
	p.Name = cleanName(info.Name, "Device")
	p.OS = cleanOS(info.OS)
	p.HTTPPort = info.Port
	p.LastSeen = time.Now()
	if probe {
		p.Probe = true
	}
	if manual {
		p.Manual = true
	}
	return true
}

// probeLoop keeps peers alive that were found by scanning or added by hand,
// for networks that drop broadcast packets.
func (a *App) probeLoop() {
	for a.sleep(6 * time.Second) {
		a.mu.Lock()
		var addrs [][2]string
		for _, p := range a.peers {
			if (p.Probe || p.Manual) && time.Since(p.LastSeen) > 5*time.Second {
				addrs = append(addrs, [2]string{p.Addr, p.ID})
			}
		}
		manual := append([]string(nil), a.cfg.ManualPeers...)
		a.mu.Unlock()
		for _, x := range addrs {
			go a.verifyAndAdd(x[0], x[1], true, false)
		}
		for _, m := range manual {
			go a.verifyAndAdd(m, "", true, true)
		}
	}
}

// scan probes every address of the local /24 networks on the default port.
func (a *App) scan() {
	a.mu.Lock()
	if a.scanning {
		a.mu.Unlock()
		return
	}
	a.scanning = true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.scanning = false
		a.mu.Unlock()
	}()
	addrs := localIPv4(a.hintIPs)
	if len(addrs) > 3 {
		addrs = addrs[:3]
	}
	sem := make(chan struct{}, 48)
	var wg sync.WaitGroup
	for _, la := range addrs {
		ip := net.ParseIP(la.IP).To4()
		if ip == nil {
			continue
		}
		for i := 1; i < 255; i++ {
			cand := net.IPv4(ip[0], ip[1], ip[2], byte(i)).String()
			if cand == la.IP {
				continue
			}
			for _, port := range []int{defaultHTTPPort + 1, defaultHTTPPort + 3} {
				addr := net.JoinHostPort(cand, strconv.Itoa(port))
				wg.Add(1)
				sem <- struct{}{}
				go func(addr string) {
					defer wg.Done()
					defer func() { <-sem }()
					c, err := net.DialTimeout("tcp", addr, 600*time.Millisecond)
					if err != nil {
						return
					}
					c.Close()
					a.verifyAndAdd(addr, "", true, false)
				}(addr)
			}
		}
	}
	wg.Wait()
}

// parseManualAddr accepts "ip", "ip:httpPort" or a pasted URL and returns ip:tlsPort.
func parseManualAddr(s string) (string, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	host, port := s, defaultHTTPPort
	if h, p, err := net.SplitHostPort(s); err == nil {
		host = h
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 || n >= 65535 {
			return "", fmt.Errorf("bad port")
		}
		port = n
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("bad address")
	}
	return net.JoinHostPort(ip.String(), strconv.Itoa(port+1)), nil
}

func (a *App) peerInfoHandler(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	info := peerInfo{App: "alldrop", V: 1, ID: a.cfg.ID, Name: a.cfg.Name, OS: a.platform, Port: a.httpPort, TLS: a.tlsPort}
	a.mu.Unlock()
	writeJSON(w, 200, info)
}
