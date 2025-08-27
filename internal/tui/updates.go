package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateFilePicker(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case fileListMsg:
		m.fileList.SetItems(msg)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			i, ok := m.fileList.SelectedItem().(item)
			if !ok {
				return m, nil
			}

			if i.isDir {
				// It's a directory, so navigate into it
				m.currentPath = filepath.Join(m.currentPath, i.name)
				return m, m.updateFileList(m.currentPath)
			} else {
				// It's a file, so toggle selection
				m.toggleFileSelection(i.name)
			}

		case " ": // Spacebar
			i, ok := m.fileList.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			if !i.isDir {
				m.toggleFileSelection(i.name)
			}
		case "s": // continue to next step
			if len(m.selectedFiles) > 0 {
				m.state = optimizerListState
			}
			return m, nil
		}
	}

	fileList, cmd := m.fileList.Update(msg)
	m.fileList = fileList
	return m, cmd
}

func (m *Model) updateOptimizerList(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Move to quality input
			m.state = qualityInputState
			m.qualityInput.Focus()
			return m, tea.Batch(textinput.Blink)
		}
	}

	optimizerList, cmd := m.optimizerList.Update(msg)
	m.optimizerList = optimizerList
	return m, cmd
}

func (m *Model) updateQualityInput(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Move to output directory input
			m.state = outputDirState
			m.qualityInput.Blur()
			m.outputDirInput.Focus()
			return m, tea.Batch(textinput.Blink)
		}
	}

	qualityInput, cmd := m.qualityInput.Update(msg)
	m.qualityInput = qualityInput
	return m, cmd
}

func (m *Model) updateOutputDirInput(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Start optimization
			m.state = optimizingState
			m.outputDirInput.Blur()
			return m, startOptimization(*m)
		}
	}

	outputDirInput, cmd := m.outputDirInput.Update(msg)
	m.outputDirInput = outputDirInput
	return m, cmd
}

func (m *Model) updateOptimizing(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case progressMsg:
		// Update progress
		cmd = m.progress.SetPercent(float64(msg) / 100.0)
		return m, cmd
	case updateStatusMsg:
		// Update status message
		m.statusMessage = string(msg)
		return m, nil
	case finishedMsg:
		// Optimization finished
		m.statusMessage = "Optimization completed! Press 'q' to quit."
		return m, nil
	case optimizationMsg:
		// Start running the optimization
		return m, runOptimization(msg)
	}

	// Handle progress model updates
	progressModel, progressCmd := m.progress.Update(msg)
	m.progress = progressModel.(progress.Model)
	cmd = progressCmd
	return m, cmd
}

// Messages for progress updates
type progressMsg float64
type finishedMsg struct{}
type updateStatusMsg string

type optimizationMsg struct {
	selectedFiles []string
	quality       string
	outputDir     string
}

func startOptimization(m Model) tea.Cmd {
	return func() tea.Msg {
		return optimizationMsg{
			selectedFiles: m.getSelectedFiles(),
			quality:       m.qualityInput.Value(),
			outputDir:     m.outputDirInput.Value(),
		}
	}
}
