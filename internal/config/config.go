// Package config handles Thicket project configuration.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/abarth/thicket/internal/ticket"
)

var (
	dataDirOverride string
)

// SetDataDir sets a custom directory for Thicket data.
// If set, it overrides the default search for .thicket and the THICKET_DIR environment variable.
func SetDataDir(dir string) {
	dataDirOverride = dir
}

func getDataDir() string {
	if dataDirOverride != "" {
		return dataDirOverride
	}
	return os.Getenv("THICKET_DIR")
}

const (
	ThicketDir  = ".thicket"
	ConfigFile  = "config.json"
	TicketsFile = "tickets.jsonl"
	CacheFile   = "cache.db"
)

var (
	ErrNotInitialized  = errors.New("thicket not initialized in this directory (run 'thicket init')")
	ErrAlreadyInit     = errors.New("thicket already initialized in this directory")
	ErrNoProjectCode   = errors.New("project code is required")
)

// Config represents the Thicket project configuration.
type Config struct {
	ProjectCode string `json:"project_code"`
}

// Paths holds the resolved paths for Thicket files.
type Paths struct {
	Root    string // The directory containing .thicket
	Dir     string // The .thicket directory
	Config  string // config.json path
	Tickets string // tickets.jsonl path
	Cache   string // cache.db path
}

// FindRoot locates the Thicket root directory by searching upward from the current directory.
func FindRoot() (string, error) {
	dataDir := getDataDir()
	if dataDir != "" {
		abs, err := filepath.Abs(dataDir)
		if err != nil {
			return "", fmt.Errorf("getting absolute path for data dir: %w", err)
		}
		// Return the parent directory so that GetPaths(root) works correctly
		return filepath.Dir(abs), nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	for {
		thicketDir := filepath.Join(dir, ThicketDir)
		if info, err := os.Stat(thicketDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotInitialized
		}
		dir = parent
	}
}

// GetPaths returns the paths for all Thicket files relative to the given root.
func GetPaths(root string) Paths {
	var dir string
	dataDir := getDataDir()
	if dataDir != "" {
		abs, _ := filepath.Abs(dataDir)
		dir = abs
	} else {
		dir = filepath.Join(root, ThicketDir)
	}

	return Paths{
		Root:    root,
		Dir:     dir,
		Config:  filepath.Join(dir, ConfigFile),
		Tickets: filepath.Join(dir, TicketsFile),
		Cache:   filepath.Join(dir, CacheFile),
	}
}

// Load reads the configuration from the given root directory.
func Load(root string) (*Config, error) {
	paths := GetPaths(root)

	data, err := os.ReadFile(paths.Config)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotInitialized
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Init initializes a new Thicket project in the given directory.
func Init(root, projectCode string) error {
	if err := ticket.ValidateProjectCode(projectCode); err != nil {
		return err
	}

	paths := GetPaths(root)

	// Check if already initialized
	if _, err := os.Stat(paths.Dir); err == nil {
		return ErrAlreadyInit
	}

	// Create .thicket directory
	if err := os.MkdirAll(paths.Dir, 0755); err != nil {
		return fmt.Errorf("creating thicket directory: %w", err)
	}

	// Write config
	cfg := Config{ProjectCode: projectCode}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := os.WriteFile(paths.Config, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	// Create empty tickets file
	if err := os.WriteFile(paths.Tickets, []byte{}, 0644); err != nil {
		return fmt.Errorf("creating tickets file: %w", err)
	}

	// Create .gitignore for cache
	gitignore := filepath.Join(paths.Dir, ".gitignore")
	if err := os.WriteFile(gitignore, []byte("cache.db\n"), 0644); err != nil {
		return fmt.Errorf("creating .gitignore: %w", err)
	}

	return nil
}
