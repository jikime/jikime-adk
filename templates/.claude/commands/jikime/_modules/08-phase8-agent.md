## Phase 8: Agent & Skill Implementation

**Purpose**: Define the enhanced `e2e-tester` agent specification for migration verification and create a dedicated Playwright migration skill with reference patterns and example code.

### Step 8.1: Enhanced e2e-tester Agent - Migration Verification Workflow

**Purpose**: Extend the e2e-tester agent to support migration-specific verification workflows invoked by F.R.I.D.A.Y.

**Enhanced Agent Frontmatter:**

```yaml
---
name: e2e-tester
description: |
  E2E test specialist (Playwright). User journey test creation, execution, and maintenance.
  Supports TWO modes:
  - Development Mode (J.A.R.V.I.S.): Standard E2E test creation and execution
  - Migration Mode (F.R.I.D.A.Y.): Migration verification with dual-server comparison
  MUST INVOKE when keywords detected:
  EN: e2e test, Playwright, browser test, user flow, visual regression, migration verify
  KO: E2E 테스트, Playwright, 브라우저 테스트, 사용자 플로우, 시각적 회귀, 마이그레이션 검증
tools: Read, Write, Edit, Bash, Grep, Glob, mcp__sequential-thinking__sequentialthinking
model: opus

# Progressive Disclosure: 3-Level Skill Loading
skills: jikime-workflow-testing, jikime-workflow-playwright-migration
---
```

**Migration Mode Context Contract:**

```typescript
interface MigrationVerificationContext {
  // From F.R.I.D.A.Y. orchestrator
  mode: 'migration'
  config: {
    sourceDir: string
    outputDir: string
    sourceFramework: string
    targetFramework: string
    artifactsDir: string
  }
  verification: {
    sourcePort: number       // Default: 3000
    targetPort: number       // Default: 3001
    visualThreshold: number  // Default: 5
    crawlDepth: number       // Default: 3
    wcagLevel: 'A' | 'AA' | 'AAA'
  }
  flags: {
    full: boolean
    visual: boolean
    crossBrowser: boolean
    a11y: boolean
    performance: boolean
    behavior: boolean
  }
  routes?: string[]          // Manual route overrides
}

interface MigrationVerificationResult {
  // Returned to F.R.I.D.A.Y. orchestrator
  status: 'passed' | 'failed' | 'partial'
  summary: {
    totalTests: number
    passed: number
    failed: number
    warnings: number
    passRate: number
  }
  categories: {
    pageLoad: CategoryResult
    navigation: CategoryResult
    visual?: CategoryResult
    performance?: CategoryResult
    crossBrowser?: CategoryResult
    accessibility?: CategoryResult
  }
  failedTests: FailedTest[]
  reportPath: string         // Path to generated report
  artifactsPath: string      // Path to screenshots/traces
  recommendations: string[]
  duration: number           // Total execution time (ms)
}

interface CategoryResult {
  passed: number
  failed: number
  warnings: number
  score: number              // 0-100
}

interface FailedTest {
  category: string
  route: string
  description: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  evidence: string | null    // Screenshot or error details
}
```

**Migration Mode Execution Flow:**

```
FUNCTION executeMigrationVerification(context: MigrationVerificationContext):
  // Phase 1: Infrastructure
  servers = await startDualServers(context.config)
  IF servers.failed:
    RETURN { status: 'failed', error: 'Server startup failed' }

  TRY:
    // Phase 2: Route Discovery
    routes = context.routes ?? await discoverRoutes(context.config.artifactsDir)

    // Phase 3-7: Run verification phases based on flags
    results = {}

    // Always run: Page Load + Navigation + Behavior
    results.pageLoad = await verifyAllPages(routes, servers.targetUrl)
    results.navigation = await verifyNavigation(routes, servers.targetUrl, context.verification.crawlDepth)
    results.behavior = await compareBehavior(routes, servers.sourceUrl, servers.targetUrl)

    // Conditional phases
    IF context.flags.visual OR context.flags.full:
      results.visual = await runVisualRegression(routes, servers, context.verification.visualThreshold)

    IF context.flags.performance OR context.flags.full:
      results.performance = await runPerformanceComparison(routes, servers)

    IF context.flags.crossBrowser OR context.flags.full:
      results.crossBrowser = await runCrossBrowserTests(routes, servers.targetUrl)

    IF context.flags.a11y OR context.flags.full:
      results.accessibility = await runAccessibilityVerification(routes, servers, context.verification.wcagLevel)

    // Generate consolidated report
    report = generateVerificationReport(results, context)
    WRITE report TO context.config.artifactsDir + '/verification_report.md'

    RETURN buildVerificationResult(results, report)

  FINALLY:
    await stopDualServers(servers)
```

