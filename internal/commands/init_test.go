package commands

import (
	"testing"

	"github.com/abarth/thicket/internal/config"
)

func TestInit(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	err := Init([]string{"--project", "TH"})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify project was created
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg.ProjectCode != "TH" {
		t.Errorf("ProjectCode = %q, want TH", cfg.ProjectCode)
	}
}

func TestInit_LowercaseCode(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	err := Init([]string{"--project", "th"})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Should be uppercased
	root, _ := config.FindRoot()
	cfg, _ := config.Load(root)
	if cfg.ProjectCode != "TH" {
		t.Errorf("ProjectCode = %q, want TH (uppercased)", cfg.ProjectCode)
	}
}

func TestInit_MissingProject(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	err := Init([]string{})
	if err == nil {
		t.Error("Init() expected error for missing --project")
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	if err := Init([]string{"--project", "TH"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Init([]string{"--project", "TH"})
	if err == nil {
		t.Error("Init() expected error for already initialized")
	}
}
