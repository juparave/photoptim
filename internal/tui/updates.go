package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateFilePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case fileListMsg:
		m.fileList.SetItems(msg)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			i, ok := m.fileList.SelectedItem().(item)
			if !ok {
				return m, nil
			}

			if i.isDir {
				m.currentPath = filepath.Join(m.currentPath, i.name)
				return m, m.updateFileList(m.currentPath)
			} else {
				m.toggleFileSelection(i.name)
			}

		case key.Matches(msg, m.keys.Space):
			i, ok := m.fileList.SelectedItem().(item)
			if ok && !i.isDir {
				m.toggleFileSelection(i.name)
			}
		case key.Matches(msg, m.keys.Select):
			if len(m.selectedFiles) > 0 {
				m.state = optimizerListState
			}
			return m, nil
		case key.Matches(msg, m.keys.Back):
			if m.currentPath != "." {
				m.currentPath = filepath.Dir(m.currentPath)
				return m, m.updateFileList(m.currentPath)
			}
		}
	}

	m.fileList, cmd = m.fileList.Update(msg)
	return m, cmd
}

func (m Model) updateOptimizerList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			m.state = qualityInputState
			m.qualityInput.Focus()
			return m, tea.Batch(textinput.Blink)
		case key.Matches(msg, m.keys.Back):
			m.state = filePickerState
			return m, nil
		}
	}

	m.optimizerList, cmd = m.optimizerList.Update(msg)
	return m, cmd
}

func (m Model) updateQualityInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			m.state = outputDirState
			m.qualityInput.Blur()
			m.outputDirInput.Focus()
			return m, tea.Batch(textinput.Blink)
		case key.Matches(msg, m.keys.Back):
			m.state = optimizerListState
			m.qualityInput.Blur()
			return m, nil
		}
	}

	m.qualityInput, cmd = m.qualityInput.Update(msg)
	return m, cmd
}

func (m Model) updateOutputDirInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			m.state = optimizingState
			m.outputDirInput.Blur()
			return m, startOptimization(m)
		case key.Matches(msg, m.keys.Back):
			m.state = qualityInputState
			m.outputDirInput.Blur()
			m.qualityInput.Focus()
			return m, textinput.Blink
		}
	}

	m.outputDirInput, cmd = m.outputDirInput.Update(msg)
	return m, cmd
}

func (m Model) updateOptimizing(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case progressMsg:
		cmd = m.progress.SetPercent(float64(msg) / 100.0)
		return m, cmd
	case updateStatusMsg:
		m.statusMessage = string(msg)
		return m, nil
	case finishedMsg:
		m.statusMessage = "Optimization completed! Press 'q' to quit."
		return m, nil
	case optimizationMsg:
		return m, runOptimization(msg)
	}

	progressModel, progressCmd := m.progress.Update(msg)
	m.progress = progressModel.(progress.Model)
	return m, progressCmd
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
