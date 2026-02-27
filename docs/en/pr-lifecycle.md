# PR Lifecycle Automation

> Automated pull request management from creation to merge with CI monitoring and review resolution.

## Overview

PR Lifecycle Automation handles the complete pull request workflow:

```
Pre-PR Checks (build, test, lint)
  ↓
Create PR (gh pr create)
  ↓
CI Monitoring Loop (max 10 min, 5 retry cycles)
  ↓
Review Resolution Loop (max 3 cycles)
  ↓
Merge & Cleanup
```

---

## Workflow Stages

### Stage 1: Pre-PR Checks

Before creating a PR, the workflow verifies:

- No uncommitted changes
- Build passes locally
- Tests pass locally
- Lint passes locally
- Branch is pushed to remote

### Stage 2: Create PR

Generates a structured PR with:

```
## Summary
- What this PR does (1-3 bullet points)

## Changes
- List of significant changes

## Test Plan
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
```

PR title follows conventional commits format: `feat(scope): description`

### Stage 3: CI Monitoring Loop

```
┌──────────────────────────────────────────────────┐
│              CI Monitoring Loop                   │
│                                                   │
│  Check CI Status ──▶ All Passing? ──▶ Done       │
│       │                    │                      │
│       │               No (Failed)                 │
│       │                    │                      │
│       │         ┌─────────▼──────────┐            │
│       │         │ Diagnose Failure   │            │
│       │         │ Fix Issue          │            │
│       │         │ Push Fix           │            │
│       │         └─────────┬──────────┘            │
│       │                   │                       │
│       │            Retry < 5?                     │
│       │           ├── Yes → Re-check              │
│       │           └── No  → STOP                  │
│       │                                           │
│       └── Pending? Wait 60s, re-check (max 10m)  │
└──────────────────────────────────────────────────┘
```

**CI Failure Categories**:

| Category | Auto-Fix | Example |
|----------|----------|---------|
| Lint error | Yes | `npm run lint --fix` |
| Type error | Fix types | Missing type annotation |
| Test failure | Fix code | Assertion mismatch |
| Build error | Fix build | Missing dependency |
| Flaky test | Re-run | `gh run rerun <id>` |

### Stage 4: Review Resolution Loop

```
Read all review comments
  ↓
Categorize (bug/style/suggestion/question)
  ↓
Fix all issues in single commit
  ↓
Push and re-request review
  ↓
Repeat (max 3 cycles)
```

**Comment Priority**:

| Type | Priority | Action |
|------|----------|--------|
| Bug/Logic error | Critical | Fix immediately |
| Style/Convention | Medium | Apply suggestion |
| Suggestion/Optional | Low | Evaluate and respond |
| Question | Low | Answer in comment |

### Stage 5: Merge & Cleanup

- Verify all CI checks passing
- Verify review approved
- Squash merge and delete branch
- Switch to main and pull

---

## Timeout Protection

| Parameter | Default | Description |
|-----------|---------|-------------|
| CI wait timeout | 10 min | Max time to wait for CI completion |
| CI retry cycles | 5 | Max fix-and-recheck attempts |
| Polling interval | 60s | Time between CI status checks |
| Review cycles | 3 | Max rounds of review resolution |

---

## Usage

### Slash Command

```bash
# Default PR lifecycle (squash merge to main)
/jikime:pr-lifecycle

# Custom base branch
/jikime:pr-lifecycle --base develop

# Draft PR
/jikime:pr-lifecycle --draft

# Merge commit instead of squash
/jikime:pr-lifecycle --merge

# Keep branch after merge
/jikime:pr-lifecycle --no-delete-branch
```

### Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `--base <branch>` | Base branch for PR | main |
| `--squash` | Squash merge | true |
| `--merge` | Merge commit | false |
| `--no-delete-branch` | Keep branch after merge | false |
| `--draft` | Create as draft PR | false |

### As a Skill

```
Skill("jikime-workflow-pr-lifecycle")
```

---

## Related Documentation

- [POC-First Workflow](./poc-first.md) — Phase 5 uses this skill
- [Structured Task Format](./task-format.md) — PR tasks use 5-field format
- [Ralph Loop](./ralph-loop.md) — CI fix loop similar pattern
- [Git Workflow](../ko/commands.md) — Git conventions and branching
