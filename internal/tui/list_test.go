package tui

import (
	"fmt"
	"strings"
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

func TestListModel_UpdatePriority(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 5, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press '+' to increase priority (lower number = higher priority)
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})

	if cmd == nil {
		t.Fatal("expected priority up to return a command")
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

func TestListModel_UpdatePriorityDown(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 3, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press '-' to decrease priority (higher number = lower priority)
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

func TestListModel_UpdatePriorityUpAtZero(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 0, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press '+' at priority 0 - should not change
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})

	if cmd != nil {
		t.Error("expected no command when priority is already 0")
	}
}

func TestListModel_UpdateType(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'b' to set type to bug
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})

	if cmd == nil {
		t.Fatal("expected set type to return a command")
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

func TestListModel_UpdateType_Feature(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'f' to set type to feature
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})

	if cmd == nil {
		t.Fatal("expected set type to return a command")
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

func TestListModel_UpdateType_Epic(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'E' to set type to epic
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("E")})

	if cmd == nil {
		t.Fatal("expected set type to return a command")
	}

	resultMsg := cmd()
	updateMsg, ok := resultMsg.(TicketTypeUpdatedMsg)
	if !ok {
		t.Fatalf("expected TicketTypeUpdatedMsg, got %T", resultMsg)
	}

	if updateMsg.NewType != ticket.TypeEpic {
		t.Errorf("expected new type epic, got %s", updateMsg.NewType)
	}
}

func TestListModel_UpdateType_Cleanup(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'C' to set type to cleanup
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")})

	if cmd == nil {
		t.Fatal("expected set type to return a command")
	}

	resultMsg := cmd()
	updateMsg, ok := resultMsg.(TicketTypeUpdatedMsg)
	if !ok {
		t.Fatalf("expected TicketTypeUpdatedMsg, got %T", resultMsg)
	}

	if updateMsg.NewType != ticket.TypeCleanup {
		t.Errorf("expected new type cleanup, got %s", updateMsg.NewType)
	}
}

func TestListModel_Search(t *testing.T) {
	store := setupTestStore(t)

	tk1, _ := ticket.New("TH", "First ticket", "Description one", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Second ticket", "Description two", ticket.TypeTask, 2, nil, "")
	store.Add(tk1)
	store.Add(tk2)

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 2 {
		t.Fatalf("expected 2 tickets, got %d", len(m.tickets))
	}

	// Press '/' to enter search mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	if !m.IsSearching() {
		t.Fatal("expected to be in search mode")
	}

	// Type search query
	m.searchInput.SetValue("First")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.IsSearching() {
		t.Error("expected to exit search mode after enter")
	}

	// Should filter to 1 ticket
	if len(m.tickets) != 1 {
		t.Errorf("expected 1 filtered ticket, got %d", len(m.tickets))
	}
	if m.tickets[0].Title != "First ticket" {
		t.Errorf("expected 'First ticket', got '%s'", m.tickets[0].Title)
	}
}

func TestListModel_SearchCancel(t *testing.T) {
	store := setupTestStore(t)

	tk1, _ := ticket.New("TH", "First ticket", "Description", ticket.TypeTask, 1, nil, "")
	store.Add(tk1)

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Enter search mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	// Type something but then cancel
	m.searchInput.SetValue("nonexistent")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.IsSearching() {
		t.Error("expected to exit search mode after esc")
	}

	// Should still have original tickets
	if len(m.tickets) != 1 {
		t.Errorf("expected 1 ticket after cancel, got %d", len(m.tickets))
	}
}

