package table

import (
	"cmp"
	"fmt"
	"strings"

	lg "charm.land/lipgloss/v2"
)

// Column defines a table column
type Column struct {
	key        string
	title      string
	width      int
	flex       int
	isFlex     bool
	filterable bool
	sortFunc   SortFunc
	style      lg.Style
	isStyled   bool
}

// NewColumn creates a new fixed-width column
func NewColumn(key, title string, width int) Column {
	return Column{
		key:    key,
		title:  title,
		width:  width,
		isFlex: false,
		sortFunc: defaultSortFunc,
	}
}

// NewFlexColumn creates a new flexible-width column
func NewFlexColumn(key, title string, flex int) Column {
	return Column{
		key:    key,
		title:  title,
		flex:   flex,
		isFlex: true,
		sortFunc: defaultSortFunc,
	}
}

// WithFiltering enables filtering for this column
func (c Column) WithFiltering(filterable bool) Column {
	c.filterable = filterable
	return c
}

// WithTitle updates the column title
func (c Column) WithTitle(title string) Column {
	c.title = title
	return c
}

// WithSortFunc sets a custom sort function for this column
func (c Column) WithSortFunc(fn SortFunc) Column {
	c.sortFunc = fn
	return c
}

// WithStyle sets a custom style to the cells of this column (overiding Styles.RowCell)
func (c Column) WithStyle(style lg.Style) Column {
	c.style = style
	c.isStyled = true
	return c
}

// Key returns the column key
func (c Column) Key() string {
	return c.key
}

// Title returns the column title
func (c Column) Title() string {
	return c.title
}

// Width returns the fixed width (or the computed width if IsFlex)
func (c Column) Width() int {
	return c.width
}

// FlexFactor returns the flex factor (only valid if IsFlex)
func (c Column) FlexFactor() int {
	return c.flex
}

// IsFlex returns true if this is a flex column
func (c Column) IsFlex() bool {
	return c.isFlex
}

// IsFilterable returns true if this column can be filtered
func (c Column) IsFilterable() bool {
	return c.filterable
}

// GetSortFunc returns the custom sort function for this column (may be nil)
func (c Column) GetSortFunc() SortFunc {
	return c.sortFunc
}

// SortFunc is a user-configurable sort function for a column.
// It receives two values and should return:
//   - a negative number if a < b
//   - zero if a == b
//   - a positive number if a > b
type SortFunc func(a, b interface{}) int

// default function used for columns unless overriden
func defaultSortFunc(a, b interface{}) int {
	// Handle nil cases
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	}

	// Try numeric comparison
	fa, aOk := toFloat64(a)
	fb, bOk := toFloat64(b)
	if aOk && bOk {
		return cmp.Compare(fa, fb)
	}

	// Fall back to string comparison
	return cmp.Compare(
		strings.ToLower(fmt.Sprintf("%v", a)),
		strings.ToLower(fmt.Sprintf("%v", b)),
	)
}

// toFloat64 attempts to convert a value to float64. Returns (0, false) if not a numeric type.
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}
