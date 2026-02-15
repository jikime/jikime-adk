# JikiME-ADK Memory System

**Version**: 3.0.0
**Last Updated**: 2026-01-28
**Status**: Implementation Complete

---

## 1. Overview

The JikiME-ADK memory system provides context continuity across Claude Code sessions.
It implements a 2-Layer Memory Architecture natively in Go, using two storage types:

- **MD Files**: Human-readable daily logs
- **memories table**: Single source of truth for search and embeddings

### Core Principles

- **memories = Single Source for Search**: Session data (user_prompt, assistant_response, tool_usage) is stored in the memories table, serving as the single source for search and embeddings
- **MD = Human-Readable Logs**: `.jikime/memory/*.md` files are chronological logs for human reading
- **chunks = MD Index**: Derived index from splitting MD files into chunks
- **Dual-Source Search**: memory_search searches **both** chunks table and memories table, then merges results
- **Hybrid Scoring**: 0.7 x vector + 0.3 x text (BM25) weighted scoring
- **Deferred Embedding**: Only text is saved during sessions; batch embedding runs in background at SessionEnd

---

## 2. Architecture

### 2.1 2-Layer Memory

```
<projectDir>/
└── .jikime/
    └── memory/
        ├── MEMORY.md              # Layer 2: Long-term knowledge (curated project knowledge)
        ├── 2026-01-28.md          # Layer 1: Daily logs (chronological append)
        ├── 2026-01-27.md
        ├── memory.db              # SQLite (memories + chunks + embedding_cache)
        └── ...
```

**Layer 1: Daily Logs** (`YYYY-MM-DD.md`)
- Chronological append-only records
- Session data (user_prompt, assistant_response, tool_usage): Appended sequentially to end of file with timestamps
- Structural memory (decision, learning, error_fix): Grouped under `## Section` headings
- Automatically recorded by Hooks and MCP tools

**Layer 2: MEMORY.md**
- Curated permanent knowledge
- Project architecture, conventions, key decisions
- Manually edited by user or Claude

### 2.2 Data Flow

```
Write                                 Read
──────────────                        ──────────────
UserPromptSubmit hook ──┐
  (memory_prompt_save)  │             memory_search MCP tool
                        ├─→ memories DB   │
Stop hook ──────────────┤    + Daily MD ┌───┴───┐
  (memory_complete)     │               │       │
                        │               ▼       ▼
memory_save MCP tool ───┘          chunks    memories
  (decision/learning)    │         (MD index) (session data)
                         ▼              │       │
                   Daily Log MD         └───┬───┘
                   (.jikime/memory/         ▼
                    YYYY-MM-DD.md)    Score merge → Top N results
                         │
                    File Watcher
                    → Indexer → chunks DB

SessionEnd hook ────→ embed-backfill (background process)
  (memory_save)          → batch embedding for memories table
```

### 2.3 Search Flow (Dual-Source Merged Search)

```
memory_search(query)
    │
    ├─ 1) Search chunks table (SearchChunks)
    │     ├─ EmbedAndCache(query) → queryVec
    │     ├─ searchChunksVector(queryVec) — Cosine Similarity brute-force
    │     ├─ searchChunksFTS(query) — FTS5 MATCH (BM25)
    │     └─ mergeChunkResults() — 0.7×vec + 0.3×text
    │
    ├─ 2) Search memories table (SearchHybrid)
    │     ├─ searchVector(queryVec) — memories.embedding Cosine Similarity
    │     ├─ searchText(query) — memories FTS5/LIKE
    │     └─ mergeResults() — 0.7×vec + 0.3×text
    │
    └─ 3) Merge both results
          allResults = chunks results + memories results
          → Sort by score descending
          → Return top N results
```

---

## 3. MCP Tools

The MCP server (`jikime-adk mcp serve`) provides 6 tools:

| Tool | Description |
|------|-------------|
| `memory_search` | Dual-source hybrid search. Searches both chunks + memories and merges results |
| `memory_get` | Read specific line range from original MD file |
| `memory_load` | On-demand load of MEMORY.md + today's Daily Log |
| `memory_save` | Save structural memory to Daily Log MD + auto-indexing |
| `memory_stats` | Memory DB statistics (memory count, chunk count, file count, DB size) |
| `memory_reindex` | Reindex all MD files |

