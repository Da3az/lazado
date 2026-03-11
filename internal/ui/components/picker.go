package components

import (
	"strings"

	"github.com/da3az/lazado/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Picker is a filterable select component for form fields.
type Picker struct {
	options  []string
	filtered []string
	cursor   int
	filter   textinput.Model
	selected string
	width    int
}

// NewPicker creates a picker with the given options and current value.
func NewPicker(options []string, currentValue string) *Picker {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 128

	p := &Picker{
		options:  options,
		filtered: options,
		filter:   ti,
		selected: currentValue,
	}

	// Pre-select the current value
	for i, opt := range options {
		if opt == currentValue {
			p.cursor = i
			break
		}
	}

	return p
}

// SetWidth sets the picker width.
func (p *Picker) SetWidth(w int) {
	p.width = w
	p.filter.Width = w - 4
}

// Focus activates the picker input.
func (p *Picker) Focus() tea.Cmd {
	return p.filter.Focus()
}

// Blur deactivates the picker input.
func (p *Picker) Blur() {
	p.filter.Blur()
}

// Value returns the selected value.
func (p *Picker) Value() string {
	return p.selected
}

// SetValue sets the selected value.
func (p *Picker) SetValue(v string) {
	p.selected = v
}

// SelectedOption returns the currently highlighted option.
func (p *Picker) SelectedOption() string {
	if p.cursor >= 0 && p.cursor < len(p.filtered) {
		return p.filtered[p.cursor]
	}
	return p.selected
}

// Update handles input for the picker.
func (p *Picker) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "ctrl+p":
			if p.cursor > 0 {
				p.cursor--
			}
			return nil
		case "down", "ctrl+n":
			if p.cursor < len(p.filtered)-1 {
				p.cursor++
			}
			return nil
		}
	}

	// Update filter input
	var cmd tea.Cmd
	p.filter, cmd = p.filter.Update(msg)

	// Re-filter options based on typed text
	query := strings.ToLower(p.filter.Value())
	if query == "" {
		p.filtered = p.options
	} else {
		p.filtered = nil
		for _, opt := range p.options {
			if strings.Contains(strings.ToLower(opt), query) {
				p.filtered = append(p.filtered, opt)
			}
		}
	}
	if p.cursor >= len(p.filtered) {
		if len(p.filtered) > 0 {
			p.cursor = len(p.filtered) - 1
		} else {
			p.cursor = 0
		}
	}

	return cmd
}

// View renders the picker.
func (p *Picker) View() string {
	var b strings.Builder
	b.WriteString(p.filter.View())
	b.WriteString("\n")

	maxVisible := 6
	start := 0
	if p.cursor >= maxVisible {
		start = p.cursor - maxVisible + 1
	}

	for i := start; i < len(p.filtered) && i < start+maxVisible; i++ {
		prefix := "  "
		style := ui.NormalItemStyle
		if i == p.cursor {
			prefix = "> "
			style = ui.SelectedItemStyle
		}
		b.WriteString(style.Render(prefix + p.filtered[i]))
		if i < start+maxVisible-1 && i < len(p.filtered)-1 {
			b.WriteString("\n")
		}
	}

	if len(p.filtered) == 0 {
		b.WriteString(ui.DimStyle.Render("  No matches"))
	}

	return b.String()
}
