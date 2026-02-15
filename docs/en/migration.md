# JikiME-ADK Migration System

## Overview

JikiME-ADK Migration System is an intelligent system that migrates legacy projects to modern frameworks through the F.R.I.D.A.Y. orchestrator. With a **Config-First approach**, you only need to input the source path and target framework once, and all subsequent steps will automatically reference them.

## Core Design Principles

| Principle | Description |
|-----------|-------------|
| **Input Once, Use Everywhere** | Source path/target is entered only once in Step 0 |
| **Config-First** | `.migrate-config.yaml` is the single source of truth for all settings |
| **Framework Agnostic** | Rules are applied through dynamic skill discovery without hardcoding |
| **DDD Methodology** | Operates on ANALYZE-PRESERVE-IMPROVE cycle for behavior preservation |

---

## Step-by-Step Workflow

```
/jikime:migrate-0-discover   → Step 0: Source discovery + config generation
        ↓
/jikime:migrate-1-analyze    → Step 1: Detailed analysis + config update
        ↓
/jikime:migrate-2-plan       → Step 2: Plan creation (awaiting approval)
        ↓
/jikime:migrate-3-execute    → Step 3: DDD execution
        ↓
/jikime:migrate-4-verify     → Step 4: Verification + report
        ↓
/jikime:verify --browser-only --fix-loop  → (Optional) Runtime error detection + auto-fix

or

/jikime:friday "description" @<path>  → Full automatic orchestration
```

### Quick Start

```bash
# Method 1: FRIDAY automatic orchestration (recommended)
/jikime:friday "Migrate Vue app to Next.js" @./my-vue-app/ --target nextjs

# Method 2: Manual step-by-step execution
/jikime:migrate-0-discover @./my-vue-app/ --target nextjs
/jikime:migrate-1-analyze
/jikime:migrate-2-plan
/jikime:migrate-3-execute
/jikime:migrate-4-verify --full
```

> **Note**: Once you specify the path and target in Step 0, subsequent steps can be executed without arguments.

---

## Config-First Approach

### `.migrate-config.yaml` (Single Source of Truth)

Automatically generated in Step 0 and referenced in all subsequent steps:

```yaml
version: "1.0"
project_name: my-vue-app
source_path: ./my-vue-app
source_framework: vue3              # Detected in Step 0
source_architecture: monolith       # Detected in Step 0 (monolith, separated, unknown)
target_framework: nextjs16          # Specified with --target in Step 0
artifacts_dir: ./migrations/my-vue-app
output_dir: ./migrations/my-vue-app/out
db_type: postgresql                 # Detected in Step 0 (postgresql, mysql, sqlite, mongodb, none)
db_orm: eloquent                    # Detected in Step 0 (prisma, drizzle, typeorm, sequelize, mongoose, eloquent, none)
created_at: "2026-01-23T10:00:00Z"
# Fields added in Step 1
analyzed_at: "2026-01-23T11:00:00Z"
component_count: 45
complexity_score: 7
db_model_count: 15                  # Added in Step 1 (0 if no database)
# Fields added in Step 2
target_architecture: fullstack-monolith  # User selection in Step 2 (fullstack-monolith, frontend-backend, frontend-only)
target_framework_backend: fastapi        # Selection in Step 2 (frontend-backend architecture only, fastapi, nestjs, express, go-fiber)
db_access_from: backend                  # Auto-derived in Step 2 (frontend, backend, both, none)
```

### Config Field Lifecycle

