# JikiME-ADK Command Reference

JikiME-ADK의 모든 슬래시 명령어에 대한 종합 레퍼런스 문서입니다.

## Overview

JikiME-ADK는 세 가지 타입의 명령어를 제공합니다:

| Type | 설명 | 명령어 |
|------|------|--------|
| **Type A: Workflow** | 핵심 개발 워크플로우 | 0-project, 1-plan, 2-run, 3-sync |
| **Type B: Utility** | 빠른 실행 및 자동화 | jarvis, test, loop |
| **Standalone** | 독립 실행 유틸리티 | architect, build-fix, docs, e2e, learn, refactor, security |
| **Migration** | 레거시 마이그레이션 | migrate, migrate-0~4 |

---

## Command Map

```
                    ┌─────────────────────────────────────┐
                    │        JikiME-ADK Commands          │
                    └─────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────┐          ┌───────────────┐          ┌───────────────┐
│   Workflow    │          │   Utility     │          │   Migration   │
│   (Type A)    │          │   (Type B)    │          │   Workflow    │
└───────────────┘          └───────────────┘          └───────────────┘
        │                           │                           │
   0-project                    jarvis                      migrate
        ↓                       test                    migrate-0~4
    1-plan                      loop
        ↓                         │
     2-run              ┌─────────┴─────────┐
        ↓               │    Standalone     │
    3-sync              │    Utilities      │
                        └───────────────────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                architect   build-fix      docs
                  e2e        learn      refactor
                          security
```

---

## Type A: Workflow Commands

핵심 JikiME 개발 워크플로우를 구성하는 명령어들입니다.

### /jikime:0-project

**프로젝트 초기화 및 문서 생성**

| 항목 | 내용 |
|------|------|
| **설명** | 코드베이스 분석을 통한 프로젝트 설정 및 문서 생성 |
| **Type** | Workflow (Type A) |
| **Context** | - |
| **Agent Chain** | manager-project → Explore → manager-docs |

#### Usage

```bash
/jikime:0-project
```

#### Process

```
PHASE 0: 프로젝트 타입 감지
    └─ New Project / Existing Project / Migration Project
         ↓
PHASE 0.5: 정보 수집 (New/Migration)
    └─ manager-project: 프로젝트 설정 수집
         ↓
PHASE 1: 코드베이스 분석 (Existing)
    └─ Explore: 구조, 기술 스택, 핵심 기능 분석
         ↓
PHASE 2: 사용자 확인
    └─ AskUserQuestion: 분석 결과 승인
         ↓
PHASE 3: 문서 생성
    └─ manager-docs: product.md, structure.md, tech.md 생성
         ↓
PHASE 3.5: 개발 환경 확인
    └─ LSP 서버 설치 확인 및 안내
         ↓
PHASE 4: 완료
    └─ 다음 단계 안내
```

#### Output Files

- `.jikime/project/product.md` - 제품 개요, 기능, 사용자 가치
- `.jikime/project/structure.md` - 프로젝트 아키텍처 및 디렉토리 구조
- `.jikime/project/tech.md` - 기술 스택, 의존성, 기술 결정

---

### /jikime:1-plan

**SPEC 정의 및 개발 브랜치 생성**

| 항목 | 내용 |
|------|------|
| **설명** | 요구사항을 EARS 형식 SPEC 문서로 정의 |
| **Type** | Workflow (Type A) |
| **Context** | planning.md |
| **Agent Chain** | Explore (optional) → manager-spec → manager-git (conditional) |

#### Usage

```bash
# SPEC만 생성 (기본)
/jikime:1-plan "User authentication system"

# SPEC + Git 브랜치
/jikime:1-plan "User authentication system" --branch

# SPEC + Git Worktree (병렬 개발)
/jikime:1-plan "User authentication system" --worktree
```

#### Options

| Option | 설명 |
|--------|------|
| `--branch` | Feature 브랜치 자동 생성 |
| `--worktree` | Git Worktree로 격리된 개발 환경 생성 |

#### Process

