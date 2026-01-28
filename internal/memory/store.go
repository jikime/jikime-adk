package memory

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const timeFormat = "2006-01-02T15:04:05Z"

// SQLiteStore implements the memory store using SQLite.
type SQLiteStore struct {
	db            *sql.DB
	dbPath        string
	hasFTS5       bool
	hasChunksFTS5 bool
}

// NewStore creates or opens the SQLite memory store for a project.
// DB is stored at <projectDir>/.jikime/memory/memory.db
func NewStore(projectDir string) (*SQLiteStore, error) {
	memDir := filepath.Join(projectDir, ".jikime", "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		return nil, fmt.Errorf("create memory dir: %w", err)
	}
	dbPath := filepath.Join(memDir, "memory.db")
	return NewStoreWithPath(dbPath)
}

// NewStoreWithPath creates a store at a specific DB path (for testing).
func NewStoreWithPath(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Set pragmas for performance and concurrent access
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-8000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("pragma %q: %w", p, err)
		}
	}

	if err := EnsureSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &SQLiteStore{
		db:            db,
		dbPath:        dbPath,
		hasFTS5:       HasFTS5(db),
		hasChunksFTS5: HasChunksFTS5(db),
	}, nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// SaveMemory inserts a memory entry. Also inserts into FTS5 if available.
