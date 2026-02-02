package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpModel provides a help overlay for displaying keyboard shortcuts.
type HelpModel struct {
	width   int
	height  int
	scrollY int
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

// ScrollUp scrolls the help content up by one line.
func (h *HelpModel) ScrollUp() {
	if h.scrollY > 0 {
		h.scrollY--
	}
}

// ScrollDown scrolls the help content down by one line.
func (h *HelpModel) ScrollDown(contentHeight, visibleHeight int) {
	maxScroll := contentHeight - visibleHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if h.scrollY < maxScroll {
		h.scrollY++
	}
}

// ResetScroll resets the scroll position to the top.
func (h *HelpModel) ResetScroll() {
	h.scrollY = 0
}

// ScrollY returns the current scroll position.
func (h *HelpModel) ScrollY() int {
	return h.scrollY
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
				{"0-5", "Set priority directly"},
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
				{"0-5", "Set priority directly"},
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
	return RenderHelpWithScroll(sections, width, height, 0)
}

// RenderHelpWithScroll renders the help overlay with scroll support.
func RenderHelpWithScroll(sections []helpSection, width, height, scrollY int) string {
	// Build all content lines
	var lines []string

	lines = append(lines, helpTitleStyle.Render("Keyboard Shortcuts"))

	for _, section := range sections {
		lines = append(lines, helpSectionTitleStyle.Render(section.title))

		for _, entry := range section.entries {
			line := helpKeyColumnStyle.Render(entry.key) + helpDescStyle.Render(entry.desc)
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")

	// Use full width with small margin for border
	overlayWidth := width - 4
	if overlayWidth < 20 {
		overlayWidth = 20
	}

	// Calculate available height inside the overlay (account for border and padding)
	// Border: 2 lines (top + bottom), Padding: 2 lines (top + bottom from Padding(1, 2))
	overlayPadding := 4
	availableHeight := height - overlayPadding
	if availableHeight < 3 {
		availableHeight = 3 // Minimum to show something
	}

	// Calculate content height needed
	contentHeight := len(lines) + 1 // +1 for footer

	// Determine if scrolling is needed
	needsScroll := contentHeight > availableHeight

	// Clamp scroll position
	maxScroll := contentHeight - availableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollY > maxScroll {
		scrollY = maxScroll
	}
	if scrollY < 0 {
		scrollY = 0
	}

	// Apply scroll and limit visible lines
	var visibleLines []string
	visibleContentLines := availableHeight - 1 // -1 for footer line
	if visibleContentLines < 1 {
		visibleContentLines = 1
	}

	if needsScroll {
		end := scrollY + visibleContentLines
		if end > len(lines) {
			end = len(lines)
		}
		start := scrollY
		if start < 0 {
			start = 0
		}
		visibleLines = lines[start:end]
	} else {
		visibleLines = lines
	}

	var b strings.Builder
	for _, line := range visibleLines {
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Footer with scroll indicator
	if needsScroll {
		scrollIndicator := helpStyle.Render(fmt.Sprintf("↑↓ scroll • %d/%d • ? to close", scrollY+1, maxScroll+1))
		b.WriteString(scrollIndicator)
	} else {
		b.WriteString(helpStyle.Render("Press ? to close"))
	}

	content := b.String()

	// Render the styled overlay
	styled := helpOverlayStyle.Width(overlayWidth).Render(content)

	// Center horizontally by adding padding
	styledLines := strings.Split(styled, "\n")
	styledWidth := lipgloss.Width(styled)
	leftPad := (width - styledWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)

	var result strings.Builder
	for i, line := range styledLines {
		if i >= height {
			break // Don't exceed terminal height
		}
		result.WriteString(padding)
		result.WriteString(line)
		if i < len(styledLines)-1 && i < height-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// HelpContentHeight returns the total number of lines in the help content.
func HelpContentHeight(sections []helpSection) int {
	count := 1 // Title
	for _, section := range sections {
		count++ // Section title
		count += len(section.entries)
	}
	count++ // Empty line before footer
	count++ // Footer
	return count
}
