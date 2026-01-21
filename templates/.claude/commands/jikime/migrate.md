---
description: "레거시 프로젝트를 Next.js 16으로 마이그레이션. 전체 자동화 또는 단계별 실행."
argument-hint: '[plan|skill|run] "project-name" [--artifacts-output path] [--output path] [--loop] [--max N] [--whitepaper-report] [--whitepaper-output path] [--client name] [--lang ko|en|ja|zh]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, MultiEdit, Glob, Grep
model: inherit
---

# /jikime:migrate

레거시 프론트엔드 프로젝트를 Next.js 16 (App Router)로 마이그레이션합니다.

## Workflow Overview

```
/jikime:migrate-0-discover   → Step 0: 소스 탐색
        ↓
/jikime:migrate-1-analyze    → Step 1: 상세 분석
        ↓
/jikime:migrate-2-plan       → Step 2: 계획 수립
        ↓
/jikime:migrate-3-execute    → Step 3: 실행
        ↓
/jikime:migrate-4-verify     → Step 4: 검증

또는

/jikime:migrate [project]    → 전체 자동화
```

## Sub-Commands

| Sub-Command | Description | Prerequisites |
|-------------|-------------|---------------|
| (none) | Full automation | `as_is_spec.md` from migrate-1-analyze |
| `plan` | Create migration plan | `as_is_spec.md` |
| `skill` | Generate project skill | `migration_plan.md` |
| `run` | Execute migration | `SKILL.md` |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| sub-command | No | `plan`, `skill`, or `run` (omit for full automation) |
| project-name | Yes | Project name (from migrate-1-analyze) |
| --artifacts-output | No | Migration artifacts directory (default: `./migrations/{project}/`) |
| --output | No | Migrated project output directory (default: `./migrations/{project}/out/`) |
| --loop | No | Enable autonomous iteration (full automation only) |
| --max N | No | Maximum iterations (default: 100) |
| --strategy | No | Migration strategy: `incremental` or `big-bang` |
| --component | No | Migrate specific component (run sub-command only) |
| --all | No | Migrate all components (run sub-command only) |
| --whitepaper-report | No | Generate Post-Migration whitepaper report (run sub-command only) |
| --whitepaper-output | No | Whitepaper output directory (default: `./whitepaper-report/`) |
| --client | No | Client company name (used in whitepaper cover) |
| --lang | No | Whitepaper language (ko\|en\|ja\|zh). Default: user's conversation_language |

---

## Artifact Flow

```
/jikime:migrate-1-analyze "./my-vue-app"
    │
    └─▶ ./migrations/my-vue-app/as_is_spec.md
                │
/jikime:migrate plan my-vue-app
                │
                └─▶ ./migrations/my-vue-app/migration_plan.md
                            │
/jikime:migrate skill my-vue-app
                            │
                            └─▶ .claude/skills/my-vue-app/SKILL.md
                                        │
/jikime:migrate run my-vue-app [--output ./out]
                                        │
                                        └─▶ ./migrations/my-vue-app/out/ (migrated project)
                                            + ./migrations/my-vue-app/progress.yaml
```

---

## Artifact Locations

Migration workflow creates artifacts in specific directories. This is **separate** from the JikiME SPEC system.

| Workflow | Directory | Files | When Created |
|----------|-----------|-------|--------------|
| **Migration Artifacts** | `./migrations/{project}/` | `as_is_spec.md`, `migration_plan.md`, `progress.yaml` | During migration workflow |
| **Migration Skill** | `.claude/skills/{project}/` | `SKILL.md` | During `skill` sub-command |
| **Migrated Project** | `./migrations/{project}/out/` | Next.js project files | During `run` sub-command |
| **Whitepaper** | `./whitepaper-report/` | Post-migration reports | When `--whitepaper-report` used |

### SPEC System Clarification

> **Note**: The JikiME SPEC system (`.jikime/specs/SPEC-{ID}/`) is **NOT** used by this migration workflow.

If you want to create a formal SPEC document for a migration project, use:
```bash
# Creates SPEC in .jikime/specs/SPEC-{ID}/
/jikime:1-plan "Migrate {project} to Next.js 16"
```

