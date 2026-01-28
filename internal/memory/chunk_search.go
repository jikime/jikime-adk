package memory

import (
	"context"
	"sort"
)

// SearchChunks performs hybrid vector + text search on the chunks table.
// If provider is nil, falls back to FTS5/LIKE text search only.
func (s *SQLiteStore) SearchChunks(ctx context.Context, q SearchQuery, provider EmbeddingProvider) ([]ChunkSearchResult, error) {
	if provider == nil {
		return s.searchChunksTextOnly(q)
	}

	// Generate query embedding
	queryEmb, err := s.EmbedAndCache(ctx, provider, q.Query)
	if err != nil || len(queryEmb) == 0 {
		return s.searchChunksTextOnly(q)
	}

	// Vector search on chunks
	vectorHits, err := s.searchChunksVector(q, queryEmb)
	if err != nil {
		return s.searchChunksTextOnly(q)
	}

	// Text search on chunks
	textHits, _ := s.searchChunksFTS(q)

	// Merge
	return mergeChunkResults(vectorHits, textHits, q), nil
}

// searchChunksTextOnly wraps FTS/LIKE results as the text-only fallback.
func (s *SQLiteStore) searchChunksTextOnly(q SearchQuery) ([]ChunkSearchResult, error) {
	return s.searchChunksFTS(q)
}

// searchChunksVector does brute-force cosine similarity on chunk embeddings.
func (s *SQLiteStore) searchChunksVector(q SearchQuery, queryVec []float32) (map[int64]chunkVectorHit, error) {
	rows, err := s.db.Query(
		`SELECT id, path, start_line, end_line, text, hash, heading, embedding
		FROM chunks WHERE embedding IS NOT NULL LIMIT 1000`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hits := make(map[int64]chunkVectorHit)
	for rows.Next() {
		var c Chunk
		var blob []byte
		if err := rows.Scan(&c.ID, &c.Path, &c.StartLine, &c.EndLine,
			&c.Text, &c.Hash, &c.Heading, &blob); err != nil {
			continue
		}
		emb := DecodeEmbedding(blob)
		if len(emb) == 0 {
			continue
		}
		score := CosineSimilarity(queryVec, emb)
		if score > 0 {
			hits[c.ID] = chunkVectorHit{chunk: c, score: score}
		}
	}
	return hits, nil
}

type chunkVectorHit struct {
	chunk Chunk
	score float64
}

// searchChunksFTS searches chunks using FTS5 or LIKE fallback.
func (s *SQLiteStore) searchChunksFTS(q SearchQuery) ([]ChunkSearchResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultMaxResults
	}

	if s.hasChunksFTS5 {
		ftsQuery := buildFTSQuery(q.Query)
		if ftsQuery == "" {
			return nil, nil
		}

		rows, err := s.db.Query(
			`SELECT c.id, c.path, c.start_line, c.end_line, c.text, c.hash, c.heading, rank
			FROM chunks_fts fts
			JOIN chunks c ON fts.rowid = c.id
			WHERE chunks_fts MATCH ?
			ORDER BY rank LIMIT ?`,
			ftsQuery, limit)
		if err != nil {
			// FTS error, fall back to LIKE
			return s.searchChunksLike(q)
		}
		defer rows.Close()

		var results []ChunkSearchResult
		for rows.Next() {
			var c Chunk
			var rank float64
			if err := rows.Scan(&c.ID, &c.Path, &c.StartLine, &c.EndLine,
				&c.Text, &c.Hash, &c.Heading, &rank); err != nil {
				continue
			}
			score := bm25RankToScore(rank)
			if q.MinScore > 0 && score < q.MinScore {
				continue
			}
			results = append(results, ChunkSearchResult{
				Chunk:     c,
				TextScore: score,
				Score:     score,
			})
		}
		return results, nil
	}

	return s.searchChunksLike(q)
}

// searchChunksLike searches chunks using LIKE (fallback when FTS5 not available).
func (s *SQLiteStore) searchChunksLike(q SearchQuery) ([]ChunkSearchResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultMaxResults
	}

	rows, err := s.db.Query(
		`SELECT id, path, start_line, end_line, text, hash, heading
		FROM chunks WHERE text LIKE ? LIMIT ?`,
		"%"+q.Query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ChunkSearchResult
	for rows.Next() {
		var c Chunk
		if err := rows.Scan(&c.ID, &c.Path, &c.StartLine, &c.EndLine,
			&c.Text, &c.Hash, &c.Heading); err != nil {
			continue
		}
		results = append(results, ChunkSearchResult{
			Chunk:     c,
			TextScore: 0.5,
			Score:     0.5,
		})
	}
	return results, nil
}

// mergeChunkResults merges vector and text search results with weighted scoring.
func mergeChunkResults(vectorHits map[int64]chunkVectorHit, textResults []ChunkSearchResult, q SearchQuery) []ChunkSearchResult {
	merged := make(map[int64]*ChunkSearchResult)

	// Add vector results
	for id, vh := range vectorHits {
		merged[id] = &ChunkSearchResult{
			Chunk:       vh.chunk,
			VectorScore: vh.score,
		}
	}

	// Merge text results
	for _, tr := range textResults {
		if existing, ok := merged[tr.Chunk.ID]; ok {
			existing.TextScore = tr.TextScore
		} else {
			merged[tr.Chunk.ID] = &ChunkSearchResult{
				Chunk:     tr.Chunk,
				TextScore: tr.TextScore,
			}
		}
	}

	// Calculate final scores
	minScore := q.MinScore
	if minScore == 0 {
		minScore = DefaultMinScore
	}
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultMaxResults
	}

	var results []ChunkSearchResult
	for _, r := range merged {
		r.Score = DefaultVectorWeight*r.VectorScore + DefaultTextWeight*r.TextScore
		if r.Score >= minScore {
			results = append(results, *r)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}
