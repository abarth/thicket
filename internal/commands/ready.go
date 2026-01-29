package commands

import (
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
)

// Ready displays the highest priority open ticket that is not blocked by other open tickets.
func Ready(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("ready")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket ready [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nShow the highest priority actionable ticket (not blocked by others).")
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

	if len(tickets) == 0 {
		if *jsonOutput {
			return printJSON(map[string]interface{}{
				"message": "No ready tickets found",
				"ticket":  nil,
			})
		}
		fmt.Println("No ready tickets found.")
		return nil
	}

	// Get the highest priority ticket (first in the sorted list)
	t := tickets[0]

	// Get full ticket details
	comments, err := store.GetComments(t.ID)
	if err != nil {
		return err
	}

	blockedBy, err := store.GetBlockers(t.ID)
	if err != nil {
		return err
	}

	blocking, err := store.GetBlocking(t.ID)
	if err != nil {
		return err
	}

	createdFrom, err := store.GetCreatedFrom(t.ID)
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
