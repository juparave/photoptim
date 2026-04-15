package tui

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// item represents a file or directory in the list.
type item struct {
	name  string
	isDir bool
}

func (i item) FilterValue() string { return i.name }
func (i item) Description() string { return "" }
func (i item) Title() string {
	icon := fileIcon
	if i.isDir {
		icon = folderIcon
	}
	return fmt.Sprintf("%s %s", icon, i.name)
}

// itemDelegate handles rendering list items.
type itemDelegate struct {
	model *Model
}

func (d itemDelegate) Height() int {
	return 1
}

func (d itemDelegate) Spacing() int {
	return 0
}

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	// Selection state for files
	selectionIndicator := "   "
	if !i.isDir {
		selectionIndicator = fmt.Sprintf(" %s ", uncheckedIcon)
		absPath, err := filepath.Abs(filepath.Join(d.model.currentPath, i.name))
		if err == nil {
			if _, ok := d.model.selectedFiles[absPath]; ok {
				selectionIndicator = fmt.Sprintf(" %s ", checkedIcon)
			}
		}
	}

	// Style selection
	isFocused := index == m.Index()
	baseStyle := itemStyle
	if i.isDir {
		baseStyle = directoryStyle
	}

	title := i.Title()
	var str string
	if isFocused {
		str = selectedItemStyle.Render(fmt.Sprintf("%s%s %s", arrowIcon, selectionIndicator, title))
	} else {
		str = baseStyle.Render(fmt.Sprintf("  %s %s", selectionIndicator, title))
	}

	fmt.Fprint(w, str)
}
