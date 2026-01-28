package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewStoreWithPath(dbPath)
	if err != nil {
		t.Fatalf("NewStoreWithPath: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestInitSchema(t *testing.T) {
	store := setupTestStore(t)

	ver, err := GetSchemaVersion(store.DB())
	if err != nil {
		t.Fatalf("GetSchemaVersion: %v", err)
	}
	if ver != schemaVersion {
		t.Errorf("got version %q, want %q", ver, schemaVersion)
	}
}

func TestSaveAndGetMemory(t *testing.T) {
	store := setupTestStore(t)

	m := Memory{
		ID:         "test-001",
		SessionID:  "session-abc",
		ProjectDir: "/tmp/test",
		Type:       TypeDecision,
		Content:    "Decided to use JWT for authentication",
		CreatedAt:  time.Now(),
	}

	if err := store.SaveMemory(m); err != nil {
		t.Fatalf("SaveMemory: %v", err)
	}

	got, err := store.GetMemory("test-001")
	if err != nil {
		t.Fatalf("GetMemory: %v", err)
	}

	if got.Content != m.Content {
		t.Errorf("content = %q, want %q", got.Content, m.Content)
	}
	if got.Type != TypeDecision {
		t.Errorf("type = %q, want %q", got.Type, TypeDecision)
	}
	if got.ContentHash == "" {
		t.Error("content_hash should be auto-generated")
	}
}

func TestSaveIfNew_Dedup(t *testing.T) {
	store := setupTestStore(t)

	m := Memory{
		SessionID:  "session-abc",
		ProjectDir: "/tmp/test",
		Type:       TypeLearning,
		Content:    "Go interfaces are implicitly implemented",
	}

	// First save — should succeed
	saved, err := store.SaveIfNew(m)
	if err != nil {
		t.Fatalf("SaveIfNew (1st): %v", err)
	}
	if !saved {
		t.Error("expected first save to succeed")
	}

	// Second save with same content — should be duplicate
	saved, err = store.SaveIfNew(m)
	if err != nil {
		t.Fatalf("SaveIfNew (2nd): %v", err)
	}
	if saved {
		t.Error("expected duplicate to be skipped")
	}
}

func TestDeleteMemory(t *testing.T) {
	store := setupTestStore(t)

	m := Memory{
		ID:         "del-001",
		SessionID:  "session-del",
		ProjectDir: "/tmp/test",
		Type:       TypeErrorFix,
		Content:    "Fixed nil pointer in handler",
		CreatedAt:  time.Now(),
	}
	store.SaveMemory(m)

	if err := store.DeleteMemory("del-001"); err != nil {
		t.Fatalf("DeleteMemory: %v", err)
	}

	_, err := store.GetMemory("del-001")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestListMemories(t *testing.T) {
	store := setupTestStore(t)

	for i := 0; i < 5; i++ {
		store.SaveMemory(Memory{
			ID:         generateID(),
			SessionID:  "session-list",
			ProjectDir: "/tmp/test",
			Type:       TypeDecision,
			Content:    "Decision " + string(rune('A'+i)),
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Second),
		})
	}

	memories, err := store.ListMemories("/tmp/test", 3)
	if err != nil {
		t.Fatalf("ListMemories: %v", err)
	}
	if len(memories) != 3 {
		t.Errorf("got %d memories, want 3", len(memories))
	}
}

func TestFTS5Search(t *testing.T) {
	store := setupTestStore(t)

	if !store.hasFTS5 {
		t.Skip("FTS5 not available")
	}

	memories := []Memory{
		{SessionID: "s1", ProjectDir: "/tmp/test", Type: TypeDecision, Content: "Use JWT for authentication tokens", CreatedAt: time.Now()},
		{SessionID: "s1", ProjectDir: "/tmp/test", Type: TypeLearning, Content: "Go channels are great for concurrency", CreatedAt: time.Now()},
		{SessionID: "s1", ProjectDir: "/tmp/test", Type: TypeDecision, Content: "PostgreSQL for database storage", CreatedAt: time.Now()},
	}

	for _, m := range memories {
		if err := store.SaveMemory(m); err != nil {
			t.Fatalf("SaveMemory: %v", err)
		}
	}

	results, err := store.Search(SearchQuery{
		ProjectDir: "/tmp/test",
		Query:      "authentication JWT",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search results for 'authentication JWT'")
	}
	if len(results) > 0 && results[0].Score <= 0 {
		t.Error("expected positive score")
	}
}

func TestSessionCRUD(t *testing.T) {
	store := setupTestStore(t)

	sr := SessionRecord{
		SessionID:     "sess-001",
		ProjectDir:    "/tmp/test",
		EndedAt:       time.Now(),
		Summary:       "Implemented auth system",
		Topics:        []string{"auth", "JWT", "middleware"},
		FilesModified: []string{"auth.go", "middleware.go"},
		Model:         "claude-opus",
	}

	if err := store.SaveSession(sr); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	got, err := store.GetSession("sess-001")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.Summary != "Implemented auth system" {
		t.Errorf("summary = %q", got.Summary)
	}
	if len(got.Topics) != 3 {
		t.Errorf("topics = %v, want 3 items", got.Topics)
	}

	last, err := store.GetLastSession("/tmp/test")
	if err != nil {
		t.Fatalf("GetLastSession: %v", err)
	}
	if last.SessionID != "sess-001" {
		t.Errorf("last session = %q", last.SessionID)
	}
}

func TestProjectKnowledge(t *testing.T) {
	store := setupTestStore(t)

	k := ProjectKnowledge{
		ProjectDir:    "/tmp/test",
		KnowledgeType: KnowledgeArchitecture,
		Content:       "Three-tier architecture: Gateway → Service → Repository",
	}

	if err := store.SaveKnowledge(k); err != nil {
		t.Fatalf("SaveKnowledge: %v", err)
	}

	results, err := store.GetProjectKnowledge("/tmp/test")
	if err != nil {
		t.Fatalf("GetProjectKnowledge: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Content != k.Content {
		t.Errorf("content = %q", results[0].Content)
	}
}

func TestGarbageCollect(t *testing.T) {
	store := setupTestStore(t)

	// Insert old and new memories
	oldTime := time.Now().Add(-100 * 24 * time.Hour)
	newTime := time.Now()

	store.SaveMemory(Memory{
		ID: "old-001", SessionID: "s1", ProjectDir: "/tmp/test",
		Type: TypeLearning, Content: "Old memory", CreatedAt: oldTime,
	})
	store.SaveMemory(Memory{
		ID: "new-001", SessionID: "s1", ProjectDir: "/tmp/test",
		Type: TypeLearning, Content: "New memory", CreatedAt: newTime,
	})

	result, err := store.GarbageCollect("/tmp/test", GCOptions{
		MaxAge:   90 * 24 * time.Hour,
		MaxCount: 1000,
	})
	if err != nil {
		t.Fatalf("GarbageCollect: %v", err)
	}
	if result.DeletedByAge != 1 {
		t.Errorf("deleted by age = %d, want 1", result.DeletedByAge)
	}
	if result.Remaining != 1 {
		t.Errorf("remaining = %d, want 1", result.Remaining)
	}
}

func TestGetStats(t *testing.T) {
	store := setupTestStore(t)

	store.SaveMemory(Memory{
		SessionID: "s1", ProjectDir: "/tmp/test",
		Type: TypeDecision, Content: "test", CreatedAt: time.Now(),
	})
	store.SaveSession(SessionRecord{
		SessionID: "s1", ProjectDir: "/tmp/test", EndedAt: time.Now(),
	})

	stats, err := store.GetStats("/tmp/test")
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalMemories != 1 {
		t.Errorf("total memories = %d, want 1", stats.TotalMemories)
	}
	if stats.TotalSessions != 1 {
		t.Errorf("total sessions = %d, want 1", stats.TotalSessions)
	}
}

func TestContentHash(t *testing.T) {
	h1 := ContentHash("hello world")
	h2 := ContentHash("hello world")
	h3 := ContentHash("different content")

	if h1 != h2 {
		t.Error("same content should produce same hash")
	}
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}
	if len(h1) != 64 { // SHA256 hex
		t.Errorf("hash length = %d, want 64", len(h1))
	}
}

func TestTranscriptParser(t *testing.T) {
	// Create a test JSONL file
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "test.jsonl")

	content := `{"type":"summary","summary":"Implemented auth system","leafUuid":"abc123"}
{"type":"user","message":"Implement JWT authentication","sessionId":"sess-001","timestamp":"2026-01-27T10:00:00Z"}
{"type":"progress","data":{"type":"hook_progress"}}
{"type":"user","message":"The architecture uses three layers","sessionId":"sess-001","timestamp":"2026-01-27T10:05:00Z"}
`
	if err := os.WriteFile(transcriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("write test transcript: %v", err)
	}

	transcript, err := ParseTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("ParseTranscript: %v", err)
	}

	if len(transcript.Summaries) != 1 {
		t.Errorf("summaries = %d, want 1", len(transcript.Summaries))
	}
	if len(transcript.UserMsgs) != 2 {
		t.Errorf("user msgs = %d, want 2", len(transcript.UserMsgs))
	}
	if transcript.SessionID != "sess-001" {
		t.Errorf("session ID = %q, want sess-001", transcript.SessionID)
	}
	if transcript.Summaries[0].Summary != "Implemented auth system" {
		t.Errorf("summary = %q", transcript.Summaries[0].Summary)
	}
}

