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

# Create a ticket with labels
thicket add --title "Fix login bug" --priority 1 --label bug

# List ready tickets (not blocked by other open tickets)
thicket ready

# Show a specific ticket
thicket show TH-abc123

# Add a comment to track progress
thicket comment TH-abc123 "Found the root cause"

# Close a ticket
thicket close TH-abc123
```

All commands support `--json` for machine-readable output.

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
