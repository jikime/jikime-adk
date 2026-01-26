---

## Phase 4: Behavioral Testing

**Activation**: Default (always runs), `--behavior`, or `--full` flag

### Overview

Behavioral testing verifies that the migrated application preserves the same functional behavior as the source. This phase tests page loads, navigation flows, form interactions, API calls, and JavaScript error collection across all discovered routes.

### Step 4.1: Page Load Verification

**Purpose**: Ensure every route loads successfully without errors.

**Verification Criteria:**

| Check | Threshold | Severity |
|-------|-----------|----------|
| HTTP Status | < 400 | Critical |
| Console Errors | 0 | High |
| Page Errors (uncaught exceptions) | 0 | Critical |
| Content Present | body text > 0 | High |
| Network Errors (4xx/5xx) | 0 critical resources | Medium |

**Page Load Test Engine:**

```typescript
interface PageLoadResult {
  route: string
  status: number
  loadTime: number
  consoleErrors: string[]
  pageErrors: string[]
  networkErrors: NetworkError[]
  hasContent: boolean
  passed: boolean
}

interface NetworkError {
  url: string
  status: number
  method: string
  resourceType: string
}

async function verifyPageLoad(
  page: Page,
  url: string,
  route: string
): Promise<PageLoadResult> {
  const consoleErrors: string[] = []
  const pageErrors: string[] = []
  const networkErrors: NetworkError[] = []

  // Collect errors
  page.on('console', msg => {
    if (msg.type() === 'error') consoleErrors.push(msg.text())
  })
  page.on('pageerror', err => pageErrors.push(err.message))
  page.on('response', response => {
    if (response.status() >= 400) {
      networkErrors.push({
        url: response.url(),
        status: response.status(),
        method: response.request().method(),
        resourceType: response.request().resourceType()
      })
    }
  })

  const startTime = Date.now()
  const response = await page.goto(url + route, {
    waitUntil: 'networkidle',
    timeout: 30000
  })
  const loadTime = Date.now() - startTime

  // Content check
  const bodyText = await page.locator('body').innerText().catch(() => '')
  const hasContent = bodyText.trim().length > 0

  const status = response?.status() ?? 0
  const passed = status < 400
    && pageErrors.length === 0
    && hasContent

  return {
    route, status, loadTime,
    consoleErrors, pageErrors, networkErrors,
    hasContent, passed
  }
}
```

**Batch Execution:**

```
FUNCTION verifyAllPages(route_registry, target_url):
  results = []

  FOR each route in route_registry.static_routes + route_registry.dynamic_routes:
    result = verifyPageLoad(page, target_url, route.path)
    results.add(result)

    // Reset page state between routes
    await page.context().clearCookies()

  RETURN results
```

### Step 4.2: Navigation Verification

**Purpose**: Verify internal navigation works correctly and detect broken links.

**Navigation Crawl Strategy:**

```typescript
interface NavigationResult {
  sourceUrl: string
  targetUrl: string
  linkText: string
  status: 'success' | 'broken' | 'redirect' | 'external'
  httpStatus: number | null
  errorMessage: string | null
}

async function verifyNavigation(
  page: Page,
  baseUrl: string,
  startRoute: string,
  maxDepth: number
): Promise<NavigationResult[]> {
  const results: NavigationResult[] = []
  const visited = new Set<string>()
  const queue: { url: string; depth: number }[] = [
    { url: startRoute, depth: 0 }
  ]

  while (queue.length > 0) {
    const { url, depth } = queue.shift()!
    if (depth > maxDepth || visited.has(url)) continue
    visited.add(url)

    await page.goto(baseUrl + url, { waitUntil: 'networkidle' })

    // Find all internal links
    const links = await page.locator('a[href]').all()

    for (const link of links) {
      const href = await link.getAttribute('href')
      const text = await link.innerText().catch(() => '')

      if (!href) continue

      // Skip external, anchor, and special links
      if (isExternalLink(href, baseUrl)) {
        results.push({
          sourceUrl: url, targetUrl: href, linkText: text,
          status: 'external', httpStatus: null, errorMessage: null
        })
        continue
      }

      const resolvedUrl = resolveUrl(href, url)

      // Verify link target loads
      try {
        const response = await page.goto(baseUrl + resolvedUrl, {
          waitUntil: 'networkidle',
          timeout: 15000
        })

        const status = response?.status() ?? 0
        results.push({
          sourceUrl: url, targetUrl: resolvedUrl, linkText: text,
          status: status < 400 ? 'success' : 'broken',
          httpStatus: status, errorMessage: null
        })

        // Add to crawl queue
        if (status < 400 && !visited.has(resolvedUrl)) {
          queue.push({ url: resolvedUrl, depth: depth + 1 })
        }
      } catch (error) {
        results.push({
          sourceUrl: url, targetUrl: resolvedUrl, linkText: text,
          status: 'broken', httpStatus: null,
          errorMessage: (error as Error).message
        })
      }
    }
  }

  return results
}
```