**Agent Behavior Changes:**

| Aspect | Development Mode (J.A.R.V.I.S.) | Migration Mode (F.R.I.D.A.Y.) |
|--------|----------------------------------|-------------------------------|
| Trigger | E2E test creation/execution | Migration verification |
| Input | Test specs, user flows | `.migrate-config.yaml`, routes |
| Server | Single target server | Dual servers (source + target) |
| Focus | Test quality, coverage | Behavior preservation, regression |
| Output | Test results, coverage | Verification report, diff evidence |
| Failure | Test failures | Regression detection |

### Step 8.2: Playwright Migration Skill Structure

**Purpose**: Create a dedicated skill that provides migration-specific Playwright patterns, best practices, and reusable code modules.

**Skill Directory Layout:**

```
.claude/skills/jikime-workflow-playwright-migration/
├── SKILL.md                    # Frontmatter + triggers + overview
├── reference.md                # Best practices for migration testing
├── modules/
│   ├── visual-regression.md    # Screenshot comparison patterns
│   ├── route-discovery.md      # Auto-route detection patterns
│   ├── server-lifecycle.md     # Dev server management patterns
│   ├── behavioral-testing.md   # Page load, navigation, form verification
│   ├── cross-browser.md        # Multi-browser + mobile emulation
│   ├── performance.md          # Core Web Vitals + budget validation
│   └── accessibility.md        # axe-core integration patterns
└── examples/
    ├── visual-comparison.ts    # TypeScript visual regression example
    ├── route-crawler.ts        # Route auto-discovery example
    └── cross-browser-verify.ts # Cross-browser verification example
```

**SKILL.md Frontmatter:**

```yaml
---
name: jikime-workflow-playwright-migration
description: |
  Migration-specific Playwright verification patterns. Provides reusable modules
  for visual regression, route discovery, server lifecycle, behavioral testing,
  cross-browser verification, performance comparison, and accessibility checks.
version: 1.0.0
tags: ["workflow", "playwright", "migration", "e2e", "verification", "visual-regression"]
triggers:
  keywords: ["migration verify", "visual regression", "playwright migration", "cross-browser migration", "migrate verify", "마이그레이션 검증"]
  phases: ["verify"]
  agents: ["e2e-tester"]
  languages: ["typescript"]
user-invocable: false

# Progressive Disclosure Configuration
progressive_disclosure:
  enabled: true
  level1_tokens: ~120
  level2_tokens: ~6000

allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
---
```

**SKILL.md Body (Overview):**

```markdown
# Playwright Migration Verification Skill

Reusable patterns for post-migration verification using Playwright.

## When to Use

- After `/jikime:migrate-3-execute` completes
- When `--visual`, `--cross-browser`, `--a11y`, `--performance`, or `--full` flags are used
- For regression comparison between source and target applications

## Quick Reference

| Module | Purpose | Key Pattern |
|--------|---------|-------------|
| visual-regression | Screenshot diff | pixelmatch + threshold |
| route-discovery | Auto-route detection | Artifact parsing + crawl |
| server-lifecycle | Dev server management | Framework-aware start/stop |
| behavioral-testing | Functional verification | Page load + navigation + forms |
| cross-browser | Multi-browser testing | Parallel Chromium/Firefox/WebKit |
| performance | Core Web Vitals | PerformanceObserver + budgets |
| accessibility | WCAG compliance | axe-core + regression diff |

## Dependencies

```json
{
  "@playwright/test": "^1.40.0",
  "@axe-core/playwright": "^4.8.0",
  "pixelmatch": "^5.3.0",
  "pngjs": "^7.0.0"
}
```

## Integration with Migration Workflow

```
migrate-3-execute (output) → migrate-4-verify (input)
                                    ↓
                              Skill Modules
                              ┌─────────────────┐
                              │ server-lifecycle │ Phase 1
                              │ route-discovery  │ Phase 2
                              │ visual-regression│ Phase 3
                              │ behavioral-test  │ Phase 4
                              │ cross-browser    │ Phase 5
                              │ performance      │ Phase 6
                              │ accessibility    │ Phase 7
                              └─────────────────┘
                                    ↓
                           verification_report.md
