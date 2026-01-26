# Thicket

A lightweight issue tracker for coding agents to track their work in projects.

## Overview

Thicket stores tickets in a `.thicket/tickets.jsonl` file that is tracked by git and shared by everyone working on the project. For efficient queries, data is cached in a local SQLite database that automatically syncs with the JSONL file.

## Installation

```bash
go install github.com/abarth/thicket/cmd/thicket@latest
```

Or build from source:

```bash
go build -o thicket ./cmd/thicket
```

## Quick Start

```bash
# Initialize Thicket in your project
thicket init --project TH

# Create a ticket
thicket add --title "Fix login bug" --description "Users cannot log in with special characters" --priority 1

# List open tickets
thicket list

# Show a specific ticket
thicket show TH-abc123

# Update a ticket
thicket update TH-abc123 --priority 2

# Close a ticket
thicket close TH-abc123
```

## Project Structure

```
.thicket/
├── config.json      # Project configuration (project code)
├── tickets.jsonl    # Authoritative ticket data (git-tracked)
├── cache.db         # SQLite cache (git-ignored)
└── .gitignore       # Ignores cache.db
```

## Ticket Format

Each ticket has:
- **ID**: Two-letter project code + six hex characters (e.g., `TH-a1b2c3`)
- **Title**: Short summary
- **Description**: Detailed explanation
- **Status**: `open` or `closed`
- **Priority**: Integer (lower = higher priority)
- **Created/Updated**: Timestamps

## Development

Run tests:

```bash
go test ./...
```

## Documentation

- [Overview](docs/overview.md) - Full product vision
- [MVP Plan](docs/mvp.md) - Minimum viable product specification
