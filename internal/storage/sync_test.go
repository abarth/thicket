package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/ticket"
)

func setupTestProject(t *testing.T) (config.Paths, func()) {
	t.Helper()
	dir := t.TempDir()

	// Create .thicket directory
	thicketDir := filepath.Join(dir, ".thicket")
	if err := os.MkdirAll(thicketDir, 0755); err != nil {
		t.Fatalf("Failed to create .thicket directory: %v", err)
	}

	// Create empty tickets file
	ticketsPath := filepath.Join(thicketDir, "tickets.jsonl")
	if err := os.WriteFile(ticketsPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create tickets file: %v", err)
	}

	paths := config.GetPaths(dir)
	return paths, func() {}
}

func TestStore_Open(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
}

func TestStore_AddAndGet(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, err := ticket.New("TH", "Test ticket", "Description", 1)
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}

	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil")
	}

	if got.ID != tk.ID {
		t.Errorf("Get().ID = %q, want %q", got.ID, tk.ID)
	}
	if got.Title != tk.Title {
		t.Errorf("Get().Title = %q, want %q", got.Title, tk.Title)
	}
}

func TestStore_Update(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, err := ticket.New("TH", "Original title", "Description", 1)
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}

	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	newTitle := "Updated title"
	if err := tk.Update(&newTitle, nil, nil, nil); err != nil {
		t.Fatalf("ticket.Update() error = %v", err)
	}

	if err := store.Update(tk); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Title != "Updated title" {
		t.Errorf("Get().Title = %q, want 'Updated title'", got.Title)
	}
}

func TestStore_List(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Add some tickets
	for i := 0; i < 3; i++ {
		tk, err := ticket.New("TH", "Ticket", "", i)
		if err != nil {
			t.Fatalf("ticket.New() error = %v", err)
		}
		if err := store.Add(tk); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	// List all
	all, err := store.List(nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 3 {
		t.Errorf("List() returned %d tickets, want 3", len(all))
	}

	// List open only
	open := ticket.StatusOpen
	openTickets, err := store.List(&open)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(openTickets) != 3 {
		t.Errorf("List(open) returned %d tickets, want 3", len(openTickets))
	}
}

func TestStore_SyncFromJSONL(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	// Pre-populate JSONL with tickets
	now := time.Now().UTC()
	tickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "First", Status: ticket.StatusOpen, Priority: 1, Created: now, Updated: now},
		{ID: "TH-222222", Title: "Second", Status: ticket.StatusOpen, Priority: 2, Created: now, Updated: now},
	}
	if err := WriteJSONL(paths.Tickets, tickets); err != nil {
		t.Fatalf("WriteJSONL() error = %v", err)
	}

	// Open store - should sync from JSONL
	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Verify tickets were loaded
	all, err := store.List(nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("List() returned %d tickets, want 2", len(all))
	}

	got, err := store.Get("TH-111111")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil || got.Title != "First" {
		t.Errorf("Get() = %+v, want ticket with Title=First", got)
	}
}

func TestStore_SyncOnReopen(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	// Create initial store and add a ticket
	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tk, _ := ticket.New("TH", "Initial", "", 1)
	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	store.Close()

	// Externally modify the JSONL file
	now := time.Now().UTC()
	externalTickets := []*ticket.Ticket{
		{ID: "TH-external", Title: "External", Status: ticket.StatusOpen, Priority: 0, Created: now, Updated: now},
	}
	if err := WriteJSONL(paths.Tickets, externalTickets); err != nil {
		t.Fatalf("WriteJSONL() error = %v", err)
	}

	// Reopen store - should sync from modified JSONL
	store2, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store2.Close()

	// Should only have the external ticket
	all, err := store2.List(nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("List() returned %d tickets, want 1", len(all))
	}
	if all[0].ID != "TH-external" {
		t.Errorf("List()[0].ID = %q, want TH-external", all[0].ID)
	}
}
