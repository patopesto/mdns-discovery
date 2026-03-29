package table

import (
	"fmt"
	"net"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"charm.land/bubbles/v2/key"

	"gitlab.com/patopest/mdns-discovery/app/table/table"
	"gitlab.com/patopest/mdns-discovery/network"
)

const (
	SortedNone int = iota
	SortedAsc
	SortedDesc
)

// Model wraps the table Model with additional logic for sorting and data transformation
type Model struct {
	model           table.Model
	columns         []table.Column
	sortedColumnKey string
	sortedDirection int
	Keys            KeyMap
}

// NewModel creates a new table wrapper with predefined columns
func New() Model {
	columns := []table.Column{
		table.NewFlexColumn("name", "Name", 20).WithFiltered(true),
		table.NewFlexColumn("service", "Service", 18).WithFiltered(true),
		table.NewFlexColumn("domain", "Domain", 6).WithFiltered(true),
		table.NewFlexColumn("hostname", "Hostname", 18).WithFiltered(true),
		table.NewColumn("ip", "IP", 15).WithFiltered(true).WithSortFunc(SortIPs),
		table.NewColumn("port", "Port", 6).WithFiltered(true),
		table.NewFlexColumn("info", "Info", 20).WithFiltered(true),
	}

	table := table.New(columns).Focused(true).Filtered(true)
	table.Keys = TableKeyMap.KeyMap

	return Model{
		model:           table,
		columns:         columns,
		sortedColumnKey: "",
		sortedDirection: SortedNone,
		Keys:            TableKeyMap,
	}
}

// WithStyles sets the styles on the underlying table model
func (m Model) WithStyles(baseStyle, headerStyle, highlightStyle lg.Style) Model {
	m.model = m.model.WithBaseStyle(baseStyle).
		HeaderStyle(headerStyle).
		HighlightStyle(highlightStyle)
	return m
}

// UpdateColumns updates the column definitions
func (m *Model) UpdateColumns(columns []table.Column) {
	m.columns = columns
	m.model = m.model.WithColumns(columns)
}

// SetRows sets the table rows from network service entries
func (m *Model) SetRows(entries []network.ServiceEntry) {
	rows := m.generateRowsFromData(entries)
	m.model = m.model.WithRows(rows)
	m.applySort()
}

// generateRowsFromData converts network entries to table rows
func (m *Model) generateRowsFromData(data []network.ServiceEntry) []table.Row {
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

// NextSort advances the sort direction and applies sorting
func (m *Model) NextSort(columnKey string) {
	if m.sortedColumnKey == columnKey {
		m.sortedDirection++
		if m.sortedDirection > SortedDesc {
			m.sortedDirection = SortedNone
		}
	} else {
		m.sortedColumnKey = columnKey
		m.sortedDirection = SortedAsc
	}
	m.applySort()
}

// applySort applies the current sort to the table
func (m *Model) applySort() {
	if m.sortedDirection == SortedAsc {
		m.model = m.model.SortByAsc(m.sortedColumnKey)
	} else if m.sortedDirection == SortedDesc {
		m.model = m.model.SortByDesc(m.sortedColumnKey)
	} else {
		m.model = m.model.SortByAsc("")
	}

	// Update column headers with sort indicators
	newColumns := slices.Clone(m.columns)
	for idx, column := range m.columns {
		if column.Key() == m.sortedColumnKey {
			title := column.Title()
			if m.sortedDirection == SortedAsc {
				title = fmt.Sprintf("%s ▼", title)
			} else if m.sortedDirection == SortedDesc {
				title = fmt.Sprintf("%s ▲", title)
			}

			newColumns[idx] = column.WithTitle(title)
			break
		}
	}
	m.model = m.model.WithColumns(newColumns)
}

// GetSortedColumn returns the currently sorted column key
func (m *Model) GetSortedColumn() string {
	return m.sortedColumnKey
}

// GetSortedDirection returns the current sort direction
func (m *Model) GetSortedDirection() int {
	return m.sortedDirection
}

// IsFilterInputFocused returns whether the filter input is focused
func (m *Model) IsFilterInputFocused() bool {
	return m.model.IsFilterInputFocused()
}

// View renders the table
func (m *Model) View() string {
	return m.model.View()
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.Keys.SortName):
			m.NextSort("name")
		case key.Matches(msg, m.Keys.SortService):
			m.NextSort("service")
		case key.Matches(msg, m.Keys.SortDomain):
			m.NextSort("domain")
		case key.Matches(msg, m.Keys.SortHostname):
			m.NextSort("hostname")
		case key.Matches(msg, m.Keys.SortIp):
			m.NextSort("ip")
		case key.Matches(msg, m.Keys.SortPort):
			m.NextSort("port")
		}
	}

	return cmd
}

// SetSize updates the table size
func (m *Model) SetSize(width, height int) {
	m.model = m.model.WithTargetWidth(width).WithMinimumHeight(height)
}

// SetMinimumHeight sets the minimum height (for compatibility)
func (m *Model) SetMinimumHeight(height int) {
	m.model = m.model.WithMinimumHeight(height)
}

func SortIPs(a, b interface{}) int {
	ipA := a.(net.IP)
	ipB := b.(net.IP)

	for i := 0; i < len(ipA) && i < len(ipB); i++ {
		if ipA[i] > ipB[i] {
			return 1
		} else if ipA[i] < ipB[i] {
			return -1
		}
	}
	return 0
}

// unescapeString handles escaped characters in mDNS service names
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