### 3.1 memory_search

Searches **both** chunks table (MD index) and memories table (session data), returning merged results by score.

```json
// Input
{
  "query": "authentication system JWT",
  "maxResults": 6,
  "minScore": 0.35,
  "type": "decision"
}

// Output
{
  "results": [
    {
      "path": ".jikime/memory/2026-01-27.md",
      "start_line": 15,
      "end_line": 22,
      "heading": "## Decision",
      "snippet": "Decided to resolve auth delay with JWT token caching...",
      "score": 0.82,
      "source": "chunks"
    },
    {
      "snippet": "Adopted JWT method for API authentication",
      "score": 0.75,
      "source": "memory"
    }
  ],
  "count": 2,
  "provider": "openai",
  "model": "text-embedding-3-small"
}
```

**Note**:
- `snippet` is a preview of up to 200 characters. Use `memory_get` for full content.
- When there are 0 results, `results` returns an empty array (`[]`), not `null`.

### 3.2 memory_get

Use `path`, `start_line`, `end_line` from `memory_search` results to read details from original MD:

```json
// Input
{
  "path": ".jikime/memory/2026-01-27.md",
  "from": 15,
  "lines": 8
}

// Output
{
  "path": ".jikime/memory/2026-01-27.md",
  "start_line": 15,
  "end_line": 22,
  "content": "## Decision\n\n- JWT token caching decision...\n..."
}
```

### 3.3 memory_load

Called at session start or when project context is needed:

```json
// Input
{ "source": "full" }    // "startup" = MEMORY.md only, "full" = MEMORY.md + today's Daily Log

// Output
{
  "content": "# Project Knowledge\n...\n---\n\n# 2026-01-28\n...",
  "files": [".jikime/memory/MEMORY.md", ".jikime/memory/2026-01-28.md"]
}
```

### 3.4 memory_save

Saves structural memory (decision, learning, error_fix, tool_usage) to Daily Log MD with auto-indexing:

```json
// Input
{
  "type": "decision",
  "content": "Adopted JWT method for API authentication",
  "metadata": "{\"context\": \"SPEC-001\"}"
}

// Output
{
  "id": ".jikime/memory/2026-01-28.md",
  "saved": true,
  "message": "memory saved to .jikime/memory/2026-01-28.md"
}
```

---

## 4. Hooks

Connected to Claude Code Hook events to automatically collect/preserve memory.

### 4.1 Hook Mapping

| Hook Event | Command | Role |
|------------|---------|------|
| **UserPromptSubmit** | `jikime-adk hooks memory-prompt-save` | Save user prompt to Daily Log + memories table |
| **PostToolUse** (Edit\|Write) | `jikime-adk hooks memory-track` | Add file modification record to track_buffer |
| **Stop** | `jikime-adk hooks memory-complete` | Save assistant_response + tool_usage to Daily Log + memories table |
| **PreCompact** | `jikime-adk hooks memory-flush` | Preserve important information to Daily Log before context compaction |
| **SessionEnd** | `jikime-adk hooks memory-save` | Execute background embed-backfill process |

### 4.2 In-Session Data Storage Flow

```
User enters prompt
    │
    ▼
UserPromptSubmit hook (memory-prompt-save)
    ├─ Append chronologically to Daily Log MD: "- [HH:MM:SS] **User Prompt**: ..."
    └─ Save to memories table (type: user_prompt, text only without embedding)

    ... Claude working ...

PostToolUse hook (memory-track) × N times
    └─ Append filename to track_buffer.jsonl (on each Edit|Write)

Claude response complete
    │
    ▼
Stop hook (memory-complete)
    ├─ Extract last assistant message from transcript
    ├─ Append chronologically to Daily Log MD: "- [HH:MM:SS] **Assistant Response**: ..."
    ├─ Save to memories table (type: assistant_response)
    ├─ Flush track_buffer → extract modified files list
    ├─ Append chronologically to Daily Log MD: "- [HH:MM:SS] **Tool Usage**: Files modified: ..."
    └─ Save to memories table (type: tool_usage)
```

