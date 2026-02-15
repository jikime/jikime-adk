# F.R.I.D.A.Y. - Migration Orchestration System

JikiME-ADK의 마이그레이션 전담 오케스트레이션 시스템. Iron Man의 두 번째 AI 비서에서 영감을 받은 체계적이고 정밀한 프레임워크 전환 자동화.

## Overview

F.R.I.D.A.Y. (Framework Relay & Integration Deployment Assistant Yesterday)는 레거시 시스템을 현대 프레임워크로 전환하는 **마이그레이션 전담 오케스트레이터**입니다. J.A.R.V.I.S.가 개발을 담당하는 동안, F.R.I.D.A.Y.는 오직 마이그레이션에 특화되어 분석, 계획, 실행, 검증의 전 과정을 자율적으로 수행합니다.

### 핵심 철학

```
"Transitioning to the new system, sir. All legacy patterns mapped and ready."
```

### J.A.R.V.I.S.와의 차별점

| 기능 | J.A.R.V.I.S. (개발) | F.R.I.D.A.Y. (마이그레이션) |
|------|---------------------|---------------------------|
| 탐색 | 5개 에이전트 병렬 | 3개 에이전트 (소스 중심) |
| 계획 | 멀티 전략 비교 | 동적 스킬 탐색 + 전략 비교 |
| 실행 | DDD 사이클 | DDD + 동작 보존 검증 |
| 추적 | SPEC 기반 | `.migrate-config.yaml` + `progress.yaml` |
| 검증 | LSP + 테스트 | Playwright E2E + 시각적 회귀 + 성능 비교 |
| 완료 마커 | `<jikime>DONE</jikime>` | `<jikime>MIGRATION_COMPLETE</jikime>` |

## Architecture

### 시스템 구조

```
┌─────────────────────────────────────────────────────────────────┐
│                    F.R.I.D.A.Y. System                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Phase 0: Discovery (3-Way Parallel)                            │
│  ┌──────────┐ ┌──────────────┐ ┌──────────────┐                │
│  │ Codebase │ │  Dependency  │ │    Risk      │                │
│  │ Explorer │ │   Analyzer   │ │  Assessor    │                │
│  └────┬─────┘ └──────┬───────┘ └──────┬───────┘                │
│       └───────────────┼────────────────┘                        │
│                       ▼                                         │
│           ┌──────────────────────┐                              │
│           │  .migrate-config.yaml │                              │
│           │  + Complexity Score   │                              │
│           └──────────┬───────────┘                              │
│                      ▼                                          │
│  Phase 1: Detailed Analysis                                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Component/Route/State/DB Mapping → as_is_spec.md        │   │
│  └─────────────────────────────────────┬───────────────────┘   │
│                                        ▼                        │
│  Phase 2: Intelligent Planning                                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ Strategy A   │ │ Strategy B   │ │ Strategy C   │            │
│  │ Incremental  │ │   Phased     │ │  Big-Bang    │            │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘            │
│         └────────────────┼────────────────┘                    │
│                          ▼                                      │
│           ┌────────────────────────┐                            │
│           │ Dynamic Skill Discovery│                            │
│           │ + migration_plan.md    │                            │
│           └────────────┬───────────┘                            │
│                        ▼                                        │
│  Phase 3: DDD Execution (Per Module)                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  FOR EACH module:                                        │   │
│  │    ├── ANALYZE  (소스 동작 이해 + DB 모델 식별)             │   │
│  │    ├── PRESERVE (특성 테스트 + DB 레이어 테스트)            │   │
│  │    ├── IMPROVE  (타겟 프레임워크 + ORM 변환)              │   │
│  │    ├── LSP Quality Gate (regression check)               │   │
│  │    └── Self-Assessment:                                  │   │
│  │        ├── SUCCESS → Next module                         │   │
│  │        ├── LSP REGRESSION → Pivot approach               │   │
│  │        ├── 3x FAIL → Pivot approach                      │   │
│  │        └── Complexity >90 → User guidance                │   │
│  └─────────────────────────────────────────────────────────┘   │
│                        ▼                                        │
│  Phase 4: Verification (Playwright E2E)                         │
│  ┌──────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────┐  │
│  │  Visual  │ │ Cross-Browser│ │  Core Web    │ │   A11y   │  │
│  │Regression│ │   Testing    │ │   Vitals     │ │  (axe)   │  │
│  └──────────┘ └──────────────┘ └──────────────┘ └──────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 관련 파일

| 파일 | 설명 |
|------|------|
| `templates/.claude/commands/jikime/friday.md` | F.R.I.D.A.Y. 슬래시 커맨드 (587줄) |
| `templates/.claude/commands/jikime/migrate-0-discover.md` | Phase 0 커맨드 |
| `templates/.claude/commands/jikime/migrate-1-analyze.md` | Phase 1 커맨드 |
| `templates/.claude/commands/jikime/migrate-2-plan.md` | Phase 2 커맨드 |
| `templates/.claude/commands/jikime/migrate-3-execute.md` | Phase 3 커맨드 |
| `templates/.claude/commands/jikime/migrate-4-verify.md` | Phase 4 커맨드 |
| `docs/migration.md` | 마이그레이션 시스템 문서 |
| `docs/migrate-playwright.md` | Playwright 검증 시스템 |
| `docs/jarvis.md` | J.A.R.V.I.S. 개발 오케스트레이터 문서 |
| `templates/.jikime/config/quality.yaml` | LSP Quality Gates 설정 |

## Config-First Approach

F.R.I.D.A.Y.의 핵심 설계 원칙은 **Config-First**입니다. Phase 0에서 설정 파일을 한 번 생성하면, 이후 모든 단계에서 자동 참조합니다.

### .migrate-config.yaml

```yaml
# Phase 0에서 자동 생성
project_name: my-vue-app
source_path: ./legacy-vue-app
source_architecture: monolith           # Phase 0에서 감지 (monolith, separated, unknown)
target_framework: nextjs16
artifacts_dir: ./migrations/my-vue-app
output_dir: ./migrations/my-vue-app/out
db_type: postgresql                     # Phase 0에서 감지
db_orm: eloquent                        # Phase 0에서 감지

