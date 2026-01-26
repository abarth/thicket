package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/abarth/thicket/internal/ticket"
)

const schema = `
CREATE TABLE IF NOT EXISTS tickets (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    priority INTEGER NOT NULL DEFAULT 0,
    created TEXT NOT NULL,
    updated TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);

CREATE TABLE IF NOT EXISTS ticket_labels (
    ticket_id TEXT NOT NULL,
    label TEXT NOT NULL,
    PRIMARY KEY (ticket_id, label)
);

CREATE INDEX IF NOT EXISTS idx_ticket_labels_label ON ticket_labels(label);

CREATE TABLE IF NOT EXISTS comments (
    id TEXT PRIMARY KEY,
    ticket_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_comments_ticket_id ON comments(ticket_id);

CREATE TABLE IF NOT EXISTS dependencies (
    id TEXT PRIMARY KEY,
    from_ticket_id TEXT NOT NULL,
    to_ticket_id TEXT NOT NULL,
    type TEXT NOT NULL,
    created TEXT NOT NULL,
    UNIQUE(from_ticket_id, to_ticket_id, type)
);

CREATE INDEX IF NOT EXISTS idx_dependencies_from ON dependencies(from_ticket_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_to ON dependencies(to_ticket_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_type ON dependencies(type);

CREATE TABLE IF NOT EXISTS metadata (
    key TEXT PRIMARY KEY,
    value TEXT
);
`

// DB wraps a SQLite database connection for ticket operations.
type DB struct {
	conn *sql.DB
	path string
}

// OpenDB opens or creates a SQLite database at the given path.
func OpenDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	return &DB{conn: conn, path: path}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetMetadata retrieves a metadata value by key.
func (db *DB) GetMetadata(key string) (string, error) {
	var value string
	err := db.conn.QueryRow("SELECT value FROM metadata WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("getting metadata %s: %w", key, err)
	}
	return value, nil
}

// SetMetadata stores a metadata key-value pair.
func (db *DB) SetMetadata(key, value string) error {
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO metadata (key, value) VALUES (?, ?)",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("setting metadata %s: %w", key, err)
	}
	return nil
}

