package ticket

import (
	"strings"
	"testing"
)

func TestGenerateCommentID(t *testing.T) {
	id, err := GenerateCommentID("TH")
	if err != nil {
		t.Fatalf("GenerateCommentID() error = %v", err)
	}

	if !strings.HasPrefix(id, "TH-c") {
		t.Errorf("GenerateCommentID() = %q, want prefix TH-c", id)
	}

	if len(id) != 10 { // "TH-c" + 6 alphanumeric chars
		t.Errorf("GenerateCommentID() = %q, want length 10", id)
	}

	// Generate another ID to ensure uniqueness
	id2, err := GenerateCommentID("TH")
	if err != nil {
		t.Fatalf("GenerateCommentID() error = %v", err)
	}

	if id == id2 {
		t.Errorf("GenerateCommentID() generated duplicate IDs: %q", id)
	}
}

func TestGenerateCommentID_InvalidCode(t *testing.T) {
	_, err := GenerateCommentID("invalid")
	if err == nil {
		t.Error("GenerateCommentID(invalid) expected error, got nil")
	}
}

func TestValidateCommentID(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{"TH-cabcdef", false},
		{"AB-c123456", false},
		{"ZZ-ca1b2c3", false},
		{"", true},
		{"TH-c", true},
		{"TH-cabc", true},
		{"TH-cabcdefg", true},
		{"th-cabcdef", true},
		{"TH-cABCDEF", true},
		{"TH-cabcdeg", false},
		{"TH-cz1y2x3", false},
		{"TH-abcdef", true},  // missing 'c' prefix
		{"THX-cabcdef", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			err := ValidateCommentID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommentID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestNewComment(t *testing.T) {
	comment, err := NewComment("TH-abcdef", "Test comment content")
	if err != nil {
		t.Fatalf("NewComment() error = %v", err)
	}

	if !strings.HasPrefix(comment.ID, "TH-c") {
		t.Errorf("NewComment().ID = %q, want prefix TH-c", comment.ID)
	}
	if comment.TicketID != "TH-abcdef" {
		t.Errorf("NewComment().TicketID = %q, want 'TH-abcdef'", comment.TicketID)
	}
	if comment.Content != "Test comment content" {
		t.Errorf("NewComment().Content = %q, want 'Test comment content'", comment.Content)
	}
	if comment.Created.IsZero() {
		t.Error("NewComment().Created is zero")
	}
}

func TestNewComment_TrimSpace(t *testing.T) {
	comment, err := NewComment("TH-abcdef", "  Comment content  ")
	if err != nil {
		t.Fatalf("NewComment() error = %v", err)
	}

	if comment.Content != "Comment content" {
		t.Errorf("NewComment().Content = %q, want 'Comment content'", comment.Content)
	}
}

func TestNewComment_EmptyContent(t *testing.T) {
	_, err := NewComment("TH-abcdef", "")
	if err != ErrEmptyComment {
		t.Errorf("NewComment() error = %v, want ErrEmptyComment", err)
	}

	_, err = NewComment("TH-abcdef", "   ")
	if err != ErrEmptyComment {
		t.Errorf("NewComment() error = %v, want ErrEmptyComment", err)
	}
}

func TestNewComment_InvalidTicketID(t *testing.T) {
	_, err := NewComment("invalid", "Content")
	if err == nil {
		t.Error("NewComment() expected error for invalid ticket ID")
	}
}

func TestComment_Validate(t *testing.T) {
	comment := &Comment{
		ID:       "TH-cabcdef",
		TicketID: "TH-abcdef",
		Content:  "Test",
	}

	if err := comment.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	comment.ID = "invalid"
	if err := comment.Validate(); err == nil {
		t.Error("Validate() expected error for invalid comment ID")
	}

	comment.ID = "TH-cabcdef"
	comment.TicketID = "invalid"
	if err := comment.Validate(); err == nil {
		t.Error("Validate() expected error for invalid ticket ID")
	}

	comment.TicketID = "TH-abcdef"
	comment.Content = ""
	if err := comment.Validate(); err == nil {
		t.Error("Validate() expected error for empty content")
	}
}
