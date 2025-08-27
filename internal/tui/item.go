package tui

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	fileIcon          = "\U0001F4C4" // 📄
	directoryIcon     = "\U0001F4C1" // 📁
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
		icon = directoryIcon
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

	// Check if the item is selected
	selected := "[ ]"
	absPath, err := filepath.Abs(filepath.Join(d.model.currentPath, i.name))
	if err == nil {
		if _, ok := d.model.selectedFiles[absPath]; ok {
			selected = "[x]"
		}
	}

	str := i.Title()
	if !i.isDir {
		str = fmt.Sprintf("%s %s", selected, str)
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + s[0])
		}
	}

	fmt.Fprint(w, fn(str))
}
