package panels

import (
	"context"
	"fmt"
	"time"

	"github.com/da3az/lazado/internal/api"
	"github.com/da3az/lazado/internal/ui"
	"github.com/da3az/lazado/internal/ui/components"
	"github.com/da3az/lazado/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type pipeKeys struct {
	ReRun   key.Binding
	Browser key.Binding
}

var pipeKeyMap = pipeKeys{
	ReRun:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "re-run")),
	Browser: key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "open in browser")),
}

// PipelinesPanel manages the pipelines view.
type PipelinesPanel struct {
	client  *api.Client
	list    *components.List
	detail  *components.DetailView
	runs    []api.PipelineRun
	focused bool
	width   int
	height  int
}

// NewPipelinesPanel creates the pipelines panel.
func NewPipelinesPanel(client *api.Client) *PipelinesPanel {
	return &PipelinesPanel{
		client: client,
		list:   components.NewList("Pipelines"),
		detail: components.NewDetailView(),
	}
}

// Init loads initial data.
func (p *PipelinesPanel) Init() tea.Cmd {
	return p.loadRuns()
}

// SetFocused sets focus state.
func (p *PipelinesPanel) SetFocused(focused bool) {
	p.focused = focused
	p.list.SetFocused(focused)
}

// SetSize updates dimensions.
func (p *PipelinesPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	layout := ui.CalculateLayout(w, h)
	p.list.SetSize(layout.ListWidth-2, layout.ContentHeight)
	p.detail.SetSize(layout.DetailWidth-2, layout.ContentHeight)
}

// Update handles messages.
func (p *PipelinesPanel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ui.PipelinesLoadedMsg:
		p.list.SetLoading(false)
		if msg.Err != nil {
			return statusCmd(msg.Err.Error(), true)
		}
		p.runs = msg.Runs
		p.list.SetItems(p.toListItems(msg.Runs))
		p.updateDetail()
		return nil

	case ui.PipelineTriggeredMsg:
		if msg.Err != nil {
			return statusCmd("Pipeline trigger failed: "+msg.Err.Error(), true)
		}
		return tea.Batch(
			statusCmd(fmt.Sprintf("Pipeline run #%d started", msg.Run.ID), false),
			p.loadRuns(),
		)

	case tea.KeyMsg:
		return p.handleKey(msg)
	}

	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

func (p *PipelinesPanel) handleKey(msg tea.KeyMsg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch {
	case key.Matches(msg, pipeKeyMap.ReRun):
		return p.reRun()
	case key.Matches(msg, pipeKeyMap.Browser):
		return p.openInBrowser()
	case key.Matches(msg, ui.Keys.Refresh):
		return p.loadRuns()
	}

	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

// View renders the panel.
func (p *PipelinesPanel) View() string {
	return p.list.View()
}

// DetailView returns the detail pane.
func (p *PipelinesPanel) DetailView() string {
	return p.detail.View()
}

// HelpKeys returns context keybindings.
func (p *PipelinesPanel) HelpKeys() []key.Binding {
	return []key.Binding{pipeKeyMap.ReRun, pipeKeyMap.Browser}
}

func (p *PipelinesPanel) loadRuns() tea.Cmd {
	p.list.SetLoading(true)
	client := p.client
	return func() tea.Msg {
		runs, err := client.ListPipelineRuns(context.Background(), 20)
		return ui.PipelinesLoadedMsg{Runs: runs, Err: err}
	}
}

func (p *PipelinesPanel) toListItems(runs []api.PipelineRun) []components.ListItem {
	result := make([]components.ListItem, len(runs))
	for i, run := range runs {
		status := run.Result
		if run.State == "inProgress" {
			status = "running"
		}
		styledStatus := ui.PipelineResultStyle(run.Result).Render(utils.PadRight(status, 10))

		result[i] = components.ListItem{
			ID:       fmt.Sprintf("#%d", run.ID),
			Title:    run.Pipeline.Name,
			Subtitle: styledStatus,
		}
	}
	return result
}

func (p *PipelinesPanel) updateDetail() {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.runs) {
		p.detail.Clear()
		return
	}
	run := p.runs[idx]

	duration := ""
	if !run.FinishedDate.IsZero() {
		duration = run.FinishedDate.Sub(run.CreatedDate).Truncate(time.Second).String()
	}

	result := run.Result
	if run.State == "inProgress" {
		result = "In Progress"
	}

	p.detail.SetContent(
		fmt.Sprintf("Pipeline Run #%d", run.ID),
		[]components.DetailField{
			{Label: "Pipeline", Value: run.Pipeline.Name},
			{Label: "State", Value: run.State},
			{Label: "Result", Value: result},
			{Label: "Started", Value: run.CreatedDate.Format("2006-01-02 15:04:05")},
			{Label: "Duration", Value: duration},
		},
		"",
	)
}

func (p *PipelinesPanel) reRun() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.runs) {
		return nil
	}
	run := p.runs[idx]
	client := p.client
	return func() tea.Msg {
		result, err := client.RunPipeline(context.Background(), run.Pipeline.ID)
		return ui.PipelineTriggeredMsg{Run: result, Err: err}
	}
}

func (p *PipelinesPanel) openInBrowser() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.runs) {
		return nil
	}
	run := p.runs[idx]
	url := fmt.Sprintf("%s/%s/_build/results?buildId=%d",
		p.client.BaseURL(), p.client.Project(), run.ID)
	utils.OpenBrowser(url, "")
	return nil
}