The migration workflow and SPEC workflow are designed to be **independent**:
- **Migration workflow**: Focused on analysis and transformation artifacts
- **SPEC workflow**: Focused on formal requirements and acceptance criteria

---

## Configuration File Reference

All sub-commands automatically read `.migrate-config.yaml` to resolve artifact paths.

### Path Resolution Logic

```python
def get_artifacts_dir(project_name):
    """Resolve artifacts directory from config or use default."""
    config_path = ".migrate-config.yaml"

    if exists(config_path):
        config = load_yaml(config_path)
        if project_name in config.get("projects", {}):
            return config["projects"][project_name]["artifacts_dir"]

    # Fallback to default
    return f"./migrations/{project_name}"
```

### Override Behavior

- If `--artifacts-output` is explicitly provided, it takes precedence
- Otherwise, the path from `.migrate-config.yaml` is used
- If config doesn't exist, falls back to `./migrations/{project}/`

---

## Sub-Command: plan

Create migration plan from AS_IS analysis.

### Execution Flow

```python
# 1. Resolve artifacts directory and verify prerequisites
artifacts_dir = get_artifacts_dir(project)  # From .migrate-config.yaml or default
as_is_spec = Read(f"{artifacts_dir}/as_is_spec.md")
if not as_is_spec:
    error("Run /jikime:migrate-1-analyze first")

# 2. Delegate to manager-spec
Task(subagent_type="manager-spec", prompt="""
Create migration plan in EARS format for {project}.

Source: {as_is_spec}
Target: Next.js 16 App Router

Include:
1. Migration strategy (incremental vs big-bang)
2. Component priority list
3. Risk assessment
4. Dependency mapping
""")

# 3. Delegate to manager-strategy
Task(subagent_type="manager-strategy", prompt="""
Design target architecture for {project} migration to Next.js 16.

Consider:
1. App Router structure
2. Server vs Client components
3. State management approach
4. API integration strategy
""")
```

### API Integration Pattern

All API integrations MUST use Next.js **Route Handlers** pattern:

```
src/app/api/
├── auth/
│   ├── login/
│   │   └── route.ts       # POST /api/auth/login
│   ├── logout/
│   │   └── route.ts       # POST /api/auth/logout
│   └── me/
│       └── route.ts       # GET /api/auth/me
├── users/
│   ├── route.ts           # GET /api/users, POST /api/users
│   └── [id]/
│       └── route.ts       # GET/PUT/DELETE /api/users/[id]
└── products/
    ├── route.ts           # GET /api/products
    └── [id]/
        └── route.ts       # GET /api/products/[id]
```

**Route Handler Example**:
```typescript
// src/app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const users = await fetchUsers()
  return NextResponse.json(users)
}

export async function POST(request: NextRequest) {
  const body = await request.json()
  const user = await createUser(body)
  return NextResponse.json(user, { status: 201 })
}
```

**Migration Rules**:
| Legacy Pattern | Next.js Route Handler |
|----------------|----------------------|
| `axios.get('/api/users')` | `src/app/api/users/route.ts` → `GET()` |
| `fetch('/api/users', { method: 'POST' })` | `src/app/api/users/route.ts` → `POST()` |
| `api/users/:id` | `src/app/api/users/[id]/route.ts` |
| Express middleware | Next.js middleware or route-level logic |

---

### shadcn/ui Component Mapping

Analyze legacy UI components and map to shadcn/ui equivalents:

**Common UI Library Mappings**:
| Legacy UI Library | shadcn/ui Component | Install Command |
|-------------------|---------------------|-----------------|
| Vuetify `v-btn` | `Button` | `npx shadcn@latest add button` |
| Vuetify `v-text-field` | `Input` | `npx shadcn@latest add input` |
| Vuetify `v-select` | `Select` | `npx shadcn@latest add select` |
| Vuetify `v-dialog` | `Dialog` | `npx shadcn@latest add dialog` |
| Vuetify `v-card` | `Card` | `npx shadcn@latest add card` |
| Vuetify `v-data-table` | `Table` | `npx shadcn@latest add table` |
| Element Plus `el-button` | `Button` | `npx shadcn@latest add button` |
| Element Plus `el-input` | `Input` | `npx shadcn@latest add input` |
| Element Plus `el-form` | `Form` | `npx shadcn@latest add form` |
| Ant Design `Button` | `Button` | `npx shadcn@latest add button` |
| Ant Design `Modal` | `Dialog` | `npx shadcn@latest add dialog` |
| Ant Design `Table` | `Table` | `npx shadcn@latest add table` |
| MUI `Button` | `Button` | `npx shadcn@latest add button` |
| MUI `TextField` | `Input` | `npx shadcn@latest add input` |
| MUI `Dialog` | `Dialog` | `npx shadcn@latest add dialog` |

