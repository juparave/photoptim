package tui

import (
	"context"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"photoptim/internal/remotefs"
	sftpfs "photoptim/internal/sftp"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Padding(1, 2)

// sftpItem represents a file or directory in the SFTP list.
type sftpItem struct {
	name     string
	isDir    bool
	size     int64
	selected bool
}

func (i sftpItem) FilterValue() string { return i.name }
func (i sftpItem) Description() string { return "" }
func (i sftpItem) Title() string {
	icon := fileIcon
	if i.isDir {
		icon = directoryIcon
		return fmt.Sprintf("%s %s", icon, i.name)
	}

	sizeStr := formatFileSize(i.size)
	return fmt.Sprintf("%s %-30s %s", icon, i.name, sizeStr)
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1fGB", float64(size)/(1024*1024*1024))
	}
}

// sftpItemDelegate handles rendering SFTP list items.
type sftpItemDelegate struct {
	model *SFTPModel
}

func (d sftpItemDelegate) Height() int {
	return 1
}

func (d sftpItemDelegate) Spacing() int {
	return 0
}

func (d sftpItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d sftpItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(sftpItem)
	if !ok {
		return
	}

	// Add selection checkbox for files (not directories)
	if !i.isDir {
		selected := "[ ]"
		if i.selected {
			selected = "[x]"
		}
		// For files, show the selection checkbox followed by the formatted title
		str := fmt.Sprintf("%s %s", selected, i.Title())
		fn := itemStyle.Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + s[0])
			}
		} else if i.size > sizeThresholdMB*1024*1024 {
			// Highlight large files
			fn = func(s ...string) string {
				return largeFileStyle.Render(s[0])
			}
		}
		fmt.Fprint(w, fn(str))
	} else {
		// For directories, just show the title
		fn := itemStyle.Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + s[0])
			}
		}
		fmt.Fprint(w, fn(i.Title()))
	}
}

// SFTPState represents the current view of the SFTP TUI.
type SFTPState int

const (
	ConnectionState SFTPState = iota
	BrowserState
)

type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
)

const (
	sizeThresholdMB = 10 // Highlight files larger than 10MB
)

// SFTPModel is the main model for the SFTP TUI.
type SFTPModel struct {
	state         SFTPState
	inputs        []textinput.Model
	focusIndex    int
	spinner       spinner.Model
	loading       bool
	status        string
	currentPath   string
	err           error
	width, height int
	sortMode      SortMode

	sftpClient    *sftpfs.Client
	fileList      list.Model
	selectedFiles map[string]struct{}
}

// --- Bubble Tea Messages ---
type (
	sftpConnectSuccessMsg struct{ client *sftpfs.Client }
	sftpConnectErrorMsg   struct{ err error }
	filesListedMsg        struct{ files []list.Item }
	fileListErrorMsg      struct{ err error }
)

// --- Bubble Tea Commands ---

func (m SFTPModel) connectCmd() tea.Cmd {
	return func() tea.Msg {
		host := m.inputs[0].Value()
		port, _ := strconv.Atoi(m.inputs[1].Value())
		user := m.inputs[2].Value()
		password := m.inputs[3].Value()
		remotePath := m.inputs[4].Value()

		cfg := remotefs.ConnectionConfig{Host: host, Port: port, User: user, Password: password, RemotePath: remotePath}
		client := &sftpfs.Client{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.Connect(ctx, cfg); err != nil {
			return sftpConnectErrorMsg{err: err}
		}

		return sftpConnectSuccessMsg{client: client}
	}
}

func (m SFTPModel) listFilesCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.sftpClient.List(context.Background(), m.currentPath)
		if err != nil {
			return fileListErrorMsg{err}
		}

		// Filter out hidden files (dotfiles)
		var filteredEntries []remotefs.RemoteEntry
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name, ".") {
				filteredEntries = append(filteredEntries, entry)
			}
		}

		items := make([]sftpItem, len(filteredEntries))
		for i, entry := range filteredEntries {
			// Check if this file is selected
			filePath := path.Join(m.currentPath, entry.Name)
			_, isSelected := m.selectedFiles[filePath]

			items[i] = sftpItem{name: entry.Name, isDir: entry.IsDir, size: entry.Size, selected: isSelected}
		}

		// Sort items: directories first, then by sort mode
		sort.Slice(items, func(i, j int) bool {
			// Directories always come first
			if items[i].isDir != items[j].isDir {
				return items[i].isDir
			}

			// Within same type (dir/file), sort by mode
			switch m.sortMode {
			case SortBySize:
				if items[i].isDir {
					// Sort directories by name
					return items[i].name < items[j].name
				}
				// Sort files by size (descending)
				return items[i].size > items[j].size
			default: // SortByName
				return items[i].name < items[j].name
			}
		})

		listItems := make([]list.Item, len(items))
		for i, item := range items {
			listItems[i] = item
		}
		return filesListedMsg{listItems}
	}
}

// --- Model Initialization and Methods ---

