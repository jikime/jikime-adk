---

## Phase 5: Cross-Browser Verification

**Activation**: `--cross-browser` or `--full` flag

### Overview

Cross-browser verification ensures the migrated application renders and functions correctly across all major browser engines. Playwright natively supports Chromium, Firefox, and WebKit, enabling comprehensive cross-browser testing without additional infrastructure.

### Step 5.1: Multi-Browser Execution Engine

**Purpose**: Run all verification tests across multiple browser engines in parallel.

**Browser Configuration:**

```typescript
interface BrowserConfig {
  name: 'chromium' | 'firefox' | 'webkit'
  displayName: string
  viewport: { width: number; height: number }
  launchOptions: {
    headless: boolean
    args?: string[]
  }
}

const BROWSER_CONFIGS: BrowserConfig[] = [
  {
    name: 'chromium',
    displayName: 'Chromium (Chrome/Edge)',
    viewport: { width: 1920, height: 1080 },
    launchOptions: {
      headless: true,
      args: ['--disable-gpu', '--no-sandbox']
    }
  },
  {
    name: 'firefox',
    displayName: 'Firefox',
    viewport: { width: 1920, height: 1080 },
    launchOptions: { headless: true }
  },
  {
    name: 'webkit',
    displayName: 'WebKit (Safari)',
    viewport: { width: 1920, height: 1080 },
    launchOptions: { headless: true }
  }
]
```

**Parallel Browser Execution:**

```typescript
interface CrossBrowserResult {
  route: string
  results: BrowserTestResult[]
  consistencyScore: number  // 0-100: how consistent across browsers
  issues: CrossBrowserIssue[]
}

interface BrowserTestResult {
  browser: string
  pageLoad: PageLoadResult
  screenshot: Buffer
  jsErrors: CategorizedError[]
  renderTime: number
}

async function runCrossBrowserTests(
  route_registry: RouteRegistry,
  targetUrl: string,
  browsers: BrowserConfig[]
): Promise<CrossBrowserResult[]> {
  const results: CrossBrowserResult[] = []

  // Launch all browsers in parallel
  const browserInstances = await Promise.all(
    browsers.map(config =>
      playwright[config.name].launch(config.launchOptions)
    )
  )

  try {
    for (const route of route_registry.static_routes) {
      // Test route across all browsers in parallel
      const browserResults = await Promise.all(
        browserInstances.map(async (browser, index) => {
          const config = browsers[index]
          const context = await browser.newContext({
            viewport: config.viewport
          })
          const page = await context.newPage()

          // Collect errors
          const errors: CategorizedError[] = []
          setupErrorCollection(page, route.path)

          // Navigate and capture
          const startTime = Date.now()
          const loadResult = await verifyPageLoad(page, targetUrl, route.path)
          const renderTime = Date.now() - startTime

          // Screenshot for visual comparison
          const screenshot = await page.screenshot({
            fullPage: true, type: 'png'
          })

          await context.close()

          return {
            browser: config.displayName,
            pageLoad: loadResult,
            screenshot,
            jsErrors: errors,
            renderTime
          }
        })
      )

      // Analyze consistency
      const { consistencyScore, issues } = analyzeCrossBrowserConsistency(
        browserResults, route.path
      )

      results.push({
        route: route.path,
        results: browserResults,
        consistencyScore,
        issues
      })
    }
  } finally {
    // Close all browsers
    await Promise.all(browserInstances.map(b => b.close()))
  }

  return results
}
```

### Step 5.2: Cross-Browser Consistency Analysis

**Purpose**: Detect rendering and behavioral differences between browsers.

**Consistency Scoring:**

