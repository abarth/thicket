package commands

import (
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// Close marks a ticket as closed.
func Close(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("close")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket close <TICKET-ID> [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nClose a ticket.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

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
