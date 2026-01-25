---
description: "[Step 4/4] Migration verification. Start dev servers, run Playwright E2E tests, compare behavior, generate final report."
argument-hint: '[--full] [--behavior] [--e2e] [--visual] [--performance] [--cross-browser] [--a11y] [--source-url URL] [--target-url URL] [--port N] [--source-port N] [--headless] [--threshold N] [--depth N]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

# Migration Step 4: Verify

**Verification Phase**: Validate migration success through automated Playwright-based testing.

[SOFT] Apply --ultrathink keyword for deep verification strategy analysis
WHY: Migration verification requires systematic planning of test execution order, server lifecycle management, and multi-dimensional quality assessment
IMPACT: Sequential thinking ensures comprehensive verification coverage with behavioral preservation validation

## CRITICAL: Input Sources

**Project settings are automatically read from `.migrate-config.yaml`.**

### Required Inputs (from Previous Steps)

1. **`.migrate-config.yaml`** - artifacts_dir, output_dir, source/target framework, verification settings
2. **`{artifacts_dir}/progress.yaml`** - Migration progress status (Step 3 output)
3. **`{output_dir}/`** - Migrated project (Step 3 output)

### Optional Inputs

4. **`{artifacts_dir}/as_is_spec.md`** - Route information for auto-discovery
5. **`{artifacts_dir}/migration_plan.md`** - User flow information for E2E scenarios

## What This Command Does

1. **Dev Server Setup** - Start source and target dev servers automatically
2. **Route Discovery** - Auto-discover testable routes from migration artifacts
3. **Characterization Tests** - Run behavior preservation tests
4. **Behavior Comparison** - Compare source/target outputs
5. **E2E Testing** - Validate full user flows with Playwright
6. **Visual Regression** - Screenshot comparison (source vs target)
7. **Performance Check** - Core Web Vitals and load time comparison
8. **Cross-Browser** - Chromium, Firefox, WebKit validation
9. **Accessibility** - WCAG compliance check with axe-core
10. **Final Report** - Comprehensive verification report with visual evidence

## Usage

```bash
# Verify current migration (reads all from config)
/jikime:migrate-4-verify

# Verify with all checks (visual + cross-browser + a11y + performance)
/jikime:migrate-4-verify --full

# Verify specific aspects
/jikime:migrate-4-verify --behavior
/jikime:migrate-4-verify --e2e
/jikime:migrate-4-verify --visual
/jikime:migrate-4-verify --performance
/jikime:migrate-4-verify --cross-browser
/jikime:migrate-4-verify --a11y

# Custom ports for dev servers
/jikime:migrate-4-verify --source-port 3000 --port 3001

# Compare live systems (skip dev server startup)
/jikime:migrate-4-verify --source-url http://old.local:3000 --target-url http://new.local:3001

# Visual regression with custom threshold
/jikime:migrate-4-verify --visual --threshold 3

# Full verification with custom depth
/jikime:migrate-4-verify --full --depth 5

# Capture migration patterns as a reusable skill
/jikime:migrate-4-verify --capture-skill
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--full` | Run ALL verification types | false |
| `--behavior` | Behavior comparison only | false |
| `--e2e` | E2E Playwright tests only | false |
| `--visual` | Visual regression (screenshot comparison) | false |
| `--performance` | Performance metrics comparison | false |
| `--cross-browser` | Cross-browser verification (Chromium, Firefox, WebKit) | false |
| `--a11y` | Accessibility (axe-core) checks | false |
| `--source-url` | Source system URL (skip source server startup) | auto |
| `--target-url` | Target system URL (skip target server startup) | auto |
| `--port` | Target dev server port | 3001 |
| `--source-port` | Source dev server port | 3000 |
| `--headless` | Run browsers in headless mode | true |
| `--threshold` | Visual diff threshold percentage | 5 |
| `--depth` | Navigation crawl depth for route discovery | 3 |
| `--capture-skill` | Generate migration skill from verified patterns | false |

**Note**: `--source-url` and `--target-url` are for comparing **already running instances**. When not provided, the command automatically starts dev servers.

---

## --capture-skill Option

After successful verification, this option captures migration patterns and creates a reusable skill for similar future migrations.

### Prerequisites

- Verification must pass (at least `--behavior` or `--full`)
- Migration artifacts must exist:
  - `{artifacts_dir}/as_is_spec.md` - Source analysis
  - `{artifacts_dir}/migration_plan.md` - Transformation rules
  - `{artifacts_dir}/progress.yaml` - Actual migration history

### Workflow

```
Step 1: Analyze Artifacts
  → Read as_is_spec.md (source patterns)
  → Read migration_plan.md (transformation rules)
  → Read progress.yaml (actual transformations applied)

Step 2: Extract Patterns
  → Identify recurring transformation patterns
  → Capture special case solutions
  → Document framework-specific mapping rules

Step 3: Invoke skill-builder Agent
  → Task(subagent_type="skill-builder", prompt="
      Create migration skill from verified patterns:
      - Source Framework: {source_framework}
      - Target Framework: {target_framework}
      - Artifacts Directory: {artifacts_dir}
      - Key Patterns: {extracted_patterns}
    ")

Step 4: Generate Skill Draft
  → Output: skills/jikime-migration-{source}-to-{target}/SKILL.md
  → Follows Progressive Disclosure format (Level 1/2/3)

Step 5: Request User Review
  → Display generated skill summary
  → Ask user to review and refine before finalizing
```

### Generated Skill Structure

```
skills/jikime-migration-{source}-to-{target}/
├── SKILL.md                 # Main skill definition (Level 1-2)
├── reference.md             # Detailed patterns (Level 3)
├── modules/
│   ├── components.md        # Component transformation rules
│   ├── routing.md           # Routing migration patterns
│   └── state.md             # State management patterns
└── examples/
    └── common-cases.md      # Real examples from this migration
```

### Example Usage

```bash
# After successful verification
/jikime:migrate-4-verify --full

# Capture patterns as reusable skill
/jikime:migrate-4-verify --capture-skill

# Combined: verify and capture in one command
/jikime:migrate-4-verify --full --capture-skill
```

### Generated Skill Frontmatter

```yaml
---
name: jikime-migration-{source}-to-{target}
description: Migration patterns from {source} to {target}
version: 1.0.0

progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000

triggers:
  keywords: ["{source}", "{target}", "migration", "convert"]
  phases: ["plan", "run"]
  agents: ["manager-ddd", "refactorer"]

metadata:
  source_framework: "{source}"
  target_framework: "{target}"
  generated_from: "{project_name}"
  generation_date: "{date}"
---
```

---

## Phase 1: Dev Server Lifecycle Management

### Overview

Before running any Playwright tests, the command MUST ensure both source and target applications are running and accessible.

### Step 1.1: Read Configuration

