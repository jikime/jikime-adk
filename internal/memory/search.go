package memory

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

var ftsSpecialChars = regexp.MustCompile(`[^\p{L}\p{N}\s]`)

// Search performs FTS5 full-text search on memories.
// Falls back to LIKE-based search if FTS5 is not available.
func (s *SQLiteStore) Search(q SearchQuery) ([]SearchResult, error) {
	if q.Limit <= 0 {
		q.Limit = 10
	}

	if s.hasFTS5 {
		return s.searchFTS5(q)
	}
	return s.searchLike(q)
}

func (s *SQLiteStore) searchFTS5(q SearchQuery) ([]SearchResult, error) {
	ftsQuery := buildFTSQuery(q.Query)
	if ftsQuery == "" {
		return nil, nil
	}

	query := `SELECT m.id, m.session_id, m.project_dir, m.type, m.content, m.content_hash,
			COALESCE(m.metadata,''), m.created_at, COALESCE(m.accessed_at, m.created_at), m.access_count,
			rank
		FROM memories_fts fts
		JOIN memories m ON fts.id = m.id
		WHERE memories_fts MATCH ?`
	args := []interface{}{ftsQuery}

	if q.ProjectDir != "" {
		query += ` AND m.project_dir = ?`
		args = append(args, q.ProjectDir)
	}
	if q.Type != "" {
		query += ` AND m.type = ?`
		args = append(args, q.Type)
	}

	query += ` ORDER BY rank LIMIT ?`
	args = append(args, q.Limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		// FTS5 query syntax error — fall back to LIKE
		return s.searchLike(q)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var m Memory
		var rank float64
		var createdStr, accessedStr string
		if err := rows.Scan(&m.ID, &m.SessionID, &m.ProjectDir, &m.Type, &m.Content, &m.ContentHash,
			&m.Metadata, &createdStr, &accessedStr, &m.AccessCount, &rank); err != nil {
			continue
		}
		m.CreatedAt = parseTime(createdStr)
		m.AccessedAt = parseTime(accessedStr)
		score := bm25RankToScore(rank)
		if q.MinScore > 0 && score < q.MinScore {
			continue
		}
		results = append(results, SearchResult{Memory: m, Score: score})
	}

	// Update access metadata for returned results
	s.updateAccessMetadata(results)

	return results, nil
}

func (s *SQLiteStore) searchLike(q SearchQuery) ([]SearchResult, error) {
	query := `SELECT id, session_id, project_dir, type, content, content_hash,
			COALESCE(metadata,''), created_at, COALESCE(accessed_at, created_at), access_count
		FROM memories WHERE content LIKE ?`
	args := []interface{}{"%" + q.Query + "%"}

	if q.ProjectDir != "" {
		query += ` AND project_dir = ?`
		args = append(args, q.ProjectDir)
	}
	if q.Type != "" {
		query += ` AND type = ?`
		args = append(args, q.Type)
	}

	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, q.Limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var m Memory
		var createdStr, accessedStr string
		if err := rows.Scan(&m.ID, &m.SessionID, &m.ProjectDir, &m.Type, &m.Content, &m.ContentHash,
			&m.Metadata, &createdStr, &accessedStr, &m.AccessCount); err != nil {
			continue
		}
		m.CreatedAt = parseTime(createdStr)
		m.AccessedAt = parseTime(accessedStr)
		results = append(results, SearchResult{Memory: m, Score: 0.5})
	}

	s.updateAccessMetadata(results)
	return results, nil
}

// SearchRecent returns the N most recent memories for a project.
func (s *SQLiteStore) SearchRecent(projectDir string, limit int) ([]Memory, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(
		`SELECT id, session_id, project_dir, type, content, content_hash,
			COALESCE(metadata,''), created_at, COALESCE(accessed_at, created_at), access_count
		FROM memories WHERE project_dir = ?
		ORDER BY created_at DESC LIMIT ?`,
		projectDir, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemories(rows)
}

// GetBySession returns all memories for a given session ID.
func (s *SQLiteStore) GetBySession(sessionID string) ([]Memory, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, project_dir, type, content, content_hash,
			COALESCE(metadata,''), created_at, COALESCE(accessed_at, created_at), access_count
		FROM memories WHERE session_id = ?
		ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemories(rows)
}

// buildFTSQuery converts a raw query string into an FTS5 MATCH expression.
// Extracts alphanumeric tokens, quotes them, and joins with AND.
func buildFTSQuery(raw string) string {
	// Remove special FTS5 characters
	cleaned := ftsSpecialChars.ReplaceAllString(raw, " ")
	tokens := strings.Fields(cleaned)
	if len(tokens) == 0 {
		return ""
	}

	var quoted []string
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		// Escape any remaining double quotes
		t = strings.ReplaceAll(t, `"`, ``)
		if t != "" {
			quoted = append(quoted, fmt.Sprintf(`"%s"`, t))
		}
	}
	if len(quoted) == 0 {
		return ""
	}
	return strings.Join(quoted, " AND ")
}

