# Agent Teams - Parallel Team-Based Development

**Parallel Multi-Agent Orchestration using Claude Code Agent Teams**

> Agent Teams is an experimental feature in Claude Code v2.1.32+ that processes complex multi-domain tasks in parallel using team-based approach.

---

## Overview

Agent Teams is a feature where J.A.R.V.I.S./F.R.I.D.A.Y. orchestrators compose multiple specialized agents into teams to perform work in parallel.

### Sub-Agent Approach vs Agent Teams

| Aspect | Sub-Agent Approach | Agent Teams Approach |
|--------|-------------------|---------------------|
| **Execution** | Sequential | Parallel |
| **Communication** | Task() call/return | Real-time collaboration via SendMessage |
| **Task Management** | Orchestrator manages directly | Autonomous distribution via shared TaskList |
| **State** | Stateless | State maintained during team session |
| **Best For** | Single domain, simple tasks | Multi-domain, complex tasks |

---

## Activation Requirements

### Prerequisites

1. **Claude Code Version**: v2.1.32 or higher
2. **Environment Variable**: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
3. **Config File**: `team.enabled: true` in `.jikime/config/workflow.yaml`

```bash
# Set environment variable
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1

# Or add to settings.json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

### Auto-Activation Conditions

Team mode is automatically activated when any of the following conditions are met:

| Condition | Threshold | Description |
|-----------|-----------|-------------|
| **Domain Count** | >= 3 | frontend, backend, database, etc. |
| **File Count** | >= 10 | Number of files to be modified |
| **Complexity Score** | >= 7 | Scale of 1-10 |

---

## Team Agent List

### Plan Phase Agents (Read-Only)

| Agent | Model | Role | Skill |
|-------|-------|------|-------|
| **team-researcher** | haiku | Codebase exploration, architecture analysis | jikime-foundation-philosopher |
| **team-analyst** | inherit | Requirements analysis, edge case identification | jikime-workflow-spec |
| **team-architect** | inherit | Technical design, alternative evaluation | jikime-domain-architecture |

### Run Phase Agents (Implementation Permissions)

| Agent | Model | Role | File Ownership |
|-------|-------|------|----------------|
| **team-backend-dev** | inherit | API, services, business logic implementation | `src/api/**`, `src/services/**` |
| **team-frontend-dev** | inherit | UI components, pages implementation | `src/components/**`, `src/pages/**` |
| **team-designer** | inherit | UI/UX design, design tokens | `design/**`, `src/styles/tokens/**` |
| **team-tester** | inherit | Test writing, coverage verification | `tests/**`, `**/*.test.*` |
| **team-quality** | inherit (read-only) | TRUST 5 verification, quality gates | - |

---

## Team Composition Patterns

### 1. plan_research (Plan Phase)

Parallel research team for SPEC document generation

```yaml
roles:
  - researcher  # Codebase exploration
  - analyst     # Requirements analysis
  - architect   # Technical design
```

**When to Use**: `/jikime:1-plan --team` or auto-detected complexity

### 2. implementation (Run Phase)

Development team for feature implementation

```yaml
roles:
  - backend-dev   # Server side
  - frontend-dev  # Client side
  - tester        # Testing
```

**When to Use**: `/jikime:2-run SPEC-001 --team`

### 3. design_implementation

Feature implementation with UI/UX focus

```yaml
roles:
  - designer      # UI/UX design
  - backend-dev
  - frontend-dev
  - tester
```

### 4. quality_gate

Production deployment with quality verification focus

```yaml
roles:
  - backend-dev
  - frontend-dev
  - tester
  - quality       # TRUST 5 verification
```

### 5. investigation

Competitive hypothesis investigation for complex bugs

```yaml
roles:
  - hypothesis-1
  - hypothesis-2
  - hypothesis-3
