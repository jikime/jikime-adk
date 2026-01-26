---

## Phase 6: Performance Comparison

**Activation**: `--performance` or `--full` flag

### Overview

Performance comparison measures Core Web Vitals, page load timing, and resource sizes for both source and target applications, then compares them against configurable performance budgets. This ensures the migration does not introduce performance regressions.

### Step 6.1: Core Web Vitals Collection

**Purpose**: Measure LCP, CLS, and INP (replacing FID) for both source and target.

**Web Vitals Measurement:**

```typescript
interface WebVitals {
  lcp: number | null    // Largest Contentful Paint (ms)
  cls: number | null    // Cumulative Layout Shift (score)
  inp: number | null    // Interaction to Next Paint (ms)
  fcp: number | null    // First Contentful Paint (ms)
  ttfb: number | null   // Time to First Byte (ms)
}

interface PerformanceMetrics {
  route: string
  server: 'source' | 'target'
  webVitals: WebVitals
  timing: NavigationTiming
  resources: ResourceMetrics
  timestamp: string
}

async function collectWebVitals(
  page: Page,
  url: string,
  route: string
): Promise<WebVitals> {
  // Inject web-vitals collection script before navigation
  await page.addInitScript(() => {
    (window as any).__webVitals = {
      lcp: null, cls: null, inp: null, fcp: null, ttfb: null
    }

    // LCP Observer
    new PerformanceObserver((list) => {
      const entries = list.getEntries()
      const lastEntry = entries[entries.length - 1]
      ;(window as any).__webVitals.lcp = lastEntry.startTime
    }).observe({ type: 'largest-contentful-paint', buffered: true })

    // CLS Observer
    let clsValue = 0
    new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        if (!(entry as any).hadRecentInput) {
          clsValue += (entry as any).value
        }
      }
      ;(window as any).__webVitals.cls = clsValue
    }).observe({ type: 'layout-shift', buffered: true })

    // FCP from Paint Timing
    new PerformanceObserver((list) => {
      const entries = list.getEntries()
      const fcp = entries.find(e => e.name === 'first-contentful-paint')
      if (fcp) (window as any).__webVitals.fcp = fcp.startTime
    }).observe({ type: 'paint', buffered: true })
  })

  // Navigate and wait for page to be fully loaded
  await page.goto(url + route, { waitUntil: 'networkidle', timeout: 30000 })

  // Wait additional time for LCP to stabilize
  await page.waitForTimeout(3000)

  // Trigger interaction for INP measurement
  await page.mouse.click(0, 0).catch(() => {})
  await page.waitForTimeout(500)

  // Collect results
  const vitals = await page.evaluate(() => {
    const v = (window as any).__webVitals

    // TTFB from Navigation Timing
    const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
    v.ttfb = nav ? nav.responseStart - nav.requestStart : null

    return v
  })

  return vitals
}
```

### Step 6.2: Navigation Timing Collection

**Purpose**: Measure detailed page load timing breakdown.

**Timing Breakdown:**

```typescript
interface NavigationTiming {
  // Connection
  dnsLookup: number        // domainLookupEnd - domainLookupStart
  tcpConnection: number    // connectEnd - connectStart
  tlsNegotiation: number   // requestStart - secureConnectionStart

  // Request/Response
  ttfb: number             // responseStart - requestStart
  contentDownload: number  // responseEnd - responseStart

  // Processing
  domParsing: number       // domInteractive - responseEnd
  domContentLoaded: number // domContentLoadedEventEnd - navigationStart
  domComplete: number      // domComplete - navigationStart

  // Full page
  pageLoad: number         // loadEventEnd - navigationStart
  totalTime: number        // loadEventEnd - fetchStart
}

async function collectNavigationTiming(page: Page): Promise<NavigationTiming> {
  return await page.evaluate(() => {
    const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming

    if (!nav) return null

    return {
      dnsLookup: nav.domainLookupEnd - nav.domainLookupStart,
      tcpConnection: nav.connectEnd - nav.connectStart,
      tlsNegotiation: nav.secureConnectionStart > 0
        ? nav.requestStart - nav.secureConnectionStart : 0,
      ttfb: nav.responseStart - nav.requestStart,
      contentDownload: nav.responseEnd - nav.responseStart,
      domParsing: nav.domInteractive - nav.responseEnd,
      domContentLoaded: nav.domContentLoadedEventEnd - nav.startTime,
      domComplete: nav.domComplete - nav.startTime,
      pageLoad: nav.loadEventEnd - nav.startTime,
      totalTime: nav.loadEventEnd - nav.fetchStart
    }
  })
}
```

