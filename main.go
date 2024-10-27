package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
	// "strconv"
	"strings"
    "slices"
	"time"

	"github.com/hashicorp/mdns"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	// "github.com/charmbracelet/bubbles/table"
	// "github.com/charmbracelet/lipgloss/table"
	"github.com/evertras/bubble-table/table"
)

const (
	QUERY_INTERVAL  = 11
	QUERY_TIMEOUT   = 10
	UPDATE_INTERVAL = 1
	MDNS_META_QUERY = "_services._dns-sd._udp" // Used to query all peers for their services
)

type Discovery struct {
	Params    *mdns.QueryParam
	Entries   []mdns.ServiceEntry
	entriesCh chan *mdns.ServiceEntry
	timer     *time.Ticker
	stop      chan struct{}
}

var discoveries []*Discovery

func InitDiscovery() {
	discoveries = make([]*Discovery, 0)

	discovery := NewDiscovery(MDNS_META_QUERY, "en0")
	discoveries = append(discoveries, discovery)
	discovery.Start()
}

func NewDiscovery(service string, iface string) *Discovery {
	itf, _ := net.InterfaceByName(iface)
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	entries := make([]mdns.ServiceEntry, 0)
	return &Discovery{
		Entries:   entries,
		entriesCh: entriesCh,
		Params: &mdns.QueryParam{
			Service:             service,
			Domain:              "local",
			Timeout:             QUERY_TIMEOUT * time.Second,
			Entries:             entriesCh,
			Interface:           itf,
			WantUnicastResponse: true,
			DisableIPv4:         false,
			DisableIPv6:         true,
		},
	}
}

func (d *Discovery) Start() {
	d.timer = time.NewTicker(QUERY_INTERVAL * time.Second)
	d.stop = make(chan struct{})

	go d.Run()
}

func (d *Discovery) Stop() {
	close(d.stop)
}

func (d *Discovery) Run() {
	defer d.timer.Stop()

	// Running the queries at interval in it's own goroutine
	go func() {
		mdns.Query(d.Params)
		for {
			select {
			case <-d.stop:
				return
			case <-d.timer.C:
				mdns.Query(d.Params)
			}
		}
	}()

	for {
		select {
		case <-d.stop:
			return
		case <-d.entriesCh:
			for entry := range d.entriesCh {
				found := false
				for _, existing := range d.Entries {
					if reflect.DeepEqual(entry, existing) {
						found = true
						break
					}
				}
				if !found {
					d.Entries = append(d.Entries, *entry)
				}
			}
		}
	}

}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	Align(lipgloss.Left)

var headerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("213")).
	Bold(true).
	Align(lipgloss.Center)

const (
	minWidth  = 30
	minHeight = 8

	// Add a fixed margin to account for description & instructions
	fixedVerticalMargin = 4
)

const (
	SortedNone int = iota
	SortedAsc
	SortedDesc
)

type Model struct {
	table table.Model
    columns []table.Column
	data  []mdns.ServiceEntry

	// Window dimensions
	totalWidth  int
	totalHeight int

	// Table dimensions
	horizontalMargin int
	verticalMargin   int

	// Sorting
	sortedColumnKey string
	sortedDirection int
}

func NewModel() Model {
	keys := table.DefaultKeyMap()
	keys.RowDown.SetKeys("j", "down", "s")
	keys.RowUp.SetKeys("k", "up", "w")

    columns := []table.Column{
        table.NewColumn("name", "Name", 50),
        // This table uses flex columns, but it will still need a target
        // width in order to know what width it should fill.  In this example
        // the target width is set below in `recalculateTable`, which sets
        // the table to the width of the screen to demonstrate resizing
        // with flex columns.
        table.NewFlexColumn("service", "Service", 6),
        table.NewFlexColumn("domain", "Domain", 1),
        table.NewFlexColumn("hostname", "Hostname", 8),
        table.NewFlexColumn("ip", "IP", 3),
        table.NewFlexColumn("port", "Port", 1),
        table.NewFlexColumn("info", "Info", 8),
    }

	table := table.New(columns).
		// SelectableRows(true).
		Focused(true).
		WithKeyMap(keys).
		WithStaticFooter("A footer!").
		WithBaseStyle(baseStyle).
		HeaderStyle(headerStyle)

	return Model{
		table: table,
        columns: columns,
	}
}

type EntryMsg []mdns.ServiceEntry

func tickEvery() tea.Cmd {
	return tea.Every(UPDATE_INTERVAL*time.Second, func(t time.Time) tea.Msg {

		entries := make([]mdns.ServiceEntry, 0)
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
		m.recalculateTable()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			cmds = append(cmds, tea.Quit)
		case "1":
            m.sortedColumnKey = "name"
            m.sortedDirection += 1
            m.sort()
        case "2":
            m.sortedColumnKey = "service"
            m.sortedDirection += 1
            m.sort()
        case "3":
            m.sortedColumnKey = "domain"
            m.sortedDirection += 1
            m.sort()
        case "4":
            m.sortedColumnKey = "hostname"
            m.sortedDirection += 1
            m.sort()
        case "5":
            m.sortedColumnKey = "ip"
            m.sortedDirection += 1
            m.sort()
        case "6":
            m.sortedColumnKey = "port"
            m.sortedDirection += 1
            m.sort()
			// case "enter":
			// 	return m, tea.Batch(
			// 		tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			// 	)
        }
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.table.View()
}

func generateRowsFromData(data []mdns.ServiceEntry) []table.Row {
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

func (m *Model) recalculateTable() {
	m.table = m.table.WithTargetWidth(m.calculateWidth()).WithMinimumHeight(m.calculateHeight())
}

func (m Model) calculateWidth() int {
	return m.totalWidth - m.horizontalMargin
}

func (m Model) calculateHeight() int {
	return m.totalHeight - m.verticalMargin - fixedVerticalMargin
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
            } else if ( m.sortedDirection == SortedDesc) {
                title = fmt.Sprintf("%s ▲", title)
            }
            
            var new_column table.Column
            if (column.IsFlex()) {
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

func main() {
	log.SetOutput(ioutil.Discard)

	InitDiscovery()

	m := NewModel()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