func (s *SQLiteStore) SaveMemory(m Memory) error {
	if m.ID == "" {
		m.ID = generateID()
	}
	if m.ContentHash == "" {
		m.ContentHash = ContentHash(m.Content)
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	accessedStr := ""
	if !m.AccessedAt.IsZero() {
		accessedStr = m.AccessedAt.Format(timeFormat)
	}

	var embeddingBlob []byte
	if len(m.Embedding) > 0 {
		embeddingBlob = EncodeEmbedding(m.Embedding)
	}

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO memories
			(id, session_id, project_dir, type, content, content_hash, metadata, created_at, accessed_at, access_count, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.SessionID, m.ProjectDir, m.Type, m.Content, m.ContentHash,
		m.Metadata, m.CreatedAt.Format(timeFormat), accessedStr, m.AccessCount, embeddingBlob,
	)
	if err != nil {
		return err
	}

	if s.hasFTS5 {
		_, _ = s.db.Exec(
			`INSERT INTO memories_fts (content, id, project_dir, type) VALUES (?, ?, ?, ?)`,
			m.Content, m.ID, m.ProjectDir, m.Type,
		)
	}

	return nil
}

// SaveIfNew saves only if content_hash is not already in the DB.
// Returns true if saved, false if duplicate.
func (s *SQLiteStore) SaveIfNew(m Memory) (bool, error) {
	if m.ContentHash == "" {
		m.ContentHash = ContentHash(m.Content)
	}

	var existing string
	err := s.db.QueryRow(
		`SELECT id FROM memories WHERE content_hash = ? AND project_dir = ?`,
		m.ContentHash, m.ProjectDir,
	).Scan(&existing)
	if err == nil {
		return false, nil // duplicate
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	return true, s.SaveMemory(m)
}

// GetMemory retrieves a single memory by ID or ID prefix.
// Exact match is tried first. If no exact match, prefix match (LIKE) is used.
func (s *SQLiteStore) GetMemory(id string) (*Memory, error) {
	m := &Memory{}
	var createdStr, accessedStr string

	// Try exact match first
	err := s.db.QueryRow(
		`SELECT id, session_id, project_dir, type, content, content_hash,
			COALESCE(metadata,''), COALESCE(created_at,''), COALESCE(accessed_at,''), access_count
		FROM memories WHERE id = ?`, id,
	).Scan(&m.ID, &m.SessionID, &m.ProjectDir, &m.Type, &m.Content, &m.ContentHash,
		&m.Metadata, &createdStr, &accessedStr, &m.AccessCount)
	if err == nil {
		m.CreatedAt = parseTime(createdStr)
		m.AccessedAt = parseTime(accessedStr)
		return m, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Try prefix match
	err = s.db.QueryRow(
		`SELECT id, session_id, project_dir, type, content, content_hash,
			COALESCE(metadata,''), COALESCE(created_at,''), COALESCE(accessed_at,''), access_count
		FROM memories WHERE id LIKE ? LIMIT 1`, id+"%",
	).Scan(&m.ID, &m.SessionID, &m.ProjectDir, &m.Type, &m.Content, &m.ContentHash,
		&m.Metadata, &createdStr, &accessedStr, &m.AccessCount)
	if err != nil {
		return nil, err
	}
	m.CreatedAt = parseTime(createdStr)
	m.AccessedAt = parseTime(accessedStr)
	return m, nil
}

// DeleteMemory removes a memory entry by ID.
func (s *SQLiteStore) DeleteMemory(id string) error {
	if s.hasFTS5 {
		_, _ = s.db.Exec(`DELETE FROM memories_fts WHERE id = ?`, id)
	}
	_, err := s.db.Exec(`DELETE FROM memories WHERE id = ?`, id)
	return err
}

// ListMemories lists memories for a project, ordered by creation time desc.
func (s *SQLiteStore) ListMemories(projectDir string, limit int) ([]Memory, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT id, session_id, project_dir, type, content, content_hash,
			COALESCE(metadata,''), COALESCE(created_at,''), COALESCE(accessed_at,''), access_count
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

// SaveSession inserts or updates a session history record.
func (s *SQLiteStore) SaveSession(sr SessionRecord) error {
	topicsJSON, _ := json.Marshal(sr.Topics)
	filesJSON, _ := json.Marshal(sr.FilesModified)

	if sr.EndedAt.IsZero() {
		sr.EndedAt = time.Now()
	}

	startedStr := ""
	if !sr.StartedAt.IsZero() {
		startedStr = sr.StartedAt.Format(timeFormat)
	}

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO session_history
			(session_id, project_dir, started_at, ended_at, summary, topics, files_modified, model)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sr.SessionID, sr.ProjectDir, startedStr, sr.EndedAt.Format(timeFormat),
		sr.Summary, string(topicsJSON), string(filesJSON), sr.Model,
	)
	return err
}

// GetSession retrieves a session record by ID.
func (s *SQLiteStore) GetSession(sessionID string) (*SessionRecord, error) {
	sr := &SessionRecord{}
	var topicsStr, filesStr, startedStr, endedStr string
	err := s.db.QueryRow(
		`SELECT session_id, project_dir, COALESCE(started_at,''), COALESCE(ended_at,''),
			COALESCE(summary,''), COALESCE(topics,'[]'), COALESCE(files_modified,'[]'), COALESCE(model,'')
		FROM session_history WHERE session_id = ?`, sessionID,
	).Scan(&sr.SessionID, &sr.ProjectDir, &startedStr, &endedStr,
		&sr.Summary, &topicsStr, &filesStr, &sr.Model)
	if err != nil {
		return nil, err
	}
	sr.StartedAt = parseTime(startedStr)
	sr.EndedAt = parseTime(endedStr)
	_ = json.Unmarshal([]byte(topicsStr), &sr.Topics)
	_ = json.Unmarshal([]byte(filesStr), &sr.FilesModified)
	return sr, nil
}

// GetLastSession retrieves the most recent session for a project.
func (s *SQLiteStore) GetLastSession(projectDir string) (*SessionRecord, error) {
	sr := &SessionRecord{}
	var topicsStr, filesStr, startedStr, endedStr string
	err := s.db.QueryRow(
		`SELECT session_id, project_dir, COALESCE(started_at,''), COALESCE(ended_at,''),
			COALESCE(summary,''), COALESCE(topics,'[]'), COALESCE(files_modified,'[]'), COALESCE(model,'')
		FROM session_history WHERE project_dir = ?
		ORDER BY ended_at DESC LIMIT 1`, projectDir,
	).Scan(&sr.SessionID, &sr.ProjectDir, &startedStr, &endedStr,
		&sr.Summary, &topicsStr, &filesStr, &sr.Model)
	if err != nil {
		return nil, err
	}
	sr.StartedAt = parseTime(startedStr)
	sr.EndedAt = parseTime(endedStr)
	_ = json.Unmarshal([]byte(topicsStr), &sr.Topics)
	_ = json.Unmarshal([]byte(filesStr), &sr.FilesModified)
	return sr, nil
}

// GetSessionsByProject retrieves recent sessions for a project.
func (s *SQLiteStore) GetSessionsByProject(projectDir string, limit int) ([]SessionRecord, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(
		`SELECT session_id, project_dir, COALESCE(started_at,''), COALESCE(ended_at,''),
			COALESCE(summary,''), COALESCE(topics,'[]'), COALESCE(files_modified,'[]'), COALESCE(model,'')
		FROM session_history WHERE project_dir = ?
		ORDER BY ended_at DESC LIMIT ?`,
		projectDir, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionRecord
	for rows.Next() {
		sr := SessionRecord{}
		var topicsStr, filesStr, startedStr, endedStr string
		if err := rows.Scan(&sr.SessionID, &sr.ProjectDir, &startedStr, &endedStr,
			&sr.Summary, &topicsStr, &filesStr, &sr.Model); err != nil {
			continue
		}
		sr.StartedAt = parseTime(startedStr)
		sr.EndedAt = parseTime(endedStr)
		_ = json.Unmarshal([]byte(topicsStr), &sr.Topics)
		_ = json.Unmarshal([]byte(filesStr), &sr.FilesModified)
		sessions = append(sessions, sr)
	}
	return sessions, nil
}

// SaveKnowledge inserts or updates project knowledge.
func (s *SQLiteStore) SaveKnowledge(k ProjectKnowledge) error {
	if k.ID == "" {
		k.ID = generateID()
	}
	if k.CreatedAt.IsZero() {
		k.CreatedAt = time.Now()
	}
	k.UpdatedAt = time.Now()

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO project_knowledge
			(id, project_dir, file_path, knowledge_type, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		k.ID, k.ProjectDir, k.FilePath, k.KnowledgeType, k.Content,
		k.CreatedAt.Format(timeFormat), k.UpdatedAt.Format(timeFormat),
	)
	return err
}

// GetProjectKnowledge retrieves all knowledge entries for a project.
func (s *SQLiteStore) GetProjectKnowledge(projectDir string) ([]ProjectKnowledge, error) {
	rows, err := s.db.Query(
		`SELECT id, project_dir, COALESCE(file_path,''), COALESCE(knowledge_type,''),
			content, COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM project_knowledge WHERE project_dir = ?
		ORDER BY updated_at DESC`,
		projectDir,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ProjectKnowledge
	for rows.Next() {
		k := ProjectKnowledge{}
		var createdStr, updatedStr string
		if err := rows.Scan(&k.ID, &k.ProjectDir, &k.FilePath, &k.KnowledgeType,
			&k.Content, &createdStr, &updatedStr); err != nil {
			continue
		}
		k.CreatedAt = parseTime(createdStr)
		k.UpdatedAt = parseTime(updatedStr)
		results = append(results, k)
	}
	return results, nil
}

// GetStats returns statistics about the memory database.
func (s *SQLiteStore) GetStats(projectDir string) (*MemoryStats, error) {
	stats := &MemoryStats{}

	s.db.QueryRow(`SELECT COUNT(*) FROM memories WHERE project_dir = ?`, projectDir).Scan(&stats.TotalMemories)
	s.db.QueryRow(`SELECT COUNT(*) FROM session_history WHERE project_dir = ?`, projectDir).Scan(&stats.TotalSessions)
	s.db.QueryRow(`SELECT COUNT(*) FROM project_knowledge WHERE project_dir = ?`, projectDir).Scan(&stats.TotalKnowledge)

	s.db.QueryRow(
		`SELECT COALESCE(MIN(created_at),'') FROM memories WHERE project_dir = ?`, projectDir,
	).Scan(&stats.OldestMemory)
	s.db.QueryRow(
		`SELECT COALESCE(MAX(created_at),'') FROM memories WHERE project_dir = ?`, projectDir,
	).Scan(&stats.NewestMemory)

	// Get DB file size
	if s.dbPath != "" && s.dbPath != ":memory:" {
		if info, err := os.Stat(s.dbPath); err == nil {
			stats.DBSizeBytes = info.Size()
		}
	}

	return stats, nil
}

// ContentHash computes SHA256 hash for content deduplication.
func ContentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

func scanMemories(rows *sql.Rows) ([]Memory, error) {
	var memories []Memory
	for rows.Next() {
		m := Memory{}
		var createdStr, accessedStr string
		if err := rows.Scan(&m.ID, &m.SessionID, &m.ProjectDir, &m.Type, &m.Content, &m.ContentHash,
			&m.Metadata, &createdStr, &accessedStr, &m.AccessCount); err != nil {
			continue
		}
		m.CreatedAt = parseTime(createdStr)
		m.AccessedAt = parseTime(accessedStr)
		memories = append(memories, m)
	}
	return memories, nil
}

// parseTime parses a time string stored in SQLite.
// Supports multiple formats since SQLite doesn't enforce time format.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	formats := []string{
		timeFormat,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02T15:04:05.999999999Z",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}

	return time.Time{}
}

// DB returns the underlying *sql.DB for advanced operations (testing).
func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}

// EncodeEmbedding converts float32 slice to binary bytes (LittleEndian).
func EncodeEmbedding(v []float32) []byte {
	buf := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

// DecodeEmbedding converts binary bytes back to float32 slice.
func DecodeEmbedding(b []byte) []float32 {
	if len(b) == 0 || len(b)%4 != 0 {
		return nil
	}
	v := make([]float32, len(b)/4)
	for i := range v {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return v
}

// --- Chunk CRUD methods (2-Layer Memory Architecture) ---

// SaveChunks inserts chunks into the chunks table within a transaction.
// Also inserts into chunks_fts if available.
func (s *SQLiteStore) SaveChunks(chunks []Chunk) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO chunks (path, start_line, end_line, text, hash, heading, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range chunks {
		var embBlob []byte
		if len(chunks[i].Embedding) > 0 {
			embBlob = EncodeEmbedding(chunks[i].Embedding)
		}
		result, err := stmt.Exec(chunks[i].Path, chunks[i].StartLine, chunks[i].EndLine,
			chunks[i].Text, chunks[i].Hash, chunks[i].Heading, embBlob)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		chunks[i].ID = id

		// Insert into FTS
		if s.hasChunksFTS5 {
			tx.Exec(`INSERT INTO chunks_fts(rowid, text) VALUES (?, ?)`, id, chunks[i].Text)
		}
	}

	return tx.Commit()
}

// DeleteChunksByPath removes all chunks for a given file path.
func (s *SQLiteStore) DeleteChunksByPath(path string) error {
	if s.hasChunksFTS5 {
		// Delete from FTS first (content-sync table needs explicit delete)
		s.db.Exec(`DELETE FROM chunks_fts WHERE rowid IN (SELECT id FROM chunks WHERE path = ?)`, path)
	}
	_, err := s.db.Exec(`DELETE FROM chunks WHERE path = ?`, path)
	return err
}

// GetFileIndex retrieves the file index entry for a path.
func (s *SQLiteStore) GetFileIndex(path string) (*FileIndexEntry, error) {
	entry := &FileIndexEntry{}
	err := s.db.QueryRow(
		`SELECT path, last_modified, chunk_count, last_indexed FROM file_index WHERE path = ?`,
		path,
	).Scan(&entry.Path, &entry.LastModified, &entry.ChunkCount, &entry.LastIndexed)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// UpdateFileIndex upserts a file index entry.
func (s *SQLiteStore) UpdateFileIndex(entry FileIndexEntry) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO file_index (path, last_modified, chunk_count, last_indexed)
		VALUES (?, ?, ?, ?)`,
		entry.Path, entry.LastModified, entry.ChunkCount, entry.LastIndexed,
	)
	return err
}

// ChunkCount returns the total number of chunks in the database.
func (s *SQLiteStore) ChunkCount() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM chunks`).Scan(&count)
	return count, err
}

// FileCount returns the total number of indexed files.
func (s *SQLiteStore) FileCount() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM file_index`).Scan(&count)
	return count, err
}
