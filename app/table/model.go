package table

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"

	"gitlab.com/patopest/mdns-discovery/app/common"
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
	table           table.Model
	columns         []table.Column
	sortedColumnKey string
	sortedDirection int

	viewport          viewport.Model
	isViewportVisible bool
	offsetX			  int
	offsetY			  int

	Keys KeyMap
}

// NewModel creates a new table wrapper with predefined columns
func New() Model {
	var styles = &common.DefaultStyles

	columns := []table.Column{
		table.NewFlexColumn("name", "Name", 20).WithFiltered(true),
		table.NewFlexColumn("service", "Service", 14).WithFiltered(true),
		table.NewFlexColumn("protocol", "Protocol", 6).WithFiltered(true),
		table.NewFlexColumn("domain", "Domain", 6).WithFiltered(true),
		table.NewFlexColumn("hostname", "Hostname", 18).WithFiltered(true),
		table.NewColumn("ip", "IP", 15).WithFiltered(true).WithSortFunc(SortIPs),
		table.NewColumn("port", "Port", 6).WithFiltered(true).WithStyle(styles.Table.RowCell.Align(lg.Right).PaddingRight(1)),
		table.NewFlexColumn("info", "Info", 20).WithFiltered(true).WithStyle(styles.Table.RowCell.UnsetPadding()),
	}

	table := table.New(columns).Focused(true).Filtered(true)
	table.Keys = TableKeyMap.KeyMap
	table.Styles.Base = styles.Table.Base
	table.Styles.Header = styles.Table.Header
	table.Styles.Row = styles.Table.Row
	table.Styles.RowCell = styles.Table.RowCell
	table.Styles.Selected = styles.Table.Selected
	table.Styles.FilterMatch = styles.Table.FilterMatch
	table.Styles.FilterInputFocused = styles.Table.FilterInputFocused
	table.Styles.FilterInputBlurred = styles.Table.FilterInputBlurred
	table.Styles.Footer = styles.Table.Footer

	viewport := viewport.New()
	viewport.Style = styles.Viewport.Base

	return Model{
		table:             table,
		columns:           columns,
		sortedColumnKey:   "",
		sortedDirection:   SortedNone,
		viewport:          viewport,
		isViewportVisible: false,
		Keys:              TableKeyMap,
		offsetX:  		   10,
		offsetY:  		   6,
	}
}

// UpdateColumns updates the column definitions
func (m *Model) UpdateColumns(columns []table.Column) {
	m.columns = columns
	m.table = m.table.WithColumns(columns)
}

// SetRows sets the table rows from network service entries
func (m *Model) SetRows(entries []network.ServiceEntry) {
	rows := m.generateRowsFromData(entries)
	m.table = m.table.WithRows(rows)
	m.applySort()
}

// generateRowsFromData converts network entries to table rows
func (m *Model) generateRowsFromData(data []network.ServiceEntry) []table.Row {
	rows := []table.Row{}

	for _, entry := range data {
		name := strings.Split(entry.Name, ".")
		row := table.NewRow(table.RowData{
			"name":     unescapeString(name[0]),
			"service":  name[1][1:],
			"protocol": name[2][1:],
			"domain":   name[3],
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
		m.table = m.table.SortByAsc(m.sortedColumnKey)
	} else if m.sortedDirection == SortedDesc {
		m.table = m.table.SortByDesc(m.sortedColumnKey)
	} else {
		m.table = m.table.SortByAsc("")
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
	m.table = m.table.WithColumns(newColumns)
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
	return m.table.IsFilterInputFocused()
}

// SetSize updates the table size
func (m *Model) SetSize(width int, height int) {
	m.table = m.table.WithTargetWidth(width).WithMinimumHeight(height)
	m.viewport.SetWidth(width - (m.offsetX * 2))
	m.viewport.SetHeight(height - (m.offsetY * 2))
}

// SetMinimumHeight sets the minimum height (for compatibility)
func (m *Model) SetMinimumHeight(height int) {
	m.table = m.table.WithMinimumHeight(height)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Special case
		if m.isViewportVisible == false {
			switch {
			case key.Matches(msg, m.Keys.Select) && !m.IsFilterInputFocused():
				m.isViewportVisible = true
				m.viewport.SetContent(m.renderSelectedRow())
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
			default:
				m.table, cmd = m.table.Update(msg)
				return cmd
			}
		} else {
			switch {
			case key.Matches(msg, m.Keys.Close), key.Matches(msg, m.Keys.Select):
				m.isViewportVisible = false
			}
		}
	case table.RowSelectedMsg:
		m.viewport.SetContent(m.renderSelectedRow())
	}

	return cmd
}

// View renders the table
func (m *Model) View() string {
	var s = &common.DefaultStyles
	var compositior *lg.Compositor

	if m.isViewportVisible {
		m.table.Styles.Base = m.table.Styles.Base.Foreground(s.Color.Grey75)
		m.table.Styles.Header = m.table.Styles.Header.Foreground(s.Color.Grey50)
		m.table.Styles.Selected = m.table.Styles.Selected.Foreground(lg.Black).Background(s.Color.Grey75)
		m.table.Styles.FilterMatch = m.table.Styles.FilterMatch.Foreground(s.Color.Grey50)
		table := lg.NewLayer(m.table.View())

		viewport := lg.NewLayer(m.viewport.View())
		viewport.X(m.offsetX).Y(m.offsetY)
		compositior = lg.NewCompositor(table, viewport)
	} else {
		m.table.Styles.Base = s.Table.Base
		m.table.Styles.Header = s.Table.Header
		m.table.Styles.Selected = s.Table.Selected
		m.table.Styles.FilterMatch = s.Table.FilterMatch
		table := lg.NewLayer(m.table.View())
		compositior = lg.NewCompositor(table)
	}

	return compositior.Render()
}

func (m *Model) renderSelectedRow() string {
	s := &common.DefaultStyles
	row := m.table.SelectedRow()

	caser := cases.Title(language.English)

	var lines []string
	for _, col := range m.columns {
		key := col.Key()
		value := row.GetString(key)

		if key == "info" {
			infos := strings.Split(value, "|")
			if len(infos) > 1 {
				parsed := []string{""}
				for _, info := range infos {
					split := strings.Split(info, "=")
					subkey := s.Viewport.Label.Foreground(s.Color.Bottom).Render(split[0] + ":")
					subval := s.Viewport.Value.Render(split[1])
					subline := lg.JoinHorizontal(lg.Left, subkey, subval)
					parsed = append(parsed, subline)
				}
				value = lg.JoinVertical(lg.Left, parsed...)
			}
		}

		label := s.Viewport.Label.Width(15).Render(caser.String(key) + ":")
		val := s.Viewport.Value.Render(value)
		line := lg.JoinHorizontal(lg.Left, label, val)
		lines = append(lines, line)
	}

	return lg.JoinVertical(lg.Left, lines...)
}

// SortIPs is a special sort function to sort the net.IP of the "ip" column
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