```typescript
interface CrossBrowserIssue {
  route: string
  type: 'visual' | 'functional' | 'performance' | 'error'
  severity: 'critical' | 'high' | 'medium' | 'low'
  browsers: string[]          // Affected browsers
  description: string
  evidence: string | null     // Screenshot path or error message
}

function analyzeCrossBrowserConsistency(
  results: BrowserTestResult[],
  route: string
): { consistencyScore: number; issues: CrossBrowserIssue[] } {
  const issues: CrossBrowserIssue[] = []
  let score = 100

  // 1. Page load status consistency
  const statuses = results.map(r => r.pageLoad.status)
  if (new Set(statuses).size > 1) {
    score -= 30
    issues.push({
      route, type: 'functional', severity: 'critical',
      browsers: results.filter(r => r.pageLoad.status >= 400).map(r => r.browser),
      description: `HTTP status differs: ${results.map(r => `${r.browser}=${r.pageLoad.status}`).join(', ')}`,
      evidence: null
    })
  }

  // 2. JavaScript error consistency
  const errorCounts = results.map(r => r.jsErrors.filter(e => e.severity === 'critical').length)
  const maxErrors = Math.max(...errorCounts)
  const minErrors = Math.min(...errorCounts)
  if (maxErrors !== minErrors) {
    score -= 15
    const affectedBrowsers = results
      .filter(r => r.jsErrors.filter(e => e.severity === 'critical').length > minErrors)
      .map(r => r.browser)
    issues.push({
      route, type: 'error', severity: 'high',
      browsers: affectedBrowsers,
      description: `Browser-specific JS errors detected (${affectedBrowsers.join(', ')})`,
      evidence: null
    })
  }

  // 3. Visual consistency (screenshot comparison between browsers)
  if (results.length >= 2) {
    const baseScreenshot = results[0].screenshot
    for (let i = 1; i < results.length; i++) {
      const comparison = compareScreenshots(baseScreenshot, results[i].screenshot, 8)
      if (comparison.diffPercentage > 8) {
        score -= 10
        issues.push({
          route, type: 'visual', severity: 'medium',
          browsers: [results[0].browser, results[i].browser],
          description: `Visual diff ${comparison.diffPercentage.toFixed(1)}% between ${results[0].browser} and ${results[i].browser}`,
          evidence: comparison.diffImagePath
        })
      }
    }
  }

  // 4. Performance consistency
  const renderTimes = results.map(r => r.renderTime)
  const avgTime = renderTimes.reduce((a, b) => a + b, 0) / renderTimes.length
  const maxDeviation = Math.max(...renderTimes.map(t => Math.abs(t - avgTime) / avgTime))
  if (maxDeviation > 0.5) {  // >50% deviation
    score -= 5
    const slowBrowser = results[renderTimes.indexOf(Math.max(...renderTimes))].browser
    issues.push({
      route, type: 'performance', severity: 'low',
      browsers: [slowBrowser],
      description: `Render time deviation >50%: ${results.map(r => `${r.browser}=${r.renderTime}ms`).join(', ')}`,
      evidence: null
    })
  }

  return { consistencyScore: Math.max(0, score), issues }
}
```

### Step 5.3: Mobile Device Emulation

**Purpose**: Verify migrated application on mobile device profiles using Playwright's built-in device emulation.

**Device Profiles:**

```typescript
interface DeviceProfile {
  name: string
  viewport: { width: number; height: number }
  userAgent: string
  deviceScaleFactor: number
  isMobile: boolean
  hasTouch: boolean
  defaultBrowserType: 'chromium' | 'webkit'
}

const MOBILE_DEVICES: DeviceProfile[] = [
  {
    name: 'iPhone 14 Pro',
    viewport: { width: 393, height: 852 },
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15',
    deviceScaleFactor: 3,
    isMobile: true,
    hasTouch: true,
    defaultBrowserType: 'webkit'
  },
  {
    name: 'iPhone SE',
    viewport: { width: 375, height: 667 },
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15',
    deviceScaleFactor: 2,
    isMobile: true,
    hasTouch: true,
    defaultBrowserType: 'webkit'
  },
  {
    name: 'Pixel 7',
    viewport: { width: 412, height: 915 },
    userAgent: 'Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 Chrome/120.0',
    deviceScaleFactor: 2.625,
    isMobile: true,
    hasTouch: true,
    defaultBrowserType: 'chromium'
  },
  {
    name: 'Galaxy S23',
    viewport: { width: 360, height: 780 },
    userAgent: 'Mozilla/5.0 (Linux; Android 14; SM-S911B) AppleWebKit/537.36 Chrome/120.0',
    deviceScaleFactor: 3,
    isMobile: true,
    hasTouch: true,
    defaultBrowserType: 'chromium'
  },
  {
    name: 'iPad Pro 12.9',
    viewport: { width: 1024, height: 1366 },
    userAgent: 'Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15',
    deviceScaleFactor: 2,
    isMobile: true,
    hasTouch: true,
    defaultBrowserType: 'webkit'
  },
  {
    name: 'iPad Mini',
    viewport: { width: 768, height: 1024 },
    userAgent: 'Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15',
    deviceScaleFactor: 2,
    isMobile: true,
    hasTouch: true,
    defaultBrowserType: 'webkit'
  }
]
```