**Helper Functions:**

```
FUNCTION isExternalLink(href, baseUrl):
  IF href starts with "http" AND NOT starts with baseUrl:
    RETURN true
  IF href starts with "mailto:" OR "tel:" OR "javascript:" OR "#":
    RETURN true
  RETURN false

FUNCTION resolveUrl(href, currentPath):
  IF href starts with "/":
    RETURN href
  IF href starts with "./":
    RETURN currentPath + "/" + href.substring(2)
  IF href starts with "../":
    RETURN resolveRelative(currentPath, href)
  RETURN "/" + href
```

### Step 4.3: Form Interaction Verification

**Purpose**: Verify that forms function correctly in the migrated application.

**Form Detection & Testing:**

```typescript
interface FormTestResult {
  route: string
  formSelector: string
  formType: 'login' | 'search' | 'signup' | 'contact' | 'generic'
  fieldsDetected: string[]
  submitResult: 'success' | 'error' | 'no-response' | 'validation-shown'
  errorMessage: string | null
}

async function detectAndTestForms(
  page: Page,
  url: string,
  route: string
): Promise<FormTestResult[]> {
  await page.goto(url + route, { waitUntil: 'networkidle' })

  const forms = await page.locator('form').all()
  const results: FormTestResult[] = []

  for (const form of forms) {
    // Detect form type
    const formType = await detectFormType(form)

    // Get all input fields
    const fields = await form.locator('input, select, textarea').all()
    const fieldNames: string[] = []

    for (const field of fields) {
      const name = await field.getAttribute('name') ?? await field.getAttribute('id') ?? 'unnamed'
      const type = await field.getAttribute('type') ?? 'text'
      fieldNames.push(`${name}(${type})`)

      // Fill with test data based on type
      await fillFieldWithTestData(field, type, name)
    }

    // Attempt submit
    const submitResult = await attemptFormSubmit(page, form)

    results.push({
      route,
      formSelector: await getFormSelector(form),
      formType,
      fieldsDetected: fieldNames,
      submitResult,
      errorMessage: null
    })
  }

  return results
}
```

**Form Type Detection:**

```
FUNCTION detectFormType(form):
  html = await form.innerHTML()

  IF html contains "password" AND (html contains "email" OR html contains "login"):
    RETURN "login"
  IF html contains "password" AND html contains "confirm":
    RETURN "signup"
  IF html contains "search" OR form has role="search":
    RETURN "search"
  IF html contains "message" OR html contains "subject":
    RETURN "contact"
  RETURN "generic"
```

**Test Data Generation:**

```typescript
async function fillFieldWithTestData(
  field: Locator,
  type: string,
  name: string
): Promise<void> {
  const testData: Record<string, string> = {
    email: 'test@example.com',
    password: 'TestPass123!',
    text: 'Test Input',
    search: 'test query',
    tel: '+1234567890',
    url: 'https://example.com',
    number: '42',
    date: '2026-01-01'
  }

  // Name-based matching (higher priority)
  if (name.includes('email')) await field.fill('test@example.com')
  else if (name.includes('password')) await field.fill('TestPass123!')
  else if (name.includes('name')) await field.fill('Test User')
  else if (name.includes('phone')) await field.fill('+1234567890')
  // Type-based fallback
  else await field.fill(testData[type] ?? 'test value')
}
```

**Form Submit Verification:**

