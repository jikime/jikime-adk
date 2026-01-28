package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceDelay = 500 * time.Millisecond

// WatchMemoryFiles watches .jikime/memory/ for MD file changes and
// auto-indexes them. Blocks until ctx is cancelled. Intended to run
// as a goroutine.
func WatchMemoryFiles(ctx context.Context, projectDir string, store *SQLiteStore, provider EmbeddingProvider) {
	memDir := filepath.Join(projectDir, ".jikime", "memory")

	// Ensure dir exists
	if _, err := os.Stat(memDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(memDir, 0o755); mkErr != nil {
			fmt.Fprintf(os.Stderr, "[watcher] cannot create memory dir: %v\n", mkErr)
			return
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[watcher] init error: %v\n", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(memDir); err != nil {
		fmt.Fprintf(os.Stderr, "[watcher] watch %s: %v\n", memDir, err)
		return
	}

	fmt.Fprintf(os.Stderr, "[watcher] watching %s\n", memDir)

	var mu sync.Mutex
	timers := make(map[string]*time.Timer)
	indexer := NewIndexer(store, provider)

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			for _, t := range timers {
				t.Stop()
			}
			mu.Unlock()
			fmt.Fprintf(os.Stderr, "[watcher] stopped\n")
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if !isRelevantEvent(event) {
				continue
			}

			absPath := event.Name
			relPath, err := filepath.Rel(projectDir, absPath)
			if err != nil {
				continue
			}

			mu.Lock()
			if t, exists := timers[absPath]; exists {
				t.Stop()
			}
			timers[absPath] = time.AfterFunc(debounceDelay, func() {
				idxCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()

				if err := indexer.IndexFile(idxCtx, projectDir, relPath); err != nil {
					fmt.Fprintf(os.Stderr, "[watcher] index %s: %v\n", relPath, err)
				} else {
					fmt.Fprintf(os.Stderr, "[watcher] indexed %s\n", relPath)
				}

				mu.Lock()
				delete(timers, absPath)
				mu.Unlock()
			})
			mu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "[watcher] error: %v\n", err)
		}
	}
}

// isRelevantEvent returns true for Create/Write events on .md files.
func isRelevantEvent(event fsnotify.Event) bool {
	if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
		return false
	}
	return strings.HasSuffix(strings.ToLower(event.Name), ".md")
}
