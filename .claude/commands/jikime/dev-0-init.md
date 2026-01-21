---
description: "[Step 0/4] Initialize development. Setup workspace, create SPEC doc (optional), prepare branch."
---

# Development Step 0: Initialize

**초기화 단계**: 개발 환경을 준비하고, 선택적으로 SPEC 문서와 브랜치를 생성합니다.

**Note**: 이 단계는 선택적이에요. 간단한 작업은 dev-1-plan부터 시작해도 됩니다.

## Usage

```bash
# Basic initialization
/jikime:dev-0-init Add user authentication

# With SPEC document
/jikime:dev-0-init --spec Add payment gateway

# With branch creation
/jikime:dev-0-init --branch Add dashboard widgets

# Full enterprise mode
/jikime:dev-0-init --spec --branch Add enterprise SSO
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `[description]` | Feature or task description | Required |
| `--spec` | Create SPEC document in `.jikime/specs/` | Off |
| `--branch` | Create feature branch | Off |
| `--id [name]` | Custom feature ID (e.g., FEAT-001) | Auto |

## Process

```
1. Analyze Request
   - Parse feature description
   - Generate feature ID
        ↓
2. Setup Workspace
   - Verify project structure
   - Check prerequisites
        ↓
3. Create SPEC (if --spec)
   - Generate .jikime/specs/FEAT-XXX.md
   - Basic requirement template
        ↓
4. Create Branch (if --branch)
   - git checkout -b feature/[id]
   - Link to SPEC (if exists)
        ↓
5. Ready for Planning
   - Display next step command
```

## SPEC Document Format (when --spec)

```markdown
# SPEC-FEAT-001: [Feature Name]

## Overview
- **Status**: Draft
- **Created**: [Date]
- **Branch**: feature/feat-001 (if --branch)

## Requirements

### Functional
- [ ] [Requirement 1]
- [ ] [Requirement 2]

### Non-Functional
- [ ] Performance: [criteria]
- [ ] Security: [criteria]

## Acceptance Criteria
- [ ] [Criterion 1]
- [ ] [Criterion 2]

## Notes
[Additional context]
```

## Output

```markdown
## Development Initialized

**Feature**: Add user authentication
**ID**: FEAT-001

### Created
- [x] Feature ID generated
- [x] SPEC document: `.jikime/specs/FEAT-001.md` (if --spec)
- [x] Branch: `feature/feat-001` (if --branch)

### Next Step
```bash
/jikime:dev-1-plan FEAT-001
```
```

## Enterprise Mode

`--spec --branch`를 함께 사용하면 엔터프라이즈 모드:

```bash
/jikime:dev-0-init --spec --branch Add critical feature
```

이는 다음과 동일해요:
```bash
/jikime:dev --enterprise Add critical feature
```

## Workflow

```
/jikime:dev-0-init  ← 현재 (선택적)
        ↓
/jikime:dev-1-plan
        ↓
/jikime:dev-2-implement
        ↓
/jikime:dev-3-test
        ↓
/jikime:dev-4-review
```

## When to Use

**Use dev-0-init when**:
- 팀 프로젝트에서 추적이 필요할 때
- 복잡한 기능을 체계적으로 관리할 때
- 엔터프라이즈 환경에서 감사 추적이 필요할 때

**Skip to dev-1-plan when**:
- 간단한 버그 수정
- 빠른 개선 작업
- 개인 프로젝트

## Next Step

초기화 후:
```bash
/jikime:dev-1-plan [FEAT-ID or description]
```

---

Version: 1.0.0
