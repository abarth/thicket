```bash
go build -o thicket ./cmd/thicket
go test ./...
```

- Write automated tests whenever you change behavior. Goal is 80% test coverage.
- Keep documentation updated, including README.md, docs/CLI.md, and help messages in internal/tui.

**CRITICAL**: Do NOT read or write `.thicket/tickets.jsonl` directly. This file is the production database, and manual edits can corrupt the data. Always use the `thicket` command to interact with the production database.

**IMPORTANT**: Do NOT run `./thicket` commands in the main project directory for manual testing. Running these commands from the project root will modify the production ticket database, which contains real project tickets.
