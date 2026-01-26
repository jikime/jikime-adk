
**Module Coverage (7 modules):**

| Module | Covers Phase | Key Pattern | ~Tokens |
|--------|-------------|-------------|---------|
| `server-lifecycle.md` | Phase 1 | Framework detection + health check | ~2K |
| `route-discovery.md` | Phase 2 | Artifact parse + BFS crawl | ~2K |
| `visual-regression.md` | Phase 3 | pixelmatch + threshold + viewports | ~3K |
| `behavioral-testing.md` | Phase 4 | Page load + nav + forms + API + errors | ~4K |
| `cross-browser.md` | Phase 5 | Parallel browsers + mobile emulation | ~3K |
| `performance.md` | Phase 6 | Web Vitals + timing + budgets | ~3K |
| `accessibility.md` | Phase 7 | axe-core + regression + categorization | ~3K |

### Phase 8 Output

```
Agent & Skill Implementation Summary
======================================

Agent Enhancement:
  File: agents/jikime/e2e-tester.md
  Changes:
    - Added migration mode context contract
    - Added jikime-workflow-playwright-migration skill reference
    - Defined MigrationVerificationContext input
    - Defined MigrationVerificationResult output
    - Added migration execution flow
    - Added mode comparison table

New Skill:
  Directory: .claude/skills/jikime-workflow-playwright-migration/
  Files:
    - SKILL.md (frontmatter + overview)
    - reference.md (best practices)
    - modules/ (7 verification module docs)
    - examples/ (3 TypeScript examples)

Example Code:
  1. visual-comparison.ts - Full visual regression pipeline
  2. route-crawler.ts - Artifact parse + BFS route discovery
  3. cross-browser-verify.ts - Parallel multi-browser verification

Integration:
  F.R.I.D.A.Y. → e2e-tester (migration mode) → Skill modules → Report
```

---

## Phase 9: Command & Configuration Updates

**Purpose**: Define the unified execution engine, flag validation, dependency management, CI/CD configuration, and error handling for the complete verification command.

### Step 9.1: Unified Execution Engine

**Purpose**: Master orchestrator that coordinates all verification phases in the correct order with proper dependency management.

**Execution Pipeline:**

```typescript
interface VerificationPipeline {
  config: MigrateConfig
  flags: VerificationFlags
  servers: DualServerState | null
  routes: RouteRegistry | null
  results: PhaseResults
  startTime: number
  errors: PipelineError[]
}

interface VerificationFlags {
  full: boolean
  behavior: boolean
  e2e: boolean
  visual: boolean
  performance: boolean
  crossBrowser: boolean
  a11y: boolean
  sourceUrl: string | null
  targetUrl: string | null
  port: number
  sourcePort: number
  headless: boolean
  threshold: number
  depth: number
}

interface PhaseResults {
  infrastructure?: InfrastructureResult
  routeDiscovery?: RouteRegistry
  visual?: VisualRegressionReport
  behavioral?: BehavioralTestReport
  crossBrowser?: CrossBrowserReport
  performance?: PerformanceSummary
  accessibility?: AccessibilityReport
}

interface PipelineError {
  phase: string
  step: string
  error: string
  severity: 'fatal' | 'degraded' | 'warning'
  fallback: string | null
}
```

**Master Execution Flow:**

