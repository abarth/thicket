package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPromptModel_InitialState(t *testing.T) {
	p := NewPrompt()
	if p.Active() {
		t.Error("expected new prompt to be inactive")
	}
	if p.Context() != "" {
		t.Error("expected new prompt to have empty context")
	}
	if p.View() != "" {
		t.Error("expected inactive prompt to render empty string")
	}
}

func TestPromptModel_Show(t *testing.T) {
	p := NewPrompt()
	p.Show("Close ticket?", "TH-abc123")

	if !p.Active() {
		t.Error("expected prompt to be active after Show")
	}
	if p.Context() != "TH-abc123" {
		t.Errorf("expected context to be 'TH-abc123', got %q", p.Context())
	}
	if p.View() == "" {
		t.Error("expected active prompt to render non-empty string")
	}
}

func TestPromptModel_Hide(t *testing.T) {
	p := NewPrompt()
	p.Show("Test?", "ctx")
	p.Hide()

	if p.Active() {
		t.Error("expected prompt to be inactive after Hide")
	}
	if p.Context() != "" {
		t.Error("expected context to be cleared after Hide")
	}
}

func TestPromptModel_Update_Confirm(t *testing.T) {
	tests := []struct {
		key      string
		expected PromptResult
	}{
		{"y", PromptConfirmed},
		{"Y", PromptConfirmed},
		{"enter", PromptConfirmed},
		{"n", PromptCancelled},
		{"N", PromptCancelled},
		{"esc", PromptCancelled},
		{"x", PromptPending},
		{"a", PromptPending},
	}

	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			p := NewPrompt()
			p.Show("Test?", "")

			result := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.key)})

			if result != tc.expected {
				t.Errorf("key %q: expected result %v, got %v", tc.key, tc.expected, result)
			}

			// Prompt should be hidden after confirm/cancel
			if tc.expected != PromptPending && p.Active() {
				t.Errorf("key %q: expected prompt to be hidden after decision", tc.key)
			}
		})
	}
}

func TestPromptModel_Update_EnterKey(t *testing.T) {
	p := NewPrompt()
	p.Show("Test?", "")

	result := p.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if result != PromptConfirmed {
		t.Errorf("expected enter key to confirm, got %v", result)
	}
}

func TestPromptModel_Update_EscKey(t *testing.T) {
	p := NewPrompt()
	p.Show("Test?", "")

	result := p.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if result != PromptCancelled {
		t.Errorf("expected esc key to cancel, got %v", result)
	}
}

func TestPromptModel_Update_Inactive(t *testing.T) {
	p := NewPrompt()
	// Don't show the prompt

	result := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if result != PromptPending {
		t.Error("expected inactive prompt to return PromptPending")
	}
}

func TestPromptModel_Update_NonKeyMsg(t *testing.T) {
	p := NewPrompt()
	p.Show("Test?", "")

	// Send a non-key message
	result := p.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if result != PromptPending {
		t.Error("expected non-key message to return PromptPending")
	}
	if !p.Active() {
		t.Error("expected prompt to remain active after non-key message")
	}
}
