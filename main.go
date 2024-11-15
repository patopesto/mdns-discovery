package main

import (
	"fmt"
	// "log"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

/* ----- BubbleTea app ----- */
const UPDATE_INTERVAL = 1 // in seconds

const (
	SortedNone int = iota
	SortedAsc
	SortedDesc
)

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	SortName     key.Binding
	SortService  key.Binding
	SortDomain   key.Binding
	SortHostname key.Binding
	SortIp       key.Binding
	SortPort     key.Binding

	Sort   key.Binding // fake key only for description purposes
	Filter key.Binding

	Help key.Binding
	Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Sort, k.Filter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},              // first column
		{k.Left, k.Right},           // second column
		{k.SortName, k.SortService}, // ...
		{k.SortDomain, k.SortHostname},
		{k.SortIp, k.SortPort},
		{k.Filter},
		{k.Help, k.Quit},
	}
}

type Model struct {
	table   table.Model
	columns []table.Column
	data    []ServiceEntry

	keys keyMap // Our own keymap to use the help interface
	help help.Model

	// Window dimensions
	totalWidth  int
	totalHeight int

	// Sorting
	sortedColumnKey string
	sortedDirection int
}

func NewModel() Model {
	keys := keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
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
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	columns := []table.Column{
		table.NewFlexColumn("name", "Name", 8).WithFiltered(true),
		table.NewFlexColumn("service", "Service", 6).WithFiltered(true),
		table.NewFlexColumn("domain", "Domain", 1).WithFiltered(true),
		table.NewFlexColumn("hostname", "Hostname", 8).WithFiltered(true),
		table.NewFlexColumn("ip", "IP", 3).WithFiltered(true),
		table.NewFlexColumn("port", "Port", 1).WithFiltered(true),
		table.NewFlexColumn("info", "Info", 8),
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

	return Model{
		table:   table,
		columns: columns,
		help:    help,
		keys:    keys,
	}
}

type EntryMsg []ServiceEntry

func tickEvery() tea.Cmd {
	return tea.Every(UPDATE_INTERVAL*time.Second, func(t time.Time) tea.Msg {

		entries := make([]ServiceEntry, 0)
		for _, discovery := range discoveries {
			entries = append(entries, discovery.Entries...)
		}

		return EntryMsg(entries)
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("mDNS Discovery"), tickEvery())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case EntryMsg:
		data := m.data

		for _, entry := range msg {
			found := false
			for _, existing := range data {
				if reflect.DeepEqual(existing, entry) {
					found = true
					break
				}
			}
			if !found {
				data = append(data, entry)
			}
		}
		m.data = data
		m.table = m.table.WithRows(generateRowsFromData(m.data))
		cmds = append(cmds, tickEvery())

	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.totalHeight = msg.Height
		m.table = m.table.WithTargetWidth(msg.Width)
		m.help.Width = msg.Width

	case tea.KeyMsg:
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
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			cmds = append(cmds, tea.Quit)
		}
	}

	return m, tea.Batch(cmds...)
}

func generateRowsFromData(data []ServiceEntry) []table.Row {
	rows := []table.Row{}

	for _, entry := range data {
		name := strings.Split(entry.Name, ".")
		row := table.NewRow(table.RowData{
			"name":     fmt.Sprintf("%s", name[0]),
			"service":  strings.Join(name[1:len(name)-2], "."),
			"domain":   name[len(name)-2],
			"hostname": entry.Host,
			"ip":       entry.AddrV4,
			"port":     entry.Port,
			"info":     entry.Info,
		})

		rows = append(rows, row)
	}

	return rows
}

func (m *Model) sort() {
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

/* ----- Styles ----- */
// Table
var tableBaseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	Foreground(lipgloss.Color("252")).
	Align(lipgloss.Left)

var tableHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("203")).
	Bold(true).
	Align(lipgloss.Center)

var tableHighlightedRowStyle = lipgloss.NewStyle().
	Bold(true).
	Background(lipgloss.Color("96")).
	Foreground(lipgloss.Color("255"))

// Window
var topStyle = lipgloss.NewStyle().Padding(1, 3, 0).
	Bold(true).
	Foreground(lipgloss.Color("202"))

var tableStyle = lipgloss.NewStyle()

var helpStyle = lipgloss.NewStyle().Padding(1, 2)

// Other
var helpKeyStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#909090",
	// Dark:  "#a0a0a0",
	Dark: "205",
})

func (m Model) View() string {
	topStr := strings.Builder{}
	topStr.WriteString("mDNS Discovery\n")

	topBlock := topStyle.Render(topStr.String())
	helpBlock := helpStyle.Render(m.help.View(m.keys))

	// Compute height of all elements to send to table
	topHeight := lipgloss.Height(topBlock)
	helpHeight := lipgloss.Height(helpBlock)
	tableHeight := m.totalHeight - topHeight - helpHeight
	m.table = m.table.WithMinimumHeight(tableHeight)

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		topBlock,
		tableStyle.Render(m.table.View()),
		helpBlock,
	)
	return lipgloss.NewStyle().Render(view)
}

/* ----- Entrypoint ----- */

var ifaces = flag.StringSliceP("interface", "i", nil, "Use specified interface(s). ex: '-i eth0,wlan0' (default: all available interfaces)")
var doms = flag.StringSliceP("domain", "d", []string{DEFAULT_DOMAIN}, "Domain(s) to use, usually '.local' \t\t!!! Do no t change unless you know what you're doing !!!")
var info = flag.BoolP("version", "v", false, "Print version info")
var usage = flag.BoolP("help", "h", false, "Print this help message")
var fake = flag.Bool("fake", false, "Use fake data instead")

func main() {
	flag.CommandLine.SortFlags = false
    flag.CommandLine.MarkHidden("fake")
	flag.Parse()

	if *info {
		PrintInfo()
		os.Exit(0)
	}
	if *usage {
		flag.Usage()
		os.Exit(0)
	}

    m := NewModel()

    if *fake { // fake data for demo purposes
        InitDiscovery(nil, []string{"test.com"})
        m.data = fakeData
        m.table = m.table.WithRows(generateRowsFromData(fakeData))
    } else {
        InitDiscovery(*ifaces, *doms)
    }

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
