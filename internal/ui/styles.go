package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#66fcf1")).
			Bold(true).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#45a29e")).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#45a29e"))

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0b0c10")).
				Background(lipgloss.Color("#66fcf1")).
				Bold(true)

	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c5c6c7"))

	cyanText  = lipgloss.NewStyle().Foreground(lipgloss.Color("#66fcf1"))
	greyText  = lipgloss.NewStyle().Foreground(lipgloss.Color("#45a29e"))
	redText   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0055"))
	greenText = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff88"))
)

func truncate(s string, l int) string {
	if len(s) > l {
		if l <= 3 {
			return s[:l]
		}
		return s[:l-3] + "..."
	}
	return s
}
