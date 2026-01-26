package ticket

import (
	"strings"
	"testing"
)

func TestValidateProjectCode(t *testing.T) {
	tests := []struct {
		code    string
		wantErr bool
	}{
		{"TH", false},
		{"AB", false},
		{"ZZ", false},
		{"", true},
		{"T", true},
		{"THI", true},
		{"th", true},
		{"T1", true},
		{"1T", true},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := ValidateProjectCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectCode(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	id, err := GenerateID("TH")
	if err != nil {
		t.Fatalf("GenerateID() error = %v", err)
	}

	if !strings.HasPrefix(id, "TH-") {
		t.Errorf("GenerateID() = %q, want prefix TH-", id)
	}

	if len(id) != 9 { // "TH-" + 6 alphanumeric chars
		t.Errorf("GenerateID() = %q, want length 9", id)
	}

	// Generate another ID to ensure uniqueness
	id2, err := GenerateID("TH")
	if err != nil {
		t.Fatalf("GenerateID() error = %v", err)
	}

	if id == id2 {
		t.Errorf("GenerateID() generated duplicate IDs: %q", id)
	}
}

func TestGenerateID_InvalidCode(t *testing.T) {
	_, err := GenerateID("invalid")
	if err == nil {
		t.Error("GenerateID(invalid) expected error, got nil")
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{"TH-abcdef", false},
		{"AB-123456", false},
		{"ZZ-a1b2c3", false},
		{"", true},
		{"TH", true},
		{"TH-", true},
		{"TH-abc", true},
		{"TH-abcdefg", true},
		{"th-abcdef", true},
		{"TH-ABCDEF", true},
		{"TH-abcdeg", false},
		{"TH-z1y2x3", false},
		{"THX-abcdef", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			err := ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestParseProjectCode(t *testing.T) {
	code, err := ParseProjectCode("TH-abcdef")
	if err != nil {
		t.Fatalf("ParseProjectCode() error = %v", err)
	}
	if code != "TH" {
		t.Errorf("ParseProjectCode() = %q, want TH", code)
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		status  Status
		wantErr bool
	}{
		{StatusOpen, false},
		{StatusClosed, false},
		{"", true},
		{"pending", true},
		{"OPEN", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatus(%q) error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	ticket, err := New("TH", "Test ticket", "A description", TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if !strings.HasPrefix(ticket.ID, "TH-") {
		t.Errorf("New().ID = %q, want prefix TH-", ticket.ID)
	}
	if ticket.Title != "Test ticket" {
		t.Errorf("New().Title = %q, want 'Test ticket'", ticket.Title)
	}
	if ticket.Description != "A description" {
		t.Errorf("New().Description = %q, want 'A description'", ticket.Description)
	}
	if ticket.Status != StatusOpen {
		t.Errorf("New().Status = %q, want %q", ticket.Status, StatusOpen)
	}
	if ticket.Priority != 1 {
		t.Errorf("New().Priority = %d, want 1", ticket.Priority)
	}
	if ticket.Type != TypeTask {
		t.Errorf("New().Type = %q, want %q", ticket.Type, TypeTask)
	}
	if ticket.Created.IsZero() {
		t.Error("New().Created is zero")
	}
	if ticket.Updated.IsZero() {
		t.Error("New().Updated is zero")
	}
}

func TestNew_TrimSpace(t *testing.T) {
	ticket, err := New("TH", "  Test ticket  ", "  Description  ", TypeTask, 0, nil, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if ticket.Title != "Test ticket" {
		t.Errorf("New().Title = %q, want 'Test ticket'", ticket.Title)
	}
	if ticket.Description != "Description" {
		t.Errorf("New().Description = %q, want 'Description'", ticket.Description)
	}
}

func TestNew_EmptyTitle(t *testing.T) {
	_, err := New("TH", "", "Description", TypeTask, 0, nil, "")
	if err != ErrEmptyTitle {
		t.Errorf("New() error = %v, want ErrEmptyTitle", err)
	}

	_, err = New("TH", "   ", "Description", TypeTask, 0, nil, "")
	if err != ErrEmptyTitle {
		t.Errorf("New() error = %v, want ErrEmptyTitle", err)
	}
}

func TestTicket_Validate(t *testing.T) {
	ticket := &Ticket{
		ID:     "TH-abcdef",
		Title:  "Test",
		Status: StatusOpen,
	}

	if err := ticket.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	ticket.ID = "invalid"
	if err := ticket.Validate(); err == nil {
		t.Error("Validate() expected error for invalid ID")
	}

	ticket.ID = "TH-abcdef"
	ticket.Title = ""
	if err := ticket.Validate(); err == nil {
		t.Error("Validate() expected error for empty title")
	}

	ticket.Title = "Test"
	ticket.Status = "invalid"
	if err := ticket.Validate(); err == nil {
		t.Error("Validate() expected error for invalid status")
	}
}

func TestTicket_Close(t *testing.T) {
	ticket := &Ticket{
		ID:     "TH-abcdef",
		Title:  "Test",
		Status: StatusOpen,
	}

	ticket.Close()

	if ticket.Status != StatusClosed {
		t.Errorf("Close() status = %q, want %q", ticket.Status, StatusClosed)
	}
	if ticket.Updated.IsZero() {
		t.Error("Close() did not update timestamp")
	}
}

func TestTicket_Update(t *testing.T) {
	ticket := &Ticket{
		ID:          "TH-abcdef",
		Title:       "Original",
		Description: "Original desc",
		Status:      StatusOpen,
		Priority:    1,
	}

	newTitle := "Updated"
	newDesc := "Updated desc"
	newPriority := 2
	newStatus := StatusClosed

	err := ticket.Update(&newTitle, &newDesc, nil, &newPriority, &newStatus, nil, nil, nil)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if ticket.Title != "Updated" {
		t.Errorf("Update() Title = %q, want 'Updated'", ticket.Title)
	}
	if ticket.Description != "Updated desc" {
		t.Errorf("Update() Description = %q, want 'Updated desc'", ticket.Description)
	}
	if ticket.Priority != 2 {
		t.Errorf("Update() Priority = %d, want 2", ticket.Priority)
	}
	if ticket.Status != StatusClosed {
		t.Errorf("Update() Status = %q, want %q", ticket.Status, StatusClosed)
	}
}

func TestTicket_Update_Partial(t *testing.T) {
	ticket := &Ticket{
		ID:          "TH-abcdef",
		Title:       "Original",
		Description: "Original desc",
		Status:      StatusOpen,
		Priority:    1,
	}

	newTitle := "Updated"
	err := ticket.Update(&newTitle, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if ticket.Title != "Updated" {
		t.Errorf("Update() Title = %q, want 'Updated'", ticket.Title)
	}
	if ticket.Description != "Original desc" {
		t.Errorf("Update() Description changed unexpectedly")
	}
}

func TestTicket_Update_EmptyTitle(t *testing.T) {
	ticket := &Ticket{
		ID:    "TH-abcdef",
		Title: "Original",
	}

	emptyTitle := ""
	err := ticket.Update(&emptyTitle, nil, nil, nil, nil, nil, nil, nil)
	if err != ErrEmptyTitle {
		t.Errorf("Update() error = %v, want ErrEmptyTitle", err)
	}
}

func TestTicket_Update_InvalidStatus(t *testing.T) {
	ticket := &Ticket{
		ID:     "TH-abcdef",
		Title:  "Test",
		Status: StatusOpen,
	}

	invalidStatus := Status("invalid")
	err := ticket.Update(nil, nil, nil, nil, &invalidStatus, nil, nil, nil)
	if err != ErrInvalidStatus {
		t.Errorf("Update() error = %v, want ErrInvalidStatus", err)
	}
}

func TestNew_Types(t *testing.T) {
	tests := []struct {
		name     string
		ticketType Type
	}{
		{"bug", TypeBug},
		{"feature", TypeFeature},
		{"task", TypeTask},
		{"epic", TypeEpic},
		{"cleanup", TypeCleanup},
		{"custom", Type("custom")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket, err := New("TH", "Title", "", tt.ticketType, 1, nil, "")
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if ticket.Type != tt.ticketType {
				t.Errorf("ticket.Type = %q, want %q", ticket.Type, tt.ticketType)
			}
		})
	}
}

func TestUpdate_Type(t *testing.T) {
	tk, _ := New("TH", "Title", "", TypeTask, 1, nil, "")
	newType := TypeBug
	err := tk.Update(nil, nil, &newType, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if tk.Type != TypeBug {
		t.Errorf("ticket.Type = %q, want %q", tk.Type, TypeBug)
	}
}

func TestValidateLabel(t *testing.T) {
	tests := []struct {
		label   string
		wantErr bool
	}{
		{"bug", false},
		{"feature", false},
		{"high-priority", false},
		{"p1_urgent", false},
		{"A123", false},
		{"a", false},
		{"", true},
		{"has spaces", true},
		{"has.dot", true},
		{"has@symbol", true},
		{"this-label-is-way-too-long-to-be-valid", true},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			err := ValidateLabel(tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLabel(%q) error = %v, wantErr %v", tt.label, err, tt.wantErr)
			}
		})
	}
}

func TestNew_WithLabels(t *testing.T) {
	ticket, err := New("TH", "Test ticket", "Description", TypeTask, 1, []string{"bug", "urgent"}, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if len(ticket.Labels) != 2 {
		t.Errorf("New().Labels length = %d, want 2", len(ticket.Labels))
	}
	if ticket.Labels[0] != "bug" || ticket.Labels[1] != "urgent" {
		t.Errorf("New().Labels = %v, want [bug, urgent]", ticket.Labels)
	}
}

func TestNew_InvalidLabel(t *testing.T) {
	_, err := New("TH", "Test ticket", "Description", TypeTask, 1, []string{"valid", "has space"}, "")
	if err != ErrInvalidLabel {
		t.Errorf("New() error = %v, want ErrInvalidLabel", err)
	}
}

func TestTicket_Update_AddLabels(t *testing.T) {
	ticket := &Ticket{
		ID:     "TH-abcdef",
		Title:  "Test",
		Status: StatusOpen,
		Labels: []string{"existing"},
	}

	err := ticket.Update(nil, nil, nil, nil, nil, []string{"new-label"}, nil, nil)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if len(ticket.Labels) != 2 {
		t.Errorf("Update() Labels length = %d, want 2", len(ticket.Labels))
	}
}

func TestTicket_Update_RemoveLabels(t *testing.T) {
	ticket := &Ticket{
		ID:     "TH-abcdef",
		Title:  "Test",
		Status: StatusOpen,
		Labels: []string{"keep", "remove"},
	}

	err := ticket.Update(nil, nil, nil, nil, nil, nil, []string{"remove"}, nil)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if len(ticket.Labels) != 1 {
		t.Errorf("Update() Labels length = %d, want 1", len(ticket.Labels))
	}
	if ticket.Labels[0] != "keep" {
		t.Errorf("Update() Labels = %v, want [keep]", ticket.Labels)
	}
}

func TestTicket_Update_AddDuplicateLabel(t *testing.T) {
	ticket := &Ticket{
		ID:     "TH-abcdef",
		Title:  "Test",
		Status: StatusOpen,
		Labels: []string{"existing"},
	}

	err := ticket.Update(nil, nil, nil, nil, nil, []string{"existing"}, nil, nil)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Should not add duplicate
	if len(ticket.Labels) != 1 {
		t.Errorf("Update() Labels length = %d, want 1", len(ticket.Labels))
	}
}

func TestTicket_Update_InvalidAddLabel(t *testing.T) {
	ticket := &Ticket{
		ID:     "TH-abcdef",
		Title:  "Test",
		Status: StatusOpen,
	}

	err := ticket.Update(nil, nil, nil, nil, nil, []string{"has space"}, nil, nil)
	if err != ErrInvalidLabel {
		t.Errorf("Update() error = %v, want ErrInvalidLabel", err)
	}
}
