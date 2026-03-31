package table

import (
	"fmt"
	"strings"
)

// RowData represents a single row's data as a map from column key to value
type RowData map[string]interface{}

// MatchInfo stores filter match positions for cells in a row
type MatchInfo struct {
	CellMatches map[string][]int // list of position indices for each cell
	HasMatch    bool
}

// Row represents a table row
type Row struct {
	Data       RowData
	MatchCache MatchInfo // Stores filter match positions for this row
}

// NewRow creates a new row from RowData
func NewRow(data RowData) Row {
	return Row{Data: data}
}

// Get returns a value from the row by column key
func (r Row) Get(key string) interface{} {
	return r.Data[key]
}

// GetString returns a string value from the row by column key
func (r Row) GetString(key string) string {
	if val, ok := r.Data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// GetSortValue extracts a sortable value from the row by column key
func (r Row) GetSortValue(key string) string {
	if val, ok := r.Data[key]; ok {
		switch v := val.(type) {
		case string:
			return strings.ToLower(v)
		default:
			return strings.ToLower(fmt.Sprintf("%v", v))
		}
	}
	return ""
}
