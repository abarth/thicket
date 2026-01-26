// Package errors provides user-friendly error types for Thicket.
package errors

import (
	"fmt"
)

// UserError represents an error that should be displayed to the user.
// These errors have user-friendly messages and optional hints.
type UserError struct {
	Message string
	Hint    string
}

func (e *UserError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\n\nHint: %s", e.Message, e.Hint)
	}
	return e.Message
}

// New creates a new UserError with the given message.
func New(message string) *UserError {
	return &UserError{Message: message}
}

// WithHint creates a new UserError with a message and hint.
func WithHint(message, hint string) *UserError {
	return &UserError{Message: message, Hint: hint}
}

// NotInitialized returns an error for when Thicket is not initialized.
func NotInitialized() *UserError {
	return WithHint(
		"Thicket is not initialized in this directory",
		"Run 'thicket init --project <CODE>' to initialize a project",
	)
}

// TicketNotFound returns an error for when a ticket is not found.
func TicketNotFound(id string) *UserError {
	return WithHint(
		fmt.Sprintf("Ticket %s not found", id),
		"Run 'thicket list' to see available tickets",
	)
}

// InvalidTicketID returns an error for invalid ticket ID format.
func InvalidTicketID(id string) *UserError {
	return WithHint(
		fmt.Sprintf("Invalid ticket ID: %s", id),
		"Ticket IDs have the format XX-xxxxxx (e.g., TH-abc123)",
	)
}

// InvalidProjectCode returns an error for invalid project code.
func InvalidProjectCode(code string) *UserError {
	return WithHint(
		fmt.Sprintf("Invalid project code: %s", code),
		"Project code must be exactly two letters (e.g., TH)",
	)
}

// MissingRequired returns an error for missing required flags.
func MissingRequired(flag string) *UserError {
	return &UserError{
		Message: fmt.Sprintf("Missing required flag: --%s", flag),
	}
}

// InvalidStatus returns an error for invalid status values.
func InvalidStatus(status string) *UserError {
	return WithHint(
		fmt.Sprintf("Invalid status: %s", status),
		"Valid statuses are: open, closed",
	)
}

// StatusReadySuggestion returns an error suggesting the ready command.
func StatusReadySuggestion() *UserError {
	return WithHint(
		"'ready' is not a valid status",
		"Did you mean 'thicket ready'? This command shows tickets that are ready to work on.",
	)
}

// EmptyComment returns an error for empty comment content.
func EmptyComment() *UserError {
	return &UserError{
		Message: "Comment content cannot be empty",
	}
}

// CommentNotFound returns an error for when a comment is not found.
func CommentNotFound(id string) *UserError {
	return &UserError{
		Message: fmt.Sprintf("Comment %s not found", id),
	}
}

// CircularDependency returns an error for circular blocking dependencies.
func CircularDependency() *UserError {
	return WithHint(
		"This would create a circular dependency",
		"A ticket cannot be blocked by a ticket that it transitively blocks",
	)
}

// DuplicateDependency returns an error for duplicate dependencies.
func DuplicateDependency() *UserError {
	return &UserError{
		Message: "This dependency already exists",
	}
}

// SelfDependency returns an error for self-referential dependencies.
func SelfDependency() *UserError {
	return &UserError{
		Message: "A ticket cannot depend on itself",
	}
}

// InvalidDependencyType returns an error for invalid dependency types.
func InvalidDependencyType(depType string) *UserError {
	return WithHint(
		fmt.Sprintf("Invalid dependency type: %s", depType),
		"Valid types are: blocked_by, created_from",
	)
}
