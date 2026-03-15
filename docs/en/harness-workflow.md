# Harness Workflow — Plans.md-Based Skill System

> Plan → Work → Review → Ship automation loop — transforms Claude Code's open-ended execution into a structured, guardrailed workflow.

---

## Concept

**Harness Workflow** is a Claude Code workflow system built around five verb-based skills that form a continuous loop:

```
harness-plan    → Create and manage Plans.md
harness-work    → Implement tasks from Plans.md (cc:WIP → cc:DONE)
harness-review  → 4-perspective code review (pm:REVIEW → pm:OK)
harness-sync    → Sync Plans.md ↔ git history + generate retrospective
harness-release → Automate releases from completed tasks
```

**Comparison with `jikime serve`:**

| | jikime serve | Harness Workflow |
|--|-------------|-----------------|
| Purpose | GitHub Issue → PR full automation | In-session task management + quality gates |
| Entry point | Daemon process (always running) | Slash commands inside Claude Code session |
| Task tracking | GitHub Issues | Plans.md (project-local) |
| Quality control | WORKFLOW.md prompt | 4-perspective review + DoD validation |
| Best fit | Issue-sized tasks (simple to medium) | Feature/milestone scope (medium to complex) |

---

## Plans.md — Single Source of Truth

All Harness Workflow operations revolve around the `Plans.md` file.

### Marker System

| Marker | Meaning | When to use |
|--------|---------|-------------|
| `cc:TODO` | Not started | Default |
| `cc:WIP` | In progress | When Worker starts a task |
| `cc:DONE [hash]` | Complete + git hash | After Worker commits |
| `pm:REVIEW` | Awaiting user review | After Reviewer completes review |
| `pm:OK` | Review approved | After user approves |
| `blocked:<reason>` | Blocked | On dependency or external block |
| `cc:SKIP` | Skipped | When task becomes unnecessary |

### Plans.md Format

```markdown
# Plans.md

## Overview

| Field | Value |
|-------|-------|
| **Goal** | [Project/feature objective] |
| **Milestone** | [Completion date or event] |
| **Owner** | [Username / Claude] |
| **Created** | [YYYY-MM-DD] |

---

## Phase 1: Core Features

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1  | Task description | Completion criteria (Yes/No judgeable) | - | cc:TODO |
| 1.2  | Task description | Completion criteria | 1.1 | cc:TODO |

## Phase 2: Quality Improvements

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 2.1  | Task description | Completion criteria | 1.2 | cc:TODO |
```

### DoD Writing Rules

| ✅ Good DoD | ❌ Bad DoD |
|------------|-----------|
| `Tests pass (npm test: 0 failed)` | `Works fine` |
| `Zero lint errors` | `Code quality improved` |
| `API /login returns 200 confirmed via curl` | `Complete` |
| `All cc:DONE tasks have valid git hash` | `ok` |

---

## Implementation Status

### Phase 1: Plans.md System Foundation ✅

> Status: **Complete** (v1.5.x)

**Implemented files:**

| File | Description |
|------|-------------|
| `templates/Plans.md` | Standard Plans.md template |
| `templates/.claude/skills/jikime-harness-plan/SKILL.md` | harness-plan skill |
| `templates/.claude/skills/jikime-harness-sync/SKILL.md` | harness-sync skill |
| `cmd/hookscmd/plans_watcher.go` | Plans.md structure validation hook |
| `templates/.claude/settings.json` | `plans-watcher` hook registered |

#### harness-plan Skill

Creates and manages Plans.md.

**Sub-commands:**

```
/jikime:harness-plan create   → Create new Plans.md (interactive)
/jikime:harness-plan add      → Add tasks/phases to existing Plans.md
/jikime:harness-plan update 1.2 cc:WIP   → Update task marker
/jikime:harness-plan sync     → Lightweight git-history-based state inference
```

**create execution flow:**
1. Auto-detect tech stack (package.json, go.mod, etc.)
2. Collect requirements (max 3 questions)
3. Phase classification: Required(1) / Recommended(2) / Optional(3)
4. Infer DoD for each task
5. Map task dependencies
6. Generate Plans.md (all tasks as `cc:TODO`)

#### harness-sync Skill

Full synchronization between Plans.md and git history.

**Sub-commands:**

```
/jikime:harness-sync sync    → Full sync (detect mismatches + propose updates)
/jikime:harness-sync retro   → Generate 4-item retrospective
/jikime:harness-sync trace   → Analyze Agent Trace for specific task
/jikime:harness-sync drift   → Detect drift between Plans.md and actual implementation
```

