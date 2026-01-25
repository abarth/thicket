# Thicket Ticket Tracker

*Thicket* is a light weight issue tracker for coding agents to use to track their work in other
projects. We want the agents to use this issue tracker instead of using Markdown files to track tasks. The issue tracker will store
authoritative data for the tickets in the `.thicket/tickets.jsonl` file, which is in JSONL format. Each ticket will be one line in
this file. This file will be tracked by git and shared by everyone working on the project. For efficiency, this data will be loaded
into a local SQLite database, which will need to be synchronized with `tickets.jsonl`. The CLI tool will automatically synchronize
the SQLite database with the JSONL file on every command to ensure the local cache remains up to date.

The tickets will be identified by a two-letter project code and a six-character random hex string (e.g., `TH-abcdef`),
which prevents people working on the project in parallel from creating colliding ticket identifiers. The tickets will need to have a
title, a description, and a status. The tickets will also need an integer priority as well as an issue type and an assignee. We will
need to keep track of important timestamps related to the tickets, such as when the ticket was created, last updated, and closed.

We will need to track dependencies between the tickets. The agents will create new tickets while they are working on existing tickets,
and we want to remember which ticket the agent was working on when it created the new ticket, which is one type of dependency. Another
important kind of dependency is when a ticket is blocked by the completion of another ticket.

We will need to enforce that we do not create circular blocking dependencies. An important query for the coding agents will be to get
the list of unblocked tickets in priority order. These are the tickets that are currently actionable.

We will also need a way to add comments to issues so that we can track what we learn about tickets over time. To keep the
`tickets.jsonl` file manageable and diff-friendly, comments will be stored in a separate lines in the `tickets.jsonl` file,
linked to their respective tickets by ID.

We should also be able to associated labels, which are short text strings, with tickets. These labels will be used to group tickets.

The primary way to interact with the ticket tracker will be through a command-line interface, which will be implemented in Go.
The CLI should support a `--json` output flag to facilitate integration with other agentic tools.