# Phase 2에서 추가
target_architecture: fullstack-monolith  # 사용자 선택 (fullstack-monolith, frontend-backend, frontend-only)
db_access_from: frontend                 # target_architecture에서 자동 파생
# target_framework_backend: fastapi      # frontend-backend 아키텍처만
```

### 산출물 흐름

```
@<source-path>/
    │
    ▼ (Phase 0-1: Discover + Analyze)
.migrate-config.yaml                  ← 프로젝트 설정 (source_architecture, db_type, db_orm 포함)
{artifacts_dir}/as_is_spec.md         ← 상세 분석 (Database Layer + Architecture Layers 포함)
    │
    ▼ (Phase 2: Plan)
.migrate-config.yaml 업데이트          ← 아키텍처 선택 (target_architecture, db_access_from)
{artifacts_dir}/migration_plan.md     ← 마이그레이션 계획 (아키텍처별 Phase 구조)
    │
    ▼ (Phase 3: Execute)
{output_dir}/                         ← 마이그레이션된 프로젝트 (아키텍처에 따라 구조 상이)
  ├─ fullstack-monolith: {output_dir}/ (단일)
  ├─ frontend-backend: {output_dir}/frontend/ + {output_dir}/backend/
  └─ frontend-only: {output_dir}/ (단일, DB 없음)
{artifacts_dir}/progress.yaml         ← 진행 상태 추적
    │
    ▼ (Phase 4: Verify)
{artifacts_dir}/verification_report.md ← 검증 결과 (아키텍처별 검증)
    │
    ▼ (Optional: Whitepaper)
{whitepaper_output}/                  ← 클라이언트용 보고서
```

## Usage

### 기본 사용법

```bash
# 전체 자동 오케스트레이션 (권장)
/jikime:friday "Vue 앱을 Next.js로 마이그레이션" @./legacy-vue-app/

# 타겟 프레임워크 명시
/jikime:friday @./my-app/ --target fastapi

# 안전 전략 (보수적 접근)
/jikime:friday @./legacy/ "Migrate to Go" --strategy safe

