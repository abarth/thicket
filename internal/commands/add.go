package commands

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// labelSlice is a custom flag type that collects multiple --label flags.
type labelSlice []string

func (l *labelSlice) String() string {
	return strings.Join(*l, ",")
}

func (l *labelSlice) Set(value string) error {
	*l = append(*l, value)
	return nil
}

// Add creates a new ticket.
func Add(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("add")
	title := fs.String("title", "", "Ticket title")
	description := fs.String("description", "", "Ticket description")
	priority := fs.Int("priority", 0, "Ticket priority (lower = higher priority)")
	assignee := fs.String("assignee", "", "Assign ticket to person")
	interactive := fs.Bool("interactive", false, "Interactive mode")
	blocks := fs.String("blocks", "", "Existing ticket that is blocked by this new ticket")
	blockedBy := fs.String("blocked-by", "", "Existing ticket that blocks this new ticket")
	createdFrom := fs.String("created-from", "", "Existing ticket this was created from")
	var labels labelSlice
	fs.Var(&labels, "label", "Add a label (can be specified multiple times)")

	fs.BoolVar(interactive, "i", false, "Interactive mode (shorthand)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket add [--interactive] [--title <TITLE>] [--description <DESC>] [--priority <N>] [--assignee <NAME>] [--label <LABEL>]... [--blocks <ID>] [--blocked-by <ID>] [--created-from <ID>] [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nCreate a new ticket.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

	if *interactive {
		scanner := bufio.NewScanner(os.Stdin)
		if *title == "" {
			fmt.Print("Title: ")
			if scanner.Scan() {
				*title = strings.TrimSpace(scanner.Text())
			}
		}
		if *description == "" {
			fmt.Print("Description: ")
			if scanner.Scan() {
				*description = strings.TrimSpace(scanner.Text())
			}
		}

		// Check if priority was explicitly set
		prioritySet := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "priority" {
				prioritySet = true
			}
		})

		if !prioritySet {
			fmt.Print("Priority [2]: ")
			if scanner.Scan() {
				val := strings.TrimSpace(scanner.Text())
				if val == "" {
					*priority = 2
				} else {
					if p, err := strconv.Atoi(val); err == nil {
						*priority = p
					} else {
						*priority = 2
					}
				}
			} else {
				*priority = 2
			}
		}
	}

	if *title == "" {
		return thickerr.MissingRequired("title")
	}

	root, err := config.FindRoot()
	if err != nil {
		return wrapConfigError(err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		return wrapConfigError(err)
	}

	paths := config.GetPaths(root)
	store, err := storage.Open(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	t, err := ticket.New(cfg.ProjectCode, *title, *description, *priority, labels, *assignee)
	if err != nil {
		return err
	}

	if err := store.Add(t); err != nil {
		return err
	}

	// Create links if specified
	if *blocks != "" {
		targetID := normalizeTicketID(*blocks)
		dep, err := ticket.NewDependency(targetID, t.ID, ticket.DependencyBlockedBy)
		if err != nil {
			return err
		}
		if err := store.AddDependency(dep); err != nil {
			return err
		}
	}
	if *blockedBy != "" {
		targetID := normalizeTicketID(*blockedBy)
		dep, err := ticket.NewDependency(t.ID, targetID, ticket.DependencyBlockedBy)
		if err != nil {
			return err
		}
		if err := store.AddDependency(dep); err != nil {
			return err
		}
	}
	if *createdFrom != "" {
		targetID := normalizeTicketID(*createdFrom)
		dep, err := ticket.NewDependency(t.ID, targetID, ticket.DependencyCreatedFrom)
		if err != nil {
			return err
		}
		if err := store.AddDependency(dep); err != nil {
			return err
		}
	}

	if *jsonOutput {
		return printJSON(SuccessResponse{
			Success: true,
			ID:      t.ID,
			Message: fmt.Sprintf("Created ticket %s", t.ID),
		})
	}

	fmt.Printf("Created ticket %s\n", t.ID)
	return nil
}
