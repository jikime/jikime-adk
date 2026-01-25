---
name: jikime-workflow-learning
description: Continuous learning system - extract, store, and reuse patterns from Claude Code sessions
version: 1.0.0
tags: ["workflow", "learning", "patterns", "knowledge", "session", "improvement"]
triggers:
  keywords: ["learn", "pattern", "lesson", "remember", "extract", "학습"]
  phases: ["sync"]
  agents: ["manager-docs", "manager-strategy"]
  languages: []
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~7000
user-invocable: false
context: fork
agent: manager-docs
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
  - TodoWrite
---

# Continuous Learning Skill

Automatically extract reusable patterns from Claude Code sessions and store them for future use.

## Philosophy

```
Every session is a learning opportunity:
├─ Error resolutions → Future prevention
├─ User corrections → Preference learning
├─ Workarounds → Knowledge base
├─ Debugging techniques → Reusable strategies
└─ Project-specific patterns → Team knowledge
```

---

## How It Works

### Session Lifecycle

```
Session Start
    ↓
  Load relevant learnings from .jikime/learnings/
    ↓
  Development work...
    ↓
Session End (Stop hook)
    ↓
  Analyze session for patterns
    ↓
  Extract learnings with confidence scoring
    ↓
  Store in .jikime/learnings/
    ↓
Next Session: Patterns available
```

### Automatic Extraction

At session end, the system analyzes:
- Error messages and resolutions
- User corrections to Claude's suggestions
- Workarounds for framework/library quirks
- Debugging techniques that worked
- Project-specific conventions

---

## Pattern Categories

### 1. Error Resolution

Captures how specific errors were resolved:

```yaml
category: error_resolution
pattern:
  trigger: "TypeError: Cannot read properties of undefined"
  context: "React component accessing props before mount"
  resolution: "Add null check or use optional chaining"
  example: |
    // Before
    const value = data.nested.property;

    // After
    const value = data?.nested?.property;
confidence: 0.92
frequency: 5
last_used: "2024-01-22"
```

### 2. User Corrections

Patterns from user corrections to Claude's output:

```yaml
category: user_correction
pattern:
  original: "Claude suggested inline styles"
  correction: "User preferred Tailwind CSS classes"
  learning: "This project uses Tailwind CSS for styling, avoid inline styles"
  scope: project
confidence: 0.88
frequency: 3
```

### 3. Workarounds

Solutions to framework/library quirks:

```yaml
category: workaround
pattern:
  technology: "Next.js 14"
  issue: "App Router dynamic imports with SSR"
  workaround: |
    Use 'use client' directive or dynamic import with ssr: false
    import dynamic from 'next/dynamic'
    const Component = dynamic(() => import('./Component'), { ssr: false })
confidence: 0.95
source: "official_docs"
```

### 4. Debugging Techniques

Effective debugging approaches:

```yaml
category: debugging
pattern:
  symptom: "State not updating in React"
  technique: "Check for object/array mutation instead of new reference"
  steps:
    1. "Verify state update uses new reference"
    2. "Check useEffect dependencies"
    3. "Look for direct mutation patterns"
  success_rate: 0.87
```

### 5. Project Conventions

Project-specific patterns:

```yaml
category: project_convention
pattern:
  domain: "API responses"
  convention: "All API responses follow { success, data, error } structure"
  example: |
    interface ApiResponse<T> {
      success: boolean;
      data?: T;
      error?: { code: string; message: string };
    }
  enforced: true
```

---

## Storage Structure

```
.jikime/
├── learnings/
│   ├── index.json              # Searchable index
│   ├── errors/
│   │   ├── typescript.yaml
│   │   ├── react.yaml
│   │   └── nextjs.yaml
│   ├── corrections/
│   │   └── style-preferences.yaml
│   ├── workarounds/
│   │   ├── nextjs-14.yaml
│   │   └── prisma.yaml
│   ├── debugging/
│   │   └── react-state.yaml
│   ├── conventions/
│   │   ├── api-patterns.yaml
│   │   └── file-structure.yaml
│   └── sessions/
│       ├── 2024-01-22-summary.md
│       └── 2024-01-21-summary.md
```

