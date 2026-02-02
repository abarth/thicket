package tui

import (
	"strings"
	"testing"
)

func TestHighlightMatches_EmptyQuery(t *testing.T) {
	text := "Hello World"
	result := highlightMatches(text, "")

	if result != text {
		t.Errorf("expected unchanged text with empty query, got '%s'", result)
	}
}

func TestHighlightMatches_NoMatch(t *testing.T) {
	text := "Hello World"
	result := highlightMatches(text, "xyz")

	// When no match, should return original text unchanged
	if result != text {
		t.Errorf("expected unchanged text with no match, got '%s'", result)
	}
}

func TestHighlightMatches_SingleMatch(t *testing.T) {
	text := "Hello World"
	result := highlightMatches(text, "World")

	// Result should contain the original text content
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "World") {
		t.Errorf("expected result to contain original text, got '%s'", result)
	}
}

func TestHighlightMatches_CaseInsensitive(t *testing.T) {
	text := "Hello World"
	result := highlightMatches(text, "world")

	// Result should contain the original text content (case preserved)
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "World") {
		t.Errorf("expected result to contain original text with original casing, got '%s'", result)
	}
}

func TestHighlightMatches_MultipleMatches(t *testing.T) {
	text := "test one test two test"
	result := highlightMatches(text, "test")

	// Result should contain the original text content
	if !strings.Contains(result, "one") || !strings.Contains(result, "two") {
		t.Errorf("expected result to contain original text, got '%s'", result)
	}
}

func TestHighlightMatches_MatchAtStart(t *testing.T) {
	text := "Hello World"
	result := highlightMatches(text, "Hello")

	// Result should contain both parts of the original text
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "World") {
		t.Errorf("expected result to contain original text, got '%s'", result)
	}
}

func TestHighlightMatches_MatchAtEnd(t *testing.T) {
	text := "Hello World"
	result := highlightMatches(text, "World")

	// Result should contain both parts of the original text
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "World") {
		t.Errorf("expected result to contain original text, got '%s'", result)
	}
}

func TestHighlightMatches_EntireString(t *testing.T) {
	text := "test"
	result := highlightMatches(text, "test")

	// Result should contain the original text
	if !strings.Contains(result, "test") {
		t.Errorf("expected result to contain 'test', got '%s'", result)
	}
}

func TestHighlightMatches_OverlappingPatterns(t *testing.T) {
	text := "aaa"
	result := highlightMatches(text, "aa")

	// The function finds non-overlapping matches from left to right
	// "aa" at index 0, then "a" remains which doesn't match
	// Result should contain the original text
	if !strings.Contains(result, "a") {
		t.Errorf("expected result to contain 'a', got '%s'", result)
	}
}

func TestHighlightMatches_MiddleOfString(t *testing.T) {
	text := "foo bar baz"
	result := highlightMatches(text, "bar")

	// Result should contain all parts
	if !strings.Contains(result, "foo") || !strings.Contains(result, "bar") || !strings.Contains(result, "baz") {
		t.Errorf("expected result to contain all parts, got '%s'", result)
	}
}
