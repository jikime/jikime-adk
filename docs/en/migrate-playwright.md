# Migration Playwright Verification System - Implementation Plan

Post-migration testing and verification system leveraging Playwright capabilities for comprehensive quality assurance.

## Overview

After migration execution (`migrate-3-execute`) completes, this system automatically:
1. Starts dev servers (source + target)
2. Discovers all routes from migration artifacts
3. Runs comprehensive Playwright-based verification
4. Generates visual comparison reports
5. Validates behavioral preservation

---

## Phase 1: Infrastructure (Dev Server Lifecycle) - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md` (v3.0.0 → v4.0.0)

| # | Task | Description | Status |
|---|------|-------------|--------|
| 1.1 | Dev Server Manager Logic | Config read → framework detect → start/wait/stop pipeline | Done |
| 1.2 | Framework Dev Command Mapping | 16 frameworks mapped with port flag detection | Done |
| 1.3 | Health Check Pattern | 30s timeout, 1s interval, HTTP status < 500 check | Done |
| 1.4 | Dual Server Mode | Source (port 3000) + Target (port 3001) simultaneous startup | Done |

### Implementation Details

**Step 1.1**: Config read from `.migrate-config.yaml` - extracts source_dir, output_dir, framework names, verification settings

**Step 1.2**: 16 frameworks supported with detection algorithm:
- Detection priority: config override → package.json scripts → framework fallback
- Lockfile detection: pnpm-lock.yaml / yarn.lock / package-lock.json / bun.lockb

```yaml
# Full framework mapping (16 frameworks)
next: "npx next dev --port {port}"
vite: "npx vite --port {port}"
cra: "npx react-scripts start"       # PORT={port} (env)
nuxt: "npx nuxt dev --port {port}"
angular: "npx ng serve --port {port}"
remix: "npx remix dev --port {port}"
gatsby: "npx gatsby develop -p {port}"
astro: "npx astro dev --port {port}"
sveltekit: "npx vite dev --port {port}"
express: "node server.js"             # PORT={port} (env)
django: "python manage.py runserver 0.0.0.0:{port}"
flask: "flask run --port {port}"
spring: "./mvnw spring-boot:run -Dserver.port={port}"
rails: "rails server -p {port}"
laravel: "php artisan serve --port={port}"
go: "go run ."                        # PORT={port} (env)
```

**Step 1.3**: Dependency installation before server startup (npm/pnpm/yarn/bun/pip/go mod)

**Step 1.4-1.5**: Background process with PID tracking + health check algorithm

**Step 1.6**: Mandatory cleanup (SIGTERM → 5s wait → SIGKILL)

**Step 1.7**: URL resolution (auto-detected or --source-url/--target-url override)

### Additional Changes in v4.0.0

- **8 new flags**: `--visual`, `--cross-browser`, `--a11y`, `--port`, `--source-port`, `--headless`, `--threshold`, `--depth`
- **`.migrate-config.yaml` verification schema**: dev_command, ports, threshold, routes, masks, performance budget
- **allowed-tools**: Added `Edit`
- **ultrathink block**: Added for deep verification strategy analysis
- **10-step workflow**: Expanded from 5 to 10 verification types
- **Agent delegation table**: Updated with route discovery + optimizer

---

## Phase 2: Route Auto-Discovery & Test Generation - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md`

| # | Task | Description | Status |
|---|------|-------------|--------|
| 2.1 | Route Discovery Engine | 5-priority source discovery (config → artifacts → file-based → router → crawl) | Done |
| 2.2 | Dynamic Route Handling | 7 framework patterns + sample URL generation + exclusion rules | Done |
| 2.3 | Test Case Auto-Generation | Page load test + behavior comparison test templates | Done |
| 2.4 | User Flow Extraction | Common flow templates (6 types) + flow detection algorithm | Done |
| 2.5 | Route Registry Output | YAML registry format for passing to Phase 3-7 (bonus) | Done |

### Implementation Details

**Step 2.1 - Route Discovery Engine**:
- 5-level priority discovery: manual override → migration artifacts → file-based routing → router config → navigation crawl
- File-based routing support: Next.js (Pages/App), Nuxt, SvelteKit, Remix, Angular
- Router config parsing: react-router, vue-router, angular, express
- Navigation crawl with configurable depth (--depth flag)
- Deduplication and sorting of discovered routes

