// Package storage handles persistence of tickets to JSONL and SQLite.
package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/ticket"
)

// rawRecord is used to detect whether a JSON line is a ticket or comment.
type rawRecord struct {
	TicketID string `json:"ticket_id"`
}

// ReadJSONL reads all tickets from a JSONL file.
func ReadJSONL(path string) ([]*ticket.Ticket, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening tickets file: %w", err)
	}
	defer file.Close()

	var tickets []*ticket.Ticket
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var t ticket.Ticket
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			return nil, fmt.Errorf("parsing line %d: %w", lineNum, err)
		}
		tickets = append(tickets, &t)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading tickets file: %w", err)
	}

	return tickets, nil
}

// AppendJSONL appends a single ticket to the JSONL file.
func AppendJSONL(path string, t *ticket.Ticket) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening tickets file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("encoding ticket: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing ticket: %w", err)
	}

	return nil
}

// WriteJSONL writes all tickets to a JSONL file, replacing existing content.
func WriteJSONL(path string, tickets []*ticket.Ticket) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating tickets file: %w", err)
	}
	defer file.Close()

	for _, t := range tickets {
		data, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("encoding ticket %s: %w", t.ID, err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("writing ticket %s: %w", t.ID, err)
		}
	}

	return nil
}

// GetJSONLModTime returns the modification time of the JSONL file.
func GetJSONLModTime(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("getting file info: %w", err)
	}
	return info.ModTime().UnixNano(), nil
}

// ReadAllJSONL reads all tickets and comments from a JSONL file.
// It distinguishes between tickets and comments by checking for the ticket_id field.
func ReadAllJSONL(path string) ([]*ticket.Ticket, []*ticket.Comment, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("opening tickets file: %w", err)
	}
	defer file.Close()

	var tickets []*ticket.Ticket
	var comments []*ticket.Comment
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		// First, check if this is a comment (has ticket_id field)
		var raw rawRecord
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, nil, fmt.Errorf("parsing line %d: %w", lineNum, err)
		}

		if raw.TicketID != "" {
			// This is a comment
			var c ticket.Comment
			if err := json.Unmarshal([]byte(line), &c); err != nil {
				return nil, nil, fmt.Errorf("parsing comment at line %d: %w", lineNum, err)
			}
			comments = append(comments, &c)
		} else {
			// This is a ticket
			var t ticket.Ticket
			if err := json.Unmarshal([]byte(line), &t); err != nil {
				return nil, nil, fmt.Errorf("parsing ticket at line %d: %w", lineNum, err)
			}
			tickets = append(tickets, &t)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("reading tickets file: %w", err)
	}

	return tickets, comments, nil
}

// AppendComment appends a single comment to the JSONL file.
func AppendComment(path string, c *ticket.Comment) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening tickets file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("encoding comment: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing comment: %w", err)
	}

	return nil
}

// WriteAllJSONL writes all tickets and comments to a JSONL file, replacing existing content.
func WriteAllJSONL(path string, tickets []*ticket.Ticket, comments []*ticket.Comment) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating tickets file: %w", err)
	}
	defer file.Close()

	for _, t := range tickets {
		data, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("encoding ticket %s: %w", t.ID, err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("writing ticket %s: %w", t.ID, err)
		}
	}

	for _, c := range comments {
		data, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("encoding comment %s: %w", c.ID, err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("writing comment %s: %w", c.ID, err)
		}
	}

	return nil
}
