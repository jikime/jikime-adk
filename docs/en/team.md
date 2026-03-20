# JiKiME Team — Multi-Agent Team Orchestration Complete Guide

**Orchestrate multiple Claude Code agents as a team to handle complex tasks in parallel**

> **Version:** JiKiME-ADK v1.5.0+
> **Last Updated:** 2026-03-20

---

## Table of Contents

1. [Overview & Architecture](#1-overview--architecture)
2. [CLI Command Reference](#2-cli-command-reference)
3. [Web UI (Webchat) Features](#3-web-ui-webchat-features)
4. [REST API & SSE Endpoints](#4-rest-api--sse-endpoints)
5. [File System Structure](#5-file-system-structure)
6. [Environment Variables](#6-environment-variables)
7. [Team Workflow Patterns](#7-team-workflow-patterns)
8. [Git Worktree Isolation](#8-git-worktree-isolation)
9. [GitHub Issues / Harness Integration](#9-github-issues--harness-integration)
10. [Template System](#10-template-system)
11. [Practical Examples](#11-practical-examples)
12. [Troubleshooting](#12-troubleshooting)

---

## 1. Overview & Architecture

### What is JiKiME Team?

JiKiME Team is an orchestration system that groups multiple Claude Code agent instances into a single team, enabling complex software development tasks to be processed **in parallel**.

```
┌─────────────────────────────────────────────────────┐
│                   jikime team                        │
│                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐       │
│  │  leader  │◄──►│ worker-1 │    │ worker-2 │       │
│  │ (Claude) │    │ (Claude) │    │ (Claude) │       │
│  └──────────┘    └──────────┘    └──────────┘       │
│        │               │               │            │
│        └───────────────┴───────────────┘            │
│                        │                            │
│            ┌───────────▼──────────┐                 │
│            │   ~/.jikime/teams/   │                 │
│            │  tasks/ registry/    │                 │
│            │  inbox/ costs/       │                 │
│            └──────────────────────┘                 │
└─────────────────────────────────────────────────────┘
```

### Core Concepts

| Concept | Description |
|---------|-------------|
| **Team** | A group of agents sharing a common task repository |
| **Agent** | An independent process running Claude Code CLI (tmux or subprocess) |
| **Task** | A unit of work processed by agents (pending → in_progress → done) |
| **Role** | Agent role: leader, worker, reviewer |
| **Workspace** | Per-agent isolated git worktree |
| **Template** | Team configuration definition (agent count, roles, initial tasks) |
| **Budget** | Token usage limit for the entire team |

### Spawn Backends

| Backend | Characteristics | When to Use |
|---------|----------------|-------------|
| `tmux` (default) | Interactive, real-time monitoring in terminal | Development, debugging |
| `subprocess` | Non-interactive, output piped to log files | CI/CD, automation |

---

## 2. CLI Command Reference

Top-level entry point: `jikime team` (alias: `jikime t`)

### 2.1 Team Creation & Management

#### `jikime team create <team-name>`

Creates an empty team workspace.

```bash
jikime team create <team-name> [flags]

Flags:
  -w, --workers int       Number of worker agents (0 = unlimited, default: 0)
  -b, --backend string    Spawn backend: tmux or subprocess (default: tmux)
      --budget int        Token budget limit (0 = no limit)
      --timeout int       Execution timeout in seconds (0 = no timeout)
      --max-agents int    Max concurrent agents (0 = unlimited)
  -t, --template string   Initial template name

Examples:
  jikime team create my-team
  jikime team create auth-team --workers 3 --budget 100000
  jikime team create api-team --template leader-worker --backend subprocess
```

**Created directory structure:**
```
~/.jikime/teams/<team-name>/
├── config.json      # Team configuration
├── webchat.json     # Webchat metadata
├── tasks/           # Task files
├── inbox/           # Agent message inbox
├── registry/        # Agent registration info
├── costs/           # Token usage records
└── events/          # Event logs
```

---

#### `jikime team launch`

Creates and starts a complete team from a template in one command.

```bash
jikime team launch [flags]

Flags:
  -t, --template string   Template name (required)
      --name string       Team name (auto-generated if empty)
  -g, --goal string       Goal to inject into agent prompts
  -b, --backend string    Spawn backend (default: tmux)
  -w, --worktree          Create isolated git worktree per agent
      --budget int        Token budget (overrides template default)

Examples:
  jikime team launch --template leader-worker \
    --goal "implement user authentication with JWT"

  jikime team launch --template leader-worker-reviewer \
    --name auth-team \
    --goal "redesign API layer" \
    --worktree \
    --budget 200000
```

**Automatic execution sequence:**
1. Load template and create team directory structure
2. Auto-create initial tasks defined in template
3. Create git worktree per agent if `--worktree` flag is set
4. Generate per-role prompts with injected goal
5. Auto-spawn all agents

---

#### `jikime team spawn <team-name>`

Adds a new agent to an existing team.

```bash
jikime team spawn <team-name> [flags]

Flags:
  -r, --role string       Agent role: leader, worker, reviewer (default: worker)
      --agent-id string   Agent ID (auto-generated if empty: agent-XXXXXXXX)
  -b, --backend string    Spawn backend (default: tmux)
      --worktree string   Git worktree path for this agent
  -p, --prompt string     Initial prompt for the agent
      --skip-permissions  Pass --dangerously-skip-permissions (default: true)
      --resume            Resume previous Claude session if available

Examples:
  jikime team spawn my-team --role leader --agent-id leader
  jikime team spawn my-team --role worker --agent-id worker-1
  jikime team spawn my-team --role worker \
    --worktree ~/.jikime/worktrees/my-team/worker-1 \
    --prompt "Focus on implementing the database layer"
```

**Tmux backend behavior:**
- Session name: `jikime-<team>-<agent-id>`
- Agents identify themselves via `JIKIME_AGENT_ID`, `JIKIME_TEAM_NAME` env vars
- Starts `claude` CLI and maintains interactive session

---

#### `jikime team status <team-name>`

Queries the current state of a team.

```bash
jikime team status <team-name> [--json]

Examples:
  jikime team status my-team
  jikime team status my-team --json | jq .agents
```

**Sample output:**
```
Team: my-team
Dir:  ~/.jikime/teams/my-team

Agents (3):
  ✅ worker-1 [worker] task:abc12345
  ✅ worker-2 [worker] task:def67890
  ❌ leader   [leader] task:-

Tasks (5): todo=2 wip=2 done=1 blocked=0

Tokens: 45,230 / 100,000 (45.2%)
```

---

#### `jikime team stop <team-name>`

Stops all agents in a team.

```bash
jikime team stop my-team
```

---

#### `jikime team discover`

Discovers teams running on the current machine.

```bash
# List all active teams
jikime team discover list [--json]

# Join an existing team as a new agent
jikime team discover join <team-name> \
  --role worker \
  --agent-id my-worker

# Approve a join request (as leader)
jikime team discover approve <agent-id> --team my-team

# Reject a join request
jikime team discover reject <agent-id>
```

---

### 2.2 Task Management

#### `jikime team tasks create <team-name> <title>`

Creates a new task.

```bash
jikime team tasks create <team-name> <title> [flags]

Flags:
  -d, --desc string       Task detailed description
      --dod string        Definition of Done
      --depends-on string Comma-separated dependency task IDs
  -p, --priority int      Priority (higher = more important, default: 0)
      --tags string       Comma-separated tags

Examples:
  jikime team tasks create my-team "Implement login endpoint"

  jikime team tasks create my-team "Design database schema" \
    --desc "Create users, sessions, tokens tables" \
    --dod "All tables created with proper indexes and constraints" \
    --priority 3 \
    --tags "database,schema"

  jikime team tasks create my-team "Write unit tests" \
    --depends-on abc12345,def67890 \
    --priority 1
```

---

#### `jikime team tasks list <team-name>`

Lists all tasks in a team.

```bash
jikime team tasks list <team-name> [flags]

Flags:
  -s, --status string   Filter by status: pending|in_progress|done|blocked|failed
  -a, --agent string    Filter by agent ID

Examples:
  jikime team tasks list my-team
  jikime team tasks list my-team --status in_progress
  jikime team tasks list my-team --agent worker-1
  jikime team tasks list my-team --status done --agent worker-1
```

---

#### `jikime team tasks get <team-name> <task-id>`

Retrieves detailed information about a specific task.

```bash
jikime team tasks get my-team abc12345
```

---

#### `jikime team tasks update <team-name> <task-id>`

Updates task status or metadata.

```bash
jikime team tasks update <team-name> <task-id> [flags]

Flags:
  -t, --title string    New title
  -d, --desc string     New description
      --dod string      New definition of done
  -p, --priority int    New priority
  -s, --status string   Status transition: pending|in_progress|done|blocked|failed
  -a, --agent string    Agent ID (required for in_progress transition)
  -r, --result string   Result summary (for done/failed)

State transition rules:
  pending    → in_progress  Agent claims task (--agent required)
  in_progress→ done         Agent completes (--result recommended)
  in_progress→ failed       Agent reports failure
  any        → blocked      Waiting on dependency
  blocked    → pending      Unblocked

Examples:
  # Claim a task (start)
  jikime team tasks update my-team abc123 \
    --status in_progress --agent worker-1

  # Complete a task
  jikime team tasks update my-team abc123 \
    --status done --agent worker-1 \
    --result "Implemented with 95% test coverage"

  # Mark task as failed
  jikime team tasks update my-team abc123 \
    --status failed --agent worker-1 \
    --result "API returned 403, need credentials"
```

---

#### `jikime team tasks claim <team-name> <task-id>`

Claims a task for the current agent (uses env vars automatically).

```bash
jikime team tasks claim <team-name> <task-id> [--agent <id>]

# Using environment variables (inside agent process)
export JIKIME_AGENT_ID=worker-1
jikime team tasks claim my-team abc12345

# Explicit agent ID
jikime team tasks claim my-team abc12345 --agent worker-1
```

---

#### `jikime team tasks complete <team-name> <task-id>`

Marks a task as completed.

```bash
jikime team tasks complete <team-name> <task-id> [flags]

Flags:
  -a, --agent string    Agent ID (required)
  -r, --result string   Result summary

Examples:
  jikime team tasks complete my-team abc12345 \
    --agent worker-1 \
    --result "All tests passing, ready for code review"
```

---

#### `jikime team tasks wait <team-name>`

Waits until all tasks are complete (useful for CI/CD integration).

```bash
jikime team tasks wait <team-name> [flags]

Flags:
  -t, --timeout int    Max wait time in seconds (0 = no limit)
  -i, --interval int   Poll interval in seconds (default: 5)

Examples:
  jikime team tasks wait my-team --timeout 3600
  jikime team tasks wait my-team --interval 10

# Exit codes:
  0 = All tasks completed (done status)
  1 = Timeout exceeded
  2 = Failed tasks exist
```

**Progress display:**
```
tasks: 5/10 done | wip:3 pending:2 blocked:0
```

---

### 2.3 Plan Management

#### `jikime team plan submit <team-name>`

Worker submits a work plan to the leader for review.

```bash
jikime team plan submit <team-name> [flags]

Flags:
  -t, --title string   Plan title (default: "Plan")
  -b, --body string    Plan body text (inline)
  -f, --file string    Read plan body from file
  -a, --agent string   Submitting agent ID

Examples:
  jikime team plan submit my-team \
    --title "Database Schema Design" \
    --body "Propose using PostgreSQL with 3 tables: users, sessions, tokens"

  jikime team plan submit my-team \
    --title "API Implementation Plan" \
    --file plan.md \
    --agent worker-1
```

---

#### `jikime team plan approve <plan-id>`

Leader approves a pending plan.

```bash
jikime team plan approve <plan-id> [--reviewer <agent-id>]

Examples:
  jikime team plan approve plan-abc12345
  jikime team plan approve plan-abc12345 --reviewer leader
```

---

#### `jikime team plan reject <plan-id>`

Leader rejects a pending plan.

```bash
jikime team plan reject <plan-id> [--reviewer <id>] [--reason <text>]

Examples:
  jikime team plan reject plan-abc12345 \
    --reason "Need to consider distributed caching layer"
```

---

#### `jikime team plan list`

Lists all plans.

```bash
jikime team plan list [--team <team-name>]
```

---

### 2.4 Board Management

#### `jikime team board show <team-name>`

Displays a snapshot of the current team board.

```bash
jikime team board show <team-name> [--json]

Examples:
  jikime team board show my-team
  jikime team board show my-team --json | jq .tasks
```

**Sample output:**
```
╔══════════════════════════════════╗
║  Team Board: my-team             ║
╚══════════════════════════════════╝

Agents (3):
  ✅ worker-1  [active]   role:worker  task:abc12345
  ✅ worker-2  [active]   role:worker  task:def67890
  ❌ leader    [offline]  role:leader  task:-

Tasks (5 total):
  pending:2  in_progress:2  done:1  failed:0  blocked:0
```

---

#### `jikime team board live <team-name>`

Displays a live-refreshing board in the terminal (Ctrl+C to stop).

```bash
jikime team board live <team-name> [--interval <seconds>]

Flags:
  -i, --interval int   Refresh interval in seconds (default: 3)
```

---

#### `jikime team board attach <team-name>`

Links all agent windows into a single tmux dashboard session.

```bash
jikime team board attach my-team

# After connecting to the tmux session:
# Ctrl-b n → next agent window
# Ctrl-b p → previous agent window
# Ctrl-b d → detach session
```

---

#### `jikime team board serve [team-name]`

Starts an HTTP web dashboard server.

```bash
jikime team board serve [team-name] [flags]

Flags:
  -p, --port int        HTTP port (default: 8080)
      --host string     Bind address (default: 127.0.0.1)
  -i, --interval float  SSE push interval in seconds (default: 2.0)

Examples:
  jikime team board serve my-team
  jikime team board serve --port 3000 --host 0.0.0.0
```

**Provided endpoints:**
- `GET /` → React SPA dashboard
- `GET /api/overview` → All teams list (JSON)
- `GET /api/team/:name` → Specific team snapshot (JSON)
- `GET /api/events/:name` → Real-time SSE stream

---

#### `jikime team board overview`

Displays an overview of all teams.

```bash
jikime team board overview
```

---

### 2.5 Budget Management

#### `jikime team budget show <team-name>`

Displays token usage and budget.

```bash
jikime team budget show <team-name> [--agent <id>]

Examples:
  jikime team budget show my-team
  jikime team budget show my-team --agent worker-1
```

**Sample output:**
```
Budget for team 'my-team' (limit: 100,000 tokens)

AGENT          INPUT      OUTPUT     TOTAL      %
worker-1       12,000     3,400      15,400     15.4%
worker-2       10,500     2,800      13,300     13.3%
leader         8,000      2,100      10,100     10.1%

TOTAL          38,800 tokens used
Budget used: 38.8% (38,800 / 100,000)
```

---

#### `jikime team budget set <team-name> <tokens>`

Sets the token budget for a team.

```bash
jikime team budget set my-team 200000
```

---

#### `jikime team budget report <team-name>`

Agent reports token usage (called from Claude hooks).

```bash
jikime team budget report <team-name> [flags]

Flags:
  -a, --agent string       Agent ID (default: $JIKIME_AGENT_ID)
      --task string        Task ID
      --model string       Model name (e.g. claude-sonnet-4-6)
      --input-tokens int   Input tokens consumed
      --output-tokens int  Output tokens consumed

Example (from Claude hook script):
  #!/bin/bash
  jikime team budget report "$JIKIME_TEAM_NAME" \
    --agent "$JIKIME_AGENT_ID" \
    --task "$JIKIME_TASK_ID" \
    --input-tokens 1234 \
    --output-tokens 567 \
    --model claude-sonnet-4-6
```

---

### 2.6 Workspace Management

Manages git worktree-based isolated agent workspaces.

Worktree path: `~/.jikime/worktrees/<team-name>/<agent-id>/`
Branch name: `jikime-<team-name>-<agent-id>`

#### `jikime team workspace list <team-name>`

Lists all active worktrees for a team.

```bash
jikime team workspace list my-team

# Output:
# Workspaces for team 'my-team':
#   worker-1  ~/.jikime/worktrees/my-team/worker-1
#   worker-2  ~/.jikime/worktrees/my-team/worker-2
```

---

#### `jikime team workspace checkpoint <team-name>`

Auto-commits current workspace changes.

```bash
jikime team workspace checkpoint <team-name> [flags]

Flags:
  -a, --agent string    Agent ID (default: $JIKIME_AGENT_ID)
  -m, --message string  Commit message (default: "checkpoint: <agent> <timestamp>")

Examples:
  jikime team workspace checkpoint my-team --agent worker-1
  jikime team workspace checkpoint my-team \
    --agent worker-1 \
    --message "feat: implement login endpoint"
```

---

#### `jikime team workspace merge <team-name>`

Merges agent workspace branch into main branch.

```bash
jikime team workspace merge <team-name> [flags]

Flags:
  -a, --agent string    Agent ID (default: $JIKIME_AGENT_ID)
  -t, --target string   Target branch (default: main)
      --cleanup         Remove worktree after merge

Examples:
  jikime team workspace merge my-team --agent worker-1
  jikime team workspace merge my-team \
    --agent worker-1 \
    --target develop \
    --cleanup
```

---

#### `jikime team workspace cleanup <team-name>`

Removes workspace(s).

```bash
jikime team workspace cleanup <team-name> [--agent <id>]

# Remove only a specific agent's workspace
jikime team workspace cleanup my-team --agent worker-1

# Remove all workspaces for the team
jikime team workspace cleanup my-team
```

---

#### `jikime team workspace status <team-name>`

Shows git diff stat for an agent's workspace.

```bash
jikime team workspace status my-team --agent worker-1
```

---

### 2.7 Additional Commands

```bash
# Team configuration
jikime team config show <team-name>
jikime team config set <team-name> <key> <value>
jikime team config get <team-name> <key>
jikime team config health        # Check ~/.jikime directory health

# Agent inbox (inter-agent messages)
jikime team inbox <team-name>    # View message list

# Agent identity
jikime team identity <team-name> # Query agent ID/team info

# Session management (state snapshots)
jikime team session <team-name>  # View session list

# Lifecycle hooks (called automatically on agent exit)
jikime team lifecycle on-exit    # Agent exit cleanup

# Templates
jikime team template list        # List available templates
```

---

## 3. Web UI (Webchat) Features

Manage Team features visually at `http://localhost:<port>` in Webchat.

### 3.1 Accessing the Team Tab

Click the **Team** icon (👥) in the left sidebar to open the team dashboard.

### 3.2 Creating a Team (TeamCreateModal)

Click **New Team** button to create a team.

| Field | Description | Default |
|-------|-------------|---------|
| Team Name | Unique team identifier (required) | - |
| Template | Template selection (built-in/custom) | None |
| Workers | Number of worker agents | 2 |
| Budget | Token budget | 0 (unlimited) |

**Template groups:**
- **Built-in Templates**: `leader-worker`, `leader-worker-reviewer`, `parallel-workers`
- **Custom Templates**: User-created templates

---

### 3.3 Kanban Board (TeamBoard)

Kanban board with 5 status columns:

| Column | Status | Color |
|--------|--------|-------|
| Pending | `pending` | Gray |
| In Progress | `in_progress` | Blue |
| Blocked | `blocked` | Orange |
| Done | `done` | Green |
| Failed | `failed` | Red |

**Information shown on task cards:**
- Task ID (first 7 characters)
- Status icon
- Title
- Assigned agent ID
- Priority ★

---

### 3.4 Agent Panel (BoardPanel)

Displays the list of agents in a team:

- Agent role (leader/worker/reviewer)
- Current status (active/idle/offline)
- Current assigned task ID
- Unread message count
- **X button**: Immediately terminate the agent's session

---

### 3.5 Adding Tasks

Click the **+ button** at the top of the board to add new tasks:

| Field | Description |
|-------|-------------|
| Title | Task title (required) |
| Description | Detailed description |
| Priority | Priority number |
| Tags | Comma-separated tags |

---

### 3.6 Real-Time Updates

Team state is updated in real-time via SSE (Server-Sent Events):
- Auto-updates on team changes (default 2-second interval)
- Immediate reflection of agent status changes
- Immediate reflection of task status changes

---

### 3.7 Starting Team Serve (TeamServeModal)

Click **▶️ button** then select "Board Server":
- Configure port, host, and refresh interval
- Auto-generate and copy CLI command

---

### 3.8 Template Manager (TemplateManagerModal)

Click ⚙️ icon:
- View custom template list
- Create new template (YAML editor)
- Edit/delete existing templates

---

## 4. REST API & SSE Endpoints

API provided by the webchat server (`http://localhost:port`).

### 4.1 Team Management API

```http
# List teams
GET /api/team/list
GET /api/team/list?projectPath=/path/to/project

Response: {
  "teams": [
    {
      "name": "my-team",
      "config": { "budget": 100000, "workers": 3 },
      "taskCounts": { "pending": 2, "in_progress": 1, "done": 5 }
    }
  ]
}

---

# Get team details
GET /api/team/:name

Response: {
  "config":      { "name": "my-team", "budget": 100000 },
  "agents":      [{ "id": "worker-1", "role": "worker", "status": "active" }],
  "tasks":       [{ "id": "abc123", "title": "...", "status": "in_progress" }],
  "cost":        { "total": 45230, "agents": { "worker-1": { "tokens": 15400 } } },
  "taskCounts":  { "pending": 2, "in_progress": 1, "done": 3 }
}

---

# Create team
POST /api/team/create
Content-Type: application/json

{
  "name":        "my-team",
  "template":    "leader-worker",   // optional
  "workers":     2,                 // optional
  "budget":      100000,            // optional
  "projectPath": "/path/to/project" // optional
}

Response: 200 OK (success) | 400/500 (error)

---

# Delete team
DELETE /api/team/:name

Response: 200 OK
```

### 4.2 Task Management API

```http
# List tasks
GET /api/team/:name/tasks
GET /api/team/:name/tasks?status=in_progress&agent=worker-1

Response: { "tasks": [...] }

---

# Create task
POST /api/team/:name/tasks
Content-Type: application/json

{ "title": "Implement login", "desc": "Create POST /auth/login" }

Response: { "task": { "id": "abc123", "title": "...", "status": "pending" } }

---

# Update task
PATCH /api/team/:name/tasks/:id
Content-Type: application/json

{
  "status":   "in_progress",  // optional
  "agent_id": "worker-1",     // optional
  "result":   "Done"          // optional
}

Response: { "task": { "id": "abc123", "status": "in_progress" } }
```

### 4.3 Agent Management API

```http
# List agents
GET /api/team/:name/agents

Response: {
  "agents": [
    {
      "id":            "worker-1",
      "role":          "worker",
      "status":        "active",
      "current_task":  "abc12345",
      "tmux_session":  "jikime-my-team-worker-1",
      "pid":           12345
    }
  ]
}

---

# Kill agent session
DELETE /api/team/:name/agents/:agentId

Response: 200 OK (runs tmux kill-session)
```

### 4.4 Messaging API

```http
# Send message
POST /api/team/:name/inbox/send
Content-Type: application/json

{
  "to":   "worker-1",           // specific agent or "broadcast"
  "body": "Please prioritize this task"
}

Response: { "message": { "id": "msg-xxx", "from": "leader", "to": "worker-1" } }
```

### 4.5 Budget API

```http
# Get budget summary
GET /api/team/:name/budget

Response: {
  "total": 45230,
  "agents": {
    "worker-1": { "tokens": 15400 },
    "worker-2": { "tokens": 13300 }
  }
}
```

### 4.6 SSE Stream (Real-Time)

```http
GET /api/team/:name/events
Accept: text/event-stream

# Response format (pushed every 2 seconds):
data: {
  "type": "update",
  "time": "2026-03-20T10:00:00Z",
  "team": {
    "name":        "my-team",
    "leaderName":  "leader",
    "description": ""
  },
  "tasks":   [...],
  "agents":  [...],
  "members": [...],
  "taskSummary": {
    "pending":     2,
    "in_progress": 1,
    "done":        3,
    "failed":      0,
    "blocked":     0
  },
  "messages": [
    {
      "from":      "worker-1",
      "to":        "leader",
      "type":      "direct",
      "timestamp": "2026-03-20T09:59:30Z",
      "content":   "Task abc123 complete"
    }
  ]
}
```

### 4.7 Template API

```http
# List templates
GET /api/template/list

Response: {
  "templates": [
    {
      "name":        "leader-worker",
      "description": "Leader coordinates, workers execute",
      "agents":      2
    }
  ]
}

---

# Get specific template
GET /api/template/:name
```

---

## 5. File System Structure

```
~/.jikime/
├── teams/
│   └── <team-name>/
│       ├── config.json              # TeamConfig
│       │   {
│       │     "name": "my-team",
│       │     "template": "leader-worker",
│       │     "budget": 100000,
│       │     "maxAgents": 0,
│       │     "timeoutSeconds": 0,
│       │     "createdAt": "..."
│       │   }
│       ├── webchat.json             # Webchat metadata
│       │   { "projectPath": "/path/to/project" }
│       ├── tasks/
│       │   └── <task-id>.json       # Task file
│       │       {
│       │         "id": "abc123",
│       │         "title": "Implement login",
│       │         "status": "in_progress",
│       │         "agent_id": "worker-1",
│       │         "priority": 2,
│       │         "dod": "...",
│       │         "created_at": "...",
│       │         "claimed_at": "..."
│       │       }
│       ├── inbox/
│       │   ├── <agent-id>/
│       │   │   └── <msg-id>.json    # Message file
│       │   └── event-log.jsonl      # Full event log (JSON Lines)
│       ├── registry/
│       │   └── <agent-id>.json      # AgentInfo file
│       │       {
│       │         "id": "worker-1",
│       │         "role": "worker",
│       │         "status": "active",
│       │         "pid": 12345,
│       │         "tmux_session": "jikime-my-team-worker-1",
│       │         "current_task": "abc123",
│       │         "last_heartbeat": "...",
│       │         "joined_at": "..."
│       │       }
│       ├── costs/
│       │   └── <agent-id>-<timestamp>.json  # CostEvent file
│       └── events/
│           └── (event records)
│
├── sessions/
│   └── <team-name>/
│       └── <session-id>.json        # Team state snapshot
│
├── plans/
│   └── <plan-id>.json               # Plan file
│
├── worktrees/
│   └── <team-name>/
│       └── <agent-id>/              # Git worktree root
│           ├── .git                 # Worktree link
│           └── (project files)
│
├── templates/
│   ├── leader-worker.yaml
│   ├── leader-worker-reviewer.yaml
│   └── parallel-workers.yaml
│
└── logs/
    └── <team-name>/
        └── <agent-id>.log           # Subprocess backend logs
```

---

## 6. Environment Variables

Environment variables automatically injected when an agent is spawned:

| Variable | Description | Example |
|----------|-------------|---------|
| `JIKIME_AGENT_ID` | Agent unique ID | `worker-1`, `agent-a1b2c3d4` |
| `JIKIME_TEAM_NAME` | Team name | `my-team` |
| `JIKIME_ROLE` | Agent role | `leader`, `worker`, `reviewer` |
| `JIKIME_DATA_DIR` | Data directory path | `~/.jikime` |
| `JIKIME_WORKTREE_PATH` | Git worktree path (if set) | `~/.jikime/worktrees/...` |
| `JIKIME_SPAWN_TIME` | Spawn timestamp (ISO 8601) | `2026-03-20T10:00:00Z` |

Using these variables in agent CLI commands:

```bash
# Example CLI usage inside an agent process
jikime team tasks claim "$JIKIME_TEAM_NAME" <task-id> --agent "$JIKIME_AGENT_ID"
jikime team tasks complete "$JIKIME_TEAM_NAME" <task-id> --agent "$JIKIME_AGENT_ID"
jikime team budget report "$JIKIME_TEAM_NAME" --agent "$JIKIME_AGENT_ID" --input-tokens 1234
```

**External configuration:**

```bash
# Override data directory
export JIKIME_DATA_DIR=/custom/path

# Override Claude binary path
export CLAUDE_PATH=/usr/local/bin/claude
```

---

## 7. Team Workflow Patterns

### 7.1 Basic Pattern: Leader-Worker

```
Leader
  ├── Analyze overall goal
  ├── Decompose into tasks → create tasks
  ├── Workers claim and process tasks
  ├── Review and integrate results
  └── Report completion
```

```bash
# 1. Create team
jikime team create auth-team --budget 100000

# 2. Create initial tasks
jikime team tasks create auth-team "Design API schema" --priority 3
jikime team tasks create auth-team "Implement login endpoint" --priority 2
jikime team tasks create auth-team "Write unit tests" --priority 1

# 3. Spawn agents
jikime team spawn auth-team --role leader --agent-id leader
jikime team spawn auth-team --role worker --agent-id worker-1
jikime team spawn auth-team --role worker --agent-id worker-2

# 4. Monitor
jikime team board live auth-team
```

---

### 7.2 Advanced Pattern: Plan Approval Workflow

```bash
# Worker perspective (inside Claude agent)
jikime team plan submit "$JIKIME_TEAM_NAME" \
  --title "Database Implementation Plan" \
  --file implementation-plan.md

# Leader perspective (reviewing plan)
jikime team plan list --team auth-team
jikime team plan approve plan-abc12345
# or
jikime team plan reject plan-abc12345 --reason "Need distributed approach"
```

---

### 7.3 Git Worktree Isolation Pattern

```bash
# Start team with worktree isolation
jikime team launch --template leader-worker \
  --goal "implement new feature" \
  --name feature-team \
  --worktree

# Each agent:
# - Works on an independent git branch
# - Branch name: jikime-feature-team-<agent-id>
# - Path: ~/.jikime/worktrees/feature-team/<agent-id>/

# Merge after completion
jikime team workspace merge feature-team --agent worker-1 --target develop
jikime team workspace merge feature-team --agent worker-2 --target develop
```

---

### 7.4 CI/CD Integration Pattern

```bash
#!/bin/bash
# ci-team.sh

# Create and start team
jikime team launch --template leader-worker \
  --goal "Fix security vulnerabilities in $PR_TITLE" \
  --name "ci-$PR_NUMBER" \
  --backend subprocess \
  --budget 50000

# Wait for all tasks to complete (max 1 hour)
jikime team tasks wait "ci-$PR_NUMBER" --timeout 3600
EXIT_CODE=$?

# Print budget report
jikime team budget show "ci-$PR_NUMBER"

# Cleanup
jikime team stop "ci-$PR_NUMBER"

exit $EXIT_CODE
```

---

## 8. Git Worktree Isolation

### How It Works

```
Main Repository
├── .git/
├── src/
└── ...
    │
    └── (worktree links)
        │
        ├── ~/.jikime/worktrees/my-team/worker-1/   ← worker-1 workspace
        │   ├── .git    (→ main .git/worktrees/worker-1)
        │   ├── src/    (branch: jikime-my-team-worker-1)
        │   └── ...
        │
        └── ~/.jikime/worktrees/my-team/worker-2/   ← worker-2 workspace
            ├── .git    (→ main .git/worktrees/worker-2)
            ├── src/    (branch: jikime-my-team-worker-2)
            └── ...
```

### Branch Naming Convention

```
jikime-<team-name>-<agent-id>

Examples:
  jikime-auth-team-worker-1
  jikime-auth-team-worker-2
  jikime-auth-team-leader
```

### Benefits of Worktree Isolation

1. **Parallel coding**: Multiple agents work on different files/features simultaneously
2. **Conflict prevention**: Each agent works on an independent branch
3. **Progress preservation**: Each agent's changes are committed immediately
4. **Independent merging**: Each task can be merged to main independently

### Claude Session Path

When agents are started via tmux with `-c <worktreePath>`, Claude sessions are saved to the correct project path:

```
~/.claude/projects/<worktree-path-hash>/
```

---

## 9. GitHub Issues / Harness Integration

Manage GitHub Issues with AI processing in Webchat's **Git Tab → Issues** section.

### 9.1 Requirements

1. **GitHub PAT (Personal Access Token)**: Sidebar Settings → Enter Git PAT
2. **WORKFLOW.md**: Harness configuration file required in project root

### 9.2 GitHub Issues Label System

| Label | Meaning | AI Behavior |
|-------|---------|-------------|
| `jikime-todo` | Issue to be processed by AI | Auto-detected and processed |
| `jikime-done` | Issue processed by AI | Added automatically on completion |

### 9.3 Harness (Automatic Polling)

With WORKFLOW.md configured, Harness periodically detects `jikime-todo` labeled issues and auto-processes them.

```bash
# Check Harness status via API
GET /api/harness/status?projectPath=/path/to/project

# Start Harness
POST /api/harness/start
{ "projectPath": "/path/to/project" }

# Stop Harness
DELETE /api/harness/stop?projectPath=/path/to/project
```

**In Webchat:**
1. Issues tab → ⚡ **Start** button
2. When Harness is running, `🔵 running` banner appears
3. Click **Stop** button to halt

### 9.4 Manual Issue Processing

Process a specific issue with AI immediately:

1. Click issue in Issues list
2. Click **▶ Process with AI** button in right panel
3. Processing log displays in real-time:
   - 🚀 Start message (status display)
   - 🔧 ToolName (tool usage — includes file path/command)
   - Claude text (markdown rendered)
   - ✅ Complete / ❌ Error (status display)

### 9.5 Processing Log Format

The Issues tab processing log displays in the same format as the Chat tab:

```
🚀 Processing issue #42: Add rate limiting
     ↓ Status message (small gray text)

🔧 Read: src/middleware/auth.js
     ↓ Tool bubble (shows file path)

Claude's text response...
     ↓ Orange "C" avatar + markdown rendering

🔧 Edit: src/middleware/rate-limit.js
     ↓ Tool bubble (shows edited file)

✅ Issue #42 processing complete
     ↓ Status message
```

---

## 10. Template System

### 10.1 Built-in Templates

#### `leader-worker`

Structure with 1 leader + N workers.

```yaml
name: leader-worker
description: "Leader coordinates tasks, workers execute"
agents:
  - id: leader
    role: leader
    auto_spawn: true
    task: |
      You are the team leader. Analyze the goal, create tasks,
      and coordinate workers.
  - id: worker-1
    role: worker
    auto_spawn: true
    task: |
      You are a worker. Check available tasks, claim one, and complete it.
tasks:
  - subject: "Analyze requirements"
    description: "Break down the goal into concrete tasks"
    owner: leader
default_budget: 100000
```

#### `leader-worker-reviewer`

Leader + worker + code reviewer structure.

#### `parallel-workers`

Workers operating in parallel without a leader.

---

### 10.2 Creating Custom Templates

Via Webchat **⚙️ → Template Manager** or by creating a file directly:

```yaml
# ~/.jikime/templates/my-custom-template.yaml
name: my-custom-template
description: "My specialized team structure"
version: "1.0.0"

agents:
  - id: architect
    role: leader
    description: "System design and task decomposition"
    auto_spawn: true
    task: |
      You are the system architect. Design the solution and
      create specific tasks for each team member.

  - id: backend-dev
    role: worker
    description: "Backend development specialist"
    auto_spawn: true
    task: |
      You are a backend developer. Focus on API implementation,
      database design, and server-side logic.

  - id: frontend-dev
    role: worker
    description: "Frontend development specialist"
    auto_spawn: true
    task: |
      You are a frontend developer. Focus on UI components,
      state management, and user experience.

  - id: qa-engineer
    role: reviewer
    description: "Quality assurance and testing"
    auto_spawn: true
    task: |
      You are a QA engineer. Write tests, review code quality,
      and ensure all requirements are met.

tasks:
  - subject: "Kickoff meeting"
    description: "Architect defines the implementation plan"
    owner: architect
  - subject: "Project setup"
    description: "Setup project structure and dependencies"

default_budget: 200000
default_max_agents: 4
```

---

## 11. Practical Examples

### 11.1 SaaS Feature Development

```bash
# New payment feature development
jikime team launch \
  --template leader-worker-reviewer \
  --goal "Implement Stripe payment integration with webhook support" \
  --name payment-team \
  --worktree \
  --budget 150000

# Real-time progress monitoring
jikime team board serve payment-team --port 8080
# Browser: http://localhost:8080

# Manually add tasks (even while running)
jikime team tasks create payment-team "Add retry logic for failed payments" \
  --priority 2 \
  --depends-on abc12345

# Wait for completion
jikime team tasks wait payment-team --timeout 7200

# Review results
jikime team budget show payment-team
jikime team board show payment-team

# Merge worktrees
jikime team workspace merge payment-team --agent worker-1 --target main --cleanup
jikime team workspace merge payment-team --agent reviewer --target main --cleanup
```

---

### 11.2 Bug Fix Team

```bash
# Critical bug fix
jikime team create bugfix-team --budget 50000

# Create bug analysis task
jikime team tasks create bugfix-team \
  "Investigate memory leak in connection pool" \
  --priority 5 \
  --dod "Root cause identified and fixed, no memory growth over 1 hour"

# Spawn agent
jikime team spawn bugfix-team \
  --role worker \
  --agent-id debugger \
  --prompt "You are a senior debugging engineer. Investigate the memory leak in src/db/pool.js"

# Monitor
jikime team board live bugfix-team
```

---

### 11.3 Automated Code Review

```bash
# PR code review team
jikime team launch \
  --template parallel-workers \
  --goal "Review PR #$PR_NUMBER: security, performance, code quality" \
  --name "review-pr-$PR_NUMBER" \
  --backend subprocess

# Wait for completion and collect results
jikime team tasks wait "review-pr-$PR_NUMBER" --timeout 1800
jikime team board show "review-pr-$PR_NUMBER" --json > review-results.json
```

---

### 11.4 Using CLI in Agent Prompts

Agents (Claude Code) directly call JIKIME CLI to interact with the team:

```bash
# Example system prompt for a Claude agent
You are a worker agent in team $JIKIME_TEAM_NAME.
Your agent ID is $JIKIME_AGENT_ID.

## Workflow:
1. Check available tasks:
   $ jikime team tasks list $JIKIME_TEAM_NAME --status pending

2. Claim a task:
   $ jikime team tasks claim $JIKIME_TEAM_NAME <task-id> --agent $JIKIME_AGENT_ID

3. Work on the task using your coding tools

4. Commit your progress (if using worktree):
   $ jikime team workspace checkpoint $JIKIME_TEAM_NAME --agent $JIKIME_AGENT_ID

5. Complete the task:
   $ jikime team tasks complete $JIKIME_TEAM_NAME <task-id> \
     --agent $JIKIME_AGENT_ID \
     --result "What was accomplished"

6. Repeat from step 1 until no tasks remain
```

---

## 12. Troubleshooting

### tmux session not found

```bash
# Check all jikime tmux sessions
tmux ls | grep jikime

# Manually attach to a session
tmux attach -t jikime-my-team-worker-1

# Force kill a session
tmux kill-session -t jikime-my-team-worker-1
```

---

### Agent shown as offline

```bash
# Check agent process
jikime team status my-team

# Inspect agent info in registry
cat ~/.jikime/teams/my-team/registry/worker-1.json

# Re-spawn agent
jikime team spawn my-team --role worker --agent-id worker-1
```

---

### Worktree creation error

```bash
# Check existing worktrees
git worktree list

# Prune damaged worktrees
git worktree prune

# Manually remove worktree
jikime team workspace cleanup my-team --agent worker-1
```

---

### Budget exceeded

```bash
# Check current budget status
jikime team budget show my-team

# Increase budget
jikime team budget set my-team 200000

# Stop specific agent
tmux kill-session -t jikime-my-team-worker-2
```

---

### Task stuck in in_progress state

```bash
# Reset task to pending
jikime team tasks update my-team <task-id> --status pending

# Let another agent claim it
jikime team tasks claim my-team <task-id> --agent worker-2
```

---

### Team not visible in Webchat

1. Verify the team is associated with the current project path
2. Check `projectPath` in `~/.jikime/teams/<team-name>/webchat.json`
3. Test API directly: `curl http://localhost:3000/api/team/list?projectPath=/your/project`

---

## Related Documentation

- [agents.md](agents.md) — General Claude Code agent guide
- [agents-team.md](agents-team.md) — Claude Code Agent Teams (experimental feature)
- [worktree.md](worktree.md) — Git Worktree workflow
- [harness-workflow.md](harness-workflow.md) — Harness Engineering workflow
- [hooks.md](hooks.md) — Claude Code hook configuration
- [webchat/usage.md](webchat/usage.md) — Webchat usage guide
