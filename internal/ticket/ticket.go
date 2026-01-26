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

// Status represents the state of a ticket.
type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

// Ticket represents a single issue in the tracker.
type Ticket struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Priority    int       `json:"priority"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

var (
	ErrInvalidID          = errors.New("invalid ticket ID format")
	ErrEmptyTitle         = errors.New("ticket title cannot be empty")
	ErrInvalidStatus      = errors.New("invalid ticket status")
	ErrInvalidProjectCode = errors.New("project code must be exactly two uppercase letters")
)

// idPattern matches valid ticket IDs: two uppercase letters, hyphen, six alphanumeric chars.
var idPattern = regexp.MustCompile(`^[A-Z]{2}-[a-z0-9]{6}$`)

// projectCodePattern matches valid project codes: exactly two uppercase letters.
var projectCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)

// ValidateProjectCode checks if a project code is valid.
func ValidateProjectCode(code string) error {
	if !projectCodePattern.MatchString(code) {
		return ErrInvalidProjectCode
	}
	return nil
}

// GenerateID creates a new ticket ID with the given project code.
func GenerateID(projectCode string) (string, error) {
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

	return fmt.Sprintf("%s-%s", projectCode, string(result)), nil
}

// ValidateID checks if a ticket ID has the correct format.
func ValidateID(id string) error {
	if !idPattern.MatchString(id) {
		return ErrInvalidID
	}
	return nil
}

// ParseProjectCode extracts the project code from a ticket ID.
func ParseProjectCode(id string) (string, error) {
	if err := ValidateID(id); err != nil {
		return "", err
	}
	return id[:2], nil
}

// ValidateStatus checks if a status value is valid.
func ValidateStatus(s Status) error {
	switch s {
	case StatusOpen, StatusClosed:
		return nil
	default:
		return ErrInvalidStatus
	}
}

// New creates a new ticket with the given parameters.
func New(projectCode, title, description string, priority int) (*Ticket, error) {
	id, err := GenerateID(projectCode)
	if err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyTitle
	}

	now := time.Now().UTC()
	return &Ticket{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(description),
		Status:      StatusOpen,
		Priority:    priority,
		Created:     now,
		Updated:     now,
	}, nil
}

// Validate checks if the ticket has valid field values.
func (t *Ticket) Validate() error {
	if err := ValidateID(t.ID); err != nil {
		return err
	}
	if strings.TrimSpace(t.Title) == "" {
		return ErrEmptyTitle
	}
	if err := ValidateStatus(t.Status); err != nil {
		return err
	}
	return nil
}

// Close marks the ticket as closed and updates the timestamp.
func (t *Ticket) Close() {
	t.Status = StatusClosed
	t.Updated = time.Now().UTC()
}

// Update modifies the ticket fields and updates the timestamp.
func (t *Ticket) Update(title, description *string, priority *int, status *Status) error {
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return ErrEmptyTitle
		}
		t.Title = trimmed
	}
	if description != nil {
		t.Description = strings.TrimSpace(*description)
	}
	if priority != nil {
		t.Priority = *priority
	}
	if status != nil {
		if err := ValidateStatus(*status); err != nil {
			return err
		}
		t.Status = *status
	}
	t.Updated = time.Now().UTC()
	return nil
}
