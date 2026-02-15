# JikiME-ADK Migration System

## Overview

JikiME-ADK Migration System은 F.R.I.D.A.Y. 오케스트레이터를 통해 레거시 프로젝트를 현대적 프레임워크로 마이그레이션하는 지능형 시스템입니다. **Config-First 접근 방식**으로 소스 경로와 타겟 프레임워크를 한 번만 입력하면, 이후 모든 단계에서 자동으로 참조합니다.

## Core Design Principles

| 원칙 | 설명 |
|------|------|
| **한 번 입력, 전체 활용** | 소스 경로/타겟은 Step 0에서 1회만 입력 |
| **Config-First** | `.migrate-config.yaml`이 모든 설정의 단일 진실 공급원 |
| **프레임워크 무관** | 하드코딩 없이 동적 스킬 탐색으로 규칙 적용 |
| **DDD 방법론** | ANALYZE-PRESERVE-IMPROVE 사이클로 동작 보존 |

---

## Step-by-Step Workflow

```
/jikime:migrate-0-discover   → Step 0: 소스 탐색 + config 생성
        ↓
/jikime:migrate-1-analyze    → Step 1: 상세 분석 + config 업데이트
        ↓
/jikime:migrate-2-plan       → Step 2: 계획 수립 (승인 대기)
        ↓
/jikime:migrate-3-execute    → Step 3: DDD 실행
        ↓
/jikime:migrate-4-verify     → Step 4: 검증 + 보고서
        ↓
/jikime:verify --browser-only --fix-loop  → (선택) 런타임 에러 감지 + 자동 수정

또는

/jikime:friday "설명" @<path>  → 전체 자동 오케스트레이션
```

### Quick Start

```bash
# 방법 1: FRIDAY 자동 오케스트레이션 (권장)
/jikime:friday "Vue 앱을 Next.js로 마이그레이션" @./my-vue-app/ --target nextjs

# 방법 2: 단계별 수동 실행
/jikime:migrate-0-discover @./my-vue-app/ --target nextjs
/jikime:migrate-1-analyze
/jikime:migrate-2-plan
/jikime:migrate-3-execute
/jikime:migrate-4-verify --full
```

> **Note**: Step 0에서 경로와 타겟을 지정하면, 이후 단계에서는 인자 없이 실행할 수 있습니다.

---

## Config-First Approach

### `.migrate-config.yaml` (단일 진실 공급원)

Step 0에서 자동 생성되며, 이후 모든 단계에서 참조합니다:

```yaml
version: "1.0"
project_name: my-vue-app
source_path: ./my-vue-app
source_framework: vue3              # Step 0에서 감지
source_architecture: monolith       # Step 0에서 감지 (monolith, separated, unknown)
target_framework: nextjs16          # Step 0에서 --target으로 지정
artifacts_dir: ./migrations/my-vue-app
output_dir: ./migrations/my-vue-app/out
db_type: postgresql                 # Step 0에서 감지 (postgresql, mysql, sqlite, mongodb, none)
db_orm: eloquent                    # Step 0에서 감지 (prisma, drizzle, typeorm, sequelize, mongoose, eloquent, none)
created_at: "2026-01-23T10:00:00Z"
# Step 1에서 추가되는 필드
analyzed_at: "2026-01-23T11:00:00Z"
component_count: 45
complexity_score: 7
db_model_count: 15                  # Step 1에서 추가 (0 if no database)
# Step 2에서 추가되는 필드
target_architecture: fullstack-monolith  # Step 2에서 사용자 선택 (fullstack-monolith, frontend-backend, frontend-only)
target_framework_backend: fastapi        # Step 2에서 선택 (frontend-backend 아키텍처만, fastapi, nestjs, express, go-fiber)
db_access_from: backend                  # Step 2에서 자동 파생 (frontend, backend, both, none)
```

### Config 필드 생명주기

