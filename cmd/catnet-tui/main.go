package main

import (
	"fmt"
	"os"

	"github.com/catnet-io/tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(ui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running CatNet TUI: %v\n", err)
		os.Exit(1)
	}
}
