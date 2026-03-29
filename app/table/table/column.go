package table

// SortFunc is a user-configurable sort function for a column.
// It receives two values and should return:
//   - a negative number if a < b
//   - zero if a == b
//   - a positive number if a > b
type SortFunc func(a, b interface{}) int

// Column defines a table column
type Column struct {
	key        string
	title      string
	width      int
	flex       int
	isFlex     bool
	filterable bool
	sortFunc   SortFunc
}

// NewColumn creates a new fixed-width column
func NewColumn(key, title string, width int) Column {
	return Column{
		key:    key,
		title:  title,
		width:  width,
		isFlex: false,
	}
}

// NewFlexColumn creates a new flexible-width column
func NewFlexColumn(key, title string, flex int) Column {
	return Column{
		key:    key,
		title:  title,
		flex:   flex,
		isFlex: true,
	}
}

// WithFiltered enables filtering for this column
func (c Column) WithFiltered(filterable bool) Column {
	c.filterable = filterable
	return c
}

// SetTitle updates the column title
func (c Column) SetTitle(title string) Column {
	c.title = title
	return c
}

// WithSortFunc sets a custom sort function for this column
func (c Column) WithSortFunc(fn SortFunc) Column {
	c.sortFunc = fn
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