### 4.3 Session End (SessionEnd)

At session end, only background embedding is triggered **without summary generation/saving**.
(Text data is already saved in UserPromptSubmit and Stop hooks)

```
Claude Code: /exit or session termination
    │
    ▼
hooks memory-save (SessionEnd hook)
    │
    ├─ Read session_id, cwd from stdin
    │
    └─ spawnEmbedBackfill()
         ├─ os.Executable() → current binary path
         ├─ exec.Command(exe, "hooks", "embed-backfill",
         │     "--project-dir", projectDir, "--session-id", sessionID)
         ├─ cmd.Start()
         └─ cmd.Process.Release() → detach (continues running after parent process terminates)
              │
              ▼
         embed-backfill (background process, 30-second timeout)
              ├─ Open Store
              ├─ Initialize Embedding Provider
              └─ BackfillMemoryEmbeddings(ctx, provider, projectDir, sessionID)
                   → Query records with embedding IS NULL from memories table
                   → Generate batch embeddings + save
```

**Why background**: The SessionEnd hook timeout in settings.json is 5 seconds.
If embedding API calls exceed 5 seconds, the process is killed and embeddings are lost.
The detached background process continues running for up to 30 seconds even after Claude Code terminates.

### 4.4 Context Compaction (PreCompact)

```
Claude Code: Context limit reached (75%+)
    │
    ▼
hooks memory-flush (PreCompact hook)
    ├─ 1) Parse Transcript JSONL
    ├─ 2) Extract important information (decisions, learnings, error_fixes)
    ├─ 3) Save to Daily Log MD
    ├─ 4) SQLite indexing (including embeddings)
    └─ 5) Flush track buffer (safety net)
```

---

## 5. Daily Log Format

`.jikime/memory/YYYY-MM-DD.md`:

Session data is appended **chronologically** to the end of the file, while structural memory is saved with **section grouping**.

```markdown
# 2026-01-28

- [10:30:15] **User Prompt**: Implement fsnotify-based watcher
- [10:30:45] **Assistant Response**: I will create the watcher.go file...
- [10:31:02] **Tool Usage**: Files modified: internal/memory/watcher.go
- [10:35:20] **User Prompt**: Check the build
- [10:35:35] **Assistant Response**: Build succeeded

## Decision

- Decided to resolve auth delay with JWT token caching. Redis TTL set to 5 minutes.

## Learning

- Static file bundling possible with Go 1.16+ embed pattern.

## Error Fix

- Panic occurred when user was nil in auth middleware. Resolved by adding nil check.
```

**Storage Rules**:
- `user_prompt`, `assistant_response`, `tool_usage` → chronological, `- [HH:MM:SS] **Type**: content`
- `decision`, `learning`, `error_fix`, `session_summary` → `- content` under `## Section`

---

## 6. Indexing Pipeline

### 6.1 Chunking

Split MD files into searchable chunks:

| Setting | Value |
|---------|-------|
| Max Tokens | ~400 tokens/chunk |
| Overlap | ~80 tokens |
| Min Chunk Size | 50 bytes |
| Split Method | Heading-aware (markdown heading boundaries prioritized) |

### 6.2 Embedding

| Provider | Model | Dimensions | Environment Variable |
|----------|-------|------------|---------------------|
| OpenAI | `text-embedding-3-small` | 1536 | `OPENAI_API_KEY` |
| Gemini | `text-embedding-004` | 768 | `GEMINI_API_KEY` |
| (None) | — | — | Text-only search fallback |

**Auto-detect order**: If `JIKIME_EMBEDDING_PROVIDER` environment variable is empty, use `auto`:
1. `OPENAI_API_KEY` → OpenAI
2. `GEMINI_API_KEY` → Gemini
3. Neither exists → `nil` (FTS5 text search only)

**Caching**: Cache in `embedding_cache` table based on `(content_hash, provider, model)`.
Identical text is not re-embedded.

**Deferred Embedding**: Only text is saved to memories table during sessions (no embedding).
At SessionEnd, background process (`embed-backfill`) batch processes unembedded records for that session.

### 6.3 File Watcher (Auto-indexing)

