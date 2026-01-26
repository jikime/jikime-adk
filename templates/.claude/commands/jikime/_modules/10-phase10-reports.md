---

## Phase 10: Reports & Artifacts

### Step 10.1: Unified Verification Report

#### Report Data Model

```typescript
interface VerificationReport {
  // Metadata
  metadata: {
    generated_at: string          // ISO 8601 timestamp
    duration_ms: number           // Total pipeline execution time
    pipeline_version: string      // migrate-4-verify version
    config_hash: string           // SHA256 of .migrate-config.yaml
  }

  // Environment
  environment: {
    source: {
      framework: string           // e.g., "react@18.2.0"
      port: number
      dev_command: string
      node_version: string
    }
    target: {
      framework: string           // e.g., "nextjs@14.0.0"
      port: number
      dev_command: string
      node_version: string
    }
    browsers: string[]            // ["chromium@120", "firefox@119", "webkit@17"]
    os: string                    // "darwin-arm64" | "linux-x64"
  }

  // Phase Results
  phases: {
    route_discovery: RouteDiscoveryResult
    page_load: PageLoadResult
    visual_regression: VisualRegressionResult
    cross_browser: CrossBrowserResult
    accessibility: AccessibilityResult
    performance: PerformanceResult
    state_integrity: StateIntegrityResult
  }

  // Aggregated Summary
  summary: {
    total_checks: number
    passed: number
    failed: number
    warnings: number
    skipped: number
    pass_rate: number             // percentage (0-100)
    verdict: 'PASSED' | 'FAILED' | 'PASSED_WITH_WARNINGS'
    blocking_failures: FailureDetail[]
  }

  // Recommendations
  recommendations: Recommendation[]
}

interface FailureDetail {
  phase: string
  route: string
  check_type: string
  expected: string | number
  actual: string | number
  severity: 'critical' | 'high' | 'medium' | 'low'
  suggestion: string
}

interface Recommendation {
  priority: 'must_fix' | 'should_fix' | 'consider'
  category: string
  message: string
  affected_routes: string[]
  auto_fixable: boolean
}
```

#### Markdown Report Template

```markdown
# Migration Verification Report

> Generated: {metadata.generated_at}
> Duration: {metadata.duration_ms}ms
> Pipeline: v{metadata.pipeline_version}

---

## Environment

| Property | Source | Target |
|----------|--------|--------|
| Framework | {source.framework} | {target.framework} |
| Port | :{source.port} | :{target.port} |
| Node.js | {source.node_version} | {target.node_version} |
| Dev Command | `{source.dev_command}` | `{target.dev_command}` |

**Browsers**: {browsers.join(', ')}
**OS**: {os}

---

## Summary

| Verdict | {summary.verdict} |
|---------|-----|
| Total Checks | {summary.total_checks} |
| Passed | {summary.passed} |
| Failed | {summary.failed} |
| Warnings | {summary.warnings} |
| Skipped | {summary.skipped} |
| **Pass Rate** | **{summary.pass_rate}%** |

---

## Phase Results

### 1. Route Discovery
| Metric | Value |
|--------|-------|
| Routes Found | {route_discovery.total_routes} |
| Source-only | {route_discovery.source_only} |
| Target-only | {route_discovery.target_only} |
| Matched | {route_discovery.matched} |

### 2. Page Load
| Route | Status | Load Time | Errors |
|-------|--------|-----------|--------|
{{#each page_load.results}}
| {route} | {status_icon} | {load_time}ms | {error_count} |
{{/each}}

### 3. Visual Regression
| Route | Viewport | Diff % | Threshold | Status |
|-------|----------|--------|-----------|--------|
{{#each visual_regression.results}}
| {route} | {viewport} | {diff_pct}% | {threshold}% | {status_icon} |
{{/each}}

**Screenshots**: See [HTML Visual Report](./visual-report.html)

### 4. Cross-Browser
| Route | Chromium | Firefox | WebKit |
|-------|----------|---------|--------|
{{#each cross_browser.results}}
| {route} | {chromium_icon} | {firefox_icon} | {webkit_icon} |
{{/each}}

### 5. Accessibility
| Route | Violations | Impact | WCAG Level |
|-------|------------|--------|------------|
{{#each accessibility.results}}
| {route} | {violation_count} | {max_impact} | {wcag_level} |
{{/each}}

### 6. Performance
| Route | LCP | FID | CLS | Score |
|-------|-----|-----|-----|-------|
{{#each performance.results}}
| {route} | {lcp}ms | {fid}ms | {cls} | {score}/100 |
{{/each}}

**Regression Budget**: LCP +{perf_budget.lcp_regression_pct}%, Load {perf_budget.page_load_max_ms}ms max

### 7. State Integrity
| Test | Source | Target | Match |
|------|--------|--------|-------|
{{#each state_integrity.results}}
| {test_name} | {source_value} | {target_value} | {match_icon} |
{{/each}}

---

## Blocking Failures

{{#if summary.blocking_failures.length}}
| # | Phase | Route | Issue | Severity |
|---|-------|-------|-------|----------|
{{#each summary.blocking_failures}}
| {index} | {phase} | {route} | {check_type}: expected {expected}, got {actual} | {severity} |
{{/each}}
{{else}}
No blocking failures detected.
{{/if}}

---

## Recommendations

{{#each recommendations}}
### [{priority}] {category}
{message}
- Affected: {affected_routes.join(', ')}
- Auto-fixable: {auto_fixable ? 'Yes' : 'No'}

{{/each}}

---

## Artifacts

| Artifact | Path | Size |
|----------|------|------|
| Screenshots (source) | `./artifacts/screenshots/source/` | {source_size} |
| Screenshots (target) | `./artifacts/screenshots/target/` | {target_size} |
| Diff images | `./artifacts/screenshots/diff/` | {diff_size} |
| HTML Visual Report | `./artifacts/visual-report.html` | {html_size} |
| Performance traces | `./artifacts/traces/` | {traces_size} |
| Accessibility reports | `./artifacts/a11y/` | {a11y_size} |
| Raw JSON | `./artifacts/report.json` | {json_size} |

---

*Generated by JikiME-ADK Migration Verification Pipeline v{pipeline_version}*
```

