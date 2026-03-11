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
	Options     []string // If non-empty, this is a select/picker field
}

// FormSubmitMsg is sent when the form is submitted.
type FormSubmitMsg struct {
	ID     string
	Values map[string]string
}

// FormCancelMsg is sent when the form is cancelled.
type FormCancelMsg struct{}

// Form is a modal input form with multiple text fields and optional pickers.
type Form struct {
	id       string
	title    string
	fields   []FormField
	inputs   []textinput.Model // text inputs (zero-value for picker fields)
	pickers  []*Picker         // pickers (nil for text fields)
	isPicker []bool            // true if field i uses a picker
	cursor   int               // which field is focused
	width    int
	height   int
}

// NewForm creates a form with the given fields.
func NewForm(id, title string, fields []FormField) *Form {
	inputs := make([]textinput.Model, len(fields))
	pickers := make([]*Picker, len(fields))
	isPicker := make([]bool, len(fields))

	for i, f := range fields {
		if len(f.Options) > 0 {
			pickers[i] = NewPicker(f.Options, f.Value)
			isPicker[i] = true
			if i == 0 {
				pickers[i].Focus()
			}
		} else {
			ti := textinput.New()
			ti.Placeholder = f.Placeholder
			ti.SetValue(f.Value)
			ti.CharLimit = 256
			if i == 0 {
				ti.Focus()
			}
			inputs[i] = ti
		}
	}

	return &Form{
		id:       id,
		title:    title,
		fields:   fields,
		inputs:   inputs,
		pickers:  pickers,
		isPicker: isPicker,
	}
}

// SetSize sets the available dimensions.
func (f *Form) SetSize(w, h int) {
	f.width = w
	f.height = h
	boxWidth := w - 10
	if boxWidth > 80 {
		boxWidth = 80
	}
	if boxWidth < 30 {
		boxWidth = 30
	}
	inputWidth := boxWidth - 10
	if inputWidth < 20 {
		inputWidth = 20
	}
	for i := range f.fields {
		if f.isPicker[i] {
			f.pickers[i].SetWidth(inputWidth)
		} else {
			f.inputs[i].Width = inputWidth
		}
	}
}

// Update handles input.
func (f *Form) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return func() tea.Msg { return FormCancelMsg{} }
		case "tab":
			f.cursor = (f.cursor + 1) % len(f.fields)
			return f.focusCurrent()
		case "shift+tab":
			f.cursor = (f.cursor - 1 + len(f.fields)) % len(f.fields)
			return f.focusCurrent()
		case "enter", "ctrl+s":
			// For picker fields, enter selects the highlighted option
			if f.isPicker[f.cursor] {
				f.pickers[f.cursor].SetValue(f.pickers[f.cursor].SelectedOption())
			}
			return f.submit()
		}
	}

	// Delegate to the focused field
	if f.isPicker[f.cursor] {
		return f.pickers[f.cursor].Update(msg)
	}
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
		if f.isPicker[i] {
			b.WriteString(f.pickers[i].View())
		} else {
			b.WriteString(f.inputs[i].View())
		}
		b.WriteString("\n\n")
	}

	help := ui.DimStyle.Render("tab: next field • enter: submit • esc: cancel")
	b.WriteString(help)

	// Center the form in a box
	content := b.String()
	boxWidth := f.width - 10
	if boxWidth > 80 {
		boxWidth = 80
	}
	if boxWidth < 30 {
		boxWidth = 30
	}
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(boxWidth)

	return boxStyle.Render(content)
}

func (f *Form) focusCurrent() tea.Cmd {
	var cmds []tea.Cmd
	for i := range f.fields {
		if f.isPicker[i] {
			if i == f.cursor {
				cmds = append(cmds, f.pickers[i].Focus())
			} else {
				f.pickers[i].Blur()
			}
		} else {
			if i == f.cursor {
				cmds = append(cmds, f.inputs[i].Focus())
			} else {
				f.inputs[i].Blur()
			}
		}
	}
	return tea.Batch(cmds...)
}

func (f *Form) submit() tea.Cmd {
	values := make(map[string]string, len(f.fields))
	for i, field := range f.fields {
		if f.isPicker[i] {
			values[field.Label] = f.pickers[i].SelectedOption()
		} else {
			values[field.Label] = f.inputs[i].Value()
		}
	}
	return func() tea.Msg {
		return FormSubmitMsg{ID: f.id, Values: values}
	}
}
