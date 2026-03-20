# JiKiME Team CLI — Complete Reference

## Global Options

```
jikime team [--help] <subcommand>
```

Data directory defaults to `~/.jikime`. Override with `JIKIME_DATA_DIR` env var.

---

## Environment Variables

Automatically set when spawned via `jikime team launch` or `jikime team spawn`:

| Variable | Description | Example |
|----------|-------------|---------|
| `JIKIME_AGENT_ID` | Unique agent identifier | `worker-1` |
| `JIKIME_TEAM_NAME` | Team the agent belongs to | `dev-team` |
| `JIKIME_ROLE` | Agent role | `leader`, `worker`, `reviewer` |
| `JIKIME_DATA_DIR` | Data directory override | `~/.jikime` |
| `JIKIME_WORKTREE_PATH` | Isolated git worktree path | `/home/user/jikime/worktrees/...` |
| `JIKIME_SPAWN_TIME` | ISO8601 spawn timestamp | `2026-03-19T10:00:00Z` |

---

## Team Lifecycle (`jikime team`)

### `team launch`

Create team + spawn all agents from a template in one command.

```bash
jikime team launch --template <name> [options]
```

| Option | Description | Default |
|--------|-------------|---------|
| `--template, -t` | Template name (required) | — |
| `--name` | Team name | auto-generated |
| `--goal, -g` | Goal injected into all agent prompts | `""` |
| `--backend, -b` | Spawn backend: `tmux` or `subprocess` | `tmux` |
| `--worktree, -w` | Create isolated git worktree per agent | `false` |
| `--budget` | Token budget override | template default |

Example:
```bash
jikime team launch --template leader-worker --goal "build REST API" --name api-team
```

### `team create`

Create team workspace without spawning agents.

```bash
jikime team create <team-name> [options]
```

### `team spawn`

Spawn a single agent into an existing team.

```bash
jikime team spawn <team-name> [options]
```

| Option | Description | Default |
|--------|-------------|---------|
| `--role, -r` | Agent role: `leader`, `worker`, `reviewer` | `worker` |
| `--agent-id` | Agent ID (auto-generated if empty) | — |
| `--backend, -b` | Spawn backend | `tmux` |
| `--worktree` | Git worktree path | — |
| `--prompt, -p` | Initial prompt | — |
| `--skip-permissions` | Pass `--dangerously-skip-permissions` | `true` |
| `--resume` | Resume previous session | `false` |

### `team status`

Show team overview: agents, tasks, token usage.

```bash
jikime team status <team-name>
```

### `team stop`

Gracefully stop a team and clean up resources.

```bash
jikime team stop <team-name> [--force]
```

---

## Discovery & Join (`jikime team discover`)

### `discover list`

List all teams on this machine with activity status.

```bash
jikime team discover list [--json]
```

Returns: name, template, agent count, task count, active tmux sessions.

### `discover join`

Join an existing team as a new agent. Sends join_request to leader inbox and spawns the agent.

```bash
jikime team discover join <team-name> [options]
```

| Option | Description | Default |
|--------|-------------|---------|
| `--role, -r` | Role to join as | `worker` |
| `--agent-id` | Agent ID | auto-generated |
| `--backend, -b` | Spawn backend | `tmux` |
| `--worktree` | Git worktree path | — |
| `--skip-permissions` | Skip permissions | `true` |

### `discover approve <agent-id>`

Leader approves a pending join request.

```bash
jikime team discover approve <agent-id> --team <team-name>
```

### `discover reject <agent-id>`

Leader rejects a pending join request.

```bash
jikime team discover reject <agent-id> --team <team-name> [--reason TEXT]
```

---

## Inbox (`jikime team inbox`)

### `inbox send`

Send a point-to-point message to an agent.

```bash
jikime team inbox send <team> <to-agent-id> "message body"
```

### `inbox broadcast`

Broadcast a message to all team members.

```bash
jikime team inbox broadcast <team> "message body"
```

### `inbox receive`

Receive and consume pending messages (destructive — messages deleted after read).

