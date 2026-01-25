---
description: "Comprehensive quality verification - build, type, lint, test, security in one command"
argument-hint: "[quick|standard|full|pre-pr|--fix|--json|--ci|--incremental]"
type: utility
allowed-tools: Task, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

# JikiME-ADK Utility: Comprehensive Verification

Unified quality gate verification integrated with LSP Quality Gates and TRUST 5 framework.

Verification mode: $ARGUMENTS

---

## Core Philosophy

```
Single command for complete quality assurance:
├─ Build verification (compilation success)
├─ Type checking (TypeScript/Pyright/etc.)
├─ Lint validation (ESLint/Ruff/etc.)
├─ Test execution (with coverage)
├─ Security scanning (secrets, vulnerabilities)
├─ LSP Quality Gates (zero regression policy)
└─ TRUST 5 compliance check
```

---

## Usage

```bash
# Standard verification (recommended)
/jikime:verify

# Quick check (build + types only)
/jikime:verify quick

# Full verification (all checks + deps)
/jikime:verify full

# Pre-PR verification (full + security scan)
/jikime:verify pre-pr

# With auto-fix attempt
/jikime:verify --fix

# CI/CD mode (exit codes)
/jikime:verify --ci

# JSON output for automation
/jikime:verify --json

# Only check changed files
/jikime:verify --incremental
```

---

## Verification Profiles

| Profile | Checks | Use Case |
|---------|--------|----------|
| `quick` | Build, Types | During active development |
| `standard` | Build, Types, Lint, Tests | Default, after changes |
| `full` | All + Deps, Coverage | Before major commits |
| `pre-pr` | Full + Security + Secrets | Before creating PR |

---

## Verification Phases

### Phase 1: Build Check

```bash
# Auto-detect and run build
npm run build 2>&1 | tail -30
# OR
pnpm build 2>&1 | tail -30
# OR
cargo build 2>&1 | tail -30
```

**Gate**: FAIL → Stop immediately, show errors

### Phase 2: Type Check

```bash
# TypeScript
npx tsc --noEmit 2>&1

# Python
pyright . 2>&1

# Go
go vet ./... 2>&1
```

**Gate**: Errors → Must fix before PR

### Phase 3: Lint Check

```bash
# JavaScript/TypeScript
npm run lint 2>&1

# With auto-fix
npm run lint -- --fix 2>&1

# Python
ruff check . 2>&1
```

**Gate**: Errors → Must fix. Warnings → Document if needed.

### Phase 4: Test Suite

```bash
# Run with coverage
npm test -- --coverage 2>&1

# Report format
Coverage: X% (target: 80%)
Tests: X passed, Y failed
```

**Gate**: Failures → Must fix. Coverage < 80% → Warning.

### Phase 5: Security Scan

```bash
# Secret detection
grep -rn "sk-\|api_key\|password\s*=" --include="*.ts" src/

# Dependency vulnerabilities
npm audit --production

# Console.log detection
grep -rn "console.log" --include="*.ts" src/
```

**Gate**: Secrets found → CRITICAL. Vulnerabilities → Review.

### Phase 6: LSP Quality Gates

Check against `.jikime/config/quality.yaml` thresholds:

```yaml
lsp_quality_gates:
  run:
    max_errors: 0
    max_type_errors: 0
    max_lint_errors: 0
  sync:
    max_warnings: 10
```

**Gate**: Regression from baseline → Block PR.

### Phase 7: TRUST 5 Compliance

```markdown
[T] Tested:    Coverage > 80%, critical paths tested
[R] Readable:  No complexity warnings, clear naming
[U] Unified:   Consistent patterns, follows architecture
[S] Secured:   No vulnerabilities, input validated
[T] Trackable: Structured logging, error context
```

### Phase 8: Adversarial Review (pre-pr, full profiles only)

**Purpose**: Multi-angle validation to reduce false positives and catch missed issues.

