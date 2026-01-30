package commands

import "fmt"

// Quickstart prints guidance for coding agents on using Thicket.
func Quickstart(args []string) error {
	fmt.Print(`# Thicket Quickstart for Coding Agents

Your goal is to improve the project by resolving tickets and discovering additional work for future agents.

## Workflow

1. Work on the ticket described by "thicket ready".
2. When resolved, run "thicket close <CURRENT_TICKET_ID>".
3. Think of additional work and create tickets for future agents:
   thicket add --title "Brief descriptive title" --description "Detailed context" --priority=<N> --type=<TYPE> --created-from <CURRENT_TICKET_ID>
4. Commit your changes.

## Commands

- thicket ready
- thicket close <TICKET_ID>
- thicket add --title <TITLE> --description <DESC> --type <TYPE> --priority <N> [--label <LABEL>...] [--blocks <ID>] [--blocked-by <ID>] [--created-from <ID>]
- thicket show <TICKET_ID>
- thicket comment <TICKET_ID> "Comment text"
- thicket list [--status <STATUS>] [--label <LABEL>]

**CRITICAL**: NEVER edit .thicket/tickets.jsonl directly. Always use the thicket CLI.
`)
	return nil
}
