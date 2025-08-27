package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the state of our application
type Model struct {
	state          state
	fileList       list.Model
	optimizerList  list.Model
	qualityInput   textinput.Model
	progress       progress.Model
	outputDirInput textinput.Model
	statusMessage  string
	optimizing     bool
	selectedFiles  map[string]struct{}
	currentPath    string
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
	}

	// Create file list
	fileList := list.New([]list.Item{}, itemDelegate{model: &m}, 0, 0)
	fileList.Title = "Select files to optimize"
	fileList.SetShowHelp(true)
	m.fileList = fileList

	// Create optimizer list
	optimizerItems := []list.Item{
		item{name: "Batch Optimization"},
	}
	optimizerList := list.New(optimizerItems, list.NewDefaultDelegate(), 0, 0)
	optimizerList.Title = "Select Optimization Mode"
	m.optimizerList = optimizerList

	// Create quality input
	qualityInput := textinput.New()
	qualityInput.Placeholder = "Enter quality (1-100)"
	qualityInput.CharLimit = 3
	qualityInput.Width = 30
	m.qualityInput = qualityInput

	// Create output directory input
	outputDirInput := textinput.New()
	outputDirInput.Placeholder = "Enter output directory"
	outputDirInput.Width = 30
	m.outputDirInput = outputDirInput

	// Create progress bar
	progress := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	m.progress = progress

	return m
}

// Init is called when the program starts
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.updateFileList(m.currentPath))
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.fileList.SetSize(msg.Width-h*2, msg.Height-v*2)
		m.optimizerList.SetSize(msg.Width, msg.Height-v-1)
		m.qualityInput.Width = msg.Width - h - 1
		m.outputDirInput.Width = msg.Width - h - 1
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.state != filePickerState {
				m.state = filePickerState
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	// Update based on current state
	switch m.state {
	case filePickerState:
		newModel, newCmd := m.updateFilePicker(msg)
		*m = *newModel
		cmd = newCmd
	case optimizerListState:
		newModel, newCmd := m.updateOptimizerList(msg)
		*m = *newModel
		cmd = newCmd
	case qualityInputState:
		newModel, newCmd := m.updateQualityInput(msg)
		*m = *newModel
		cmd = newCmd
	case outputDirState:
		newModel, newCmd := m.updateOutputDirInput(msg)
		*m = *newModel
		cmd = newCmd
	case optimizingState:
		newModel, newCmd := m.updateOptimizing(msg)
		*m = *newModel
		cmd = newCmd
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	switch m.state {
	case filePickerState:
		return appStyle.Render(m.fileList.View())
	case optimizerListState:
		return appStyle.Render(m.optimizerList.View())
	case qualityInputState:
		return appStyle.Render(fmt.Sprintf(
			"Enter quality for JPEG compression (1-100):\n\n%s\n\n%s",
			m.qualityInput.View(),
			"(press enter to continue, esc to go back)",
		))
	case outputDirState:
		return appStyle.Render(fmt.Sprintf(
			"Enter output directory:\n\n%s\n\n%s",
			m.outputDirInput.View(),
			"(press enter to start optimization, esc to go back)",
		))
	case optimizingState:
		return appStyle.Render(fmt.Sprintf(
			"Optimizing images...\n\n%s\n\n%s",
			m.progress.View(),
			m.statusMessage,
		))
	default:
		return "Unknown state"
	}
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

// Styles
var (
	appStyle = lipgloss.NewStyle().
		Margin(1, 2).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#87CEEB")).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#000000"))
)

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
