package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/juparave/photoptim/internal/tui"
)

func main() {
	// Create the model
	model := tui.NewModel()

	// Create the program
	program := tea.NewProgram(&model)

	// Run the program
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
