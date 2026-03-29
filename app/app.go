package app

import (
	"reflect"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"gitlab.com/patopest/mdns-discovery/app/keys"
	"gitlab.com/patopest/mdns-discovery/app/settings"
	"gitlab.com/patopest/mdns-discovery/app/table"
	"gitlab.com/patopest/mdns-discovery/network"
)

const APP_TITLE string = "mDNS Discovery"

type App struct {
	// data
	discovery network.Discovery
	data      []network.ServiceEntry
	entriesCh chan network.ServiceEntry

	// table component
	table table.Model

	keys    keys.KeyMap
	help    help.Model
	spinner spinner.Model

	// Window dimensions
	totalWidth  int
	totalHeight int

	// Interface selector
	showSettings bool
	settings     *settings.Model
}

func NewApp(ifaces []string, domains []string) *App {
	table := table.New().
		WithStyles(tableBaseStyle, tableHeaderStyle, tableHighlightedRowStyle)

	help := help.New()
	help.ShortSeparator = "  •  "
	help.Styles.ShortKey = helpKeyStyle
	help.Styles.FullKey = helpKeyStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	// Create the entries channel
	entriesCh := make(chan network.ServiceEntry, 30)
	discovery := network.InitDiscovery(ifaces, domains, entriesCh)
	settings := settings.New(&discovery)

	app := &App{
		discovery:    discovery,
		table:        table,
		help:         help,
		keys:         keys.DefaultKeyMap,
		spinner:      s,
		entriesCh:    entriesCh,
		showSettings: false,
		settings:     settings,
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
