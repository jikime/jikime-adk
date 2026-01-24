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

## Phase 1: Infrastructure (Dev Server Lifecycle)

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 1.1 | Dev Server Manager Logic | Target framework detection → dev command extraction → start/wait/stop pipeline | `migrate-4-verify.md` |
| 1.2 | Framework Dev Command Mapping | Auto-detect dev commands for Next.js, Vite, CRA, Nuxt, Angular, etc. | `migrate-4-verify.md` |
| 1.3 | Health Check Pattern | Port listening detection + HTTP 200 confirmation + timeout/retry logic | `migrate-4-verify.md` |
| 1.4 | Dual Server Mode | When `--source-url` not provided, auto-start source too (separate ports) | `migrate-4-verify.md` |

### Framework Dev Command Reference

```yaml
next: "npx next dev --port {port}"
vite: "npx vite --port {port}"
cra: "npx react-scripts start"  # PORT env var
nuxt: "npx nuxt dev --port {port}"
angular: "npx ng serve --port {port}"
remix: "npx remix dev --port {port}"
svelte: "npx vite dev --port {port}"
```

### Health Check Flow

```
Start server (background)
  → Wait 1s
  → HTTP GET http://localhost:{port}
  → If 200: Ready
  → If fail: Retry (max 30s, 1s interval)
  → If timeout: Abort with error
```

---

## Phase 2: Route Auto-Discovery & Test Generation

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 2.1 | Route Discovery Engine | Extract routes/pages from `as_is_spec.md` + source code | `migrate-4-verify.md` |
| 2.2 | Dynamic Route Handling | `/users/[id]` style routes → generate sample test URLs | `migrate-4-verify.md` |
| 2.3 | Test Case Auto-Generation | Per-route basic verification code (200 response, no errors, key elements) | New skill or agent |
| 2.4 | User Flow Extraction | Identify critical user flows from `migration_plan.md` → E2E scenarios | `migrate-4-verify.md` |

### Route Discovery Sources

```
Priority 1: as_is_spec.md (routes section)
Priority 2: File-based routing (pages/, app/ directories)
Priority 3: Router configuration files (react-router, vue-router, etc.)
Priority 4: Sitemap or navigation components
```

### Auto-Generated Test Template

```typescript
// Per-route verification
test('GET /dashboard - page loads correctly', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', err => errors.push(err.message))
  page.on('console', msg => { if (msg.type() === 'error') errors.push(msg.text()) })

  const response = await page.goto('/dashboard')
  expect(response?.status()).toBe(200)
  expect(errors).toHaveLength(0)

  // Key element verification
  await expect(page.locator('h1')).toBeVisible()
})
```

---

## Phase 3: Visual Regression (Core Feature)

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 3.1 | Screenshot Capture Engine | Full-page + viewport screenshots per route | Skill reference |
| 3.2 | Source vs Target Comparison | Pixel-by-pixel comparison + diff image generation | Skill reference |
| 3.3 | Responsive Verification | Desktop(1920), Tablet(768), Mobile(375) - 3 viewports auto-verified | `migrate-4-verify.md` |
| 3.4 | Threshold Configuration | Allowed diff ratio (default 5%), dynamic content area masking | `migrate-4-verify.md` |
| 3.5 | Visual Report Generation | HTML report with before/after/diff side-by-side | `migrate-4-verify.md` |

### Visual Comparison Flow

```
For each discovered route:
  1. Navigate source server → screenshot (source_{route}_{viewport}.png)
  2. Navigate target server → screenshot (target_{route}_{viewport}.png)
  3. Pixel compare → diff image (diff_{route}_{viewport}.png)
  4. Calculate diff percentage
  5. Pass/Fail based on threshold
```

### Viewport Matrix

| Viewport | Width | Height | Device |
|----------|-------|--------|--------|
| Desktop | 1920 | 1080 | Standard monitor |
| Tablet | 768 | 1024 | iPad |
| Mobile | 375 | 812 | iPhone X |

### Masking Strategy (Dynamic Content)

```yaml
mask_selectors:
  - "[data-testid='timestamp']"
  - "[data-testid='random-content']"
  - ".ad-banner"
  - ".user-avatar"
  - "[class*='skeleton']"
```

---

## Phase 4: Behavioral Testing (Functional Verification)

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 4.1 | Page Load Verification | All pages return 200 + no console.error + no uncaught exceptions | `migrate-4-verify.md` |
| 4.2 | Navigation Verification | Internal link clicks → correct page transitions (broken link detection) | `migrate-4-verify.md` |
| 4.3 | Form Interaction Verification | Login, signup, search - core form behavior confirmation | `migrate-4-verify.md` |
| 4.4 | API Call Verification | Network tab monitoring → API request/response status code validation | `migrate-4-verify.md` |
| 4.5 | JavaScript Error Collection | `page.on('pageerror')` + `console.error` full capture and reporting | `migrate-4-verify.md` |

