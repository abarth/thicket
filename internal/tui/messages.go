package tui

import "github.com/abarth/thicket/internal/ticket"

// ViewTicketMsg is sent when the user wants to view a ticket's details.
type ViewTicketMsg struct {
	ID string
}

// BackToListMsg is sent when the user wants to return to the list view.
type BackToListMsg struct{}

// CreateTicketMsg is sent when the user wants to create a new ticket.
type CreateTicketMsg struct{}

// EditTicketMsg is sent when the user wants to edit a ticket.
type EditTicketMsg struct {
	Ticket *ticket.Ticket
}

// TicketSavedMsg is sent after a ticket has been saved.
type TicketSavedMsg struct {
	ID      string
	IsNew   bool
	Message string
}

// TicketClosedMsg is sent after a ticket has been closed.
type TicketClosedMsg struct {
	ID string
}

// TicketPriorityUpdatedMsg is sent after a ticket's priority has been updated.
type TicketPriorityUpdatedMsg struct {
	ID          string
	NewPriority int
}

// TicketTypeUpdatedMsg is sent after a ticket's type has been updated.
type TicketTypeUpdatedMsg struct {
	ID      string
	NewType ticket.Type
}

// RefreshListMsg is sent to refresh the ticket list.
type RefreshListMsg struct{}

// TicketsLoadedMsg is sent when tickets have been loaded from storage.
type TicketsLoadedMsg struct {
	Tickets []*ticket.Ticket
	Err     error
}

// TicketLoadedMsg is sent when a single ticket's details have been loaded.
type TicketLoadedMsg struct {
	Ticket    *ticket.Ticket
	Comments  []*ticket.Comment
	BlockedBy []*ticket.Ticket
	Blocking  []*ticket.Ticket
	Err       error
}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Err error
}

// StatusMsg is sent to display a status message.
type StatusMsg struct {
	Message string
	IsError bool
}

// AddCommentMsg is sent when the user wants to add a comment.
type AddCommentMsg struct {
	TicketID string
}

// CommentSavedMsg is sent after a comment has been saved.
type CommentSavedMsg struct {
	TicketID string
}

// ConfirmCloseMsg is sent to confirm closing a ticket.
type ConfirmCloseMsg struct {
	TicketID string
	Title    string
}
