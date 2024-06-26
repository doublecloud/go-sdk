//go:build windows
// +build windows

package browser

import (
	"os/exec"
	"strings"
	"syscall"
)

func openURLSupported() bool {
	return true
}

func openBrowser(url string) error {
	r := strings.NewReplacer("&", "^&")
	return runCmd("cmd", "/c", "start", r.Replace(url))
}

func setFlags(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