| 필드 | 생성 단계 | 사용 단계 |
|------|-----------|-----------|
| `source_path` | Step 0 | Step 1, 3 |
| `source_framework` | Step 0 | Step 1, 2, 3 |
| `source_architecture` | Step 0 | Step 1, 2 |
| `target_framework` | Step 0 or 1 | Step 2, 3 |
| `artifacts_dir` | Step 0 | Step 2, 3, 4 |
| `output_dir` | Step 0 | Step 3, 4 |
| `db_type` | Step 0 | Step 1, 2, 3, 4 |
| `db_orm` | Step 0 | Step 1, 2, 3, 4 |
| `component_count` | Step 1 | Step 2 |
| `complexity_score` | Step 1 | Step 2 |
| `db_model_count` | Step 1 | Step 2 |
| `target_architecture` | Step 2 | Step 3, 4 |
| `target_framework_backend` | Step 2 | Step 3, 4 |
| `db_access_from` | Step 2 | Step 3, 4 |

---

## Command Reference

### Step 0: Discover (소스 탐색)

```bash
/jikime:migrate-0-discover @<path> [--target <framework>] [--quick]
```

| Option | Required | Description |
|--------|----------|-------------|
| `@path` | Yes | 분석할 소스 코드 경로 |
| `--target` | No | 타겟 프레임워크 (`nextjs`\|`fastapi`\|`go`\|`flutter`) |
| `--quick` | No | 빠른 개요만 (상세 분석 생략) |

**하는 일**:
- 기술 스택 감지 (언어, 프레임워크, 버전)
- 아키텍처 패턴 파악
- 소스 아키텍처 패턴 감지 (monolith / separated / unknown)
- 데이터베이스 유형 및 ORM 감지
- 마이그레이션 복잡도 평가
- `.migrate-config.yaml` 생성
- 타겟 미지정 시 추천 프레임워크 제시

**산출물**: `.migrate-config.yaml`, Discovery Report

---

### Step 1: Analyze (상세 분석)

```bash
/jikime:migrate-1-analyze [project-path] [options]
```

| Option | Required | Description |
|--------|----------|-------------|
| `project-path` | No* | 레거시 프로젝트 경로 (config에서 자동 읽음) |
| `--framework` | No | 소스 프레임워크 강제 지정 (`vue`\|`react`\|`angular`\|`svelte`\|`auto`) |
| `--target` | No | 타겟 프레임워크 (config 값 override) |
| `--artifacts-output` | No | 산출물 경로 (기본: `./migrations/{project}/`) |
| `--whitepaper` | No | 클라이언트 제안용 백서 패키지 생성 |
| `--whitepaper-output` | No | 백서 출력 경로 (기본: `./whitepaper/`) |
| `--client` | No | 클라이언트 회사명 (백서 표지용) |
| `--lang` | No | 백서 언어 (`ko`\|`en`\|`ja`\|`zh`) |

*\* `.migrate-config.yaml`이 있으면 자동으로 읽힙니다. 없으면 필수입니다.*

**경로 우선순위**:
1. 명시적 인자: `/jikime:migrate-1-analyze "./my-app" --target nextjs`
2. Config 파일: `.migrate-config.yaml` → `source_path`, `target_framework`
3. 에러: 둘 다 없으면 Step 0 먼저 실행 안내

**하는 일**:
- 컴포넌트 구조 및 계층 분석
- 라우팅 구조 매핑
- 상태 관리 패턴 파악
- 데이터베이스 레이어 분석 (모델, 쿼리 패턴, 외부 데이터 서비스)
- 아키텍처 레이어 분석 (Frontend / Backend / Data / Shared 레이어 식별)
- 의존성 호환성 분석
- 위험 요소 식별

**산출물**:
- `{artifacts_dir}/as_is_spec.md`
- `.migrate-config.yaml` 업데이트 (component_count, complexity_score 추가)

---

### Step 2: Plan (계획 수립)

