# Thicket Development

## Running tests

```bash
# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover
```

## Project layout

thicket/
├── cmd/thicket/           # CLI entry point
├── internal/
│   ├── commands/          # CLI command implementations
│   ├── config/            # Configuration management
│   ├── errors/            # User-friendly error types
│   ├── storage/           # JSONL and SQLite operations
│   └── ticket/            # Ticket data model
├── docs/                  # Documentation
```
