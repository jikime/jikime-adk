# JikiME-ADK Sync Workflow Implementation Document

## Overview

The `/jikime:3-sync` command is a workflow that synchronizes code changes with documentation, verifies quality, and performs Git operations.

**Core Philosophy**: "Sync to Verify to Commit"

## Implemented File List

### 1. Command

| File | Path | Description |
|------|------|-------------|
| `3-sync.md` | `templates/.claude/commands/jikime/` | Main Sync Command (Step 5/5) |

### 2. Context

| File | Path | Description |
|------|------|-------------|
| `sync.md` | `templates/.claude/contexts/` | Sync Mode Behavior Rules Definition |

### 3. Agents

| File | Path | Description |
|------|------|-------------|
| `manager-docs.md` | `templates/.claude/agents/jikime/` | Documentation Synchronization Specialist Agent |
| `manager-quality.md` | `templates/.claude/agents/jikime/` | Quality Verification Specialist Agent |
| `manager-git.md` | `templates/.claude/agents/jikime/` | Git Operations Specialist Agent |

### 4. Skill

| File | Path | Description |
|------|------|-------------|
| `SKILL.md` | `templates/.claude/skills/jikime-workflow-sync/` | Sync Workflow Skill |

---

## Architecture

### Workflow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    /jikime:3-sync                           │
│                   (Main Entry Point)                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 0.5: Quality Verification                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Tests     │  │   Linter    │  │ Type Check  │         │
│  │   pytest    │  │   ruff      │  │   mypy      │         │
│  │   vitest    │  │   eslint    │  │   tsc       │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────┘
                      │ PASS
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 1: Analysis & Planning                   │
│                                                             │
│  git diff --name-only HEAD                                  │
│  git status --porcelain                                     │
│                     ↓                                       │
│  Documentation Mapping (Change Type → Document Update)      │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 2: Execute Sync                          │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                  manager-docs                          │ │
│  │  • README.md synchronization                           │ │
│  │  • CODEMAP generation/update                           │ │
│  │  • SPEC status synchronization                         │ │
│  │  • API documentation                                   │ │
│  └───────────────────────────────────────────────────────┘ │
│                          ↓                                  │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                 manager-quality                        │ │
│  │  • TRUST 5 verification                                │ │
│  │  • Link integrity check                                │ │
│  │  • Consistency check                                   │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 3: Git Operations                        │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                  manager-git                           │ │
│  │  • Stage documentation files                           │ │
│  │  • Create commit (HEREDOC message)                     │ │
│  │  • PR management (Team Mode)                           │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Agent Collaboration Structure

```
┌──────────────────────────────────────────────────────────────┐
│                     Sync Orchestrator                        │
│                    (/jikime:3-sync)                          │
└──────────────────────────┬───────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
           ▼               ▼               ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ manager-docs │  │manager-quality│  │ manager-git  │
│              │  │              │  │              │
│ Tools:       │  │ Tools:       │  │ Tools:       │
│ • Read       │  │ • Read       │  │ • Bash       │
│ • Write      │  │ • Bash       │  │ • Read       │
│ • Edit       │  │ • Grep       │  │ • Grep       │
│ • Bash       │  │ • Glob       │  │ • TodoWrite  │
│ • Grep       │  │ • TodoWrite  │  │              │
│ • Glob       │  │              │  │              │
│ • TodoWrite  │  │              │  │              │
└──────────────┘  └──────────────┘  └──────────────┘
```

---

## Component Details

### 1. `/jikime:3-sync` Command

**Location**: `templates/.claude/commands/jikime/3-sync.md`

**Core Configuration**:
```yaml
---
description: "[Step 5/5] Sync docs, verify quality, commit changes."
context: sync
---
```

**Execution Modes**:

| Mode | Flag | Description |
|------|------|-------------|
| Auto | (default) | Synchronize only changed files |
| Full | `--full` | Regenerate all documentation |
| Status | `--status` | Check status only (read-only) |

**Options**:

| Option | Description |
|--------|-------------|
| `--skip-quality` | Skip Phase 0.5 quality verification |
| `--commit` | Auto-commit changes |
| `--worktree` | Run in Worktree environment |

---

### 2. `sync.md` Context

**Location**: `templates/.claude/contexts/sync.md`

**Mode Definition**:
```yaml
Mode: Documentation Synchronization
Focus: Document generation, quality verification, git operations
Methodology: Sync to Verify to Commit
```

**Behavior Rules**:

| DO | DON'T |
|----|-------|
| Synchronize after git change analysis | Regenerate unchanged documents |
| Update only affected documents | Skip quality verification |
| Verify link integrity | Commit without review |
| Follow TRUST 5 principles | Generate unnecessary documents |
| Use meaningful commit messages | Break existing links |

---

### 3. manager-docs Agent

**Location**: `templates/.claude/agents/jikime/manager-docs.md`

**Role**: Technical Writer & Documentation Architect

**Document Types**:

1. **README.md**
   - Project overview
   - Quick Start guide
   - Architecture reference

2. **CODEMAPS**
   ```
   docs/CODEMAPS/
   ├── INDEX.md       # Architecture overview
   ├── frontend.md    # Frontend structure
   ├── backend.md     # Backend structure
   └── database.md    # Database schema
   ```

3. **SPEC Status Synchronization**
   - Status: Planning → In Progress → Completed
   - Progress percentage
   - Last Updated timestamp

4. **API Documentation**
   - Endpoint descriptions
   - Parameter/response schemas

---

### 4. manager-quality Agent

**Location**: `templates/.claude/agents/jikime/manager-quality.md`

