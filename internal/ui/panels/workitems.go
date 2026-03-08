package panels

import (
	"context"
	"fmt"
	"strconv"

	"github.com/da3az/lazado/internal/api"
	"github.com/da3az/lazado/internal/ui"
	"github.com/da3az/lazado/internal/ui/components"
	"github.com/da3az/lazado/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Work Items panel keybindings
type wiKeys struct {
	Create      key.Binding
	ChangeState key.Binding
	Edit        key.Binding
	Assign      key.Binding
	Unassign    key.Binding
	Branch      key.Binding
	Browser     key.Binding
	Filter      key.Binding
	TypeFilter  key.Binding
}

var wiKeyMap = wiKeys{
	Create:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create")),
	ChangeState: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "change state")),
	Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Assign:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "assign to me")),
	Unassign:    key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unassign")),
	Branch:      key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "create branch")),
	Browser:     key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "open in browser")),
	Filter:      key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter state")),
	TypeFilter:  key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "filter type")),
}

type stateFilter int

const (
	filterActive stateFilter = iota
	filterClosed
	filterAll
)

// WorkItemsPanel manages the work items view.
type WorkItemsPanel struct {
	client    *api.Client
	list      *components.List
	detail    *components.DetailView
	items     []api.WorkItem
	focused   bool
	width     int
	height    int

	// Filters
	stateFilter stateFilter
	typeFilter  string // "" means all types
	wiTypes     []string

	// Overlay state
	form    *components.Form
	confirm *components.Confirm
	search  *components.Search
}

// NewWorkItemsPanel creates the work items panel.
func NewWorkItemsPanel(client *api.Client) *WorkItemsPanel {
	return &WorkItemsPanel{
		client:     client,
		list:       components.NewList("Work Items"),
		detail:     components.NewDetailView(),
		search:     components.NewSearch(),
		stateFilter: filterActive,
		wiTypes:    []string{"Task", "User Story", "Bug", "Epic", "Feature"},
	}
}

// Init loads initial data.
func (p *WorkItemsPanel) Init() tea.Cmd {
	return p.loadItems()
}

// SetFocused sets panel focus state.
func (p *WorkItemsPanel) SetFocused(focused bool) {
	p.focused = focused
	p.list.SetFocused(focused)
}

// SetSize updates panel dimensions.
func (p *WorkItemsPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	layout := ui.CalculateLayout(w, h)
	p.list.SetSize(layout.ListWidth-2, layout.ContentHeight)
	p.detail.SetSize(layout.DetailWidth-2, layout.ContentHeight)
	p.search.SetWidth(w)
}