**Mobile Emulation Test:**

```typescript
interface MobileTestResult {
  device: string
  browser: string
  route: string
  pageLoad: PageLoadResult
  mobileSpecificChecks: MobileCheck[]
  passed: boolean
}

interface MobileCheck {
  name: string
  passed: boolean
  details: string
}

async function runMobileEmulationTests(
  route_registry: RouteRegistry,
  targetUrl: string,
  devices: DeviceProfile[]
): Promise<MobileTestResult[]> {
  const results: MobileTestResult[] = []

  for (const device of devices) {
    const browser = await playwright[device.defaultBrowserType].launch({ headless: true })
    const context = await browser.newContext({
      viewport: device.viewport,
      userAgent: device.userAgent,
      deviceScaleFactor: device.deviceScaleFactor,
      isMobile: device.isMobile,
      hasTouch: device.hasTouch
    })

    for (const route of route_registry.static_routes) {
      const page = await context.newPage()
      const loadResult = await verifyPageLoad(page, targetUrl, route.path)

      // Mobile-specific checks
      const mobileChecks = await runMobileSpecificChecks(page, device)

      results.push({
        device: device.name,
        browser: device.defaultBrowserType,
        route: route.path,
        pageLoad: loadResult,
        mobileSpecificChecks: mobileChecks,
        passed: loadResult.passed && mobileChecks.every(c => c.passed)
      })

      await page.close()
    }

    await context.close()
    await browser.close()
  }

  return results
}
```

**Mobile-Specific Checks:**

```typescript
async function runMobileSpecificChecks(
  page: Page,
  device: DeviceProfile
): Promise<MobileCheck[]> {
  const checks: MobileCheck[] = []

  // 1. Viewport meta tag
  const viewportMeta = await page.locator('meta[name="viewport"]').getAttribute('content')
  checks.push({
    name: 'Viewport Meta',
    passed: viewportMeta !== null && viewportMeta.includes('width=device-width'),
    details: viewportMeta ?? 'Missing viewport meta tag'
  })

  // 2. No horizontal overflow
  const hasOverflow = await page.evaluate(() => {
    return document.documentElement.scrollWidth > document.documentElement.clientWidth
  })
  checks.push({
    name: 'No Horizontal Overflow',
    passed: !hasOverflow,
    details: hasOverflow
      ? `Page width ${await page.evaluate(() => document.documentElement.scrollWidth)}px exceeds viewport`
      : 'OK'
  })

  // 3. Touch target sizes (minimum 44x44px per WCAG)
  const smallTargets = await page.evaluate(() => {
    const interactives = document.querySelectorAll('a, button, input, select, textarea, [role="button"]')
    const small: string[] = []
    interactives.forEach(el => {
      const rect = el.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0 && (rect.width < 44 || rect.height < 44)) {
        small.push(`${el.tagName.toLowerCase()}(${Math.round(rect.width)}x${Math.round(rect.height)})`)
      }
    })
    return small.slice(0, 10) // Limit to top 10
  })
  checks.push({
    name: 'Touch Targets (44x44px min)',
    passed: smallTargets.length === 0,
    details: smallTargets.length > 0
      ? `${smallTargets.length} small targets: ${smallTargets.join(', ')}`
      : 'All targets meet minimum size'
  })

  // 4. Font size readability (minimum 12px)
  const smallFonts = await page.evaluate(() => {
    const textElements = document.querySelectorAll('p, span, li, td, th, label, a')
    const small: string[] = []
    textElements.forEach(el => {
      const fontSize = parseFloat(window.getComputedStyle(el).fontSize)
      if (fontSize > 0 && fontSize < 12) {
        const text = (el.textContent ?? '').substring(0, 20)
        small.push(`"${text}" (${fontSize}px)`)
      }
    })
    return small.slice(0, 5)
  })
  checks.push({
    name: 'Font Readability (12px min)',
    passed: smallFonts.length === 0,
    details: smallFonts.length > 0
      ? `${smallFonts.length} elements with small font: ${smallFonts.join(', ')}`
      : 'All text meets minimum size'
  })

  // 5. Fixed/Sticky elements not overlapping content
  const fixedOverlap = await page.evaluate(() => {
    const fixedElements = Array.from(document.querySelectorAll('*')).filter(el => {
      const style = window.getComputedStyle(el)
      return style.position === 'fixed' || style.position === 'sticky'
    })
    return fixedElements.length > 3 // Warning if too many fixed elements on mobile
  })
  checks.push({
    name: 'Fixed Elements',
    passed: !fixedOverlap,
    details: fixedOverlap
      ? 'Multiple fixed/sticky elements detected (may overlap content on mobile)'
      : 'OK'
  })

  return checks
}
```