```
PHASE 1: 프로젝트 분석 & SPEC 기획
    └─ manager-spec: SPEC 후보 생성, EARS 구조 설계
         ↓
PHASE 1.5: 사전 검증
    └─ SPEC 타입 분류, ID 형식 검증, 중복 확인
         ↓
PHASE 2: SPEC 문서 생성
    └─ spec.md, plan.md, acceptance.md 생성
         ↓
PHASE 3: Git 브랜치/Worktree 설정 (조건부)
    └─ 플래그에 따라 브랜치 또는 worktree 생성
```

#### Output Files

```
.jikime/specs/SPEC-{ID}/
├── spec.md        # 핵심 명세 (EARS 형식)
├── plan.md        # 구현 계획
└── acceptance.md  # 인수 기준 (Given/When/Then)
```

---

### /jikime:2-run

**DDD 기반 SPEC 구현**

| 항목 | 내용 |
|------|------|
| **설명** | ANALYZE-PRESERVE-IMPROVE 사이클로 SPEC 구현 |
| **Type** | Workflow (Type A) |
| **Context** | dev.md |
| **Agent Chain** | manager-strategy → manager-ddd → manager-quality → manager-git |

#### Usage

```bash
# 표준 실행
/jikime:2-run SPEC-AUTH-001

# 체크포인트 생성 후 실행
/jikime:2-run SPEC-AUTH-001 --checkpoint

# Personal/Team 모드 강제
/jikime:2-run SPEC-AUTH-001 --personal
/jikime:2-run SPEC-AUTH-001 --team
```

#### Options

| Option | 설명 |
|--------|------|
| `--checkpoint` | 시작 전 복구 지점 생성 |
| `--skip-quality` | 품질 검증 건너뛰기 (비권장) |
| `--personal` | Personal git 모드 강제 |
| `--team` | Team git 모드 강제 |

#### DDD Cycle

```
┌─────────────┐
│   ANALYZE   │  ← 현재 동작 이해
└──────┬──────┘
       ↓
┌─────────────┐
│  PRESERVE   │  ← 특성화 테스트로 동작 보존
└──────┬──────┘
       ↓
┌─────────────┐
│   IMPROVE   │  ← 자신감 있게 변경
└──────┬──────┘
       ↓
    (반복)
```

#### Process

```
PHASE 1: 전략 분석
    └─ manager-strategy: 구현 전략 수립
         ↓
PHASE 1.5: 태스크 분해
    └─ TodoWrite로 태스크 추적
         ↓
PHASE 2: DDD 구현
    └─ manager-ddd: 각 태스크별 DDD 사이클 실행
         ↓
PHASE 2.5: 품질 검증
    └─ manager-quality: 테스트 커버리지, 린트, 타입 체크
         ↓
PHASE 3: Git 작업
    └─ manager-git: 커밋, PR 생성 (team 모드)
         ↓
PHASE 4: 완료
    └─ 결과 리포트, 다음 단계 안내
```

---

### /jikime:3-sync

**문서 동기화 및 SPEC 완료 처리**

| 항목 | 내용 |
|------|------|
| **설명** | 코드 변경사항과 문서 동기화, SPEC 완료 처리 |
| **Type** | Workflow (Type A) |
| **Context** | sync.md |
| **Agent Chain** | manager-quality → manager-docs → manager-git |

#### Usage

```bash
# 특정 SPEC 동기화
/jikime:3-sync SPEC-AUTH-001

# 프롬프트 없이 자동 실행
/jikime:3-sync SPEC-AUTH-001 --auto

# 문서 재생성
/jikime:3-sync SPEC-AUTH-001 --force

# 전체 SPEC 상태 확인
/jikime:3-sync --status

# 프로젝트 문서만 동기화
/jikime:3-sync --project
```

#### Options

| Option | 설명 |
|--------|------|
| `--auto` | 프롬프트 없이 자동 실행 |
| `--force` | 모든 문서 재생성 |
| `--status` | 전체 SPEC 동기화 상태 표시 |
| `--project` | 프로젝트 레벨 문서만 동기화 |

#### Process

