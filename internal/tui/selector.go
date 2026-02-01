package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectorModel provides a selection list for choosing from predefined options.
// Users cycle through options with left/right arrow keys when focused.
type SelectorModel struct {
	options  []string
	selected int
	focused  bool
}

// NewSelector creates a new selector with the given options.
// The first option is selected by default.
func NewSelector(options []string) SelectorModel {
	return SelectorModel{
		options:  options,
		selected: 0,
	}
}

// Focus sets the selector as focused.
func (s *SelectorModel) Focus() {
	s.focused = true
}

// Blur removes focus from the selector.
func (s *SelectorModel) Blur() {
	s.focused = false
}

// Focused returns whether the selector is focused.
func (s SelectorModel) Focused() bool {
	return s.focused
}

// Value returns the currently selected option.
func (s SelectorModel) Value() string {
	if len(s.options) == 0 {
		return ""
	}
	return s.options[s.selected]
}

// SetValue sets the selection to the option matching the given value.
// If the value is not found, the selection is unchanged.
func (s *SelectorModel) SetValue(value string) {
	for i, opt := range s.options {
		if opt == value {
			s.selected = i
			return
		}
	}
}

// Update handles keyboard input for the selector.
func (s SelectorModel) Update(msg tea.Msg) (SelectorModel, tea.Cmd) {
	if !s.focused {
		return s, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left", "h":
			s.prev()
		case "right", "l":
			s.next()
		}
	}

	return s, nil
}

func (s *SelectorModel) next() {
	s.selected = (s.selected + 1) % len(s.options)
}

func (s *SelectorModel) prev() {
	s.selected--
	if s.selected < 0 {
		s.selected = len(s.options) - 1
	}
}

// View renders the selector.
// Shows all options with the selected one highlighted.
func (s SelectorModel) View() string {
	if len(s.options) == 0 {
		return ""
	}

	var result string
	for i, opt := range s.options {
		if i > 0 {
			result += "  "
		}

		if i == s.selected {
			if s.focused {
				result += selectorFocusedStyle.Render(opt)
			} else {
				result += selectorSelectedStyle.Render(opt)
			}
		} else {
			result += selectorOptionStyle.Render(opt)
		}
	}

	return result
}

// Selector styles
var (
	selectorFocusedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("0")).
				Background(colorPrimary).
				Padding(0, 1)

	selectorSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				Padding(0, 1)

	selectorOptionStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Padding(0, 1)
)
