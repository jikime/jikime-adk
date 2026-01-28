package memory

import (
	"context"
	"time"
)

// Memory type constants
const (
	TypeSessionSummary = "session_summary"
	TypeDecision       = "decision"
	TypeLearning       = "learning"
	TypeToolUsage      = "tool_usage"
	TypeErrorFix       = "error_fix"
	TypeUserPrompt         = "user_prompt"
	TypeAssistantResponse  = "assistant_response"
)

// Hybrid search constants (Clawdbot-compatible)
const (
	DefaultVectorWeight = 0.7
	DefaultTextWeight   = 0.3
	DefaultMinScore     = 0.35
	DefaultMaxResults   = 6
)

// Knowledge type constants
const (
	KnowledgeArchitecture = "architecture"
	KnowledgePattern      = "pattern"
	KnowledgeConvention   = "convention"
	KnowledgeDecision     = "decision"
)

// Memory represents a single memory entry in the DB.
type Memory struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	ProjectDir  string    `json:"project_dir"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	ContentHash string    `json:"content_hash"`
	Metadata    string    `json:"metadata,omitempty"`
	Embedding   []float32 `json:"-"` // Phase 2: vector embedding (BLOB in DB)
	CreatedAt   time.Time `json:"created_at"`
	AccessedAt  time.Time `json:"accessed_at,omitempty"`
	AccessCount int       `json:"access_count"`
}

// SessionRecord represents a session history entry.
type SessionRecord struct {
	SessionID     string    `json:"session_id"`
	ProjectDir    string    `json:"project_dir"`
	StartedAt     time.Time `json:"started_at,omitempty"`
	EndedAt       time.Time `json:"ended_at,omitempty"`
	Summary       string    `json:"summary,omitempty"`
	Topics        []string  `json:"topics,omitempty"`
	FilesModified []string  `json:"files_modified,omitempty"`
	Model         string    `json:"model,omitempty"`
}

// ProjectKnowledge represents project-level knowledge.
type ProjectKnowledge struct {
	ID            string    `json:"id"`
	ProjectDir    string    `json:"project_dir"`
	FilePath      string    `json:"file_path,omitempty"`
	KnowledgeType string    `json:"knowledge_type"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

// SearchResult represents a search result with relevance score.
type SearchResult struct {
	Memory Memory  `json:"memory"`
	Score  float64 `json:"score"`
}

// SearchQuery holds search parameters.
type SearchQuery struct {
	ProjectDir string  `json:"project_dir"`
	Query      string  `json:"query"`
	Type       string  `json:"type,omitempty"`
	Limit      int     `json:"limit"`
	MinScore   float64 `json:"min_score,omitempty"`
}

// ExtractOptions configures memory extraction behavior.
type ExtractOptions struct {
	SessionID  string
	ProjectDir string
	Trigger    string // "manual" or "auto"
}

// Transcript holds parsed transcript data.
type Transcript struct {
	Summaries []TranscriptRecord
	UserMsgs  []TranscriptRecord
	SessionID string
}