// Update handles messages.
func (p *WorkItemsPanel) Update(msg tea.Msg) tea.Cmd {
	// Handle overlays first
	if p.form != nil {
		return p.updateForm(msg)
	}
	if p.confirm != nil {
		return p.updateConfirm(msg)
	}
	if p.search.Active() {
		return p.search.Update(msg)
	}

	switch msg := msg.(type) {
	case ui.WorkItemsLoadedMsg:
		p.list.SetLoading(false)
		if msg.Err != nil {
			return statusCmd(msg.Err.Error(), true)
		}
		p.items = msg.Items
		p.list.SetItems(p.toListItems(msg.Items))
		p.updateDetail()
		return nil

	case ui.WorkItemUpdatedMsg:
		if msg.Err != nil {
			return statusCmd("Update failed: "+msg.Err.Error(), true)
		}
		return tea.Batch(
			statusCmd("Work item updated", false),
			p.loadItems(),
		)

	case ui.WorkItemCreatedMsg:
		if msg.Err != nil {
			return statusCmd("Create failed: "+msg.Err.Error(), true)
		}
		return tea.Batch(
			statusCmd(fmt.Sprintf("Created work item #%d", msg.Item.ID), false),
			p.loadItems(),
		)

	case components.FormSubmitMsg:
		return p.handleFormSubmit(msg)

	case components.FormCancelMsg:
		p.form = nil
		return nil

	case components.ConfirmMsg:
		p.confirm = nil
		return nil

	case components.SearchSubmitMsg:
		if msg.Query != "" {
			return p.searchItems(msg.Query)
		}
		return nil

	case tea.KeyMsg:
		return p.handleKey(msg)
	}

	// Update list navigation
	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

func (p *WorkItemsPanel) handleKey(msg tea.KeyMsg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch {
	case key.Matches(msg, wiKeyMap.Create):
		return p.showCreateForm()
	case key.Matches(msg, wiKeyMap.ChangeState):
		return p.showStateForm()
	case key.Matches(msg, wiKeyMap.Edit):
		return p.showEditForm()
	case key.Matches(msg, wiKeyMap.Assign):
		return p.assignToMe()
	case key.Matches(msg, wiKeyMap.Unassign):
		return p.unassign()
	case key.Matches(msg, wiKeyMap.Browser):
		return p.openInBrowser()
	case key.Matches(msg, wiKeyMap.Filter):
		return p.cycleStateFilter()
	case key.Matches(msg, wiKeyMap.TypeFilter):
		return p.cycleTypeFilter()
	case key.Matches(msg, ui.Keys.Search):
		return p.search.Open()
	case key.Matches(msg, ui.Keys.Refresh):
		return p.loadItems()
	}

	// List navigation
	cmd := p.list.Update(msg)
	p.updateDetail()
	return cmd
}

// View renders the panel.
func (p *WorkItemsPanel) View() string {
	// If overlay is active, render it centered
	if p.form != nil {
		return p.form.View()
	}
	if p.confirm != nil {
		return p.confirm.View()
	}
	if p.search.Active() {
		return p.search.View() + "\n" + p.list.View()
	}
	return p.list.View()
}

// DetailView returns the detail pane content.
func (p *WorkItemsPanel) DetailView() string {
	return p.detail.View()
}

// HelpKeys returns context-sensitive keybindings for the help bar.
func (p *WorkItemsPanel) HelpKeys() []key.Binding {
	return []key.Binding{
		wiKeyMap.Create, wiKeyMap.ChangeState, wiKeyMap.Edit,
		wiKeyMap.Assign, wiKeyMap.Filter, wiKeyMap.TypeFilter,
		wiKeyMap.Browser,
	}
}

// FilterLabel returns a description of the current filter.
func (p *WorkItemsPanel) FilterLabel() string {
	var state string
	switch p.stateFilter {
	case filterActive:
		state = "Active"
	case filterClosed:
		state = "Closed"
	case filterAll:
		state = "All"
	}
	if p.typeFilter != "" {
		return state + " / " + p.typeFilter
	}
	return state
}

// --- Data loading ---

func (p *WorkItemsPanel) loadItems() tea.Cmd {
	p.list.SetLoading(true)
	client := p.client
	sf := p.stateFilter
	tf := p.typeFilter
	return func() tea.Msg {
		wiql := p.buildQuery(sf, tf)
		items, err := client.QueryWorkItems(context.Background(), wiql)
		return ui.WorkItemsLoadedMsg{Items: items, Err: err}
	}
}

func (p *WorkItemsPanel) searchItems(query string) tea.Cmd {
	p.list.SetLoading(true)
	client := p.client
	return func() tea.Msg {
		wiql := api.NewWIQL().
			Select("System.Id", "System.Title", "System.State", "System.WorkItemType", "System.AssignedTo").
			Where(api.And(
				api.EqMe("System.AssignedTo"),
				api.Contains("System.Title", query),
			)).
			OrderByDesc("System.CreatedDate").
			Build()
		items, err := client.QueryWorkItems(context.Background(), wiql)
		return ui.WorkItemsLoadedMsg{Items: items, Err: err}
	}
}

func (p *WorkItemsPanel) buildQuery(sf stateFilter, tf string) string {
	conditions := []string{api.EqMe("System.AssignedTo")}

	switch sf {
	case filterActive:
		conditions = append(conditions, api.Or(
			api.Eq("System.State", "New"),
			api.Eq("System.State", "Active"),
		))
	case filterClosed:
		conditions = append(conditions, api.Or(
			api.Eq("System.State", "Closed"),
			api.Eq("System.State", "Resolved"),
			api.Eq("System.State", "Done"),
		))
	}

	if tf != "" {
		conditions = append(conditions, api.Eq("System.WorkItemType", tf))
	}

	return api.NewWIQL().
		Select("System.Id", "System.Title", "System.State", "System.WorkItemType", "System.AssignedTo").
		Where(api.And(conditions...)).
		OrderByDesc("System.CreatedDate").
		Build()
}

// --- Conversions ---

func (p *WorkItemsPanel) toListItems(items []api.WorkItem) []components.ListItem {
	result := make([]components.ListItem, len(items))
	for i, wi := range items {
		result[i] = components.ListItem{
			ID:       fmt.Sprintf("#%d", wi.ID),
			Title:    wi.Title(),
			Subtitle: ui.StateStyle(wi.State()).Render(utils.PadRight(wi.State(), 10)),
		}
	}
	return result
}

func (p *WorkItemsPanel) updateDetail() {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		p.detail.Clear()
		return
	}
	wi := p.items[idx]
	p.detail.SetContent(
		fmt.Sprintf("Work Item #%d", wi.ID),
		[]components.DetailField{
			{Label: "Title", Value: wi.Title()},
			{Label: "State", Value: wi.State()},
			{Label: "Type", Value: wi.WorkItemType()},
			{Label: "Assigned To", Value: wi.AssignedTo()},
			{Label: "Area Path", Value: wi.AreaPath()},
			{Label: "Iteration", Value: wi.IterationPath()},
			{Label: "Tags", Value: wi.Tags()},
		},
		utils.StripHTML(wi.Description()),
	)
}

