package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Panel is the interface that all panels must implement.
type Panel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View() string
	DetailView() string
	OverlayView() string // returns overlay content (form/confirm) or "" if none
	HelpKeys() []key.Binding
	HasActiveOverlay() bool
	SetFocused(bool)
	SetSize(w, h int)
}

// AppModel is the root BubbleTea model.
type AppModel struct {
	panels      map[PanelID]Panel
	activePanel PanelID
	tabBar      TabBar
	statusBar   StatusBar
	layout      Layout
	showHelp    bool
	ready       bool
	initialized map[PanelID]bool // tracks which panels have been Init()'d
}

// TabBar wraps the tab bar component reference (defined here to avoid circular import).
type TabBar struct {
	tabs   []string
	active int
	width  int
}

// StatusBar wraps the status bar state.
type StatusBar struct {
	org     string
	project string
	branch  string
	message string
	isError bool
	width   int
}

// NewAppModel creates the root model.
func NewAppModel(panels map[PanelID]Panel, org, project, branch string) *AppModel {
	tabs := []string{
		PanelNames[PanelWorkItems],
		PanelNames[PanelPullRequests],
		PanelNames[PanelPipelines],
		PanelNames[PanelRepos],
	}
	return &AppModel{
		panels:      panels,
		activePanel: PanelWorkItems,
		tabBar:      TabBar{tabs: tabs, active: 0},
		statusBar: StatusBar{
			org:     org,
			project: project,
			branch:  branch,
		},
		initialized: make(map[PanelID]bool),
	}
}

// Init initializes only the active panel (lazy init — others init on first switch).
func (m *AppModel) Init() tea.Cmd {
	// Focus and init only the active panel
	if p, ok := m.panels[m.activePanel]; ok {
		p.SetFocused(true)
		m.initialized[m.activePanel] = true
		return p.Init()
	}
	return nil
}

// Update handles messages.
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = CalculateLayout(msg.Width, msg.Height)
		m.tabBar.width = msg.Width
		m.statusBar.width = msg.Width
		m.ready = true
		// Propagate size to all panels
		for _, panel := range m.panels {
			panel.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyMsg:
		// When a panel overlay (form/confirm/search) is active,
		// skip global keybindings and delegate directly to the panel.
		// Only ctrl+c is allowed to quit globally.
		if panel, ok := m.panels[m.activePanel]; ok && panel.HasActiveOverlay() {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, panel.Update(msg)
		}

		// Global keybindings take priority
		cmd := m.handleGlobalKey(msg)
		if cmd != nil {
			return m, cmd
		}

		// Delegate to active panel
		if panel, ok := m.panels[m.activePanel]; ok {
			return m, panel.Update(msg)
		}

	case StatusMsg:
		m.statusBar.message = msg.Text
		m.statusBar.isError = msg.IsError
		return m, nil

	default:
		// Only delegate non-key messages to the active panel
		if panel, ok := m.panels[m.activePanel]; ok {
			return m, panel.Update(msg)
		}
	}

	return m, nil
}

func (m *AppModel) handleGlobalKey(msg tea.KeyMsg) tea.Cmd {
	// When help overlay is open, only allow closing it
	if m.showHelp {
		switch {
		case key.Matches(msg, Keys.Help), key.Matches(msg, Keys.Back):
			m.showHelp = false
		case key.Matches(msg, Keys.Quit):
			m.showHelp = false
		}
		return nil
	}

	switch {
	case key.Matches(msg, Keys.Quit):
		return tea.Quit
	case key.Matches(msg, Keys.Help):
		m.showHelp = true
		return nil
	case key.Matches(msg, Keys.Panel1):
		return m.switchPanel(PanelWorkItems)
	case key.Matches(msg, Keys.Panel2):
		return m.switchPanel(PanelPullRequests)
	case key.Matches(msg, Keys.Panel3):
		return m.switchPanel(PanelPipelines)
	case key.Matches(msg, Keys.Panel4):
		return m.switchPanel(PanelRepos)
	case key.Matches(msg, Keys.Tab):
		next := PanelID((int(m.activePanel) + 1) % len(m.panels))
		return m.switchPanel(next)
	case key.Matches(msg, Keys.ShiftTab):
		prev := PanelID((int(m.activePanel) - 1 + len(m.panels)) % len(m.panels))
		return m.switchPanel(prev)
	}
	return nil
}

func (m *AppModel) switchPanel(id PanelID) tea.Cmd {
	if _, ok := m.panels[id]; !ok {
		return nil
	}
	// Unfocus current
	if p, ok := m.panels[m.activePanel]; ok {
		p.SetFocused(false)
	}
	m.activePanel = id
	m.tabBar.active = int(id)
	// Focus new
	if p, ok := m.panels[id]; ok {
		p.SetFocused(true)
	}
	// Clear status message on panel switch
	m.statusBar.message = ""

	// Lazy init: if this panel hasn't been initialized yet, do it now
	if !m.initialized[id] {
		m.initialized[id] = true
		if p, ok := m.panels[id]; ok {
			return p.Init()
		}
	}
	return nil
}

