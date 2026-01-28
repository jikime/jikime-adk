package mcpcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
	"jikime-adk/version"
)

// --- Input/Output types ---

// MemorySearchInput is the input for the memory_search tool.
// Compatible with Clawdbot: query, maxResults, minScore parameters.
type MemorySearchInput struct {
	Query      string  `json:"query" jsonschema:"Semantic search query text"`
	MaxResults int     `json:"maxResults,omitempty" jsonschema:"Maximum number of results to return (default 6)"`
	MinScore   float64 `json:"minScore,omitempty" jsonschema:"Minimum relevance score threshold 0.0-1.0 (default 0.35)"`
	Type       string  `json:"type,omitempty" jsonschema:"Filter by memory type (decision, learning, error_fix, tool_usage, session_summary, user_prompt)"`
}

// MemorySearchOutput is the output of the memory_search tool.
type MemorySearchOutput struct {
	Results  []memorySearchResult `json:"results"`
	Count    int                  `json:"count"`
	Provider string               `json:"provider"` // embedding provider used (e.g. "openai", "gemini", "fts5")
	Model    string               `json:"model"`    // embedding model (e.g. "text-embedding-3-small")
}

type memorySearchResult struct {
	Path      string  `json:"path"`       // file path (chunk source)
	StartLine int     `json:"start_line"` // chunk start line
	EndLine   int     `json:"end_line"`   // chunk end line
	Heading   string  `json:"heading"`    // section heading
	Snippet   string  `json:"snippet"`    // truncated preview (max 200 chars). Use memory_get for full content.
	Score     float64 `json:"score"`
	Source    string  `json:"source"` // "chunks" or "memory"
}

// MemoryGetInput is the input for the memory_get tool.
type MemoryGetInput struct {
	Path  string `json:"path" jsonschema:"Relative file path (e.g. '.jikime/memory/2026-01-27.md') or legacy memory ID"`
	From  int    `json:"from,omitempty" jsonschema:"Start line number (1-based). If omitted, reads from beginning."`
	Lines int    `json:"lines,omitempty" jsonschema:"Number of lines to read. If omitted, reads entire file (or to end from 'from')."`
}

// MemoryGetOutput is the output of the memory_get tool.
type MemoryGetOutput struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line,omitempty"` // 1-based, only set when from is used
	EndLine   int    `json:"end_line,omitempty"`   // 1-based, only set when from is used
	Content   string `json:"content"`
}

// MemoryLoadInput is the input for the memory_load tool.
type MemoryLoadInput struct {
	Source string `json:"source,omitempty" jsonschema:"Context source: 'startup' (MEMORY.md only), 'full' (MEMORY.md + today's daily log). Default: 'startup'"`
}

// MemoryLoadOutput is the output of the memory_load tool.
type MemoryLoadOutput struct {
	Content string   `json:"content"`
	Files   []string `json:"files"` // which files were loaded
}

// MemorySaveInput is the input for the memory_save tool.
type MemorySaveInput struct {
	Type     string `json:"type" jsonschema:"Memory type: decision, learning, error_fix, tool_usage"`
	Content  string `json:"content" jsonschema:"Memory content text"`
	Metadata string `json:"metadata,omitempty" jsonschema:"Optional JSON metadata string"`
}

// MemorySaveOutput is the output of the memory_save tool.
type MemorySaveOutput struct {
	ID      string `json:"id"`
	Saved   bool   `json:"saved"`
	Message string `json:"message"`
}

// MemoryStatsInput is the input for the memory_stats tool (no params).
type MemoryStatsInput struct{}

// MemoryStatsOutput is the output of the memory_stats tool.
type MemoryStatsOutput struct {
	TotalMemories  int    `json:"total_memories"`
	TotalSessions  int    `json:"total_sessions"`
	TotalKnowledge int    `json:"total_knowledge"`
	TotalChunks    int    `json:"total_chunks"`
	IndexedFiles   int    `json:"indexed_files"`
	DBSizeBytes    int64  `json:"db_size_bytes"`
	OldestMemory   string `json:"oldest_memory,omitempty"`
	NewestMemory   string `json:"newest_memory,omitempty"`
}

// MemoryReindexInput is the input for the memory_reindex tool (no params).
type MemoryReindexInput struct{}