While MCP server is running, detects MD file changes in `.jikime/memory/` directory for auto-indexing:

- **Library**: `github.com/fsnotify/fsnotify`
- **Events**: Create, Write (`.md` files only)
- **Debounce**: 500ms (only last indexing on rapid consecutive writes)
- **Action**: Call `IndexFile()` (chunking + embedding + SQLite save)

### 6.4 SQLite Schema

```sql
-- Memories (single source for session data)
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    project_dir TEXT NOT NULL,
    type TEXT NOT NULL,           -- user_prompt, assistant_response, tool_usage, decision, ...
    content TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    metadata TEXT,
    embedding BLOB,               -- Deferred: batch generated at SessionEnd
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    accessed_at DATETIME,
    access_count INTEGER DEFAULT 0
);

-- Chunk index (derived from MD files)
CREATE TABLE chunks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL,           -- MD file relative path
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    text TEXT NOT NULL,
    hash TEXT NOT NULL,           -- SHA256 (change detection)
    heading TEXT DEFAULT '',      -- markdown heading
    embedding BLOB               -- float32 little-endian
);

-- File tracking (indexing status)
CREATE TABLE files (
    path TEXT PRIMARY KEY,
    hash TEXT NOT NULL,
    mtime INTEGER NOT NULL,
    size INTEGER NOT NULL,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Embedding cache
CREATE TABLE embedding_cache (
    content_hash TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    embedding BLOB NOT NULL,
    dims INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_hash, provider, model)
);
```

### 6.5 Vector Search Method

**Brute-force Cosine Similarity** (Go native):

```
SELECT id, ..., embedding FROM chunks/memories
  WHERE embedding IS NOT NULL LIMIT 1000
→ DecodeEmbedding(BLOB → []float32)
→ CosineSimilarity(queryVec, rowVec) for each row
→ Collect only those with score > 0
```

**Why sqlite-vec is not used**: `modernc.org/sqlite` (pure Go) cannot load C extensions.
Brute-force is sufficient at current memory scale (~1000 chunks).

---

## 7. Claude Code Integration

### 7.1 CLAUDE.md Memory-First Reasoning

Memory-First Reasoning instructions added as Section 14 in CLAUDE.md:

```
For every user message:
  1. Can the current session context sufficiently answer the question?
     → YES: Respond directly
     → NO or uncertain: Call memory_search first
  2. Is this the first message of a new session?
     → YES: Call memory_load(source: "full") first
```

**Reasoning-based** judgment, not keyword-based. Always search if past context might help.
Unnecessary searches (false positives) are acceptable; missing searches (false negatives) are not.

### 7.2 MCP Server Configuration

`.mcp.json` (project root):

```json
{
  "mcpServers": {
    "jikime-memory": {
      "command": "jikime-adk",
      "args": ["mcp", "serve"],
      "description": "Project memory search and management (hybrid vector + text)"
    }
  }
}
```

### 7.3 Embedding Environment Variables

Set in shell profile (`~/.zshrc` etc.):

```bash
# OpenAI (recommended)
export OPENAI_API_KEY="sk-..."

# Or Gemini
export GEMINI_API_KEY="AIza..."

# Optional: Specify provider directly
export JIKIME_EMBEDDING_PROVIDER="openai"   # openai, gemini, auto, none
export JIKIME_EMBEDDING_MODEL="text-embedding-3-small"
export JIKIME_EMBEDDING_BASE_URL=""
```

---

## 8. CLI Commands

```bash
# Memory search
jikime-adk memory search "authentication system"

# Memory list
jikime-adk memory list --project .

# Memory detail view
jikime-adk memory show <id>

# Memory delete
jikime-adk memory delete <id>

# Memory statistics
jikime-adk memory stats

# Clean up old memories
jikime-adk memory gc

# MD file indexing
jikime-adk memory index              # Full indexing
jikime-adk memory index --file path  # Single file

# Background embedding (internal use, auto-called at SessionEnd)
jikime-adk hooks embed-backfill --project-dir . --session-id abc123
```

---

## 9. Claude Code Built-in Features (No Implementation Needed)

The following features are handled by the Claude Code runtime itself and are not implemented in jikime-adk:

