package main

import "github.com/charmbracelet/lipgloss"

var (
	WarningText   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // yellow
	HighlightText = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // bright yellow
	// NoticePrefix  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	BoldCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	StatusText = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00C853")) // bold green adhering to fluent 2 theme
)
