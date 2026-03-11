package panels

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/da3az/lazado/internal/api"
	gitpkg "github.com/da3az/lazado/internal/git"
	"github.com/da3az/lazado/internal/ui"
	"github.com/da3az/lazado/internal/ui/components"
	"github.com/da3az/lazado/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type repoKeys struct {
	Clone   key.Binding
	Browser key.Binding
}

var repoKeyMap = repoKeys{
	Clone:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clone")),
	Browser: key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "open in browser")),
}

// ReposPanel manages the repositories view.
type ReposPanel struct {
	client   *api.Client
	git      *gitpkg.Git
	cloneDir string
	list     *components.List
	detail   *components.DetailView
	repos    []api.Repository
	focused  bool
	width    int
	height   int
}

// NewReposPanel creates the repos panel.
func NewReposPanel(client *api.Client, g *gitpkg.Git, cloneDir string) *ReposPanel {
	list := components.NewList("Repositories")
	list.SetColumns([]components.ColumnDef{
		{Field: "Subtitle", MinWidth: 20},
		{Field: "Title", Flex: true},
	})
	return &ReposPanel{
		client:   client,
		git:      g,
		cloneDir: cloneDir,
		list:     list,
		detail:   components.NewDetailView(),
	}
}

// Init loads initial data.
func (p *ReposPanel) Init() tea.Cmd {
	return p.loadRepos()
}

// SetFocused sets focus state.
func (p *ReposPanel) SetFocused(focused bool) {
	p.focused = focused
	p.list.SetFocused(focused)
}

// SetSize updates dimensions.
func (p *ReposPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	layout := ui.CalculateLayout(w, h)
	p.list.SetSize(layout.ListWidth-layout.HFrame, layout.ContentHeight)
	p.detail.SetSize(layout.DetailWidth-layout.HFrame, layout.ContentHeight)
}

// Update handles messages.
func (p *ReposPanel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ui.ReposLoadedMsg:
		p.list.SetLoading(false)
		if msg.Err != nil {
			return statusCmd(msg.Err.Error(), true)
		}
		p.repos = msg.Repos
		p.list.SetItems(p.toListItems(msg.Repos))
		p.updateDetail()
		return nil

	case tea.KeyMsg:
		return p.handleKey(msg)
	}

	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

func (p *ReposPanel) handleKey(msg tea.KeyMsg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch {
	case key.Matches(msg, repoKeyMap.Clone):
		return p.cloneRepo()
	case key.Matches(msg, repoKeyMap.Browser):
		return p.openInBrowser()
	case key.Matches(msg, ui.Keys.Refresh):
		return p.loadRepos()
	}

	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

// View renders the panel.
func (p *ReposPanel) View() string {
	return p.list.View()
}

// DetailView returns the detail pane.
func (p *ReposPanel) DetailView() string {
	return p.detail.View()
}

// OverlayView returns the overlay content or empty string.
func (p *ReposPanel) OverlayView() string {
	return ""
}

// HasActiveOverlay returns true when a modal overlay is open.
func (p *ReposPanel) HasActiveOverlay() bool {
	return false
}

// HelpKeys returns context keybindings.
func (p *ReposPanel) HelpKeys() []key.Binding {
	return []key.Binding{repoKeyMap.Clone, repoKeyMap.Browser}
}

func (p *ReposPanel) loadRepos() tea.Cmd {
	p.list.SetLoading(true)
	client := p.client
	return func() tea.Msg {
		repos, err := client.ListRepositories(context.Background())
		return ui.ReposLoadedMsg{Repos: repos, Err: err}
	}
}

func (p *ReposPanel) toListItems(repos []api.Repository) []components.ListItem {
	result := make([]components.ListItem, len(repos))
	for i, repo := range repos {
		branch := repo.DefaultBranch
		if len(branch) > len("refs/heads/") {
			branch = branch[len("refs/heads/"):]
		}
		sizeMB := float64(repo.Size) / (1024 * 1024)
		result[i] = components.ListItem{
			ID:       "",
			Title:    repo.Name,
			Subtitle: fmt.Sprintf("%s (%.1f MB)", branch, sizeMB),
		}
	}
	return result
}

func (p *ReposPanel) updateDetail() {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.repos) {
		p.detail.Clear()
		return
	}
	repo := p.repos[idx]

	branch := repo.DefaultBranch
	if len(branch) > len("refs/heads/") {
		branch = branch[len("refs/heads/"):]
	}

	p.detail.SetContent(
		repo.Name,
		[]components.DetailField{
			{Label: "Default Branch", Value: branch},
			{Label: "Remote URL", Value: repo.RemoteURL},
			{Label: "SSH URL", Value: repo.SSHURL},
			{Label: "Size", Value: fmt.Sprintf("%.1f MB", float64(repo.Size)/(1024*1024))},
		},
		"",
	)
}

func (p *ReposPanel) cloneRepo() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.repos) {
		return nil
	}
	repo := p.repos[idx]
	g := p.git
	dir := filepath.Join(p.cloneDir, repo.Name)
	cloneURL := repo.SSHURL
	if cloneURL == "" {
		cloneURL = repo.RemoteURL
	}
	return func() tea.Msg {
		err := g.Clone(cloneURL, dir)
		if err != nil {
			return ui.StatusMsg{Text: "Clone failed: " + err.Error(), IsError: true}
		}
		return ui.StatusMsg{Text: "Cloned " + repo.Name + " to " + utils.Truncate(dir, 40), IsError: false}
	}
}

func (p *ReposPanel) openInBrowser() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.repos) {
		return nil
	}
	repo := p.repos[idx]
	if repo.WebURL != "" {
		utils.OpenBrowser(repo.WebURL, "")
	}
	return nil
}
