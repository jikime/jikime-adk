# Testing Guidelines

Testing best practices with DDD (Domain-Driven Development) methodology.

## Coverage Targets

| Type | Target | Priority |
|------|--------|----------|
| Business Logic | 90%+ | Critical |
| API Endpoints | 80%+ | High |
| UI Components | 70%+ | Medium |
| Utilities | 80%+ | Medium |
| **Overall** | **80%+** | Required |

## Test Types

### 1. Unit Tests

Test individual functions and modules in isolation.

```typescript
// user.service.test.ts
describe('UserService', () => {
  describe('validateEmail', () => {
    it('should accept valid email', () => {
      expect(validateEmail('user@example.com')).toBe(true)
    })

    it('should reject invalid email', () => {
      expect(validateEmail('invalid')).toBe(false)
    })
  })
})
```

### 2. Integration Tests

Test interactions between components.

```typescript
// api/users.test.ts
describe('POST /api/users', () => {
  it('should create user with valid data', async () => {
    const response = await request(app)
      .post('/api/users')
      .send({ email: 'test@example.com', password: 'SecurePass1!' })

    expect(response.status).toBe(201)
    expect(response.body.data.email).toBe('test@example.com')
  })
})
```

### 3. E2E Tests

Test complete user flows.

```typescript
// e2e/auth.spec.ts
test('user can login and access dashboard', async ({ page }) => {
  await page.goto('/login')
  await page.fill('[name="email"]', 'user@example.com')
  await page.fill('[name="password"]', 'password')
  await page.click('button[type="submit"]')

  await expect(page).toHaveURL('/dashboard')
  await expect(page.locator('h1')).toContainText('Welcome')
})
```

## DDD Testing Approach

### ANALYZE → PRESERVE → IMPROVE

```markdown
1. ANALYZE
   - Run existing tests
   - Identify test coverage gaps
   - Understand current behavior

2. PRESERVE
   - Write characterization tests for uncovered code
   - Capture current behavior as baseline
   - Ensure no regressions

3. IMPROVE
   - Implement changes
   - Run all tests after each change
   - Add tests for new functionality
```

### Characterization Tests

When code lacks tests, write characterization tests first:

```typescript
// Capture existing behavior
describe('LegacyCalculator', () => {
  it('characterizes current behavior', () => {
    const calc = new LegacyCalculator()

    // Document what it actually does, not what it should do
    expect(calc.compute(10, 5)).toBe(15)  // Addition
    expect(calc.compute(-1, 5)).toBe(4)   // Handles negative
    expect(calc.compute(0, 0)).toBe(0)    // Zero case
  })
})
```

### Behavior Preservation

```typescript
// Before refactoring, ensure tests pass
describe('OrderService - Behavior Preservation', () => {
  const testCases = [
    { input: { items: [], discount: 0 }, expected: 0 },
    { input: { items: [{ price: 100 }], discount: 0 }, expected: 100 },
    { input: { items: [{ price: 100 }], discount: 10 }, expected: 90 },
  ]

  testCases.forEach(({ input, expected }) => {
    it(`calculates total for ${JSON.stringify(input)}`, () => {
      expect(calculateTotal(input)).toBe(expected)
    })
  })
})
```

## Test Quality Principles

### Good Tests Are

| Principle | Description |
|-----------|-------------|
| **Fast** | Run quickly, encourage frequent execution |
| **Isolated** | No dependencies between tests |
| **Repeatable** | Same result every time |
| **Self-validating** | Clear pass/fail, no manual check |
| **Timely** | Written close to the code change |

### Good Tests Have

```typescript
// Clear naming: should_expectedBehavior_when_condition
it('should return null when user not found', () => {
  const result = userService.findById('non-existent')
  expect(result).toBeNull()
})

// Meaningful assertions
expect(user.status).toBe('active')  // Good: clear intent
expect(user.status).toBeTruthy()   // Bad: unclear what's expected

// Single responsibility
it('should validate email format', () => {
  // Only test email validation, nothing else
})
```

## Test Organization

### File Structure

```
src/
├── services/
│   ├── user.service.ts
│   └── user.service.test.ts     # Co-located unit tests
├── api/
│   └── users/
│       ├── route.ts
│       └── route.test.ts
tests/
├── integration/                  # Integration tests
│   └── api.test.ts
└── e2e/                          # E2E tests
    └── auth.spec.ts
```

### Test Setup

```typescript
// Shared test utilities
// tests/setup.ts
import { beforeAll, afterAll, afterEach } from 'vitest'

beforeAll(async () => {
  await setupTestDatabase()
})

afterEach(async () => {
  await cleanupTestData()
})

afterAll(async () => {
  await teardownTestDatabase()
})
```

## Mocking Guidelines

### When to Mock

| Mock | Don't Mock |
|------|------------|
| External APIs | Your own code |
| Database (in unit tests) | Business logic |
| Time/Date | Pure functions |
| File system | Simple utilities |

### Mocking Examples

```typescript
// Mock external service
vi.mock('./email.service', () => ({
  sendEmail: vi.fn().mockResolvedValue({ success: true })
}))

// Mock time
vi.useFakeTimers()
vi.setSystemTime(new Date('2024-01-01'))

// Restore after test
afterEach(() => {
  vi.restoreAllMocks()
})
```

## Test Commands

```bash
# Run all tests
npm test

# Run with coverage
npm test -- --coverage

# Run specific file
npm test -- user.service.test.ts

# Watch mode
npm test -- --watch

# Run E2E
npm run test:e2e
```

## Testing Checklist

Before committing:

- [ ] All existing tests pass
- [ ] New code has tests
- [ ] Coverage maintained (80%+)
- [ ] No skipped tests without reason
- [ ] Tests are meaningful (not just for coverage)
- [ ] Edge cases covered
- [ ] Error scenarios tested

---

Version: 1.0.0
Methodology: DDD (ANALYZE-PRESERVE-IMPROVE)
Source: Adapted from everything-claude-code (TDD → DDD)
