package app

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"

	"gitlab.com/patopest/mdns-discovery/app/settings"
	"gitlab.com/patopest/mdns-discovery/app/keys"
	"gitlab.com/patopest/mdns-discovery/network"
)

const APP_TITLE string = "mDNS Discovery"

const (
	SortedNone int = iota
	SortedAsc
	SortedDesc
)

type App struct {
	// data
	discovery network.Discovery
	data      []network.ServiceEntry
	entriesCh chan network.ServiceEntry

	// bubble tea components
	table   table.Model
	columns []table.Column

	keys    keys.KeyMap
	help    help.Model
	spinner spinner.Model

	// Window dimensions
	totalWidth  int
	totalHeight int

	// Sorting
	sortedColumnKey string
	sortedDirection int

	// Interface selector
	showSettings    bool
	settings *settings.Model
}

func NewApp(ifaces []string, domains []string) *App {
	columns := []table.Column{
		table.NewFlexColumn("name", "Name", 20).WithFiltered(true),
		table.NewFlexColumn("service", "Service", 18).WithFiltered(true),
		table.NewFlexColumn("domain", "Domain", 6).WithFiltered(true),
		table.NewFlexColumn("hostname", "Hostname", 18).WithFiltered(true),
		table.NewColumn("ip", "IP", 15).WithFiltered(true),
		table.NewColumn("port", "Port", 6).WithFiltered(true),
		table.NewFlexColumn("info", "Info", 20).WithFiltered(true),
	}

	table := table.New(columns).
		Focused(true).
		Filtered(true).
		WithBaseStyle(tableBaseStyle).
		HeaderStyle(tableHeaderStyle).
		HighlightStyle(tableHighlightedRowStyle)

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
		discovery:      discovery,
		table:          table,
		columns:        columns,
		help:           help,
		keys:           keys.DefaultKeyMap,
		spinner:        s,
		entriesCh:      entriesCh,
		showSettings: false,
		settings: settings,
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
		m.table = m.table.WithRows(generateRowsFromData(m.data))
	}
}

// Implement tea.Model interface
func (m *App) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle(APP_TITLE), m.listenForEntries(), m.spinner.Tick)
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
			m.table = m.table.WithRows(generateRowsFromData(m.data))
		}
		// Listen for the next entry
		cmds = append(cmds, m.listenForEntries())

	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.totalHeight = msg.Height
		m.table = m.table.WithTargetWidth(msg.Width)
		m.settings.SetSize(msg.Width, msg.Height-6)
		m.help.Width = msg.Width

	case tea.KeyMsg:
		// Special cases
		if m.table.GetIsFilterInputFocused() {
			m.table, cmd = m.table.Update(msg)
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
		} else {
			switch {
			case key.Matches(msg, m.keys.SortName):
				m.sortedColumnKey = "name"
				m.sortedDirection += 1
				m.sort()
			case key.Matches(msg, m.keys.SortService):
				m.sortedColumnKey = "service"
				m.sortedDirection += 1
				m.sort()
			case key.Matches(msg, m.keys.SortDomain):
				m.sortedColumnKey = "domain"
				m.sortedDirection += 1
				m.sort()
			case key.Matches(msg, m.keys.SortHostname):
				m.sortedColumnKey = "hostname"
				m.sortedDirection += 1
				m.sort()
			case key.Matches(msg, m.keys.SortIp):
				m.sortedColumnKey = "ip"
				m.sortedDirection += 1
				m.sort()
			case key.Matches(msg, m.keys.SortPort):
				m.sortedColumnKey = "port"
				m.sortedDirection += 1
				m.sort()
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
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Fix for escaped characters or ascii binary characters as "\123"
// https://github.com/miekg/dns/issues/1607
func unescapeString(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			next := s[i+1]
			// Check if next char is a digit (for decimal byte values)
			if next >= '0' && next <= '9' {
				start := i + 1
				end := start
				for end < len(s) && s[end] >= '0' && s[end] <= '9' {
					end++
				}
				// Parse decimal value and convert to byte
				val := 0
				for j := start; j < end; j++ {
					val = val*10 + int(s[j]-'0')
				}
				buf.WriteByte(byte(val))
				i = end - 1
			} else {
				// Any other escaped character - output it literally
				buf.WriteByte(next)
				i++
			}
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String()
}

func generateRowsFromData(data []network.ServiceEntry) []table.Row {
	rows := []table.Row{}

	for _, entry := range data {
		name := strings.Split(entry.Name, ".")
		row := table.NewRow(table.RowData{
			"name":     unescapeString(name[0]),
			"service":  strings.Join(name[1:len(name)-2], "."),
			"domain":   name[len(name)-2],
			"hostname": entry.Host,
			"ip":       entry.AddrV4,
			"port":     entry.Port,
			"info":     unescapeString(entry.Info),
		})

		rows = append(rows, row)
	}

	return rows
}

func (m *App) sort() {
	if m.sortedDirection == SortedAsc {
		m.table = m.table.SortByAsc(m.sortedColumnKey)
	} else if m.sortedDirection == SortedDesc {
		m.table = m.table.SortByDesc(m.sortedColumnKey)
	} else {
		m.sortedDirection = SortedNone
		m.table = m.table.SortByAsc("") // trick to reset sorting
	}

	// Update column header
	new_columns := slices.Clone(m.columns)
	for idx, column := range m.columns {
		if column.Key() == m.sortedColumnKey {
			title := column.Title()
			if m.sortedDirection == SortedAsc {
				title = fmt.Sprintf("%s ▼", title)
			} else if m.sortedDirection == SortedDesc {
				title = fmt.Sprintf("%s ▲", title)
			}

			var new_column table.Column
			if column.IsFlex() {
				new_column = table.NewFlexColumn(column.Key(), title, column.FlexFactor())
			} else {
				new_column = table.NewColumn(column.Key(), title, column.Width())
			}
			new_columns[idx] = new_column
			break
		}
	}
	m.table = m.table.WithColumns(new_columns)
}
