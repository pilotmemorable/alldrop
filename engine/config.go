package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config is persisted as JSON in the data directory.
type Config struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	DownloadDir string            `json:"downloadDir"`
	AutoAccept  bool              `json:"autoAccept"`
	NoCode      bool              `json:"noCode"`  // false = senders must type the receiver's 6-digit code
	Trusted     map[string]string `json:"trusted"` // sender id -> display name
	GuestKey    string            `json:"guestKey"`
	ManualPeers []string          `json:"manualPeers"`
}

func defaultDataDir() string {
	if d, err := os.UserConfigDir(); err == nil && d != "" {
		return filepath.Join(d, "AllDrop")
	}
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return filepath.Join(h, ".alldrop")
	}
	return filepath.Join(os.TempDir(), "alldrop")
}

func defaultDownloadDir() string {
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return filepath.Join(h, "Downloads", "AllDrop")
	}
	return filepath.Join(os.TempDir(), "AllDrop")
}

func defaultName() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		switch runtime.GOOS {
		case "windows":
			return "Windows PC"
		case "darwin":
			return "Mac"
		}
		return "Computer"
	}
	h = strings.TrimSuffix(h, ".local")
	h = strings.TrimSuffix(h, ".lan")
	h = strings.ReplaceAll(h, "-", " ")
	if len(h) > 40 {
		h = h[:40]
	}
	return h
}

func loadConfig(path string) Config {
	var c Config
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &c)
	}
	if c.ID == "" {
		c.ID = randID(12)
	}
	if c.Name == "" {
		c.Name = defaultName()
	}
	if c.DownloadDir == "" {
		c.DownloadDir = defaultDownloadDir()
	}
	if c.Trusted == nil {
		c.Trusted = map[string]string{}
	}
	if c.GuestKey == "" {
		c.GuestKey = randID(8)
	}
	return c
}

func saveConfig(path string, c Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
