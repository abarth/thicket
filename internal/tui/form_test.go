package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abarth/thicket/internal/ticket"
)

func TestNewFormModel_NewTicket(t *testing.T) {
	store := setupTestStore(t)

	m := NewFormModel(store, "TH", nil)

	if !m.isNew {
		t.Error("expected isNew to be true for new ticket")
	}
	if m.ticketID != "" {
		t.Error("expected ticketID to be empty for new ticket")
	}
	// Check defaults
	if m.ticketType.Value() != "task" {
		t.Errorf("expected default type 'task', got '%s'", m.ticketType.Value())
	}
	if m.priority.Value() != "2" {
		t.Errorf("expected default priority '2', got '%s'", m.priority.Value())
	}
	if m.status.Value() != "open" {
		t.Errorf("expected default status 'open', got '%s'", m.status.Value())
	}
}

func TestNewFormModel_EditTicket(t *testing.T) {
	store := setupTestStore(t)

	tk, err := ticket.New("TH", "Test Title", "Test Description", ticket.TypeBug, 3, []string{"label1", "label2"}, "john")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewFormModel(store, "TH", tk)

	if m.isNew {
		t.Error("expected isNew to be false for editing existing ticket")
	}
	if m.ticketID != tk.ID {
		t.Errorf("expected ticketID '%s', got '%s'", tk.ID, m.ticketID)
	}
	if m.title.Value() != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", m.title.Value())
	}
	if m.description.Value() != "Test Description" {
		t.Errorf("expected description 'Test Description', got '%s'", m.description.Value())
	}
	if m.ticketType.Value() != "bug" {
		t.Errorf("expected type 'bug', got '%s'", m.ticketType.Value())
	}
	if m.priority.Value() != "3" {
		t.Errorf("expected priority '3', got '%s'", m.priority.Value())
	}
	if m.assignee.Value() != "john" {
		t.Errorf("expected assignee 'john', got '%s'", m.assignee.Value())
	}
	if m.labels.Value() != "label1, label2" {
		t.Errorf("expected labels 'label1, label2', got '%s'", m.labels.Value())
	}
}

func TestFormModel_Init(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a command (textinput.Blink)")
	}
}

func TestFormModel_SetSize(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
	// Input width should be capped at 60
	if m.title.Width != 60 {
		t.Errorf("expected title width 60, got %d", m.title.Width)
	}
}

func TestFormModel_SetSize_SmallWidth(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	m.SetSize(40, 30)

	if m.width != 40 {
		t.Errorf("expected width 40, got %d", m.width)
	}
	// Input width should be width - 20
	if m.title.Width != 20 {
		t.Errorf("expected title width 20, got %d", m.title.Width)
	}
}

