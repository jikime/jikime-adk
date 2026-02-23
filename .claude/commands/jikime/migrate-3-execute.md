---
description: "[Step 3/4] Execute migration using DDD methodology. ANALYZE → PRESERVE → IMPROVE cycle."
argument-hint: '[--module name] [--resume] [--dry-run]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

# Migration Step 3: Execute

**Execution Phase**: Execute the actual migration using DDD methodology.

## CRITICAL: Input Sources

**All settings are read from `.migrate-config.yaml`.** No need to re-enter source/target.

### Required Inputs (from Previous Steps)

1. **`.migrate-config.yaml`** - source_path, target_framework, artifacts_dir, output_dir
2. **`{artifacts_dir}/migration_plan.md`** - Migration plan (Step 2 output)

### Input Loading Flow

```
Step 1: Read `.migrate-config.yaml`
        → Extract: source_path, target_framework, artifacts_dir, output_dir

Step 2: Read `{artifacts_dir}/migration_plan.md`
        → Extract: module list, migration order, skill conventions

Step 3: Execute migration per module following plan
        → DO NOT ask user for source/target again
```

### Error Handling

If `.migrate-config.yaml` or `migration_plan.md` is not found:
- Inform user that previous steps must be completed first
- Suggest running `/jikime:migrate-2-plan` before this command
- DO NOT attempt to guess source/target frameworks

[SOFT] Apply --ultrathink keyword for deep migration execution analysis
WHY: Migration execution requires systematic DDD cycle management, behavior preservation verification, and incremental transformation validation
IMPACT: Sequential thinking ensures each module transformation preserves existing behavior while achieving target framework conventions

## What This Command Does

### DDD Cycle: ANALYZE → PRESERVE → IMPROVE

1. **ANALYZE** - Understand existing code behavior (from as_is_spec.md)
2. **PRESERVE** - Preserve behavior with characterization tests
3. **IMPROVE** - Transform to new code (following migration_plan.md conventions)
4. **Repeat** - Repeat for each module

## Usage

```bash
# Execute migration (reads all config from .migrate-config.yaml)
/jikime:migrate-3-execute

# Migrate specific module only
/jikime:migrate-3-execute --module auth

# Resume interrupted migration
/jikime:migrate-3-execute --resume

# Preview what would be done
/jikime:migrate-3-execute --dry-run
```

## Options

| Option | Description |
|--------|-------------|
| `--module` | Migrate specific module only |
| `--resume` | Resume from last checkpoint (reads progress.yaml) |
| `--dry-run` | Show what would be done without writing files |
| `--skip-site-flow` | Skip site-flow API integration (AC-8) |

**Note**: `source` and `target` are read from `.migrate-config.yaml`. No need to specify them.

## Execution Flow

### Step 0: Load Configuration

```python
config = load(".migrate-config.yaml")
source_path = config["source_path"]
target_framework = config["target_framework"]
artifacts_dir = config["artifacts_dir"]
output_dir = config["output_dir"]
target_arch = config.get("target_architecture", "fullstack-monolith")
backend_framework = config.get("target_framework_backend", None)
db_access = config.get("db_access_from", "frontend")
ui_library = config.get("target_ui_library", "legacy-css")  # NEW: UI library setting

plan = load(f"{artifacts_dir}/migration_plan.md")
modules = extract_modules(plan)
ui_component_map = extract_ui_component_mapping(plan)  # NEW: Component transformation rules
```

### Step 1: Initialize Target Project (by architecture)

Based on `target_architecture` from config and `migration_plan.md`:

#### fullstack-monolith (default)
- Single Next.js project creation in `{output_dir}/`
- Install all dependencies
- **UI Library setup** (CRITICAL - Claude Code 직접 수행):
  - **Claude Code는 스스로 인지**: Next.js 프로젝트면 **반드시** UI 라이브러리 초기화
  - **스크립트/템플릿 결과와 무관하게** 다음을 실행:
  ```bash
  cd {output_dir} && npx shadcn@latest init --defaults
  cd {output_dir} && npx shadcn@latest add button input card dialog tabs alert badge
  ```
  - **레거시 CSS 복사 금지**: shadcn 컴포넌트로 UI 현대화
- **Database setup** (if `db_type` is not `none`):
  - Install target ORM (Prisma/Drizzle) in the same project
  - Create schema from migration plan
  - Configure database connection

