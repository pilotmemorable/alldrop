package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Several loose files and two whole folders (with sub folders) go to another
// device in ONE transfer, behind ONE security code, and arrive with the same
// folder structure.
func TestMultiFilesAndFoldersInOneTransfer(t *testing.T) {
	bdir := t.TempDir()
	B, err := Start(Options{DataDir: bdir, DownloadDir: filepath.Join(bdir, "dl"), Name: "Receiver", NoBrowser: true})
	if err != nil {
		t.Fatalf("start B: %v", err)
	}
	t.Cleanup(B.Close)
	adir := t.TempDir()
	A, err := Start(Options{DataDir: adir, DownloadDir: filepath.Join(adir, "dl"), Name: "Sender", NoBrowser: true})
	if err != nil {
		t.Fatalf("start A: %v", err)
	}
	t.Cleanup(A.Close)
	time.Sleep(150 * time.Millisecond)
	if !A.verifyAndAdd(fmt.Sprintf("127.0.0.1:%d", B.tlsPort), "", true, true) {
		t.Skip("peer handshake unavailable in this environment")
	}

	src := filepath.Join(adir, "src")
	files := map[string]string{
		"Photos/2024/beach.jpg":   strings.Repeat("J", 70000),
		"Photos/2024/Mayıs/a.png": "png-bytes",
		"Photos/notes.txt":        "photo notes",
		"Docs/report.pdf":         strings.Repeat("P", 12345),
		"loose1.txt":              "one",
		"loose two.mov":           strings.Repeat("M", 4096),
	}
	for rel, body := range files {
		p := filepath.Join(src, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := writeFile(p, []byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	A.stagePaths([]string{
		filepath.Join(src, "Photos"), filepath.Join(src, "Docs"),
		filepath.Join(src, "loose1.txt"), filepath.Join(src, "loose two.mov"),
	}, false)

	// The tray shows one row per top-level item, not one per file.
	A.mu.Lock()
	groups := A.stagedGroupsLocked()
	staged := A.staged
	A.staged = nil
	A.mu.Unlock()
	if len(staged) != len(files) {
		t.Fatalf("staged %d files, want %d", len(staged), len(files))
	}
	want := map[string]int{"d:Photos": 3, "d:Docs": 1, "f:loose1.txt": 1, "f:loose two.mov": 1}
	if len(groups) != len(want) {
		t.Fatalf("groups = %+v", groups)
	}
	for _, g := range groups {
		k := "f:" + g.Name
		if g.Dir {
			k = "d:" + g.Name
		}
		if want[k] != g.Count {
			t.Fatalf("group %s has %d files, want %d", k, g.Count, want[k])
		}
	}

	metas := make([]FileMeta, len(staged))
	for i, f := range staged {
		metas[i] = FileMeta{Name: path.Base(f.Rel), Path: f.Rel, Size: f.Size}
		if !strings.Contains(f.Rel, "/") {
			metas[i].Path = ""
		}
	}
	s, err := A.startSend("local", "p:"+B.cfg.ID, metas, true)
	if err != nil {
		t.Fatalf("startSend: %v", err)
	}
	go A.runServerSend(s, staged)

	var code string
	offers := 0
	for i := 0; i < 60 && code == ""; i++ {
		B.mu.Lock()
		offers = 0
		for _, o := range B.offers {
			offers++
			if o.Mode == "code" && o.Status == "pending" {
				code = o.Code
			}
		}
		B.mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}
	if code == "" || offers != 1 {
		t.Fatalf("want exactly one code offer for the whole batch, got %d (code %q)", offers, code)
	}
	if ok, _, _, err := A.submitCode(s, code); err != nil || !ok {
		t.Fatalf("code rejected: %v %v", ok, err)
	}

	dl := B.cfg.DownloadDir
	deadline := time.Now().Add(10 * time.Second)
	for rel, body := range files {
		p := filepath.Join(dl, filepath.FromSlash(rel))
		for {
			b, e := os.ReadFile(p)
			if e == nil && bytes.Equal(b, []byte(body)) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("%s did not arrive intact", rel)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// History groups the folder files under their top-level folder.
	B.mu.Lock()
	hist := append([]HistoryItem(nil), B.history...)
	B.mu.Unlock()
	photos := 0
	for _, h := range hist {
		if h.Group == filepath.Join(dl, "Photos") {
			photos++
		}
		if h.Offer == "" {
			t.Fatalf("history item %s has no offer id", h.Name)
		}
	}
	if photos != 3 {
		t.Fatalf("history: %d items grouped under Photos, want 3", photos)
	}
}

// Removing one row from the tray is owner-only and only drops that row.
func TestStageRemoveOwnerOnly(t *testing.T) {
	app := startTestApp(t)
	dir := t.TempDir()
	for _, rel := range []string{"Album/1.jpg", "Album/sub/2.jpg", "keep.txt"} {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		writeFile(p, []byte(rel))
	}
	app.stagePaths([]string{filepath.Join(dir, "Album"), filepath.Join(dir, "keep.txt")}, false)

	base := fmt.Sprintf("http://127.0.0.1:%d", app.Port())
	post := func(host string) int {
		req, _ := http.NewRequest("POST", base+"/api/stage/remove", strings.NewReader(`{"name":"Album","dir":true}`))
		req.Host = host
		req.Header.Set("X-AD", "1")
		req.Header.Set("Content-Type", "application/json")
		resp, err := (&http.Client{Timeout: 3 * time.Second}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if c := post("attacker.example"); c == 200 {
		t.Fatal("SECURITY: non-owner removed staged files")
	}
	if c := post(fmt.Sprintf("127.0.0.1:%d", app.Port())); c != 200 {
		t.Fatalf("owner remove: %d", c)
	}

	req, _ := http.NewRequest("GET", base+"/api/state", nil)
	req.Host = fmt.Sprintf("127.0.0.1:%d", app.Port())
	resp, err := (&http.Client{Timeout: 3 * time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var st struct {
		Staged struct {
			Count  int           `json:"count"`
			Groups []stagedGroup `json:"groups"`
		} `json:"staged"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatal(err)
	}
	if st.Staged.Count != 1 || len(st.Staged.Groups) != 1 || st.Staged.Groups[0].Name != "keep.txt" {
		t.Fatalf("after remove: %+v", st.Staged)
	}
}
