# F.R.I.D.A.Y. - Migration Orchestration System

JikiME-ADK's dedicated migration orchestration system. A systematic and precise framework transition automation inspired by Iron Man's second AI assistant.

## Overview

F.R.I.D.A.Y. (Framework Relay & Integration Deployment Assistant Yesterday) is a **dedicated migration orchestrator** that transforms legacy systems to modern frameworks. While J.A.R.V.I.S. handles development, F.R.I.D.A.Y. specializes exclusively in migration, autonomously performing the entire process of analysis, planning, execution, and verification.

### Core Philosophy

```
"Transitioning to the new system, sir. All legacy patterns mapped and ready."
```

### Differences from J.A.R.V.I.S.

| Feature | J.A.R.V.I.S. (Development) | F.R.I.D.A.Y. (Migration) |
|---------|---------------------------|--------------------------|
| Discovery | 5 agents in parallel | 3 agents (source-focused) |
| Planning | Multi-strategy comparison | Dynamic skill discovery + strategy comparison |
| Execution | DDD cycle | DDD + behavior preservation verification |
| Tracking | SPEC-based | `.migrate-config.yaml` + `progress.yaml` |
| Verification | LSP + Tests | Playwright E2E + visual regression + performance comparison |
| Completion Marker | `<jikime>DONE</jikime>` | `<jikime>MIGRATION_COMPLETE</jikime>` |

## Architecture

### System Structure

```
┌─────────────────────────────────────────────────────────────────┐
│                    F.R.I.D.A.Y. System                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Phase 0: Discovery (3-Way Parallel)                            │
│  ┌──────────┐ ┌──────────────┐ ┌──────────────┐                │
│  │ Codebase │ │  Dependency  │ │    Risk      │                │
│  │ Explorer │ │   Analyzer   │ │  Assessor    │                │
│  └────┬─────┘ └──────┬───────┘ └──────┬───────┘                │
│       └───────────────┼────────────────┘                        │
│                       ▼                                         │
│           ┌──────────────────────┐                              │
│           │  .migrate-config.yaml │                              │
│           │  + Complexity Score   │                              │
│           └──────────┬───────────┘                              │
│                      ▼                                          │
│  Phase 1: Detailed Analysis                                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Component/Route/State/DB Mapping → as_is_spec.md        │   │
│  └─────────────────────────────────────┬───────────────────┘   │
│                                        ▼                        │
│  Phase 2: Intelligent Planning                                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ Strategy A   │ │ Strategy B   │ │ Strategy C   │            │
│  │ Incremental  │ │   Phased     │ │  Big-Bang    │            │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘            │
│         └────────────────┼────────────────┘                    │
│                          ▼                                      │
│           ┌────────────────────────┐                            │
│           │ Dynamic Skill Discovery│                            │
│           │ + migration_plan.md    │                            │
│           └────────────┬───────────┘                            │
│                        ▼                                        │
│  Phase 3: DDD Execution (Per Module)                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  FOR EACH module:                                        │   │
│  │    ├── ANALYZE  (understand source behavior + identify DB models) │   │
│  │    ├── PRESERVE (characterization tests + DB layer tests)  │   │
│  │    ├── IMPROVE  (target framework + ORM conversion)       │   │
│  │    ├── LSP Quality Gate (regression check)               │   │
│  │    └── Self-Assessment:                                  │   │
│  │        ├── SUCCESS → Next module                         │   │
│  │        ├── LSP REGRESSION → Pivot approach               │   │
│  │        ├── 3x FAIL → Pivot approach                      │   │
│  │        └── Complexity >90 → User guidance                │   │
│  └─────────────────────────────────────────────────────────┘   │
│                        ▼                                        │
│  Phase 4: Verification (Playwright E2E)                         │
│  ┌──────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────┐  │
│  │  Visual  │ │ Cross-Browser│ │  Core Web    │ │   A11y   │  │
│  │Regression│ │   Testing    │ │   Vitals     │ │  (axe)   │  │
│  └──────────┘ └──────────────┘ └──────────────┘ └──────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Related Files

| File | Description |
|------|-------------|
| `templates/.claude/commands/jikime/friday.md` | F.R.I.D.A.Y. slash command (587 lines) |
| `templates/.claude/commands/jikime/migrate-0-discover.md` | Phase 0 command |
| `templates/.claude/commands/jikime/migrate-1-analyze.md` | Phase 1 command |
| `templates/.claude/commands/jikime/migrate-2-plan.md` | Phase 2 command |
| `templates/.claude/commands/jikime/migrate-3-execute.md` | Phase 3 command |
| `templates/.claude/commands/jikime/migrate-4-verify.md` | Phase 4 command |
| `docs/migration.md` | Migration system documentation |
| `docs/migrate-playwright.md` | Playwright verification system |
| `docs/jarvis.md` | J.A.R.V.I.S. development orchestrator documentation |
| `templates/.jikime/config/quality.yaml` | LSP Quality Gates configuration |

## Config-First Approach

The core design principle of F.R.I.D.A.Y. is **Config-First**. Once the configuration file is generated in Phase 0, it is automatically referenced in all subsequent phases.

### .migrate-config.yaml

```yaml
# Automatically generated in Phase 0
project_name: my-vue-app
source_path: ./legacy-vue-app
source_architecture: monolith           # Detected in Phase 0 (monolith, separated, unknown)
target_framework: nextjs16
artifacts_dir: ./migrations/my-vue-app
output_dir: ./migrations/my-vue-app/out
db_type: postgresql                     # Detected in Phase 0
db_orm: eloquent                        # Detected in Phase 0

