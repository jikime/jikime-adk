# JikiME-ADK Command Reference

A comprehensive reference document for all slash commands in JikiME-ADK.

## Overview

JikiME-ADK provides three types of commands:

| Type | Description | Commands |
|------|------|--------|
| **Type A: Workflow** | Core development workflow | 0-project, 1-plan, 2-run, 3-sync |
| **Type B: Utility** | Quick execution and automation | jarvis, test, loop, verify |
| **Standalone** | Independent execution utilities | architect, build-fix, cleanup, codemap, docs, e2e, learn, perspective, refactor, security |
| **Generator** | Skill and code generation | skill-create, migration-skill |
| **Migration** | Legacy migration | migrate, migrate-0~4 |

---

## Command Map

```
                    ┌─────────────────────────────────────┐
                    │        JikiME-ADK Commands          │
                    └─────────────────────────────────────┘
                                    │
    ┌───────────────┬───────────────┼───────────────┬───────────────┐
    │               │               │               │               │
    ▼               ▼               ▼               ▼               ▼
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│Workflow │   │ Utility │   │Generator│   │Standalone│   │Migration│
│(Type A) │   │(Type B) │   │         │   │         │   │         │
└─────────┘   └─────────┘   └─────────┘   └─────────┘   └─────────┘
    │               │               │               │               │
 0-project       jarvis      skill-create      architect        migrate
    ↓            test       migration-skill   build-fix      migrate-0~4
  1-plan         loop                          cleanup
    ↓           verify                         codemap
  2-run                                          docs
    ↓                                            e2e
  3-sync                                        learn
                                            perspective
                                              refactor
                                              security
```

---

## Type A: Workflow Commands

Commands that comprise the core development workflow.

### /jikime:0-project

**Project initialization and document generation**

| Item | Content |
|------|------|
| **Description** | Project setup and document generation through codebase analysis |
| **Type** | Workflow (Type A) |
| **Context** | - |
| **Agent Chain** | manager-project → Explore → manager-docs |

#### Usage

```bash
/jikime:0-project
```

#### Process

```
PHASE 0: Project type detection
    └─ New Project / Existing Project / Migration Project
         ↓
PHASE 0.5: Information gathering (New/Migration)
    └─ manager-project: Collect project settings
         ↓
PHASE 1: Codebase analysis (Existing)
    └─ Explore: Analyze structure, tech stack, core features
         ↓
PHASE 2: User confirmation
    └─ AskUserQuestion: Approve analysis results
         ↓
PHASE 3: Document generation
    └─ manager-docs: Generate product.md, structure.md, tech.md
         ↓
PHASE 3.5: Development environment check
    └─ Check and guide LSP server installation
         ↓
PHASE 4: Complete
    └─ Next step guidance
```

#### Output Files

- `.jikime/project/product.md` - Product overview, features, user value
- `.jikime/project/structure.md` - Project architecture and directory structure
- `.jikime/project/tech.md` - Tech stack, dependencies, technical decisions

---

### /jikime:1-plan

**SPEC definition and development branch creation**

| Item | Content |
|------|------|
| **Description** | Define requirements as EARS format SPEC document |
| **Type** | Workflow (Type A) |
| **Context** | planning.md |
| **Agent Chain** | Explore (optional) → manager-spec → manager-git (conditional) |

#### Usage

```bash
# Generate SPEC only (default)
/jikime:1-plan "User authentication system"

# SPEC + Git branch
/jikime:1-plan "User authentication system" --branch

# SPEC + Git Worktree (parallel development)
/jikime:1-plan "User authentication system" --worktree
```

#### Options

| Option | Description |
|--------|------|
| `--branch` | Auto-create feature branch |
| `--worktree` | Create isolated development environment with Git Worktree |

#### Process

```
PHASE 1: Project analysis & SPEC planning
    └─ manager-spec: Generate SPEC candidates, design EARS structure
         ↓
PHASE 1.5: Pre-validation
    └─ SPEC type classification, ID format validation, duplicate check
         ↓
PHASE 2: SPEC document generation
    └─ Generate spec.md, plan.md, acceptance.md
         ↓
PHASE 3: Git branch/Worktree setup (conditional)
    └─ Create branch or worktree based on flags
```

#### Output Files

```
.jikime/specs/SPEC-{ID}/
├── spec.md        # Core specification (EARS format)
├── plan.md        # Implementation plan
└── acceptance.md  # Acceptance criteria (Given/When/Then)
```

---

### /jikime:2-run

**DDD-based SPEC implementation**

| Item | Content |
|------|------|
| **Description** | Implement SPEC with ANALYZE-PRESERVE-IMPROVE cycle |
| **Type** | Workflow (Type A) |
| **Context** | dev.md |
| **Agent Chain** | manager-strategy → manager-ddd → manager-quality → manager-git |