### Error Collection Pattern

```typescript
interface PageErrors {
  route: string
  jsErrors: string[]       // page.on('pageerror')
  consoleErrors: string[]  // console.error messages
  networkErrors: {         // Failed network requests
    url: string
    status: number
    method: string
  }[]
  uncaughtExceptions: string[]
}
```

### Navigation Crawl Strategy

```
1. Start at root (/)
2. Find all <a href="..."> with internal links
3. Click each → verify page loads without error
4. Recurse (max depth: 3)
5. Report broken links
```

---

## Phase 5: Cross-Browser Verification

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 5.1 | Multi-Browser Execution | Chromium + Firefox + WebKit parallel execution | `migrate-4-verify.md` |
| 5.2 | Browser Diff Report | Rendering differences between browsers detected and reported | `migrate-4-verify.md` |
| 5.3 | Mobile Emulation | iPhone, Android device emulation testing | `migrate-4-verify.md` |

### Browser Matrix

```yaml
browsers:
  - name: chromium
    viewport: { width: 1920, height: 1080 }
  - name: firefox
    viewport: { width: 1920, height: 1080 }
  - name: webkit
    viewport: { width: 1920, height: 1080 }

mobile_devices:
  - name: "iPhone 14"
    userAgent: "..."
    viewport: { width: 390, height: 844 }
  - name: "Pixel 7"
    userAgent: "..."
    viewport: { width: 412, height: 915 }
```

---

## Phase 6: Performance Comparison

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 6.1 | Core Web Vitals Collection | LCP, FID/INP, CLS metrics source vs target comparison | `migrate-4-verify.md` |
| 6.2 | Page Load Time | Navigation Timing API per-page load time measurement | `migrate-4-verify.md` |
| 6.3 | Bundle Size Comparison | JS/CSS transfer size comparison from network requests | `migrate-4-verify.md` |
| 6.4 | Performance Budget | Degradation threshold vs source (e.g., LCP within +20%) | `migrate-4-verify.md` |

### Performance Metrics Collection

```typescript
const metrics = await page.evaluate(() => ({
  // Navigation Timing
  loadTime: performance.timing.loadEventEnd - performance.timing.navigationStart,
  domContentLoaded: performance.timing.domContentLoadedEventEnd - performance.timing.navigationStart,
  firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime,

  // Core Web Vitals
  lcp: /* PerformanceObserver LCP */,
  cls: /* Layout Shift accumulation */,
  fid: /* First Input Delay */,

  // Resource metrics
  totalTransferSize: performance.getEntriesByType('resource')
    .reduce((sum, r) => sum + r.transferSize, 0),
  jsSize: performance.getEntriesByType('resource')
    .filter(r => r.name.endsWith('.js'))
    .reduce((sum, r) => sum + r.transferSize, 0),
}))
```

### Performance Budget Template

```yaml
performance_budget:
  lcp_max_ms: 2500           # Absolute max
  lcp_regression_pct: 20     # Max regression vs source
  cls_max: 0.1
  fid_max_ms: 100
  total_js_kb: 500
  total_css_kb: 100
  page_load_max_ms: 3000
```

---

## Phase 7: Accessibility Verification (Bonus)

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 7.1 | axe-core Integration | Playwright + @axe-core/playwright for WCAG violation auto-detection | `migrate-4-verify.md` |
| 7.2 | Accessibility Regression Comparison | Pre/post migration accessibility score comparison | `migrate-4-verify.md` |

### axe-core Integration Pattern

```typescript
import AxeBuilder from '@axe-core/playwright'

test('accessibility check', async ({ page }) => {
  await page.goto('/dashboard')

  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa'])
    .analyze()

  expect(results.violations).toHaveLength(0)
})
```

### Accessibility Report Format

```markdown
| Route | Violations | Impact | Source Score | Target Score |
|-------|-----------|--------|-------------|-------------|
| / | 0 | - | 95 | 98 |
| /dashboard | 2 | minor | 88 | 85 |
| /settings | 0 | - | 92 | 94 |
```

---

## Phase 8: Agent & Skill Implementation

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 8.1 | Enhance `e2e-tester` Agent | Add migration verification workflow | `agents/jikime/e2e-tester.md` |
| 8.2 | Create Playwright Migration Skill | Migration-specific Playwright pattern reference | New skill directory |
| 8.3 | Visual Regression Example Code | Python/TypeScript examples (based on moai-adk + extensions) | skill `examples/` |