```
PHASE 0.5: 사전 품질 확인
    └─ 구현 완료 확인, 테스트 통과 확인
         ↓
PHASE 1: 문서 분석
    └─ 코드 변경 스캔, 문서 영향 매핑
         ↓
PHASE 2: 문서 업데이트
    └─ manager-docs: product.md, structure.md, tech.md, CHANGELOG 업데이트
         ↓
PHASE 3: SPEC 완료 처리
    └─ 상태 "completed"로 변경, 완료 요약 생성
         ↓
PHASE 4: PR/Merge 관리 (Team 모드)
    └─ PR 업데이트, 머지 옵션 제공
         ↓
PHASE 5: 완료
    └─ 동기화 리포트, 다음 단계 안내
```

---

## Type B: Utility Commands

빠른 실행과 자동화를 위한 명령어들입니다.

### /jikime:jarvis

**J.A.R.V.I.S. - 지능형 자율 오케스트레이션**

| 항목 | 내용 |
|------|------|
| **설명** | Iron Man의 AI 비서에서 영감을 받은 지능형 오케스트레이터 |
| **Type** | Utility (Type B) |
| **Context** | - |
| **특징** | 5-way 병렬 탐색, 멀티 전략 비교, 적응형 실행, 예측 제안 |

#### Usage

```bash
# 기본 사용 (auto 전략)
/jikime:jarvis "Add JWT authentication"

# Safe 전략 (보수적)
/jikime:jarvis "Refactor payment module" --strategy safe

# Fast 전략 (공격적)
/jikime:jarvis "Fix typo in README" --strategy fast

# 자동 루프 활성화
/jikime:jarvis "Implement user dashboard" --loop --max 20

# 이전 작업 재개
/jikime:jarvis resume SPEC-AUTH-001
```

#### Options

| Option | 설명 | 기본값 |
|--------|------|--------|
| `--strategy` | 실행 전략: auto, safe, fast | auto |
| `--loop` | 에러 자동 수정 반복 활성화 | config |
| `--max N` | 최대 반복 횟수 | 50 |
| `--branch` | 피처 브랜치 자동 생성 | config |
| `--pr` | 완료 시 PR 자동 생성 | config |
| `--resume SPEC` | 이전 작업 재개 | - |

#### Strategy Comparison

| 전략 | 리스크 | 속도 | 되돌리기 | 테스트 커버리지 |
|------|--------|------|----------|----------------|
| **Conservative** | 낮음 | 느림 | 쉬움 | 100% |
| **Balanced** | 중간 | 중간 | 중간 | 85% |
| **Aggressive** | 높음 | 빠름 | 어려움 | 70% |

#### Autonomous Flow

```
PHASE 0: 선제적 정보 수집 (5-way 병렬)
    ├── Explore Agent: 코드베이스 구조
    ├── Research Agent: 외부 문서, 베스트 프랙티스
    ├── Quality Agent: 현재 상태 진단
    ├── Security Agent: 보안 영향 사전 스캔
    └── Performance Agent: 성능 영향 예측
         ↓
PHASE 1: 멀티 전략 기획
    ├── Strategy A: Conservative
    ├── Strategy B: Balanced
    └── Strategy C: Aggressive
    └─ Trade-off 분석 → 최적 전략 선택
         ↓
PHASE 2: 적응형 DDD 구현
    └─ 자가 진단 루프 (진행 확인, 전략 피봇 결정)
         ↓
PHASE 3: 완료 & 예측
    └─ 문서 동기화 + 다음 단계 예측 제안
```

---

### /jikime:test

**테스트 실행 유틸리티**

| 항목 | 내용 |
|------|------|
| **설명** | 단위/통합 테스트 빠른 실행 |
| **Type** | Utility (Type B) |
| **Context** | - |
| **관련 명령어** | `/jikime:e2e` (E2E 테스트) |

#### Usage

```bash
# 전체 테스트 실행
/jikime:test

# 커버리지 리포트 포함
/jikime:test --coverage

# 특정 테스트 타입만
/jikime:test --unit
/jikime:test --integration

# Watch 모드
/jikime:test --watch

# 실패 테스트 자동 수정 시도
/jikime:test --fix
```

#### Options

| Option | 설명 |
|--------|------|
| `--coverage` | 커버리지 리포트 생성 |
| `--unit` | 단위 테스트만 실행 |
| `--integration` | 통합 테스트만 실행 |
| `--watch` | 연속 테스트를 위한 Watch 모드 |
| `--fix` | 가능한 경우 실패 테스트 자동 수정 |

