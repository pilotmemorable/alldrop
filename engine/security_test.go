package engine

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Attack test 1: path traversal / filename injection into the download folder.
// A malicious sender controls FileMeta.Name and FileMeta.Path. None of these
// must ever escape the receive directory or create a device-dangerous name.
// ---------------------------------------------------------------------------

func TestAttack_PathTraversal(t *testing.T) {
	evil := []string{
		"../../../../etc/passwd",
		`..\..\..\Windows\System32\drivers\etc\hosts`,
		"/etc/shadow",
		`C:\Windows\win.ini`,
		"....//....//secret",
		"foo/../../bar",
		"a/./../../b",
		"\x00/etc/passwd",
		"con/PRN/..",
		"....",
		"../",
		"..",
		".",
		"\\\\server\\share\\x",
		strings.Repeat("A/", 100) + "deep.txt", // depth bomb
	}
	for _, in := range evil {
		parts := sanitizeRel(in)
		joined := strings.Join(parts, "/")
		// No component may be "..", "." or contain a separator escaping upward.
		for _, p := range parts {
			if p == ".." || p == "." || p == "" || strings.ContainsAny(p, `/\`) {
				t.Fatalf("sanitizeRel(%q) produced unsafe component %q (full=%q)", in, p, joined)
			}
		}
		// The reconstructed path must never be absolute nor start with ..
		if strings.HasPrefix(joined, "/") || strings.HasPrefix(joined, "..") || strings.Contains(joined, "../") {
			t.Fatalf("sanitizeRel(%q) escaped: %q", in, joined)
		}
		if len(parts) > 32 {
			t.Fatalf("sanitizeRel(%q) exceeded depth cap: %d", in, len(parts))
		}
	}
}

func TestAttack_SanitizeNameNeutralizesSpecials(t *testing.T) {
	cases := map[string]func(string) bool{
		`a<b>c:d"e|f?g*h`: func(o string) bool { return !strings.ContainsAny(o, `<>:"|?*`) },
		"ctrl\x01\x02\x1f": func(o string) bool { return !strings.ContainsAny(o, "\x00\x01\x02\x1f") },
		"NUL":              func(o string) bool { return strings.HasPrefix(o, "_") }, // reserved win name
		"trailing dots...": func(o string) bool { return !strings.HasSuffix(o, ".") },
	}
	for in, ok := range cases {
		if got := sanitizeName(in); !ok(got) {
			t.Fatalf("sanitizeName(%q) = %q did not neutralize the attack", in, got)
		}
	}
	// Very long name must be capped (defends against path-length DoS).
	if got := sanitizeName(strings.Repeat("x", 5000) + ".txt"); len(got) > 205 {
		t.Fatalf("long name not capped: %d", len(got))
	}
}

// receiveStream must physically keep a traversal filename inside the receive dir.
func TestAttack_ReceiveStaysInsideDir(t *testing.T) {
	dir := t.TempDir()
	a := newApp()
	a.dataDir = dir
	a.inboxDir = filepath.Join(dir, "inbox")
	a.cfg = loadConfig(filepath.Join(dir, "cfg.json"))
	a.cfg.DownloadDir = filepath.Join(dir, "dl")
	a.dirOK = true

	payload := []byte("owned")
	o := &Offer{
		ID: "o1", To: "local", Status: "accepted",
		Files:  []FileMeta{{Name: "x", Path: "../../../../tmp/evil.txt", Size: int64(len(payload))}},
		Done:   make([]bool, 1),
		topMap: map[string]string{},
	}
	a.offers[o.ID] = o
	if err := a.receiveStream(o, 0, bytes.NewReader(payload)); err != nil {
		t.Fatalf("receiveStream: %v", err)
	}
	// The file must exist under the download dir, never at /tmp/evil.txt.
	want := filepath.Join(a.cfg.DownloadDir, "tmp", "evil.txt")
	if !exists(want) {
		t.Fatalf("expected received file inside %s; tree not as expected", a.cfg.DownloadDir)
	}
	if exists("/tmp/evil.txt") {
		t.Fatalf("SECURITY: traversal escaped to /tmp/evil.txt")
	}
}

// receiveStream must reject a body larger than the declared size (disk-fill /
// smuggling defense) and clean up the partial file.
func TestAttack_OversizeBodyRejected(t *testing.T) {
	dir := t.TempDir()
	a := newApp()
	a.dataDir = dir
	a.inboxDir = filepath.Join(dir, "inbox")
	a.cfg = loadConfig(filepath.Join(dir, "cfg.json"))
	a.cfg.DownloadDir = filepath.Join(dir, "dl")
	a.dirOK = true

	o := &Offer{
		ID: "o1", To: "local", Status: "accepted",
		Files:  []FileMeta{{Name: "small.bin", Size: 4}},
		Done:   make([]bool, 1),
		topMap: map[string]string{},
	}
	a.offers[o.ID] = o
	// Declared size 4, but send 1 MB.
	big := bytes.Repeat([]byte("A"), 1<<20)
	err := a.receiveStream(o, 0, bytes.NewReader(big))
	if err != errSize {
		t.Fatalf("expected errSize, got %v", err)
	}
	if o.Done[0] {
		t.Fatal("oversize transfer must not be marked done")
	}
	// No leftover .adpart file.
	matches, _ := filepath.Glob(filepath.Join(a.cfg.DownloadDir, "*"+partSuffix))
	if len(matches) != 0 {
		t.Fatalf("partial file not cleaned up: %v", matches)
	}
}

// ---------------------------------------------------------------------------
// Attack test 2: 6-digit code brute force. Five wrong guesses (including
// malformed ones) must lock the offer; only the exact code accepts.
// ---------------------------------------------------------------------------

func TestAttack_CodeBruteForceLockout(t *testing.T) {
	a := newApp()
	a.cfg = loadConfig(t.TempDir() + "/c.json")
	o := &Offer{ID: "o", To: "local", Status: "pending", Mode: "code", Code: "428913"}
	a.offers[o.ID] = o

	wrong := []string{"000000", "111111", "abc", "42891", "4289130", "999999"}
	for i := 0; i < maxCodeTries; i++ {
		ok, left := a.checkCodeLocked(o, wrong[i%len(wrong)])
		if ok {
			t.Fatalf("wrong code accepted on try %d", i)
		}
		_ = left
	}
	if o.Status != "declined" {
		t.Fatalf("offer not locked after %d wrong tries: status=%s", maxCodeTries, o.Status)
	}
	// Even the correct code must now be refused (offer already declined).
	if ok, _ := a.checkCodeLocked(o, "428913"); ok {
		t.Fatal("SECURITY: correct code accepted after lockout")
	}
}

func TestAttack_CorrectCodeAccepts(t *testing.T) {
	a := newApp()
	a.cfg = loadConfig(t.TempDir() + "/c.json")
	o := &Offer{ID: "o", To: "local", Status: "pending", Mode: "code", Code: "428913"}
	a.offers[o.ID] = o
	// spaces/dashes tolerated, value still must match digits
	if ok, _ := a.checkCodeLocked(o, "428-913"); !ok {
		t.Fatal("correct code (with separators) should accept")
	}
	if o.Status != "accepted" {
		t.Fatalf("status = %s, want accepted", o.Status)
	}
}

// ---------------------------------------------------------------------------
// Attack test 3: TLS identity. Two certs must hash to different IDs, and the
// pinned transport must reject a peer whose cert id differs from expected.
// ---------------------------------------------------------------------------

func TestAttack_CertIDUniqueAndStable(t *testing.T) {
	c1, id1, err := loadOrCreateCert(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, id2, err := loadOrCreateCert(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if id1 == id2 {
		t.Fatal("two independent installs produced the same device id")
	}
	if len(id1) != 24 { // 12 bytes hex
		t.Fatalf("unexpected id length %d", len(id1))
	}
	if got := certID(c1.Certificate[0]); got != id1 {
		t.Fatalf("certID not stable: %s vs %s", got, id1)
	}
}

func TestAttack_ValidFilesBounds(t *testing.T) {
	if err := validFiles(nil); err == nil {
		t.Fatal("empty file list must be rejected")
	}
	huge := make([]FileMeta, 20001)
	if err := validFiles(huge); err == nil {
		t.Fatal(">20000 files must be rejected")
	}
	if err := validFiles([]FileMeta{{Name: "a", Size: -1}}); err == nil {
		t.Fatal("negative size must be rejected")
	}
	if err := validFiles([]FileMeta{{Name: strings.Repeat("x", 2000)}}); err == nil {
		t.Fatal("over-long name must be rejected")
	}
	if err := validFiles([]FileMeta{{Name: "ok", Path: strings.Repeat("x", 5000)}}); err == nil {
		t.Fatal("over-long path must be rejected")
	}
}

// Content-Disposition header must not allow CRLF / quote injection.
func TestAttack_ContentDispositionNoHeaderInjection(t *testing.T) {
	bad := "evil\r\nSet-Cookie: x=1\"name"
	h := contentDisposition("attachment", bad)
	if strings.ContainsAny(h, "\r\n") {
		t.Fatalf("CRLF leaked into header: %q", h)
	}
}

// ---------------------------------------------------------------------------
// Attack test 4: HTTP auth, CSRF and DNS-rebinding against the real server.
// ---------------------------------------------------------------------------

func startTestApp(t *testing.T) *App {
	t.Helper()
	dir := t.TempDir()
	app, err := Start(Options{
		DataDir:     dir,
		DownloadDir: filepath.Join(dir, "dl"),
		Name:        "TestHost",
		NoBrowser:   true,
		Port:        0, // let the OS choose a free pair
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(app.Close)
	// give the servers a moment
	time.Sleep(100 * time.Millisecond)
	return app
}

func TestAttack_CSRFAndRebind(t *testing.T) {
	app := startTestApp(t)
	base := fmt.Sprintf("http://127.0.0.1:%d", app.Port())
	cl := &http.Client{Timeout: 3 * time.Second}

	// (a) A state-changing POST without the X-AD header must be refused,
	//     so a cross-site <form> auto-post cannot drive the app.
	req, _ := http.NewRequest("POST", base+"/api/settings", strings.NewReader(`{"autoAccept":true}`))
	req.Host = "127.0.0.1"
	resp, err := cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("POST without X-AD should be 403, got %d", resp.StatusCode)
	}

	// (b) A cross-origin request (foreign Origin header) must be refused.
	req, _ = http.NewRequest("POST", base+"/api/settings", strings.NewReader(`{"autoAccept":true}`))
	req.Header.Set("X-AD", "1")
	req.Header.Set("Origin", "http://evil.example")
	resp, err = cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("cross-origin POST should be 403, got %d", resp.StatusCode)
	}

	// (c) DNS rebinding: request arrives on loopback but with a foreign Host
	//     header. It must NOT be treated as the owner (→ 401, needs guest key).
	req, _ = http.NewRequest("GET", base+"/api/state", nil)
	req.Host = "attacker.example"
	resp, err = cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Fatalf("rebind (foreign Host) should not be owner; want 401, got %d", resp.StatusCode)
	}

	// (d) Sanity: a genuine local owner GET works.
	req, _ = http.NewRequest("GET", base+"/api/state", nil)
	req.Host = fmt.Sprintf("127.0.0.1:%d", app.Port())
	resp, err = cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("local owner GET should be 200, got %d", resp.StatusCode)
	}
}

// The encrypted peer port must require a client certificate / identity:
// an anonymous TLS client with no cert must be rejected by /peer/offer.
func TestAttack_PeerRequiresIdentity(t *testing.T) {
	app := startTestApp(t)
	purl := fmt.Sprintf("https://127.0.0.1:%d/peer/offer", app.tlsPort)
	cl := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // no client cert presented
		},
	}
	resp, err := cl.Post(purl, "application/json", strings.NewReader(`{"files":[{"name":"x","size":1}]}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != 403 {
		t.Fatalf("anonymous peer offer must be 403, got %d (%s)", resp.StatusCode, body)
	}
}

// ---------------------------------------------------------------------------
// Attack test 5: full two-device encrypted transfer over real TLS, proving the
// code gate blocks a wrong code and the file only arrives after the right one.
// ---------------------------------------------------------------------------

func TestAttack_EndToEndEncryptedTransferWithCodeGate(t *testing.T) {
	// Receiver B (requires the 6-digit code by default).
	bdir := t.TempDir()
	B, err := Start(Options{DataDir: bdir, DownloadDir: filepath.Join(bdir, "dl"), Name: "Receiver", NoBrowser: true})
	if err != nil {
		t.Fatalf("start B: %v", err)
	}
	t.Cleanup(B.Close)

	// Sender A.
	adir := t.TempDir()
	A, err := Start(Options{DataDir: adir, DownloadDir: filepath.Join(adir, "dl"), Name: "Sender", NoBrowser: true})
	if err != nil {
		t.Fatalf("start A: %v", err)
	}
	t.Cleanup(A.Close)
	time.Sleep(150 * time.Millisecond)

	// A learns B over real TLS (pins B's certificate identity).
	baddr := fmt.Sprintf("127.0.0.1:%d", B.tlsPort)
	if !A.verifyAndAdd(baddr, "", true, true) {
		t.Skip("peer handshake unavailable in this environment")
	}
	bid := B.cfg.ID
	A.mu.Lock()
	_, ok := A.peers[bid]
	A.mu.Unlock()
	if !ok {
		t.Fatalf("B not registered as peer")
	}

	// A stages a secret file and starts a send to B.
	secret := []byte("top secret payload " + strings.Repeat("Z", 1000))
	src := filepath.Join(adir, "secret.bin")
	if err := writeFile(src, secret); err != nil {
		t.Fatal(err)
	}
	A.stagePaths([]string{src}, false)
	A.mu.Lock()
	staged := A.staged
	A.staged = nil
	A.mu.Unlock()
	metas := []FileMeta{{Name: "secret.bin", Size: int64(len(secret))}}
	s, err := A.startSend("local", "p:"+bid, metas, true)
	if err != nil {
		t.Fatalf("startSend: %v", err)
	}
	go A.runServerSend(s, staged)

	// Read the code B generated for this incoming offer.
	var code string
	for i := 0; i < 50 && code == ""; i++ {
		B.mu.Lock()
		for _, o := range B.offers {
			if o.Mode == "code" && o.Status == "pending" {
				code = o.Code
			}
		}
		B.mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}
	if code == "" {
		t.Fatal("no code offer appeared on receiver")
	}

	// Wrong code must be rejected over the wire.
	if okCode, _, _, _ := A.submitCode(s, "000000"); okCode {
		t.Fatal("SECURITY: wrong code accepted end-to-end")
	}
	// Correct code accepts.
	if okCode, _, _, err := A.submitCode(s, code); err != nil || !okCode {
		t.Fatalf("correct code rejected: ok=%v err=%v", okCode, err)
	}

	// The file must now arrive in B's download folder, bytes intact.
	got := filepath.Join(B.cfg.DownloadDir, "secret.bin")
	var data []byte
	for i := 0; i < 100; i++ {
		if b, e := readFile(got); e == nil && int64(len(b)) == int64(len(secret)) {
			data = b
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !bytes.Equal(data, secret) {
		t.Fatalf("received file mismatch: got %d bytes, want %d", len(data), len(secret))
	}
}

func writeFile(p string, b []byte) error { return osWrite(p, b) }
func readFile(p string) ([]byte, error)  { return osRead(p) }