# 자동 루프 활성화
/jikime:friday @./src/ --target nextjs --loop --max 30

# 중단된 마이그레이션 재개
/jikime:friday resume

# 백서(Whitepaper) 생성
/jikime:friday @./app/ --target nextjs --whitepaper --client "ABC Corp" --lang ko
```

### 단계별 수동 실행

```bash
/jikime:migrate-0-discover @./my-vue-app/ --target nextjs
/jikime:migrate-1-analyze
/jikime:migrate-2-plan
/jikime:migrate-3-execute
/jikime:migrate-4-verify --full
```

### 명령어 옵션

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `@<path>` | 소스 프로젝트 경로 | 현재 디렉토리 |
| `--target` | 타겟 프레임워크 (nextjs, fastapi, go, flutter 등) | 자동 감지 |
| `--strategy` | 실행 전략: auto, safe, fast | auto |
| `--loop` | 에러 자동 수정 반복 활성화 | false |
| `--max N` | 최대 반복 횟수 | 50 |
| `--whitepaper` | 마이그레이션 백서 생성 | false |
| `--client` | 클라이언트 회사명 (백서 표지용) | - |
| `--lang` | 백서 언어 (ko, en, ja, zh) | conversation_language |
| `resume` | 이전 마이그레이션 재개 | - |

## Intelligence Features

### 1. Discovery (Phase 0) - 3-Way Parallel

3개의 전문 에이전트가 **동시에** 소스 프로젝트를 분석합니다:

| 에이전트 | 역할 | 출력 |
|----------|------|------|
| **Codebase Explorer** | 파일 구조, 프레임워크 감지, DB/ORM 유형, 소스 아키텍처 패턴 | 기술 스택, 컴포넌트 목록, DB 정보, 소스 아키텍처 (monolith/separated/unknown), 복잡도 점수 |
| **Dependency Analyzer** | 패키지 의존성, 버전 호환성, 브레이킹 체인지 | 의존성 맵, 업그레이드 요구사항 |
| **Risk Assessor** | 마이그레이션 리스크, 안티패턴, 레거시 락 | 리스크 점수, 차단 요인 식별 |

### 2. Analysis (Phase 1) - as_is_spec.md 생성

소스 프로젝트의 전체 구조를 문서화합니다:

- 컴포넌트/라우트/상태 매핑
- 데이터베이스 레이어 분석 (모델, 쿼리 패턴, 외부 데이터 서비스)
- 아키텍처 레이어 분석 (Frontend / Backend / Data / Shared 레이어 식별 + Coupling 분석)
- 비즈니스 로직 문서화
- API 엔드포인트 매핑
- 의존성 및 리스크 평가

### 3. Planning (Phase 2) - Architecture Selection + Dynamic Skill Discovery

F.R.I.D.A.Y.는 하드코딩된 프레임워크 패턴 없이, **동적으로 스킬을 탐색**합니다:

```bash
# 자동으로 실행되는 내부 프로세스
jikime-adk skill search "{target_framework}"
```

**아키텍처 패턴 선택** (Phase 2의 핵심 단계):

1. `source_architecture`와 `Architecture Layers` 분석 결과를 기반으로 추천
2. 사용자에게 3가지 옵션 제시: `fullstack-monolith` / `frontend-backend` / `frontend-only`
3. `frontend-backend` 선택 시 백엔드 프레임워크 후속 질문 (FastAPI/NestJS/Express/Go)
4. `.migrate-config.yaml` 업데이트 (`target_architecture`, `target_framework_backend`, `db_access_from`)

2-3개의 마이그레이션 전략을 생성하고 비교합니다:

| 전략 | 리스크 | 속도 | 적합한 경우 |
|------|--------|------|-------------|
| **Incremental** | 낮음 | 느림 | 복잡도 > 70 |
| **Phased** | 중간 | 중간 | 복잡도 40-70 |
| **Big-Bang** | 높음 | 빠름 | 복잡도 < 40 |

### 4. Execution (Phase 3) - DDD Migration Cycle

`target_architecture`에 따라 실행 전략이 달라집니다:

| 아키텍처 | 실행 방식 |
|----------|----------|
| `fullstack-monolith` | 단일 프로젝트 DDD 사이클 (기본) |
| `frontend-backend` | Shared → Backend → Frontend → Integration 4단계 분리 실행 |
| `frontend-only` | 프론트엔드 모듈만 DDD 사이클 (DB 단계 스킵) |

각 모듈에 대해 ANALYZE-PRESERVE-IMPROVE 사이클을 수행합니다:

```
ANALYZE:     소스 컴포넌트 동작 이해
ANALYZE-DB:  데이터 모델 및 쿼리 패턴 식별 (DB가 있는 경우)
PRESERVE:    특성 테스트 작성 (현재 동작 캡처)
PRESERVE-DB: 데이터 레이어 테스트 작성 (DB가 있는 경우)
IMPROVE:     타겟 프레임워크로 변환 (스킬 컨벤션 적용)
IMPROVE-DB:  ORM/데이터 접근 패턴 변환 (DB가 있는 경우)
```

#### LSP Quality Gates

Phase 3 실행 중 LSP 기반 품질 게이트가 자동으로 적용됩니다:

| Phase | 조건 | 설명 |
|-------|------|------|
| **plan** | `require_baseline: true` | Migration plan 수립 시 LSP 베이스라인 캡처 |
| **execute** | `max_errors: 0` | 타입에러/린트에러 모두 0 필요 |
| **verify** | `require_clean_lsp: true` | 검증 전 LSP 클린 상태 필수 |

설정 위치: `.jikime/config/quality.yaml` → `constitution.lsp_quality_gates`

#### Ralph Loop 통합

F.R.I.D.A.Y.의 DDD Migration Cycle은 LSP Quality Gates와 통합됩니다:

```
Ralph Loop Cycle (Migration):
  1. ANALYZE: 소스 컴포넌트 분석 + LSP 베이스라인 캡처
  2. PRESERVE: Characterization test 생성
  3. IMPROVE: 타겟 프레임워크로 변환
  4. LSP Check: 변환 후 LSP 진단 (regression 체크)
  5. Decision: Continue, Retry, or Pivot