**Plan Output Should Include**:
```markdown
## shadcn/ui Component Requirements

| Component | Usage Count | Priority | Install Command |
|-----------|-------------|----------|-----------------|
| Button | 45 | High | `npx shadcn@latest add button` |
| Input | 32 | High | `npx shadcn@latest add input` |
| Card | 18 | Medium | `npx shadcn@latest add card` |
| Dialog | 12 | Medium | `npx shadcn@latest add dialog` |
| Table | 8 | Low | `npx shadcn@latest add table` |

### Batch Install Command
\`\`\`bash
npx shadcn@latest add button input card dialog table select form
\`\`\`
```

---

### Naming Conventions

All folders and files created during migration planning MUST follow **kebab-case** convention:

| Type | Convention | Example |
|------|------------|---------|
| Folders | kebab-case | `user-profile/`, `auth-service/`, `data-table/` |
| Files | kebab-case | `user-profile.tsx`, `auth-service.ts`, `use-auth.ts` |
| Components | kebab-case file, PascalCase export | `user-card.tsx` → `export function UserCard()` |

**Examples**:
```
# Correct (kebab-case)
src/
├── components/
│   ├── user-profile/
│   │   ├── user-profile.tsx
│   │   ├── user-avatar.tsx
│   │   └── index.ts
│   └── data-table/
│       ├── data-table.tsx
│       └── table-header.tsx
├── hooks/
│   ├── use-auth.ts
│   └── use-user-data.ts
└── services/
    └── api-client.ts

# Wrong (other conventions)
src/
├── components/
│   ├── UserProfile/          # ❌ PascalCase folder
│   ├── user_profile/         # ❌ snake_case folder
│   └── userProfile/          # ❌ camelCase folder
```

### Output

- **File**: `{artifacts_dir}/migration_plan.md`
- **Completion Marker**: `<jikime>PHASE_PLAN_COMPLETE</jikime>`

---

## Sub-Command: skill

Generate project-specific migration skill.

### Execution Flow

```python
# 1. Resolve artifacts directory and verify prerequisites
artifacts_dir = get_artifacts_dir(project)  # From .migrate-config.yaml or default
migration_plan = Read(f"{artifacts_dir}/migration_plan.md")
if not migration_plan:
    error("Run plan sub-command first")

# 2. Delegate to builder-skill
# Note: SKILL.md is written to .claude/skills/{project}/ (Claude Code skill system)
Task(subagent_type="builder-skill", prompt="""
Create migration skill for {project}.
Output: .claude/skills/{project}/SKILL.md

Based on: {migration_plan}

Generate:
1. Component mapping rules (source → target)
2. Pattern transformation examples
3. Framework-specific conversion rules
4. Coding conventions for target
5. shadcn/ui component transformation rules
""")
```

### shadcn/ui Transformation Rules

The generated SKILL.md should include shadcn/ui transformation patterns:

```markdown
## shadcn/ui Component Transformation

### Button Transformation
\`\`\`typescript
// Before (Vuetify)
<v-btn color="primary" @click="handleClick">Submit</v-btn>

// After (shadcn/ui)
import { Button } from "@/components/ui/button"
<Button variant="default" onClick={handleClick}>Submit</Button>
\`\`\`

### Input Transformation
\`\`\`typescript
// Before (Vuetify)
<v-text-field v-model="email" label="Email" />

// After (shadcn/ui)
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
<div>
  <Label htmlFor="email">Email</Label>
  <Input id="email" value={email} onChange={(e) => setEmail(e.target.value)} />
</div>
\`\`\`

### Dialog Transformation
\`\`\`typescript
// Before (Vuetify)
<v-dialog v-model="isOpen">
  <v-card>
    <v-card-title>Title</v-card-title>
    <v-card-text>Content</v-card-text>
  </v-card>
</v-dialog>

// After (shadcn/ui)
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
<Dialog open={isOpen} onOpenChange={setIsOpen}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Title</DialogTitle>
    </DialogHeader>
    <p>Content</p>
  </DialogContent>
</Dialog>
\`\`\`
```