```
Read .migrate-config.yaml to extract:
- source_dir: Original project path
- output_dir: Migrated project path
- source_framework: Source framework name
- target_framework: Target framework name
- verification.dev_command: Override dev command (optional)
- verification.source_port: Source port (default: 3000)
- verification.target_port: Target port (default: 3001)
```

### Step 1.2: Framework Dev Command Detection

Auto-detect the dev command based on framework and package.json:

**Detection Priority:**
1. `verification.dev_command` in `.migrate-config.yaml` (explicit override)
2. `package.json` scripts → `dev` or `start` script
3. Framework-based fallback mapping

**Framework Command Mapping:**

| Framework | Dev Command | Port Flag |
|-----------|------------|-----------|
| Next.js | `npx next dev` | `--port {port}` |
| Vite (React/Vue/Svelte) | `npx vite` | `--port {port}` |
| Create React App | `npx react-scripts start` | `PORT={port}` (env) |
| Nuxt | `npx nuxt dev` | `--port {port}` |
| Angular | `npx ng serve` | `--port {port}` |
| Remix | `npx remix dev` | `--port {port}` |
| Gatsby | `npx gatsby develop` | `-p {port}` |
| Astro | `npx astro dev` | `--port {port}` |
| SvelteKit | `npx vite dev` | `--port {port}` |
| Express/Fastify | `node server.js` | `PORT={port}` (env) |
| Django | `python manage.py runserver` | `0.0.0.0:{port}` |
| Flask | `flask run` | `--port {port}` |
| Spring Boot | `./mvnw spring-boot:run` | `-Dserver.port={port}` |
| Rails | `rails server` | `-p {port}` |
| Laravel | `php artisan serve` | `--port={port}` |
| Go (Fiber/Gin) | `go run .` | `PORT={port}` (env) |

**Detection Algorithm:**

```
1. IF verification.dev_command exists in config:
     → Use it directly (user override)

2. ELIF package.json exists:
     → Read scripts.dev or scripts.start
     → Extract base command
     → Append port flag based on framework detection

3. ELIF requirements.txt / manage.py exists:
     → Use Django/Flask command

4. ELIF go.mod exists:
     → Use Go run command

5. ELIF pom.xml / build.gradle exists:
     → Use Spring Boot / Gradle command

6. ELSE:
     → Ask user via AskUserQuestion for dev command
```

### Step 1.3: Install Dependencies

Before starting servers, ensure dependencies are installed:

```
For each project (source_dir, output_dir):
  IF package.json exists AND node_modules/ missing:
    → Run: npm install (or pnpm install / yarn install based on lockfile)
  ELIF requirements.txt exists AND venv/ missing:
    → Run: pip install -r requirements.txt
  ELIF go.mod exists:
    → Run: go mod download
```

**Lockfile Detection:**
- `pnpm-lock.yaml` → `pnpm install`
- `yarn.lock` → `yarn install`
- `package-lock.json` → `npm install`
- `bun.lockb` → `bun install`

### Step 1.4: Start Dev Servers

**Startup Sequence:**

```
Step 1: Start SOURCE server (if --source-url not provided)
  → cd {source_dir}
  → Run dev command with --source-port (background process)
  → Record PID for cleanup

Step 2: Start TARGET server (if --target-url not provided)
  → cd {output_dir}
  → Run dev command with --port (background process)
  → Record PID for cleanup

Step 3: Health check both servers
  → Wait for both to be accessible
  → Abort if either fails
```

**Background Process Management:**

```bash
# Start server in background, capture PID
{dev_command} > /tmp/jikime-verify-{role}.log 2>&1 &
SERVER_PID=$!

# Store PIDs for cleanup
echo $SERVER_PID >> /tmp/jikime-verify-pids.txt
```

### Step 1.5: Health Check

**Health Check Algorithm:**

```
FUNCTION healthCheck(url, maxWaitSeconds=30, intervalMs=1000):
  startTime = now()

  WHILE (now() - startTime) < maxWaitSeconds:
    TRY:
      response = HTTP_GET(url)
      IF response.status < 500:
        RETURN SUCCESS
    CATCH:
      WAIT(intervalMs)

  RETURN TIMEOUT_ERROR
```

**Health Check Targets:**
- Source: `http://localhost:{source_port}` (or `--source-url`)
- Target: `http://localhost:{port}` (or `--target-url`)

**Failure Handling:**
- If source fails to start: Warn user, continue with target-only verification
- If target fails to start: ABORT with error (target is required)
- Log server output to `/tmp/jikime-verify-{role}.log` for debugging

### Step 1.6: Server Cleanup (Post-Verification)

**Cleanup is MANDATORY after all tests complete (success or failure):**

```
FINALLY:
  Read PIDs from /tmp/jikime-verify-pids.txt
  FOR each PID:
    kill -TERM {PID}
    Wait 5s for graceful shutdown
    IF still running: kill -9 {PID}
  Remove /tmp/jikime-verify-*.log
  Remove /tmp/jikime-verify-pids.txt
```

### Step 1.7: URL Resolution

After servers are running, resolve final URLs:

```
source_url = --source-url OR "http://localhost:{source_port}"
target_url = --target-url OR "http://localhost:{port}"
```

These URLs are passed to all subsequent verification phases.

---

## Phase 2: Route Auto-Discovery & Test Generation

### Overview

Before running tests, the command discovers all testable routes from migration artifacts and source code. This enables automatic verification without manual route specification.

### Step 2.1: Route Discovery Engine

**Discovery Sources (Priority Order):**

```
1. Manual Override (highest priority)
   → verification.test_routes in .migrate-config.yaml
   → User explicitly defined routes

2. Migration Artifacts
   → {artifacts_dir}/as_is_spec.md - "Routes" or "Pages" or "Endpoints" section
   → {artifacts_dir}/migration_plan.md - Referenced routes in plan

3. File-Based Routing (source_dir)
   → pages/**/*.{tsx,jsx,vue,svelte} (Next.js, Nuxt, SvelteKit)
   → app/**/page.{tsx,jsx} (Next.js App Router)
   → src/routes/**/*.{tsx,svelte} (Remix, SvelteKit)
   → src/app/**/*.component.ts (Angular)

4. Router Configuration Files
   → react-router: src/routes.{tsx,jsx}, src/App.{tsx,jsx} (Route/Routes components)
   → vue-router: src/router/index.{ts,js} (routes array)
   → angular: app-routing.module.ts (routes array)
   → express: routes/*.{ts,js}, app.{ts,js} (app.get/post/use patterns)

5. Navigation Crawl (runtime, if servers running)
   → Start at root (/)
   → Find all <a href="..."> with internal links
   → Crawl up to --depth levels (default: 3)
   → Deduplicate discovered routes
```

**Route Discovery Algorithm:**