// MemoryReindexOutput is the output of the memory_reindex tool.
type MemoryReindexOutput struct {
	IndexedFiles int    `json:"indexed_files"`
	TotalChunks  int    `json:"total_chunks"`
	Message      string `json:"message"`
}

// --- Command ---

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start MCP memory server (STDIO transport)",
		Long: `Starts an MCP server that exposes memory tools over STDIO.
Claude Code connects to this server to search, retrieve, save, and inspect memories.`,
		RunE:         runServe,
		SilenceUsage: true,
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "jikime-memory",
			Version: version.String(),
		},
		nil,
	)

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_search",
		Description: "Search project memories using hybrid vector + text search. Searches indexed MD file chunks. Use this to find relevant past decisions, learnings, error fixes, and session context.",
	}, handleMemorySearch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_get",
		Description: "Read specific lines from a memory file after memory_search. Supports line-range reading with 'from' and 'lines' parameters, or reads the entire file if omitted.",
	}, handleMemoryGet)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_load",
		Description: "Load project knowledge context on demand. Reads MEMORY.md and optionally today's daily log. Call this at session start or when you need project-level context about architecture, patterns, and conventions.",
	}, handleMemoryLoad)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_save",
		Description: "Save a new memory (decision, learning, error fix, or tool usage pattern) to the daily log MD file.",
	}, handleMemorySave)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_stats",
		Description: "Get statistics about the project memory database, including chunk index stats.",
	}, handleMemoryStats)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_reindex",
		Description: "Re-index all memory MD files. Run this after manual edits to memory files.",
	}, handleMemoryReindex)

	// Start file watcher for auto-indexing MD changes
	projectDir, _ := os.Getwd()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if store, err := memory.NewStore(projectDir); err == nil {
		cfg := memory.LoadEmbeddingConfig()
		provider, _ := memory.NewEmbeddingProvider(cfg)
		go memory.WatchMemoryFiles(ctx, projectDir, store, provider)
		defer store.Close()
	}

	return server.Run(ctx, &mcp.StdioTransport{})
}

// --- Handlers ---

func handleMemorySearch(ctx context.Context, req *mcp.CallToolRequest, input MemorySearchInput) (
	*mcp.CallToolResult, MemorySearchOutput, error,
) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, MemorySearchOutput{}, fmt.Errorf("get working directory: %w", err)
	}

	store, err := memory.NewStore(projectDir)
	if err != nil {
		return nil, MemorySearchOutput{}, fmt.Errorf("open memory store: %w", err)
	}
	defer store.Close()

	limit := input.MaxResults
	if limit <= 0 {
		limit = 6
	}
	minScore := input.MinScore
	if minScore <= 0 {
		minScore = memory.DefaultMinScore
	}

	q := memory.SearchQuery{
		ProjectDir: projectDir,
		Query:      input.Query,
		Type:       input.Type,
		Limit:      limit,
		MinScore:   minScore,
	}

	cfg := memory.LoadEmbeddingConfig()
	provider, _ := memory.NewEmbeddingProvider(cfg)

	searchCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	// Search both sources and merge results.
	// - chunks table: indexed MD files (memory_save tool, memory_reindex)
	// - memories table: session data (user_prompt, assistant_response, tool_usage, session_summary)
	allResults := make([]memorySearchResult, 0)

	// 1. Search chunks table (indexed MD content)
	chunkResults, _ := store.SearchChunks(searchCtx, q, provider)
	for _, r := range chunkResults {
		allResults = append(allResults, memorySearchResult{
			Path:      r.Chunk.Path,
			StartLine: r.Chunk.StartLine,
			EndLine:   r.Chunk.EndLine,
			Heading:   r.Chunk.Heading,
			Snippet:   makeSnippet(r.Chunk.Text, 200),
			Score:     r.Score,
			Source:    "chunks",
		})
	}

	// 2. Search memories table (session data with embeddings)
	hybridResults, _ := store.SearchHybrid(searchCtx, q, provider)
	for _, r := range hybridResults {
		allResults = append(allResults, memorySearchResult{
			Snippet: makeSnippet(r.Memory.Content, 200),
			Score:   r.Score,
			Source:  "memory",
		})
	}

	// 3. Sort by score descending and limit
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	providerName := "fts5"
	modelName := ""
	if provider != nil {
		providerName = provider.ID()
		modelName = provider.Model()
	}

	output := MemorySearchOutput{
		Results:  allResults,
		Count:    len(allResults),
		Provider: providerName,
		Model:    modelName,
	}

	return nil, output, nil
}

