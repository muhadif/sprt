package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Common styles for all UI components
var (
	// Colors
	primaryColor   = lipgloss.Color("#25A065")
	secondaryColor = lipgloss.Color("#FFFDF5")
	textColor      = lipgloss.Color("#FFFFFF")
	mutedColor     = lipgloss.Color("#888888")
	linkColor      = lipgloss.Color("#4A86E8")
	inputBgColor   = lipgloss.Color("#333333")
)

// GetTitleStyle returns a consistent title style for all UIs
func GetTitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(secondaryColor).
		Background(primaryColor).
		Bold(true).
		Padding(0, 1).
		Width(width).
		Align(lipgloss.Center)
}

// GetHeaderStyle returns a consistent header style for all UIs
func GetHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)
}

// GetValueStyle returns a consistent value style for all UIs
func GetValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textColor)
}

// GetInfoStyle returns a consistent info style for all UIs
func GetInfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true)
}

// GetLinkStyle returns a consistent link style for all UIs
func GetLinkStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(linkColor).
		Underline(true)
}

// GetInputStyle returns a consistent input style for all UIs
func GetInputStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textColor).
		Background(inputBgColor).
		Padding(0, 1)
}

// GetBorderStyle returns a consistent border style for all UIs
func GetBorderStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(width - 4)
}

// GetSelectedStyle returns a consistent selected item style for all UIs
func GetSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)
}

// GetNormalStyle returns a consistent normal item style for all UIs
func GetNormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textColor)
}

// GetCurrentLineStyle returns a consistent current line style for lyrics
func GetCurrentLineStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Width(width).
		Align(lipgloss.Center)
}

// GetOtherLineStyle returns a consistent other line style for lyrics
func GetOtherLineStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(textColor).
		Width(width).
		Align(lipgloss.Center)
}
