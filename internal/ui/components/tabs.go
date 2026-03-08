package components

import (
	"strings"

	"github.com/da3az/lazado/internal/ui"
)

// TabBar renders the panel selection tabs.
type TabBar struct {
	tabs   []string
	active int
	width  int
}

// NewTabBar creates a tab bar with the given tab labels.
func NewTabBar(tabs []string) *TabBar {
	return &TabBar{tabs: tabs}
}

// SetActive sets which tab is highlighted.
func (t *TabBar) SetActive(idx int) {
	if idx >= 0 && idx < len(t.tabs) {
		t.active = idx
	}
}

// SetWidth sets the available width.
func (t *TabBar) SetWidth(w int) {
	t.width = w
}

// View renders the tab bar.
func (t *TabBar) View() string {
	var tabs []string
	for i, tab := range t.tabs {
		label := "[" + string(rune('1'+i)) + "] " + tab
		if i == t.active {
			tabs = append(tabs, ui.ActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, ui.InactiveTabStyle.Render(label))
		}
	}
	row := strings.Join(tabs, "")
	return ui.TabBarStyle.Width(t.width).Render(row)
}
