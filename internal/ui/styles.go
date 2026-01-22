package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor = lipgloss.Color("#7C3AED")
	textColor    = lipgloss.Color("#E5E7EB")
	mutedColor   = lipgloss.Color("#9CA3AF")
	borderColor  = lipgloss.Color("#4B5563")
	selectedBg   = lipgloss.Color("#1F2937")

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	// Item styles
	itemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Background(selectedBg).
				Bold(true).
				Padding(0, 1)

	// Info styles
	infoStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1, 0, 0, 2)
)
