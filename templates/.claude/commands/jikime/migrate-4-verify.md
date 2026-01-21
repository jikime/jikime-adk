---
description: "[Step 4/4] 마이그레이션 검증. 테스트 실행, 동작 비교, 최종 보고서 생성."
---

# Migration Step 4: Verify

**검증 단계**: 마이그레이션 성공을 검증합니다.

## What This Command Does

1. **Characterization Tests** - 동작 보존 테스트 실행
2. **Behavior Comparison** - 소스/타겟 출력 비교
3. **E2E Testing** - 전체 사용자 흐름 검증
4. **Performance Check** - 성능 비교 분석
5. **Final Report** - 종합 검증 보고서

## Usage

```bash
# Verify current migration
/jikime:migrate-4-verify

# Verify with all checks
/jikime:migrate-4-verify --full

# Verify specific aspects
/jikime:migrate-4-verify --behavior
/jikime:migrate-4-verify --e2e
/jikime:migrate-4-verify --performance

# Compare live systems
/jikime:migrate-4-verify --source http://old.local --target http://new.local
```

## Options

| Option | Description |
|--------|-------------|
| `--full` | Run all verification types |
| `--behavior` | Behavior comparison only |
| `--e2e` | E2E tests only |
| `--performance` | Performance comparison only |
| `--source` | Source system URL |
| `--target` | Target system URL |

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

## Workflow

```
/jikime:migrate-0-discover
        ↓
/jikime:migrate-1-analyze
        ↓
/jikime:migrate-2-plan
        ↓
/jikime:migrate-3-execute
        ↓
/jikime:migrate-4-verify  ← 현재 (마지막)
```

## Migration Complete!

마이그레이션이 완료되었습니다! 🎉

**다음 단계:**
1. 스테이징 환경에 배포
2. 사용자 승인 테스트 (UAT)
3. 프로덕션 배포

**필요시 유틸리티 커맨드 사용:**
- `/jikime:build-fix` - 빌드 에러 수정
- `/jikime:review` - 코드 리뷰
- `/jikime:docs` - 문서 업데이트

---

Version: 2.1.0