```bash
/jikime:migrate-2-plan [--modules <list>] [--incremental]
```

| Option | Required | Description |
|--------|----------|-------------|
| `--modules` | No | 특정 모듈만 계획 (예: `auth,users,orders`) |
| `--incremental` | No | 점진적 마이그레이션 계획 |

**하는 일**:
1. `.migrate-config.yaml`에서 `target_framework` 읽기
2. 동적 스킬 탐색 (`jikime-adk skill search "{target_framework}"`, `"{db_orm}"`, `"{db_type}"`)
3. 타겟 아키텍처 패턴 선택 (fullstack-monolith / frontend-backend / frontend-only)
4. `{artifacts_dir}/as_is_spec.md` 기반 계획 수립
5. 데이터베이스 마이그레이션 전략 수립
6. 스킬 규칙(구조, 네이밍, 라우팅) 적용
7. **사용자 승인 대기**

**산출물**: `{artifacts_dir}/migration_plan.md`

**승인 방법**:
- `yes` - 계획대로 진행
- `modify: [변경사항]` - 계획 수정
- `no` - 취소

**제약사항**: 소스 코드를 직접 분석하지 않습니다. 반드시 `as_is_spec.md`만 참조합니다.

---

### Step 3: Execute (실행)

```bash
/jikime:migrate-3-execute [--module <name>] [--resume] [--dry-run]
```

| Option | Required | Description |
|--------|----------|-------------|
| `--module` | No | 특정 모듈만 마이그레이션 |
| `--resume` | No | 중단된 마이그레이션 재개 (`progress.yaml` 기반) |
| `--dry-run` | No | 실제 실행 없이 미리보기 |

**방법론**: DDD (ANALYZE → PRESERVE → IMPROVE)

**아키텍처별 실행 전략**:

| 아키텍처 | 실행 방식 | 출력 구조 |
|----------|----------|----------|
| `fullstack-monolith` | 단일 프로젝트 DDD 사이클 | `{output_dir}/` |
| `frontend-backend` | Shared → Backend → Frontend → Integration 4단계 | `{output_dir}/frontend/` + `{output_dir}/backend/` |
| `frontend-only` | 프론트엔드 모듈만 DDD 사이클 (DB 스킵) | `{output_dir}/` |

```
각 모듈별 반복 (fullstack-monolith 기준):
  1. ANALYZE     - 소스 모듈 동작 이해
  1.5 ANALYZE-DB - 데이터 모델 및 쿼리 식별 (DB가 있는 경우)
  2. PRESERVE    - 특성 테스트 작성 (동작 보존)
  2.5 PRESERVE-DB - 데이터 레이어 테스트 작성 (DB가 있는 경우)
  3. IMPROVE     - 타겟 프레임워크로 변환
  3.5 IMPROVE-DB - ORM/데이터 접근 패턴 변환 (DB가 있는 경우)
  4. Validate    - 빌드 + 테스트 + DB 스키마 확인

frontend-backend의 경우:
  Sub-Phase 1: Shared Layer (공유 타입, API 계약 정의)
  Sub-Phase 2: Backend (API + 비즈니스 로직 + 데이터 접근)
  Sub-Phase 3: Frontend (컴포넌트 + 라우팅 + 상태 + API Client)
  Sub-Phase 4: Integration (API 계약 일치 검증)
```

**산출물**:
- `{output_dir}/` - 마이그레이션된 프로젝트 (아키텍처에 따라 구조 상이)
- `{artifacts_dir}/progress.yaml` - 진행 상황 추적

**progress.yaml 구조**:
```yaml
project: my-vue-app
source_framework: vue3
target_framework: nextjs16
status: in_progress
modules:
  total: 15
  completed: 8
  in_progress: 1
  failed: 0
  pending: 6
```

---

### Step 4: Verify (검증)

```bash
/jikime:migrate-4-verify [options]
```

