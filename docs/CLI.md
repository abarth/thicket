# Thicket CLI Reference

This document provides detailed documentation for all Thicket commands.

## Global Flags

These flags can be used with almost all commands. They can be placed before or after the command.

- `--data-dir <DIR>`: Specify a custom `.thicket` directory location. This is useful for manual testing without affecting the production ticket data.
- `--json`: Output in JSON format for machine readability.

## Commands

### `thicket init`

Initialize a new Thicket project in the current directory.

```bash
thicket init --project <CODE>
```

**Flags:**
- `--project` (required): Two-letter project code (e.g., TH, BG, FX)

### `thicket add`

Create a new ticket.

```bash
thicket add [--interactive] [--title <TITLE>] [--description <DESC>] [--priority <N>] [--label <LABEL>]... [--blocks <ID>] [--blocked-by <ID>] [--created-from <ID>]
```

**Flags:**
- `--interactive`, `-i`: Enter interactive mode to provide ticket details. If title, description, or priority are not provided as flags, the tool will prompt for them.
- `--title`: Short summary of the ticket (required if not in interactive mode)
- `--description`: Detailed explanation
- `--priority`: Integer priority (default: 0, lower = higher priority)
- `--label`: Add a label (can be specified multiple times)
- `--blocks`: Mark an existing ticket as blocked by this new ticket
- `--blocked-by`: Mark this new ticket as blocked by an existing ticket
- `--created-from`: Track which existing ticket this new ticket was created from

**Examples:**
```bash
# Create a ticket with labels
thicket add --title "Fix login bug" --label bug --label urgent
```

### `thicket list`

List tickets ordered by priority.

```bash
thicket list [--status <STATUS>] [--label <LABEL>]
```

**Flags:**
- `--status`: Filter by status (`open` or `closed`)
- `--label`: Filter by label

**Alias:** `thicket ls`

**Examples:**
```bash
# List all open tickets with the "bug" label
thicket list --status open --label bug
```

### `thicket ready`

List open tickets that are not blocked by other open tickets. These are actionable work items.

```bash
thicket ready
```

### `thicket show`

Display details of a specific ticket, including any comments.

```bash
thicket show <TICKET-ID>
```

### `thicket comment`

Add a comment to a ticket. Comments are displayed when viewing the ticket with `show`.

```bash
thicket comment <TICKET-ID> "Comment text"
```

Comments are stored as separate lines in `tickets.jsonl` and are useful for:
- Recording progress on a ticket
- Noting discoveries or blockers
- Documenting decisions made while working

### `thicket link`

Create dependencies between tickets.

```bash
thicket link [flags] <TICKET-ID>
```

**Flags:**
- `--blocked-by`: Mark this ticket as blocked by another ticket
- `--created-from`: Track which ticket this was created from

**Examples:**
```bash
# TH-child is blocked by TH-blocker (TH-child cannot proceed until TH-blocker is closed)
thicket link --blocked-by TH-blocker TH-child

# Track that TH-child was created while working on TH-parent
thicket link --created-from TH-parent TH-child
```

**Notes:**
- Circular blocking dependencies are automatically detected and prevented
- The `show` command displays both "Blocked by" and "Blocking" relationships

### `thicket update`

Modify an existing ticket.

```bash
thicket update [flags] <TICKET-ID>
```

**Flags:**
- `--title`: New title
- `--description`: New description
- `--priority`: New priority
- `--status`: New status (`open` or `closed`)
- `--add-label`: Add a label (can be specified multiple times)
- `--remove-label`: Remove a label (can be specified multiple times)

**Examples:**
```bash
# Add a label to an existing ticket
thicket update --add-label urgent TH-abc123

# Remove a label
thicket update --remove-label urgent TH-abc123
```

### `thicket close`

Close a ticket (shortcut for `update --status closed`).

```bash
thicket close <TICKET-ID>
```

### `thicket quickstart`

Display a guide for coding agents on how to use Thicket effectively.

```bash
thicket quickstart
```
