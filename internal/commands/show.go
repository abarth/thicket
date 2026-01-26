package commands

import (
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// Show displays a single ticket.
func Show(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("show")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket show <TICKET-ID> [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nDisplay details of a specific ticket.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

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
