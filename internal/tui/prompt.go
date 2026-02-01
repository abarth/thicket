package tui

import tea "github.com/charmbracelet/bubbletea"

// PromptModel provides a reusable yes/no confirmation prompt.
// It captures keyboard input when active and invokes callbacks on confirm/cancel.
type PromptModel struct {
	active  bool
	message string
	context string // Optional context data (e.g., ticket ID)
}

// NewPrompt creates a new prompt model.
func NewPrompt() PromptModel {
	return PromptModel{}
}

// Show activates the prompt with the given message.
// The context parameter can store additional data (e.g., ticket ID).
func (p *PromptModel) Show(message string, context string) {
	p.active = true
	p.message = message
	p.context = context
}

// Hide deactivates the prompt.
func (p *PromptModel) Hide() {
	p.active = false
	p.message = ""
	p.context = ""
}

// Active returns whether the prompt is currently displayed.
func (p PromptModel) Active() bool {
	return p.active
}

// Context returns the context data set when showing the prompt.
func (p PromptModel) Context() string {
	return p.context
}

// PromptResult represents the user's response to a prompt.
type PromptResult int

const (
	PromptPending PromptResult = iota
	PromptConfirmed
	PromptCancelled
)

// Update handles keyboard input for the prompt.
// Returns the result of the user's action.
func (p *PromptModel) Update(msg tea.Msg) PromptResult {
	if !p.active {
		return PromptPending
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "y", "Y", "enter":
			p.Hide()
			return PromptConfirmed
		case "n", "N", "esc":
			p.Hide()
			return PromptCancelled
		}
	}
	return PromptPending
}

// View renders the prompt if active.
// Returns an empty string if the prompt is not active.
func (p PromptModel) View() string {
	if !p.active {
		return ""
	}
	return promptStyle.Render(p.message)
}