```

LSP regression이 감지되면 F.R.I.D.A.Y.는 자동으로 대안 마이그레이션 패턴을 시도합니다.

#### 자가 평가 루프

각 모듈 변환 시 F.R.I.D.A.Y.가 자동으로 평가합니다:

1. **"현재 모듈이 정상적으로 마이그레이션되고 있는가?"**
   - TypeScript 컴파일 성공?
   - 특성 테스트 통과?
   - 빌드 성공?

2. **"접근 방식을 변경해야 하는가?"**
   - 트리거: 같은 모듈에서 3회 연속 실패
   - 행동: 대안 마이그레이션 패턴 시도

3. **"자동 마이그레이션이 불가능한 복잡도인가?"**
   - 트리거: 단일 컴포넌트 복잡도 > 90
   - 행동: 서브 컴포넌트로 분할 또는 사용자 지침 요청

#### Progress Tracking

```yaml
# {artifacts_dir}/progress.yaml
project: my-vue-app
source: vue3
target: nextjs16
target_architecture: fullstack-monolith  # 선택된 아키텍처 패턴
status: in_progress
strategy: phased

phases:
  discover: completed
  analyze: completed
  plan: completed
  execute: in_progress
  verify: pending

modules:
  total: 15
  completed: 8
  in_progress: 1
  failed: 0
  pending: 6

current:
  module: UserProfile
  iteration: 2