### New Skill Structure

```
.claude/skills/jikime-workflow-playwright-migration/
├── SKILL.md                    # Frontmatter + triggers
├── reference.md                # Best practices for migration testing
├── modules/
│   ├── visual-regression.md    # Screenshot comparison patterns
│   ├── route-discovery.md      # Auto-route detection patterns
│   └── server-lifecycle.md     # Dev server management patterns
└── examples/
    ├── visual-comparison.ts    # TypeScript visual regression example
    ├── route-crawler.ts        # Route auto-discovery example
    └── cross-browser-verify.ts # Cross-browser verification example
```

---

## Phase 9: Command & Configuration Updates

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 9.1 | Overhaul `migrate-4-verify.md` | Integrate Phase 1-7 workflows into command | `commands/jikime/migrate-4-verify.md` |
| 9.2 | Add New Flags | `--visual`, `--cross-browser`, `--a11y`, `--port`, `--headless` | `migrate-4-verify.md` |
| 9.3 | Update `allowed-tools` | Add Playwright MCP tools | `migrate-4-verify.md` |
| 9.4 | Extend `.migrate-config.yaml` Schema | Add dev_command, port, test_routes, visual_threshold fields | Config schema docs |

### New Flags

```yaml
flags:
  --visual:        "Enable visual regression comparison (screenshots)"
  --cross-browser: "Run verification across Chromium, Firefox, WebKit"
  --a11y:          "Include accessibility (axe-core) checks"
  --port:          "Custom port for target dev server (default: 3001)"
  --source-port:   "Custom port for source dev server (default: 3000)"
  --headless:      "Run browsers in headless mode (default: true)"
  --threshold:     "Visual diff threshold percentage (default: 5)"
  --depth:         "Navigation crawl depth (default: 3)"
  --full:          "Run ALL verification types (visual + cross-browser + a11y + performance)"
```

### Extended `.migrate-config.yaml` Schema

```yaml
# Existing fields...
source_framework: react
target_framework: nextjs

# NEW: Playwright verification fields
verification:
  dev_command: "npm run dev"           # Override auto-detection
  source_port: 3000                    # Source dev server port
  target_port: 3001                    # Target dev server port
  visual_threshold: 5                  # Allowed diff percentage
  crawl_depth: 3                       # Navigation link depth
  test_routes:                         # Manual route overrides
    - "/"
    - "/dashboard"
    - "/users"
  mask_selectors:                      # Dynamic content to ignore
    - "[data-testid='timestamp']"
  performance_budget:
    lcp_regression_pct: 20
    page_load_max_ms: 3000
```

---

## Phase 10: Reports & Artifacts

| # | Task | Description | Target File |
|---|------|-------------|-------------|
| 10.1 | Unified Verification Report Template | All verification results in one markdown | `migrate-4-verify.md` |
| 10.2 | HTML Visual Report Template | Screenshot comparison visualization (viewable in browser) | Skill reference |
| 10.3 | CI/CD Integration Guide | GitHub Actions/GitLab CI headless execution guide | Skill reference |

### Unified Report Template

```markdown
# Migration Verification Report

## Environment
- Source: {source_framework} @ localhost:{source_port}
- Target: {target_framework} @ localhost:{target_port}
- Date: {timestamp}
- Duration: {total_time}

## Summary
| Category | Passed | Failed | Rate |
|----------|--------|--------|------|
| Page Load | 25 | 0 | 100% |
| Visual Regression | 72 | 3 | 95.8% |
| Navigation | 45 | 1 | 97.8% |
| API Calls | 30 | 0 | 100% |
| Performance | 25 | 2 | 92% |
| Accessibility | 25 | 0 | 100% |
| Cross-Browser | 75 | 0 | 100% |
| **TOTAL** | **297** | **6** | **98.0%** |

## Visual Regression Details
[Link to HTML visual report]

## Failed Tests
1. /settings - visual diff 7.2% (threshold: 5%)
2. /dashboard - LCP regression 25% (budget: 20%)
...

## Recommendation
[PASSED/FAILED] - Ready for production / Needs attention
```

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

## Deliverables Summary

| Type | Count | Content |
|------|-------|---------|
| Command Modified | 1 | `migrate-4-verify.md` overhaul |
| Agent Modified | 1 | `e2e-tester.md` enhancement |
| New Skill | 1 | `jikime-workflow-playwright-migration/` |
| Example Code | 3-5 | Visual Regression, Cross-Browser, Route Discovery |
| Config Schema | 1 | `.migrate-config.yaml` extension |

## Total Tasks: 37

---

Version: 1.0.0
Created: 2026-01-24
Status: Planning
