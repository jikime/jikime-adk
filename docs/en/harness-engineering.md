# Harness Engineering

> Autonomous agent orchestration — from GitHub Issue to merged PR, fully automated without human intervention.

---

## Why Harness Engineering?

Development teams spend significant time on repetitive tasks: bug fixes, small feature additions, dependency updates, documentation. These tasks are well-defined and testable, but too mundane for developers to handle manually.

**Harness Engineering** solves this.

| Traditional Approach | Harness Engineering |
|---------------------|---------------------|
| Developer reads issue, creates branch, writes code | Claude handles it automatically |
| Manual PR creation, review requests, merge | Fully automated pipeline |
| Work stops nights and weekends | 24/7 continuous operation |
| Costly context switching for developers | Developers focus on high-complexity work |
| Simple PRs piling up in CI queue | Autonomous agents process in background |

### What tasks are suitable?

```
✅ Good fit
  - Bug fixes (with clear reproduction steps)
  - Small feature additions (single-component scope)
  - Dependency updates
  - Documentation and comment writing
  - Adding tests
  - Linting and formatting fixes
  - Type error fixes

⚠️ Proceed carefully (requires precise specification)
  - Medium-scale refactoring
  - New API endpoints
  - Multi-file changes

❌ Not suitable
  - System architecture decisions
  - Complex business logic design
  - Large database schema migrations
```

---

## Concept

**Harness Engineering** is the practice of building a *harness* — a control framework — that directs AI agents to work autonomously on software tasks. Like a horse harness that channels raw power into precise movement, a Harness Engineering system takes Claude's capabilities and guides them through a structured, safe, and repeatable workflow.

In JiKiME-ADK, Harness Engineering is implemented as **`jikime serve`**: a long-running daemon that polls GitHub Issues, spins up isolated workspaces, runs Claude headlessly, and manages the full lifecycle from issue assignment to PR merge — without human intervention.

```
Human writes a GitHub Issue (label: jikime-todo)
        ↓
jikime serve detects it (every 15s)
        ↓
git clone into isolated workspace
        ↓
Claude reads the issue, creates branch, writes code
        ↓
Claude creates PR → auto-merged
        ↓
GitHub: Issue automatically closed (state: Done)
        ↓
Workspace cleanup (before_remove hook)
```

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                      jikime serve                        │
│                                                          │
│  ┌──────────┐    ┌──────────────┐    ┌────────────────┐  │
│  │ Tracker  │───▶│ Orchestrator │───▶│  Agent Runner  │  │
│  │ (GitHub) │    │              │    │    (Claude)    │  │
│  └──────────┘    └──────┬───────┘    └────────────────┘  │
│                         │                                 │
│  ┌──────────┐    ┌──────▼───────┐    ┌────────────────┐  │
│  │ HTTP API │    │  Workspace   │    │     Hooks      │  │
│  │  :8888   │    │  Manager     │    │   lifecycle    │  │
│  └──────────┘    └──────────────┘    └────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

| Component | Role |
|-----------|------|
| **Tracker** | Polls GitHub API for issues with `active_states` labels |
| **Orchestrator** | State machine: dispatch → retry (exponential backoff) → reconcile terminal states |
| **Agent Runner** | Executes `claude --print --output-format stream-json` headlessly, accumulates token usage |
| **Workspace Manager** | Creates/reuses/deletes per-issue directories, runs lifecycle hooks |
| **HTTP API** | Real-time state snapshots at `http://127.0.0.1:<port>` |

---

## WORKFLOW.md — The Configuration Contract

Every project using `jikime serve` is configured via a single `WORKFLOW.md` file — YAML front matter (runtime config) followed by a Markdown prompt template.

