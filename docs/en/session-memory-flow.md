# Session Memory Complete Flow

**Version**: 3.0.0
**Last Updated**: 2026-01-28
**Status**: Implementation Complete

---

## 1. Session Start (SessionStart)

```
Claude Code starts
    │
    ▼
hooks session-start
    │  → findProjectRoot() → Search for .jikime directory
    │  → loadConfig() → Load 4 YAML configs (user, language, project, git-strategy)
    │  → Initialize orchestrator state (only when state file doesn't exist)
    │  → getGitInfoParallel() → Execute 3 git commands in parallel (branch/status/log)
    │  → formatSessionOutput() → Format session information
    │
    ▼
hooks memory-load (no-op)
    │  → SessionStart hook returns empty output
    │  → Context loading is delegated to memory_load MCP tool
    │
    ▼
Claude calls memory_load MCP tool (CLAUDE.md Memory-First rule)
    │
    ├─ source="startup" → Read MEMORY.md
    ├─ source="full"    → Read MEMORY.md + Today's Daily Log
    │
    ▼
Project context delivered to Claude
```

---

## 2. User Prompt Save (UserPromptSubmit)

```
User enters prompt
    │
    ▼
hooks memory-prompt-save (UserPromptSubmit hook)
    │
    ├─ 1) Read prompt text from stdin
    │
    ├─ 2) Append to Daily Log MD in chronological order (AppendDailyLog)
    │     .jikime/memory/YYYY-MM-DD.md
    │     → "- [HH:MM:SS] **User Prompt**: prompt content"
    │
    ├─ 3) Save to memories table (SaveIfNew)
    │     type: "user_prompt"
    │     content: prompt text
    │     embedding: NULL (Deferred — batch generated at SessionEnd)
    │
    └─ 4) Output JSON response (stdout)
```

---

## 3. Memory Search (memory_search — Dual-Source Merged)

```
Claude: calls memory_search MCP tool
    │  (Memory-First Reasoning: determined that past context is needed)
    │
    ▼
handleMemorySearch()
    │
    ├─ 1) Search chunks table (SearchChunks)
    │     ├─ EmbedAndCache(query) → queryVec
    │     │     ├─ Cache hit? → Return immediately
    │     │     └─ Cache miss? → API call → Save to cache
    │     │
    │     ├─ searchChunksVector(queryVec)
    │     │     SELECT embedding FROM chunks
    │     │       WHERE embedding IS NOT NULL LIMIT 1000
    │     │     → Go CosineSimilarity() brute-force
    │     │
    │     ├─ searchChunksFTS(query)
    │     │     FTS5 MATCH or LIKE fallback
    │     │     → BM25 scoring
    │     │
    │     └─ mergeChunkResults()
    │           finalScore = 0.7 × vecScore + 0.3 × textScore
    │
    ├─ 2) Search memories table (SearchHybrid)
    │     ├─ searchVector(queryVec) — memories.embedding Cosine Similarity
    │     ├─ searchText(query) — memories FTS5/LIKE
    │     └─ mergeResults() — 0.7×vec + 0.3×text
    │
    └─ 3) Merge results from both sources
          allResults = make([]memorySearchResult, 0)  ← empty array, not nil
          allResults = chunks results + memories results
          → Sort by score descending with sort.Slice()
          → Return top N results (default 6)

    │
    ▼
Result: [{path, start_line, end_line, snippet, score, source:"chunks"|"memory"}, ...]
    │  snippet = max 200 char preview (token saving)
    │
    ▼
(If needed) call memory_get MCP tool — read full details from original MD after checking snippet
    ├─ path: ".jikime/memory/2026-01-28.md"
    ├─ from: start_line
    ├─ lines: end_line - start_line
    └─ → Return line range from original MD file
```

---

## 4. Tool Usage Tracking (PostToolUse)

```
Claude: Edit("src/auth/jwt.go", ...)
    │
    ▼
hooks memory-track (PostToolUse, matcher: Edit|Write)
    │
    ├─ tool_name == "Edit" or "Write"?
    │     ├─ YES → Append to JSONL buffer
    │     │        .jikime/memory/track_buffer.jsonl:
    │     │        {"session_id":"abc","file":"src/auth/jwt.go","tool":"Edit","ts":"..."}
    │     │
    │     └─ NO → Return immediately
    │
    ▼
suppressOutput: true (file append only without DB access — completes within 3 seconds)
```

---

## 5. Response Complete (Stop)

