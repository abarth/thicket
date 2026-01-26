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

	stmt, err := tx.Prepare(`
		INSERT INTO tickets (id, title, description, status, priority, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing insert: %w", err)
	}
	defer stmt.Close()

	for _, t := range tickets {
		_, err := stmt.Exec(
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
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// InsertTicket adds a new ticket to the database.
func (db *DB) InsertTicket(t *ticket.Ticket) error {
	_, err := db.conn.Exec(`
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
	return nil
}

// UpdateTicket updates an existing ticket in the database.
func (db *DB) UpdateTicket(t *ticket.Ticket) error {
	result, err := db.conn.Exec(`
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

	return &t, nil
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

	var tickets []*ticket.Ticket
	for rows.Next() {
		var t ticket.Ticket
		var statusStr string
		var created, updated string

		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &statusStr, &t.Priority, &created, &updated); err != nil {
			return nil, fmt.Errorf("scanning ticket: %w", err)
		}

		t.Status = ticket.Status(statusStr)
		t.Created, _ = time.Parse(time.RFC3339Nano, created)
		t.Updated, _ = time.Parse(time.RFC3339Nano, updated)

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
