package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Org                 string `yaml:"org"`
	Project             string `yaml:"project"`
	CloneDir            string `yaml:"clone_dir"`
	BranchPrefix        string `yaml:"branch_prefix"`
	BranchIDPattern     string `yaml:"branch_id_pattern"`
	DefaultTargetBranch string `yaml:"default_target_branch"`
	CacheTTL            int    `yaml:"cache_ttl"`
	CacheDir            string `yaml:"cache_dir"`
	BrowserCmd          string `yaml:"browser_cmd"`
	SSHURLTemplate      string `yaml:"ssh_url_template"`
	Theme               string `yaml:"theme"`
	RefreshInterval     int    `yaml:"refresh_interval"`
}

func defaults() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		CloneDir:            filepath.Join(home, "dev"),
		BranchPrefix:        "feature",
		BranchIDPattern:     `(?:/)\d+`,
		DefaultTargetBranch: "main",
		CacheTTL:            86400,
		CacheDir:            filepath.Join(home, ".cache", "lazado"),
		BrowserCmd:          detectBrowser(),
		SSHURLTemplate:      "git@ssh.dev.azure.com:v3/{org}/{project}/{repo}",
		Theme:               "dark",
		RefreshInterval:     30,
	}
}

func detectBrowser() string {
	for _, cmd := range []string{"xdg-open", "open", "start"} {
		if _, err := execLookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

// execLookPath wraps os/exec.LookPath for testability.
var execLookPath = defaultLookPath

func defaultLookPath(cmd string) (string, error) {
	return lookPath(cmd)
}

// Load builds a Config by applying the hierarchy:
// 1. Built-in defaults
// 2. Global config file (~/.config/lazado/config.yml or legacy config)
// 3. Project config file (.lazado.yml or legacy .adoconfig)
// 4. Environment variables (LAZADO_*)
// 5. Auto-detect from git remote URL
func Load() (*Config, error) {
	cfg := defaults()

	// 2. Global config
	globalDir := os.Getenv("XDG_CONFIG_HOME")
	if globalDir == "" {
		home, _ := os.UserHomeDir()
		globalDir = filepath.Join(home, ".config")
	}
	yamlPath := filepath.Join(globalDir, "lazado", "config.yml")
	legacyPath := filepath.Join(globalDir, "lazado", "config")

	if err := loadYAML(yamlPath, cfg); err != nil {
		// Try legacy key=value format
		loadLegacy(legacyPath, cfg)
	}

	// 3. Project-level config
	gitRoot := findGitRoot()
	if gitRoot != "" {
		projectYAML := filepath.Join(gitRoot, ".lazado.yml")
		projectLegacy := filepath.Join(gitRoot, ".adoconfig")
		if err := loadYAML(projectYAML, cfg); err != nil {
			loadLegacy(projectLegacy, cfg)
		}
	}

	// 4. Environment variables override everything
	applyEnv(cfg)

	// 5. Auto-detect from git remote
	if cfg.Org == "" || cfg.Project == "" {
		DetectFromRemote(cfg)
	}

	return cfg, nil
}

func loadYAML(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

// Save writes the config as YAML to the global config file.
func Save(cfg *Config) error {
	globalDir := os.Getenv("XDG_CONFIG_HOME")
	if globalDir == "" {
		home, _ := os.UserHomeDir()
		globalDir = filepath.Join(home, ".config")
	}
	dir := filepath.Join(globalDir, "lazado")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.yml"), data, 0600)
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("LAZADO_ORG"); v != "" {
		cfg.Org = v
	}
	if v := os.Getenv("LAZADO_PROJECT"); v != "" {
		cfg.Project = v
	}
	if v := os.Getenv("LAZADO_CLONE_DIR"); v != "" {
		cfg.CloneDir = v
	}
	if v := os.Getenv("LAZADO_BRANCH_PREFIX"); v != "" {
		cfg.BranchPrefix = v
	}
	if v := os.Getenv("LAZADO_BRANCH_ID_PATTERN"); v != "" {
		cfg.BranchIDPattern = v
	}
	if v := os.Getenv("LAZADO_DEFAULT_TARGET_BRANCH"); v != "" {
		cfg.DefaultTargetBranch = v
	}
	if v := os.Getenv("LAZADO_CACHE_TTL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.CacheTTL = n
		}
	}
	if v := os.Getenv("LAZADO_CACHE_DIR"); v != "" {
		cfg.CacheDir = v
	}
	if v := os.Getenv("LAZADO_BROWSER_CMD"); v != "" {
		cfg.BrowserCmd = v
	}
	if v := os.Getenv("LAZADO_SSH_URL_TEMPLATE"); v != "" {
		cfg.SSHURLTemplate = v
	}
	if v := os.Getenv("LAZADO_THEME"); v != "" {
		cfg.Theme = v
	}
}

func findGitRoot() string {
	// Walk up from cwd looking for .git
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
