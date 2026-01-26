package commands

import (
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// Ready displays open tickets that are not blocked by other open tickets.
func Ready(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("ready")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket ready [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nList actionable open tickets (not blocked by others).")
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