# Added in Phase 2
target_architecture: fullstack-monolith  # User selection (fullstack-monolith, frontend-backend, frontend-only)
db_access_from: frontend                 # Automatically derived from target_architecture
# target_framework_backend: fastapi      # frontend-backend architecture only
```

### Artifact Flow

```
@<source-path>/
    │
    ▼ (Phase 0-1: Discover + Analyze)
.migrate-config.yaml                  ← Project settings (including source_architecture, db_type, db_orm)
{artifacts_dir}/as_is_spec.md         ← Detailed analysis (including Database Layer + Architecture Layers)
    │
    ▼ (Phase 2: Plan)
.migrate-config.yaml update           ← Architecture selection (target_architecture, db_access_from)
{artifacts_dir}/migration_plan.md     ← Migration plan (Phase structure by architecture)
    │
    ▼ (Phase 3: Execute)
{output_dir}/                         ← Migrated project (structure varies by architecture)
  ├─ fullstack-monolith: {output_dir}/ (single)
  ├─ frontend-backend: {output_dir}/frontend/ + {output_dir}/backend/
  └─ frontend-only: {output_dir}/ (single, no DB)
{artifacts_dir}/progress.yaml         ← Progress tracking
    │
    ▼ (Phase 4: Verify)
{artifacts_dir}/verification_report.md ← Verification results (architecture-specific verification)
    │
    ▼ (Optional: Whitepaper)
{whitepaper_output}/                  ← Client report
```

## Usage

### Basic Usage

```bash
# Full automatic orchestration (recommended)
/jikime:friday "Migrate Vue app to Next.js" @./legacy-vue-app/

# Specify target framework
/jikime:friday @./my-app/ --target fastapi

# Safe strategy (conservative approach)
/jikime:friday @./legacy/ "Migrate to Go" --strategy safe

# Enable auto loop
/jikime:friday @./src/ --target nextjs --loop --max 30

# Resume interrupted migration
/jikime:friday resume

