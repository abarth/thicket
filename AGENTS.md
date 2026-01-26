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

1. **Check open tickets**: Run `./thicket ready` to see actionable work items.
2. **Pick a ticket**: Choose the highest priority (lowest number) open ticket.
3. **Review the ticket**: Run `./thicket show --json <ID>` to understand the task

### While Working

1. **Create tickets for new work**: When you discover bugs, tasks, or improvements:
   ```bash
   # Basic use
   ./thicket add --json --title "Brief description" --priority N

   # Create and link in one command
   ./thicket add --json --title "Follow-on task" --created-from <CURRENT-ID>
   ./thicket add --json --title "Urgent blocker" --blocks <EXISTING-ID> --priority 0
   ```

2. **Add context to tickets**: Use descriptions for complex issues:
   ```bash
   ./thicket add --json --title "Fix auth timeout" --description "The auth module times out after 30s but should wait 60s. See auth.go:142" --priority 1
   ```

3. **Add comments to track progress**: Document your findings and progress:
   ```bash
   ./thicket comment --json <ID> "Found root cause: missing nil check"
   ./thicket comment --json <ID> "Fix implemented, running tests"
   ```

4. **Link tickets with dependencies**: Track blocking relationships:
   ```bash
   ./thicket link --json --blocked-by <BLOCKER-ID> <BLOCKED-ID>
   ./thicket link --json --created-from <PARENT-ID> <CHILD-ID>
   ```

5. **Update tickets as needed**:
   ```bash
   ./thicket update --json --description "New information" <ID>
   ```

### When Completing Work

1. **Check for follow-on work**: If you discovered additional bugs, tasks, or improvements that you are not addressing in your current task, create new tickets for them. You can link them to the current ticket using `--created-from` directly when adding them.
   ```bash
   # Create follow-on ticket and link it to the current ticket in one step
   ./thicket add --json --title "Follow-on task" --priority 2 --created-from <CURRENT-ID>
   ```
2. **Close the ticket**: `./thicket close --json <ID>`
3. **Verify no regressions**: Run `go test ./...`

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
├── .thicket/              # Thicket data directory
│   ├── config.json        # Project code configuration
│   ├── tickets.jsonl      # Ticket data (git-tracked)
│   └── cache.db           # SQLite cache (git-ignored)
├── AGENTS.md              # This file
└── README.md              # User documentation
```

## Key Files to Understand

- **internal/ticket/ticket.go**: Core ticket data model - understand this first
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
/path/to/thicket link --json --blocked-by TS-abc123 TS-def456

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
