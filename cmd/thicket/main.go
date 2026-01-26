// Package main provides the CLI entry point for Thicket.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/abarth/thicket/internal/commands"
	"github.com/abarth/thicket/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("thicket", flag.ExitOnError)
	dataDir := fs.String("data-dir", "", "Custom .thicket directory")
	fs.Usage = printUsage

	// We want to parse global flags before the command.
	// flag.Parse() would consume everything, but NewFlagSet.Parse()
	// stops at the first non-flag argument if we don't use flag.CommandLine.
	// Actually, by default it continues. We need to handle this.
	
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *dataDir != "" {
		config.SetDataDir(*dataDir)
	}

	args := fs.Args()
	if len(args) == 0 {
		printUsage()
		return nil
	}

	cmd := args[0]
	// Use the remaining arguments for the command
	remainingArgs := args[1:]

	switch cmd {
	case "init":
		return commands.Init(remainingArgs)
	case "add":
		return commands.Add(remainingArgs)
	case "list", "ls":
		return commands.List(remainingArgs)
	case "ready":
		return commands.Ready(remainingArgs)
	case "show":
		return commands.Show(remainingArgs)
	case "update":
		return commands.Update(remainingArgs)
	case "close":
		return commands.Close(remainingArgs)
	case "comment":
		return commands.Comment(remainingArgs)
	case "link":
		return commands.Link(remainingArgs)
	case "quickstart":
		return commands.Quickstart(remainingArgs)
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "version", "-v", "--version":
		fmt.Println("thicket version 0.1.0")
		return nil
	default:
		return fmt.Errorf("unknown command: %s\nRun 'thicket help' for usage", cmd)
	}
}

func printUsage() {
	fmt.Println(`Thicket - A lightweight issue tracker for coding agents

Usage:
  thicket <command> [arguments]

Global Flags:
  --data-dir  Custom .thicket directory location
  --json      Output in JSON format (available for most commands)

Commands:
  init        Initialize a new Thicket project
  add         Create a new ticket
  list        List tickets (alias: ls)
  ready       List actionable open tickets
  show        Display a ticket
  update      Modify a ticket
  close       Close a ticket
  comment     Add a comment to a ticket
  link        Create dependencies between tickets
  quickstart  Show guide for coding agents
  help        Show this help message
  version     Show version information

Run 'thicket <command> --help' for more information on a command.

Examples:
  thicket init --json --project TH
  thicket add --interactive
  thicket add --json --title "Fix bug" --priority 1
  thicket list --json --status open
  thicket show --json TH-abc123
  thicket update --json --priority 2 TH-abc123
  thicket close --json TH-abc123
  thicket comment --json TH-abc123 "Working on this now"
  thicket link --json TH-abc123 --blocked-by TH-def456`)
}
