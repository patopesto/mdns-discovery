package common

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	// navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	// sorting (table)
	Sort         key.Binding // fake key only for description purposes (in help)
	SortName     key.Binding
	SortService  key.Binding
	SortDomain   key.Binding
	SortHostname key.Binding
	SortIp       key.Binding
	SortPort     key.Binding

	// fitlering (table)
	Filter      key.Binding
	FilterBlur  key.Binding
	FilterClear key.Binding

	// modes / settings / panes
	Settings key.Binding
	Select   key.Binding
	Close    key.Binding

	// other
	Help key.Binding
	Quit key.Binding
}

var DefaultKeyMap = KeyMap{
	// navigation
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),

	// sorting (table)
	Sort: key.NewBinding(
		key.WithKeys(""),
		key.WithHelp("[1-6]", "sort"),
	),
	SortName: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "sort by name"),
	),
	SortService: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "sort by service"),
	),
	SortDomain: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "sort by domain"),
	),
	SortHostname: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "sort by hostname"),
	),
	SortIp: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "sort by ip"),
	),
	SortPort: key.NewBinding(
		key.WithKeys("6"),
		key.WithHelp("6", "sort by port "),
	),

	// fitlering (table)
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	FilterBlur: key.NewBinding(
		key.WithKeys("esc", "enter"),
		key.WithHelp("enter/esc", "unfocus"),
	),
	FilterClear: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),

	// modes / settings / panes
	Settings: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "settings"),
	),
	Select: key.NewBinding(
		key.WithKeys("space", "enter"),
		key.WithHelp("<space>/enter", "toggle"),
	),
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),

	// other
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
