package table

import (
	lg "charm.land/lipgloss/v2"
)

type Styles struct {
	// Style applied to whole table
	Base lg.Style

	// Style applied to header row
	Header lg.Style
	// Style applied to each header row cell
	HeaderCell lg.Style

	// Style applied to each row (unless row is selected)
	Row      lg.Style
	Selected lg.Style
	// Style applied to each cell
	RowCell lg.Style

	// Style applied to footer
	Footer lg.Style

	// Style applied to runes matching filter
	FilterMatch lg.Style

	// Style applied to filter input
	FilterInputFocused lg.Style
	FilterInputBlurred lg.Style
}

func DefaultStyles() (s Styles) {
	s.Base = lg.NewStyle()

	s.Header = lg.NewStyle()
	s.HeaderCell = lg.NewStyle()

	s.Row = lg.NewStyle()
	s.Selected = lg.NewStyle()
	s.RowCell = lg.NewStyle()

	s.Footer = lg.NewStyle()

	s.FilterMatch = lg.NewStyle()
	s.FilterInputFocused = lg.NewStyle()
	s.FilterInputBlurred = lg.NewStyle()

	return s
}
