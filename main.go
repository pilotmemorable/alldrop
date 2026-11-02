// AllDrop: All Device - AirDrop-style file sharing between Windows, macOS,
// Linux, Android, iPhone and iPad over the local network. No cloud.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"alldrop/engine"
)

func main() {
	var (
		port      = flag.Int("port", engine.DefaultPort, "web port (the encrypted app-to-app port is port+1)")
		dataDir   = flag.String("data", "", "settings directory")
		downloads = flag.String("downloads", "", "folder for received files")
		name      = flag.String("name", "", "device name shown to others")
		ips       = flag.String("ips", "", "comma separated LAN IPv4 hints (Android)")
		noBrowser = flag.Bool("no-browser", false, "do not open the browser")
		android   = flag.Bool("android", false, "run as the Android app engine")
		showVer   = flag.Bool("version", false, "print version")
		platform  = flag.String("platform", "", "override the device type shown to others (testing)")
		mobileUI  = flag.Bool("mobile-ui", false, "show the phone interface (testing)")
	)
	flag.Parse()
	if *showVer {
		fmt.Println("AllDrop", engine.Version)
		return
	}
	dir := *dataDir
	if dir == "" {
		dir = engine.DefaultDataDir()
	}
	staged := absPaths(flag.Args())

	// Already running? Hand over the files and show the window.
	if !*android {
		if id, err := engine.IdentityID(dir); err == nil {
			if p := engine.FindRunning(id, *port); p > 0 {
				if len(staged) > 0 {
					body, _ := json.Marshal(map[string]interface{}{"paths": staged})
					req, _ := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/api/stage", p), bytes.NewReader(body))
					req.Header.Set("X-AD", "1")
					req.Header.Set("Content-Type", "application/json")
					if resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req); err == nil {
						resp.Body.Close()
					}
				}
				if !*noBrowser {
					engine.OpenURL(fmt.Sprintf("http://localhost:%d/", p))
				}
				return
			}
		}
	}

	opts := engine.Options{
		Port:        *port,
		DataDir:     dir,
		DownloadDir: *downloads,
		Name:        *name,
		NoBrowser:   *noBrowser,
		Platform:    *platform,
		Mobile:      *mobileUI,
		OnQuit:      func() { os.Exit(0) },
	}
	if *ips != "" {
		opts.HintIPs = strings.Split(*ips, ",")
	}
	if *android {
		// Events go to the Android shell as tab separated lines on stdout.
		var mu sync.Mutex
		opts.Mobile = true
		opts.Platform = "android"
		opts.OwnerToken = os.Getenv("AD_OWNER_TOKEN")
		opts.OnEvent = func(parts ...string) {
			mu.Lock()
			defer mu.Unlock()
			clean := make([]string, len(parts))
			for i, p := range parts {
				clean[i] = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ").Replace(p)
			}
			fmt.Fprintln(os.Stdout, "ALLDROP\t"+strings.Join(clean, "\t"))
		}
	}
	app, err := engine.Start(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "AllDrop:", err)
		os.Exit(1)
	}
	if len(staged) > 0 {
		app.Stage(staged, false)
	}
	if *android {
		go func() {
			// The Android shell keeps stdin open; when it dies, so do we.
			io.Copy(io.Discard, os.Stdin)
			app.Close()
			os.Exit(0)
		}()
	} else if !*noBrowser {
		time.Sleep(150 * time.Millisecond)
		engine.OpenURL(fmt.Sprintf("http://localhost:%d/", app.Port()))
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	app.Close()
}

func absPaths(in []string) []string {
	out := make([]string, 0, len(in))
	for _, p := range in {
		if ap, err := filepath.Abs(p); err == nil {
			out = append(out, ap)
		}
	}
	return out
}