// --- Actions ---

func (p *WorkItemsPanel) showCreateForm() tea.Cmd {
	p.form = components.NewForm("create-wi", "Create Work Item", []components.FormField{
		{Label: "Title", Placeholder: "Enter title..."},
		{Label: "Type", Placeholder: "Task", Value: "Task"},
		{Label: "Description", Placeholder: "Enter description (optional)..."},
	})
	p.form.SetSize(p.width, p.height)
	return nil
}

func (p *WorkItemsPanel) showEditForm() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return nil
	}
	wi := p.items[idx]
	p.form = components.NewForm("edit-wi", fmt.Sprintf("Edit #%d", wi.ID), []components.FormField{
		{Label: "Title", Value: wi.Title()},
		{Label: "Description", Value: utils.StripHTML(wi.Description())},
	})
	p.form.SetSize(p.width, p.height)
	return nil
}

func (p *WorkItemsPanel) showStateForm() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return nil
	}
	wi := p.items[idx]
	p.form = components.NewForm("state-wi", fmt.Sprintf("Change State for #%d", wi.ID), []components.FormField{
		{Label: "State", Placeholder: "New, Active, Resolved, Closed", Value: wi.State()},
	})
	p.form.SetSize(p.width, p.height)
	return nil
}

func (p *WorkItemsPanel) handleFormSubmit(msg components.FormSubmitMsg) tea.Cmd {
	p.form = nil
	client := p.client

	switch msg.ID {
	case "create-wi":
		title := msg.Values["Title"]
		wiType := msg.Values["Type"]
		desc := msg.Values["Description"]
		if title == "" {
			return statusCmd("Title is required", true)
		}
		if wiType == "" {
			wiType = "Task"
		}
		return func() tea.Msg {
			ops := []api.PatchOperation{
				{Op: "add", Path: "/fields/System.Title", Value: title},
			}
			if desc != "" {
				ops = append(ops, api.PatchOperation{Op: "add", Path: "/fields/System.Description", Value: desc})
			}
			item, err := client.CreateWorkItem(context.Background(), wiType, ops)
			return ui.WorkItemCreatedMsg{Item: item, Err: err}
		}

	case "edit-wi":
		idx := p.list.SelectedIndex()
		if idx < 0 || idx >= len(p.items) {
			return nil
		}
		wi := p.items[idx]
		title := msg.Values["Title"]
		desc := msg.Values["Description"]
		return func() tea.Msg {
			var ops []api.PatchOperation
			if title != "" && title != wi.Title() {
				ops = append(ops, api.PatchOperation{Op: "replace", Path: "/fields/System.Title", Value: title})
			}
			if desc != utils.StripHTML(wi.Description()) {
				ops = append(ops, api.PatchOperation{Op: "replace", Path: "/fields/System.Description", Value: desc})
			}
			if len(ops) == 0 {
				return ui.WorkItemUpdatedMsg{}
			}
			item, err := client.UpdateWorkItem(context.Background(), wi.ID, ops)
			return ui.WorkItemUpdatedMsg{Item: item, Err: err}
		}

	case "state-wi":
		idx := p.list.SelectedIndex()
		if idx < 0 || idx >= len(p.items) {
			return nil
		}
		wi := p.items[idx]
		state := msg.Values["State"]
		if state == "" || state == wi.State() {
			return nil
		}
		return func() tea.Msg {
			ops := []api.PatchOperation{
				{Op: "replace", Path: "/fields/System.State", Value: state},
			}
			item, err := client.UpdateWorkItem(context.Background(), wi.ID, ops)
			return ui.WorkItemUpdatedMsg{Item: item, Err: err}
		}
	}

	return nil
}