```
FUNCTION discoverRoutes(config, artifacts_dir, source_dir):
  routes = []

  // Priority 1: Manual override
  IF config.verification.test_routes is not empty:
    routes = config.verification.test_routes
    RETURN routes (skip auto-discovery)

  // Priority 2: Migration artifacts
  spec_routes = parseRoutesFromSpec(artifacts_dir + "/as_is_spec.md")
  routes.addAll(spec_routes)

  // Priority 3: File-based routing
  file_routes = scanFileBasedRoutes(source_dir, config.source_framework)
  routes.addAll(file_routes)

  // Priority 4: Router config
  IF routes.isEmpty():
    config_routes = parseRouterConfig(source_dir, config.source_framework)
    routes.addAll(config_routes)

  // Priority 5: Navigation crawl (fallback)
  IF routes.isEmpty() AND servers_running:
    crawl_routes = crawlNavigation(source_url, config.verification.crawl_depth)
    routes.addAll(crawl_routes)

  // Deduplicate and sort
  routes = deduplicate(routes)
  routes = sort(routes, by: "path")

  RETURN routes
```

**File-Based Route Patterns:**

| Framework | Directory | Pattern | Route Mapping |
|-----------|-----------|---------|---------------|
| Next.js (Pages) | `pages/` | `pages/about.tsx` | `/about` |
| Next.js (App) | `app/` | `app/dashboard/page.tsx` | `/dashboard` |
| Nuxt | `pages/` | `pages/users/index.vue` | `/users` |
| SvelteKit | `src/routes/` | `src/routes/blog/+page.svelte` | `/blog` |
| Remix | `app/routes/` | `app/routes/products.tsx` | `/products` |
| Angular | `src/app/` | Route config in module | Parsed from config |

### Step 2.2: Dynamic Route Handling

Dynamic routes (parameterized paths) need sample data to become testable URLs.

**Dynamic Route Patterns:**

| Framework | Pattern | Example |
|-----------|---------|---------|
| Next.js | `[param]` | `pages/users/[id].tsx` → `/users/:id` |
| Next.js | `[...slug]` | `pages/docs/[...slug].tsx` → `/docs/:slug*` |
| Nuxt | `[param]` | `pages/posts/[id].vue` → `/posts/:id` |
| SvelteKit | `[param]` | `routes/items/[id]/+page.svelte` → `/items/:id` |
| React Router | `:param` | `path="/users/:id"` → `/users/:id` |
| Angular | `:param` | `path: 'users/:id'` → `/users/:id` |
| Express | `:param` | `app.get('/api/users/:id')` → `/api/users/:id` |

**Sample Data Generation:**

```
FUNCTION generateSampleUrls(dynamic_routes):
  FOR each route:
    IF route contains :id or [id]:
      → Generate: route.replace(":id", "1")
      → Also try: route.replace(":id", "test-item")

    IF route contains :slug or [...slug]:
      → Generate: route.replace(":slug", "example-page")

    IF route contains :category:
      → Generate: route.replace(":category", "general")

  RETURN sampleUrls
```

**Exclusion Rules (skip these routes):**

```yaml
exclude_patterns:
  - "/api/*"           # API routes (tested separately in behavior comparison)
  - "/_next/*"         # Framework internals
  - "/_nuxt/*"         # Framework internals
  - "/favicon.ico"     # Static assets
  - "/*.xml"           # Sitemaps, RSS
  - "/*.json"          # JSON endpoints
  - "/admin/*"         # Admin routes (require auth, test separately)
```

### Step 2.3: Test Case Auto-Generation

For each discovered route, generate baseline verification tests.

**Test Categories:**

| Category | What It Checks | Severity |
|----------|---------------|----------|
| Page Load | HTTP 200, no crash | Critical |
| Console Errors | No console.error or uncaught exceptions | High |
| Key Elements | Page has visible content (not blank) | High |
| Navigation | Internal links work (no 404) | Medium |
| Assets | Images/CSS/JS load correctly | Medium |

**Generated Test Structure:**

```typescript
// Auto-generated: {route} verification
test('{route} - page loads without errors', async ({ page }) => {
  // Error collection
  const errors: string[] = []
  const networkErrors: { url: string; status: number }[] = []

  page.on('pageerror', err => errors.push(err.message))
  page.on('console', msg => {
    if (msg.type() === 'error') errors.push(msg.text())
  })
  page.on('response', response => {
    if (response.status() >= 400) {
      networkErrors.push({ url: response.url(), status: response.status() })
    }
  })

  // Navigate
  const response = await page.goto('{target_url}{route}', {
    waitUntil: 'networkidle',
    timeout: 30000
  })

  // Assertions
  expect(response?.status()).toBeLessThan(400)
  expect(errors).toHaveLength(0)

  // Content check (page is not blank)
  const bodyText = await page.locator('body').innerText()
  expect(bodyText.trim().length).toBeGreaterThan(0)
})
```

**Behavior Comparison Test (Source vs Target):**

```typescript
// Auto-generated: {route} behavior comparison
test('{route} - source vs target behavior match', async ({ browser }) => {
  const sourcePage = await browser.newPage()
  const targetPage = await browser.newPage()

  // Navigate both
  const [sourceResponse, targetResponse] = await Promise.all([
    sourcePage.goto('{source_url}{route}', { waitUntil: 'networkidle' }),
    targetPage.goto('{target_url}{route}', { waitUntil: 'networkidle' })
  ])

  // Status code match
  expect(targetResponse?.status()).toBe(sourceResponse?.status())

  // Key content comparison (headings, navigation items)
  const sourceH1 = await sourcePage.locator('h1').allInnerTexts()
  const targetH1 = await targetPage.locator('h1').allInnerTexts()
  expect(targetH1).toEqual(sourceH1)

  // Navigation links match
  const sourceLinks = await sourcePage.locator('nav a').allInnerTexts()
  const targetLinks = await targetPage.locator('nav a').allInnerTexts()
  expect(targetLinks).toEqual(sourceLinks)

  await sourcePage.close()
  await targetPage.close()
})
```

### Step 2.4: User Flow Extraction

Extract critical user journeys from migration artifacts for E2E scenario testing.

**Extraction Sources:**

```
1. migration_plan.md - "Critical Flows" or "User Journeys" section
2. as_is_spec.md - "Features" or "Use Cases" section
3. Source test files - Existing E2E/integration tests
4. Common flow detection - Login, signup, checkout patterns
```

**Common Flow Templates:**

| Flow | Steps | Priority |
|------|-------|----------|
| Authentication | Navigate → Fill form → Submit → Verify redirect | Critical |
| Navigation | Home → Click menu → Verify page load | High |
| Search | Navigate → Type query → Submit → Verify results | High |
| Form Submit | Navigate → Fill fields → Submit → Verify success | Medium |
| CRUD | List → Create → Read → Update → Delete → Verify | Medium |
| Pagination | List → Next page → Verify content change | Low |

**Flow Detection Algorithm:**

