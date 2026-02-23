---
description: "[Step 0/4] Source project discovery. Identify tech stack, architecture, and migration complexity."
argument-hint: '@<source-path> [--quick]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Glob, Grep
model: inherit
---

# Migration Step 0: Discover

**Starting Phase**: Explore source code and perform initial analysis.

## What This Command Does

1. **Technology Detection** - Identify languages, frameworks, and libraries
2. **Architecture Analysis** - Understand structure, patterns, and dependencies
3. **Architecture Pattern Detection** - Classify source as monolith, separated, or unknown
4. **Complexity Assessment** - Evaluate migration difficulty
5. **Interactive Stack Selection** - Guide user through target stack choices
6. **Config Initialization** - Create `.migrate-config.yaml` with full target stack
7. **Discovery Report** - Generate comprehensive discovery report

## Usage

```bash
# Discover source codebase (interactive stack selection)
/jikime:migrate-0-discover @./legacy-app/

# Quick discovery (overview only, skips interactive selection)
/jikime:migrate-0-discover @./legacy-app/ --quick
```

## Options

| Option | Description |
|--------|-------------|
| `@path` | Source code path to analyze (required) |
| `--quick` | Quick overview without deep analysis or interactive selection |
| `--skip-site-flow` | Skip site-flow API integration (offline mode) |

## Execution Flow

### Step 1: Analyze Source

Explore the source project to detect:
- Primary language and framework
- Framework version
- Build tools and package manager
- Database type and ORM/schema tool
- Source architecture pattern (monolith / separated / unknown)
- File count and complexity

**Architecture Detection** (uses architecture-detection module logic):
- Check for frontend/backend directory separation
- Analyze monorepo configuration (turbo.json, pnpm-workspace.yaml, etc.)
- Detect fullstack framework indicators (Next.js, Laravel, Django, Rails)
- Classify as `monolith`, `separated`, or `unknown`

### Step 1.5: Interactive Stack Selection

After source analysis, guide the user through target stack selection using **sequential AskUserQuestion** calls.

**Skip this step if `--quick` is specified** (set all target fields to "pending").

#### Dynamic Option Detection

Before asking questions, scan installed skills to populate framework options dynamically:

```
jikime-adk skill list --tag framework  → frontend framework options
jikime-adk skill list --agent backend  → backend framework options
jikime-adk skill list --tag database   → ORM/DB options
```

Use detected skills as option sources. Fall back to common defaults if no skills found.

#### Question Flow

```
Question 1: Architecture Strategy
  "What architecture strategy for the target project?"
  Options (max 4):
  - Fullstack: Single framework handles frontend + backend + DB (e.g., Next.js)
  - Separated: Independent frontend and backend projects

  → Sets: target_architecture

Question 2: Frontend Framework
  "Which frontend framework for the target?"
  Options (dynamic from installed skills, top 3 + Other):
  - Next.js (Recommended)
  - Nuxt
  - Angular
  - Other (manual input)

  → Sets: target_framework

Question 3: (Separated only) Backend Language/Framework
  "Which backend framework?"
  Options (dynamic from installed skills, top 3 + Other):
  - Java (Spring Boot)
  - Go (Fiber/Echo)
  - Python (FastAPI)
  - Other (manual input)

  → Sets: target_framework_backend, target_backend_language

Question 4: (Separated only) DB Access Layer
  "Which layer handles database access?"
  Options:
  - Backend only (Recommended)
  - Both frontend and backend

  → Sets: db_access_from

Question 5: DB Schema Extraction (skip if db_type is "none")
  "Extract existing DB schema for migration?"
  Options:
  - Yes, from environment variable (DATABASE_URL in .env)
  - Yes, from schema file (schema.prisma / SQL dump)
  - No, analyze from source code only

  → Sets: db_schema_source

Question 6: (Fullstack only) Target DB ORM
  "Which ORM for the target project?"
  Options (context-dependent):
  - Prisma (Recommended)
  - Drizzle
  - Supabase
  - Other (manual input)

  → Sets: target_db_orm

Question 7: UI Component Library
  "Which UI component library for the target?"
  Options (dynamic from installed skills, top 3 + Other):
  - shadcn/ui (Recommended) - Modern, accessible, customizable
  - Material UI (MUI) - Comprehensive, enterprise-ready
  - Chakra UI - Simple, modular, accessible
  - Keep legacy CSS (copy existing styles)
  - Other (manual input)

  → Sets: target_ui_library
```

