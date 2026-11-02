package engine

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var bufPool = sync.Pool{New: func() interface{} { b := make([]byte, 1<<20); return &b }}

var (
	errNotAccepted = errors.New("transfer is not accepted")
	errBadIndex    = errors.New("bad file index")
	errSize        = errors.New("size mismatch")
	errCanceled    = errors.New("canceled")
)

func validFiles(files []FileMeta) error {
	if len(files) == 0 || len(files) > 20000 {
		return errors.New("bad file list")
	}
	for i := range files {
		if files[i].Size < 0 {
			return errors.New("bad size")
		}
		if len(files[i].Name) > 1024 || len(files[i].Path) > 4096 {
			return errors.New("name too long")
		}
		if strings.TrimSpace(files[i].Name) == "" {
			files[i].Name = "file"
		}
	}
	return nil
}

// newOfferLocked registers an incoming request for "local" or a guest.
func (a *App) newOfferLocked(from Party, to string, files []FileMeta, srcIP, senderKey string) *Offer {
	var total int64
	for _, f := range files {
		total += f.Size
	}
	now := time.Now()
	o := &Offer{
		ID: randID(8), Secret: randID(16), SenderKey: senderKey,
		From: from, To: to, Files: files, Total: total,
		Status: "pending", Created: now, Updated: now,
		Done: make([]bool, len(files)), SrcIP: srcIP, topMap: map[string]string{},
	}
	a.offers[o.ID] = o
	if to == "local" {
		if _, trusted := a.cfg.Trusted[from.ID]; a.cfg.AutoAccept || trusted {
			o.Status = "accepted"
		}
	}
	if o.Status == "pending" {
		if a.cfg.NoCode {
			o.Mode = "accept"
		} else {
			o.Mode, o.Code = "code", randCode()
		}
		if to == "local" {
			a.notifyOwnerLocked(o)
		}
	}
	return o
}

const maxCodeTries = 5

func randCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// checkCodeLocked verifies the code typed by the sender. Five wrong tries
// cancel the request, so guessing a 6-digit code is not practical.
func (a *App) checkCodeLocked(o *Offer, code string) (bool, int) {
	if o.Mode != "code" || o.Status != "pending" {
		return o.Status == "accepted", maxCodeTries - o.Fails
	}
	var digits []byte
	for i := 0; i < len(code); i++ {
		if code[i] >= '0' && code[i] <= '9' {
			digits = append(digits, code[i])
		}
	}
	o.Updated = time.Now()
	if len(digits) == 6 && subtle.ConstantTimeCompare(digits, []byte(o.Code)) == 1 {
		o.Status = "accepted"
		if o.TrustOnOK && o.To == "local" && o.From.ID != "" {
			a.cfg.Trusted[o.From.ID] = o.From.Name
			a.saveLocked()
		}
		return true, maxCodeTries - o.Fails
	}
	o.Fails++
	if o.Fails >= maxCodeTries {
		o.Status = "declined"
	}
	return false, maxCodeTries - o.Fails
}

func (a *App) pendingCount() int {
	n := 0
	for _, o := range a.offers {
		if o.Status == "pending" {
			n++
		}
	}
	return n
}

