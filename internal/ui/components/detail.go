package components

import (
	"fmt"
	"strings"

	"github.com/da3az/lazado/internal/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// DetailField is a label-value pair for the detail view.
type DetailField struct {
	Label string
	Value string
}

// DetailView renders a scrollable detail pane for the selected item.
type DetailView struct {
	title    string
	fields   []DetailField
	body     string // longer text content (e.g., description)
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewDetailView creates a detail view.
func NewDetailView() *DetailView {
	return &DetailView{}
}

// SetContent updates the detail view content and resets scroll to top.
func (d *DetailView) SetContent(title string, fields []DetailField, body string) {
	d.title = title
	d.fields = fields
	d.body = body
	d.updateViewport()
}

// Clear empties the detail view.
func (d *DetailView) Clear() {
	d.title = ""
	d.fields = nil
	d.body = ""
	if d.ready {
		d.viewport.SetContent("")
		d.viewport.GotoTop()
	}
}

// SetSize sets the available dimensions.
func (d *DetailView) SetSize(w, h int) {
	d.width = w
	d.height = h
	if !d.ready {
		d.viewport = viewport.New(w, h)
		d.viewport.Style = ui.NormalItemStyle
		d.ready = true
	} else {
		d.viewport.Width = w
		d.viewport.Height = h
	}
	d.updateViewport()
}

// Update handles scroll input for the detail view.
func (d *DetailView) Update(msg tea.Msg) tea.Cmd {
	if !d.ready {
		return nil
	}
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return cmd
}

// View renders the detail pane.
func (d *DetailView) View() string {
	if d.title == "" {
		return centerText("Select an item", d.width, d.height)
	}
	if !d.ready {
		return ""
	}
	return d.viewport.View()
}

// ScrollPercent returns the viewport scroll percentage (0-100).
func (d *DetailView) ScrollPercent() float64 {
	if !d.ready {
		return 0
	}
	return d.viewport.ScrollPercent()
}

func (d *DetailView) updateViewport() {
	if !d.ready || d.title == "" {
		return
	}

	var b strings.Builder

	// Title
	b.WriteString(ui.DetailTitleStyle.Render(d.title))
	b.WriteString("\n\n")

	// Fields
	for _, f := range d.fields {
		label := ui.DetailLabelStyle.Render(f.Label + ":")
		value := ui.DetailValueStyle.Render(f.Value)
		b.WriteString(fmt.Sprintf("%s %s\n", label, value))
	}

	// Body
	if d.body != "" {
		b.WriteString("\n")
		maxW := d.width - 4
		if maxW < 20 {
			maxW = 20
		}
		wrapped := wordWrap(d.body, maxW)
		b.WriteString(ui.NormalItemStyle.Render(wrapped))
	}

	d.viewport.SetContent(b.String())
	d.viewport.GotoTop()
}

// wordWrap wraps text to the given width, preserving existing newlines.
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if line == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		var current strings.Builder
		lineLen := 0
		for i, word := range words {
			wLen := len(word)
			if i > 0 && lineLen+1+wLen > width {
				result = append(result, current.String())
				current.Reset()
				lineLen = 0
			} else if i > 0 {
				current.WriteByte(' ')
				lineLen++
			}
			current.WriteString(word)
			lineLen += wLen
		}
		result = append(result, current.String())
	}
	return strings.Join(result, "\n")
}