func handleMemoryGet(ctx context.Context, req *mcp.CallToolRequest, input MemoryGetInput) (
	*mcp.CallToolResult, MemoryGetOutput, error,
) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, MemoryGetOutput{}, fmt.Errorf("get working directory: %w", err)
	}

	// Try reading as file path first
	absPath := filepath.Join(projectDir, input.Path)
	data, err := os.ReadFile(absPath)
	if err == nil {
		content := string(data)

		// If from/lines specified, extract the line range
		if input.From > 0 {
			lines := strings.Split(content, "\n")
			startIdx := input.From - 1 // convert to 0-based
			if startIdx < 0 {
				startIdx = 0
			}
			if startIdx >= len(lines) {
				return nil, MemoryGetOutput{
					Path:    input.Path,
					Content: "",
				}, nil
			}

			endIdx := len(lines)
			if input.Lines > 0 {
				endIdx = startIdx + input.Lines
				if endIdx > len(lines) {
					endIdx = len(lines)
				}
			}

			return nil, MemoryGetOutput{
				Path:      input.Path,
				StartLine: startIdx + 1, // back to 1-based
				EndLine:   endIdx,
				Content:   strings.Join(lines[startIdx:endIdx], "\n"),
			}, nil
		}

		return nil, MemoryGetOutput{
			Path:    input.Path,
			Content: content,
		}, nil
	}

	// Fallback: try as legacy memory ID
	store, storeErr := memory.NewStore(projectDir)
	if storeErr != nil {
		return nil, MemoryGetOutput{}, fmt.Errorf("open memory store: %w", storeErr)
	}
	defer store.Close()

	m, getErr := store.GetMemory(input.Path)
	if getErr != nil {
		return nil, MemoryGetOutput{}, fmt.Errorf("not found: %w", getErr)
	}

	return nil, MemoryGetOutput{
		Path:    "legacy:" + m.ID,
		Content: m.Content,
	}, nil
}

func handleMemoryLoad(ctx context.Context, req *mcp.CallToolRequest, input MemoryLoadInput) (
	*mcp.CallToolResult, MemoryLoadOutput, error,
) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, MemoryLoadOutput{}, fmt.Errorf("get working directory: %w", err)
	}

	source := input.Source
	if source == "" {
		source = "startup"
	}

	var b strings.Builder
	var loadedFiles []string
	const maxBytes = 16 * 1024

	// Always include MEMORY.md
	memoryPath := filepath.Join(projectDir, ".jikime", "memory", "MEMORY.md")
	if data, readErr := os.ReadFile(memoryPath); readErr == nil && len(data) > 0 {
		b.Write(data)
		b.WriteString("\n")
		loadedFiles = append(loadedFiles, ".jikime/memory/MEMORY.md")
	}

	// For "full" source, also include today's daily log
	if source == "full" {
		dailyName := time.Now().Format("2006-01-02") + ".md"
		dailyPath := filepath.Join(projectDir, ".jikime", "memory", dailyName)
		if data, readErr := os.ReadFile(dailyPath); readErr == nil && len(data) > 0 {
			b.WriteString("\n---\n\n")
			b.Write(data)
			b.WriteString("\n")
			loadedFiles = append(loadedFiles, ".jikime/memory/"+dailyName)
		}
	}

	content := b.String()
	if len(content) > maxBytes {
		content = content[:maxBytes] + "\n\n... (context truncated)\n"
	}

	return nil, MemoryLoadOutput{
		Content: content,
		Files:   loadedFiles,
	}, nil
}