#### Coverage Targets

| 타입 | 목표 |
|------|------|
| Business Logic | 90%+ |
| API Endpoints | 80%+ |
| UI Components | 70%+ |
| Overall | 80%+ |

---

### /jikime:loop

**Ralph Loop - LSP/AST-grep 피드백 기반 반복 개선**

| 항목 | 내용 |
|------|------|
| **설명** | 지능적 피드백 루프로 점진적 코드 개선 |
| **Type** | Utility (Type B) |
| **Context** | debug.md |
| **Skill** | jikime-workflow-loop |

#### Usage

```bash
# 기본 사용 (모든 에러 수정)
/jikime:loop "Fix all TypeScript errors"

# 옵션 지정
/jikime:loop "Remove security vulnerabilities" --max-iterations 5 --zero-security

# 특정 디렉토리
/jikime:loop @src/services/ "Fix all lint errors" --zero-warnings

# 테스트 통과까지
/jikime:loop "Fix failing tests" --tests-pass --max-iterations 10

# 활성 루프 취소
/jikime:loop --cancel
```

#### Options

| Option | 설명 | 기본값 |
|--------|------|--------|
| `--max-iterations` | 최대 반복 횟수 | 10 |
| `--zero-errors` | 에러 0개 필수 | true |
| `--zero-warnings` | 경고 0개 필수 | false |
| `--zero-security` | 보안 이슈 0개 필수 | false |
| `--tests-pass` | 모든 테스트 통과 필수 | false |
| `--stagnation-limit` | 개선 없는 반복 한계 | 3 |
| `--cancel` | 활성 루프 취소 | - |

#### Process

```
1. Initialize Loop
   jikime hooks start-loop --task "..." --options ...
        ↓
2. Load Skill
   Skill("jikime-workflow-loop")
        ↓
3. Execute Iteration
   - 현재 상태 분석
   - 이슈 하나씩 수정
   - LSP/AST-grep 피드백 수집
        ↓
4. Stop Hook Evaluation
   - 완료 조건 체크
   - 개선률 계산
   - Continue 또는 Complete 결정
        ↓
5. (Continue) 피드백 재주입
        ↓
6. (Complete) 최종 리포트 생성
```

---

## Standalone Utility Commands

워크플로우와 독립적으로 사용할 수 있는 유틸리티 명령어들입니다.

### /jikime:architect

**아키텍처 리뷰 및 설계**

| 항목 | 내용 |
|------|------|
| **설명** | 시스템 설계, 트레이드오프 분석, ADR 생성 |
| **Context** | planning.md |
| **단독 사용** | ✅ 높음 - 전체 워크플로우 없이 독립 사용 가능 |

#### Usage

```bash
# 현재 아키텍처 리뷰
/jikime:architect

# 새 기능 아키텍처 설계
/jikime:architect Design payment system

# ADR 생성
/jikime:architect --adr "Use PostgreSQL over MongoDB"

# 트레이드오프 분석
/jikime:architect --tradeoff "Monolith vs Microservices"
```

#### Options

| Option | 설명 |
|--------|------|
| `[description]` | 설계할 기능/시스템 |
| `--adr` | Architecture Decision Record 생성 |
| `--tradeoff` | 트레이드오프 분석 |
| `--review` | 기존 아키텍처 리뷰 |

#### Architecture Principles

| 원칙 | 설명 |
|------|------|
| **Modularity** | High cohesion, low coupling |
| **Scalability** | 수평 확장 가능 |
| **Maintainability** | 이해하고 테스트하기 쉬움 |
| **Security** | Defense in depth |

---

### /jikime:build-fix

**빌드 에러 점진적 수정**

| 항목 | 내용 |
|------|------|
| **설명** | TypeScript 및 빌드 에러를 하나씩 안전하게 수정 |
| **Context** | debug.md |
| **단독 사용** | ✅ 높음 - 빌드 실패 시 즉시 사용 |

#### Usage

```bash
# 모든 빌드 에러 수정
/jikime:build-fix

# 특정 파일 에러 수정
/jikime:build-fix @src/services/order.ts

# 적용 없이 미리보기
/jikime:build-fix --dry-run
```

#### Options