```bash
jikime team inbox receive <team> [--agent AGENT_ID] [--limit N]
```

### `inbox peek`

Peek at messages without consuming them (non-destructive).

```bash
jikime team inbox peek <team> [--agent AGENT_ID]
```

### `inbox watch`

Watch inbox for new messages in real-time (blocking, Ctrl+C to stop).

```bash
jikime team inbox watch <team> [--agent AGENT_ID] [--interval SECONDS]
```

---

## Tasks (`jikime team tasks`)

### `tasks create`

Create a new task.

```bash
jikime team tasks create <team> "Task title" [options]
```

| Option | Description | Default |
|--------|-------------|---------|
| `--desc, -d` | Task description | `""` |
| `--dod` | Definition of Done | `""` |
| `--depends-on` | Comma-separated task IDs this depends on | — |
| `--priority, -p` | Priority (higher = more important) | `0` |
| `--tags` | Comma-separated tags | — |

Example:
```bash
jikime team tasks create api-team "Implement auth endpoint" \
  --desc "POST /api/auth/login with JWT response" \
  --dod "Tests pass, OpenAPI spec updated"
```

### `tasks get`

Get full details of a task.

```bash
jikime team tasks get <team> <task-id>
```

### `tasks update`

Update task metadata or status.

```bash
jikime team tasks update <team> <task-id> [options]
```

| Option | Description |
|--------|-------------|
| `--title, -t` | New title |
| `--desc, -d` | New description |
| `--dod` | New definition of done |
| `--priority, -p` | New priority |
| `--status, -s` | New status: `pending` \| `in_progress` \| `done` \| `blocked` |

Setting `--status pending` releases a task back to the queue (re-queues for revision).

### `tasks list`

List tasks with optional filters.

```bash
jikime team tasks list <team> [--status STATUS] [--agent AGENT_ID]
```

| Status Filter | Description |
|---|---|
| `pending` | Available to claim |
| `in_progress` | Currently being worked on |
| `done` | Completed |
| `failed` | Failed |
| `blocked` | Waiting on dependencies |

### `tasks claim`

Claim a pending task for an agent.

```bash
jikime team tasks claim <team> <task-id> --agent <agent-id>
```

Agent ID defaults to `$JIKIME_AGENT_ID` if set.

### `tasks complete`

Mark a task as done with a result summary.

```bash
jikime team tasks complete <team> <task-id> --agent <agent-id> --result "summary"
```

Completing a task automatically unblocks any tasks that depend on it.

### `tasks wait`

Block until all tasks are completed (or timeout).

```bash
jikime team tasks wait <team> [--timeout SECONDS] [--interval SECONDS]
```

| Option | Description | Default |
|--------|-------------|---------|
| `--timeout, -t` | Max wait time in seconds (0 = no limit) | `0` |
| `--interval, -i` | Poll interval in seconds | `5` |

Prints live progress: `tasks: N/M done | wip:X pending:Y blocked:Z`

---

## Board (`jikime team board`)

### `board show`

Snapshot of current team state: agents, task counts, recent tasks.

```bash
jikime team board show <team> [--json]
```

### `board live`

Live-refreshing board (Ctrl+C to stop).

```bash
jikime team board live <team> [--interval SECONDS]
```

Default refresh interval: 3 seconds.

### `board overview`

Summary table of ALL teams on this machine.

```bash
jikime team board overview
```

### `board attach`

Create a tmux dashboard session linking all agent windows for a team.
Navigate agents with `Ctrl-b n` / `Ctrl-b p`. Detach with `Ctrl-b d`.
Agent sessions are unaffected (linked windows, not moved).

```bash
jikime team board attach <team>
```

---

## Plan Approval (`jikime team plan`)

### `plan submit`

Worker submits a plan for leader review.

```bash
jikime team plan submit <team> <agent-id> "plan content" [--summary TEXT]
```

### `plan approve`

Leader approves a submitted plan.

```bash
jikime team plan approve <team> <plan-id> [--feedback TEXT]
```