| Option | Required | Description |
|--------|----------|-------------|
| `--full` | No | 모든 검증 유형 실행 (visual + cross-browser + a11y + performance) |
| `--behavior` | No | 동작 보존 비교만 |
| `--e2e` | No | E2E 테스트만 |
| `--visual` | No | 스크린샷 기반 시각적 회귀 검증 |
| `--performance` | No | Core Web Vitals 및 로드 시간 비교 |
| `--cross-browser` | No | Chromium, Firefox, WebKit 교차 브라우저 검증 |
| `--a11y` | No | WCAG 접근성 검증 (axe-core) |
| `--source-url` | No | 소스 시스템 URL (라이브 비교용) |
| `--target-url` | No | 타겟 시스템 URL (라이브 비교용) |
| `--headed` | No | 브라우저 창 표시 (디버깅용) |
| `--capture-skill` | No | 검증된 마이그레이션 패턴을 재사용 가능한 스킬로 저장 |

> **Note**: `--source-url`/`--target-url`은 실행 중인 인스턴스 비교용입니다. 소스/타겟 프레임워크 정보는 `.migrate-config.yaml`에서 자동으로 읽습니다.

**아키텍처별 검증**:

| 아키텍처 | 검증 대상 | DB 검증 |
|----------|----------|---------|
| `fullstack-monolith` | 단일 프로젝트 빌드/타입체크/린트 | 포함 |
| `frontend-backend` | Frontend + Backend 개별 검증 + 연동 검증 | Backend에서 실행 |
| `frontend-only` | 단일 프로젝트 빌드/타입체크/린트 | 스킵 |

**검증 항목**:
1. Dev Server Setup - 소스/타겟 개발 서버 자동 시작
2. Route Discovery - 마이그레이션 산출물에서 테스트 가능한 라우트 탐색
3. Characterization Tests - 동작 보존 테스트
4. Behavior Comparison - 소스/타겟 출력 비교
5. E2E Tests - Playwright 기반 사용자 흐름 검증
6. Visual Regression - 스크린샷 비교 (소스 vs 타겟)
7. Performance Check - Core Web Vitals, 로드 시간 비교
8. Cross-Browser - Chromium, Firefox, WebKit 검증
9. Accessibility - axe-core 기반 WCAG 준수 확인
10. Skill Capture - 검증된 패턴을 재사용 가능한 스킬로 저장 (`--capture-skill` 옵션)

**산출물**:
- `{artifacts_dir}/verification_report.md`
- `skills/jikime-migration-{source}-to-{target}/` (`--capture-skill` 사용 시)

### 런타임 에러 감지: verify --browser-only

마이그레이션 검증 후에도 **브라우저에서만 발생하는 런타임 에러**가 남아있을 수 있습니다. 정적 분석이나 빌드 도구가 감지하지 못하는 에러(undefined 참조, 잘못된 라이브러리 import 등)를 잡으려면 `/jikime:verify --browser-only`를 활용하세요.

```bash
# 마이그레이션 검증 완료 후, 런타임 에러 추가 확인
cd {output_dir}
/jikime:verify --browser-only

# 특정 라우트만 확인
/jikime:verify --browser-only --routes /,/dashboard,/settings

# 에러 보고만 (수정 안함, fix-loop 없이)
/jikime:verify --browser-only

# 자동 수정 루프 활성화
/jikime:verify --browser-only --fix-loop

# 브라우저 창 표시 (headed 모드)
/jikime:verify --browser-only --headed
```

**verify --browser-only 동작 방식**:
1. package.json에서 패키지 매니저 감지 (pnpm/yarn/npm/bun)
2. `dev` 스크립트로 개발 서버 시작 (백그라운드)
3. Playwright로 각 라우트 탐색하며 에러 캡처
4. 스택 트레이스에서 소스 파일:라인 추출
5. `--fix-loop` 사용 시: 전문 에이전트에 수정 위임 (자동)
6. 재검증 루프 (에러 0개까지 반복)

