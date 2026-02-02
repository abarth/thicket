package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestDetailModel_Update_TicketClosedMsg(t *testing.T) {
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
	if m.ticket.Status != ticket.StatusOpen {
		t.Errorf("expected initial status open, got %s", m.ticket.Status)
	}

	// Simulate closing ticket in storage
	tk.Status = ticket.StatusClosed
	if err := store.Update(tk); err != nil {
		t.Fatalf("store.Update() error = %v", err)
	}

	// Send TicketClosedMsg and verify it triggers reload
	m, reloadCmd := m.Update(TicketClosedMsg{ID: tk.ID})

	if reloadCmd == nil {
		t.Fatal("expected TicketClosedMsg to return a reload command")
	}

	// Execute the reload command
	reloadMsg := reloadCmd()
	m, _ = m.Update(reloadMsg)

	if m.ticket.Status != ticket.StatusClosed {
		t.Errorf("expected status to be closed after reload, got %s", m.ticket.Status)
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

func TestDetailModel_Update_PriorityDown(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 3, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Test priority down key (-)
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if cmd == nil {
		t.Fatal("expected priority down to return a command")
	}

	resultMsg := cmd()
	updateMsg, ok := resultMsg.(TicketPriorityUpdatedMsg)
	if !ok {
		t.Fatalf("expected TicketPriorityUpdatedMsg, got %T", resultMsg)
	}
	if updateMsg.NewPriority != 4 {
		t.Errorf("expected new priority 4, got %d", updateMsg.NewPriority)
	}
}

func TestDetailModel_Update_SetType(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Test setting type to bug with 'b'
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if cmd == nil {
		t.Fatal("expected set bug type to return a command")
	}

	resultMsg := cmd()
	updateMsg, ok := resultMsg.(TicketTypeUpdatedMsg)
	if !ok {
		t.Fatalf("expected TicketTypeUpdatedMsg, got %T", resultMsg)
	}
	if updateMsg.NewType != ticket.TypeBug {
		t.Errorf("expected new type bug, got %s", updateMsg.NewType)
	}
}

func TestDetailModel_Update_SetFeature(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Test setting type to feature with 'f'
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if cmd == nil {
		t.Fatal("expected set feature type to return a command")
	}

	resultMsg := cmd()
	updateMsg, ok := resultMsg.(TicketTypeUpdatedMsg)
	if !ok {
		t.Fatalf("expected TicketTypeUpdatedMsg, got %T", resultMsg)
	}
	if updateMsg.NewType != ticket.TypeFeature {
		t.Errorf("expected new type feature, got %s", updateMsg.NewType)
	}
}

func TestDetailModel_View(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Test description", ticket.TypeBug, 3, []string{"label1"}, "john")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	view := m.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check that key fields are displayed
	if !strings.Contains(view, tk.ID) {
		t.Error("expected view to contain ticket ID")
	}
	if !strings.Contains(view, "Test ticket") {
		t.Error("expected view to contain ticket title")
	}
	if !strings.Contains(view, "bug") {
		t.Error("expected view to contain ticket type")
	}
	if !strings.Contains(view, "open") {
		t.Error("expected view to contain ticket status")
	}
	if !strings.Contains(view, "john") {
		t.Error("expected view to contain assignee")
	}
	if !strings.Contains(view, "label1") {
		t.Error("expected view to contain label")
	}
	if !strings.Contains(view, "Test description") {
		t.Error("expected view to contain description")
	}
}

func TestDetailModel_View_Loading(t *testing.T) {
	store := setupTestStore(t)

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.loading = true

	view := m.View()

	if !strings.Contains(view, "Loading") {
		t.Error("expected view to show loading message")
	}
}

func TestDetailModel_View_Error(t *testing.T) {
	store := setupTestStore(t)

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.err = fmt.Errorf("test error")

	view := m.View()

	if !strings.Contains(view, "Error") || !strings.Contains(view, "test error") {
		t.Error("expected view to show error message")
	}
}

func TestDetailModel_View_WithComments(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	comment, err := ticket.NewComment(tk.ID, "This is a test comment")
	if err != nil {
		t.Fatalf("ticket.NewComment() error = %v", err)
	}
	if err := store.AddComment(comment); err != nil {
		t.Fatalf("store.AddComment() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	view := m.View()

	if !strings.Contains(view, "Comments") {
		t.Error("expected view to contain Comments section")
	}
	if !strings.Contains(view, "This is a test comment") {
		t.Error("expected view to contain comment content")
	}
}

func TestDetailModel_CloseTicket(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Initiate close with 'c' key
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	if !m.closePrompt.Active() {
		t.Fatal("expected close prompt to be active")
	}

	// Confirm with 'y'
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if cmd == nil {
		t.Fatal("expected close to return a command")
	}

	// Execute the close command
	resultMsg := cmd()
	closedMsg, ok := resultMsg.(TicketClosedMsg)
	if !ok {
		if errMsg, isErr := resultMsg.(ErrorMsg); isErr {
			t.Fatalf("close returned error: %v", errMsg.Err)
		}
		t.Fatalf("expected TicketClosedMsg, got %T", resultMsg)
	}

	if closedMsg.ID != tk.ID {
		t.Errorf("expected closed ID %s, got %s", tk.ID, closedMsg.ID)
	}

	// Verify in store
	updated, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}
	if updated.Status != ticket.StatusClosed {
		t.Errorf("expected status closed, got %s", updated.Status)
	}
}

func TestDetailModel_CloseTicket_AlreadyClosed(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	tk.Status = ticket.StatusClosed
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Try to close an already closed ticket
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	if m.closePrompt.Active() {
		t.Error("expected close prompt to NOT be active for already closed ticket")
	}
}

func TestDetailModel_SaveComment(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Enter comment mode with 'm'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	if !m.commenting {
		t.Fatal("expected to be in commenting mode")
	}

	// Type a comment
	m.commentInput.SetValue("This is my comment")

	// Save with ctrl+s
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if cmd == nil {
		t.Fatal("expected save comment to return a command")
	}

	// Execute the save command
	resultMsg := cmd()
	savedMsg, ok := resultMsg.(CommentSavedMsg)
	if !ok {
		if errMsg, isErr := resultMsg.(ErrorMsg); isErr {
			t.Fatalf("save comment returned error: %v", errMsg.Err)
		}
		t.Fatalf("expected CommentSavedMsg, got %T", resultMsg)
	}

	if savedMsg.TicketID != tk.ID {
		t.Errorf("expected ticket ID %s, got %s", tk.ID, savedMsg.TicketID)
	}

	// Verify comment was saved
	comments, err := store.GetComments(tk.ID)
	if err != nil {
		t.Fatalf("store.GetComments() error = %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	if comments[0].Content != "This is my comment" {
		t.Errorf("expected comment content 'This is my comment', got '%s'", comments[0].Content)
	}
}

func TestDetailModel_CancelComment(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Enter comment mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	if !m.commenting {
		t.Fatal("expected to be in commenting mode")
	}

	// Type something
	m.commentInput.SetValue("Draft comment")

	// Cancel with esc
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.commenting {
		t.Error("expected to exit commenting mode")
	}
	if m.commentInput.Value() != "" {
		t.Error("expected comment input to be reset")
	}
}

func TestDetailModel_SaveEmptyComment(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Enter comment mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	// Don't type anything, try to save
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	// Should not save empty comment
	if cmd != nil {
		t.Error("expected no command for empty comment")
	}
}

func TestDetailModel_SetSearchQuery(t *testing.T) {
	store := setupTestStore(t)
	m := NewDetailModel(store)

	m.SetSearchQuery("test query")

	if m.searchQuery != "test query" {
		t.Errorf("expected search query 'test query', got '%s'", m.searchQuery)
	}
}

func TestDetailModel_Scroll(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Long description\nline 2\nline 3\nline 4\nline 5", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 10) // Small height to enable scrolling
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	initialScroll := m.scrollY

	// Scroll down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.scrollY != initialScroll+1 {
		t.Errorf("expected scrollY to increase, got %d", m.scrollY)
	}

	// Scroll up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.scrollY != initialScroll {
		t.Errorf("expected scrollY to return to initial, got %d", m.scrollY)
	}

	// Scroll up at 0 should stay at 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.scrollY != 0 {
		t.Errorf("expected scrollY to stay at 0, got %d", m.scrollY)
	}
}

func TestDetailModel_BackToList(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press esc to go back
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected back to return a command")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(BackToListMsg); !ok {
		t.Errorf("expected BackToListMsg, got %T", resultMsg)
	}
}

func TestDetailModel_EditTicket(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewDetailModel(store)
	m.SetSize(80, 24)
	m.SetTicketID(tk.ID)

	cmd := m.LoadTicket()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'e' to edit
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})

	if cmd == nil {
		t.Fatal("expected edit to return a command")
	}

	resultMsg := cmd()
	editMsg, ok := resultMsg.(EditTicketMsg)
	if !ok {
		t.Fatalf("expected EditTicketMsg, got %T", resultMsg)
	}

	if editMsg.Ticket.ID != tk.ID {
		t.Errorf("expected ticket ID %s, got %s", tk.ID, editMsg.Ticket.ID)
	}
}