```
FUNCTION detectUserFlows(artifacts_dir, source_dir):
  flows = []

  // 1. Check migration_plan.md for documented flows
  plan = read(artifacts_dir + "/migration_plan.md")
  flows.addAll(parseFlowsFromPlan(plan))

  // 2. Check for existing E2E test files
  e2e_files = glob(source_dir, "**/*.{spec,test,e2e}.{ts,js}")
  FOR each file:
    flows.addAll(parseTestFlows(file))

  // 3. Auto-detect common patterns
  IF hasLoginPage(discovered_routes):
    flows.add(AuthenticationFlow)
  IF hasSearchComponent(source_dir):
    flows.add(SearchFlow)
  IF hasFormPages(discovered_routes):
    flows.add(FormSubmissionFlow)

  RETURN flows
```

**Generated E2E Flow Test:**

```typescript
// Auto-generated: Authentication flow
test('Authentication flow - login and access protected page', async ({ page }) => {
  // Step 1: Navigate to login
  await page.goto('{target_url}/login')
  await expect(page.locator('form')).toBeVisible()

  // Step 2: Fill credentials
  await page.fill('[name="email"], [type="email"], #email', 'test@example.com')
  await page.fill('[name="password"], [type="password"], #password', 'password123')

  // Step 3: Submit
  await page.click('button[type="submit"], input[type="submit"]')

  // Step 4: Verify redirect (should not stay on login)
  await page.waitForURL(url => !url.includes('/login'), { timeout: 10000 })

  // Step 5: Verify authenticated state
  const currentUrl = page.url()
  expect(currentUrl).not.toContain('/login')
})
```

### Step 2.5: Route Registry Output

After discovery, create a route registry for use by subsequent phases:

```yaml
# Internal route registry (passed to Phase 3-7)
route_registry:
  static_routes:
    - path: "/"
      source: "file-based"
      priority: critical
    - path: "/dashboard"
      source: "as_is_spec"
      priority: high
    - path: "/settings"
      source: "file-based"
      priority: medium

  dynamic_routes:
    - pattern: "/users/:id"
      sample_url: "/users/1"
      source: "router-config"
      priority: medium

  user_flows:
    - name: "Authentication"
      steps: 5
      source: "auto-detected"
      priority: critical
    - name: "Search"
      steps: 3
      source: "migration_plan"
      priority: high

  excluded_routes:
    - "/api/*"
    - "/_next/*"

  stats:
    total_routes: 25
    static: 18
    dynamic: 5
    excluded: 12
    user_flows: 3
```

---

## Phase 3: Visual Regression Testing

**Activation**: `--visual` or `--full` flag

### Overview

Visual regression testing captures screenshots of each route on both source and target servers, then performs pixel-by-pixel comparison to detect unintended visual changes during migration.

### Step 3.1: Screenshot Capture Engine

**Capture Strategy:**

For each route in route_registry, capture screenshots on both servers across multiple viewports.

```
FUNCTION captureScreenshots(route_registry, source_url, target_url, viewports):
  screenshots = []

  FOR each route in route_registry.static_routes + route_registry.dynamic_routes:
    FOR each viewport in viewports:
      // Capture source
      source_screenshot = capture(source_url + route.path, viewport, "source")
      // Capture target
      target_screenshot = capture(target_url + route.path, viewport, "target")

      screenshots.add({
        route: route.path,
        viewport: viewport.name,
        source: source_screenshot,
        target: target_screenshot
      })

  RETURN screenshots
```

**Capture Function:**

```typescript
async function capturePageScreenshot(
  page: Page,
  url: string,
  viewport: { width: number; height: number },
  maskSelectors: string[]
): Promise<Buffer> {
  // Set viewport
  await page.setViewportSize(viewport)

  // Navigate and wait for stable state
  await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 })

  // Wait for animations to settle
  await page.waitForTimeout(1000)

  // Hide dynamic content (timestamps, ads, avatars)
  for (const selector of maskSelectors) {
    await page.locator(selector).evaluateAll(elements => {
      elements.forEach(el => {
        (el as HTMLElement).style.visibility = 'hidden'
      })
    }).catch(() => {}) // Ignore if selector not found
  }

  // Capture full page screenshot
  return await page.screenshot({
    fullPage: true,
    type: 'png'
  })
}
```

**Screenshot Storage:**

```
{artifacts_dir}/screenshots/
├── source/
│   ├── desktop/
│   │   ├── home.png
│   │   ├── dashboard.png
│   │   └── settings.png
│   ├── tablet/
│   │   └── ...
│   └── mobile/
│       └── ...
├── target/
│   ├── desktop/
│   │   └── ...
│   ├── tablet/
│   │   └── ...
│   └── mobile/
│       └── ...
└── diff/
    ├── desktop/
    │   ├── home-diff.png
    │   ├── dashboard-diff.png
    │   └── settings-diff.png
    ├── tablet/
    │   └── ...
    └── mobile/
        └── ...
```

### Step 3.2: Pixel Comparison Engine

**Comparison Algorithm:**

```typescript
interface ComparisonResult {
  route: string
  viewport: string
  diffPercentage: number
  diffPixels: number
  totalPixels: number
  passed: boolean
  diffImagePath: string | null
}

async function compareScreenshots(
  sourceBuffer: Buffer,
  targetBuffer: Buffer,
  threshold: number  // percentage (default: 5)
): Promise<ComparisonResult> {
  // Decode PNG buffers to pixel arrays
  const sourcePixels = decodePNG(sourceBuffer)
  const targetPixels = decodePNG(targetBuffer)

  // Handle size differences
  const width = Math.max(sourcePixels.width, targetPixels.width)
  const height = Math.max(sourcePixels.height, targetPixels.height)
  const totalPixels = width * height

  let diffPixels = 0
  const diffImage = createEmptyImage(width, height)

  // Pixel-by-pixel comparison
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const sourceColor = getPixel(sourcePixels, x, y)
      const targetColor = getPixel(targetPixels, x, y)

      if (!colorsMatch(sourceColor, targetColor, colorThreshold: 10)) {
        diffPixels++
        setPixel(diffImage, x, y, RED)  // Mark diff in red
      } else {
        setPixel(diffImage, x, y, dimmed(targetColor))  // Dimmed original
      }
    }
  }

  const diffPercentage = (diffPixels / totalPixels) * 100
  const passed = diffPercentage <= threshold

  return {
    diffPercentage: round(diffPercentage, 2),
    diffPixels,
    totalPixels,
    passed,
    diffImagePath: passed ? null : saveDiffImage(diffImage)
  }
}
```

**Color Matching with Tolerance:**

```
FUNCTION colorsMatch(color1, color2, tolerance=10):
  // Allow slight color variations (anti-aliasing, rendering differences)
  RETURN abs(color1.r - color2.r) <= tolerance
     AND abs(color1.g - color2.g) <= tolerance
     AND abs(color1.b - color2.b) <= tolerance
     AND abs(color1.a - color2.a) <= tolerance
```

