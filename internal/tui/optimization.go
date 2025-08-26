package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"photoptim/internal/optimizer"

	tea "github.com/charmbracelet/bubbletea"
)

func startOptimization(m Model) tea.Cmd {
	// Get selected mode
	selectedMode := m.optimizerList.SelectedItem().(item).name

	// Get quality
	quality, err := strconv.Atoi(m.qualityInput.Value())
	if err != nil || quality < 1 || quality > 100 {
		quality = 80 // Default quality
	}

	// Get output directory
	outputDir := m.outputDirInput.Value()
	if outputDir == "" {
		outputDir = "optimized"
	}

	// Get selected file
	file := m.selectedFile

	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return optimizationMsg{
			mode:      selectedMode,
			quality:   quality,
			outputDir: outputDir,
			file:      file,
			startTime: t,
		}
	})
}

type optimizationMsg struct {
	mode      string
	quality   int
	outputDir string
	file      string
	startTime time.Time
}

func runOptimization(msg optimizationMsg) tea.Cmd {
	return func() tea.Msg {
		// Create output directory if it doesn't exist
		os.MkdirAll(msg.outputDir, os.ModePerm)

		// Create optimizer
		opt := optimizer.New()
		opt.Quality = msg.quality

		// Process file based on selected mode
		if msg.mode == "Single Image Optimization" && msg.file != "" {
			// Optimize single image
			inputPath := msg.file
			filename := filepath.Base(inputPath)
			outputPath := filepath.Join(msg.outputDir, filename)

			// Perform optimization
			if err := opt.Optimize(inputPath, outputPath); err != nil {
				return updateStatusMsg(fmt.Sprintf("Error: %v", err))
			} else {
				return updateStatusMsg(fmt.Sprintf("Optimized %s successfully", filename))
			}
		} else if msg.mode == "Batch Optimization" {
			// For simplicity, we'll just optimize the single selected file
			// In a more complete implementation, we would select multiple files
			inputPath := msg.file
			filename := filepath.Base(inputPath)
			outputPath := filepath.Join(msg.outputDir, filename)

			if err := opt.Optimize(inputPath, outputPath); err != nil {
				return updateStatusMsg(fmt.Sprintf("Error optimizing %s: %v", filename, err))
			}
			return updateStatusMsg("Batch optimization completed!")
		}

		return updateStatusMsg("Optimization completed!")
	}
}
