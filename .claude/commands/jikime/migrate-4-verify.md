---
description: "[Step 4/4] Migration verification. Runs /jikime:verify pre-pr on migrated project."
argument-hint: '[--skip-browser] [--headed]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

# Migration Step 4: Verify

**Final Step**: Verify the migrated project works correctly.

## What This Command Does

1. Read `output_dir` from `.migrate-config.yaml`
2. Run `/jikime:verify pre-pr` on the migrated project
3. Update `progress.yaml` with verification status

## Usage

```bash
# Verify migrated project (full verification including browser)
/jikime:migrate-4-verify

# Skip browser verification (static analysis only)
/jikime:migrate-4-verify --skip-browser

# Run with visible browser window (headed mode for manual verification)
/jikime:migrate-4-verify --headed
```

## Options

| Option | Description |
|--------|-------------|
| `--skip-browser` | Skip browser verification (pass `--no-browser` to verify) |
| `--headed` | Run browser verification in headed mode (visible browser window) |
| `--skip-site-flow` | Skip site-flow API integration (AC-8) |

---

## Execution Flow

### Step 1: Load Configuration

```
Read .migrate-config.yaml:
  - output_dir: Path to migrated project
  - artifacts_dir: Path to migration artifacts
  - target_architecture: Architecture pattern (fullstack-monolith | frontend-backend | frontend-only)
```

If `.migrate-config.yaml` not found:
- Inform user that previous steps must be completed first
- Suggest running `/jikime:migrate-3-execute` before this command

### Step 2: Run Verification

```bash
cd {output_dir}

# If --skip-browser:
/jikime:verify pre-pr --no-browser

# If --headed (visible browser for manual verification):
/jikime:verify --browser-only --headed

# Otherwise (headless browser):
/jikime:verify pre-pr
```

**What verify pre-pr checks:**
- Build compilation
- Type checking
- Lint validation
- Test execution
- Security scanning
- Database schema validation (if `db_type` is not `none`)
- Database connectivity test (if `db_type` is not `none`)
- Browser runtime errors (unless --skip-browser)

**Browser Modes:**
- **headless** (default): No visible browser, automated verification
- **headed** (`--headed`): Visible browser window for manual verification and debugging

### Step 2.5: site-flow Test Execution & Bug Reports (AC-6)

**Condition**: Skip if `--skip-site-flow` flag is set OR `site_flow.enabled` is false in `.migrate-config.yaml`.

#### Step 2.5.1: Initialize site-flow Client

```
import { loadSiteFlowConfig, createSiteFlowClient } from '../lib/site-flow';

const config = loadSiteFlowConfig();
const client = createSiteFlowClient(config);

IF client is null:
  → Log warning: "site-flow server unreachable, skipping test execution sync"
  → Continue to Step 3 without site-flow (AC-7 Graceful Degradation)
```

#### Step 2.5.2: Bulk Execute Test Cases

Execute test cases created in Phase 2 (`/migrate-2-plan`) against the migrated project.

```
import { bulkExecuteTestCases } from '../lib/site-flow';

// Collect test case IDs from .migrate-config.yaml or progress.yaml
const testCaseIds = config.site_flow.test_case_ids || [];

IF testCaseIds.length > 0:
  const executionResult = await bulkExecuteTestCases(client, {
    siteId: config.site_flow.site_id,
    testCaseIds: testCaseIds,
    trigger: 'migration_verify',
    scope: {
      type: 'site',
      scopeId: config.site_flow.site_id
    }
  });

  // Track execution ID for result retrieval
  executionId = executionResult.id;
```

#### Step 2.5.3: Create Bug Reports for Failed Tests

For each failed test case, automatically create a bug report in site-flow.

```
import { createBugReport } from '../lib/site-flow';

// Map local verification failures to site-flow bug reports
FOR each failedItem in verificationFailures:
  await createBugReport(client, {
    siteId: config.site_flow.site_id,
    title: `[Migration] ${failedItem.type}: ${failedItem.summary}`,
    description: failedItem.details,
    kind: 'migration_defect',
    severity: failedItem.severity || 'medium',
    status: 'open',
    category: mapFailureCategory(failedItem.type),
    pageId: failedItem.pageId || undefined,
    featureId: failedItem.featureId || undefined
  });
```

