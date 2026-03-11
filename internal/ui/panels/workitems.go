package panels

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	Comment     key.Binding
	Branch      key.Binding
	Browser     key.Binding
	Filter      key.Binding
	TypeFilter  key.Binding
	ToggleMine  key.Binding
}

var wiKeyMap = wiKeys{
	Create:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create")),
	ChangeState: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "change state")),
	Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Assign:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "assign to me")),
	Unassign:    key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unassign")),
	Comment:     key.NewBinding(key.WithKeys("C"), key.WithHelp("C", "add comment")),
	Branch:      key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "create branch")),
	Browser:     key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "open in browser")),
	Filter:      key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter state")),
	TypeFilter:  key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "filter type")),
	ToggleMine:  key.NewBinding(key.WithKeys("F"), key.WithHelp("F", "mine / all")),
}

// Default state filters
var defaultStateFilters = []string{"All", "New", "Active", "Resolved", "Closed", "Done", "Removed"}

// WorkItemsPanel manages the work items view.
type WorkItemsPanel struct {
	client  *api.Client
	list    *components.List
	detail  *components.DetailView
	items   []api.WorkItem
	focused bool
	width   int
	height  int

	// Filters
	stateFilters  []string // available state filter options
	stateFilterIdx int     // index into stateFilters
	typeFilter    string   // "" means all types
	wiTypes       []string
	myItems       bool // true = assigned to me, false = all items

	// Search
	searchQuery string

	// Metadata for form pickers
	teamMembers []string
	areaPaths   []string
	iterations  []string

	// Comments
	comments       []api.Comment
	commentsLoaded int // work item ID for which comments are loaded

	// Overlay state
	form    *components.Form
	confirm *components.Confirm
	search  *components.Search
}

// NewWorkItemsPanel creates the work items panel.
func NewWorkItemsPanel(client *api.Client) *WorkItemsPanel {
	list := components.NewList("Work Items")
	list.SetColumns([]components.ColumnDef{
		{Field: "ID", MinWidth: 10},
		{Field: "Subtitle", MinWidth: 20},
		{Field: "Title", Flex: true},
	})
	return &WorkItemsPanel{
		client:        client,
		list:          list,
		detail:        components.NewDetailView(),
		search:        components.NewSearch(),
		stateFilters:  defaultStateFilters,
		stateFilterIdx: 0, // "All" initially — but will load on first refresh
		wiTypes:       []string{"Task", "User Story", "Bug", "Epic", "Feature"},
		myItems:       false,
	}
}

// Init loads initial data.
func (p *WorkItemsPanel) Init() tea.Cmd {
	return tea.Batch(
		p.loadItems(),
		p.loadStates(),
		p.loadTeamMembers(),
		p.loadAreaPaths(),
		p.loadIterations(),
	)
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
	p.list.SetSize(layout.ListWidth-layout.HFrame, layout.ContentHeight)
	p.detail.SetSize(layout.DetailWidth-layout.HFrame, layout.ContentHeight)
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
		// Load comments for the first selected item
		if idx := p.list.SelectedIndex(); idx >= 0 && idx < len(p.items) {
			return p.loadComments(p.items[idx].ID)
		}
		return nil

	case ui.StatesLoadedMsg:
		if msg.Err == nil && len(msg.States) > 0 {
			// Build unique state names
			seen := map[string]bool{}
			states := []string{"All"}
			for _, s := range msg.States {
				if !seen[s.Name] {
					seen[s.Name] = true
					states = append(states, s.Name)
				}
			}
			p.stateFilters = states
		}
		return nil

	case ui.TeamMembersLoadedMsg:
		if msg.Err == nil {
			p.teamMembers = msg.Members
		}
		return nil

	case ui.AreaPathsLoadedMsg:
		if msg.Err == nil {
			p.areaPaths = msg.Paths
		}
		return nil

	case ui.IterationsLoadedMsg:
		if msg.Err == nil {
			p.iterations = msg.Paths
		}
		return nil

	case ui.CommentsLoadedMsg:
		if msg.Err == nil {
			p.comments = msg.Comments
			p.commentsLoaded = msg.WIID
			p.updateDetail()
		}
		return nil

	case ui.CommentAddedMsg:
		if msg.Err != nil {
			return statusCmd("Failed to add comment: "+msg.Err.Error(), true)
		}
		return tea.Batch(
			statusCmd("Comment added", false),
			p.loadComments(msg.WIID),
		)

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
			p.searchQuery = msg.Query
			return p.searchItems(msg.Query)
		}
		p.searchQuery = ""
		return nil

	case components.SearchCancelMsg:
		p.searchQuery = ""
		return p.loadItems()

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
	case key.Matches(msg, wiKeyMap.Comment):
		return p.showCommentForm()
	case key.Matches(msg, wiKeyMap.Browser):
		return p.openInBrowser()
	case key.Matches(msg, wiKeyMap.Filter):
		return p.cycleStateFilter()
	case key.Matches(msg, wiKeyMap.TypeFilter):
		return p.cycleTypeFilter()
	case key.Matches(msg, wiKeyMap.ToggleMine):
		return p.toggleMyItems()
	case key.Matches(msg, ui.Keys.Search):
		return p.search.Open()
	case key.Matches(msg, ui.Keys.Refresh):
		p.searchQuery = ""
		return p.loadItems()
	}

	// List navigation
	prevIdx := p.list.SelectedIndex()
	cmd := p.list.Update(msg)
	p.updateDetail()
	// Load comments if selection changed
	newIdx := p.list.SelectedIndex()
	if newIdx >= 0 && newIdx < len(p.items) && newIdx != prevIdx {
		return tea.Batch(cmd, p.loadComments(p.items[newIdx].ID))
	}
	return cmd
}

