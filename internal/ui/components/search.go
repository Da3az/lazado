package components

import (
	"github.com/da3az/lazado/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchSubmitMsg is sent when the user submits a search query.
type SearchSubmitMsg struct {
	Query string
}

// SearchCancelMsg is sent when the user cancels search.
type SearchCancelMsg struct{}

// Search is a search input component.
type Search struct {
	input  textinput.Model
	active bool
	width  int
}

// NewSearch creates a search component.
func NewSearch() *Search {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 128
	return &Search{input: ti}
}

// Open activates the search input.
func (s *Search) Open() tea.Cmd {
	s.active = true
	s.input.SetValue("")
	return s.input.Focus()
}

// Close deactivates the search input.
func (s *Search) Close() {
	s.active = false
	s.input.Blur()
}

// Active returns whether search is active.
func (s *Search) Active() bool {
	return s.active
}

// SetWidth sets the input width.
func (s *Search) SetWidth(w int) {
	s.width = w
	s.input.Width = w - 10
}

// Update handles input.
func (s *Search) Update(msg tea.Msg) tea.Cmd {
	if !s.active {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.Close()
			return func() tea.Msg { return SearchCancelMsg{} }
		case "enter":
			query := s.input.Value()
			s.Close()
			return func() tea.Msg { return SearchSubmitMsg{Query: query} }
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

// View renders the search input.
func (s *Search) View() string {
	if !s.active {
		return ""
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(0, 1)

	label := ui.HelpKeyStyle.Render("/") + " "
	return style.Render(label + s.input.View())
}