**Role**: Quality Assurance Architect

**TRUST 5 Framework**:

| Principle | Verification Items |
|-----------|-------------------|
| **T**ested | Test coverage ≥80%, all tests pass |
| **R**eadable | Functions <50 lines, files <400 lines, nesting <4 |
| **U**nified | Consistent style, standard patterns, no duplication |
| **S**ecured | No hardcoded secrets, input validation, injection prevention |
| **T**rackable | Meaningful commits, SPEC tracking, timestamps |

**Approval Criteria**:

| Status | Condition |
|--------|-----------|
| APPROVE | No CRITICAL/HIGH issues |
| WARNING | Only MEDIUM issues |
| BLOCK | CRITICAL/HIGH issues present |

---

### 5. manager-git Agent

**Location**: `templates/.claude/agents/jikime/manager-git.md`

**Role**: Version Control Specialist

**Commit Message Template**:
```
docs: sync documentation with code changes

Synchronized:
- [List of updated documents]

SPEC updates:
- [SPEC status changes]

Quality verification:
- Tests: PASS
- Linter: PASS

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

**Workflow Modes**:

| Mode | Description |
|------|-------------|
| Personal | Single branch direct commit |
| Team | PR-based branch workflow |

**Safety Rules**:

Prohibited:
- `git push --force` (main/master)
- `git reset --hard` (without confirmation)
- `git checkout .` (discard all changes)
- Commit secrets/credentials
- Skip pre-commit hooks

Required:
- Review changes before commit
- Meaningful commit messages
- Test before push
- New commit instead of amend (except when requested)

---

### 6. jikime-workflow-sync Skill

**Location**: `templates/.claude/skills/jikime-workflow-sync/SKILL.md`

**Triggers**:
```yaml
triggers:
  keywords: ["sync", "synchronization", "documentation", "docs", "CODEMAP", "README", "commit"]
  phases: ["sync"]
  agents: ["manager-docs", "manager-quality", "manager-git"]
```

**Allowed Tools**:
- Read, Write, Edit
- Bash, Grep, Glob
- TodoWrite

---

## Usage Examples

### Basic Usage

```bash
# Synchronize only changed files (Auto mode)
/jikime:3-sync

# Regenerate all documentation (Full mode)
/jikime:3-sync --full

# Check status only (Status mode)
/jikime:3-sync --status
```

### Advanced Usage

```bash
# Skip quality verification
/jikime:3-sync --skip-quality

# Include auto-commit
/jikime:3-sync --commit

# Run in Worktree environment
/jikime:3-sync --worktree

# Combined usage
/jikime:3-sync --full --commit
```

### Workflow Integration

```bash
# Full development workflow
/jikime:dev-0-init          # Step 1: Project initialization
/jikime:dev-1-plan          # Step 2: SPEC-based planning
/jikime:dev-2-implement     # Step 3: Implementation
/jikime:dev-3-test          # Step 4: Testing
/jikime:3-sync              # Step 5: Documentation sync & commit
```

---

## Improvements Compared to moai-adk

| Item | moai-adk | jikime-adk-v2 |
|------|----------|---------------|
| Command file size | ~1400 lines | ~250 lines |
| Execution modes | 4 (auto, force, status, project) | 3 (auto, full, status) |
| Context loading | Manual reference | Auto loading (`context: sync`) |
| Agent structure | Defined in single file | Separate agent files |
| Skill system | None | Progressive Disclosure integration |
| CLI integration | Python (moai) | Go (jikime) |

### Key Improvement Points

1. **Conciseness**: Removed unnecessary duplication, focused on core logic
2. **Modularity**: Separated responsibilities by agent for better maintainability
3. **Automation**: Context auto-loading improves user experience
4. **Extensibility**: Skill system integration supports Progressive Disclosure
5. **Performance**: Go-based CLI for fast execution

---

## Related Files

### Integration Points

```
templates/.claude/
├── commands/jikime/
│   ├── 3-sync.md           ← Main command
│   ├── dev-0-init.md       ← Step 1
│   ├── dev-1-plan.md       ← Step 2
│   ├── dev-2-implement.md  ← Step 3
│   └── dev-3-test.md       ← Step 4
├── contexts/
│   ├── sync.md             ← Sync context
│   ├── dev.md              ← Dev context
│   └── review.md           ← Review context
├── agents/jikime/
│   ├── manager-docs.md     ← Docs agent
│   ├── manager-quality.md  ← Quality agent
│   └── manager-git.md      ← Git agent
└── skills/
    └── jikime-workflow-sync/
        └── SKILL.md        ← Sync skill
```

---

## Troubleshooting

### Problem: Documents not synchronizing

**Cause**: Change detection failure in git diff

**Solution**:
```bash
# Check changes
git status --porcelain
git diff --name-only HEAD

# Use full regeneration mode
/jikime:3-sync --full
```

### Problem: Quality verification failure

**Cause**: Test/linter/type check errors

**Solution**:
```bash
# Check specific failure details and fix
# Or skip (not recommended)
/jikime:3-sync --skip-quality
```

### Problem: Link integrity failure

**Cause**: References to moved/deleted files

**Solution**:
1. Identify broken link locations
2. Update references
3. Re-synchronize

---

## Version Information

- **Version**: 1.0.0
- **Created**: 2026-01-22
- **Last Updated**: 2026-01-22
- **Author**: Claude Opus 4.5

---

## Related Documents

- [Worktree Workflow](./worktree.md)
- [Ralph Loop](./ralph-loop.md)
- [Migration Guide](./migration.md)
