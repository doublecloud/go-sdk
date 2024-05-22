package browser

import (
	"os/exec"
)

func OpenURLSupported() bool {
	return openURLSupported()
}

// OpenURL opens a new browser window pointing to url.
func OpenURL(url string) error {
	return openBrowser(url)
}

func runCmd(prog string, args ...string) error {
	cmd := exec.Command(prog, args...)
	setFlags(cmd)
	return cmd.Run()
}