```
FUNCTION executeVerificationPipeline(config, flags):
  pipeline = initPipeline(config, flags)

  TRY:
    // ═══ PHASE 1: Infrastructure (REQUIRED) ═══
    pipeline.servers = await startInfrastructure(config, flags)
    IF pipeline.servers.failed:
      RETURN fatalError("Dev server startup failed", pipeline)

    sourceUrl = flags.sourceUrl ?? pipeline.servers.sourceUrl
    targetUrl = flags.targetUrl ?? pipeline.servers.targetUrl

    // ═══ PHASE 2: Route Discovery (REQUIRED) ═══
    pipeline.routes = await discoverRoutes(config, targetUrl, flags.depth)
    IF pipeline.routes.totalRoutes === 0:
      RETURN fatalError("No routes discovered", pipeline)

    // ═══ PHASE 3: Visual Regression (CONDITIONAL) ═══
    IF flags.visual OR flags.full:
      TRY:
        pipeline.results.visual = await runVisualRegression(
          pipeline.routes, sourceUrl, targetUrl, flags.threshold, flags.headless
        )
      CATCH error:
        pipeline.errors.add(degradedError("Visual regression", error))

    // ═══ PHASE 4: Behavioral Testing (DEFAULT) ═══
    IF flags.behavior OR flags.full OR noSpecificFlag(flags):
      TRY:
        pipeline.results.behavioral = await runBehavioralTests(
          pipeline.routes, sourceUrl, targetUrl, flags.depth
        )
      CATCH error:
        pipeline.errors.add(degradedError("Behavioral testing", error))

    // ═══ PHASE 5: Cross-Browser (CONDITIONAL) ═══
    IF flags.crossBrowser OR flags.full:
      TRY:
        pipeline.results.crossBrowser = await runCrossBrowserTests(
          pipeline.routes, targetUrl, flags.headless
        )
      CATCH error:
        pipeline.errors.add(degradedError("Cross-browser", error))

    // ═══ PHASE 6: Performance (CONDITIONAL) ═══
    IF flags.performance OR flags.full:
      TRY:
        pipeline.results.performance = await runPerformanceComparison(
          pipeline.routes, sourceUrl, targetUrl, config.verification.performance_budget
        )
      CATCH error:
        pipeline.errors.add(degradedError("Performance", error))

    // ═══ PHASE 7: Accessibility (CONDITIONAL) ═══
    IF flags.a11y OR flags.full:
      TRY:
        pipeline.results.accessibility = await runAccessibilityVerification(
          pipeline.routes, sourceUrl, targetUrl, config.verification.accessibility
        )
      CATCH error:
        pipeline.errors.add(degradedError("Accessibility", error))

    // ═══ GENERATE REPORT ═══
    report = generateUnifiedReport(pipeline)
    WRITE report TO config.artifacts_dir + '/verification_report.md'

    RETURN pipeline

  FINALLY:
    // ═══ CLEANUP (ALWAYS) ═══
    IF pipeline.servers AND NOT flags.sourceUrl:
      await stopDualServers(pipeline.servers)

FUNCTION noSpecificFlag(flags):
  RETURN NOT (flags.visual OR flags.performance OR flags.crossBrowser
    OR flags.a11y OR flags.e2e OR flags.behavior)
  // When no specific flag is set, run behavioral by default
```

**Phase Dependency Graph:**

```
Phase 1 (Infrastructure) ─── REQUIRED
    │
    ▼
Phase 2 (Route Discovery) ── REQUIRED
    │
    ├──▶ Phase 3 (Visual) ─────── if --visual or --full
    │
    ├──▶ Phase 4 (Behavioral) ─── DEFAULT (always unless specific flag)
    │
    ├──▶ Phase 5 (Cross-Browser) ─ if --cross-browser or --full
    │
    ├──▶ Phase 6 (Performance) ── if --performance or --full
    │
    └──▶ Phase 7 (Accessibility) ─ if --a11y or --full
              │
              ▼
         Report Generation ──── ALWAYS
              │
              ▼
         Cleanup ──────────── ALWAYS (finally block)
```

**Flag Resolution Matrix:**

| User Flags | Phases Executed |
|-----------|----------------|
| (none) | 1 → 2 → 4 (behavioral) → Report |
| `--full` | 1 → 2 → 3 → 4 → 5 → 6 → 7 → Report |
| `--visual` | 1 → 2 → 3 → Report |
| `--behavior` | 1 → 2 → 4 → Report |
| `--performance` | 1 → 2 → 6 → Report |
| `--cross-browser` | 1 → 2 → 5 → Report |
| `--a11y` | 1 → 2 → 7 → Report |
| `--visual --a11y` | 1 → 2 → 3 → 7 → Report |
| `--source-url X` | Skip source server start, use X |
| `--source-url X --target-url Y` | Skip ALL server starts |

### Step 9.2: Flag Validation & Conflict Resolution

**Purpose**: Validate flag combinations and resolve conflicts before execution.

**Validation Rules:**

