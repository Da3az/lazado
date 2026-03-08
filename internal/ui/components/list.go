package components

import (
	"strings"

	"github.com/da3az/lazado/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem is an item that can be displayed in a List.
type ListItem struct {
	ID       string
	Title    string
	Subtitle string // secondary info (state, status, etc.)
	Extra    string // right-aligned extra info
}

// List is a scrollable, selectable list component.
type List struct {
	items    []ListItem
	cursor   int
	offset   int // scroll offset
	width    int
	height   int
	focused  bool
	title    string
	loading  bool
}

// NewList creates a new List.
func NewList(title string) *List {
	return &List{title: title}
}

// SetItems replaces the list contents.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	if l.cursor >= len(items) {
		l.cursor = max(0, len(items)-1)
	}
	l.clampOffset()
}

// Items returns the current items.
func (l *List) Items() []ListItem {
	return l.items
}

// SetSize sets the available dimensions.
func (l *List) SetSize(w, h int) {
	l.width = w
	l.height = h
}

// SetFocused sets whether this list is focused.
func (l *List) SetFocused(focused bool) {
	l.focused = focused
}

// SetLoading sets the loading state.
func (l *List) SetLoading(loading bool) {
	l.loading = loading
}

// SelectedIndex returns the index of the selected item, or -1 if empty.
func (l *List) SelectedIndex() int {
	if len(l.items) == 0 {
		return -1
	}
	return l.cursor
}

// SelectedItem returns the currently selected item, or nil if empty.
func (l *List) SelectedItem() *ListItem {
	if l.cursor >= 0 && l.cursor < len(l.items) {
		return &l.items[l.cursor]
	}
	return nil
}

// Update handles keyboard input.
func (l *List) Update(msg tea.Msg) tea.Cmd {
	if !l.focused || len(l.items) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, ui.ListNav.Down):
			if l.cursor < len(l.items)-1 {
				l.cursor++
				l.clampOffset()
			}
		case key.Matches(msg, ui.ListNav.Up):
			if l.cursor > 0 {
				l.cursor--
				l.clampOffset()
			}
		case key.Matches(msg, ui.ListNav.Top):
			l.cursor = 0
			l.offset = 0
		case key.Matches(msg, ui.ListNav.Bottom):
			l.cursor = len(l.items) - 1
			l.clampOffset()
		}
	}
	return nil
}

// View renders the list.
func (l *List) View() string {
	if l.loading {
		return centerText("Loading...", l.width, l.height)
	}
	if len(l.items) == 0 {
		return centerText("No items", l.width, l.height)
	}

	visible := l.visibleCount()
	var b strings.Builder

	for i := l.offset; i < l.offset+visible && i < len(l.items); i++ {
		item := l.items[i]
		line := l.renderItem(item, i == l.cursor)
		b.WriteString(line)
		if i < l.offset+visible-1 {
			b.WriteByte('\n')
		}
	}

	// Pad remaining lines
	rendered := strings.Count(b.String(), "\n") + 1
	for rendered < l.height {
		b.WriteByte('\n')
		rendered++
	}

	return b.String()
}

func (l *List) renderItem(item ListItem, selected bool) string {
	style := ui.NormalItemStyle
	prefix := "  "
	if selected && l.focused {
		style = ui.SelectedItemStyle
		prefix = "> "
	}

	// Format: "> ID  Subtitle  Title"
	content := prefix + item.ID
	if item.Subtitle != "" {
		content += "  " + item.Subtitle
	}
	content += "  " + item.Title

	// Truncate to width
	maxW := l.width - 2 // account for borders
	if maxW > 0 && lipgloss.Width(content) > maxW {
		content = content[:maxW]
	}

	return style.Render(content)
}

func (l *List) visibleCount() int {
	if l.height <= 0 {
		return 0
	}
	return l.height
}

func (l *List) clampOffset() {
	visible := l.visibleCount()
	if visible <= 0 {
		return
	}
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+visible {
		l.offset = l.cursor - visible + 1
	}
}

func centerText(text string, w, h int) string {
	if h <= 0 || w <= 0 {
		return text
	}
	padTop := h / 2
	padLeft := (w - len(text)) / 2
	if padLeft < 0 {
		padLeft = 0
	}
	var b strings.Builder
	for range padTop {
		b.WriteByte('\n')
	}
	b.WriteString(strings.Repeat(" ", padLeft) + text)
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