**Step 2.2 - Dynamic Route Handling**:
- Pattern recognition for 7 frameworks: `[param]`, `[...slug]`, `:param`
- Sample URL generation: `:id` → `1`, `:slug` → `example-page`, `:category` → `general`
- Exclusion rules: `/api/*`, `/_next/*`, `/_nuxt/*`, static assets, admin routes

**Step 2.3 - Test Case Auto-Generation**:
- Per-route verification: HTTP status < 400, no console.error, no pageerror, content not blank
- Network error collection: response status >= 400 tracking
- Behavior comparison: Source vs Target heading/navigation matching
- `waitUntil: 'networkidle'` for reliable page load detection

**Step 2.4 - User Flow Extraction**:
- 3 detection sources: migration_plan.md, existing E2E files, auto-detection
- 6 common flow templates: Authentication, Navigation, Search, Form Submit, CRUD, Pagination
- Generated E2E flow tests with smart selector patterns (`[name="email"], [type="email"], #email`)

**Step 2.5 - Route Registry (Bonus)**:
- YAML output format with static_routes, dynamic_routes, user_flows, excluded_routes, stats
- Priority classification: critical / high / medium / low
- Source tracking for each discovered route

### Auto-Generated Test Template

```typescript
// Per-route verification
test('{route} - page loads without errors', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', err => errors.push(err.message))
  page.on('console', msg => { if (msg.type() === 'error') errors.push(msg.text()) })

  const response = await page.goto('{target_url}{route}', { waitUntil: 'networkidle' })
  expect(response?.status()).toBeLessThan(400)
  expect(errors).toHaveLength(0)

  // Key element verification
  await expect(page.locator('h1')).toBeVisible()
})
```

---

## Phase 3: Visual Regression (Core Feature) - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md`

| # | Task | Description | Status |
|---|------|-------------|--------|
| 3.1 | Screenshot Capture Engine | Full-page + viewport screenshots with dynamic content masking | Done |
| 3.2 | Pixel Comparison Engine | ComparisonResult interface, color matching with tolerance (10) | Done |
| 3.3 | Responsive Viewport Matrix | 5 viewports (Desktop/Laptop/Tablet/Mobile/Mobile Small) | Done |
| 3.4 | Threshold & Masking Configuration | 4 levels (Strict/Normal/Relaxed/Loose) + per-route overrides | Done |
| 3.5 | Visual Report Generation | HTML template with grid layout (source/target/diff side-by-side) | Done |
| 3.6 | Diff Analysis & Categorization | 6 diff categories with detection logic (bonus) | Done |

### Implementation Details

**Step 3.1 - Screenshot Capture Engine**:
- `capturePageScreenshot()` function with viewport, masking, and networkidle wait
- Storage structure: `{artifacts_dir}/screenshots/{source|target|diff}/{viewport}/{route}.png`
- Dynamic content masking via `visibility: hidden` before capture
- 1000ms animation settle wait after networkidle

**Step 3.2 - Pixel Comparison Engine**:
- `ComparisonResult` interface: route, viewport, diffPercentage, diffPixels, totalPixels, passed, diffImagePath
- `compareScreenshots()` function with configurable threshold
- Color tolerance: `colorsMatch()` with per-channel tolerance (default: 10)
- Size difference handling: uses max dimensions for comparison
- Diff image: red pixels for differences, dimmed original for matches

**Step 3.3 - Responsive Viewport Matrix (5 viewports)**:

| Viewport | Width | Height | Represents |
|----------|-------|--------|------------|
| Desktop | 1920 | 1080 | Standard monitor |
| Laptop | 1366 | 768 | Common laptop |
| Tablet | 768 | 1024 | iPad portrait |
| Mobile | 375 | 812 | iPhone X/12/13 |
| Mobile Small | 320 | 568 | iPhone SE |

- `--full`: All 5 viewports
- `--visual` (no cross-browser): Desktop + Tablet + Mobile (3)
- Default: Desktop only (1)
- Responsive-specific checks: horizontal overflow detection, element overlap detection

**Step 3.4 - Threshold & Masking Configuration**:

| Level | Percentage | Use Case |
|-------|-----------|----------|
| Strict | 1% | Pixel-perfect (same CSS framework) |
| Normal | 5% | Default - minor rendering differences allowed |
| Relaxed | 10% | Framework changes with known visual diffs |
| Loose | 20% | Major redesigns where layout is similar |

- Per-route threshold override via `verification.route_thresholds` in config
- Enhanced mask_selectors: timestamps, random IDs, ads, avatars, skeletons, spinners, `<time>` elements

**Step 3.5 - Visual Report Generation**:
- HTML report at `{artifacts_dir}/visual-report.html`
- Grid layout: 3-column (Source | Target | Diff)
- Summary table with route, viewport, diff %, status
- Color-coded borders: green (pass), red (fail), orange (warn)
- WARN threshold: 80-100% of configured threshold

**Step 3.6 - Diff Analysis & Categorization (Bonus)**:

| Category | Detection Method | Severity |
|----------|-----------------|----------|
| Layout Shift | Large contiguous diff regions (100x100+) | High |
| Color Change | Scattered diffs with consistent color offset | Medium |
| Font Rendering | Text-area-only diffs with similar shapes | Low |
| Missing Element | Target has blank where source has content | Critical |
| Extra Element | Target has content not in source | Medium |
| Size Difference | Page height/width mismatch | High |

- `categorizeDiff()` function with cascading detection logic
- Actionable feedback per category for developer guidance

---

## Phase 4: Behavioral Testing (Functional Verification) - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md`

| # | Task | Description | Status |
|---|------|-------------|--------|
| 4.1 | Page Load Verification | HTTP < 400, no pageerror, content present, network error tracking | Done |
| 4.2 | Navigation Verification | Link crawl with depth, broken link detection, URL resolution | Done |
| 4.3 | Form Interaction Verification | Form type detection (5 types), auto-fill, submit verification | Done |
| 4.4 | API Call Verification | Capture/ignore patterns, source vs target comparison (strict/relaxed) | Done |
| 4.5 | JavaScript Error Collection | 6 error categories, 4 severity levels, deduplication, summary | Done |
| 4.6 | Behavioral Comparison | Source vs Target heading/navigation/content matching (bonus) | Done |

### Implementation Details

**Step 4.1 - Page Load Verification**:
- `PageLoadResult` interface: status, loadTime, consoleErrors, pageErrors, networkErrors, hasContent, passed
- `NetworkError` interface: url, status, method, resourceType
- 5 verification criteria with severity levels (Critical → Medium)
- Cookie clearing between routes for clean state