**Multi-Run Averaging:**

```
FUNCTION collectTimingAverage(page, url, route, runs=3):
  timings = []

  FOR i in range(runs):
    // Clear cache between runs
    await page.context().clearCookies()
    await page.evaluate(() => {
      caches?.keys().then(names => names.forEach(name => caches.delete(name)))
    }).catch(() => {})

    timing = await collectNavigationTiming(page, url, route)
    timings.add(timing)

    // Brief pause between runs
    await page.waitForTimeout(500)

  // Return median values (more stable than average)
  RETURN medianTiming(timings)
```

### Step 6.3: Resource & Bundle Size Analysis

**Purpose**: Compare JavaScript, CSS, and total transfer sizes between source and target.

**Resource Collection:**

```typescript
interface ResourceMetrics {
  totalTransferSize: number   // Total bytes transferred
  totalDecodedSize: number    // Total decoded (uncompressed) size
  jsSize: number              // JavaScript transfer size
  cssSize: number             // CSS transfer size
  imageSize: number           // Image transfer size
  fontSize: number            // Font transfer size
  otherSize: number           // Other resources
  requestCount: number        // Total network requests
  jsRequestCount: number      // JavaScript file count
  cssRequestCount: number     // CSS file count
  thirdPartySize: number      // Third-party resources size
  largestResources: ResourceEntry[]  // Top 5 largest resources
}

interface ResourceEntry {
  url: string
  type: string
  transferSize: number
  decodedSize: number
}

async function collectResourceMetrics(
  page: Page,
  url: string,
  route: string
): Promise<ResourceMetrics> {
  const resources: ResourceEntry[] = []

  // Monitor network requests
  page.on('response', async response => {
    const request = response.request()
    const headers = await response.allHeaders()
    const contentLength = parseInt(headers['content-length'] ?? '0')

    resources.push({
      url: response.url(),
      type: request.resourceType(),
      transferSize: contentLength,
      decodedSize: contentLength  // Approximation
    })
  })

  await page.goto(url + route, { waitUntil: 'networkidle', timeout: 30000 })

  // Also collect from Performance API for more accurate data
  const perfResources = await page.evaluate(() => {
    return performance.getEntriesByType('resource').map(r => ({
      url: r.name,
      type: (r as any).initiatorType,
      transferSize: (r as PerformanceResourceTiming).transferSize,
      decodedSize: (r as PerformanceResourceTiming).decodedBodySize
    }))
  })

  // Merge with Performance API data (more accurate sizes)
  const allResources = perfResources.length > 0 ? perfResources : resources
  const baseHost = new URL(url).host

  return {
    totalTransferSize: sum(allResources, 'transferSize'),
    totalDecodedSize: sum(allResources, 'decodedSize'),
    jsSize: sumByType(allResources, ['script', 'js']),
    cssSize: sumByType(allResources, ['css', 'link']),
    imageSize: sumByType(allResources, ['img', 'image']),
    fontSize: sumByType(allResources, ['font']),
    otherSize: sumByType(allResources, ['other', 'fetch', 'xmlhttprequest']),
    requestCount: allResources.length,
    jsRequestCount: allResources.filter(r => isJs(r)).length,
    cssRequestCount: allResources.filter(r => isCss(r)).length,
    thirdPartySize: allResources
      .filter(r => !new URL(r.url).host.includes(baseHost))
      .reduce((sum, r) => sum + r.transferSize, 0),
    largestResources: allResources
      .sort((a, b) => b.transferSize - a.transferSize)
      .slice(0, 5)
  }
}
```

