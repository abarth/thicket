---
description: Standard agentic coding workflow for Thicket
---

Follow these steps to work on Thicket using its own issue tracking system.

// turbo
1. Check for the next high-priority ticket:
   ```bash
   ./thicket ready
   ```

2. Implement the changes requested in the ticket. While working, if you discover new problems or potential improvements, log them as new tickets:
   ```bash
   ./thicket add --json --title "Brief descriptive title" --description "Detailed context" --created-from <CURRENT_TICKET_ID>
   ```

3. Document your progress or findings using comments:
   ```bash
   ./thicket comment --json <CURRENT_TICKET_ID> "Describe your progress or technical discovery"
   ```

4. Verify your changes do not introduce regressions:
   ```bash
   go test ./...
   ```

5. Once the task is complete and verified, close the ticket:
   ```bash
   ./thicket close --json <CURRENT_TICKET_ID>
   ```

6. Final thought: Before finishing, consider if any follow-up tasks are needed and create tickets for them using step 2.

**CRITICAL**: NEVER edit `.thicket/tickets.jsonl` directly. Always use the `./thicket` CLI.