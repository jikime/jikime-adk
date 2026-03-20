# JiKiME Team — Coordination Workflows

---

## Workflow 1: Launch a Team from Template

The fastest way to start a multi-agent team.

```bash
# Standard 1-leader + 2-workers team
jikime team launch \
  --template leader-worker \
  --goal "Implement a REST API with JWT authentication" \
  --name api-team

# With isolated git worktrees per agent (recommended for code tasks)
jikime team launch \
  --template leader-worker \
  --goal "Refactor the authentication module" \
  --name refactor-team \
  --worktree

# Monitor immediately
jikime team status api-team
jikime team board live api-team --interval 3

# Watch all agents simultaneously
jikime team board attach api-team
```

---

## Workflow 2: Manual Team Setup with Dependencies

When you need fine-grained control over task ordering.

```bash
# 1. Create team workspace
jikime team create api-team

# 2. Spawn leader
jikime team spawn api-team --role leader --agent-id leader

# 3. Create tasks with dependency chain
#    Design -> Backend (blocked) -> Frontend (blocked) -> Integration (blocked)
jikime team tasks create api-team "Design API schema" --priority 10
# => Task ID: aaa11111

jikime team tasks create api-team "Implement backend endpoints" \
  --depends-on aaa11111 \
  --desc "POST /auth/login, GET /users/:id" \
  --dod "All endpoints return correct status codes and schemas"
# => Task ID: bbb22222 (auto-set to blocked)

jikime team tasks create api-team "Build frontend auth UI" \
  --depends-on aaa11111 \
  --desc "Login form, token storage, protected routes"
# => Task ID: ccc33333 (auto-set to blocked)

jikime team tasks create api-team "Integration testing" \
  --depends-on bbb22222,ccc33333 \
  --dod "E2E tests pass, all acceptance criteria met"
# => Task ID: ddd44444 (blocked until both bbb and ccc complete)

# 4. Spawn workers
jikime team spawn api-team --role worker --agent-id worker-1
jikime team spawn api-team --role worker --agent-id worker-2

# 5. When design is done, backend and frontend auto-unblock
jikime team tasks complete api-team aaa11111 --agent leader \
  --result "API schema defined: auth/login, users CRUD"
# => bbb22222 and ccc33333 move from blocked -> pending automatically

# 6. Wait for all tasks to complete
jikime team tasks wait api-team --timeout 7200
```

---

## Workflow 3: Worker Agent Loop

The canonical loop every worker agent follows inside its tmux session.

```bash
# Identity is pre-set via JIKIME_* environment variables
echo "I am $JIKIME_AGENT_ID on team $JIKIME_TEAM_NAME"

# Check inbox for leader instructions first
jikime team inbox receive $JIKIME_TEAM_NAME

# Main work loop
while true; do
  # Find available tasks
  TASKS=$(jikime team tasks list $JIKIME_TEAM_NAME --status pending)

  if [ -z "$TASKS" ]; then
    echo "No pending tasks found."
    break
  fi

  # Claim first available task
  TASK_ID=$(echo "$TASKS" | head -1 | awk '{print $1}')
  jikime team tasks claim $JIKIME_TEAM_NAME $TASK_ID --agent $JIKIME_AGENT_ID

  # Get full task details
  jikime team tasks get $JIKIME_TEAM_NAME $TASK_ID

  # ... do the work ...

  # Mark complete
  jikime team tasks complete $JIKIME_TEAM_NAME $TASK_ID \
    --agent $JIKIME_AGENT_ID \
    --result "Implemented X: created files A, B, C. All tests pass."

  # Notify leader
  jikime team inbox send $JIKIME_TEAM_NAME leader \
    "Completed $TASK_ID: brief one-line summary"
done

# Notify leader when idle
jikime team inbox send $JIKIME_TEAM_NAME leader \
  "No more pending tasks. $JIKIME_AGENT_ID idle."
```

---

## Workflow 4: Join Request Protocol

When a new agent joins an existing running team dynamically.

```bash
# === New Agent Side ===

# Request to join an existing team
jikime team discover join dev-team --role worker --agent-id specialist-007

# Check inbox for approval/rejection
jikime team inbox receive dev-team

# If approved: begin claiming tasks
jikime team tasks list dev-team --status pending
jikime team tasks claim dev-team <task-id> --agent specialist-007

# === Leader Side ===

# Check inbox — see join_request
jikime team inbox receive dev-team

# Approve
jikime team discover approve specialist-007 --team dev-team

# Or reject
jikime team discover reject specialist-007 --team dev-team --reason "team capacity reached"
```

---

## Workflow 5: Plan Approval Flow

For teams requiring plan review before execution (risk-sensitive tasks).

```bash
# === Worker Side: submit plan ===
jikime team plan submit dev-team $JIKIME_AGENT_ID \
  "1. Refactor auth module\n2. Add OAuth2 provider\n3. Update integration tests" \
  --summary "Auth system modernization"

# Wait for leader's decision
jikime team inbox receive dev-team

# === Leader Side: review plan ===
jikime team inbox receive dev-team
# => sees plan_approval_request

jikime team plan list dev-team
jikime team plan approve dev-team <plan-id>
# or
jikime team plan reject dev-team <plan-id> --feedback "Add error handling section first"
```

---

## Workflow 6: Reviewer-Gated Quality Control