```typescript
interface FlagValidation {
  valid: boolean
  warnings: string[]
  corrections: FlagCorrection[]
}

interface FlagCorrection {
  flag: string
  original: any
  corrected: any
  reason: string
}

function validateFlags(flags: VerificationFlags): FlagValidation {
  const warnings: string[] = []
  const corrections: FlagCorrection[] = []

  // Rule 1: --full overrides individual flags
  if (flags.full) {
    if (flags.visual || flags.crossBrowser || flags.a11y || flags.performance) {
      warnings.push('--full already includes all verification types; individual flags ignored')
    }
  }

  // Rule 2: Port range validation
  if (flags.port < 1024 || flags.port > 65535) {
    corrections.push({
      flag: '--port', original: flags.port, corrected: 3001,
      reason: 'Port must be between 1024-65535'
    })
    flags.port = 3001
  }
  if (flags.sourcePort < 1024 || flags.sourcePort > 65535) {
    corrections.push({
      flag: '--source-port', original: flags.sourcePort, corrected: 3000,
      reason: 'Port must be between 1024-65535'
    })
    flags.sourcePort = 3000
  }

  // Rule 3: Port collision detection
  if (flags.port === flags.sourcePort) {
    corrections.push({
      flag: '--port', original: flags.port, corrected: flags.sourcePort + 1,
      reason: 'Source and target ports must differ'
    })
    flags.port = flags.sourcePort + 1
  }

  // Rule 4: Threshold range
  if (flags.threshold < 0 || flags.threshold > 100) {
    corrections.push({
      flag: '--threshold', original: flags.threshold, corrected: 5,
      reason: 'Threshold must be 0-100'
    })
    flags.threshold = 5
  }

  // Rule 5: Depth range
  if (flags.depth < 1 || flags.depth > 10) {
    corrections.push({
      flag: '--depth', original: flags.depth, corrected: 3,
      reason: 'Depth must be 1-10'
    })
    flags.depth = 3
  }

  // Rule 6: URL format validation
  if (flags.sourceUrl && !isValidUrl(flags.sourceUrl)) {
    warnings.push(`--source-url "${flags.sourceUrl}" is not a valid URL`)
  }
  if (flags.targetUrl && !isValidUrl(flags.targetUrl)) {
    warnings.push(`--target-url "${flags.targetUrl}" is not a valid URL`)
  }

  // Rule 7: Cross-browser requires headless in CI
  if (flags.crossBrowser && !flags.headless && isCI()) {
    corrections.push({
      flag: '--headless', original: false, corrected: true,
      reason: 'Cross-browser tests require headless mode in CI environment'
    })
    flags.headless = true
  }

  return {
    valid: corrections.length === 0 || corrections.every(c => c.corrected !== null),
    warnings,
    corrections
  }
}

function isCI(): boolean {
  return !!(process.env.CI || process.env.GITHUB_ACTIONS || process.env.GITLAB_CI)
}
```

**Flag Precedence Order:**

```
1. Explicit user flags (highest priority)
2. .migrate-config.yaml verification section
3. Environment detection (CI mode adjustments)
4. Defaults (lowest priority)
```

### Step 9.3: Tool & Dependency Requirements

**Purpose**: Define required packages and validate their availability before execution.

**Dependency Manifest:**

```typescript
interface DependencyCheck {
  name: string
  package: string
  version: string
  required: boolean           // true = fatal if missing
  usedBy: string[]           // Which phases need this
  installCommand: string
}

const REQUIRED_DEPENDENCIES: DependencyCheck[] = [
  {
    name: 'Playwright',
    package: '@playwright/test',
    version: '^1.40.0',
    required: true,
    usedBy: ['all phases'],
    installCommand: 'npm install -D @playwright/test && npx playwright install'
  },
  {
    name: 'pixelmatch',
    package: 'pixelmatch',
    version: '^5.3.0',
    required: false,
    usedBy: ['Phase 3: Visual Regression'],
    installCommand: 'npm install -D pixelmatch'
  },
  {
    name: 'pngjs',
    package: 'pngjs',
    version: '^7.0.0',
    required: false,
    usedBy: ['Phase 3: Visual Regression'],
    installCommand: 'npm install -D pngjs'
  },
  {
    name: 'axe-core Playwright',
    package: '@axe-core/playwright',
    version: '^4.8.0',
    required: false,
    usedBy: ['Phase 7: Accessibility'],
    installCommand: 'npm install -D @axe-core/playwright'
  }
]
```

