package commands

import "fmt"

// Quickstart prints guidance for coding agents on using Thicket.
func Quickstart(args []string) error {
	fmt.Print(`Thicket Quickstart for Coding Agents
=====================================

Thicket is a lightweight issue tracker designed to help you manage your work.
Use it to track tasks, bugs, and features as you work on this project.
Coding agents should use the --json flag for machine-readable output.

GETTING STARTED
---------------

1. Check current tickets:
   thicket ready                   # Show actionable (not blocked) open tickets
   thicket list --status open      # Show all open tickets

2. Pick a ticket to work on (lowest priority number = most important):
   thicket show --json TH-abc123   # View ticket details and comments

3. Create tickets as you discover work:
   thicket add --json --title "Fix bug in auth" --priority 1
   thicket add --json --title "Refactor later" --description "Details here" --priority 5

4. Add comments to track progress:
   thicket comment --json TH-abc123 "Found root cause in auth.go:142"
   thicket comment --json TH-abc123 "Fix implemented, tests passing"

5. Link tickets with dependencies:
   thicket link --json --blocked-by TH-blocker TH-blocked   # TH-blocked is blocked by TH-blocker
   thicket link --json --created-from TH-parent TH-child    # TH-child was created from TH-parent

6. Close tickets when done:
   thicket close --json TH-abc123

WORKFLOW
--------

When you start a session:
  1. Run 'thicket list --status open' to see what needs to be done
  2. Pick the highest priority (lowest number) open ticket
  3. Work on it, adding comments as you make progress
  4. Create new tickets for issues you discover along the way
  5. Close the ticket when complete

Priority guidelines:
  0 = Critical, blocking other work
  1 = High priority, do soon
  2 = Normal priority
  3+ = Lower priority, can wait

BEST PRACTICES
--------------

- **NEVER read or edit .thicket/tickets.jsonl directly.** Always use the CLI.
- Use the --json flag for all commands to get machine-readable output
- Create tickets for any work you defer ("I'll fix this later")
- Use descriptive titles that explain WHAT needs to be done
- Add descriptions for complex issues to capture context
- Add comments to document your progress and findings
- Close tickets promptly when work is complete
- Check 'thicket list --status open' at the start of each session

COMMANDS REFERENCE
------------------

  thicket list [--json] [--status open|closed]  List tickets by priority
  thicket show [--json] <ID>                      View ticket details and dependencies
  thicket add --json --title "..." [options]      Create a new ticket
  thicket comment --json <ID> "text"              Add a comment to a ticket
  thicket link --json [flags] <ID>                Create ticket dependencies
  thicket update --json [options] <ID>            Modify a ticket
  thicket close --json <ID>                       Mark ticket as closed
  thicket quickstart                              Show this guide

For more details, see AGENTS.md in the project root.
`)
	return nil
}
