package commands

import (
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

func TestClose(t *testing.T) {
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

	err := Close([]string{ticketID})
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	store, _ = storage.Open(paths)
	tk, _ := store.Get(ticketID)
	store.Close()

	if tk.Status != ticket.StatusClosed {
		t.Errorf("Status = %q, want closed", tk.Status)
	}
}

func TestClose_AlreadyClosed(t *testing.T) {
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

	Close([]string{ticketID})

	// Close again should not error
	err := Close([]string{ticketID})
	if err != nil {
		t.Fatalf("Close() error = %v (should not error for already closed)", err)
	}
}

func TestClose_NotFound(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Close([]string{"TH-999999"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Close() error = %v, want error containing 'not found'", err)
	}
}