```
FUNCTION attemptFormSubmit(page, form):
  // Listen for navigation or network response
  responsePromise = page.waitForResponse(r => r.status() > 0, timeout: 5000)
  navigationPromise = page.waitForNavigation(timeout: 5000)

  // Try submit button first
  submitBtn = form.locator('button[type="submit"], input[type="submit"]')
  IF await submitBtn.count() > 0:
    await submitBtn.first().click()
  ELSE:
    // Try pressing Enter in last input
    lastInput = form.locator('input').last()
    await lastInput.press('Enter')

  // Wait for response (navigation or network)
  TRY:
    await Promise.race([responsePromise, navigationPromise])
    RETURN "success"
  CATCH timeout:
    // Check if validation errors appeared
    IF await page.locator('[class*="error"], [role="alert"], .invalid-feedback').count() > 0:
      RETURN "validation-shown"
    RETURN "no-response"
```

### Step 4.4: API Call Verification

**Purpose**: Monitor and verify API calls made by the application during navigation and interaction.

**API Monitoring Engine:**

```typescript
interface ApiCallResult {
  route: string
  endpoint: string
  method: string
  requestStatus: number
  responseTime: number
  requestBody: string | null
  responseBody: string | null
  matched: boolean  // true if source and target make same API calls
}

interface ApiMonitorConfig {
  capturePatterns: string[]   // URL patterns to monitor (e.g., "/api/*")
  ignorePatterns: string[]    // Patterns to skip (e.g., analytics, tracking)
  compareMode: 'strict' | 'relaxed'  // How to compare source vs target
}

async function monitorApiCalls(
  page: Page,
  url: string,
  route: string,
  config: ApiMonitorConfig
): Promise<ApiCallResult[]> {
  const apiCalls: ApiCallResult[] = []

  page.on('request', request => {
    const requestUrl = request.url()
    if (matchesPattern(requestUrl, config.capturePatterns)
        && !matchesPattern(requestUrl, config.ignorePatterns)) {
      // Track request start
      apiCalls.push({
        route,
        endpoint: new URL(requestUrl).pathname,
        method: request.method(),
        requestStatus: 0,  // Filled on response
        responseTime: 0,
        requestBody: request.postData() ?? null,
        responseBody: null,
        matched: false
      })
    }
  })

  page.on('response', async response => {
    const responseUrl = response.url()
    const existing = apiCalls.find(c =>
      c.endpoint === new URL(responseUrl).pathname && c.requestStatus === 0
    )
    if (existing) {
      existing.requestStatus = response.status()
      existing.responseTime = response.timing().responseEnd
      existing.responseBody = await response.text().catch(() => null)
    }
  })

  await page.goto(url + route, { waitUntil: 'networkidle' })

  return apiCalls
}
```

**Source vs Target API Comparison:**

```
FUNCTION compareApiCalls(sourceApis, targetApis, compareMode):
  results = []

  FOR each sourceApi in sourceApis:
    matchingTarget = targetApis.find(t =>
      t.endpoint === sourceApi.endpoint AND t.method === sourceApi.method
    )

    IF matchingTarget:
      IF compareMode === "strict":
        matched = sourceApi.requestStatus === matchingTarget.requestStatus
      ELSE:  // relaxed
        matched = (sourceApi.requestStatus < 400) === (matchingTarget.requestStatus < 400)

      results.add({
        endpoint: sourceApi.endpoint,
        method: sourceApi.method,
        sourceStatus: sourceApi.requestStatus,
        targetStatus: matchingTarget.requestStatus,
        matched: matched
      })
    ELSE:
      results.add({
        endpoint: sourceApi.endpoint,
        method: sourceApi.method,
        sourceStatus: sourceApi.requestStatus,
        targetStatus: null,
        matched: false,
        issue: "API call missing in target"
      })

  // Check for extra API calls in target
  FOR each targetApi in targetApis:
    IF NOT sourceApis.find(s => s.endpoint === targetApi.endpoint):
      results.add({
        endpoint: targetApi.endpoint,
        method: targetApi.method,
        sourceStatus: null,
        targetStatus: targetApi.requestStatus,
        matched: false,
        issue: "Extra API call in target"
      })

  RETURN results
```

**Default API Monitor Configuration:**

