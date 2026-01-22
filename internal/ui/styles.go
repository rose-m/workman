package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Adaptive colors that work on both light and dark backgrounds
	primaryColor = lipgloss.AdaptiveColor{
		Light: "#7C3AED", // Purple
		Dark:  "#A78BFA", // Lighter purple for dark backgrounds
	}

	mutedColor = lipgloss.AdaptiveColor{
		Light: "#6B7280", // Medium gray
		Dark:  "#9CA3AF", // Lighter gray
	}

	borderColor = lipgloss.AdaptiveColor{
		Light: "#D1D5DB", // Light gray
		Dark:  "#4B5563", // Dark gray
	}

	selectedBg = lipgloss.AdaptiveColor{
		Light: "#EDE9FE", // Very light purple
		Dark:  "#1F2937", // Dark gray
	}

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

	// Item styles - use terminal default foreground color
	itemStyle = lipgloss.NewStyle().
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