| Option | 설명 |
|--------|------|
| `@path` | 특정 파일 지정 |
| `--dry-run` | 적용 없이 미리보기 |
| `--max` | 최대 수정 에러 수 (기본: 10) |

#### Safety Rules

- **한 번에 하나의 에러** - 안전 우선
- **각 수정 후 검증** - 새 에러 감지 시 중단
- **최소 변경** - 필요한 최소 수정만

---

### /jikime:docs

**문서 업데이트**

| 항목 | 내용 |
|------|------|
| **설명** | README, API 문서, 코드 코멘트를 코드와 동기화 |
| **Context** | - |
| **단독 사용** | ⚠️ 중간 - 3-sync와 부분 중복 |

#### Usage

```bash
# 모든 문서 업데이트
/jikime:docs

# 특정 문서 타입
/jikime:docs --type api
/jikime:docs --type readme
/jikime:docs --type changelog

# 누락된 문서 생성
/jikime:docs --generate

# 특정 코드 변경사항에 대해
/jikime:docs @src/api/
```

#### Options

| Option | 설명 |
|--------|------|
| `@path` | 특정 코드 대상 |
| `--type` | 문서 타입: api, readme, changelog, jsdoc |
| `--generate` | 누락된 문서 생성 |
| `--dry-run` | 적용 없이 변경 표시 |

#### vs 3-sync

| 기능 | docs | 3-sync |
|------|------|--------|
| SPEC 기반 | ❌ | ✅ |
| CHANGELOG | ✅ | ✅ |
| 프로젝트 문서 | ✅ | ✅ |
| Git PR 관리 | ❌ | ✅ |
| SPEC 완료 처리 | ❌ | ✅ |

**권장**: SPEC 워크플로우 사용 시 `3-sync`, 빠른 문서 업데이트만 필요시 `docs`

---

### /jikime:e2e

**E2E 테스트 (Playwright)**

| 항목 | 내용 |
|------|------|
| **설명** | Playwright로 E2E 테스트 생성 및 실행 |
| **Context** | - |
| **단독 사용** | ✅ 높음 - test와 별도 영역 |

#### Usage

```bash
# 플로우에 대한 E2E 테스트 생성
/jikime:e2e Test login flow

# 기존 E2E 테스트 실행
/jikime:e2e --run

# 특정 테스트 실행
/jikime:e2e --run @tests/e2e/auth.spec.ts

# 디버그 모드
/jikime:e2e --run --debug
```

#### Options

| Option | 설명 |
|--------|------|
| `[description]` | 테스트할 사용자 플로우 |
| `--run` | 기존 테스트 실행 |
| `--debug` | 디버그 모드 (headed browser) |
| `--headed` | 브라우저 창 표시 |

#### vs test

| 기능 | e2e | test |
|------|-----|------|
| 범위 | 전체 사용자 플로우 | 단위/통합 |
| 도구 | Playwright | Vitest, Jest, Pytest 등 |
| 속도 | 느림 | 빠름 |
| 용도 | 핵심 사용자 여정 | 비즈니스 로직, API |

---

### /jikime:learn

**코드베이스 탐색 및 학습**

| 항목 | 내용 |
|------|------|
| **설명** | 아키텍처, 패턴, 구현 세부사항 대화형 학습 |
| **Context** | research.md |
| **단독 사용** | ✅ 높음 - 온보딩, 코드 이해에 유용 |

#### Usage

```bash
# 전체 개요
/jikime:learn

# 특정 주제 학습
/jikime:learn authentication flow

# 특정 파일 학습
/jikime:learn @src/services/order.ts

# 대화형 Q&A 모드
/jikime:learn --interactive
```

#### Options

| Option | 설명 |
|--------|------|
| `[topic]` | 학습할 특정 주제 |
| `@path` | 특정 파일/모듈 학습 |
| `--interactive` | 대화형 Q&A 모드 |
| `--depth` | 상세 수준: overview, detailed, deep |

#### Topics

- **Architecture**: 프로젝트 구조, 패턴
- **Features**: 기능 작동 방식, 구현
- **Conventions**: 코딩 스타일, 명명, 조직
- **Data Flow**: 데이터 흐름

---

### /jikime:refactor