# Generate whitepaper
/jikime:friday @./app/ --target nextjs --whitepaper --client "ABC Corp" --lang ko
```

### Manual Step-by-Step Execution

```bash
/jikime:migrate-0-discover @./my-vue-app/ --target nextjs
/jikime:migrate-1-analyze
/jikime:migrate-2-plan
/jikime:migrate-3-execute
/jikime:migrate-4-verify --full
```

### Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `@<path>` | Source project path | Current directory |
| `--target` | Target framework (nextjs, fastapi, go, flutter, etc.) | Auto-detect |
| `--strategy` | Execution strategy: auto, safe, fast | auto |
| `--loop` | Enable automatic error correction loop | false |
| `--max N` | Maximum number of iterations | 50 |
| `--whitepaper` | Generate migration whitepaper | false |
| `--client` | Client company name (for whitepaper cover) | - |
| `--lang` | Whitepaper language (ko, en, ja, zh) | conversation_language |
| `resume` | Resume previous migration | - |

## Intelligence Features

### 1. Discovery (Phase 0) - 3-Way Parallel

3 specialized agents analyze the source project **simultaneously**:

| Agent | Role | Output |
|-------|------|--------|
| **Codebase Explorer** | File structure, framework detection, DB/ORM type, source architecture patterns | Tech stack, component list, DB info, source architecture (monolith/separated/unknown), complexity score |
| **Dependency Analyzer** | Package dependencies, version compatibility, breaking changes | Dependency map, upgrade requirements |
| **Risk Assessor** | Migration risks, anti-patterns, legacy locks | Risk score, blocker identification |

### 2. Analysis (Phase 1) - as_is_spec.md Generation

Documents the complete structure of the source project:

- Component/Route/State mapping
- Database layer analysis (models, query patterns, external data services)
- Architecture layer analysis (Frontend / Backend / Data / Shared layer identification + coupling analysis)
- Business logic documentation
- API endpoint mapping
- Dependency and risk assessment

### 3. Planning (Phase 2) - Architecture Selection + Dynamic Skill Discovery

F.R.I.D.A.Y. **dynamically discovers skills** without hardcoded framework patterns:

```bash
# Internal process executed automatically
jikime-adk skill search "{target_framework}"
```

**Architecture Pattern Selection** (core step of Phase 2):

1. Recommendations based on `source_architecture` and `Architecture Layers` analysis results
2. Present 3 options to user: `fullstack-monolith` / `frontend-backend` / `frontend-only`
3. If `frontend-backend` is selected, follow-up question for backend framework (FastAPI/NestJS/Express/Go)
4. Update `.migrate-config.yaml` (`target_architecture`, `target_framework_backend`, `db_access_from`)

Generates and compares 2-3 migration strategies:

| Strategy | Risk | Speed | Suitable When |
|----------|------|-------|---------------|
| **Incremental** | Low | Slow | Complexity > 70 |
| **Phased** | Medium | Medium | Complexity 40-70 |
| **Big-Bang** | High | Fast | Complexity < 40 |

### 4. Execution (Phase 3) - DDD Migration Cycle

Execution strategy varies based on `target_architecture`:

| Architecture | Execution Method |
|--------------|-----------------|
| `fullstack-monolith` | Single project DDD cycle (default) |
| `frontend-backend` | 4-stage separate execution: Shared → Backend → Frontend → Integration |
| `frontend-only` | DDD cycle for frontend modules only (DB steps skipped) |

Performs ANALYZE-PRESERVE-IMPROVE cycle for each module:

```
ANALYZE:     Understand source component behavior
ANALYZE-DB:  Identify data models and query patterns (if DB exists)
PRESERVE:    Write characterization tests (capture current behavior)
PRESERVE-DB: Write data layer tests (if DB exists)
IMPROVE:     Convert to target framework (apply skill conventions)
IMPROVE-DB:  Convert ORM/data access patterns (if DB exists)
```

#### LSP Quality Gates

LSP-based quality gates are automatically applied during Phase 3 execution:

| Phase | Condition | Description |
|-------|-----------|-------------|
| **plan** | `require_baseline: true` | Capture LSP baseline when establishing migration plan |
| **execute** | `max_errors: 0` | Zero type errors/lint errors required |
| **verify** | `require_clean_lsp: true` | LSP clean state required before verification |

Configuration location: `.jikime/config/quality.yaml` → `constitution.lsp_quality_gates`

#### Ralph Loop Integration

F.R.I.D.A.Y.'s DDD Migration Cycle integrates with LSP Quality Gates:

```
Ralph Loop Cycle (Migration):
  1. ANALYZE: Source component analysis + LSP baseline capture
  2. PRESERVE: Characterization test generation
  3. IMPROVE: Convert to target framework
  4. LSP Check: LSP diagnostics after conversion (regression check)
  5. Decision: Continue, Retry, or Pivot