| Feature | Description |
|---------|-------------|
| **Compaction** | Summarize/compress old conversations when context window limit is reached |
| **Pruning** | In-memory reduction of old tool results (exec output, etc.) |
| **Cache-TTL Pruning** | Tool result cleanup for re-caching cost optimization after 5-minute cache expiration |

jikime-adk's role is to save important information to MD files **before** compaction (`memory-flush` hook).

---

## 10. Implementation File Structure

```
internal/memory/
├── store.go            # SQLiteStore (DB connection, CRUD)
├── schema.go           # Table creation, migrations
├── search.go           # FTS5/LIKE text search
├── hybrid.go           # Hybrid search merge (memories table)
├── chunk_search.go     # Chunk hybrid search (chunks table)
├── chunker.go          # MD → chunk splitting (heading-aware)
├── indexer.go          # Indexing pipeline (chunking + embedding + save)
├── watcher.go          # fsnotify file watching (auto-indexing)
├── embedding.go        # Embedding providers (OpenAI, Gemini)
├── embedding_cache.go  # Embedding cache (hash-based) + BackfillMemoryEmbeddings
├── mdwriter.go         # Daily Log MD file read/write (chronological + sectioned)
├── extractor.go        # Memory extraction from Transcript
├── injector.go         # additionalContext generation
├── transcript.go       # JSONL transcript parser
├── gc.go               # Garbage collection
├── track.go            # PostToolUse file modification tracking buffer
└── types.go            # Type definitions, constants

cmd/hookscmd/
├── memory_prompt_save.go  # UserPromptSubmit hook (save user_prompt)
├── memory_complete.go     # Stop hook (save assistant_response + tool_usage)
├── memory_save.go         # SessionEnd hook (run background embed-backfill)
├── memory_flush.go        # PreCompact hook
├── memory_track.go        # PostToolUse hook
├── memory_load.go         # SessionStart hook (no-op)
├── embed_backfill.go      # Background embedding subcommand (hidden)
└── hooks.go               # Hook command registration

cmd/mcpcmd/
├── mcp.go              # MCP command group
└── serve.go            # MCP server (6 tools + file watcher)

cmd/memorycmd/
├── memory.go           # CLI command group
├── search.go           # memory search
├── list.go             # memory list
├── show.go             # memory show
├── delete.go           # memory delete
├── stats.go            # memory stats
├── gc.go               # memory gc
├── index.go            # memory index
└── migrate.go          # memory migrate (v2→v3)
```

---

## 11. Design Decisions and Constraints

### 11.1 Key Design Decisions

| Decision | Reason |
|----------|--------|
| **memories as single source for search** | MD files are human logs; search is performed on memories table |
| **Dual-source merged search** | Search both chunks (MD index) and memories (session data) to prevent omissions |
| **Deferred embedding** | Save only text during session without API calls; batch process at SessionEnd |
| **Background embed-backfill** | Circumvent 5-second timeout constraint of SessionEnd hook |
| **Chronological Daily Log** | Record user_prompt/assistant_response in conversation flow order |
| **Memory-First Reasoning** | Reasoning-based memory usage judgment, not keyword triggers |
| **Removed validateEnvironment** | Session start performance optimization (removed unnecessary exec.LookPath) |

### 11.2 sqlite-vec Not Implemented

`modernc.org/sqlite` (pure Go) cannot load C extensions, making sqlite-vec unusable.
Switching to `mattn/go-sqlite3` (CGo) would enable it, but breaks cross-platform portability.
Brute-force search is sufficient at current memory scale (~1000 chunks + memories).

### 11.3 Data Volume Estimation

| Period | Chunk Count | Embedding Size | DB Size |
|--------|-------------|----------------|---------|
| 1 month | ~150-450 | ~1-3MB | ~3-8MB |
| 6 months | ~900-2,700 | ~5-16MB | ~18-50MB |
| 1 year | ~1,800-5,400 | ~11-32MB | ~35-100MB |

---

## Related Documents

- [Hooks System](./hooks.md) - Hook implementation details
- [Session Memory Flow](./session-memory-flow.md) - Complete data flow
- [Commands](./commands.md) - CLI command reference
