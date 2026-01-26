package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/abarth/thicket/internal/ticket"
)

func TestReadWriteJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	// Reading non-existent file should return nil
	tickets, err := ReadJSONL(path)
	if err != nil {
		t.Fatalf("ReadJSONL() error = %v", err)
	}
	if tickets != nil {
		t.Errorf("ReadJSONL() = %v, want nil", tickets)
	}

	// Write some tickets
	now := time.Now().UTC()
	testTickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "First", Description: "Desc 1", Status: ticket.StatusOpen, Priority: 1, Created: now, Updated: now},
		{ID: "TH-222222", Title: "Second", Description: "Desc 2", Status: ticket.StatusClosed, Priority: 2, Created: now, Updated: now},
	}

	if err := WriteJSONL(path, testTickets); err != nil {
		t.Fatalf("WriteJSONL() error = %v", err)
	}

	// Read them back
	read, err := ReadJSONL(path)
	if err != nil {
		t.Fatalf("ReadJSONL() error = %v", err)
	}

	if len(read) != 2 {
		t.Fatalf("ReadJSONL() returned %d tickets, want 2", len(read))
	}

	if read[0].ID != "TH-111111" || read[0].Title != "First" {
		t.Errorf("ReadJSONL()[0] = %+v, want ID=TH-111111, Title=First", read[0])
	}
	if read[1].ID != "TH-222222" || read[1].Title != "Second" {
		t.Errorf("ReadJSONL()[1] = %+v, want ID=TH-222222, Title=Second", read[1])
	}
}

func TestAppendJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	now := time.Now().UTC()
	t1 := &ticket.Ticket{ID: "TH-111111", Title: "First", Status: ticket.StatusOpen, Priority: 1, Created: now, Updated: now}
	t2 := &ticket.Ticket{ID: "TH-222222", Title: "Second", Status: ticket.StatusOpen, Priority: 2, Created: now, Updated: now}

	if err := AppendJSONL(path, t1); err != nil {
		t.Fatalf("AppendJSONL() error = %v", err)
	}

	if err := AppendJSONL(path, t2); err != nil {
		t.Fatalf("AppendJSONL() error = %v", err)
	}

	tickets, err := ReadJSONL(path)
	if err != nil {
		t.Fatalf("ReadJSONL() error = %v", err)
	}

	if len(tickets) != 2 {
		t.Fatalf("ReadJSONL() returned %d tickets, want 2", len(tickets))
	}

	if tickets[0].ID != "TH-111111" {
		t.Errorf("tickets[0].ID = %q, want TH-111111", tickets[0].ID)
	}
	if tickets[1].ID != "TH-222222" {
		t.Errorf("tickets[1].ID = %q, want TH-222222", tickets[1].ID)
	}
}

func TestReadJSONL_EmptyLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	content := `{"id":"TH-111111","title":"First","description":"","status":"open","priority":1,"created":"2024-01-01T00:00:00Z","updated":"2024-01-01T00:00:00Z"}

{"id":"TH-222222","title":"Second","description":"","status":"open","priority":2,"created":"2024-01-01T00:00:00Z","updated":"2024-01-01T00:00:00Z"}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tickets, err := ReadJSONL(path)
	if err != nil {
		t.Fatalf("ReadJSONL() error = %v", err)
	}

	if len(tickets) != 2 {
		t.Fatalf("ReadJSONL() returned %d tickets, want 2", len(tickets))
	}
}

func TestReadJSONL_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	content := `{"id":"TH-111111","title":"First"}
invalid json
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ReadJSONL(path)
	if err == nil {
		t.Error("ReadJSONL() expected error for invalid JSON")
	}
}

func TestGetJSONLModTime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	// Non-existent file should return 0
	modTime, err := GetJSONLModTime(path)
	if err != nil {
		t.Fatalf("GetJSONLModTime() error = %v", err)
	}
	if modTime != 0 {
		t.Errorf("GetJSONLModTime() = %d, want 0", modTime)
	}

	// Create file
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	modTime, err = GetJSONLModTime(path)
	if err != nil {
		t.Fatalf("GetJSONLModTime() error = %v", err)
	}
	if modTime == 0 {
		t.Error("GetJSONLModTime() = 0, want non-zero")
	}
}