### Output

- **File**: `.claude/skills/{project}/SKILL.md`
- **Completion Marker**: `<jikime>PHASE_SKILL_COMPLETE</jikime>`

---

## Sub-Command: run

Execute the actual migration using generated skill.

### Arguments

| Argument | Description |
|----------|-------------|
| --component Name | Migrate specific component only |
| --all | Migrate all remaining components |
| --output path | Output directory for migrated project |
| --dry-run | Preview changes without writing |

### Execution Flow

```python
# 1. Resolve artifacts directory and verify prerequisites
artifacts_dir = get_artifacts_dir(project)  # From .migrate-config.yaml or default
skill = Read(f".claude/skills/{project}/SKILL.md")
if not skill:
    error("Run skill sub-command first")

# 2. Setup Next.js project (if not exists)
output_dir = args.output or f"migrations/{project}"
if not exists(output_dir):
    Bash(f"""
    npx create-next-app@latest {output_dir} \
      --typescript --tailwind --app --src-dir \
      --import-alias "@/*" --no-turbopack

    cd {output_dir}
    npx shadcn@latest init -d
    npm install zustand lucide-react
    """)

# 3. Install shadcn/ui components from migration plan
migration_plan = Read(f"{artifacts_dir}/migration_plan.md")
shadcn_components = extract_shadcn_components(migration_plan)
if shadcn_components:
    # Install all required components in a single batch
    Bash(f"""
    cd {output_dir}
    npx shadcn@latest add {' '.join(shadcn_components)}
    """)
    # Example: npx shadcn@latest add button input card dialog table select form

# 4. Migrate components using DDD cycle
for component in get_components(skill):
    Task(subagent_type="manager-ddd", prompt=f"""
    Migrate {component.name} using ANALYZE-PRESERVE-IMPROVE cycle.

    Source: {component.source_path}
    Target: {output_dir}/src/components/{component.target_path}
    Rules: Follow SKILL.md mapping rules
    """)

    # Update progress
    update_progress(component, "completed")

# 5. Quality validation
Bash(f"cd {output_dir} && npx tsc --noEmit && npm run lint && npm run build")
```

### shadcn/ui Component Installation

The run phase automatically installs shadcn/ui components identified in `migration_plan.md`:

**Installation Process**:
1. Parse `migration_plan.md` to extract required shadcn/ui components
2. Execute batch installation: `npx shadcn@latest add <components>`
3. Components are installed to `src/components/ui/`

**Supported Components**:
| Component | Install Command | Usage |
|-----------|-----------------|-------|
| `button` | `npx shadcn@latest add button` | Primary actions, form submissions |
| `input` | `npx shadcn@latest add input` | Text input fields |
| `select` | `npx shadcn@latest add select` | Dropdown selections |
| `dialog` | `npx shadcn@latest add dialog` | Modal dialogs |
| `card` | `npx shadcn@latest add card` | Content containers |
| `table` | `npx shadcn@latest add table` | Data tables |
| `form` | `npx shadcn@latest add form` | Form with validation (react-hook-form + zod) |
| `toast` | `npx shadcn@latest add toast` | Notifications |
| `tabs` | `npx shadcn@latest add tabs` | Tabbed navigation |
| `dropdown-menu` | `npx shadcn@latest add dropdown-menu` | Context menus |

**Batch Installation Example**:
```bash
# Install multiple components at once
npx shadcn@latest add button input card dialog table select form toast
```

**Note**: The `-d` flag in `npx shadcn@latest init -d` uses default configuration (New York style, Tailwind CSS variables).

### Output

- **Directory**: `{--output}/` or `migrations/{project}/`
- **Progress**: `{artifacts_dir}/progress.yaml`
- **Completion Marker**: `<jikime>PHASE_RUN_COMPLETE</jikime>`

---

### Post-Migration Whitepaper Generation (--whitepaper-report)

When `--whitepaper-report` flag is provided after migration completion, generate a comprehensive Post-Migration report package for client delivery.

