---
description: "[Step 3/4] DDD 방법론으로 마이그레이션 실행. ANALYZE → PRESERVE → IMPROVE 사이클."
---

# Migration Step 3: Execute

**실행 단계**: DDD 방법론으로 실제 마이그레이션을 실행합니다.

## What This Command Does

### DDD Cycle: ANALYZE → PRESERVE → IMPROVE

1. **ANALYZE** - 기존 코드 동작 이해
2. **PRESERVE** - 특성 테스트로 동작 보존
3. **IMPROVE** - 새로운 코드로 변환
4. **Repeat** - 모듈별 반복

## Usage

```bash
# Execute migration (uses plan from step 2)
/jikime:migrate-3-execute

# Migrate specific module
/jikime:migrate-3-execute --module auth

# Migrate with explicit source/target
/jikime:migrate-3-execute source:php target:nextjs

# Resume interrupted migration
/jikime:migrate-3-execute --resume
```

## Options

| Option | Description |
|--------|-------------|
| `source:<lang>` | Source language/framework |
| `target:<lang>` | Target language/framework |
| `--module` | Migrate specific module only |
| `--resume` | Resume from last checkpoint |
| `--dry-run` | Show what would be done |

## Supported Migrations

| Source | Target Options |
|--------|----------------|
| PHP | Next.js, FastAPI, Go, Spring Boot |
| jQuery | React, Vue, Svelte |
| Java Servlet | Spring Boot, Go, FastAPI |
| Python 2 | Python 3, FastAPI |
| Legacy C++ | Modern C++20, Rust |

## Progress Display

```
╔══════════════════════════════════════════════════════════╗
║  Migration: MIG-2026-001                                  ║
║  Phase: IMPROVE                                           ║
║  Module: user-service                                     ║
║  Progress: [████████████░░░░░░░░] 60%                    ║
╚══════════════════════════════════════════════════════════╝

✅ ANALYZE: user-service - completed
✅ PRESERVE: 23 characterization tests created
🔄 IMPROVE: generating target code...
```

## Output

```markdown
## Migration Progress: MIG-2026-001

### Completed Modules
- ✅ auth (15 files → 8 files)
- ✅ users (12 files → 6 files)
- 🔄 orders (in progress)
- ⏳ payments (pending)

### Generated Files
- src/app/api/auth/route.ts
- src/lib/services/user.service.ts
- src/components/LoginForm.tsx
...

### Characterization Tests
- 50 tests created
- 48 passing
- 2 pending review

### Next: Run /jikime:migrate-4-verify
```

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Analysis | `source-analyzer` | Legacy code understanding |
| Test Creation | `tdd-guide` | Characterization tests |
| Code Generation | `target-generator` | Modern code creation |
| Review | `code-reviewer` | Quality check |

## Workflow

```
/jikime:migrate-0-discover
        ↓
/jikime:migrate-1-analyze
        ↓
/jikime:migrate-2-plan
        ↓
/jikime:migrate-3-execute  ← 현재
        ↓
/jikime:migrate-4-verify
```

## Next Step

실행 후 다음 단계로:
```bash
/jikime:migrate-4-verify
```

---

Version: 2.1.0
Methodology: DDD (ANALYZE-PRESERVE-IMPROVE)
