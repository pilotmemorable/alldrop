package engine

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const appVersion = "1.1.1"

// Party identifies the sender of an offer.
type Party struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
}

// FileMeta describes one file of a transfer. Path is the relative path
// (including the file name) when a folder is sent, otherwise empty.
type FileMeta struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// Offer is an incoming transfer request that targets this server: either the
// owner of this machine ("local") or one of the browser guests connected to it.
type Offer struct {
	ID        string
	Secret    string
	SenderKey string // TLS identity of the sending peer (peer offers only)
	From      Party
	To        string // "local" or guest client id
	Files     []FileMeta
	Total     int64
	Status    string // pending, accepted, done, declined, expired, canceled, error
	Created   time.Time
	Updated   time.Time
	Received  int64
	Done      []bool
	SrcIP     string
	topMap    map[string]string
	Mode      string // "code": sender must type Code shown to the receiver; "accept": receiver taps Accept
	Code      string
	Fails     int
	TrustOnOK bool // receiver asked to trust the sender once the code matches
}

// Client is a browser guest that joined through the QR code.
type Client struct {
	ID       string
	Name     string
	OS       string
	LastSeen time.Time
}

// InboxItem is a file waiting to be downloaded by a browser guest.
type InboxItem struct {
	ID      string
	OfferID string
	To      string
	Name    string
	Size    int64
	From    string
	Path    string
	Temp    bool
	Created time.Time
}

// Peer is another AllDrop app (Windows, Mac, Linux, Android) on the network.
type Peer struct {
	ID       string
	Name     string
	OS       string
	Addr     string // ip:tlsport
	HTTPPort int
	LastSeen time.Time
	Probe    bool // found by scan/manual: keep alive by probing
	Manual   bool
}

// Send is an outgoing transfer started by a client of this server.
type Send struct {
	ID           string
	ClientID     string // "local" for the owner, otherwise guest id
	To           string
	ToName       string
	Kind         string // "local" or "peer"
	OfferID      string
	PeerID       string
	PeerAddr     string
	RemoteID     string
	RemoteSecret string
	Status       string
	Err          string
	Files        []FileMeta
	Total        int64
	Sent         int64
	Server       bool
	Mode         string
	Left         int
	Created      time.Time
	Updated      time.Time
	lastPoll     time.Time
}

// Staged is a file on this machine queued for sending (Windows "Send to",
// Android share sheet, command line).
type Staged struct {
	Abs  string
	Rel  string
	Size int64
	Temp bool
}

type HistoryItem struct {
	Name  string    `json:"name"`
	Path  string    `json:"path"`
	From  string    `json:"from"`
	Group string    `json:"group"` // top-level folder of a folder transfer, "" for single files
	Offer string    `json:"offer"`
	Size  int64     `json:"size"`
	Time  time.Time `json:"time"`
	TS    int64     `json:"ts"`
}

type App struct {
	mu       sync.Mutex
	cfg      Config
	cfgPath  string
	dataDir  string
	inboxDir string
	httpPort int
	tlsPort  int
	tlsCert  tls.Certificate
	hintIPs  []string
	platform string

	clients map[string]*Client
	peers   map[string]*Peer
	offers  map[string]*Offer
	sends   map[string]*Send
	inbox   map[string]*InboxItem
	staged  []Staged
	history []HistoryItem

	probing       map[string]bool
	lastOwnerPoll time.Time
	lastAutoOpen  time.Time
	dirOK         bool
	udp           net.PacketConn
	transports    map[string]*http.Transport
	scanning      bool
	active        bool

	opts    Options
	httpLn  net.Listener
	tlsLn   net.Listener
	uiSrv   *http.Server
	peerSrv *http.Server
	done    chan struct{}
	closed  bool
}

func newApp() *App {
	return &App{
		clients:    map[string]*Client{},
		peers:      map[string]*Peer{},
		offers:     map[string]*Offer{},
		sends:      map[string]*Send{},
		inbox:      map[string]*InboxItem{},
		probing:    map[string]bool{},
		transports: map[string]*http.Transport{},
		done:       make(chan struct{}),
	}
}

func (a *App) saveLocked() {
	if err := saveConfig(a.cfgPath, a.cfg); err != nil {
		logf("config save: %v", err)
	}
}

func (a *App) recvDirLocked() string {
	if a.dirOK {
		return a.cfg.DownloadDir
	}
	return filepath.Join(a.dataDir, "Received")
}