Using the `leader-worker-reviewer` template for QA-gated completion.

```bash
jikime team launch \
  --template leader-worker-reviewer \
  --goal "Build checkout flow with payment integration" \
  --name checkout-team

# Worker completes task -> status becomes "done"
# Reviewer checks done tasks:
jikime team tasks list $JIKIME_TEAM_NAME --status done

# Reviewer approves good work:
jikime team inbox send $JIKIME_TEAM_NAME leader \
  "Approved <task-id>: implementation correct, tests adequate."

# Reviewer sends back for revision:
jikime team tasks update $JIKIME_TEAM_NAME <task-id> --status pending
jikime team inbox send $JIKIME_TEAM_NAME leader \
  "Revision needed <task-id>: missing input validation on card number."
jikime team inbox send $JIKIME_TEAM_NAME worker-1 \
  "Please revise <task-id>: add Luhn algorithm check for card validation."
```

---

## Workflow 7: Graceful Shutdown (Leader Orchestrated)

```bash
# Leader: integration complete, notify all agents
jikime team inbox broadcast $JIKIME_TEAM_NAME \
  "Integration complete. All agents please wrap up and idle."

# Wait for idle confirmations (check inbox)
jikime team inbox receive $JIKIME_TEAM_NAME

# Send individual shutdown requests
jikime team lifecycle shutdown --agent worker-1 --team $JIKIME_TEAM_NAME --reason "All tasks done"
jikime team lifecycle shutdown --agent worker-2 --team $JIKIME_TEAM_NAME --reason "All tasks done"

# Final shutdown
jikime team lifecycle shutdown --team $JIKIME_TEAM_NAME
```

**Note**: Agents running in tmux automatically run `jikime team lifecycle on-exit` when their
session closes, releasing held tasks and marking themselves offline.

---

## Workflow 8: Monitoring and Debugging

```bash
# Quick overview of all teams on this machine
jikime team board overview

# Detailed snapshot of one team
jikime team board show dev-team

# JSON output for scripting
jikime team board show dev-team --json | jq '.tasks'
jikime team tasks list dev-team --json | jq '.[].title'

# Live monitoring (auto-refreshes every 3s)
jikime team board live dev-team --interval 3

# Watch agents work in split-pane view
jikime team board attach dev-team   # Ctrl-b n/p to switch, Ctrl-b d to detach

# Check agent liveness
jikime team status dev-team

# Check budget remaining
jikime team budget show dev-team

# Watch a specific agent's inbox in real-time
jikime team inbox watch dev-team --agent leader

# Find stuck tasks
jikime team tasks list dev-team --status in_progress
jikime team tasks list dev-team --status blocked

# Release a stuck in_progress task back to pending
jikime team tasks update dev-team <task-id> --status pending
```

---

## Common Patterns

### Task Dependency Chain: Sequential Pipeline

```bash
# A -> B -> C (strict sequence)
jikime team tasks create team "Research" --priority 10
# => aaa

jikime team tasks create team "Implement" --depends-on aaa
# => bbb (blocked until aaa done)

jikime team tasks create team "Deploy" --depends-on bbb
# => ccc (blocked until bbb done)

# As each completes, next auto-unblocks
jikime team tasks complete team aaa --agent leader --result "..."
# => bbb moves to pending

jikime team tasks complete team bbb --agent worker-1 --result "..."
# => ccc moves to pending
```

### Fan-Out / Fan-In Pattern

```bash
# One coordinator task -> N parallel tasks -> one merge task
jikime team tasks create team "Plan architecture"
# => plan-task

jikime team tasks create team "Implement module A" --depends-on plan-task
# => task-a

jikime team tasks create team "Implement module B" --depends-on plan-task
# => task-b

jikime team tasks create team "Implement module C" --depends-on plan-task
# => task-c

jikime team tasks create team "Integrate all modules" --depends-on task-a,task-b,task-c
# => merge-task (blocked until ALL 3 are done)
```

### Investment Research Pattern (hedge-fund template)

```bash
jikime team launch \
  --template hedge-fund \
  --goal "Research best semiconductor stocks for Q3 2025 portfolio" \
  --name invest-q3

# Portfolio manager creates sector research tasks
# 5 analysts work in parallel on their sectors
# Risk manager reviews combined portfolio
# Portfolio manager synthesizes final report

jikime team board attach invest-q3  # Watch all 7 agents
```

### Worktree Isolation Pattern (for code tasks)

```bash
# Each agent works in its own branch, no conflicts
jikime team launch \
  --template leader-worker \
  --goal "Implement microservices: user-service, order-service, payment-service" \
  --name micro-team \
  --worktree

# Each worker's changes go to: ~/.jikime/worktrees/teams/micro-team/<agent>/
# After all tasks done, leader merges worktrees:
jikime team workspace list micro-team
jikime team workspace merge micro-team worker-1
jikime team workspace merge micro-team worker-2
jikime team workspace cleanup micro-team worker-1
jikime team workspace cleanup micro-team worker-2
```

### Inbox Watch Pattern (real-time coordination)

```bash
# Leader watches for worker updates without polling
jikime team inbox watch $JIKIME_TEAM_NAME --agent leader --interval 2

# In a separate window, workers send updates
jikime team inbox send $JIKIME_TEAM_NAME leader "Completed task-123: auth endpoint done"
```
