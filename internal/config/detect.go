package config

import (
	"os/exec"
	"regexp"
	"strings"
)

var (
	// SSH: git@ssh.dev.azure.com:v3/OrgName/ProjectName/RepoName
	sshPattern = regexp.MustCompile(`ssh\.dev\.azure\.com.*v3/([^/]+)/([^/]+)/`)
	// HTTPS: https://dev.azure.com/OrgName/ProjectName/_git/RepoName
	httpsPattern = regexp.MustCompile(`dev\.azure\.com/([^/]+)/([^/]+)/`)
	// Legacy HTTPS: https://OrgName.visualstudio.com/ProjectName/_git/RepoName
	legacyPattern = regexp.MustCompile(`([^/]+)\.visualstudio\.com/([^/]+)/`)
)

// DetectFromRemote parses the git remote origin URL to extract org and project.
func DetectFromRemote(cfg *Config) {
	url := getRemoteURL()
	if url == "" {
		return
	}

	org, project := parseRemoteURL(url)
	if org == "" {
		return
	}

	if cfg.Org == "" {
		cfg.Org = "https://dev.azure.com/" + org
	}
	if cfg.Project == "" {
		cfg.Project = project
	}
}

// DetectRepoName extracts the repository name from the git remote origin URL.
func DetectRepoName() string {
	url := getRemoteURL()
	if url == "" {
		return ""
	}
	// Get the last path segment, strip .git suffix
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	name := parts[len(parts)-1]
	return strings.TrimSuffix(name, ".git")
}

func getRemoteURL() string {
	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func parseRemoteURL(url string) (org, project string) {
	if m := sshPattern.FindStringSubmatch(url); len(m) == 3 {
		return m[1], m[2]
	}
	if m := httpsPattern.FindStringSubmatch(url); len(m) == 3 {
		return m[1], m[2]
	}
	if m := legacyPattern.FindStringSubmatch(url); len(m) == 3 {
		return m[1], m[2]
	}
	return "", ""
}