**Pre-flight Dependency Check:**

```
FUNCTION checkDependencies(flags):
  missing = []
  optional_missing = []

  FOR each dep in REQUIRED_DEPENDENCIES:
    installed = await checkPackageInstalled(dep.package)

    IF NOT installed:
      IF dep.required:
        missing.add(dep)
      ELIF flagRequiresPhase(flags, dep.usedBy):
        optional_missing.add(dep)

  IF missing.length > 0:
    // Fatal: Required dependency missing
    PRINT "ERROR: Required dependencies not installed:"
    FOR each dep in missing:
      PRINT "  ${dep.name}: ${dep.installCommand}"
    RETURN { success: false, missing }

  IF optional_missing.length > 0:
    // Warning: Optional dependency missing, phase will be skipped
    PRINT "WARNING: Optional dependencies not installed (related phases will be skipped):"
    FOR each dep in optional_missing:
      PRINT "  ${dep.name} (used by ${dep.usedBy.join(', ')})"
      PRINT "  Install: ${dep.installCommand}"

  RETURN { success: true, missing: [], skippedPhases: optional_missing.map(d => d.usedBy) }

FUNCTION checkPackageInstalled(packageName):
  TRY:
    require.resolve(packageName)
    RETURN true
  CATCH:
    RETURN false
```

**Playwright Browser Installation:**

```
FUNCTION ensureBrowsersInstalled(flags):
  browsers_needed = ['chromium']  // Always needed

  IF flags.crossBrowser OR flags.full:
    browsers_needed.add('firefox', 'webkit')

  FOR each browser in browsers_needed:
    IF NOT browserInstalled(browser):
      PRINT "Installing Playwright browser: ${browser}..."
      await exec(`npx playwright install ${browser}`)

FUNCTION browserInstalled(browser):
  TRY:
    await exec(`npx playwright install --check ${browser}`)
    RETURN true
  CATCH:
    RETURN false
```

**Updated Allowed Tools:**

```yaml
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
```

> Note: Playwright is invoked via `Bash` tool (npx playwright) rather than as a separate MCP tool.
> The e2e-tester agent handles Playwright execution through Bash commands.

### Step 9.4: CI/CD Configuration & Headless Mode

**Purpose**: Support automated verification in CI/CD pipelines.

**CI Environment Detection:**

```typescript
interface CIEnvironment {
  detected: boolean
  provider: 'github-actions' | 'gitlab-ci' | 'jenkins' | 'circleci' | 'unknown' | null
  adjustments: CIAdjustment[]
}

interface CIAdjustment {
  setting: string
  value: any
  reason: string
}

function detectCIEnvironment(): CIEnvironment {
  const env = process.env

  if (env.GITHUB_ACTIONS) {
    return {
      detected: true,
      provider: 'github-actions',
      adjustments: [
        { setting: 'headless', value: true, reason: 'No display in GH Actions' },
        { setting: 'retries', value: 2, reason: 'Handle flaky network in CI' },
        { setting: 'timeout', value: 60000, reason: 'Longer timeout for CI containers' },
        { setting: 'workers', value: 1, reason: 'Single worker for stability' }
      ]
    }
  }

  if (env.GITLAB_CI) {
    return {
      detected: true,
      provider: 'gitlab-ci',
      adjustments: [
        { setting: 'headless', value: true, reason: 'No display in GitLab CI' },
        { setting: 'retries', value: 2, reason: 'Handle CI network variance' },
        { setting: 'timeout', value: 60000, reason: 'Longer timeout for CI' }
      ]
    }
  }

  if (env.CI) {
    return {
      detected: true,
      provider: 'unknown',
      adjustments: [
        { setting: 'headless', value: true, reason: 'Generic CI detected' }
      ]
    }
  }

  return { detected: false, provider: null, adjustments: [] }
}
```

**GitHub Actions Workflow Example:**

