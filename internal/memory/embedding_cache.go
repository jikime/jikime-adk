package memory

import (
	"context"
	"fmt"
)

// GetCachedEmbedding checks cache by content_hash + provider + model.
func (s *SQLiteStore) GetCachedEmbedding(contentHash, provider, model string) ([]float32, error) {
	var blob []byte
	err := s.db.QueryRow(
		`SELECT embedding FROM embedding_cache WHERE content_hash = ? AND provider = ? AND model = ?`,
		contentHash, provider, model,
	).Scan(&blob)
	if err != nil {
		return nil, err
	}
	return DecodeEmbedding(blob), nil
}

// SaveCachedEmbedding stores embedding in cache.
func (s *SQLiteStore) SaveCachedEmbedding(contentHash, provider, model string, embedding []float32, dims int) error {
	blob := EncodeEmbedding(embedding)
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO embedding_cache (content_hash, provider, model, embedding, dims) VALUES (?, ?, ?, ?, ?)`,
		contentHash, provider, model, blob, dims,
	)
	return err
}

// EmbedAndCache embeds text, checks cache first, saves to cache if new.
func (s *SQLiteStore) EmbedAndCache(ctx context.Context, provider EmbeddingProvider, content string) ([]float32, error) {
	if provider == nil {
		return nil, nil
	}

	hash := ContentHash(content)

	// Check cache
	cached, err := s.GetCachedEmbedding(hash, provider.ID(), provider.Model())
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Generate embedding
	embedding, err := provider.EmbedQuery(ctx, content)
	if err != nil {
		return nil, err
	}

	// Save to cache (best-effort)
	_ = s.SaveCachedEmbedding(hash, provider.ID(), provider.Model(), embedding, provider.Dims())

	return embedding, nil
}

// unembeddedMemory holds the minimal info needed to backfill embeddings.
type unembeddedMemory struct {
	ID      string
	Content string
}

// GetUnembeddedMemories returns memories that have no embedding (NULL or zero-length BLOB).
// If sessionID is non-empty, only returns memories from that session.
func (s *SQLiteStore) GetUnembeddedMemories(projectDir, sessionID string, limit int) ([]unembeddedMemory, error) {
	var query string
	var args []interface{}

	if sessionID != "" {
		query = `SELECT id, content FROM memories
			WHERE project_dir = ? AND session_id = ? AND (embedding IS NULL OR LENGTH(embedding) = 0)
			ORDER BY created_at DESC LIMIT ?`
		args = []interface{}{projectDir, sessionID, limit}
	} else {
		query = `SELECT id, content FROM memories
			WHERE project_dir = ? AND (embedding IS NULL OR LENGTH(embedding) = 0)
			ORDER BY created_at DESC LIMIT ?`
		args = []interface{}{projectDir, limit}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []unembeddedMemory
	for rows.Next() {
		var m unembeddedMemory
		if err := rows.Scan(&m.ID, &m.Content); err != nil {
			continue
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// UpdateMemoryEmbedding sets the embedding BLOB for a memory by ID.
func (s *SQLiteStore) UpdateMemoryEmbedding(id string, embedding []float32) error {
	blob := EncodeEmbedding(embedding)
	_, err := s.db.Exec(`UPDATE memories SET embedding = ? WHERE id = ?`, blob, id)
	return err
}

// BackfillMemoryEmbeddings generates embeddings for memories that don't have one yet.
// If sessionID is non-empty, only processes memories from that session.
// Uses batch API to minimize network round-trips.
func (s *SQLiteStore) BackfillMemoryEmbeddings(ctx context.Context, provider EmbeddingProvider, projectDir, sessionID string) (int, error) {
	if provider == nil {
		return 0, nil
	}

	memories, err := s.GetUnembeddedMemories(projectDir, sessionID, 100)
	if err != nil || len(memories) == 0 {
		return 0, err
	}

	// Collect texts for batch embedding
	texts := make([]string, len(memories))
	for i, m := range memories {
		texts[i] = m.Content
	}

	embeddings, err := provider.EmbedBatch(ctx, texts)
	if err != nil {
		return 0, fmt.Errorf("batch embed: %w", err)
	}

	// Update each memory with its embedding
	count := 0
	for i, emb := range embeddings {
		if i >= len(memories) {
			break
		}
		if err := s.UpdateMemoryEmbedding(memories[i].ID, emb); err != nil {
			continue
		}
		// Also save to cache for future lookups
		hash := ContentHash(memories[i].Content)
		_ = s.SaveCachedEmbedding(hash, provider.ID(), provider.Model(), emb, provider.Dims())
		count++
	}

	return count, nil
}

// CleanEmbeddingCache removes cache entries older than maxAgeDays.
func (s *SQLiteStore) CleanEmbeddingCache(maxAgeDays int) (int, error) {
	result, err := s.db.Exec(
		`DELETE FROM embedding_cache WHERE created_at < datetime('now', ? || ' days')`,
		fmt.Sprintf("-%d", maxAgeDays),
	)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}