### Step 3.3: Responsive Viewport Matrix

**Default Viewports:**

| Name | Width | Height | Represents |
|------|-------|--------|------------|
| Desktop | 1920 | 1080 | Standard monitor |
| Laptop | 1366 | 768 | Common laptop |
| Tablet | 768 | 1024 | iPad portrait |
| Mobile | 375 | 812 | iPhone X/12/13 |
| Mobile Small | 320 | 568 | iPhone SE |

**Viewport Execution:**

```
IF --full:
  → Run all 5 viewports
ELIF --visual (no --cross-browser):
  → Run Desktop + Tablet + Mobile (3 viewports)
ELSE:
  → Run Desktop only (1 viewport)
```

**Responsive-Specific Checks:**

```typescript
// Check for horizontal overflow (common migration issue)
async function checkHorizontalOverflow(page: Page): Promise<boolean> {
  return await page.evaluate(() => {
    return document.documentElement.scrollWidth > document.documentElement.clientWidth
  })
}

// Check for overlapping elements
async function checkElementOverlap(page: Page): Promise<string[]> {
  return await page.evaluate(() => {
    const elements = document.querySelectorAll('*')
    const overlaps: string[] = []
    // ... bounding rect comparison logic
    return overlaps
  })
}
```

### Step 3.4: Threshold & Masking Configuration

**Threshold Levels:**

| Level | Percentage | Use Case |
|-------|-----------|----------|
| Strict | 1% | Pixel-perfect migrations (same CSS framework) |
| Normal | 5% | Default - allows minor rendering differences |
| Relaxed | 10% | Framework changes with known visual differences |
| Loose | 20% | Major redesigns where layout is similar |

**Configuration from `.migrate-config.yaml`:**

```yaml
verification:
  visual_threshold: 5          # Default threshold
  mask_selectors:              # Elements to hide before comparison
    - "[data-testid='timestamp']"
    - "[data-testid='random-id']"
    - ".ad-banner"
    - ".user-avatar"
    - "[class*='skeleton']"
    - ".loading-spinner"
    - "time"                   # All <time> elements
    - "[datetime]"             # Elements with datetime attr
```

**Per-Route Threshold Override (advanced):**

```yaml
verification:
  visual_threshold: 5          # Global default
  route_thresholds:            # Per-route overrides
    "/": 2                     # Homepage: stricter
    "/dashboard": 10           # Dashboard: has dynamic charts
    "/settings": 3             # Settings: static content
```

### Step 3.5: Visual Report Generation

**Report Format: HTML (viewable in browser)**

```
{artifacts_dir}/visual-report.html
```

**Report Structure:**

```html
<!DOCTYPE html>
<html>
<head>
  <title>Visual Regression Report - Migration Verification</title>
  <style>
    .comparison { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 8px; }
    .pass { border: 2px solid green; }
    .fail { border: 2px solid red; }
    .warn { border: 2px solid orange; }
    img { width: 100%; height: auto; }
  </style>
</head>
<body>
  <h1>Visual Regression Report</h1>
  <p>Generated: {timestamp} | Threshold: {threshold}%</p>

  <!-- Summary -->
  <table>
    <tr><th>Route</th><th>Viewport</th><th>Diff %</th><th>Status</th></tr>
    <!-- Per-route results -->
  </table>

  <!-- Detailed Comparisons -->
  <div class="route-detail" id="route-{encoded_path}">
    <h2>{route} ({viewport})</h2>
    <p>Diff: {diff_pct}% | Pixels: {diff_pixels}/{total_pixels}</p>
    <div class="comparison">
      <div>
        <h3>Source</h3>
        <img src="screenshots/source/{viewport}/{route}.png" />
      </div>
      <div>
        <h3>Target</h3>
        <img src="screenshots/target/{viewport}/{route}.png" />
      </div>
      <div>
        <h3>Diff</h3>
        <img src="screenshots/diff/{viewport}/{route}-diff.png" />
      </div>
    </div>
  </div>
</body>
</html>
```

**Report Summary Section:**

```
Visual Regression Summary
=========================
Total Routes Tested: 25
Total Comparisons: 75 (25 routes x 3 viewports)

Results:
  PASS: 70 (93.3%)
  WARN: 3 (4.0%) - within 80-100% of threshold
  FAIL: 2 (2.7%) - exceeded threshold

Failed Routes:
  /profile (Desktop) - diff: 7.2% (threshold: 5%)
  /profile (Mobile) - diff: 8.1% (threshold: 5%)

Warning Routes:
  /settings (Desktop) - diff: 4.9% (threshold: 5%)
  /dashboard (Tablet) - diff: 4.2% (threshold: 5%)
  /dashboard (Mobile) - diff: 4.5% (threshold: 5%)
```

### Step 3.6: Diff Analysis & Categorization

**Diff Categories (for actionable feedback):**

| Category | Detection Method | Severity |
|----------|-----------------|----------|
| Layout Shift | Large contiguous diff regions | High |
| Color Change | Scattered pixel diffs with consistent color offset | Medium |
| Font Rendering | Text-area-only diffs with similar shapes | Low |
| Missing Element | One side has content, other is blank | Critical |
| Extra Element | Target has content not in source | Medium |
| Size Difference | Page height/width mismatch | High |

**Categorization Logic:**

```
FUNCTION categorizeDiff(diffImage, sourceImage, targetImage):
  IF targetImage.height != sourceImage.height:
    → "Size Difference" (page height changed)

  IF hasLargeContiguousRegion(diffImage, minSize: 100x100):
    → "Layout Shift" (major positioning change)

  IF diffOnlyInTextAreas(diffImage, sourceImage):
    → "Font Rendering" (acceptable in most cases)

  IF hasBlankRegionInTarget(targetImage, diffRegions):
    → "Missing Element" (critical - content lost)

  IF hasNewContentInTarget(targetImage, sourceImage, diffRegions):
    → "Extra Element" (review needed)

  DEFAULT:
    → "Color Change" (minor styling difference)
```

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

## Phase 7: Accessibility Verification

**Activation**: `--a11y` or `--full` flag

### Overview

Accessibility verification uses axe-core (the industry-standard accessibility testing engine) integrated with Playwright to detect WCAG 2.1 AA violations. It compares accessibility scores between source and target to ensure the migration does not introduce accessibility regressions.

### Step 7.1: axe-core Integration

**Purpose**: Automated WCAG compliance scanning for all discovered routes.

**axe-core Setup:**