func TestListModel_FilterByStatus(t *testing.T) {
	store := setupTestStore(t)

	tk1, _ := ticket.New("TH", "Open ticket", "Description", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Closed ticket", "Description", ticket.TypeTask, 2, nil, "")
	tk2.Status = ticket.StatusClosed

	store.Add(tk1)
	store.Add(tk2)

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Default is open only
	if len(m.tickets) != 1 {
		t.Errorf("expected 1 open ticket, got %d", len(m.tickets))
	}

	// Press 'x' to show closed
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	msg = cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 1 {
		t.Errorf("expected 1 closed ticket, got %d", len(m.tickets))
	}
	if m.tickets[0].Status != ticket.StatusClosed {
		t.Error("expected closed ticket")
	}

	// Press 'a' to show all
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	msg = cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 2 {
		t.Errorf("expected 2 tickets with 'all' filter, got %d", len(m.tickets))
	}
}

func TestListModel_FilterIcebox(t *testing.T) {
	store := setupTestStore(t)

	tk1, _ := ticket.New("TH", "Open ticket", "Description", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Icebox ticket", "Description", ticket.TypeTask, 2, nil, "")
	tk2.Status = ticket.StatusIcebox

	store.Add(tk1)
	store.Add(tk2)

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'i' to show icebox
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	msg = cmd()
	m, _ = m.Update(msg)

	if len(m.tickets) != 1 {
		t.Errorf("expected 1 icebox ticket, got %d", len(m.tickets))
	}
	if m.tickets[0].Status != ticket.StatusIcebox {
		t.Error("expected icebox ticket")
	}
}

func TestListModel_Navigation_TopBottom(t *testing.T) {
	store := setupTestStore(t)

	for i := 0; i < 10; i++ {
		tk, _ := ticket.New("TH", fmt.Sprintf("Ticket %d", i), "Description", ticket.TypeTask, i, nil, "")
		store.Add(tk)
	}

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'G' to go to bottom
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m.cursor != 9 {
		t.Errorf("expected cursor at 9, got %d", m.cursor)
	}

	// Press 'g' to go to top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
}

func TestListModel_NewTicket(t *testing.T) {
	store := setupTestStore(t)
	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'n' to create new ticket
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	if cmd == nil {
		t.Fatal("expected new ticket to return a command")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(CreateTicketMsg); !ok {
		t.Errorf("expected CreateTicketMsg, got %T", resultMsg)
	}
}

func TestListModel_Refresh(t *testing.T) {
	store := setupTestStore(t)

	tk, _ := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	store.Add(tk)

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Press 'r' to refresh
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	if !m.loading {
		t.Error("expected loading to be true during refresh")
	}

	if cmd == nil {
		t.Fatal("expected refresh to return a command")
	}
}

func TestListModel_View_WithTickets(t *testing.T) {
	store := setupTestStore(t)

	tk, _ := ticket.New("TH", "Test ticket", "Description", ticket.TypeBug, 3, nil, "")
	store.Add(tk)

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	view := m.View()

	// Check for table header
	if !strings.Contains(view, "ID") || !strings.Contains(view, "TITLE") {
		t.Error("expected view to contain table header")
	}

	// Check for ticket data
	if !strings.Contains(view, tk.ID) {
		t.Error("expected view to contain ticket ID")
	}
	if !strings.Contains(view, "Test ticket") {
		t.Error("expected view to contain ticket title")
	}
}

func TestListModel_View_Loading(t *testing.T) {
	store := setupTestStore(t)
	m := NewListModel(store)
	m.SetSize(80, 24)
	m.loading = true

	view := m.View()

	if !strings.Contains(view, "Loading") {
		t.Error("expected view to show loading message")
	}
}

func TestListModel_View_Error(t *testing.T) {
	store := setupTestStore(t)
	m := NewListModel(store)
	m.SetSize(80, 24)
	m.err = fmt.Errorf("test error")

	view := m.View()

	if !strings.Contains(view, "Error") || !strings.Contains(view, "test error") {
		t.Error("expected view to show error message")
	}
}

func TestListModel_HasFilters(t *testing.T) {
	store := setupTestStore(t)
	m := NewListModel(store)

	// Default has status filter (open)
	if !m.hasFilters() {
		t.Error("expected hasFilters to be true with default status filter")
	}

	// Clear status filter
	m.filters.Status = nil
	if m.hasFilters() {
		t.Error("expected hasFilters to be false without any filters")
	}

	// Add query filter
	m.filters.Query = "test"
	if !m.hasFilters() {
		t.Error("expected hasFilters to be true with query filter")
	}
}

func TestListModel_ApplyFilters_CommentSearch(t *testing.T) {
	store := setupTestStore(t)

	tk1, _ := ticket.New("TH", "First ticket", "Description", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Second ticket", "Description", ticket.TypeTask, 2, nil, "")
	store.Add(tk1)
	store.Add(tk2)

	// Add comment to first ticket
	comment, _ := ticket.NewComment(tk1.ID, "special keyword in comment")
	store.AddComment(comment)

	m := NewListModel(store)
	m.SetSize(80, 24)

	cmd := m.Init()
	msg := cmd()
	m, _ = m.Update(msg)

	// Search for keyword in comment
	m.filters.Query = "special keyword"
	m.applyFilters()

	if len(m.tickets) != 1 {
		t.Errorf("expected 1 ticket matching comment, got %d", len(m.tickets))
	}
	if m.tickets[0].ID != tk1.ID {
		t.Errorf("expected ticket %s, got %s", tk1.ID, m.tickets[0].ID)
	}
}

func TestListModel_RenderRow(t *testing.T) {
	store := setupTestStore(t)
	m := NewListModel(store)
	m.SetSize(80, 24)

	row := m.renderRow("> ", "TH-123", "1", "bug", "open", "Test title", true)

	if !strings.Contains(row, "> ") {
		t.Error("expected row to contain cursor")
	}
	if !strings.Contains(row, "TH-123") {
		t.Error("expected row to contain ID")
	}
	if !strings.Contains(row, "bug") {
		t.Error("expected row to contain type")
	}
}

func TestListModel_RenderRow_LongTitle(t *testing.T) {
	store := setupTestStore(t)
	m := NewListModel(store)
	m.SetSize(60, 24) // Narrow width to force truncation

	longTitle := "This is a very long title that should be truncated because it exceeds the available width"
	row := m.renderRow("  ", "TH-123", "1", "bug", "open", longTitle, false)

	// Should contain truncation indicator
	if !strings.Contains(row, "...") {
		t.Error("expected truncated title to contain '...'")
	}
}

func TestListModel_RenderRow_EmptyType(t *testing.T) {
	store := setupTestStore(t)
	m := NewListModel(store)
	m.SetSize(80, 24)

	row := m.renderRow("  ", "TH-123", "1", "", "open", "Test title", false)

	// Empty type should show as "-"
	if !strings.Contains(row, "-") {
		t.Error("expected empty type to show as '-'")
	}
}
