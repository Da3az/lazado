package git

import (
	"os/exec"
	"strings"
)

// Git provides helpers for interacting with the local git repository.
type Git struct{}

// New creates a Git helper.
func New() *Git {
	return &Git{}
}

// CurrentBranch returns the name of the currently checked-out branch.
func (g *Git) CurrentBranch() (string, error) {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RemoteURL returns the remote origin URL.
func (g *Git) RemoteURL() (string, error) {
	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Checkout switches to the specified branch, creating it if necessary.
func (g *Git) Checkout(branch string, create bool) error {
	args := []string{"checkout"}
	if create {
		args = append(args, "-b")
	}
	args = append(args, branch)
	return exec.Command("git", args...).Run()
}

// Fetch fetches from the specified remote (or "origin" by default).
func (g *Git) Fetch(remote string) error {
	if remote == "" {
		remote = "origin"
	}
	return exec.Command("git", "fetch", remote).Run()
}

// FetchAndCheckout fetches from origin and checks out the specified branch.
func (g *Git) FetchAndCheckout(branch string) error {
	if err := g.Fetch("origin"); err != nil {
		return err
	}
	return g.Checkout(branch, false)
}

// GitRoot returns the root directory of the current git repository.
func (g *Git) GitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Clone clones a repository into the specified directory.
func (g *Git) Clone(url, dir string) error {
	return exec.Command("git", "clone", url, dir).Run()
}