// View renders the full UI.
func (m *AppModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	var sections []string

	// Tab bar
	sections = append(sections, m.renderTabBar())

	// Content: list panel + detail panel
	panel, ok := m.panels[m.activePanel]
	if !ok {
		sections = append(sections, "No panel")
	} else {
		listContent := panel.View()
		detailContent := panel.DetailView()

		// Apply border styles using dynamic frame sizes
		listBox := ActiveBorderStyle.
			Width(m.layout.ListWidth - m.layout.HFrame).
			Height(m.layout.ContentHeight).
			Render(listContent)

		detailBox := InactiveBorderStyle.
			Width(m.layout.DetailWidth - m.layout.HFrame).
			Height(m.layout.ContentHeight).
			Render(detailContent)

		content := lipgloss.JoinHorizontal(lipgloss.Top, listBox, detailBox)
		sections = append(sections, content)
	}

	// Status bar
	sections = append(sections, m.renderStatusBar())

	base := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Ensure output fits terminal height exactly
	base = lipgloss.NewStyle().
		MaxHeight(m.layout.Height).
		MaxWidth(m.layout.Width).
		Render(base)

	// Panel overlays (form/confirm) render centered on top of everything
	if panel, ok := m.panels[m.activePanel]; ok {
		if overlay := panel.OverlayView(); overlay != "" {
			base = lipgloss.Place(
				m.layout.Width, m.layout.Height,
				lipgloss.Center, lipgloss.Center,
				overlay,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
			)
		}
	}

	if m.showHelp {
		overlay := m.renderHelp()
		base = lipgloss.Place(
			m.layout.Width, m.layout.Height,
			lipgloss.Center, lipgloss.Center,
			overlay,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}

	return base
}

func (m *AppModel) renderTabBar() string {
	var tabs []string
	for i, tab := range m.tabBar.tabs {
		label := " (" + string(rune('1'+i)) + ") " + tab + " "
		if i == m.tabBar.active {
			tabs = append(tabs, ActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, InactiveTabStyle.Render(label))
		}
	}
	sep := DimStyle.Render(" │ ")
	row := strings.Join(tabs, sep)
	return TabBarStyle.Width(m.tabBar.width).Render(row)
}

func (m *AppModel) renderStatusBar() string {
	left := " lazado"
	if m.statusBar.org != "" {
		// Extract org name from URL
		org := m.statusBar.org
		parts := strings.Split(strings.TrimRight(org, "/"), "/")
		if len(parts) > 0 {
			org = parts[len(parts)-1]
		}
		left += "  │  " + org + "/" + m.statusBar.project
	}
	if m.statusBar.branch != "" {
		left += "  │  " + m.statusBar.branch
	}

	helpHint := StatusKeyStyle.Render("?") + " help "
	right := helpHint
	if m.statusBar.message != "" {
		if m.statusBar.isError {
			right = " " + ErrorStyle.Render(m.statusBar.message) + "  " + helpHint
		} else {
			right = " " + SuccessStyle.Render(m.statusBar.message) + "  " + helpHint
		}
	}

	gap := m.statusBar.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + strings.Repeat(" ", gap) + right
	return StatusBarStyle.Width(m.statusBar.width).Render(bar)
}

func (m *AppModel) renderHelp() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("lazado — Keybindings"))
	b.WriteString("\n\n")

	// Global keys
	b.WriteString(TitleStyle.Render("Global"))
	b.WriteString("\n")
	globalBindings := []key.Binding{
		Keys.Quit, Keys.Tab, Keys.Panel1, Keys.Panel2,
		Keys.Panel3, Keys.Panel4, Keys.Help, Keys.Refresh, Keys.Search,
	}
	for _, k := range globalBindings {
		help := k.Help()
		b.WriteString(HelpKeyStyle.Render("  "+help.Key) + "  " + HelpDescStyle.Render(help.Desc) + "\n")
	}

	// Navigation
	b.WriteString("\n")
	b.WriteString(TitleStyle.Render("Navigation"))
	b.WriteString("\n")
	navBindings := []key.Binding{
		ListNav.Up, ListNav.Down, ListNav.Top, ListNav.Bottom, ListNav.Enter,
	}
	for _, k := range navBindings {
		help := k.Help()
		b.WriteString(HelpKeyStyle.Render("  "+help.Key) + "  " + HelpDescStyle.Render(help.Desc) + "\n")
	}

	// Panel-specific keys
	if panel, ok := m.panels[m.activePanel]; ok {
		b.WriteString("\n")
		b.WriteString(TitleStyle.Render(PanelNames[m.activePanel]))
		b.WriteString("\n")
		for _, k := range panel.HelpKeys() {
			help := k.Help()
			b.WriteString(HelpKeyStyle.Render("  "+help.Key) + "  " + HelpDescStyle.Render(help.Desc) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(DimStyle.Render("Press ? or esc to close"))

	return HelpOverlayStyle.Render(b.String())
}
