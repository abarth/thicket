package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpModel provides a help overlay for displaying keyboard shortcuts.
type HelpModel struct {
	width  int
	height int
}

// NewHelp creates a new help model.
func NewHelp() HelpModel {
	return HelpModel{}
}

// SetSize sets the dimensions for the help overlay.
func (h *HelpModel) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// helpEntry represents a single key binding with description.
type helpEntry struct {
	key  string
	desc string
}

// helpSection represents a group of related key bindings.
type helpSection struct {
	title   string
	entries []helpEntry
}

// ListFullHelp returns the full help content for list view.
func ListFullHelp() []helpSection {
	return []helpSection{
		{
			title: "Navigation",
			entries: []helpEntry{
				{"j / ↓", "Move down"},
				{"k / ↑", "Move up"},
				{"g / Home", "Go to top"},
				{"G / End", "Go to bottom"},
				{"Enter / l", "View ticket details"},
			},
		},
		{
			title: "Actions",
			entries: []helpEntry{
				{"n", "Create new ticket"},
				{"e", "Edit selected ticket"},
				{"c", "Close selected ticket"},
				{"+/=", "Increase priority (lower number)"},
				{"-/_", "Decrease priority (higher number)"},
			},
		},
		{
			title: "Set Type",
			entries: []helpEntry{
				{"b", "Set type to bug"},
				{"f", "Set type to feature"},
				{"t", "Set type to task"},
				{"E", "Set type to epic"},
				{"C", "Set type to cleanup"},
			},
		},
		{
			title: "Filtering",
			entries: []helpEntry{
				{"o", "Show open tickets only"},
				{"x", "Show closed tickets only"},
				{"i", "Show icebox tickets only"},
				{"a", "Show all tickets"},
				{"/", "Search tickets"},
			},
		},
		{
			title: "General",
			entries: []helpEntry{
				{"r", "Refresh list"},
				{"?", "Toggle this help"},
				{"q", "Quit"},
			},
		},
	}
}

// DetailFullHelp returns the full help content for detail view.
func DetailFullHelp() []helpSection {
	return []helpSection{
		{
			title: "Navigation",
			entries: []helpEntry{
				{"j / ↓", "Scroll down"},
				{"k / ↑", "Scroll up"},
				{"Esc / h", "Back to list"},
			},
		},
		{
			title: "Actions",
			entries: []helpEntry{
				{"e", "Edit ticket"},
				{"c", "Close ticket"},
				{"m", "Add comment"},
				{"+/=", "Increase priority"},
				{"-/_", "Decrease priority"},
			},
		},
		{
			title: "Set Type",
			entries: []helpEntry{
				{"b", "Set type to bug"},
				{"f", "Set type to feature"},
				{"t", "Set type to task"},
				{"E", "Set type to epic"},
				{"C", "Set type to cleanup"},
			},
		},
		{
			title: "General",
			entries: []helpEntry{
				{"?", "Toggle this help"},
				{"q", "Quit"},
			},
		},
	}
}

// FormFullHelp returns the full help content for form view.
func FormFullHelp() []helpSection {
	return []helpSection{
		{
			title: "Navigation",
			entries: []helpEntry{
				{"Tab", "Next field"},
				{"Shift+Tab", "Previous field"},
				{"← / h", "Previous option (selectors)"},
				{"→ / l", "Next option (selectors)"},
			},
		},
		{
			title: "Actions",
			entries: []helpEntry{
				{"Ctrl+S", "Save ticket"},
				{"Esc", "Cancel and go back"},
			},
		},
		{
			title: "General",
			entries: []helpEntry{
				{"?", "Toggle this help"},
			},
		},
	}
}

// Styles for help overlay
var (
	helpOverlayStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary)

	helpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	helpSectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				MarginTop(1)

	helpKeyColumnStyle = lipgloss.NewStyle().
				Width(16).
				Foreground(colorPrimary)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)

// RenderHelp renders the help overlay for the given sections.
func RenderHelp(sections []helpSection, width, height int) string {
	var b strings.Builder

	b.WriteString(helpTitleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n")

	for _, section := range sections {
		b.WriteString(helpSectionTitleStyle.Render(section.title))
		b.WriteString("\n")

		for _, entry := range section.entries {
			line := helpKeyColumnStyle.Render(entry.key) + helpDescStyle.Render(entry.desc)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press ? to close"))

	content := b.String()

	// Calculate overlay dimensions
	overlayWidth := 50
	if overlayWidth > width-4 {
		overlayWidth = width - 4
	}

	styled := helpOverlayStyle.Width(overlayWidth).Render(content)

	// Center the overlay
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, styled)
}
