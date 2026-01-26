// Package commands implements the CLI commands for Thicket.
package commands

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
)

// normalizeTicketID normalizes a ticket ID to the canonical format:
// uppercase project code, lowercase hex portion (e.g., "th-ABCDEF" -> "TH-abcdef").
func normalizeTicketID(id string) string {
	if len(id) < 3 || id[2] != '-' {
		return id
	}
	return strings.ToUpper(id[:2]) + "-" + strings.ToLower(id[3:])
}

// Init initializes a new Thicket project.
func Init(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	projectCode := fs.String("project", "", "Two-letter project code (e.g., TH)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket init --project <CODE>")
		fmt.Fprintln(os.Stderr, "\nInitialize a new Thicket project in the current directory.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *projectCode == "" {
		return errors.New("--project flag is required")
	}

	*projectCode = strings.ToUpper(*projectCode)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	if err := config.Init(wd, *projectCode); err != nil {
		return err
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
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket add --title <TITLE> [--description <DESC>] [--priority <N>]")
		fmt.Fprintln(os.Stderr, "\nCreate a new ticket.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *title == "" {
		return errors.New("--title flag is required")
	}

	root, err := config.FindRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(root)
	if err != nil {
		return err
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

	fmt.Printf("Created ticket %s\n", t.ID)
	return nil
}

// List displays tickets.
func List(args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	statusFilter := fs.String("status", "", "Filter by status (open, closed)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket list [--status <STATUS>]")
		fmt.Fprintln(os.Stderr, "\nList tickets, ordered by priority.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	root, err := config.FindRoot()
	if err != nil {
		return err
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
			return fmt.Errorf("invalid status: %s (use 'open' or 'closed')", *statusFilter)
		}
		status = &s
	}

	tickets, err := store.List(status)
	if err != nil {
		return err
	}

	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
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
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket show <TICKET-ID>")
		fmt.Fprintln(os.Stderr, "\nDisplay details of a specific ticket.")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return errors.New("ticket ID is required")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return fmt.Errorf("invalid ticket ID: %s", ticketID)
	}

	root, err := config.FindRoot()
	if err != nil {
		return err
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
		return ErrTicketNotFound
	}

	printTicketDetail(os.Stdout, t)
	return nil
}

func printTicketDetail(w io.Writer, t *ticket.Ticket) {
	fmt.Fprintf(w, "ID:          %s\n", t.ID)
	fmt.Fprintf(w, "Title:       %s\n", t.Title)
	fmt.Fprintf(w, "Status:      %s\n", t.Status)
	fmt.Fprintf(w, "Priority:    %d\n", t.Priority)
	fmt.Fprintf(w, "Created:     %s\n", t.Created.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated:     %s\n", t.Updated.Format(time.RFC3339))
	if t.Description != "" {
		fmt.Fprintf(w, "\nDescription:\n%s\n", t.Description)
	}
}

// Update modifies an existing ticket.
func Update(args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	title := fs.String("title", "", "New title")
	description := fs.String("description", "", "New description")
	priority := fs.Int("priority", -1, "New priority")
	status := fs.String("status", "", "New status (open, closed)")
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
		return errors.New("ticket ID is required")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return fmt.Errorf("invalid ticket ID: %s", ticketID)
	}

	root, err := config.FindRoot()
	if err != nil {
		return err
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
		return ErrTicketNotFound
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
			return fmt.Errorf("invalid status: %s (use 'open' or 'closed')", *status)
		}
		statusPtr = &s
	}

	if titlePtr == nil && descPtr == nil && priorityPtr == nil && statusPtr == nil {
		return errors.New("no fields to update (use --title, --description, --priority, or --status)")
	}

	if err := t.Update(titlePtr, descPtr, priorityPtr, statusPtr); err != nil {
		return err
	}

	if err := store.Update(t); err != nil {
		return err
	}

	fmt.Printf("Updated ticket %s\n", t.ID)
	return nil
}

// Close marks a ticket as closed.
func Close(args []string) error {
	fs := flag.NewFlagSet("close", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket close <TICKET-ID>")
		fmt.Fprintln(os.Stderr, "\nClose a ticket.")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return errors.New("ticket ID is required")
	}

	ticketID := normalizeTicketID(fs.Arg(0))
	if err := ticket.ValidateID(ticketID); err != nil {
		return fmt.Errorf("invalid ticket ID: %s", ticketID)
	}

	root, err := config.FindRoot()
	if err != nil {
		return err
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
		return ErrTicketNotFound
	}

	if t.Status == ticket.StatusClosed {
		fmt.Printf("Ticket %s is already closed\n", t.ID)
		return nil
	}

	t.Close()

	if err := store.Update(t); err != nil {
		return err
	}

	fmt.Printf("Closed ticket %s\n", t.ID)
	return nil
}
