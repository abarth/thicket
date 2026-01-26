package commands

import (
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// Link creates a dependency between tickets.
func Link(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("link")
	blockedBy := fs.String("blocked-by", "", "Ticket that blocks this one")
	createdFrom := fs.String("created-from", "", "Ticket this was created from")
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

	handleGlobalFlags(*dataDir)

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
