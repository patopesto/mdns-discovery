package settings

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"

	"gitlab.com/patopest/mdns-discovery/app/common"
)

type keyMap struct {
	list.KeyMap

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
	KeyMap: list.KeyMap{
		CursorUp:   common.DefaultKeyMap.Up,
		CursorDown: common.DefaultKeyMap.Down,
	},

	Up:   common.DefaultKeyMap.Up,
	Down: common.DefaultKeyMap.Down,

	Select: common.DefaultKeyMap.Select,
	Close:  common.DefaultKeyMap.Close,
}
