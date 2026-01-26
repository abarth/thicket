// Package commands implements the CLI commands for Thicket.
package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// TicketDetails holds all information about a ticket for display.
type TicketDetails struct {
	Ticket      *ticket.Ticket    `json:"ticket"`
	Comments    []*ticket.Comment `json:"comments"`
	BlockedBy   []*ticket.Ticket  `json:"blocked_by"`
	Blocking    []*ticket.Ticket  `json:"blocking"`
	CreatedFrom *ticket.Ticket    `json:"created_from"`
}

// SuccessResponse is a common JSON response for mutating commands.
type SuccessResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ErrTicketNotFound is returned when a ticket cannot be found.
// Deprecated: Use thickerr.TicketNotFound() for better error messages.
var ErrTicketNotFound = thickerr.New("ticket not found")

// normalizeTicketID normalizes a ticket ID to the canonical format:
// uppercase project code, lowercase alphanumeric portion (e.g., "th-Z1Y2X3" -> "TH-z1y2x3").
func normalizeTicketID(id string) string {
	if len(id) < 3 || id[2] != '-' {
		return id
	}
	return strings.ToUpper(id[:2]) + "-" + strings.ToLower(id[3:])
}

// wrapConfigError converts config errors to user-friendly errors.
func wrapConfigError(err error) error {
	if err == config.ErrNotInitialized {
		return thickerr.NotInitialized()
	}
	if err == config.ErrAlreadyInit {
		return thickerr.New("Thicket is already initialized in this directory")
	}
	return err
}