```yaml
# .github/workflows/migration-verify.yml
name: Migration Verification

on:
  push:
    branches: [migration/*]
  workflow_dispatch:

jobs:
  verify:
    runs-on: ubuntu-latest
    timeout-minutes: 30

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies (source)
        working-directory: ./source-project
        run: npm ci

      - name: Install dependencies (target)
        working-directory: ./migrated-project
        run: npm ci

      - name: Install Playwright browsers
        run: npx playwright install --with-deps chromium firefox webkit

      - name: Install verification dependencies
        run: npm install -D pixelmatch pngjs @axe-core/playwright

      - name: Run migration verification
        run: |
          # Start source server in background
          cd source-project && npm run dev &
          SOURCE_PID=$!

          # Start target server in background
          cd migrated-project && npm run dev &
          TARGET_PID=$!

          # Wait for servers
          npx wait-on http://localhost:3000 http://localhost:3001 --timeout 60000

          # Run verification
          /jikime:migrate-4-verify --full --headless

          # Cleanup
          kill $SOURCE_PID $TARGET_PID 2>/dev/null || true
        env:
          CI: true

      - name: Upload verification artifacts
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: verification-report
          path: |
            .migrate-artifacts/verification_report.md
            .migrate-artifacts/screenshots/
          retention-days: 14

      - name: Check verification status
        run: |
          if grep -q "Status: FAILED" .migrate-artifacts/verification_report.md; then
            echo "::error::Migration verification FAILED"
            exit 1
          fi
```

**GitLab CI Example:**

```yaml
# .gitlab-ci.yml (migration-verify job)
migration-verify:
  stage: verify
  image: mcr.microsoft.com/playwright:v1.40.0-jammy
  timeout: 30 minutes

  variables:
    CI: "true"
    PLAYWRIGHT_BROWSERS_PATH: /ms-playwright

  before_script:
    - cd source-project && npm ci && cd ..
    - cd migrated-project && npm ci && cd ..
    - npm install -D pixelmatch pngjs @axe-core/playwright

  script:
    - cd source-project && npm run dev &
    - cd migrated-project && npm run dev &
    - npx wait-on http://localhost:3000 http://localhost:3001 --timeout 60000
    - /jikime:migrate-4-verify --full --headless
    - |
      if grep -q "Status: FAILED" .migrate-artifacts/verification_report.md; then
        exit 1
      fi

  artifacts:
    when: always
    paths:
      - .migrate-artifacts/verification_report.md
      - .migrate-artifacts/screenshots/
    expire_in: 14 days

  rules:
    - if: '$CI_COMMIT_BRANCH =~ /^migration\//'
```

### Step 9.5: Error Handling & Graceful Degradation

**Purpose**: Define how the pipeline handles failures at each phase without aborting the entire verification.

**Error Severity Levels:**

```typescript
type ErrorSeverity = 'fatal' | 'degraded' | 'warning'

interface PhaseError {
  phase: string
  step: string
  error: Error
  severity: ErrorSeverity
  recovery: RecoveryAction
}

type RecoveryAction =
  | { type: 'abort'; reason: string }
  | { type: 'skip_phase'; phase: string; reason: string }
  | { type: 'retry'; maxRetries: number; delay: number }
  | { type: 'fallback'; alternative: string }
  | { type: 'continue'; warning: string }
```

**Error Classification Matrix:**

| Phase | Error Type | Severity | Recovery |
|-------|-----------|----------|----------|
| 1 (Infrastructure) | Server start failed | fatal | abort |
| 1 (Infrastructure) | Health check timeout | fatal | abort (retry 3x first) |
| 2 (Route Discovery) | No routes found | fatal | abort |
| 2 (Route Discovery) | Artifact parsing failed | degraded | fallback to crawl-only |
| 3 (Visual) | pixelmatch not installed | degraded | skip phase |
| 3 (Visual) | Screenshot capture failed | warning | skip route, continue |
| 4 (Behavioral) | Page load timeout | warning | skip route, continue |
| 4 (Behavioral) | Form test error | warning | record as failed, continue |
| 5 (Cross-Browser) | Firefox not installed | degraded | skip browser, continue |
| 5 (Cross-Browser) | Mobile emulation crash | warning | skip device, continue |
| 6 (Performance) | Web Vitals timeout | warning | use available metrics |
| 6 (Performance) | All runs failed | degraded | skip route |
| 7 (Accessibility) | axe-core not installed | degraded | skip phase |
| 7 (Accessibility) | Scan timeout | warning | skip route, continue |

**Graceful Degradation Engine:**

