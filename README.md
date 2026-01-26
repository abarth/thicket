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

# Add a comment to a ticket
thicket comment TH-abc123 "Found the root cause"

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

Display details of a specific ticket, including any comments.

```bash
thicket show <TICKET-ID>
```

### `thicket comment`

Add a comment to a ticket. Comments are displayed when viewing the ticket with `show`.

```bash
thicket comment <TICKET-ID> "Comment text"
```

Comments are stored as separate lines in `tickets.jsonl` and are useful for:
- Recording progress on a ticket
- Noting discoveries or blockers
- Documenting decisions made while working

### `thicket link`

Create dependencies between tickets.

```bash
thicket link [flags] <TICKET-ID>
```

**Flags:**
- `--blocked-by`: Mark this ticket as blocked by another ticket
- `--created-from`: Track which ticket this was created from

**Examples:**
```bash
# TH-child is blocked by TH-blocker (TH-child cannot proceed until TH-blocker is closed)
thicket link --blocked-by TH-blocker TH-child

# Track that TH-child was created while working on TH-parent
thicket link --created-from TH-parent TH-child
```

**Notes:**
- Circular blocking dependencies are automatically detected and prevented
- The `show` command displays both "Blocked by" and "Blocking" relationships

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

### `thicket quickstart`

Display a guide for coding agents on how to use Thicket effectively.

```bash
thicket quickstart
```

## Ticket Format

Each ticket has:

| Field | Description |
|-------|-------------|
| **ID** | Two-letter project code + six alphanumeric characters (e.g., `TH-a1b2c3`) |
| **Title** | Short summary of the issue |
| **Description** | Detailed explanation |
| **Status** | `open` or `closed` |
| **Priority** | Integer (lower numbers = higher priority) |
| **Created** | Timestamp when ticket was created |
| **Updated** | Timestamp of last modification |
| **Comments** | Timestamped notes added over time |
| **Dependencies** | Links to blocking or related tickets |

### Comments

Comments have their own IDs (format: `TH-cXXXXXX`) and are linked to tickets by ticket ID. They are stored as separate lines in `tickets.jsonl` to keep diffs clean when adding comments.

### Dependencies

Dependencies track relationships between tickets:

- **blocked_by**: Indicates that a ticket cannot proceed until another ticket is completed
- **created_from**: Tracks that a ticket was created while working on another ticket

Dependencies have their own IDs (format: `TH-dXXXXXX`) and are stored as separate lines in `tickets.jsonl`. Circular blocking dependencies are automatically prevented.

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

### Claude Code Integration

If you are using [Claude Code](https://claude.ai/code), you can add a `CLAUDE.md` file to your project root to help the agent use Thicket effectively:

```markdown
# Claude Code Instructions

This project uses Thicket to track work.

## Work Management

- Run `thicket list` to see what needs to be done.
- Run `thicket quickstart` for a guide on the Thicket workflow.
```

Example workflow:

```bash
# Agent discovers a bug
thicket add --title "Fix null pointer in auth module" --priority 0

# Agent starts working, adds a comment with findings
thicket comment TH-abc123 "Root cause: missing nil check in validateUser()"

# Agent finds related issue while investigating
thicket add --title "Refactor auth error handling" --priority 2

# Agent fixes the bug, documents the fix
thicket comment TH-abc123 "Fixed by adding nil check at auth.go:142"
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

- [AGENTS.md](AGENTS.md) - Instructions for coding agents working on this project
- [Overview](docs/overview.md) - Full product vision

**For coding agents**: Start with `thicket quickstart` or read [AGENTS.md](AGENTS.md).