#### frontend-backend
- Create `{output_dir}/shared/` for types and API contracts
- Create `{output_dir}/frontend/` with Next.js project
  - **UI Library setup** in frontend (if `target_ui_library` is not `legacy-css`):
    - Initialize UI library: `cd frontend && npx shadcn@latest init`
    - Add required base components
    - Configure Tailwind CSS / theming
- Create `{output_dir}/backend/` with `{backend_framework}` project
- **Database setup** in backend:
  - Install target ORM in backend project
  - Create schema from migration plan
  - Configure database connection in backend

#### frontend-only
- Single Next.js project creation in `{output_dir}/`
- **UI Library setup** (if `target_ui_library` is not `legacy-css`):
  - Initialize UI library: `npx shadcn@latest init` (or equivalent)
  - Add required base components
  - Configure Tailwind CSS / theming
- Configure API client for existing backend
- **No database setup** (existing backend handles data access)

### Step 2: Migrate Each Module (by architecture)

#### fullstack-monolith (default)

Standard DDD cycle for each module:

```python
for module in modules:
    # ANALYZE: Read source module from source_path
    source_code = read_module(source_path, module)

    # ANALYZE-DB: Identify data models and queries in this module
    db_dependencies = analyze_db_usage(source_code, module)

    # ANALYZE-UI: Identify UI components and patterns in this module
    ui_components = analyze_ui_patterns(source_code, module)

    # PRESERVE: Create characterization tests
    create_characterization_tests(source_code, module)

    # PRESERVE-DB: Create data layer tests (query results, relationships)
    if db_dependencies:
        create_db_characterization_tests(db_dependencies, module)

    # IMPROVE: Transform to target framework
    transform_module(source_code, module, target_framework)

    # IMPROVE-DB: Convert ORM/data access patterns to target ORM
    if db_dependencies:
        transform_db_layer(db_dependencies, module, target_orm)

    # IMPROVE-UI: Transform UI components using target UI library
    if ui_components and ui_library != "legacy-css":
        transform_ui_components(ui_components, module, ui_library, ui_component_map)
        # - Legacy button → <Button variant="...">
        # - Legacy input → <Input />
        # - Legacy modal → <Dialog>
        # - DO NOT copy legacy CSS classes
    elif ui_components and ui_library == "legacy-css":
        copy_legacy_styles(ui_components, module)  # Only if explicitly requested

    # Validate: Build and test
    validate_module(output_dir, module)

    # Track progress
    update_progress(artifacts_dir, module, "completed")
```

#### frontend-backend

Separated execution in 4 sub-phases:

```python
# Sub-Phase 1: Shared Layer
create_shared_types(output_dir + "/shared/", plan)
define_api_contracts(output_dir + "/shared/", plan)

# Sub-Phase 2: Backend
for module in backend_modules:
    # ANALYZE: Source API/business logic understanding
    source_code = read_module(source_path, module)
    db_dependencies = analyze_db_usage(source_code, module)

    # PRESERVE: Characterization tests (API responses)
    create_api_characterization_tests(source_code, module)
    if db_dependencies:
        create_db_characterization_tests(db_dependencies, module)

    # IMPROVE: Transform to target backend framework
    transform_backend_module(source_code, module, backend_framework)
    if db_dependencies:
        transform_db_layer(db_dependencies, module, target_orm)

    validate_module(output_dir + "/backend/", module)
    update_progress(artifacts_dir, module, "completed")

# Sub-Phase 3: Frontend
for module in frontend_modules:
    # ANALYZE: Source component understanding
    source_code = read_module(source_path, module)

    # ANALYZE-UI: Identify UI components and patterns
    ui_components = analyze_ui_patterns(source_code, module)

    # PRESERVE: Characterization tests (UI behavior)
    create_characterization_tests(source_code, module)

    # IMPROVE: Transform to Next.js (using API client to call backend)
    transform_frontend_module(source_code, module, "nextjs16")

    # IMPROVE-UI: Transform UI components using target UI library
    if ui_components and ui_library != "legacy-css":
        transform_ui_components(ui_components, module, ui_library, ui_component_map)
        # Use modern UI library components instead of legacy HTML/CSS
    elif ui_components and ui_library == "legacy-css":
        copy_legacy_styles(ui_components, module)

    validate_module(output_dir + "/frontend/", module)
    update_progress(artifacts_dir, module, "completed")

# Sub-Phase 4: Integration
verify_api_contract_match(output_dir)
test_frontend_backend_communication(output_dir)
```

#### frontend-only

Frontend modules only, DB steps skipped:

```python
for module in frontend_modules:
    # ANALYZE: Source component understanding
    source_code = read_module(source_path, module)

    # ANALYZE-UI: Identify UI components and patterns
    ui_components = analyze_ui_patterns(source_code, module)

    # PRESERVE: Create characterization tests
    create_characterization_tests(source_code, module)

    # IMPROVE: Transform to Next.js (API client calls existing backend)
    transform_module(source_code, module, target_framework)

    # IMPROVE-UI: Transform UI components using target UI library
    if ui_components and ui_library != "legacy-css":
        transform_ui_components(ui_components, module, ui_library, ui_component_map)
    elif ui_components and ui_library == "legacy-css":
        copy_legacy_styles(ui_components, module)

    # Validate: Build and test (no DB validation)
    validate_module(output_dir, module)

    # Track progress
    update_progress(artifacts_dir, module, "completed")
```

### Step 2.5: site-flow Screenshot Capture & Storage (AC-5)

**Condition**: Skip if `--skip-site-flow` is set OR `site_flow.enabled` is false in `.migrate-config.yaml`

After each module's DDD cycle completes, capture and store screenshots in site-flow for visual verification tracking.

#### Step 2.5.1: Initialize site-flow Client

```python
sf_config = loadSiteFlowConfig()
sf_client = createSiteFlowClient(sf_config)
if not sf_client:
    log_warning("site-flow unavailable, skipping screenshot capture (AC-7)")
    skip_site_flow = True
```

#### Step 2.5.2: Capture Original Page Screenshots (per module)

After each module migration completes:

```python
for module in completed_modules:
    page_id = get_site_flow_page_id(module)  # From .migrate-config.yaml mapping
    if not page_id:
        continue

    # Capture original page screenshot (before migration)
    original_screenshot = capture_screenshot(source_url + module.path)
    if original_screenshot:
        addPageImage(sf_client, page_id, {
            "source": "migration",
            "label": f"Original - {module.name}",
            "base64Data": fileToBase64DataUri(original_screenshot),
            "metadata": { "phase": "execute", "type": "original", "module": module.name }
        })

    # Capture migrated page screenshot (after migration)
    migrated_screenshot = capture_screenshot(localhost_url + module.path)
    if migrated_screenshot:
        image_response = addPageImage(sf_client, page_id, {
            "source": "migration",
            "label": f"Migrated - {module.name}",
            "base64Data": fileToBase64DataUri(migrated_screenshot),
            "metadata": { "phase": "execute", "type": "migrated", "module": module.name }
        })

        # Set migrated screenshot as page thumbnail
        if image_response:
            setPageThumbnail(sf_client, page_id, image_response.id)
```

#### Step 2.5.3: SSE Batch Capture (for bulk pages)

When multiple modules are completed, use SSE streaming batch capture:

```python
completed_page_ids = [get_site_flow_page_id(m) for m in completed_modules if get_site_flow_page_id(m)]

if len(completed_page_ids) > 3:
    # Use SSE batch capture for efficiency (AC-5)
    sse_stream = autoCapturePages(sf_client, site_id, {
        "pageIds": completed_page_ids,
        "source": "migration",
        "captureOptions": { "fullPage": True, "waitForFonts": False }
    })

    # Parse SSE stream for real-time progress display
    handler = createSSEProgressHandler({
        "onProgress": lambda event: display_capture_progress(event),
        "onComplete": lambda event: log_capture_complete(event),
        "onError": lambda event: log_capture_error(event)
    })
    parseSSEStream(sse_stream, handler)
```

#### Step 2.5.4: Large Image Handling (EC-1)

```python
# If screenshot exceeds 10MB size limit:
if not isImageWithinSizeLimit(screenshot_data):
    # Chunk the image for upload
    chunks = chunkBase64Image(screenshot_data)
    for chunk in chunks:
        addPageImage(sf_client, page_id, {
            "source": "migration",
            "label": f"Chunk {chunk.index} - {module.name}",
            "base64Data": chunk.data
        })
```

#### Step 2.5.5: Update Config & Handle Failures

```python
# Update .migrate-config.yaml with capture stats
config["site_flow"]["screenshots_captured"] = captured_count
config["site_flow"]["screenshots_captured_at"] = datetime.now().isoformat()
saveSiteFlowConfig(config)

# On individual API failure: queue for retry (AC-7)
if failed_uploads:
    queue_to_retry_file(".site-flow-queue.json", failed_uploads)
    log_warning(f"{len(failed_uploads)} screenshots queued for retry")
```

### Step 3: Quality Validation (by architecture)