// openDestLocked chooses a collision-free destination and creates its
// partial file. Folders keep their structure; a second "Photos" folder
// becomes "Photos (2)".
func (a *App) openDestLocked(o *Offer, idx int) (string, *os.File, error) {
	fm := o.Files[idx]
	if o.To != "local" {
		dest := filepath.Join(a.inboxDir, o.ID, strconv.Itoa(idx))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return "", nil, err
		}
		f, err := os.OpenFile(dest+partSuffix, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		return dest, f, err
	}
	dir := a.recvDirLocked()
	comps := sanitizeRel(fm.Path)
	if len(comps) == 0 {
		comps = sanitizeRel(fm.Name)
	}
	if len(comps) == 0 {
		comps = []string{"file"}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, err
	}
	sub := dir
	if len(comps) > 1 {
		top, ok := o.topMap[comps[0]]
		if !ok {
			top = uniqueName(dir, comps[0])
			o.topMap[comps[0]] = top
		}
		sub = filepath.Join(append([]string{dir, top}, comps[1:len(comps)-1]...)...)
		if err := os.MkdirAll(sub, 0o755); err != nil {
			return "", nil, err
		}
	}
	name := uniqueName(sub, comps[len(comps)-1])
	dest := filepath.Join(sub, name)
	f, err := os.OpenFile(dest+partSuffix, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	return dest, f, err
}

// receiveStream writes one file of an accepted offer straight from the network.
func (a *App) receiveStream(o *Offer, idx int, body io.Reader) error {
	a.mu.Lock()
	if o.Status != "accepted" {
		a.mu.Unlock()
		return errNotAccepted
	}
	if idx < 0 || idx >= len(o.Files) || o.Done[idx] {
		a.mu.Unlock()
		return errBadIndex
	}
	size := o.Files[idx].Size
	dest, f, err := a.openDestLocked(o, idx)
	a.mu.Unlock()
	if err != nil {
		logf("receive: open: %v", err)
		return err
	}
	part := dest + partSuffix
	bp := bufPool.Get().(*[]byte)
	defer bufPool.Put(bp)
	buf := *bp
	var n int64
	fail := func(e error) error {
		f.Close()
		os.Remove(part)
		a.mu.Lock()
		o.Received -= n
		o.Updated = time.Now()
		a.mu.Unlock()
		return e
	}
	for {
		m, rerr := body.Read(buf)
		if m > 0 {
			if n+int64(m) > size {
				return fail(errSize)
			}
			if _, werr := f.Write(buf[:m]); werr != nil {
				return fail(werr)
			}
			n += int64(m)
			a.mu.Lock()
			o.Received += int64(m)
			o.Updated = time.Now()
			st := o.Status
			a.mu.Unlock()
			if st == "canceled" {
				return fail(errCanceled)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return fail(rerr)
		}
	}
	if n != size {
		return fail(errSize)
	}
	if err := f.Close(); err != nil {
		return fail(err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if o.To == "local" && exists(dest) {
		dest = filepath.Join(filepath.Dir(dest), uniqueName(filepath.Dir(dest), filepath.Base(dest)))
	}
	if err := os.Rename(part, dest); err != nil {
		os.Remove(part)
		return err
	}
	o.Done[idx] = true
	o.Updated = time.Now()
	a.fileArrivedLocked(o, idx, dest, true)
	return nil
}

// fileArrivedLocked records a completed file and closes the offer when all arrived.
func (a *App) fileArrivedLocked(o *Offer, idx int, dest string, temp bool) {
	fm := o.Files[idx]
	if o.To == "local" {
		now := time.Now()
		a.history = append([]HistoryItem{{Name: filepath.Base(dest), Path: dest, From: o.From.Name, Size: fm.Size, Time: now, TS: now.UnixMilli()}}, a.history...)
		if len(a.history) > 200 {
			a.history = a.history[:200]
		}
		go a.emit("SAVED", dest)
	} else {
		name := fm.Path
		if name == "" {
			name = fm.Name
		}
		comps := sanitizeRel(name)
		if len(comps) == 0 {
			comps = []string{"file"}
		}
		it := &InboxItem{ID: randID(8), OfferID: o.ID, To: o.To, Name: strings.Join(comps, "/"), Size: fm.Size,
			From: o.From.Name, Path: dest, Temp: temp, Created: time.Now()}
		a.inbox[it.ID] = it
	}
	all := true
	for _, d := range o.Done {
		if !d {
			all = false
			break
		}
	}
	if all {
		o.Status = "done"
	}
}

// ---------- peer (app-to-app, TLS) endpoints ----------

type offerReq struct {
	From  Party      `json:"from"`
	Sub   string     `json:"sub"` // guest id when forwarded on behalf of a browser guest
	Files []FileMeta `json:"files"`
}

func (a *App) peerOffer(w http.ResponseWriter, r *http.Request) {
	sid := senderID(r)
	if sid == "" || r.Method != http.MethodPost {
		httpErr(w, 403, "identity required")
		return
	}
	var req offerReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 16<<20)).Decode(&req); err != nil {
		httpErr(w, 400, "bad json")
		return
	}
	if err := validFiles(req.Files); err != nil {
		httpErr(w, 400, err.Error())
		return
	}
	from := Party{ID: sid, Name: cleanName(req.From.Name, "Device"), OS: cleanOS(req.From.OS)}
	if req.Sub != "" {
		from.ID = sid + "/" + sanitizeName(req.Sub)
	}
	ip := remoteIP(r.RemoteAddr).String()
	a.mu.Lock()
	n := 0
	for _, o := range a.offers {
		if o.Status == "pending" && o.SrcIP == ip {
			n++
		}
	}
	if n >= 8 {
		a.mu.Unlock()
		httpErr(w, 429, "too many pending requests")
		return
	}
	o := a.newOfferLocked(from, "local", req.Files, ip, sid)
	a.mu.Unlock()
	writeJSON(w, 200, map[string]string{"id": o.ID, "secret": o.Secret})
}

func (a *App) peerOfferFor(r *http.Request) *Offer {
	q := r.URL.Query()
	a.mu.Lock()
	defer a.mu.Unlock()
	o := a.offers[q.Get("id")]
	if o == nil || o.Secret != q.Get("secret") || o.SenderKey == "" || o.SenderKey != senderID(r) {
		return nil
	}
	return o
}

func (a *App) peerStatus(w http.ResponseWriter, r *http.Request) {
	o := a.peerOfferFor(r)
	if o == nil {
		httpErr(w, 404, "unknown transfer")
		return
	}
	a.mu.Lock()
	out := map[string]interface{}{"status": o.Status, "mode": o.Mode, "left": maxCodeTries - o.Fails}
	a.mu.Unlock()
	writeJSON(w, 200, out)
}

func (a *App) peerCode(w http.ResponseWriter, r *http.Request) {
	o := a.peerOfferFor(r)
	if o == nil || r.Method != http.MethodPost {
		httpErr(w, 404, "unknown transfer")
		return
	}
	var req struct{ Code string }
	if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	ok, left := a.checkCodeLocked(o, req.Code)
	st := o.Status
	a.mu.Unlock()
	writeJSON(w, 200, map[string]interface{}{"ok": ok, "status": st, "left": left})
}

// peerHello lets a device that found us by scanning (iPhones cannot use
// broadcast) introduce itself so it shows up here too.
func (a *App) peerHello(w http.ResponseWriter, r *http.Request) {
	sid := senderID(r)
	if sid == "" || r.Method != http.MethodPost {
		httpErr(w, 403, "identity required")
		return
	}
	var req struct {
		TLS int `json:"tls"`
	}
	if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req) != nil || req.TLS <= 0 || req.TLS > 65535 {
		httpErr(w, 400, "bad request")
		return
	}
	ip := remoteIP(r.RemoteAddr)
	if ip == nil {
		httpErr(w, 400, "bad address")
		return
	}
	go a.verifyAndAdd(net.JoinHostPort(ip.String(), strconv.Itoa(req.TLS)), sid, true, false)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) sayHelloTo(addr, id string) {
	body, _ := json.Marshal(map[string]int{"port": a.httpPort, "tls": a.tlsPort})
	if resp, err := a.peerClient(id, 4*time.Second).Post("https://"+addr+"/peer/hello", "application/json", bytes.NewReader(body)); err == nil {
		resp.Body.Close()
	}
}

func (a *App) peerUpload(w http.ResponseWriter, r *http.Request) {
	o := a.peerOfferFor(r)
	if o == nil || r.Method != http.MethodPost {
		httpErr(w, 404, "unknown transfer")
		return
	}
	idx, _ := strconv.Atoi(r.URL.Query().Get("i"))
	if err := a.receiveStream(o, idx, r.Body); err != nil {
		httpErr(w, 409, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) peerCancel(w http.ResponseWriter, r *http.Request) {
	o := a.peerOfferFor(r)
	if o == nil {
		httpErr(w, 404, "unknown transfer")
		return
	}
	a.mu.Lock()
	if !terminal(o.Status) {
		o.Status, o.Updated = "canceled", time.Now()
	}
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---------- outgoing (client of this server -> somebody) ----------

func (a *App) partyFor(me string) (Party, string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if me == "local" {
		return a.selfParty(), ""
	}
	c := a.clients[me]
	if c == nil {
		return Party{ID: "g:" + me, Name: "Guest", OS: "browser"}, me
	}
	return Party{ID: "g:" + c.ID, Name: c.Name, OS: c.OS}, me
}

// startSend creates the request on the target and returns a Send handle.
func (a *App) startSend(me, to string, files []FileMeta, server bool) (*Send, error) {
	if err := validFiles(files); err != nil {
		return nil, err
	}
	from, sub := a.partyFor(me)
	var total int64
	for _, f := range files {
		total += f.Size
	}
	now := time.Now()
	s := &Send{ID: randID(8), ClientID: me, To: to, Files: files, Total: total, Status: "pending", Server: server, Created: now, Updated: now}
	switch {
	case to == "local" || strings.HasPrefix(to, "g:"):
		target := "local"
		a.mu.Lock()
		if to == "local" {
			if me == "local" {
				a.mu.Unlock()
				return nil, errors.New("cannot send to yourself")
			}
			s.ToName = a.cfg.Name
		} else {
			target = strings.TrimPrefix(to, "g:")
			c := a.clients[target]
			if c == nil || target == me {
				a.mu.Unlock()
				return nil, errors.New("device is not available")
			}
			s.ToName = c.Name
		}
		o := a.newOfferLocked(from, target, files, "", "")
		s.Kind, s.OfferID, s.Status = "local", o.ID, o.Status
		a.sends[s.ID] = s
		a.mu.Unlock()
		return s, nil
	case strings.HasPrefix(to, "p:"):
		pid := strings.TrimPrefix(to, "p:")
		a.mu.Lock()
		p := a.peers[pid]
		if p == nil {
			a.mu.Unlock()
			return nil, errors.New("device is not available")
		}
		addr, pname := p.Addr, p.Name
		selfName := a.cfg.Name
		a.mu.Unlock()
		req := offerReq{From: from, Sub: sub, Files: files}
		if sub != "" {
			req.From.Name = from.Name + " (" + selfName + ")"
		}
		body, _ := json.Marshal(req)
		resp, err := a.peerClient(pid, 15*time.Second).Post("https://"+addr+"/peer/offer", "application/json", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("could not reach %s", pname)
		}
		defer resp.Body.Close()
		var out struct{ ID, Secret, Error string }
		_ = json.NewDecoder(io.LimitReader(resp.Body, 65536)).Decode(&out)
		if resp.StatusCode != 200 || out.ID == "" {
			if out.Error == "" {
				out.Error = resp.Status
			}
			return nil, errors.New(out.Error)
		}
		s.Kind, s.PeerID, s.PeerAddr, s.RemoteID, s.RemoteSecret, s.ToName = "peer", pid, addr, out.ID, out.Secret, pname
		a.mu.Lock()
		a.sends[s.ID] = s
		a.mu.Unlock()
		return s, nil
	}
	return nil, errors.New("unknown device")
}

// refreshSend updates the status of a send from its target.
func (a *App) refreshSend(s *Send) string {
	a.mu.Lock()
	if s.Kind == "local" {
		if o := a.offers[s.OfferID]; o != nil {
			if !(s.Status == "canceled" || s.Status == "error") {
				s.Status = o.Status
			}
			s.Mode, s.Left = o.Mode, maxCodeTries-o.Fails
		} else if !terminal(s.Status) {
			s.Status = "expired"
		}
		st := s.Status
		a.mu.Unlock()
		return st
	}
	st := s.Status
	if terminal(st) || (st == "accepted" && !s.Server) || time.Since(s.lastPoll) < 400*time.Millisecond {
		a.mu.Unlock()
		return st
	}
	s.lastPoll = time.Now()
	pid, addr, rid, sec := s.PeerID, s.PeerAddr, s.RemoteID, s.RemoteSecret
	a.mu.Unlock()
	u := "https://" + addr + "/peer/status?id=" + url.QueryEscape(rid) + "&secret=" + url.QueryEscape(sec)
	resp, err := a.peerClient(pid, 6*time.Second).Get(u)
	if err != nil {
		return st
	}
	defer resp.Body.Close()
	var out struct {
		Status string
		Mode   string
		Left   int
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&out)
	a.mu.Lock()
	defer a.mu.Unlock()
	if out.Mode != "" {
		s.Mode, s.Left = out.Mode, out.Left
	}
	if resp.StatusCode == 404 {
		s.Status = "expired"
	} else if out.Status != "" && !terminal(s.Status) {
		if !(s.Status == "accepted" && out.Status == "done" && s.Server) {
			s.Status = out.Status
		}
	}
	s.Updated = time.Now()
	return s.Status
}

// submitCode passes the code typed by the sender to the receiving device.
func (a *App) submitCode(s *Send, code string) (bool, string, int, error) {
	a.mu.Lock()
	if s.Kind == "local" {
		defer a.mu.Unlock()
		o := a.offers[s.OfferID]
		if o == nil {
			return false, "expired", 0, nil
		}
		ok, left := a.checkCodeLocked(o, code)
		s.Status, s.Left, s.Updated = o.Status, left, time.Now()
		return ok, o.Status, left, nil
	}
	pid, addr, rid, sec := s.PeerID, s.PeerAddr, s.RemoteID, s.RemoteSecret
	a.mu.Unlock()
	body, _ := json.Marshal(map[string]string{"code": code})
	u := "https://" + addr + "/peer/code?id=" + url.QueryEscape(rid) + "&secret=" + url.QueryEscape(sec)
	resp, err := a.peerClient(pid, 8*time.Second).Post(u, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, "", 0, errors.New("device is not reachable")
	}
	defer resp.Body.Close()
	var out struct {
		OK     bool
		Status string
		Left   int
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&out)
	if resp.StatusCode != 200 {
		return false, "expired", 0, nil
	}
	a.mu.Lock()
	if !terminal(s.Status) {
		s.Status = out.Status
	}
	s.Left, s.Updated = out.Left, time.Now()
	a.mu.Unlock()
	return out.OK, out.Status, out.Left, nil
}

// pushFile sends one file of an accepted send to its target.
func (a *App) pushFile(s *Send, idx int, body io.Reader, length int64) error {
	a.mu.Lock()
	kind, offerID := s.Kind, s.OfferID
	pid, addr, rid, sec := s.PeerID, s.PeerAddr, s.RemoteID, s.RemoteSecret
	a.mu.Unlock()
	if kind == "local" {
		a.mu.Lock()
		o := a.offers[offerID]
		a.mu.Unlock()
		if o == nil {
			return errNotAccepted
		}
		return a.receiveStream(o, idx, body)
	}
	u := fmt.Sprintf("https://%s/peer/upload?id=%s&secret=%s&i=%d", addr, url.QueryEscape(rid), url.QueryEscape(sec), idx)
	req, err := http.NewRequest(http.MethodPost, u, body)
	if err != nil {
		return err
	}
	req.ContentLength = length
	if length == 0 {
		req.Body = http.NoBody
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := a.peerClient(pid, 0).Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {
			err = ue.Err // do not leak the transfer secret into the UI
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var out struct{ Error string }
		_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&out)
		if out.Error == "" {
			out.Error = resp.Status
		}
		return errors.New(out.Error)
	}
	return nil
}

func (a *App) cancelSend(s *Send) {
	a.mu.Lock()
	if terminal(s.Status) {
		a.mu.Unlock()
		return
	}
	s.Status, s.Updated = "canceled", time.Now()
	kind, offerID := s.Kind, s.OfferID
	pid, addr, rid, sec := s.PeerID, s.PeerAddr, s.RemoteID, s.RemoteSecret
	if kind == "local" {
		if o := a.offers[offerID]; o != nil && !terminal(o.Status) {
			o.Status, o.Updated = "canceled", time.Now()
		}
	}
	a.mu.Unlock()
	if kind == "peer" {
		u := "https://" + addr + "/peer/cancel?id=" + url.QueryEscape(rid) + "&secret=" + url.QueryEscape(sec)
		if resp, err := a.peerClient(pid, 5*time.Second).Post(u, "application/json", nil); err == nil {
			resp.Body.Close()
		}
	}
}

type countingReader struct {
	r io.Reader
	a *App
	s *Send
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 {
		c.a.mu.Lock()
		c.s.Sent += int64(n)
		c.s.Updated = time.Now()
		canceled := c.s.Status == "canceled"
		c.a.mu.Unlock()
		if canceled {
			return n, errCanceled
		}
	}
	return n, err
}

// runServerSend delivers staged files from this machine's disk.
func (a *App) runServerSend(s *Send, files []Staged) {
	defer func() {
		for _, f := range files {
			if f.Temp {
				a.mu.Lock()
				used := false
				for _, it := range a.inbox {
					if it.Path == f.Abs {
						used = true
					}
				}
				a.mu.Unlock()
				if !used {
					os.Remove(f.Abs)
				}
			}
		}
	}()
	deadline := time.Now().Add(3 * time.Minute)
	for {
		st := a.refreshSend(s)
		if st == "accepted" {
			break
		}
		if terminal(st) || time.Now().After(deadline) {
			a.mu.Lock()
			if !terminal(s.Status) {
				s.Status = "expired"
			}
			a.mu.Unlock()
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	a.mu.Lock()
	s.Status = "accepted"
	a.mu.Unlock()
	for i, f := range files {
		a.mu.Lock()
		canceled := s.Status == "canceled"
		a.mu.Unlock()
		if canceled {
			return
		}
		var err error
		if s.Kind == "local" {
			// guest on this server: hand the file over without copying
			a.mu.Lock()
			o := a.offers[s.OfferID]
			if o != nil && o.Status == "accepted" && !o.Done[i] {
				o.Done[i] = true
				o.Received += f.Size
				o.Updated = time.Now()
				s.Sent += f.Size
				a.fileArrivedLocked(o, i, f.Abs, f.Temp)
			} else {
				err = errNotAccepted
			}
			a.mu.Unlock()
		} else {
			var fh *os.File
			fh, err = os.Open(f.Abs)
			if err == nil {
				err = a.pushFile(s, i, &countingReader{r: fh, a: a, s: s}, f.Size)
				fh.Close()
			}
		}
		if err != nil {
			a.mu.Lock()
			if s.Status != "canceled" {
				s.Status, s.Err = "error", err.Error()
			}
			a.mu.Unlock()
			if s.Kind == "peer" {
				a.cancelSend(s)
			}
			return
		}
	}
	a.mu.Lock()
	s.Status, s.Updated = "done", time.Now()
	a.mu.Unlock()
}

// stagePaths queues files/folders from this machine.
func (a *App) stagePaths(paths []string, temp bool) int {
	var out []Staged
	for _, p := range paths {
		p = filepath.Clean(p)
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !st.IsDir() {
			out = append(out, Staged{Abs: p, Rel: filepath.Base(p), Size: st.Size(), Temp: temp})
			continue
		}
		base := filepath.Base(p)
		_ = filepath.Walk(p, func(fp string, fi os.FileInfo, err error) error {
			if err != nil || len(out) > 20000 {
				return nil
			}
			if fi.Mode().IsRegular() {
				rel, _ := filepath.Rel(p, fp)
				out = append(out, Staged{Abs: fp, Rel: base + "/" + filepath.ToSlash(rel), Size: fi.Size(), Temp: temp})
			}
			return nil
		})
	}
	a.mu.Lock()
	a.staged = append(a.staged, out...)
	if len(a.staged) > 20000 {
		a.staged = a.staged[:20000]
	}
	n := len(a.staged)
	a.mu.Unlock()
	return n
}

// ---------- guest inbox ----------

func (a *App) serveInboxFile(w http.ResponseWriter, r *http.Request, me string) {
	a.mu.Lock()
	it := a.inbox[r.URL.Query().Get("id")]
	a.mu.Unlock()
	if it == nil || it.To != me {
		httpErr(w, 404, "not found")
		return
	}
	f, err := os.Open(it.Path)
	if err != nil {
		httpErr(w, 404, "not found")
		return
	}
	defer f.Close()
	name := filepath.Base(it.Name)
	kind := "attachment"
	if r.URL.Query().Get("inline") == "1" {
		kind = "inline"
	}
	w.Header().Set("Content-Disposition", contentDisposition(kind, name))
	if kind == "attachment" {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, name, it.Created, f)
}

func (a *App) serveInboxZip(w http.ResponseWriter, r *http.Request, me string) {
	offer := r.URL.Query().Get("offer")
	a.mu.Lock()
	var items []*InboxItem
	for _, it := range a.inbox {
		if it.To == me && (offer == "" || it.OfferID == offer) {
			items = append(items, it)
		}
	}
	a.mu.Unlock()
	if len(items) == 0 {
		httpErr(w, 404, "nothing to download")
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", contentDisposition("attachment", "AllDrop-"+time.Now().Format("20060102-150405")+".zip"))
	zw := zip.NewWriter(w)
	used := map[string]int{}
	for _, it := range items {
		f, err := os.Open(it.Path)
		if err != nil {
			continue
		}
		name := it.Name
		if used[name] > 0 {
			ext := filepath.Ext(name)
			name = strings.TrimSuffix(name, ext) + " (" + strconv.Itoa(used[name]+1) + ")" + ext
		}
		used[it.Name]++
		hdr := &zip.FileHeader{Name: name, Method: zip.Store, Modified: it.Created}
		hdr.SetMode(0o644)
		zf, err := zw.CreateHeader(hdr)
		if err == nil {
			bp := bufPool.Get().(*[]byte)
			_, err = io.CopyBuffer(zf, f, *bp)
			bufPool.Put(bp)
		}
		f.Close()
		if err != nil {
			return
		}
	}
	zw.Close()
}
