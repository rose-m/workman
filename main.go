package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michael-rose/workman/internal/config"
	"github.com/michael-rose/workman/internal/state"
	"github.com/michael-rose/workman/internal/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize application state
	appState := state.New(cfg)

	// Create the Bubble Tea model
	model := ui.NewModel(appState)

	// Create the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
