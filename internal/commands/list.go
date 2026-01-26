package commands

import (
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// List displays tickets.
func List(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("list")
	statusFilter := fs.String("status", "", "Filter by status (open, closed)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket list [--status <STATUS>] [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nList tickets, ordered by priority.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

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
