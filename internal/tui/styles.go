// Package tui provides a terminal user interface for Thicket.
package tui

import "github.com/charmbracelet/lipgloss"

// Colors used throughout the TUI.
var (
	colorPrimary   = lipgloss.Color("39")  // Blue
	colorSecondary = lipgloss.Color("243") // Gray
	colorSuccess   = lipgloss.Color("42")  // Green
	colorWarning   = lipgloss.Color("214") // Orange
	colorError     = lipgloss.Color("196") // Red
	colorMuted     = lipgloss.Color("240") // Dark gray
)

// Styles for various UI elements.
var (
	// Header styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary)

	normalRowStyle = lipgloss.NewStyle()

	// Status styles
	statusOpenStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	statusClosedStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	// Priority styles
	priorityHighStyle = lipgloss.NewStyle().
				Foreground(colorError)

	priorityMedStyle = lipgloss.NewStyle().
				Foreground(colorWarning)

	priorityLowStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)

	// Help bar style
	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	statusMsgStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(colorError)

	// Detail view styles
	labelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Width(14)

	valueStyle = lipgloss.NewStyle()

	// Form styles
	focusedInputStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary)

	blurredInputStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted)

	// Filter chip style
	filterChipStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Background(lipgloss.Color("237")).
			Padding(0, 1)

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")) // Gray

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWarning).
			Padding(0, 1)
)
