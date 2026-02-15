# J.A.R.V.I.S. - Intelligent Autonomous Orchestration

JikiME-ADK's intelligent autonomous orchestration system. Proactive and adaptive development automation inspired by Iron Man's AI assistant.

## Overview

J.A.R.V.I.S. (Just A Rather Very Intelligent System) is JikiME-ADK's **development-dedicated** intelligent orchestrator. Rather than simple command execution, it is an autonomous system that **predicts, adapts, and learns**, while migration tasks are handled by the partner orchestrator F.R.I.D.A.Y.

### Core Philosophy

```
"I'm not just following orders, sir. I'm anticipating your needs."
```

### Differences from Existing Orchestrators

| Feature | Existing Orchestrators | J.A.R.V.I.S. |
|---------|------------------------|--------------|
| Exploration | 3 agents in parallel | 5 agents + dependency analysis |
| Planning | Single strategy | Multi-strategy comparison and optimal selection |
| Execution | Fixed sequential/parallel | Situation-adaptive dynamic switching |
| Error Handling | Simple retry | Self-diagnosis + alternative strategy pivot |
| Learning | None | In-session pattern learning |
| Prediction | None | Proactive suggestions for next steps |

## Architecture

### System Structure

```
┌─────────────────────────────────────────────────────────────────┐
│                    J.A.R.V.I.S. System                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Phase 0: Proactive Intelligence Gathering                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐│
│  │ Explore  │ │ Research │ │ Quality  │ │ Security │ │  Perf  ││
│  │  Agent   │ │  Agent   │ │  Agent   │ │  Agent   │ │ Agent  ││
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └───┬────┘│
│       └────────────┴────────────┼────────────┴───────────┘     │
│                                 ▼                               │
│                    ┌────────────────────┐                       │
│                    │ Integration Engine │                       │
│                    │  + Dependency Map  │                       │
│                    │  + Risk Assessment │                       │
│                    └─────────┬──────────┘                       │
│                              ▼                                  │
│  Phase 1: Multi-Strategy Planning                               │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ Strategy A   │ │ Strategy B   │ │ Strategy C   │            │
│  │ Conservative │ │  Balanced    │ │  Aggressive  │            │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘            │
│         └────────────────┼────────────────┘                    │
│                          ▼                                      │
│                ┌──────────────────┐                             │
│                │ Trade-off Matrix │                             │
│                │ Optimal Selection│                             │
│                └────────┬─────────┘                             │
│                         ▼                                       │
│  Phase 2: Adaptive DDD Implementation (Ralph Loop)              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  WHILE (issues_exist AND iteration < max):              │   │
│  │    ├── LSP Quality Gate (baseline capture/compare)      │   │
│  │    ├── Diagnostics (LSP + Tests + Coverage)             │   │
│  │    ├── Self-Assessment: "Is approach working?"          │   │
│  │    │   ├── YES → Continue                               │   │
│  │    │   ├── REGRESSION → Ralph alerts → Pivot            │   │
│  │    │   └── NO  → Pivot Strategy                         │   │
│  │    ├── Expert Agent Delegation                          │   │
│  │    └── Verification (zero errors required)              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                         ▼                                       │
│  Phase 3: Completion & Prediction                               │
│  ┌──────────────┐ ┌──────────────────────┐                     │
│  │  Doc Sync    │ │ Predictive Suggest   │                     │
│  │ manager-docs │ │ "You might also..."  │                     │
│  └──────────────┘ └──────────────────────┘                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Related Files

| File | Description |
|------|-------------|
| `templates/.claude/commands/jikime/jarvis.md` | J.A.R.V.I.S. slash command |
| `templates/.claude/commands/jikime/friday.md` | F.R.I.D.A.Y. slash command |
| `templates/CLAUDE.md` | Dual orchestrator registration |
| `templates/.jikime/config/quality.yaml` | LSP Quality Gates configuration |
| `docs/friday.md` | F.R.I.D.A.Y. migration orchestrator documentation |
| `docs/migration.md` | Migration system documentation |

## Dual Orchestrator Architecture

### Dual Orchestration System

JikiME-ADK separates roles with two specialized orchestrators:

| Orchestrator | Role | Command | Scope |
|---|---|---|---|
| **J.A.R.V.I.S.** | Development | `/jikime:jarvis` | New feature implementation, refactoring, bug fixes |
| **F.R.I.D.A.Y.** | Migration | `/jikime:friday` | Legacy → modern framework conversion |

> For migration-related detailed documentation, refer to `docs/friday.md` and `docs/migration.md`.

### Routing Logic

```
IF migration keywords detected (migrate, convert, legacy, transform):
    → F.R.I.D.A.Y. activation
