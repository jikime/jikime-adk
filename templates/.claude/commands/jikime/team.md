---
description: "Launch and manage multi-agent teams for complex parallel tasks."
context: dev
---

# /jikime:team — Multi-Agent Team Orchestration

Launch a coordinated team of Claude agents to tackle complex tasks in parallel.

## Quick Start

```bash
# One-click team launch from a template
jikime team launch --template leader-worker --goal "implement user authentication"

# Custom team with 3 workers
jikime team create my-team --workers 3
jikime team spawn  my-team --role leader --backend tmux
jikime team spawn  my-team --role worker --backend tmux
jikime team spawn  my-team --role worker --backend tmux
jikime team spawn  my-team --role worker --backend tmux
```

## Template Quick Launch

| Template | Agents | Best For |
|----------|--------|----------|
| `leader-worker` | 1 leader + 2 workers | Standard parallel tasks |
| `leader-worker-reviewer` | 1 leader + 2 workers + 1 reviewer | Quality-critical work |
| `parallel-workers` | 3 workers (no leader) | Homogeneous parallel tasks |

## Example Workflows

### Implement a Feature

```bash
jikime team launch \
  --template leader-worker \
  --goal "implement OAuth2 login with Google and GitHub" \
  --name oauth-team \
  --budget 500000
```

### GitHub Issues Batch Processing

Use `/jikime:team-harness` instead for GitHub Issues workflow.

### Monitor Team

```bash
# Live board
jikime team board live my-team

# Task status
jikime team tasks list my-team

# Agent health
jikime team status my-team

# Budget usage
jikime team budget show my-team
```

### Send Guidance to Leader

```bash
jikime team inbox send my-team leader \
  "Prioritize the database migration task first"
```

### Wait for Completion

```bash
jikime team tasks wait my-team --timeout 3600
```

### Stop Team

```bash
jikime team stop my-team
```

## Arguments

`$ARGUMENTS` — Optional goal description to inject into agent prompts.

If provided, goal is passed to `--goal` automatically:

```
/jikime:team implement payment processing with Stripe
```

## Notes

- Agents run in tmux sessions (use `tmux ls` to see them)
- Each agent reads its role from `JIKIME_ROLE` env var
- Token budget is enforced via `team-cost-track` hook
- Plan gate requires `JIKIME_PLAN_GATE=1` on worker agents
