package commands

import (
	"bytes"
	"os"
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

func TestClose_JSONOutputIncludesFollowOnHint(t *testing.T) {
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Close([]string{"--json", ticketID})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify the hint is included in JSON output
	if !strings.Contains(output, "hint") {
		t.Error("Close JSON output missing 'hint' field")
	}
	if !strings.Contains(output, "additional work") {
		t.Error("Close JSON output hint should mention 'additional work'")
	}
	if !strings.Contains(output, ticketID) {
		t.Errorf("Close JSON output hint should include ticket ID %s for --created-from", ticketID)
	}
}