**DDD 방법론 리팩토링**

| 항목 | 내용 |
|------|------|
| **설명** | 동작 보존과 함께 클린 코드 원칙 적용 |
| **Context** | dev.md |
| **단독 사용** | ⚠️ 중간 - 2-run DDD와 부분 중복 |

#### Usage

```bash
# 특정 파일 리팩토링
/jikime:refactor @src/services/order.ts

# 특정 패턴으로 리팩토링
/jikime:refactor @src/utils/ --pattern extract-function

# Safe 모드 (추가 테스트)
/jikime:refactor @src/core/ --safe

# 미리보기
/jikime:refactor @src/auth/ --dry-run
```

#### Options

| Option | 설명 |
|--------|------|
| `@path` | 리팩토링할 파일 |
| `--pattern` | 패턴: extract-function, remove-duplication |
| `--safe` | 추가 특성화 테스트 |
| `--dry-run` | 적용 없이 미리보기 |

#### DDD Approach

```
ANALYZE → PRESERVE → IMPROVE

1. ANALYZE: 현재 동작 이해
2. PRESERVE: 특성화 테스트 생성
3. IMPROVE: 리팩토링 적용
4. VERIFY: 테스트 통과 확인
```

#### vs 2-run

| 기능 | refactor | 2-run |
|------|----------|-------|
| SPEC 기반 | ❌ | ✅ |
| DDD 사이클 | ✅ | ✅ |
| 품질 게이트 | ✅ | ✅ |
| Git PR 관리 | ❌ | ✅ |
| 용도 | 즉석 리팩토링 | SPEC 기반 구현 |

**권장**: SPEC 워크플로우 사용 시 `2-run`, 빠른 리팩토링만 필요시 `refactor`

---

### /jikime:security

**보안 감사**

| 항목 | 내용 |
|------|------|
| **설명** | OWASP Top 10, 의존성 스캔, 시크릿 탐지 |
| **Context** | review.md |
| **단독 사용** | ✅ 높음 - 언제든 독립 실행 가능 |

#### Usage

```bash
# 전체 보안 감사
/jikime:security

# 특정 경로 스캔
/jikime:security @src/api/

# 의존성 감사만
/jikime:security --deps

# 시크릿 스캔만
/jikime:security --secrets

# OWASP 체크만
/jikime:security --owasp
```

#### Options

| Option | 설명 |
|--------|------|
| `[path]` | 대상 경로 |
| `--deps` | 의존성 취약점 스캔 |
| `--secrets` | 하드코딩된 시크릿 탐지 |
| `--owasp` | OWASP Top 10 체크 |
| `--fix` | 가능한 경우 자동 수정 |

#### OWASP Top 10 Checks

| # | 취약점 | 탐지 대상 |
|---|--------|----------|
| 1 | Injection | SQL, NoSQL, Command |
| 2 | Broken Auth | Password handling |
| 3 | Data Exposure | Hardcoded secrets |
| 4 | XSS | innerHTML, dangerouslySetInnerHTML |
| 5 | SSRF | Unvalidated URLs |
| 6 | Authorization | Missing permission checks |

#### Severity Levels

| Level | 액션 |
|-------|------|
| CRITICAL | 즉시 수정 필요 |
| HIGH | 배포 전 수정 |
| MEDIUM | 가능한 빨리 수정 |
| LOW | 검토 후 결정 |

---

## Migration Workflow

레거시 프로젝트를 Next.js 16으로 마이그레이션하는 워크플로우입니다.

### /jikime:migrate

**마이그레이션 통합 명령어**

| 항목 | 내용 |
|------|------|
| **설명** | 레거시 프론트엔드를 Next.js 16 App Router로 마이그레이션 |
| **Type** | Workflow |
| **대상** | Vue.js, React CRA, Angular, Svelte 등 |

#### Workflow Overview

```
/jikime:migrate-0-discover   → Step 0: 소스 탐색
        ↓
/jikime:migrate-1-analyze    → Step 1: 상세 분석
        ↓
/jikime:migrate-2-plan       → Step 2: 계획 수립
        ↓
/jikime:migrate-3-execute    → Step 3: 실행
        ↓
/jikime:migrate-4-verify     → Step 4: 검증

또는

/jikime:migrate [project]    → 전체 자동화
```

