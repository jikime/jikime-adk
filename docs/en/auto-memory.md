# Auto-Memory

Cross-session project context injection via Claude Code's native memory system.

## Overview

Auto-Memory automatically discovers and injects project-specific memory files into Claude's context at every session start. This enables Claude to maintain persistent knowledge about your project across sessions — without any manual context-loading.

```
Session Start
    ↓
jikime-adk reads ~/.claude/projects/{hash}/memory/*.md
    ↓
Content injected into systemMessage
    ↓
Claude starts with full project context already loaded
```

## How It Works

### Path Discovery

Claude Code stores project memories at:

```
~/.claude/projects/{path-hash}/memory/
```

Where `{path-hash}` is the project path with `/` replaced by `-`:

```
/Users/foo/myproject  →  -Users-foo-myproject
```

jikime-adk uses the `cwd` value from Claude Code's stdin payload (not `os.Getwd()`) to compute the correct hash — ensuring the path always matches what Claude Code uses internally.

### Session Start Flow

```
Claude Code starts session
    ↓
Sends JSON to stdin: {"cwd": "/your/project", "session_id": "..."}
    ↓
jikime hooks session-start reads cwd
    ↓
ensureMemoryDir() — creates directory if missing
    ↓
discoverAutoMemory() — reads all .md files
    ↓
formatMemorySection() — builds systemMessage section
    ↓
Returns: {"continue": true, "systemMessage": "...Auto-Memory Loaded..."}
```

### Output Example

```
🚀 JikiME-ADK Session Started
   📦 Version: 1.0.0
   🔄 Changes: 5 file(s) modified
   🌿 Branch: master
   ...

---
📚 **Auto-Memory Loaded**
   📁 Path: /Users/foo/.claude/projects/-Users-foo-myproject/memory
   📄 Files: 2 (3104 bytes)

### MEMORY.md
# My Project Memory

## Architecture
- Next.js 16 migration from legacy PHP
...

### lessons.md
## Lessons Learned
- Always run `pnpm build` before committing
...
---
```

## Memory File Conventions

### Priority Order

Files are loaded and displayed in this order:

| Priority | Filename | Max Length | Purpose |
|----------|----------|------------|---------|
| 1st | `MEMORY.md` | 800 chars | Main project memory |
| 2nd | `lessons.md` | 800 chars | Lessons learned |
| 3rd | `context.md` | 800 chars | Current context |
| Others | `*.md` | 400 chars | Topic-specific notes |

### Recommended Structure for MEMORY.md

```markdown
# Project Memory

## Architecture
- Brief description of tech stack and structure

## Key Decisions
- Important decisions made and why

## Patterns & Conventions
- Code patterns, naming conventions

## Recent Work
- Summary of recent changes
```

## Who Writes the Memory Files?

**Claude writes them** — using the `Write` and `Edit` tools.

Claude Code's system prompt instructs Claude to save important information to the memory directory. You can trigger this explicitly:

```
"Remember the current project structure for next session"
"Save what we discussed about the API design to memory"
"Update MEMORY.md with today's decisions"
```

You can also write files manually:

```bash
# Create or edit directly
vim ~/.claude/projects/{hash}/memory/MEMORY.md
```

## Configuration

No configuration needed. Auto-Memory activates automatically when:

1. jikime-adk is installed (`go install .` or install script)
2. The `SessionStart` hook is registered in `.claude/settings.json`
3. A Claude Code session is started in a project with jikime-adk initialized

### Verify Hook Registration

```bash
cat .claude/settings.json | grep -A5 "SessionStart"
```

Expected output:
```json
"SessionStart": [
  {
    "hooks": [
      {
        "type": "command",
        "command": "jikime hooks session-start"
      }
    ]
  }
]
```

## Testing

### CLI Test

```bash
echo '{"cwd":"/your/project/path"}' | jikime-adk hooks session-start | python3 -m json.tool
```

Check that `systemMessage` contains `Auto-Memory Loaded` when `.md` files exist in the memory directory.

### Verify Memory Directory

```bash
# Find your project hash
PROJECT_HASH=$(echo "/your/project/path" | sed 's|/|-|g')
ls ~/.claude/projects/${PROJECT_HASH}/memory/
```

### End-to-End Test

1. Create a test memory file:
   ```bash
   echo "# Test" > ~/.claude/projects/{hash}/memory/MEMORY.md
   ```
2. Start a new Claude Code session in the project
3. Ask Claude: *"What's in my session system message?"*
4. Verify MEMORY.md content appears in the response

## Troubleshooting

### Memory not appearing

| Cause | Solution |
|-------|----------|
| Memory directory is empty | Ask Claude to write something to MEMORY.md |
| Wrong project hash | Verify `cwd` in Claude Code matches expected path |
| Old binary (< 1.0.0) | Run `go install github.com/jikime/jikime-adk@latest` |
| Hook not registered | Run `jikime-adk init` to re-install templates |

### Content truncated

Files over 800 chars (MEMORY.md) or 400 chars (others) are truncated. Keep memory files concise or split into multiple topic files.

## Related

- [Hooks System](./hooks.md) — Full hook system documentation
- [Session Start Hook](./hooks.md#session-start) — session-start hook details