```
```

**reference.md (Best Practices):**

```markdown
# Migration Testing Best Practices

## Dual-Server Architecture

### Why Dual Servers?
Migration verification requires comparing the SAME routes on two different implementations.
Running source and target simultaneously enables direct A/B comparison.

### Port Isolation
- Source: port 3000 (original application)
- Target: port 3001 (migrated application)
- Never share ports or sessions between source/target

### Health Check Before Testing
Always verify both servers respond before running tests:
- HTTP GET to root path
- Status < 500
- Body contains content
- Timeout: 30 seconds with 1-second interval

## Visual Regression Strategy

### Threshold Selection
| Content Type | Recommended Threshold | Reasoning |
|-------------|----------------------|-----------|
| Static pages | 1-3% | Minimal acceptable difference |
| Dynamic content | 5-8% | Account for timestamps, ads |
| Data-heavy pages | 8-12% | Table sorting, pagination variance |
| Animation pages | 15-20% | Frame timing differences |

### Mask Dynamic Content
Always mask elements that change between renders:
- Timestamps and dates
- Random IDs or tokens
- Advertisement banners
- User-specific content
- Live data feeds

### Multi-Viewport Strategy
Test at minimum 3 viewports to catch responsive issues:
- Desktop: 1920x1080 (standard)
- Tablet: 768x1024 (iPad portrait)
- Mobile: 375x667 (iPhone SE)

## Performance Testing Guidelines

### Multi-Run Median
Always collect 3+ runs and use MEDIAN (not average):
- Average is skewed by outliers
- Median represents typical user experience
- Discard first run (cold cache effects)

### Cache Clearing Between Runs
Clear all caches between performance measurement runs:
- Browser cache (via context recreation)
- Service worker cache
- Application state

### Regression vs Absolute Budgets
- **Absolute**: "LCP must be < 2.5s" (industry standards)
- **Regression**: "LCP must not worsen by > 20%" (migration-specific)
- Use BOTH for comprehensive validation

## Accessibility Testing Guidelines

### Scan After Render
Wait for dynamic content to fully render before axe-core scan:
- Wait for `networkidle` state
- Additional 1-second buffer for JS-rendered content
- Handle async data loading

### Focus on NEW Violations
Migration should not INTRODUCE new a11y issues:
- Compare source vs target violations by rule ID
- New critical/serious violations = regression = FAIL
- Existing violations are acceptable (pre-existing debt)

### Exclude Third-Party Content
Skip elements outside your control:
- Third-party widgets (chat, analytics)
- Embedded iframes
- CDN-served content with fixed HTML
```

### Step 8.3: Example Code Templates

**Purpose**: Provide ready-to-use TypeScript examples that demonstrate key verification patterns.

**Example 1: `visual-comparison.ts`**

