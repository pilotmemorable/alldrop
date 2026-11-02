package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeRel(t *testing.T) {
	cases := map[string]string{
		"../../etc/passwd":      "etc/passwd",
		`..\..\Windows\win.ini`: "Windows/win.ini",
		"Photos/CON.txt":        "Photos/_CON.txt",
		"a:b?c*.txt":            "a_b_c_.txt",
		"trailing. ":            "trailing",
		"/abs/path/x.jpg":       "abs/path/x.jpg",
		"Fotoğraflar/şeker.jpg": "Fotoğraflar/şeker.jpg",
	}
	for in, want := range cases {
		if got := strings.Join(sanitizeRel(in), "/"); got != want {
			t.Errorf("sanitizeRel(%q) = %q, want %q", in, got, want)
		}
	}
	if len(sanitizeRel("../..")) != 0 {
		t.Error("dot paths must vanish")
	}
}

func TestUniqueName(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, "a (2).txt"), nil, 0o644)
	if got := uniqueName(dir, "a.txt"); got != "a (3).txt" {
		t.Errorf("got %q", got)
	}
	if got := uniqueName(dir, "b.txt"); got != "b.txt" {
		t.Errorf("got %q", got)
	}
}

func TestParseManualAddr(t *testing.T) {
	for in, want := range map[string]string{
		"192.168.1.20":                     "192.168.1.20:45455",
		"192.168.1.20:45464":               "192.168.1.20:45465",
		"http://192.168.1.20:45454/?k=abc": "192.168.1.20:45455",
	} {
		got, err := parseManualAddr(in)
		if err != nil || got != want {
			t.Errorf("parseManualAddr(%q) = %q, %v", in, got, err)
		}
	}
	if _, err := parseManualAddr("example.com"); err == nil {
		t.Error("host names must be rejected")
	}
}