### Step 6.4: Performance Budget Validation

**Purpose**: Compare source vs target metrics against configurable performance budgets.

**Budget Configuration (from `.migrate-config.yaml`):**

```yaml
verification:
  performance_budget:
    # Absolute thresholds (target must not exceed)
    lcp_max_ms: 2500            # Max LCP in ms
    cls_max: 0.1                # Max CLS score
    inp_max_ms: 200             # Max INP in ms
    fcp_max_ms: 1800            # Max FCP in ms
    ttfb_max_ms: 800            # Max TTFB in ms
    page_load_max_ms: 3000      # Max total page load
    js_bundle_max_kb: 500       # Max JS bundle size (KB)
    css_bundle_max_kb: 150      # Max CSS bundle size (KB)
    total_transfer_max_kb: 2000 # Max total transfer size (KB)
    request_count_max: 80       # Max network requests

    # Regression thresholds (target vs source comparison)
    lcp_regression_pct: 20      # Max LCP regression vs source
    cls_regression_pct: 50      # Max CLS regression vs source
    load_time_regression_pct: 25 # Max load time regression vs source
    js_size_regression_pct: 30  # Max JS size regression vs source
    request_count_regression_pct: 50  # Max request count regression
```

**Budget Validation Engine:**

```typescript
interface BudgetResult {
  metric: string
  sourceValue: number
  targetValue: number
  budgetValue: number
  budgetType: 'absolute' | 'regression'
  passed: boolean
  change: string          // e.g., "-14%", "+120ms"
  severity: 'pass' | 'warn' | 'fail'
}

function validatePerformanceBudget(
  sourceMetrics: PerformanceMetrics,
  targetMetrics: PerformanceMetrics,
  budget: PerformanceBudget
): BudgetResult[] {
  const results: BudgetResult[] = []

  // Absolute budget checks
  const absoluteChecks = [
    { metric: 'LCP', value: targetMetrics.webVitals.lcp, max: budget.lcp_max_ms },
    { metric: 'CLS', value: targetMetrics.webVitals.cls, max: budget.cls_max },
    { metric: 'INP', value: targetMetrics.webVitals.inp, max: budget.inp_max_ms },
    { metric: 'FCP', value: targetMetrics.webVitals.fcp, max: budget.fcp_max_ms },
    { metric: 'TTFB', value: targetMetrics.webVitals.ttfb, max: budget.ttfb_max_ms },
    { metric: 'Page Load', value: targetMetrics.timing.pageLoad, max: budget.page_load_max_ms },
    { metric: 'JS Bundle', value: targetMetrics.resources.jsSize / 1024, max: budget.js_bundle_max_kb },
    { metric: 'CSS Bundle', value: targetMetrics.resources.cssSize / 1024, max: budget.css_bundle_max_kb },
    { metric: 'Total Transfer', value: targetMetrics.resources.totalTransferSize / 1024, max: budget.total_transfer_max_kb },
    { metric: 'Request Count', value: targetMetrics.resources.requestCount, max: budget.request_count_max }
  ]

  for (const check of absoluteChecks) {
    if (check.value === null) continue
    const passed = check.value <= check.max
    results.push({
      metric: check.metric,
      sourceValue: 0,
      targetValue: check.value,
      budgetValue: check.max,
      budgetType: 'absolute',
      passed,
      change: formatValue(check.metric, check.value),
      severity: passed ? 'pass' : 'fail'
    })
  }

  // Regression checks (target vs source)
  const regressionChecks = [
    { metric: 'LCP', source: sourceMetrics.webVitals.lcp, target: targetMetrics.webVitals.lcp, maxPct: budget.lcp_regression_pct },
    { metric: 'CLS', source: sourceMetrics.webVitals.cls, target: targetMetrics.webVitals.cls, maxPct: budget.cls_regression_pct },
    { metric: 'Page Load', source: sourceMetrics.timing.pageLoad, target: targetMetrics.timing.pageLoad, maxPct: budget.load_time_regression_pct },
    { metric: 'JS Size', source: sourceMetrics.resources.jsSize, target: targetMetrics.resources.jsSize, maxPct: budget.js_size_regression_pct },
    { metric: 'Request Count', source: sourceMetrics.resources.requestCount, target: targetMetrics.resources.requestCount, maxPct: budget.request_count_regression_pct }
  ]

  for (const check of regressionChecks) {
    if (check.source === null || check.target === null) continue
    const regressionPct = check.source > 0
      ? ((check.target - check.source) / check.source) * 100
      : 0
    const passed = regressionPct <= check.maxPct
    const improved = regressionPct < 0

    results.push({
      metric: `${check.metric} (regression)`,
      sourceValue: check.source,
      targetValue: check.target,
      budgetValue: check.maxPct,
      budgetType: 'regression',
      passed,
      change: `${regressionPct >= 0 ? '+' : ''}${regressionPct.toFixed(1)}%`,
      severity: improved ? 'pass' : (passed ? 'warn' : 'fail')
    })
  }

  return results
}
```