**IMPORTANT**: This question is critical for modernizing the frontend.
- If user selects a modern UI library, migration will convert legacy components to modern equivalents
- If "Keep legacy CSS" is selected, existing styles will be preserved (not recommended for modernization)

#### Conditional Flow Summary

| Architecture | Questions Asked |
|-------------|----------------|
| **Fullstack** | Q1 → Q2 → Q5 → Q6 → Q7 (5 questions) |
| **Separated** | Q1 → Q2 → Q3 → Q4 → Q5 → Q7 (6 questions) |

#### Derived Values

Values automatically derived from user selections:

| Field | Fullstack | Separated |
|-------|-----------|-----------|
| `target_architecture` | `fullstack-monolith` | `frontend-backend` |
| `db_access_from` | `frontend` (via API Routes) | From Q4 (`backend` or `both`) |
| `target_framework_backend` | _(not set)_ | From Q3 |
| `target_backend_language` | _(not set)_ | From Q3 |
| `target_db_orm` | From Q6 | _(set in Plan phase based on backend)_ |
| `target_ui_library` | From Q7 | From Q7 |

### Step 2: Create `.migrate-config.yaml`

After discovery and interactive selection, **automatically create** the config file:

```yaml
# .migrate-config.yaml (created by Step 0)
version: "1.0"
project_name: legacy-app          # Derived from @path
source_path: ./legacy-app         # From @path argument
source_framework: laravel8        # Detected framework
db_type: mysql                    # Detected database type (postgresql, mysql, sqlite, mongodb, none)
db_orm: eloquent                  # Detected ORM/schema tool (prisma, drizzle, typeorm, sequelize, mongoose, eloquent, none)
source_architecture: monolith     # Detected architecture pattern (monolith, separated, unknown)
artifacts_dir: ./migrations/legacy-app  # Default artifacts location
output_dir: ./migrations/legacy-app/out # Default output location
created_at: "2026-01-23T10:00:00Z"

# Target stack (from interactive selection)
target_architecture: fullstack-monolith    # fullstack-monolith | frontend-backend
target_framework: nextjs16                 # Frontend framework
target_framework_backend: ""               # (separated only) Backend framework
target_backend_language: ""                # (separated only) Backend language
db_access_from: frontend                   # frontend | backend | both | none
target_db_orm: prisma                      # Target ORM (fullstack: from Q6, separated: set in Plan)
db_schema_source: env                      # env | file | none
target_ui_library: shadcn                  # UI library (shadcn, mui, chakra, legacy-css, other)

# site-flow integration (auto-configured, unless --skip-site-flow)
site_flow:
  enabled: true                            # false if --skip-site-flow or connection failed
  api_url: "http://localhost:3000"         # site-flow server URL
  api_key: "sf_xxxxxxxx"                   # API key (Bearer token) for subsequent phases
  site_id: "507f1f77bcf86cd799439011"      # site-flow site ObjectId
  site_url: "https://legacy-app.com"       # Registered site URL (source site)
  registered_at: "2026-01-23T10:00:00Z"    # Registration timestamp
```

**If `--quick` is specified**: Set all target fields to `"pending"` and skip interactive selection.
**If `--skip-site-flow` is specified**: Set `site_flow.enabled` to `false` and skip site-flow registration.

### Step 2.5: site-flow Bootstrap (CLI Authentication & Site Registration)

**Skip this step if `--skip-site-flow` is specified.**

After creating `.migrate-config.yaml`, authenticate with site-flow and register the source site using the CLI bootstrap module:

#### Step 2.5.0: Collect Credentials

Use AskUserQuestion to collect site-flow server URL and login credentials:

```
Q-SF-1: "site-flow 서버 URL을 입력하세요"
  Options:
  - http://localhost:3000 (로컬 개발 서버)
  - 직접 입력

Q-SF-2: "site-flow 로그인 이메일을 입력하세요"
  → Free text input (Other)

Q-SF-3: "site-flow 로그인 비밀번호를 입력하세요"
  → Free text input (Other)
```

#### Step 2.5.1: Run Bootstrap

