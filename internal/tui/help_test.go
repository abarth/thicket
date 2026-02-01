package tui

import (
	"strings"
	"testing"
)

func TestListFullHelp(t *testing.T) {
	sections := ListFullHelp()

	if len(sections) == 0 {
		t.Error("expected non-empty sections for list help")
	}

	// Check that common sections exist
	sectionTitles := make(map[string]bool)
	for _, s := range sections {
		sectionTitles[s.title] = true
	}

	expectedSections := []string{"Navigation", "Actions", "Filtering", "General"}
	for _, expected := range expectedSections {
		if !sectionTitles[expected] {
			t.Errorf("expected section %q in list help", expected)
		}
	}
}

func TestDetailFullHelp(t *testing.T) {
	sections := DetailFullHelp()

	if len(sections) == 0 {
		t.Error("expected non-empty sections for detail help")
	}

	// Check that navigation section has back entry
	hasBack := false
	for _, s := range sections {
		for _, e := range s.entries {
			if strings.Contains(e.desc, "Back") {
				hasBack = true
				break
			}
		}
	}

	if !hasBack {
		t.Error("expected detail help to have a 'back' entry")
	}
}

func TestFormFullHelp(t *testing.T) {
	sections := FormFullHelp()

	if len(sections) == 0 {
		t.Error("expected non-empty sections for form help")
	}

	// Check that save and cancel entries exist
	hasSave := false
	hasCancel := false
	for _, s := range sections {
		for _, e := range s.entries {
			if strings.Contains(e.desc, "Save") {
				hasSave = true
			}
			if strings.Contains(e.desc, "Cancel") {
				hasCancel = true
			}
		}
	}

	if !hasSave {
		t.Error("expected form help to have a 'save' entry")
	}
	if !hasCancel {
		t.Error("expected form help to have a 'cancel' entry")
	}
}

func TestRenderHelp(t *testing.T) {
	sections := []helpSection{
		{
			title: "Test",
			entries: []helpEntry{
				{"a", "do something"},
				{"b", "do something else"},
			},
		},
	}

	result := RenderHelp(sections, 80, 24)

	if result == "" {
		t.Error("expected non-empty rendered help")
	}

	// Check that title and entries are included
	if !strings.Contains(result, "Keyboard Shortcuts") {
		t.Error("expected help to contain title")
	}
	if !strings.Contains(result, "Test") {
		t.Error("expected help to contain section title")
	}
	if !strings.Contains(result, "do something") {
		t.Error("expected help to contain entry description")
	}
}

func TestHelpModel_SetSize(t *testing.T) {
	h := NewHelp()
	h.SetSize(100, 50)

	if h.width != 100 {
		t.Errorf("expected width 100, got %d", h.width)
	}
	if h.height != 50 {
		t.Errorf("expected height 50, got %d", h.height)
	}
}
