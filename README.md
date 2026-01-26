# Thicket

A lightweight issue tracker designed for coding agents to track their work in projects.

## Overview

Thicket stores tickets in a `.thicket/tickets.jsonl` file that is tracked by git and shared by everyone working on the project. For efficient queries, data is cached in a local SQLite database that automatically syncs with the JSONL file.

## Installation

```bash
go install github.com/abarth/thicket/cmd/thicket@latest
```

Or build from source:

```bash
git clone https://github.com/abarth/thicket.git
cd thicket
go build -o thicket ./cmd/thicket
```

## Quick Start

```bash
# Initialize Thicket in your project
thicket init --project TH

# Create a ticket
thicket add --title "Fix login bug" --description "Users cannot log in" --priority 1

# List open tickets
thicket list

# Show a specific ticket
thicket show TH-abc123

# Update a ticket
thicket update --priority 2 TH-abc123

# Close a ticket
thicket close TH-abc123
```

## Commands

### `thicket init`

Initialize a new Thicket project in the current directory.

```bash
thicket init --project <CODE>
```

**Flags:**
- `--project` (required): Two-letter project code (e.g., TH, BG, FX)

### `thicket add`

Create a new ticket.

```bash
thicket add --title <TITLE> [--description <DESC>] [--priority <N>]
```

**Flags:**
- `--title` (required): Short summary of the ticket
- `--description`: Detailed explanation
- `--priority`: Integer priority (default: 0, lower = higher priority)

### `thicket list`

List tickets ordered by priority.

```bash
thicket list [--status <STATUS>]
```

**Flags:**
- `--status`: Filter by status (`open` or `closed`)

**Alias:** `thicket ls`

### `thicket show`

Display details of a specific ticket.

```bash
thicket show <TICKET-ID>
```

### `thicket update`

Modify an existing ticket.

```bash
thicket update [flags] <TICKET-ID>
```

**Flags:**
- `--title`: New title
- `--description`: New description
- `--priority`: New priority
- `--status`: New status (`open` or `closed`)

### `thicket close`

Close a ticket (shortcut for `update --status closed`).

```bash
thicket close <TICKET-ID>
```

## Ticket Format

Each ticket has:

| Field | Description |
|-------|-------------|
| **ID** | Two-letter project code + six hex characters (e.g., `TH-a1b2c3`) |
| **Title** | Short summary of the issue |
| **Description** | Detailed explanation |
| **Status** | `open` or `closed` |
| **Priority** | Integer (lower numbers = higher priority) |
| **Created** | Timestamp when ticket was created |
| **Updated** | Timestamp of last modification |

## Project Structure

```
your-project/
└── .thicket/
    ├── config.json      # Project configuration
    ├── tickets.jsonl    # Ticket data (git-tracked)
    ├── cache.db         # SQLite cache (git-ignored)
    └── .gitignore       # Ignores cache.db
```

### Files

- **config.json**: Stores the project code and settings
- **tickets.jsonl**: Authoritative source of ticket data in JSON Lines format
- **cache.db**: Local SQLite database for fast queries (automatically rebuilt from JSONL)

## For Coding Agents

Thicket is designed to help coding agents track their work:

1. **Create tickets** when you identify tasks or issues
2. **Set priorities** to determine work order (lower = more urgent)
3. **Update tickets** as you learn more about the problem
4. **Close tickets** when work is complete

Example workflow:

```bash
# Agent discovers a bug
thicket add --title "Fix null pointer in auth module" --priority 0

# Agent starts working, finds related issue
thicket add --title "Refactor auth error handling" --priority 2

# Agent fixes the bug
thicket close TH-abc123

# Check remaining work
thicket list --status open
```

## Development

### Running Tests

```bash
go test ./...
```

### Test Coverage

```bash
go test ./... -cover
```

### Project Layout

```
thicket/
├── cmd/thicket/           # CLI entry point
├── internal/
│   ├── commands/          # CLI command implementations
│   ├── config/            # Configuration management
│   ├── errors/            # User-friendly error types
│   ├── storage/           # JSONL and SQLite operations
│   └── ticket/            # Ticket data model
└── docs/                  # Documentation
```

## Documentation

- [Overview](docs/overview.md) - Full product vision
- [MVP Plan](docs/mvp.md) - Minimum viable product specification
