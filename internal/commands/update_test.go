package commands

import (
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

func TestUpdate(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Original", "--priority", "1"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Update([]string{"--title", "Updated", "--priority", "2", ticketID})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	store, _ = storage.Open(paths)
	tk, _ := store.Get(ticketID)
	store.Close()

	if tk.Title != "Updated" {
		t.Errorf("Title = %q, want 'Updated'", tk.Title)
	}
	if tk.Priority != 2 {
		t.Errorf("Priority = %d, want 2", tk.Priority)
	}
}

func TestUpdate_Status(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Update([]string{"--status", "closed", ticketID})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	store, _ = storage.Open(paths)
	tk, _ := store.Get(ticketID)
	store.Close()

	if tk.Status != ticket.StatusClosed {
		t.Errorf("Status = %q, want closed", tk.Status)
	}
}

func TestUpdate_NoFields(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Update([]string{ticketID}) // Just ID, no flags
	if err == nil {
		t.Error("Update() expected error when no fields specified")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Update([]string{"--title", "Test", "TH-999999"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Update() error = %v, want error containing 'not found'", err)
	}
}