```
Claude response complete
    │
    ▼
hooks memory-complete (Stop hook)
    │
    ├─ 1) Extract last assistant message from transcript
    │
    ├─ 2) Append to Daily Log MD in chronological order (AppendDailyLog)
    │     → "- [HH:MM:SS] **Assistant Response**: response content (truncated)"
    │
    ├─ 3) Save to memories table (SaveIfNew)
    │     type: "assistant_response"
    │     embedding: NULL (Deferred)
    │
    ├─ 4) FlushTrack() — Read + delete track_buffer.jsonl
    │     [jwt.go, middleware.go, auth_test.go, ...]
    │
    ├─ 5) Deduplicate files
    │     jwt.go, middleware.go, auth_test.go (3 files)
    │
    ├─ 6) Append to Daily Log MD in chronological order (AppendDailyLog)
    │     → "- [HH:MM:SS] **Tool Usage**: Files modified: jwt.go, middleware.go, auth_test.go"
    │
    └─ 7) Save to memories table (SaveIfNew)
          type: "tool_usage"
          embedding: NULL (Deferred)
```

---

## 6. Memory Save (memory_save MCP tool)

```
Claude: "Remember this decision"
    │
    ▼
memory_save MCP tool call
    │
    ├─ type: "decision"
    ├─ content: "JWT token caching decision..."
    ├─ metadata: "{...}" (optional)
    │
    ├─ 1) Append to Daily Log MD (AppendDailyLog)
    │     .jikime/memory/YYYY-MM-DD.md
    │     → Append "- content" to ## Decision section
    │
    ├─ 2) SQLite indexing (IndexFile)
    │     → Chunking (~400 tokens)
    │     → Generate embeddings (OpenAI/Gemini)
    │     → Save to chunks table
    │
    └─ 3) File Watcher detects (backup)
          → Re-index after 500ms debounce
          → Skip if duplicate since already indexed by save
```

---

## 7. Context Compaction (PreCompact)

```
Claude Code: Context limit reached (75%+)
    │
    ▼
hooks memory-flush (PreCompact hook)
    │
    ├─ 1) Parse Transcript JSONL (ParseTranscript)
    │
    ├─ 2) Extract important content (Extract)
    │     → decision, learning, error_fix, etc.
    │
    ├─ 3) Save to Daily Log MD (AppendDailyLog)
    │     → For each extracted item by type:
    │       Structured memory → append to ## Section
    │       Session data → append to end of file in chronological order
    │
    ├─ 4) SQLite indexing (IndexFile + embeddings)
    │
    ├─ 5) FlushTrack() — track_buffer safety net
    │
    ▼
Claude Code: Execute Compaction
    → Summarize/compress old conversations
    → Important information preserved in MD + SQLite
```

---

## 8. Session End (SessionEnd)

Summary generation/saving is **not performed** at session end.
Text data is already saved by UserPromptSubmit (memory-prompt-save) and Stop (memory-complete).
SessionEnd only **triggers background embedding**.

```
Claude Code: /exit or session termination
    │
    ▼
hooks memory-save (SessionEnd hook)
    │
    ├─ 1) Read session_id, cwd from stdin
    │     memorySaveInput { session_id, cwd }
    │
    ├─ 2) Determine projectDir (cwd or os.Getwd())
    │
    └─ 3) spawnEmbedBackfill(projectDir, sessionID)
         │
         ├─ os.Executable() → Current binary path
         ├─ exec.Command(exe, "hooks", "embed-backfill",
         │     "--project-dir", projectDir,
         │     "--session-id", sessionID)
         ├─ cmd.Stdin = nil
         ├─ cmd.Stdout = nil
         ├─ cmd.Stderr = os.Stderr  ← Debug logs
         ├─ cmd.Start()
         └─ cmd.Process.Release()  ← Detach (continues running after parent terminates)
              │
              ▼
         embed-backfill (independent background process)
              │
              ├─ context.WithTimeout(30 seconds)
              ├─ memory.NewStore(projectDir) → Open SQLite
              ├─ memory.LoadEmbeddingConfig() → Read environment variables
              ├─ memory.NewEmbeddingProvider(cfg)
              │
              └─ store.BackfillMemoryEmbeddings(ctx, provider, projectDir, sessionID)
                   │
                   ├─ SELECT id, content FROM memories
                   │    WHERE embedding IS NULL
                   │      AND session_id = ? AND project_dir = ?
                   │
                   ├─ for each unembedded memory:
                   │     embedding = provider.Embed(content)
                   │     UPDATE memories SET embedding = ? WHERE id = ?
                   │
                   └─ return count, nil
```

