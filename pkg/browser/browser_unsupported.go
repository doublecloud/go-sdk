//go:build !linux && !windows && !darwin
// +build !linux,!windows,!darwin

package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func openURLSupported() bool {
	return false
}

func openBrowser(url string) error {
	return fmt.Errorf("openBrowser: unsupported operating system: %v", runtime.GOOS)
}

func setFlags(cmd *exec.Cmd) {}
