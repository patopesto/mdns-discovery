package settings

import (
	"github.com/charmbracelet/bubbles/key"

	"gitlab.com/patopest/mdns-discovery/app/keys"
)

type keyMap struct {
	Up   key.Binding
	Down key.Binding

	Select key.Binding
	Close      key.Binding

	Help key.Binding
	Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Select, k.Close, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down}, // first column
		{k.Select},     // second column
		{k.Close},      // ...
		{k.Help, k.Quit},
	}
}

var SettingsKeyMap = keyMap{
	Up:   keys.DefaultKeyMap.Up,
	Down: keys.DefaultKeyMap.Down,

	Select: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("<space>/enter", "toggle"),
	),
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),

	Help: keys.DefaultKeyMap.Help,
	Quit: keys.DefaultKeyMap.Quit,
}