**4-Item Retrospective:**

| Item | Content |
|------|---------|
| Estimation Accuracy | Planned vs. completed ratio, scope creep measurement |
| Blocking Analysis | blocked marker analysis, average block duration |
| Quality Marker Hit Rate | cc:DONE → pm:OK conversion rate, re-review ratio |
| Scope Drift | Plans.md change history, task addition/removal patterns |

#### plans-watcher Hook

Automatically runs whenever Plans.md is written or edited.

```
PostToolUse (Write|Edit) → plans-watcher → structure validation
```

**Validation checks:**
- Marker syntax validation (`cc:TODO`, `cc:WIP`, `cc:DONE [hash]`, etc.)
- Vague DoD detection (warns when DoD cannot be judged Yes/No)
- Dependency reference validation (detects references to non-existent Task IDs)
- Task status summary

**Example output:**
```
📋 Plans.md: 8 tasks [TODO:3 | WIP:1 | DONE:4]

⚠️ Plans.md validation warnings:
  - Line 12: Task 2.1 has vague DoD: `complete` — DoD must be Yes/No judgeable
```

---

### Phase 2: harness-work + 3 Agent Team ✅

> Status: **Complete** (v1.5.x)

**Implemented files:**

| File | Description |
|------|-------------|
| `templates/.claude/skills/jikime-harness-work/SKILL.md` | harness-work skill |
| `templates/.claude/skills/jikime-harness-review/SKILL.md` | harness-review skill |
| `templates/.claude/agents/jikime/harness-worker.md` | Worker Agent |
| `templates/.claude/agents/jikime/harness-reviewer.md` | Reviewer Agent |
| `templates/.claude/agents/jikime/harness-scaffolder.md` | Scaffolder Agent |

#### harness-work Skill

Executes Plans.md tasks through the `cc:WIP → cc:DONE` lifecycle.

```
/jikime:harness-work 1.2              → Execute task 1.2 solo
/jikime:harness-work 1.2 1.3         → Parallel execution
/jikime:harness-work --auto           → Auto-select TODO tasks
/jikime:harness-work --mode breezing  → Force Breezing Mode (4+ tasks)
```

**Execution flow:**
```
Phase 0: Validate Plans.md (check dependencies)
Phase 1: Scaffolder → Select mode (Solo/Parallel/Breezing)
Phase 2: Scaffolder → Task analysis + scaffolding
Phase 3: Worker → Implement + validate DoD (cc:WIP → cc:DONE)
Phase 4: Reviewer → 4-perspective review
Phase 5: User notification (pm:REVIEW request)
```

**Breezing Mode:**

| Condition | Mode | Description |
|-----------|------|-------------|
| 1 task | Solo | Single Worker, sequential |
| 2–3 tasks, independent | Parallel | 2–3 Workers, parallel worktrees |
| 4+ tasks | Breezing | Lead + Workers + Reviewer team |

#### harness-review Skill

Runs 4-perspective code review on `cc:DONE` tasks and manages `pm:REVIEW → pm:OK`.

```
/jikime:harness-review 1.2           → Review task 1.2
/jikime:harness-review --all         → Review all pm:REVIEW tasks
/jikime:harness-review --approve 1.2 → Direct pm:OK approval
```

#### 3 Agent Team

| Agent | Role | Permissions |
|-------|------|-------------|
| **harness-worker** | Implementation, DoD validation, Plans.md marker updates | Read/Write/Edit/Bash |
| **harness-reviewer** | 4-perspective review (security/perf/quality/DoD), never modifies code | Read/Grep/Glob/Bash |
| **harness-scaffolder** | Task analysis, scaffolding, mode selection, state transitions | Read/Write/Edit/Bash |

---

### Phase 3: Declarative Guardrail Engine ✅

> Status: **Complete** (v1.5.x)

**Implemented files:**

| File | Description |
|------|-------------|
| `cmd/hookscmd/guardrail_engine.go` | R01–R08 rule evaluation engine (PostToolUse hook) |
| `templates/.claude/settings.json` | `guardrail-engine` hook registered |

**Runs automatically whenever Plans.md is written or edited.**

#### Rules

| Rule | Level | Description | Detection |
|------|-------|-------------|-----------|
| **R02** | error | More than 2 simultaneous cc:WIP tasks | WIP marker count |
| **R03** | error | WIP task has incomplete dependencies | Dependency graph validation |
| **R05** | error | pm:OK set without prior pm:REVIEW | git log -S search |
| **R06** | warn | Task blocked for 5+ days | git log timestamp analysis |
| **R07** | warn | Task count grew >30% from first commit | Compare git first commit vs. current |
| **R08** | warn | Later-phase task completed while earlier phase has unfinished tasks | Phase order inversion detection |