```
import { bootstrapSiteFlow, BootstrapError } from '../lib/site-flow';
import { saveSiteFlowConfig } from '../lib/site-flow';

try {
  const result = await bootstrapSiteFlow({
    apiUrl: sf_api_url,           // From Q-SF-1
    email: sf_email,              // From Q-SF-2
    password: sf_password,        // From Q-SF-3
    siteName: project_name,       // From Step 2
    siteUrl: source_site_url,     // From source analysis
    apiKeyName: `migration-${project_name}`,
  });

  // Bootstrap completes 4 steps automatically:
  //   1. GET /api/auth/csrf → CSRF token + cookies
  //   2. POST /api/auth/callback/credentials → session cookie
  //   3. Find or create site (session auth)
  //   4. Create API key (session auth)

  // Save to .migrate-config.yaml
  saveSiteFlowConfig(configPath, {
    enabled: true,
    apiUrl: sf_api_url,
    apiKey: result.apiKey,        // sf_xxx format
    siteId: result.siteId,
    siteUrl: source_site_url,
    registeredAt: new Date().toISOString(),
  });

  // Inform user
  IF result.isExistingSite:
    → "기존 사이트를 재사용합니다: {result.siteName}"
  ELSE:
    → "새 사이트가 등록되었습니다: {result.siteName}"

} catch (error) {
  IF error instanceof BootstrapError:
    → Log warning: "site-flow 부트스트랩 실패 (phase: {error.phase}): {error.message}"
  ELSE:
    → Log warning: "site-flow 연결 실패: {error.message}"

  // Graceful degradation (AC-7)
  Set site_flow.enabled = false in config
  Continue to Step 3 without site-flow
}
```

**Bootstrap Authentication Flow**:
1. **CSRF Token**: `GET /api/auth/csrf` → CSRF token + session cookies
2. **Credentials Login**: `POST /api/auth/callback/credentials` → NextAuth session cookie (HTTP-only)
3. **Site Registration**: `GET /api/sites` (find by URL) or `POST /api/sites` (create new) → site_id
4. **API Key Generation**: `POST /api/api-keys` → API key (`sf_` prefix, returned once)

After bootstrap, subsequent phases (1-4) use the API key for authentication (no session needed).

**Graceful Degradation (AC-7)**: If bootstrap fails at any phase (CSRF, login, site creation, or API key generation), set `site_flow.enabled` to `false`, display a warning with the failure phase, and continue the discovery process without site-flow integration.

### Step 3: Generate Report

```markdown
# Discovery Report: {project_name}

## Source Overview
- **Language**: PHP 7.4
- **Framework**: Laravel 8
- **Database**: MySQL 5.7
- **Frontend**: jQuery + Blade

## Database Overview
- **Database**: MySQL 5.7
- **ORM**: Eloquent (Laravel)
- **Models**: 15 data models detected
- **Migrations**: 23 migration files in `database/migrations/`
- **Additional Services**: Redis (session store, cache)

## Architecture Overview
- **Pattern**: Monolith (single codebase, frontend + backend coexist)
- **Confidence**: High
- **Indicators**: Single package.json with Laravel framework, Blade templates mixed with controllers

## Complexity Score: 7/10 (Medium-High)

## Target Stack (from interactive selection)
- **Architecture**: Separated (Frontend + Backend)
- **Frontend**: Next.js 16
- **Backend**: Java (Spring Boot)
- **DB Access**: Backend only
- **Target ORM**: (to be decided in Plan phase)
- **Schema Extraction**: From .env (DATABASE_URL)
- **UI Library**: shadcn/ui (modern components will replace legacy CSS)

## site-flow Integration
- **Status**: Connected / Offline / Skipped
- **Site ID**: 507f1f77bcf86cd799439011 (or "N/A" if offline)
- **API Key**: Issued (sf_xxx...xxx) / Not issued
- **Server**: http://localhost:3000

## Config Created
`.migrate-config.yaml` has been initialized with full target stack configuration.

## Next Step
Run `/jikime:migrate-1-analyze` to perform deep analysis.
(Source path and target are already saved in .migrate-config.yaml)
```

## Config File Purpose

`.migrate-config.yaml` is the **single source of truth** for all subsequent steps:

| Field | Set by | Used by |
|-------|--------|---------|
| `source_path` | Step 0 | Step 1, 3 |
| `source_framework` | Step 0 | Step 1, 2, 3 |
| `target_framework` | Step 0 | Step 2, 3 |
| `target_architecture` | Step 0 | Step 2, 3 |
| `target_framework_backend` | Step 0 (separated only) | Step 2, 3 |
| `target_backend_language` | Step 0 (separated only) | Step 2, 3 |
| `db_access_from` | Step 0 | Step 2, 3 |
| `target_db_orm` | Step 0 (fullstack) or Step 2 (separated) | Step 3 |
| `db_schema_source` | Step 0 | Step 1, 3 |
| `target_ui_library` | Step 0 | Step 2, 3 |
| `artifacts_dir` | Step 0 (default) or Step 1 | Step 2, 3, 4 |
| `output_dir` | Step 0 (default) or Step 3 | Step 3, 4 |
| `db_type` | Step 0 | Step 1, 2, 3, 4 |
| `db_orm` | Step 0 | Step 1, 2, 3, 4 |
| `source_architecture` | Step 0 | Step 1, 2 |
| `site_flow.enabled` | Step 0 | Step 1, 2, 3, 4 |
| `site_flow.api_url` | Step 0 | Step 1, 2, 3, 4 |
| `site_flow.api_key` | Step 0 | Step 1, 2, 3, 4 |
| `site_flow.site_id` | Step 0 | Step 1, 2, 3, 4 |

