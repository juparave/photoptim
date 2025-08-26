package tui

// item represents an item in the list
type item struct {
	name string
}

func (i item) FilterValue() string { return i.name }
func (i item) Title() string       { return i.name }
func (i item) Description() string { return "" }
