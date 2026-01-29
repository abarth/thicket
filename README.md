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

### 1. Initialize your project

```bash
thicket init --project TH
```

### 2. Launch the TUI

The Terminal UI is the recommended interface for human users. It provides an interactive way to browse, create, and manage tickets.

```bash
thicket tui
```

**Key Features:**
- **Navigation**: Use arrow keys or `j`/`k` to move through the list.
- **Creation**: Press `n` to create a new ticket.
- **Management**: Press `e` to edit, `c` to close, or `m` to add a comment (in detail view).
- **Filtering**: Use `o`, `x`, `i`, or `a` to filter by open, closed, icebox, or all tickets.

### CLI Usage

For automation or quick commands, Thicket provides a robust CLI.

```bash
# Show the next actionable ticket (most important for agents)
thicket ready

# Add a ticket
thicket add --title "Fix login bug" --type bug --priority 1

# List open tickets
thicket list --status open
```

All commands support `--json` for machine-readable output. For the full CLI reference, see [docs/CLI.md](docs/CLI.md).

## Project Structure

```
your-project/
└── .thicket/
    ├── config.json      # Project configuration
    ├── tickets.jsonl    # Ticket data (git-tracked)
    ├── cache.db         # SQLite cache (git-ignored)
    └── .gitignore       # Ignores cache.db
```

## For Coding Agents

Thicket is designed to help coding agents track their work. Run `thicket quickstart` for a workflow guide, or see [AGENTS.md](AGENTS.md) for detailed instructions.

## Documentation

- [CLI Reference](docs/CLI.md) - Detailed command documentation
- [AGENTS.md](AGENTS.md) - Instructions for coding agents working on this project
- [Development](docs/development.md) - Information for developers of Thicket
