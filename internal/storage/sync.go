package storage

import (
	"fmt"
	"strconv"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/ticket"
)

const metaKeyJSONLModTime = "jsonl_modtime"

// Store provides synchronized access to ticket storage.
type Store struct {
	db    *DB
	paths config.Paths
}

// Open creates a new Store, opening the SQLite database and syncing from JSONL if needed.
func Open(paths config.Paths) (*Store, error) {
	db, err := OpenDB(paths.Cache)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db, paths: paths}

	if err := store.syncFromJSONL(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

// syncFromJSONL checks if the JSONL file has been modified and rebuilds the cache if needed.
func (s *Store) syncFromJSONL() error {
	currentModTime, err := GetJSONLModTime(s.paths.Tickets)
	if err != nil {
		return fmt.Errorf("getting JSONL mod time: %w", err)
	}

	storedModTimeStr, err := s.db.GetMetadata(metaKeyJSONLModTime)
	if err != nil {
		return fmt.Errorf("getting stored mod time: %w", err)
	}

	var storedModTime int64
	if storedModTimeStr != "" {
		storedModTime, _ = strconv.ParseInt(storedModTimeStr, 10, 64)
	}

	if currentModTime != storedModTime {
		tickets, comments, err := ReadAllJSONL(s.paths.Tickets)
		if err != nil {
			return fmt.Errorf("reading JSONL: %w", err)
		}

		if err := s.db.RebuildFromAll(tickets, comments); err != nil {
			return fmt.Errorf("rebuilding cache: %w", err)
		}

		if err := s.db.SetMetadata(metaKeyJSONLModTime, strconv.FormatInt(currentModTime, 10)); err != nil {
			return fmt.Errorf("storing mod time: %w", err)
		}
	}

	return nil
}

// updateJSONLModTime updates the stored modification time after a write.
func (s *Store) updateJSONLModTime() error {
	modTime, err := GetJSONLModTime(s.paths.Tickets)
	if err != nil {
		return err
	}
	return s.db.SetMetadata(metaKeyJSONLModTime, strconv.FormatInt(modTime, 10))
}

// Add creates a new ticket and persists it to both JSONL and SQLite.
func (s *Store) Add(t *ticket.Ticket) error {
	if err := AppendJSONL(s.paths.Tickets, t); err != nil {
		return err
	}

	if err := s.db.InsertTicket(t); err != nil {
		return err
	}

	return s.updateJSONLModTime()
}

// Update modifies an existing ticket in both JSONL and SQLite.
func (s *Store) Update(t *ticket.Ticket) error {
	// Read all tickets, update the matching one, and rewrite
	tickets, err := ReadJSONL(s.paths.Tickets)
	if err != nil {
		return err
	}

	found := false
	for i, existing := range tickets {
		if existing.ID == t.ID {
			tickets[i] = t
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("ticket %s not found", t.ID)
	}

	if err := WriteJSONL(s.paths.Tickets, tickets); err != nil {
		return err
	}

	if err := s.db.UpdateTicket(t); err != nil {
		return err
	}

	return s.updateJSONLModTime()
}

// Get retrieves a ticket by ID.
func (s *Store) Get(id string) (*ticket.Ticket, error) {
	return s.db.GetTicket(id)
}

// List retrieves tickets with optional status filter.
func (s *Store) List(status *ticket.Status) ([]*ticket.Ticket, error) {
	return s.db.ListTickets(status)
}

// AddComment creates a new comment and persists it to both JSONL and SQLite.
func (s *Store) AddComment(c *ticket.Comment) error {
	if err := AppendComment(s.paths.Tickets, c); err != nil {
		return err
	}

	if err := s.db.InsertComment(c); err != nil {
		return err
	}

	return s.updateJSONLModTime()
}

// GetComments retrieves all comments for a ticket.
func (s *Store) GetComments(ticketID string) ([]*ticket.Comment, error) {
	return s.db.GetCommentsForTicket(ticketID)
}