```typescript
async function executeWithDegradation<T>(
  phaseName: string,
  stepName: string,
  operation: () => Promise<T>,
  options: {
    retries?: number
    retryDelay?: number
    fallback?: () => Promise<T>
    onError?: (error: Error) => ErrorSeverity
    skipOnFail?: boolean
  } = {}
): Promise<{ result: T | null; error: PhaseError | null }> {
  const { retries = 1, retryDelay = 1000, fallback, onError, skipOnFail = true } = options

  for (let attempt = 0; attempt < retries; attempt++) {
    try {
      const result = await operation()
      return { result, error: null }
    } catch (error) {
      const severity = onError?.(error as Error) ?? 'warning'

      if (attempt < retries - 1) {
        // Retry with delay
        await new Promise(resolve => setTimeout(resolve, retryDelay * (attempt + 1)))
        continue
      }

      // All retries exhausted
      if (fallback) {
        try {
          const fallbackResult = await fallback()
          return {
            result: fallbackResult,
            error: {
              phase: phaseName, step: stepName, error: error as Error,
              severity: 'warning',
              recovery: { type: 'fallback', alternative: 'Used fallback strategy' }
            }
          }
        } catch (fallbackError) {
          // Fallback also failed
        }
      }

      if (skipOnFail) {
        return {
          result: null,
          error: {
            phase: phaseName, step: stepName, error: error as Error,
            severity,
            recovery: { type: 'skip_phase', phase: phaseName, reason: (error as Error).message }
          }
        }
      }

      // Fatal - abort pipeline
      return {
        result: null,
        error: {
          phase: phaseName, step: stepName, error: error as Error,
          severity: 'fatal',
          recovery: { type: 'abort', reason: (error as Error).message }
        }
      }
    }
  }

  return { result: null, error: null }
}
```

**Degradation Report Section:**

```
Verification Degradation Notes
================================

Phases Skipped (dependency not met):
  - Phase 3 (Visual): pixelmatch package not installed
    Install: npm install -D pixelmatch pngjs

Routes Skipped (errors):
  - /admin/settings: Page load timeout (30s)
  - /api/internal: HTTP 403 (authentication required)

Retries Used:
  - Phase 1: Health check succeeded on attempt 2/3 (source server slow start)
  - Phase 6: Performance collection for / succeeded on attempt 2/3

Warnings:
  - Phase 5: WebKit browser not installed; tested Chromium + Firefox only
  - Phase 4: /checkout form test got "no-response" (may need auth setup)
```

### Step 9.6: Extended Configuration Schema (Final)

**Purpose**: Complete `.migrate-config.yaml` verification schema with all Phase 1-7 settings.