### Step 6.5: Performance Comparison Report

**Purpose**: Generate detailed performance comparison between source and target.

**Report Structure:**

```
Performance Comparison Report
==============================

Route: / (Homepage)
Runs: 3 (median values)

Core Web Vitals:
| Metric | Source | Target | Change | Budget | Status |
|--------|--------|--------|--------|--------|--------|
| LCP | 2.1s | 1.8s | -14% | <2.5s | PASS |
| CLS | 0.05 | 0.03 | -40% | <0.1 | PASS |
| INP | 80ms | 45ms | -44% | <200ms | PASS |
| FCP | 1.2s | 0.9s | -25% | <1.8s | PASS |
| TTFB | 180ms | 150ms | -17% | <800ms | PASS |

Page Load Timing:
| Phase | Source | Target | Change |
|-------|--------|--------|--------|
| DNS Lookup | 5ms | 5ms | 0% |
| TCP Connect | 12ms | 12ms | 0% |
| TTFB | 180ms | 150ms | -17% |
| Content Download | 45ms | 38ms | -16% |
| DOM Parsing | 320ms | 280ms | -13% |
| DOM Complete | 1.8s | 1.5s | -17% |
| Page Load | 2.1s | 1.8s | -14% |

Resource Analysis:
| Resource | Source | Target | Change | Budget | Status |
|----------|--------|--------|--------|--------|--------|
| JS Bundle | 450KB | 380KB | -16% | <500KB | PASS |
| CSS Bundle | 85KB | 72KB | -15% | <150KB | PASS |
| Images | 1.2MB | 1.1MB | -8% | - | INFO |
| Fonts | 120KB | 120KB | 0% | - | INFO |
| Total | 1.9MB | 1.7MB | -11% | <2MB | PASS |
| Requests | 42 | 38 | -10% | <80 | PASS |

Top 5 Largest Resources (Target):
1. /static/js/main.chunk.js - 185KB
2. /static/js/vendor.chunk.js - 142KB
3. /images/hero.webp - 98KB
4. /static/css/main.css - 52KB
5. /fonts/inter-var.woff2 - 45KB
```

**Aggregated Performance Summary:**

