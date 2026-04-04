package settings

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"

	"gitlab.com/patopest/mdns-discovery/app/common"
	"gitlab.com/patopest/mdns-discovery/network"
)

// Styles
type Styles struct {
	Base lg.Style

	Title lg.Style

	NormalTitle lg.Style
	NormalDesc  lg.Style

	SelectedTitle lg.Style
	SelectedDesc  lg.Style

	NormalItem   lg.Style
	SelectedItem lg.Style
}

func NewStyles() (s Styles) {
	var c = &common.DefaultStyles.Color

	s.Base = lg.NewStyle().
		Padding(0, 2)

	s.Title = lg.NewStyle().
		Foreground(c.Mid)

	s.NormalTitle = lg.NewStyle().
		Foreground(c.Text).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalTitle.
		Foreground(lg.Darken(c.Text, 0.5))

	s.SelectedTitle = s.NormalTitle.
		Foreground(c.MidLow)

	s.SelectedDesc = s.NormalDesc.
		Foreground(lg.Darken(c.MidLow, 0.30))

	s.NormalItem = lg.NewStyle().
		Padding(0, 0, 0, 2)

	s.SelectedItem = s.NormalItem.
		Border(lg.NormalBorder(), false, false, false, true).
		BorderForeground(lg.Darken(c.MidLow, 0.30)).
		Padding(0, 0, 0, 1)

	return s
}

// Item represents an item in the list
type Item struct {
	iface *network.Interface
}

// Implements list.Item interface
func (i Item) FilterValue() string { return i.iface.Name }
func (i Item) Title() string       { return i.iface.Name }
func (i Item) Description() string {
	macAddr := i.iface.HardwareAddr.String()
	if macAddr == "" {
		macAddr = "no MAC address"
	}

	var ipAddr string
	if i.iface.IPv4 != nil {
		ipAddr = i.iface.IPv4.String()
	} else {
		ipAddr = "no IPv4 address"
	}

	return fmt.Sprintf("%s | %s", macAddr, ipAddr)
}

// Delegate customizes how interface items are rendered
type Delegate struct {
	Styles    Styles
	IsEnabled func(string) bool
}

func NewDelegate() Delegate {
	return Delegate{
		Styles: NewStyles(),
	}
}

// Implements list.ItemDelegate interface
func (d Delegate) Height() int                               { return 2 }
func (d Delegate) Spacing() int                              { return 1 }
func (d Delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d Delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {

	i, ok := item.(Item)
	if !ok {
		return
	}
	var s = &d.Styles

	// Check enabled state from Discovery via callback
	isEnabled := d.IsEnabled != nil && d.IsEnabled(i.iface.Name)

	var checkbox string
	if isEnabled {
		checkbox = "[✓]"
	} else {
		checkbox = "[ ]"
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		checkbox = s.SelectedTitle.UnsetPadding().Render(checkbox)
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
		if isEnabled {
			checkbox = s.NormalTitle.UnsetPadding().Render(checkbox)
		} else {
			checkbox = s.NormalDesc.UnsetPadding().Render(checkbox)
		}
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}

	view := lg.JoinHorizontal(
		lg.Left,
		checkbox,
		title+"\n"+desc,
	)

	if index == m.Index() {
		view = s.SelectedItem.Render(view)
	} else {
		view = s.NormalItem.Render(view)
	}

	fmt.Fprintf(w, "%s", s.Base.Render(view))
}

// InterfaceSelector manages the interface list
type Model struct {
	discovery *network.Discovery

	list list.Model

	Keys   keyMap
	Styles Styles
}

// New creates a new interface selector
func New(discovery *network.Discovery) *Model {
	items := []list.Item{}

	// Get all available interfaces
	allInterfaces := network.GetInterfaces()

	for _, iface := range allInterfaces {
		items = append(items, Item{iface: iface})
	}

	styles := NewStyles()

	delegate := NewDelegate()
	delegate.Styles = styles
	delegate.IsEnabled = func(name string) bool {
		return discovery.IsInterfaceEnabled(name)
	}

	l := list.New(items, delegate, 0, 0)
	l.Title = "Select network interfaces"
	l.KeyMap = SettingsKeyMap.KeyMap
	l.Styles.Title = styles.Title
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	return &Model{
		discovery: discovery,
		list:      l,
		Keys:      SettingsKeyMap,
		Styles:    styles,
	}
}

// ToggleInterfaceMsg is sent when an interface is toggled
type ToggleInterfaceMsg struct {
	Iface   *network.Interface
	Enabled bool
}

// ToggleInterface enables/disables the currently selected interface
func (m *Model) ToggleInterface() tea.Cmd {
	item, ok := m.list.SelectedItem().(Item)
	if !ok {
		return nil
	}

	enabled := m.discovery.IsInterfaceEnabled(item.iface.Name)

	return func() tea.Msg {
		return ToggleInterfaceMsg{
			Iface:   item.iface,
			Enabled: !enabled,
		}
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.Keys.Select):
			cmd = m.ToggleInterface()
			cmds = append(cmds, cmd)
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m *Model) Refresh() {
	items := []list.Item{}
	allInterfaces := network.GetInterfaces()

	for _, iface := range allInterfaces {
		items = append(items, Item{iface: iface})
	}

	m.list.SetItems(items)
}

func (m *Model) View() string {
	return m.list.View()
}
