// Package adkit exposes the AllDrop engine to the iOS app.
// Build: gomobile bind -target=ios,iossimulator -o Adkit.xcframework ./mobile/adkit
package adkit

import (
	"strings"
	"sync"

	"alldrop/engine"
)

// Listener receives engine events: PORT, OFFER(name, count, code), SAVED(path), ACTIVE(1/0).
type Listener interface {
	OnEvent(kind string, a string, b string, c string)
}

var (
	mu      sync.Mutex
	app     *engine.App
	lastErr string
)

// Start runs the engine and returns its local web port, or -1 (see LastError).
// hintIPs is a comma separated list of the device's LAN addresses.
func Start(dataDir, downloadDir, name, platform, token, hintIPs string, l Listener) int {
	mu.Lock()
	defer mu.Unlock()
	if app != nil {
		return app.Port()
	}
	opts := engine.Options{
		DataDir:      dataDir,
		DownloadDir:  downloadDir,
		Name:         name,
		Platform:     platform,
		Mobile:       true,
		NoBrowser:    true,
		OwnerToken:   token,
		PeriodicScan: true,
	}
	if hintIPs != "" {
		opts.HintIPs = strings.Split(hintIPs, ",")
	}
	if l != nil {
		opts.OnEvent = func(parts ...string) {
			p := append(append([]string{}, parts...), "", "", "", "")
			l.OnEvent(p[0], p[1], p[2], p[3])
		}
	}
	a, err := engine.Start(opts)
	if err != nil {
		lastErr = err.Error()
		return -1
	}
	app = a
	return a.Port()
}

// SetDownloadDir changes the folder that received files are saved to.
func SetDownloadDir(dir string) bool {
	mu.Lock()
	a := app
	mu.Unlock()
	if a == nil {
		return false
	}
	return a.SetDownloadDir(dir) == nil
}

// Stop shuts the engine down; Start may be called again afterwards.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if app != nil {
		app.Close()
		app = nil
	}
}

// Port returns the running engine's web port or 0.
func Port() int {
	mu.Lock()
	defer mu.Unlock()
	if app == nil {
		return 0
	}
	return app.Port()
}

// Stage queues files (newline separated absolute paths) for sending.
// temp files are deleted once sent.
func Stage(paths string, temp bool) int {
	mu.Lock()
	a := app
	mu.Unlock()
	if a == nil {
		return 0
	}
	var list []string
	for _, p := range strings.Split(paths, "\n") {
		if p = strings.TrimSpace(p); p != "" {
			list = append(list, p)
		}
	}
	return a.Stage(list, temp)
}

// LastError describes why Start failed.
func LastError() string {
	mu.Lock()
	defer mu.Unlock()
	return lastErr
}

// Version of the engine.
func Version() string { return engine.Version }