```
┌─────────────────────────────────────────────────────────────┐
│                   ADVERSARIAL REVIEW LAYER                  │
├─────────────────────────────────────────────────────────────┤
│  Subagent 1: False Positive Filter                          │
│  ├─ Review all warnings/errors from Phases 1-7              │
│  ├─ Identify false positives (intentional patterns,         │
│  │   test fixtures, generated code, third-party)            │
│  └─ Output: Filtered list with confidence scores            │
├─────────────────────────────────────────────────────────────┤
│  Subagent 2: Missing Issues Finder                          │
│  ├─ Analyze code changes with fresh perspective             │
│  ├─ Look for edge cases, race conditions, error handling    │
│  ├─ Check boundary conditions and null safety               │
│  └─ Output: Additional issues not caught by standard tools  │
├─────────────────────────────────────────────────────────────┤
│  Subagent 3: Context Validator                              │
│  ├─ Compare findings against original intent/requirements   │
│  ├─ Verify changes don't break existing functionality       │
│  ├─ Check if suggested fixes align with codebase patterns   │
│  └─ Output: Contextual assessment with recommendations      │
└─────────────────────────────────────────────────────────────┘
```

**Execution**: All 3 subagents run in PARALLEL (single Task message):

```markdown
## Adversarial Review

### False Positive Analysis
| Finding | Verdict | Reason |
|---------|---------|--------|
| `unused import` in test.ts | FALSE POSITIVE | Test fixture |
| `any` type warning | VALID | Should be typed |

### Missing Issues Found
1. Race condition in `async updateUser()` - no mutex
2. Missing null check at `data.items[0]`

### Context Validation
- Changes align with PR description: ✅
- Pattern consistency maintained: ✅
- Suggested fixes are safe: ✅
```

**Gate**: Adversarial findings integrated into final report with severity adjustment.

---

## Output Format

### J.A.R.V.I.S. Format

```markdown
## J.A.R.V.I.S.: Verification Report

### Quick Summary
| Check | Status | Details |
|-------|--------|---------|
| Build | ✅ PASS | 0 errors |
| Types | ✅ PASS | 0 errors |
| Lint | ⚠️ WARN | 3 warnings |
| Tests | ✅ PASS | 47/47 (98% coverage) |
| Security | ✅ PASS | 0 issues |
| LSP Gates | ✅ PASS | No regression |
| TRUST 5 | ✅ PASS | All principles met |

### Overall: ✅ READY FOR PR

### Warnings to Address
1. `src/utils/helper.ts:42` - Unused variable (lint)
2. `src/api/handler.ts:18` - Consider extracting function

### Predictive Suggestions
- Consider adding E2E test for new auth flow
- Review error handling in payment module

### Adversarial Review (pre-pr/full only)
| Subagent | Findings |
|----------|----------|
| False Positive Filter | 2 warnings filtered (test fixtures) |
| Missing Issues Finder | 1 race condition detected |
| Context Validator | Changes align with intent ✅ |

**Adjusted Issues**: 3 warnings → 1 warning (after filtering)
**New Issues**: 1 (race condition in async handler)
```

### F.R.I.D.A.Y. Format

```markdown
## F.R.I.D.A.Y.: Migration Verification

### Module Status
| Module | Build | Types | Tests | Status |
|--------|-------|-------|-------|--------|
| Auth | ✅ | ✅ | ✅ | VERIFIED |
| Users | ✅ | ✅ | ✅ | VERIFIED |
| Products | ✅ | ⚠️ 2 | ✅ | NEEDS FIX |

### Migration Progress: 8/10 modules verified

### Blocking Issues
1. Products module: 2 type errors in migration
   - `ProductDTO.price`: Expected number, got string
   - `Product.category`: Missing property

### Next Steps
1. Fix type errors in Products module
2. Run: /jikime:verify --incremental
```

---

## Auto-Fix Mode (--fix)

When `--fix` is used:

1. Run `eslint --fix` for auto-fixable lint issues
2. Run `prettier --write` for formatting
3. Re-run verification to confirm fixes
4. Report remaining issues that need manual attention

```markdown
## Auto-Fix Results

### Fixed Automatically
- 12 lint issues (formatting, imports)
- 3 unused imports removed

### Requires Manual Fix
- `src/auth.ts:42` - Type error: string vs number
- `src/api.ts:18` - Unused function (intentional?)

Re-verification: Build ✅ | Types ❌ (1 error) | Lint ✅
```

---

## Incremental Mode (--incremental)

Only verify changed files since last commit:

```bash
/jikime:verify --incremental

# Checks only:
git diff --name-only HEAD~1 | xargs -I {} verify {}
```

Benefits:
- 10x faster for large codebases
- Immediate feedback during development
- Full verification still recommended pre-PR

