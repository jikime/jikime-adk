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