```typescript
// Note: Requires @axe-core/playwright package
// Install: npm install -D @axe-core/playwright

import AxeBuilder from '@axe-core/playwright'

interface AccessibilityResult {
  route: string
  server: 'source' | 'target'
  violations: AxeViolation[]
  passes: number
  incomplete: number
  inapplicable: number
  score: number            // Calculated accessibility score (0-100)
  wcagLevel: 'A' | 'AA' | 'AAA'
  timestamp: string
}

interface AxeViolation {
  id: string               // Rule ID (e.g., "color-contrast")
  impact: 'critical' | 'serious' | 'moderate' | 'minor'
  description: string
  help: string
  helpUrl: string
  nodes: ViolationNode[]
  tags: string[]           // WCAG tags (e.g., "wcag2aa", "wcag143")
}

interface ViolationNode {
  html: string             // The offending HTML element
  target: string[]         // CSS selector path
  failureSummary: string   // What to fix
}
```

**Accessibility Scan Engine:**

```typescript
async function scanAccessibility(
  page: Page,
  url: string,
  route: string,
  wcagLevel: 'A' | 'AA' | 'AAA' = 'AA'
): Promise<AccessibilityResult> {
  await page.goto(url + route, { waitUntil: 'networkidle', timeout: 30000 })

  // Wait for dynamic content to render
  await page.waitForTimeout(1000)

  // Configure axe-core tags based on WCAG level
  const tags = getWcagTags(wcagLevel)

  // Run axe analysis
  const results = await new AxeBuilder({ page })
    .withTags(tags)
    .exclude('[data-testid="dynamic-content"]')  // Exclude known dynamic areas
    .analyze()

  // Calculate accessibility score
  const totalRules = results.violations.length + results.passes.length
  const score = totalRules > 0
    ? Math.round((results.passes.length / totalRules) * 100)
    : 100

  return {
    route,
    server: 'target',
    violations: results.violations.map(v => ({
      id: v.id,
      impact: v.impact as any,
      description: v.description,
      help: v.help,
      helpUrl: v.helpUrl,
      nodes: v.nodes.map(n => ({
        html: n.html,
        target: n.target as string[],
        failureSummary: n.failureSummary ?? ''
      })),
      tags: v.tags
    })),
    passes: results.passes.length,
    incomplete: results.incomplete.length,
    inapplicable: results.inapplicable.length,
    score,
    wcagLevel,
    timestamp: new Date().toISOString()
  }
}

function getWcagTags(level: 'A' | 'AA' | 'AAA'): string[] {
  const tags = ['wcag2a', 'best-practice']
  if (level === 'AA' || level === 'AAA') tags.push('wcag2aa', 'wcag21aa')
  if (level === 'AAA') tags.push('wcag2aaa', 'wcag21aaa')
  return tags
}
```

### Step 7.2: Accessibility Regression Comparison

**Purpose**: Compare accessibility scores between source and target to detect regressions.

**Regression Analysis:**

```typescript
interface AccessibilityComparison {
  route: string
  sourceScore: number
  targetScore: number
  scoreDelta: number           // Positive = improved, Negative = regressed
  newViolations: AxeViolation[]    // Violations in target but not in source
  resolvedViolations: AxeViolation[] // Violations in source but not in target
  persistentViolations: AxeViolation[] // Violations in both
  regressionDetected: boolean
}

async function compareAccessibility(
  sourceUrl: string,
  targetUrl: string,
  route: string,
  wcagLevel: 'A' | 'AA' | 'AAA'
): Promise<AccessibilityComparison> {
  const browser = await chromium.launch({ headless: true })

  // Scan source
  const sourcePage = await browser.newPage()
  const sourceResult = await scanAccessibility(sourcePage, sourceUrl, route, wcagLevel)
  await sourcePage.close()

  // Scan target
  const targetPage = await browser.newPage()
  const targetResult = await scanAccessibility(targetPage, targetUrl, route, wcagLevel)
  await targetPage.close()

  await browser.close()

  // Compare violations by rule ID
  const sourceViolationIds = new Set(sourceResult.violations.map(v => v.id))
  const targetViolationIds = new Set(targetResult.violations.map(v => v.id))

  const newViolations = targetResult.violations
    .filter(v => !sourceViolationIds.has(v.id))

  const resolvedViolations = sourceResult.violations
    .filter(v => !targetViolationIds.has(v.id))

  const persistentViolations = targetResult.violations
    .filter(v => sourceViolationIds.has(v.id))

  return {
    route,
    sourceScore: sourceResult.score,
    targetScore: targetResult.score,
    scoreDelta: targetResult.score - sourceResult.score,
    newViolations,
    resolvedViolations,
    persistentViolations,
    regressionDetected: newViolations.some(v =>
      v.impact === 'critical' || v.impact === 'serious'
    )
  }
}
```

### Step 7.3: Violation Categorization & Prioritization

**Purpose**: Group and prioritize accessibility violations for actionable remediation.

**Violation Categories:**

```typescript
interface ViolationCategory {
  name: string
  description: string
  rules: string[]
  priority: 'critical' | 'high' | 'medium' | 'low'
  estimatedEffort: 'quick' | 'moderate' | 'significant'
}

const VIOLATION_CATEGORIES: ViolationCategory[] = [
  {
    name: 'Color & Contrast',
    description: 'Text and UI elements do not meet contrast ratio requirements',
    rules: ['color-contrast', 'link-in-text-block'],
    priority: 'high',
    estimatedEffort: 'quick'
  },
  {
    name: 'Keyboard Navigation',
    description: 'Interactive elements not accessible via keyboard',
    rules: ['keyboard', 'tabindex', 'focus-order-semantics', 'scrollable-region-focusable'],
    priority: 'critical',
    estimatedEffort: 'moderate'
  },
  {
    name: 'Images & Media',
    description: 'Missing alt text or media alternatives',
    rules: ['image-alt', 'input-image-alt', 'svg-img-alt', 'video-caption', 'audio-caption'],
    priority: 'high',
    estimatedEffort: 'quick'
  },
  {
    name: 'Form Labels',
    description: 'Form inputs missing associated labels',
    rules: ['label', 'input-button-name', 'select-name', 'autocomplete-valid'],
    priority: 'high',
    estimatedEffort: 'quick'
  },
  {
    name: 'Document Structure',
    description: 'Missing landmarks, heading hierarchy issues',
    rules: ['landmark-one-main', 'page-has-heading-one', 'heading-order', 'region', 'bypass'],
    priority: 'medium',
    estimatedEffort: 'moderate'
  },
  {
    name: 'ARIA Usage',
    description: 'Incorrect or missing ARIA attributes',
    rules: ['aria-valid-attr', 'aria-required-attr', 'aria-roles', 'aria-hidden-focus', 'aria-allowed-attr'],
    priority: 'medium',
    estimatedEffort: 'moderate'
  },
  {
    name: 'Links & Buttons',
    description: 'Empty or non-descriptive interactive elements',
    rules: ['link-name', 'button-name', 'duplicate-id-active', 'identical-links-same-purpose'],
    priority: 'high',
    estimatedEffort: 'quick'
  },
  {
    name: 'Tables & Lists',
    description: 'Data table accessibility and list semantics',
    rules: ['td-headers-attr', 'th-has-data-cells', 'definition-list', 'list', 'listitem'],
    priority: 'medium',
    estimatedEffort: 'moderate'
  }
]

function categorizeViolations(violations: AxeViolation[]): Map<string, AxeViolation[]> {
  const categorized = new Map<string, AxeViolation[]>()

  for (const category of VIOLATION_CATEGORIES) {
    const matching = violations.filter(v => category.rules.includes(v.id))
    if (matching.length > 0) {
      categorized.set(category.name, matching)
    }
  }

  // Uncategorized violations
  const allCategorizedRules = VIOLATION_CATEGORIES.flatMap(c => c.rules)
  const uncategorized = violations.filter(v => !allCategorizedRules.includes(v.id))
  if (uncategorized.length > 0) {
    categorized.set('Other', uncategorized)
  }

  return categorized
}
```

