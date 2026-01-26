package errors

import (
	"strings"
	"testing"
)

func TestUserError_Error(t *testing.T) {
	err := New("something went wrong")
	if err.Error() != "something went wrong" {
		t.Errorf("Error() = %q, want 'something went wrong'", err.Error())
	}
}

func TestUserError_WithHint(t *testing.T) {
	err := WithHint("something went wrong", "try doing X")
	msg := err.Error()
	if !strings.Contains(msg, "something went wrong") {
		t.Errorf("Error() should contain message, got %q", msg)
	}
	if !strings.Contains(msg, "try doing X") {
		t.Errorf("Error() should contain hint, got %q", msg)
	}
}

func TestNotInitialized(t *testing.T) {
	err := NotInitialized()
	msg := err.Error()
	if !strings.Contains(msg, "not initialized") {
		t.Errorf("Error() should mention 'not initialized', got %q", msg)
	}
	if !strings.Contains(msg, "thicket init") {
		t.Errorf("Error() should mention 'thicket init', got %q", msg)
	}
}

func TestTicketNotFound(t *testing.T) {
	err := TicketNotFound("TH-abc123")
	msg := err.Error()
	if !strings.Contains(msg, "TH-abc123") {
		t.Errorf("Error() should contain ticket ID, got %q", msg)
	}
	if !strings.Contains(msg, "not found") {
		t.Errorf("Error() should mention 'not found', got %q", msg)
	}
}

func TestInvalidTicketID(t *testing.T) {
	err := InvalidTicketID("bad-id")
	msg := err.Error()
	if !strings.Contains(msg, "bad-id") {
		t.Errorf("Error() should contain invalid ID, got %q", msg)
	}
	if !strings.Contains(msg, "XX-xxxxxx") {
		t.Errorf("Error() should show correct format, got %q", msg)
	}
}

func TestInvalidProjectCode(t *testing.T) {
	err := InvalidProjectCode("XYZ")
	msg := err.Error()
	if !strings.Contains(msg, "XYZ") {
		t.Errorf("Error() should contain invalid code, got %q", msg)
	}
	if !strings.Contains(msg, "two letters") {
		t.Errorf("Error() should mention format, got %q", msg)
	}
}

func TestMissingRequired(t *testing.T) {
	err := MissingRequired("title")
	msg := err.Error()
	if !strings.Contains(msg, "--title") {
		t.Errorf("Error() should contain flag name, got %q", msg)
	}
}

func TestInvalidStatus(t *testing.T) {
	err := InvalidStatus("pending")
	msg := err.Error()
	if !strings.Contains(msg, "pending") {
		t.Errorf("Error() should contain invalid status, got %q", msg)
	}
	if !strings.Contains(msg, "open") && !strings.Contains(msg, "closed") {
		t.Errorf("Error() should mention valid statuses, got %q", msg)
	}
}
