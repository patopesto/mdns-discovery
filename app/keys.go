package app

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	SortName     key.Binding
	SortService  key.Binding
	SortDomain   key.Binding
	SortHostname key.Binding
	SortIp       key.Binding
	SortPort     key.Binding

	Sort   key.Binding // fake key only for description purposes (in help)
	Filter key.Binding

	Help key.Binding
	Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Sort, k.Filter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},              // first column
		{k.Left, k.Right},           // second column
		{k.SortName, k.SortService}, // ...
		{k.SortDomain, k.SortHostname},
		{k.SortIp, k.SortPort},
		{k.Filter},
		{k.Help, k.Quit},
	}
}

var DefaultKeyMap = keyMap{
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
	Sort: key.NewBinding(
		key.WithKeys(""),
		key.WithHelp("[1-6]", "sort"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
