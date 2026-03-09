# Harness Engineering

> Autonomous agent orchestration — from GitHub Issue to merged PR, fully automated.

## Concept

**Harness Engineering** is the practice of building a *harness* — a control framework — that directs AI agents to work autonomously on software tasks. Like a horse harness that channels raw power into precise movement, a Harness Engineering system takes Claude's capabilities and guides them through a structured, safe, and repeatable workflow.

In JiKiME-ADK, Harness Engineering is implemented as **`jikime serve`**: a long-running daemon that polls GitHub Issues, spins up isolated workspaces, runs Claude headlessly, and manages the full lifecycle from issue assignment to PR merge — without human intervention.

```
Human writes a GitHub Issue
        ↓
jikime serve detects it (every 15s)
        ↓
Claude reads the issue, writes code, creates PR
        ↓
PR is automatically merged
        ↓
Issue is automatically closed
```

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   jikime serve                      │
│                                                     │
│  ┌──────────┐    ┌─────────────┐    ┌───────────┐  │
│  │ Tracker  │───▶│ Orchestrator│───▶│  Runner   │  │
│  │ (GitHub) │    │             │    │ (Claude)  │  │
│  └──────────┘    └──────┬──────┘    └───────────┘  │
│                         │                           │
│  ┌──────────┐    ┌──────▼──────┐    ┌───────────┐  │
│  │ HTTP API │    │  Workspace  │    │  Hooks    │  │
│  │  :8888   │    │  Manager    │    │ lifecycle │  │
│  └──────────┘    └─────────────┘    └───────────┘  │
└─────────────────────────────────────────────────────┘
```

| Component | Role |
|-----------|------|
| **Tracker** | Polls GitHub Issues for active-state labels (e.g. `jikime-todo`) |
| **Orchestrator** | State machine: dispatch, retry with backoff, reconcile terminal states |
| **Runner** | Executes `claude --print --output-format stream-json` headlessly |
| **Workspace Manager** | Creates per-issue directories, runs lifecycle hooks |
| **HTTP API** | Real-time status dashboard at `http://127.0.0.1:<port>` |

---

## WORKFLOW.md

Every project using `jikime serve` is configured via a single `WORKFLOW.md` file — a YAML front matter block followed by the prompt template.

```yaml
---
tracker:
  kind: github
  project_slug: owner/repo
  # api_key: $GITHUB_TOKEN   # omit → uses gh auth token automatically
  active_states:
    - jikime-todo             # Issues Claude should work on
  terminal_states:
    - jikime-done             # Human-closed
    - Done                    # GitHub closed

polling:
  interval_ms: 15000          # Poll every 15 seconds

workspace:
  root: /tmp/my-workspaces    # Per-issue clone directory

hooks:
  after_create: |             # Runs once when workspace is first created
    git clone https://github.com/owner/repo.git .

  before_run: |               # Runs before each Claude session
    git fetch origin
    git checkout main
    git reset --hard origin/main

  after_run: |                # Runs after each Claude session
    echo "done"

  timeout_ms: 60000           # Hook timeout (60s)

agent:
  max_concurrent_agents: 1    # Parallel Claude sessions
  max_turns: 5                # Max multi-turn loops per session
  max_retry_backoff_ms: 60000 # Max retry delay cap

claude:
  stall_timeout_ms: 180000    # Kill Claude if no output for 3 min

server:
  port: 8888                  # HTTP status API (0 = disabled)
---

You are an autonomous software engineer working on a GitHub issue.

## Issue

**{{ issue.identifier }}**: {{ issue.title }}

{{ issue.description }}

## Instructions

1. Read the issue carefully and implement what is requested.
2. Create a feature branch: `git checkout -b fix/issue-{{ issue.id }}`
3. Make your changes.
4. Commit: `git add -A && git commit -m "fix: {{ issue.identifier }} - {{ issue.title }}"`
5. Push: `git push origin fix/issue-{{ issue.id }}`
6. Create PR: `gh pr create --title "fix: {{ issue.title }}" --body "Closes #{{ issue.id }}" --base main --head fix/issue-{{ issue.id }}`
7. Merge: `gh pr merge --squash --delete-branch --admin`
```

