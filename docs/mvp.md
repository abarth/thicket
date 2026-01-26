# Thicket MVP Plan

This document describes the minimum-viable implementation of Thicket needed to bootstrap development. The goal is to use Thicket to track work on Thicket itself.

## MVP Scope

### In Scope

1. **Ticket data model** with essential fields only:
   - ID: two-letter project code + six-character hex string (e.g., `TH-a1b2c3`)
   - Title: short summary of the ticket
   - Description: detailed explanation
   - Status: `open` or `closed`
   - Priority: integer (lower numbers = higher priority)
   - Created timestamp
   - Updated timestamp

2. **Storage layer**:
   - `.thicket/tickets.jsonl` as the source of truth
   - SQLite database as local cache for efficient queries
   - Automatic sync from JSONL to SQLite on each command

3. **CLI commands**:
   - `thicket init` - Initialize a new Thicket project in the current directory
   - `thicket add` - Create a new ticket
   - `thicket list` - List tickets (with filtering by status)
   - `thicket show <id>` - Display a single ticket
   - `thicket update <id>` - Modify ticket fields
   - `thicket close <id>` - Close a ticket

4. **Basic configuration**:
   - `.thicket/config.json` for project code and settings

### Deferred to Post-MVP

- Comments on tickets
- Labels
- Dependencies (blocked-by, created-from relationships)
- Assignee field
- Issue type field
- Closed timestamp
- `--json` output flag
- Unblocked ticket queries

## Technical Design

### Directory Structure

```
.thicket/
├── config.json      # Project configuration
├── tickets.jsonl    # Authoritative ticket data (git-tracked)
└── cache.db         # SQLite cache (git-ignored)
```

### JSONL Format

Each line in `tickets.jsonl` is a JSON object representing a ticket:

```json
{"id":"TH-a1b2c3","title":"Implement add command","description":"...","status":"open","priority":1,"created":"2024-01-15T10:00:00Z","updated":"2024-01-15T10:00:00Z"}
```

### SQLite Schema

```sql
CREATE TABLE tickets (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    priority INTEGER NOT NULL DEFAULT 0,
    created TEXT NOT NULL,
    updated TEXT NOT NULL
);

CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_priority ON tickets(priority);
```

### Sync Strategy

On every command:
1. Read the modification time of `tickets.jsonl`
2. Compare with a stored timestamp in SQLite
3. If JSONL is newer, rebuild the SQLite tables from JSONL
4. For write operations, append to JSONL and update SQLite

### Project Structure

```
thicket/
├── cmd/
│   └── thicket/
│       └── main.go         # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go       # Configuration loading
│   ├── storage/
│   │   ├── jsonl.go        # JSONL file operations
│   │   ├── sqlite.go       # SQLite cache operations
│   │   └── sync.go         # Sync logic
│   └── ticket/
│       └── ticket.go       # Ticket data model
├── docs/
│   ├── overview.md
│   └── mvp.md
├── go.mod
├── go.sum
└── README.md
```

## Implementation Tasks

### Phase 1: Foundation

1. Initialize Go module and project structure
2. Define the Ticket struct and validation
3. Implement JSONL read/write operations
4. Implement SQLite schema and basic operations
5. Implement sync logic between JSONL and SQLite

### Phase 2: CLI Commands

1. Implement `thicket init` command
2. Implement `thicket add` command
3. Implement `thicket list` command
4. Implement `thicket show` command
5. Implement `thicket update` command
6. Implement `thicket close` command

### Phase 3: Polish

1. Add input validation and error messages
2. Write unit tests for core functionality
3. Write integration tests for CLI commands
4. Add README with usage instructions

## Testing Strategy

- **Unit tests**: Cover ticket model, JSONL parsing, SQLite operations, and sync logic
- **Integration tests**: Test CLI commands end-to-end using a temporary directory
- **Test coverage target**: 80% for core packages

## Success Criteria

The MVP is complete when:

1. A developer can initialize Thicket in a project directory
2. Tickets can be created, listed, viewed, updated, and closed
3. Data persists correctly in `.thicket/tickets.jsonl`
4. The SQLite cache rebuilds correctly from JSONL
5. Core functionality has unit test coverage
6. README documents basic usage

## Usage Examples

```bash
# Initialize Thicket in the current project
thicket init --project TH

# Create a new ticket
thicket add --title "Implement list command" --description "Add the list command to show all tickets" --priority 1

# List open tickets
thicket list
thicket list --status open

# Show a specific ticket
thicket show TH-a1b2c3

# Update a ticket
thicket update TH-a1b2c3 --priority 2

# Close a ticket
thicket close TH-a1b2c3
```
