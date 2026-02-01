package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectorModel_InitialState(t *testing.T) {
	s := NewSelector([]string{"a", "b", "c"})

	if s.Value() != "a" {
		t.Errorf("expected initial value 'a', got %q", s.Value())
	}
	if s.Focused() {
		t.Error("expected new selector to be unfocused")
	}
}

func TestSelectorModel_EmptyOptions(t *testing.T) {
	s := NewSelector([]string{})

	if s.Value() != "" {
		t.Errorf("expected empty value for empty options, got %q", s.Value())
	}
	if s.View() != "" {
		t.Error("expected empty view for empty options")
	}
}

func TestSelectorModel_SetValue(t *testing.T) {
	s := NewSelector([]string{"open", "closed", "icebox"})

	s.SetValue("closed")
	if s.Value() != "closed" {
		t.Errorf("expected value 'closed', got %q", s.Value())
	}

	s.SetValue("icebox")
	if s.Value() != "icebox" {
		t.Errorf("expected value 'icebox', got %q", s.Value())
	}

	// Setting invalid value should not change selection
	s.SetValue("invalid")
	if s.Value() != "icebox" {
		t.Errorf("expected value to remain 'icebox', got %q", s.Value())
	}
}

func TestSelectorModel_FocusBlur(t *testing.T) {
	s := NewSelector([]string{"a", "b"})

	s.Focus()
	if !s.Focused() {
		t.Error("expected selector to be focused after Focus()")
	}

	s.Blur()
	if s.Focused() {
		t.Error("expected selector to be unfocused after Blur()")
	}
}

func TestSelectorModel_Navigation(t *testing.T) {
	s := NewSelector([]string{"a", "b", "c"})
	s.Focus()

	// Move right
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s.Value() != "b" {
		t.Errorf("expected 'b' after right, got %q", s.Value())
	}

	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s.Value() != "c" {
		t.Errorf("expected 'c' after right, got %q", s.Value())
	}

	// Wrap around right
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s.Value() != "a" {
		t.Errorf("expected 'a' after wrap, got %q", s.Value())
	}

	// Move left
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if s.Value() != "c" {
		t.Errorf("expected 'c' after left wrap, got %q", s.Value())
	}

	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if s.Value() != "b" {
		t.Errorf("expected 'b' after left, got %q", s.Value())
	}
}

func TestSelectorModel_VimKeys(t *testing.T) {
	s := NewSelector([]string{"a", "b", "c"})
	s.Focus()

	// l for right
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if s.Value() != "b" {
		t.Errorf("expected 'b' after 'l', got %q", s.Value())
	}

	// h for left
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if s.Value() != "a" {
		t.Errorf("expected 'a' after 'h', got %q", s.Value())
	}
}

func TestSelectorModel_IgnoresInputWhenUnfocused(t *testing.T) {
	s := NewSelector([]string{"a", "b", "c"})
	// Don't focus

	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s.Value() != "a" {
		t.Errorf("expected value to remain 'a' when unfocused, got %q", s.Value())
	}
}

func TestSelectorModel_View(t *testing.T) {
	s := NewSelector([]string{"a", "b"})

	// View should contain both options
	view := s.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}
