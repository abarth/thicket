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

func TestHelpModel_Scroll(t *testing.T) {
	h := NewHelp()

	// Initial scroll position is 0
	if h.ScrollY() != 0 {
		t.Errorf("expected initial scrollY 0, got %d", h.ScrollY())
	}

	// ScrollUp at 0 stays at 0
	h.ScrollUp()
	if h.ScrollY() != 0 {
		t.Errorf("expected scrollY 0 after ScrollUp at 0, got %d", h.ScrollY())
	}

	// ScrollDown increases scroll
	h.ScrollDown(20, 10) // content 20 lines, visible 10
	if h.ScrollY() != 1 {
		t.Errorf("expected scrollY 1 after ScrollDown, got %d", h.ScrollY())
	}

	// ScrollUp decreases scroll
	h.ScrollUp()
	if h.ScrollY() != 0 {
		t.Errorf("expected scrollY 0 after ScrollUp, got %d", h.ScrollY())
	}

	// ResetScroll resets to 0
	h.ScrollDown(20, 10)
	h.ScrollDown(20, 10)
	h.ResetScroll()
	if h.ScrollY() != 0 {
		t.Errorf("expected scrollY 0 after ResetScroll, got %d", h.ScrollY())
	}
}

func TestHelpModel_ScrollDown_MaxScroll(t *testing.T) {
	h := NewHelp()

	// Content of 15 lines, visible 10 = maxScroll of 5
	contentHeight := 15
	visibleHeight := 10
	maxScroll := contentHeight - visibleHeight // 5

	// Scroll all the way down
	for i := 0; i < maxScroll+5; i++ {
		h.ScrollDown(contentHeight, visibleHeight)
	}

	// Should stop at maxScroll
	if h.ScrollY() != maxScroll {
		t.Errorf("expected scrollY %d at max, got %d", maxScroll, h.ScrollY())
	}
}

func TestHelpModel_ScrollDown_ContentFits(t *testing.T) {
	h := NewHelp()

	// Content of 5 lines, visible 10 = no scroll needed
	h.ScrollDown(5, 10)
	if h.ScrollY() != 0 {
		t.Errorf("expected scrollY 0 when content fits, got %d", h.ScrollY())
	}
}

func TestRenderHelpWithScroll(t *testing.T) {
	sections := []helpSection{
		{
			title: "Test",
			entries: []helpEntry{
				{"a", "do something"},
				{"b", "do something else"},
			},
		},
	}

	// Render with no scroll
	result := RenderHelpWithScroll(sections, 80, 24, 0)
	if result == "" {
		t.Error("expected non-empty rendered help")
	}
	if !strings.Contains(result, "Keyboard Shortcuts") {
		t.Error("expected help to contain title")
	}

	// Render with scroll offset
	result = RenderHelpWithScroll(sections, 80, 24, 1)
	if result == "" {
		t.Error("expected non-empty rendered help with scroll")
	}
}

func TestRenderHelpWithScroll_ShortTerminal(t *testing.T) {
	// Create sections that would exceed a short terminal
	sections := ListFullHelp() // This has many entries

	// Render in a very short terminal (height 10)
	result := RenderHelpWithScroll(sections, 80, 10, 0)
	if result == "" {
		t.Error("expected non-empty rendered help in short terminal")
	}

	// Should contain scroll indicator when content exceeds height
	if !strings.Contains(result, "↑↓") {
		t.Error("expected scroll indicator in short terminal")
	}
}

func TestHelpContentHeight(t *testing.T) {
	sections := []helpSection{
		{
			title: "Section 1",
			entries: []helpEntry{
				{"a", "entry 1"},
				{"b", "entry 2"},
			},
		},
		{
			title: "Section 2",
			entries: []helpEntry{
				{"c", "entry 3"},
			},
		},
	}

	// Expected: 1 title + 1 section1 title + 2 entries + 1 section2 title + 1 entry + 1 empty + 1 footer = 8
	height := HelpContentHeight(sections)
	if height != 8 {
		t.Errorf("expected content height 8, got %d", height)
	}
}

func TestHelpLayout_FitsInTerminal(t *testing.T) {
	sections := ListFullHelp()
	termWidth := 80
	termHeight := 24

	result := RenderHelpWithScroll(sections, termWidth, termHeight, 0)
	lines := strings.Split(result, "\n")

	// The output should not exceed terminal height
	// Note: the last split element might be empty if string ends with \n
	lineCount := len(lines)
	if lines[len(lines)-1] == "" {
		lineCount--
	}

	if lineCount > termHeight {
		t.Errorf("help output has %d lines, exceeds terminal height %d", lineCount, termHeight)
	}

	// The first line should be part of the border or content, not empty padding
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		t.Errorf("first line is empty - content may be offset from top")
	}
}

func TestHelpLayout_TitleVisible(t *testing.T) {
	sections := []helpSection{
		{
			title: "Navigation",
			entries: []helpEntry{
				{"j", "Move down"},
			},
		},
	}
	termWidth := 80
	termHeight := 15

	result := RenderHelpWithScroll(sections, termWidth, termHeight, 0)

	// The title "Keyboard Shortcuts" should be visible in output
	if !strings.Contains(result, "Keyboard Shortcuts") {
		t.Errorf("title 'Keyboard Shortcuts' not found in help output")
	}

	// The section title should also be visible
	if !strings.Contains(result, "Navigation") {
		t.Errorf("section title 'Navigation' not found in help output")
	}
}

func TestHelpLayout_VeryShortTerminal(t *testing.T) {
	sections := ListFullHelp()
	termWidth := 80
	termHeight := 10 // Very short terminal

	result := RenderHelpWithScroll(sections, termWidth, termHeight, 0)
	lines := strings.Split(result, "\n")

	// Count non-empty lines
	lineCount := 0
	for _, line := range lines {
		if line != "" {
			lineCount++
		}
	}

	// Output should fit within terminal height
	if lineCount > termHeight {
		t.Errorf("help output has %d non-empty lines, exceeds terminal height %d", lineCount, termHeight)
	}

	// First few lines should contain the border and title
	topContent := strings.Join(lines[:min(5, len(lines))], "\n")
	if !strings.Contains(topContent, "Keyboard") {
		t.Errorf("title not visible in top of output for short terminal.\nTop 5 lines:\n%s", topContent)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