### Step 5.4: Cross-Browser Report Generation

**Purpose**: Generate comprehensive cross-browser compatibility report.

**Report Structure:**

```
Cross-Browser Verification Report
===================================

Browsers Tested: Chromium, Firefox, WebKit
Mobile Devices: iPhone 14 Pro, Pixel 7, iPad Pro 12.9
Routes Tested: 25

Desktop Browser Results:
| Route | Chromium | Firefox | WebKit | Consistency |
|-------|----------|---------|--------|-------------|
| / | PASS | PASS | PASS | 100% |
| /dashboard | PASS | PASS | WARN | 85% |
| /settings | PASS | PASS | PASS | 100% |
| /profile | PASS | WARN | PASS | 90% |

Mobile Device Results:
| Route | iPhone 14 | Pixel 7 | iPad Pro | Issues |
|-------|-----------|---------|----------|--------|
| / | PASS | PASS | PASS | 0 |
| /dashboard | PASS | WARN | PASS | 1 (overflow) |
| /settings | PASS | PASS | PASS | 0 |

Cross-Browser Issues:
1. [MEDIUM] /dashboard: Visual diff 9.2% between Chromium and WebKit
   → Likely cause: CSS Grid rendering difference
2. [LOW] /profile: Render time deviation >50% (WebKit slower)
   → Firefox: 450ms, Chromium: 380ms, WebKit: 720ms

Mobile-Specific Issues:
1. [HIGH] /dashboard (Pixel 7): Horizontal overflow detected
2. [MEDIUM] /settings: 3 touch targets below 44x44px minimum

Overall Consistency Score: 92/100
```

**Report Data Structure:**

```typescript
interface CrossBrowserReport {
  summary: {
    browsersCount: number
    devicesCount: number
    routesCount: number
    overallConsistency: number
    passRate: { desktop: number; mobile: number }
  }
  desktopResults: {
    route: string
    browserResults: { browser: string; status: 'pass' | 'warn' | 'fail' }[]
    consistency: number
  }[]
  mobileResults: {
    route: string
    deviceResults: { device: string; status: 'pass' | 'warn' | 'fail'; issues: number }[]
  }[]
  issues: CrossBrowserIssue[]
  mobileIssues: MobileCheck[]
}
```

### Step 5.5: Browser-Specific Known Issues

**Purpose**: Filter known browser-specific differences that are acceptable.

**Known Differences Registry:**

```yaml
known_browser_differences:
  webkit:
    - pattern: "font-smoothing"
      description: "WebKit uses different font anti-aliasing"
      severity: ignore
    - pattern: "scrollbar-width"
      description: "WebKit shows overlay scrollbars"
      severity: ignore
    - pattern: "date-input"
      description: "Date picker rendering differs"
      severity: low

  firefox:
    - pattern: "focus-ring"
      description: "Firefox shows different focus outline"
      severity: ignore
    - pattern: "select-dropdown"
      description: "Native select rendering differs"
      severity: ignore

  chromium:
    - pattern: "backdrop-filter"
      description: "Chromium may show different blur quality"
      severity: ignore
```

**Known Issues Filter:**

```
FUNCTION filterKnownIssues(issues, knownDifferences):
  filtered = []
  FOR each issue in issues:
    isKnown = knownDifferences[issue.browsers[0]]?.find(
      known => issue.description.contains(known.pattern)
    )
    IF isKnown AND isKnown.severity === "ignore":
      CONTINUE  // Skip known acceptable difference
    ELSE:
      filtered.add(issue)

  RETURN filtered
```

### Cross-Browser Execution Mode

```
IF --full:
  → All 3 desktop browsers + All 6 mobile devices
ELIF --cross-browser:
  → All 3 desktop browsers + 2 mobile devices (iPhone 14, Pixel 7)
ELSE:
  → Chromium only (default, fastest)
```

