package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/da3az/lazado/internal/api"
	gitpkg "github.com/da3az/lazado/internal/git"
	"github.com/da3az/lazado/internal/ui"
	"github.com/da3az/lazado/internal/ui/components"
	"github.com/da3az/lazado/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type prKeys struct {
	Create   key.Binding
	Approve  key.Binding
	Complete key.Binding
	Checkout key.Binding
	Browser  key.Binding
	Toggle   key.Binding
}

var prKeyMap = prKeys{
	Create:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create PR")),
	Approve:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "approve")),
	Complete: key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "merge")),
	Checkout: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "checkout")),
	Browser:  key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "open in browser")),
	Toggle:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "my PRs / repo PRs")),
}

// PullRequestsPanel manages the pull requests view.
type PullRequestsPanel struct {
	client  *api.Client
	git     *gitpkg.Git
	repoID  string
	list    *components.List
	detail  *components.DetailView
	prs     []api.PullRequest
	focused bool
	width   int
	height  int
	myPRs   bool // true = show my PRs across repos, false = current repo PRs

	form    *components.Form
	confirm *components.Confirm
}

// NewPullRequestsPanel creates the PR panel.
func NewPullRequestsPanel(client *api.Client, g *gitpkg.Git) *PullRequestsPanel {
	return &PullRequestsPanel{
		client: client,
		git:    g,
		list:   components.NewList("Pull Requests"),
		detail: components.NewDetailView(),
	}
}

// SetRepoID sets the current repository ID.
func (p *PullRequestsPanel) SetRepoID(id string) {
	p.repoID = id
}

// Init loads initial data.
func (p *PullRequestsPanel) Init() tea.Cmd {
	return p.loadPRs()
}

// SetFocused sets focus state.
func (p *PullRequestsPanel) SetFocused(focused bool) {
	p.focused = focused
	p.list.SetFocused(focused)
}

// SetSize updates dimensions.
func (p *PullRequestsPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	layout := ui.CalculateLayout(w, h)
	p.list.SetSize(layout.ListWidth-2, layout.ContentHeight)
	p.detail.SetSize(layout.DetailWidth-2, layout.ContentHeight)
}