```

When LSP regression is detected, F.R.I.D.A.Y. automatically attempts alternative migration patterns.

#### Self-Assessment Loop

F.R.I.D.A.Y. automatically evaluates during each module conversion:

1. **"Is the current module being migrated successfully?"**
   - TypeScript compilation successful?
   - Characterization tests passing?
   - Build successful?

2. **"Should the approach be changed?"**
   - Trigger: 3 consecutive failures on the same module
   - Action: Attempt alternative migration pattern

3. **"Is the complexity too high for automatic migration?"**
   - Trigger: Single component complexity > 90
   - Action: Split into sub-components or request user guidance

#### Progress Tracking

```yaml
# {artifacts_dir}/progress.yaml
project: my-vue-app
source: vue3
target: nextjs16
target_architecture: fullstack-monolith  # Selected architecture pattern
status: in_progress
strategy: phased

phases:
  discover: completed
  analyze: completed
  plan: completed
  execute: in_progress
  verify: pending

modules:
  total: 15
  completed: 8
  in_progress: 1
  failed: 0
  pending: 6

current:
  module: UserProfile
  iteration: 2
```

### 5. Verification (Phase 4) - Playwright E2E

A 10-step verification system ensures migration quality:

| Step | Verification Item | Tool |
|------|-------------------|------|
| 1 | Infrastructure (Auto-start dev server) | Playwright |
| 2 | Route discovery (Explore testable routes) | Explore |
| 3 | Visual regression (5 viewport comparison) | Playwright Screenshots |
| 4 | Behavior testing (forms, API, JS errors) | Playwright Actions |
| 5 | Cross-browser (Chromium/Firefox/WebKit) | Playwright Multi-Browser |
| 6 | Performance (Core Web Vitals, bundle size) | Playwright Metrics |
| 7 | Accessibility (WCAG compliance) | axe-core |
| 8 | Agent delegation (e2e-tester + skills) | Task Agent |
| 9 | Integration verification (flags, dependencies) | CLI |
| 10 | Report (Markdown + HTML + JSON) | Write |

## Strategy Details

### auto (default)

F.R.I.D.A.Y. analyzes migration complexity and automatically selects the optimal strategy:

| Migration Type | Analysis Result | Selected Strategy |
|----------------|-----------------|-------------------|
| Simple (single framework, <20 components) | Complexity < 40 | Big-Bang (direct sequential) |
| Medium (2-3 concerns, 20-50 components) | Complexity 40-70 | Phased (checkpoints) |
| Complex (multi-domain, >50 components) | Complexity > 70 | Incremental (parallel orchestration) |

### safe (conservative)

Applies maximum verification and safeguards:

- User confirmation between each Phase
- Individual component migration
- Full test suite execution at each step
- Rollback points for all Phases

### fast (aggressive)

Fast execution for small migrations:

- Minimal checkpoints (Phase-level only)
- Batch component migration
- Skip selective verification
- Prioritize quick completion

## Framework Agnosticism

F.R.I.D.A.Y. is designed to be **framework-agnostic**. Without hardcoded target framework patterns, all knowledge is gathered dynamically:

| Source | Knowledge Origin |
|--------|-----------------|
| **Skills** | `jikime-adk skill search "{target_framework}"` |
| **Context7** | Fallback when no skill exists |
| **as_is_spec.md** | Source analysis data |

### Supported Migrations (non-exhaustive)

| Source | Target Options |
|--------|---------------|
| Vue 2/3 | Next.js (App Router) |
| React (CRA) | Next.js (App Router) |
| Angular | Next.js, SvelteKit |
| jQuery | React, Vue, Svelte |
| PHP | Next.js, FastAPI, Go |
| Monolith | Microservices |
| Any source | Any target |

## Whitepaper Generation

Using the `--whitepaper` flag generates a client report after migration completion:

### Artifact Structure

```
{whitepaper_output}/
├── 00_cover.md                    # Cover + Table of Contents
├── 01_executive_summary.md        # Non-technical summary
├── 02_migration_summary.md        # Execution timeline
├── 03_architecture_comparison.md  # Before/After diagrams
├── 04_component_inventory.md      # Migrated component list
├── 05_performance_report.md       # Performance metrics
├── 06_quality_report.md           # Quality metrics
└── 07_lessons_learned.md          # Recommendations
```

### Supported Languages

| Code | Language |
|------|----------|
| ko | Korean |
| en | English |
| ja | Japanese |
| zh | Chinese |

## Resume Capability

You can continue an interrupted migration:

```bash
/jikime:friday resume
```

Internal operation:
1. Read project settings from `.migrate-config.yaml`
2. Check current status from `{artifacts_dir}/progress.yaml`
3. Identify last completed Phase
4. Continue from next pending Phase/module
5. Restore strategy and context

## Agent Delegation

### Agent Delegation by Phase

| Phase | Agent | Role |
|-------|-------|------|
| Phase 0 | Explore (x3) | Code analysis, dependency analysis, risk assessment |
| Phase 1 | manager-spec | as_is_spec.md generation |
| Phase 2 | manager-strategy | Migration strategy development |
| Phase 3 | backend, frontend | DDD-based code migration |
| Phase 4 | e2e-tester | Playwright verification |

## Output Format

### During Execution

```markdown
## F.R.I.D.A.Y.: Phase 3 - Execution (Module 8/15)

