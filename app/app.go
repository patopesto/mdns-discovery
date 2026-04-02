package app

import (
	"reflect"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"

	"gitlab.com/patopest/mdns-discovery/app/common"
	"gitlab.com/patopest/mdns-discovery/app/settings"
	"gitlab.com/patopest/mdns-discovery/app/table"
	"gitlab.com/patopest/mdns-discovery/network"
)

const APP_TITLE string = "mDNS Discovery"

type App struct {
	// data
	discovery    network.Discovery
	data         []network.ServiceEntry
	entriesCh    chan network.ServiceEntry
	showSettings bool

	// table component
	table    table.Model
	settings *settings.Model
	spinner  spinner.Model
	help     help.Model

	// dimensions
	totalWidth  int
	totalHeight int

	keys   common.KeyMap
	styles common.Styles
}

func NewApp(ifaces []string, domains []string) *App {
	table := table.New()

	help := help.New()
	help.ShortSeparator = "  •  "
	help.Styles.ShortKey = common.DefaultStyles.Footer.Help
	help.Styles.FullKey = common.DefaultStyles.Footer.Help

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = common.DefaultStyles.Header.Spinner

	// Create the entries channel
	entriesCh := make(chan network.ServiceEntry, 30)
	discovery := network.InitDiscovery(ifaces, domains, entriesCh)
	settings := settings.New(&discovery)

	app := &App{
		discovery:    discovery,
		entriesCh:    entriesCh,
		showSettings: false,
		table:        table,
		settings:     settings,
		spinner:      spin,
		help:         help,
		keys:         common.DefaultKeyMap,
		styles:       common.DefaultStyles,
	}

	return app
}

type EntryMsg network.ServiceEntry

func (m *App) listenForEntries() tea.Cmd {
	return func() tea.Msg {
		entry := <-m.entriesCh
		return EntryMsg(entry)
	}
}

func (m *App) InjectFakeData(entries []network.ServiceEntry) {
	for _, entry := range entries {
		m.data = append(m.data, entry)
		m.table.SetRows(m.data)
	}
}

// Implement tea.Model interface
func (m *App) Init() tea.Cmd {
	return tea.Batch(m.listenForEntries(), m.spinner.Tick)
}

// Implement tea.Model interface
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case EntryMsg:
		entry := network.ServiceEntry(msg)
		found := false
		for _, existing := range m.data {
			if reflect.DeepEqual(existing, entry) {
				found = true
				break
			}
		}
		if !found {
			m.data = append(m.data, entry)
			m.table.SetRows(m.data)
		}
		// Listen for the next entry
		cmds = append(cmds, m.listenForEntries())

	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.totalHeight = msg.Height
		m.help.SetWidth(msg.Width)

	case tea.KeyPressMsg:
		// Special cases
		if m.table.IsFilterInputFocused() {
			cmd = m.table.Update(msg)
			return m, cmd
		}

		// Handle top-level keys
		switch {
		case key.Matches(msg, m.keys.Settings):
			m.showSettings = !m.showSettings
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			cmds = append(cmds, tea.Quit)
		}
		// Component specific keys
		if m.showSettings {
			switch {
			case key.Matches(msg, m.settings.Keys.Close):
				m.showSettings = false
			}
		}

	case settings.ToggleInterfaceMsg:
		if msg.Enabled {
			// Find the interface by name
			allInterfaces := network.GetInterfaces()
			for _, iface := range allInterfaces {
				if iface.Name == msg.IfaceName {
					m.discovery.EnableInterface(iface)
					break
				}
			}
		} else {
			m.discovery.DisableInterface(msg.IfaceName)
		}
	}

	// Update components
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	if m.showSettings {
		cmd = m.settings.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Implement tea.Model interface
func (m *App) View() tea.View {
	var s = &m.styles

	header := m.viewHeader()
	footer := m.viewFooter()

	// Compute height of all elements to send to component
	headerHeight := lg.Height(header)
	footerHeight := lg.Height(footer)
	innerHeight := m.totalHeight - headerHeight - footerHeight

	var mainView string
	if m.showSettings {
		m.settings.SetSize(m.totalWidth, innerHeight)
		mainView = s.Settings.Base.Render(m.settings.View())
		mainView = lg.Place(m.totalWidth, innerHeight, lg.Center, lg.Center, mainView)
	} else {
		m.table.SetSize(m.totalWidth, innerHeight)
		mainView = m.table.View()
	}

	content := lg.JoinVertical(
		lg.Left,
		header,
		mainView,
		footer,
	)

	view := tea.NewView(s.Base.Render(content))
	view.AltScreen = true
	view.WindowTitle = APP_TITLE
	return view
}

func (m *App) viewHeader() string {
	var s = &m.styles

	title := s.Header.Title.Render(APP_TITLE)
	spinner := m.spinner.View()

	itfs := strings.Builder{}
	itfs.WriteString("interfaces ")

	gradient := lg.Blend1D(len(m.discovery.Interfaces), s.Color.Top, s.Color.Bottom)
	for i, itf := range m.discovery.Interfaces {
		s := s.Header.Interface.Foreground(gradient[i]).Render(itf.Name)
		itfs.WriteString(s)
	}
	interfaces := s.Header.Interfaces.Render(itfs.String())

	spacerWidth := m.totalWidth - lg.Width(spinner) - lg.Width(title) - lg.Width(interfaces) - s.Header.Base.GetHorizontalPadding()

	header := lg.JoinHorizontal(
		lg.Center,
		spinner,
		title,
		lg.NewStyle().Width(spacerWidth).Render(""),
		interfaces,
	)

	return s.Header.Base.Render(header)
}

func (m *App) viewFooter() string {
	var s = &m.styles

	footer := m.help.View(m)

	return s.Footer.Base.Render(footer)
}

// Implements help.KeyMap interface
func (m App) ShortHelp() []key.Binding {
	keys := []key.Binding{m.keys.Help}
	if m.showSettings {
		keys = append(keys, m.settings.ShortHelp()...)
	} else {
		keys = append(keys, m.table.ShortHelp()...)
		keys = append(keys, m.keys.Settings)
	}
	keys = append(keys, m.keys.Quit)
	return keys
}

// Implements help.KeyMap interface
func (m App) FullHelp() [][]key.Binding {
	var keys [][]key.Binding
	if m.showSettings {
		keys = append(keys, m.settings.FullHelp()...)
	} else {
		keys = append(keys, m.table.FullHelp()...)
		keys = append(keys, []key.Binding{m.keys.Settings})
	}
	keys = append(keys, []key.Binding{m.keys.Help, m.keys.Quit})
	return keys
}