```yaml
---
tracker:
  kind: github
  # api_key: $GITHUB_TOKEN   # omit → uses gh auth token automatically
  project_slug: owner/repo   # GitHub "owner/repo" format
  active_states:
    - jikime-todo             # Issues with this label are dispatched to Claude
  terminal_states:
    - jikime-done             # Human-marked complete
    - Done                    # GitHub auto-closes on PR merge

polling:
  interval_ms: 15000          # Poll GitHub every 15 seconds

workspace:
  root: /tmp/jikime-myrepo   # Per-issue isolated directory

hooks:
  after_create: |             # Runs once on first workspace creation
    git clone https://github.com/owner/repo.git .
    echo "[after_create] cloned to $(pwd)"

  before_run: |               # Runs before every Claude session
    git fetch origin
    git checkout main
    git reset --hard origin/main
    echo "[before_run] synced to $(git rev-parse --short HEAD)"

  after_run: |                # Runs after every Claude session (failures ignored)
    echo "[after_run] done"
    if [ -d "/path/to/local-project/.git" ]; then
      cd "/path/to/local-project" && git pull --ff-only 2>&1 \
        && echo "[after_run] local repo synced at $(git rev-parse --short HEAD)" \
        || echo "[after_run] git pull skipped (local changes or diverged branch)"
    fi

  timeout_ms: 60000           # Hook timeout (60s)

agent:
  max_concurrent_agents: 1    # Parallel Claude sessions
  max_turns: 5                # Max multi-turn loops per session
  max_retry_backoff_ms: 300000 # Max retry delay cap (5 minutes)

claude:
  command: claude              # Claude CLI command
  turn_timeout_ms: 3600000    # Max session duration (1 hour)
  stall_timeout_ms: 180000    # Kill Claude if no output for 3 minutes

server:
  port: 8888                  # HTTP status API port (0 = disabled)
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

| Variable | Example | Description |
|----------|---------|-------------|
| `{{ issue.id }}` | `9` | GitHub Issue number |
| `{{ issue.identifier }}` | `owner/repo#9` | Human-readable key |
| `{{ issue.title }}` | `Add footer component` | Issue title |
| `{{ issue.description }}` | *(full issue body)* | Issue body |
| `{{ issue.state }}` | `jikime-todo` | Current state |
| `{{ issue.url }}` | `https://github.com/...` | Issue URL |
| `{{ issue.branch_name }}` | `fix/footer` | Tracker-provided branch hint |
| `{{ attempt }}` | `2` | Retry count (empty string on first run) |

> **Strict mode**: Any unresolved `{{ variable }}` in the rendered prompt causes a `template_render_error` — the run attempt fails and is retried.

---

## How to Create WORKFLOW.md

### Option 1: CLI Wizard — `jikime serve init` (Recommended)

```bash
cd my-project
jikime serve init
```

An interactive prompt asks 5 questions:

```
? GitHub repo slug (owner/repo)  › owner/my-repo   ← auto-detected from git remote
? Active label                   › jikime-todo
? Workspace root                 › /tmp/jikime-my-repo
? HTTP status API port           › 8888 (recommended)
? Max concurrent agents          › 1 (safe, recommended)
```

If a `.claude/` directory is present, **JiKiME-ADK mode** (J.A.R.V.I.S. agent stack) is selected automatically. Otherwise, **basic mode** (standard git/PR workflow) is used.

After generation, the next steps are displayed:

```
✓ WORKFLOW.md created

Configuration:
  Repo:    owner/my-repo
  Label:   jikime-todo
  Mode:    JiKiME-ADK (J.A.R.V.I.S. agent stack)
  Port:    8888

Next steps:
  1. Create GitHub labels:
     gh label create "jikime-todo" --repo owner/my-repo ...
     gh label create "jikime-done" --repo owner/my-repo ...

  2. Start the service:
     jikime serve WORKFLOW.md
```

### Option 2: Claude Code Slash Command — `/jikime:harness`

Run inside a Claude Code session in a JiKiME-ADK project:

```
/jikime:harness
/jikime:harness --port 9999 --label ai-todo
/jikime:harness --basic --output my-workflow.md
```

Claude analyzes the project and generates an optimized WORKFLOW.md:
- Detects `owner/repo` slug from git remote automatically
- Determines mode from `.claude/` directory presence
- Detects tech stack (package.json / go.mod / requirements.txt) → selects specialist agent

| Flag | Default | Description |
|------|---------|-------------|
| `--basic` | off | Ignore `.claude/`, force basic mode |
| `--port N` | `8888` | HTTP API port |
| `--label LABEL` | `jikime-todo` | Active label name |
| `--output PATH` | `WORKFLOW.md` | Output file path |

### Option 3: Copy Example File

```bash
cp WORKFLOW.md.example ./WORKFLOW.md
vim WORKFLOW.md
```

---

## Quick Start

### Full Setup Flow

