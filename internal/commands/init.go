package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/abarth/thicket/internal/config"
	thickerr "github.com/abarth/thicket/internal/errors"
	"github.com/abarth/thicket/internal/ticket"
)

// Init initializes a new Thicket project.
func Init(args []string) error {
	fs, jsonOutput, dataDir := newFlagSet("init")
	projectCode := fs.String("project", "", "Two-letter project code (e.g., TH)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket init --project <CODE> [--json] [--data-dir <DIR>]")
		fmt.Fprintln(os.Stderr, "\nInitialize a new Thicket project in the current directory.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	handleGlobalFlags(*dataDir)

	if *projectCode == "" {
		return thickerr.MissingRequired("project")
	}

	*projectCode = strings.ToUpper(*projectCode)

	if err := ticket.ValidateProjectCode(*projectCode); err != nil {
		return thickerr.InvalidProjectCode(*projectCode)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	if err := config.Init(wd, *projectCode); err != nil {
		return wrapConfigError(err)
	}

	if *jsonOutput {
		return printJSON(SuccessResponse{
			Success: true,
			Message: fmt.Sprintf("Initialized Thicket project with code %s", *projectCode),
		})
	}

	fmt.Printf("Initialized Thicket project with code %s\n", *projectCode)
	return nil
}
