package settings

import (
	"charm.land/bubbles/v2/key"

	"gitlab.com/patopest/mdns-discovery/app/keys"
)

type keyMap struct {
	Up   key.Binding
	Down key.Binding

	Select key.Binding
	Close  key.Binding
}

// Implements help.KeyMap interface
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{m.Keys.Select, m.Keys.Close}
}

// Implements help.KeyMap interface
func (m *Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Keys.Up, m.Keys.Down}, // first column
		{m.Keys.Select},          // second column
		{m.Keys.Close},           // ...
	}
}

var SettingsKeyMap = keyMap{
	Up:   keys.DefaultKeyMap.Up,
	Down: keys.DefaultKeyMap.Down,

	Select: key.NewBinding(
		key.WithKeys("space", "enter"),
		key.WithHelp("<space>/enter", "toggle"),
	),
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),
}
