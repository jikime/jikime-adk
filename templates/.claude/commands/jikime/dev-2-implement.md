---
description: "[Step 2/4] Implement features. Write code following the approved plan with DDD methodology."
context: dev
---

# Development Step 2: Implement

**Context**: @.claude/contexts/dev.md (Auto-loaded)

**구현 단계**: 승인된 계획에 따라 코드를 작성합니다.

**Methodology**: DDD (ANALYZE → PRESERVE → IMPROVE)

## Usage

```bash
# Implement approved plan
/jikime:dev-2-implement

# Implement specific phase
/jikime:dev-2-implement --phase 1

# Implement with specific focus
/jikime:dev-2-implement --focus backend
/jikime:dev-2-implement --focus frontend
```

## Options

| Option | Description |
|--------|-------------|
| `--phase` | Implement specific phase only |
| `--focus` | Focus area: backend, frontend, api |
| `--dry-run` | Show what would be done |

## DDD Cycle

```
ANALYZE → PRESERVE → IMPROVE

1. ANALYZE: Understand existing behavior
2. PRESERVE: Create characterization tests
3. IMPROVE: Implement new code
```

## Implementation Process

```
1. Review Plan
   - Load approved plan from dev-1-plan
   - Verify requirements
        ↓
2. ANALYZE
   - Understand affected code
   - Identify dependencies
        ↓
3. PRESERVE
   - Run existing tests
   - Create characterization tests if needed
        ↓
4. IMPROVE
   - Write implementation code
   - Follow coding standards
        ↓
5. Validate
   - Run tests
   - Check for regressions
```

## Progress Display

```
╔══════════════════════════════════════════════════════════╗
║  Implementation Progress                                  ║
║  Phase: 2 of 3                                           ║
║  Current: user-service                                   ║
║  Progress: [████████████░░░░░░░░] 60%                   ║
╚══════════════════════════════════════════════════════════╝

✅ Phase 1: Database schema - completed
🔄 Phase 2: API endpoints - in progress
⏳ Phase 3: Frontend components - pending
```

## Output

```markdown
## Implementation Progress

### Completed
- ✅ Created UserService class
- ✅ Added API endpoints
- ✅ Database migrations

### Generated Files
- src/services/user.service.ts
- src/api/users/route.ts
- prisma/migrations/001_users.sql

### Tests
- 12 tests created
- 12 passing

### Next: Run /jikime:dev-3-test
```

## Quality Standards

- [ ] Code follows project conventions
- [ ] No hardcoded values
- [ ] Error handling implemented
- [ ] Input validation added
- [ ] Tests written for new code

## Workflow

```
/jikime:dev-0-init   (선택적)
        ↓
/jikime:dev-1-plan
        ↓
/jikime:dev-2-implement  ← 현재
        ↓
/jikime:dev-3-test
        ↓
/jikime:dev-4-review
```

## Next Step

구현 완료 후 다음 단계로:
```bash
/jikime:dev-3-test
```

---

Version: 2.0.0
Methodology: DDD (ANALYZE-PRESERVE-IMPROVE)
