package commands

import (
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
)

func TestList(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Add some tickets
	Add([]string{"--title", "High priority", "--priority", "1"})
	Add([]string{"--title", "Low priority", "--priority", "3"})

	err := List([]string{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Test with status filter
	err = List([]string{"--status", "open"})
	if err != nil {
		t.Fatalf("List(--status open) error = %v", err)
	}

	// Close one ticket and test filter
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	tickets[0].Close()
	store.Update(tickets[0])
	store.Close()

	err = List([]string{"--status", "closed"})
	if err != nil {
		t.Fatalf("List(--status closed) error = %v", err)
	}
}

func TestList_InvalidStatus(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := List([]string{"--status", "invalid"})
	if err == nil {
		t.Error("List() expected error for invalid status")
	}
}

func TestList_StatusReady(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := List([]string{"--status", "ready"})
	if err == nil {
		t.Error("List() expected error for 'ready' status")
	}
	if !strings.Contains(err.Error(), "thicket ready") {
		t.Errorf("List() error should suggest 'thicket ready' command, got: %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := List([]string{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
}