```yaml
# Complete verification schema (all phases)
verification:
  # ── Infrastructure (Phase 1) ──
  dev_command: ""                    # Override auto-detection (empty = auto)
  source_port: 3000                  # Source dev server port
  target_port: 3001                  # Target dev server port
  health_check_timeout: 30           # Seconds to wait for server ready
  health_check_interval: 1           # Seconds between health checks

  # ── Route Discovery (Phase 2) ──
  crawl_depth: 3                     # BFS link discovery depth
  test_routes: []                    # Manual route overrides (empty = auto-discover)
  exclude_routes:                    # Routes to skip during verification
    - "/logout"
    - "/api/internal/*"

  # ── Visual Regression (Phase 3) ──
  visual_threshold: 5                # Allowed diff percentage (0-100)
  viewports:                         # Viewports for screenshot comparison
    - { name: "desktop", width: 1920, height: 1080 }
    - { name: "tablet", width: 768, height: 1024 }
    - { name: "mobile", width: 375, height: 667 }
  mask_selectors:                    # Dynamic content to ignore in visual diff
    - "[data-testid='timestamp']"
    - ".ad-banner"
    - "[data-dynamic]"

  # ── Behavioral Testing (Phase 4) ──
  api_monitor:
    capture_patterns:
      - "/api/*"
      - "/graphql"
    ignore_patterns:
      - "*/analytics*"
      - "*/tracking*"
    compare_mode: relaxed            # strict | relaxed

  # ── Cross-Browser (Phase 5) ──
  browsers:                          # Browsers to test (--cross-browser mode)
    - chromium
    - firefox
    - webkit
  mobile_devices:                    # Devices for mobile emulation
    - "iPhone 14 Pro"
    - "Pixel 7"

  # ── Performance (Phase 6) ──
  performance_budget:
    # Absolute thresholds
    lcp_max_ms: 2500
    cls_max: 0.1
    inp_max_ms: 200
    fcp_max_ms: 1800
    ttfb_max_ms: 800
    page_load_max_ms: 3000
    js_bundle_max_kb: 500
    css_bundle_max_kb: 150
    total_transfer_max_kb: 2000
    request_count_max: 80
    # Regression thresholds
    lcp_regression_pct: 20
    cls_regression_pct: 50
    load_time_regression_pct: 25
    js_size_regression_pct: 30
    request_count_regression_pct: 50
  performance_runs: 3                # Number of measurement runs (median used)

  # ── Accessibility (Phase 7) ──
  accessibility:
    wcag_level: "AA"                 # A | AA | AAA
    fail_on_serious: true            # Fail on serious/critical violations
    fail_on_regression: true         # Fail if new violations introduced
    score_threshold: 80              # Minimum acceptable score (0-100)
    exclude_rules: []                # axe rules to skip
    exclude_selectors:               # Elements to skip during scan
      - "[data-testid='dynamic']"
      - ".third-party-widget"
      - "iframe"
    include_best_practices: true     # Include best-practice rules

  # ── Pipeline Settings ──
  pipeline:
    fail_fast: false                 # Stop on first phase failure (default: continue)
    retry_count: 3                   # Max retries for failed operations
    retry_delay_ms: 1000             # Delay between retries
    timeout_per_route_ms: 30000      # Per-route timeout
    parallel_routes: false           # Run route tests in parallel (experimental)
    ci_mode: auto                    # auto | true | false
```

**Config Validation:**

```
FUNCTION validateConfig(config):
  errors = []
  warnings = []

  // Required fields
  IF NOT config.source_dir:
    errors.add("source_dir is required")
  IF NOT config.output_dir:
    errors.add("output_dir is required")

  // Path validation
  IF config.source_dir AND NOT exists(config.source_dir):
    errors.add("source_dir does not exist: " + config.source_dir)
  IF config.output_dir AND NOT exists(config.output_dir):
    errors.add("output_dir does not exist: " + config.output_dir)

  // Port validation
  IF config.verification.source_port === config.verification.target_port:
    errors.add("source_port and target_port must differ")

  // Threshold validation
  IF config.verification.visual_threshold < 0 OR > 100:
    warnings.add("visual_threshold should be 0-100, using default: 5")

  // Performance budget validation
  budget = config.verification.performance_budget
  IF budget.lcp_max_ms < 0:
    warnings.add("lcp_max_ms should be positive")
  IF budget.cls_max < 0 OR budget.cls_max > 1:
    warnings.add("cls_max should be 0-1")

  RETURN { valid: errors.length === 0, errors, warnings }
```

### Phase 9 Output

```
Command & Configuration Summary
==================================

Execution Engine:
  - Unified pipeline with 7 conditional phases
  - Dependency graph: Phase 1-2 required, Phase 3-7 conditional
  - Flag resolution matrix: 7 flag combinations defined
  - Default behavior: behavioral testing when no flags specified

Flag Validation:
  - 7 validation rules (port range, collision, threshold, depth, URL, CI)
  - Auto-correction for invalid values
  - CI environment detection (GitHub Actions, GitLab CI)

Dependencies:
  - Required: @playwright/test (^1.40.0)
  - Optional: pixelmatch, pngjs, @axe-core/playwright
  - Pre-flight check with install suggestions
  - Browser installation verification

CI/CD Support:
  - GitHub Actions workflow template
  - GitLab CI job template
  - Auto headless mode in CI
  - Artifact upload configuration

Error Handling:
  - 3 severity levels: fatal, degraded, warning
  - Phase-specific error classification (14 error types)
  - Retry with exponential backoff
  - Fallback strategies per phase
  - Degradation report in final output

Configuration:
  - Complete .migrate-config.yaml schema (all 7 phases)
  - Pipeline settings (fail_fast, retries, timeouts, CI mode)
  - Config validation with error/warning separation
```

