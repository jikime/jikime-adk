package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Indexer handles indexing MD files into SQLite chunks.
type Indexer struct {
	store    *SQLiteStore
	provider EmbeddingProvider // may be nil
}

// NewIndexer creates a new indexer.
func NewIndexer(store *SQLiteStore, provider EmbeddingProvider) *Indexer {
	return &Indexer{store: store, provider: provider}
}

// IndexFile indexes a single MD file. Only re-indexes if the file has changed
// (based on modification time comparison with file_index table).
func (idx *Indexer) IndexFile(ctx context.Context, projectDir, relPath string) error {
	absPath := filepath.Join(projectDir, relPath)
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", relPath, err)
	}

	// Check if file needs re-indexing
	existing, _ := idx.store.GetFileIndex(relPath)
	if existing != nil && existing.LastModified >= info.ModTime().Unix() {
		return nil // file hasn't changed
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", relPath, err)
	}

	chunks := ChunkFile(relPath, content, DefaultChunkOpts())

	// Generate embeddings if provider available
	if idx.provider != nil && len(chunks) > 0 {
		texts := make([]string, len(chunks))
		for i, c := range chunks {
			texts[i] = c.Text
		}
		embeddings, err := idx.provider.EmbedBatch(ctx, texts)
		if err == nil && len(embeddings) == len(chunks) {
			for i := range chunks {
				chunks[i].Embedding = embeddings[i]
			}
		}
		// Non-fatal: continue without embeddings if batch fails
	}

	// Delete old chunks for this path and insert new ones
	if err := idx.store.DeleteChunksByPath(relPath); err != nil {
		return fmt.Errorf("delete old chunks for %s: %w", relPath, err)
	}

	if len(chunks) > 0 {
		if err := idx.store.SaveChunks(chunks); err != nil {
			return fmt.Errorf("save chunks for %s: %w", relPath, err)
		}
	}

	// Update file index
	return idx.store.UpdateFileIndex(FileIndexEntry{
		Path:         relPath,
		LastModified: info.ModTime().Unix(),
		ChunkCount:   len(chunks),
		LastIndexed:  time.Now().Unix(),
	})
}

// IndexAll indexes all MD files under the .jikime/memory/ directory.
func (idx *Indexer) IndexAll(ctx context.Context, projectDir string) error {
	memDir := filepath.Join(projectDir, ".jikime", "memory")

	entries, err := os.ReadDir(memDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no memory directory yet
		}
		return fmt.Errorf("read memory directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		relPath := filepath.Join(".jikime", "memory", entry.Name())
		if err := idx.IndexFile(ctx, projectDir, relPath); err != nil {
			// Log but continue indexing other files
			fmt.Fprintf(os.Stderr, "warning: index %s: %v\n", relPath, err)
		}
	}

	return nil
}