```yaml
api_monitor:
  capture_patterns:
    - "/api/*"
    - "/graphql"
    - "*/rest/*"
  ignore_patterns:
    - "*/analytics*"
    - "*/tracking*"
    - "*/hotjar*"
    - "*/sentry*"
    - "*google-analytics*"
    - "*facebook*"
  compare_mode: relaxed   # strict | relaxed
```

### Step 4.5: JavaScript Error Collection & Reporting

**Purpose**: Comprehensive collection and categorization of all JavaScript errors encountered during testing.

**Error Collection Engine:**

```typescript
interface JsErrorReport {
  route: string
  timestamp: string
  errors: CategorizedError[]
  summary: ErrorSummary
}

interface CategorizedError {
  message: string
  source: 'pageerror' | 'console.error' | 'unhandledrejection' | 'network'
  category: ErrorCategory
  severity: 'critical' | 'high' | 'medium' | 'low'
  stack: string | null
  count: number  // How many times this error occurred
}

type ErrorCategory =
  | 'runtime'          // TypeError, ReferenceError, etc.
  | 'network'          // Failed fetch, 404 resources
  | 'framework'        // React/Vue/Next.js specific errors
  | 'third-party'      // Errors from external scripts
  | 'deprecation'      // Deprecation warnings treated as errors
  | 'security'         // CSP violations, mixed content

interface ErrorSummary {
  totalErrors: number
  critical: number
  high: number
  medium: number
  low: number
  uniqueErrors: number
  topErrors: { message: string; count: number }[]
}
```

**Error Categorization Logic:**

```
FUNCTION categorizeError(error, source):
  // Category detection
  IF error.message contains "TypeError" OR "ReferenceError" OR "SyntaxError":
    category = "runtime"
    severity = "critical"

  ELIF error.message contains "Failed to fetch" OR "NetworkError" OR "ERR_":
    category = "network"
    severity = "medium"

  ELIF error.message contains "React" OR "Vue" OR "Next" OR "hydration":
    category = "framework"
    severity = "high"

  ELIF error.stack contains "node_modules" OR error.message contains "third-party":
    category = "third-party"
    severity = "low"

  ELIF error.message contains "deprecated" OR "will be removed":
    category = "deprecation"
    severity = "low"

  ELIF error.message contains "CSP" OR "mixed content" OR "blocked":
    category = "security"
    severity = "high"

  ELSE:
    category = "runtime"
    severity = "medium"

  RETURN { category, severity }
```

**Full Error Collection Setup:**

```typescript
function setupErrorCollection(page: Page, route: string): JsErrorReport {
  const errors: CategorizedError[] = []

  // Uncaught exceptions
  page.on('pageerror', err => {
    const { category, severity } = categorizeError(err, 'pageerror')
    addOrIncrementError(errors, {
      message: err.message,
      source: 'pageerror',
      category, severity,
      stack: err.stack ?? null,
      count: 1
    })
  })

  // Console errors
  page.on('console', msg => {
    if (msg.type() === 'error') {
      const { category, severity } = categorizeError(
        { message: msg.text() }, 'console.error'
      )
      addOrIncrementError(errors, {
        message: msg.text(),
        source: 'console.error',
        category, severity,
        stack: null,
        count: 1
      })
    }
  })

  // Unhandled promise rejections (captured as pageerror in Playwright)
  // Network errors for resources
  page.on('requestfailed', request => {
    addOrIncrementError(errors, {
      message: `Failed to load: ${request.url()} (${request.failure()?.errorText})`,
      source: 'network',
      category: 'network',
      severity: request.resourceType() === 'document' ? 'critical' : 'medium',
      stack: null,
      count: 1
    })
  })

  return { route, timestamp: new Date().toISOString(), errors, summary: null as any }
}

function addOrIncrementError(errors: CategorizedError[], newError: CategorizedError): void {
  const existing = errors.find(e => e.message === newError.message && e.source === newError.source)
  if (existing) {
    existing.count++
  } else {
    errors.push(newError)
  }
}
```

**Error Report Generation:**

