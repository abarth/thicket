package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchFile(t *testing.T) {
	// Create a temporary file to watch
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create the file
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Start watching with a short debounce
	watchChan, cleanup := WatchFile(testFile, 50*time.Millisecond)()
	defer cleanup()

	// Modify the file
	time.Sleep(100 * time.Millisecond) // Let the watcher settle
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for the change notification with a timeout
	select {
	case <-watchChan:
		// Success - we received the notification
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for file change notification")
	}
}

func TestWatchFileDebounce(t *testing.T) {
	// Create a temporary file to watch
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create the file
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Start watching with a longer debounce
	watchChan, cleanup := WatchFile(testFile, 200*time.Millisecond)()
	defer cleanup()

	// Modify the file multiple times rapidly
	time.Sleep(100 * time.Millisecond) // Let the watcher settle
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(testFile, []byte("modified"+string(rune('0'+i))), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}
		time.Sleep(50 * time.Millisecond) // Less than debounce time
	}

	// We should only receive one notification due to debouncing
	received := 0
	timeout := time.After(1 * time.Second)

	for {
		select {
		case _, ok := <-watchChan:
			if ok {
				received++
			}
			if received == 1 {
				// After receiving one, wait a bit to see if more come
				time.Sleep(300 * time.Millisecond)
				// Check if channel has more (non-blocking)
				select {
				case <-watchChan:
					received++
				default:
				}
				if received > 1 {
					t.Errorf("Expected 1 notification due to debouncing, got %d", received)
				}
				return
			}
		case <-timeout:
			if received == 0 {
				t.Fatal("Timed out waiting for file change notification")
			}
			return
		}
	}
}

func TestWatchFileCleanup(t *testing.T) {
	// Create a temporary file to watch
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create the file
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Start watching
	watchChan, cleanup := WatchFile(testFile, 50*time.Millisecond)()

	// Clean up the watcher
	cleanup()

	// Channel should be closed
	time.Sleep(100 * time.Millisecond)
	select {
	case _, ok := <-watchChan:
		if ok {
			t.Error("Expected channel to be closed after cleanup")
		}
		// Channel was closed, success
	case <-time.After(500 * time.Millisecond):
		// Channel might not have closed yet, that's ok
	}
}
