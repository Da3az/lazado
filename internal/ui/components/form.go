package components

import (
	"strings"

	"github.com/da3az/lazado/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField defines a single field in a form.
type FormField struct {
	Label       string
	Placeholder string
	Value       string
	Options     []string // If non-empty, this is a select field
}

// FormSubmitMsg is sent when the form is submitted.
type FormSubmitMsg struct {
	ID     string
	Values map[string]string
}

// FormCancelMsg is sent when the form is cancelled.
type FormCancelMsg struct{}

// Form is a modal input form with multiple text fields.
type Form struct {
	id       string
	title    string
	fields   []FormField
	inputs   []textinput.Model
	selected int // for select fields, which option is selected
	cursor   int // which field is focused
	width    int
	height   int
}

// NewForm creates a form with the given fields.
func NewForm(id, title string, fields []FormField) *Form {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.SetValue(f.Value)
		ti.CharLimit = 256
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}

	return &Form{
		id:     id,
		title:  title,
		fields: fields,
		inputs: inputs,
	}
}

// SetSize sets the available dimensions.
func (f *Form) SetSize(w, h int) {
	f.width = w
	f.height = h
	inputWidth := w - 20
	if inputWidth < 20 {
		inputWidth = 20
	}
	for i := range f.inputs {
		f.inputs[i].Width = inputWidth
	}
}

// Update handles input.
func (f *Form) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return func() tea.Msg { return FormCancelMsg{} }
		case "tab", "down":
			f.cursor = (f.cursor + 1) % len(f.inputs)
			return f.focusCurrent()
		case "shift+tab", "up":
			f.cursor = (f.cursor - 1 + len(f.inputs)) % len(f.inputs)
			return f.focusCurrent()
		case "enter":
			// If on last field, submit
			if f.cursor == len(f.inputs)-1 {
				return f.submit()
			}
			// Otherwise move to next field
			f.cursor = (f.cursor + 1) % len(f.inputs)
			return f.focusCurrent()
		case "ctrl+s":
			return f.submit()
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	f.inputs[f.cursor], cmd = f.inputs[f.cursor].Update(msg)
	return cmd
}

// View renders the form.
func (f *Form) View() string {
	var b strings.Builder

	titleStr := ui.TitleStyle.Render(f.title)
	b.WriteString(titleStr)
	b.WriteString("\n\n")

	for i, field := range f.fields {
		label := ui.DetailLabelStyle.Render(field.Label + ":")
		b.WriteString(label + "\n")
		b.WriteString(f.inputs[i].View())
		b.WriteString("\n\n")
	}

	help := ui.DimStyle.Render("tab: next field • ctrl+s: submit • esc: cancel")
	b.WriteString(help)

	// Center the form in a box
	content := b.String()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(f.width - 10)

	return boxStyle.Render(content)
}

func (f *Form) focusCurrent() tea.Cmd {
	cmds := make([]tea.Cmd, len(f.inputs))
	for i := range f.inputs {
		if i == f.cursor {
			cmds[i] = f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (f *Form) submit() tea.Cmd {
	values := make(map[string]string, len(f.fields))
	for i, field := range f.fields {
		values[field.Label] = f.inputs[i].Value()
	}
	return func() tea.Msg {
		return FormSubmitMsg{ID: f.id, Values: values}
	}
}