| Field | Creation Step | Usage Steps |
|-------|---------------|-------------|
| `source_path` | Step 0 | Step 1, 3 |
| `source_framework` | Step 0 | Step 1, 2, 3 |
| `source_architecture` | Step 0 | Step 1, 2 |
| `target_framework` | Step 0 or 1 | Step 2, 3 |
| `artifacts_dir` | Step 0 | Step 2, 3, 4 |
| `output_dir` | Step 0 | Step 3, 4 |
| `db_type` | Step 0 | Step 1, 2, 3, 4 |
| `db_orm` | Step 0 | Step 1, 2, 3, 4 |
| `component_count` | Step 1 | Step 2 |
| `complexity_score` | Step 1 | Step 2 |
| `db_model_count` | Step 1 | Step 2 |
| `target_architecture` | Step 2 | Step 3, 4 |
| `target_framework_backend` | Step 2 | Step 3, 4 |
| `db_access_from` | Step 2 | Step 3, 4 |

---

## Command Reference

### Step 0: Discover (Source Discovery)

```bash
/jikime:migrate-0-discover @<path> [--target <framework>] [--quick]
```

| Option | Required | Description |
|--------|----------|-------------|
| `@path` | Yes | Source code path to analyze |
| `--target` | No | Target framework (`nextjs`\|`fastapi`\|`go`\|`flutter`) |
| `--quick` | No | Quick overview only (skip detailed analysis) |

**What it does**:
- Detect tech stack (language, framework, version)
- Identify architecture patterns
- Detect source architecture pattern (monolith / separated / unknown)
- Detect database type and ORM
- Evaluate migration complexity
- Generate `.migrate-config.yaml`
- Suggest recommended frameworks if target not specified

**Outputs**: `.migrate-config.yaml`, Discovery Report

---

### Step 1: Analyze (Detailed Analysis)

```bash
/jikime:migrate-1-analyze [project-path] [options]
```

| Option | Required | Description |
|--------|----------|-------------|
| `project-path` | No* | Legacy project path (automatically read from config) |
| `--framework` | No | Force source framework (`vue`\|`react`\|`angular`\|`svelte`\|`auto`) |
| `--target` | No | Target framework (overrides config value) |
| `--artifacts-output` | No | Artifacts path (default: `./migrations/{project}/`) |
| `--whitepaper` | No | Generate whitepaper package for client proposals |
| `--whitepaper-output` | No | Whitepaper output path (default: `./whitepaper/`) |
| `--client` | No | Client company name (for whitepaper cover) |
| `--lang` | No | Whitepaper language (`ko`\|`en`\|`ja`\|`zh`) |

*\* Automatically read if `.migrate-config.yaml` exists. Required if it doesn't.*

**Path Priority**:
1. Explicit argument: `/jikime:migrate-1-analyze "./my-app" --target nextjs`
2. Config file: `.migrate-config.yaml` → `source_path`, `target_framework`
3. Error: If neither exists, guidance to run Step 0 first

**What it does**:
- Analyze component structure and hierarchy
- Map routing structure
- Identify state management patterns
- Analyze database layer (models, query patterns, external data services)
- Analyze architecture layers (identify Frontend / Backend / Data / Shared layers)
- Analyze dependency compatibility
- Identify risk factors

**Outputs**:
- `{artifacts_dir}/as_is_spec.md`
- `.migrate-config.yaml` update (adds component_count, complexity_score)

---

### Step 2: Plan (Plan Creation)

```bash
/jikime:migrate-2-plan [--modules <list>] [--incremental]
```

| Option | Required | Description |
|--------|----------|-------------|
| `--modules` | No | Plan only specific modules (e.g., `auth,users,orders`) |
| `--incremental` | No | Incremental migration plan |

**What it does**:
1. Read `target_framework` from `.migrate-config.yaml`
2. Dynamic skill discovery (`jikime-adk skill search "{target_framework}"`, `"{db_orm}"`, `"{db_type}"`)
3. Select target architecture pattern (fullstack-monolith / frontend-backend / frontend-only)
4. Create plan based on `{artifacts_dir}/as_is_spec.md`
5. Establish database migration strategy
6. Apply skill rules (structure, naming, routing)
7. **Await user approval**

**Outputs**: `{artifacts_dir}/migration_plan.md`

