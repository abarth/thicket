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
	StatusIcebox Status = "icebox"
)

// Type represents the category of a ticket.
type Type string

const (
	TypeBug     Type = "bug"
	TypeFeature Type = "feature"
	TypeTask    Type = "task"
	TypeEpic    Type = "epic"
	TypeCleanup Type = "cleanup"
)

// Ticket represents a single issue in the tracker.
type Ticket struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        Type      `json:"type"`
	Status      Status    `json:"status"`
	Priority    int       `json:"priority"`
	Labels      []string  `json:"labels"`
	Assignee    string    `json:"assignee"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

var (
	ErrInvalidID          = errors.New("invalid ticket ID format")
	ErrEmptyTitle         = errors.New("ticket title cannot be empty")
	ErrInvalidStatus      = errors.New("invalid ticket status")
	ErrInvalidType        = errors.New("invalid ticket type")
	ErrInvalidProjectCode = errors.New("project code must be exactly two uppercase letters")
	ErrInvalidLabel       = errors.New("label must be 1-30 alphanumeric characters, hyphens, or underscores")
)

// idPattern matches valid ticket IDs: two uppercase letters, hyphen, six alphanumeric chars.
var idPattern = regexp.MustCompile(`^[A-Z]{2}-[a-z0-9]{6}$`)

// projectCodePattern matches valid project codes: exactly two uppercase letters.
var projectCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)

// labelPattern matches valid labels: 1-30 alphanumeric characters, hyphens, or underscores.
var labelPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,30}$`)

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
	case StatusOpen, StatusClosed, StatusIcebox:
		return nil
	default:
		return ErrInvalidStatus
	}
}

// ValidateType checks if a type value is valid.
// An empty type is allowed for tickets where type is not specified.
func ValidateType(t Type) error {
	if t == "" {
		return nil
	}
	switch t {
	case TypeBug, TypeFeature, TypeTask, TypeEpic, TypeCleanup:
		return nil
	default:
		return ErrInvalidType
	}
}

// ValidateLabel checks if a label is valid.
func ValidateLabel(label string) error {
	if !labelPattern.MatchString(label) {
		return ErrInvalidLabel
	}
	return nil
}

// ValidateLabels checks if all labels in a slice are valid.
func ValidateLabels(labels []string) error {
	for _, label := range labels {
		if err := ValidateLabel(label); err != nil {
			return err
		}
	}
	return nil
}

// New creates a new ticket with the given parameters.
// New creates a new ticket with a generated ID.
func New(projectCode, title, description string, issueType Type, priority int, labels []string, assignee string) (*Ticket, error) {
	id, err := GenerateID(projectCode)
	if err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyTitle
	}

	if err := ValidateLabels(labels); err != nil {
		return nil, err
	}

	if err := ValidateType(issueType); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Ticket{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(description),
		Type:        issueType,
		Status:      StatusOpen,
		Priority:    priority,
		Labels:      labels,
		Assignee:    strings.TrimSpace(assignee),
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
	if err := ValidateType(t.Type); err != nil {
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
func (t *Ticket) Update(title, description *string, issueType *Type, priority *int, status *Status, addLabels, removeLabels []string, assignee *string) error {
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
	if issueType != nil {
		if err := ValidateType(*issueType); err != nil {
			return err
		}
		t.Type = *issueType
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

	// Handle label additions
	if len(addLabels) > 0 {
		if err := ValidateLabels(addLabels); err != nil {
			return err
		}
		// Add labels that don't already exist
		existing := make(map[string]bool)
		for _, l := range t.Labels {
			existing[l] = true
		}
		for _, l := range addLabels {
			if !existing[l] {
				t.Labels = append(t.Labels, l)
				existing[l] = true
			}
		}
	}

	// Handle label removals
	if len(removeLabels) > 0 {
		toRemove := make(map[string]bool)
		for _, l := range removeLabels {
			toRemove[l] = true
		}
		filtered := t.Labels[:0]
		for _, l := range t.Labels {
			if !toRemove[l] {
				filtered = append(filtered, l)
			}
		}
		t.Labels = filtered
	}

	// Handle assignee update
	if assignee != nil {
		t.Assignee = strings.TrimSpace(*assignee)
	}

	t.Updated = time.Now().UTC()
	return nil
}