> R01 (Plans.md required), R04 (DoD validation), and R09 (Breezing Mode) are handled by plans-watcher hook, Worker Agent, and Scaffolder Agent respectively.

#### Example output

```
🚨 Harness Guardrail violations (2):

[R02] WIP concurrency exceeded: 3 tasks are cc:WIP (1.1, 1.2, 1.3). Max 2 recommended.
  → Complete or block pending tasks first.

[R03] Dependency incomplete: Task 1.2 is cc:WIP but dependency 1.1 is still cc:WIP.
  → Complete 1.1 first.

⚠️  Harness Guardrail warnings (1):

[R08] Phase order inversion: Phase 2 task is complete but Phase 1 Task 1.1 is cc:WIP.
  → Complete Phase 1 tasks first.
```

---

### Phase 4: harness-setup + harness-release ✅

> Status: **Complete** (v1.5.x)

**Implemented files:**

| File | Description |
|------|-------------|
| `templates/.claude/skills/jikime-harness-setup/SKILL.md` | harness-setup skill |
| `templates/.claude/skills/jikime-harness-release/SKILL.md` | harness-release skill |

#### harness-setup Skill

Initializes Harness Engineering workflow for new or existing projects.

```
/jikime:harness-setup           → Full setup (interactive)
/jikime:harness-setup --check   → Diagnostic only
/jikime:harness-setup --reset   → Reset and reconfigure
```

**5-step execution flow:**
1. Environment diagnostic (git repo, jikime binary, settings.json, Plans.md, hooks)
2. Plans.md creation guidance (if not found)
3. Hook registration validation and auto-add (plans-watcher, guardrail-engine)
4. Git environment preparation (worktree support, .gitignore update)
5. Setup completion report

#### harness-release Skill

Automates releases from `pm:OK` tasks in Plans.md.

```
/jikime:harness-release                → Interactive release (auto version suggestion)
/jikime:harness-release --patch        → Patch version (v1.0.0 → v1.0.1)
/jikime:harness-release --minor        → Minor version (v1.0.0 → v1.1.0)
/jikime:harness-release --major        → Major version (v1.0.0 → v2.0.0)
/jikime:harness-release --dry-run      → Preview release (no changes)
/jikime:harness-release --version v2.1.0 → Specify version directly
```

**7-step execution flow:**
1. Collect pm:OK tasks (prerequisites: no cc:WIP, no pm:REVIEW, clean working directory)
2. Version determination (auto-detect from package.json, go.mod, VERSION file, git tags + bump suggestion)
3. CHANGELOG.md generation/update (commit prefix → section mapping: feat→✨, fix→🐛, etc.)
4. Version file update
5. Plans.md release record append
6. Git commit + tag creation (user confirmation before push)
7. Release completion report

---

## How to Use

### Basic Flow

```
1. Project Setup    → /jikime:harness-setup
2. Create Plans.md  → /jikime:harness-plan create
3. Implement tasks  → /jikime:harness-work
4. Code review      → /jikime:harness-review
5. Sync             → /jikime:harness-sync
6. Release          → /jikime:harness-release
```

### Step 1: Initial Setup (once per project)

Run this when starting a new project.

```
/jikime:harness-setup
```

Validates git repo, hook registration, and fixes any issues automatically.
To check current status only:

```
/jikime:harness-setup --check
```

### Step 2: Create Plans.md

Describe the feature you want to build — Claude decomposes it into structured tasks.

```
/jikime:harness-plan create "Implement JWT Authentication"
```

After up to 3 clarifying questions, Claude generates Plans.md:

```markdown
## Phase 1: Core Features
| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1  | JWT token generation | jwt.sign() works, expiry test passes | - | cc:TODO |
| 1.2  | Login API | POST /auth/login returns 200 | 1.1 | cc:TODO |
| 1.3  | Middleware validation | Distinguishes valid/expired/forged tokens | 1.2 | cc:TODO |
```

To add tasks to an existing Plans.md:

```
/jikime:harness-plan add "Phase 2: Social Login"
```

### Step 3: Implement Tasks

Once Plans.md is ready, start implementing.

```bash
# Single task
/jikime:harness-work 1.1

# Parallel execution (2–3 independent tasks)
/jikime:harness-work 1.1 1.3

# Auto-select TODO tasks
/jikime:harness-work --auto
```

