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

	if len(id) != 9 { // "TH-" + 6 hex chars
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
		{"TH-abcdeg", true}, // 'g' is not hex
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
	ticket, err := New("TH", "Test ticket", "A description", 1)
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
	if ticket.Created.IsZero() {
		t.Error("New().Created is zero")
	}
	if ticket.Updated.IsZero() {
		t.Error("New().Updated is zero")
	}
}

func TestNew_TrimSpace(t *testing.T) {
	ticket, err := New("TH", "  Test ticket  ", "  Description  ", 0)
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
	_, err := New("TH", "", "Description", 0)
	if err != ErrEmptyTitle {
		t.Errorf("New() error = %v, want ErrEmptyTitle", err)
	}

	_, err = New("TH", "   ", "Description", 0)
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

	err := ticket.Update(&newTitle, &newDesc, &newPriority, &newStatus)
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
	err := ticket.Update(&newTitle, nil, nil, nil)
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
	err := ticket.Update(&emptyTitle, nil, nil, nil)
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
	err := ticket.Update(nil, nil, nil, &invalidStatus)
	if err != ErrInvalidStatus {
		t.Errorf("Update() error = %v, want ErrInvalidStatus", err)
	}
}
