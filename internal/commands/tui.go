package commands

import (
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/tui"
)

// TUI launches the interactive terminal UI.
func TUI(args []string) error {
	fs, _, dataDir := newFlagSet("tui")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket tui [flags]")
		fmt.Fprintln(os.Stderr, "\nLaunch interactive terminal UI for managing tickets.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nKey bindings:")
		fmt.Fprintln(os.Stderr, "  List view:")
		fmt.Fprintln(os.Stderr, "    j/k, arrows   Navigate up/down")
		fmt.Fprintln(os.Stderr, "    g, home       Go to top")
		fmt.Fprintln(os.Stderr, "    G, end        Go to bottom")
		fmt.Fprintln(os.Stderr, "    Enter, l      View ticket details")
		fmt.Fprintln(os.Stderr, "    n             Create new ticket")
		fmt.Fprintln(os.Stderr, "    e             Edit selected ticket")
		fmt.Fprintln(os.Stderr, "    c             Close selected ticket")
		fmt.Fprintln(os.Stderr, "    +/=           Lower priority (increment priority value)")
		fmt.Fprintln(os.Stderr, "    -/_           Higher priority (decrement priority value)")
		fmt.Fprintln(os.Stderr, "    o/x/i/a       Filter: open/closed/icebox/all")
		fmt.Fprintln(os.Stderr, "    r             Refresh list")
		fmt.Fprintln(os.Stderr, "    q             Quit")
		fmt.Fprintln(os.Stderr, "    ?             Show help")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  Detail view:")
		fmt.Fprintln(os.Stderr, "    Esc, h, bksp  Back to list")
		fmt.Fprintln(os.Stderr, "    e             Edit ticket")
		fmt.Fprintln(os.Stderr, "    c             Close ticket")
		fmt.Fprintln(os.Stderr, "    m             Add comment")
		fmt.Fprintln(os.Stderr, "    j/k, arrows   Scroll description/comments")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  Form view:")
		fmt.Fprintln(os.Stderr, "    Tab           Next field")
		fmt.Fprintln(os.Stderr, "    Shift+Tab     Previous field")
		fmt.Fprintln(os.Stderr, "    Ctrl+S        Save")
		fmt.Fprintln(os.Stderr, "    Esc           Cancel")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

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

	return tui.Run(store, cfg, paths.Tickets)
}