```

### 5. Verification (Phase 4) - Playwright E2E

10단계 검증 시스템으로 마이그레이션 품질을 보장합니다:

| 단계 | 검증 항목 | 도구 |
|------|----------|------|
| 1 | 인프라 (Dev server 자동 시작) | Playwright |
| 2 | 라우트 발견 (테스트 가능한 라우트 탐색) | Explore |
| 3 | 시각적 회귀 (5개 뷰포트 비교) | Playwright Screenshots |
| 4 | 동작 테스트 (폼, API, JS 에러) | Playwright Actions |
| 5 | 크로스 브라우저 (Chromium/Firefox/WebKit) | Playwright Multi-Browser |
| 6 | 성능 (Core Web Vitals, 번들 사이즈) | Playwright Metrics |
| 7 | 접근성 (WCAG 준수) | axe-core |
| 8 | 에이전트 위임 (e2e-tester + 스킬) | Task Agent |
| 9 | 통합 검증 (플래그, 의존성) | CLI |
| 10 | 리포트 (Markdown + HTML + JSON) | Write |

## Strategy Details

### auto (기본값)

F.R.I.D.A.Y.가 마이그레이션 복잡도를 분석하여 최적 전략을 자동 선택:

| 마이그레이션 유형 | 분석 결과 | 선택 전략 |
|-----------------|----------|----------|
| 단순 (단일 프레임워크, <20 컴포넌트) | 복잡도 < 40 | Big-Bang (직접 순차) |
| 중간 (2-3 관심사, 20-50 컴포넌트) | 복잡도 40-70 | Phased (체크포인트) |
| 복잡 (멀티 도메인, >50 컴포넌트) | 복잡도 > 70 | Incremental (병렬 오케스트레이션) |

### safe (보수적)

최대한의 검증과 안전장치를 적용:

- 매 Phase 사이에 사용자 확인
- 컴포넌트별 개별 마이그레이션
- 각 단계마다 전체 테스트 스위트 실행
- 모든 Phase에 롤백 포인트

### fast (공격적)

소규모 마이그레이션을 위한 빠른 실행:

- 최소한의 체크포인트 (Phase 단위만)
- 배치 컴포넌트 마이그레이션
- 선택적 검증 건너뛰기
- 빠른 완료 우선

## Framework Agnosticism

F.R.I.D.A.Y.는 **프레임워크 무관** 설계입니다. 하드코딩된 타겟 프레임워크 패턴 없이, 모든 지식은 동적으로 수집됩니다:

| 소스 | 지식 원천 |
|------|----------|
| **Skills** | `jikime-adk skill search "{target_framework}"` |
| **Context7** | 스킬이 없을 때 폴백 |
| **as_is_spec.md** | 소스 분석 데이터 |

### 지원 마이그레이션 (비제한적)

| 소스 | 타겟 옵션 |
|------|----------|
| Vue 2/3 | Next.js (App Router) |
| React (CRA) | Next.js (App Router) |
| Angular | Next.js, SvelteKit |
| jQuery | React, Vue, Svelte |
| PHP | Next.js, FastAPI, Go |
| Monolith | Microservices |
| Any source | Any target |

## Whitepaper Generation

`--whitepaper` 플래그를 사용하면 마이그레이션 완료 후 클라이언트용 보고서를 생성합니다:

### 산출물 구조

```
{whitepaper_output}/
├── 00_cover.md                    # 표지 + 목차
├── 01_executive_summary.md        # 비기술적 요약
├── 02_migration_summary.md        # 실행 타임라인
├── 03_architecture_comparison.md  # Before/After 다이어그램
├── 04_component_inventory.md      # 마이그레이션된 컴포넌트 목록
├── 05_performance_report.md       # 성능 메트릭
├── 06_quality_report.md           # 품질 메트릭
└── 07_lessons_learned.md          # 권장사항
```

### 지원 언어

| 코드 | 언어 |
|------|------|
| ko | 한국어 |
| en | English |
| ja | 日本語 |
| zh | 中文 |

## Resume Capability

중단된 마이그레이션을 이어서 진행할 수 있습니다:

```bash
/jikime:friday resume
```

내부 동작:
1. `.migrate-config.yaml`에서 프로젝트 설정 읽기
2. `{artifacts_dir}/progress.yaml`에서 현재 상태 확인
3. 마지막 완료된 Phase 확인
4. 다음 대기 중인 Phase/모듈부터 계속 진행
5. 전략과 컨텍스트 복원

## Agent Delegation

### Phase별 에이전트 위임

| Phase | 에이전트 | 역할 |
|-------|---------|------|
| Phase 0 | Explore (x3) | 코드 분석, 의존성 분석, 리스크 평가 |
| Phase 1 | manager-spec | as_is_spec.md 생성 |
| Phase 2 | manager-strategy | 마이그레이션 전략 수립 |
| Phase 3 | backend, frontend | DDD 기반 코드 마이그레이션 |
| Phase 4 | e2e-tester | Playwright 검증 |

## Output Format

### 실행 중

```markdown
## F.R.I.D.A.Y.: Phase 3 - Execution (Module 8/15)

