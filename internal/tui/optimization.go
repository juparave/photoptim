package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/juparave/photoptim/internal/optimizer"

	tea "github.com/charmbracelet/bubbletea"
)

// supportedExts are the image extensions the local optimizer can encode.
var supportedExts = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
}

func isSupportedImage(path string) bool {
	_, ok := supportedExts[strings.ToLower(filepath.Ext(path))]
	return ok
}

// beginOptimization prepares the output directory and kicks off the first file.
func beginOptimization(msg optimizationMsg) tea.Cmd {
	return func() tea.Msg {
		if err := os.MkdirAll(msg.outputDir, 0o755); err != nil {
			return updateStatusMsg(fmt.Sprintf("Error creating output dir: %v", err))
		}
		quality, err := strconv.Atoi(strings.TrimSpace(msg.quality))
		if err != nil || quality < 1 || quality > 100 {
			quality = 80
		}
		return optimizeStartedMsg{
			files:     msg.selectedFiles,
			quality:   quality,
			outputDir: msg.outputDir,
		}
	}
}

// optimizeOneCmd optimizes a single file and reports the result.
func optimizeOneCmd(file string, quality int, outputDir string) tea.Cmd {
	return func() tea.Msg {
		name := filepath.Base(file)
		if !isSupportedImage(file) {
			return fileDoneMsg{result: fmt.Sprintf("⊘ %s: unsupported format", name), success: false}
		}

		opt := optimizer.New()
		opt.Quality = quality
		outputPath := filepath.Join(outputDir, name)
		if err := opt.Optimize(file, outputPath); err != nil {
			return fileDoneMsg{result: fmt.Sprintf("✗ %s: %v", name, err), success: false}
		}
		return fileDoneMsg{result: fmt.Sprintf("✓ %s", name), success: true}
	}
}