func (p *WorkItemsPanel) assignToMe() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return nil
	}
	wi := p.items[idx]
	client := p.client
	return func() tea.Msg {
		ops := []api.PatchOperation{
			{Op: "replace", Path: "/fields/System.AssignedTo", Value: ""},
		}
		// Use @Me - Azure DevOps resolves this to the authenticated user
		ops[0].Value = "@Me"
		item, err := client.UpdateWorkItem(context.Background(), wi.ID, ops)
		return ui.WorkItemUpdatedMsg{Item: item, Err: err}
	}
}

func (p *WorkItemsPanel) unassign() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return nil
	}
	wi := p.items[idx]
	client := p.client
	return func() tea.Msg {
		ops := []api.PatchOperation{
			{Op: "replace", Path: "/fields/System.AssignedTo", Value: ""},
		}
		item, err := client.UpdateWorkItem(context.Background(), wi.ID, ops)
		return ui.WorkItemUpdatedMsg{Item: item, Err: err}
	}
}

func (p *WorkItemsPanel) openInBrowser() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return nil
	}
	wi := p.items[idx]
	url := fmt.Sprintf("%s/%s/_workitems/edit/%d", p.client.BaseURL(), p.client.Project(), wi.ID)
	utils.OpenBrowser(url, "")
	return nil
}

func (p *WorkItemsPanel) cycleStateFilter() tea.Cmd {
	p.stateFilter = (p.stateFilter + 1) % 3
	return p.loadItems()
}

func (p *WorkItemsPanel) cycleTypeFilter() tea.Cmd {
	if p.typeFilter == "" {
		p.typeFilter = p.wiTypes[0]
	} else {
		for i, t := range p.wiTypes {
			if t == p.typeFilter {
				if i == len(p.wiTypes)-1 {
					p.typeFilter = ""
				} else {
					p.typeFilter = p.wiTypes[i+1]
				}
				break
			}
		}
	}
	return p.loadItems()
}

func (p *WorkItemsPanel) updateForm(msg tea.Msg) tea.Cmd {
	return p.form.Update(msg)
}

func (p *WorkItemsPanel) updateConfirm(msg tea.Msg) tea.Cmd {
	return p.confirm.Update(msg)
}

// selectedWIID returns the ID of the selected work item as a string.
func (p *WorkItemsPanel) selectedWIID() string {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return ""
	}
	return strconv.Itoa(p.items[idx].ID)
}

func statusCmd(msg string, isErr bool) tea.Cmd {
	return func() tea.Msg {
		return ui.StatusMsg{Text: msg, IsError: isErr}
	}
}