**Approval Methods**:
- `yes` - Proceed as planned
- `modify: [changes]` - Modify plan
- `no` - Cancel

**Constraints**: Does not directly analyze source code. Must only reference `as_is_spec.md`.

---

### Step 3: Execute (Execution)

```bash
/jikime:migrate-3-execute [--module <name>] [--resume] [--dry-run]
```

| Option | Required | Description |
|--------|----------|-------------|
| `--module` | No | Migrate only specific module |
| `--resume` | No | Resume interrupted migration (based on `progress.yaml`) |
| `--dry-run` | No | Preview without actual execution |

**Methodology**: DDD (ANALYZE → PRESERVE → IMPROVE)

**Architecture-specific Execution Strategies**:

| Architecture | Execution Method | Output Structure |
|--------------|------------------|------------------|
| `fullstack-monolith` | Single project DDD cycle | `{output_dir}/` |
| `frontend-backend` | 4 phases: Shared → Backend → Frontend → Integration | `{output_dir}/frontend/` + `{output_dir}/backend/` |
| `frontend-only` | Frontend modules only DDD cycle (skip DB) | `{output_dir}/` |

```
Iteration per module (fullstack-monolith basis):
  1. ANALYZE     - Understand source module behavior
  1.5 ANALYZE-DB - Identify data models and queries (if DB exists)
  2. PRESERVE    - Write characterization tests (preserve behavior)
  2.5 PRESERVE-DB - Write data layer tests (if DB exists)
  3. IMPROVE     - Transform to target framework
  3.5 IMPROVE-DB - Transform ORM/data access patterns (if DB exists)
  4. Validate    - Build + test + DB schema verification

For frontend-backend:
  Sub-Phase 1: Shared Layer (shared types, API contract definition)
  Sub-Phase 2: Backend (API + business logic + data access)
  Sub-Phase 3: Frontend (components + routing + state + API Client)
  Sub-Phase 4: Integration (API contract consistency verification)
```

**Outputs**:
- `{output_dir}/` - Migrated project (structure varies by architecture)
- `{artifacts_dir}/progress.yaml` - Progress tracking

**progress.yaml Structure**:
```yaml
project: my-vue-app
source_framework: vue3
target_framework: nextjs16
status: in_progress
modules:
  total: 15
  completed: 8
  in_progress: 1
  failed: 0
  pending: 6
```

---

### Step 4: Verify (Verification)

```bash
/jikime:migrate-4-verify [options]
```

| Option | Required | Description |
|--------|----------|-------------|
| `--full` | No | Run all verification types (visual + cross-browser + a11y + performance) |
| `--behavior` | No | Behavior preservation comparison only |
| `--e2e` | No | E2E tests only |
| `--visual` | No | Screenshot-based visual regression verification |
| `--performance` | No | Core Web Vitals and load time comparison |
| `--cross-browser` | No | Cross-browser verification (Chromium, Firefox, WebKit) |
| `--a11y` | No | WCAG accessibility verification (axe-core) |
| `--source-url` | No | Source system URL (for live comparison) |
| `--target-url` | No | Target system URL (for live comparison) |
| `--headed` | No | Show browser window (for debugging) |
| `--capture-skill` | No | Save verified migration patterns as reusable skills |

> **Note**: `--source-url`/`--target-url` are for comparing running instances. Source/target framework information is automatically read from `.migrate-config.yaml`.

**Architecture-specific Verification**:

| Architecture | Verification Target | DB Verification |
|--------------|---------------------|-----------------|
| `fullstack-monolith` | Single project build/type-check/lint | Included |
| `frontend-backend` | Frontend + Backend individual verification + integration verification | Runs on Backend |
| `frontend-only` | Single project build/type-check/lint | Skipped |

