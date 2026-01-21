---
description: "[Step 4/4] Code review. Quality check, security scan, best practices verification."
context: review
---

# Development Step 4: Review

**Context**: @.claude/contexts/review.md (Auto-loaded)

**리뷰 단계**: 코드 품질과 보안을 최종 검토합니다.

## Usage

```bash
# Review all changes
/jikime:dev-4-review

# Review with automated quality gates
/jikime:dev-4-review --quality

# Review specific files
/jikime:dev-4-review @src/services/

# Review with focus
/jikime:dev-4-review --focus security
/jikime:dev-4-review --focus quality

# Full enterprise review
/jikime:dev-4-review --quality --strict
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `@path` | Specific files to review | All changes |
| `--quality` | Run automated quality gates | Off |
| `--focus` | Focus: security, quality, performance | All |
| `--staged` | Review staged changes only | Off |
| `--strict` | Apply stricter criteria | Off |

## Review Process

### Standard Review

```
1. Quality Check
   - Code standards
   - Best practices
        ↓
2. Security Scan
   - Vulnerability check
   - Secrets detection
        ↓
3. Performance Review
   - Bottleneck identification
   - Optimization suggestions
        ↓
4. Final Report
   - Approval status
   - Action items
```

### With Quality Gates (--quality)

```
1. Automated Checks
   ┌─────────────────────────────────────────┐
   │           Quality Verification          │
   ├─────────────────────────────────────────┤
   │  ✓ Lint Check      → eslint/biome       │
   │  ✓ Type Check      → tsc --noEmit       │
   │  ✓ Unit Tests      → vitest/jest/pytest │
   │  ✓ Security Scan   → basic checks       │
   └─────────────────────────────────────────┘
        ↓
2. Results Analysis
   - Pass: Continue to manual review
   - Fail: Fix required before proceeding
        ↓
3. Manual Review
   - Code quality assessment
   - Architecture review
        ↓
4. Final Report
```

## Quality Gates (--quality)

### Automated Checks

| Check | Tool | Pass Criteria |
|-------|------|---------------|
| Lint | eslint/biome/ruff | No errors |
| Type | tsc/mypy/pyright | No errors |
| Test | vitest/jest/pytest | All pass |
| Security | basic scan | No critical issues |

### Check Detection

프로젝트 설정에 따라 자동 감지:

```yaml
# JavaScript/TypeScript
- package.json의 lint/test 스크립트
- eslint.config.js, biome.json
- tsconfig.json

# Python
- pyproject.toml의 ruff/mypy 설정
- pytest.ini, setup.cfg
```

### Quality Gate Output

```markdown
## Quality Gate Results

| Check | Status | Details |
|-------|--------|---------|
| Lint | ✅ Pass | 0 errors, 2 warnings |
| Type | ✅ Pass | No type errors |
| Test | ✅ Pass | 42 passed, 0 failed |
| Security | ⚠️ Warning | 1 low severity issue |

**Overall: ✅ PASS** (Ready for manual review)
```

## Review Categories

### Security (CRITICAL)
- Hardcoded credentials
- SQL injection risks
- XSS vulnerabilities
- Missing input validation

### Quality (HIGH)
- Large functions (>50 lines)
- Deep nesting (>4 levels)
- Missing error handling
- Code duplication

### Best Practices (MEDIUM)
- Missing tests
- Poor naming conventions
- Inconsistent patterns

## Output

### Standard Review Report

```markdown
## Code Review Report

### Summary
- Files Reviewed: 12
- Issues Found: 5
- Status: ⚠️ Warning

### Issues by Severity

#### CRITICAL (0)
None

#### HIGH (2)
1. Missing error handling
   - File: src/services/order.ts:78
   - Fix: Add try/catch block

2. Large function (68 lines)
   - File: src/utils/parser.ts:45
   - Fix: Split into smaller functions

#### MEDIUM (3)
1. Missing test coverage
   - File: src/api/users.ts
   - Fix: Add unit tests

### Security Scan
✅ No vulnerabilities detected

### Approval Status
⚠️ **Fix HIGH issues before merge**
✅ Ready after fixes
```

### With Quality Gates Report

```markdown
## Code Review Report

### Quality Gates
| Check | Status | Details |
|-------|--------|---------|
| Lint | ✅ Pass | Clean |
| Type | ✅ Pass | Clean |
| Test | ✅ Pass | 100% pass rate |
| Security | ✅ Pass | No issues |

**Quality Gates: ✅ ALL PASSED**

### Manual Review
[Standard review content...]

### Final Status
✅ **Ready for merge**
```

## Severity Actions

| Level | Action | --strict |
|-------|--------|----------|
| CRITICAL | Must fix immediately | Block |
| HIGH | Fix before merge | Block |
| MEDIUM | Should fix | Warning → Block |
| LOW | Optional improvement | Warning |

## Workflow

```
/jikime:dev-0-init   (선택적)
        ↓
/jikime:dev-1-plan
        ↓
/jikime:dev-2-implement
        ↓
/jikime:dev-3-test
        ↓
/jikime:dev-4-review  ← 현재 (마지막)
```

## Development Complete!

개발이 완료되었습니다!

**다음 단계:**
1. 커밋 및 PR 생성
2. CI/CD 파이프라인 확인
3. 배포

**필요시 유틸리티 명령어 사용:**
- `/jikime:security` - 심층 보안 감사
- `/jikime:docs` - 문서 업데이트
- `/jikime:e2e` - E2E 테스트

## Enterprise Mode Note

`/jikime:dev --enterprise` 사용 시 `--quality`가 자동 활성화됩니다.

---

Version: 2.0.0
