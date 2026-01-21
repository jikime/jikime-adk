---
description: "Generate and run E2E tests with Playwright. Create test journeys, capture screenshots/videos on failure."
---

# E2E

Playwright로 E2E 테스트를 생성하고 실행합니다.

## Usage

```bash
# Generate E2E test for a flow
/jikime:e2e Test login flow

# Run existing E2E tests
/jikime:e2e --run

# Run specific test
/jikime:e2e --run @tests/e2e/auth.spec.ts

# Debug mode
/jikime:e2e --run --debug
```

## Options

| Option | Description |
|--------|-------------|
| `[description]` | User flow to test |
| `--run` | Run existing tests |
| `--debug` | Debug mode (headed browser) |
| `--headed` | Show browser window |

## Test Generation

```typescript
// Generated: tests/e2e/login.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Login Flow', () => {
  test('user can login with credentials', async ({ page }) => {
    // 1. Navigate to login page
    await page.goto('/login')

    // 2. Fill credentials
    await page.fill('[data-testid="email"]', 'test@example.com')
    await page.fill('[data-testid="password"]', 'password123')

    // 3. Submit form
    await page.click('[data-testid="submit"]')

    // 4. Verify redirect
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('h1')).toContainText('Dashboard')
  })
})
```

## Best Practices

### DO ✅
- `data-testid` 속성 사용
- API 응답 대기 (timeout 아님)
- Page Object Model 패턴
- 핵심 사용자 여정 테스트

### DON'T ❌
- CSS 클래스로 선택 (변경됨)
- 구현 세부사항 테스트
- 프로덕션 환경 테스트
- 모든 엣지케이스 E2E로 (유닛 테스트 사용)

## Artifacts

테스트 실패 시 자동 캡처:
- 📸 Screenshot
- 📹 Video recording
- 🔍 Trace file (step-by-step)

```bash
# View trace
npx playwright show-trace artifacts/trace.zip

# View report
npx playwright show-report
```

## Output

```markdown
## E2E Test Results

### Summary
- Total: 5 tests
- Passed: 4 (80%)
- Failed: 1
- Duration: 12.3s

### Failed Tests
❌ login.spec.ts:15 - user can login
   Error: Timeout waiting for '[data-testid="submit"]'
   Screenshot: artifacts/login-failure.png

### Artifacts
📸 Screenshots: 2 files
📹 Videos: 1 file
📊 HTML Report: playwright-report/index.html
```

## Quick Commands

```bash
# Install Playwright
npx playwright install

# Run all tests
npx playwright test

# Run headed
npx playwright test --headed

# Generate test code
npx playwright codegen http://localhost:3000
```

## Critical Flows to Test

**필수 (반드시 통과):**
1. 로그인/로그아웃
2. 회원가입
3. 핵심 기능 flow

**중요:**
1. 사용자 프로필
2. 설정 변경
3. 반응형 레이아웃

## Related Commands

- `/jikime:test` - Unit/Integration tests
- `/jikime:plan` - Identify flows to test
- `/jikime:review` - Review test code

---

Version: 1.0.0