---

## CI/CD Integration

### Exit Codes (--ci)

| Code | Meaning |
|------|---------|
| 0 | All checks passed |
| 1 | Build failed |
| 2 | Type errors |
| 3 | Lint errors (not warnings) |
| 4 | Test failures |
| 5 | Security issues (critical) |
| 6 | LSP regression detected |

### JSON Output (--json)

```json
{
  "timestamp": "2024-01-22T10:30:00Z",
  "profile": "pre-pr",
  "status": "PASS",
  "checks": {
    "build": {"status": "pass", "errors": 0, "duration_ms": 2340},
    "types": {"status": "pass", "errors": 0, "warnings": 2},
    "lint": {"status": "pass", "errors": 0, "warnings": 5},
    "tests": {"status": "pass", "total": 47, "passed": 47, "coverage": 98.2},
    "security": {"status": "pass", "secrets": 0, "vulnerabilities": 0},
    "lsp": {"status": "pass", "regression": false},
    "trust5": {"status": "pass", "score": 5}
  },
  "ready_for_pr": true,
  "warnings": [
    {"file": "src/utils.ts", "line": 42, "message": "Unused variable"}
  ],
  "adversarial_review": {
    "false_positives_filtered": 2,
    "missing_issues_found": 1,
    "context_validated": true,
    "findings": [
      {"type": "missing", "severity": "medium", "message": "Race condition in async handler", "file": "src/api.ts", "line": 78}
    ]
  }
}
```

### GitHub Actions Integration

```yaml
- name: Run Verification
  run: |
    jikime-adk verify pre-pr --ci --json > verify-results.json
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
      echo "Verification failed with code $exit_code"
      cat verify-results.json | jq '.warnings'
      exit $exit_code
    fi
```

---

## LSP Quality Gates Integration

Reads from `.jikime/config/quality.yaml`:

```yaml
lsp_quality_gates:
  enabled: true

  plan:
    require_baseline: true

  run:
    max_errors: 0
    max_type_errors: 0
    max_lint_errors: 0
    allow_regression: false

  sync:
    max_errors: 0
    max_warnings: 10
    require_clean_lsp: true
```

### Baseline Comparison

```
Previous State (baseline):
  Errors: 0, Type Errors: 0, Warnings: 5

Current State:
  Errors: 0, Type Errors: 0, Warnings: 7

Regression: +2 warnings (within threshold)
Status: PASS (no error regression)
```

---

## TRUST 5 Integration

Each principle is verified:

| Principle | Checks | Target |
|-----------|--------|--------|
| **T**ested | Coverage, test count | 80%+ coverage |
| **R**eadable | Complexity, naming | No warnings |
| **U**nified | Pattern consistency | Architecture compliance |
| **S**ecured | Vulnerabilities, secrets | Zero critical |
| **T**rackable | Logging, error handling | Structured logs |

---

## EXECUTION DIRECTIVE

1. Detect project type (npm, pnpm, cargo, go, etc.)
2. Parse profile from $ARGUMENTS (default: standard)
3. Execute phases in order:
   - Build → Type → Lint → Test → Security → LSP → TRUST 5
4. Stop on critical failures (build, types in strict mode)
5. **For `pre-pr` and `full` profiles**: Execute Adversarial Review (Phase 8)
   - Launch 3 subagents in PARALLEL (single Task message):
     - False Positive Filter: Review all findings, identify false positives
     - Missing Issues Finder: Fresh perspective analysis for edge cases
     - Context Validator: Compare findings against intent and patterns
   - Collect results with TaskOutput
   - Integrate adversarial findings into final report
   - Adjust severity based on adversarial consensus
6. Aggregate results into report
7. Use orchestrator-appropriate format
8. If `--fix`, attempt auto-fixes and re-verify
9. If `--ci`, set exit code based on results
10. If `--json`, output JSON instead of markdown

Execute NOW. Do NOT just describe.

---

## Related Commands

- `/jikime:test` - Run tests only
- `/jikime:build-fix` - Fix build errors
- `/jikime:security` - Deep security analysis
- `/jikime:eval` - Eval-driven verification
- `/jikime:loop` - Iterative fix loop

---

Version: 1.1.0
Type: Utility Command (Type B)
Integration: LSP Quality Gates, TRUST 5, Adversarial Review