```typescript
/**
 * Visual Regression Comparison Example
 *
 * Compares screenshots between source and target servers
 * for a list of routes across multiple viewports.
 *
 * Usage: npx ts-node examples/visual-comparison.ts
 */

import { chromium, Browser, Page } from 'playwright'
import { PNG } from 'pngjs'
import pixelmatch from 'pixelmatch'
import * as fs from 'fs'
import * as path from 'path'

interface ComparisonConfig {
  sourceUrl: string
  targetUrl: string
  routes: string[]
  viewports: { name: string; width: number; height: number }[]
  threshold: number          // Allowed diff percentage (0-100)
  outputDir: string
  maskSelectors: string[]    // CSS selectors to mask before comparison
}

interface ComparisonResult {
  route: string
  viewport: string
  diffPercentage: number
  diffPixels: number
  totalPixels: number
  passed: boolean
  screenshotPaths: {
    source: string
    target: string
    diff: string
  }
}

const DEFAULT_CONFIG: ComparisonConfig = {
  sourceUrl: 'http://localhost:3000',
  targetUrl: 'http://localhost:3001',
  routes: ['/'],
  viewports: [
    { name: 'desktop', width: 1920, height: 1080 },
    { name: 'tablet', width: 768, height: 1024 },
    { name: 'mobile', width: 375, height: 667 }
  ],
  threshold: 5,
  outputDir: './verification-artifacts/screenshots',
  maskSelectors: ['[data-testid="timestamp"]', '.ad-banner']
}

async function captureScreenshot(
  browser: Browser,
  url: string,
  route: string,
  viewport: { width: number; height: number },
  maskSelectors: string[]
): Promise<Buffer> {
  const context = await browser.newContext({ viewport })
  const page = await context.newPage()

  await page.goto(url + route, { waitUntil: 'networkidle', timeout: 30000 })

  // Mask dynamic content
  for (const selector of maskSelectors) {
    await page.locator(selector).evaluateAll(elements => {
      elements.forEach(el => {
        ;(el as HTMLElement).style.visibility = 'hidden'
      })
    }).catch(() => {}) // Ignore if selector not found
  }

  // Wait for fonts and images
  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(500)

  const screenshot = await page.screenshot({ fullPage: true, type: 'png' })

  await context.close()
  return screenshot
}

function compareImages(
  sourceBuffer: Buffer,
  targetBuffer: Buffer,
  threshold: number
): { diffPercentage: number; diffPixels: number; diffImage: Buffer } {
  const sourcePng = PNG.sync.read(sourceBuffer)
  const targetPng = PNG.sync.read(targetBuffer)

  // Normalize dimensions (use larger of the two)
  const width = Math.max(sourcePng.width, targetPng.width)
  const height = Math.max(sourcePng.height, targetPng.height)

  // Resize if needed (pad with white)
  const normalizedSource = normalizeImage(sourcePng, width, height)
  const normalizedTarget = normalizeImage(targetPng, width, height)

  const diffPng = new PNG({ width, height })
  const diffPixels = pixelmatch(
    normalizedSource.data, normalizedTarget.data, diffPng.data,
    width, height,
    { threshold: 0.1, alpha: 0.5, diffColor: [255, 0, 0] }
  )

  const totalPixels = width * height
  const diffPercentage = (diffPixels / totalPixels) * 100

  return {
    diffPercentage,
    diffPixels,
    diffImage: PNG.sync.write(diffPng)
  }
}

function normalizeImage(png: PNG, targetWidth: number, targetHeight: number): PNG {
  if (png.width === targetWidth && png.height === targetHeight) return png

  const normalized = new PNG({ width: targetWidth, height: targetHeight, fill: true })
  // Fill with white background
  for (let i = 0; i < normalized.data.length; i += 4) {
    normalized.data[i] = 255     // R
    normalized.data[i + 1] = 255 // G
    normalized.data[i + 2] = 255 // B
    normalized.data[i + 3] = 255 // A
  }
  // Copy source pixels
  PNG.bitblt(png, normalized, 0, 0, png.width, png.height, 0, 0)
  return normalized
}

async function runVisualComparison(config: ComparisonConfig): Promise<ComparisonResult[]> {
  const results: ComparisonResult[] = []
  const browser = await chromium.launch({ headless: true })

  // Ensure output directories exist
  for (const sub of ['source', 'target', 'diff']) {
    fs.mkdirSync(path.join(config.outputDir, sub), { recursive: true })
  }

  try {
    for (const route of config.routes) {
      for (const viewport of config.viewports) {
        const safeName = route.replace(/\//g, '_') || '_root'

        // Capture both screenshots
        const sourceScreenshot = await captureScreenshot(
          browser, config.sourceUrl, route, viewport, config.maskSelectors
        )
        const targetScreenshot = await captureScreenshot(
          browser, config.targetUrl, route, viewport, config.maskSelectors
        )

        // Compare
        const { diffPercentage, diffPixels, diffImage } = compareImages(
          sourceScreenshot, targetScreenshot, config.threshold
        )

        // Save files
        const paths = {
          source: path.join(config.outputDir, 'source', `${safeName}-${viewport.name}.png`),
          target: path.join(config.outputDir, 'target', `${safeName}-${viewport.name}.png`),
          diff: path.join(config.outputDir, 'diff', `${safeName}-${viewport.name}-diff.png`)
        }

        fs.writeFileSync(paths.source, sourceScreenshot)
        fs.writeFileSync(paths.target, targetScreenshot)
        if (diffPercentage > 0) {
          fs.writeFileSync(paths.diff, diffImage)
        }

        const totalPixels = Math.max(
          PNG.sync.read(sourceScreenshot).width * PNG.sync.read(sourceScreenshot).height,
          PNG.sync.read(targetScreenshot).width * PNG.sync.read(targetScreenshot).height
        )

        results.push({
          route,
          viewport: viewport.name,
          diffPercentage,
          diffPixels,
          totalPixels,
          passed: diffPercentage <= config.threshold,
          screenshotPaths: paths
        })
      }
    }
  } finally {
    await browser.close()
  }

  return results
}

// Entry point
;(async () => {
  const config: ComparisonConfig = {
    ...DEFAULT_CONFIG,
    routes: process.argv.slice(2).length > 0
      ? process.argv.slice(2)
      : DEFAULT_CONFIG.routes
  }

  console.log(`Visual Comparison: ${config.sourceUrl} vs ${config.targetUrl}`)
  console.log(`Routes: ${config.routes.join(', ')}`)
  console.log(`Threshold: ${config.threshold}%\n`)

  const results = await runVisualComparison(config)

  // Print results
  const passed = results.filter(r => r.passed).length
  const total = results.length

  for (const result of results) {
    const status = result.passed ? 'PASS' : 'FAIL'
    console.log(`[${status}] ${result.route} (${result.viewport}) - diff: ${result.diffPercentage.toFixed(2)}%`)
  }

  console.log(`\nResults: ${passed}/${total} passed (${((passed/total)*100).toFixed(1)}%)`)
  process.exit(passed === total ? 0 : 1)
})()
```

