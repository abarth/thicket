# Instructions for Coding Agents

## Workflow

```bash
# Build thicket (if not already built)
go build -o thicket ./cmd/thicket

# Get the next ticket to work on
./thicket ready
```

## Rules

**CRITICAL**: Do NOT read or write `.thicket/tickets.jsonl` directly. This file is the source of truth for the project's ticket database, and manual edits can corrupt the data or cause sync issues. Always use the `thicket` command (e.g., `./thicket ready`) to interact with the production database.

**IMPORTANT**: Do NOT run `./thicket` commands in the main project directory for manual testing. Running commands like `./thicket add` or `./thicket link` from the project root will modify the production ticket database (`.thicket/tickets.jsonl`), which is tracked by git and contains real project tickets.

### At the Start of Each Session

1. **Get the next ticket**: Run `./thicket ready` to see the highest priority actionable ticket with full details.

### When Completing Work

1. **Verify no regressions**: Run `go test ./...`
2. **Close the ticket**: `./thicket close --json <ID>`
3. **Final thought**: Before finishing, consider if any follow-up tasks are needed and create tickets for them:
   ```bash
   ./thicket add --json --title "Brief descriptive title" --description "Detailed context" --created-from <CURRENT_TICKET_ID>
   ```

## Run tests

Use `go test ./...` — the test suite creates isolated temporary directories and does not touch the production data.

### For Manual/Ad-hoc Testing
If you need to manually test Thicket commands, you **must** create a separate test instance in a different directory:

```bash
# Create a test directory
mkdir -p thicket-test
cd thicket-test

# Initialize a test instance
../thicket init --project TS --json

# Now you can safely test commands
../thicket add --json --title "Test ticket" --priority 1
../thicket list --json

# Clean up when done
cd ..
rm -rf thicket-test
```

### Why This Matters
- Running `./thicket` from the project root modifies the production database
- The production `.thicket/tickets.jsonl` is version-controlled and shared
- Test data would pollute the real ticket list
- Corrupted test data could break the production instance
- Other agents and developers rely on the ticket data being accurate