### `plan reject`

Leader rejects a plan with feedback.

```bash
jikime team plan reject <team> <plan-id> [--feedback TEXT]
```

### `plan list`

List all plans for a team.

```bash
jikime team plan list <team>
```

---

## Lifecycle (`jikime team lifecycle`)

### `lifecycle idle`

Send idle notification to leader (worker has no more tasks).

```bash
jikime team lifecycle idle [--agent AGENT_ID] [--team TEAM_NAME] [--last-task TASK_ID]
```

Reads from `$JIKIME_AGENT_ID` / `$JIKIME_TEAM_NAME` if flags omitted.

### `lifecycle on-exit`

Run on agent process exit to release held tasks and mark agent offline.
This runs automatically via the on-exit hook when a tmux session closes.

```bash
jikime team lifecycle on-exit [--agent AGENT_ID] [--team TEAM_NAME]
```

### `lifecycle shutdown`

Request graceful shutdown of an agent (sends message to agent's inbox).

```bash
jikime team lifecycle shutdown --agent <agent-id> --team <team-name> [--reason TEXT]
```

---

## Templates (`jikime team template`)

### `template list`

List all available templates.

```bash
jikime team template list
```

### `template show`

Show full template definition.

```bash
jikime team template show <name>
```

---

## Workspace / Worktrees (`jikime team workspace`)

### `workspace list`

List all worktrees for a team.

```bash
jikime team workspace list <team>
```

### `workspace checkpoint`

Commit and push the current worktree branch.

```bash
jikime team workspace checkpoint <team> <agent-id>
```

### `workspace merge`

Merge an agent's worktree branch back to main.

```bash
jikime team workspace merge <team> <agent-id>
```

### `workspace cleanup`

Remove a worktree after merge.

```bash
jikime team workspace cleanup <team> <agent-id>
```

---

## Token Budget (`jikime team budget`)

### `budget show`

Show token usage and remaining budget.

```bash
jikime team budget show <team> [--agent AGENT_ID]
```

---

## Data Model

### Task Statuses

| Status | Description |
|--------|-------------|
| `pending` | Available to claim |
| `in_progress` | Currently being worked on |
| `done` | Completed successfully |
| `failed` | Failed with error |
| `blocked` | Waiting on dependency tasks |

Dependency resolution: When a task completes, any task listing it in `DependsOn` has it removed. If `DependsOn` becomes empty, the task moves from `blocked` → `pending` automatically.

### Message Types (inbox subjects)

| Subject | Description |
|---------|-------------|
| `message` | General point-to-point message |
| `broadcast` | Team-wide announcement |
| `join_request` | Agent requests to join team |
| `join_approved` / `join_rejected` | Leader's response to join request |
| `shutdown_request` | Leader requests agent shutdown |
| `idle` | Agent idle notification |

### File Storage Layout

```
~/.jikime/
├── teams/<team>/
│   ├── config.json          # TeamConfig (name, template, budget, timestamps)
│   ├── tasks/
│   │   └── <uuid>.json      # Individual task files (atomic write via tmp+rename)
│   ├── inbox/
│   │   └── <agent-id>/      # Per-agent message queue
│   │       └── msg-<ts>-<uuid>.json
│   ├── registry/
│   │   └── <agent-id>.json  # AgentInfo (status, role, PID, tmux session)
│   ├── costs/
│   │   └── <agent>-<ts>.json  # Token usage events
│   ├── plans/
│   │   └── <id>.json        # Plan submission files
│   ├── sessions/            # Team state snapshots
│   └── events/              # Event log
├── templates/               # User-installed templates (overrides built-in)
│   └── <name>.yaml
└── worktrees/
    └── teams/<team>/<agent>/  # Isolated git worktrees
```

### Agent Liveness Detection (priority order)

1. **tmux session exists** — `tmux has-session -t <session-name>`
2. **PID alive** — `kill -0 <pid>` (no signal sent, just checks existence)
3. **Last heartbeat** — `time.Since(lastHeartbeat) < 30s`
