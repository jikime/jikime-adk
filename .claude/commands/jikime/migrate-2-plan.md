---
description: "[Step 2/4] 마이그레이션 계획 수립. 단계 정의, 작업량 추정, 위험 식별. 승인 후 진행."
---

# Migration Step 2: Plan

**계획 단계**: 마이그레이션 계획을 수립합니다.

## What This Command Does

1. **Phase Definition** - 단계별 마이그레이션 계획
2. **Effort Estimation** - 작업량 및 기간 추정
3. **Risk Assessment** - 잠재적 위험 요소 식별
4. **Wait for Approval** - 사용자 승인 후 진행

## Usage

```bash
# Create plan based on analysis
/jikime:migrate-2-plan

# Plan with specific source and target
/jikime:migrate-2-plan source:php target:nextjs

# Plan specific modules only
/jikime:migrate-2-plan --modules auth,users,orders
```

## Options

| Option | Description |
|--------|-------------|
| `source:<lang>` | Source language/framework |
| `target:<lang>` | Target language/framework |
| `--modules` | Specific modules to migrate |
| `--incremental` | Plan for incremental migration |

## Output

```markdown
# Migration Plan: PHP → Next.js

## Phase 1: Database Layer (3 days)
- PDO → Prisma ORM
- MySQL schema migration
- Data validation with Zod

## Phase 2: API Endpoints (5 days)
- Laravel Controllers → API Routes
- Authentication → NextAuth.js
- Middleware migration

## Phase 3: Frontend (4 days)
- Blade + jQuery → React Components
- Asset pipeline migration
- Styling to Tailwind CSS

## Phase 4: Testing & Verification (2 days)
- Characterization tests
- E2E comparison tests
- Performance validation

## Total Estimated: 14 days

## Risks
- HIGH: Payment integration complexity
- MEDIUM: Session handling differences

**WAITING FOR CONFIRMATION**: Proceed? (yes/no/modify)
```

## Important

**승인 전까지 코드를 작성하지 않습니다.**

응답 방법:
- `yes` - 계획대로 진행
- `modify: [변경사항]` - 계획 수정
- `no` - 취소

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Planning | `planner` | Migration strategy |
| Architecture | `architect` | System design |

## Workflow

```
/jikime:migrate-0-discover
        ↓
/jikime:migrate-1-analyze
        ↓
/jikime:migrate-2-plan  ← 현재
        ↓
/jikime:migrate-3-execute
        ↓
/jikime:migrate-4-verify
```

## Next Step

승인 후 다음 단계로:
```bash
/jikime:migrate-3-execute
```

---

Version: 2.1.0
