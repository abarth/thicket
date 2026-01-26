package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/abarth/thicket/internal/ticket"
)

func TestOpenDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()
}

func TestDB_Metadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	// Get non-existent key
	val, err := db.GetMetadata("nonexistent")
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}
	if val != "" {
		t.Errorf("GetMetadata() = %q, want empty", val)
	}

	// Set and get
	if err := db.SetMetadata("key1", "value1"); err != nil {
		t.Fatalf("SetMetadata() error = %v", err)
	}

	val, err = db.GetMetadata("key1")
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}
	if val != "value1" {
		t.Errorf("GetMetadata() = %q, want value1", val)
	}

	// Update existing key
	if err := db.SetMetadata("key1", "value2"); err != nil {
		t.Fatalf("SetMetadata() error = %v", err)
	}

	val, err = db.GetMetadata("key1")
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}
	if val != "value2" {
		t.Errorf("GetMetadata() = %q, want value2", val)
	}
}

func TestDB_InsertAndGetTicket(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	tk := &ticket.Ticket{
		ID:          "TH-111111",
		Title:       "Test ticket",
		Description: "A description",
		Status:      ticket.StatusOpen,
		Priority:    1,
		Created:     now,
		Updated:     now,
	}

	if err := db.InsertTicket(tk); err != nil {
		t.Fatalf("InsertTicket() error = %v", err)
	}

	got, err := db.GetTicket("TH-111111")
	if err != nil {
		t.Fatalf("GetTicket() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetTicket() returned nil")
	}

	if got.ID != tk.ID {
		t.Errorf("GetTicket().ID = %q, want %q", got.ID, tk.ID)
	}
	if got.Title != tk.Title {
		t.Errorf("GetTicket().Title = %q, want %q", got.Title, tk.Title)
	}
	if got.Description != tk.Description {
		t.Errorf("GetTicket().Description = %q, want %q", got.Description, tk.Description)
	}
	if got.Status != tk.Status {
		t.Errorf("GetTicket().Status = %q, want %q", got.Status, tk.Status)
	}
	if got.Priority != tk.Priority {
		t.Errorf("GetTicket().Priority = %d, want %d", got.Priority, tk.Priority)
	}
}

func TestDB_GetTicket_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	got, err := db.GetTicket("TH-999999")
	if err != nil {
		t.Fatalf("GetTicket() error = %v", err)
	}
	if got != nil {
		t.Errorf("GetTicket() = %+v, want nil", got)
	}
}

func TestDB_UpdateTicket(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	tk := &ticket.Ticket{
		ID:       "TH-111111",
		Title:    "Original",
		Status:   ticket.StatusOpen,
		Priority: 1,
		Created:  now,
		Updated:  now,
	}

	if err := db.InsertTicket(tk); err != nil {
		t.Fatalf("InsertTicket() error = %v", err)
	}

	tk.Title = "Updated"
	tk.Status = ticket.StatusClosed
	tk.Updated = time.Now().UTC()

	if err := db.UpdateTicket(tk); err != nil {
		t.Fatalf("UpdateTicket() error = %v", err)
	}

	got, err := db.GetTicket("TH-111111")
	if err != nil {
		t.Fatalf("GetTicket() error = %v", err)
	}

	if got.Title != "Updated" {
		t.Errorf("GetTicket().Title = %q, want Updated", got.Title)
	}
	if got.Status != ticket.StatusClosed {
		t.Errorf("GetTicket().Status = %q, want closed", got.Status)
	}
}

func TestDB_UpdateTicket_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	tk := &ticket.Ticket{
		ID:      "TH-999999",
		Title:   "Test",
		Status:  ticket.StatusOpen,
		Updated: time.Now().UTC(),
	}

	err = db.UpdateTicket(tk)
	if err == nil {
		t.Error("UpdateTicket() expected error for non-existent ticket")
	}
}

func TestDB_ListTickets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	tickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "Low priority", Status: ticket.StatusOpen, Priority: 3, Created: now, Updated: now},
		{ID: "TH-222222", Title: "High priority", Status: ticket.StatusOpen, Priority: 1, Created: now, Updated: now},
		{ID: "TH-333333", Title: "Closed", Status: ticket.StatusClosed, Priority: 2, Created: now, Updated: now},
	}

	for _, tk := range tickets {
		if err := db.InsertTicket(tk); err != nil {
			t.Fatalf("InsertTicket() error = %v", err)
		}
	}

	// List all tickets (should be ordered by priority)
	all, err := db.ListTickets(nil)
	if err != nil {
		t.Fatalf("ListTickets() error = %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("ListTickets() returned %d tickets, want 3", len(all))
	}
	if all[0].ID != "TH-222222" {
		t.Errorf("ListTickets()[0].ID = %q, want TH-222222 (highest priority)", all[0].ID)
	}

	// List only open tickets
	open := ticket.StatusOpen
	openTickets, err := db.ListTickets(&open)
	if err != nil {
		t.Fatalf("ListTickets() error = %v", err)
	}
	if len(openTickets) != 2 {
		t.Fatalf("ListTickets(open) returned %d tickets, want 2", len(openTickets))
	}

	// List only closed tickets
	closed := ticket.StatusClosed
	closedTickets, err := db.ListTickets(&closed)
	if err != nil {
		t.Fatalf("ListTickets() error = %v", err)
	}
	if len(closedTickets) != 1 {
		t.Fatalf("ListTickets(closed) returned %d tickets, want 1", len(closedTickets))
	}
	if closedTickets[0].ID != "TH-333333" {
		t.Errorf("ListTickets(closed)[0].ID = %q, want TH-333333", closedTickets[0].ID)
	}
}

func TestDB_RebuildFromTickets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()

	// Insert initial ticket
	initial := &ticket.Ticket{ID: "TH-000000", Title: "Initial", Status: ticket.StatusOpen, Priority: 0, Created: now, Updated: now}
	if err := db.InsertTicket(initial); err != nil {
		t.Fatalf("InsertTicket() error = %v", err)
	}

	// Rebuild with different tickets
	newTickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "First", Status: ticket.StatusOpen, Priority: 1, Created: now, Updated: now},
		{ID: "TH-222222", Title: "Second", Status: ticket.StatusOpen, Priority: 2, Created: now, Updated: now},
	}

	if err := db.RebuildFromTickets(newTickets); err != nil {
		t.Fatalf("RebuildFromTickets() error = %v", err)
	}

	// Check that old ticket is gone
	old, err := db.GetTicket("TH-000000")
	if err != nil {
		t.Fatalf("GetTicket() error = %v", err)
	}
	if old != nil {
		t.Error("GetTicket() returned old ticket after rebuild")
	}

	// Check that new tickets are present
	all, err := db.GetAllTickets()
	if err != nil {
		t.Fatalf("GetAllTickets() error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("GetAllTickets() returned %d tickets, want 2", len(all))
	}
}
