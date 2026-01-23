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
target_framework: nextjs16          # Step 0에서 --target으로 지정
artifacts_dir: ./migrations/my-vue-app
output_dir: ./migrations/my-vue-app/out
created_at: "2026-01-23T10:00:00Z"
# Step 1에서 추가되는 필드
analyzed_at: "2026-01-23T11:00:00Z"
component_count: 45
complexity_score: 7
```

### Config 필드 생명주기

| 필드 | 생성 단계 | 사용 단계 |
|------|-----------|-----------|
| `source_path` | Step 0 | Step 1, 3 |
| `source_framework` | Step 0 | Step 1, 2, 3 |
| `target_framework` | Step 0 or 1 | Step 2, 3 |
| `artifacts_dir` | Step 0 | Step 2, 3, 4 |
| `output_dir` | Step 0 | Step 3, 4 |
| `component_count` | Step 1 | Step 2 |
| `complexity_score` | Step 1 | Step 2 |

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
2. 동적 스킬 탐색 (`jikime-adk skill search "{target_framework}"`)
3. `{artifacts_dir}/as_is_spec.md` 기반 계획 수립
4. 스킬 규칙(구조, 네이밍, 라우팅) 적용
5. **사용자 승인 대기**

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

```
각 모듈별 반복:
  1. ANALYZE  - 소스 모듈 동작 이해
  2. PRESERVE - 특성 테스트 작성 (동작 보존)
  3. IMPROVE  - 타겟 프레임워크로 변환
  4. Validate - 빌드 + 테스트 확인
```

**산출물**:
- `{output_dir}/` - 마이그레이션된 프로젝트
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
| `--full` | No | 모든 검증 유형 실행 |
| `--behavior` | No | 동작 비교만 |
| `--e2e` | No | E2E 테스트만 |
| `--performance` | No | 성능 비교만 |
| `--source-url` | No | 소스 시스템 URL (라이브 비교용) |
| `--target-url` | No | 타겟 시스템 URL (라이브 비교용) |

> **Note**: `--source-url`/`--target-url`은 실행 중인 인스턴스 비교용입니다. 소스/타겟 프레임워크 정보는 `.migrate-config.yaml`에서 자동으로 읽습니다.

**검증 항목**:
1. Characterization Tests - 동작 보존 테스트
2. Behavior Comparison - API 응답 비교
3. E2E Tests - 사용자 흐름 검증
4. Performance Check - 성능 비교

**산출물**: `{artifacts_dir}/verification_report.md`

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
     │  (source_path, source_framework, target_framework, artifacts_dir, output_dir)
     │
     ▼
Step 1: config 업데이트 + as_is_spec.md 생성
     │  (component_count, complexity_score, analyzed_at 추가)
     │
     ▼
Step 2: migration_plan.md 생성 (승인 대기)
     │  (동적 스킬 탐색 → 타겟 규칙 적용)
     │
     ▼
Step 3: output_dir/ 생성 + progress.yaml 업데이트
     │  (모듈별 DDD 사이클 반복)
     │
     ▼
Step 4: verification_report.md 생성
     │  (동작 보존 + 성능 검증)
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
```

| target_framework | 탐색되는 스킬 |
|------------------|---------------|
| `nextjs16` | `jikime-migrate-to-nextjs`, `jikime-nextjs@16`, `jikime-library-shadcn` |
| `fastapi` | `jikime-lang-python` (+ 관련 스킬) |
| `go-fiber` | `jikime-lang-go` (+ 관련 스킬) |
| `flutter` | `jikime-lang-flutter` (+ 관련 스킬) |

스킬이 없는 경우 Context7 MCP를 통해 공식 문서를 조회합니다.

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

### 스킬 작성자 가이드

1. **메타데이터 포함** - 프론트매터에 triggers, version 등 명시
2. **최신 문서 참조** - Context7/WebFetch 활용 지침 포함
3. **브레이킹 체인지** - 알려진 이슈와 해결법 문서화
4. **예시 제공** - Before/After 코드 예시 포함
5. **버전 명시** - 호환 버전 범위를 명확히 기재

---

Version: 3.0.0
Last Updated: 2026-01-23
Changelog:
- v3.0.0: Config-First approach; FRIDAY orchestrator; Removed /jikime:migrate; Removed redundant source/target options from Steps 2-4; Renamed --source/--target to --source-url/--target-url in Step 4
- v2.0.0: Added Step-by-Step Workflow, Command Reference with full options
- v1.0.0: Initial migration system documentation
