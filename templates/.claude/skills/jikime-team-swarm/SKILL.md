---
name: jikime-team-swarm
description: Claude Code skill for operating as part of a jikime multi-agent team (swarm). Provides command reference, coordination protocols, and role-specific workflows for leader, worker, and reviewer agents. Triggers when the task involves multi-agent coordination, spawning workers, assigning tasks, monitoring team progress, or when the scope exceeds what a single agent can efficiently handle (e.g., "build a full-stack app", "implement multiple features in parallel", "refactor the entire codebase"). Also triggers on keywords: "create a team", "spawn agents", "assign tasks", "team status", "board attach", "agent inbox", "multi-agent", "swarm", "jikime team".
version: "1.0.0"
triggers:
  - "jikime team"
  - "JIKIME_TEAM_NAME"
  - "JIKIME_AGENT_ID"
  - "JIKIME_ROLE"
  - agent swarm
  - multi-agent team
  - team coordination
  - spawn workers
  - task board
  - agent inbox
references:
  - references/cli-reference.md
  - references/workflows.md
---

# JiKiME Team Swarm

You are operating as part of a **jikime multi-agent team**. This skill provides the complete
command reference and coordination protocol for all team roles.

## Your Identity

Your identity is injected as environment variables at spawn time:

```bash
echo $JIKIME_AGENT_ID    # Your agent ID (e.g., "worker-1")
echo $JIKIME_TEAM_NAME   # Your team name
echo $JIKIME_ROLE        # Your role: leader | worker | reviewer
echo $JIKIME_DATA_DIR    # Data directory (~/.jikime)
echo $JIKIME_WORKTREE_PATH  # Your git worktree path (if isolated)
```

---

## Role: Leader

**Activate when**: `$JIKIME_ROLE == "leader"`

### Core Workflow

```bash
# 1. Create tasks for workers
jikime team tasks create $JIKIME_TEAM_NAME "Implement feature X" \
  --desc "Detailed description" \
  --dod "Done when: unit tests pass, PR merged"

# 2. Monitor progress
jikime team status $JIKIME_TEAM_NAME
jikime team board show $JIKIME_TEAM_NAME

# 3. Read inbox (worker updates, join requests)
jikime team inbox receive $JIKIME_TEAM_NAME

# 4. Final integration — do this yourself when ALL tasks are done:
#    a. Review all changed files
#    b. Fix integration issues (missing imports, type errors)
#    c. Run: go build ./... OR npm run build OR equivalent
#    d. Fix build errors
#    e. git add -A && git commit -m "feat: final integration"

# 5. Shutdown team after integration
jikime team inbox broadcast $JIKIME_TEAM_NAME "Integration complete. Shutting down."
jikime team lifecycle shutdown $JIKIME_TEAM_NAME
```

### Handling Join Requests

When a worker sends a join request to your inbox:

```bash
# Approve a join request
jikime team discover approve <agent-id> --team $JIKIME_TEAM_NAME

# Reject a join request
jikime team discover reject <agent-id> --team $JIKIME_TEAM_NAME --reason "team full"
```

### Task Dependency Management

```bash
# Create a task that depends on another
jikime team tasks create $JIKIME_TEAM_NAME "Deploy service" \
  --desc "Deploy after tests pass" \
  --depends-on "<task-id-of-tests>"

# View task details including dependencies
jikime team tasks get $JIKIME_TEAM_NAME <task-id>
```

---

## Role: Worker

**Activate when**: `$JIKIME_ROLE == "worker"`

### Core Workflow (Loop until idle)

```bash
# 1. Check for available tasks
jikime team tasks list $JIKIME_TEAM_NAME --status pending

# 2. Claim a task
jikime team tasks claim $JIKIME_TEAM_NAME <task-id> --agent $JIKIME_AGENT_ID

# 3. Get full task details
jikime team tasks get $JIKIME_TEAM_NAME <task-id>

# 4. Implement the work
#    Read task description and DoD carefully.
#    Work in your worktree if JIKIME_WORKTREE_PATH is set.

# 5. Mark complete with a result summary
jikime team tasks complete $JIKIME_TEAM_NAME <task-id> \
  --agent $JIKIME_AGENT_ID \
  --result "Implemented X: created files A, B, C. All tests pass."

# 6. Notify leader
jikime team inbox send $JIKIME_TEAM_NAME leader \
  "Completed <task-id>: brief one-line summary"

# 7. Repeat from step 1
```

### When No Tasks Remain

```bash
jikime team inbox send $JIKIME_TEAM_NAME leader \
  "No more pending tasks. $JIKIME_AGENT_ID idle."
```

### Worktree Usage

If `$JIKIME_WORKTREE_PATH` is set, all your file changes go there:

```bash
cd $JIKIME_WORKTREE_PATH

# Work normally in your isolated branch
git add -A
git commit -m "feat(<task-id>): implement X"

# Check your branch status
git status
git log --oneline -5
```

---

## Role: Reviewer

**Activate when**: `$JIKIME_ROLE == "reviewer"`

### Core Workflow