```bash
# 1. Create WORKFLOW.md
cd my-project
jikime serve init

# 2. Create GitHub labels
gh label create "jikime-todo" --repo owner/repo \
  --description "Ready for AI agent" --color "0e8a16"

gh label create "jikime-done" --repo owner/repo \
  --description "Completed by AI agent" --color "6f42c1"

# 3. Verify GitHub authentication
gh auth login
gh auth status

# 4. Start the service
jikime serve WORKFLOW.md

# Override port (takes precedence over WORKFLOW.md server.port)
jikime serve --port 8888 WORKFLOW.md
```

### Assigning Issues

```bash
# Add label to an existing issue
gh issue edit 42 --repo owner/repo --add-label "jikime-todo"

# Create a new issue directly
gh issue create --repo owner/repo \
  --title "Add dark mode toggle" \
  --label "jikime-todo" \
  --body "Add a dark/light mode toggle button to the header.

## Requirements
- Persist preference in localStorage
- Default: follow system preference
- Implement with CSS variables"
```

> **Tip**: The more specific your Issue description, the better Claude's implementation quality. Include reproduction steps, expected behavior, and relevant file paths.

---

## Complete Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  1. POLL (every 15s)                                            │
│     GitHub API → fetch issues with active_states labels         │
│     Sort: priority ascending → created_at oldest first          │
│           → identifier lexicographic (tiebreaker)               │
└──────────────────────────────┬──────────────────────────────────┘
                               │ New issue found
┌──────────────────────────────▼──────────────────────────────────┐
│  2. DISPATCH                                                     │
│     ✓ not in running map                                         │
│     ✓ not in claimed set                                         │
│     ✓ concurrent slots available (max_concurrent_agents)         │
│     → mark as claimed, spawn worker goroutine                    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  3. WORKSPACE SETUP                                              │
│     Path: <workspace.root>/<sanitized_identifier>/              │
│     e.g.: /tmp/jikime-myrepo/owner_repo_42/                     │
│                                                                  │
│     [First creation]  after_create hook                          │
│       → git clone https://github.com/owner/repo.git .           │
│                                                                  │
│     [Every session]   before_run hook                            │
│       → git fetch origin                                         │
│       → git checkout main                                        │
│       → git reset --hard origin/main   ← always latest main     │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  4. PROMPT RENDERING                                             │
│     Substitute issue fields into WORKFLOW.md body               │
│     {{ issue.id }} → "42"                                        │
│     {{ issue.title }} → "Add dark mode toggle"                   │
│     Unknown variable → template_render_error → retry             │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  5. CLAUDE EXECUTION                                             │
│     Working directory: <workspace_path>/ (isolated from source)  │
│                                                                  │
│     claude --print \                                             │
│            --output-format stream-json \                         │
│            --verbose \                                           │
│            --dangerously-skip-permissions \                      │
│            --max-turns <max_turns> \                             │
│            "RENDERED PROMPT"                                     │
│                                                                  │
│     [Stall detection] no output for stall_timeout_ms → kill     │
│     [Turn timeout]    exceeds turn_timeout_ms → terminate        │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  6. GIT FLOW (Claude executes)                                   │
│     git checkout -b fix/issue-42                                 │
│     (write code / modify files)                                  │
│     git add -A                                                   │
│     git commit -m "fix: owner/repo#42 - Add dark mode toggle"   │
│     git push origin fix/issue-42                                 │
│                                                                  │
│     gh pr create \                                               │
│       --title "fix: Add dark mode toggle" \                      │
│       --body "Closes #42" \                                      │
│       --base main \                                              │
│       --head fix/issue-42                                        │
│                                                                  │
│     gh pr merge --squash --delete-branch --admin                 │
└──────────────────────────────┬──────────────────────────────────┘
                               │ PR merged