#### fullstack-monolith (default)
- TypeScript compiles: `npx tsc --noEmit`
- Lint passes: `npm run lint`
- Build succeeds: `npm run build`
- Characterization tests pass
- Database schema validation (if applicable)
- Database connectivity verified (if applicable)

#### frontend-backend
- **Frontend**: `cd {output_dir}/frontend && npm run build && npx tsc --noEmit && npm run lint`
- **Backend**: `cd {output_dir}/backend && {build_command} && {lint_command} && {test_command}`
- **DB validation**: `cd {output_dir}/backend && npx prisma validate` (or equivalent)
- **Integration**: Frontend → Backend API communication test

#### frontend-only
- TypeScript compiles: `npx tsc --noEmit`
- Lint passes: `npm run lint`
- Build succeeds: `npm run build`
- Characterization tests pass
- No database validation (existing backend handles data)

## Progress Tracking

Progress is saved to `{artifacts_dir}/progress.yaml`:

```yaml
project: my-vue-app
source_framework: vue3            # From config
target_framework: nextjs16        # From config
target_architecture: fullstack-monolith  # From config
status: in_progress

modules:
  total: 15
  completed: 8
  in_progress: 1
  failed: 0
  pending: 6

current:
  module: UserProfile
  phase: IMPROVE
  iteration: 2
  started_at: "2026-01-23T10:30:00Z"

history:
  - module: auth
    status: completed
    duration: "5m"
  - module: users
    status: completed
    duration: "8m"
```

## Progress Display

```
╔══════════════════════════════════════════════════════════╗
║  Migration: {project_name}                               ║
║  Source: {source_framework} → Target: {target_framework} ║
║  Phase: IMPROVE                                          ║
║  Module: user-service                                    ║
║  Progress: [████████████░░░░░░░░] 60%                   ║
╚══════════════════════════════════════════════════════════╝
```

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Analysis | `Explore` | Source code understanding |
| Test Creation | `test-guide` | Characterization tests |
| Code Generation | `frontend` or `backend` | Target code creation |
| Validation | `debugger` | Build/test error fixing |

## Workflow (Data Flow)

```
/jikime:migrate-0-discover
        ↓ (.migrate-config.yaml created)
/jikime:migrate-1-analyze
        ↓ (config updated + as_is_spec.md)
/jikime:migrate-2-plan
        ↓ (migration_plan.md)
/jikime:migrate-3-execute  ← current
        │
        ├─ Reads: .migrate-config.yaml (source, target, paths, site_flow config)
        ├─ Reads: {artifacts_dir}/migration_plan.md (modules, order)
        ├─ Creates: {output_dir}/ (migrated project)
        ├─ Updates: {artifacts_dir}/progress.yaml
        │
        ├─ [site-flow] addPageImage() ─ Upload original & migrated screenshots (AC-5)
        ├─ [site-flow] autoCapturePages() ─ SSE batch capture for bulk pages (AC-5)
        ├─ [site-flow] setPageThumbnail() ─ Set migrated screenshot as thumbnail
        ├─ [site-flow] chunkBase64Image() ─ Handle >10MB images (EC-1)
        ├─ [site-flow] Queue failed uploads to .site-flow-queue.json (AC-7)
        │
        ↓
/jikime:migrate-4-verify
```

## Next Step

After execution, proceed to next step:
```bash
/jikime:migrate-4-verify
```

---

## EXECUTION DIRECTIVE

Arguments: $ARGUMENTS

1. **Parse $ARGUMENTS**:
   - Extract `--module` (specific module name)
   - Extract `--resume` (resume from last checkpoint)
   - Extract `--dry-run` (preview mode, no file writes)
   - Extract `--skip-site-flow` (skip site-flow API integration, AC-8)

2. **Load configuration**:
   - Read `.migrate-config.yaml`
   - Extract: `source_path`, `target_framework`, `artifacts_dir`, `output_dir`
   - Extract: `target_architecture` (default: `fullstack-monolith`), `target_framework_backend`, `db_access_from`
   - Extract: `target_ui_library` (default: `legacy-css`) - CRITICAL for UI modernization
   - IF file not found: Inform user to run `/jikime:migrate-2-plan` first

3. **Load migration plan**:
   - Read `{artifacts_dir}/migration_plan.md`
   - Extract module list, migration order, skill conventions
   - IF file not found: Inform user to run `/jikime:migrate-2-plan` first
   - IF `--resume`: Read `{artifacts_dir}/progress.yaml` and skip completed modules
   - IF `--module`: Filter to specified module only

