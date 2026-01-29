package tui

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileChangedMsg is sent when the ticket database file changes.
type FileChangedMsg struct{}

// WatchFile returns a command that watches the given file for changes.
// When the file changes, it sends a FileChangedMsg.
// The debounce parameter controls how long to wait after a change before
// sending the message (to avoid rapid repeated updates).
func WatchFile(path string, debounce time.Duration) func() (chan FileChangedMsg, func()) {
	return func() (chan FileChangedMsg, func()) {
		ch := make(chan FileChangedMsg)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			close(ch)
			return ch, func() {}
		}

		if err := watcher.Add(path); err != nil {
			watcher.Close()
			close(ch)
			return ch, func() {}
		}

		done := make(chan struct{})
		cleanup := func() {
			close(done)
			watcher.Close()
		}

		go func() {
			defer close(ch)
			var timer *time.Timer

			for {
				select {
				case <-done:
					if timer != nil {
						timer.Stop()
					}
					return

				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					// We only care about write events
					if event.Op&fsnotify.Write == fsnotify.Write {
						// Debounce: reset timer on each event
						if timer != nil {
							timer.Stop()
						}
						timer = time.AfterFunc(debounce, func() {
							select {
							case ch <- FileChangedMsg{}:
							case <-done:
							}
						})
					}

				case _, ok := <-watcher.Errors:
					if !ok {
						return
					}
					// Ignore errors, continue watching
				}
			}
		}()

		return ch, cleanup
	}
}