**Why this architecture**:
- settings.json SessionEnd hook timeout: 5000ms (5 seconds)
- Single embedding API call: ~200-500ms, memories accumulated during session: ~10-50 items
- Total time required: ~2-25 seconds → May exceed 5 second timeout
- With `cmd.Process.Release()` detach, continues running for up to 30 seconds after Claude Code exits

---

## 9. Embedding Fallback Chain

```
LoadEmbeddingConfig()
    │
    ├─ JIKIME_EMBEDDING_PROVIDER environment variable?
    │     ├─ "openai" → OpenAI (text-embedding-3-small, 1536 dims)
    │     ├─ "gemini" → Gemini (text-embedding-004, 768 dims)
    │     ├─ "none"   → Disable embeddings
    │     └─ "auto" or "" → Auto-detect below
    │
    ├─ Auto-detect:
    │     ├─ OPENAI_API_KEY exists? → Use OpenAI
    │     ├─ GEMINI_API_KEY exists? → Use Gemini
    │     └─ Neither exists?        → provider=nil
    │
    ▼
When provider == nil:
    ├─ Skip embedding generation
    ├─ Skip vector search
    └─ Use FTS5 text search only (or LIKE fallback)
```

---

## 10. File Watcher (Auto Indexing)

```
MCP server starts (jikime-adk mcp serve)
    │
    ├─ Open Store
    ├─ Initialize Embedding Provider
    ├─ Start WatchMemoryFiles() goroutine
    │     │
    │     ├─ fsnotify.NewWatcher()
    │     ├─ Watch .jikime/memory/ directory
    │     │
    │     └─ Event loop:
    │           ├─ Detect Create/Write .md file
    │           ├─ 500ms debounce (per-file timer)
    │           └─ Call IndexFile()
    │                 ├─ Chunking
    │                 ├─ Generate embeddings
    │                 └─ Save to SQLite chunks table
    │
    └─ Run MCP server (STDIO transport)
        └─ cancel() on server shutdown → cleanup watcher
```

---

## 11. Timeline Summary

```
Session start ──→ session-start hook (git info + config + orchestrator)
              memory-load hook (no-op)
              Claude calls memory_load MCP tool (Memory-First)
    │
User input ──→ memory-prompt-save hook (user_prompt → Daily MD + memories DB)
              Claude calls memory_search MCP tool (if needed)
              → Read original MD details with memory_get
    │
Claude tools ──→ memory-track hook (record file modifications, repeat)
    │
Claude save ──→ memory_save MCP tool (structured memory → Daily MD + chunks indexing)
    │
Response complete ──→ memory-complete hook (assistant_response + tool_usage → Daily MD + memories DB)
    │
Context compaction ──→ memory-flush hook (transcript → Daily MD + indexing)
    │
Session end ──→ memory-save hook (background embed-backfill execution)
              → memories table batch embedding (30 second timeout)
```

---

## 12. DB Schema (v3 — 2-Layer Architecture)

```
memories (single source for session data — search/embedding target)
├─ id TEXT PK
├─ session_id TEXT              ← Session identifier
├─ project_dir TEXT             ← Project path
├─ type TEXT                    ← user_prompt, assistant_response, tool_usage, decision, ...
├─ content TEXT                 ← Memory text
├─ content_hash TEXT            ← SHA256 (duplicate prevention)
├─ metadata TEXT                ← JSON metadata
├─ embedding BLOB               ← Deferred: batch generated at SessionEnd
├─ created_at DATETIME
├─ accessed_at DATETIME
└─ access_count INTEGER

chunks (index derived from MD files)
├─ id INTEGER PK AUTOINCREMENT
├─ path TEXT                    ← MD file relative path
├─ start_line INTEGER
├─ end_line INTEGER
├─ text TEXT
├─ hash TEXT                    ← SHA256 (change detection)
├─ heading TEXT                 ← Markdown heading
└─ embedding BLOB               ← float32 little-endian

files (indexing status tracking)
├─ path TEXT PK
├─ hash TEXT
├─ mtime INTEGER
├─ size INTEGER
└─ indexed_at DATETIME

embedding_cache (embedding cache)
├─ content_hash TEXT  ─┐
├─ provider TEXT       ├─ Composite PK
├─ model TEXT         ─┘
├─ embedding BLOB
├─ dims INTEGER
└─ created_at DATETIME
```