4. **Initialize target project** (by `target_architecture`):
   - **fullstack-monolith**: Single Next.js project in `{output_dir}/`, UI library setup, DB setup if applicable
   - **frontend-backend**: `{output_dir}/shared/` + `{output_dir}/frontend/` (with UI library) + `{output_dir}/backend/`, DB in backend
   - **frontend-only**: Single Next.js project in `{output_dir}/`, UI library setup, API client setup, no DB
   - **UI library setup** (if `target_ui_library` != `legacy-css`):
     - `npx shadcn@latest init` (or equivalent for chosen library)
     - Add base components: `npx shadcn@latest add button input card dialog tabs form`
     - Configure Tailwind CSS and design tokens
   - IF `--dry-run`: Report initialization plan without executing

5. **Execute DDD cycle per module** (by `target_architecture`):
   - **fullstack-monolith**: Standard ANALYZE → ANALYZE-DB → ANALYZE-UI → PRESERVE → PRESERVE-DB → IMPROVE → IMPROVE-DB → IMPROVE-UI → validate
   - **frontend-backend**: Sub-Phase 1 (shared) → Sub-Phase 2 (backend modules) → Sub-Phase 3 (frontend modules with UI transformation) → Sub-Phase 4 (integration)
   - **frontend-only**: ANALYZE → ANALYZE-UI → PRESERVE → IMPROVE → IMPROVE-UI → validate (no DB steps)
   - **IMPROVE-UI step** (if `target_ui_library` != `legacy-css`):
     - Transform legacy HTML/CSS to modern UI library components
     - Use component mapping from migration_plan.md
     - DO NOT copy legacy CSS - use UI library's theming system
   - After each module: update `{artifacts_dir}/progress.yaml`
   - IF `--dry-run`: Report each module's plan without executing

6. **Quality validation** (by `target_architecture`):
   - **fullstack-monolith**: `npx tsc --noEmit`, `npm run lint`, `npm run build`, tests, DB validation
   - **frontend-backend**: Frontend build + Backend build + DB validation + integration test
   - **frontend-only**: `npx tsc --noEmit`, `npm run lint`, `npm run build`, tests

7. **site-flow Screenshot Capture** (Step 2.5, skip if `--skip-site-flow` or `site_flow.enabled` is false):
   - Initialize site-flow client via `loadSiteFlowConfig()` + `createSiteFlowClient()`
   - If client unavailable: log warning, continue without site-flow (AC-7)
   - Per completed module: capture original + migrated screenshots via `addPageImage()`
   - Set migrated screenshot as thumbnail via `setPageThumbnail()`
   - If >3 completed pages: use SSE batch capture via `autoCapturePages()` with `parseSSEStream()`
   - Handle >10MB images via `chunkBase64Image()` (EC-1)
   - Queue failed uploads to `.site-flow-queue.json` for retry (AC-7)

8. **Report results** to user in F.R.I.D.A.Y. format:
   - Progress summary, module status, site-flow capture stats (if enabled)
   - Next Step: `/jikime:migrate-4-verify`

Execute NOW. Do NOT just describe.

---

Version: 4.0.0
Changelog:
- v4.0.0: Added site-flow integration for screenshot capture & storage (AC-5); Added Step 2.5 with addPageImage, autoCapturePages SSE, setPageThumbnail; Added --skip-site-flow flag (AC-8); Added graceful degradation with retry queue (AC-7); Added large image chunking (EC-1); Updated data flow diagram with site-flow API references; Added EXECUTION DIRECTIVE step 7 for site-flow operations
- v3.4.0: Added UI-aware DDD cycle (ANALYZE-UI, IMPROVE-UI); Added target_ui_library config support; Added UI library initialization in project setup; Modern UI library components replace legacy CSS (shadcn, MUI, Chakra); Legacy CSS only preserved when explicitly requested
- v3.3.0: Added EXECUTION DIRECTIVE with $ARGUMENTS parsing and step-by-step execution flow
- v3.2.0: Added architecture-specific execution flows (fullstack-monolith, frontend-backend, frontend-only); Added sub-phase execution for frontend-backend; Architecture field in progress.yaml
- v3.1.0: Added DB-aware DDD cycle (ANALYZE-DB, PRESERVE-DB, IMPROVE-DB); Added database setup in project initialization; Added DB quality validation
- v3.0.0: Removed redundant source/target options; Config-first approach; All settings from .migrate-config.yaml
- v2.1.0: Initial DDD-based execution command
Methodology: DDD (ANALYZE-PRESERVE-IMPROVE) with UI modernization