#### Verdict Determination Algorithm

```typescript
function determineVerdict(phases: PhaseResults): Verdict {
  const blocking = collectBlockingFailures(phases)

  // Rule 1: Any critical failure = FAILED
  if (blocking.some(f => f.severity === 'critical')) {
    return 'FAILED'
  }

  // Rule 2: Pass rate below threshold = FAILED
  const passRate = calculatePassRate(phases)
  if (passRate < config.verification.pass_threshold) { // default: 90%
    return 'FAILED'
  }

  // Rule 3: High-severity failures present = PASSED_WITH_WARNINGS
  if (blocking.some(f => f.severity === 'high')) {
    return 'PASSED_WITH_WARNINGS'
  }

  // Rule 4: Warnings exist = PASSED_WITH_WARNINGS
  const warningCount = countWarnings(phases)
  if (warningCount > 0) {
    return 'PASSED_WITH_WARNINGS'
  }

  return 'PASSED'
}

function collectBlockingFailures(phases: PhaseResults): FailureDetail[] {
  const failures: FailureDetail[] = []

  // Visual regression: diff exceeds threshold
  for (const result of phases.visual_regression.results) {
    if (result.diff_pct > result.threshold) {
      failures.push({
        phase: 'visual_regression',
        route: result.route,
        check_type: 'pixel_diff',
        expected: `<= ${result.threshold}%`,
        actual: `${result.diff_pct}%`,
        severity: result.diff_pct > result.threshold * 2 ? 'critical' : 'high',
        suggestion: `Review visual changes at ${result.route}. Consider updating mask_selectors for dynamic content.`
      })
    }
  }

  // Performance: regression exceeds budget
  for (const result of phases.performance.results) {
    if (result.lcp_regression_pct > config.performance_budget.lcp_regression_pct) {
      failures.push({
        phase: 'performance',
        route: result.route,
        check_type: 'lcp_regression',
        expected: `<= +${config.performance_budget.lcp_regression_pct}%`,
        actual: `+${result.lcp_regression_pct}%`,
        severity: result.lcp_regression_pct > 50 ? 'critical' : 'high',
        suggestion: `LCP degraded at ${result.route}. Check for render-blocking resources or large component trees.`
      })
    }
  }

  // Accessibility: critical/serious violations
  for (const result of phases.accessibility.results) {
    const critical = result.violations.filter(v => v.impact === 'critical')
    if (critical.length > 0) {
      failures.push({
        phase: 'accessibility',
        route: result.route,
        check_type: 'wcag_violation',
        expected: '0 critical violations',
        actual: `${critical.length} critical violations`,
        severity: 'critical',
        suggestion: `Fix: ${critical.map(v => v.id).join(', ')} at ${result.route}`
      })
    }
  }

  // Page load: HTTP errors or JS exceptions
  for (const result of phases.page_load.results) {
    if (result.status >= 400 || result.js_errors.length > 0) {
      failures.push({
        phase: 'page_load',
        route: result.route,
        check_type: result.status >= 400 ? 'http_error' : 'js_exception',
        expected: 'HTTP 200, no JS errors',
        actual: `HTTP ${result.status}, ${result.js_errors.length} JS errors`,
        severity: 'critical',
        suggestion: `Page failed to load correctly at ${result.route}. Check server logs and browser console.`
      })
    }
  }

  return failures
}
```

