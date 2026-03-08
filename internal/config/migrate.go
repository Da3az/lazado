package config

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// loadLegacy reads a bash-style key=value config file (the old lazado format).
func loadLegacy(path string, cfg *Config) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and blank lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Strip surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		// Expand $HOME and ~
		home, _ := os.UserHomeDir()
		value = strings.ReplaceAll(value, "$HOME", home)
		value = strings.ReplaceAll(value, "~", home)

		applyLegacyKey(key, value, cfg)
	}
}

func applyLegacyKey(key, value string, cfg *Config) {
	switch key {
	case "org":
		if cfg.Org == "" {
			cfg.Org = value
		}
	case "project":
		if cfg.Project == "" {
			cfg.Project = value
		}
	case "clone_dir":
		cfg.CloneDir = value
	case "branch_prefix":
		cfg.BranchPrefix = value
	case "branch_id_pattern":
		cfg.BranchIDPattern = value
	case "default_target_branch":
		cfg.DefaultTargetBranch = value
	case "cache_ttl":
		if n, err := strconv.Atoi(value); err == nil {
			cfg.CacheTTL = n
		}
	case "cache_dir":
		cfg.CacheDir = value
	case "browser_cmd":
		cfg.BrowserCmd = value
	case "ssh_url_template":
		cfg.SSHURLTemplate = value
	}
}

// lookPath is used by config.go's detectBrowser via the execLookPath variable.
func lookPath(cmd string) (string, error) {
	return exec.LookPath(cmd)
}
