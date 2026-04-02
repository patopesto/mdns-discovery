package table

import (
	"charm.land/bubbles/v2/key"

	"gitlab.com/patopest/mdns-discovery/app/common"
	"gitlab.com/patopest/mdns-discovery/app/table/table"
)

type KeyMap struct {
	table.KeyMap

	Sort         key.Binding // fake key only for description purposes (in help)
	SortName     key.Binding
	SortService  key.Binding
	SortDomain   key.Binding
	SortHostname key.Binding
	SortIp       key.Binding
	SortPort     key.Binding

	Select key.Binding
	Close  key.Binding
}

// Implements help.KeyMap interface
func (m Model) ShortHelp() []key.Binding {
	keys := []key.Binding{}
	if m.isViewportVisible {
		keys = append(keys, m.Keys.Select)
	} else if m.table.IsFilterInputFocused() {
		keys = append(keys, m.Keys.Sort, m.Keys.FilterBlur)
	} else if m.table.IsFiltered() {
		keys = append(keys, m.Keys.Sort, m.Keys.Filter, m.Keys.FilterClear, m.Keys.Select)
	} else {
		keys = append(keys, m.Keys.Sort, m.Keys.Filter, m.Keys.Select)
	}
	return keys
}

// Implements help.KeyMap interface
func (m Model) FullHelp() [][]key.Binding {
	keys := [][]key.Binding{
		{m.Keys.Up, m.Keys.Down},              // first column
		{m.Keys.SortName, m.Keys.SortService}, // second column
		{m.Keys.SortDomain, m.Keys.SortHostname},
		{m.Keys.SortIp, m.Keys.SortPort},
	}
	if m.isViewportVisible {
		keys = append(keys, []key.Binding{m.Keys.Select})
	} else if m.table.IsFilterInputFocused() {
		keys = append(keys, []key.Binding{m.Keys.FilterBlur})
	} else if m.table.IsFiltered() {
		keys = append(keys, []key.Binding{m.Keys.Select}, []key.Binding{m.Keys.Filter, m.Keys.FilterClear})
	} else {
		keys = append(keys, []key.Binding{m.Keys.Select}, []key.Binding{m.Keys.Filter})
	}
	return keys
}

var TableKeyMap = KeyMap{
	KeyMap: table.KeyMap{
		Up:    common.DefaultKeyMap.Up,
		Down:  common.DefaultKeyMap.Down,
		Left:  common.DefaultKeyMap.Left,
		Right: common.DefaultKeyMap.Right,

		Filter:      common.DefaultKeyMap.Filter,
		FilterBlur:  common.DefaultKeyMap.FilterBlur,
		FilterClear: common.DefaultKeyMap.FilterClear,
	},

	Sort:         common.DefaultKeyMap.Sort,
	SortName:     common.DefaultKeyMap.SortName,
	SortService:  common.DefaultKeyMap.SortService,
	SortDomain:   common.DefaultKeyMap.SortDomain,
	SortHostname: common.DefaultKeyMap.SortHostname,
	SortIp:       common.DefaultKeyMap.SortIp,
	SortPort:     common.DefaultKeyMap.SortPort,

	Select: common.DefaultKeyMap.Select,
	Close:  common.DefaultKeyMap.Close,
}