ELIF development keywords detected (implement, build, fix, refactor):
    → J.A.R.V.I.S. activation
ELSE:
    → Default: J.A.R.V.I.S.
```

## Usage

### Basic Usage

```bash
# Intelligent autonomous execution
/jikime:jarvis "Add JWT authentication"

# Safe strategy (conservative approach)
/jikime:jarvis "Refactor payment module" --strategy safe

# Fast strategy (aggressive approach)
/jikime:jarvis "Fix typo in README" --strategy fast

# Enable automatic loop
/jikime:jarvis "Implement user dashboard" --loop --max 20

# Resume previous work
/jikime:jarvis resume SPEC-AUTH-001
```

### Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `--strategy` | Execution strategy: auto, safe, fast | auto |
| `--loop` | Enable automatic error fix loop | config |
| `--max N` | Maximum iteration count | 50 |
| `--branch` | Auto-create feature branch | config |
| `--pr` | Auto-create PR on completion | config |
| `--resume SPEC` | Resume previous work | - |

## Intelligence Features

### 1. Proactive Intelligence Gathering (Phase 0)

5 specialized agents perform analysis **simultaneously**:

| Agent | Role | Output |
|-------|------|--------|
| **Explore Agent** | Codebase structure, architecture pattern analysis | Related file list, dependency map |
| **Research Agent** | External documentation, library best practices | Implementation patterns, API references |
| **Quality Agent** | Test coverage, code quality baseline | Quality metrics, technical debt assessment |
| **Security Agent** | Pre-scan potential security impacts | Security considerations, OWASP checklist |
| **Performance Agent** | Performance impact prediction analysis | Bottleneck risks, optimization opportunities |

### 2. Multi-Strategy Planning (Phase 1)

For all tasks, 2-3 approach strategies are generated and compared:

#### Strategy Types

| Strategy | Risk | Speed | Rollback | Test Coverage |
|----------|------|-------|----------|---------------|
| **Conservative** | Low | Slow | Easy | 100% |
| **Balanced** | Medium | Medium | Medium | 85% |
| **Aggressive** | High | Fast | Difficult | 70% |

#### Automatic Strategy Selection Algorithm

```
IF risk_score > 70:
    SELECT Conservative (safety first)
ELIF risk_score > 40:
    SELECT Balanced (balance)
ELSE:
    SELECT Aggressive (speed first)

OVERRIDE: Can be manually specified with --strategy flag
```

### 3. Adaptive Execution (Phase 2)

#### LSP Quality Gates

During Phase 2 execution, LSP-based quality gates are automatically applied:

| Phase | Condition | Description |
|-------|-----------|-------------|
| **plan** | `require_baseline: true` | Capture LSP baseline at phase start |
| **run** | `max_errors: 0` | Requires zero errors/type errors/lint errors |
| **sync** | `require_clean_lsp: true` | Clean LSP state required before PR/Sync |

Configuration location: `.jikime/config/quality.yaml` → `constitution.lsp_quality_gates`

#### Ralph Loop Integration

J.A.R.V.I.S.'s self-diagnosis loop integrates with LSP Quality Gates:

```
Ralph Loop Cycle:
  1. Code Transformation (agent task execution)
  2. LSP Diagnostic Capture (post-transformation diagnostics)
  3. Regression Check (comparison against baseline)
  4. Decision: Continue or Pivot
```

When LSP regression is detected, J.A.R.V.I.S. automatically considers pivoting.

#### Self-Diagnosis Loop

With each iteration, J.A.R.V.I.S. asks itself:

1. **"Is the current approach showing progress?"**
   - Is the error count decreasing?
   - Is the test pass rate improving?
   - Are LSP diagnostic results improving?

2. **"Should I switch to a different strategy?"**
   - Trigger: 3 consecutive iterations without improvement
   - Trigger: LSP regression detected
   - Action: Pivot to alternative strategy

3. **"Is this a pattern I've seen before?"**
   - Check for similar error patterns within session
   - Immediately apply learned solutions

#### Pivot Decision Tree

```
IF no_progress_count >= 3:
    IF current_strategy == "aggressive":
        PIVOT → "balanced"
    ELIF current_strategy == "balanced":
        PIVOT → "conservative"
    ELSE:
        REQUEST → user_intervention
```

### 4. Predictive Suggestions (Phase 3)

Predicts and suggests next steps based on completed work:

```markdown
## Completed: JWT Authentication

### Predictive Suggestions

Based on this implementation, you might also want to:

1. **Add refresh token mechanism** - Extend session when JWT token expires
2. **Implement rate limiting** - Prevent brute force attacks on auth endpoints
3. **Add password reset flow** - Common feature paired with authentication
4. **Set up audit logging** - Track authentication events for security

Would you like me to start any of these?
```

## Strategy Details

### auto (Default)

J.A.R.V.I.S. analyzes task complexity and automatically selects the optimal strategy:

| Task Type | Analysis Result | Selected Strategy |
|-----------|-----------------|-------------------|
| Simple (single domain) | Risk < 40 | Direct expert delegation |
| Medium (2-3 domains) | Risk 40-70 | Sequential workflow |
| Complex (4+ domains) | Risk > 70 | Full parallel orchestration |

### safe (Conservative)

Applies maximum validation and safeguards:

- More user checkpoints
- Smaller incremental changes
- Comprehensive testing at each step
- Rollback points at every Phase

### fast (Aggressive)

Fast execution for simple or urgent tasks:

- Minimal checkpoints
- Parallelize everything possible
- Skip optional validation
- Prioritize quick completion

## Agent Delegation

### Delegation Rules

[HARD] All implementation tasks must be delegated to expert agents:

| Task Type | Assigned Agent |
|-----------|----------------|
| Backend logic | backend |
| Frontend components | frontend |
| Full-stack features (DB→API→UI) | fullstack |
| Test generation | test-guide |
| Bug fixes | debugger |
| Refactoring | refactorer |
| Security fixes | security-auditor |
| Performance optimization | optimizer |
| API design | specialist-api |
| GraphQL schema | specialist-graphql |
| Microservices architecture | specialist-microservices |
| Mobile apps (RN/Flutter) | specialist-mobile |
| Desktop apps (Electron) | specialist-electron |
| Real-time communication (WebSocket) | specialist-websocket |
| UI design system | designer-ui |
| Angular 15+ development | specialist-angular |
| JavaScript/Node.js development | specialist-javascript |
| PHP/Laravel/Symfony development | specialist-php |
| Python/FastAPI/Django development | specialist-python |
| Rust system programming | specialist-rust |
| SQL query optimization | specialist-sql |
| TypeScript advanced types | specialist-typescript |
| Vue/Nuxt development | specialist-vue |

## Output Format

### During Execution

```markdown
## J.A.R.V.I.S.: Phase 2 (Iteration 3/50)

### Strategy: Balanced (auto-selected)
### Risk Score: 45/100

### Current Status
- [x] User model created
- [x] JWT token generation
- [ ] Login endpoint ← in progress
- [ ] Token validation middleware

### Self-Assessment
- Progress: YES (2 errors → 1 error)
- Pivot needed: NO
- Confidence: 85%

### Active Issues
- ERROR: src/auth/login.ts:45 - undefined 'hashPassword'

Fixing...
```

### On Completion

```markdown
## J.A.R.V.I.S.: COMPLETE

### Summary
- SPEC: SPEC-AUTH-001
- Strategy Used: Balanced
- Files Modified: 12
- Tests: 34/34 passing
- Coverage: 92%
- Iterations: 5
- Self-Corrections: 1 (pivoted from aggressive at iteration 3)

### Changes Implemented
- JWT token generation and validation
- Login/logout endpoints
- Token refresh mechanism
- Authentication middleware
- Comprehensive test suite

### Predictive Suggestions
1. Add rate limiting to auth endpoints
2. Implement password reset flow
3. Set up audit logging

<jikime>DONE</jikime>
```

## Workflow Integration

### Dual Orchestrator Workflow

JikiME-ADK operates two orchestrators independently according to their roles:

```
J.A.R.V.I.S. (Development):
  Individual:  /jikime:0-project → /jikime:1-plan → /jikime:2-run → /jikime:3-sync
  Integrated:  /jikime:jarvis "task" → Full automatic execution

F.R.I.D.A.Y. (Migration):
  Individual:  /jikime:migrate-0-discover → 1-analyze → 2-plan → 3-execute → 4-verify
  Integrated:  /jikime:friday "migrate to X" → Full automatic execution