// View renders the panel (list only — overlays are rendered via OverlayView).
func (p *WorkItemsPanel) View() string {
	if p.search.Active() {
		return p.search.View() + "\n" + p.list.View()
	}
	return p.list.View()
}

// DetailView returns the detail pane content.
func (p *WorkItemsPanel) DetailView() string {
	return p.detail.View()
}

// OverlayView returns the overlay content (form/confirm) or empty string.
func (p *WorkItemsPanel) OverlayView() string {
	if p.form != nil {
		return p.form.View()
	}
	if p.confirm != nil {
		return p.confirm.View()
	}
	return ""
}

// HasActiveOverlay returns true when a form, confirm, or search is open.
func (p *WorkItemsPanel) HasActiveOverlay() bool {
	return p.form != nil || p.confirm != nil || p.search.Active()
}

// HelpKeys returns context-sensitive keybindings for the help bar.
func (p *WorkItemsPanel) HelpKeys() []key.Binding {
	return []key.Binding{
		wiKeyMap.Create, wiKeyMap.ChangeState, wiKeyMap.Edit,
		wiKeyMap.Assign, wiKeyMap.Comment, wiKeyMap.Filter,
		wiKeyMap.TypeFilter, wiKeyMap.ToggleMine, wiKeyMap.Browser,
	}
}

// currentStateFilter returns the current state filter string.
func (p *WorkItemsPanel) currentStateFilter() string {
	if p.stateFilterIdx >= 0 && p.stateFilterIdx < len(p.stateFilters) {
		return p.stateFilters[p.stateFilterIdx]
	}
	return "All"
}

// FilterLabel returns a description of the current filter.
func (p *WorkItemsPanel) FilterLabel() string {
	state := p.currentStateFilter()
	scope := "Mine"
	if !p.myItems {
		scope = "All"
	}
	label := state + " / " + scope
	if p.typeFilter != "" {
		label += " / " + p.typeFilter
	}
	if p.searchQuery != "" {
		label += " / Search: " + p.searchQuery
	}
	return label
}

// --- Data loading ---

func (p *WorkItemsPanel) loadStates() tea.Cmd {
	client := p.client
	return func() tea.Msg {
		// Fetch states for Task type as representative
		states, err := client.GetWorkItemTypeStates(context.Background(), "Task")
		return ui.StatesLoadedMsg{Type: "Task", States: states, Err: err}
	}
}

func (p *WorkItemsPanel) loadItems() tea.Cmd {
	p.list.SetLoading(true)
	client := p.client
	sf := p.currentStateFilter()
	tf := p.typeFilter
	myItems := p.myItems
	return func() tea.Msg {
		wiql := buildQuery(sf, tf, myItems, "")
		items, err := client.QueryWorkItems(context.Background(), wiql)
		return ui.WorkItemsLoadedMsg{Items: items, Err: err}
	}
}