model: haiku  # Fast and inexpensive model
```

**When to Use**: `/jikime:build-fix --team`

---

## Workflows

### Team Plan Workflow

```
┌─────────────────────────────────────────────────────────┐
│  Phase 0: TeamCreate("jikime-plan-{feature}")           │
│  ↓                                                       │
│  Phase 1: Parallel Spawn                                 │
│  ├─ Task(team-researcher) ──┐                           │
│  ├─ Task(team-analyst) ─────┼─→ Parallel Execution      │
│  └─ Task(team-architect) ───┘                           │
│  ↓                                                       │
│  Phase 2: Monitoring & Coordination                      │
│  (Real-time collaboration via SendMessage)               │
│  ↓                                                       │
│  Phase 3: Result Integration → SPEC Document Generation  │
│  ↓                                                       │
│  Phase 4: User Approval (AskUserQuestion)               │
│  ↓                                                       │
│  Phase 5: TeamDelete + /clear                           │
└─────────────────────────────────────────────────────────┘
```

### Team Run Workflow

```
┌─────────────────────────────────────────────────────────┐
│  Phase 0: TeamCreate + Task Decomposition               │
│           (File Ownership Assignment)                    │
│  ↓                                                       │
│  Phase 1: Parallel Implementation Team Spawn            │
│  ├─ backend-dev (src/api/**)                            │
│  ├─ frontend-dev (src/components/**)                    │
│  ├─ tester (tests/**)                                   │
│  └─ quality (read-only)                                 │
│  ↓                                                       │
│  Phase 2: Parallel Implementation                        │
│  - SendMessage("api_ready") → frontend starts work      │
│  - SendMessage("component_ready") → tester starts work  │
│  ↓                                                       │
│  Phase 3: Quality Gate                                   │
│  - team-quality performs TRUST 5 verification           │
│  - Pass: complete, Fail: request fixes                  │
│  ↓                                                       │
│  Phase 4: TeamDelete                                    │
└─────────────────────────────────────────────────────────┘
```

---

## Team API Reference

### TeamCreate

Initializes a team session.

```javascript
TeamCreate(team_name: "jikime-plan-auth-feature")
```

### Task (Team Mode)

Creates a teammate. Requires `team_name` and `name` parameters.

```javascript
Task(
  subagent_type: "team-backend-dev",
  team_name: "jikime-run-spec-001",
  name: "backend-dev",
  prompt: "..."
)
```

### SendMessage

Sends messages between teammates or to team lead.

```javascript
// API ready notification
SendMessage(
  recipient: "frontend-dev",
  type: "api_ready",
  content: {
    endpoint: "POST /api/auth/login",
    schema: { email: "string", password: "string" }
  }
)

// Bug report
SendMessage(
  recipient: "backend-dev",
  type: "bug_report",
  content: {
    test: "auth.test.ts:45",
    expected: "200 OK",
    actual: "500 Error"
  }
)

// Shutdown request
SendMessage(
  type: "shutdown_request",
  recipient: "researcher",
  content: "Plan phase complete"
)
```

### TaskCreate/Update/List/Get

Manages shared task list.

```javascript
// Create task
TaskCreate(
  subject: "Implement login API",
  description: "...",
  owner: "backend-dev"
)

// Update status
TaskUpdate(taskId: "1", status: "in_progress")
TaskUpdate(taskId: "1", status: "completed")

// List tasks
TaskList()  // Visible to all teammates

// Get details
TaskGet(taskId: "1")
```

### TeamDelete

Terminates team session. **Must be called after all teammates have shut down.**

```javascript
// First send shutdown request to all teammates
SendMessage(type: "shutdown_request", recipient: "all", ...)

// After confirming teammate termination
TeamDelete(team_name: "jikime-run-spec-001")
```

---

## File Ownership

To prevent file conflicts in team mode, each teammate owns specific file patterns.

```yaml
file_ownership:
  team-backend-dev:
    - "src/api/**"
    - "src/services/**"
    - "src/repositories/**"
    - "src/models/**"
    - "src/middleware/**"
    - "prisma/migrations/**"

  team-frontend-dev:
    - "src/components/**"
    - "src/pages/**"
    - "src/app/**"
    - "src/hooks/**"
    - "src/stores/**"
    - "src/styles/**"

  team-designer:
    - "design/**"
    - "src/styles/tokens/**"

  team-tester:
    - "tests/**"
    - "__tests__/**"
    - "**/*.test.*"
    - "**/*.spec.*"
    - "cypress/**"
    - "playwright/**"

  shared:  # Requires coordination via SendMessage
    - "src/types/**"
    - "src/utils/**"
    - "src/lib/**"
```

### Conflict Resolution

When you need to modify a file you don't own:

```javascript
// ❌ Don't modify directly

// ✅ Request from file owner
SendMessage(
  recipient: "backend-dev",
  type: "change_request",
  content: {
    file: "src/types/user.ts",
    requested_change: "Add 'refreshToken' field",
    reason: "Needed for token refresh flow"
  }
)
```

---

## Hook Events

### TeammateIdle

Called when a teammate completes work and becomes idle.

```yaml
TeammateIdle:
  exit_code_0: "Accept idle state - no more work"
  exit_code_2: "Reject idle - assign additional work from TaskList"
```

### TaskCompleted

Called when a teammate marks a task as completed.

```yaml
TaskCompleted:
  exit_code_0: "Accept completion - end task"
  exit_code_2: "Reject completion - additional work required"
  validation:
    - Tests pass
    - Coverage target met
    - No lint errors
```

---

## Usage Examples

### Using Team Mode in Plan Phase

```bash
# Explicit team mode
/jikime:1-plan "Implement user authentication system" --team

# Auto-detection (activates automatically if complexity is high)
/jikime:1-plan "Complex multi-domain feature"
```

### Using Team Mode in Run Phase

```bash
# Explicit team mode
/jikime:2-run SPEC-001 --team

# Team with designer
/jikime:2-run SPEC-001 --team --pattern=design_implementation
```

### Using Team Mode for Debugging

```bash
# Competitive hypothesis investigation
/jikime:build-fix --team

# Investigate multiple hypotheses in parallel
```

### Force Sub-Agent Mode

```bash
# Disable team mode (for simple tasks)
/jikime:2-run SPEC-001 --solo
```

---

## Configuration File

### .jikime/config/workflow.yaml

```yaml
workflow:
  execution_mode: "auto"  # auto | subagent | team

  team:
    enabled: true
    max_teammates: 10
    default_model: "inherit"
    require_plan_approval: true
    delegate_mode: true
    teammate_display: "auto"

    auto_selection:
      min_domains_for_team: 3
      min_files_for_team: 10
      min_complexity_score: 7

    file_ownership:
      team-backend-dev:
        - "src/api/**"
        - "src/services/**"
      team-frontend-dev:
        - "src/components/**"
        - "src/pages/**"
      team-tester:
        - "tests/**"

    patterns:
      plan_research:
        roles: [researcher, analyst, architect]
      implementation:
        roles: [backend-dev, frontend-dev, tester]

    hooks:
      teammate_idle:
        enabled: true
        validate_work: true
      task_completed:
        enabled: true
        require_quality_check: true
```

---

## Fallback Behavior

When team mode fails or requirements are not met:

1. **Warning log** output
2. **Auto-switch to Sub-Agent mode**
3. **Resume from last completed task**
4. **No data loss**

### Fallback Trigger Conditions

- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` environment variable not set
- `workflow.team.enabled: false`
- TeamCreate failure
- Teammate spawn failure
- Network error

---

## Related Documentation

- [J.A.R.V.I.S. Orchestrator](./jarvis.md)
- [F.R.I.D.A.Y. Orchestrator](./friday.md)
- [Agent Catalog](./agents.md)
- [SPEC Workflow](./spec-workflow.md)
- [DDD Development Methodology](./tdd-ddd.md)

---

## Version Information

| Item | Value |
|------|-------|
| **Document Version** | 1.0.0 |
| **Required Claude Code** | v2.1.32+ |
| **Status** | Experimental |
| **Last Updated** | 2026-02-14 |
