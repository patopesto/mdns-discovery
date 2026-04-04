package table

// bubbletea messages sent back from Update() method

// RowSelectedMsg is sent when the cursor is changed with KeyMap.Up and KeyMap.Down keys
type RowSelectedMsg struct {
	// The current index of the selected row
	Index int
}

// FilterInputFocusedMsg is sent when the filter input is put into focus with the KeyMap.Filter key
type FilterInputFocusedMsg struct{}

// FilterInputBlurredMsg is sent when the filter input is put out of focus with the KeyMap.FilterBlur key
type FilterInputBlurredMsg struct{}

// FilterInputClearedMsg is sent when the filter is cleared with the KeyMap.FilterClear key
type FilterInputClearedMsg struct{}