#### Usage

```bash
# Standard execution
/jikime:2-run SPEC-AUTH-001

# Execute after creating checkpoint
/jikime:2-run SPEC-AUTH-001 --checkpoint

# Force Personal/Team mode
/jikime:2-run SPEC-AUTH-001 --personal
/jikime:2-run SPEC-AUTH-001 --team
```

#### Options

| Option | Description |
|--------|------|
| `--checkpoint` | Create recovery point before start |
| `--skip-quality` | Skip quality verification (not recommended) |
| `--personal` | Force Personal git mode |
| `--team` | Force Team git mode |

#### DDD Cycle

```
┌─────────────┐
│   ANALYZE   │  ← Understand current behavior
└──────┬──────┘
       ↓
┌─────────────┐
│  PRESERVE   │  ← Preserve behavior with characterization tests
└──────┬──────┘
       ↓
┌─────────────┐
│   IMPROVE   │  ← Change with confidence
└──────┬──────┘
       ↓
    (Repeat)
```

#### Process

```
PHASE 1: Strategy analysis
    └─ manager-strategy: Establish implementation strategy
         ↓
PHASE 1.5: Task decomposition
    └─ Task tracking with TodoWrite
         ↓
PHASE 2: DDD implementation
    └─ manager-ddd: Execute DDD cycle for each task
         ↓
PHASE 2.5: Quality verification
    └─ manager-quality: Test coverage, lint, type check
         ↓
PHASE 3: Git operations
    └─ manager-git: Commit, PR creation (team mode)
         ↓
PHASE 4: Complete
    └─ Result report, next step guidance
```

---

### /jikime:3-sync

**Document synchronization and SPEC completion**

| Item | Content |
|------|------|
| **Description** | Sync code changes with documentation, SPEC completion processing |
| **Type** | Workflow (Type A) |
| **Context** | sync.md |
| **Agent Chain** | manager-quality → manager-docs → manager-git |

#### Usage

```bash
# Sync specific SPEC
/jikime:3-sync SPEC-AUTH-001

# Auto-execute without prompts
/jikime:3-sync SPEC-AUTH-001 --auto

# Regenerate documents
/jikime:3-sync SPEC-AUTH-001 --force

# Check all SPEC status
/jikime:3-sync --status

# Sync project documents only
/jikime:3-sync --project
```

#### Options

| Option | Description |
|--------|------|
| `--auto` | Auto-execute without prompts |
| `--force` | Regenerate all documents |
| `--status` | Display all SPEC sync status |
| `--project` | Sync project-level documents only |

#### Process

```
PHASE 0.5: Pre-quality check
    └─ Verify implementation complete, tests passed
         ↓
PHASE 1: Document analysis
    └─ Scan code changes, map document impacts
         ↓
PHASE 2: Document update
    └─ manager-docs: Update product.md, structure.md, tech.md, CHANGELOG
         ↓
PHASE 3: SPEC completion
    └─ Change status to "completed", generate completion summary
         ↓
PHASE 4: PR/Merge management (Team mode)
    └─ Update PR, provide merge options
         ↓
PHASE 5: Complete
    └─ Sync report, next step guidance
```

---

## Type B: Utility Commands

Commands for quick execution and automation.

### /jikime:jarvis

**J.A.R.V.I.S. - Intelligent Autonomous Orchestration**

| Item | Content |
|------|------|
| **Description** | Intelligent orchestrator inspired by Iron Man's AI assistant |
| **Type** | Utility (Type B) |
| **Context** | - |
| **Features** | 5-way parallel exploration, multi-strategy comparison, adaptive execution, predictive suggestions |

#### Usage

```bash
# Basic usage (auto strategy)
/jikime:jarvis "Add JWT authentication"

# Safe strategy (conservative)
/jikime:jarvis "Refactor payment module" --strategy safe

# Fast strategy (aggressive)
/jikime:jarvis "Fix typo in README" --strategy fast

# Enable auto loop
/jikime:jarvis "Implement user dashboard" --loop --max 20

# Resume previous work
/jikime:jarvis resume SPEC-AUTH-001
```

#### Options

| Option | Description | Default |
|--------|------|--------|
| `--strategy` | Execution strategy: auto, safe, fast | auto |
| `--loop` | Enable auto error fix loop | config |
| `--max N` | Maximum iterations | 50 |
| `--branch` | Auto-create feature branch | config |
| `--pr` | Auto-create PR on completion | config |
| `--resume SPEC` | Resume previous work | - |

#### Strategy Comparison