### Step 7.4: Accessibility Report Generation

**Purpose**: Generate comprehensive accessibility report with actionable fix suggestions.

**Report Structure:**

```
Accessibility Verification Report
====================================

WCAG Level: AA (2.1)
Routes Tested: 25

Overall Summary:
| Metric | Source | Target | Change |
|--------|--------|--------|--------|
| Average Score | 88 | 91 | +3 |
| Total Violations | 15 | 12 | -3 (improved) |
| Critical/Serious | 3 | 1 | -2 (improved) |
| Routes 100% Pass | 18 | 20 | +2 |

Per-Route Results:
| Route | Source Score | Target Score | Violations | Impact | Status |
|-------|-------------|-------------|------------|--------|--------|
| / | 95 | 98 | 0 | - | PASS |
| /dashboard | 82 | 85 | 2 | moderate | WARN |
| /settings | 90 | 92 | 1 | minor | PASS |
| /profile | 88 | 75 | 4 | serious | FAIL |

New Violations (Target Only):
┌─────────────────────────────────────────────────────────┐
│ SERIOUS: color-contrast (Route: /profile)               │
│ Description: Elements must have sufficient contrast     │
│ Affected: 3 elements                                    │
│ Fix: Ensure text color has ≥4.5:1 ratio with background│
│ Help: https://dequeuniversity.com/rules/axe/4.8/...    │
│                                                         │
│ Elements:                                               │
│   <p class="subtitle">...</p>  (contrast: 3.2:1)      │
│   <span class="meta">...</span> (contrast: 2.8:1)     │
└─────────────────────────────────────────────────────────┘

Resolved Violations (Fixed by Migration):
  ✓ landmark-one-main (was on /dashboard)
  ✓ heading-order (was on /settings, /about)

By Category:
| Category | Count | Priority | Effort |
|----------|-------|----------|--------|
| Color & Contrast | 3 | high | quick |
| Form Labels | 2 | high | quick |
| Document Structure | 1 | medium | moderate |
| ARIA Usage | 1 | medium | moderate |

Regression Status: WARN (new serious violation detected)
```

**Report Data Structure:**

```typescript
interface AccessibilityReport {
  summary: {
    wcagLevel: string
    routesTested: number
    sourceAvgScore: number
    targetAvgScore: number
    scoreDelta: number
    totalViolations: { source: number; target: number }
    criticalSerious: { source: number; target: number }
    routesPassed: { source: number; target: number }
  }
  routeResults: AccessibilityComparison[]
  newViolations: { route: string; violation: AxeViolation }[]
  resolvedViolations: { route: string; violation: AxeViolation }[]
  byCategory: { category: string; count: number; priority: string; effort: string }[]
  overallStatus: 'pass' | 'warn' | 'fail'
  recommendations: string[]
}

function determineOverallStatus(report: AccessibilityReport): 'pass' | 'warn' | 'fail' {
  // FAIL: New critical or serious violations introduced
  if (report.newViolations.some(v =>
    v.violation.impact === 'critical' || v.violation.impact === 'serious'
  )) return 'fail'

  // WARN: Score decreased or new moderate violations
  if (report.summary.scoreDelta < -5) return 'warn'
  if (report.newViolations.some(v => v.violation.impact === 'moderate')) return 'warn'

  // PASS: No regressions
  return 'pass'
}

function generateRecommendations(report: AccessibilityReport): string[] {
  const recs: string[] = []

  const categories = categorizeViolations(
    report.newViolations.map(v => v.violation)
  )

  if (categories.has('Color & Contrast')) {
    recs.push('Update color palette to meet WCAG AA contrast ratios (≥4.5:1 for normal text, ≥3:1 for large text)')
  }
  if (categories.has('Form Labels')) {
    recs.push('Add <label> elements or aria-label attributes to all form inputs')
  }
  if (categories.has('Images & Media')) {
    recs.push('Add descriptive alt text to all <img> elements (use alt="" for decorative images)')
  }
  if (categories.has('Keyboard Navigation')) {
    recs.push('Ensure all interactive elements are focusable and keyboard-operable (tabindex, focus styles)')
  }
  if (categories.has('Document Structure')) {
    recs.push('Add landmark regions (main, nav, footer) and ensure heading hierarchy (h1 → h2 → h3)')
  }
  if (categories.has('ARIA Usage')) {
    recs.push('Review ARIA attributes: ensure valid roles, required attributes present, no conflicts')
  }

  return recs
}
```

### Step 7.5: Accessibility Configuration

**Purpose**: Configure accessibility verification behavior.

**Configuration (from `.migrate-config.yaml`):**

```yaml
verification:
  accessibility:
    wcag_level: "AA"              # A | AA | AAA
    fail_on_serious: true         # Fail verification on serious/critical violations
    fail_on_regression: true      # Fail if new violations introduced
    score_threshold: 80           # Minimum acceptable score
    exclude_rules: []             # Rules to skip (e.g., ["color-contrast"])
    exclude_selectors:            # Elements to skip during scan
      - "[data-testid='dynamic']"
      - ".third-party-widget"
      - "iframe"
    include_best_practices: true  # Include best-practice rules (not strictly WCAG)
```

**Accessibility Execution Flow:**

```
FUNCTION runAccessibilityVerification(route_registry, sourceUrl, targetUrl, config):
  results = []

  FOR each route in route_registry.static_routes:
    comparison = await compareAccessibility(
      sourceUrl, targetUrl, route.path, config.wcag_level
    )
    results.add(comparison)

  // Generate report
  report = generateAccessibilityReport(results, config)

  // Determine pass/fail
  IF config.fail_on_serious AND report.overallStatus === "fail":
    → Mark verification as FAILED
  ELIF config.fail_on_regression AND report.summary.scoreDelta < -10:
    → Mark verification as FAILED
  ELIF report.summary.targetAvgScore < config.score_threshold:
    → Mark verification as FAILED

  RETURN report
```

