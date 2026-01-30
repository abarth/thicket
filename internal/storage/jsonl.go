// Package storage handles persistence of tickets to JSONL and SQLite.
package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/abarth/thicket/internal/ticket"
)

// rawRecord is used to detect whether a JSON line is a ticket, comment, or dependency.
type rawRecord struct {
	TicketID     string `json:"ticket_id"`      // Present for comments
	FromTicketID string `json:"from_ticket_id"` // Present for dependencies
}

// ReadJSONL reads all tickets from a JSONL file, ignoring comments and dependencies.
func ReadJSONL(path string) ([]*ticket.Ticket, error) {
	tickets, _, _, err := ReadAllJSONL(path)
	return tickets, err
}

// AppendJSONL appends a single ticket to the JSONL file by rewriting it sorted.
func AppendJSONL(path string, t *ticket.Ticket) error {
	tickets, comments, dependencies, err := ReadAllJSONL(path)
	if err != nil {
		return err
	}
	tickets = append(tickets, t)
	return WriteAllJSONL(path, tickets, comments, dependencies)
}

// WriteJSONL writes all tickets to a JSONL file, replacing existing content and sorting by ID.
func WriteJSONL(path string, tickets []*ticket.Ticket) error {
	sort.Slice(tickets, func(i, j int) bool {
		return tickets[i].ID < tickets[j].ID
	})

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

// ReadAllJSONL reads all tickets, comments, and dependencies from a JSONL file.
// It distinguishes between record types by checking for specific fields:
// - Dependencies have from_ticket_id
// - Comments have ticket_id
// - Tickets have neither
func ReadAllJSONL(path string) ([]*ticket.Ticket, []*ticket.Comment, []*ticket.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("opening tickets file: %w", err)
	}
	defer file.Close()

	var tickets []*ticket.Ticket
	var comments []*ticket.Comment
	var dependencies []*ticket.Dependency
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		// First, check the record type by looking at specific fields
		var raw rawRecord
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, nil, nil, fmt.Errorf("parsing line %d: %w", lineNum, err)
		}

		if raw.FromTicketID != "" {
			// This is a dependency
			var d ticket.Dependency
			if err := json.Unmarshal([]byte(line), &d); err != nil {
				return nil, nil, nil, fmt.Errorf("parsing dependency at line %d: %w", lineNum, err)
			}
			dependencies = append(dependencies, &d)
		} else if raw.TicketID != "" {
			// This is a comment
			var c ticket.Comment
			if err := json.Unmarshal([]byte(line), &c); err != nil {
				return nil, nil, nil, fmt.Errorf("parsing comment at line %d: %w", lineNum, err)
			}
			comments = append(comments, &c)
		} else {
			// This is a ticket
			var t ticket.Ticket
			if err := json.Unmarshal([]byte(line), &t); err != nil {
				return nil, nil, nil, fmt.Errorf("parsing ticket at line %d: %w", lineNum, err)
			}
			tickets = append(tickets, &t)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("reading tickets file: %w", err)
	}

	return tickets, comments, dependencies, nil
}

// AppendComment appends a single comment to the JSONL file by rewriting it sorted.
func AppendComment(path string, c *ticket.Comment) error {
	tickets, comments, dependencies, err := ReadAllJSONL(path)
	if err != nil {
		return err
	}
	comments = append(comments, c)
	return WriteAllJSONL(path, tickets, comments, dependencies)
}

// AppendDependency appends a single dependency to the JSONL file by rewriting it sorted.
func AppendDependency(path string, d *ticket.Dependency) error {
	tickets, comments, dependencies, err := ReadAllJSONL(path)
	if err != nil {
		return err
	}
	dependencies = append(dependencies, d)
	return WriteAllJSONL(path, tickets, comments, dependencies)
}

// WriteAllJSONL writes all tickets, comments, and dependencies to a JSONL file, replacing existing content and sorting by ID.
func WriteAllJSONL(path string, tickets []*ticket.Ticket, comments []*ticket.Comment, dependencies []*ticket.Dependency) error {
	// Sort everything by ID to reduce merge conflicts
	sort.Slice(tickets, func(i, j int) bool {
		return tickets[i].ID < tickets[j].ID
	})
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].ID < comments[j].ID
	})
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].ID < dependencies[j].ID
	})

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

	for _, d := range dependencies {
		data, err := json.Marshal(d)
		if err != nil {
			return fmt.Errorf("encoding dependency %s: %w", d.ID, err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("writing dependency %s: %w", d.ID, err)
		}
	}

	return nil
}
