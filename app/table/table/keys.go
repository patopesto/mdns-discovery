package table

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	Filter      key.Binding
	FilterBlur  key.Binding
	FilterClear key.Binding
}

// Implements help.KeyMap interface
func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.Keys.Up, m.Keys.Down, m.Keys.Filter}
}

// Implements help.KeyMap interface
func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Keys.Up, m.Keys.Down},    // first column
		{m.Keys.Left, m.Keys.Right}, // second column
		{m.Keys.Filter},             // ...
	}
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("left/h", "scroll left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("right/l", "scroll right"),
	),

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
}
