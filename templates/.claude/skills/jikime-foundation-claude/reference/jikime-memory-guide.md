# Project Memory (jikime-memory MCP) Guide

Detailed usage patterns for the jikime-memory MCP server.

## When to Search Memory

**ALWAYS search** when the user's message:
- Refers to anything outside the current session (past work, previous decisions)
- Asks about the user themselves (name, preferences, habits)
- Asks about project conventions, patterns, architecture, or history
- References something "we" did, decided, discussed, or built
- Asks to continue, resume, or pick up previous work
- Cannot be answered from current session context alone
- Uses recall words: remember, what was, how did we, show me again

**SKIP search** only when:
- Request is purely about the current session
- Generic coding question with no project context needed
- Already searched memory for same topic in this session

## Search Query Strategy

Extract **semantic intent**, not literal keywords:

```
"내 이름이 뭐야?" → memory_search(query: "user name personal information")
"DB 스키마 어떻게 설계했었지?" → memory_search(query: "database schema design architecture")
"그 버그 어떻게 고쳤더라?" → memory_search(query: "bug fix error resolution", type: "error_fix")
"auth 어떻게 하기로 했지?" → memory_search(query: "authentication design decision", type: "decision")
"어제 뭐 했지?" → memory_search(query: "yesterday work progress session summary")
```

## Tool Reference

### memory_search - Hybrid vector + text search
```
query: "descriptive search"   # Required: semantic search query
maxResults: 10                # Optional: default 6
minScore: 0.35                # Optional: minimum relevance score
type: "decision"              # Optional: decision|learning|error_fix|user_prompt|assistant_response
```
Returns: snippet (200 chars preview), path, start_line, end_line, score, source.
**Important**: Use `memory_get` with path/from/lines to read full content.

### memory_load - Load project knowledge
```
source: "full"    # "startup" = MEMORY.md only, "full" = MEMORY.md + today's daily log
```

### memory_save - Save structured memory
```
type: "decision"              # Required: decision | learning | error_fix | tool_usage
content: "Use JWT for auth"   # Required: memory content
metadata: "{}"                # Optional: JSON metadata
```

### memory_get - Read specific memory file
```
path: ".jikime/memory/2026-01-28.md"  # Relative file path
from: 10                              # Optional: start line (1-based)
lines: 20                             # Optional: number of lines
```

### memory_stats - Database statistics (no params)
### memory_reindex - Re-index all MD files (no params)

---

Version: 1.0.0
Source: Consolidated from CLAUDE.md Section 14