### Migration: Vue 3 → Next.js 16
### Complexity Score: 55/100

### Module Status
- [x] Auth module (5 components)
- [x] Users module (3 components)
- [ ] Products module <- in progress
- [ ] Orders module
- [ ] Dashboard module

### Self-Assessment
- Progress: YES (build errors: 3 -> 1)
- Pivot needed: NO
- Current module confidence: 80%

### Active Issues
- WARNING: ProductCard.tsx - dynamic import pattern needs manual review

Continuing...
```

### Completion

```markdown
## F.R.I.D.A.Y.: MIGRATION COMPLETE

### Summary
- Source: Vue 3 (Vuetify)
- Target: Next.js 16 (App Router)
- Strategy Used: Phased
- Modules Migrated: 15/15
- Tests: 89/89 passing
- Build: SUCCESS
- Iterations: 12
- Self-Corrections: 2

### Predictive Suggestions
1. Set up CI/CD pipeline for the new project
2. Configure production environment variables
3. Set up monitoring and error tracking
4. Plan user acceptance testing

<jikime>MIGRATION_COMPLETE</jikime>
```

## Limitations & Safety

### Limitations

- Maximum 3 strategy pivots (user intervention requested afterward)
- Maximum 5 retries per module
- Session-only learning (no cross-session learning)

### Safety Measures

- [HARD] All implementations delegated to specialist agents
- [HARD] User confirmation required before Phase 3 execution (except --strategy fast)
- [HARD] Completion marker required: `<jikime>MIGRATION_COMPLETE</jikime>`
- [HARD] Dynamic skill discovery - no hardcoded framework patterns
- [HARD] Read from `.migrate-config.yaml` and `as_is_spec.md` - no source re-analysis
- [HARD] LSP Quality Gate: Zero errors required in execute phase
- Rollback points created for each Phase
- LSP Quality Gates automatically alert when regression is detected

## Best Practices

### When to Use F.R.I.D.A.Y.?

**Suitable cases:**
- Transitioning legacy projects to modern frameworks
- Framework upgrades (Vue 2 → Vue 3, React CRA → Next.js)
- Monolithic → Microservices conversion
- Frontend framework replacement

**When J.A.R.V.I.S. is better:**
- Implementing new features
- Refactoring existing code (without framework changes)
- Bug fixes
- Performance optimization

### Dual Orchestrator Switching

```bash
# Development work → J.A.R.V.I.S.
/jikime:jarvis "Add user authentication"

# Migration work → F.R.I.D.A.Y.
/jikime:friday "Migrate Vue app to Next.js 16"

# Automatic routing (keyword-based)
"migrate this app" → F.R.I.D.A.Y.
"implement login" → J.A.R.V.I.S.
```

---

Version: 1.3.0
Last Updated: 2026-02-03
Codename: F.R.I.D.A.Y. (Framework Relay & Integration Deployment Assistant Yesterday)
Inspiration: Iron Man's second AI Assistant (successor to J.A.R.V.I.S.)
Changelog:
- v1.3.0: Added Architecture Pattern support (fullstack-monolith, frontend-backend, frontend-only); Architecture selection step (Phase 2); Architecture-specific execution/verification strategies; source_architecture detection
- v1.2.0: Added Database Layer support (DB/ORM detection, DB-aware DDD cycle, DB schema verification)
- v1.1.0: LSP Quality Gates integration, Ralph Loop Integration added
- v1.0.0: Initial release - Migration-focused orchestrator extracted from J.A.R.V.I.S.
