---
description: "GitHub Issues parallel processing via jikime team. Sequential Harness → parallel team upgrade."
context: dev
---

# /jikime:team-harness — GitHub Issues Team Processing

Upgrade your Harness Engineering workflow from sequential issue processing to
a parallel multi-agent team. Each agent handles one issue in isolation.

## How It Works

```
GitHub Issues (open/labeled)
       │
       ▼
  Leader Agent
  ├─ List open issues
  ├─ Create one task per issue
  └─ Distribute tasks to workers
       │
  ┌────┴────────────────────────────┐
  │            Workers              │
  Worker-1       Worker-2      Worker-3
  ├─ Claim       ├─ Claim       ├─ Claim
  ├─ Fix issue   ├─ Fix issue   ├─ Fix issue
  ├─ Create PR   ├─ Create PR   ├─ Create PR
  └─ Complete    └─ Complete    └─ Complete
       │
  Reviewer (optional)
  └─ Review PRs before merge
```

## Quick Start

```bash
# Step 1: Generate or update your WORKFLOW.md
/jikime:harness

# Step 2: Launch the team
jikime team launch \
  --template leader-worker \
  --goal "Process all open GitHub issues in owner/repo, create PRs for each" \
  --name harness-team \
  --backend tmux

# Step 3: Monitor
jikime team board live harness-team
```

## Leader Prompt Template

The leader agent should:

```markdown
Team: {{team-name}} | Role: leader | Goal: Process all open GitHub issues

1. Fetch open issues:
   gh issue list --repo owner/repo --state open --json number,title,labels

2. For each issue, create a task:
   jikime team tasks create harness-team "Fix issue #N: <title>" \
     --desc "GitHub issue #N. Labels: <labels>" \
     --dod "PR created, CI passes, PR number reported in result"

3. Monitor worker progress:
   jikime team tasks list harness-team

4. When all done:
   jikime team inbox broadcast harness-team "All issues processed"
   jikime team lifecycle shutdown harness-team
```

## Worker Prompt Template

```markdown
Team: {{team-name}} | Role: worker | Goal: Fix assigned GitHub issues

Loop:
  1. jikime team tasks claim harness-team <task-id> --agent $JIKIME_AGENT_ID
  2. Read issue number from task description
  3. git worktree add fix-issue-N (if --worktree flag used)
  4. Implement fix
  5. Create PR: gh pr create --title "fix: issue #N" --body "Fixes #N"
  6. jikime team tasks complete harness-team <task-id> \
       --result "PR #<pr-number> created"
```

## Comparison: Before vs After

| Aspect | Sequential Harness | Team Harness |
|--------|-------------------|--------------|
| Issue processing | One at a time | Parallel (N workers) |
| Speed | 1x | N× faster |
| Worktree isolation | Per issue | Per agent |
| Progress visibility | Log file | `jikime team board` |
| Cost tracking | None | Per-agent token budget |

## Arguments

`$ARGUMENTS` — GitHub repo slug (owner/repo) or WORKFLOW.md path.

```
/jikime:team-harness myorg/myrepo
/jikime:team-harness ./WORKFLOW.md
```

## Notes

- Works with existing WORKFLOW.md from `/jikime:harness`
- Requires `gh` CLI authenticated with repo access
- Token budget per worker prevents runaway costs
- Use `--worktree` for complete git isolation between workers