func TestFormModel_FieldNavigation(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	// Should start at title field
	if m.focus != fieldTitle {
		t.Errorf("expected focus on fieldTitle, got %d", m.focus)
	}

	// Press tab to go to next field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != fieldDescription {
		t.Errorf("expected focus on fieldDescription, got %d", m.focus)
	}

	// Continue through all fields
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != fieldType {
		t.Errorf("expected focus on fieldType, got %d", m.focus)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != fieldStatus {
		t.Errorf("expected focus on fieldStatus, got %d", m.focus)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != fieldPriority {
		t.Errorf("expected focus on fieldPriority, got %d", m.focus)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != fieldAssignee {
		t.Errorf("expected focus on fieldAssignee, got %d", m.focus)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != fieldLabels {
		t.Errorf("expected focus on fieldLabels, got %d", m.focus)
	}

	// Wrap around to title
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != fieldTitle {
		t.Errorf("expected focus to wrap to fieldTitle, got %d", m.focus)
	}
}

func TestFormModel_FieldNavigation_Prev(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	// Should start at title field
	if m.focus != fieldTitle {
		t.Errorf("expected focus on fieldTitle, got %d", m.focus)
	}

	// Press shift+tab to go to previous field (wrap to labels)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focus != fieldLabels {
		t.Errorf("expected focus to wrap to fieldLabels, got %d", m.focus)
	}

	// Continue backwards
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focus != fieldAssignee {
		t.Errorf("expected focus on fieldAssignee, got %d", m.focus)
	}
}

func TestFormModel_CancelWithoutChanges(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	// Press escape without making changes
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected cancel to return a command")
	}

	msg := cmd()
	if _, ok := msg.(BackToListMsg); !ok {
		t.Errorf("expected BackToListMsg, got %T", msg)
	}
}

func TestFormModel_CancelWithChanges(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	// Make a change by typing in the title field
	m.title.SetValue("New Title")

	// Press escape - should show discard prompt
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.discardPrompt.Active() != true {
		t.Error("expected discard prompt to be active")
	}
	if cmd != nil {
		t.Error("expected no command when showing discard prompt")
	}

	// Confirm discard with 'y'
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if cmd == nil {
		t.Fatal("expected confirm to return a command")
	}

	msg := cmd()
	if _, ok := msg.(BackToListMsg); !ok {
		t.Errorf("expected BackToListMsg, got %T", msg)
	}
}

func TestFormModel_CancelDiscardPrompt(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	// Make a change
	m.title.SetValue("New Title")

	// Press escape - should show discard prompt
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.discardPrompt.Active() != true {
		t.Error("expected discard prompt to be active")
	}

	// Press 'n' to cancel
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	if m.discardPrompt.Active() == true {
		t.Error("expected discard prompt to be inactive after cancel")
	}
	if cmd != nil {
		t.Error("expected no command after canceling discard prompt")
	}
}

func TestFormModel_SaveNewTicket(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)
	m.SetSize(80, 24)

	// Fill in required fields
	m.title.SetValue("Test New Ticket")
	m.description.SetValue("This is a test description")

	// Press ctrl+s to save
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if cmd == nil {
		t.Fatal("expected save to return a command")
	}

	msg := cmd()
	savedMsg, ok := msg.(TicketSavedMsg)
	if !ok {
		if errMsg, isErr := msg.(ErrorMsg); isErr {
			t.Fatalf("save returned error: %v", errMsg.Err)
		}
		t.Fatalf("expected TicketSavedMsg, got %T", msg)
	}

	if !savedMsg.IsNew {
		t.Error("expected IsNew to be true")
	}
	if savedMsg.ID == "" {
		t.Error("expected ticket ID to be set")
	}
	if !strings.Contains(savedMsg.Message, "Created") {
		t.Errorf("expected message to contain 'Created', got '%s'", savedMsg.Message)
	}
}

func TestFormModel_SaveEditedTicket(t *testing.T) {
	store := setupTestStore(t)

	// Create existing ticket
	tk, err := ticket.New("TH", "Original Title", "Original Description", ticket.TypeTask, 2, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("store.Add() error = %v", err)
	}

	m := NewFormModel(store, "TH", tk)
	m.SetSize(80, 24)

	// Edit the title
	m.title.SetValue("Updated Title")

	// Press ctrl+s to save
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if cmd == nil {
		t.Fatal("expected save to return a command")
	}

	msg := cmd()
	savedMsg, ok := msg.(TicketSavedMsg)
	if !ok {
		if errMsg, isErr := msg.(ErrorMsg); isErr {
			t.Fatalf("save returned error: %v", errMsg.Err)
		}
		t.Fatalf("expected TicketSavedMsg, got %T", msg)
	}

	if savedMsg.IsNew {
		t.Error("expected IsNew to be false for update")
	}
	if savedMsg.ID != tk.ID {
		t.Errorf("expected ticket ID %s, got %s", tk.ID, savedMsg.ID)
	}

	// Verify the update in store
	updated, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", updated.Title)
	}
}