func handleMemorySave(ctx context.Context, req *mcp.CallToolRequest, input MemorySaveInput) (
	*mcp.CallToolResult, MemorySaveOutput, error,
) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, MemorySaveOutput{}, fmt.Errorf("get working directory: %w", err)
	}

	// Validate type
	validTypes := map[string]bool{
		memory.TypeDecision:       true,
		memory.TypeLearning:       true,
		memory.TypeErrorFix:       true,
		memory.TypeToolUsage:      true,
		memory.TypeSessionSummary: true,
	}
	if !validTypes[input.Type] {
		return nil, MemorySaveOutput{
			Saved:   false,
			Message: fmt.Sprintf("invalid type %q; valid types: decision, learning, error_fix, tool_usage, session_summary", input.Type),
		}, nil
	}

	// Validate metadata is valid JSON if provided
	if input.Metadata != "" {
		var js json.RawMessage
		if err := json.Unmarshal([]byte(input.Metadata), &js); err != nil {
			return nil, MemorySaveOutput{
				Saved:   false,
				Message: "metadata must be valid JSON",
			}, nil
		}
	}

	// Write to daily log MD file
	entry := memory.DailyLogEntry{
		Type:     input.Type,
		Content:  input.Content,
		Metadata: input.Metadata,
	}
	relPath, err := memory.AppendDailyLog(projectDir, entry)
	if err != nil {
		return nil, MemorySaveOutput{}, fmt.Errorf("append daily log: %w", err)
	}

	// Index the updated file (non-fatal if indexing fails)
	store, err := memory.NewStore(projectDir)
	if err == nil {
		defer store.Close()

		cfg := memory.LoadEmbeddingConfig()
		provider, _ := memory.NewEmbeddingProvider(cfg)

		indexer := memory.NewIndexer(store, provider)
		indexCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_ = indexer.IndexFile(indexCtx, projectDir, relPath)
	}

	return nil, MemorySaveOutput{
		ID:      relPath,
		Saved:   true,
		Message: "memory saved to " + relPath,
	}, nil
}

func handleMemoryStats(ctx context.Context, req *mcp.CallToolRequest, input MemoryStatsInput) (
	*mcp.CallToolResult, MemoryStatsOutput, error,
) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, MemoryStatsOutput{}, fmt.Errorf("get working directory: %w", err)
	}

	store, err := memory.NewStore(projectDir)
	if err != nil {
		return nil, MemoryStatsOutput{}, fmt.Errorf("open memory store: %w", err)
	}
	defer store.Close()

	stats, err := store.GetStats(projectDir)
	if err != nil {
		return nil, MemoryStatsOutput{}, fmt.Errorf("get stats: %w", err)
	}

	chunkCount, _ := store.ChunkCount()
	fileCount, _ := store.FileCount()

	return nil, MemoryStatsOutput{
		TotalMemories:  stats.TotalMemories,
		TotalSessions:  stats.TotalSessions,
		TotalKnowledge: stats.TotalKnowledge,
		TotalChunks:    chunkCount,
		IndexedFiles:   fileCount,
		DBSizeBytes:    stats.DBSizeBytes,
		OldestMemory:   stats.OldestMemory,
		NewestMemory:   stats.NewestMemory,
	}, nil
}

func handleMemoryReindex(ctx context.Context, req *mcp.CallToolRequest, input MemoryReindexInput) (
	*mcp.CallToolResult, MemoryReindexOutput, error,
) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, MemoryReindexOutput{}, fmt.Errorf("get working directory: %w", err)
	}

	store, err := memory.NewStore(projectDir)
	if err != nil {
		return nil, MemoryReindexOutput{}, fmt.Errorf("open memory store: %w", err)
	}
	defer store.Close()

	cfg := memory.LoadEmbeddingConfig()
	provider, _ := memory.NewEmbeddingProvider(cfg)

	indexer := memory.NewIndexer(store, provider)
	indexCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := indexer.IndexAll(indexCtx, projectDir); err != nil {
		return nil, MemoryReindexOutput{}, fmt.Errorf("reindex: %w", err)
	}

	chunkCount, _ := store.ChunkCount()
	fileCount, _ := store.FileCount()

	return nil, MemoryReindexOutput{
		IndexedFiles: fileCount,
		TotalChunks:  chunkCount,
		Message:      fmt.Sprintf("Reindexed %d files, %d chunks total", fileCount, chunkCount),
	}, nil
}

// makeSnippet truncates text to maxLen characters for search result previews.
func makeSnippet(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