func NewSFTPModel() SFTPModel {
	m := SFTPModel{
		state:         ConnectionState,
		focusIndex:    0,
		currentPath:   ".",
		selectedFiles: make(map[string]struct{}),
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.spinner = s

	var inputs []textinput.Model = make([]textinput.Model, 5)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "sftp.example.com"
	inputs[0].Focus()
	inputs[0].CharLimit = 156
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "22"
	inputs[1].CharLimit = 5
	inputs[1].Width = 5
	inputs[1].SetValue("22")

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "username"
	inputs[2].CharLimit = 56
	inputs[2].Width = 20

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "password"
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'
	inputs[3].CharLimit = 128
	inputs[3].Width = 20

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "/"
	inputs[4].CharLimit = 256
	inputs[4].Width = 30
	inputs[4].SetValue("/")
	m.inputs = inputs

	l := list.New([]list.Item{}, sftpItemDelegate{model: &m}, 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	m.fileList = l

	return m
}

func (m SFTPModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SFTPModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		docStyle.Width(m.width)
		m.fileList.SetSize(m.width-4, m.height-4)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case sftpConnectSuccessMsg:
		m.sftpClient = msg.client
		m.state = BrowserState
		m.loading = true
		m.status = "Listing files..."
		return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())

	case sftpConnectErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case filesListedMsg:
		m.loading = false
		m.fileList.SetItems(msg.files)
		sortModeStr := "name"
		if m.sortMode == SortBySize {
			sortModeStr = "size"
		}
		m.fileList.Title = fmt.Sprintf("Remote Files: %s (sorted by %s) - Press 's' to toggle sort", m.currentPath, sortModeStr)
		return m, nil

	case fileListErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			if m.sftpClient != nil {
				m.sftpClient.Close()
			}
			return m, tea.Quit
		}
	}

	switch m.state {
	case ConnectionState:
		return m.updateConnection(msg)
	case BrowserState:
		return m.updateBrowser(msg)
	}

	return m, nil
}

func (m SFTPModel) View() string {
	if m.err != nil {
		return docStyle.Render(fmt.Sprintf("Error: %v\n\n(press any key to return)", m.err))
	}

	if m.loading {
		return fmt.Sprintf("\n   %s %s\n\n", m.spinner.View(), m.status)
	}

	switch m.state {
	case BrowserState:
		return docStyle.Render(m.fileList.View())
	default:
		return m.connectionView()
	}
}

func (m *SFTPModel) updateConnection(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			return m.updateFocus(msg)
		case "enter":
			if m.focusIndex == len(m.inputs) {
				m.loading = true
				m.status = "Connecting..."
				return m, tea.Batch(m.spinner.Tick, m.connectCmd())
			}
			return m.updateFocus(tea.KeyMsg{Type: tea.KeyTab})
		}
	}

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
	}
	return m, cmd
}

func (m *SFTPModel) updateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.sftpClient != nil {
				m.sftpClient.Close()
			}
			return m, tea.Quit
		case " ":
			// Handle spacebar for file selection - return early to prevent list from processing
			i, ok := m.fileList.SelectedItem().(sftpItem)
			if ok && !i.isDir {
				m.toggleFileSelection(i.name)
			}
			return m, nil
		case "enter":
			i, ok := m.fileList.SelectedItem().(sftpItem)
			if ok && i.isDir {
				m.currentPath = path.Join(m.currentPath, i.name)
				m.loading = true
				m.status = fmt.Sprintf("Loading %s...", m.currentPath)
				return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
			}
		case "backspace", "left":
			if m.currentPath != "." {
				m.currentPath = path.Dir(m.currentPath)
				m.loading = true
				m.status = fmt.Sprintf("Loading %s...", m.currentPath)
				return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
			}
		case "s":
			// Toggle sort mode
			if m.sortMode == SortByName {
				m.sortMode = SortBySize
			} else {
				m.sortMode = SortByName
			}
			m.loading = true
			m.status = "Resorting files..."
			return m, tea.Batch(m.spinner.Tick, m.listFilesCmd())
		}
	}

	// For all other messages, update the file list
	fileList, newCmd := m.fileList.Update(msg)
	m.fileList = fileList
	cmd = newCmd
	return m, cmd
}

func (m *SFTPModel) updateFocus(msg tea.Msg) (tea.Model, tea.Cmd) {
	s := msg.(tea.KeyMsg).String()

	if s == "up" || s == "shift+tab" {
		m.focusIndex--
	} else {
		m.focusIndex++
	}

	if m.focusIndex > len(m.inputs) {
		m.focusIndex = 0
	} else if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs)
	}

	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			continue
		}
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = lipgloss.NewStyle()
	}

	return m, tea.Batch(cmds...)
}

func (m SFTPModel) connectionView() string {
	var b strings.Builder

	b.WriteString("Connect to SFTP Server\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View() + "\n")
	}

	button := "[ Connect ]"
	if m.focusIndex != len(m.inputs) {
		button = "  Connect  "
	}

	fmt.Fprintf(&b, "\n%s\n\n", button)
	b.WriteString("(q to quit, tab to navigate)\n")

	return docStyle.Render(b.String())
}

// toggleFileSelection toggles the selection state of a file
func (m *SFTPModel) toggleFileSelection(filename string) {
	filePath := path.Join(m.currentPath, filename)
	if _, ok := m.selectedFiles[filePath]; ok {
		delete(m.selectedFiles, filePath)
	} else {
		m.selectedFiles[filePath] = struct{}{}
	}

	// Update the list items to reflect the new selection state
	items := m.fileList.Items()
	newItems := make([]list.Item, len(items))
	for i, item := range items {
		if sftpItem, ok := item.(sftpItem); ok {
			// Create a new item with updated selection state
			itemFilePath := path.Join(m.currentPath, sftpItem.name)
			_, isSelected := m.selectedFiles[itemFilePath]
			newSftpItem := sftpItem
			newSftpItem.selected = isSelected
			newItems[i] = newSftpItem
		} else {
			newItems[i] = item
		}
	}
	m.fileList.SetItems(newItems)
}

// isFileSelected checks if a file is currently selected
func (m *SFTPModel) isFileSelected(filename string) bool {
	filePath := path.Join(m.currentPath, filename)
	_, ok := m.selectedFiles[filePath]
	return ok
}

// getSelectedFiles returns a slice of all selected file paths
func (m *SFTPModel) getSelectedFiles() []string {
	files := make([]string, 0, len(m.selectedFiles))
	for file := range m.selectedFiles {
		files = append(files, file)
	}
	return files
}