> **Tip**: `migrate-4-verify`는 마이그레이션 후 **정적 분석 검증**에 초점을 맞추고, `verify --browser-only --fix-loop`는 **런타임 에러 감지 및 자동 수정**에 초점을 맞춥니다. 마이그레이션 후 순차적으로 사용하면 가장 효과적입니다.

---

## F.R.I.D.A.Y. Orchestrator

전체 마이그레이션 프로세스를 자동으로 오케스트레이션합니다.

```bash
/jikime:friday "작업 설명" @<source-path> [options]
```

| Option | Description |
|--------|-------------|
| `@<source-path>` | 소스 프로젝트 경로 |
| `--target` | 타겟 프레임워크 (`nextjs`\|`fastapi`\|`go`\|`flutter`) |
| `--strategy` | 마이그레이션 전략 (`auto`\|`safe`\|`fast`) |
| `--loop` | 자동 반복 모드 |
| `--max N` | 최대 반복 횟수 (기본: 100) |
| `--whitepaper` | 클라이언트 납품용 백서 생성 |
| `--client` | 클라이언트명 |
| `--lang` | 백서 언어 (`ko`\|`en`\|`ja`\|`zh`) |
| `resume` | 중단된 작업 재개 |

### FRIDAY 실행 흐름

```
/jikime:friday "Vue→Next.js 마이그레이션" @./my-vue-app/ --target nextjs
    │
    ├─ Phase 1: Discovery
    │   └─ /jikime:migrate-0-discover @./my-vue-app/ --target nextjs
    │       → .migrate-config.yaml 생성
    │
    ├─ Phase 2: Analysis
    │   └─ /jikime:migrate-1-analyze
    │       → as_is_spec.md + config 업데이트
    │
    ├─ Phase 3: Planning
    │   └─ /jikime:migrate-2-plan
    │       → migration_plan.md (사용자 승인)
    │
    ├─ Phase 4: Execution
    │   └─ /jikime:migrate-3-execute
    │       → output_dir/ + progress.yaml
    │
    └─ Phase 5: Verification
        └─ /jikime:migrate-4-verify --full
            → verification_report.md
```

---

## Data Flow Diagram

```
사용자 입력: 소스 경로 + 타겟 (최초 1회만)
     │
     ▼
Step 0: .migrate-config.yaml 생성
     │  (source_path, source_framework, source_architecture, target_framework,
     │   db_type, db_orm, artifacts_dir, output_dir)
     │
     ▼
Step 1: config 업데이트 + as_is_spec.md 생성
     │  (component_count, complexity_score, db_model_count, analyzed_at 추가)
     │  (Architecture Layers 분석: Frontend/Backend/Data/Shared)
     │
     ▼
Step 2: 아키텍처 패턴 선택 + migration_plan.md 생성 (승인 대기)
     │  (target_architecture, target_framework_backend, db_access_from 추가)
     │  (동적 스킬 탐색 → 타겟 규칙 적용 + DB 마이그레이션 전략)
     │
     ▼
Step 3: output_dir/ 생성 + progress.yaml 업데이트
     │  (아키텍처별 실행 전략: monolith / frontend-backend / frontend-only)
     │  (모듈별 DDD 사이클 반복 + DB 레이어 변환)
     │
     ▼
Step 4: verification_report.md 생성
     │  (아키텍처별 검증 + 동작 보존 + 성능 검증 + DB 스키마/연결 검증)
     │
     ▼
(선택) --capture-skill
     │  (검증된 패턴을 스킬로 저장 → 다음 마이그레이션에 재사용)
     │
     ▼
(선택) /jikime:verify --browser-only --fix-loop
     │  (런타임 브라우저 에러 감지 + 자동 수정)
     │
     ▼
완료 → 스테이징 배포 → UAT → 프로덕션
```

---

