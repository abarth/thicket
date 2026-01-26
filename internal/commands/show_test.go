package commands

import (
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
)

func TestShow(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket", "--description", "Description"})

	// Get the ticket ID
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	store.Close()

	err := Show([]string{tickets[0].ID})
	if err != nil {
		t.Fatalf("Show() error = %v", err)
	}
}

func TestShow_LowercaseID(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	store.Close()

	// Use lowercase ID
	lowercaseID := strings.ToLower(tickets[0].ID)
	err := Show([]string{lowercaseID})
	if err != nil {
		t.Fatalf("Show() error = %v (should accept lowercase)", err)
	}
}

func TestShow_NotFound(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Show([]string{"TH-999999"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Show() error = %v, want error containing 'not found'", err)
	}
}

func TestShow_InvalidID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Show([]string{"invalid"})
	if err == nil {
		t.Error("Show() expected error for invalid ID")
	}
}

func TestShow_MissingID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Show([]string{})
	if err == nil {
		t.Error("Show() expected error for missing ID")
	}
}

func TestShow_WithComments(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket", "--description", "Description"})

	// Get the ticket ID
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	// Add a comment
	Comment([]string{ticketID, "A test comment"})

	// Show should not error (output includes comments)
	err := Show([]string{ticketID})
	if err != nil {
		t.Fatalf("Show() error = %v", err)
	}
}