┌──────────────────────────────▼──────────────────────────────────┐
│  7. AUTO-CLOSE                                                   │
│     GitHub detects "Closes #42" → Issue #42 auto-closed         │
│     Issue state: Done (in terminal_states)                       │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  8. RECONCILE & CLEANUP                                          │
│     Next poll tick detects Issue #42 is in terminal state        │
│     Runs after_run hook (failures logged and ignored)            │
│     Runs before_remove hook (failures logged and ignored)        │
│     Deletes workspace directory                                  │
│     Removes from claimed set                                     │
└─────────────────────────────────────────────────────────────────┘
```

---

## How Git Works with Harness Engineering

### Branch Strategy

Harness Engineering uses **branch isolation** as the core safety mechanism.

```
main ──●────────────────●────────────────●──▶
        │                │                │
        └─ fix/issue-42  └─ fix/issue-43  └─ fix/issue-44
           (Claude #1)      (Claude #2)      (Claude #3)
```

- Agents never commit directly to `main`.
- Each issue gets its own dedicated `fix/issue-N` branch.
- PRs are squash-merged and the branch is auto-deleted.

### Workspace Isolation

```
/tmp/jikime-myrepo/
  owner_repo_42/    ← Issue #42 only (independent git repo)
  owner_repo_43/    ← Issue #43 only (independent git repo)
  owner_repo_44/    ← Issue #44 only (independent git repo)
```

Each workspace is a **fully independent git repository**. Multiple agents can run concurrently without affecting each other's files.

### Conflict Prevention

| Risk | Protection |
|------|-----------|
| Two agents pushing to the same branch | Issue ID-based dedicated branch (`fix/issue-N`) |
| Stale codebase on retry | `before_run`: `git reset --hard origin/main` |
| Agent and developer branch collision | Naming convention enforces separation (`fix/issue-*`) |
| Multiple agents competing on main | `max_concurrent_agents: 1` (default) |
| Source repository contamination | Agents only run inside `workspace.root` |

### Developers and Agents Coexisting

```
✅ Developer: feature/my-feature-branch → PR → merge to main
✅ Agent:     fix/issue-42              → PR → merge to main
→ main is always clean, no conflicts
```

The `git reset --hard origin/main` in `before_run` **only runs inside the agent's isolated workspace** under `workspace.root` (e.g., `/tmp/...`). Your local development directory is completely separate and unaffected.

---

## Monitoring and Status

### Terminal Logs

`jikime serve` emits structured logs to stderr:

```
  ╔══════════════════════════════════════╗
  ║   jikime serve — Agent Orchestrator  ║
  ║   Powered by Claude Code + Symphony  ║
  ╚══════════════════════════════════════╝

  Workflow:    /my-project/WORKFLOW.md
  Tracker:     github / owner/repo
  Workspace:   /tmp/jikime-myrepo
  Concurrency: 1 agents
  Poll:        every 15000ms
  HTTP API:    http://127.0.0.1:8888

time=2026-03-09T10:00:00 level=INFO msg="polling..."
time=2026-03-09T10:00:00 level=INFO msg="dispatching issue" issue_id=42
time=2026-03-09T10:00:01 level=INFO msg="agent event" type=session_started issue_id=42
time=2026-03-09T10:00:45 level=INFO msg="agent event" type=turn_completed issue_id=42
```

### HTTP Status API

When `server.port` is configured, a status API is available:

**Text Dashboard** — human-readable:

```bash
curl http://127.0.0.1:8888/

# jikime serve — 2026-03-09T10:05:00Z
#
# Running (1):
#   owner/repo#42        turns=3   Implementing dark mode toggle...
#
# Retrying (0):
#
# Tokens: input=8420 output=2140 total=10560 runtime=42.3s
```

**JSON Snapshot** — programmatic:

```bash
curl -s http://127.0.0.1:8888/api/v1/state | jq .

# {
#   "generated_at": "2026-03-09T10:05:00Z",
#   "counts": { "running": 1, "retrying": 0 },
#   "running": [
#     {
#       "IssueIdentifier": "owner/repo#42",
#       "TurnCount": 3,
#       "LastMessage": "Implementing dark mode toggle..."
#     }
#   ],
#   "jikime_totals": {
#     "InputTokens": 8420,
#     "OutputTokens": 2140,
#     "TotalTokens": 10560,
#     "SecondsRunning": 42.3
#   }
# }
```

**Trigger Immediate Poll**:

```bash
curl -s -X POST http://127.0.0.1:8888/api/v1/refresh
```

**Live Monitoring**:

```bash
# Refresh dashboard every 3 seconds
watch -n 3 'curl -s http://127.0.0.1:8888/'

# Track running issues only
watch -n 5 'curl -s http://127.0.0.1:8888/api/v1/state | jq ".running[].IssueIdentifier"'

# Monitor token usage
watch -n 10 'curl -s http://127.0.0.1:8888/api/v1/state | jq ".jikime_totals"'
```

### WORKFLOW.md Hot-Reload

Edit `WORKFLOW.md` while `jikime serve` is running — changes apply on the next tick **without restarting**:

```bash
vim WORKFLOW.md   # e.g., increase max_concurrent_agents to 3
# → jikime serve detects change via fsnotify
# → logs "WORKFLOW.md reloaded"
# → applies from next dispatch onward
```

Applied changes: poll interval, concurrency limits, active/terminal states, hooks, prompt template.
In-flight agent sessions are not interrupted.

---

## Features

### Workspace Safety Invariants (Symphony SPEC §9.5)

1. **Agent runs only in the per-issue workspace**: `cwd == workspace_path` is validated before launch
2. **Workspace path must stay inside workspace root**: prevents path traversal attacks
3. **Workspace key sanitization**: only `[A-Za-z0-9._-]` allowed; other characters replaced with `_`

### Lifecycle Hooks

| Hook | Timing | On Failure | Common Use |
|------|--------|-----------|-----------|
| `after_create` | First workspace creation only | **Fatal** — aborts issue run | `git clone` |
| `before_run` | Before every Claude session | **Fatal** — aborts that attempt | `git reset --hard origin/main` |
| `after_run` | After every Claude session | **Ignored** (logged only) | Artifact collection, notifications |
| `before_remove` | Before workspace deletion | **Ignored** (logged only) | Backup, archiving |

All hooks must complete within `hooks.timeout_ms` (default: 60 seconds).

### Exponential Backoff Retry

Failed sessions are automatically retried:

```
Formula: min(10000 × 2^(attempt-1), max_retry_backoff_ms)

attempt 1 → 10,000ms (10s)
attempt 2 → 20,000ms (20s)
attempt 3 → 40,000ms (40s)
attempt 4 → 80,000ms (80s)
...
cap       → max_retry_backoff_ms (default: 300,000ms = 5min)
```

Retry is automatically cancelled when the issue moves to `terminal_states`.

After a clean session exit, if the issue is still active, a 1-second continuation retry is scheduled to recheck the state.

### Token Tracking

With `--output-format stream-json`, token usage is captured and accumulated per session in real time:

```bash
curl -s http://127.0.0.1:8888/api/v1/state | jq '.jikime_totals'
# {
#   "InputTokens": 45820,
#   "OutputTokens": 12340,
#   "TotalTokens": 58160,
#   "SecondsRunning": 1842.5
# }
```

---

## Configuration Reference

### All Configuration Keys

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `tracker.kind` | string | — | `"github"` or `"linear"` (required) |
| `tracker.api_key` | string | `$GITHUB_TOKEN` or `gh auth token` | GitHub token |
| `tracker.project_slug` | string | — | `"owner/repo"` (required) |
| `tracker.active_states` | list | `["Todo", "In Progress"]` | Issues to process |
| `tracker.terminal_states` | list | `["Closed", "Cancelled", "Done"]` | Completed issue states |
| `polling.interval_ms` | int | `30000` | Poll interval (ms) |
| `workspace.root` | path | `/tmp/jikime_workspaces` | Workspace root directory |
| `hooks.after_create` | script | — | First-creation hook |
| `hooks.before_run` | script | — | Pre-session hook |
| `hooks.after_run` | script | — | Post-session hook |
| `hooks.before_remove` | script | — | Pre-deletion hook |
| `hooks.timeout_ms` | int | `60000` | Hook execution timeout |
| `agent.max_concurrent_agents` | int | `10` | Parallel sessions |
| `agent.max_turns` | int | `20` | Max turns per session |
| `agent.max_retry_backoff_ms` | int | `300000` | Max retry delay cap |
| `claude.command` | string | `"claude"` | Claude CLI command |
| `claude.turn_timeout_ms` | int | `3600000` | Max session duration (1 hour) |
| `claude.stall_timeout_ms` | int | `300000` | Stall kill timeout |
| `server.port` | int | `0` (disabled) | HTTP API port |

### CLI Flags

```bash
jikime serve [WORKFLOW.md] [flags]

Flags:
  -p, --port int   HTTP API server port (0 = disabled, overrides WORKFLOW.md server.port)

Subcommands:
  init             Interactive wizard to create WORKFLOW.md
```

### Handling Long-Running Tasks

For tasks that require many steps (new framework setup, large-scale refactoring):

```yaml
agent:
  max_turns: 15              # Allow more multi-turn loops

claude:
  turn_timeout_ms: 7200000   # 2 hours (default: 1 hour)
  stall_timeout_ms: 600000   # 10 min stall detection (default: 5 min)

hooks:
  timeout_ms: 180000         # 3 min for npm install / pip install
```

---

## Related

- [PR Lifecycle Automation](./pr-lifecycle.md)
- [Structured Task Format](./task-format.md)
- [POC-First Workflow](./poc.md)
