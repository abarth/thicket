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

// DependencyType represents the type of relationship between tickets.
type DependencyType string

const (
	// DependencyBlockedBy indicates that a ticket is blocked by another ticket.
	DependencyBlockedBy DependencyType = "blocked_by"
	// DependencyCreatedFrom indicates that a ticket was created from another ticket.
	DependencyCreatedFrom DependencyType = "created_from"
)

// Dependency represents a relationship between two tickets.
type Dependency struct {
	ID           string         `json:"id"`             // Format: TH-dXXXXXX (project code + d + 6 alphanumeric chars)
	FromTicketID string         `json:"from_ticket_id"` // The ticket that has the dependency
	ToTicketID   string         `json:"to_ticket_id"`   // The ticket being referenced
	Type         DependencyType `json:"type"`           // Type of dependency
	Created      time.Time      `json:"created"`        // Timestamp
}

var (
	ErrInvalidDependencyID   = errors.New("invalid dependency ID format")
	ErrInvalidDependencyType = errors.New("invalid dependency type")
	ErrSelfDependency        = errors.New("a ticket cannot depend on itself")
	ErrCircularDependency    = errors.New("this would create a circular dependency")
	ErrDuplicateDependency   = errors.New("this dependency already exists")
)

// dependencyIDPattern matches valid dependency IDs: two uppercase letters, hyphen, 'd', six alphanumeric chars.
var dependencyIDPattern = regexp.MustCompile(`^[A-Z]{2}-d[a-z0-9]{6}$`)

// ValidateDependencyType checks if a dependency type is valid.
func ValidateDependencyType(t DependencyType) error {
	switch t {
	case DependencyBlockedBy, DependencyCreatedFrom:
		return nil
	default:
		return ErrInvalidDependencyType
	}
}

// GenerateDependencyID creates a new dependency ID with the given project code.
func GenerateDependencyID(projectCode string) (string, error) {
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

	return fmt.Sprintf("%s-d%s", projectCode, string(result)), nil
}

// ValidateDependencyID checks if a dependency ID has the correct format.
func ValidateDependencyID(id string) error {
	if !dependencyIDPattern.MatchString(id) {
		return ErrInvalidDependencyID
	}
	return nil
}

// NewDependency creates a new dependency between two tickets.
func NewDependency(fromTicketID, toTicketID string, depType DependencyType) (*Dependency, error) {
	if err := ValidateID(fromTicketID); err != nil {
		return nil, fmt.Errorf("from ticket: %w", err)
	}
	if err := ValidateID(toTicketID); err != nil {
		return nil, fmt.Errorf("to ticket: %w", err)
	}

	if strings.EqualFold(fromTicketID, toTicketID) {
		return nil, ErrSelfDependency
	}

	if err := ValidateDependencyType(depType); err != nil {
		return nil, err
	}

	projectCode, err := ParseProjectCode(fromTicketID)
	if err != nil {
		return nil, err
	}

	id, err := GenerateDependencyID(projectCode)
	if err != nil {
		return nil, err
	}

	return &Dependency{
		ID:           id,
		FromTicketID: fromTicketID,
		ToTicketID:   toTicketID,
		Type:         depType,
		Created:      time.Now().UTC(),
	}, nil
}

// Validate checks if the dependency has valid field values.
func (d *Dependency) Validate() error {
	if err := ValidateDependencyID(d.ID); err != nil {
		return err
	}
	if err := ValidateID(d.FromTicketID); err != nil {
		return fmt.Errorf("from ticket: %w", err)
	}
	if err := ValidateID(d.ToTicketID); err != nil {
		return fmt.Errorf("to ticket: %w", err)
	}
	if strings.EqualFold(d.FromTicketID, d.ToTicketID) {
		return ErrSelfDependency
	}
	if err := ValidateDependencyType(d.Type); err != nil {
		return err
	}
	return nil
}
