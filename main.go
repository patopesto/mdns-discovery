package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
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

	// Spinner
	spinner spinner.Model
}

func NewModel() *Model {
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

	return &Model{
		table:   table,
		columns: columns,
		help:    help,
		keys:    keys,
		spinner: s,
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

func (m *Model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("mDNS Discovery"), tickEvery(), m.spinner.Tick)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	m.spinner, cmd = m.spinner.Update(msg)
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

func generateRowsFromData(data []ServiceEntry) []table.Row {
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
// Helpers
var Style = lg.NewStyle

// Window
// - Header
var headerStyle = Style().
	Padding(1, 4, 1, 2)

var titleStyle = Style().
	Bold(true).
	Foreground(lg.Color("202"))

var spinnerStyle = Style().
	PaddingRight(2).
	Foreground(lg.Color("255"))

var interfacesStyle = Style().
	Foreground(lg.Color("#606060"))

var interfaceStyle = Style().
	Padding(0, 1).
	Foreground(lg.Color("202"))

// - Table
var tableStyle = Style()

var tableBaseStyle = Style().
	BorderStyle(lg.NormalBorder()).
	BorderForeground(lg.Color("240")).
	Foreground(lg.Color("252")).
	Align(lg.Left)

var tableHeaderStyle = Style().
	Foreground(lg.Color("203")).
	Bold(true).
	Align(lg.Center)

var tableHighlightedRowStyle = Style().
	Bold(true).
	Background(lg.Color("96")).
	Foreground(lg.Color("255"))

// - Footer
var footerStyle = Style().
	Padding(1, 2)

// Other
var helpKeyStyle = Style().
	Foreground(lg.AdaptiveColor{
		Light: "#909090",
		Dark:  "205",
	})

func (m *Model) View() string {
	title := titleStyle.Render("mDNS Discovery")
	spinner := m.spinner.View()
	itfs := strings.Builder{}
	itfs.WriteString("interfaces ")
	for _, itf := range interfaces {
		s := interfaceStyle.Render(itf.Name)
		itfs.WriteString(s)
	}
	interfaces := interfacesStyle.Render(itfs.String())

	spacerWidth := m.totalWidth - lg.Width(spinner) - lg.Width(title) - lg.Width(interfaces) - headerStyle.GetHorizontalPadding()

	header := lg.JoinHorizontal(
		lg.Center,
		spinner,
		title,
		Style().Width(spacerWidth).Render(""),
		interfaces,
	)

	headerBlock := headerStyle.Render(header)
	footerBlock := footerStyle.Render(m.help.View(m.keys))

	// Compute height of all elements to send to table
	topHeight := lg.Height(headerBlock)
	helpHeight := lg.Height(footerBlock)
	tableHeight := m.totalHeight - topHeight - helpHeight
	pageSize := tableHeight - 6 // magic offset based on current headerBlock + tableHeader rendering
	if pageSize < 1 {
		pageSize = 1
	}
	m.table = m.table.WithMinimumHeight(tableHeight).WithPageSize(pageSize)

	view := lg.JoinVertical(
		lg.Left,
		headerBlock,
		tableStyle.Render(m.table.View()),
		footerBlock,
	)
	return Style().Render(view)
}

/* ----- Entrypoint ----- */

var ifaces = flag.StringSliceP("interface", "i", nil, "Use specified interface(s). ex: '-i eth0,wlan0' (default: all available interfaces)")
var doms = flag.StringSliceP("domain", "d", []string{DEFAULT_DOMAIN}, "Domain(s) to use, usually '.local' \t\t!!! Do not change unless you know what you're doing !!!")
var info = flag.BoolP("version", "v", false, "Print version info")
var usage = flag.BoolP("help", "h", false, "Print this help message")
var debugFile = flag.Bool("debug", false, "Write logs to file")
var fake = flag.Bool("fake", false, "Use fake data instead")

func main() {
	flag.CommandLine.SortFlags = false
	flag.CommandLine.MarkHidden("debug")
	flag.CommandLine.MarkHidden("fake")
	flag.Parse()

	if *debugFile {
		f, err := tea.LogToFile("debug.log", "")
		if err != nil {
			log.Fatal("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	} else {
		log.SetOutput(io.Discard)
	}

	log.Println("Hello! Starting up...")

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
		m.data = fakeDataLong
		m.table = m.table.WithRows(generateRowsFromData(fakeDataLong))
	} else {
		InitDiscovery(*ifaces, *doms)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
