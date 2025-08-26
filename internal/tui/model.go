package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the state of our application
type Model struct {
	state          state
	filepicker     filepicker.Model
	optimizerList  list.Model
	qualityInput   textinput.Model
	progress       progress.Model
	outputDirInput textinput.Model
	statusMessage  string
	optimizing     bool
	selectedFile   string
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
	// Create file picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".jpg", ".jpeg", ".png"}

	// Create optimizer list
	optimizerItems := []list.Item{
		item{name: "Single Image Optimization"},
		item{name: "Batch Optimization"},
	}
	optimizerList := list.New(optimizerItems, list.NewDefaultDelegate(), 0, 0)
	optimizerList.Title = "Select Optimization Mode"

	// Create quality input
	qualityInput := textinput.New()
	qualityInput.Placeholder = "Enter quality (1-100)"
	qualityInput.CharLimit = 3
	qualityInput.Width = 30

	// Create output directory input
	outputDirInput := textinput.New()
	outputDirInput.Placeholder = "Enter output directory"
	outputDirInput.Width = 30

	// Create progress bar
	progress := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return Model{
		state:          filePickerState,
		filepicker:     fp,
		optimizerList:  optimizerList,
		qualityInput:   qualityInput,
		progress:       progress,
		outputDirInput: outputDirInput,
	}
}

// Init is called when the program starts
func (m Model) Init() tea.Cmd {
	return m.filepicker.Init()
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.filepicker.Height = msg.Height - h - v - 1
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

	// Update based on current state
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

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	switch m.state {
	case filePickerState:
		return appStyle.Render(m.filepicker.View())
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
