package memory

import "database/sql"

const schemaVersion = "3"

// InitSchema creates all tables, indexes, and FTS5 virtual tables.
func InitSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []string{
		// Meta table
		`CREATE TABLE IF NOT EXISTS meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,

		// Memories table (v2: includes embedding BLOB)
		`CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			project_dir TEXT NOT NULL,
			type TEXT NOT NULL,
			content TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			metadata TEXT,
			embedding BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accessed_at DATETIME,
			access_count INTEGER DEFAULT 0
		)`,

		// Embedding cache table (Phase 2)
		`CREATE TABLE IF NOT EXISTS embedding_cache (
			content_hash TEXT NOT NULL,
			provider TEXT NOT NULL,
			model TEXT NOT NULL,
			embedding BLOB NOT NULL,
			dims INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (content_hash, provider, model)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_embedding_cache_created ON embedding_cache(created_at)`,

		// Project knowledge table
		`CREATE TABLE IF NOT EXISTS project_knowledge (
			id TEXT PRIMARY KEY,
			project_dir TEXT NOT NULL,
			file_path TEXT,
			knowledge_type TEXT,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME
		)`,

		// Session history table
		`CREATE TABLE IF NOT EXISTS session_history (
			session_id TEXT PRIMARY KEY,
			project_dir TEXT NOT NULL,
			started_at DATETIME,
			ended_at DATETIME,
			summary TEXT,
			topics TEXT,
			files_modified TEXT,
			model TEXT
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_memories_project ON memories(project_dir)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_hash ON memories(content_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_session ON memories(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_project ON project_knowledge(project_dir)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_project ON session_history(project_dir)`,

		// Chunks table (v3: 2-Layer Memory Architecture)
		`CREATE TABLE IF NOT EXISTS chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			text TEXT NOT NULL,
			hash TEXT NOT NULL,
			heading TEXT DEFAULT '',
			embedding BLOB
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_hash ON chunks(hash)`,

		// File index table (v3: tracks indexing state)
		`CREATE TABLE IF NOT EXISTS file_index (
			path TEXT PRIMARY KEY,
			last_modified INTEGER NOT NULL,
			chunk_count INTEGER NOT NULL,
			last_indexed INTEGER NOT NULL
		)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	// FTS5 virtual table (standalone, not content-sync)
	_, err = tx.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		content,
		id UNINDEXED,
		project_dir UNINDEXED,
		type UNINDEXED
	)`)
	if err != nil {
		// FTS5 not available — continue without full-text search
		// This is non-fatal; search will fall back to LIKE queries
	}

	// FTS5 for chunks (non-fatal if not available)
	_, _ = tx.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
		text,
		content=chunks,
		content_rowid=id
	)`)

	// Set schema version
	_, err = tx.Exec(
		`INSERT OR REPLACE INTO meta (key, value) VALUES (?, ?)`,
		"schema_version", schemaVersion,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetSchemaVersion reads the current schema version from meta table.
func GetSchemaVersion(db *sql.DB) (string, error) {
	return GetMeta(db, "schema_version")
}

// SetMeta sets a key-value pair in the meta table.
func SetMeta(db *sql.DB, key, value string) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO meta (key, value) VALUES (?, ?)`,
		key, value,
	)
	return err
}

// GetMeta reads a value from the meta table.
func GetMeta(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM meta WHERE key = ?`, key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// MigrateV1toV2 adds embedding support: embedding BLOB column + embedding_cache table.
func MigrateV1toV2(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	migrations := []string{
		// Add embedding BLOB column to memories
		`ALTER TABLE memories ADD COLUMN embedding BLOB`,

		// Embedding cache table (Clawdbot-compatible pattern)
		`CREATE TABLE IF NOT EXISTS embedding_cache (
			content_hash TEXT NOT NULL,
			provider TEXT NOT NULL,
			model TEXT NOT NULL,
			embedding BLOB NOT NULL,
			dims INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (content_hash, provider, model)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_embedding_cache_created ON embedding_cache(created_at)`,
	}

	for _, stmt := range migrations {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	// Update schema version
	if _, err := tx.Exec(
		`INSERT OR REPLACE INTO meta (key, value) VALUES (?, ?)`,
		"schema_version", "2",
	); err != nil {
		return err
	}

	return tx.Commit()
}

// EnsureSchema checks the current schema version and applies migrations if needed.
func EnsureSchema(db *sql.DB) error {
	version, err := GetSchemaVersion(db)
	if err != nil {
		// No schema yet — initialize from scratch
		return InitSchema(db)
	}

	switch version {
	case "3":
		return nil // up to date
	case "2":
		return MigrateV2toV3(db)
	case "1":
		if err := MigrateV1toV2(db); err != nil {
			return err
		}
		return MigrateV2toV3(db)
	default:
		return InitSchema(db)
	}
}

// HasFTS5 checks whether the FTS5 virtual table exists.
func HasFTS5(db *sql.DB) bool {
	var name string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='memories_fts'`,
	).Scan(&name)
	return err == nil && name == "memories_fts"
}

// HasChunksFTS5 checks whether the chunks_fts virtual table exists.
func HasChunksFTS5(db *sql.DB) bool {
	var name string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='chunks_fts'`,
	).Scan(&name)
	return err == nil && name == "chunks_fts"
}

// MigrateV2toV3 adds chunks table, file_index table, and chunks_fts for 2-Layer Memory Architecture.
func MigrateV2toV3(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	migrations := []string{
		// Chunks table
		`CREATE TABLE IF NOT EXISTS chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			text TEXT NOT NULL,
			hash TEXT NOT NULL,
			heading TEXT DEFAULT '',
			embedding BLOB
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_hash ON chunks(hash)`,

		// File index table
		`CREATE TABLE IF NOT EXISTS file_index (
			path TEXT PRIMARY KEY,
			last_modified INTEGER NOT NULL,
			chunk_count INTEGER NOT NULL,
			last_indexed INTEGER NOT NULL
		)`,
	}

	for _, stmt := range migrations {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	// FTS5 for chunks (non-fatal if not available)
	_, _ = tx.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
		text,
		content=chunks,
		content_rowid=id
	)`)

	// Update schema version
	if _, err := tx.Exec(
		`INSERT OR REPLACE INTO meta (key, value) VALUES (?, ?)`,
		"schema_version", "3",
	); err != nil {
		return err
	}

	return tx.Commit()
}
