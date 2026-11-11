package engine

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

//go:embed web
var webFS embed.FS

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func httpErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(v)
}

func securityHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("X-Frame-Options", "DENY")
}

// ---------- auth ----------

type reqCtx struct {
	owner bool
	me    string // "local" or guest client id
}

func (a *App) isOwner(r *http.Request) bool {
	ip := remoteIP(r.RemoteAddr)
	if ip == nil || !ip.IsLoopback() || !hostIsLocal(r.Host) {
		return false
	}
	if ownerToken == "" {
		return true
	}
	// Android: other apps on the phone can reach 127.0.0.1 too, so the
	// owner must also prove it is our own WebView / shell.
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-AD-Owner")), []byte(ownerToken)) == 1 {
		return true
	}
	c, err := r.Cookie("ad_o")
	return err == nil && subtle.ConstantTimeCompare([]byte(c.Value), []byte(ownerToken)) == 1
}

func (a *App) auth(w http.ResponseWriter, r *http.Request, write bool) (*reqCtx, bool) {
	if write {
		if r.Method != http.MethodPost || r.Header.Get("X-AD") != "1" {
			httpErr(w, 403, "forbidden")
			return nil, false
		}
	}
	if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host && o != "null" {
		httpErr(w, 403, "bad origin")
		return nil, false
	}
	if a.isOwner(r) {
		a.mu.Lock()
		a.lastOwnerPoll = time.Now()
		a.mu.Unlock()
		return &reqCtx{owner: true, me: "local"}, true
	}
	k, err := r.Cookie("ad_k")
	a.mu.Lock()
	key := a.cfg.GuestKey
	a.mu.Unlock()
	if err != nil || subtle.ConstantTimeCompare([]byte(k.Value), []byte(key)) != 1 {
		httpErr(w, 401, "auth")
		return nil, false
	}
	id := ""
	if c, err := r.Cookie("ad_c"); err == nil && len(c.Value) == 16 {
		id = c.Value
	}
	if id == "" {
		id = randID(8)
		http.SetCookie(w, &http.Cookie{Name: "ad_c", Value: id, Path: "/", MaxAge: 365 * 86400, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	a.mu.Lock()
	c := a.clients[id]
	if c == nil {
		c = &Client{ID: id, Name: "Guest", OS: "browser"}
		a.clients[id] = c
	}
	c.LastSeen = time.Now()
	a.mu.Unlock()
	return &reqCtx{me: id}, true
}

func (a *App) ownerOnly(w http.ResponseWriter, r *http.Request) bool {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return false
	}
	if !ctx.owner {
		httpErr(w, 403, "owner only")
		return false
	}
	return true
}

// ---------- routing ----------

func (a *App) uiHandler() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(webFS, "web")
	static := http.FileServer(http.FS(sub))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		securityHeaders(w)
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			if ot := r.URL.Query().Get("ot"); ot != "" && ownerToken != "" {
				if subtle.ConstantTimeCompare([]byte(ot), []byte(ownerToken)) == 1 {
					http.SetCookie(w, &http.Cookie{Name: "ad_o", Value: ot, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
				}
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			if k := r.URL.Query().Get("k"); k != "" {
				a.mu.Lock()
				valid := subtle.ConstantTimeCompare([]byte(k), []byte(a.cfg.GuestKey)) == 1
				a.mu.Unlock()
				if valid {
					http.SetCookie(w, &http.Cookie{Name: "ad_k", Value: k, Path: "/", MaxAge: 30 * 86400, HttpOnly: true, SameSite: http.SameSiteLaxMode})
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			}
			b, _ := fs.ReadFile(sub, "index.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.Write(b)
			return
		}
		ext := path.Ext(r.URL.Path)
		if ct := mime.TypeByExtension(ext); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.Header().Set("Cache-Control", "no-cache")
		static.ServeHTTP(w, r)
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if !a.isOwner(r) {
			httpErr(w, 403, "forbidden")
			return
		}
		a.mu.Lock()
		id := a.cfg.ID
		a.mu.Unlock()
		writeJSON(w, 200, map[string]string{"app": "alldrop", "id": id})
	})
	mux.HandleFunc("/api/state", a.apiState)
	mux.HandleFunc("/api/hello", a.apiHello)
	mux.HandleFunc("/api/send", a.apiSend)
	mux.HandleFunc("/api/send/status", a.apiSendStatus)
	mux.HandleFunc("/api/send/file", a.apiSendFile)
	mux.HandleFunc("/api/send/cancel", a.apiSendCancel)
	mux.HandleFunc("/api/send/code", a.apiSendCode)
	mux.HandleFunc("/api/offer/trust", a.apiOfferTrust)
	mux.HandleFunc("/api/send/staged", a.apiSendStaged)
	mux.HandleFunc("/api/offer/respond", a.apiRespond)
	mux.HandleFunc("/api/inbox/file", func(w http.ResponseWriter, r *http.Request) {
		if ctx, ok := a.auth(w, r, false); ok {
			a.serveInboxFile(w, r, ctx.me)
		}
	})
	mux.HandleFunc("/api/inbox/zip", func(w http.ResponseWriter, r *http.Request) {
		if ctx, ok := a.auth(w, r, false); ok {
			a.serveInboxZip(w, r, ctx.me)
		}
	})
	mux.HandleFunc("/api/inbox/remove", a.apiInboxRemove)
	mux.HandleFunc("/api/qr", a.apiQR)
	mux.HandleFunc("/api/settings", a.apiSettings)
	mux.HandleFunc("/api/pickdir", a.apiPickDir)
	mux.HandleFunc("/api/trust/remove", a.apiTrustRemove)
	mux.HandleFunc("/api/key/regen", a.apiKeyRegen)
	mux.HandleFunc("/api/peer/add", a.apiPeerAdd)
	mux.HandleFunc("/api/peer/forget", a.apiPeerForget)
	mux.HandleFunc("/api/scan", a.apiScan)
	mux.HandleFunc("/api/stage", a.apiStage)
	mux.HandleFunc("/api/stage/clear", a.apiStageClear)
	mux.HandleFunc("/api/stage/remove", a.apiStageRemove)
	mux.HandleFunc("/api/open", a.apiOpen)
	mux.HandleFunc("/api/quit", a.apiQuit)
	mux.HandleFunc("/api/win/firewall", a.apiFirewall)
	mux.HandleFunc("/api/win/sendto", a.apiSendTo)
	return mux
}

func (a *App) peerHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/peer/info", a.peerInfoHandler)
	mux.HandleFunc("/peer/offer", a.peerOffer)
	mux.HandleFunc("/peer/status", a.peerStatus)
	mux.HandleFunc("/peer/upload", a.peerUpload)
	mux.HandleFunc("/peer/cancel", a.peerCancel)
	mux.HandleFunc("/peer/code", a.peerCode)
	mux.HandleFunc("/peer/hello", a.peerHello)
	return mux
}

// ---------- state ----------

type offerJSON struct {
	ID       string     `json:"id"`
	From     Party      `json:"from"`
	Files    []FileMeta `json:"files"`
	Count    int        `json:"count"`
	Total    int64      `json:"total"`
	Received int64      `json:"received"`
	Status   string     `json:"status"`
	Trusted  bool       `json:"trusted"`
	Mode     string     `json:"mode"`
	Code     string     `json:"code"`
	Trust    bool       `json:"trust"`
}

type inboxJSON struct {
	ID      string `json:"id"`
	Offer   string `json:"offer"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	From    string `json:"from"`
	Created int64  `json:"created"`
}

type sendJSON struct {
	ID     string `json:"id"`
	ToName string `json:"toName"`
	Status string `json:"status"`
	Err    string `json:"err"`
	Count  int    `json:"count"`
	Total  int64  `json:"total"`
	Sent   int64  `json:"sent"`
	First  string `json:"first"`
	Mode   string `json:"mode"`
	Left   int    `json:"left"`
}

func (a *App) joinURLsLocked() []string {
	var out []string
	for _, la := range localIPv4(a.hintIPs) {
		out = append(out, fmt.Sprintf("http://%s:%d/?k=%s", la.IP, a.httpPort, a.cfg.GuestKey))
	}
	return out
}

func (a *App) apiState(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, false)
	if !ok {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	st := map[string]interface{}{}
	me := map[string]interface{}{"id": ctx.me, "owner": ctx.owner}
	if ctx.owner {
		me["name"], me["os"] = a.cfg.Name, a.platform
	} else if c := a.clients[ctx.me]; c != nil {
		me["name"], me["os"] = c.Name, c.OS
	}
	st["me"] = me
	host := map[string]interface{}{"name": a.cfg.Name, "os": a.platform, "version": appVersion}
	if ctx.owner {
		host["port"] = a.httpPort
		host["urls"] = a.joinURLsLocked()
		host["downloadDir"] = a.recvDirLocked()
		host["dirOK"] = a.dirOK
		host["autoAccept"] = a.cfg.AutoAccept
		host["requireCode"] = !a.cfg.NoCode
		host["goos"] = runtime.GOOS
		host["android"] = a.platform == "android"
		host["mobile"] = mobileMode
		host["scanning"] = a.scanning
		var trusted []Party
		for id, n := range a.cfg.Trusted {
			trusted = append(trusted, Party{ID: id, Name: n})
		}
		host["trusted"] = trusted
		host["manual"] = a.cfg.ManualPeers
		hist := a.history
		if len(hist) > 300 {
			hist = hist[:300]
		}
		st["history"] = hist
		var total int64
		var names []string
		for i, s := range a.staged {
			total += s.Size
			if i < 4 {
				names = append(names, s.Rel)
			}
		}
		groups := a.stagedGroupsLocked()
		if len(groups) > 200 {
			groups = groups[:200]
		}
		st["staged"] = map[string]interface{}{"count": len(a.staged), "total": total, "names": names, "groups": groups}
		var sends []sendJSON
		for _, s := range a.sends {
			if s.Server && time.Since(s.Updated) < 10*time.Minute {
				first := ""
				if len(s.Files) > 0 {
					first = s.Files[0].Name
				}
				sends = append(sends, sendJSON{ID: s.ID, ToName: s.ToName, Status: s.Status, Err: s.Err, Count: len(s.Files), Total: s.Total, Sent: s.Sent, First: first, Mode: s.Mode, Left: s.Left})
			}
		}
		st["sends"] = sends
	}
	st["host"] = host
	st["devices"] = a.devicesForLocked(ctx.me)
	var incoming []offerJSON
	for _, o := range a.offers {
		if o.To != ctx.me {
			continue
		}
		if terminal(o.Status) && time.Since(o.Updated) > 8*time.Second {
			continue
		}
		files := o.Files
		if len(files) > 12 {
			files = files[:12]
		}
		_, trusted := a.cfg.Trusted[o.From.ID]
		incoming = append(incoming, offerJSON{ID: o.ID, From: o.From, Files: files, Count: len(o.Files), Total: o.Total, Received: o.Received, Status: o.Status, Trusted: trusted, Mode: o.Mode, Code: o.Code, Trust: o.TrustOnOK})
	}
	st["incoming"] = incoming
	if !ctx.owner {
		var inbox []inboxJSON
		for _, it := range a.inbox {
			if it.To == ctx.me {
				inbox = append(inbox, inboxJSON{ID: it.ID, Offer: it.OfferID, Name: it.Name, Size: it.Size, From: it.From, Created: it.Created.UnixMilli()})
			}
		}
		st["inbox"] = inbox
	}
	writeJSON(w, 200, st)
}

func (a *App) apiHello(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return
	}
	var req struct{ Name, OS string }
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	if !ctx.owner {
		a.mu.Lock()
		if c := a.clients[ctx.me]; c != nil {
			c.Name = cleanName(req.Name, c.Name)
			c.OS = cleanOS(req.OS)
		}
		a.mu.Unlock()
	}
	writeJSON(w, 200, map[string]string{"id": ctx.me})
}

// ---------- sending ----------

func (a *App) apiSend(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return
	}
	var req struct {
		To    string     `json:"to"`
		Files []FileMeta `json:"files"`
	}
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	s, err := a.startSend(ctx.me, req.To, req.Files, false)
	if err != nil {
		httpErr(w, 409, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"id": s.ID})
}

func (a *App) sendFor(ctx *reqCtx, r *http.Request) *Send {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.sends[r.URL.Query().Get("id")]
	if s == nil || s.ClientID != ctx.me {
		return nil
	}
	return s
}

func (a *App) apiSendStatus(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, false)
	if !ok {
		return
	}
	s := a.sendFor(ctx, r)
	if s == nil {
		httpErr(w, 404, "unknown transfer")
		return
	}
	st := a.refreshSend(s)
	a.mu.Lock()
	out := map[string]interface{}{"status": st, "mode": s.Mode, "left": s.Left}
	a.mu.Unlock()
	writeJSON(w, 200, out)
}

func (a *App) apiSendCode(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return
	}
	s := a.sendFor(ctx, r)
	if s == nil {
		httpErr(w, 404, "unknown transfer")
		return
	}
	var req struct{ Code string }
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	good, st, left, err := a.submitCode(s, req.Code)
	if err != nil {
		httpErr(w, 502, err.Error())
		return
	}
	writeJSON(w, 200, map[string]interface{}{"ok": good, "status": st, "left": left})
}

func (a *App) apiOfferTrust(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct {
		ID    string `json:"id"`
		Trust bool   `json:"trust"`
	}
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	if o := a.offers[req.ID]; o != nil && o.To == "local" {
		o.TrustOnOK = req.Trust
	}
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiSendFile(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return
	}
	s := a.sendFor(ctx, r)
	if s == nil {
		httpErr(w, 404, "unknown transfer")
		return
	}
	idx, err := strconv.Atoi(r.URL.Query().Get("i"))
	a.mu.Lock()
	bad := err != nil || idx < 0 || idx >= len(s.Files)
	var size int64
	if !bad {
		size = s.Files[idx].Size
	}
	accepted := s.Status == "accepted"
	a.mu.Unlock()
	if bad {
		httpErr(w, 400, "bad index")
		return
	}
	if !accepted && a.refreshSend(s) != "accepted" {
		httpErr(w, 409, "not accepted")
		return
	}
	if err := a.pushFile(s, idx, r.Body, size); err != nil {
		httpErr(w, 409, err.Error())
		return
	}
	a.mu.Lock()
	s.Sent += size
	s.Updated = time.Now()
	if s.Sent >= s.Total && s.Kind == "peer" {
		s.Status = "done"
	}
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiSendCancel(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return
	}
	s := a.sendFor(ctx, r)
	if s == nil {
		httpErr(w, 404, "unknown transfer")
		return
	}
	a.cancelSend(s)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiSendStaged(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct{ To string }
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	files := a.staged
	a.staged = nil
	a.mu.Unlock()
	if len(files) == 0 {
		httpErr(w, 400, "nothing staged")
		return
	}
	metas := make([]FileMeta, len(files))
	for i, f := range files {
		metas[i] = FileMeta{Name: path.Base(f.Rel), Path: f.Rel, Size: f.Size}
		if !strings.Contains(f.Rel, "/") {
			metas[i].Path = ""
		}
	}
	s, err := a.startSend("local", req.To, metas, true)
	if err != nil {
		a.mu.Lock()
		a.staged = append(files, a.staged...)
		a.mu.Unlock()
		httpErr(w, 409, err.Error())
		return
	}
	go a.runServerSend(s, files)
	writeJSON(w, 200, map[string]string{"id": s.ID})
}

func (a *App) apiRespond(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return
	}
	var req struct {
		ID     string `json:"id"`
		Accept bool   `json:"accept"`
		Trust  bool   `json:"trust"`
	}
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	o := a.offers[req.ID]
	if o == nil || o.To != ctx.me {
		httpErr(w, 404, "unknown request")
		return
	}
	if req.Accept && o.Mode == "code" && o.Status == "pending" {
		httpErr(w, 409, "the sender must type the code")
		return
	}
	if req.Accept && (o.Status == "pending" || o.Status == "accepted") {
		o.Status = "accepted"
		if req.Trust && ctx.owner && o.From.ID != "" {
			a.cfg.Trusted[o.From.ID] = o.From.Name
			a.saveLocked()
		}
	} else if !req.Accept && !terminal(o.Status) {
		if o.Status == "accepted" {
			o.Status = "canceled"
		} else {
			o.Status = "declined"
		}
	}
	o.Updated = time.Now()
	writeJSON(w, 200, map[string]string{"status": o.Status})
}

func (a *App) apiInboxRemove(w http.ResponseWriter, r *http.Request) {
	ctx, ok := a.auth(w, r, true)
	if !ok {
		return
	}
	var req struct {
		ID  string `json:"id"`
		All bool   `json:"all"`
	}
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	for _, it := range a.inbox {
		if it.To == ctx.me && (req.All || it.ID == req.ID) {
			a.removeInboxLocked(it)
		}
	}
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---------- owner settings & tools ----------

func (a *App) apiQR(w http.ResponseWriter, r *http.Request) {
	if !a.isOwner(r) {
		httpErr(w, 403, "owner only")
		return
	}
	i, _ := strconv.Atoi(r.URL.Query().Get("i"))
	a.mu.Lock()
	urls := a.joinURLsLocked()
	a.mu.Unlock()
	if i < 0 || i >= len(urls) {
		httpErr(w, 404, "no network address")
		return
	}
	png, err := qrcode.Encode(urls[i], qrcode.Medium, 480)
	if err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(png)
}

func (a *App) apiSettings(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct {
		Name        *string `json:"name"`
		DownloadDir *string `json:"downloadDir"`
		AutoAccept  *bool   `json:"autoAccept"`
		RequireCode *bool   `json:"requireCode"`
	}
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	if req.DownloadDir != nil {
		if err := a.SetDownloadDir(*req.DownloadDir); err != nil {
			httpErr(w, 400, "folder is not writable")
			return
		}
	}
	a.mu.Lock()
	if req.Name != nil {
		a.cfg.Name = cleanName(*req.Name, a.cfg.Name)
	}
	if req.AutoAccept != nil {
		a.cfg.AutoAccept = *req.AutoAccept
	}
	if req.RequireCode != nil {
		a.cfg.NoCode = !*req.RequireCode
	}
	a.saveLocked()
	a.mu.Unlock()
	if req.Name != nil {
		go a.announce("hello")
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiTrustRemove(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct{ ID string }
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	delete(a.cfg.Trusted, req.ID)
	a.saveLocked()
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiKeyRegen(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	a.mu.Lock()
	a.cfg.GuestKey = randID(8)
	a.clients = map[string]*Client{}
	for _, it := range a.inbox {
		a.removeInboxLocked(it)
	}
	a.saveLocked()
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiPeerAdd(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct{ Addr string }
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	addr, err := parseManualAddr(req.Addr)
	if err != nil {
		httpErr(w, 400, "bad address")
		return
	}
	if !a.verifyAndAdd(addr, "", true, true) {
		httpErr(w, 404, "no AllDrop device answered at that address")
		return
	}
	a.mu.Lock()
	found := false
	for _, m := range a.cfg.ManualPeers {
		if m == addr {
			found = true
		}
	}
	if !found {
		a.cfg.ManualPeers = append(a.cfg.ManualPeers, addr)
		a.saveLocked()
	}
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiPeerForget(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct{ Addr string }
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	var keep []string
	for _, m := range a.cfg.ManualPeers {
		if m != req.Addr {
			keep = append(keep, m)
		}
	}
	a.cfg.ManualPeers = keep
	for id, p := range a.peers {
		if p.Addr == req.Addr {
			delete(a.peers, id)
		}
	}
	a.saveLocked()
	a.mu.Unlock()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiScan(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	go a.announce("hello")
	go a.scan()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiStage(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct {
		Paths []string `json:"paths"`
		Temp  bool     `json:"temp"`
	}
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	n := a.stagePaths(req.Paths, req.Temp)
	writeJSON(w, 200, map[string]int{"count": n})
}

func (a *App) apiStageClear(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	a.mu.Lock()
	files := a.staged
	a.staged = nil
	a.mu.Unlock()
	for _, f := range files {
		if f.Temp {
			os.Remove(f.Abs)
		}
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiStageRemove(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct {
		Name string `json:"name"`
		Dir  bool   `json:"dir"`
	}
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.unstage(req.Name, req.Dir)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiOpen(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	var req struct{ Kind, Path string }
	if readJSON(r, &req) != nil {
		httpErr(w, 400, "bad json")
		return
	}
	a.mu.Lock()
	dir := a.recvDirLocked()
	a.mu.Unlock()
	var err error
	switch req.Kind {
	case "folder":
		os.MkdirAll(dir, 0o755)
		err = openFolder(dir)
	case "file", "reveal":
		p := filepath.Clean(req.Path)
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil || strings.HasPrefix(rel, "..") || !exists(p) {
			httpErr(w, 404, "not found")
			return
		}
		if req.Kind == "file" {
			err = OpenURL(p)
		} else {
			err = revealFile(p)
		}
	default:
		httpErr(w, 400, "bad kind")
		return
	}
	if err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiQuit(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
	go func() {
		time.Sleep(300 * time.Millisecond)
		a.quit()
	}()
}

func (a *App) apiFirewall(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	if err := addFirewallRule(); err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) apiSendTo(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	if err := installSendTo(); err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// SetDownloadDir changes the folder that received files are saved to.
func (a *App) SetDownloadDir(dir string) error {
	d := strings.TrimSpace(dir)
	if strings.HasPrefix(d, "~") {
		if h, err := os.UserHomeDir(); err == nil {
			d = filepath.Join(h, strings.TrimPrefix(d, "~"))
		}
	}
	if d == "" || !filepath.IsAbs(d) || !dirWritable(d) {
		return fmt.Errorf("folder is not writable")
	}
	a.mu.Lock()
	a.cfg.DownloadDir = filepath.Clean(d)
	a.dirOK = true
	a.saveLocked()
	a.mu.Unlock()
	return nil
}

// apiPickDir shows the computer's own folder chooser and, if the owner picks
// a folder, makes it the place where received files are saved.
func (a *App) apiPickDir(w http.ResponseWriter, r *http.Request) {
	if !a.ownerOnly(w, r) {
		return
	}
	if mobileMode {
		httpErr(w, 404, "not available")
		return
	}
	a.mu.Lock()
	cur := a.recvDirLocked()
	a.mu.Unlock()
	d, err := pickFolder(cur)
	if err != nil {
		httpErr(w, 500, "folder chooser unavailable")
		return
	}
	if d == "" { // cancelled
		w.WriteHeader(204)
		return
	}
	if err := a.SetDownloadDir(d); err != nil {
		httpErr(w, 400, "folder is not writable")
		return
	}
	writeJSON(w, 200, map[string]string{"path": d})
}
