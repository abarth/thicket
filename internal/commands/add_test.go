package commands

import (
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
)

func TestAdd(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Add([]string{"--title", "Test ticket", "--description", "A description", "--priority", "1"})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify ticket was created
	paths := config.GetPaths(dir)
	store, err := storage.Open(paths)
	if err != nil {
		t.Fatalf("storage.Open() error = %v", err)
	}
	defer store.Close()

	tickets, err := store.List(nil)
	if err != nil {
		t.Fatalf("store.List() error = %v", err)
	}
	if len(tickets) != 1 {
		t.Fatalf("Expected 1 ticket, got %d", len(tickets))
	}
	if tickets[0].Title != "Test ticket" {
		t.Errorf("Title = %q, want 'Test ticket'", tickets[0].Title)
	}
	if tickets[0].Priority != 1 {
		t.Errorf("Priority = %d, want 1", tickets[0].Priority)
	}
}

func TestAdd_MissingTitle(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Add([]string{"--description", "No title"})
	if err == nil {
		t.Error("Add() expected error for missing --title")
	}
}

func TestAdd_NotInitialized(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	err := Add([]string{"--title", "Test"})
	if err == nil {
		t.Error("Add() expected error for not initialized")
	}
}

func TestAdd_Links(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Create initial ticket T1
	if err := Add([]string{"--title", "T1"}); err != nil {
		t.Fatalf("Add(T1) error = %v", err)
	}

	// Get T1 ID
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	t1ID := tickets[0].ID
	store.Close()

	// 1. New ticket T2 blocks T1
	if err := Add([]string{"--title", "T2", "--blocks", t1ID}); err != nil {
		t.Fatalf("Add(T2 --blocks T1) error = %v", err)
	}

	// Verify T1 is blocked by T2
	store, _ = storage.Open(paths)
	tickets, _ = store.List(nil)
	var t2ID string
	for _, tk := range tickets {
		if tk.Title == "T2" {
			t2ID = tk.ID
			break
		}
	}
	blockers, _ := store.GetBlockers(t1ID)
	found := false
	for _, b := range blockers {
		if b.ID == t2ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("T1 should be blocked by T2")
	}
	store.Close()

	// 2. New ticket T3 is blocked by T2
	if err := Add([]string{"--title", "T3", "--blocked-by", t2ID}); err != nil {
		t.Fatalf("Add(T3 --blocked-by T2) error = %v", err)
	}

	// Verify T3 is blocked by T2
	store, _ = storage.Open(paths)
	tickets, _ = store.List(nil)
	var t3ID string
	for _, tk := range tickets {
		if tk.Title == "T3" {
			t3ID = tk.ID
			break
		}
	}
	blockers, _ = store.GetBlockers(t3ID)
	found = false
	for _, b := range blockers {
		if b.ID == t2ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("T3 should be blocked by T2")
	}
	store.Close()

	// 3. New ticket T4 is created from T3
	if err := Add([]string{"--title", "T4", "--created-from", t3ID}); err != nil {
		t.Fatalf("Add(T4 --created-from T3) error = %v", err)
	}

	// Verify T4 created from T3
	store, _ = storage.Open(paths)
	tickets, _ = store.List(nil)
	var t4ID string
	for _, tk := range tickets {
		if tk.Title == "T4" {
			t4ID = tk.ID
			break
		}
	}
	parent, _ := store.GetCreatedFrom(t4ID)
	if parent == nil || parent.ID != t3ID {
		t.Errorf("T4 should be created from T3, got %v", parent)
	}
	store.Close()
}
