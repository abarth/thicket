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
	Hint    string `json:"hint,omitempty"`
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// newFlagSet creates a new FlagSet with global flags already defined.
func newFlagSet(name string) (*flag.FlagSet, *bool, *string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	dataDir := fs.String("data-dir", "", "Custom .thicket directory location")
	return fs, jsonOutput, dataDir
}

// handleGlobalFlags sets global configuration based on flags.
func handleGlobalFlags(dataDir string) {
	if dataDir != "" {
		config.SetDataDir(dataDir)
	}
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

func printTicketDetail(w io.Writer, details *TicketDetails) {
	t := details.Ticket
	fmt.Fprintf(w, "ID:          %s\n", t.ID)
	fmt.Fprintf(w, "Title:       %s\n", t.Title)
	fmt.Fprintf(w, "Status:      %s\n", t.Status)
	fmt.Fprintf(w, "Priority:    %d\n", t.Priority)
	if len(t.Labels) > 0 {
		fmt.Fprintf(w, "Labels:      %s\n", strings.Join(t.Labels, ", "))
	}
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