#### Language Selection (--lang)

The whitepaper can be generated in different languages:

| Language | Code | Description |
|----------|------|-------------|
| Korean | `ko` | 한국어 백서 |
| English | `en` | English whitepaper |
| Japanese | `ja` | 日本語ホワイトペーパー |
| Chinese | `zh` | 中文白皮书 |

**Default Behavior**:
- If `--lang` is not specified, uses user's `conversation_language` from `.jikime/config/language.yaml`
- Templates from `.claude/skills/jikime-migrate/templates/post-migration/` are used as structure reference
- Content is generated in the specified language by the delegated agents

#### Whitepaper Output Structure

**Output Directory**: `{--whitepaper-output}` or `./whitepaper-report/` (default)

```
{whitepaper-output}/               # Post-Migration Whitepaper (default: ./whitepaper-report/)
    ├── 00_cover.md                # Cover page and table of contents
    ├── 01_executive_summary.md    # Executive summary
    ├── 02_migration_summary.md    # Migration execution summary
    ├── 03_architecture_comparison.md  # Before/After architecture comparison
    ├── 04_component_inventory.md  # Migrated component inventory
    ├── 05_performance_report.md   # Performance improvement report
    ├── 06_quality_report.md       # Quality metrics report
    └── 07_lessons_learned.md      # Lessons learned and recommendations
```

#### Post-Migration Whitepaper Execution

```python
# After migration completion, if --whitepaper-report is provided
if args.whitepaper_report:
    # 1. Cover Page
    Task(subagent_type="manager-docs", prompt="""
    Create Post-Migration whitepaper cover page.
    Client: {--client or "Client Company"}
    Project: {project-name}
    Language: {--lang or conversation_language}
    """)

    # 2. Executive Summary (for stakeholders)
    Task(subagent_type="manager-docs", prompt="""
    Create executive summary for non-technical stakeholders.
    Include: project overview, migration success metrics, business value delivered
    """)

    # 3. Migration Summary
    Task(subagent_type="manager-docs", prompt="""
    Summarize migration execution based on progress.yaml.
    Include: timeline, phases completed, challenges overcome
    """)

    # 4. Architecture Comparison
    Task(subagent_type="expert-frontend", prompt="""
    Create before/after architecture comparison.
    Include: Mermaid diagrams, technology stack comparison, improvements
    """)

    # 5. Component Inventory
    Task(subagent_type="expert-frontend", prompt="""
    Create migrated component inventory.
    Include: component list, migration status, code quality improvements
    """)

    # 6. Performance Report
    Task(subagent_type="expert-performance", prompt="""
    Create performance improvement report.
    Include: bundle size, load times, Core Web Vitals comparison
    """)

    # 7. Quality Report
    Task(subagent_type="manager-quality", prompt="""
    Create quality metrics report.
    Include: test coverage, lint compliance, TypeScript strictness
    """)

    # 8. Lessons Learned
    Task(subagent_type="manager-strategy", prompt="""
    Document lessons learned and future recommendations.
    Include: challenges, solutions, maintenance guidelines
    """)
```

#### Post-Migration Whitepaper Quality Checklist

- [ ] All 8 documents generated (cover + 7 reports)
- [ ] Client name appears on cover page
- [ ] All Mermaid diagrams render correctly
- [ ] Performance metrics include before/after comparison
- [ ] Quality metrics are accurate and verified
- [ ] Executive summary is non-technical
- [ ] Lessons learned include actionable recommendations
- [ ] No placeholder text remains

---

## Full Automation (No Sub-Command)

When invoked without sub-command, runs complete workflow:

```bash
/jikime:migrate my-vue-app [--loop] [--output ./out]
```

### Execution Flow

```python
# 1. Resolve artifacts directory and check for as_is_spec.md
artifacts_dir = get_artifacts_dir(project)  # From .migrate-config.yaml or default
as_is = Read(f"{artifacts_dir}/as_is_spec.md")
if not as_is:
    error("Run /jikime:migrate-1-analyze '{source-path}' first")

# 2. Execute all phases
phases = ["plan", "skill", "run"]

for phase in phases:
    if args.loop:
        # Autonomous mode: continue until success
        while not phase_complete(phase):
            execute_phase(phase)
            if has_errors():
                Task(subagent_type="expert-debug", prompt="Fix errors")
    else:
        # Interactive mode: ask for approval
        execute_phase(phase)
        AskUserQuestion(f"Phase {phase} complete. Continue?")

# 3. Final validation
if all_phases_complete() and build_succeeds():
    print("<jikime>MIGRATED</jikime>")
```