**Example 2: `route-crawler.ts`**

```typescript
/**
 * Route Auto-Discovery Crawler
 *
 * Discovers testable routes by:
 * 1. Parsing migration artifacts (as_is_spec.md, migration_plan.md)
 * 2. Crawling the running application via BFS link traversal
 * 3. Merging and deduplicating results
 *
 * Usage: npx ts-node examples/route-crawler.ts [base-url] [max-depth]
 */

import { chromium, Browser, Page } from 'playwright'
import * as fs from 'fs'
import * as path from 'path'

interface RouteInfo {
  path: string
  type: 'static' | 'dynamic' | 'api'
  source: 'artifact' | 'crawl' | 'both'
  depth: number
  title: string | null
  status: number | null
}

interface CrawlConfig {
  baseUrl: string
  maxDepth: number
  artifactsDir: string
  timeout: number            // Per-page timeout (ms)
  excludePatterns: string[]  // URL patterns to skip
}

const DEFAULT_CRAWL_CONFIG: CrawlConfig = {
  baseUrl: 'http://localhost:3001',
  maxDepth: 3,
  artifactsDir: './.migrate-artifacts',
  timeout: 15000,
  excludePatterns: [
    '/api/',           // API endpoints (tested separately)
    '#',              // Anchor links
    'mailto:',
    'tel:',
    'javascript:',
    '/logout',        // Destructive actions
    '/delete'
  ]
}

// Phase 1: Parse routes from migration artifacts
function parseArtifactRoutes(artifactsDir: string): RouteInfo[] {
  const routes: RouteInfo[] = []

  // Parse as_is_spec.md for route information
  const specPath = path.join(artifactsDir, 'as_is_spec.md')
  if (fs.existsSync(specPath)) {
    const content = fs.readFileSync(specPath, 'utf-8')

    // Match route patterns: /path, /path/:param, /path/[param]
    const routePatterns = content.match(/(?:^|\s)(\/[\w\-\/\[\]:*]+)/gm) ?? []
    for (const match of routePatterns) {
      const routePath = match.trim()
      if (routePath.length > 1 && routePath.length < 100) {
        const isDynamic = routePath.includes(':') || routePath.includes('[')
        routes.push({
          path: routePath,
          type: isDynamic ? 'dynamic' : 'static',
          source: 'artifact',
          depth: 0,
          title: null,
          status: null
        })
      }
    }
  }

  // Parse migration_plan.md for additional routes
  const planPath = path.join(artifactsDir, 'migration_plan.md')
  if (fs.existsSync(planPath)) {
    const content = fs.readFileSync(planPath, 'utf-8')

    // Look for route tables or lists
    const routeMatches = content.match(/\|\s*(\/[\w\-\/]+)\s*\|/g) ?? []
    for (const match of routeMatches) {
      const routePath = match.replace(/\|/g, '').trim()
      if (!routes.some(r => r.path === routePath)) {
        routes.push({
          path: routePath,
          type: 'static',
          source: 'artifact',
          depth: 0,
          title: null,
          status: null
        })
      }
    }
  }

  return routes
}

// Phase 2: Crawl application via BFS
async function crawlRoutes(
  browser: Browser,
  config: CrawlConfig
): Promise<RouteInfo[]> {
  const routes: RouteInfo[] = []
  const visited = new Set<string>()
  const queue: { path: string; depth: number }[] = [{ path: '/', depth: 0 }]

  const context = await browser.newContext()
  const page = await context.newPage()

  while (queue.length > 0) {
    const { path: currentPath, depth } = queue.shift()!

    if (depth > config.maxDepth) continue
    if (visited.has(currentPath)) continue
    if (config.excludePatterns.some(p => currentPath.includes(p))) continue

    visited.add(currentPath)

    try {
      const response = await page.goto(config.baseUrl + currentPath, {
        waitUntil: 'networkidle',
        timeout: config.timeout
      })

      const status = response?.status() ?? 0
      const title = await page.title().catch(() => null)

      routes.push({
        path: currentPath,
        type: 'static',
        source: 'crawl',
        depth,
        title,
        status
      })

      // Skip further crawling if error page
      if (status >= 400) continue

      // Discover links on this page
      const links = await page.locator('a[href]').evaluateAll(elements =>
        elements.map(el => el.getAttribute('href')).filter(Boolean)
      )

      for (const href of links) {
        if (!href) continue
        const resolved = resolveLink(href, currentPath, config.baseUrl)
        if (resolved && !visited.has(resolved)) {
          queue.push({ path: resolved, depth: depth + 1 })
        }
      }
    } catch (error) {
      // Page failed to load, record but continue
      routes.push({
        path: currentPath,
        type: 'static',
        source: 'crawl',
        depth,
        title: null,
        status: null
      })
    }
  }

  await context.close()
  return routes
}

function resolveLink(href: string, currentPath: string, baseUrl: string): string | null {
  // Skip external links
  if (href.startsWith('http') && !href.startsWith(baseUrl)) return null
  // Skip special protocols
  if (/^(mailto|tel|javascript|#)/.test(href)) return null

  // Resolve relative paths
  if (href.startsWith('/')) return href
  if (href.startsWith('./')) return path.posix.join(currentPath, '..', href.substring(2))
  if (href.startsWith('../')) return path.posix.resolve(currentPath, '..', href)
  if (href.startsWith(baseUrl)) return href.substring(baseUrl.length)

  return path.posix.join(currentPath, '..', href)
}

// Phase 3: Merge artifact + crawl results
function mergeRoutes(artifactRoutes: RouteInfo[], crawledRoutes: RouteInfo[]): RouteInfo[] {
  const merged = new Map<string, RouteInfo>()

  // Add artifact routes first (priority source)
  for (const route of artifactRoutes) {
    merged.set(route.path, route)
  }

  // Merge crawled routes
  for (const route of crawledRoutes) {
    const existing = merged.get(route.path)
    if (existing) {
      existing.source = 'both'
      existing.status = route.status ?? existing.status
      existing.title = route.title ?? existing.title
    } else {
      merged.set(route.path, route)
    }
  }

  // Sort by depth then alphabetically
  return Array.from(merged.values())
    .sort((a, b) => a.depth - b.depth || a.path.localeCompare(b.path))
}

// Entry point
;(async () => {
  const config: CrawlConfig = {
    ...DEFAULT_CRAWL_CONFIG,
    baseUrl: process.argv[2] ?? DEFAULT_CRAWL_CONFIG.baseUrl,
    maxDepth: parseInt(process.argv[3] ?? String(DEFAULT_CRAWL_CONFIG.maxDepth))
  }

  console.log(`Route Discovery: ${config.baseUrl} (depth: ${config.maxDepth})`)
  console.log(`Artifacts: ${config.artifactsDir}\n`)

  // Step 1: Parse artifacts
  const artifactRoutes = parseArtifactRoutes(config.artifactsDir)
  console.log(`Artifact routes found: ${artifactRoutes.length}`)

  // Step 2: Crawl application
  const browser = await chromium.launch({ headless: true })
  const crawledRoutes = await crawlRoutes(browser, config)
  await browser.close()
  console.log(`Crawled routes found: ${crawledRoutes.length}`)

  // Step 3: Merge
  const allRoutes = mergeRoutes(artifactRoutes, crawledRoutes)
  console.log(`Total unique routes: ${allRoutes.length}\n`)

  // Output route registry
  console.log('Route Registry:')
  console.log('─'.repeat(80))
  console.log(`${'Path'.padEnd(40)} ${'Type'.padEnd(10)} ${'Source'.padEnd(10)} ${'Status'.padEnd(8)} Title`)
  console.log('─'.repeat(80))

  for (const route of allRoutes) {
    console.log(
      `${route.path.padEnd(40)} ${route.type.padEnd(10)} ${route.source.padEnd(10)} ${String(route.status ?? '-').padEnd(8)} ${route.title ?? '-'}`
    )
  }

  // Save route registry as JSON
  const outputPath = path.join(config.artifactsDir, 'route_registry.json')
  fs.mkdirSync(path.dirname(outputPath), { recursive: true })
  fs.writeFileSync(outputPath, JSON.stringify(allRoutes, null, 2))
  console.log(`\nRoute registry saved: ${outputPath}`)
})()
```