**Failure category mapping:**

| Verification Failure | Bug Category | Severity |
|---------------------|-------------|----------|
| Build compilation error | `build` | `critical` |
| Type checking error | `type_safety` | `high` |
| Lint violation | `code_quality` | `low` |
| Test failure | `regression` | `high` |
| Security issue | `security` | `critical` |
| Browser runtime error | `runtime` | `high` |
| DB schema mismatch | `data_integrity` | `critical` |

#### Step 2.5.4: Update Page Status to Verified

When all tests pass, update each migrated page status to `verified` in site-flow.

```
import { updatePage, getPages } from '../lib/site-flow';

IF all verifications passed:
  const pages = await getPages(client, config.site_flow.site_id);

  FOR each page in pages:
    await updatePage(client, page.id, {
      inspectionStatus: 'verified'
    });
```

#### Step 2.5.5: Export Final Report

Generate a comprehensive migration report from site-flow data.

```
import { exportSiteData } from '../lib/site-flow';

const exportResult = await exportSiteData(client, config.site_flow.site_id, {
  format: 'json',
  includePages: true,
  includeFeatures: true,
  includeTestCases: true,
  includeBugReports: true
});

// Save export to artifacts directory
Write exportResult to {artifacts_dir}/site-flow-report.json
```

#### Step 2.5.6: Handle Failures (AC-7)

```
IF any site-flow API call fails after 3 retries:
  → Log warning with failure details
  → Queue failed operations to {artifacts_dir}/.site-flow-queue.json
  → Continue verification without site-flow sync
  → Report queued items count in final summary

Queue format:
{
  "queuedAt": "2026-02-23T10:00:00Z",
  "phase": "verify",
  "operations": [
    { "type": "bulk_execute", "testCaseIds": [...], "retryCount": 3 },
    { "type": "create_bug_report", "data": {...}, "retryCount": 3 }
  ]
}
```

### Step 3: Update Progress

On success, update `{artifacts_dir}/progress.yaml`:

```yaml
status: verified
verified_at: "2026-01-26T12:00:00Z"
target_architecture: fullstack-monolith  # From config
verification:
  profile: pre-pr
  browser_check: true  # or false if --skip-browser
  db_check: true       # or false if db_type is none or frontend-only
  result: passed
```

On failure:
```yaml
status: verification_failed
verification:
  profile: pre-pr
  result: failed
  errors: [list of errors from verify]
```

---

## Workflow (Data Flow)

```
/jikime:migrate-0-discover
        ↓ (.migrate-config.yaml)
/jikime:migrate-1-analyze
        ↓ (as_is_spec.md)
/jikime:migrate-2-plan
        ↓ (migration_plan.md)
/jikime:migrate-3-execute
        ↓ (output_dir/ + progress.yaml)
/jikime:migrate-4-verify  ← current
        │
        ├─ Runs: /jikime:verify pre-pr
        ├─ Updates: progress.yaml (status: verified)
        │
        ├─ [site-flow] bulkExecuteTestCases() ─ Run test cases from Phase 2 (AC-6)
        ├─ [site-flow] createBugReport() ─ Auto-create bugs for failed tests (AC-6)
        ├─ [site-flow] updatePage() ─ Set page status to 'verified' (AC-6)
        ├─ [site-flow] exportSiteData() ─ Generate final migration report (AC-6)
        └─ [site-flow] Queue failed operations to .site-flow-queue.json (AC-7)
```

---

## EXECUTION DIRECTIVE

Arguments: $ARGUMENTS

1. **Parse flags from $ARGUMENTS**:
   - `--skip-browser`: Skip browser verification
   - `--headed`: Run browser in headed mode (visible window)
   - `--skip-site-flow`: Skip site-flow API integration (AC-8)

2. **Load configuration**:
   ```bash
   # Read .migrate-config.yaml
   cat .migrate-config.yaml
   ```
   - Extract `output_dir` and `artifacts_dir`
   - IF file not found: Inform user to run `/jikime:migrate-3-execute` first

