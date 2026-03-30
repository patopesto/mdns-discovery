package table

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
)

// Model is the internal table model implementing tea.Model
type Model struct {
	// Core data
	columns []Column
	rows    []Row
	allRows []Row // stores all rows for filtering

	// Display
	width            int
	height           int
	innerHeight      int // height available to rows (after header and footer)
	horizontalScroll int // horizontal scroll offset

	// Control
	Keys   KeyMap
	Styles Styles

	// Focus and state
	focused            bool
	cursor             int // selected row index
	highlightedRow     int
	filterInputFocused bool

	// Filtering
	filterInput textinput.Model
	filtered    bool
	filterText  string

	// Sorting
	sortColumn string
	sortAsc    bool

	// Styles
	baseStyle         lg.Style
	headerStyle       lg.Style
	highlightStyle    lg.Style
	filterPromptStyle lg.Style
}

// New creates a new table model with the given columns
func New(columns []Column) Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 50
	ti.SetWidth(30)

	return Model{
		columns:           columns,
		rows:              []Row{},
		allRows:           []Row{},
		horizontalScroll:  0,
		focused:           true,
		cursor:            0,
		highlightedRow:    0,
		Keys:              DefaultKeyMap,
		Styles:            DefaultStyles(),
		filterInput:       ti,
		filtered:          false,
		filterText:        "",
		sortColumn:        "",
		sortAsc:           true,
		baseStyle:         lg.NewStyle(),
		headerStyle:       lg.NewStyle().Bold(true),
		highlightStyle:    lg.NewStyle().Reverse(true),
		filterPromptStyle: lg.NewStyle().Foreground(lg.Color("240")),
	}
}

// Focused sets whether the table has focus
func (m Model) Focused(focused bool) Model {
	m.focused = focused
	return m
}

// Filtered enables/disables filtering
func (m Model) Filtered(filtered bool) Model {
	m.filtered = filtered
	return m
}

// WithRows sets the table rows
func (m Model) WithRows(rows []Row) Model {
	m.allRows = rows
	m.applyFilterAndSort()
	return m
}

// WithColumns updates the columns
func (m Model) WithColumns(columns []Column) Model {
	m.columns = columns
	m.calculateColumnWidths()
	return m
}

// WithMinimumHeight sets the minimum height (used for layout)
func (m Model) WithMinimumHeight(height int) Model {
	m.height = height
	return m
}

// WithTargetWidth sets the target width
func (m Model) WithTargetWidth(width int) Model {
	if m.width != width {
		m.width = width
		m.calculateColumnWidths()
	} else {
		m.width = width
	}
	return m
}

// SortByAsc sorts by a column in ascending order
func (m Model) SortByAsc(column string) Model {
	m.sortColumn = column
	m.sortAsc = true
	m.applyFilterAndSort()
	return m
}

// SortByDesc sorts by a column in descending order
func (m Model) SortByDesc(column string) Model {
	m.sortColumn = column
	m.sortAsc = false
	m.applyFilterAndSort()
	return m
}

// IsFiltered returns whether a filter is applied
func (m Model) IsFiltered() bool {
	return m.filterText != ""
}

// IsFilterInputFocused returns whether the filter input is focused
func (m Model) IsFilterInputFocused() bool {
	return m.filterInputFocused
}

