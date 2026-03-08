package components

import (
	"fmt"
	"strings"

	"github.com/da3az/lazado/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar shows org/project/branch info and temporary messages at the bottom.
type StatusBar struct {
	org     string
	project string
	branch  string
	message string
	isError bool
	width   int
}

// NewStatusBar creates a status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{}
}

// SetInfo updates the org/project/branch display.
func (s *StatusBar) SetInfo(org, project, branch string) {
	s.org = org
	s.project = project
	s.branch = branch
}

// SetMessage sets a temporary status message.
func (s *StatusBar) SetMessage(msg string, isError bool) {
	s.message = msg
	s.isError = isError
}

// ClearMessage clears the temporary message.
func (s *StatusBar) ClearMessage() {
	s.message = ""
	s.isError = false
}

// SetWidth sets the available width.
func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

// View renders the status bar.
func (s *StatusBar) View() string {
	left := fmt.Sprintf(" lazado  │  %s/%s", extractOrg(s.org), s.project)
	if s.branch != "" {
		left += fmt.Sprintf("  │  %s", s.branch)
	}

	right := "? help"
	if s.message != "" {
		if s.isError {
			right = ui.ErrorStyle.Render(s.message)
		} else {
			right = ui.SuccessStyle.Render(s.message)
		}
	}

	// Fill the remaining space
	gap := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + strings.Repeat(" ", gap) + right
	return ui.StatusBarStyle.Width(s.width).Render(bar)
}

func extractOrg(orgURL string) string {
	// https://dev.azure.com/OrgName -> OrgName
	parts := strings.Split(strings.TrimRight(orgURL, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return orgURL
}