**Users never need to re-enter these values** in subsequent steps.

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Exploration | `Explore` | File structure and tech detection |
| Architecture | `Explore` | Pattern identification |

## Workflow (Data Flow)

```
/jikime:migrate-0-discover @./src/  ← current
        │
        ├─ Explores: Source project analysis
        ├─ AskUserQuestion: Interactive stack selection (Q1~Q7)
        ├─ Creates: .migrate-config.yaml
        │   (source_*, target_*, db_*, artifacts_dir)
        ├─ site-flow: bootstrapSiteFlow (CSRF → login → site → apikey)
        │   (site_flow.enabled, site_id, api_key saved to config)
        │
        ↓
/jikime:migrate-1-analyze
        │ (reads config → no path re-entry needed)
        ├─ Updates: .migrate-config.yaml (enriches with details)
        ├─ Creates: {artifacts_dir}/as_is_spec.md
        ↓
/jikime:migrate-2-plan
        │ (reads config + as_is_spec.md)
        ├─ Creates: {artifacts_dir}/migration_plan.md
        ↓
/jikime:migrate-3-execute
        │ (reads config + plan)
        ├─ Creates: {output_dir}/ (migrated project)
        ├─ Updates: {artifacts_dir}/progress.yaml
        ↓
/jikime:migrate-4-verify
        │ (reads config + progress)
        ├─ Creates: {artifacts_dir}/verification_report.md
```

## Next Step

After discovery, proceed to next step:
```bash
/jikime:migrate-1-analyze
```

---

## EXECUTION DIRECTIVE

Arguments: $ARGUMENTS

1. **Parse $ARGUMENTS**:
   - Extract `@path` (source code path, required)
   - Extract `--quick` (quick overview mode, optional)
   - IF no `@path` provided: Use AskUserQuestion to ask for source path

2. **Explore source project** using Explore agent:
   ```
   Task(subagent_type="Explore", prompt="
   Analyze the project at {source_path}:
   1. Primary language and framework (with version)
   2. Build tools and package manager
   3. Database type (postgresql, mysql, sqlite, mongodb, none)
   4. ORM/schema tool (prisma, drizzle, typeorm, sequelize, mongoose, eloquent, none)
   5. Architecture pattern:
      - Check frontend/backend directory separation
      - Monorepo config (turbo.json, pnpm-workspace.yaml)
      - Fullstack indicators (Next.js API Routes, Laravel, Django, Rails)
      - Classify as: monolith, separated, or unknown
   6. File count and project complexity
   ")
   ```
   - IF `--quick`: Limit to package.json + config file analysis only

3. **Present source analysis** to user:
   - Show detected source stack summary before asking questions
   - IF `--quick`: Skip to step 5 (set all target fields to "pending")

4. **Interactive Stack Selection** (Step 1.5):

   a. **Scan installed skills** for dynamic options:
      ```
      jikime-adk skill list --tag framework   → frontend options
      jikime-adk skill list --agent backend    → backend options
      ```

   b. **Q1: Architecture Strategy**
      - Use AskUserQuestion with 2 options: Fullstack, Separated
      - Store result as `target_architecture`

   c. **Q2: Frontend Framework**
      - Use AskUserQuestion with top 3 from skills + Other
      - Store result as `target_framework`

   d. **Q3: (Separated only) Backend Framework**
      - Use AskUserQuestion with top 3 from skills + Other
      - Store result as `target_framework_backend` and `target_backend_language`

   e. **Q4: (Separated only) DB Access Layer**
      - Use AskUserQuestion: "Backend only" (Recommended) vs "Both frontend and backend"
      - Store result as `db_access_from`

   f. **Q5: DB Schema Extraction** (skip if db_type is "none")
      - Use AskUserQuestion: env variable, schema file, or source code only
      - Store result as `db_schema_source`

   g. **Q6: (Fullstack only) Target DB ORM**
      - Use AskUserQuestion with context-dependent options
      - Store result as `target_db_orm`

   h. **Q7: UI Component Library**
      - Use AskUserQuestion with top options: shadcn/ui (Recommended), MUI, Chakra UI, Keep legacy CSS, Other
      - Store result as `target_ui_library`
      - This determines how legacy components will be transformed

   i. **Derive values**:
      - Fullstack: `db_access_from` = "frontend"
      - Separated without Q4: `db_access_from` = "backend" (default)

