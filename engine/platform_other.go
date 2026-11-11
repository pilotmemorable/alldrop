//go:build !windows

package engine

import (
	"errors"
	"os/exec"
	"runtime"
	"syscall"
)

func hideWindow(cmd *exec.Cmd) {}

// shellOpen only exists on Windows.
func shellOpen(file, args string) error { return nil }

func reuseControl(network, address string, c syscall.RawConn) error {
	var serr error
	err := c.Control(func(fd uintptr) {
		serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		switch runtime.GOOS {
		case "darwin", "freebsd", "openbsd", "netbsd", "dragonfly":
			_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, 0x200, 1) // SO_REUSEPORT
		}
	})
	if err != nil {
		return err
	}
	return serr
}

var errWindowsOnly = errors.New("only available on Windows")

func addFirewallRule() error { return errWindowsOnly }
func installSendTo() error   { return errWindowsOnly }
