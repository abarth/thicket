package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
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

func TestInit(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	err := Init([]string{"--project", "TH"})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify project was created
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg.ProjectCode != "TH" {
		t.Errorf("ProjectCode = %q, want TH", cfg.ProjectCode)
	}
}

func TestInit_LowercaseCode(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	err := Init([]string{"--project", "th"})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Should be uppercased
	root, _ := config.FindRoot()
	cfg, _ := config.Load(root)
	if cfg.ProjectCode != "TH" {
		t.Errorf("ProjectCode = %q, want TH (uppercased)", cfg.ProjectCode)
	}
}

func TestInit_MissingProject(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	err := Init([]string{})
	if err == nil {
		t.Error("Init() expected error for missing --project")
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Init([]string{"--project", "TH"})
	if err == nil {
		t.Error("Init() expected error for already initialized")
	}
}

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

func TestShow(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket", "--description", "Description"})

	// Get the ticket ID
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	store.Close()

	err := Show([]string{tickets[0].ID})
	if err != nil {
		t.Fatalf("Show() error = %v", err)
	}
}

func TestShow_LowercaseID(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	store.Close()

	// Use lowercase ID
	lowercaseID := strings.ToLower(tickets[0].ID)
	err := Show([]string{lowercaseID})
	if err != nil {
		t.Fatalf("Show() error = %v (should accept lowercase)", err)
	}
}

func TestShow_NotFound(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Show([]string{"TH-999999"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Show() error = %v, want error containing 'not found'", err)
	}
}

func TestShow_InvalidID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Show([]string{"invalid"})
	if err == nil {
		t.Error("Show() expected error for invalid ID")
	}
}

func TestShow_MissingID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Show([]string{})
	if err == nil {
		t.Error("Show() expected error for missing ID")
	}
}

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

func TestQuickstart(t *testing.T) {
	err := Quickstart([]string{})
	if err != nil {
		t.Fatalf("Quickstart() error = %v", err)
	}
}

func TestComment(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	// Get the ticket ID
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Comment([]string{ticketID, "This is a comment"})
	if err != nil {
		t.Fatalf("Comment() error = %v", err)
	}

	// Verify comment was created
	store, _ = storage.Open(paths)
	comments, err := store.GetComments(ticketID)
	store.Close()

	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(comments))
	}
	if comments[0].Content != "This is a comment" {
		t.Errorf("Content = %q, want 'This is a comment'", comments[0].Content)
	}
}

func TestComment_MissingTicketID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Comment([]string{})
	if err == nil {
		t.Error("Comment() expected error for missing ticket ID")
	}
}

func TestComment_MissingContent(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Comment([]string{ticketID})
	if err == nil {
		t.Error("Comment() expected error for missing content")
	}
}

func TestComment_EmptyContent(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket"})

	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	err := Comment([]string{ticketID, "   "})
	if err == nil {
		t.Error("Comment() expected error for empty content")
	}
}

func TestComment_InvalidTicketID(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Comment([]string{"invalid", "Comment"})
	if err == nil {
		t.Error("Comment() expected error for invalid ticket ID")
	}
}

func TestComment_TicketNotFound(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Comment([]string{"TH-999999", "Comment"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Comment() error = %v, want error containing 'not found'", err)
	}
}

func TestShow_WithComments(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	Add([]string{"--title", "Test ticket", "--description", "Description"})

	// Get the ticket ID
	paths := config.GetPaths(dir)
	store, _ := storage.Open(paths)
	tickets, _ := store.List(nil)
	ticketID := tickets[0].ID
	store.Close()

	// Add a comment
	Comment([]string{ticketID, "A test comment"})

	// Show should not error (output includes comments)
	err := Show([]string{ticketID})
	if err != nil {
		t.Fatalf("Show() error = %v", err)
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

	var buf bytes.Buffer
	printTicketDetail(&buf, tk, comments)

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

	var buf bytes.Buffer
	printTicketDetail(&buf, tk, nil)

	output := buf.String()
	if strings.Contains(output, "Comments:") {
		t.Error("Output should not contain Comments section when there are no comments")
	}
}