// Update handles messages.
func (p *PullRequestsPanel) Update(msg tea.Msg) tea.Cmd {
	if p.form != nil {
		return p.form.Update(msg)
	}
	if p.confirm != nil {
		return p.confirm.Update(msg)
	}

	switch msg := msg.(type) {
	case ui.PullRequestsLoadedMsg:
		p.list.SetLoading(false)
		if msg.Err != nil {
			return statusCmd(msg.Err.Error(), true)
		}
		p.prs = msg.PRs
		p.list.SetItems(p.toListItems(msg.PRs))
		p.updateDetail()
		return nil

	case ui.PRCreatedMsg:
		if msg.Err != nil {
			return statusCmd("PR creation failed: "+msg.Err.Error(), true)
		}
		return tea.Batch(
			statusCmd(fmt.Sprintf("Created PR #%d", msg.PR.PullRequestID), false),
			p.loadPRs(),
		)

	case ui.PRApprovedMsg:
		if msg.Err != nil {
			return statusCmd("Approve failed: "+msg.Err.Error(), true)
		}
		return tea.Batch(statusCmd("PR approved", false), p.loadPRs())

	case ui.PRCompletedMsg:
		if msg.Err != nil {
			return statusCmd("Merge failed: "+msg.Err.Error(), true)
		}
		return tea.Batch(statusCmd("PR merged", false), p.loadPRs())

	case components.FormSubmitMsg:
		return p.handleFormSubmit(msg)

	case components.FormCancelMsg:
		p.form = nil
		return nil

	case components.ConfirmMsg:
		p.confirm = nil
		if msg.ID == "merge-pr" && msg.Confirmed {
			return p.completePR()
		}
		return nil

	case tea.KeyMsg:
		return p.handleKey(msg)
	}

	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

func (p *PullRequestsPanel) handleKey(msg tea.KeyMsg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch {
	case key.Matches(msg, prKeyMap.Create):
		return p.showCreateForm()
	case key.Matches(msg, prKeyMap.Approve):
		return p.approvePR()
	case key.Matches(msg, prKeyMap.Complete):
		return p.showMergeConfirm()
	case key.Matches(msg, prKeyMap.Checkout):
		return p.checkoutBranch()
	case key.Matches(msg, prKeyMap.Browser):
		return p.openInBrowser()
	case key.Matches(msg, prKeyMap.Toggle):
		p.myPRs = !p.myPRs
		return p.loadPRs()
	case key.Matches(msg, ui.Keys.Refresh):
		return p.loadPRs()
	}

	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

// View renders the panel.
func (p *PullRequestsPanel) View() string {
	if p.form != nil {
		return p.form.View()
	}
	if p.confirm != nil {
		return p.confirm.View()
	}
	return p.list.View()
}

// DetailView returns the detail pane.
func (p *PullRequestsPanel) DetailView() string {
	return p.detail.View()
}

// HelpKeys returns context-sensitive keybindings.
func (p *PullRequestsPanel) HelpKeys() []key.Binding {
	return []key.Binding{
		prKeyMap.Create, prKeyMap.Approve, prKeyMap.Complete,
		prKeyMap.Checkout, prKeyMap.Toggle, prKeyMap.Browser,
	}
}

// --- Data loading ---

func (p *PullRequestsPanel) loadPRs() tea.Cmd {
	p.list.SetLoading(true)
	client := p.client
	repoID := p.repoID
	myPRs := p.myPRs
	return func() tea.Msg {
		opts := api.PRListOptions{Status: "active"}
		if myPRs {
			opts.CreatorID = client.UserID()
		} else {
			opts.RepoID = repoID
		}
		prs, err := client.ListPullRequests(context.Background(), opts)
		return ui.PullRequestsLoadedMsg{PRs: prs, Err: err}
	}
}

func (p *PullRequestsPanel) toListItems(prs []api.PullRequest) []components.ListItem {
	result := make([]components.ListItem, len(prs))
	for i, pr := range prs {
		subtitle := pr.SourceBranch() + " → " + pr.TargetBranch()
		result[i] = components.ListItem{
			ID:       fmt.Sprintf("#%d", pr.PullRequestID),
			Title:    pr.Title,
			Subtitle: utils.Truncate(subtitle, 30),
		}
	}
	return result
}

func (p *PullRequestsPanel) updateDetail() {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.prs) {
		p.detail.Clear()
		return
	}
	pr := p.prs[idx]

	var reviewers []string
	for _, r := range pr.Reviewers {
		reviewers = append(reviewers, fmt.Sprintf("%s (%s)", r.DisplayName, r.VoteLabel()))
	}
	reviewerStr := "None"
	if len(reviewers) > 0 {
		reviewerStr = strings.Join(reviewers, ", ")
	}

	draft := ""
	if pr.IsDraft {
		draft = " [DRAFT]"
	}

	p.detail.SetContent(
		fmt.Sprintf("PR #%d%s", pr.PullRequestID, draft),
		[]components.DetailField{
			{Label: "Title", Value: pr.Title},
			{Label: "Status", Value: pr.Status},
			{Label: "Source", Value: pr.SourceBranch()},
			{Label: "Target", Value: pr.TargetBranch()},
			{Label: "Created By", Value: pr.CreatedBy.DisplayName},
			{Label: "Merge Status", Value: pr.MergeStatus},
			{Label: "Reviewers", Value: reviewerStr},
		},
		pr.Description,
	)
}

// --- Actions ---

func (p *PullRequestsPanel) showCreateForm() tea.Cmd {
	branch, _ := p.git.CurrentBranch()
	p.form = components.NewForm("create-pr", "Create Pull Request", []components.FormField{
		{Label: "Title", Placeholder: "PR title...", Value: branch},
		{Label: "Source Branch", Value: "refs/heads/" + branch},
		{Label: "Target Branch", Placeholder: "refs/heads/main", Value: "refs/heads/main"},
		{Label: "Description", Placeholder: "Optional description..."},
	})
	p.form.SetSize(p.width, p.height)
	return nil
}

func (p *PullRequestsPanel) handleFormSubmit(msg components.FormSubmitMsg) tea.Cmd {
	p.form = nil
	if msg.ID != "create-pr" {
		return nil
	}

	title := msg.Values["Title"]
	source := msg.Values["Source Branch"]
	target := msg.Values["Target Branch"]
	desc := msg.Values["Description"]

	if title == "" || source == "" || target == "" {
		return statusCmd("Title, source, and target branches are required", true)
	}

	client := p.client
	repoID := p.repoID
	return func() tea.Msg {
		req := api.CreatePRRequest{
			Title:         title,
			SourceRefName: source,
			TargetRefName: target,
			Description:   desc,
		}
		pr, err := client.CreatePullRequest(context.Background(), repoID, req)
		return ui.PRCreatedMsg{PR: pr, Err: err}
	}
}

func (p *PullRequestsPanel) approvePR() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.prs) {
		return nil
	}
	pr := p.prs[idx]
	client := p.client
	return func() tea.Msg {
		err := client.ApprovePullRequest(context.Background(), pr.Repository.ID, pr.PullRequestID, client.UserID())
		return ui.PRApprovedMsg{Err: err}
	}
}

func (p *PullRequestsPanel) showMergeConfirm() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.prs) {
		return nil
	}
	pr := p.prs[idx]
	p.confirm = components.NewConfirm("merge-pr",
		fmt.Sprintf("Merge PR #%d \"%s\"?", pr.PullRequestID, pr.Title))
	p.confirm.SetWidth(p.width)
	return nil
}

func (p *PullRequestsPanel) completePR() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.prs) {
		return nil
	}
	pr := p.prs[idx]
	client := p.client
	return func() tea.Msg {
		_, err := client.CompletePullRequest(context.Background(), pr.Repository.ID, pr.PullRequestID)
		return ui.PRCompletedMsg{Err: err}
	}
}

func (p *PullRequestsPanel) checkoutBranch() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.prs) {
		return nil
	}
	pr := p.prs[idx]
	branch := pr.SourceBranch()
	g := p.git
	return func() tea.Msg {
		err := g.FetchAndCheckout(branch)
		if err != nil {
			return ui.StatusMsg{Text: "Checkout failed: " + err.Error(), IsError: true}
		}
		return ui.StatusMsg{Text: "Checked out " + branch, IsError: false}
	}
}

func (p *PullRequestsPanel) openInBrowser() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.prs) {
		return nil
	}
	pr := p.prs[idx]
	url := fmt.Sprintf("%s/%s/_git/%s/pullrequest/%d",
		p.client.BaseURL(), p.client.Project(), pr.Repository.Name, pr.PullRequestID)
	utils.OpenBrowser(url, "")
	return nil
}
