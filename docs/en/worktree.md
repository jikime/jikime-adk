# JikiME-ADK Worktree Management

This feature enables parallel SPEC development based on Git Worktree.

## Overview

The worktree feature in JikiME-ADK manages Git worktrees to allow simultaneous development of multiple SPECs. Each SPEC has an independent branch and working directory, and can be synchronized with the main repository.

## Usage Flow

### Components

| Component | Description |
|-----------|-------------|
| **CLI Tool** | `jikime worktree` (actual command implemented in Go) |
| **Skill** | `jikime-workflow-worktree` (knowledge that helps Claude understand worktree usage) |

### Claude Code Integration

Worktree operates on a **request basis, not automatically**.

```
1. User Request
   "Create a worktree for SPEC-001"
        ↓
2. Claude Loads Skill
   Skill("jikime-workflow-worktree") auto-activates
        ↓
3. Claude Executes CLI
   Bash("jikime worktree new SPEC-001")
        ↓
4. Result Confirmation and Guidance
   "SPEC-001 worktree has been created. Path: ~/jikime/worktrees/..."
```

### Execution Timing

| Timing | Command | Trigger |
|--------|---------|---------|
| Starting SPEC development | `worktree new SPEC-001` | User request |
| When base branch changes | `worktree sync SPEC-001` | User request |
| After development completion | `worktree done SPEC-001` | User request |
| When cleanup is needed | `worktree clean --stale` | User request |

### Direct CLI Usage

You can also run it directly from the terminal:

```bash
# Direct CLI usage
jikime worktree new SPEC-001
jikime worktree sync SPEC-001
jikime worktree done SPEC-001
```

## Key Features

| Feature | Description |
|---------|-------------|
| Auto LLM Config Copy | Auto-detection and copy of `.claude/settings.local.json` |
| Various Sync Strategies | Supports merge, rebase, squash, fast-forward |
| Auto Conflict Resolution | 3-step conflict resolution strategy (ours → theirs → marker removal) |
| Batch Operations | Supports `sync --all`, `clean --stale` |
| Registry Recovery | Auto-recovery of worktrees from disk |

## CLI Commands

### worktree new - Create Worktree

```bash
# Basic creation
jikime worktree new SPEC-001

# Specify custom branch name
jikime worktree new SPEC-001 --branch feature/custom-branch

# Create from specific branch
jikime worktree new SPEC-001 --base develop

# Specify LLM config file
jikime worktree new SPEC-001 --llm-config ~/.claude/my-settings.json

# Force recreation
jikime worktree new SPEC-001 --force
```

**Auto LLM Config Copy**: Even without specifying `--llm-config`, if `.claude/settings.local.json` exists in the main repository, it will be automatically copied to the new worktree.

### worktree sync - Synchronize with Base Branch

```bash
# Basic sync (merge)
jikime worktree sync SPEC-001

# Use rebase strategy
jikime worktree sync SPEC-001 --rebase

# Allow fast-forward only
jikime worktree sync SPEC-001 --ff-only

# Squash strategy (all commits into one)
jikime worktree sync SPEC-001 --squash

# Specify different base branch
jikime worktree sync SPEC-001 --base develop

# Auto-resolve conflicts
jikime worktree sync SPEC-001 --auto-resolve

# Sync all worktrees
jikime worktree sync --all
```

### worktree clean - Clean Up Worktrees

```bash
# Clean only worktrees with merged branches
jikime worktree clean --merged-only

# Clean stale worktrees (default 30 days)
jikime worktree clean --stale

# Clean stale worktrees (14 days)
jikime worktree clean --stale --days 14

# Interactive cleanup
jikime worktree clean --interactive

# Clean all worktrees
jikime worktree clean
```

### worktree list - List Worktrees

```bash
jikime worktree list
```

### worktree status - Check Worktree Status

```bash
jikime worktree status SPEC-001
```

### worktree go - Navigate to Worktree

```bash
# Output directory path
jikime worktree go SPEC-001

# Actually navigate (using shell eval)
eval $(jikime worktree go SPEC-001)
```

### worktree remove - Remove Worktree

```bash
# Basic removal (checks for uncommitted changes)
jikime worktree remove SPEC-001

# Force removal
jikime worktree remove SPEC-001 --force
```

### worktree done - Complete Work and Merge

```bash
# Merge to main branch
jikime worktree done SPEC-001

# Merge and push
jikime worktree done SPEC-001 --push

# Force merge
jikime worktree done SPEC-001 --force
```

### worktree recover - Recover Registry

```bash
jikime worktree recover
```

### worktree config - Check Configuration

```bash
jikime worktree config
jikime worktree config root
jikime worktree config registry
```

## Architecture

### Directory Structure

```
~/jikime/worktrees/           # Worktree root (priority 1)
~/worktrees/                  # Worktree root (priority 2)
├── {project-name}/
│   ├── SPEC-001/
│   │   ├── .git             # Git worktree
│   │   ├── .claude/
│   │   │   └── settings.local.json  # Auto-copied LLM config
│   │   └── ...
│   └── SPEC-002/
└── .jikime-worktree-registry.json  # Registry file
```

### Registry Structure

```json
{
  "worktrees": {
    "project-name": {
      "SPEC-001": {
        "spec_id": "SPEC-001",
        "path": "/path/to/worktree",
        "branch": "feature/SPEC-001",
        "created_at": "2024-01-01T00:00:00Z",
        "last_accessed": "2024-01-02T00:00:00Z",
        "status": "active"
      }
    }
  }
}
```

## Sync Strategies

| Strategy | Flag | Description |
|----------|------|-------------|
| Merge | (default) | Preserves history, creates merge commit |
| Rebase | `--rebase` | Linear history, rewrites commits |
| Squash | `--squash` | Merges all changes into a single commit |
| Fast-forward | `--ff-only` | Syncs only when fast-forward is possible |

## Conflict Resolution

When using the `--auto-resolve` flag, 3-step conflict resolution:

1. **Our Changes**: Keep current branch changes (`git checkout --ours`)
2. **Their Changes**: Apply base branch changes (`git checkout --theirs`)
3. **Marker Removal**: Remove Git conflict markers and commit

Manual resolution is required if auto-resolution fails.

## Auto LLM Config Copy

When creating a worktree, LLM configuration is processed in the following order:

1. Explicitly specified with `--llm-config` flag → Copy that file
2. Flag not specified → Auto-detect and copy `.claude/settings.local.json` from main repository
3. If file doesn't exist, nothing is copied

Environment variables (`${VAR_NAME}`) are automatically substituted during copy.

## Best Practices

1. **Independent Development per SPEC**: Develop each SPEC in an isolated worktree
2. **Regular Synchronization**: Periodically sync all worktrees with `sync --all`
3. **Automated Cleanup**: Clean stale worktrees with `clean --stale --days 14`
4. **Branch Naming**: Recommended to use default branch name `feature/{SPEC-ID}`

## Troubleshooting

### Registry Corruption

```bash
jikime worktree recover
```

### Worktree State Inconsistency

```bash
git worktree prune
jikime worktree recover
```

### Conflict Resolution Failure

```bash
cd /path/to/worktree
git merge --abort  # or git rebase --abort
# Manually resolve conflicts
```

---

Version: 1.0.0
Last Updated: 2026-01-22