## Dynamic Skill Discovery

Step 2에서 타겟 프레임워크에 맞는 스킬을 동적으로 탐색합니다:

```bash
# target_framework에 따라 자동 탐색
jikime-adk skill search "{target_framework}"
jikime-adk skill search "migrate {target_framework}"
jikime-adk skill search "{target_language}"
jikime-adk skill search "{db_orm}"
jikime-adk skill search "{db_type}"
```

| target_framework | 탐색되는 스킬 |
|------------------|---------------|
| `nextjs16` | `jikime-migrate-to-nextjs`, `jikime-nextjs@16`, `jikime-library-shadcn` |
| `fastapi` | `jikime-lang-python` (+ 관련 스킬) |
| `go-fiber` | `jikime-lang-go` (+ 관련 스킬) |
| `flutter` | `jikime-lang-flutter` (+ 관련 스킬) |

스킬이 없는 경우 Context7 MCP를 통해 공식 문서를 조회합니다.

---

## Architecture Patterns

Step 2 (Plan)에서 사용자가 타겟 아키텍처 패턴을 선택합니다. 소스 분석 결과에 따라 자동 추천되며, 사용자가 최종 결정합니다.

### 3가지 패턴

| 패턴 | 설명 | 적합한 경우 |
|------|------|------------|
| **fullstack-monolith** | 단일 Next.js 프로젝트 (API Routes + Server Components → DB) | 소규모~중규모, 모놀리식 소스 |
| **frontend-backend** | 프론트엔드(Next.js) + 백엔드(FastAPI/NestJS/Express/Go) 분리 | 대규모, 이미 분리된 소스 |
| **frontend-only** | 프론트엔드만 마이그레이션 (기존 백엔드 유지, API 호출) | 백엔드 유지 필요 시 |

### 선택 기준

```
source_architecture?
├─ monolith + component_count < 50 → 추천: fullstack-monolith
├─ monolith + component_count >= 50 → 추천: frontend-backend
├─ separated → 추천: frontend-backend
└─ unknown → 사용자에게 3가지 옵션 제시
```

### 디렉토리 구조

**fullstack-monolith** (기본):
```
{output_dir}/
├── src/
│   ├── app/          # Next.js App Router
│   ├── components/   # React 컴포넌트
│   ├── lib/          # 유틸리티, DB 클라이언트
│   └── stores/       # 상태 관리
├── prisma/           # DB 스키마
└── package.json
```

**frontend-backend**:
```
{output_dir}/
├── shared/           # 공유 타입, API 계약
│   └── types/
├── frontend/         # Next.js 프로젝트
│   ├── src/
│   └── package.json
└── backend/          # 백엔드 프로젝트 (FastAPI/NestJS/Express/Go)
    ├── src/
    ├── prisma/       # DB 스키마
    └── package.json
```

**frontend-only**:
```
{output_dir}/
├── src/
│   ├── app/          # Next.js App Router
│   ├── components/   # React 컴포넌트
│   ├── lib/          # API 클라이언트, 유틸리티
│   └── stores/       # 상태 관리
└── package.json      # DB 관련 없음
```

### Config 필드

```yaml
# Step 0에서 자동 감지
source_architecture: monolith    # monolith | separated | unknown

# Step 2에서 사용자 선택
target_architecture: fullstack-monolith  # fullstack-monolith | frontend-backend | frontend-only

# frontend-backend 선택 시 추가
target_framework_backend: fastapi  # fastapi | nestjs | express | go-fiber

# target_architecture에서 자동 파생
db_access_from: frontend          # frontend | backend | both | none
```

**기본값 (하위 호환)**: `target_architecture` 미설정 시 `fullstack-monolith`로 동작 (기존과 동일)

---

## Architecture

### Skills Structure

