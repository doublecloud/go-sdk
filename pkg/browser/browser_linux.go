//go:build linux
// +build linux

package browser

import (
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"
)

const openURLCmdLinux = "xdg-open"
const openURLCmdWSL = "cmd.exe"

func isWSL() bool {
	read, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	return strings.Contains(string(read), "Microsoft")
}

func openURLSupported() bool {
	var urlCmd string
	if isWSL() {
		urlCmd = openURLCmdWSL
	} else {
		urlCmd = openURLCmdLinux
	}

	_, err := exec.LookPath(urlCmd)
	return err == nil
}

func openBrowser(url string) error {
	if isWSL() {
		r := strings.NewReplacer("&", "^&")
		return runCmd(openURLCmdWSL, "/c", "start", r.Replace(url))
	}

	return runCmd(openURLCmdLinux, url)
}

func setFlags(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
