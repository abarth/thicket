package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

func setupTestStore(t *testing.T) *storage.Store {
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
	store, err := storage.Open(paths)
	if err != nil {
		t.Fatalf("storage.Open() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestDetailModel_Update_TicketPriorityUpdatedMsg(t *testing.T) {
	store := setupTestStore(t)

	// Create and add a test ticket
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 5, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create detail model and set ticket
	m := NewDetailModel(store)
	m.SetTicketID(tk.ID)

	// Load the ticket first
	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	if m.ticket == nil {
		t.Fatal("expected ticket to be loaded")
	}
	if m.ticket.Priority != 5 {
		t.Errorf("expected initial priority 5, got %d", m.ticket.Priority)
	}

	// Simulate priority update in storage
	tk.Priority = 3
	if err := store.Update(tk); err != nil {
		t.Fatalf("store.Update() error = %v", err)
	}

	// Send TicketPriorityUpdatedMsg and verify it triggers reload
	m, reloadCmd := m.Update(TicketPriorityUpdatedMsg{ID: tk.ID, NewPriority: 3})

	if reloadCmd == nil {
		t.Fatal("expected TicketPriorityUpdatedMsg to return a reload command")
	}

	// Execute the reload command
	reloadMsg := reloadCmd()
	m, _ = m.Update(reloadMsg)

	if m.ticket.Priority != 3 {
		t.Errorf("expected priority to be updated to 3 after reload, got %d", m.ticket.Priority)
	}
}

func TestDetailModel_Update_TicketTypeUpdatedMsg(t *testing.T) {
	store := setupTestStore(t)

	// Create and add a test ticket
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create detail model and set ticket
	m := NewDetailModel(store)
	m.SetTicketID(tk.ID)

	// Load the ticket first
	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	if m.ticket == nil {
		t.Fatal("expected ticket to be loaded")
	}
	if m.ticket.Type != ticket.TypeTask {
		t.Errorf("expected initial type task, got %s", m.ticket.Type)
	}

	// Simulate type update in storage
	tk.Type = ticket.TypeBug
	if err := store.Update(tk); err != nil {
		t.Fatalf("store.Update() error = %v", err)
	}

	// Send TicketTypeUpdatedMsg and verify it triggers reload
	m, reloadCmd := m.Update(TicketTypeUpdatedMsg{ID: tk.ID, NewType: ticket.TypeBug})

	if reloadCmd == nil {
		t.Fatal("expected TicketTypeUpdatedMsg to return a reload command")
	}

	// Execute the reload command
	reloadMsg := reloadCmd()
	m, _ = m.Update(reloadMsg)

	if m.ticket.Type != ticket.TypeBug {
		t.Errorf("expected type to be updated to bug after reload, got %s", m.ticket.Type)
	}
}

func TestDetailModel_Update_CommentSavedMsg(t *testing.T) {
	store := setupTestStore(t)

	// Create and add a test ticket
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create detail model and set ticket
	m := NewDetailModel(store)
	m.SetTicketID(tk.ID)

	// Load the ticket first
	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	if len(m.comments) != 0 {
		t.Errorf("expected no comments initially, got %d", len(m.comments))
	}

	// Add a comment to storage
	comment, err := ticket.NewComment(tk.ID, "Test comment")
	if err != nil {
		t.Fatalf("ticket.NewComment() error = %v", err)
	}
	if err := store.AddComment(comment); err != nil {
		t.Fatalf("store.AddComment() error = %v", err)
	}

	// Send CommentSavedMsg and verify it triggers reload
	m, reloadCmd := m.Update(CommentSavedMsg{TicketID: tk.ID})

	if reloadCmd == nil {
		t.Fatal("expected CommentSavedMsg to return a reload command")
	}

	// Execute the reload command
	reloadMsg := reloadCmd()
	m, _ = m.Update(reloadMsg)

	if len(m.comments) != 1 {
		t.Errorf("expected 1 comment after reload, got %d", len(m.comments))
	}
}

func TestDetailModel_Update_PriorityKeys(t *testing.T) {
	store := setupTestStore(t)

	// Create and add a test ticket with priority 5
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 5, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create detail model and set ticket
	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	// Load the ticket first
	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Test priority up key (+)
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if cmd == nil {
		t.Fatal("expected priority up to return a command")
	}

	// Execute the command and check the resulting message
	resultMsg := cmd()
	updateMsg, ok := resultMsg.(TicketPriorityUpdatedMsg)
	if !ok {
		t.Fatalf("expected TicketPriorityUpdatedMsg, got %T", resultMsg)
	}
	if updateMsg.NewPriority != 4 {
		t.Errorf("expected new priority 4, got %d", updateMsg.NewPriority)
	}
}

func TestDetailModel_Update_PriorityUpAtZero(t *testing.T) {
	store := setupTestStore(t)

	// Create and add a test ticket with priority 0
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 0, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create detail model and set ticket
	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	// Load the ticket first
	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Test priority up key at 0 - should not change
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if cmd != nil {
		t.Error("expected priority up at 0 to return nil command")
	}
}
