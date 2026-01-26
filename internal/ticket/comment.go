// Package ticket defines the core ticket data model and validation.
package ticket

import (
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Comment represents a comment on a ticket.
type Comment struct {
	ID       string    `json:"id"`        // Format: TH-cXXXXXX (project code + c + 6 alphanumeric chars)
	TicketID string    `json:"ticket_id"` // The ticket this comment belongs to
	Content  string    `json:"content"`   // Comment text
	Created  time.Time `json:"created"`   // Timestamp
}

var (
	ErrEmptyComment    = errors.New("comment content cannot be empty")
	ErrInvalidCommentID = errors.New("invalid comment ID format")
)

// commentIDPattern matches valid comment IDs: two uppercase letters, hyphen, 'c', six alphanumeric chars.
var commentIDPattern = regexp.MustCompile(`^[A-Z]{2}-c[a-z0-9]{6}$`)

// GenerateCommentID creates a new comment ID with the given project code.
func GenerateCommentID(projectCode string) (string, error) {
	if err := ValidateProjectCode(projectCode); err != nil {
		return "", err
	}

	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 6)
	if _, err := rand.Read(result); err != nil {
		return "", fmt.Errorf("generating random ID: %w", err)
	}

	for i := 0; i < len(result); i++ {
		result[i] = charset[result[i]%byte(len(charset))]
	}

	return fmt.Sprintf("%s-c%s", projectCode, string(result)), nil
}

// ValidateCommentID checks if a comment ID has the correct format.
func ValidateCommentID(id string) error {
	if !commentIDPattern.MatchString(id) {
		return ErrInvalidCommentID
	}
	return nil
}

// NewComment creates a new comment for a ticket.
func NewComment(ticketID, content string) (*Comment, error) {
	if err := ValidateID(ticketID); err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyComment
	}

	projectCode, err := ParseProjectCode(ticketID)
	if err != nil {
		return nil, err
	}

	id, err := GenerateCommentID(projectCode)
	if err != nil {
		return nil, err
	}

	return &Comment{
		ID:       id,
		TicketID: ticketID,
		Content:  content,
		Created:  time.Now().UTC(),
	}, nil
}

// Validate checks if the comment has valid field values.
func (c *Comment) Validate() error {
	if err := ValidateCommentID(c.ID); err != nil {
		return err
	}
	if err := ValidateID(c.TicketID); err != nil {
		return err
	}
	if strings.TrimSpace(c.Content) == "" {
		return ErrEmptyComment
	}
	return nil
}