// TranscriptRecord represents a single record from the JSONL transcript.
type TranscriptRecord struct {
	Type      string                 `json:"type"`
	Message   interface{}            `json:"message,omitempty"`
	Summary   string                 `json:"summary,omitempty"`
	SessionID string                 `json:"sessionId,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	UUID      string                 `json:"uuid,omitempty"`
	CWD       string                 `json:"cwd,omitempty"`
	GitBranch string                 `json:"gitBranch,omitempty"`
	LeafUUID  string                 `json:"leafUuid,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// SessionSummary is the output of the Summarize function.
type SessionSummary struct {
	Text   string   `json:"text"`
	Topics []string `json:"topics"`
	Files  []string `json:"files"`
}

// MemoryStats holds statistics about the memory DB.
type MemoryStats struct {
	TotalMemories  int    `json:"total_memories"`
	TotalSessions  int    `json:"total_sessions"`
	TotalKnowledge int    `json:"total_knowledge"`
	DBSizeBytes    int64  `json:"db_size_bytes"`
	OldestMemory   string `json:"oldest_memory,omitempty"`
	NewestMemory   string `json:"newest_memory,omitempty"`
}

// GCOptions configures garbage collection behavior.
type GCOptions struct {
	MaxAge   time.Duration
	MaxCount int
	DryRun   bool
}

// GCResult reports what was cleaned up.
type GCResult struct {
	DeletedByAge   int `json:"deleted_by_age"`
	DeletedByCount int `json:"deleted_by_count"`
	Remaining      int `json:"remaining"`
}

// EmbeddingProvider defines the interface for embedding generation.
type EmbeddingProvider interface {
	ID() string
	Model() string
	Dims() int
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// EmbeddingConfig holds embedding provider configuration.
type EmbeddingConfig struct {
	Provider string `json:"provider"` // "openai", "gemini", "auto", "none"
	Model    string `json:"model"`    // e.g. "text-embedding-3-small"
	APIKey   string `json:"-"`        // never serialized
	BaseURL  string `json:"base_url"` // optional override
	Fallback string `json:"fallback"` // fallback provider
	Dims     int    `json:"dims"`     // embedding dimensions
}

// HybridSearchResult extends SearchResult with vector+text scores.
type HybridSearchResult struct {
	Memory      Memory  `json:"memory"`
	Score       float64 `json:"score"`        // final = 0.7*vec + 0.3*text
	VectorScore float64 `json:"vector_score"`
	TextScore   float64 `json:"text_score"`
}

// FileTrackRecord represents a file modification from PostToolUse.
type FileTrackRecord struct {
	SessionID string `json:"session_id"`
	FilePath  string `json:"file_path"`
	ToolName  string `json:"tool_name"`
	Timestamp string `json:"timestamp"`
}

// --- 2-Layer Memory Architecture: Chunk types ---

// Chunk represents a text chunk from an indexed MD file.
type Chunk struct {
	ID        int64     `json:"id"`         // SQLite rowid
	Path      string    `json:"path"`       // relative file path (e.g. "memory/2026-01-27.md")
	StartLine int       `json:"start_line"` // 1-based
	EndLine   int       `json:"end_line"`   // 1-based
	Text      string    `json:"text"`
	Hash      string    `json:"hash"`      // SHA256 of Text for change detection
	Heading   string    `json:"heading"`   // parent heading context
	Embedding []float32 `json:"-"`         // optional vector embedding
}

// ChunkOpts configures chunking behavior.
type ChunkOpts struct {
	MaxTokens    int // approximate max tokens per chunk (default 400)
	Overlap      int // approximate overlap tokens (default 80)
	MinChunkSize int // minimum chunk size in bytes to keep (default 50)
}

// DefaultChunkOpts returns Clawdbot-compatible defaults.
func DefaultChunkOpts() ChunkOpts {
	return ChunkOpts{
		MaxTokens:    400,
		Overlap:      80,
		MinChunkSize: 50,
	}
}

// ChunkSearchResult represents a search result from the chunks table.
type ChunkSearchResult struct {
	Chunk       Chunk   `json:"chunk"`
	Score       float64 `json:"score"`
	VectorScore float64 `json:"vector_score"`
	TextScore   float64 `json:"text_score"`
}

// FileIndexEntry tracks indexing state for a file.
type FileIndexEntry struct {
	Path         string `json:"path"`
	LastModified int64  `json:"last_modified"` // unix timestamp
	ChunkCount   int    `json:"chunk_count"`
	LastIndexed  int64  `json:"last_indexed"` // unix timestamp
}

// DailyLogEntry represents an entry to append to a daily log MD file.
type DailyLogEntry struct {
	Type     string `json:"type"`              // decision, learning, error_fix, tool_usage, session_summary
	Content  string `json:"content"`
	Metadata string `json:"metadata,omitempty"` // optional JSON string
}

// MemoryMDEntry represents an entry to append to MEMORY.md.
type MemoryMDEntry struct {
	Section string `json:"section"` // e.g. "Architecture Decisions", "Key Patterns"
	Content string `json:"content"`
}
