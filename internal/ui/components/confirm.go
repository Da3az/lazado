package components

import (
	"github.com/da3az/lazado/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmMsg is sent when the user responds to a confirmation dialog.
type ConfirmMsg struct {
	ID        string
	Confirmed bool
}

// Confirm is a yes/no confirmation dialog.
type Confirm struct {
	id       string
	message  string
	selected int // 0 = yes, 1 = no
	width    int
}

// NewConfirm creates a confirmation dialog.
func NewConfirm(id, message string) *Confirm {
	return &Confirm{
		id:       id,
		message:  message,
		selected: 1, // default to "no" for safety
	}
}

// SetWidth sets the dialog width.
func (c *Confirm) SetWidth(w int) {
	c.width = w
}

// Update handles input.
func (c *Confirm) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "tab":
			c.selected = (c.selected + 1) % 2
		case "right", "l", "shift+tab":
			c.selected = (c.selected + 1) % 2
		case "y":
			return func() tea.Msg { return ConfirmMsg{ID: c.id, Confirmed: true} }
		case "n", "esc":
			return func() tea.Msg { return ConfirmMsg{ID: c.id, Confirmed: false} }
		case "enter":
			return func() tea.Msg { return ConfirmMsg{ID: c.id, Confirmed: c.selected == 0} }
		}
	}
	return nil
}

// View renders the dialog.
func (c *Confirm) View() string {
	boxWidth := c.width - 20
	if boxWidth < 30 {
		boxWidth = 30
	}

	yesStyle := ui.DimStyle
	noStyle := ui.DimStyle
	if c.selected == 0 {
		yesStyle = ui.SelectedItemStyle
	} else {
		noStyle = ui.SelectedItemStyle
	}

	buttons := yesStyle.Render(" [Yes] ") + "  " + noStyle.Render(" [No] ")

	content := ui.WarningStyle.Render(c.message) + "\n\n" + buttons

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("220")).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Center)

	return boxStyle.Render(content)
}
