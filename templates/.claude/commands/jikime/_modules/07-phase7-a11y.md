
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