---

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

---

## Verification Types

### 1. Characterization Tests
```
Running characterization tests...

auth/login.test.ts          PASS 12/12 passed
auth/logout.test.ts         PASS 5/5 passed
users/crud.test.ts          PASS 18/18 passed
orders/calculate.test.ts    WARN 9/10 passed (1 improved)
```

### 2. Behavior Comparison
```
GET /api/users     PASS Identical response
POST /api/orders   PASS Identical response
GET /api/products  PASS Identical response
```

### 3. E2E Tests (Playwright)
```
Login Flow         PASS (Chromium, 1.2s)
Checkout Flow      PASS (Chromium, 2.8s)
User Registration  PASS (Chromium, 1.5s)
```

### 4. Visual Regression
```
/              PASS diff: 0.2% (threshold: 5%)
/dashboard     PASS diff: 1.8% (threshold: 5%)
/settings      WARN diff: 4.9% (threshold: 5%)
/profile       FAIL diff: 7.2% (threshold: 5%)
```

### 5. Performance
```
| Metric      | Source | Target | Change | Budget |
|-------------|--------|--------|--------|--------|
| LCP         | 2.1s   | 1.8s   | -14%   | PASS   |
| CLS         | 0.05   | 0.03   | -40%   | PASS   |
| FID         | 80ms   | 45ms   | -44%   | PASS   |
| Load Time   | 3.2s   | 2.1s   | -34%   | PASS   |
| JS Bundle   | 450KB  | 380KB  | -16%   | PASS   |
```

### 6. Cross-Browser
```
| Route      | Chromium | Firefox | WebKit |
|------------|----------|---------|--------|
| /          | PASS     | PASS    | PASS   |
| /dashboard | PASS     | PASS    | WARN   |
| /settings  | PASS     | PASS    | PASS   |
```

### 7. Accessibility
```
| Route      | Violations | Impact   | Score |
|------------|-----------|----------|-------|
| /          | 0         | -        | 98    |
| /dashboard | 1         | minor    | 94    |
| /settings  | 0         | -        | 96    |
```

---

## Final Report

```markdown
# Migration Verification Report

## Environment
- Source: {source_framework} @ {source_url}
- Target: {target_framework} @ {target_url}
- Date: {timestamp}
- Duration: {total_time}

## Summary
| Category | Passed | Failed | Rate |
|----------|--------|--------|------|
| Characterization | 148 | 2 | 98.7% |
| Behavior | 45 | 0 | 100% |
| E2E | 19 | 1 | 95% |
| Visual Regression | 72 | 3 | 95.8% |
| Performance | 25 | 0 | 100% |
| Cross-Browser | 75 | 1 | 98.7% |
| Accessibility | 25 | 1 | 96% |
| **Total** | **409** | **8** | **98.0%** |

## Status: PASSED / FAILED

## Known Differences (Intentional)
1. Improved error messages
2. Better validation responses

## Performance Gains
- 14% faster LCP
- 34% faster page loads
- 16% smaller JS bundle

## Recommendation
Ready for production deployment / Needs attention
```

---

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Route Discovery | `e2e-tester` | Discover testable routes from artifacts |
| Behavior Validation | `e2e-tester` | Compare source/target behavior |
| E2E + Visual | `e2e-tester` | Playwright-based testing |
| Performance | `optimizer` | Performance metrics collection |
| Security Review | `security-auditor` | Vulnerability check |

---

## Workflow (Data Flow)

```
/jikime:migrate-0-discover
        | (.migrate-config.yaml created)
/jikime:migrate-1-analyze
        | (config updated + as_is_spec.md)
/jikime:migrate-2-plan
        | (migration_plan.md)
/jikime:migrate-3-execute
        | (output_dir/ + progress.yaml)
/jikime:migrate-4-verify  <- current (final)
        |
        |-- Step 1: Read .migrate-config.yaml
        |-- Step 2: Detect dev commands (framework-aware)
        |-- Step 3: Install dependencies (if needed)
        |-- Step 4: Start dev servers (source + target)
        |-- Step 5: Health check (wait for ready)
        |-- Step 6: Run verification suite
        |   |-- Characterization tests
        |   |-- Behavior comparison
        |   |-- E2E Playwright tests
        |   |-- Visual regression (if --visual/--full)
        |   |-- Performance (if --performance/--full)
        |   |-- Cross-browser (if --cross-browser/--full)
        |   |-- Accessibility (if --a11y/--full)
        |-- Step 7: Generate verification report
        |-- Step 8: Cleanup (stop dev servers)
        |
        |-- Creates: {artifacts_dir}/verification_report.md
        |-- Creates: {artifacts_dir}/screenshots/ (if --visual)
```

---

## .migrate-config.yaml Verification Schema

The following fields are used by this command (added to existing config):

```yaml
# Existing fields (from previous steps)
source_dir: "./source-project"
output_dir: "./migrated-project"
source_framework: react
target_framework: nextjs
artifacts_dir: "./.migrate-artifacts"

# Verification settings (optional, auto-detected if missing)
verification:
  dev_command: ""                    # Override auto-detection (empty = auto)
  source_port: 3000                  # Source dev server port
  target_port: 3001                  # Target dev server port
  visual_threshold: 5                # Allowed visual diff percentage
  crawl_depth: 3                     # Navigation link discovery depth
  health_check_timeout: 30           # Max seconds to wait for server
  test_routes:                       # Manual route overrides (optional)
    - "/"
    - "/dashboard"
    - "/login"
  mask_selectors:                    # Ignore dynamic content in visual diff
    - "[data-testid='timestamp']"
    - ".ad-banner"
  performance_budget:
    lcp_regression_pct: 20           # Max LCP regression vs source
    page_load_max_ms: 3000           # Absolute max page load
    js_bundle_max_kb: 500            # Max JS bundle size
```

---

## Migration Complete!

Migration is complete when verification passes.

**Next Steps:**
1. Deploy to staging environment
2. User Acceptance Testing (UAT)
3. Production deployment

## Related Commands

- `/jikime:browser-verify` - Standalone browser runtime error detection and auto-fix loop. Use this for catching runtime errors (undefined references, missing modules, DOM errors) that static analysis and build tools miss. Works independently of migration workflow.
- `/jikime:e2e` - E2E test generation and execution
- `/jikime:loop` - General iterative fix loop (LSP, tests, coverage)

> **Tip**: After migration verification passes, run `/jikime:browser-verify` to catch any remaining runtime browser errors that only appear during actual page rendering.

---

Version: 4.0.0
Changelog:
- v4.0.0: Playwright-based verification with dev server lifecycle, visual regression, cross-browser, accessibility, performance budgets
- v3.0.0: Config-first approach; Renamed --source/--target to --source-url/--target-url for clarity; Added data flow diagram
- v2.1.0: Initial verification command
