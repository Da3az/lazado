package app

import (
	"context"
	"fmt"
	"os"

	"github.com/da3az/lazado/internal/api"
	"github.com/da3az/lazado/internal/auth"
	"github.com/da3az/lazado/internal/cache"
	"github.com/da3az/lazado/internal/config"
	gitpkg "github.com/da3az/lazado/internal/git"
	"github.com/da3az/lazado/internal/ui"
	"github.com/da3az/lazado/internal/ui/panels"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.Org == "" || cfg.Project == "" {
		fmt.Fprintln(os.Stderr, "Organization and project not configured.")
		fmt.Fprintln(os.Stderr, "Run 'lazado init' to set up, or set LAZADO_ORG and LAZADO_PROJECT.")
		os.Exit(1)
	}

	authProvider := auth.NewChain()
	if _, err := authProvider.Token(); err != nil {
		fmt.Fprintln(os.Stderr, "Authentication not configured.")
		fmt.Fprintln(os.Stderr, "Run 'lazado init' to set up, or set LAZADO_PAT environment variable.")
		os.Exit(1)
	}

	client := api.NewClient(cfg.Org, cfg.Project, authProvider)

	// Try to get current user ID
	conn, err := client.ValidateConnection(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to Azure DevOps: %v\n", err)
		fmt.Fprintln(os.Stderr, "Check your PAT and org URL.")
		os.Exit(1)
	}
	client.SetUser(conn.AuthenticatedUser.ID, conn.AuthenticatedUser.DisplayName)

	g := gitpkg.New()
	branch, _ := g.CurrentBranch()

	// Detect repo ID for PR panel
	diskCache := cache.New(cfg.CacheDir, cfg.CacheTTL)
	repoID := resolveRepoID(client, cfg, diskCache)

	// Build panels
	wiPanel := panels.NewWorkItemsPanel(client)
	prPanel := panels.NewPullRequestsPanel(client, g)
	prPanel.SetRepoID(repoID)
	pipePanel := panels.NewPipelinesPanel(client)
	repoPanel := panels.NewReposPanel(client, g, cfg.CloneDir)

	panelMap := map[ui.PanelID]ui.Panel{
		ui.PanelWorkItems:    wiPanel,
		ui.PanelPullRequests: prPanel,
		ui.PanelPipelines:    pipePanel,
		ui.PanelRepos:        repoPanel,
	}

	model := ui.NewAppModel(panelMap, cfg.Org, cfg.Project, branch)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

// resolveRepoID gets the current repo's Azure DevOps ID.
func resolveRepoID(client *api.Client, cfg *config.Config, c *cache.DiskCache) string {
	repoName := config.DetectRepoName()
	if repoName == "" {
		return ""
	}

	// Check cache
	var cachedID string
	cacheKey := "repo-id-" + repoName
	if c.Get(cacheKey, &cachedID, false) {
		return cachedID
	}

	// Fetch from API
	repo, err := client.GetRepository(context.Background(), repoName)
	if err != nil {
		return ""
	}

	// Cache it
	c.Set(cacheKey, repo.ID)
	return repo.ID
}

// RunInit runs the setup wizard.
func RunInit() error {
	cfg, _ := config.Load()

	fmt.Println("lazado — Setup Wizard")
	fmt.Println("=====================")
	fmt.Println()

	// Detect from git remote
	if cfg.Org == "" || cfg.Project == "" {
		fmt.Println("Attempting to detect organization and project from git remote...")
		config.DetectFromRemote(cfg)
	}

	if cfg.Org != "" {
		fmt.Printf("Organization: %s\n", cfg.Org)
	} else {
		fmt.Print("Organization URL (e.g., https://dev.azure.com/MyOrg): ")
		fmt.Scanln(&cfg.Org)
	}

	if cfg.Project != "" {
		fmt.Printf("Project: %s\n", cfg.Project)
	} else {
		fmt.Print("Project name: ")
		fmt.Scanln(&cfg.Project)
	}

	// PAT
	fmt.Println()
	fmt.Println("Create a Personal Access Token at:")
	fmt.Printf("  %s/_usersettings/tokens\n", cfg.Org)
	fmt.Println("Required scopes: Work Items (Read/Write), Code (Read/Write), Build (Read/Execute)")
	fmt.Println()
	fmt.Print("Personal Access Token: ")
	var pat string
	fmt.Scanln(&pat)

	if pat == "" {
		return fmt.Errorf("PAT is required")
	}

	// Validate connection
	fmt.Println()
	fmt.Println("Validating connection...")
	provider := &auth.EnvProvider{}
	// Temporarily set env for validation
	os.Setenv("LAZADO_PAT", pat)
	client := api.NewClient(cfg.Org, cfg.Project, provider)
	conn, err := client.ValidateConnection(context.Background())
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	fmt.Printf("Authenticated as: %s\n", conn.AuthenticatedUser.DisplayName)

	// Store PAT in keyring
	fmt.Println("Storing PAT in system keyring...")
	if err := auth.Store(pat); err != nil {
		fmt.Printf("Warning: Could not store in keyring (%v). Set LAZADO_PAT env var instead.\n", err)
	} else {
		fmt.Println("PAT stored securely.")
		os.Unsetenv("LAZADO_PAT")
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Println()
	fmt.Println("Configuration saved. Run 'lazado' to start the TUI.")
	return nil
}