func (p *WorkItemsPanel) searchItems(query string) tea.Cmd {
	p.list.SetLoading(true)
	client := p.client
	sf := p.currentStateFilter()
	tf := p.typeFilter
	myItems := p.myItems
	return func() tea.Msg {
		wiql := buildQuery(sf, tf, myItems, query)
		items, err := client.QueryWorkItems(context.Background(), wiql)
		return ui.WorkItemsLoadedMsg{Items: items, Err: err}
	}
}

func buildQuery(stateFilter, typeFilter string, myItems bool, searchQuery string) string {
	var conditions []string

	if myItems {
		conditions = append(conditions, api.EqMe("System.AssignedTo"))
	}

	if stateFilter != "" && stateFilter != "All" {
		conditions = append(conditions, api.Eq("System.State", stateFilter))
	}

	if typeFilter != "" {
		conditions = append(conditions, api.Eq("System.WorkItemType", typeFilter))
	}

	if searchQuery != "" {
		conditions = append(conditions, api.Contains("System.Title", searchQuery))
	}

	q := api.NewWIQL().
		Select("System.Id", "System.Title", "System.State", "System.WorkItemType", "System.AssignedTo")

	if len(conditions) > 0 {
		q = q.Where(api.And(conditions...))
	}

	return q.OrderByDesc("System.CreatedDate").Build()
}

// --- Conversions ---

func (p *WorkItemsPanel) toListItems(items []api.WorkItem) []components.ListItem {
	result := make([]components.ListItem, len(items))
	for i, wi := range items {
		result[i] = components.ListItem{
			ID:       fmt.Sprintf("#%d", wi.ID),
			Title:    wi.Title(),
			Subtitle: ui.StateStyle(wi.State()).Render(wi.State()),
		}
	}
	return result
}

func (p *WorkItemsPanel) updateDetail() {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		p.detail.Clear()
		p.comments = nil
		return
	}
	wi := p.items[idx]

	// Build body with description and comments
	bodyText := utils.StripHTML(wi.Description())
	if p.commentsLoaded == wi.ID && len(p.comments) > 0 {
		var parts []string
		for _, c := range p.comments {
			header := fmt.Sprintf("%s (%s):",
				c.CreatedBy.DisplayName,
				c.CreatedDate.Format("Jan 2 15:04"))
			body := utils.StripHTML(c.Text)
			parts = append(parts, header+"\n"+body)
		}
		bodyText += "\n\n--- Comments ---\n\n" + strings.Join(parts, "\n---\n")
	}

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
		bodyText,
	)
}

// --- Filter actions ---

func (p *WorkItemsPanel) cycleStateFilter() tea.Cmd {
	p.stateFilterIdx = (p.stateFilterIdx + 1) % len(p.stateFilters)
	return tea.Batch(
		statusCmd("Filter: "+p.currentStateFilter(), false),
		p.loadItems(),
	)
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
	label := "All types"
	if p.typeFilter != "" {
		label = p.typeFilter
	}
	return tea.Batch(
		statusCmd("Type: "+label, false),
		p.loadItems(),
	)
}

func (p *WorkItemsPanel) toggleMyItems() tea.Cmd {
	p.myItems = !p.myItems
	scope := "My Items"
	if !p.myItems {
		scope = "All Items"
	}
	return tea.Batch(
		statusCmd("Scope: "+scope, false),
		p.loadItems(),
	)
}

// --- Form actions ---