**Verification Items**:
1. Dev Server Setup - Auto-start source/target development servers
2. Route Discovery - Discover testable routes from migration outputs
3. Characterization Tests - Behavior preservation tests
4. Behavior Comparison - Source/target output comparison
5. E2E Tests - Playwright-based user flow verification
6. Visual Regression - Screenshot comparison (source vs target)
7. Performance Check - Core Web Vitals, load time comparison
8. Cross-Browser - Chromium, Firefox, WebKit verification
9. Accessibility - axe-core based WCAG compliance check
10. Skill Capture - Save verified patterns as reusable skills (`--capture-skill` option)

**Outputs**:
- `{artifacts_dir}/verification_report.md`
- `skills/jikime-migration-{source}-to-{target}/` (when using `--capture-skill`)

### Runtime Error Detection: verify --browser-only

Even after migration verification, **runtime errors that only occur in the browser** may remain. To catch errors that static analysis or build tools cannot detect (undefined references, incorrect library imports, etc.), use `/jikime:verify --browser-only`.

```bash
# After migration verification, additional runtime error check
cd {output_dir}
/jikime:verify --browser-only

# Check specific routes only
/jikime:verify --browser-only --routes /,/dashboard,/settings

# Error report only (no fix, without fix-loop)
/jikime:verify --browser-only

# Enable auto-fix loop
/jikime:verify --browser-only --fix-loop

# Show browser window (headed mode)
/jikime:verify --browser-only --headed
```

**verify --browser-only Behavior**:
1. Detect package manager from package.json (pnpm/yarn/npm/bun)
2. Start development server with `dev` script (background)
3. Navigate to each route with Playwright and capture errors
4. Extract source file:line from stack trace
5. When using `--fix-loop`: Delegate fix to expert agent (automatic)
6. Re-verification loop (repeat until 0 errors)

> **Tip**: `migrate-4-verify` focuses on **static analysis verification** after migration, while `verify --browser-only --fix-loop` focuses on **runtime error detection and auto-fix**. Using them sequentially after migration is most effective.

---

## F.R.I.D.A.Y. Orchestrator

Automatically orchestrates the entire migration process.

```bash
/jikime:friday "task description" @<source-path> [options]
```

| Option | Description |
|--------|-------------|
| `@<source-path>` | Source project path |
| `--target` | Target framework (`nextjs`\|`fastapi`\|`go`\|`flutter`) |
| `--strategy` | Migration strategy (`auto`\|`safe`\|`fast`) |
| `--loop` | Automatic iteration mode |
| `--max N` | Maximum iterations (default: 100) |
| `--whitepaper` | Generate whitepaper for client delivery |
| `--client` | Client name |
| `--lang` | Whitepaper language (`ko`\|`en`\|`ja`\|`zh`) |
| `resume` | Resume interrupted task |

### FRIDAY Execution Flow

```
/jikime:friday "Vue→Next.js migration" @./my-vue-app/ --target nextjs
    │
    ├─ Phase 1: Discovery
    │   └─ /jikime:migrate-0-discover @./my-vue-app/ --target nextjs
    │       → .migrate-config.yaml generation
    │
    ├─ Phase 2: Analysis
    │   └─ /jikime:migrate-1-analyze
    │       → as_is_spec.md + config update
    │
    ├─ Phase 3: Planning
    │   └─ /jikime:migrate-2-plan
    │       → migration_plan.md (user approval)
    │
    ├─ Phase 4: Execution
    │   └─ /jikime:migrate-3-execute
    │       → output_dir/ + progress.yaml
    │
    └─ Phase 5: Verification
        └─ /jikime:migrate-4-verify --full
            → verification_report.md
```

---

## Data Flow Diagram

