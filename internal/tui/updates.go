package tui

import (
	"fmt"
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
		case key.Matches(msg, m.keys.All):
			m.selectAllFiles()
			return m, nil
		case key.Matches(msg, m.keys.Clear):
			m.selectedFiles = make(map[string]struct{})
			return m, nil
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
	switch msg := msg.(type) {
	case updateStatusMsg:
		m.statusMessage = string(msg)
		return m, nil

	case optimizationMsg:
		// Prepare the run (mkdir, parse quality) then start.
		return m, beginOptimization(msg)

	case optimizeStartedMsg:
		m.optFiles = msg.files
		m.optQuality = msg.quality
		m.optOutputDir = msg.outputDir
		m.optProcessed = 0
		m.optSucceeded = 0
		m.optFailed = 0
		m.optResults = nil
		m.optDone = false
		if len(m.optFiles) == 0 {
			m.optDone = true
			m.statusMessage = "No files to optimize."
			return m, nil
		}
		m.statusMessage = fmt.Sprintf("Optimizing %d files...", len(m.optFiles))
		return m, optimizeOneCmd(m.optFiles[0], m.optQuality, m.optOutputDir)

	case fileDoneMsg:
		m.optProcessed++
		if msg.success {
			m.optSucceeded++
		} else {
			m.optFailed++
		}
		m.optResults = append(m.optResults, msg.result)

		pct := float64(m.optProcessed) / float64(len(m.optFiles))
		cmd := m.progress.SetPercent(pct)

		if m.optProcessed >= len(m.optFiles) {
			m.optDone = true
			m.statusMessage = fmt.Sprintf("Done: %d optimized, %d failed. Press 'q' to quit.", m.optSucceeded, m.optFailed)
			return m, cmd
		}
		m.statusMessage = fmt.Sprintf("Optimizing %d/%d...", m.optProcessed+1, len(m.optFiles))
		return m, tea.Batch(cmd, optimizeOneCmd(m.optFiles[m.optProcessed], m.optQuality, m.optOutputDir))
	}

	progressModel, progressCmd := m.progress.Update(msg)
	m.progress = progressModel.(progress.Model)
	return m, progressCmd
}

// Messages for the optimization run
type updateStatusMsg string

type optimizationMsg struct {
	selectedFiles []string
	quality       string
	outputDir     string
}

type optimizeStartedMsg struct {
	files     []string
	quality   int
	outputDir string
}

type fileDoneMsg struct {
	result  string
	success bool
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
