package commands

import "fmt"

// Quickstart prints guidance for coding agents on using Thicket.
func Quickstart(args []string) error {
	fmt.Print(`# Thicket Quickstart for Coding Agents

Thicket is a lightweight issue tracker designed to help you manage your work.
Use it to track tasks, bugs, and features as you work on this project.
Coding agents should use the --json flag for machine-readable output.

## WORKFLOW

1. Get the next ticket to work on (shows full details):
   thicket ready

2. As you work on tickets, find additional work and create new tickets:
   thicket add --json --title "Refactor later" --description "Details here" --created-from TH-abc123

3. Add comments to track progress:
   thicket comment --json TH-abc123 "Found root cause in auth.go:142"
   thicket comment --json TH-abc123 "Fix implemented, tests passing"

4. Close tickets when completed:
   thicket close --json TH-abc123

5. Think about what additional work should be done and then file tickets for
   that work.

## Priority guidelines:

  0 = Critical, blocking other work
  1 = High priority, do soon
  2 = Normal priority
  3+ = Lower priority, can wait

## BEST PRACTICES

- **NEVER read or edit .thicket/tickets.jsonl directly.** Always use the CLI.
- Practively discover additional work as you go and create tickets for it
- Use descriptive titles that explain WHAT needs to be done
- Add descriptions for complex issues to capture context
- Add comments to document your progress and findings
- Close tickets promptly when work is complete
- Check 'thicket ready' at the start of each session

For more details, see AGENTS.md in the project root.
`)
	return nil
}