func (p *WorkItemsPanel) showCreateForm() tea.Cmd {
	p.form = components.NewForm("create-wi", "Create Work Item", []components.FormField{
		{Label: "Title", Placeholder: "Enter title..."},
		{Label: "Type", Value: "Task", Options: p.wiTypes},
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

	// Build state options from loaded states (skip "All")
	var stateOptions []string
	if len(p.stateFilters) > 1 {
		stateOptions = p.stateFilters[1:]
	}

	p.form = components.NewForm("edit-wi", fmt.Sprintf("Edit #%d", wi.ID), []components.FormField{
		{Label: "Title", Value: wi.Title()},
		{Label: "State", Value: wi.State(), Options: stateOptions},
		{Label: "Assigned To", Value: wi.AssignedTo(), Options: p.teamMembers},
		{Label: "Area Path", Value: wi.AreaPath(), Options: p.areaPaths},
		{Label: "Iteration", Value: wi.IterationPath(), Options: p.iterations},
		{Label: "Tags", Value: wi.Tags()},
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

	var stateOptions []string
	if len(p.stateFilters) > 1 {
		stateOptions = p.stateFilters[1:]
	}

	p.form = components.NewForm("state-wi", fmt.Sprintf("Change State for #%d", wi.ID), []components.FormField{
		{Label: "State", Value: wi.State(), Options: stateOptions},
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
		return p.submitEditForm(wi, msg.Values)

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

	case "comment-wi":
		idx := p.list.SelectedIndex()
		if idx < 0 || idx >= len(p.items) {
			return nil
		}
		wi := p.items[idx]
		text := msg.Values["Comment"]
		if text == "" {
			return statusCmd("Comment text is required", true)
		}
		wiID := wi.ID
		return func() tea.Msg {
			comment, err := client.AddWorkItemComment(context.Background(), wiID, text)
			return ui.CommentAddedMsg{WIID: wiID, Comment: comment, Err: err}
		}
	}

	return nil
}

func (p *WorkItemsPanel) submitEditForm(wi api.WorkItem, values map[string]string) tea.Cmd {
	client := p.client
	wiID := wi.ID

	// Field mappings: form label -> Azure DevOps field path
	fieldMap := map[string]struct {
		path   string
		oldVal string
	}{
		"Title":       {path: "/fields/System.Title", oldVal: wi.Title()},
		"State":       {path: "/fields/System.State", oldVal: wi.State()},
		"Assigned To": {path: "/fields/System.AssignedTo", oldVal: wi.AssignedTo()},
		"Area Path":   {path: "/fields/System.AreaPath", oldVal: wi.AreaPath()},
		"Iteration":   {path: "/fields/System.IterationPath", oldVal: wi.IterationPath()},
		"Tags":        {path: "/fields/System.Tags", oldVal: wi.Tags()},
		"Description": {path: "/fields/System.Description", oldVal: utils.StripHTML(wi.Description())},
	}

	return func() tea.Msg {
		var ops []api.PatchOperation
		for label, fm := range fieldMap {
			newVal, ok := values[label]
			if !ok {
				continue
			}
			if newVal != fm.oldVal {
				ops = append(ops, api.PatchOperation{
					Op:    "replace",
					Path:  fm.path,
					Value: newVal,
				})
			}
		}
		if len(ops) == 0 {
			return ui.WorkItemUpdatedMsg{}
		}
		item, err := client.UpdateWorkItem(context.Background(), wiID, ops)
		return ui.WorkItemUpdatedMsg{Item: item, Err: err}
	}
}

// --- Other actions ---

func (p *WorkItemsPanel) assignToMe() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return nil
	}
	wi := p.items[idx]
	client := p.client
	return func() tea.Msg {
		ops := []api.PatchOperation{
			{Op: "replace", Path: "/fields/System.AssignedTo", Value: "@Me"},
		}
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

func (p *WorkItemsPanel) showCommentForm() tea.Cmd {
	idx := p.list.SelectedIndex()
	if idx < 0 || idx >= len(p.items) {
		return nil
	}
	wi := p.items[idx]
	p.form = components.NewForm("comment-wi", fmt.Sprintf("Comment on #%d", wi.ID), []components.FormField{
		{Label: "Comment", Placeholder: "Enter your comment..."},
	})
	p.form.SetSize(p.width, p.height)
	return nil
}

func (p *WorkItemsPanel) loadComments(wiID int) tea.Cmd {
	client := p.client
	return func() tea.Msg {
		comments, err := client.GetWorkItemComments(context.Background(), wiID)
		return ui.CommentsLoadedMsg{WIID: wiID, Comments: comments, Err: err}
	}
}

func (p *WorkItemsPanel) loadTeamMembers() tea.Cmd {
	client := p.client
	return func() tea.Msg {
		members, err := client.GetTeamMembers(context.Background())
		return ui.TeamMembersLoadedMsg{Members: members, Err: err}
	}
}

func (p *WorkItemsPanel) loadAreaPaths() tea.Cmd {
	client := p.client
	return func() tea.Msg {
		paths, err := client.GetAreaPaths(context.Background())
		return ui.AreaPathsLoadedMsg{Paths: paths, Err: err}
	}
}

func (p *WorkItemsPanel) loadIterations() tea.Cmd {
	client := p.client
	return func() tea.Msg {
		paths, err := client.GetIterationPaths(context.Background())
		return ui.IterationsLoadedMsg{Paths: paths, Err: err}
	}
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