```bash
# 1. Check for completed tasks awaiting review
jikime team tasks list $JIKIME_TEAM_NAME --status done

# 2. Get task details and review the work
jikime team tasks get $JIKIME_TEAM_NAME <task-id>

# 3a. Approve — notify leader
jikime team inbox send $JIKIME_TEAM_NAME leader \
  "Approved <task-id>: quality good, no issues."

# 3b. Request revision — requeue task and notify both parties
jikime team tasks update $JIKIME_TEAM_NAME <task-id> --status pending
jikime team inbox send $JIKIME_TEAM_NAME leader \
  "Revision needed <task-id>: [what to fix]"
jikime team inbox send $JIKIME_TEAM_NAME <original-worker-id> \
  "Please revise <task-id>: [specific feedback]"

# 4. Repeat until all done tasks reviewed
```

---

## Inbox (All Roles)

```bash
# Receive and consume pending messages
jikime team inbox receive $JIKIME_TEAM_NAME

# Peek without consuming
jikime team inbox peek $JIKIME_TEAM_NAME

# Send a direct message
jikime team inbox send $JIKIME_TEAM_NAME <to-agent-id> "message body"

# Broadcast to all agents
jikime team inbox broadcast $JIKIME_TEAM_NAME "Announcement message"
```

---

## Team Discovery (All Roles)

```bash
# List all teams on this machine
jikime team discover list

# Join an existing team as a new worker
jikime team discover join my-team --role worker --agent-id my-worker-3

# View all active tmux sessions for a team
jikime team board attach my-team   # Opens linked tmux dashboard
```

---

## Monitoring (All Roles)

```bash
# Team overview
jikime team status $JIKIME_TEAM_NAME

# Task details
jikime team tasks list $JIKIME_TEAM_NAME
jikime team tasks list $JIKIME_TEAM_NAME --status pending
jikime team tasks list $JIKIME_TEAM_NAME --status in_progress

# Wait for all tasks to complete (blocks)
jikime team tasks wait $JIKIME_TEAM_NAME --timeout 3600

# Live board (auto-refreshes)
jikime team board live $JIKIME_TEAM_NAME --interval 5

# Token budget
jikime team budget show $JIKIME_TEAM_NAME
```

---

## Lifecycle

```bash
# Notify leader you're idle (sent automatically at end of worker loop)
jikime team lifecycle idle --agent $JIKIME_AGENT_ID --team $JIKIME_TEAM_NAME

# Clean up on exit (auto-runs via on-exit hook)
jikime team lifecycle on-exit --agent $JIKIME_AGENT_ID --team $JIKIME_TEAM_NAME

# Request shutdown of a specific agent (leader only)
jikime team lifecycle shutdown --agent <agent-id> --team $JIKIME_TEAM_NAME
```

---

## Worktree Collaboration (Workspace Commands)

```bash
# List worktrees for the team
jikime team workspace list $JIKIME_TEAM_NAME

# Checkpoint your work (commit + push worktree branch)
jikime team workspace checkpoint $JIKIME_TEAM_NAME $JIKIME_AGENT_ID

# Merge worktree back to main (leader runs this after integration)
jikime team workspace merge $JIKIME_TEAM_NAME $JIKIME_AGENT_ID

# Clean up a merged worktree
jikime team workspace cleanup $JIKIME_TEAM_NAME $JIKIME_AGENT_ID
```

---

## Available Templates

Launch a pre-configured team:

```bash
# List available templates
jikime team template list

# Standard leader + 2 workers
jikime team launch --template leader-worker --goal "your goal" --name my-team

# Leader + 2 workers + 1 reviewer
jikime team launch --template leader-worker-reviewer --goal "your goal" --name qa-team

# Parallel workers (no leader, self-coordinated)
jikime team launch --template parallel-workers --goal "your goal" --name parallel-team

# 7-agent investment research team
jikime team launch --template hedge-fund --goal "Research best AI stocks for 2025 portfolio" --name invest-team
```

---

## Anti-Patterns to Avoid

| Anti-Pattern | Correct Approach |
|---|---|
| Working directly in main branch | Always use `$JIKIME_WORKTREE_PATH` if set |
| Claiming multiple tasks simultaneously | Claim one task, complete it, then claim next |
| Skipping `inbox send` after completion | Always notify leader after each task |
| Polling inbox every 5 seconds | Check inbox naturally between tasks |
| Modifying files outside your worktree | Stay within your assigned directory |
| Shutting down before final integration | Leader must integrate before shutdown |

---

## Troubleshooting

```bash
# Check if your session is registered
jikime team status $JIKIME_TEAM_NAME

# If task stuck in_progress, release it (leader only)
jikime team tasks update $JIKIME_TEAM_NAME <task-id> --status pending

# Check agent liveness
jikime team board show $JIKIME_TEAM_NAME

# View budget remaining
jikime team budget show $JIKIME_TEAM_NAME
```

---

## Additional Resources

### Reference Files

For complete command arguments, all options, data models, and storage layout:
- **`references/cli-reference.md`** — Full CLI reference: every command, every flag, data model, file storage layout, agent liveness detection

For step-by-step coordination workflows and common patterns:
- **`references/workflows.md`** — 8 complete workflows: team launch, manual setup with dependencies, worker loop, join protocol, plan approval, reviewer-gated QA, graceful shutdown, monitoring & debugging. Plus: dependency chain, fan-out/fan-in, investment research, and worktree isolation patterns.
