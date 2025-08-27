package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/juparave/photoptim/internal/optimizer"

	tea "github.com/charmbracelet/bubbletea"
)

func runOptimization(msg optimizationMsg) tea.Cmd {
	return func() tea.Msg {
		// Create output directory if it doesn't exist
		os.MkdirAll(msg.outputDir, os.ModePerm)

		// Create optimizer
		opt := optimizer.New()
		quality, err := strconv.Atoi(msg.quality)
		if err != nil {
			quality = 80 // default quality
		}
		opt.Quality = quality

		// Process files
		for _, file := range msg.selectedFiles {
			filename := filepath.Base(file)
			outputPath := filepath.Join(msg.outputDir, filename)

			if err := opt.Optimize(file, outputPath); err != nil {
				return updateStatusMsg(fmt.Sprintf("Error optimizing %s: %v", filename, err))
			}
		}

		return finishedMsg{}
	}
}