5. **Create `.migrate-config.yaml`**:
   - Derive `project_name` from source path
   - Set all detected values (source_framework, db_type, db_orm, source_architecture)
   - Set all interactive selection values (target_architecture, target_framework, target_framework_backend, target_backend_language, db_access_from, target_db_orm, db_schema_source, target_ui_library)
   - Set default `artifacts_dir` and `output_dir`

6. **site-flow Bootstrap** (skip if `--skip-site-flow`):
   - Extract `--skip-site-flow` flag from $ARGUMENTS
   - IF `--skip-site-flow`: Set `site_flow.enabled = false` in config, skip to step 7
   - **Collect credentials** via AskUserQuestion:
     - Q-SF-1: site-flow server URL (default: `http://localhost:3000`)
     - Q-SF-2: Login email (free text)
     - Q-SF-3: Login password (free text)
   - **Run bootstrap**: `bootstrapSiteFlow({ apiUrl, email, password, siteName: project_name, siteUrl: source_site_url })`
     - Bootstrap automatically: CSRF → login → find/create site → create API key
     - On success: `saveSiteFlowConfig({ enabled: true, apiUrl, apiKey: result.apiKey, siteId: result.siteId, ... })`
     - On `BootstrapError`: Log warning with `error.phase`, set `site_flow.enabled = false`, continue to step 7
   - **Graceful Degradation (AC-7)**: On any failure, set `site_flow.enabled = false`, warn, continue

7. **Generate Discovery Report** to user in F.R.I.D.A.Y. format:
   - Source Overview (language, framework, database, frontend)
   - Database Overview (DB type, ORM, models, migrations)
   - Architecture Overview (pattern, confidence, indicators)
   - Complexity Score (1-10)
   - Target Stack (architecture, frontend, backend, DB access, ORM, schema extraction, UI library)
   - Config Created confirmation
   - Next Step: `/jikime:migrate-1-analyze`

Execute NOW. Do NOT just describe.

---

Version: 5.1.0
Changelog:
- v5.1.0: Replaced manual client-based site-flow setup with bootstrapSiteFlow() programmatic authentication; Step 2.5 now uses CLI-based NextAuth credentials login (CSRF → login → find/create site → create API key); Added credential collection via AskUserQuestion (Q-SF-1~3); Removed session auth limitation note (solved by bootstrap); Updated Workflow diagram with bootstrapSiteFlow reference; Updated EXECUTION DIRECTIVE step 6 with bootstrap flow
- v5.0.0: Added site-flow integration (Step 2.5: site registration, API key generation via findSiteByUrl/createSite/createApiKey); Added --skip-site-flow flag; Added site_flow config section (enabled, api_url, api_key, site_id, site_url, registered_at); Updated Discovery Report with site-flow Integration status; Added site_flow fields to Config File Purpose table; Updated Workflow diagram with site-flow steps; Added EXECUTION DIRECTIVE step 6 for site-flow registration; Graceful degradation on connection failure (AC-7, AC-8)
- v4.1.0: Added Q7 UI Component Library selection (shadcn, MUI, Chakra, legacy-css); Added target_ui_library to config schema; Updated conditional flow (Fullstack: 5 questions, Separated: 6 questions); UI library choice determines component modernization strategy
- v4.0.0: Replaced --target with interactive stack selection (Step 1.5); Added 6 conditional AskUserQuestion flow; Extended .migrate-config.yaml schema with target_architecture, target_framework_backend, target_backend_language, db_access_from, target_db_orm, db_schema_source; Added Target Stack to discovery report; Dynamic skill-based option detection
- v3.3.0: Added EXECUTION DIRECTIVE with $ARGUMENTS parsing and step-by-step execution flow
- v3.2.0: Added source architecture pattern detection (source_architecture field); Added Architecture Overview in discovery report
- v3.1.0: Added database type and ORM detection (db_type, db_orm fields); Added Database Overview in discovery report
- v3.0.0: Added .migrate-config.yaml creation; Added --target option; Defined data flow across steps
- v2.1.0: Initial structured discover command
