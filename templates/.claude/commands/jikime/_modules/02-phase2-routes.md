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
