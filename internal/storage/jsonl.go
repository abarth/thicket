// Package storage handles persistence of tickets to JSONL and SQLite.
package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/ticket"
)

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
