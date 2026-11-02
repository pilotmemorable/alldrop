//go:build windows

package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const createNoWindow = 0x08000000

func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}

func reuseControl(network, address string, c syscall.RawConn) error {
	var serr error
	err := c.Control(func(fd uintptr) {
		serr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if err != nil {
		return err
	}
	return serr
}

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

// runElevated starts file with the "runas" verb (UAC prompt).
func runElevated(file, args string) error {
	verb, _ := syscall.UTF16PtrFromString("runas")
	f, _ := syscall.UTF16PtrFromString(file)
	a, _ := syscall.UTF16PtrFromString(args)
	r, _, _ := procShellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(a)), 0, 0) // SW_HIDE
	if r <= 32 {
		return fmt.Errorf("could not start elevated command (code %d)", r)
	}
	return nil
}

// addFirewallRule lets other devices on the network reach this app.
func addFirewallRule() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	script := filepath.Join(os.TempDir(), "alldrop-firewall.cmd")
	body := "@echo off\r\n" +
		"netsh advfirewall firewall delete rule name=\"AllDrop\" >nul 2>&1\r\n" +
		"netsh advfirewall firewall add rule name=\"AllDrop\" dir=in action=allow program=\"" + exe + "\" enable=yes profile=any >nul\r\n"
	if err := os.WriteFile(script, []byte(body), 0o644); err != nil {
		return err
	}
	return runElevated("cmd.exe", "/c \""+script+"\"")
}

// installSendTo adds AllDrop to Explorer's right-click "Send to" menu.
func installSendTo() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	ps := "$s=(New-Object -ComObject WScript.Shell).CreateShortcut([Environment]::GetFolderPath('SendTo')+'\\AllDrop.lnk');" +
		"$s.TargetPath=" + q(exe) + ";$s.IconLocation=" + q(exe+",0") + ";$s.Description='AllDrop';$s.Save()"
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", ps)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