**Example 3: `cross-browser-verify.ts`**

```typescript
/**
 * Cross-Browser Verification Example
 *
 * Runs page load tests across Chromium, Firefox, and WebKit
 * and compares consistency of rendering and behavior.
 *
 * Usage: npx ts-node examples/cross-browser-verify.ts [target-url] [routes...]
 */

import { chromium, firefox, webkit, Browser, BrowserType } from 'playwright'
import * as fs from 'fs'

interface BrowserSpec {
  name: string
  type: BrowserType
  viewport: { width: number; height: number }
}

interface RouteResult {
  route: string
  browsers: {
    name: string
    status: number
    loadTime: number
    jsErrors: string[]
    contentLength: number
    screenshot: Buffer
  }[]
  consistency: {
    statusConsistent: boolean
    contentSimilar: boolean
    errorConsistent: boolean
    performanceConsistent: boolean
    overallScore: number
  }
}

const BROWSERS: BrowserSpec[] = [
  { name: 'Chromium', type: chromium, viewport: { width: 1920, height: 1080 } },
  { name: 'Firefox', type: firefox, viewport: { width: 1920, height: 1080 } },
  { name: 'WebKit', type: webkit, viewport: { width: 1920, height: 1080 } }
]

async function testRouteInBrowser(
  browserSpec: BrowserSpec,
  url: string,
  route: string
): Promise<RouteResult['browsers'][0]> {
  const browser = await browserSpec.type.launch({ headless: true })
  const context = await browser.newContext({ viewport: browserSpec.viewport })
  const page = await context.newPage()

  const jsErrors: string[] = []
  page.on('pageerror', err => jsErrors.push(err.message))
  page.on('console', msg => {
    if (msg.type() === 'error') jsErrors.push(msg.text())
  })

  const startTime = Date.now()
  const response = await page.goto(url + route, {
    waitUntil: 'networkidle',
    timeout: 30000
  }).catch(() => null)
  const loadTime = Date.now() - startTime

  const status = response?.status() ?? 0
  const content = await page.locator('body').innerText().catch(() => '')
  const screenshot = await page.screenshot({ fullPage: true, type: 'png' })

  await context.close()
  await browser.close()

  return {
    name: browserSpec.name,
    status,
    loadTime,
    jsErrors,
    contentLength: content.length,
    screenshot
  }
}

function analyzeConsistency(results: RouteResult['browsers']): RouteResult['consistency'] {
  let score = 100

  // Status consistency
  const statuses = new Set(results.map(r => r.status))
  const statusConsistent = statuses.size === 1
  if (!statusConsistent) score -= 30

  // Content similarity (within 10% length)
  const lengths = results.map(r => r.contentLength)
  const avgLength = lengths.reduce((a, b) => a + b, 0) / lengths.length
  const contentSimilar = lengths.every(l =>
    avgLength === 0 || Math.abs(l - avgLength) / avgLength < 0.1
  )
  if (!contentSimilar) score -= 20

  // Error consistency
  const errorCounts = results.map(r => r.jsErrors.length)
  const errorConsistent = new Set(errorCounts).size === 1
  if (!errorConsistent) score -= 15

  // Performance consistency (within 50% of each other)
  const times = results.map(r => r.loadTime)
  const avgTime = times.reduce((a, b) => a + b, 0) / times.length
  const performanceConsistent = times.every(t =>
    avgTime === 0 || Math.abs(t - avgTime) / avgTime < 0.5
  )
  if (!performanceConsistent) score -= 10

  return {
    statusConsistent,
    contentSimilar,
    errorConsistent,
    performanceConsistent,
    overallScore: Math.max(0, score)
  }
}

async function runCrossBrowserVerification(
  targetUrl: string,
  routes: string[]
): Promise<RouteResult[]> {
  const results: RouteResult[] = []

  for (const route of routes) {
    console.log(`Testing ${route}...`)

    // Run all browsers in parallel for this route
    const browserResults = await Promise.all(
      BROWSERS.map(spec => testRouteInBrowser(spec, targetUrl, route))
    )

    const consistency = analyzeConsistency(browserResults)

    results.push({ route, browsers: browserResults, consistency })

    // Print immediate feedback
    const status = consistency.overallScore >= 85 ? 'PASS'
      : consistency.overallScore >= 60 ? 'WARN' : 'FAIL'
    console.log(`  [${status}] Consistency: ${consistency.overallScore}%`)
    for (const br of browserResults) {
      console.log(`    ${br.name}: ${br.status} (${br.loadTime}ms, ${br.jsErrors.length} errors)`)
    }
  }

  return results
}

// Entry point
;(async () => {
  const targetUrl = process.argv[2] ?? 'http://localhost:3001'
  const routes = process.argv.slice(3).length > 0
    ? process.argv.slice(3)
    : ['/']

  console.log(`Cross-Browser Verification: ${targetUrl}`)
  console.log(`Routes: ${routes.join(', ')}`)
  console.log(`Browsers: ${BROWSERS.map(b => b.name).join(', ')}\n`)

  const results = await runCrossBrowserVerification(targetUrl, routes)

  // Summary
  console.log('\n' + '═'.repeat(60))
  console.log('SUMMARY')
  console.log('═'.repeat(60))

  const avgScore = results.reduce((sum, r) => sum + r.consistency.overallScore, 0) / results.length
  console.log(`Overall Consistency: ${avgScore.toFixed(1)}%`)

  const issues = results.filter(r => r.consistency.overallScore < 85)
  if (issues.length > 0) {
    console.log(`\nRoutes with issues (${issues.length}):`)
    for (const issue of issues) {
      console.log(`  ${issue.route}: ${issue.consistency.overallScore}%`)
      if (!issue.consistency.statusConsistent) console.log('    - HTTP status differs across browsers')
      if (!issue.consistency.contentSimilar) console.log('    - Content length varies significantly')
      if (!issue.consistency.errorConsistent) console.log('    - JS errors differ across browsers')
      if (!issue.consistency.performanceConsistent) console.log('    - Load time varies >50%')
    }
  }

  // Save results
  const outputPath = './verification-artifacts/cross-browser-results.json'
  fs.mkdirSync('./verification-artifacts', { recursive: true })
  fs.writeFileSync(outputPath, JSON.stringify(
    results.map(r => ({
      ...r,
      browsers: r.browsers.map(b => ({ ...b, screenshot: undefined }))  // Exclude binary
    })),
    null, 2
  ))
  console.log(`\nResults saved: ${outputPath}`)

  const allPassed = results.every(r => r.consistency.overallScore >= 85)
  process.exit(allPassed ? 0 : 1)
})()
```

### Step 8.4: Module Documentation Pattern

**Purpose**: Each module file follows a consistent structure for progressive disclosure loading.

**Module File Template (`modules/*.md`):**

```markdown
# Module: [Name]

## Quick Reference

| Item | Detail |
|------|--------|
| Phase | [N] |
| Activation | [flags] |
| Key Interfaces | [list] |
| Dependencies | [packages] |

## Pattern

[Core pattern/algorithm in pseudocode or TypeScript]

## Configuration

[YAML configuration snippet from .migrate-config.yaml]

## Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| [issue] | [cause] | [fix] |

## Integration Points

- **Input from**: [previous phase/module]
- **Output to**: [next phase/module]
- **Agent**: e2e-tester (migration mode)
```