| Strategy | Risk | Speed | Rollback | Test Coverage |
|------|--------|------|----------|----------------|
| **Conservative** | Low | Slow | Easy | 100% |
| **Balanced** | Medium | Medium | Medium | 85% |
| **Aggressive** | High | Fast | Difficult | 70% |

#### Autonomous Flow

```
PHASE 0: Preemptive information gathering (5-way parallel)
    ├── Explore Agent: Codebase structure
    ├── Research Agent: External docs, best practices
    ├── Quality Agent: Current state diagnosis
    ├── Security Agent: Pre-scan security impact
    └── Performance Agent: Performance impact prediction
         ↓
PHASE 1: Multi-strategy planning
    ├── Strategy A: Conservative
    ├── Strategy B: Balanced
    └── Strategy C: Aggressive
    └─ Trade-off analysis → Select optimal strategy
         ↓
PHASE 2: Adaptive DDD implementation
    └─ Self-diagnosis loop (progress check, strategy pivot decision)
         ↓
PHASE 3: Completion & prediction
    └─ Document sync + predictive next step suggestions
```

---

### /jikime:test

**Test execution utility**

| Item | Content |
|------|------|
| **Description** | Quick execution of unit/integration tests |
| **Type** | Utility (Type B) |
| **Context** | - |
| **Related Command** | `/jikime:e2e` (E2E tests) |

#### Usage

```bash
# Run all tests
/jikime:test

# Include coverage report
/jikime:test --coverage

# Specific test type only
/jikime:test --unit
/jikime:test --integration

# Watch mode
/jikime:test --watch

# Auto-fix failing tests
/jikime:test --fix
```

#### Options

| Option | Description |
|--------|------|
| `--coverage` | Generate coverage report |
| `--unit` | Run unit tests only |
| `--integration` | Run integration tests only |
| `--watch` | Watch mode for continuous testing |
| `--fix` | Auto-fix failing tests when possible |

#### Coverage Targets

| Type | Target |
|------|------|
| Business Logic | 90%+ |
| API Endpoints | 80%+ |
| UI Components | 70%+ |
| Overall | 80%+ |

---

### /jikime:loop

**Ralph Loop - Iterative improvement based on LSP/AST-grep feedback**

| Item | Content |
|------|------|
| **Description** | Progressive code improvement with intelligent feedback loop |
| **Type** | Utility (Type B) |
| **Context** | debug.md |
| **Skill** | jikime-workflow-loop |

#### Usage

```bash
# Basic usage (fix all errors)
/jikime:loop "Fix all TypeScript errors"

# With options
/jikime:loop "Remove security vulnerabilities" --max-iterations 5 --zero-security

# Specific directory
/jikime:loop @src/services/ "Fix all lint errors" --zero-warnings

# Until tests pass
/jikime:loop "Fix failing tests" --tests-pass --max-iterations 10

# Cancel active loop
/jikime:loop --cancel
```

#### Options

| Option | Description | Default |
|--------|------|--------|
| `--max-iterations` | Maximum iteration count | 10 |
| `--zero-errors` | Require zero errors | true |
| `--zero-warnings` | Require zero warnings | false |
| `--zero-security` | Require zero security issues | false |
| `--tests-pass` | Require all tests to pass | false |
| `--stagnation-limit` | Iterations without improvement limit | 3 |
| `--cancel` | Cancel active loop | - |

#### Process

```
1. Initialize Loop
   jikime hooks start-loop --task "..." --options ...
        ↓
2. Load Skill
   Skill("jikime-workflow-loop")
        ↓
3. Execute Iteration
   - Analyze current state
   - Fix issues one by one
   - Collect LSP/AST-grep feedback
        ↓
4. Stop Hook Evaluation
   - Check completion conditions
   - Calculate improvement rate
   - Decide Continue or Complete
        ↓
5. (Continue) Reinject feedback
        ↓
6. (Complete) Generate final report
```

---

### /jikime:verify

**Comprehensive quality verification**

| Item | Content |
|------|------|
| **Description** | Verify build, types, lint, tests, and security at once |
| **Type** | Utility (Type B) |
| **Context** | - |
| **Features** | LSP Quality Gates, TRUST 5 framework, Adversarial Review integration |

#### Usage

```bash
# Standard verification (recommended)
/jikime:verify

# Quick check (build + types only)
/jikime:verify quick

# Full verification (all checks + deps)
/jikime:verify full

# Pre-PR verification (full + security + Adversarial Review)
/jikime:verify pre-pr

# Attempt auto-fix
/jikime:verify --fix

# CI/CD mode (exit codes)
/jikime:verify --ci

# JSON output (for automation)
/jikime:verify --json

# Check changed files only
/jikime:verify --incremental
```

