package ticket

import (
	"testing"
)

func TestValidateDependencyType(t *testing.T) {
	tests := []struct {
		name    string
		depType DependencyType
		wantErr bool
	}{
		{"blocked_by valid", DependencyBlockedBy, false},
		{"created_from valid", DependencyCreatedFrom, false},
		{"empty invalid", "", true},
		{"unknown invalid", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencyType(tt.depType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependencyType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateDependencyID(t *testing.T) {
	id, err := GenerateDependencyID("TH")
	if err != nil {
		t.Fatalf("GenerateDependencyID() error = %v", err)
	}

	if err := ValidateDependencyID(id); err != nil {
		t.Errorf("Generated ID %q is not valid: %v", id, err)
	}

	// Should start with project code + "-d"
	if id[:4] != "TH-d" {
		t.Errorf("Generated ID %q should start with TH-d", id)
	}
}

func TestValidateDependencyID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid", "TH-d123abc", false},
		{"valid other project", "AB-d000fff", false},
		{"valid alphanumeric", "TH-dz1y2x3", false},
		{"missing d prefix", "TH-123abc", true},
		{"uppercase alphanumeric", "TH-dABCDEF", true},
		{"too short", "TH-d12345", true},
		{"too long", "TH-d1234567", true},
		{"lowercase project code", "th-d123abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencyID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependencyID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestNewDependency(t *testing.T) {
	tests := []struct {
		name         string
		fromTicketID string
		toTicketID   string
		depType      DependencyType
		wantErr      bool
		errType      error
	}{
		{
			name:         "valid blocked_by",
			fromTicketID: "TH-111111",
			toTicketID:   "TH-222222",
			depType:      DependencyBlockedBy,
			wantErr:      false,
		},
		{
			name:         "valid created_from",
			fromTicketID: "TH-111111",
			toTicketID:   "TH-222222",
			depType:      DependencyCreatedFrom,
			wantErr:      false,
		},
		{
			name:         "self dependency",
			fromTicketID: "TH-111111",
			toTicketID:   "TH-111111",
			depType:      DependencyBlockedBy,
			wantErr:      true,
			errType:      ErrSelfDependency,
		},
		{
			name:         "invalid from ticket",
			fromTicketID: "invalid",
			toTicketID:   "TH-222222",
			depType:      DependencyBlockedBy,
			wantErr:      true,
		},
		{
			name:         "invalid to ticket",
			fromTicketID: "TH-111111",
			toTicketID:   "invalid",
			depType:      DependencyBlockedBy,
			wantErr:      true,
		},
		{
			name:         "invalid type",
			fromTicketID: "TH-111111",
			toTicketID:   "TH-222222",
			depType:      "invalid",
			wantErr:      true,
			errType:      ErrInvalidDependencyType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep, err := NewDependency(tt.fromTicketID, tt.toTicketID, tt.depType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDependency() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewDependency() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err := dep.Validate(); err != nil {
				t.Errorf("Created dependency is invalid: %v", err)
			}

			if dep.FromTicketID != tt.fromTicketID {
				t.Errorf("FromTicketID = %q, want %q", dep.FromTicketID, tt.fromTicketID)
			}
			if dep.ToTicketID != tt.toTicketID {
				t.Errorf("ToTicketID = %q, want %q", dep.ToTicketID, tt.toTicketID)
			}
			if dep.Type != tt.depType {
				t.Errorf("Type = %q, want %q", dep.Type, tt.depType)
			}
		})
	}
}

func TestDependency_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dep     Dependency
		wantErr bool
	}{
		{
			name: "valid",
			dep: Dependency{
				ID:           "TH-d123abc",
				FromTicketID: "TH-111111",
				ToTicketID:   "TH-222222",
				Type:         DependencyBlockedBy,
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			dep: Dependency{
				ID:           "invalid",
				FromTicketID: "TH-111111",
				ToTicketID:   "TH-222222",
				Type:         DependencyBlockedBy,
			},
			wantErr: true,
		},
		{
			name: "self dependency",
			dep: Dependency{
				ID:           "TH-d123abc",
				FromTicketID: "TH-111111",
				ToTicketID:   "TH-111111",
				Type:         DependencyBlockedBy,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dep.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