func TestFormModel_SaveValidation_EmptyTitle(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)
	m.SetSize(80, 24)

	// Don't fill in title (empty)
	m.title.SetValue("")

	// Press ctrl+s to save
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if cmd == nil {
		t.Fatal("expected save to return a command")
	}

	msg := cmd()
	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg for validation failure, got %T", msg)
	}
	if !strings.Contains(errMsg.Err.Error(), "Title is required") {
		t.Errorf("expected error about title, got: %v", errMsg.Err)
	}
}

func TestFormModel_SaveWithLabels(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)
	m.SetSize(80, 24)

	// Fill in required fields
	m.title.SetValue("Test Ticket with Labels")
	m.labels.SetValue("bug, frontend, urgent")

	// Press ctrl+s to save
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if cmd == nil {
		t.Fatal("expected save to return a command")
	}

	msg := cmd()
	savedMsg, ok := msg.(TicketSavedMsg)
	if !ok {
		if errMsg, isErr := msg.(ErrorMsg); isErr {
			t.Fatalf("save returned error: %v", errMsg.Err)
		}
		t.Fatalf("expected TicketSavedMsg, got %T", msg)
	}

	// Verify labels in store
	saved, err := store.Get(savedMsg.ID)
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}
	if len(saved.Labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(saved.Labels))
	}
	expectedLabels := []string{"bug", "frontend", "urgent"}
	for i, label := range expectedLabels {
		if saved.Labels[i] != label {
			t.Errorf("expected label %d to be '%s', got '%s'", i, label, saved.Labels[i])
		}
	}
}

func TestFormModel_View(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)
	m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check that key fields are rendered
	if !strings.Contains(view, "Title") {
		t.Error("expected view to contain 'Title' label")
	}
	if !strings.Contains(view, "Description") {
		t.Error("expected view to contain 'Description' label")
	}
	if !strings.Contains(view, "Type") {
		t.Error("expected view to contain 'Type' label")
	}
	if !strings.Contains(view, "Priority") {
		t.Error("expected view to contain 'Priority' label")
	}
}

func TestFormModel_View_WithDiscardPrompt(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)
	m.SetSize(80, 24)

	// Make a change and press escape to show discard prompt
	m.title.SetValue("Changed")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	view := m.View()

	if !strings.Contains(view, "Discard changes") {
		t.Error("expected view to contain discard prompt")
	}
}

func TestFormModel_IsDirty(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	// Initially not dirty
	if m.isDirty() {
		t.Error("expected form to not be dirty initially")
	}

	// Make a change
	m.title.SetValue("Changed Title")

	// Now should be dirty
	if !m.isDirty() {
		t.Error("expected form to be dirty after change")
	}
}

func TestFormModel_GetValues(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)

	m.title.SetValue("Test Title")
	m.description.SetValue("Test Description")
	m.assignee.SetValue("john")

	values := m.getValues()

	if values[fieldTitle] != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", values[fieldTitle])
	}
	if values[fieldDescription] != "Test Description" {
		t.Errorf("expected description 'Test Description', got '%s'", values[fieldDescription])
	}
	if values[fieldAssignee] != "john" {
		t.Errorf("expected assignee 'john', got '%s'", values[fieldAssignee])
	}
}

func TestFormModel_SelectorNavigation(t *testing.T) {
	store := setupTestStore(t)
	m := NewFormModel(store, "TH", nil)
	m.SetSize(80, 24)

	// Navigate to type field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // description
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // type

	if m.focus != fieldType {
		t.Fatalf("expected focus on type field, got %d", m.focus)
	}

	// Default is "task", use arrow keys to change
	initialType := m.ticketType.Value()
	if initialType != "task" {
		t.Errorf("expected initial type 'task', got '%s'", initialType)
	}

	// Press right arrow to go to next type option
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	newType := m.ticketType.Value()

	// Should have changed from "task" to "epic" (next in list after task)
	if newType == initialType {
		t.Error("expected type to change after right arrow")
	}
}