#### Sub-Commands

| Sub-Command | 설명 | 전제조건 |
|-------------|------|----------|
| `plan` | 마이그레이션 계획 생성 | `as_is_spec.md` |
| `skill` | 프로젝트 스킬 생성 | `migration_plan.md` |
| `run` | 마이그레이션 실행 | `SKILL.md` |

#### Usage

```bash
# Step-by-Step
/jikime:migrate-1-analyze "./my-vue-app"
/jikime:migrate plan my-vue-app
/jikime:migrate skill my-vue-app
/jikime:migrate run my-vue-app --output ./migrated

# 전체 자동화
/jikime:migrate my-vue-app --loop --output ./migrated

# 백서 생성
/jikime:migrate run my-vue-app --whitepaper-report --client "ABC Corp"
```

#### Options

| Option | 설명 |
|--------|------|
| `--artifacts-output` | 마이그레이션 아티팩트 디렉토리 |
| `--output` | 마이그레이션된 프로젝트 출력 디렉토리 |
| `--loop` | 자율 반복 활성화 |
| `--max N` | 최대 반복 횟수 |
| `--strategy` | 전략: incremental, big-bang |
| `--whitepaper-report` | Post-Migration 백서 생성 |
| `--client` | 클라이언트 회사명 |
| `--lang` | 백서 언어: ko, en, ja, zh |

#### Target Stack

| 기술 | 버전 |
|------|------|
| Framework | Next.js 16 (App Router) |
| Language | TypeScript 5.x |
| Styling | Tailwind CSS 4.x |
| UI Components | shadcn/ui |
| Icons | lucide-react |
| State | Zustand |

---

## Command Comparison Matrix

### Use Case Guide

| 상황 | 권장 명령어 |
|------|------------|
| 새 프로젝트 시작 | `0-project` → `1-plan` → `2-run` → `3-sync` |
| 빠른 기능 구현 | `/jikime:jarvis "description"` |
| 빌드 에러 수정 | `/jikime:build-fix` |
| 보안 점검 | `/jikime:security` |
| 아키텍처 리뷰 | `/jikime:architect` |
| 코드 리팩토링 (SPEC 없이) | `/jikime:refactor @path` |
| 코드 리팩토링 (SPEC 기반) | `/jikime:1-plan` → `/jikime:2-run` |
| 테스트 실행 (단위/통합) | `/jikime:test` |
| 테스트 실행 (E2E) | `/jikime:e2e` |
| 에러 반복 수정 | `/jikime:loop "description"` |
| 코드베이스 학습 | `/jikime:learn` |
| 문서 업데이트 (빠른) | `/jikime:docs` |
| 문서 동기화 (SPEC 기반) | `/jikime:3-sync` |
| 레거시 마이그레이션 | `/jikime:migrate` |

### Overlap Analysis

| 명령어 A | 명령어 B | 중복 영역 | 권장 |
|----------|----------|----------|------|
| `docs` | `3-sync` | 문서 업데이트 | SPEC 기반: 3-sync, 빠른 업데이트: docs |
| `refactor` | `2-run` | DDD 리팩토링 | SPEC 기반: 2-run, 즉석: refactor |
| `test` | `e2e` | 테스트 실행 | 단위/통합: test, E2E: e2e |

---

## Version Information

| 명령어 | 버전 | 최종 업데이트 |
|--------|------|--------------|
| 0-project | 1.0.0 | 2026-01-22 |
| 1-plan | 1.0.0 | 2026-01-22 |
| 2-run | 1.0.0 | 2026-01-22 |
| 3-sync | 2.0.0 | 2026-01-22 |
| jarvis | 1.0.0 | 2026-01-22 |
| test | 1.0.0 | 2026-01-22 |
| loop | 1.0.0 | 2026-01-22 |
| architect | 1.0.0 | - |
| build-fix | 1.0.0 | - |
| docs | 1.0.0 | - |
| e2e | 1.0.0 | - |
| learn | 1.0.0 | - |
| refactor | 1.0.0 | - |
| security | 1.0.0 | - |
| migrate | 1.4.5 | - |

---

Version: 1.0.0
Last Updated: 2026-01-22