```

### Command System

| Type | Commands | Orchestrator | Purpose |
|------|----------|--------------|---------|
| **Workflow (Type A)** | 0-project, 1-plan, 2-run, 3-sync | J.A.R.V.I.S. | Fine-grained control per development stage |
| **Migration** | migrate-0 ~ migrate-4 | F.R.I.D.A.Y. | Stage-by-stage migration control |
| **Utility (Type B)** | jarvis, test, loop, fix | J.A.R.V.I.S. | Quick execution and automation |
| **Utility (Type B)** | friday | F.R.I.D.A.Y. | Migration automation |

### Orchestrator Comparison

| Aspect | J.A.R.V.I.S. (Development) | F.R.I.D.A.Y. (Migration) |
|--------|----------------------------|--------------------------|
| **Purpose** | New feature implementation, improvements | Legacy → modern framework conversion |
| **Input** | Task description, SPEC | Legacy source code |
| **Stages** | 4 stages (0-project ~ 3-sync) | 5 stages (0-discover ~ 4-verify) |
| **Output** | Code, documentation | Migrated project, verification report |
| **Methodology** | DDD | DDD + behavioral comparison verification |
| **Completion Marker** | `<jikime>DONE</jikime>` | `<jikime>MIGRATION_COMPLETE</jikime>` |

## Limitations & Safety

### Limitations

- Maximum 3 strategy pivots (requests user intervention after that)
- No pivoting during critical operations (migration, deletion)
- Only in-session learning supported (no cross-session learning)

### Safeguards

- [HARD] All implementations delegated to expert agents
- [HARD] User confirmation required before SPEC creation
- [HARD] Completion marker required: `<jikime>DONE</jikime>`
- [HARD] LSP Quality Gate: Zero errors required in run phase
- Rollback points created at each Phase
- Automatic alerts when LSP Quality Gates detect regression

## Related Commands & Features

### Parallel Execution Features

Commands that can be used in conjunction with J.A.R.V.I.S.'s 5-way parallel exploration:

| Command | Description | Parallel Execution Pattern |
|---------|-------------|---------------------------|
| `/jikime:perspective` | Simultaneous analysis from 4 perspectives (Arch, Sec, Perf, Test) | 4-way parallel sub-agents |
| `/jikime:verify pre-pr` | Verification including Adversarial Review | 3-way parallel verification |

### Multi-Perspective Analysis

The `/jikime:perspective` command uses a parallel analysis pattern similar to J.A.R.V.I.S.'s Intelligence Gathering:

```
J.A.R.V.I.S. Phase 0:        /jikime:perspective:
├── Explore Agent            ├── Architecture Agent
├── Research Agent           ├── Security Agent
├── Quality Agent      ←→    ├── Performance Agent
├── Security Agent           └── Testing Agent
└── Performance Agent
```

### Adversarial Review

The Adversarial Review executed in Phase 8 of `/jikime:verify pre-pr` runs 3 sub-agents in parallel for verification:

| Subagent | Role |
|----------|------|
| **False Positive Filter** | False positive filtering |
| **Missing Issues Finder** | Missed issue detection |
| **Context Validator** | Context verification |

### Skill Reference

Detailed documentation on parallel execution patterns:
- `templates/.claude/skills/jikime-workflow-parallel/SKILL.md`
- `templates/.claude/skills/jikime-workflow-parallel/modules/parallel-patterns.md`
- `templates/.claude/skills/jikime-workflow-parallel/modules/synthesis-strategies.md`

---

## Best Practices

### When Should You Use J.A.R.V.I.S.?

**Suitable cases:**
- New feature implementation (spanning multiple domains)
- Large-scale refactoring
- Complex bug fixes
- When full workflow automation is needed

**When individual commands are better:**
- Single file modifications
- When you want to execute only specific stages
- When fine-grained control is needed

### Recommended Usage Patterns

```bash
# Complex new feature
/jikime:jarvis "Implement payment processing system"

# Safety-critical refactoring
/jikime:jarvis "Refactor database layer" --strategy safe

# Simple fix
/jikime:jarvis "Add validation to login form" --strategy fast

# Resume support for long tasks
/jikime:jarvis "Complex feature" --loop --max 30
# ... after interruption ...
/jikime:jarvis resume SPEC-XXX

# Use F.R.I.D.A.Y. for migrations
/jikime:friday "Migrate Vue app to Next.js 16"
```

---

Version: 3.2.0
Last Updated: 2026-01-25
Codename: J.A.R.V.I.S. (Just A Rather Very Intelligent System)
Inspiration: Iron Man's AI Assistant
Changelog:
- v3.2.0: Added Related Commands section (perspective, verify), linked parallel execution pattern documentation
- v3.1.0: Updated LSP Quality Gates configuration path (quality.yaml), documentation improvements
- v3.0.0: Dual Orchestrator (J.A.R.V.I.S. + F.R.I.D.A.Y.), LSP Quality Gates, Ralph Loop integration
- v2.0.0: Added Migration Mode (--mode migrate), unified workflow orchestration
- v1.0.0: Initial release with Development Mode
