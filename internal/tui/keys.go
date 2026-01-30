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
	Enter        key.Binding
	Back         key.Binding
	New          key.Binding
	Edit         key.Binding
	Close        key.Binding
	Comment      key.Binding
	Search       key.Binding
	Refresh      key.Binding
	PriorityUp   key.Binding
	PriorityDown key.Binding

	// Filtering
	FilterOpen   key.Binding
	FilterClosed key.Binding
	FilterIcebox key.Binding
	FilterAll    key.Binding

	// Type settings
	SetBug     key.Binding
	SetFeature key.Binding
	SetTask    key.Binding
	SetEpic    key.Binding
	SetCleanup key.Binding

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
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		PriorityUp: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "higher priority"),
		),
		PriorityDown: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "lower priority"),
		),
		FilterOpen: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open only"),
		),
		FilterClosed: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "closed only"),
		),
		FilterIcebox: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "icebox only"),
		),
		FilterAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "show all"),
		),
		SetBug: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "set bug"),
		),
		SetFeature: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "set feature"),
		),
		SetTask: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "set task"),
		),
		SetEpic: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "set epic"),
		),
		SetCleanup: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "set cleanup"),
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
		helpKeyStyle.Render("+/-") + helpStyle.Render(" prio  ") +
		helpKeyStyle.Render("b/f/t/E/C") + helpStyle.Render(" type  ") +
		helpKeyStyle.Render("o/x/i/a") + helpStyle.Render(" filter  ") +
		helpKeyStyle.Render("/") + helpStyle.Render(" search  ") +
		helpKeyStyle.Render("q") + helpStyle.Render(" quit")
}

// DetailHelp returns help text for detail view.
func DetailHelp() string {
	return helpKeyStyle.Render("esc") + helpStyle.Render(" back  ") +
		helpKeyStyle.Render("e") + helpStyle.Render(" edit  ") +
		helpKeyStyle.Render("c") + helpStyle.Render(" close  ") +
		helpKeyStyle.Render("m") + helpStyle.Render(" comment  ") +
		helpKeyStyle.Render("b/f/t/E/C") + helpStyle.Render(" type  ") +
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
