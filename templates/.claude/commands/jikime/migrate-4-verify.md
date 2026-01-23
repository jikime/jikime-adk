---
description: "[Step 4/4] Migration verification. Run tests, compare behavior, generate final report."
argument-hint: '[--full] [--behavior] [--e2e] [--performance] [--source-url URL] [--target-url URL]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Glob, Grep
model: inherit
---

# Migration Step 4: Verify

**Verification Phase**: Validate migration success.

## CRITICAL: Input Sources

**Project settings are automatically read from `.migrate-config.yaml`.**

### Required Inputs (from Previous Steps)

1. **`.migrate-config.yaml`** - artifacts_dir, output_dir, source/target framework
2. **`{artifacts_dir}/progress.yaml`** - Migration progress status (Step 3 output)
3. **`{output_dir}/`** - Migrated project (Step 3 output)

## What This Command Does

1. **Characterization Tests** - Run behavior preservation tests
2. **Behavior Comparison** - Compare source/target outputs
3. **E2E Testing** - Validate full user flows
4. **Performance Check** - Comparative performance analysis
5. **Final Report** - Comprehensive verification report

## Usage

```bash
# Verify current migration (reads all from config)
/jikime:migrate-4-verify

# Verify with all checks
/jikime:migrate-4-verify --full

# Verify specific aspects
/jikime:migrate-4-verify --behavior
/jikime:migrate-4-verify --e2e
/jikime:migrate-4-verify --performance

# Compare live systems (optional: for running instances)
/jikime:migrate-4-verify --source-url http://old.local --target-url http://new.local
```

## Options

| Option | Description |
|--------|-------------|
| `--full` | Run all verification types |
| `--behavior` | Behavior comparison only |
| `--e2e` | E2E tests only |
| `--performance` | Performance comparison only |
| `--source-url` | Source system URL (for live comparison) |
| `--target-url` | Target system URL (for live comparison) |

**Note**: `--source-url` and `--target-url` are for comparing **live running instances**. They are NOT the source/target frameworks (those come from `.migrate-config.yaml`).

## Verification Types

### 1. Characterization Tests
```
Running characterization tests...

auth/login.test.ts          ✅ 12/12 passed
auth/logout.test.ts         ✅ 5/5 passed
users/crud.test.ts          ✅ 18/18 passed
orders/calculate.test.ts    ⚠️ 9/10 passed (1 improved)
```

### 2. Behavior Comparison
```
GET /api/users     ✅ Identical response
POST /api/orders   ✅ Identical response
GET /api/products  ✅ Identical response
```

### 3. E2E Tests
```
Login Flow         ✅ Passed
Checkout Flow      ✅ Passed
User Registration  ✅ Passed
```

### 4. Performance
```
| Metric      | Source | Target | Change |
|-------------|--------|--------|--------|
| Avg Response| 250ms  | 80ms   | -68%   |
| Throughput  | 100/s  | 350/s  | +250%  |
```

## Final Report

```markdown
# Migration Verification Report

## Summary
| Category | Passed | Failed | Rate |
|----------|--------|--------|------|
| Characterization | 148 | 2 | 98.7% |
| Behavior | 45 | 0 | 100% |
| E2E | 19 | 1 | 95% |
| **Total** | **212** | **3** | **98.6%** |

## Status: ✅ PASSED

## Known Differences (Intentional)
1. Improved error messages
2. Better validation responses

## Performance Gains
- 68% faster response times
- 250% higher throughput

## Recommendation
✅ Ready for production deployment
```

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Behavior Validation | `behavior-validator` | Compare source/target |
| E2E Testing | `e2e-runner` | Playwright tests |
| Security Review | `security-reviewer` | Vulnerability check |

## Workflow (Data Flow)

```
/jikime:migrate-0-discover
        ↓ (.migrate-config.yaml created)
/jikime:migrate-1-analyze
        ↓ (config updated + as_is_spec.md)
/jikime:migrate-2-plan
        ↓ (migration_plan.md)
/jikime:migrate-3-execute
        ↓ (output_dir/ + progress.yaml)
/jikime:migrate-4-verify  ← current (final)
        │
        ├─ Reads: .migrate-config.yaml (paths, frameworks)
        ├─ Reads: {artifacts_dir}/progress.yaml (migration status)
        ├─ Tests: {output_dir}/ (migrated project)
        ├─ Creates: {artifacts_dir}/verification_report.md
```

## Migration Complete!

Migration is complete when verification passes.

**Next Steps:**
1. Deploy to staging environment
2. User Acceptance Testing (UAT)
3. Production deployment

---

Version: 3.0.0
Changelog:
- v3.0.0: Config-first approach; Renamed --source/--target to --source-url/--target-url for clarity; Added data flow diagram
- v2.1.0: Initial verification command