```
User Input: Source path + Target (only once initially)
     │
     ▼
Step 0: .migrate-config.yaml generation
     │  (source_path, source_framework, source_architecture, target_framework,
     │   db_type, db_orm, artifacts_dir, output_dir)
     │
     ▼
Step 1: config update + as_is_spec.md generation
     │  (add component_count, complexity_score, db_model_count, analyzed_at)
     │  (Architecture Layers analysis: Frontend/Backend/Data/Shared)
     │
     ▼
Step 2: Architecture pattern selection + migration_plan.md generation (awaiting approval)
     │  (add target_architecture, target_framework_backend, db_access_from)
     │  (Dynamic skill discovery → Apply target rules + DB migration strategy)
     │
     ▼
Step 3: output_dir/ generation + progress.yaml update
     │  (Architecture-specific execution strategy: monolith / frontend-backend / frontend-only)
     │  (Module-wise DDD cycle iteration + DB layer transformation)
     │
     ▼
Step 4: verification_report.md generation
     │  (Architecture-specific verification + behavior preservation + performance verification + DB schema/connection verification)
     │
     ▼
(Optional) --capture-skill
     │  (Save verified patterns as skills → Reuse in next migration)
     │
     ▼
(Optional) /jikime:verify --browser-only --fix-loop
     │  (Runtime browser error detection + auto-fix)
     │
     ▼
Complete → Staging deployment → UAT → Production
```

---

## Dynamic Skill Discovery

Step 2 dynamically discovers skills matching the target framework:

```bash
# Auto-discovery based on target_framework
jikime-adk skill search "{target_framework}"
jikime-adk skill search "migrate {target_framework}"
jikime-adk skill search "{target_language}"
jikime-adk skill search "{db_orm}"
jikime-adk skill search "{db_type}"
```

| target_framework | Discovered Skills |
|------------------|-------------------|
| `nextjs16` | `jikime-migrate-to-nextjs`, `jikime-nextjs@16`, `jikime-library-shadcn` |
| `fastapi` | `jikime-lang-python` (+ related skills) |
| `go-fiber` | `jikime-lang-go` (+ related skills) |
| `flutter` | `jikime-lang-flutter` (+ related skills) |

If no skills are found, official documentation is queried through Context7 MCP.

---

## Architecture Patterns

In Step 2 (Plan), users select the target architecture pattern. It is automatically recommended based on source analysis results, and the user makes the final decision.

### 3 Patterns

| Pattern | Description | Suitable For |
|---------|-------------|--------------|
| **fullstack-monolith** | Single Next.js project (API Routes + Server Components → DB) | Small to medium scale, monolithic source |
| **frontend-backend** | Separate Frontend (Next.js) + Backend (FastAPI/NestJS/Express/Go) | Large scale, already separated source |
| **frontend-only** | Migrate frontend only (keep existing backend, API calls) | When backend needs to be maintained |

### Selection Criteria

```
source_architecture?
├─ monolith + component_count < 50 → Recommend: fullstack-monolith
├─ monolith + component_count >= 50 → Recommend: frontend-backend
├─ separated → Recommend: frontend-backend
└─ unknown → Present 3 options to user
```

### Directory Structure

**fullstack-monolith** (default):
```
{output_dir}/
├── src/
│   ├── app/          # Next.js App Router
│   ├── components/   # React components
│   ├── lib/          # Utilities, DB client
│   └── stores/       # State management
├── prisma/           # DB schema
└── package.json
```

**frontend-backend**:
```
{output_dir}/
├── shared/           # Shared types, API contracts
│   └── types/
├── frontend/         # Next.js project
│   ├── src/
│   └── package.json
└── backend/          # Backend project (FastAPI/NestJS/Express/Go)
    ├── src/
    ├── prisma/       # DB schema
    └── package.json
```

**frontend-only**:
```
{output_dir}/
├── src/
│   ├── app/          # Next.js App Router
│   ├── components/   # React components
│   ├── lib/          # API client, utilities
│   └── stores/       # State management
└── package.json      # No DB related
```

### Config Fields

