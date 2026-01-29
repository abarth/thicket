package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the TUI.
type KeyMap struct {
	// Global keys
	Quit key.Binding
	Help key.Binding

	// Navigation
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding

	// Actions
	Enter  key.Binding
	Back   key.Binding
	New    key.Binding
	Edit   key.Binding
	Close  key.Binding
	Comment key.Binding
	Refresh key.Binding

	// Priority
	PriorityUp   key.Binding
	PriorityDown key.Binding

	// Filtering
	FilterOpen   key.Binding
	FilterClosed key.Binding
	FilterAll    key.Binding

	// Form navigation
	NextField key.Binding
	PrevField key.Binding
	Save      key.Binding
	Cancel    key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", "l"),
			key.WithHelp("enter", "view"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace", "h"),
			key.WithHelp("esc", "back"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Close: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "close"),
		),
		Comment: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "comment"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		PriorityUp: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "lower priority"),
		),
		PriorityDown: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "higher priority"),
		),
		FilterOpen: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open only"),
		),
		FilterClosed: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "closed only"),
		),
		FilterAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "show all"),
		),
		NextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ListHelp returns help text for list view.
func ListHelp() string {
	return helpKeyStyle.Render("j/k") + helpStyle.Render(" nav  ") +
		helpKeyStyle.Render("enter") + helpStyle.Render(" view  ") +
		helpKeyStyle.Render("n") + helpStyle.Render(" new  ") +
		helpKeyStyle.Render("e") + helpStyle.Render(" edit  ") +
		helpKeyStyle.Render("c") + helpStyle.Render(" close  ") +
		helpKeyStyle.Render("+/-") + helpStyle.Render(" priority  ") +
		helpKeyStyle.Render("o/x/a") + helpStyle.Render(" filter  ") +
		helpKeyStyle.Render("q") + helpStyle.Render(" quit")
}

// DetailHelp returns help text for detail view.
func DetailHelp() string {
	return helpKeyStyle.Render("esc") + helpStyle.Render(" back  ") +
		helpKeyStyle.Render("e") + helpStyle.Render(" edit  ") +
		helpKeyStyle.Render("c") + helpStyle.Render(" close  ") +
		helpKeyStyle.Render("m") + helpStyle.Render(" comment  ") +
		helpKeyStyle.Render("j/k") + helpStyle.Render(" scroll  ") +
		helpKeyStyle.Render("q") + helpStyle.Render(" quit")
}

// FormHelp returns help text for form view.
func FormHelp() string {
	return helpKeyStyle.Render("tab") + helpStyle.Render(" next  ") +
		helpKeyStyle.Render("shift+tab") + helpStyle.Render(" prev  ") +
		helpKeyStyle.Render("ctrl+s") + helpStyle.Render(" save  ") +
		helpKeyStyle.Render("esc") + helpStyle.Render(" cancel")
}
