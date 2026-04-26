package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Primary   = lipgloss.Color("#7D56F4")
	Secondary = lipgloss.Color("#2E8B57")
	Danger    = lipgloss.Color("#FF5F56")
	Muted     = lipgloss.Color("#8B8B8B")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginLeft(2).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginLeft(2).
			MarginBottom(1)

	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(Primary).
				Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true).
			MarginLeft(2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1).
			MarginLeft(2)
)