func TestExtractor(t *testing.T) {
	transcript := &Transcript{
		Summaries: []TranscriptRecord{
			{Type: "summary", Summary: "Decided to use JWT for auth"},
		},
		UserMsgs: []TranscriptRecord{
			{Type: "user", Message: "We decided to use PostgreSQL for the database"},
			{Type: "user", Message: "The architecture follows a three-tier design pattern"},
			{Type: "user", Message: "hello"}, // too short, should be skipped
		},
		SessionID: "sess-001",
	}

	extracted := Extract(transcript, ExtractOptions{
		SessionID:  "sess-001",
		ProjectDir: "/tmp/test",
		Trigger:    "auto",
	})

	if len(extracted) < 2 {
		t.Errorf("expected at least 2 extracted memories, got %d", len(extracted))
	}

	// Verify summary was extracted
	hasSummary := false
	for _, m := range extracted {
		if m.Type == TypeSessionSummary {
			hasSummary = true
		}
	}
	if !hasSummary {
		t.Error("expected session summary memory")
	}
}

func TestBuildFTSQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", `"hello" AND "world"`},
		{"JWT authentication", `"JWT" AND "authentication"`},
		{"", ""},
		{"!!!@@@", ""},
		{"test-query", `"test" AND "query"`},
	}

	for _, tt := range tests {
		got := buildFTSQuery(tt.input)
		if got != tt.expected {
			t.Errorf("buildFTSQuery(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildStartupContext(t *testing.T) {
	session := &SessionRecord{
		SessionID: "sess-001",
		EndedAt:   time.Now(),
		Summary:   "Implemented authentication system",
		Topics:    []string{"auth", "JWT"},
	}

	knowledge := []ProjectKnowledge{
		{KnowledgeType: KnowledgeArchitecture, Content: "Three-tier architecture"},
	}

	memories := []Memory{
		{Type: TypeDecision, Content: "Use JWT tokens"},
	}

	ctx := BuildStartupContext(session, knowledge, memories)
	if ctx == "" {
		t.Error("expected non-empty context")
	}
	if !containsStr(ctx, "Session Memory") {
		t.Error("expected 'Session Memory' header")
	}
	if !containsStr(ctx, "Implemented authentication") {
		t.Error("expected session summary in context")
	}
}

func TestEnforceLimit(t *testing.T) {
	short := "short text"
	if enforceLimit(short) != short {
		t.Error("short text should not be truncated")
	}

	// Create a string larger than maxContextBytes
	large := ""
	for len(large) < maxContextBytes+1000 {
		large += "This is a line of text for testing purposes.\n"
	}
	result := enforceLimit(large)
	if len(result) > maxContextBytes+100 { // small margin for truncation message
		t.Errorf("result length %d exceeds limit", len(result))
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