#### Verification Profiles

| Profile | Verification Items | Use Case |
|---------|----------|------|
| `quick` | Build, Types | Quick check during development |
| `standard` | Build, Types, Lint, Tests | Default, after changes |
| `full` | All + Deps, Coverage | Before major commits |
| `pre-pr` | Full + Security + Adversarial | Before PR creation |

#### Verification Phases

| Phase | Verification Content | Gate |
|-------|----------|------|
| 1. Build | Compilation success | FAIL → Immediate stop |
| 2. Type Check | TypeScript/Pyright | Error → Fix required before PR |
| 3. Lint | ESLint/Ruff | Error → Fix, Warning → Document |
| 4. Test Suite | Including coverage | Fail → Fix, <80% → Warning |
| 5. Security | Secrets, vulnerabilities | Secret → CRITICAL |
| 6. LSP Gates | Quality thresholds | Regression → Block PR |
| 7. TRUST 5 | 5-principle compliance | Report non-compliant items |
| 8. Adversarial | 3-way parallel verification (pre-pr, full only) | Adjusted severity |

#### Adversarial Review (v1.1.0+)

In `pre-pr` and `full` profiles, 3 subagents run **in parallel**:

| Subagent | Role |
|----------|------|
| **False Positive Filter** | Identify false positives in Phase 1-7 results |
| **Missing Issues Finder** | Detect missed issues from new perspectives |
| **Context Validator** | Compare with original intent, verify pattern consistency |

```
Example results:
- False Positives Filtered: 2 warnings (test fixtures)
- Missing Issues Found: 1 race condition
- Context Validated: ✅

Adjusted issues: 3 warnings → 1 warning (after filtering)
New issues: 1 (race condition in async handler)
```

---

## Standalone Utility Commands

Utility commands that can be used independently of the workflow.

### /jikime:architect

**Architecture review and design**

| Item | Content |
|------|------|
| **Description** | System design, trade-off analysis, ADR generation |
| **Context** | planning.md |
| **Standalone Use** | ✅ High - Can be used independently without full workflow |

#### Usage

```bash
# Review current architecture
/jikime:architect

# Design new feature architecture
/jikime:architect Design payment system

# Generate ADR
/jikime:architect --adr "Use PostgreSQL over MongoDB"

# Trade-off analysis
/jikime:architect --tradeoff "Monolith vs Microservices"
```

#### Options

| Option | Description |
|--------|------|
| `[description]` | Feature/system to design |
| `--adr` | Generate Architecture Decision Record |
| `--tradeoff` | Trade-off analysis |
| `--review` | Review existing architecture |

#### Architecture Principles

| Principle | Description |
|------|------|
| **Modularity** | High cohesion, low coupling |
| **Scalability** | Horizontally scalable |
| **Maintainability** | Easy to understand and test |
| **Security** | Defense in depth |

---

### /jikime:build-fix

**Progressive build error fixing**

| Item | Content |
|------|------|
| **Description** | Safely fix TypeScript and build errors one by one |
| **Context** | debug.md |
| **Standalone Use** | ✅ High - Use immediately when build fails |

#### Usage

```bash
# Fix all build errors
/jikime:build-fix

# Fix errors in specific file
/jikime:build-fix @src/services/order.ts

# Preview without applying
/jikime:build-fix --dry-run
```

#### Options

| Option | Description |
|--------|------|
| `@path` | Specify specific file |
| `--dry-run` | Preview without applying |
| `--max` | Maximum errors to fix (default: 10) |

#### Safety Rules

- **One error at a time** - Safety first
- **Verify after each fix** - Stop if new errors detected
- **Minimal changes** - Only necessary minimum modifications

---

### /jikime:cleanup

**Dead code detection and safe removal**

| Item | Content |
|------|------|
| **Description** | Comprehensive dead code analysis with knip, depcheck, ts-prune and DELETION_LOG tracking |
| **Context** | dev.md |
| **Agent** | refactorer |
| **Standalone Use** | ✅ High - Can run independently anytime |

#### Usage

```bash
# Scan dead code (analysis only)
/jikime:cleanup scan

# Remove safe items only
/jikime:cleanup remove --safe

# Remove including careful items
/jikime:cleanup remove --careful

# Specific category only
/jikime:cleanup remove --deps
/jikime:cleanup remove --exports
/jikime:cleanup remove --files

# Check deletion log
/jikime:cleanup log

# Comprehensive report
/jikime:cleanup report
```

#### Options

| Option | Description |
|--------|------|
| `scan` | Analyze codebase (no changes) |
| `remove` | Remove dead code |
| `report` | Comprehensive cleanup report |
| `log` | Check DELETION_LOG.md |
| `--safe` | Low-risk items only |
| `--careful` | Include medium-risk |
| `--deps` | Unused dependencies |
| `--exports` | Unused exports |
| `--files` | Unused files |
| `--dry-run` | Preview only |