func (a *App) selfParty() Party {
	return Party{ID: a.cfg.ID, Name: a.cfg.Name, OS: a.platform}
}

// emit hands an event to the mobile shell (Android/iOS).
func (a *App) emit(parts ...string) {
	if a.opts.OnEvent != nil {
		a.opts.OnEvent(parts...)
	}
}

// notifyOwnerLocked brings an incoming request to the owner's attention.
func (a *App) notifyOwnerLocked(o *Offer) {
	if a.opts.OnEvent != nil {
		go a.emit("OFFER", o.From.Name, fmt.Sprint(len(o.Files)), o.Code)
		return
	}
	if noBrowser {
		return
	}
	if time.Since(a.lastOwnerPoll) > 4*time.Second && time.Since(a.lastAutoOpen) > 20*time.Second {
		a.lastAutoOpen = time.Now()
		go OpenURL(fmt.Sprintf("http://localhost:%d/", a.httpPort))
	}
}

type deviceJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
	Kind string `json:"kind"`
}

// devicesForLocked lists what a client can send to.
func (a *App) devicesForLocked(me string) []deviceJSON {
	var out []deviceJSON
	if me != "local" {
		out = append(out, deviceJSON{ID: "local", Name: a.cfg.Name, OS: a.platform, Kind: "host"})
	}
	var guests []deviceJSON
	for _, c := range a.clients {
		if c.ID == me || time.Since(c.LastSeen) > 12*time.Second {
			continue
		}
		guests = append(guests, deviceJSON{ID: "g:" + c.ID, Name: c.Name, OS: c.OS, Kind: "guest"})
	}
	sort.Slice(guests, func(i, j int) bool { return guests[i].Name < guests[j].Name })
	var peers []deviceJSON
	for _, p := range a.peers {
		if time.Since(p.LastSeen) > 15*time.Second {
			continue
		}
		peers = append(peers, deviceJSON{ID: "p:" + p.ID, Name: p.Name, OS: p.OS, Kind: "peer"})
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].Name < peers[j].Name })
	out = append(out, peers...)
	out = append(out, guests...)
	return out
}

func (a *App) removeInboxLocked(it *InboxItem) {
	delete(a.inbox, it.ID)
	if it.Temp {
		os.Remove(it.Path)
		os.Remove(filepath.Dir(it.Path)) // only succeeds when empty
	}
}

func terminal(status string) bool {
	switch status {
	case "done", "declined", "expired", "canceled", "error":
		return true
	}
	return false
}

// housekeeping expires stale requests, guests and temporary files.
func (a *App) housekeeping() {
	for a.sleep(3 * time.Second) {
		now := time.Now()
		a.mu.Lock()
		for id, o := range a.offers {
			if o.Status == "pending" && now.Sub(o.Created) > 3*time.Minute {
				o.Status, o.Updated = "expired", now
			}
			if o.Status == "accepted" && now.Sub(o.Updated) > 30*time.Minute {
				o.Status, o.Updated = "error", now
			}
			if terminal(o.Status) && now.Sub(o.Updated) > 20*time.Minute {
				delete(a.offers, id)
			}
		}
		for id, s := range a.sends {
			if terminal(s.Status) && now.Sub(s.Updated) > 30*time.Minute {
				delete(a.sends, id)
			}
		}
		for _, it := range a.inbox {
			c := a.clients[it.To]
			if now.Sub(it.Created) > 3*time.Hour || c == nil {
				a.removeInboxLocked(it)
			}
		}
		for id, c := range a.clients {
			if now.Sub(c.LastSeen) > 45*time.Minute {
				delete(a.clients, id)
			}
		}
		for id, p := range a.peers {
			if !p.Manual && now.Sub(p.LastSeen) > 2*time.Minute {
				delete(a.peers, id)
			}
		}
		// Tell the Android shell when a transfer is running so it can keep
		// the CPU and Wi-Fi awake while the screen is off.
		busy := false
		for _, o := range a.offers {
			if o.Status == "accepted" || (o.Status == "pending" && o.To == "local") {
				busy = true
			}
		}
		for _, s := range a.sends {
			if !terminal(s.Status) && now.Sub(s.Updated) < 2*time.Minute {
				busy = true
			}
		}
		if busy != a.active {
			a.active = busy
			if busy {
				go a.emit("ACTIVE", "1")
			} else {
				go a.emit("ACTIVE", "0")
			}
		}
		a.mu.Unlock()
	}
}