### Migration: Vue 3 → Next.js 16
### Complexity Score: 55/100

### Module Status
- [x] Auth module (5 components)
- [x] Users module (3 components)
- [ ] Products module <- in progress
- [ ] Orders module
- [ ] Dashboard module

### Self-Assessment
- Progress: YES (build errors: 3 -> 1)
- Pivot needed: NO
- Current module confidence: 80%

### Active Issues
- WARNING: ProductCard.tsx - dynamic import pattern needs manual review

Continuing...
```

### 완료

```markdown
## F.R.I.D.A.Y.: MIGRATION COMPLETE

### Summary
- Source: Vue 3 (Vuetify)
- Target: Next.js 16 (App Router)
- Strategy Used: Phased
- Modules Migrated: 15/15
- Tests: 89/89 passing
- Build: SUCCESS
- Iterations: 12
- Self-Corrections: 2

### Predictive Suggestions
1. Set up CI/CD pipeline for the new project
2. Configure production environment variables
3. Set up monitoring and error tracking
4. Plan user acceptance testing

<jikime>MIGRATION_COMPLETE</jikime>
```

## Limitations & Safety

### 제한사항

- 최대 3회 전략 피봇 (이후 사용자 개입 요청)
- 모듈당 최대 5회 재시도
- 세션 내 학습만 지원 (세션 간 학습 미지원)

### 안전장치

- [HARD] 모든 구현은 전문가 에이전트에 위임
- [HARD] Phase 3 실행 전 사용자 확인 필수 (--strategy fast 제외)
- [HARD] 완료 마커 필수: `<jikime>MIGRATION_COMPLETE</jikime>`
- [HARD] 동적 스킬 탐색 - 프레임워크 패턴 하드코딩 금지
- [HARD] `.migrate-config.yaml`과 `as_is_spec.md`에서 읽기 - 소스 재분석 금지
- [HARD] LSP Quality Gate: execute phase에서 에러 0 필수
- 각 Phase에 롤백 포인트 생성
- LSP Quality Gates가 regression 감지 시 자동 알림

## Best Practices

### 언제 F.R.I.D.A.Y.를 사용하나요?

**적합한 경우:**
- 레거시 프로젝트를 현대 프레임워크로 전환
- 프레임워크 업그레이드 (Vue 2 → Vue 3, React CRA → Next.js)
- 모놀리식 → 마이크로서비스 전환
- 프론트엔드 프레임워크 교체

**J.A.R.V.I.S.가 나은 경우:**
- 새로운 기능 구현
- 기존 코드 리팩토링 (프레임워크 변경 없이)
- 버그 수정
- 성능 최적화

### Dual Orchestrator 전환

```bash
# 개발 작업 → J.A.R.V.I.S.
/jikime:jarvis "Add user authentication"

# 마이그레이션 작업 → F.R.I.D.A.Y.
/jikime:friday "Migrate Vue app to Next.js 16"

# 자동 라우팅 (키워드 기반)
"migrate this app" → F.R.I.D.A.Y.
"implement login" → J.A.R.V.I.S.
```

---

Version: 1.3.0
Last Updated: 2026-02-03
Codename: F.R.I.D.A.Y. (Framework Relay & Integration Deployment Assistant Yesterday)
Inspiration: Iron Man's second AI Assistant (successor to J.A.R.V.I.S.)
Changelog:
- v1.3.0: Architecture Pattern 지원 추가 (fullstack-monolith, frontend-backend, frontend-only); 아키텍처 선택 단계(Phase 2); 아키텍처별 실행/검증 전략; source_architecture 감지
- v1.2.0: Database Layer 지원 추가 (DB/ORM 감지, DB-aware DDD cycle, DB 스키마 검증)
- v1.1.0: LSP Quality Gates 통합, Ralph Loop Integration 추가
- v1.0.0: Initial release - Migration-focused orchestrator extracted from J.A.R.V.I.S.