```yaml
# Auto-detected in Step 0
source_architecture: monolith    # monolith | separated | unknown

# User selection in Step 2
target_architecture: fullstack-monolith  # fullstack-monolith | frontend-backend | frontend-only

# Added when frontend-backend is selected
target_framework_backend: fastapi  # fastapi | nestjs | express | go-fiber

# Auto-derived from target_architecture
db_access_from: frontend          # frontend | backend | both | none
```

**Default (backward compatible)**: If `target_architecture` is not set, operates as `fullstack-monolith` (same as existing behavior)

---

## Architecture

### Skills Structure

| Layer | Role | Example |
|-------|------|---------|
| **Migration Skills** | Framework transition strategies | `jikime-migrate-to-nextjs` |
| **Version Skills** | Version-specific guides | `jikime-nextjs@16` |
| **Language Skills** | Language-specific patterns | `jikime-lang-typescript` |
| **Domain Skills** | Domain-specific patterns | `jikime-migration-patterns-auth` |

### Skills Naming Convention

```
Migration:      jikime-migrate-{source}-to-{target} or jikime-migrate-to-{target}
Version Guide:  jikime-{framework}@{version}
Language:       jikime-lang-{language}
Domain Pattern: jikime-migration-patterns-{domain}
```

### MCP Integration

| MCP Server | Purpose |
|------------|---------|
| **Context7** | Official documentation, migration guides, API changes |
| **Playwright** | Step 4 verification (E2E, visual regression, cross-browser), runtime error detection |
| **WebFetch** | Latest release notes, breaking changes |
| **Sequential** | Complex migration analysis |

---

## Supported Migrations

| Source | Target Options |
|--------|----------------|
| PHP (Laravel) | Next.js, FastAPI, Go, Spring Boot |
| jQuery | React, Vue, Svelte |
| Vue 2/3 | Next.js (App Router), Nuxt |
| React (CRA) | Next.js (App Router) |
| Angular | Next.js, SvelteKit |
| Java Servlet | Spring Boot, Go, FastAPI |
| Python 2 | Python 3, FastAPI |
| Svelte | SvelteKit |

---

## Best Practices

### User Guide

1. **Start from Step 0** - Always start with Discover to generate config
2. **Target only once** - No need to re-enter after specifying `--target` in Step 0
3. **Git commit first** - Always commit current state before migration
4. **Review the plan** - Carefully review the plan in Step 2 before approval
5. **Progress by module** - For large projects, use `--module` option for incremental execution
6. **Resume after interruption** - Use `--resume` in Step 3 to continue
7. **Check runtime errors** - After Step 4, use `/jikime:verify --browser-only --fix-loop` to catch browser runtime errors
8. **Save experience as skills** - After successful migration, use `--capture-skill` to convert patterns into skills for reuse in next migration

### Skill Author Guide

1. **Include metadata** - Specify triggers, version, etc. in frontmatter
2. **Reference latest docs** - Include Context7/WebFetch usage instructions
3. **Breaking changes** - Document known issues and solutions
4. **Provide examples** - Include Before/After code examples
5. **Specify versions** - Clearly state compatible version ranges

---

Version: 3.4.0
Last Updated: 2026-02-03
Changelog:
- v3.4.0: Added Architecture Patterns section (fullstack-monolith, frontend-backend, frontend-only); Architecture-specific execution and verification; New config fields (source_architecture, target_architecture, target_framework_backend, db_access_from)
- v3.3.0: Added database layer support across all phases (db_type, db_orm, db_model_count); DB-aware DDD cycle; DB skill discovery
- v3.2.0: Added --capture-skill option to Step 4 for generating reusable migration skills from verified patterns
- v3.1.0: Step 4 Playwright-based verification details; Added verify --browser-only integration for runtime error detection
- v3.0.0: Config-First approach; FRIDAY orchestrator; Removed /jikime:migrate; Removed redundant source/target options from Steps 2-4; Renamed --source/--target to --source-url/--target-url in Step 4
- v2.0.0: Added Step-by-Step Workflow, Command Reference with full options
- v1.0.0: Initial migration system documentation
