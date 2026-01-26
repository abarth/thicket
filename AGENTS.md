# Instructions for Coding Agents

This project uses **Thicket** to track work. Before starting any task, check the ticket list and use tickets to manage your work.

## Quick Start

```bash
# Build thicket (if not already built)
go build -o thicket ./cmd/thicket

# See what needs to be done
./thicket list

# Get detailed guidance
./thicket quickstart
```

## Required Workflow

### At the Start of Each Session

1. **Check open tickets**: Run `./thicket list` to see current work items
2. **Pick a ticket**: Choose the highest priority (lowest number) open ticket
3. **Review the ticket**: Run `./thicket show <ID>` to understand the task

### While Working

1. **Create tickets for new work**: When you discover bugs, tasks, or improvements:
   ```bash
   ./thicket add --title "Brief description" --priority N
   ```

2. **Add context to tickets**: Use descriptions for complex issues:
   ```bash
   ./thicket add --title "Fix auth timeout" --description "The auth module times out after 30s but should wait 60s. See auth.go:142" --priority 1
   ```

3. **Update tickets as needed**:
   ```bash
   ./thicket update --description "New information" <ID>
   ```

### When Completing Work

1. **Close the ticket**: `./thicket close <ID>`
2. **Verify no regressions**: Run `go test ./...`
3. **Check for follow-up work**: `./thicket list`

## Priority Guidelines

| Priority | Meaning | Examples |
|----------|---------|----------|
| 0 | Critical | Blocking bugs, broken builds |
| 1 | High | Important features, significant bugs |
| 2 | Normal | Regular tasks, minor bugs |
| 3+ | Low | Nice-to-haves, future improvements |

## Project Architecture

```
thicket/
├── cmd/thicket/           # CLI entry point (main.go)
├── internal/
│   ├── commands/          # CLI command implementations
│   ├── config/            # Project configuration (.thicket/config.json)
│   ├── errors/            # User-friendly error types
│   ├── storage/           # Data persistence (JSONL + SQLite)
│   └── ticket/            # Core ticket data model
├── docs/
│   ├── overview.md        # Product vision and full specification
│   └── mvp.md             # MVP implementation plan
├── .thicket/              # Thicket data directory
│   ├── config.json        # Project code configuration
│   ├── tickets.jsonl      # Ticket data (git-tracked)
│   └── cache.db           # SQLite cache (git-ignored)
├── AGENTS.md              # This file
└── README.md              # User documentation
```

## Key Files to Understand

- **docs/overview.md**: Full product vision including features not yet implemented (comments, labels, dependencies, blocking relationships)
- **docs/mvp.md**: What was implemented in the MVP and the technical design decisions
- **internal/ticket/ticket.go**: Core data model - understand this first
- **internal/storage/**: How data flows between JSONL (source of truth) and SQLite (cache)

## Development Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Build the CLI
go build -o thicket ./cmd/thicket

# Run a specific test
go test ./internal/ticket -run TestValidateID -v
```

## What's Implemented (MVP)

- Ticket CRUD operations (create, read, update, close)
- Priority-based listing
- Status filtering (open/closed)
- JSONL storage with SQLite caching
- Automatic sync between JSONL and SQLite

## What's NOT Yet Implemented

See `docs/overview.md` for the full vision. Key missing features:

- **Comments**: Adding notes to tickets over time
- **Labels**: Tagging tickets for grouping
- **Dependencies**: Tracking which tickets block others
- **Created-from relationships**: Knowing which ticket spawned another
- **Assignee**: Who is working on a ticket
- **Issue type**: Bug, feature, task, etc.
- **--json output**: Machine-readable output for tooling
- **Unblocked query**: Finding tickets ready to work on

## Conventions

1. **Ticket IDs**: Format is `XX-xxxxxx` (e.g., `TH-a1b2c3`). The project code is `TH`.
2. **Commits**: Reference ticket IDs in commit messages when relevant
3. **Testing**: All new code should have tests. Target 80%+ coverage.
4. **Error handling**: Use `internal/errors` for user-facing errors with hints

## Common Issues

### "Thicket is not initialized"
Run `./thicket init --project TH` or ensure you're in the project root.

### SQLite cache out of sync
Delete `.thicket/cache.db` - it will rebuild automatically.

### Tests failing on macOS
Some tests use `filepath.EvalSymlinks` due to `/var` -> `/private/var` symlinks.