| Layer | Role | Example |
|-------|------|---------|
| **Migration Skills** | 프레임워크 전환 전략 | `jikime-migrate-to-nextjs` |
| **Version Skills** | 버전별 가이드 | `jikime-nextjs@16` |
| **Language Skills** | 언어별 패턴 | `jikime-lang-typescript` |
| **Domain Skills** | 도메인별 패턴 | `jikime-migration-patterns-auth` |

### Skills Naming Convention

```
Migration:      jikime-migrate-{source}-to-{target} 또는 jikime-migrate-to-{target}
Version Guide:  jikime-{framework}@{version}
Language:       jikime-lang-{language}
Domain Pattern: jikime-migration-patterns-{domain}
```

### MCP Integration

| MCP Server | Purpose |
|------------|---------|
| **Context7** | 공식 문서, 마이그레이션 가이드, API 변경사항 |
| **Playwright** | Step 4 검증 (E2E, 시각적 회귀, 크로스 브라우저), 런타임 에러 감지 |
| **WebFetch** | 최신 릴리스 노트, 브레이킹 체인지 |
| **Sequential** | 복잡한 마이그레이션 분석 |

---

## Supported Migrations

| Source | Target Options |
|--------|----------------|
| PHP (Laravel) | Next.js, FastAPI, Go, Spring Boot |
| jQuery | React, Vue, Svelte |
| Vue 2/3 | Next.js (App Router), Nuxt |
| React (CRA) | Next.js (App Router) |
| Angular | Next.js, SvelteKit |
| Java Servlet | Spring Boot, Go, FastAPI |
| Python 2 | Python 3, FastAPI |
| Svelte | SvelteKit |

---

## Best Practices

### 사용자 가이드

1. **Step 0부터 시작** - 항상 Discover로 시작하여 config를 생성하세요
2. **타겟은 한 번만** - Step 0에서 `--target` 지정 후 재입력 불필요
3. **Git 커밋 먼저** - 마이그레이션 전 반드시 현재 상태를 커밋하세요
4. **계획 검토** - Step 2에서 계획을 꼼꼼히 검토 후 승인하세요
5. **모듈별 진행** - 큰 프로젝트는 `--module` 옵션으로 점진적 실행
6. **중단 후 재개** - Step 3에서 `--resume`으로 이어서 진행 가능
7. **런타임 에러 확인** - Step 4 후 `/jikime:verify --browser-only --fix-loop`로 브라우저 런타임 에러까지 잡으세요
8. **경험을 스킬로 저장** - 성공적인 마이그레이션 후 `--capture-skill`로 패턴을 스킬화하여 다음 마이그레이션에 재사용하세요

### 스킬 작성자 가이드

1. **메타데이터 포함** - 프론트매터에 triggers, version 등 명시
2. **최신 문서 참조** - Context7/WebFetch 활용 지침 포함
3. **브레이킹 체인지** - 알려진 이슈와 해결법 문서화
4. **예시 제공** - Before/After 코드 예시 포함
5. **버전 명시** - 호환 버전 범위를 명확히 기재

---

Version: 3.4.0
Last Updated: 2026-02-03
Changelog:
- v3.4.0: Added Architecture Patterns section (fullstack-monolith, frontend-backend, frontend-only); Architecture-specific execution and verification; New config fields (source_architecture, target_architecture, target_framework_backend, db_access_from)
- v3.3.0: Added database layer support across all phases (db_type, db_orm, db_model_count); DB-aware DDD cycle; DB skill discovery
- v3.2.0: Added --capture-skill option to Step 4 for generating reusable migration skills from verified patterns
- v3.1.0: Step 4 Playwright-based verification details; Added verify --browser-only integration for runtime error detection
- v3.0.0: Config-First approach; FRIDAY orchestrator; Removed /jikime:migrate; Removed redundant source/target options from Steps 2-4; Renamed --source/--target to --source-url/--target-url in Step 4
- v2.0.0: Added Step-by-Step Workflow, Command Reference with full options
- v1.0.0: Initial migration system documentation
