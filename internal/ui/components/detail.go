package components

import (
	"fmt"
	"strings"

	"github.com/da3az/lazado/internal/ui"
)

// DetailField is a label-value pair for the detail view.
type DetailField struct {
	Label string
	Value string
}

// DetailView renders a detail pane for the selected item.
type DetailView struct {
	title  string
	fields []DetailField
	body   string // longer text content (e.g., description)
	width  int
	height int
}

// NewDetailView creates a detail view.
func NewDetailView() *DetailView {
	return &DetailView{}
}

// SetContent updates the detail view content.
func (d *DetailView) SetContent(title string, fields []DetailField, body string) {
	d.title = title
	d.fields = fields
	d.body = body
}

// Clear empties the detail view.
func (d *DetailView) Clear() {
	d.title = ""
	d.fields = nil
	d.body = ""
}

// SetSize sets the available dimensions.
func (d *DetailView) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// View renders the detail pane.
func (d *DetailView) View() string {
	if d.title == "" {
		return centerText("Select an item", d.width, d.height)
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
		// Word-wrap the body to fit width
		maxW := d.width - 4
		if maxW < 20 {
			maxW = 20
		}
		wrapped := wordWrap(d.body, maxW)
		b.WriteString(ui.NormalItemStyle.Render(wrapped))
	}

	return b.String()
}

func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	lineLen := 0
	for i, word := range words {
		wLen := len(word)
		if i > 0 && lineLen+1+wLen > width {
			b.WriteByte('\n')
			lineLen = 0
		} else if i > 0 {
			b.WriteByte(' ')
			lineLen++
		}
		b.WriteString(word)
		lineLen += wLen
	}
	return b.String()
}
