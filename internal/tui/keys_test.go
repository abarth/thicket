package tui

import (
	"strings"
	"testing"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Check some key bindings are defined
	if len(km.Quit.Keys()) == 0 {
		t.Error("expected Quit key binding to have keys")
	}
	if len(km.Help.Keys()) == 0 {
		t.Error("expected Help key binding to have keys")
	}
	if len(km.Up.Keys()) == 0 {
		t.Error("expected Up key binding to have keys")
	}
	if len(km.Down.Keys()) == 0 {
		t.Error("expected Down key binding to have keys")
	}
	if len(km.Enter.Keys()) == 0 {
		t.Error("expected Enter key binding to have keys")
	}
}

func TestListHelp(t *testing.T) {
	help := ListHelp()

	if help == "" {
		t.Error("expected non-empty help string")
	}

	// Check for expected key descriptions
	expectedKeys := []string{"j/k", "enter", "n", "e", "/", "?", "q"}
	for _, key := range expectedKeys {
		if !strings.Contains(help, key) {
			t.Errorf("expected ListHelp to contain '%s'", key)
		}
	}
}

func TestDetailHelp(t *testing.T) {
	help := DetailHelp()

	if help == "" {
		t.Error("expected non-empty help string")
	}

	// Check for expected key descriptions
	expectedKeys := []string{"esc", "e", "c", "m", "j/k", "?"}
	for _, key := range expectedKeys {
		if !strings.Contains(help, key) {
			t.Errorf("expected DetailHelp to contain '%s'", key)
		}
	}
}

func TestFormHelp(t *testing.T) {
	help := FormHelp()

	if help == "" {
		t.Error("expected non-empty help string")
	}

	// Check for expected key descriptions
	expectedKeys := []string{"tab", "ctrl+s", "esc", "?"}
	for _, key := range expectedKeys {
		if !strings.Contains(help, key) {
			t.Errorf("expected FormHelp to contain '%s'", key)
		}
	}
}