### Index Structure

```json
{
  "version": "1.0.0",
  "last_updated": "2024-01-22T15:30:00Z",
  "total_patterns": 47,
  "categories": {
    "error_resolution": 15,
    "user_correction": 8,
    "workaround": 12,
    "debugging": 7,
    "project_convention": 5
  },
  "top_patterns": [
    {"id": "err-ts-001", "confidence": 0.95, "frequency": 12},
    {"id": "wk-nextjs-003", "confidence": 0.93, "frequency": 8}
  ],
  "technologies": ["typescript", "react", "nextjs", "prisma"]
}
```

---

## Confidence Scoring

Patterns are scored for reliability:

```
Confidence = (
  base_score * 0.4 +
  frequency_score * 0.3 +
  recency_score * 0.2 +
  source_reliability * 0.1
)

Base Score:
  - Official docs: 1.0
  - User correction: 0.9
  - Successful resolution: 0.8
  - Experimental: 0.6

Frequency Score:
  - Used 10+ times: 1.0
  - Used 5-9 times: 0.8
  - Used 2-4 times: 0.6
  - Used once: 0.4

Recency Score:
  - Used this week: 1.0
  - Used this month: 0.8
  - Used this quarter: 0.6
  - Older: 0.4
```

### Confidence Thresholds

| Level | Score | Treatment |
|-------|-------|-----------|
| High | 0.85+ | Apply automatically |
| Medium | 0.65-0.84 | Suggest with context |
| Low | 0.40-0.64 | Available for search |
| Experimental | <0.40 | Flag for review |

---

## Orchestrator Integration

### J.A.R.V.I.S. (Development)

```
Session Start:
  → Load high-confidence patterns for active technologies
  → Summarize: "Loaded 12 patterns for React/TypeScript"

During Development:
  → Apply patterns proactively
  → "Based on learned pattern: using optional chaining here"

Session End:
  → Extract new patterns
  → Report: "3 new patterns learned this session"

Predictive Suggestions:
  → "Based on past sessions, you might also want to..."
```

### F.R.I.D.A.Y. (Migration)

```
Migration Start:
  → Load patterns for source/target frameworks
  → "Loaded 8 migration patterns for Vue → React"

During Migration:
  → Apply migration-specific workarounds
  → Track framework-specific quirks

Migration End:
  → Store migration patterns for future use
  → Export as reusable migration guide
```

---

## Session Summary

At session end, generate a summary:

```markdown
# Session Summary: 2024-01-22

## Duration
Started: 10:30 AM
Ended: 2:15 PM (3h 45m)

## Work Completed
- Implemented user authentication
- Fixed 3 TypeScript errors
- Resolved hydration mismatch issue

## Patterns Learned

### New Patterns (3)
1. **Error Resolution**: TypeScript strict null checks
   - Confidence: 0.85
   - Category: error_resolution

2. **Workaround**: Next.js 14 cache invalidation
   - Confidence: 0.78
   - Category: workaround

3. **Convention**: API response structure
   - Confidence: 0.92
   - Category: project_convention

### Reinforced Patterns (2)
- React useState with objects (frequency: 5 → 6)
- Prisma relation queries (frequency: 3 → 4)

## For Next Session
- Continue with payment integration
- Review auth edge cases
- Consider adding rate limiting
```

---

## Export/Import

### Export Patterns

Share learnings between projects:

```bash
# Export all patterns
jikime-adk learnings export --output learnings-export.yaml

# Export specific category
jikime-adk learnings export --category workaround --output workarounds.yaml

# Export high-confidence only
jikime-adk learnings export --min-confidence 0.85 --output reliable-patterns.yaml
```

### Import Patterns

```bash
# Import from another project
jikime-adk learnings import --source ../other-project/.jikime/learnings/

# Import with merge strategy
jikime-adk learnings import --source patterns.yaml --strategy merge

# Import with confidence adjustment
jikime-adk learnings import --source patterns.yaml --confidence-penalty 0.1
```

