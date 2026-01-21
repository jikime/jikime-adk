---
description: "[Step 1/4] Create implementation plan. Analyze requirements, assess risks, WAIT for confirmation."
context: planning
---

# Development Step 1: Plan

**Context**: @.claude/contexts/planning.md (Auto-loaded)

**계획 단계**: 요구사항을 분석하고 구현 계획을 수립합니다.

**Note**: 이 단계에서 사용자 확인을 받기 전까지 코드를 작성하지 않습니다.

## Usage

```bash
# Plan a new feature
/jikime:dev-1-plan Add user authentication

# Plan with SPEC document
/jikime:dev-1-plan --spec Add payment gateway

# Plan with context
/jikime:dev-1-plan @src/services/ Add caching layer

# Plan refactoring
/jikime:dev-1-plan Refactor order processing

# From existing SPEC
/jikime:dev-1-plan FEAT-001
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `[description]` | Feature or task description | Required |
| `@path` | Reference existing code | - |
| `--spec` | Create/update SPEC document | Off |
| `--detail` | More detailed breakdown | Off |
| `[FEAT-ID]` | Use existing SPEC document | - |

## Process

```
1. Analyze Request
   - Understand requirements
   - Identify affected areas
        ↓
2. Check SPEC (if --spec or FEAT-ID)
   - Create new SPEC document
   - Or load existing SPEC
        ↓
3. Create Plan
   - Break into phases
   - Identify dependencies
   - Assess risks
        ↓
4. Present & WAIT
   - Show plan to user
   - Wait for confirmation
        ↓
5. User Response
   - "yes" → Proceed to dev-2-implement
   - "modify: [changes]" → Revise plan
   - "no" → Cancel
```

## Plan Format

### Standard Plan

```markdown
# Implementation Plan: [Feature Name]

## Requirements
- [Requirement 1]
- [Requirement 2]

## Phases

### Phase 1: [Name]
- Task 1 (File: path/to/file.ts)
- Task 2

### Phase 2: [Name]
...

## Dependencies
- [Dependency 1]

## Risks
- HIGH: [Risk description]
- MEDIUM: [Risk description]

## Complexity: MEDIUM

**WAITING FOR CONFIRMATION**
Proceed? (yes/no/modify)
```

### With SPEC (--spec)

```markdown
# Implementation Plan: [Feature Name]
**SPEC**: .jikime/specs/FEAT-001.md

## SPEC Summary
- **ID**: FEAT-001
- **Status**: Planning
- **Created**: [Date]

## Requirements (from SPEC)
...

[Rest of standard plan]
```

## SPEC Integration

### Create New SPEC (--spec flag)

```bash
/jikime:dev-1-plan --spec Add user dashboard
```

Creates `.jikime/specs/FEAT-XXX.md` with:
- Requirements extracted from plan
- Acceptance criteria
- Risk assessment

### Use Existing SPEC

```bash
/jikime:dev-1-plan FEAT-001
```

Loads requirements from existing SPEC document.

## Critical Rules

1. **MUST WAIT** - 사용자 확인 전 코드 작성 금지
2. **Be Specific** - 파일 경로, 함수명 명시
3. **Consider Risks** - 잠재적 문제 식별
4. **SPEC Sync** - SPEC 사용 시 문서와 계획 동기화

## Workflow

```
/jikime:dev-0-init   (선택적)
        ↓
/jikime:dev-1-plan  ← 현재
        ↓
/jikime:dev-2-implement
        ↓
/jikime:dev-3-test
        ↓
/jikime:dev-4-review
```

## Next Step

승인 후 다음 단계로:
```bash
/jikime:dev-2-implement
# Or with SPEC
/jikime:dev-2-implement FEAT-001
```

---

Version: 2.0.0
