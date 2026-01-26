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

	if !strings.Contains(output, "Blocker") {
		t.Error("Ready output missing 'Blocker'")
	}
	if !strings.Contains(output, "Independent") {
		t.Error("Ready output missing 'Independent'")
	}
	if strings.Contains(output, "Blocked") {
		t.Error("Ready output should not contain 'Blocked'")
	}
}
