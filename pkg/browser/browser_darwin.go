//go:build darwin
// +build darwin

package browser

import "os/exec"

func openURLSupported() bool {
	return true
}

func openBrowser(url string) error {
	return runCmd("open", url)
}

func setFlags(cmd *exec.Cmd) {}
