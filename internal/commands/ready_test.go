package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
)

func TestReady(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Add tickets
	Add([]string{"--title", "Blocker", "--priority", "1"})
	Add([]string{"--title", "Blocked", "--priority", "2"})
	Add([]string{"--title", "Independent", "--priority", "3"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)

	var blockerID, blockedID string
	for _, tk := range tickets {
		if tk.Title == "Blocker" {
			blockerID = tk.ID
		} else if tk.Title == "Blocked" {
			blockedID = tk.ID
		}
	}

	// Link them: Blocked is blocked by Blocker
	Link([]string{"--blocked-by", blockerID, blockedID})
	store.Close()

	err := Ready([]string{})
	if err != nil {
		t.Fatalf("Ready() error = %v", err)
	}

	// Test with JSON output to verify content
	// We need to capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = Ready([]string{"--json"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Ready(--json) error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// ready should now show only the highest priority unblocked ticket (Blocker at P1)
	// The JSON structure has "ticket": {...} for the primary ticket
	// Check that "Blocker" appears right after "ticket": { (indicating it's the primary ticket)
	if !strings.Contains(output, `"ticket"`) || !strings.Contains(output, `"title": "Blocker"`) {
		t.Errorf("Ready output missing 'Blocker' as primary ticket, got: %s", output)
	}
	// Independent is unblocked but lower priority, so not mentioned at all
	if strings.Contains(output, "Independent") {
		t.Error("Ready output should not contain 'Independent' anywhere (lower priority)")
	}
	// Blocked appears in the blocking array but should not be the primary ticket
	// Verify that the primary ticket is Blocker by checking the "ticket" structure
	// comes before any "Blocked" reference (which would be in "blocking" array)
	ticketIdx := strings.Index(output, `"ticket"`)
	blockerIdx := strings.Index(output, `"title": "Blocker"`)
	if ticketIdx == -1 || blockerIdx == -1 || blockerIdx < ticketIdx {
		t.Error("Primary ticket should be 'Blocker'")
	}
}

func TestReadyShowsTicketDetails(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Add a ticket with description
	Add([]string{"--title", "Test ticket", "--description", "Test description", "--priority", "1"})

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Ready([]string{"--json"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Ready(--json) error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should show ticket details structure
	if !strings.Contains(output, `"ticket"`) {
		t.Error("Ready JSON output missing 'ticket' field")
	}
	if !strings.Contains(output, "Test ticket") {
		t.Error("Ready output missing ticket title")
	}
	if !strings.Contains(output, "Test description") {
		t.Error("Ready output missing ticket description")
	}
	if !strings.Contains(output, `"comments"`) {
		t.Error("Ready JSON output missing 'comments' field")
	}
}

func TestReadyNoTickets(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Ready([]string{"--json"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Ready(--json) error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No ready tickets found") {
		t.Error("Ready output should indicate no tickets found")
	}
}
