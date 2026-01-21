---
description: "Development workflow orchestrator. Init → Plan → Implement → Test → Review cycle."
---

# Dev - Development Workflow

개발 워크플로우 오케스트레이터. 새 기능 개발 및 기존 코드 개선을 위한 체계적 프로세스.

## Quick Start

```bash
# Simple mode (most common)
/jikime:dev Add user authentication

# Enterprise mode (full tracking)
/jikime:dev --enterprise Add payment system
```

## Workflow Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Development Workflow                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   /jikime:dev-0-init       초기화, SPEC 생성 (선택적)       │
│           ↓                                                 │
│   /jikime:dev-1-plan       요구사항 분석, 계획 수립         │
│           ↓                (승인 대기)                      │
│   /jikime:dev-2-implement  코드 구현 (DDD 방법론)           │
│           ↓                                                 │
│   /jikime:dev-3-test       테스트 실행, 커버리지 확인       │
│           ↓                                                 │
│   /jikime:dev-4-review     코드 리뷰, 품질 검증             │
│           ↓                                                 │
│        Complete!           커밋 및 배포 준비                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Modes

### Simple Mode (Default)

빠른 개발을 위한 기본 모드:

```bash
/jikime:dev Add shopping cart
```

- dev-1-plan부터 시작
- 확인 후 구현 진행
- 간단한 리뷰로 완료

### Enterprise Mode

전체 추적 및 품질 검증:

```bash
/jikime:dev --enterprise Add enterprise SSO
```

**Enterprise mode는 다음을 자동 활성화**:
- `--spec`: SPEC 문서 생성
- `--branch`: 기능 브랜치 생성
- `--quality`: 자동 품질 검증 (lint, type-check, test)

## Usage

```bash
# Full workflow (starts from plan)
/jikime:dev Add payment processing

# Enterprise mode
/jikime:dev --enterprise Add critical feature

# Individual flags
/jikime:dev --spec Add user profile        # SPEC 문서만
/jikime:dev --branch Add dashboard         # 브랜치만
/jikime:dev --quality Add auth             # 품질 검증만

# Resume from specific step
/jikime:dev --resume implement
/jikime:dev --resume test

# Quick mode (skip confirmation)
/jikime:dev --quick Fix login button

# With context
/jikime:dev @src/auth/ Improve session handling
```

## Options

| Option | Description | Mode |
|--------|-------------|------|
| `[description]` | Feature/improvement description | All |
| `@path` | Reference existing code | All |
| `--enterprise` | Full enterprise mode (spec+branch+quality) | Enterprise |
| `--spec` | Create SPEC document | Enterprise |
| `--branch` | Create feature branch | Enterprise |
| `--quality` | Enable quality gates (lint, type, test) | Enterprise |
| `--resume` | Resume from: init, plan, implement, test, review | All |
| `--quick` | Skip confirmations (for small changes) | Simple |

## Individual Commands

| Step | Command | Description | Mode |
|------|---------|-------------|------|
| 0 | `/jikime:dev-0-init` | 초기화, SPEC/브랜치 생성 | Enterprise |
| 1 | `/jikime:dev-1-plan` | 계획 수립 | All |
| 2 | `/jikime:dev-2-implement` | 코드 구현 | All |
| 3 | `/jikime:dev-3-test` | 테스트 실행 | All |
| 4 | `/jikime:dev-4-review` | 코드 리뷰, 품질 검증 | All |

## When to Use

### Simple Mode (권장 - 대부분의 경우)
- 간단한 버그 수정
- 빠른 기능 추가
- 개인 프로젝트
- 프로토타이핑

### Enterprise Mode (복잡한 프로젝트)
- 팀 협업 프로젝트
- 중요한 기능 개발
- 감사 추적이 필요한 경우
- 품질이 중요한 프로덕션 코드

## Methodology: DDD

```
ANALYZE → PRESERVE → IMPROVE

1. ANALYZE: 기존 동작 이해
2. PRESERVE: 특성 테스트로 동작 보존
3. IMPROVE: 새로운 코드 구현
```

## Quality Gates (--quality 또는 --enterprise)

활성화 시 자동 실행:

```
┌─────────────────────────────────────────┐
│           Quality Verification          │
├─────────────────────────────────────────┤
│  1. Lint Check      → eslint/biome      │
│  2. Type Check      → tsc --noEmit      │
│  3. Unit Tests      → vitest/jest       │
│  4. Security Scan   → basic checks      │
└─────────────────────────────────────────┘
```

## Utility Commands

워크플로우 중 필요시 사용:

| Command | Purpose |
|---------|---------|
| `/jikime:build-fix` | 빌드/타입 에러 수정 |
| `/jikime:refactor` | 코드 리팩토링 |
| `/jikime:security` | 심층 보안 감사 |
| `/jikime:e2e` | E2E 테스트 |
| `/jikime:architect` | 아키텍처 설계 |
| `/jikime:docs` | 문서화 |

## Example Sessions

### Simple Mode

```bash
# 1. Start planning
/jikime:dev Add shopping cart feature

# User approves plan
> yes

# 2. Implement (auto-continues)
# 3. Test (auto-continues)
# 4. Review (auto-continues)

# Done! Ready for commit
```

### Enterprise Mode

```bash
# 1. Initialize with full setup
/jikime:dev --enterprise Add payment gateway

# Creates:
# - .jikime/specs/FEAT-001.md
# - Branch: feature/feat-001
# - Quality gates enabled

# 2-5. Workflow continues with quality verification
```

## vs Migration Workflow

| Workflow | Purpose | Commands | Mode |
|----------|---------|----------|------|
| **dev** | 새 기능/개선 | `dev-0` ~ `dev-4` | Simple/Enterprise |
| **migrate** | 기술 스택 변환 | `migrate-0` ~ `migrate-4` | - |

---

Version: 2.0.0
Methodology: DDD (ANALYZE-PRESERVE-IMPROVE)
Modes: Simple (default), Enterprise (--enterprise)