### Step 10.2: HTML Visual Report

#### Report Structure

```typescript
interface HTMLVisualReport {
  // Configuration
  title: string                         // "Migration Visual Comparison"
  generated_at: string
  source_framework: string
  target_framework: string

  // Comparison data per route
  comparisons: VisualComparison[]

  // Summary statistics
  stats: {
    total_comparisons: number
    passed: number
    failed: number
    pass_rate: number
  }
}

interface VisualComparison {
  route: string
  viewport: { width: number; height: number; label: string }
  source_screenshot: string             // Base64 or relative path
  target_screenshot: string
  diff_image: string                    // Generated diff overlay
  diff_percentage: number
  threshold: number
  status: 'pass' | 'fail'
  pixel_count: { total: number; changed: number }
  highlighted_regions: BoundingBox[]    // Areas of significant change
}

interface BoundingBox {
  x: number
  y: number
  width: number
  height: number
  change_intensity: number              // 0-1, how different this region is
}
```

#### HTML Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Migration Visual Comparison Report</title>
  <style>
    :root {
      --pass: #22c55e;
      --fail: #ef4444;
      --warn: #f59e0b;
      --bg: #0f172a;
      --surface: #1e293b;
      --text: #f8fafc;
      --muted: #94a3b8;
    }

    * { box-sizing: border-box; margin: 0; padding: 0; }

    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background: var(--bg);
      color: var(--text);
      padding: 2rem;
    }

    .header {
      text-align: center;
      margin-bottom: 2rem;
      padding: 1.5rem;
      background: var(--surface);
      border-radius: 12px;
    }

    .header h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
    .header .meta { color: var(--muted); font-size: 0.875rem; }

    .stats {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 1rem;
      margin-bottom: 2rem;
    }

    .stat-card {
      background: var(--surface);
      padding: 1rem;
      border-radius: 8px;
      text-align: center;
    }

    .stat-card .value { font-size: 2rem; font-weight: bold; }
    .stat-card .label { color: var(--muted); font-size: 0.75rem; }
    .stat-card.pass .value { color: var(--pass); }
    .stat-card.fail .value { color: var(--fail); }

    .comparison {
      background: var(--surface);
      border-radius: 12px;
      margin-bottom: 1.5rem;
      overflow: hidden;
    }

    .comparison-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1rem 1.5rem;
      border-bottom: 1px solid rgba(255,255,255,0.1);
    }

    .comparison-header .route { font-weight: 600; }
    .comparison-header .badge {
      padding: 0.25rem 0.75rem;
      border-radius: 999px;
      font-size: 0.75rem;
      font-weight: 600;
    }
    .badge.pass { background: rgba(34,197,94,0.2); color: var(--pass); }
    .badge.fail { background: rgba(239,68,68,0.2); color: var(--fail); }

    .comparison-body { padding: 1.5rem; }

    .view-controls {
      display: flex;
      gap: 0.5rem;
      margin-bottom: 1rem;
    }

    .view-btn {
      padding: 0.5rem 1rem;
      border: 1px solid rgba(255,255,255,0.2);
      border-radius: 6px;
      background: transparent;
      color: var(--text);
      cursor: pointer;
      font-size: 0.875rem;
    }
    .view-btn.active {
      background: rgba(99,102,241,0.3);
      border-color: #6366f1;
    }

    .screenshots {
      display: grid;
      grid-template-columns: 1fr 1fr 1fr;
      gap: 1rem;
    }
    .screenshots.side-by-side { grid-template-columns: 1fr 1fr; }
    .screenshots.overlay { grid-template-columns: 1fr; }

    .screenshot-panel { position: relative; }
    .screenshot-panel img {
      width: 100%;
      border-radius: 8px;
      border: 1px solid rgba(255,255,255,0.1);
    }
    .screenshot-panel .label {
      position: absolute;
      top: 0.5rem;
      left: 0.5rem;
      background: rgba(0,0,0,0.7);
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      font-size: 0.75rem;
    }

    .diff-info {
      display: flex;
      gap: 2rem;
      padding: 1rem;
      background: rgba(0,0,0,0.3);
      border-radius: 8px;
      margin-top: 1rem;
      font-size: 0.875rem;
    }
    .diff-info .item { display: flex; align-items: center; gap: 0.5rem; }

    .slider-container {
      position: relative;
      overflow: hidden;
      border-radius: 8px;
    }
    .slider-container img { width: 100%; display: block; }
    .slider-overlay {
      position: absolute;
      top: 0;
      left: 0;
      height: 100%;
      overflow: hidden;
      border-right: 2px solid #6366f1;
    }
    .slider-handle {
      position: absolute;
      top: 50%;
      right: -12px;
      width: 24px;
      height: 24px;
      background: #6366f1;
      border-radius: 50%;
      transform: translateY(-50%);
      cursor: ew-resize;
    }

    .filter-bar {
      display: flex;
      gap: 1rem;
      margin-bottom: 1.5rem;
      align-items: center;
    }
    .filter-bar select, .filter-bar input {
      padding: 0.5rem;
      border-radius: 6px;
      border: 1px solid rgba(255,255,255,0.2);
      background: var(--surface);
      color: var(--text);
    }

    @media (max-width: 768px) {
      .stats { grid-template-columns: repeat(2, 1fr); }
      .screenshots { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="header">
    <h1>Migration Visual Comparison Report</h1>
    <div class="meta">
      <span>{source_framework} → {target_framework}</span> |
      <span>Generated: {generated_at}</span> |
      <span>{stats.total_comparisons} comparisons</span>
    </div>
  </div>

  <div class="stats">
    <div class="stat-card">
      <div class="value">{stats.total_comparisons}</div>
      <div class="label">Total Comparisons</div>
    </div>
    <div class="stat-card pass">
      <div class="value">{stats.passed}</div>
      <div class="label">Passed</div>
    </div>
    <div class="stat-card fail">
      <div class="value">{stats.failed}</div>
      <div class="label">Failed</div>
    </div>
    <div class="stat-card">
      <div class="value">{stats.pass_rate}%</div>
      <div class="label">Pass Rate</div>
    </div>
  </div>

  <div class="filter-bar">
    <select id="status-filter">
      <option value="all">All Status</option>
      <option value="fail">Failed Only</option>
      <option value="pass">Passed Only</option>
    </select>
    <select id="viewport-filter">
      <option value="all">All Viewports</option>
      <option value="desktop">Desktop (1920x1080)</option>
      <option value="tablet">Tablet (768x1024)</option>
      <option value="mobile">Mobile (375x667)</option>
    </select>
    <input type="search" placeholder="Filter by route..." id="route-filter">
  </div>

  <div id="comparisons">
    <!-- Dynamically generated comparison cards -->
  </div>

  <script>
    // Report data (injected during generation)
    const REPORT_DATA = {comparisons_json};

    // View mode state
    let currentView = 'three-panel'; // 'three-panel' | 'side-by-side' | 'overlay' | 'slider'

    function renderComparisons(data) {
      const container = document.getElementById('comparisons');
      container.innerHTML = data.map(comp => `
        <div class="comparison" data-status="${comp.status}" data-viewport="${comp.viewport.label}" data-route="${comp.route}">
          <div class="comparison-header">
            <div>
              <span class="route">${comp.route}</span>
              <span style="color: var(--muted); margin-left: 0.5rem;">${comp.viewport.label} (${comp.viewport.width}x${comp.viewport.height})</span>
            </div>
            <span class="badge ${comp.status}">${comp.status.toUpperCase()} (${comp.diff_percentage.toFixed(2)}%)</span>
          </div>
          <div class="comparison-body">
            <div class="view-controls">
              <button class="view-btn ${currentView === 'three-panel' ? 'active' : ''}" onclick="setView('three-panel', this)">3-Panel</button>
              <button class="view-btn ${currentView === 'side-by-side' ? 'active' : ''}" onclick="setView('side-by-side', this)">Side by Side</button>
              <button class="view-btn ${currentView === 'overlay' ? 'active' : ''}" onclick="setView('overlay', this)">Overlay</button>
              <button class="view-btn ${currentView === 'slider' ? 'active' : ''}" onclick="setView('slider', this)">Slider</button>
            </div>
            <div class="screenshots">
              <div class="screenshot-panel">
                <span class="label">Source</span>
                <img src="${comp.source_screenshot}" alt="Source: ${comp.route}">
              </div>
              <div class="screenshot-panel">
                <span class="label">Target</span>
                <img src="${comp.target_screenshot}" alt="Target: ${comp.route}">
              </div>
              <div class="screenshot-panel">
                <span class="label">Diff</span>
                <img src="${comp.diff_image}" alt="Diff: ${comp.route}">
              </div>
            </div>
            <div class="diff-info">
              <div class="item">Pixels Changed: <strong>${comp.pixel_count.changed.toLocaleString()} / ${comp.pixel_count.total.toLocaleString()}</strong></div>
              <div class="item">Diff: <strong>${comp.diff_percentage.toFixed(2)}%</strong></div>
              <div class="item">Threshold: <strong>${comp.threshold}%</strong></div>
            </div>
          </div>
        </div>
      `).join('');
    }

    function setView(view, btn) {
      currentView = view;
      renderComparisons(getFilteredData());
    }

    function getFilteredData() {
      const status = document.getElementById('status-filter').value;
      const viewport = document.getElementById('viewport-filter').value;
      const route = document.getElementById('route-filter').value.toLowerCase();

      return REPORT_DATA.filter(comp => {
        if (status !== 'all' && comp.status !== status) return false;
        if (viewport !== 'all' && comp.viewport.label.toLowerCase() !== viewport) return false;
        if (route && !comp.route.toLowerCase().includes(route)) return false;
        return true;
      });
    }

    // Event listeners for filters
    document.getElementById('status-filter').addEventListener('change', () => renderComparisons(getFilteredData()));
    document.getElementById('viewport-filter').addEventListener('change', () => renderComparisons(getFilteredData()));
    document.getElementById('route-filter').addEventListener('input', () => renderComparisons(getFilteredData()));

    // Initial render
    renderComparisons(REPORT_DATA);
  </script>
</body>
</html>
```

#### Diff Image Generation Algorithm

```typescript
async function generateDiffImage(
  sourcePath: string,
  targetPath: string,
  outputPath: string
): Promise<DiffResult> {
  // Use pixelmatch for pixel-level comparison
  const sourceImg = PNG.sync.read(fs.readFileSync(sourcePath))
  const targetImg = PNG.sync.read(fs.readFileSync(targetPath))

  const { width, height } = sourceImg
  const diff = new PNG({ width, height })

  const changedPixels = pixelmatch(
    sourceImg.data,
    targetImg.data,
    diff.data,
    width,
    height,
    {
      threshold: 0.1,           // Per-pixel sensitivity
      alpha: 0.5,               // Diff overlay opacity
      diffColor: [255, 0, 128], // Changed pixel color (magenta)
      diffColorAlt: [0, 200, 255], // Anti-aliased pixel color (cyan)
      aaColor: [128, 128, 128]  // Anti-aliased area color
    }
  )

  fs.writeFileSync(outputPath, PNG.sync.write(diff))

  const totalPixels = width * height
  const diffPercentage = (changedPixels / totalPixels) * 100

  // Detect changed regions using flood-fill clustering
  const regions = detectChangedRegions(diff.data, width, height)

  return {
    changed_pixels: changedPixels,
    total_pixels: totalPixels,
    diff_percentage: diffPercentage,
    highlighted_regions: regions,
    diff_path: outputPath
  }
}

function detectChangedRegions(
  diffData: Uint8Array,
  width: number,
  height: number
): BoundingBox[] {
  const visited = new Set<number>()
  const regions: BoundingBox[] = []
  const CLUSTER_THRESHOLD = 10  // Minimum pixels for a region

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const idx = (y * width + x) * 4
      // Check if pixel is marked as different (non-black in diff image)
      if (diffData[idx] > 0 && !visited.has(y * width + x)) {
        const region = floodFill(diffData, width, height, x, y, visited)
        if (region.pixelCount >= CLUSTER_THRESHOLD) {
          regions.push(region.bounds)
        }
      }
    }
  }

  return mergeOverlappingRegions(regions)
}
```

### Step 10.3: CI/CD Artifact Integration

#### Artifact Directory Structure

```
.migration-verification/
├── artifacts/
│   ├── screenshots/
│   │   ├── source/
│   │   │   ├── home-desktop-1920x1080.png
│   │   │   ├── home-tablet-768x1024.png
│   │   │   ├── home-mobile-375x667.png
│   │   │   ├── dashboard-desktop-1920x1080.png
│   │   │   └── ...
│   │   ├── target/
│   │   │   └── ... (same structure as source)
│   │   └── diff/
│   │       └── ... (generated diff images)
│   ├── traces/
│   │   ├── home-performance.json
│   │   ├── dashboard-performance.json
│   │   └── ...
│   ├── a11y/
│   │   ├── home-axe-results.json
│   │   ├── dashboard-axe-results.json
│   │   └── ...
│   ├── report.md                       # Unified markdown report
│   ├── report.json                     # Machine-readable JSON
│   └── visual-report.html             # Interactive HTML viewer
├── .gitignore                          # Ignore artifacts in VCS
└── config.yaml                         # Verification configuration snapshot
```

#### GitHub Actions Artifact Configuration

```yaml
# .github/workflows/migration-verify.yml (artifact section)
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      # ... (setup and verification steps from Phase 9.4)

      - name: Upload Verification Artifacts
        if: always()  # Upload even on failure
        uses: actions/upload-artifact@v4
        with:
          name: migration-verification-${{ github.sha }}
          path: |
            .migration-verification/artifacts/report.md
            .migration-verification/artifacts/report.json
            .migration-verification/artifacts/visual-report.html
            .migration-verification/artifacts/screenshots/diff/
          retention-days: 30

      - name: Upload Full Screenshots (on failure)
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: migration-screenshots-full-${{ github.sha }}
          path: .migration-verification/artifacts/screenshots/
          retention-days: 14

      - name: Upload Performance Traces
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: migration-perf-traces-${{ github.sha }}
          path: .migration-verification/artifacts/traces/
          retention-days: 14

      - name: Comment PR with Results
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const report = fs.readFileSync('.migration-verification/artifacts/report.json', 'utf8');
            const data = JSON.parse(report);

            const verdict = data.summary.verdict;
            const icon = verdict === 'PASSED' ? ':white_check_mark:' :
                         verdict === 'PASSED_WITH_WARNINGS' ? ':warning:' : ':x:';

            const body = `## ${icon} Migration Verification: ${verdict}

            | Metric | Value |
            |--------|-------|
            | Pass Rate | ${data.summary.pass_rate}% |
            | Total Checks | ${data.summary.total_checks} |
            | Passed | ${data.summary.passed} |
            | Failed | ${data.summary.failed} |
            | Duration | ${(data.metadata.duration_ms / 1000).toFixed(1)}s |

            ${data.summary.blocking_failures.length > 0 ? `
            ### Blocking Failures
            ${data.summary.blocking_failures.map(f =>
              \`- **\${f.phase}** \${f.route}: \${f.check_type} (expected \${f.expected}, got \${f.actual})\`
            ).join('\\n')}
            ` : ''}

            <details>
            <summary>View Full Report</summary>

            Download artifacts for detailed visual comparison and performance traces.
            </details>

            ---
            *Generated by JikiME-ADK Migration Verification Pipeline*`;

            await github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body
            });
```

#### GitLab CI Artifact Configuration

```yaml
# .gitlab-ci.yml (artifact section)
migration-verify:
  stage: verify
  artifacts:
    when: always
    paths:
      - .migration-verification/artifacts/report.md
      - .migration-verification/artifacts/report.json
      - .migration-verification/artifacts/visual-report.html
      - .migration-verification/artifacts/screenshots/diff/
    reports:
      junit: .migration-verification/artifacts/junit-report.xml
    expire_in: 30 days

  after_script:
    # Generate JUnit-compatible report for GitLab test visualization
    - |
      node -e "
        const report = require('./.migration-verification/artifacts/report.json');
        const builder = require('junit-report-builder');
        const suite = builder.testSuite().name('Migration Verification');

        for (const [phase, results] of Object.entries(report.phases)) {
          if (results.results) {
            for (const result of results.results) {
              const tc = suite.testCase()
                .className(phase)
                .name(result.route || result.test_name);
              if (result.status === 'fail') {
                tc.failure(JSON.stringify(result));
              }
            }
          }
        }

        builder.writeTo('.migration-verification/artifacts/junit-report.xml');
      "

migration-verify-pages:
  stage: deploy
  needs: [migration-verify]
  rules:
    - if: $CI_MERGE_REQUEST_ID
  script:
    - mkdir -p public
    - cp .migration-verification/artifacts/visual-report.html public/index.html
    - cp -r .migration-verification/artifacts/screenshots public/screenshots
  artifacts:
    paths:
      - public
  environment:
    name: review/migration-$CI_MERGE_REQUEST_IID
    url: $CI_ENVIRONMENT_URL
    auto_stop_in: 7 days
```

### Step 10.4: Report Generation Engine

#### Generation Algorithm

```typescript
async function generateReports(
  pipelineResult: PipelineResult,
  config: MigrateConfig
): Promise<GeneratedArtifacts> {
  const artifacts: GeneratedArtifacts = {
    markdown: '',
    json: {},
    html: '',
    paths: {}
  }

  // Step 1: Collect all phase results
  const phaseResults = aggregatePhaseResults(pipelineResult)

  // Step 2: Calculate summary statistics
  const summary = calculateSummary(phaseResults)

  // Step 3: Determine verdict
  summary.verdict = determineVerdict(phaseResults)
  summary.blocking_failures = collectBlockingFailures(phaseResults)

  // Step 4: Generate recommendations
  const recommendations = generateRecommendations(phaseResults, config)

  // Step 5: Compose report data
  const reportData: VerificationReport = {
    metadata: {
      generated_at: new Date().toISOString(),
      duration_ms: pipelineResult.duration_ms,
      pipeline_version: PIPELINE_VERSION,
      config_hash: hashConfig(config)
    },
    environment: extractEnvironment(config, pipelineResult),
    phases: phaseResults,
    summary,
    recommendations
  }

  // Step 6: Generate all formats in parallel
  const [markdown, html, json] = await Promise.all([
    renderMarkdownReport(reportData),
    renderHTMLVisualReport(reportData),
    JSON.stringify(reportData, null, 2)
  ])

  // Step 7: Write artifacts to disk
  const artifactsDir = path.join(config.artifacts_dir || '.migration-verification/artifacts')
  await fs.promises.mkdir(artifactsDir, { recursive: true })

  await Promise.all([
    fs.promises.writeFile(path.join(artifactsDir, 'report.md'), markdown),
    fs.promises.writeFile(path.join(artifactsDir, 'report.json'), json),
    fs.promises.writeFile(path.join(artifactsDir, 'visual-report.html'), html)
  ])

  return {
    markdown,
    json: reportData,
    html,
    paths: {
      markdown: path.join(artifactsDir, 'report.md'),
      json: path.join(artifactsDir, 'report.json'),
      html: path.join(artifactsDir, 'visual-report.html')
    }
  }
}

function calculateSummary(phases: PhaseResults): ReportSummary {
  let total = 0, passed = 0, failed = 0, warnings = 0, skipped = 0

  for (const [phaseName, phaseResult] of Object.entries(phases)) {
    if (!phaseResult || !phaseResult.results) {
      skipped++
      continue
    }

    for (const result of phaseResult.results) {
      total++
      switch (result.status) {
        case 'pass': passed++; break
        case 'fail': failed++; break
        case 'warn': warnings++; break
        case 'skip': skipped++; break
      }
    }
  }

  return {
    total_checks: total,
    passed,
    failed,
    warnings,
    skipped,
    pass_rate: total > 0 ? Math.round((passed / (total - skipped)) * 10000) / 100 : 0,
    verdict: 'PASSED',      // Will be overwritten by determineVerdict()
    blocking_failures: []   // Will be populated by collectBlockingFailures()
  }
}

function generateRecommendations(
  phases: PhaseResults,
  config: MigrateConfig
): Recommendation[] {
  const recommendations: Recommendation[] = []

  // Visual regression recommendations
  const failedVisuals = phases.visual_regression?.results?.filter(r => r.status === 'fail') || []
  if (failedVisuals.length > 0) {
    const dynamicContent = failedVisuals.filter(r => r.diff_percentage < 15)
    if (dynamicContent.length > 0) {
      recommendations.push({
        priority: 'should_fix',
        category: 'Visual Regression',
        message: `${dynamicContent.length} routes have minor visual differences (<15%). Consider adding mask_selectors for dynamic content (timestamps, avatars, ads).`,
        affected_routes: dynamicContent.map(r => r.route),
        auto_fixable: true
      })
    }

    const majorChanges = failedVisuals.filter(r => r.diff_percentage >= 15)
    if (majorChanges.length > 0) {
      recommendations.push({
        priority: 'must_fix',
        category: 'Visual Regression',
        message: `${majorChanges.length} routes have significant visual changes (>=15%). Review layout, styling, and component rendering differences.`,
        affected_routes: majorChanges.map(r => r.route),
        auto_fixable: false
      })
    }
  }

  // Performance recommendations
  const slowRoutes = phases.performance?.results?.filter(
    r => r.lcp > config.verification.performance_budget.page_load_max_ms
  ) || []
  if (slowRoutes.length > 0) {
    recommendations.push({
      priority: 'should_fix',
      category: 'Performance',
      message: `${slowRoutes.length} routes exceed the page load budget (${config.verification.performance_budget.page_load_max_ms}ms). Consider code splitting, image optimization, or SSR.`,
      affected_routes: slowRoutes.map(r => r.route),
      auto_fixable: false
    })
  }

  // Accessibility recommendations
  const a11yIssues = phases.accessibility?.results?.filter(
    r => r.violations.length > 0
  ) || []
  if (a11yIssues.length > 0) {
    const totalViolations = a11yIssues.reduce((sum, r) => sum + r.violations.length, 0)
    recommendations.push({
      priority: a11yIssues.some(r => r.violations.some(v => v.impact === 'critical')) ? 'must_fix' : 'should_fix',
      category: 'Accessibility',
      message: `${totalViolations} accessibility violations across ${a11yIssues.length} routes. Focus on critical/serious violations first.`,
      affected_routes: a11yIssues.map(r => r.route),
      auto_fixable: false
    })
  }

  // Cross-browser recommendations
  const browserIssues = phases.cross_browser?.results?.filter(
    r => !r.all_browsers_pass
  ) || []
  if (browserIssues.length > 0) {
    recommendations.push({
      priority: 'should_fix',
      category: 'Cross-Browser',
      message: `${browserIssues.length} routes have browser-specific issues. Check CSS compatibility and polyfill requirements.`,
      affected_routes: browserIssues.map(r => r.route),
      auto_fixable: false
    })
  }

  return recommendations.sort((a, b) => {
    const priorityOrder = { must_fix: 0, should_fix: 1, consider: 2 }
    return priorityOrder[a.priority] - priorityOrder[b.priority]
  })
}
```

#### Console Output Format

```
╔══════════════════════════════════════════════════════════════╗
║            Migration Verification Complete                  ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  Verdict:    PASSED_WITH_WARNINGS                           ║
║  Pass Rate:  96.3% (289/300 checks passed)                  ║
║  Duration:   47.2s                                          ║
║                                                              ║
║  Phase Results:                                              ║
║    Route Discovery .... 25/25 routes matched          [PASS] ║
║    Page Load .......... 25/25 pages loaded            [PASS] ║
║    Visual Regression .. 69/72 screenshots matched     [WARN] ║
║    Cross-Browser ...... 75/75 checks passed           [PASS] ║
║    Accessibility ...... 23/25 routes compliant        [WARN] ║
║    Performance ........ 23/25 within budget           [WARN] ║
║    State Integrity .... 50/50 states preserved        [PASS] ║
║                                                              ║
║  Artifacts:                                                  ║
║    Report:  .migration-verification/artifacts/report.md      ║
║    Visual:  .migration-verification/artifacts/visual-report.html ║
║    JSON:    .migration-verification/artifacts/report.json    ║
║                                                              ║
║  Recommendations: 3 must_fix, 2 should_fix                  ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
```

### Execution Order (Final)

```
Phase 1: Dev Server Lifecycle (COMPLETE)
  → Servers running and accessible

Phase 2: Route Discovery (COMPLETE)
  → Routes discovered and registered

Phase 3: Visual Regression (COMPLETE)
  → Screenshot comparison across viewports

Phase 4: Behavioral Testing (COMPLETE)
  → Page load, navigation, forms, API calls, JS errors

Phase 5: Cross-Browser (COMPLETE)
  → Chromium, Firefox, WebKit + Mobile device emulation

Phase 6: Performance (COMPLETE)
  → Core Web Vitals, load times, bundle sizes, budgets

Phase 7: Accessibility (COMPLETE)
  → axe-core WCAG compliance + regression comparison

Phase 8: Agent & Skill (COMPLETE)
  → e2e-tester enhancement + Playwright migration skill + examples

Phase 9: Command & Configuration (COMPLETE)
  → Execution engine, flags, dependencies, CI/CD, error handling

Phase 10: Reports & Artifacts (COMPLETE)
  → Unified report template, HTML visual report, CI/CD artifact integration
```

