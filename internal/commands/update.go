package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// Update modifies an existing ticket.
func Update(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("update")
	title := fs.String("title", "", "New title")
	description := fs.String("description", "", "New description")
	issueType := fs.String("type", "", "New type")
	priority := fs.Int("priority", -1, "New priority")
	status := fs.String("status", "", "New status (open, closed)")
	assignee := fs.String("assignee", "", "Assign ticket to person (use empty string to clear)")
	var addLabels labelSlice
	var removeLabels labelSlice
	fs.Var(&addLabels, "add-label", "Add a label (can be specified multiple times)")
	fs.Var(&removeLabels, "remove-label", "Remove a label (can be specified multiple times)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket update [flags] <TICKET-ID>")
		fmt.Fprintln(os.Stderr, "\nUpdate an existing ticket. Only specified fields are changed.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

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
	var typePtr *ticket.Type
	var priorityPtr *int
	var statusPtr *ticket.Status
	var assigneePtr *string

	if *title != "" {
		titlePtr = title
	}
	if *description != "" {
		descPtr = description
	}
	if *issueType != "" {
		t := ticket.Type(*issueType)
		typePtr = &t
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

	// Check if --assignee was explicitly provided (even if empty, to clear it)
	assigneeSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "assignee" {
			assigneeSet = true
		}
	})
	if assigneeSet {
		assigneePtr = assignee
	}

	if titlePtr == nil && descPtr == nil && typePtr == nil && priorityPtr == nil && statusPtr == nil && assigneePtr == nil && len(addLabels) == 0 && len(removeLabels) == 0 {
		return thickerr.WithHint(
			"No fields to update",
			"Use --title, --description, --type, --priority, --status, --assignee, --add-label, or --remove-label to specify changes",
		)
	}

	if err := t.Update(titlePtr, descPtr, typePtr, priorityPtr, statusPtr, addLabels, removeLabels, assigneePtr); err != nil {
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