3. **Change to output directory**:
   ```bash
   cd {output_dir}
   ```

4. **Run static analysis verification** (by `target_architecture`):

   **fullstack-monolith** (default) or **frontend-only**:
   ```bash
   cd {output_dir}
   npm run build 2>&1 | tail -30
   npx tsc --noEmit 2>&1
   npm run lint 2>&1
   npm test 2>&1
   ```

   **frontend-backend**:
   ```bash
   # Frontend verification
   cd {output_dir}/frontend
   npm run build 2>&1 | tail -30
   npx tsc --noEmit 2>&1
   npm run lint 2>&1
   npm test 2>&1

   # Backend verification
   cd {output_dir}/backend
   {build_command} 2>&1 | tail -30   # framework-specific
   {lint_command} 2>&1
   {test_command} 2>&1
   ```

   **Database verification** (skip if `db_type` is `none` or `target_architecture` is `frontend-only`):
   ```bash
   # For fullstack-monolith: run in {output_dir}
   # For frontend-backend: run in {output_dir}/backend
   npx prisma validate 2>&1  # or equivalent for target ORM
   npx prisma db pull --print 2>&1  # or equivalent dry-run connection test
   ```

5. **Run browser verification** (unless `--skip-browser`):

   **IF `--headed` flag**:
   ```bash
   # Start dev server in background
   npm run dev &
   DEV_PID=$!
   sleep 5  # Wait for server to start

   # Open browser in headed mode (visible window)
   npx playwright open http://localhost:3000

   # After manual verification, kill dev server
   kill $DEV_PID
   ```

   **ELSE (headless mode)**:
   ```bash
   # Start dev server in background
   npm run dev &
   DEV_PID=$!
   sleep 5

   # Take screenshot for verification
   npx playwright screenshot http://localhost:3000 verification-screenshot.png

   # Kill dev server
   kill $DEV_PID
   ```

6. **Update progress.yaml**:
   - On success: Set `status: verified`
   - On failure: Set `status: verification_failed` with error details

7. **site-flow Test Execution & Bug Reports** (Step 2.5, skip if `--skip-site-flow` or `site_flow.enabled` is false):
   - Initialize site-flow client via `loadSiteFlowConfig()` + `createSiteFlowClient()`
   - If client unavailable: log warning, continue without site-flow (AC-7)
   - Bulk execute test cases via `bulkExecuteTestCases()` with test case IDs from config (AC-6)
   - For each verification failure: create bug report via `createBugReport()` with severity mapping (AC-6)
   - If all verifications passed: update each page status to `verified` via `updatePage()` (AC-6)
   - Export final migration report via `exportSiteData()` to `{artifacts_dir}/site-flow-report.json` (AC-6)
   - Queue failed API operations to `{artifacts_dir}/.site-flow-queue.json` for retry (AC-7)

8. **Report results** to user in F.R.I.D.A.Y. format:
   - Verification summary (pass/fail counts, error details)
   - site-flow sync stats (if enabled): test executions, bug reports created, pages verified
   - Next Step: Deploy to staging → UAT → Production

Execute NOW. Do NOT just describe.

---

## Migration Complete!

When verification passes, migration is complete.

**Next Steps:**
1. Review the migrated code
2. Deploy to staging environment
3. User Acceptance Testing (UAT)
4. Production deployment

---

Version: 6.0.0
Changelog:
- v6.0.0: Added site-flow integration for test execution & bug reports (AC-6); Added Step 2.5 with bulkExecuteTestCases, createBugReport, updatePage (verified), exportSiteData; Added --skip-site-flow flag (AC-8); Added graceful degradation with retry queue (AC-7); Updated data flow diagram with site-flow API references
- v5.2.0: Added architecture-specific verification flows; frontend-backend runs separate verification per project; architecture field in progress.yaml
- v5.1.0: Added database schema validation and connectivity test; Added db_check to verification result
- v5.0.0: Simplified to wrapper for /jikime:verify pre-pr. Removed source↔target comparison.
- v4.0.0: Playwright-based verification with visual regression, cross-browser, accessibility
- v3.0.0: Config-first approach