// bm25RankToScore converts FTS5 BM25 rank to a 0-1 score.
// FTS5 rank values are negative (closer to 0 = better match).
func bm25RankToScore(rank float64) float64 {
	if math.IsNaN(rank) || math.IsInf(rank, 0) {
		return 0
	}
	return 1.0 / (1.0 + math.Abs(rank))
}

func (s *SQLiteStore) updateAccessMetadata(results []SearchResult) {
	nowStr := time.Now().Format(timeFormat)
	for _, r := range results {
		_, _ = s.db.Exec(
			`UPDATE memories SET accessed_at = ?, access_count = access_count + 1 WHERE id = ?`,
			nowStr, r.Memory.ID,
		)
	}
}

// GarbageCollect removes old or excess memories.
func (s *SQLiteStore) GarbageCollect(projectDir string, opts GCOptions) (*GCResult, error) {
	if opts.MaxAge == 0 {
		opts.MaxAge = 90 * 24 * time.Hour // 90 days
	}
	if opts.MaxCount == 0 {
		opts.MaxCount = 1000
	}

	result := &GCResult{}

	cutoffStr := time.Now().Add(-opts.MaxAge).Format(timeFormat)

	// Count by age
	var ageCount int
	s.db.QueryRow(
		`SELECT COUNT(*) FROM memories WHERE project_dir = ? AND created_at < ?`,
		projectDir, cutoffStr,
	).Scan(&ageCount)
	result.DeletedByAge = ageCount

	if !opts.DryRun && ageCount > 0 {
		// Get IDs to delete from FTS
		if s.hasFTS5 {
			rows, _ := s.db.Query(
				`SELECT id FROM memories WHERE project_dir = ? AND created_at < ?`,
				projectDir, cutoffStr,
			)
			if rows != nil {
				for rows.Next() {
					var id string
					rows.Scan(&id)
					s.db.Exec(`DELETE FROM memories_fts WHERE id = ?`, id)
				}
				rows.Close()
			}
		}
		s.db.Exec(
			`DELETE FROM memories WHERE project_dir = ? AND created_at < ?`,
			projectDir, cutoffStr,
		)
	}

	// Count excess beyond MaxCount
	var totalCount int
	s.db.QueryRow(`SELECT COUNT(*) FROM memories WHERE project_dir = ?`, projectDir).Scan(&totalCount)

	excess := totalCount - opts.MaxCount
	if excess > 0 {
		result.DeletedByCount = excess
		if !opts.DryRun {
			if s.hasFTS5 {
				rows, _ := s.db.Query(
					`SELECT id FROM memories WHERE project_dir = ?
					ORDER BY created_at ASC LIMIT ?`,
					projectDir, excess,
				)
				if rows != nil {
					for rows.Next() {
						var id string
						rows.Scan(&id)
						s.db.Exec(`DELETE FROM memories_fts WHERE id = ?`, id)
					}
					rows.Close()
				}
			}
			s.db.Exec(
				`DELETE FROM memories WHERE id IN (
					SELECT id FROM memories WHERE project_dir = ?
					ORDER BY created_at ASC LIMIT ?
				)`, projectDir, excess,
			)
		}
	}

	// Count remaining
	s.db.QueryRow(`SELECT COUNT(*) FROM memories WHERE project_dir = ?`, projectDir).Scan(&result.Remaining)

	// Also clean old sessions (keep last 100)
	if !opts.DryRun {
		s.db.Exec(
			`DELETE FROM session_history WHERE session_id IN (
				SELECT session_id FROM session_history WHERE project_dir = ?
				ORDER BY ended_at DESC LIMIT -1 OFFSET 100
			)`, projectDir,
		)
	}

	return result, nil
}

// DeleteBySessionID removes all memories associated with a session.
func (s *SQLiteStore) DeleteBySessionID(sessionID string) error {
	if s.hasFTS5 {
		rows, _ := s.db.Query(`SELECT id FROM memories WHERE session_id = ?`, sessionID)
		if rows != nil {
			for rows.Next() {
				var id string
				rows.Scan(&id)
				s.db.Exec(`DELETE FROM memories_fts WHERE id = ?`, id)
			}
			rows.Close()
		}
	}
	_, err := s.db.Exec(`DELETE FROM memories WHERE session_id = ?`, sessionID)
	return err
}