#### Risk Classification

| Level | Category | Auto Remove |
|-------|----------|----------|
| **SAFE** | npm deps, imports, eslint-disable | ✅ |
| **CAREFUL** | exports, files, types | ⚠️ Confirmation required |
| **RISKY** | Public API, dynamic imports | ❌ Manual review |

---

### /jikime:codemap

**AST-based architecture mapping**

| Item | Content |
|------|------|
| **Description** | Auto-generate architecture documentation from codebase using ts-morph, madge |
| **Context** | sync.md |
| **Skill** | jikime-workflow-codemap |
| **Standalone Use** | ✅ High - Can run independently anytime |

#### Usage

```bash
# Generate full architecture map
/jikime:codemap all

# Specific area only
/jikime:codemap frontend
/jikime:codemap backend
/jikime:codemap database
/jikime:codemap integrations

# Include AST analysis
/jikime:codemap all --ast

# Generate dependency graph
/jikime:codemap all --deps

# JSON output (for automation)
/jikime:codemap all --json
```

#### Options

| Option | Description |
|--------|------|
| `all` | Codemap all areas |
| `frontend` | Frontend architecture |
| `backend` | Backend/API architecture |
| `database` | DB schema/models |
| `integrations` | External services |
| `--ast` | ts-morph AST analysis |
| `--deps` | madge dependency graph |
| `--refresh` | Force regeneration |
| `--json` | JSON output |

#### Output

```
docs/CODEMAPS/
├── INDEX.md          # Architecture overview
├── frontend.md       # Frontend structure
├── backend.md        # Backend structure
├── database.md       # DB schema
├── integrations.md   # External services
└── assets/
    └── dependency-graph.svg
```

---

### /jikime:docs

**Document update**

| Item | Content |
|------|------|
| **Description** | Sync README, API docs, code comments with code |
| **Context** | - |
| **Standalone Use** | ⚠️ Medium - Partial overlap with 3-sync |

#### Usage

```bash
# Update all documents
/jikime:docs

# Specific document type
/jikime:docs --type api
/jikime:docs --type readme
/jikime:docs --type changelog

# Generate missing documents
/jikime:docs --generate

# For specific code changes
/jikime:docs @src/api/
```

#### Options

| Option | Description |
|--------|------|
| `@path` | Specific code target |
| `--type` | Document type: api, readme, changelog, jsdoc |
| `--generate` | Generate missing documents |
| `--dry-run` | Show changes without applying |

#### vs 3-sync

| Feature | docs | 3-sync |
|------|------|--------|
| SPEC-based | ❌ | ✅ |
| CHANGELOG | ✅ | ✅ |
| Project docs | ✅ | ✅ |
| Git PR management | ❌ | ✅ |
| SPEC completion | ❌ | ✅ |

**Recommendation**: Use `3-sync` with SPEC workflow, use `docs` for quick document updates only

---

### /jikime:e2e

**E2E Testing (Playwright)**

| Item | Content |
|------|------|
| **Description** | Generate and run E2E tests with Playwright |
| **Context** | - |
| **Standalone Use** | ✅ High - Separate domain from test |

#### Usage

```bash
# Generate E2E test for flow
/jikime:e2e Test login flow

# Run existing E2E tests
/jikime:e2e --run

# Run specific test
/jikime:e2e --run @tests/e2e/auth.spec.ts

# Debug mode
/jikime:e2e --run --debug
```

#### Options

| Option | Description |
|--------|------|
| `[description]` | User flow to test |
| `--run` | Run existing tests |
| `--debug` | Debug mode (headed browser) |
| `--headed` | Show browser window |

#### vs test

| Feature | e2e | test |
|------|-----|------|
| Scope | Full user flow | Unit/integration |
| Tool | Playwright | Vitest, Jest, Pytest, etc. |
| Speed | Slow | Fast |
| Use Case | Critical user journeys | Business logic, API |

---

### /jikime:learn

**Codebase exploration and learning**

| Item | Content |
|------|------|
| **Description** | Interactive learning of architecture, patterns, implementation details |
| **Context** | research.md |
| **Standalone Use** | ✅ High - Useful for onboarding, code understanding |

#### Usage

```bash
# Full overview
/jikime:learn

# Learn specific topic
/jikime:learn authentication flow

# Learn specific file
/jikime:learn @src/services/order.ts

# Interactive Q&A mode
/jikime:learn --interactive
```

#### Options

