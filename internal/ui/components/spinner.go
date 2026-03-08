package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Spinner wraps the bubbles spinner for loading states.
type Spinner struct {
	spinner spinner.Model
	message string
	active  bool
}

// NewSpinner creates a spinner.
func NewSpinner() *Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	return &Spinner{spinner: s}
}

// Start activates the spinner with a message.
func (s *Spinner) Start(msg string) tea.Cmd {
	s.active = true
	s.message = msg
	return s.spinner.Tick
}

// Stop deactivates the spinner.
func (s *Spinner) Stop() {
	s.active = false
	s.message = ""
}

// Active returns whether the spinner is active.
func (s *Spinner) Active() bool {
	return s.active
}

// Update handles spinner animation ticks.
func (s *Spinner) Update(msg tea.Msg) tea.Cmd {
	if !s.active {
		return nil
	}
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return cmd
}

// View renders the spinner.
func (s *Spinner) View() string {
	if !s.active {
		return ""
	}
	return s.spinner.View() + " " + s.message
}
