package commands

import (
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
)

func TestComment(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	// Get the ticket ID
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Comment([]string{ticketID, "This is a comment"})
	if err != nil {
		t.Fatalf("Comment() error = %v", err)
	}

	// Verify comment was created
	store, _ = storage.Open(paths)
	comments, err := store.GetComments(ticketID)
	store.Close()

	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(comments))
	}
	if comments[0].Content != "This is a comment" {
		t.Errorf("Content = %q, want 'This is a comment'", comments[0].Content)
	}
}

func TestComment_MissingTicketID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Comment([]string{})
	if err == nil {
		t.Error("Comment() expected error for missing ticket ID")
	}
}

func TestComment_MissingContent(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Comment([]string{ticketID})
	if err == nil {
		t.Error("Comment() expected error for missing content")
	}
}

func TestComment_EmptyContent(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Comment([]string{ticketID, "   "})
	if err == nil {
		t.Error("Comment() expected error for empty content")
	}
}

func TestComment_InvalidTicketID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Comment([]string{"invalid", "Comment"})
	if err == nil {
		t.Error("Comment() expected error for invalid ticket ID")
	}
}

func TestComment_TicketNotFound(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Comment([]string{"TH-999999", "Comment"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Comment() error = %v, want error containing 'not found'", err)
	}
}
