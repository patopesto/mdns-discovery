package table

import (
	"github.com/charmbracelet/bubbles/key"

	"gitlab.com/patopest/mdns-discovery/app/keys"
	"gitlab.com/patopest/mdns-discovery/app/table/table"
)

type KeyMap struct {
	table.KeyMap

	SortName     key.Binding
	SortService  key.Binding
	SortDomain   key.Binding
	SortHostname key.Binding
	SortIp       key.Binding
	SortPort     key.Binding

	Sort key.Binding // fake key only for description purposes (in help)
}

// Implements help.KeyMap interface
func (m Model) ShortHelp() []key.Binding {
	keys := []key.Binding{m.Keys.Sort}
	if m.model.IsFilterInputFocused() {
		keys = append(keys, m.Keys.FilterBlur)
	} else if m.model.IsFiltered() {
		keys = append(keys, m.Keys.Filter, m.Keys.FilterClear)
	} else {
		keys = append(keys, m.Keys.Filter)
	}
	return keys
}

// Implements help.KeyMap interface
func (m Model) FullHelp() [][]key.Binding {
	keys := [][]key.Binding{
		{m.Keys.Up, m.Keys.Down},              // first column
		{m.Keys.Left, m.Keys.Right},           // second column
		{m.Keys.SortName, m.Keys.SortService}, // ...
		{m.Keys.SortDomain, m.Keys.SortHostname},
		{m.Keys.SortIp, m.Keys.SortPort},
	}
	if m.model.IsFilterInputFocused() {
		keys = append(keys, []key.Binding{m.Keys.FilterBlur})
	} else if m.model.IsFiltered() {
		keys = append(keys, []key.Binding{m.Keys.Filter, m.Keys.FilterClear})
	} else {
		keys = append(keys, []key.Binding{m.Keys.Filter})
	}
	return keys
}

var TableKeyMap = KeyMap{
	KeyMap: table.KeyMap{
		Up:    keys.DefaultKeyMap.Up,
		Down:  keys.DefaultKeyMap.Down,
		Left:  keys.DefaultKeyMap.Left,
		Right: keys.DefaultKeyMap.Right,

		Filter: keys.DefaultKeyMap.Filter,
		FilterBlur: key.NewBinding(
			key.WithKeys("esc", "enter"),
			key.WithHelp("enter/esc", "unfocus"),
		),
		FilterClear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
	},

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
}
