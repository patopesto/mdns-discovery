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
	width       int
	height      int
	innerHeight int // height available to rows (after header and footer)

	// Control
	Keys   KeyMap
	Styles Styles

	// Focus and state
	focused            bool
	cursor             int // selected row index
	scrollOffset       int // first visible row index
	highlightedRow     int
	filterInputFocused bool

	// Filtering
	filteringEnabled bool
	filterInput      textinput.Model
	filterText       string

	// Sorting
	sortColumn string
	sortAsc    bool
}

// New creates a new table model with the given columns
func New(columns []Column) Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 50
	ti.SetWidth(30)

	return Model{
		columns:          columns,
		rows:             []Row{},
		allRows:          []Row{},
		focused:          true,
		cursor:           0,
		scrollOffset:     0,
		highlightedRow:   0,
		Keys:             DefaultKeyMap,
		Styles:           DefaultStyles(),
		filterInput:      ti,
		filteringEnabled: false,
		filterText:       "",
		sortColumn:       "",
		sortAsc:          true,
	}
}

// WithFiltering enables/disables filtering
func (m Model) WithFiltering(filtered bool) Model {
	m.filteringEnabled = filtered
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

// Focus sets whether the table has focus and interactions such as keys/mouse will work
func (m *Model) Focus(focus bool) {
	m.focused = focus
}

// IsFocused returns the focus state of the table
func (m Model) IsFocused() bool {
	return m.focused
}

// IsFiltered returns whether a filter is currently applied
func (m Model) IsFiltered() bool {
	return m.filterText != ""
}

// IsFilterInputFocused returns whether the filter input is focused
func (m Model) IsFilterInputFocused() bool {
	return m.filterInputFocused
}

// TotalRowCount returns the total number of rows without filtering
func (m Model) TotalRowCount() int {
	return len(m.allRows)
}

// VisibleRowCount returns the number of rows taking filtering into account
func (m Model) VisibleRowCount() int {
	return len(m.allRows)
}

// SelectedIndex returns the index of the currently selected row
func (m Model) SelectedIndex() int {
	return m.cursor
}

// SelectedRow returns the currently selected row
func (m Model) SelectedRow() Row {
	if len(m.rows) > 0 { // safety
		return m.rows[m.cursor]
	}
	return Row{}
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

// applyFilterAndSort applies both filtering and sorting to allRows
func (m *Model) applyFilterAndSort() {
	// First filter
	filtered := make([]Row, len(m.allRows))
	copy(filtered, m.allRows)
	if m.filteringEnabled && m.filterText != "" {
		filtered = m.filterRows(m.allRows, m.filterText)
	} else {
		for i := range filtered {
			filtered[i].MatchCache = MatchInfo{}
		}
	}

	// Then sort
	if m.sortColumn != "" {
		var sortFunc SortFunc
		for _, col := range m.columns {
			if col.Key() == m.sortColumn {
				sortFunc = col.GetSortFunc()
				break
			}
		}

		slices.SortStableFunc(filtered, func(a, b Row) int {
			valA := a.Data[m.sortColumn]
			valB := b.Data[m.sortColumn]
			cmp := sortFunc(valA, valB)
			// Reverse if descending
			if !m.sortAsc {
				cmp = -cmp
			}
			return cmp
		})
	}

	m.rows = filtered

	// Reset scroll offset when filtering
	m.scrollOffset = 0

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
		matchInfo := MatchInfo{CellMatches: make(map[string][]int)}

		for _, col := range m.columns {
			if col.IsFilterable() {
				val := row.GetSortValue(col.Key())
				if idx := strings.Index(val, searchLower); idx >= 0 {
					var indices []int
					for i := idx; i < idx+len(search); i++ {
						indices = append(indices, i)
					}
					matchInfo.CellMatches[col.Key()] = indices
					matchInfo.HasMatch = true
				}
			}
		}

		row.MatchCache = matchInfo
		if matchInfo.HasMatch == true {
			filtered = append(filtered, row)
		}
	}

	return filtered
}

// updateScrollOffset adjusts scrollOffset to ensure cursor is visible
func (m *Model) updateScrollOffset() {
	if m.innerHeight > 0 && len(m.rows) > m.innerHeight {
		// Cursor below visible area - scroll down
		if m.cursor >= m.scrollOffset+m.innerHeight {
			m.scrollOffset = m.cursor - m.innerHeight + 1
		}
		// Cursor above visible area - scroll up
		if m.cursor < m.scrollOffset {
			m.scrollOffset = m.cursor
		}
		// Clamp scrollOffset to valid bounds
		maxOffset := len(m.rows) - m.innerHeight
		if m.scrollOffset > maxOffset {
			m.scrollOffset = maxOffset
		}
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
	} else {
		m.scrollOffset = 0
	}
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return nil
}

// small helper to create a tea.Cmd from a custom tea.Msg
func createCmd(msg any) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.focused == false {
		return m, nil
	}
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.filterInputFocused {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, m.Keys.FilterBlur):
				m.filterInputFocused = false
				m.filterInput.Blur()
				cmd = createCmd(FilterInputBlurredMsg{})
				cmds = append(cmds, cmd)
			default:
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
				m.filterText = m.filterInput.Value()
				m.applyFilterAndSort()
			}
		}
	} else {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, m.Keys.Up):
				if m.cursor > 0 {
					m.cursor--
					m.updateScrollOffset()
					cmd = createCmd(RowSelectedMsg{Index: m.cursor})
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, m.Keys.Down):
				if m.cursor < len(m.rows)-1 {
					m.cursor++
					m.updateScrollOffset()
					cmd = createCmd(RowSelectedMsg{Index: m.cursor})
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, m.Keys.Filter) && m.filteringEnabled:
				m.filterInputFocused = true
				m.filterInput.Focus()
				cmd = textinput.Blink
				cmds = append(cmds, cmd)
				cmd = createCmd(FilterInputFocusedMsg{})
				cmds = append(cmds, cmd)
			case key.Matches(msg, m.Keys.FilterClear):
				m.filterText = ""
				m.filterInput.Reset()
				m.applyFilterAndSort()
				cmd = createCmd(FilterInputClearedMsg{})
				cmds = append(cmds, cmd)
			}
		case tea.MouseWheelMsg:
			switch msg.Button {
			case tea.MouseWheelUp:
				if m.cursor > 0 {
					m.cursor--
					m.updateScrollOffset()
					cmd = createCmd(RowSelectedMsg{Index: m.cursor})
					cmds = append(cmds, cmd)
				}
			case tea.MouseWheelDown:
				if m.cursor < len(m.rows)-1 {
					m.cursor++
					m.updateScrollOffset()
					cmd = createCmd(RowSelectedMsg{Index: m.cursor})
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements the v1 tea.Model interface and returns the renderered table
func (m *Model) View() string {
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
func (m *Model) renderHeader() string {
	var s = &m.Styles
	var cells []string

	rowstyle := s.Header.Inherit(s.Base.UnsetBorderStyle())
	cellstyle := s.HeaderCell.Inherit(rowstyle.UnsetBorderStyle())

	for _, col := range m.columns {
		width := col.width
		title := truncate(col.Title(), width, s.HeaderCell)
		cell := cellstyle.Width(width).Render(title)
		cells = append(cells, cell)
	}

	return rowstyle.Render(strings.Join(cells, ""))
}

func (m *Model) isFooterActive() bool {
	if m.filteringEnabled {
		return m.filterInputFocused || m.filterText != ""
	}
	return false
}

// renderFooter renders the filter input
func (m *Model) renderFooter() string {
	var s = &m.Styles
	var footer string

	rowstyle := s.Footer.Inherit(s.Base.UnsetBorderStyle())
	width := m.width - s.Base.GetHorizontalFrameSize()

	if m.filterInputFocused {
		style := s.FilterInputFocused.Inherit(rowstyle.UnsetBorderStyle())
		footer = style.Render(m.filterInput.View())
	} else if m.filterText != "" {
		style := s.FilterInputBlurred.Inherit(rowstyle.UnsetBorderStyle())
		footer = style.Render("/ " + m.filterText)
	}

	return rowstyle.Width(width).Render(footer)
}

// renderRows renders the data rows
func (m *Model) renderRows() string {
	var rows []string

	startIdx := 0
	endIdx := len(m.rows)

	if m.innerHeight > 0 && len(m.rows) > m.innerHeight {
		startIdx = m.scrollOffset
		endIdx = startIdx + m.innerHeight
		if endIdx > len(m.rows) {
			endIdx = len(m.rows)
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
func (m *Model) renderRow(row Row, isSelected bool) string {
	var s = &m.Styles
	var cells []string

	// Inherit doesn't do margins and paddins but Inline() will block the rendering of
	// future margins and paddings found in the cell and column styles.
	// So we Inherit without Borders to get the style stacking to work
	rowstyle := s.Row.Inherit(s.Base.UnsetBorderStyle())
	if isSelected {
		rowstyle = s.Selected.Inherit(rowstyle)
	}

	for _, col := range m.columns {
		style := s.RowCell.Inherit(rowstyle)
		if col.isStyled {
			style = col.style.Inherit(rowstyle)
		}
		value := row.GetString(col.Key())
		value = truncate(value, col.width, style)

		// Check if this cell has a cached filter match
		if row.MatchCache.HasMatch {
			if matchRange, ok := row.MatchCache.CellMatches[col.Key()]; ok {
				unmatched := style.Inline(true)
				matched := s.FilterMatch.Inherit(unmatched)
				value = lg.StyleRunes(value, matchRange, matched, unmatched)
			}
		}

		cell := style.Width(col.width).Render(value)
		cells = append(cells, cell)
	}

	return rowstyle.Render(strings.Join(cells, ""))
}

// renderRow renders a single empty row (for padding)
func (m *Model) renderEmptyRows(num int) string {
	var s = &m.Styles
	var cells []string
	var rows []string

	rowstyle := lg.NewStyle().Inherit(s.Base.UnsetBorderStyle())

	for _, col := range m.columns {
		cell := rowstyle.Width(col.width).Render("")
		cells = append(cells, cell)
	}
	row := strings.Join(cells, "")

	for i := 0; i < num; i++ {
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// calculateColumnWidths calculates widths for each column
func (m *Model) calculateColumnWidths() {
	if len(m.columns) == 0 || m.width == 0 {
		return
	}

	// Account for borders, etc if present
	availableWidth := m.width - m.Styles.Base.GetHorizontalFrameSize()

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
	if availableWidth <= 0 {
		return ""
	}

	var truncated string
	if style.GetAlign() == lg.Right {
		safeWidth := width - availableWidth
		truncated = "…" + s[safeWidth+1:]
	} else {
		truncated = s[0:availableWidth-1] + "…"
	}

	return truncated
}
