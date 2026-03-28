package settings

import (
	"fmt"
	"io"
	"net"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"

	"gitlab.com/patopest/mdns-discovery/network"
)

// Styles
type Styles struct {
	Base lg.Style

	NormalTitle lg.Style
	NormalDesc  lg.Style

	SelectedTitle lg.Style
	SelectedDesc  lg.Style

	NormalItem   lg.Style
	SelectedItem lg.Style

	EnabledItem  lg.Style
	DisabledItem lg.Style
}

func NewStyles() (s Styles) {
	s.Base = lg.NewStyle().
		Padding(0, 2)

	s.NormalTitle = lg.NewStyle().
		Foreground(lg.Color("#dddddd")).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalTitle.
		Foreground(lg.Color("#777777"))

	s.SelectedTitle = s.NormalTitle.
		Foreground(lg.Color("#EE6FF8"))

	s.SelectedDesc = s.SelectedTitle.
		Foreground(lg.Color("#AD58B4"))

	s.NormalItem = lg.NewStyle().
		Padding(0, 0, 0, 2)

	s.SelectedItem = s.NormalItem.
		Border(lg.NormalBorder(), false, false, false, true).
		BorderForeground(lg.Color("#AD58B4")).
		Padding(0, 0, 0, 1)

	s.EnabledItem = lg.NewStyle().
		Foreground(lg.Color("202"))

	s.DisabledItem = s.EnabledItem.
		Foreground(lg.Color("240"))

	return s
}

// Item represents an item in the list
type Item struct {
	iface   *net.Interface
	enabled bool
}

func (i Item) FilterValue() string { return i.iface.Name }
func (i Item) Title() string       { return i.iface.Name }
func (i Item) Description() string { return i.iface.HardwareAddr.String() }

// Delegate customizes how interface items are rendered
type Delegate struct {
	Styles Styles
}

func NewDelegate() Delegate {
	return Delegate{
		Styles: NewStyles(),
	}
}

func (d Delegate) Height() int                               { return 2 }
func (d Delegate) Spacing() int                              { return 1 }
func (d Delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d Delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {

	i, ok := item.(Item)
	if !ok {
		return
	}
	var s = &d.Styles

	var checkbox string
	if i.enabled {
		checkbox = s.EnabledItem.Render("[✓]")
	} else {
		checkbox = s.DisabledItem.Render("[ ]")
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
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

	list      list.Model
	Keys      keyMap
}

// New creates a new interface selector
func New(discovery *network.Discovery) *Model {
	items := []list.Item{}

	// Get all available interfaces
	allInterfaces := network.GetInterfaces()

	// Get currently enabled interfaces
	enabledMap := make(map[string]bool)
	for _, iface := range discovery.Interfaces {
		enabledMap[iface.Name] = true
	}

	for _, iface := range allInterfaces {
		items = append(items, Item{
			iface:   iface,
			enabled: enabledMap[iface.Name],
		})
	}

	delegate := NewDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Select network interfaces"
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
	}
}

// ToggleInterfaceMsg is sent when an interface is toggled
type ToggleInterfaceMsg struct {
	IfaceName string
	Enabled   bool
}

// ToggleInterface enables/disables the currently selected interface
func (m *Model) ToggleInterface() tea.Cmd {
	item, ok := m.list.SelectedItem().(Item)
	if !ok {
		return nil
	}

	item.enabled = !item.enabled

	// Update the item in the list
	items := m.list.Items()
	idx := m.list.Index()
	if idx < len(items) {
		items[idx] = item
		m.list.SetItems(items)
	}

	// Return a command to update the discovery
	return func() tea.Msg {
		return ToggleInterfaceMsg{
			IfaceName: item.iface.Name,
			Enabled:   item.enabled,
		}
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Up):
			m.list.CursorUp()
		case key.Matches(msg, m.Keys.Down):
			m.list.CursorDown()
		case key.Matches(msg, m.Keys.Select):
			cmd = m.ToggleInterface()
			cmds = append(cmds, cmd)
		}
	default:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m *Model) View() string {
	return m.list.View()
}

// Implements help.KeyMap interface
func (m *Model) ShortHelp() []key.Binding {
	return m.Keys.ShortHelp()
}

// Implements help.KeyMap interface
func (m *Model) FullHelp() [][]key.Binding {
	return m.Keys.FullHelp()
}
