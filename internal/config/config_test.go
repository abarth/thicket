package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPaths(t *testing.T) {
	paths := GetPaths("/project")

	if paths.Root != "/project" {
		t.Errorf("Root = %q, want /project", paths.Root)
	}
	if paths.Dir != "/project/.thicket" {
		t.Errorf("Dir = %q, want /project/.thicket", paths.Dir)
	}
	if paths.Config != "/project/.thicket/config.json" {
		t.Errorf("Config = %q, want /project/.thicket/config.json", paths.Config)
	}
	if paths.Tickets != "/project/.thicket/tickets.jsonl" {
		t.Errorf("Tickets = %q, want /project/.thicket/tickets.jsonl", paths.Tickets)
	}
	if paths.Cache != "/project/.thicket/cache.db" {
		t.Errorf("Cache = %q, want /project/.thicket/cache.db", paths.Cache)
	}
}

func TestInit(t *testing.T) {
	dir := t.TempDir()

	if err := Init(dir, "TH"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Check directory was created
	thicketDir := filepath.Join(dir, ".thicket")
	if _, err := os.Stat(thicketDir); err != nil {
		t.Errorf("Init() did not create .thicket directory: %v", err)
	}

	// Check config was created
	configPath := filepath.Join(thicketDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	if string(data) != "{\n  \"project_code\": \"TH\"\n}" {
		t.Errorf("Config content = %q, want project_code TH", string(data))
	}

	// Check tickets file was created
	ticketsPath := filepath.Join(thicketDir, "tickets.jsonl")
	if _, err := os.Stat(ticketsPath); err != nil {
		t.Errorf("Init() did not create tickets.jsonl: %v", err)
	}

	// Check .gitignore was created
	gitignorePath := filepath.Join(thicketDir, ".gitignore")
	gitignoreData, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	if string(gitignoreData) != "cache.db\n" {
		t.Errorf(".gitignore content = %q, want 'cache.db\\n'", string(gitignoreData))
	}
}

func TestInit_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	if err := Init(dir, "TH"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Init(dir, "TH")
	if err != ErrAlreadyInit {
		t.Errorf("Init() error = %v, want ErrAlreadyInit", err)
	}
}

func TestInit_InvalidProjectCode(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "invalid")
	if err == nil {
		t.Error("Init() expected error for invalid project code")
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()

	if err := Init(dir, "TH"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ProjectCode != "TH" {
		t.Errorf("Load().ProjectCode = %q, want TH", cfg.ProjectCode)
	}
}

func TestLoad_NotInitialized(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(dir)
	if err != ErrNotInitialized {
		t.Errorf("Load() error = %v, want ErrNotInitialized", err)
	}
}

func TestFindRoot(t *testing.T) {
	// Create a nested directory structure
	dir := t.TempDir()

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}

	if err := Init(dir, "TH"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	nested := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Change to nested directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(nested); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	root, err := FindRoot()
	if err != nil {
		t.Fatalf("FindRoot() error = %v", err)
	}

	if root != dir {
		t.Errorf("FindRoot() = %q, want %q", root, dir)
	}
}

func TestFindRoot_NotInitialized(t *testing.T) {
	dir := t.TempDir()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	_, err := FindRoot()
	if err != ErrNotInitialized {
		t.Errorf("FindRoot() error = %v, want ErrNotInitialized", err)
	}
}

func TestWorkflowPaths(t *testing.T) {
	workflowPath, symlinkPath := WorkflowPaths("/project")

	expectedWorkflow := "/project/.agent/workflows/crank.md"
	expectedSymlink := "/project/.claude/commands/crank.md"

	if workflowPath != expectedWorkflow {
		t.Errorf("workflowPath = %q, want %q", workflowPath, expectedWorkflow)
	}
	if symlinkPath != expectedSymlink {
		t.Errorf("symlinkPath = %q, want %q", symlinkPath, expectedSymlink)
	}
}

func TestWorkflowFilesExist(t *testing.T) {
	dir := t.TempDir()

	// Initially no files exist
	existing := WorkflowFilesExist(dir)
	if len(existing) != 0 {
		t.Errorf("expected no existing files, got %v", existing)
	}

	// Create workflow file
	workflowPath, _ := WorkflowPaths(dir)
	if err := os.MkdirAll(filepath.Dir(workflowPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(workflowPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	existing = WorkflowFilesExist(dir)
	if len(existing) != 1 {
		t.Errorf("expected 1 existing file, got %v", existing)
	}
}

func TestSetupWorkflowFiles(t *testing.T) {
	dir := t.TempDir()

	if err := SetupWorkflowFiles(dir); err != nil {
		t.Fatalf("SetupWorkflowFiles() error = %v", err)
	}

	workflowPath, symlinkPath := WorkflowPaths(dir)

	// Check workflow file was created
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read workflow file: %v", err)
	}
	if string(data) != CrankWorkflowContent {
		t.Errorf("Workflow content mismatch")
	}

	// Check symlink was created
	linkTarget, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	expectedTarget := "../../.agent/workflows/crank.md"
	if linkTarget != expectedTarget {
		t.Errorf("Symlink target = %q, want %q", linkTarget, expectedTarget)
	}

	// Verify symlink works (can read through it)
	symlinkData, err := os.ReadFile(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read through symlink: %v", err)
	}
	if string(symlinkData) != CrankWorkflowContent {
		t.Errorf("Symlink content mismatch")
	}
}

func TestSetupWorkflowFiles_Overwrite(t *testing.T) {
	dir := t.TempDir()

	workflowPath, symlinkPath := WorkflowPaths(dir)

	// Create existing files
	if err := os.MkdirAll(filepath.Dir(workflowPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(workflowPath, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(symlinkPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(symlinkPath, []byte("old symlink"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// SetupWorkflowFiles should overwrite
	if err := SetupWorkflowFiles(dir); err != nil {
		t.Fatalf("SetupWorkflowFiles() error = %v", err)
	}

	// Check workflow file was updated
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read workflow file: %v", err)
	}
	if string(data) != CrankWorkflowContent {
		t.Errorf("Workflow content not updated")
	}

	// Check symlink was created (replacing regular file)
	linkTarget, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}
	if linkTarget != "../../.agent/workflows/crank.md" {
		t.Errorf("Symlink target = %q, want ../../.agent/workflows/crank.md", linkTarget)
	}
}

func TestTHICKET_DIR(t *testing.T) {
	// Reset global state after test
	oldOverride := dataDirOverride
	defer func() { dataDirOverride = oldOverride }()

	// Reset environment after test
	oldEnv := os.Getenv("THICKET_DIR")
	defer os.Setenv("THICKET_DIR", oldEnv)

	tempDir := t.TempDir()
	customDir := filepath.Join(tempDir, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatalf("failed to create custom dir: %v", err)
	}

	// 1. Test THICKET_DIR environment variable
	os.Setenv("THICKET_DIR", customDir)
	dataDirOverride = "" // Ensure flag is not set

	paths := GetPaths("/project")
	if paths.Dir != customDir {
		t.Errorf("expected Dir to be %q, got %q", customDir, paths.Dir)
	}

	// 2. Test flag takes precedence over THICKET_DIR
	flagDir := filepath.Join(tempDir, "flag")
	if err := os.MkdirAll(flagDir, 0755); err != nil {
		t.Fatalf("failed to create flag dir: %v", err)
	}
	SetDataDir(flagDir)

	paths = GetPaths("/project")
	if paths.Dir != flagDir {
		t.Errorf("expected Dir to be %q (flag), got %q", flagDir, paths.Dir)
	}

	// 3. Test FindRoot with THICKET_DIR
	dataDirOverride = ""
	os.Setenv("THICKET_DIR", customDir)

	root, err := FindRoot()
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	expectedRoot, _ := filepath.Abs(filepath.Dir(customDir))
	if root != expectedRoot {
		t.Errorf("expected Root to be %q, got %q", expectedRoot, root)
	}
}

