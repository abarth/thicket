package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abarth/thicket/internal/ticket"
)

func TestListModel_CloseLastTicket(t *testing.T) {
	store := setupTestStore(t)

	// Create and add a single test ticket
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create list model (defaults to filtering by open status)
	m := NewListModel(store)
	m.SetSize(80, 24)

	// Load tickets
	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(m.tickets))
	}
	if m.cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", m.cursor)
	}

	// Close the ticket in storage
	tk.Status = ticket.StatusClosed
	if err := store.Update(tk); err != nil {
		t.Fatalf("store.Update() error = %v", err)
	}

	// Simulate refresh after closing (this is what happens in tui.go after TicketClosedMsg)
	cmd = m.Refresh()
	msg = cmd()
	m, _ = m.Update(msg)

	// After closing, the list should be empty (filtering by open tickets)
	if len(m.tickets) != 0 {
		t.Fatalf("expected 0 tickets after closing, got %d", len(m.tickets))
	}

	// Cursor should be at 0 (safe position for empty list)
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 for empty list, got %d", m.cursor)
	}

	// View should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestListModel_CloseCursorAtEnd(t *testing.T) {
	store := setupTestStore(t)

	// Create and add two test tickets
	tk1, err := ticket.New("TH", "Ticket 1", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk1); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	tk2, err := ticket.New("TH", "Ticket 2", "Description", ticket.TypeTask, 2, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk2); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create list model
	m := NewListModel(store)
	m.SetSize(80, 24)

	// Load tickets
	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 2 {
		t.Fatalf("expected 2 tickets, got %d", len(m.tickets))
	}

	// Move cursor to the second ticket (index 1)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("expected cursor at 1, got %d", m.cursor)
	}

	// Close the second ticket in storage
	tk2.Status = ticket.StatusClosed
	if err := store.Update(tk2); err != nil {
		t.Fatalf("store.Update() error = %v", err)
	}

	// Simulate refresh after closing
	cmd = m.Refresh()
	msg = cmd()
	m, _ = m.Update(msg)

	// After closing, should have 1 ticket
	if len(m.tickets) != 1 {
		t.Fatalf("expected 1 ticket after closing, got %d", len(m.tickets))
	}

	// Cursor should be adjusted to valid position (0)
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}

	// View should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestListModel_ClosePromptConfirm(t *testing.T) {
	store := setupTestStore(t)

	// Create and add a single test ticket
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	// Create list model
	m := NewListModel(store)
	m.SetSize(80, 24)

	// Load tickets
	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(m.tickets))
	}

	// Press 'c' to initiate close
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	if !m.closePrompt.Active() {
		t.Fatal("expected close prompt to be active")
	}

	// Confirm with 'y'
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if cmd == nil {
		t.Fatal("expected close command to be returned")
	}

	// Execute the close command
	closeMsg := cmd()
	closedMsg, ok := closeMsg.(TicketClosedMsg)
	if !ok {
		// Check if it's an error
		if errMsg, isErr := closeMsg.(ErrorMsg); isErr {
			t.Fatalf("close returned error: %v", errMsg.Err)
		}
		t.Fatalf("expected TicketClosedMsg, got %T", closeMsg)
	}

	if closedMsg.ID != tk.ID {
		t.Errorf("expected closed ticket ID %s, got %s", tk.ID, closedMsg.ID)
	}
}

func TestListModel_CloseOnEmptyList(t *testing.T) {
	store := setupTestStore(t)

	// Create list model with no tickets
	m := NewListModel(store)
	m.SetSize(80, 24)

	// Load tickets (empty)
	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 0 {
		t.Fatalf("expected 0 tickets, got %d", len(m.tickets))
	}

	// Press 'c' - should do nothing since there are no tickets
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	if m.closePrompt.Active() {
		t.Error("expected close prompt to NOT be active on empty list")
	}

	if cmd != nil {
		t.Error("expected no command on empty list")
	}
}

func TestListModel_EnterOnEmptyList(t *testing.T) {
	store := setupTestStore(t)

	// Create list model with no tickets
	m := NewListModel(store)
	m.SetSize(80, 24)

	// Load tickets (empty)
	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 0 {
		t.Fatalf("expected 0 tickets, got %d", len(m.tickets))
	}

	// Press enter - should do nothing since there are no tickets
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("expected no command on enter with empty list")
	}
}

func TestListModel_EditOnEmptyList(t *testing.T) {
	store := setupTestStore(t)

	// Create list model with no tickets
	m := NewListModel(store)
	m.SetSize(80, 24)

	// Load tickets (empty)
	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'e' - should do nothing since there are no tickets
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})

	if cmd != nil {
		t.Error("expected no command on edit with empty list")
	}
}

func TestListModel_ViewEmptyList(t *testing.T) {
	store := setupTestStore(t)

	// Create list model with no tickets
	m := NewListModel(store)
	m.SetSize(80, 24)

	// Load tickets (empty)
	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// View should not panic and should show message
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view for empty list")
	}
}