**Step 4.2 - Navigation Verification**:
- BFS crawl with configurable depth (`--depth` flag, default: 3)
- `NavigationResult` with status: success/broken/redirect/external
- External link detection (mailto:, tel:, javascript:, #, different host)
- Relative URL resolution (/, ./, ../)
- Broken link report with source page and link text

**Step 4.3 - Form Interaction Verification**:
- 5 form types auto-detected: login, signup, search, contact, generic
- Smart test data generation by field name + type fallback
- Submit strategies: button[type="submit"] → Enter key
- Result states: success, error, no-response, validation-shown
- Validation error detection via `[class*="error"], [role="alert"], .invalid-feedback`

**Step 4.4 - API Call Verification**:
- `ApiMonitorConfig` with capture/ignore patterns
- Default captures: `/api/*`, `/graphql`, `*/rest/*`
- Default ignores: analytics, tracking, hotjar, sentry, google-analytics, facebook
- Source vs Target comparison modes: strict (exact status) / relaxed (success class match)
- Missing/Extra API call detection between source and target

**Step 4.5 - JavaScript Error Collection**:
- `CategorizedError` with source, category, severity, stack, count
- 6 categories: runtime, network, framework, third-party, deprecation, security
- 4 severity levels: critical, high, medium, low
- Deduplication via `addOrIncrementError()` (same message + source)
- Summary with totalErrors, severity breakdown, top 5 errors
- Event listeners: pageerror, console.error, requestfailed

**Step 4.6 - Behavioral Comparison (Bonus)**:
- Source vs Target simultaneous page load
- Heading comparison (h1, h2, h3)
- Navigation link comparison (nav a, [role="navigation"] a)
- Content length comparison (main, [role="main"], #content) with 30% tolerance
- Detailed differences array for actionable feedback

---

## Phase 5: Cross-Browser Verification - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md`

| # | Task | Description | Status |
|---|------|-------------|--------|
| 5.1 | Multi-Browser Execution Engine | 3 browsers parallel launch + per-route testing | Done |
| 5.2 | Cross-Browser Consistency Analysis | 4-factor scoring (status/errors/visual/performance) | Done |
| 5.3 | Mobile Device Emulation | 6 devices (iPhone 14 Pro, SE, Pixel 7, Galaxy S23, iPad Pro, iPad Mini) | Done |
| 5.4 | Cross-Browser Report Generation | Desktop + Mobile results table + issue summary | Done |
| 5.5 | Known Issues Filter | Browser-specific acceptable differences registry (bonus) | Done |

### Implementation Details

**Step 5.1 - Multi-Browser Execution Engine**:
- `BrowserConfig` interface with name, displayName, viewport, launchOptions
- 3 desktop browsers: Chromium (Chrome/Edge), Firefox, WebKit (Safari)
- All browsers launched in parallel via `Promise.all()`
- Per-route testing across all browsers simultaneously
- `CrossBrowserResult` with results array + consistencyScore + issues

**Step 5.2 - Cross-Browser Consistency Analysis**:
- 4-factor consistency scoring (starts at 100, deductions per issue):
  - HTTP status difference: -30 (critical)
  - JS error inconsistency: -15 (high)
  - Visual diff > 8%: -10 per pair (medium)
  - Render time deviation > 50%: -5 (low)
- `CrossBrowserIssue` with type: visual/functional/performance/error
- Screenshot comparison between browsers (threshold: 8%)

**Step 5.3 - Mobile Device Emulation (6 devices)**:

| Device | Viewport | Scale | Browser |
|--------|----------|-------|---------|
| iPhone 14 Pro | 393x852 | 3x | WebKit |
| iPhone SE | 375x667 | 2x | WebKit |
| Pixel 7 | 412x915 | 2.625x | Chromium |
| Galaxy S23 | 360x780 | 3x | Chromium |
| iPad Pro 12.9 | 1024x1366 | 2x | WebKit |
| iPad Mini | 768x1024 | 2x | WebKit |

- 5 mobile-specific checks per device:
  1. Viewport meta tag (width=device-width)
  2. No horizontal overflow
  3. Touch target size (44x44px WCAG minimum)
  4. Font readability (12px minimum)
  5. Fixed/sticky element count

**Step 5.4 - Cross-Browser Report**:
- Desktop results: route × browser matrix with consistency score
- Mobile results: route × device matrix with issue counts
- Issue summary with severity and likely cause
- Overall consistency score (0-100)

**Step 5.5 - Known Issues Filter (Bonus)**:
- Per-browser known differences registry (YAML config)
- Severity levels: ignore, low, medium, high
- Examples: WebKit font-smoothing, Firefox focus-ring, date picker rendering
- Filter function removes "ignore" severity from final report

### Execution Mode

| Flag | Desktop Browsers | Mobile Devices |
|------|-----------------|----------------|
| `--full` | All 3 | All 6 |
| `--cross-browser` | All 3 | 2 (iPhone 14, Pixel 7) |
| Default | Chromium only | None |

---

## Phase 6: Performance Comparison - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md`

| # | Task | Description | Status |
|---|------|-------------|--------|
| 6.1 | Core Web Vitals Collection | LCP, CLS, INP, FCP, TTFB via PerformanceObserver injection | Done |
| 6.2 | Navigation Timing Collection | 10-phase breakdown (DNS→PageLoad) + 3-run median | Done |
| 6.3 | Resource & Bundle Size Analysis | JS/CSS/Image/Font/ThirdParty sizes + top 5 largest | Done |
| 6.4 | Performance Budget Validation | 10 absolute + 5 regression checks with configurable thresholds | Done |
| 6.5 | Performance Comparison Report | Per-route table + aggregated summary + recommendations | Done |
| 6.6 | Performance Test Execution Flow | Multi-run median, retry, cache-clear between runs (bonus) | Done |

### Implementation Details

**Step 6.1 - Core Web Vitals**:
- `addInitScript()` injection before navigation for accurate PerformanceObserver setup
- 5 metrics: LCP, CLS, INP (replacing FID), FCP, TTFB
- CLS: `hadRecentInput` filtering for accurate shift calculation
- LCP: 3-second stabilization wait post-networkidle
- INP: Mouse click trigger for interaction measurement

**Step 6.2 - Navigation Timing (10 phases)**:
- DNS Lookup, TCP Connection, TLS Negotiation
- TTFB, Content Download
- DOM Parsing, DOM Content Loaded, DOM Complete
- Page Load, Total Time
- 3-run median averaging with cache clearing between runs

**Step 6.3 - Resource Analysis**:
- `ResourceMetrics` interface: 7 size categories + request counts + top 5 largest
- Dual collection: Network interception + Performance API (more accurate)
- Third-party size isolation via host comparison
- Resource type detection: script, css, img, font, other

**Step 6.4 - Performance Budget**:
- 10 absolute thresholds (LCP, CLS, INP, FCP, TTFB, PageLoad, JS, CSS, Total, Requests)
- 5 regression thresholds (LCP, CLS, Load Time, JS Size, Request Count)
- 3 severity levels: pass (under budget), warn (regression but within budget), fail (over budget)
- Configurable via `verification.performance_budget` in `.migrate-config.yaml`

**Step 6.5 - Performance Report**:
- Per-route table: Source vs Target with % change
- `PerformanceSummary`: improvements, regressions, violations, recommendations
- Auto-generated recommendations based on detected issues
- Overall score: 0-100 (deductions per violation/regression)

**Step 6.6 - Execution Flow (Bonus)**:
- 3-run median for stability (not average)
- Cache clearing between runs (cookies + Cache API)
- Fresh browser instance per run for isolation
- Error handling with retry (failed runs logged, remaining used)

### Performance Budget Configuration

```yaml
verification:
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

    # Regression thresholds (vs source)
    lcp_regression_pct: 20
    cls_regression_pct: 50
    load_time_regression_pct: 25
    js_size_regression_pct: 30
    request_count_regression_pct: 50
```

---

## Phase 7: Accessibility Verification - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md`

| # | Task | Description | Status |
|---|------|-------------|--------|
| 7.1 | axe-core Integration | AxeBuilder + WCAG tag selection + score calculation | Done |
| 7.2 | Regression Comparison | Source vs Target violation diff (new/resolved/persistent) | Done |
| 7.3 | Violation Categorization | 8 categories with priority + effort estimates (bonus) | Done |
| 7.4 | Report Generation | Per-route scores + category breakdown + recommendations | Done |
| 7.5 | Configuration | WCAG level, fail conditions, exclusions, threshold (bonus) | Done |

### Implementation Details

**Step 7.1 - axe-core Integration**:
- Utilizes `@axe-core/playwright` package
- `AccessibilityResult` interface: violations, passes, incomplete, inapplicable, score
- `AxeViolation` with impact, nodes (html + target selector + failureSummary), tags
- WCAG level-based tag selection: A → AA → AAA (cumulative)
- Dynamic content exclusion support

**Step 7.2 - Regression Comparison**:
- Scans both Source and Target, then compares based on violation ID
- 3 classifications: newViolations (target only), resolvedViolations (source only), persistentViolations (both)
- `regressionDetected`: true when critical/serious new violations exist
- Score delta calculation: positive = improvement, negative = regression

**Step 7.3 - Violation Categorization (8 categories)**:

| Category | Priority | Effort | Example Rules |
|----------|----------|--------|---------------|
| Color & Contrast | high | quick | color-contrast |
| Keyboard Navigation | critical | moderate | keyboard, tabindex |
| Images & Media | high | quick | image-alt, video-caption |
| Form Labels | high | quick | label, select-name |
| Document Structure | medium | moderate | landmark-one-main, heading-order |
| ARIA Usage | medium | moderate | aria-valid-attr, aria-roles |
| Links & Buttons | high | quick | link-name, button-name |
| Tables & Lists | medium | moderate | td-headers-attr, list |

**Step 7.4 - Report Generation**:
- Overall summary: avg score, violations, critical count, routes passed
- Per-route table: source score vs target score with status
- New violation details: affected elements + fix suggestions + help URL
- Category breakdown table with priority and effort
- `determineOverallStatus()`: fail (critical/serious new) / warn (score drop) / pass

**Step 7.5 - Configuration (Bonus)**:
```yaml
verification:
  accessibility:
    wcag_level: "AA"
    fail_on_serious: true
    fail_on_regression: true
    score_threshold: 80
    exclude_rules: []
    exclude_selectors: ["[data-testid='dynamic']", ".third-party-widget"]
    include_best_practices: true
```

---

## Phase 8: Agent & Skill Implementation - COMPLETED

**Status**: Completed (2026-01-24)
**File Modified**: `templates/.claude/commands/jikime/migrate-4-verify.md`

| # | Task | Description | Status |
|---|------|-------------|--------|
| 8.1 | Enhanced `e2e-tester` Agent Spec | Migration mode context contract + execution flow | Done |
| 8.2 | Playwright Migration Skill Structure | SKILL.md frontmatter + reference.md + 7 modules | Done |
| 8.3 | Example Code Templates (3) | visual-comparison.ts, route-crawler.ts, cross-browser-verify.ts | Done |
| 8.4 | Module Documentation Pattern | Consistent module template + coverage table | Done |

### Implementation Details

**Step 8.1 - Enhanced e2e-tester Agent**:
- Dual mode support: Development (J.A.R.V.I.S.) + Migration (F.R.I.D.A.Y.)
- `MigrationVerificationContext` input interface: config, verification settings, flags, routes
- `MigrationVerificationResult` output interface: status, summary, categories, failedTests, recommendations
- Migration execution flow: servers → routes → phases (conditional) → report → cleanup
- Mode comparison table: trigger, input, server, focus, output, failure differences

**Step 8.2 - Playwright Migration Skill**:
- Directory: `.claude/skills/jikime-workflow-playwright-migration/`
- SKILL.md: triggers on `migration verify`, `visual regression`, `playwright migration`
- reference.md: Dual-server architecture, threshold selection, multi-viewport strategy, performance/a11y guidelines
- 7 modules covering all 7 verification phases (~2-4K tokens each)
- Progressive disclosure: Level 1 ~120 tokens, Level 2 ~6K tokens

**Step 8.3 - Example Code (3 files)**:
- `visual-comparison.ts`: Full pipeline with pixelmatch, multi-viewport, mask selectors, image normalization
- `route-crawler.ts`: Artifact parsing + BFS crawl + merge/dedup + route registry JSON output
- `cross-browser-verify.ts`: Parallel Chromium/Firefox/WebKit, consistency scoring (4 factors), JSON report

**New Skill Structure (expanded)**:
```
.claude/skills/jikime-workflow-playwright-migration/
├── SKILL.md                    # Frontmatter + triggers + dependency list
├── reference.md                # Best practices (dual-server, thresholds, caching, a11y)
├── modules/
│   ├── server-lifecycle.md     # Phase 1: Framework detection + health check
│   ├── route-discovery.md      # Phase 2: Artifact parse + BFS crawl
│   ├── visual-regression.md    # Phase 3: pixelmatch + threshold + viewports
│   ├── behavioral-testing.md   # Phase 4: Page load + nav + forms + API + errors
│   ├── cross-browser.md        # Phase 5: Parallel browsers + mobile emulation
│   ├── performance.md          # Phase 6: Web Vitals + timing + budgets
│   └── accessibility.md        # Phase 7: axe-core + regression + categorization
└── examples/
    ├── visual-comparison.ts    # Complete visual regression pipeline
    ├── route-crawler.ts        # Route auto-discovery with artifact merge
    └── cross-browser-verify.ts # Multi-browser parallel verification
```

---

## Phase 9: Command & Configuration Updates (COMPLETED)

| # | Step | Description | Status |
|---|------|-------------|--------|
| 9.1 | Unified Execution Engine | `VerificationPipeline` interface with dependency graph, `executeVerification()` master flow | Done |
| 9.2 | Flag Validation & Conflict Resolution | 7 validation rules, auto-correction, `validateFlags()` algorithm | Done |
| 9.3 | Tool & Dependency Requirements | 4 required packages, pre-flight check script, version matrix | Done |
| 9.4 | CI/CD Configuration | GitHub Actions workflow + GitLab CI pipeline templates (headless) | Done |
| 9.5 | Error Handling & Graceful Degradation | 3 severity levels, 14 error types, 4 recovery actions | Done |
| 9.6 | Extended Configuration Schema | Complete `.migrate-config.yaml` with all 7 phase settings + pipeline config | Done |

### Key Deliverables

**Unified Execution Engine** (`VerificationPipeline` interface):
- Dependency graph: route-discovery → page-load → visual/cross-browser/a11y/perf (parallel) → report
- Master `executeVerification()` flow with pre-flight checks
- Phase result aggregation into unified `PipelineResult`

**Flag Validation** (7 rules):
- `--full` expands to all verification types
- `--threshold` requires `--visual`
- `--depth` requires route-discovery phase enabled
- Port conflict detection and auto-correction

**CI/CD Templates**:
- GitHub Actions: `migration-verify.yml` with Playwright container, artifact upload, PR comment
- GitLab CI: `.gitlab-ci.yml` with stages (setup/verify/report), Docker-in-Docker runner

**Error Handling** (3 severity levels):
- Fatal: dependency missing, port conflict, build failure → abort pipeline
- Degraded: browser unavailable, timeout → skip phase, continue
- Warning: threshold exceeded, flaky test → log, include in report

**Configuration Schema** (`.migrate-config.yaml`):
- 7 phase sections: route_discovery, page_load, visual_regression, cross_browser, accessibility, performance, state_integrity
- Pipeline settings: parallel_phases, timeout, retry, artifacts directory
- CI integration: headless mode, reporter format, failure threshold

---

## Phase 10: Reports & Artifacts (COMPLETED)

| # | Step | Description | Status |
|---|------|-------------|--------|
| 10.1 | Unified Verification Report | `VerificationReport` data model, markdown template with 7 phase sections, verdict algorithm | Done |
| 10.2 | HTML Visual Report | Interactive comparison viewer (3-panel/side-by-side/overlay/slider), filter controls, diff generation | Done |
| 10.3 | CI/CD Artifact Integration | GitHub Actions artifact upload + PR comment, GitLab CI artifacts + JUnit + Pages deploy | Done |
| 10.4 | Report Generation Engine | `generateReports()` algorithm, summary calculation, recommendation engine, console output | Done |

### Key Deliverables

**Unified Report** (`VerificationReport` interface):
- Full data model: metadata, environment, 7 phase results, summary, recommendations
- Markdown template with Handlebars-style placeholders for all sections
- `determineVerdict()`: 4-rule algorithm (critical → pass_rate → high_severity → warnings)
- `collectBlockingFailures()`: visual/performance/accessibility/page_load checks

**HTML Visual Report** (interactive browser viewer):
- 4 view modes: 3-panel (source/target/diff), side-by-side, overlay, slider
- Filter controls: status, viewport, route search
- Dark theme with responsive layout (mobile/tablet/desktop)
- `generateDiffImage()`: pixelmatch + flood-fill region detection
- Per-comparison metrics: pixel count, diff percentage, threshold, regions

**CI/CD Artifacts**:
- Directory structure: `screenshots/{source,target,diff}/`, `traces/`, `a11y/`, reports
- GitHub Actions: `upload-artifact@v4`, PR comment with summary table, conditional full upload on failure
- GitLab CI: JUnit report generation for test visualization, GitLab Pages deployment for visual report
- Retention: 30 days for reports, 14 days for screenshots/traces

**Report Generation Engine** (`generateReports()`):
- 7-step pipeline: collect → calculate → verdict → recommendations → compose → render → write
- Parallel format generation: markdown + HTML + JSON simultaneously
- `generateRecommendations()`: priority-sorted advice for visual/perf/a11y/cross-browser issues
- Console output: box-formatted summary with phase status indicators

---

## Execution Order (Recommended)

```
Phase 1 (Infrastructure) → Phase 2 (Route Discovery) → Phase 4 (Behavioral)
    ↓
Phase 3 (Visual Regression) → Phase 5 (Cross-Browser)
    ↓
Phase 6 (Performance) → Phase 7 (Accessibility)
    ↓
Phase 8 (Agent/Skill) → Phase 9 (Command) → Phase 10 (Reports)
```

## Execution Architecture

### 3-Layer System

This system operates via **AI orchestration**. `migrate-4-verify.md` is not a standalone executable script but a Claude Code slash command (AI instruction document).

```
┌─────────────────────────────────────────────────────────┐
│ Layer 1: Command (migrate-4-verify.md)                  │
│ - Instructs AI on "what" and "how" to perform           │
│ - Defines workflows, interfaces, and algorithms for     │
│   10 phases                                             │
│ - Slash command: /jikime:migrate-4-verify               │
└──────────────────────┬──────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────┐
│ Layer 2: Agent (e2e-tester)                             │
│ - Specialized sub-agent delegated by F.R.I.D.A.Y.       │
│ - Interprets command instructions to generate           │
│   Playwright code                                       │
│ - Executes actual commands via Bash tool                │
└──────────────────────┬──────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────┐
│ Layer 3: Playwright (Actual Browser Automation)         │
│ - Browser execution, screenshots, performance           │
│   measurement, accessibility testing                    │
│ - Requires npm package installation in project          │
│ - Runs via npx playwright test, etc.                    │
└─────────────────────────────────────────────────────────┘
```

### Execution Flow

```
User: /jikime:migrate-4-verify --visual --cross-browser
     ↓
F.R.I.D.A.Y. (Orchestrator):
  1. Read command instructions
  2. Validate flags (Step 9.2)
  3. Pre-flight dependency check (Step 9.3)
     ↓
e2e-tester Agent:
  4. Load .migrate-config.yaml
  5. Start dev servers (Phase 1)
  6. Route discovery (Phase 2)
  7. Generate & execute Playwright test code (Phase 3-7)
  8. Collect results
     ↓
F.R.I.D.A.Y.:
  9. Generate reports (Phase 10)
  10. Report results to user
```

### Required Packages

| Package | Version | Required | Purpose | Installation Command |
|---------|---------|----------|---------|---------------------|
| `@playwright/test` | ^1.40.0 | Yes | Full browser automation | `npm install -D @playwright/test` |
| `pixelmatch` | ^5.3.0 | No | Visual Regression (Phase 3) | `npm install -D pixelmatch` |
| `pngjs` | ^7.0.0 | No | Screenshot PNG processing (Phase 3) | `npm install -D pngjs` |
| `@axe-core/playwright` | ^4.8.0 | No | Accessibility testing (Phase 7) | `npm install -D @axe-core/playwright` |

**Full Installation (all at once):**

```bash
# Install all required + optional packages
npm install -D @playwright/test pixelmatch pngjs @axe-core/playwright

# Install Playwright browser binaries
npx playwright install

# Install all browsers for cross-browser testing
npx playwright install chromium firefox webkit
```

### Pre-flight Check

The agent automatically checks dependencies before execution:

| Situation | Behavior |
|-----------|----------|
| Required package not installed (`@playwright/test`) | Output error + show installation command + abort pipeline |
| Optional package not installed (`pixelmatch`, etc.) | Output warning + skip corresponding phase |
| Browser not installed | Attempt auto-installation (`npx playwright install`) |

### Prerequisites

Items to verify before command execution:

- [ ] Node.js 18+ installed
- [ ] `@playwright/test` installed in project
- [ ] Playwright browser binaries installed (`npx playwright install`)
- [ ] Source project dev server runnable (`npm run dev`, etc.)
- [ ] Target project dev server runnable
- [ ] `.migrate-config.yaml` file exists (generated by migrate-2-plan)
- [ ] Ports 3000, 3001 available (or change with `--port`, `--source-port`)

---

## Deliverables Summary

| Type | Count | Content |
|------|-------|---------|
| Command Modified | 1 | `migrate-4-verify.md` complete overhaul (10 phases) |
| Agent Modified | 1 | `e2e-tester.md` dual-mode enhancement |
| New Skill | 1 | `jikime-workflow-playwright-migration/` (7 modules + 3 examples) |
| Example Code | 3 | Visual Comparison, Route Crawler, Cross-Browser Verify |
| Config Schema | 1 | `.migrate-config.yaml` full verification schema |
| Report Templates | 3 | Markdown, HTML Visual, JSON machine-readable |
| CI/CD Templates | 2 | GitHub Actions workflow, GitLab CI pipeline |
| TypeScript Interfaces | 15+ | Pipeline, Report, Flags, Config, Phase results |

## Total Steps: 44 (across 10 phases)

---

Version: 2.0.0
Created: 2026-01-24
Last Updated: 2026-01-24
Status: Complete (All 10 Phases Implemented)