What happens internally:
1. Scaffolder analyzes tasks + selects execution mode (Solo/Parallel/Breezing)
2. Worker implements in an isolated worktree
3. Plans.md markers update automatically: `cc:TODO → cc:WIP → cc:DONE [abc1234]`
4. Reviewer runs 4-perspective review (security/performance/quality/DoD)
5. On completion, status becomes `pm:REVIEW` + user notification

### Step 4: Review and Approve

When Worker completes, Plans.md shows `pm:REVIEW` status.

```bash
# View review details
/jikime:harness-review 1.1

# Approve (mark as pm:OK)
/jikime:harness-review --approve 1.1

# Process all pm:REVIEW tasks at once
/jikime:harness-review --all
```

### Step 5: Sync & Retrospective

Verify Plans.md matches actual git history.

```bash
/jikime:harness-sync sync    # Detect mismatches + propose updates
/jikime:harness-sync retro   # Generate 4-item retrospective
/jikime:harness-sync drift   # Detect divergence between Plans.md and actual implementation
```

Sample retrospective output:

```
📊 Retrospective v1.2.0

Estimation Accuracy: Planned 8 tasks → Completed 8 tasks (100%)
Blocking Analysis:   0 blocked incidents
Quality Hit Rate:    cc:DONE → pm:OK conversion 100%
Scope Drift:         No tasks added (stable)
```

### Step 6: Release

When all tasks reach `pm:OK`, run the release.

```bash
# Interactive release (auto version suggestion)
/jikime:harness-release

# Preview only (no changes)
/jikime:harness-release --dry-run

# Specify version bump
/jikime:harness-release --minor   # v1.0.0 → v1.1.0
/jikime:harness-release --patch   # v1.0.0 → v1.0.1
```

Claude automatically generates CHANGELOG.md, updates version files, creates `git commit + tag`, then confirms before pushing.

---

### End-to-End Example

```bash
# 1. Setup
/jikime:harness-setup

# 2. Plan
/jikime:harness-plan create "User Authentication System"

# 3. Implement (work through Plans.md)
/jikime:harness-work 1.1
/jikime:harness-review --approve 1.1

/jikime:harness-work 1.2 1.3     # parallel execution
/jikime:harness-review --all

# 4. Wrap up
/jikime:harness-sync retro
/jikime:harness-release --minor
```

### Quick Reference by Situation

| Situation | Command |
|-----------|---------|
| Check remaining tasks | Open Plans.md or `/jikime:harness-plan update` |
| Implementation blocked | Manually change task to `blocked:<reason>` |
| Add new tasks | `/jikime:harness-plan add` |
| Plans.md out of sync with git | `/jikime:harness-sync sync` |
| Preview release before executing | `/jikime:harness-release --dry-run` |

> **Key principle**: Plans.md is always the Single Source of Truth (SSOT). Claude automatically updates markers to track progress — you just review and approve.

---

## Full Workflow

```
User: /jikime:harness-plan create "Implement JWT Authentication"
        ↓
Claude: Generates Plans.md (Phase 1–3, 12 tasks, all cc:TODO)
        ↓
User: /jikime:harness-work 1.1
        ↓
Worker Agent: isolated worktree, implement, test, commit
              Plans.md: 1.1 → cc:WIP → cc:DONE [abc1234]
        ↓
Reviewer Agent: 4-perspective review (security/performance/quality/docs)
                Plans.md: 1.1 → pm:REVIEW
        ↓
User: Reviews and approves
        Plans.md: 1.1 → pm:OK
        ↓
(Repeat for next tasks...)
        ↓
User: /jikime:harness-sync retro
        ↓
Claude: Generates 4-item retrospective
        Appends to Plans.md
        ↓
User: /jikime:harness-release
        ↓
Claude: Updates CHANGELOG, creates tag, writes release notes
```

---

## Slash Command Mapping

| Command | Skill | Phase |
|---------|-------|-------|
| `/jikime:1-plan` | harness-plan create | Plan |
| `/jikime:2-run` | harness-work | Work |
| `/jikime:3-sync` | harness-sync | Sync |
| `/jikime:harness-plan` | harness-plan | Plan management |

---

## Related

- [Harness Engineering (jikime serve)](./harness-engineering.md) — GitHub Issue automation daemon
- [Hooks System](./hooks.md)
- [Skills Catalog](./skills-catalog.md)
- [POC-First Workflow](./poc-first.md)

---

Last Updated: 2026-03-15
Version: Phase 4 Complete
