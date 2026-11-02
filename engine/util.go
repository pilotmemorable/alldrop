package engine

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func randID(nbytes int) string {
	b := make([]byte, nbytes)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// platformName returns the OS label used by the UI.
func platformName() string {
	switch runtime.GOOS {
	case "darwin":
		return "mac"
	case "windows":
		return "windows"
	}
	return "linux"
}

var reservedWin = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// sanitizeName makes one path component safe on Windows, macOS, Linux and Android.
func sanitizeName(s string) string {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "_")
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r < 0x20 || r == 0x7f:
			b.WriteRune('_')
		case strings.ContainsRune(`<>:"/\|?*`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimLeft(b.String(), " ")
	out = strings.TrimRight(out, " .")
	if out == "" || out == "." || out == ".." {
		return ""
	}
	base := out
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if reservedWin[strings.ToUpper(base)] {
		out = "_" + out
	}
	if len(out) > 200 {
		ext := filepath.Ext(out)
		if len(ext) > 20 {
			ext = ""
		}
		cut := 200 - len(ext)
		for cut > 0 && !utf8.RuneStart(out[cut]) {
			cut--
		}
		out = out[:cut] + ext
	}
	return out
}

// sanitizeRel splits a relative path sent by a client into safe components.
func sanitizeRel(p string) []string {
	p = strings.ReplaceAll(p, "\\", "/")
	var parts []string
	for _, c := range strings.Split(p, "/") {
		c = sanitizeName(c)
		if c != "" {
			parts = append(parts, c)
		}
	}
	if len(parts) > 32 {
		parts = parts[len(parts)-32:]
	}
	return parts
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

// uniqueName returns name, or "name (2).ext" ... so that dir/name does not exist yet.
func uniqueName(dir, name string) string {
	if !exists(filepath.Join(dir, name)) && !exists(filepath.Join(dir, name+partSuffix)) {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	if base == "" {
		base, ext = name, ""
	}
	for i := 2; i < 100000; i++ {
		n := base + " (" + strconv.Itoa(i) + ")" + ext
		if !exists(filepath.Join(dir, n)) && !exists(filepath.Join(dir, n+partSuffix)) {
			return n
		}
	}
	return randID(4) + "_" + name
}

const partSuffix = ".adpart"

type localAddr struct {
	IP    string
	Bcast string
	score int
}

func ifaceScore(name string, ip net.IP) int {
	n := strings.ToLower(name)
	s := 50
	for _, bad := range []string{"virtual", "vmware", "vbox", "hyper-v", "vethernet", "docker", "veth", "br-", "wsl", "tailscale", "zerotier", "utun", "tun", "tap", "llw", "awdl", "bridge", "loopback", "vpn"} {
		if strings.Contains(n, bad) {
			s -= 40
			break
		}
	}
	for _, good := range []string{"wi-fi", "wifi", "wlan", "wireless", "en0", "en1", "eth", "ethernet", "wl"} {
		if strings.Contains(n, good) {
			s += 10
			break
		}
	}
	ip4 := ip.To4()
	switch {
	case ip4[0] == 192 && ip4[1] == 168:
		s += 20
	case ip4[0] == 10:
		s += 15
	case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
		s += 5
	}
	return s
}

// localIPv4 lists usable LAN addresses, best first. hints are used when the
// interface list is unavailable (Android 11+ blocks netlink for apps).
func localIPv4(hints []string) []localAddr {
	var out []localAddr
	seen := map[string]bool{}
	ifs, err := net.Interfaces()
	if err == nil {
		for _, ifc := range ifs {
			if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, err := ifc.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				ipn, ok := a.(*net.IPNet)
				if !ok {
					continue
				}
				ip4 := ipn.IP.To4()
				if ip4 == nil || ip4.IsLinkLocalUnicast() || ip4.IsLoopback() {
					continue
				}
				mask := ipn.Mask
				if len(mask) == 16 {
					mask = mask[12:]
				}
				bc := make(net.IP, 4)
				for i := 0; i < 4; i++ {
					bc[i] = ip4[i] | ^mask[i]
				}
				ones, _ := ipn.Mask.Size()
				if ones >= 31 {
					bc = nil
				}
				if seen[ip4.String()] {
					continue
				}
				seen[ip4.String()] = true
				la := localAddr{IP: ip4.String(), score: ifaceScore(ifc.Name, ip4)}
				if bc != nil {
					la.Bcast = bc.String()
				}
				out = append(out, la)
			}
		}
	}
	for _, h := range hints {
		ip := net.ParseIP(strings.TrimSpace(h)).To4()
		if ip == nil || seen[ip.String()] || ip.IsLoopback() {
			continue
		}
		seen[ip.String()] = true
		bc := net.IPv4(ip[0], ip[1], ip[2], 255).To4()
		out = append(out, localAddr{IP: ip.String(), Bcast: bc.String(), score: ifaceScore("wlan", ip) + 5})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })
	return out
}

func hostIsLocal(host string) bool {
	h := host
	if hh, _, err := net.SplitHostPort(host); err == nil {
		h = hh
	}
	h = strings.Trim(h, "[]")
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

func remoteIP(remoteAddr string) net.IP {
	h, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		h = remoteAddr
	}
	return net.ParseIP(h)
}

func contentDisposition(kind, name string) string {
	ascii := make([]rune, 0, len(name))
	for _, r := range name {
		if r < 0x20 || r > 0x7e || r == '"' || r == '\\' {
			ascii = append(ascii, '_')
		} else {
			ascii = append(ascii, r)
		}
	}
	return kind + `; filename="` + string(ascii) + `"; filename*=UTF-8''` + url.PathEscape(name)
}

// OpenURL opens a URL or path with the system default handler.
func OpenURL(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		if mobileMode {
			return nil
		}
		cmd = exec.Command("xdg-open", target)
	}
	hideWindow(cmd)
	return cmd.Start()
}

func openFolder(dir string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		if mobileMode {
			return nil
		}
		cmd = exec.Command("xdg-open", dir)
	}
	hideWindow(cmd)
	return cmd.Start()
}

func revealFile(p string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", p)
	case "darwin":
		cmd = exec.Command("open", "-R", p)
	default:
		return openFolder(filepath.Dir(p))
	}
	hideWindow(cmd)
	return cmd.Start()
}

func dirWritable(dir string) bool {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	f, err := os.CreateTemp(dir, ".ad-test-*")
	if err != nil {
		return false
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	return true
}