```
FUNCTION generateErrorSummary(errors):
  summary = {
    totalErrors: sum(errors.map(e => e.count)),
    critical: errors.filter(e => e.severity === "critical").length,
    high: errors.filter(e => e.severity === "high").length,
    medium: errors.filter(e => e.severity === "medium").length,
    low: errors.filter(e => e.severity === "low").length,
    uniqueErrors: errors.length,
    topErrors: errors
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
      .map(e => ({ message: e.message, count: e.count }))
  }

  RETURN summary
```

### Step 4.6: Behavioral Comparison (Source vs Target)

**Purpose**: Compare behavior between source and target for the same routes.

**Comparison Engine:**

```typescript
interface BehaviorComparisonResult {
  route: string
  sourceLoadResult: PageLoadResult
  targetLoadResult: PageLoadResult
  contentMatch: boolean
  headingsMatch: boolean
  navigationMatch: boolean
  apiCallsMatch: boolean
  overallPassed: boolean
  differences: string[]
}

async function compareBehavior(
  browser: Browser,
  sourceUrl: string,
  targetUrl: string,
  route: string
): Promise<BehaviorComparisonResult> {
  const sourcePage = await browser.newPage()
  const targetPage = await browser.newPage()

  // Load both pages
  const sourceResult = await verifyPageLoad(sourcePage, sourceUrl, route)
  const targetResult = await verifyPageLoad(targetPage, targetUrl, route)

  const differences: string[] = []

  // Compare headings
  const sourceH1 = await sourcePage.locator('h1, h2, h3').allInnerTexts()
  const targetH1 = await targetPage.locator('h1, h2, h3').allInnerTexts()
  const headingsMatch = JSON.stringify(sourceH1) === JSON.stringify(targetH1)
  if (!headingsMatch) differences.push(`Headings differ: source=${sourceH1.length}, target=${targetH1.length}`)

  // Compare navigation
  const sourceNav = await sourcePage.locator('nav a, [role="navigation"] a').allInnerTexts()
  const targetNav = await targetPage.locator('nav a, [role="navigation"] a').allInnerTexts()
  const navigationMatch = JSON.stringify(sourceNav) === JSON.stringify(targetNav)
  if (!navigationMatch) differences.push(`Navigation differs: source=${sourceNav.length} links, target=${targetNav.length} links`)

  // Compare main content presence
  const sourceContent = await sourcePage.locator('main, [role="main"], #content, .content').innerText().catch(() => '')
  const targetContent = await targetPage.locator('main, [role="main"], #content, .content').innerText().catch(() => '')
  const contentMatch = sourceContent.length > 0 && targetContent.length > 0
    && Math.abs(sourceContent.length - targetContent.length) / Math.max(sourceContent.length, 1) < 0.3
  if (!contentMatch) differences.push(`Content length differs: source=${sourceContent.length}, target=${targetContent.length}`)

  await sourcePage.close()
  await targetPage.close()

  return {
    route,
    sourceLoadResult: sourceResult,
    targetLoadResult: targetResult,
    contentMatch,
    headingsMatch,
    navigationMatch,
    apiCallsMatch: true,  // Filled by Step 4.4
    overallPassed: headingsMatch && navigationMatch && contentMatch
      && targetResult.passed,
    differences
  }
}
```

### Behavioral Testing Output

```
Behavioral Testing Results
===========================

Page Load Verification:
  Total Routes: 25
  Passed: 24 (96%)
  Failed: 1

  Failed:
    /admin/settings - HTTP 500 (Internal Server Error)

Navigation Verification:
  Total Links: 156
  Working: 152 (97.4%)
  Broken: 3
  External: 1 (skipped)

  Broken Links:
    /products → /products/featured (404)
    /blog → /blog/archive (404)
    /docs → /docs/api/v1 (timeout)

Form Verification:
  Forms Detected: 8
  Tested: 8
  Working: 7 (87.5%)
  Issues: 1

  Issues:
    /contact (contact form) - no-response after submit

API Verification:
  API Calls Monitored: 45
  Matched (source = target): 42 (93.3%)
  Missing in Target: 2
  Extra in Target: 1

JavaScript Errors:
  Critical: 0
  High: 1 (hydration mismatch on /dashboard)
  Medium: 3 (network timeouts)
  Low: 5 (deprecation warnings)
```