| Option | Description |
|--------|------|
| `[topic]` | Specific topic to learn |
| `@path` | Learn specific file/module |
| `--interactive` | Interactive Q&A mode |
| `--depth` | Detail level: overview, detailed, deep |

#### Topics

- **Architecture**: Project structure, patterns
- **Features**: How features work, implementation
- **Conventions**: Coding style, naming, organization
- **Data Flow**: Data flow

---

### /jikime:perspective

**Multi-perspective parallel analysis**

| Item | Content |
|------|------|
| **Description** | Simultaneous analysis from 4 perspectives: Architecture, Security, Performance, Testing |
| **Context** | - |
| **Skill** | jikime-workflow-parallel |
| **Standalone Use** | ✅ High - Can run independently anytime |

#### Usage

```bash
# Analyze entire project
/jikime:perspective

# Analyze specific path
/jikime:perspective @src/api/

# Focus on specific perspective
/jikime:perspective --focus security

# Deep analysis
/jikime:perspective --depth deep

# Quick scan
/jikime:perspective --depth quick

# Combined options
/jikime:perspective @src/auth/ --focus security --depth deep
```

#### Options

| Option | Description |
|--------|------|
| `@path` | Target path for analysis |
| `--focus` | Focus on specific perspective: arch, security, perf, test |
| `--depth` | Analysis depth: quick, standard, deep |

#### Depth Profiles

| Profile | Description | Estimated Time |
|---------|------|----------|
| `quick` | Surface scan, obvious issues | ~1 min |
| `standard` | Balanced analysis (default) | ~3 min |
| `deep` | Comprehensive analysis, edge cases | ~5 min |

#### 4 Perspectives

| Perspective | Analysis Content | Key Metrics |
|------|----------|----------|
| **Architecture** | Structure, coupling, SOLID, DRY | Structure score (0-100) |
| **Security** | OWASP Top 10, input validation, secrets | Risk score (0-100) |
| **Performance** | O(n) complexity, N+1, caching, memory | Efficiency score (0-100) |
| **Testing** | Coverage, edge cases, mocking | Coverage score (0-100) |

#### Synthesis Report

Integrated report generated after 4-perspective analysis:

```markdown
## Cross-Perspective Insights

| Finding | Perspectives | Priority |
|---------|--------------|----------|
| SQL injection + Untested | Security + Testing | CRITICAL |
| N+1 query + High coupling | Performance + Architecture | HIGH |

## Correlation Matrix

              Arch    Sec     Perf    Test
Architecture    -     LOW     HIGH    MED
Security       LOW     -      LOW     HIGH
Performance   HIGH    LOW      -      MED
Testing        MED    HIGH    MED      -
```

#### Parallel Execution

4 subagents execute **in parallel** with a **single message**:

```
Single Message:
  - Task("Architecture analysis", run_in_background: true)
  - Task("Security analysis", run_in_background: true)
  - Task("Performance analysis", run_in_background: true)
  - Task("Testing analysis", run_in_background: true)

→ Collect results with TaskOutput
→ Generate Synthesis integrated report
```

---

### /jikime:refactor

**DDD methodology refactoring**

| Item | Content |
|------|------|
| **Description** | Apply clean code principles with behavior preservation |
| **Context** | dev.md |
| **Standalone Use** | ⚠️ Medium - Partial overlap with 2-run DDD |

#### Usage

```bash
# Refactor specific file
/jikime:refactor @src/services/order.ts

# Refactor with specific pattern
/jikime:refactor @src/utils/ --pattern extract-function

# Safe mode (additional tests)
/jikime:refactor @src/core/ --safe

# Preview
/jikime:refactor @src/auth/ --dry-run
```

#### Options

| Option | Description |
|--------|------|
| `@path` | File to refactor |
| `--pattern` | Pattern: extract-function, remove-duplication |
| `--safe` | Additional characterization tests |
| `--dry-run` | Preview without applying |

#### DDD Approach

```
ANALYZE → PRESERVE → IMPROVE

1. ANALYZE: Understand current behavior
2. PRESERVE: Generate characterization tests
3. IMPROVE: Apply refactoring
4. VERIFY: Confirm tests pass
```

#### vs 2-run

| Feature | refactor | 2-run |
|------|----------|-------|
| SPEC-based | ❌ | ✅ |
| DDD cycle | ✅ | ✅ |
| Quality gates | ✅ | ✅ |
| Git PR management | ❌ | ✅ |
| Use Case | Ad-hoc refactoring | SPEC-based implementation |

**Recommendation**: Use `2-run` with SPEC workflow, use `refactor` for quick refactoring only

---

### /jikime:security

**Security audit**

