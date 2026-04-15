package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap defines the keybindings for our application
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Space  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Select key.Binding
}

var keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/confirm"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle selection"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Select: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "proceed to optimization"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Space, k.Select, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Space},
		{k.Select, k.Back, k.Quit},
	}
}

// Model represents the state of our application
type Model struct {
	state          state
	fileList       list.Model
	optimizerList  list.Model
	qualityInput   textinput.Model
	progress       progress.Model
	outputDirInput textinput.Model
	help           help.Model
	keys           KeyMap
	statusMessage  string
	optimizing     bool
	selectedFiles  map[string]struct{}
	currentPath    string
	width, height  int
}

type state int

const (
	filePickerState state = iota
	optimizerListState
	qualityInputState
	outputDirState
	optimizingState
)

func NewModel() Model {
	m := Model{
		state:         filePickerState,
		selectedFiles: make(map[string]struct{}),
		currentPath:   ".",
		keys:          keys,
		help:          help.New(),
	}

	// Create file list
	fileList := list.New([]list.Item{}, itemDelegate{model: &m}, 0, 0)
	fileList.Title = "Select files to optimize"
	fileList.SetShowHelp(false) // We use our own help component
	fileList.Styles.Title = titleStyle
	m.fileList = fileList

	// Create optimizer list
	optimizerItems := []list.Item{
		item{name: "Batch Optimization"},
	}
	optimizerList := list.New(optimizerItems, list.NewDefaultDelegate(), 0, 0)
	optimizerList.Title = "Select Optimization Mode"
	optimizerList.Styles.Title = titleStyle
	m.optimizerList = optimizerList

	// Create quality input
	qualityInput := textinput.New()
	qualityInput.Placeholder = "Enter quality (1-100)"
	qualityInput.CharLimit = 3
	qualityInput.Width = 30
	qualityInput.PromptStyle = primaryColorStyle
	m.qualityInput = qualityInput

	// Create output directory input
	outputDirInput := textinput.New()
	outputDirInput.Placeholder = "Enter output directory"
	outputDirInput.Width = 30
	outputDirInput.PromptStyle = primaryColorStyle
	m.outputDirInput = outputDirInput

	// Create progress bar
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	m.progress = prog

	return m
}

var primaryColorStyle = lipgloss.NewStyle().Foreground(primaryColor)

// Init is called when the program starts
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.updateFileList(m.currentPath), textinput.Blink)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h, v := appStyle.GetFrameSize()
		m.fileList.SetSize(msg.Width-h-4, msg.Height-v-6)
		m.optimizerList.SetSize(msg.Width-h-4, msg.Height-v-6)
		m.progress.Width = msg.Width - h - 10
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			// If we are in an input state, only quit on ctrl+c, not 'q'
			if msg.String() == "q" && (m.state == qualityInputState || m.state == outputDirState) {
				break
			}
			return m, tea.Quit
		}
	}

	// Delegate updates based on state
	switch m.state {
	case filePickerState:
		return m.updateFilePicker(msg)
	case optimizerListState:
		return m.updateOptimizerList(msg)
	case qualityInputState:
		return m.updateQualityInput(msg)
	case outputDirState:
		return m.updateOutputDirInput(msg)
	case optimizingState:
		return m.updateOptimizing(msg)
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	var content string

	switch m.state {
	case filePickerState:
		content = m.fileList.View()
	case optimizerListState:
		content = m.optimizerList.View()
	case qualityInputState:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			headerStyle.Render("Optimization Settings"),
			"Enter quality for JPEG compression (1-100):",
			m.qualityInput.View(),
		)
	case outputDirState:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			headerStyle.Render("Optimization Settings"),
			"Enter output directory:",
			m.outputDirInput.View(),
		)
	case optimizingState:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			headerStyle.Render("Optimizing images..."),
			m.progress.View(),
			statusStyle.Render(m.statusMessage),
		)
	}

	helpView := m.help.View(m.keys)
	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content, footerStyle.Render(helpView)))
}

// updateFileList reads the contents of a directory and updates the file list.
func (m *Model) updateFileList(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := os.ReadDir(path)
		if err != nil {
			return errMsg{err}
		}

		items := []list.Item{}
		// Add parent directory navigation
		if path != "." {
			items = append(items, item{name: "..", isDir: true})
		}

		for _, f := range files {
			items = append(items, item{name: f.Name(), isDir: f.IsDir()})
		}
		return fileListMsg(items)
	}
}

type fileListMsg []list.Item
type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

func (m Model) getSelectedFiles() []string {
	var files []string
	for file := range m.selectedFiles {
		files = append(files, file)
	}
	return files
}

func (m *Model) toggleFileSelection(path string) {
	absPath, err := filepath.Abs(filepath.Join(m.currentPath, path))
	if err != nil {
		// Handle error appropriately
		return
	}
	if _, ok := m.selectedFiles[absPath]; ok {
		delete(m.selectedFiles, absPath)
	} else {
		m.selectedFiles[absPath] = struct{}{}
	}
}