### Template Variables

| Variable | Example |
|----------|---------|
| `{{ issue.id }}` | `9` |
| `{{ issue.identifier }}` | `owner/repo#9` |
| `{{ issue.title }}` | `Add footer component` |
| `{{ issue.description }}` | Full issue body |
| `{{ issue.state }}` | `jikime-todo` |
| `{{ issue.url }}` | `https://github.com/...` |
| `{{ attempt }}` | `2` (retry count) |

---

## Complete Flow

```
1. POLL ─────────────────────────────────────────────────────
   jikime serve polls GitHub every 15s
   Fetches issues with label: jikime-todo
   Sorts by priority → created_at → identifier

2. DISPATCH ─────────────────────────────────────────────────
   Orchestrator checks: not already running, not claimed
   Marks issue as claimed
   Spawns worker goroutine

3. WORKSPACE SETUP ──────────────────────────────────────────
   Creates /tmp/workspaces/owner_repo_9/
   Runs after_create hook: git clone ...
   (On retry: before_run syncs to latest main)

4. CLAUDE RUNS ──────────────────────────────────────────────
   Renders prompt template with issue data
   Executes: claude --print --output-format stream-json \
             --verbose --dangerously-skip-permissions \
             "RENDERED PROMPT"
   Streams output (stall detection: 3 min timeout)

5. BRANCH + PR + MERGE ──────────────────────────────────────
   Claude: git checkout -b fix/issue-9
   Claude: (writes code)
   Claude: git push origin fix/issue-9
   Claude: gh pr create --body "Closes #9"
   Claude: gh pr merge --squash --delete-branch --admin

6. AUTO-CLOSE ───────────────────────────────────────────────
   GitHub: PR merged → Issue #9 closed (state: Done)

7. RECONCILE ────────────────────────────────────────────────
   jikime serve detects issue is in terminal state (Done)
   Runs after_run hook
   Cleans up workspace
   Releases claim
```

---

## Features

### Workspace Isolation

Each issue gets its own directory under `workspace.root`:

```
/tmp/my-workspaces/
  owner_repo_7/    ← Issue #7
  owner_repo_9/    ← Issue #9
  owner_repo_11/   ← Issue #11
```

- Fresh `git clone` on first creation
- `before_run` always syncs to latest `origin/main` before Claude starts
- Isolated: multiple issues can run concurrently without interfering

### Branch Strategy & Conflict Prevention

| Risk | Protection |
|------|-----------|
| Concurrent agents pushing to same branch | Each issue gets its own `fix/issue-N` branch |
| Stale workspace on retry | `before_run`: `git reset --hard origin/main` |
| Human push conflicts | Branch isolation — humans and agents never touch the same branch |
| Two agents racing on main | `max_concurrent_agents: 1` (default) |

### Lifecycle Hooks

| Hook | Timing | Common Use |
|------|--------|-----------|
| `after_create` | First workspace creation | `git clone` |
| `before_run` | Before every Claude session | `git fetch && git reset --hard origin/main` |
| `after_run` | After every Claude session | `git pull` local sync |
| `before_remove` | Before workspace deletion | Archive artifacts |

### Retry with Exponential Backoff

Failed sessions are automatically retried:

```
attempt 1 → 10s delay
attempt 2 → 20s delay
attempt 3 → 40s delay
attempt 4 → 60s (capped by max_retry_backoff_ms)
```

Formula: `min(10000 × 2^(attempt-1), max_retry_backoff_ms)`

Retry is cancelled if the issue moves to a terminal state.

### Token Tracking

With `--output-format stream-json`, token usage is captured from every session and accumulated:

```json
{
  "jikime_totals": {
    "InputTokens": 12840,
    "OutputTokens": 3210,
    "TotalTokens": 16050,
    "SecondsRunning": 342.5
  }
}
```

### HTTP Status API

When `server.port` is set, a status API is available:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Human-readable text dashboard |
| `/api/v1/state` | GET | JSON state snapshot |
| `/api/v1/refresh` | POST | Trigger immediate poll |

```bash
# Live dashboard (refresh every 3s)
watch -n 3 'curl -s http://127.0.0.1:8888/'

# JSON state
curl -s http://127.0.0.1:8888/api/v1/state | jq .

# Trigger immediate poll
curl -s -X POST http://127.0.0.1:8888/api/v1/refresh
```

### WORKFLOW.md Hot-Reload

Edit `WORKFLOW.md` while `jikime serve` is running — changes are applied on the next tick without restarting. Uses `fsnotify` to watch for file changes.

---

## Usage Guide

### 1. Install

```bash
go install github.com/jikime/jikime-adk@latest
```

### 2. Create WORKFLOW.md

```bash
# Copy the example
cp WORKFLOW.md.example my-project/WORKFLOW.md

# Edit for your project
vim my-project/WORKFLOW.md
```

### 3. Create GitHub Labels

```bash
gh label create "jikime-todo" --repo owner/repo \
  --description "Ready for AI agent" --color "0e8a16"

gh label create "jikime-done" --repo owner/repo \
  --description "Completed by AI agent" --color "6f42c1"
```

### 4. Authenticate

```bash
gh auth login    # jikime serve uses gh auth token automatically
```

### 5. Start the Service

```bash
jikime serve my-project/WORKFLOW.md

# With explicit port
jikime serve --port 8888 my-project/WORKFLOW.md
```

### 6. Create Issues

```bash
gh issue create --repo owner/repo \
  --title "Add dark mode toggle" \
  --label "jikime-todo" \
  --body "Add a dark/light mode toggle button..."
```

### 7. Monitor

```bash
# Terminal dashboard
curl -s http://127.0.0.1:8888/

# Or JSON
curl -s http://127.0.0.1:8888/api/v1/state | jq '.running'
```

---

## Developer Guidelines

### Can developers work on main branch?

The `git reset --hard origin/main` in `before_run` **only affects the agent's isolated workspace** under `workspace.root` (e.g., `/tmp/...`). Your local development directory is completely separate and unaffected.

However, it is recommended that developers also use feature branches:

```
✅ Developer: feature/my-work branch → PR → merge to main
✅ Agent:     fix/issue-N branch     → PR → merge to main
→ main is always clean, no conflicts
```

### Handling Long-Running Tasks

For tasks that require many steps (e.g., setting up a new framework), increase limits:

```yaml
agent:
  max_turns: 10

claude:
  stall_timeout_ms: 300000   # 5 minutes

hooks:
  timeout_ms: 120000         # 2 minutes for npm install etc.
```

---

## Reference

### Configuration Defaults

| Key | Default | Description |
|-----|---------|-------------|
| `polling.interval_ms` | `30000` | Poll interval |
| `agent.max_concurrent_agents` | `10` | Parallel sessions |
| `agent.max_turns` | `20` | Turns per session |
| `agent.max_retry_backoff_ms` | `300000` | Max retry delay |
| `claude.stall_timeout_ms` | `300000` | Stall kill timeout |
| `hooks.timeout_ms` | `60000` | Hook execution timeout |
| `server.port` | `0` | HTTP API (disabled) |

### CLI Flags

```bash
jikime serve [WORKFLOW.md] [flags]

Flags:
  -p, --port int   HTTP API server port (0 = disabled)
```

### Related

- [PR Lifecycle Automation](./pr-lifecycle.md)
- [Structured Task Format](./task-format.md)
- [Hooks Configuration](./hooks.md)
