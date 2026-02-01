package commands

import (
	"bufio"
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
	skipWorkflow := fs.Bool("skip-workflow", false, "Skip creating workflow command files")
	forceOverwrite := fs.Bool("force", false, "Overwrite existing workflow files without prompting")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: thicket init --project <CODE> [--json] [--data-dir <DIR>] [--skip-workflow] [--force]")
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

	// Setup workflow files unless skipped
	var workflowCreated bool
	if !*skipWorkflow {
		workflowCreated, err = setupWorkflow(wd, *forceOverwrite, *jsonOutput)
		if err != nil {
			return err
		}
	}

	if *jsonOutput {
		return printJSON(InitResponse{
			SuccessResponse: SuccessResponse{
				Success: true,
				Message: fmt.Sprintf("Initialized Thicket project with code %s", *projectCode),
			},
			WorkflowCreated: workflowCreated,
		})
	}

	fmt.Printf("Initialized Thicket project with code %s\n", *projectCode)
	if workflowCreated {
		fmt.Println("Created workflow files:")
		fmt.Println("  .agent/workflows/crank.md")
		fmt.Println("  .claude/commands/crank.md -> .agent/workflows/crank.md")
	}
	return nil
}

// InitResponse extends SuccessResponse with workflow creation info.
type InitResponse struct {
	SuccessResponse
	WorkflowCreated bool `json:"workflow_created"`
}

// setupWorkflow handles creating workflow files with user confirmation if needed.
func setupWorkflow(root string, forceOverwrite, jsonOutput bool) (bool, error) {
	existing := config.WorkflowFilesExist(root)

	if len(existing) > 0 && !forceOverwrite {
		if jsonOutput {
			// In JSON mode, don't prompt - return without creating
			return false, nil
		}

		fmt.Println("\nThe following workflow files already exist:")
		for _, f := range existing {
			fmt.Printf("  %s\n", f)
		}
		fmt.Print("Overwrite these files? [y/N] ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("reading response: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Skipping workflow file creation.")
			return false, nil
		}
	}

	if err := config.SetupWorkflowFiles(root); err != nil {
		return false, fmt.Errorf("setting up workflow files: %w", err)
	}

	return true, nil
}
