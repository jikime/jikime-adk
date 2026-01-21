---
description: "[Step 3/4] Run tests. Unit tests, integration tests, coverage verification."
context: dev
---

# Development Step 3: Test

**Context**: @.claude/contexts/dev.md (Auto-loaded)

**테스트 단계**: 구현된 코드의 품질을 검증합니다.

**Note**: E2E 테스트는 `/jikime:e2e` 유틸리티 명령어를 사용하세요.

## Usage

```bash
# Run all tests
/jikime:dev-3-test

# Run with coverage
/jikime:dev-3-test --coverage

# Run specific test type
/jikime:dev-3-test --unit
/jikime:dev-3-test --integration

# Watch mode
/jikime:dev-3-test --watch
```

## Options

| Option | Description |
|--------|-------------|
| `--coverage` | Generate coverage report |
| `--unit` | Unit tests only |
| `--integration` | Integration tests only |
| `--watch` | Watch mode for continuous testing |
| `--fix` | Auto-fix failing tests if possible |

## Test Process

```
1. Run Existing Tests
   - Ensure no regressions
        ↓
2. Run New Tests
   - Tests for new functionality
        ↓
3. Coverage Check
   - Verify coverage targets
        ↓
4. Report Results
   - Pass: Proceed to review
   - Fail: Fix and retry
```

## Coverage Targets

| Type | Target |
|------|--------|
| Business Logic | 90%+ |
| API Endpoints | 80%+ |
| UI Components | 70%+ |
| Overall | 80%+ |

## Output

```markdown
## Test Results

### Summary
- Total: 68 tests
- Passed: 67 (98.5%)
- Failed: 1

### Unit Tests
✅ 45/45 passed (100%)

### Integration Tests
⚠️ 22/23 passed (95.6%)
❌ 1 failed: order/payment.test.ts:42

### Failed Test Details
\`\`\`
FAIL order/payment.test.ts
  ✕ should process refund correctly
    Expected: 100
    Received: 0
\`\`\`

### Coverage
| Category | Coverage | Target | Status |
|----------|----------|--------|--------|
| Statements | 85% | 80% | ✅ |
| Branches | 78% | 75% | ✅ |
| Functions | 82% | 80% | ✅ |

### Status: ⚠️ Fix failing test before review
```

## TDD Cycle (Optional)

```
RED → GREEN → REFACTOR

1. RED: Write failing test
2. GREEN: Write minimal code to pass
3. REFACTOR: Improve while keeping tests green
```

## Workflow

```
/jikime:dev-0-init   (선택적)
        ↓
/jikime:dev-1-plan
        ↓
/jikime:dev-2-implement
        ↓
/jikime:dev-3-test  ← 현재
        ↓
/jikime:dev-4-review
```

## Next Step

테스트 통과 후 다음 단계로:
```bash
/jikime:dev-4-review
```

## Related Utilities

- `/jikime:e2e` - E2E tests with Playwright
- `/jikime:build-fix` - Fix build errors

---

Version: 2.0.0
