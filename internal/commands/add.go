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

// Add creates a new ticket.
func Add(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("add")
	title := fs.String("title", "", "Ticket title")
	description := fs.String("description", "", "Ticket description")
	priority := fs.Int("priority", 0, "Ticket priority (lower = higher priority)")
	interactive := fs.Bool("interactive", false, "Interactive mode")
	fs.BoolVar(interactive, "i", false, "Interactive mode (shorthand)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket add [--interactive] [--title <TITLE>] [--description <DESC>] [--priority <N>] [--json] [--data-dir <DIR>]")
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

	t, err := ticket.New(cfg.ProjectCode, *title, *description, *priority)
	if err != nil {
		return err
	}

	if err := store.Add(t); err != nil {
		return err
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