```typescript
interface PerformanceSummary {
  routesAnalyzed: number
  overallScore: number          // 0-100
  improvements: MetricChange[]  // Metrics that improved
  regressions: MetricChange[]   // Metrics that regressed
  budgetViolations: BudgetResult[]
  recommendations: string[]
}

interface MetricChange {
  route: string
  metric: string
  sourceValue: number
  targetValue: number
  changePct: number
}

function generatePerformanceSummary(
  allResults: Map<string, { source: PerformanceMetrics; target: PerformanceMetrics }>,
  budgetResults: BudgetResult[]
): PerformanceSummary {
  const improvements: MetricChange[] = []
  const regressions: MetricChange[] = []

  for (const [route, { source, target }] of allResults) {
    // LCP comparison
    if (source.webVitals.lcp && target.webVitals.lcp) {
      const change = ((target.webVitals.lcp - source.webVitals.lcp) / source.webVitals.lcp) * 100
      const entry = { route, metric: 'LCP', sourceValue: source.webVitals.lcp, targetValue: target.webVitals.lcp, changePct: change }
      if (change < -5) improvements.push(entry)
      else if (change > 10) regressions.push(entry)
    }

    // Page load comparison
    if (source.timing.pageLoad && target.timing.pageLoad) {
      const change = ((target.timing.pageLoad - source.timing.pageLoad) / source.timing.pageLoad) * 100
      const entry = { route, metric: 'Page Load', sourceValue: source.timing.pageLoad, targetValue: target.timing.pageLoad, changePct: change }
      if (change < -5) improvements.push(entry)
      else if (change > 10) regressions.push(entry)
    }
  }

  // Generate recommendations
  const recommendations: string[] = []
  if (regressions.some(r => r.metric === 'LCP'))
    recommendations.push('Consider lazy-loading below-fold images and deferring non-critical JS')
  if (budgetResults.some(r => r.metric.includes('JS') && !r.passed))
    recommendations.push('JS bundle exceeds budget - consider code splitting or tree shaking')
  if (budgetResults.some(r => r.metric.includes('Request') && !r.passed))
    recommendations.push('Too many network requests - consider bundling or HTTP/2 push')

  const violations = budgetResults.filter(r => !r.passed)
  const overallScore = Math.max(0, 100 - (violations.length * 10) - (regressions.length * 5))

  return {
    routesAnalyzed: allResults.size,
    overallScore,
    improvements,
    regressions,
    budgetViolations: violations,
    recommendations
  }
}
```

### Step 6.6: Performance Test Execution Flow

**Purpose**: Orchestrate complete performance comparison workflow.

**Execution Flow:**

```
FUNCTION runPerformanceComparison(route_registry, sourceUrl, targetUrl, budget):
  allResults = new Map()

  FOR each route in route_registry.static_routes:
    // Collect source metrics (3 runs, median)
    sourceMetrics = collectMetricsWithRetry(sourceUrl, route, runs=3)

    // Collect target metrics (3 runs, median)
    targetMetrics = collectMetricsWithRetry(targetUrl, route, runs=3)

    allResults.set(route.path, { source: sourceMetrics, target: targetMetrics })

  // Validate budgets for each route
  allBudgetResults = []
  FOR each [route, { source, target }] of allResults:
    budgetResults = validatePerformanceBudget(source, target, budget)
    allBudgetResults.addAll(budgetResults)

  // Generate summary
  summary = generatePerformanceSummary(allResults, allBudgetResults)

  RETURN { allResults, allBudgetResults, summary }
```

**Performance Collection with Retry:**

```
FUNCTION collectMetricsWithRetry(url, route, runs=3):
  metrics = []

  FOR i in range(runs):
    TRY:
      browser = await chromium.launch({ headless: true })
      context = await browser.newContext()
      page = await context.newPage()

      // Collect all metrics
      vitals = await collectWebVitals(page, url, route.path)
      timing = await collectNavigationTiming(page)
      resources = await collectResourceMetrics(page, url, route.path)

      metrics.add({ webVitals: vitals, timing, resources })

      await browser.close()
    CATCH error:
      // Log error, continue with remaining runs
      console.warn(`Run ${i+1} failed for ${route.path}: ${error.message}`)

  IF metrics.length === 0:
    RETURN null  // All runs failed

  // Return median metrics
  RETURN medianMetrics(metrics)
```

---
