package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/ticket"
)

// setupTestProject creates a temporary directory with an initialized Thicket project.
func setupTestProject(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()

	// Resolve symlinks for macOS
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}

	// Save and restore working directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	cleanup := func() {
		os.Chdir(oldWd)
	}

	return dir, cleanup
}

func TestPrintTicketTable(t *testing.T) {
	tickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "First ticket", Status: ticket.StatusOpen, Priority: 1},
		{ID: "TH-222222", Title: "Second ticket", Status: ticket.StatusClosed, Priority: 2},
	}

	var buf bytes.Buffer
	printTicketTable(&buf, tickets)

	output := buf.String()
	if !strings.Contains(output, "TH-111111") {
		t.Error("Output should contain ticket ID TH-111111")
	}
	if !strings.Contains(output, "First ticket") {
		t.Error("Output should contain ticket title")
	}
	if !strings.Contains(output, "open") {
		t.Error("Output should contain status")
	}
}

func TestPrintTicketTable_LongTitle(t *testing.T) {
	tickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "This is a very long title that should be truncated when displayed in the table", Status: ticket.StatusOpen, Priority: 1},
	}

	var buf bytes.Buffer
	printTicketTable(&buf, tickets)

	output := buf.String()
	if strings.Contains(output, "displayed in the table") {
		t.Error("Long title should be truncated")
	}
	if !strings.Contains(output, "...") {
		t.Error("Truncated title should end with ...")
	}
}

func TestPrintTicketDetail_WithComments(t *testing.T) {
	tk := &ticket.Ticket{
		ID:          "TH-111111",
		Title:       "Test ticket",
		Description: "Description",
		Status:      ticket.StatusOpen,
		Priority:    1,
	}
	comments := []*ticket.Comment{
		{ID: "TH-c111111", TicketID: "TH-111111", Content: "First comment"},
		{ID: "TH-c222222", TicketID: "TH-111111", Content: "Second comment"},
	}

	details := &TicketDetails{
		Ticket:   tk,
		Comments: comments,
	}

	var buf bytes.Buffer
	printTicketDetail(&buf, details)

	output := buf.String()
	if !strings.Contains(output, "TH-111111") {
		t.Error("Output should contain ticket ID")
	}
	if !strings.Contains(output, "Comments:") {
		t.Error("Output should contain Comments section")
	}
	if !strings.Contains(output, "First comment") {
		t.Error("Output should contain first comment")
	}
	if !strings.Contains(output, "Second comment") {
		t.Error("Output should contain second comment")
	}
}

func TestPrintTicketDetail_NoComments(t *testing.T) {
	tk := &ticket.Ticket{
		ID:     "TH-111111",
		Title:  "Test ticket",
		Status: ticket.StatusOpen,
	}

	details := &TicketDetails{
		Ticket: tk,
	}

	var buf bytes.Buffer
	printTicketDetail(&buf, details)

	output := buf.String()
	if strings.Contains(output, "Comments:") {
		t.Error("Output should not contain Comments section when there are no comments")
	}
}