| Item | Content |
|------|------|
| **Description** | OWASP Top 10, dependency scan, secret detection |
| **Context** | review.md |
| **Standalone Use** | ✅ High - Can run independently anytime |

#### Usage

```bash
# Full security audit
/jikime:security

# Scan specific path
/jikime:security @src/api/

# Dependency audit only
/jikime:security --deps

# Secret scan only
/jikime:security --secrets

# OWASP check only
/jikime:security --owasp
```

#### Options

| Option | Description |
|--------|------|
| `[path]` | Target path |
| `--deps` | Dependency vulnerability scan |
| `--secrets` | Detect hardcoded secrets |
| `--owasp` | OWASP Top 10 check |
| `--fix` | Auto-fix when possible |

#### OWASP Top 10 Checks

| # | Vulnerability | Detection Target |
|---|--------|----------|
| 1 | Injection | SQL, NoSQL, Command |
| 2 | Broken Auth | Password handling |
| 3 | Data Exposure | Hardcoded secrets |
| 4 | XSS | innerHTML, dangerouslySetInnerHTML |
| 5 | SSRF | Unvalidated URLs |
| 6 | Authorization | Missing permission checks |

#### Severity Levels

| Level | Action |
|-------|------|
| CRITICAL | Immediate fix required |
| HIGH | Fix before deployment |
| MEDIUM | Fix as soon as possible |
| LOW | Review and decide |

---

## Generator Commands

Commands for skill and code generation.

### /jikime:skill-create

**General-purpose Claude Code skill generator**

| Item | Content |
|------|------|
| **Description** | Generate various types of specialized skills with Progressive Disclosure pattern |
| **Type** | Generator |
| **Context** | - |
| **MCP** | Context7 (documentation lookup) |

#### Usage

```bash
# Generate language expert skill
/jikime:skill-create --type lang --name rust

# Generate platform integration skill
/jikime:skill-create --type platform --name firebase

# Generate domain expert skill
/jikime:skill-create --type domain --name security

# Generate workflow skill
/jikime:skill-create --type workflow --name ci-cd

# Generate library skill
/jikime:skill-create --type library --name prisma

# Generate framework skill
/jikime:skill-create --type framework --name remix

# Enhance existing skill
/jikime:skill-create --type lang --name python --enhance-only
```

#### Options

| Option | Description |
|--------|------|
| `--type` | Skill type: lang, platform, domain, workflow, library, framework |
| `--name` | Skill name |
| `--enhance-only` | Enhance existing skill only (no new creation) |

#### Generated Structure by Type

| Type | Generated Files | Use Case |
|------|----------|------|
| `lang` | SKILL.md + examples.md + reference.md | Language expert |
| `platform` | SKILL.md + setup.md + reference.md | Platform integration |
| `domain` | SKILL.md + patterns.md + examples.md | Domain expert |
| `workflow` | SKILL.md + steps.md + examples.md | Workflow |
| `library` | SKILL.md + examples.md + reference.md | Library |
| `framework` | SKILL.md + patterns.md + upgrade.md | Framework |

> Detailed documentation: [skill-create.md](./skill-create.md)

---

### /jikime:migration-skill

**Migration-specific skill generator**

| Item | Content |
|------|------|
| **Description** | Generate specialized skills for legacy→modern framework migration |
| **Type** | Generator |
| **Context** | - |
| **MCP** | Context7 (migration guide lookup) |

#### Usage

```bash
# Generate migration skill from CRA to Next.js
/jikime:migration-skill --from cra --to nextjs

# Generate migration skill from Vue to Nuxt
/jikime:migration-skill --from vue --to nuxt

# Enhance existing skill
/jikime:migration-skill --from angular --to react --enhance-only
```

#### Options

| Option | Description |
|--------|------|
| `--from` | Source framework: cra, vue, angular, svelte, jquery, php |
| `--to` | Target framework: nextjs, nuxt, react, vue |
| `--enhance-only` | Enhance existing skill only |

> Detailed documentation: [migration-skill.md](./migration-skill.md)

---

## Migration Workflow

Workflow for migrating legacy projects to Next.js 16.

### /jikime:migrate

**Migration unified command**

| Item | Content |
|------|------|
| **Description** | Migrate legacy frontend to Next.js 16 App Router |
| **Type** | Workflow |
| **Target** | Vue.js, React CRA, Angular, Svelte, etc. |

#### Workflow Overview

```
/jikime:migrate-0-discover   → Step 0: Source discovery
        ↓
/jikime:migrate-1-analyze    → Step 1: Detailed analysis
        ↓
/jikime:migrate-2-plan       → Step 2: Plan establishment
        ↓
/jikime:migrate-3-execute    → Step 3: Execution
        ↓
/jikime:migrate-4-verify     → Step 4: Verification

Or

/jikime:migrate [project]    → Full automation
```