func TestReadAllJSONL_MixedContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	now := time.Now().UTC()
	content := `{"id":"TH-111111","title":"First","description":"","status":"open","priority":1,"created":"2024-01-01T00:00:00Z","updated":"2024-01-01T00:00:00Z"}
{"id":"TH-c222222","ticket_id":"TH-111111","content":"A comment","created":"2024-01-01T00:01:00Z"}
{"id":"TH-333333","title":"Second","description":"","status":"open","priority":2,"created":"2024-01-01T00:00:00Z","updated":"2024-01-01T00:00:00Z"}
{"id":"TH-c444444","ticket_id":"TH-111111","content":"Another comment","created":"2024-01-01T00:02:00Z"}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tickets, comments, deps, err := ReadAllJSONL(path)
	if err != nil {
		t.Fatalf("ReadAllJSONL() error = %v", err)
	}

	if len(tickets) != 2 {
		t.Errorf("ReadAllJSONL() returned %d tickets, want 2", len(tickets))
	}
	if len(comments) != 2 {
		t.Errorf("ReadAllJSONL() returned %d comments, want 2", len(comments))
	}
	if len(deps) != 0 {
		t.Errorf("ReadAllJSONL() returned %d dependencies, want 0", len(deps))
	}

	if tickets[0].ID != "TH-111111" {
		t.Errorf("tickets[0].ID = %q, want TH-111111", tickets[0].ID)
	}
	if tickets[1].ID != "TH-333333" {
		t.Errorf("tickets[1].ID = %q, want TH-333333", tickets[1].ID)
	}
	if comments[0].ID != "TH-c222222" {
		t.Errorf("comments[0].ID = %q, want TH-c222222", comments[0].ID)
	}
	if comments[0].TicketID != "TH-111111" {
		t.Errorf("comments[0].TicketID = %q, want TH-111111", comments[0].TicketID)
	}
	if comments[0].Content != "A comment" {
		t.Errorf("comments[0].Content = %q, want 'A comment'", comments[0].Content)
	}

	_ = now // suppress unused variable warning
}

func TestReadAllJSONL_NonExistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")

	tickets, comments, deps, err := ReadAllJSONL(path)
	if err != nil {
		t.Fatalf("ReadAllJSONL() error = %v", err)
	}
	if tickets != nil {
		t.Errorf("ReadAllJSONL() tickets = %v, want nil", tickets)
	}
	if comments != nil {
		t.Errorf("ReadAllJSONL() comments = %v, want nil", comments)
	}
	if deps != nil {
		t.Errorf("ReadAllJSONL() dependencies = %v, want nil", deps)
	}
}

func TestAppendComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	now := time.Now().UTC()
	c := &ticket.Comment{
		ID:       "TH-cabcdef",
		TicketID: "TH-111111",
		Content:  "Test comment",
		Created:  now,
	}

	if err := AppendComment(path, c); err != nil {
		t.Fatalf("AppendComment() error = %v", err)
	}

	_, comments, _, err := ReadAllJSONL(path)
	if err != nil {
		t.Fatalf("ReadAllJSONL() error = %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("ReadAllJSONL() returned %d comments, want 1", len(comments))
	}

	if comments[0].ID != "TH-cabcdef" {
		t.Errorf("comments[0].ID = %q, want TH-cabcdef", comments[0].ID)
	}
	if comments[0].TicketID != "TH-111111" {
		t.Errorf("comments[0].TicketID = %q, want TH-111111", comments[0].TicketID)
	}
	if comments[0].Content != "Test comment" {
		t.Errorf("comments[0].Content = %q, want 'Test comment'", comments[0].Content)
	}
}

func TestWriteAllJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tickets.jsonl")

	now := time.Now().UTC()
	tickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "First", Status: ticket.StatusOpen, Priority: 1, Created: now, Updated: now},
	}
	comments := []*ticket.Comment{
		{ID: "TH-cabcdef", TicketID: "TH-111111", Content: "A comment", Created: now},
	}

	if err := WriteAllJSONL(path, tickets, comments, nil); err != nil {
		t.Fatalf("WriteAllJSONL() error = %v", err)
	}

	readTickets, readComments, _, err := ReadAllJSONL(path)
	if err != nil {
		t.Fatalf("ReadAllJSONL() error = %v", err)
	}

	if len(readTickets) != 1 {
		t.Errorf("ReadAllJSONL() returned %d tickets, want 1", len(readTickets))
	}
	if len(readComments) != 1 {
		t.Errorf("ReadAllJSONL() returned %d comments, want 1", len(readComments))
	}

	if readTickets[0].ID != "TH-111111" {
		t.Errorf("readTickets[0].ID = %q, want TH-111111", readTickets[0].ID)
	}
	if readComments[0].ID != "TH-cabcdef" {
		t.Errorf("readComments[0].ID = %q, want TH-cabcdef", readComments[0].ID)
	}
}
