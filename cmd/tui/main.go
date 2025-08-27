package main

import (
	"fmt"
	"os"

	"photoptim/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
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