#### Sub-Commands

| Sub-Command | Description | Prerequisites |
|-------------|------|----------|
| `plan` | Generate migration plan | `as_is_spec.md` |
| `skill` | Generate project skill | `migration_plan.md` |
| `run` | Execute migration | `SKILL.md` |

#### Usage

```bash
# Step-by-Step
/jikime:migrate-1-analyze "./my-vue-app"
/jikime:migrate plan my-vue-app
/jikime:migrate skill my-vue-app
/jikime:migrate run my-vue-app --output ./migrated

# Full automation
/jikime:migrate my-vue-app --loop --output ./migrated

# Generate whitepaper
/jikime:migrate run my-vue-app --whitepaper-report --client "ABC Corp"
```

#### Options

| Option | Description |
|--------|------|
| `--artifacts-output` | Migration artifacts directory |
| `--output` | Migrated project output directory |
| `--loop` | Enable autonomous loop |
| `--max N` | Maximum iterations |
| `--strategy` | Strategy: incremental, big-bang |
| `--whitepaper-report` | Generate Post-Migration whitepaper |
| `--client` | Client company name |
| `--lang` | Whitepaper language: ko, en, ja, zh |

#### Target Stack

| Technology | Version |
|------|------|
| Framework | Next.js 16 (App Router) |
| Language | TypeScript 5.x |
| Styling | Tailwind CSS 4.x |
| UI Components | shadcn/ui |
| Icons | lucide-react |
| State | Zustand |

---

## Command Comparison Matrix

### Use Case Guide

| Situation | Recommended Command |
|------|------------|
| Starting new project | `0-project` → `1-plan` → `2-run` → `3-sync` |
| Quick feature implementation | `/jikime:jarvis "description"` |
| Fix build errors | `/jikime:build-fix` |
| Pre-PR comprehensive verification | `/jikime:verify pre-pr` |
| Security check | `/jikime:security` |
| Multi-perspective analysis | `/jikime:perspective @path` |
| Architecture review | `/jikime:architect` |
| Code refactoring (without SPEC) | `/jikime:refactor @path` |
| Code refactoring (SPEC-based) | `/jikime:1-plan` → `/jikime:2-run` |
| Run tests (unit/integration) | `/jikime:test` |
| Run tests (E2E) | `/jikime:e2e` |
| Iterative error fixing | `/jikime:loop "description"` |
| Learn codebase | `/jikime:learn` |
| Architecture documentation | `/jikime:codemap all` |
| Dead code cleanup | `/jikime:cleanup scan` → `/jikime:cleanup remove --safe` |
| Document update (quick) | `/jikime:docs` |
| Document sync (SPEC-based) | `/jikime:3-sync` |
| Legacy migration | `/jikime:migrate` |

### Overlap Analysis

| Command A | Command B | Overlap Area | Recommendation |
|----------|----------|----------|------|
| `docs` | `3-sync` | Document update | SPEC-based: 3-sync, Quick update: docs |
| `refactor` | `2-run` | DDD refactoring | SPEC-based: 2-run, Ad-hoc: refactor |
| `test` | `e2e` | Test execution | Unit/integration: test, E2E: e2e |

---

## Version Information

| Command | Version | Last Updated |
|--------|------|--------------|
| 0-project | 1.0.0 | 2026-01-22 |
| 1-plan | 1.0.0 | 2026-01-22 |
| 2-run | 1.0.0 | 2026-01-22 |
| 3-sync | 2.0.0 | 2026-01-22 |
| jarvis | 1.0.0 | 2026-01-22 |
| test | 1.0.0 | 2026-01-22 |
| loop | 1.0.0 | 2026-01-22 |
| verify | 2.0.0 | 2026-01-26 |
| architect | 1.0.0 | - |
| build-fix | 1.0.0 | - |
| cleanup | 1.0.0 | 2026-01-25 |
| codemap | 1.0.0 | 2026-01-25 |
| docs | 1.0.0 | - |
| e2e | 1.0.0 | - |
| learn | 1.0.0 | - |
| perspective | 1.0.0 | 2026-01-25 |
| refactor | 1.0.0 | - |
| security | 1.0.0 | - |
| skill-create | 1.0.0 | 2026-01-26 |
| migration-skill | 1.0.0 | 2026-01-26 |
| migrate | 1.4.5 | - |

---

Version: 1.3.0
Last Updated: 2026-01-26
Changelog:
- v1.3.0: Added Generator Commands section (skill-create, migration-skill)
- v1.2.0: Added verify (Adversarial Review), perspective (Multi-Perspective Analysis) commands
- v1.1.0: Added cleanup, codemap commands