### Export Format

```yaml
export_version: "1.0"
exported_at: "2024-01-22T15:30:00Z"
source_project: "auth-service"
patterns:
  - id: "wk-nextjs-001"
    category: "workaround"
    confidence: 0.93
    pattern:
      technology: "Next.js 14"
      issue: "Dynamic routes with middleware"
      solution: "..."
    metadata:
      created: "2024-01-15"
      frequency: 8
```

---

## Hook Integration

### Session End Hook

```json
{
  "hooks": {
    "Stop": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hooks learning-extract"
          }
        ]
      }
    ]
  }
}
```

### Session Start Hook

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hooks learning-load"
          }
        ]
      }
    ]
  }
}
```

---

## Configuration

```yaml
# .jikime/config/learning.yaml

learning:
  enabled: true

  # Extraction settings
  extraction:
    min_session_length: 10          # Minimum messages to analyze
    auto_extract: true              # Extract on session end
    require_confirmation: false     # Ask before saving patterns

  # Pattern settings
  patterns:
    min_confidence: 0.40            # Minimum to store
    auto_apply_threshold: 0.85      # Apply without asking
    max_age_days: 365               # Archive old patterns

  # Categories to track
  categories:
    - error_resolution
    - user_correction
    - workaround
    - debugging
    - project_convention

  # Ignore patterns
  ignore:
    - simple_typos
    - one_time_fixes
    - external_api_issues
    - environment_specific
```

---

## Searching Patterns

Query stored patterns:

```bash
# Search by keyword
jikime-adk learnings search "useState"

# Search by category
jikime-adk learnings search --category workaround

# Search by technology
jikime-adk learnings search --tech nextjs

# Full-text search
jikime-adk learnings search "hydration mismatch react"
```

### Search Output

```markdown
## Search Results: "hydration mismatch"

### 1. React Hydration Mismatch Fix
**Category**: error_resolution
**Confidence**: 0.91
**Frequency**: 7

**Pattern**:
When encountering hydration mismatch in Next.js:
1. Check for browser-only APIs (window, localStorage)
2. Use useEffect for client-side only code
3. Consider dynamic import with ssr: false

### 2. Dynamic Import Workaround
**Category**: workaround
**Confidence**: 0.88
**Frequency**: 5

**Pattern**:
```jsx
import dynamic from 'next/dynamic'
const Component = dynamic(() => import('./Component'), { ssr: false })
```
```

---

## Privacy & Security

### Sensitive Data Handling

```yaml
# Patterns never stored:
- API keys, tokens, secrets
- Passwords or credentials
- Personal information
- Environment-specific values

# Before storage:
- Redact secrets: sk-*** → [REDACTED]
- Generalize specific values
- Remove project-specific paths
```

### Local Storage Only

```
All learnings stored locally in .jikime/learnings/
- Not synced to cloud by default
- Export explicitly for sharing
- Add to .gitignore if sensitive
```

---

## Best Practices

### DO

1. **Review high-frequency patterns** - They shape future behavior
2. **Adjust confidence when wrong** - Learning improves over time
3. **Export valuable patterns** - Share across projects
4. **Clean stale patterns** - Remove outdated learnings
5. **Categorize correctly** - Aids future retrieval

### DON'T

1. **Trust low-confidence blindly** - Verify before applying
2. **Store one-time fixes** - Not reusable
3. **Keep outdated patterns** - Technology evolves
4. **Ignore user corrections** - They signal preferences
5. **Over-generalize** - Some patterns are context-specific

---

## Works Well With

- `jikime-foundation-core`: Core workflow integration
- `jikime-workflow-spec`: SPEC-based development
- `jikime-workflow-eval`: Evaluation framework
- `jikime-workflow-project`: Project initialization
- `jikime-foundation-claude`: Claude Code patterns

---

Last Updated: 2026-01-25
Version: 1.0.0
Integration: SessionEnd hook, J.A.R.V.I.S./F.R.I.D.A.Y., Export/Import
