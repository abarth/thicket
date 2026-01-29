# Instructions for Coding Agents

This project uses **Thicket** to track work. Before starting any task, check the ticket list and use tickets to manage your work.

## Quick Start

```bash
# Build thicket (if not already built)
go build -o thicket ./cmd/thicket

# See what needs to be done
./thicket ready

# Get detailed guidance
./thicket quickstart
```

## Required Workflow

**CRITICAL**: Do NOT read or write `.thicket/tickets.jsonl` directly. This file is the source of truth for the project's ticket database, and manual edits can corrupt the data or cause sync issues. Always use the `thicket` command (e.g., `./thicket ready`, `./thicket show`) to interact with the production database.

### At the Start of Each Session

1. **Get the next ticket**: Run `./thicket ready` to see the highest priority actionable ticket with full details.

### While Working

1. **Discover additional work** In addition to resolving the current ticket, your
   job is to discover additional work, such as bugs, tasks, refactorings, or
   improvements. When you discover new work, file a ticket for a future agent to
   work on:
   ```bash
   ./thicket add --json --title "Follow-on task" --created-from <CURRENT-ID>
   ./thicket add --json --title "Blocking issue" --blocks <EXISTING-ID>
   ```

2. **Add comments to track progress**: Document your findings and progress:
   ```bash
   ./thicket comment --json <ID> "Found root cause: missing nil check"
   ./thicket comment --json <ID> "Fix implemented, running tests"
   ```

### When Completing Work

1. **Verify no regressions**: Run `go test ./...`
2. **Close the ticket**: `./thicket close --json <ID>`
3. **Think about follow-on work**: When you are done with the current ticket, think about
   additional work that needs to be done. Create new tickets additional follow-on work:
   ```bash
   ./thicket add --json --title "Follow-on task" --created-from <CURRENT-ID>
   ```

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
├── .thicket/              # Thicket data directory
│   ├── config.json        # Project code configuration
│   ├── tickets.jsonl      # Ticket data (git-tracked)
│   └── cache.db           # SQLite cache (git-ignored)
├── AGENTS.md              # This file
└── README.md              # User documentation
```

## Key Files to Understand

- **internal/ticket/ticket.go**: Core ticket data model
- **internal/storage/**: How data flows between JSONL (source of truth) and SQLite (cache)
- **.thicket/tickets.jsonl**: The production ticket database. **NEVER read or edit this file directly.** Always use the `thicket` CLI tool.

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

## Manual Testing

**IMPORTANT**: Do NOT run `./thicket` commands in the main project directory for manual testing. Running commands like `./thicket add` or `./thicket link` from the project root will modify the production ticket database (`.thicket/tickets.jsonl`), which is tracked by git and contains real project tickets.

### For Automated Tests
Use `go test ./...` — the test suite creates isolated temporary directories and does not touch the production data.

### For Manual/Ad-hoc Testing
If you need to manually test Thicket commands, you **must** create a separate test instance in a different directory:

```bash
# Create a test directory
mkdir -p /tmp/thicket-test
cd /tmp/thicket-test

# Initialize a test instance
/path/to/thicket init --project TS --json

# Now you can safely test commands
/path/to/thicket add --json --title "Test ticket" --priority 1
/path/to/thicket list --json

# Clean up when done
rm -rf /tmp/thicket-test
```

### Why This Matters
- Running `./thicket` from the project root modifies the production database
- The production `.thicket/tickets.jsonl` is version-controlled and shared
- Test data would pollute the real ticket list
- Corrupted test data could break the production instance
- Other agents and developers rely on the ticket data being accurate

## Conventions

1. **Ticket IDs**: Format is `XX-xxxxxx` (e.g., `TH-a1b2c3`). The project code is `TH`.
2. **Commits**: Reference ticket IDs in commit messages when relevant
3. **Testing**: All new code should have tests. Target 80%+ coverage.
4. **Error handling**: Use `internal/errors` for user-facing errors with hints
5. **Documentation**: When adding new features:
   - Update `README.md` with command usage and examples
   - Update the quickstart guide in `internal/commands/commands.go`
6. **Help text**: When adding new commands:
   - Add the command to `printUsage()` in `cmd/thicket/main.go`
   - Include usage examples in the command's `--help` output

## Common Issues

### "Thicket is not initialized"
Run `./thicket init --json --project TH` or ensure you're in the project root.

### SQLite cache out of sync
Delete `.thicket/cache.db` - it will rebuild automatically.

### Tests failing on macOS
Some tests use `filepath.EvalSymlinks` due to `/var` -> `/private/var` symlinks.
