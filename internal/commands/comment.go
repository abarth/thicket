package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// Comment adds a comment to a ticket.
func Comment(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("comment")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket comment <TICKET-ID> <MESSAGE> [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nAdd a comment to a ticket.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

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