// applyFilterAndSort applies both filtering and sorting to allRows
func (m *Model) applyFilterAndSort() {
	// First filter
	filtered := make([]Row, len(m.allRows))
	copy(filtered, m.allRows)
	if m.filtered && m.filterText != "" {
		filtered = m.filterRows(m.allRows, m.filterText)
	}

	// Then sort
	if m.sortColumn != "" {
		// Find the sort function for this column
		var sortFunc SortFunc
		for _, col := range m.columns {
			if col.Key() == m.sortColumn {
				sortFunc = col.GetSortFunc()
				break
			}
		}

		slices.SortStableFunc(filtered, func(a, b Row) int {
			var cmp int
			if sortFunc != nil {
				// Use custom sort function
				valA := a.Data[m.sortColumn]
				valB := b.Data[m.sortColumn]
				cmp = sortFunc(valA, valB)
			} else {
				// Use default string comparison
				valA := a.GetSortValue(m.sortColumn)
				valB := b.GetSortValue(m.sortColumn)
				if valA < valB {
					cmp = -1
				} else if valA > valB {
					cmp = 1
				}
			}
			// Reverse if descending
			if !m.sortAsc {
				cmp = -cmp
			}
			return cmp
		})
	}

	m.rows = filtered

	// Ensure cursor is valid
	if m.cursor >= len(m.rows) && len(m.rows) > 0 {
		m.cursor = len(m.rows) - 1
	} else if len(m.rows) == 0 {
		m.cursor = 0
	}
}

