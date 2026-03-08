package ui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeys defines keybindings available across all panels.
type GlobalKeys struct {
	Quit       key.Binding
	Tab        key.Binding
	ShiftTab   key.Binding
	Panel1     key.Binding
	Panel2     key.Binding
	Panel3     key.Binding
	Panel4     key.Binding
	Help       key.Binding
	Refresh    key.Binding
	Search     key.Binding
	Back       key.Binding
}

var Keys = GlobalKeys{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next panel"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev panel"),
	),
	Panel1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "work items"),
	),
	Panel2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "pull requests"),
	),
	Panel3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "pipelines"),
	),
	Panel4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "repos"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "refresh"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

// ListKeys defines navigation keybindings for list views.
type ListKeys struct {
	Up    key.Binding
	Down  key.Binding
	Top   key.Binding
	Bottom key.Binding
	Enter key.Binding
}

var ListNav = ListKeys{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G", "bottom"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
}