// RebuildFromTickets clears all tickets and inserts the given list.
func (db *DB) RebuildFromTickets(tickets []*ticket.Ticket) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM tickets"); err != nil {
		return fmt.Errorf("clearing tickets: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM ticket_labels"); err != nil {
		return fmt.Errorf("clearing ticket labels: %w", err)
	}

	ticketStmt, err := tx.Prepare(`
		INSERT INTO tickets (id, title, description, status, priority, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing insert: %w", err)
	}
	defer ticketStmt.Close()

	labelStmt, err := tx.Prepare(`INSERT INTO ticket_labels (ticket_id, label) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing label insert: %w", err)
	}
	defer labelStmt.Close()

	for _, t := range tickets {
		_, err := ticketStmt.Exec(
			t.ID,
			t.Title,
			t.Description,
			string(t.Status),
			t.Priority,
			t.Created.Format(time.RFC3339Nano),
			t.Updated.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("inserting ticket %s: %w", t.ID, err)
		}

		for _, label := range t.Labels {
			if _, err := labelStmt.Exec(t.ID, label); err != nil {
				return fmt.Errorf("inserting label for ticket %s: %w", t.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// InsertTicket adds a new ticket to the database.
func (db *DB) InsertTicket(t *ticket.Ticket) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO tickets (id, title, description, status, priority, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		t.ID,
		t.Title,
		t.Description,
		string(t.Status),
		t.Priority,
		t.Created.Format(time.RFC3339Nano),
		t.Updated.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("inserting ticket: %w", err)
	}

	for _, label := range t.Labels {
		_, err = tx.Exec(`INSERT INTO ticket_labels (ticket_id, label) VALUES (?, ?)`, t.ID, label)
		if err != nil {
			return fmt.Errorf("inserting label: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// UpdateTicket updates an existing ticket in the database.
func (db *DB) UpdateTicket(t *ticket.Ticket) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE tickets
		SET title = ?, description = ?, status = ?, priority = ?, updated = ?
		WHERE id = ?
	`,
		t.Title,
		t.Description,
		string(t.Status),
		t.Priority,
		t.Updated.Format(time.RFC3339Nano),
		t.ID,
	)
	if err != nil {
		return fmt.Errorf("updating ticket: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("ticket %s not found", t.ID)
	}

	// Replace labels
	_, err = tx.Exec(`DELETE FROM ticket_labels WHERE ticket_id = ?`, t.ID)
	if err != nil {
		return fmt.Errorf("deleting labels: %w", err)
	}

	for _, label := range t.Labels {
		_, err = tx.Exec(`INSERT INTO ticket_labels (ticket_id, label) VALUES (?, ?)`, t.ID, label)
		if err != nil {
			return fmt.Errorf("inserting label: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// GetTicket retrieves a ticket by ID.
func (db *DB) GetTicket(id string) (*ticket.Ticket, error) {
	var t ticket.Ticket
	var status string
	var created, updated string

	err := db.conn.QueryRow(`
		SELECT id, title, description, status, priority, created, updated
		FROM tickets WHERE id = ?
	`, id).Scan(&t.ID, &t.Title, &t.Description, &status, &t.Priority, &created, &updated)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying ticket: %w", err)
	}

	t.Status = ticket.Status(status)
	t.Created, _ = time.Parse(time.RFC3339Nano, created)
	t.Updated, _ = time.Parse(time.RFC3339Nano, updated)

	// Fetch labels
	labels, err := db.getLabelsForTicket(id)
	if err != nil {
		return nil, err
	}
	t.Labels = labels

	return &t, nil
}

// getLabelsForTicket retrieves all labels for a ticket.
func (db *DB) getLabelsForTicket(ticketID string) ([]string, error) {
	rows, err := db.conn.Query(`SELECT label FROM ticket_labels WHERE ticket_id = ? ORDER BY label`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("querying labels: %w", err)
	}
	defer rows.Close()

	var labels []string
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating labels: %w", err)
	}

	return labels, nil
}

// ListTickets retrieves tickets with optional status filter, ordered by priority.
func (db *DB) ListTickets(status *ticket.Status) ([]*ticket.Ticket, error) {
	var rows *sql.Rows
	var err error

	if status != nil {
		rows, err = db.conn.Query(`
			SELECT id, title, description, status, priority, created, updated
			FROM tickets WHERE status = ?
			ORDER BY priority ASC, created ASC
		`, string(*status))
	} else {
		rows, err = db.conn.Query(`
			SELECT id, title, description, status, priority, created, updated
			FROM tickets
			ORDER BY priority ASC, created ASC
		`)
	}

	if err != nil {
		return nil, fmt.Errorf("querying tickets: %w", err)
	}
	defer rows.Close()

	tickets, err := scanTickets(rows)
	if err != nil {
		return nil, err
	}

	if err := db.loadLabelsForTickets(tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

// ListReadyTickets retrieves open tickets that are not blocked by other open tickets.
func (db *DB) ListReadyTickets() ([]*ticket.Ticket, error) {
	rows, err := db.conn.Query(`
		SELECT t.id, t.title, t.description, t.status, t.priority, t.created, t.updated
		FROM tickets t
		WHERE t.status = 'open'
		AND NOT EXISTS (
			SELECT 1
			FROM dependencies d
			JOIN tickets bt ON d.to_ticket_id = bt.id
			WHERE d.from_ticket_id = t.id
			AND d.type = 'blocked_by'
			AND bt.status = 'open'
		)
		ORDER BY t.priority ASC, t.created ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying ready tickets: %w", err)
	}
	defer rows.Close()

	tickets, err := scanTickets(rows)
	if err != nil {
		return nil, err
	}

	if err := db.loadLabelsForTickets(tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

// ListTicketsByLabel retrieves tickets that have the specified label.
func (db *DB) ListTicketsByLabel(label string, status *ticket.Status) ([]*ticket.Ticket, error) {
	var rows *sql.Rows
	var err error

	if status != nil {
		rows, err = db.conn.Query(`
			SELECT t.id, t.title, t.description, t.status, t.priority, t.created, t.updated
			FROM tickets t
			JOIN ticket_labels tl ON t.id = tl.ticket_id
			WHERE tl.label = ? AND t.status = ?
			ORDER BY t.priority ASC, t.created ASC
		`, label, string(*status))
	} else {
		rows, err = db.conn.Query(`
			SELECT t.id, t.title, t.description, t.status, t.priority, t.created, t.updated
			FROM tickets t
			JOIN ticket_labels tl ON t.id = tl.ticket_id
			WHERE tl.label = ?
			ORDER BY t.priority ASC, t.created ASC
		`, label)
	}

	if err != nil {
		return nil, fmt.Errorf("querying tickets by label: %w", err)
	}
	defer rows.Close()

	tickets, err := scanTickets(rows)
	if err != nil {
		return nil, err
	}

	if err := db.loadLabelsForTickets(tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

// loadLabelsForTickets fetches and populates labels for a slice of tickets.
func (db *DB) loadLabelsForTickets(tickets []*ticket.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}

	// Build a map for quick lookup
	ticketMap := make(map[string]*ticket.Ticket)
	for _, t := range tickets {
		ticketMap[t.ID] = t
	}

	// Fetch all labels for these tickets in one query
	rows, err := db.conn.Query(`
		SELECT ticket_id, label FROM ticket_labels
		WHERE ticket_id IN (SELECT id FROM tickets)
		ORDER BY ticket_id, label
	`)
	if err != nil {
		return fmt.Errorf("querying labels: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ticketID, label string
		if err := rows.Scan(&ticketID, &label); err != nil {
			return fmt.Errorf("scanning label: %w", err)
		}
		if t, ok := ticketMap[ticketID]; ok {
			t.Labels = append(t.Labels, label)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating labels: %w", err)
	}

	return nil
}

func scanTickets(rows *sql.Rows) ([]*ticket.Ticket, error) {
	var tickets []*ticket.Ticket
	for rows.Next() {
		var t ticket.Ticket
		var statusStr string
		var created, updated string

		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &statusStr, &t.Priority, &created, &updated); err != nil {
			return nil, fmt.Errorf("scanning ticket: %w", err)
		}

		t.Status = ticket.Status(statusStr)
		createdTime, err := time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, fmt.Errorf("parsing ticket created time: %w", err)
		}
		t.Created = createdTime

		updatedTime, err := time.Parse(time.RFC3339Nano, updated)
		if err != nil {
			return nil, fmt.Errorf("parsing ticket updated time: %w", err)
		}
		t.Updated = updatedTime

		tickets = append(tickets, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tickets: %w", err)
	}

	return tickets, nil
}

// GetAllTickets retrieves all tickets from the database.
func (db *DB) GetAllTickets() ([]*ticket.Ticket, error) {
	return db.ListTickets(nil)
}

// GetAllComments retrieves all comments from the database.
func (db *DB) GetAllComments() ([]*ticket.Comment, error) {
	rows, err := db.conn.Query(`
		SELECT id, ticket_id, content, created
		FROM comments
		ORDER BY created ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying comments: %w", err)
	}
	defer rows.Close()

	var comments []*ticket.Comment
	for rows.Next() {
		var c ticket.Comment
		var created string

		if err := rows.Scan(&c.ID, &c.TicketID, &c.Content, &created); err != nil {
			return nil, fmt.Errorf("scanning comment: %w", err)
		}

		createdTime, err := time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, fmt.Errorf("parsing comment time: %w", err)
		}
		c.Created = createdTime
		comments = append(comments, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating comments: %w", err)
	}

	return comments, nil
}

// InsertComment adds a new comment to the database.
func (db *DB) InsertComment(c *ticket.Comment) error {
	_, err := db.conn.Exec(`
		INSERT INTO comments (id, ticket_id, content, created)
		VALUES (?, ?, ?, ?)
	`,
		c.ID,
		c.TicketID,
		c.Content,
		c.Created.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("inserting comment: %w", err)
	}
	return nil
}

// GetCommentsForTicket retrieves all comments for a ticket, ordered by creation time.
func (db *DB) GetCommentsForTicket(ticketID string) ([]*ticket.Comment, error) {
	rows, err := db.conn.Query(`
		SELECT id, ticket_id, content, created
		FROM comments WHERE ticket_id = ?
		ORDER BY created ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("querying comments: %w", err)
	}
	defer rows.Close()

	var comments []*ticket.Comment
	for rows.Next() {
		var c ticket.Comment
		var created string

		if err := rows.Scan(&c.ID, &c.TicketID, &c.Content, &created); err != nil {
			return nil, fmt.Errorf("scanning comment: %w", err)
		}

		createdTime, err := time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, fmt.Errorf("parsing comment time: %w", err)
		}
		c.Created = createdTime
		comments = append(comments, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating comments: %w", err)
	}

	return comments, nil
}

// RebuildFromAll clears all tickets, comments, and dependencies and inserts the given lists.
func (db *DB) RebuildFromAll(tickets []*ticket.Ticket, comments []*ticket.Comment, dependencies []*ticket.Dependency) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM tickets"); err != nil {
		return fmt.Errorf("clearing tickets: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM ticket_labels"); err != nil {
		return fmt.Errorf("clearing ticket labels: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM comments"); err != nil {
		return fmt.Errorf("clearing comments: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM dependencies"); err != nil {
		return fmt.Errorf("clearing dependencies: %w", err)
	}

	ticketStmt, err := tx.Prepare(`
		INSERT INTO tickets (id, title, description, status, priority, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing ticket insert: %w", err)
	}
	defer ticketStmt.Close()

	labelStmt, err := tx.Prepare(`INSERT INTO ticket_labels (ticket_id, label) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing label insert: %w", err)
	}
	defer labelStmt.Close()

	for _, t := range tickets {
		_, err := ticketStmt.Exec(
			t.ID,
			t.Title,
			t.Description,
			string(t.Status),
			t.Priority,
			t.Created.Format(time.RFC3339Nano),
			t.Updated.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("inserting ticket %s: %w", t.ID, err)
		}

		for _, label := range t.Labels {
			if _, err := labelStmt.Exec(t.ID, label); err != nil {
				return fmt.Errorf("inserting label for ticket %s: %w", t.ID, err)
			}
		}
	}

	commentStmt, err := tx.Prepare(`
		INSERT INTO comments (id, ticket_id, content, created)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing comment insert: %w", err)
	}
	defer commentStmt.Close()

	for _, c := range comments {
		_, err := commentStmt.Exec(
			c.ID,
			c.TicketID,
			c.Content,
			c.Created.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("inserting comment %s: %w", c.ID, err)
		}
	}

	depStmt, err := tx.Prepare(`
		INSERT INTO dependencies (id, from_ticket_id, to_ticket_id, type, created)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing dependency insert: %w", err)
	}
	defer depStmt.Close()

	for _, d := range dependencies {
		_, err := depStmt.Exec(
			d.ID,
			d.FromTicketID,
			d.ToTicketID,
			string(d.Type),
			d.Created.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("inserting dependency %s: %w", d.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// InsertDependency adds a new dependency to the database.
func (db *DB) InsertDependency(d *ticket.Dependency) error {
	_, err := db.conn.Exec(`
		INSERT INTO dependencies (id, from_ticket_id, to_ticket_id, type, created)
		VALUES (?, ?, ?, ?, ?)
	`,
		d.ID,
		d.FromTicketID,
		d.ToTicketID,
		string(d.Type),
		d.Created.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("inserting dependency: %w", err)
	}
	return nil
}

// GetDependenciesFrom retrieves all dependencies from a specific ticket.
func (db *DB) GetDependenciesFrom(ticketID string) ([]*ticket.Dependency, error) {
	rows, err := db.conn.Query(`
		SELECT id, from_ticket_id, to_ticket_id, type, created
		FROM dependencies WHERE from_ticket_id = ?
		ORDER BY created ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("querying dependencies: %w", err)
	}
	defer rows.Close()

	return scanDependencies(rows)
}

// GetDependenciesTo retrieves all dependencies pointing to a specific ticket.
func (db *DB) GetDependenciesTo(ticketID string) ([]*ticket.Dependency, error) {
	rows, err := db.conn.Query(`
		SELECT id, from_ticket_id, to_ticket_id, type, created
		FROM dependencies WHERE to_ticket_id = ?
		ORDER BY created ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("querying dependencies: %w", err)
	}
	defer rows.Close()

	return scanDependencies(rows)
}

// GetAllDependencies retrieves all dependencies from the database.
func (db *DB) GetAllDependencies() ([]*ticket.Dependency, error) {
	rows, err := db.conn.Query(`
		SELECT id, from_ticket_id, to_ticket_id, type, created
		FROM dependencies
		ORDER BY created ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying dependencies: %w", err)
	}
	defer rows.Close()

	return scanDependencies(rows)
}

// GetBlockingDependencies retrieves all blocked_by dependencies from the database.
func (db *DB) GetBlockingDependencies() ([]*ticket.Dependency, error) {
	rows, err := db.conn.Query(`
		SELECT id, from_ticket_id, to_ticket_id, type, created
		FROM dependencies WHERE type = ?
		ORDER BY created ASC
	`, string(ticket.DependencyBlockedBy))
	if err != nil {
		return nil, fmt.Errorf("querying blocking dependencies: %w", err)
	}
	defer rows.Close()

	return scanDependencies(rows)
}

// DependencyExists checks if a specific dependency already exists.
func (db *DB) DependencyExists(fromTicketID, toTicketID string, depType ticket.DependencyType) (bool, error) {
	var count int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM dependencies
		WHERE from_ticket_id = ? AND to_ticket_id = ? AND type = ?
	`, fromTicketID, toTicketID, string(depType)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking dependency existence: %w", err)
	}
	return count > 0, nil
}

func scanDependencies(rows *sql.Rows) ([]*ticket.Dependency, error) {
	var dependencies []*ticket.Dependency
	for rows.Next() {
		var d ticket.Dependency
		var depType string
		var created string

		if err := rows.Scan(&d.ID, &d.FromTicketID, &d.ToTicketID, &depType, &created); err != nil {
			return nil, fmt.Errorf("scanning dependency: %w", err)
		}

		d.Type = ticket.DependencyType(depType)
		createdTime, err := time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, fmt.Errorf("parsing dependency time: %w", err)
		}
		d.Created = createdTime
		dependencies = append(dependencies, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating dependencies: %w", err)
	}

	return dependencies, nil
}