// filterRows filters rows based on the search text
func (m *Model) filterRows(rows []Row, search string) []Row {
	if search == "" {
		return rows
	}

	var filtered []Row
	searchLower := strings.ToLower(search)

	for _, row := range rows {
		for _, col := range m.columns {
			if col.IsFilterable() {
				val := row.GetSortValue(col.Key())
				if strings.Contains(val, searchLower) {
					filtered = append(filtered, row)
					break
				}
			}
		}
	}

	return filtered
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.filterInputFocused {
			switch {
			case key.Matches(msg, m.Keys.FilterBlur):
				m.filterInputFocused = false
				m.filterInput.Blur()
			default:
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
				m.filterText = m.filterInput.Value()
				m.applyFilterAndSort()
			}
		} else {
			switch {
			case key.Matches(msg, m.Keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.Keys.Down):
				if m.cursor < len(m.rows)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.Keys.Left):
				if m.horizontalScroll > 0 {
					m.horizontalScroll -= 5
					if m.horizontalScroll < 0 {
						m.horizontalScroll = 0
					}
				}
			case key.Matches(msg, m.Keys.Right):
				m.horizontalScroll += 5
			case key.Matches(msg, m.Keys.Filter) && m.filtered:
				m.filterInputFocused = true
				m.filterInput.Focus()
				cmd = textinput.Blink
				cmds = append(cmds, cmd)
			case key.Matches(msg, m.Keys.FilterClear):
				m.filterText = ""
				m.filterInput.SetValue("")
				m.applyFilterAndSort()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	var s = &m.Styles
	var sections []string

	// Header row
	headerView := m.renderHeader()
	sections = append(sections, headerView)

	m.innerHeight = m.height - lg.Height(headerView) - m.Styles.Base.GetHorizontalFrameSize()

	var footer string
	if m.isFooterActive() {
		footer = m.renderFooter()
		m.innerHeight -= lg.Height(footer)
	}

	// Data rows
	rows := m.renderRows()
	// Empty rows
	padding := m.innerHeight - lg.Height(rows)
	if padding > 0 {
		rows += "\n" + m.renderEmptyRows(padding)
	}
	sections = append(sections, rows)

	if m.isFooterActive() {
		sections = append(sections, footer)
	}

	view := lg.JoinVertical(
		lg.Left,
		sections...,
	)

	return s.Base.Render(view)
}

// renderHeader renders the column headers
func (m Model) renderHeader() string {
	var s = &m.Styles
	var cells []string

	for _, col := range m.columns {
		width := col.width
		title := truncate(col.Title(), width, s.HeaderCell)
		cell := s.HeaderCell.Width(width).Render(title)
		cells = append(cells, cell)
	}

	return s.Header.Render(strings.Join(cells, ""))
}

func (m Model) isFooterActive() bool {
	if m.filtered {
		return m.filterInputFocused || m.filterText != ""
	}
	return false
}

// renderFooter renders the filter input
func (m Model) renderFooter() string {
	var s = &m.Styles
	var footer string
	if m.filterInputFocused {
		footer = m.filterInput.View()
	}
	if m.filterText != "" {
		footer = m.filterPromptStyle.Render("/" + m.filterText)
	}

	return s.Footer.Render(footer)
}

// renderRows renders the data rows
func (m Model) renderRows() string {
	var rows []string

	// Calculate visible row range
	startIdx := 0
	endIdx := len(m.rows)

	if m.innerHeight > 0 && len(m.rows) > m.innerHeight {
		// Simple pagination: show from cursor
		startIdx = m.cursor - m.innerHeight/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + m.innerHeight
		if endIdx > len(m.rows) {
			endIdx = len(m.rows)
			startIdx = endIdx - m.innerHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		row := m.rows[i]
		rowStr := m.renderRow(row, i == m.cursor)
		rows = append(rows, rowStr)
	}

	return strings.Join(rows, "\n")
}

// renderRow renders a single row
func (m Model) renderRow(row Row, isSelected bool) string {
	var s = &m.Styles
	var cells []string

	for _, col := range m.columns {
		style := s.RowCell
		if col.isStyled {
			style = col.style
		}

		value := row.GetString(col.Key())
		value = truncate(value, col.width, style)

		cell := style.Width(col.width).Render(value)
		cells = append(cells, cell)
	}

	var rowStr string
	if isSelected {
		rowStr = s.Selected.Render(strings.Join(cells, ""))
	} else {
		rowStr = s.Row.Render(strings.Join(cells, ""))
	}

	return rowStr
}

// renderRow renders a single empty row (for padding)
func (m Model) renderEmptyRows(num int) string {
	var cells []string
	var rows []string

	for _, col := range m.columns {
		cell := lg.NewStyle().Width(col.width).Render("")
		cells = append(cells, cell)
	}
	row := strings.Join(cells, "")

	for i := 0; i < num; i++ {
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// calculateColumnWidths calculates widths for each column
func (m Model) calculateColumnWidths() {
	if len(m.columns) == 0 || m.width == 0 {
		return
	}

	// Account for borders, etc if present
	availableWidth := m.width - m.Styles.Base.GetVerticalFrameSize()

	// Calculate total flex and fixed widths
	totalFlex := 0
	fixedWidth := 0

	for _, col := range m.columns {
		if col.IsFlex() {
			totalFlex += col.FlexFactor()
		} else {
			fixedWidth += col.Width()
		}
	}

	// Calculate widths
	remainingWidth := availableWidth - fixedWidth
	leftOverWidth := remainingWidth % totalFlex

	var totalCalculatedWidth int
	flexColumnIndices := []int{}

	for i, col := range m.columns {
		if col.IsFlex() {
			flexColumnIndices = append(flexColumnIndices, i)
			width := (col.FlexFactor() * remainingWidth) / totalFlex

			if width < 1 {
				width = 1
			}

			m.columns[i].width = width
			totalCalculatedWidth += width
		} else {
			totalCalculatedWidth += col.Width()
		}
	}

	// Distribute any remaining leftover width to flex columns
	leftOverWidth = availableWidth - totalCalculatedWidth
	for _, i := range flexColumnIndices {
		if leftOverWidth <= 0 {
			break
		}
		m.columns[i].width++
		leftOverWidth--
	}
}

// truncate truncates a string to fit within maxWidth accounting for style
func truncate(s string, maxWidth int, style lg.Style) string {
	if maxWidth <= 0 {
		return ""
	}

	width := lg.Width(s)
	availableWidth := maxWidth - style.GetHorizontalFrameSize()
	if width <= availableWidth {
		return s
	}

	var truncated string
	if style.GetAlign() == lg.Right {
		safeWidth := width - availableWidth
		truncated = "…" + s[safeWidth+1:]
	} else {
		truncated = s[0:maxWidth-1] + "…"
	}

	return truncated
}