---

## Progress Tracking

Progress is maintained in `{artifacts-dir}/progress.yaml` (default: `./migrations/{project}/progress.yaml`):

```yaml
project: my-vue-app
source: vue3
target: nextjs16
status: in_progress
output_dir: migrations/my-vue-app  # or custom --output path

phases:
  analyze: completed  # Done by migrate-1-analyze
  plan: completed
  skill: completed
  run: in_progress

components:
  total: 15
  completed: 8
  in_progress: 1
  pending: 6

current:
  phase: run
  component: UserProfile
  started_at: "2026-01-20T15:30:00Z"
```

---

## Quality Gates

### Plan Phase
- [ ] Strategy selected with justification
- [ ] All components prioritized
- [ ] Risks identified with mitigations
- [ ] migration_plan.md generated

### Skill Phase
- [ ] Component mappings defined
- [ ] Transformation rules documented
- [ ] SKILL.md generated

### Run Phase
- [ ] Next.js project initialized
- [ ] Components migrated
- [ ] TypeScript compiles (`tsc --noEmit`)
- [ ] Lint passes (`npm run lint`)
- [ ] Build succeeds (`npm run build`)

---

## Error Handling

| Error | Recovery |
|-------|----------|
| Missing prerequisite | Display required command |
| Type error during migration | Delegate to expert-debug |
| Build failure | Analyze error, fix, retry |
| Component migration fails | Retry with different approach |

---

## Example Usage

### Step-by-Step

```bash
# Step 1: Analyze (common command)
/jikime:migrate-1-analyze "./my-vue-app"

# Step 2: Create migration plan
/jikime:migrate plan my-vue-app

# Step 3: Generate project skill
/jikime:migrate skill my-vue-app

# Step 4: Execute migration
/jikime:migrate run my-vue-app --output ./migrated
/jikime:migrate run my-vue-app --component Header
/jikime:migrate run my-vue-app --all

# Step 5: Generate Post-Migration whitepaper (optional)
/jikime:migrate run my-vue-app --whitepaper-report --client "ABC Corp"
/jikime:migrate run my-vue-app --whitepaper-report --client "ABC Corp" --lang en
/jikime:migrate run my-vue-app --whitepaper-report --client "株式会社ABC" --lang ja

# Custom output directory for whitepaper
/jikime:migrate run my-vue-app --whitepaper-report --client "ABC Corp" --whitepaper-output ./docs/post-migration
```

### Full Automation

```bash
# After migrate-1-analyze, run full automation
/jikime:migrate my-vue-app --loop --output ./migrated

# With iteration limit
/jikime:migrate my-vue-app --loop --max 50
```

---

## Target Stack

| Technology | Version |
|------------|---------|
| Framework | Next.js 16 (App Router) |
| Language | TypeScript 5.x |
| Styling | Tailwind CSS 4.x |
| UI Components | shadcn/ui |
| Icons | lucide-react |
| State | Zustand |

---

Version: 1.4.5
Changelog:
- v1.4.5: Added shadcn/ui component workflow (plan: mapping analysis, skill: transformation rules, run: batch installation)
- v1.4.4: Added Artifact Locations section clarifying separation from JikiME SPEC system
- v1.4.3: Added Next.js Route Handlers pattern requirement for API integration in plan sub-command
- v1.4.2: Added kebab-case naming convention requirement for folders and files in plan sub-command
- v1.4.1: Removed unused assets/diagrams folder from whitepaper output structure
- v1.4.0: Added .migrate-config.yaml auto-reference for cross-command artifact path resolution
- v1.3.0: Added --artifacts-output option; Changed default artifacts path from .claude/skills/ to ./migrations/
- v1.2.0: Added --whitepaper-output option for custom whitepaper output directory
- v1.1.0: Added --whitepaper-report, --client, --lang options for Post-Migration whitepaper generation
