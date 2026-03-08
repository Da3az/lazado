package utils

import "os/exec"

// OpenBrowser opens the given URL in the user's default browser.
func OpenBrowser(url, browserCmd string) error {
	if browserCmd == "" {
		// Try common defaults
		for _, cmd := range []string{"xdg-open", "open", "start"} {
			if _, err := exec.LookPath(cmd); err == nil {
				browserCmd = cmd
				break
			}
		}
	}
	if browserCmd == "" {
		return nil // silently skip if no browser found
	}
	return exec.Command(browserCmd, url).Start()
}
