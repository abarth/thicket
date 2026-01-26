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
		tickets, comments, dependencies, err := ReadAllJSONL(s.paths.Tickets)
		if err != nil {
			return fmt.Errorf("reading JSONL: %w", err)
		}

		if err := s.db.RebuildFromAll(tickets, comments, dependencies); err != nil {
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
	// Read everything, update the matching ticket, and rewrite
	tickets, comments, dependencies, err := ReadAllJSONL(s.paths.Tickets)
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

	if err := WriteAllJSONL(s.paths.Tickets, tickets, comments, dependencies); err != nil {
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

// AddDependency creates a new dependency and persists it to both JSONL and SQLite.
// For blocked_by dependencies, it validates that no circular dependency would be created.
func (s *Store) AddDependency(d *ticket.Dependency) error {
	// Check if dependency already exists
	exists, err := s.db.DependencyExists(d.FromTicketID, d.ToTicketID, d.Type)
	if err != nil {
		return err
	}
	if exists {
		return ticket.ErrDuplicateDependency
	}

	// For blocked_by dependencies, check for circular dependencies
	if d.Type == ticket.DependencyBlockedBy {
		if err := s.checkCircularDependency(d.FromTicketID, d.ToTicketID); err != nil {
			return err
		}
	}

	if err := AppendDependency(s.paths.Tickets, d); err != nil {
		return err
	}

	if err := s.db.InsertDependency(d); err != nil {
		return err
	}

	return s.updateJSONLModTime()
}

// checkCircularDependency checks if adding a blocked_by dependency from fromID to toID
// would create a circular dependency. It traverses the blocked_by graph from toID
// to see if it can reach fromID.
func (s *Store) checkCircularDependency(fromID, toID string) error {
	// Get all blocking dependencies
	deps, err := s.db.GetBlockingDependencies()
	if err != nil {
		return err
	}

	// Build an adjacency list: blockedBy[A] = [B, C] means A is blocked by B and C
	blockedBy := make(map[string][]string)
	for _, d := range deps {
		blockedBy[d.FromTicketID] = append(blockedBy[d.FromTicketID], d.ToTicketID)
	}

	// Check if adding fromID -> toID creates a cycle
	// This would happen if toID transitively blocks fromID
	// (i.e., if we can reach fromID starting from toID through the blocked_by graph)
	visited := make(map[string]bool)
	var canReach func(current, target string) bool
	canReach = func(current, target string) bool {
		if current == target {
			return true
		}
		if visited[current] {
			return false
		}
		visited[current] = true

		for _, next := range blockedBy[current] {
			if canReach(next, target) {
				return true
			}
		}
		return false
	}

	if canReach(toID, fromID) {
		return ticket.ErrCircularDependency
	}

	return nil
}

// GetDependenciesFrom retrieves all dependencies from a specific ticket.
func (s *Store) GetDependenciesFrom(ticketID string) ([]*ticket.Dependency, error) {
	return s.db.GetDependenciesFrom(ticketID)
}

// GetDependenciesTo retrieves all dependencies pointing to a specific ticket.
func (s *Store) GetDependenciesTo(ticketID string) ([]*ticket.Dependency, error) {
	return s.db.GetDependenciesTo(ticketID)
}

// GetBlockers retrieves tickets that block the given ticket (blocked_by dependencies).
func (s *Store) GetBlockers(ticketID string) ([]*ticket.Ticket, error) {
	deps, err := s.db.GetDependenciesFrom(ticketID)
	if err != nil {
		return nil, err
	}

	var blockers []*ticket.Ticket
	for _, d := range deps {
		if d.Type == ticket.DependencyBlockedBy {
			t, err := s.db.GetTicket(d.ToTicketID)
			if err != nil {
				return nil, err
			}
			if t != nil {
				blockers = append(blockers, t)
			}
		}
	}
	return blockers, nil
}

// GetBlocking retrieves tickets that are blocked by the given ticket.
func (s *Store) GetBlocking(ticketID string) ([]*ticket.Ticket, error) {
	deps, err := s.db.GetDependenciesTo(ticketID)
	if err != nil {
		return nil, err
	}

	var blocking []*ticket.Ticket
	for _, d := range deps {
		if d.Type == ticket.DependencyBlockedBy {
			t, err := s.db.GetTicket(d.FromTicketID)
			if err != nil {
				return nil, err
			}
			if t != nil {
				blocking = append(blocking, t)
			}
		}
	}
	return blocking, nil
}

// GetCreatedFrom retrieves the parent ticket that this ticket was created from.
func (s *Store) GetCreatedFrom(ticketID string) (*ticket.Ticket, error) {
	deps, err := s.db.GetDependenciesFrom(ticketID)
	if err != nil {
		return nil, err
	}

	for _, d := range deps {
		if d.Type == ticket.DependencyCreatedFrom {
			return s.db.GetTicket(d.ToTicketID)
		}
	}
	return nil, nil
}

// IsBlocked checks if a ticket has any open blocking dependencies.
func (s *Store) IsBlocked(ticketID string) (bool, error) {
	blockers, err := s.GetBlockers(ticketID)
	if err != nil {
		return false, err
	}

	for _, b := range blockers {
		if b.Status == ticket.StatusOpen {
			return true, nil
		}
	}
	return false, nil
}
