package memory

import (
	"context"
	"math"
	"sort"
)

// CosineSimilarity computes cosine similarity between two vectors.
// Returns value in [-1, 1], where 1 = identical.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// SearchHybrid performs hybrid vector + text search.
// If provider is nil, falls back to pure FTS5 search.
func (s *SQLiteStore) SearchHybrid(ctx context.Context, q SearchQuery, provider EmbeddingProvider) ([]HybridSearchResult, error) {
	if provider == nil {
		return s.searchTextOnly(q)
	}

	// Generate query embedding
	queryEmbedding, err := s.EmbedAndCache(ctx, provider, q.Query)
	if err != nil || len(queryEmbedding) == 0 {
		// Embedding failed — fallback to text-only
		return s.searchTextOnly(q)
	}

	// Vector search: load all embeddings for project and compute similarity
	vectorResults, err := s.searchVector(q, queryEmbedding)
	if err != nil {
		return s.searchTextOnly(q)
	}

	// Text search: existing FTS5
	textResults, _ := s.Search(q)

	// Merge: 0.7 * vectorScore + 0.3 * textScore
	return mergeHybridResults(vectorResults, textResults, q), nil
}

// searchVector performs brute-force cosine similarity search over stored embeddings.
func (s *SQLiteStore) searchVector(q SearchQuery, queryVec []float32) (map[string]vectorHit, error) {
	query := `SELECT id, session_id, project_dir, type, content, content_hash,
			COALESCE(metadata,''), COALESCE(created_at,''), COALESCE(accessed_at,''), access_count, embedding
		FROM memories WHERE project_dir = ? AND embedding IS NOT NULL`
	args := []interface{}{q.ProjectDir}

	if q.Type != "" {
		query += ` AND type = ?`
		args = append(args, q.Type)
	}
	query += ` LIMIT 1000`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hits := make(map[string]vectorHit)
	for rows.Next() {
		var m Memory
		var blob []byte
		var createdStr, accessedStr string
		if err := rows.Scan(&m.ID, &m.SessionID, &m.ProjectDir, &m.Type, &m.Content, &m.ContentHash,
			&m.Metadata, &createdStr, &accessedStr, &m.AccessCount, &blob); err != nil {
			continue
		}
		m.CreatedAt = parseTime(createdStr)
		m.AccessedAt = parseTime(accessedStr)

		emb := DecodeEmbedding(blob)
		if len(emb) == 0 {
			continue
		}
		score := CosineSimilarity(queryVec, emb)
		if score > 0 {
			hits[m.ID] = vectorHit{memory: m, score: score}
		}
	}
	return hits, nil
}

type vectorHit struct {
	memory Memory
	score  float64
}

// mergeHybridResults combines vector and text search results with weighted scoring.
func mergeHybridResults(vectorHits map[string]vectorHit, textResults []SearchResult, q SearchQuery) []HybridSearchResult {
	merged := make(map[string]*HybridSearchResult)

	// Add vector results
	for id, vh := range vectorHits {
		merged[id] = &HybridSearchResult{
			Memory:      vh.memory,
			VectorScore: vh.score,
		}
	}

	// Merge text results
	for _, tr := range textResults {
		if existing, ok := merged[tr.Memory.ID]; ok {
			existing.TextScore = tr.Score
		} else {
			merged[tr.Memory.ID] = &HybridSearchResult{
				Memory:    tr.Memory,
				TextScore: tr.Score,
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

	var results []HybridSearchResult
	for _, r := range merged {
		r.Score = DefaultVectorWeight*r.VectorScore + DefaultTextWeight*r.TextScore
		if r.Score >= minScore {
			results = append(results, *r)
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

// searchTextOnly wraps FTS5 results as HybridSearchResult (fallback).
func (s *SQLiteStore) searchTextOnly(q SearchQuery) ([]HybridSearchResult, error) {
	results, err := s.Search(q)
	if err != nil {
		return nil, err
	}
	hybrid := make([]HybridSearchResult, len(results))
	for i, r := range results {
		hybrid[i] = HybridSearchResult{
			Memory:    r.Memory,
			Score:     r.Score,
			TextScore: r.Score,
		}
	}
	return hybrid, nil
}
