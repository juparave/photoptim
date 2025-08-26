package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateFilePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Check if a file was selected
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// A file was selected
		m.selectedFile = path
		m.state = optimizerListState
		return m, nil
	}
	
	// Check if a disabled file was selected (wrong file type)
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Could show an error message here
		_ = path // Ignore for now
	}
	
	// Update the filepicker
	m.filepicker, cmd = m.filepicker.Update(msg)
	return m, cmd
}

func (m Model) updateOptimizerList(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	m.optimizerList, cmd = m.optimizerList.Update(msg)
	return m, cmd
}

func (m Model) updateQualityInput(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	m.qualityInput, cmd = m.qualityInput.Update(msg)
	return m, cmd
}

func (m Model) updateOutputDirInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Start optimization
			m.state = optimizingState
			m.outputDirInput.Blur()
			return m, startOptimization(m)
		}
	}

	m.outputDirInput, cmd = m.outputDirInput.Update(msg)
	return m, cmd
}

func (m Model) updateOptimizing(msg tea.Msg) (tea.Model, tea.Cmd) {
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
