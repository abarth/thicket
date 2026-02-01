package commands

import (
	"os"
	"path/filepath"
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

	// Verify workflow files were created
	workflowPath := filepath.Join(dir, ".agent/workflows/crank.md")
	if _, err := os.Stat(workflowPath); err != nil {
		t.Errorf("workflow file not created: %v", err)
	}

	symlinkPath := filepath.Join(dir, ".claude/commands/crank.md")
	linkTarget, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Errorf("symlink not created: %v", err)
	}
	if linkTarget != "../../.agent/workflows/crank.md" {
		t.Errorf("symlink target = %q, want ../../.agent/workflows/crank.md", linkTarget)
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

func TestInit_SkipWorkflow(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	err := Init([]string{"--project", "TH", "--skip-workflow"})
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

	// Verify workflow files were NOT created
	workflowPath := filepath.Join(dir, ".agent/workflows/crank.md")
	if _, err := os.Stat(workflowPath); !os.IsNotExist(err) {
		t.Errorf("workflow file should not exist with --skip-workflow")
	}
}

func TestInit_ForceOverwrite(t *testing.T) {
	dir, cleanup := setupTestProject(t)
	defer cleanup()

	// Create existing workflow file with old content
	workflowDir := filepath.Join(dir, ".agent/workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	workflowPath := filepath.Join(workflowDir, "crank.md")
	if err := os.WriteFile(workflowPath, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	err := Init([]string{"--project", "TH", "--force"})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify workflow file was overwritten
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read workflow file: %v", err)
	}
	if string(data) == "old content" {
		t.Error("workflow file should be overwritten with --force")
	}
	if string(data) != config.CrankWorkflowContent {
		t.Error("workflow file content does not match expected")
	}
}