// Init initializes a new Thicket project.
func Init(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	projectCode := fs.String("project", "", "Two-letter project code (e.g., TH)")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket init --project <CODE> [--json]")
		fmt.Fprintln(os.Stderr, "\nInitialize a new Thicket project in the current directory.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *projectCode == "" {
		return thickerr.MissingRequired("project")
	}

	*projectCode = strings.ToUpper(*projectCode)

	if err := ticket.ValidateProjectCode(*projectCode); err != nil {
		return thickerr.InvalidProjectCode(*projectCode)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	if err := config.Init(wd, *projectCode); err != nil {
		return wrapConfigError(err)
	}

	if *jsonOutput {
		return printJSON(SuccessResponse{
			Success: true,
			Message: fmt.Sprintf("Initialized Thicket project with code %s", *projectCode),
		})
	}

	fmt.Printf("Initialized Thicket project with code %s\n", *projectCode)
	return nil
}

// Add creates a new ticket.
func Add(args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	title := fs.String("title", "", "Ticket title (required)")
	description := fs.String("description", "", "Ticket description")
	priority := fs.Int("priority", 0, "Ticket priority (lower = higher priority)")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket add --title <TITLE> [--description <DESC>] [--priority <N>] [--json]")
		fmt.Fprintln(os.Stderr, "\nCreate a new ticket.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *title == "" {
		return thickerr.MissingRequired("title")
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	t, err := ticket.New(cfg.ProjectCode, *title, *description, *priority)
	if err != nil {
		return err
	}

	if err := store.Add(t); err != nil {
		return err
	}

	if *jsonOutput {
		return printJSON(SuccessResponse{
			Success: true,
			ID:      t.ID,
			Message: fmt.Sprintf("Created ticket %s", t.ID),
		})
	}

	fmt.Printf("Created ticket %s\n", t.ID)
	return nil
}

// List displays tickets.
func List(args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	statusFilter := fs.String("status", "", "Filter by status (open, closed)")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket list [--status <STATUS>] [--json]")
		fmt.Fprintln(os.Stderr, "\nList tickets, ordered by priority.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	var status *ticket.Status
	if *statusFilter != "" {
		s := ticket.Status(*statusFilter)
		if err := ticket.ValidateStatus(s); err != nil {
			return thickerr.InvalidStatus(*statusFilter)
		}
		status = &s
	}

	tickets, err := store.List(status)
	if err != nil {
		return err
	}

	if *jsonOutput {
		if tickets == nil {
			tickets = []*ticket.Ticket{}
		}
		return printJSON(tickets)
	}

	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
		return nil
	}

	printTicketTable(os.Stdout, tickets)
	return nil
}

// Ready displays open tickets that are not blocked by other open tickets.
func Ready(args []string) error {
	fs := flag.NewFlagSet("ready", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket ready [--json]")
		fmt.Fprintln(os.Stderr, "\nList open tickets that are not blocked by other open tickets, ordered by priority.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	tickets, err := store.ListReady()
	if err != nil {
		return err
	}

	if *jsonOutput {
		if tickets == nil {
			tickets = []*ticket.Ticket{}
		}
		return printJSON(tickets)
	}

	if len(tickets) == 0 {
		fmt.Println("No ready tickets found.")
		return nil
	}

	printTicketTable(os.Stdout, tickets)
	return nil
}

func printTicketTable(w io.Writer, tickets []*ticket.Ticket) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tPRI\tSTATUS\tTITLE")
	fmt.Fprintln(tw, "--\t---\t------\t-----")
	for _, t := range tickets {
		title := t.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", t.ID, t.Priority, t.Status, title)
	}
	tw.Flush()
}

// Show displays a single ticket.
func Show(args []string) error {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket show [--json] <TICKET-ID>")
		fmt.Fprintln(os.Stderr, "\nDisplay details of a specific ticket.")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return thickerr.WithHint("Ticket ID is required", "Usage: thicket show <TICKET-ID>")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return thickerr.InvalidTicketID(ticketID)
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	t, err := store.Get(ticketID)
	if err != nil {
		return err
	}
	if t == nil {
		return thickerr.TicketNotFound(ticketID)
	}

	comments, err := store.GetComments(ticketID)
	if err != nil {
		return err
	}

	blockedBy, err := store.GetBlockers(ticketID)
	if err != nil {
		return err
	}

	blocking, err := store.GetBlocking(ticketID)
	if err != nil {
		return err
	}

	createdFrom, err := store.GetCreatedFrom(ticketID)
	if err != nil {
		return err
	}

	details := &TicketDetails{
		Ticket:      t,
		Comments:    comments,
		BlockedBy:   blockedBy,
		Blocking:    blocking,
		CreatedFrom: createdFrom,
	}

	if *jsonOutput {
		return printJSON(details)
	}

	printTicketDetail(os.Stdout, details)
	return nil
}

func printTicketDetail(w io.Writer, details *TicketDetails) {
	t := details.Ticket
	fmt.Fprintf(w, "ID:          %s\n", t.ID)
	fmt.Fprintf(w, "Title:       %s\n", t.Title)
	fmt.Fprintf(w, "Status:      %s\n", t.Status)
	fmt.Fprintf(w, "Priority:    %d\n", t.Priority)
	fmt.Fprintf(w, "Created:     %s\n", t.Created.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated:     %s\n", t.Updated.Format(time.RFC3339))

	if details.CreatedFrom != nil {
		fmt.Fprintf(w, "Created from: %s (%s)\n", details.CreatedFrom.ID, details.CreatedFrom.Title)
	}

	if len(details.BlockedBy) > 0 {
		fmt.Fprintf(w, "\nBlocked by:\n")
		for _, b := range details.BlockedBy {
			status := ""
			if b.Status == ticket.StatusClosed {
				status = " [closed]"
			}
			fmt.Fprintf(w, "  - %s: %s%s\n", b.ID, b.Title, status)
		}
	}

	if len(details.Blocking) > 0 {
		fmt.Fprintf(w, "\nBlocking:\n")
		for _, b := range details.Blocking {
			fmt.Fprintf(w, "  - %s: %s\n", b.ID, b.Title)
		}
	}

	if t.Description != "" {
		fmt.Fprintf(w, "\nDescription:\n%s\n", t.Description)
	}
	if len(details.Comments) > 0 {
		fmt.Fprintf(w, "\nComments:\n")
		for _, c := range details.Comments {
			fmt.Fprintf(w, "  [%s] %s\n", c.Created.Format("2006-01-02 15:04:05"), c.Content)
		}
	}
}

// Update modifies an existing ticket.
func Update(args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	title := fs.String("title", "", "New title")
	description := fs.String("description", "", "New description")
	priority := fs.Int("priority", -1, "New priority")
	status := fs.String("status", "", "New status (open, closed)")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket update [flags] <TICKET-ID>")
		fmt.Fprintln(os.Stderr, "\nUpdate an existing ticket. Only specified fields are changed.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return thickerr.WithHint("Ticket ID is required", "Usage: thicket update [flags] <TICKET-ID>")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return thickerr.InvalidTicketID(ticketID)
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	t, err := store.Get(ticketID)
	if err != nil {
		return err
	}
	if t == nil {
		return thickerr.TicketNotFound(ticketID)
	}

	// Build update parameters
	var titlePtr, descPtr *string
	var priorityPtr *int
	var statusPtr *ticket.Status

	if *title != "" {
		titlePtr = title
	}
	if *description != "" {
		descPtr = description
	}
	if *priority >= 0 {
		priorityPtr = priority
	}
	if *status != "" {
		s := ticket.Status(*status)
		if err := ticket.ValidateStatus(s); err != nil {
			return thickerr.InvalidStatus(*status)
		}
		statusPtr = &s
	}

	if titlePtr == nil && descPtr == nil && priorityPtr == nil && statusPtr == nil {
		return thickerr.WithHint(
			"No fields to update",
			"Use --title, --description, --priority, or --status to specify changes",
		)
	}

	if err := t.Update(titlePtr, descPtr, priorityPtr, statusPtr); err != nil {
		return err
	}

	if err := store.Update(t); err != nil {
		return err
	}

	if *jsonOutput {
		return printJSON(SuccessResponse{
			Success: true,
			ID:      t.ID,
			Message: fmt.Sprintf("Updated ticket %s", t.ID),
		})
	}

	fmt.Printf("Updated ticket %s\n", t.ID)
	return nil
}

// Close marks a ticket as closed.
func Close(args []string) error {
	fs := flag.NewFlagSet("close", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket close [--json] <TICKET-ID>")
		fmt.Fprintln(os.Stderr, "\nClose a ticket.")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return thickerr.WithHint("Ticket ID is required", "Usage: thicket close <TICKET-ID>")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return thickerr.InvalidTicketID(ticketID)
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	t, err := store.Get(ticketID)
	if err != nil {
		return err
	}
	if t == nil {
		return thickerr.TicketNotFound(ticketID)
	}

	if t.Status == ticket.StatusClosed {
		if *jsonOutput {
			return printJSON(SuccessResponse{
				Success: true,
				ID:      t.ID,
				Message: fmt.Sprintf("Ticket %s is already closed", t.ID),
			})
		}
		fmt.Printf("Ticket %s is already closed\n", t.ID)
		return nil
	}

	t.Close()

	if err := store.Update(t); err != nil {
		return err
	}

	if *jsonOutput {
		return printJSON(SuccessResponse{
			Success: true,
			ID:      t.ID,
			Message: fmt.Sprintf("Closed ticket %s", t.ID),
		})
	}

	fmt.Printf("Closed ticket %s\n", t.ID)
	return nil
}

// Comment adds a comment to a ticket.
func Comment(args []string) error {
	fs := flag.NewFlagSet("comment", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket comment [--json] <TICKET-ID> \"Comment text\"")
		fmt.Fprintln(os.Stderr, "\nAdd a comment to a ticket.")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return thickerr.WithHint("Ticket ID is required", "Usage: thicket comment <TICKET-ID> \"Comment text\"")
	}
	if fs.NArg() < 2 {
		return thickerr.WithHint("Comment text is required", "Usage: thicket comment <TICKET-ID> \"Comment text\"")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return thickerr.InvalidTicketID(ticketID)
	}

	content := fs.Arg(1)
	if strings.TrimSpace(content) == "" {
		return thickerr.EmptyComment()
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	// Verify ticket exists
	t, err := store.Get(ticketID)
	if err != nil {
		return err
	}
	if t == nil {
		return thickerr.TicketNotFound(ticketID)
	}

	c, err := ticket.NewComment(ticketID, content)
	if err != nil {
		return err
	}

	if err := store.AddComment(c); err != nil {
		return err
	}

	if *jsonOutput {
		return printJSON(SuccessResponse{
			Success: true,
			ID:      c.ID,
			Message: fmt.Sprintf("Added comment %s to ticket %s", c.ID, ticketID),
		})
	}

	fmt.Printf("Added comment %s to ticket %s\n", c.ID, ticketID)
	return nil
}

// Link creates a dependency between two tickets.
func Link(args []string) error {
	fs := flag.NewFlagSet("link", flag.ExitOnError)
	blockedBy := fs.String("blocked-by", "", "Ticket that blocks this one")
	createdFrom := fs.String("created-from", "", "Ticket this was created from")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket link [flags] <TICKET-ID>")
		fmt.Fprintln(os.Stderr, "\nCreate a dependency relationship between tickets.")
		fmt.Fprintln(os.Stderr, "\nDependency Types:")
		fmt.Fprintln(os.Stderr, "  --blocked-by    Mark a ticket as blocked by another ticket")
		fmt.Fprintln(os.Stderr, "  --created-from  Track which ticket this was created from")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintln(os.Stderr, "  thicket link --blocked-by TH-def456 TH-abc123")
		fmt.Fprintln(os.Stderr, "  thicket link --created-from TH-def456 TH-abc123")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return thickerr.WithHint("Ticket ID is required", "Usage: thicket link <TICKET-ID> --blocked-by <ID>")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return thickerr.InvalidTicketID(ticketID)
	}

	if *blockedBy == "" && *createdFrom == "" {
		return thickerr.WithHint(
			"No dependency type specified",
			"Use --blocked-by or --created-from to specify the dependency type",
		)
	}

	if *blockedBy != "" && *createdFrom != "" {
		return thickerr.WithHint(
			"Cannot specify both --blocked-by and --created-from",
			"Use separate commands for different dependency types",
		)
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	// Verify the main ticket exists
	t, err := store.Get(ticketID)
	if err != nil {
		return err
	}
	if t == nil {
		return thickerr.TicketNotFound(ticketID)
	}

	var targetID string
	var depType ticket.DependencyType

	if *blockedBy != "" {
		targetID = normalizeTicketID(*blockedBy)
		depType = ticket.DependencyBlockedBy
	} else {
		targetID = normalizeTicketID(*createdFrom)
		depType = ticket.DependencyCreatedFrom
	}

	if err := ticket.ValidateID(targetID); err != nil {
		return thickerr.InvalidTicketID(targetID)
	}

	// Verify target ticket exists
	target, err := store.Get(targetID)
	if err != nil {
		return err
	}
	if target == nil {
		return thickerr.TicketNotFound(targetID)
	}

	dep, err := ticket.NewDependency(ticketID, targetID, depType)
	if err != nil {
		switch err {
		case ticket.ErrSelfDependency:
			return thickerr.SelfDependency()
		case ticket.ErrInvalidDependencyType:
			return thickerr.InvalidDependencyType(string(depType))
		default:
			return err
		}
	}

	if err := store.AddDependency(dep); err != nil {
		switch err {
		case ticket.ErrCircularDependency:
			return thickerr.CircularDependency()
		case ticket.ErrDuplicateDependency:
			return thickerr.DuplicateDependency()
		default:
			return err
		}
	}

	if *jsonOutput {
		msg := ""
		if depType == ticket.DependencyBlockedBy {
			msg = fmt.Sprintf("Ticket %s is now blocked by %s", ticketID, targetID)
		} else {
			msg = fmt.Sprintf("Ticket %s was created from %s", ticketID, targetID)
		}
		return printJSON(SuccessResponse{
			Success: true,
			ID:      dep.ID,
			Message: msg,
		})
	}

	if depType == ticket.DependencyBlockedBy {
		fmt.Printf("Ticket %s is now blocked by %s\n", ticketID, targetID)
	} else {
		fmt.Printf("Ticket %s was created from %s\n", ticketID, targetID)
	}

	return nil
}

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
