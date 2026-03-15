# Harness Workflow — Plans.md 기반 워크플로우 스킬 시스템

> Plan → Work → Review → Ship 자동화 루프 — Claude Code의 자유로운 실행을 구조화된 워크플로우로 전환합니다.

---

## 개념

**Harness Workflow**는 5개의 동사형 스킬이 하나의 루프를 형성하는 Claude Code 워크플로우 시스템입니다.

```
harness-plan    → Plans.md 생성 및 태스크 관리
harness-work    → Plans.md의 태스크를 구현 (cc:WIP → cc:DONE)
harness-review  → 구현된 변경사항 4관점 리뷰 (pm:REVIEW → pm:OK)
harness-sync    → Plans.md ↔ git 히스토리 동기화 + 레트로스펙티브
harness-release → 완료된 태스크 배포 자동화
```

**jikime serve와의 차이:**

| 구분 | jikime serve | Harness Workflow |
|------|-------------|-----------------|
| 목적 | GitHub Issue → PR 완전 자동화 | 세션 내 태스크 관리 + 품질 게이트 |
| 진입점 | 데몬 프로세스 (항상 실행) | Claude Code 세션 내 슬래시 커맨드 |
| 태스크 관리 | GitHub Issues | Plans.md (프로젝트 로컬) |
| 품질 관리 | WORKFLOW.md 프롬프트 | 4관점 리뷰 + DoD 검증 |
| 적합한 규모 | 이슈 단위 (단순~중간) | 피처/마일스톤 단위 (중간~복잡) |

---

## Plans.md — 단일 진실 원천 (SSOT)

모든 Harness Workflow는 `Plans.md` 파일을 중심으로 작동합니다.

### 마커 시스템

| 마커 | 의미 | 사용 시점 |
|------|------|-----------|
| `cc:TODO` | 미시작 | 기본값 |
| `cc:WIP` | 진행 중 | Worker가 태스크 시작 시 |
| `cc:DONE [hash]` | 완료 + git hash | Worker가 커밋 완료 후 |
| `pm:REVIEW` | 사용자 검토 요청 | Reviewer가 리뷰 후 |
| `pm:OK` | 검토 완료 | 사용자 승인 후 |
| `blocked:<이유>` | 차단 | 의존성/외부 차단 발생 시 |
| `cc:SKIP` | 건너뜀 | 범위 변경으로 불필요해진 경우 |

### Plans.md 포맷

```markdown
# Plans.md

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | [프로젝트/기능 목표] |
| **마일스톤** | [완료 기준 날짜 또는 이벤트] |
| **담당** | [사용자명 / Claude] |
| **생성일** | [YYYY-MM-DD] |

---

## Phase 1: 핵심 기능

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 1.1  | 태스크 설명 | 완료 기준 (Yes/No 판정 가능) | - | cc:TODO |
| 1.2  | 태스크 설명 | 완료 기준 | 1.1 | cc:TODO |

## Phase 2: 품질 개선

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 2.1  | 태스크 설명 | 완료 기준 | 1.2 | cc:TODO |
```

### DoD 작성 규칙

| ✅ 좋은 DoD | ❌ 나쁜 DoD |
|------------|------------|
| `테스트 통과 (npm test 0 failed)` | `잘 작동함` |
| `lint 에러 0개` | `코드 품질 향상` |
| `API /login 엔드포인트 curl로 200 응답 확인` | `완료됨` |
| `Plans.md의 모든 cc:DONE 태스크에 valid hash 존재` | `ok` |

---

## 구현 현황

### Phase 1: Plans.md 시스템 기초 ✅

> 상태: **완료** (v1.5.x)

**구현된 파일:**

| 파일 | 설명 |
|------|------|
| `templates/Plans.md` | 표준 Plans.md 템플릿 |
| `templates/.claude/skills/jikime-harness-plan/SKILL.md` | harness-plan 스킬 |
| `templates/.claude/skills/jikime-harness-sync/SKILL.md` | harness-sync 스킬 |
| `cmd/hookscmd/plans_watcher.go` | Plans.md 구조 검증 hook |
| `templates/.claude/settings.json` | `plans-watcher` hook 등록 |

#### harness-plan 스킬

Plans.md를 생성하고 관리합니다.

**서브커맨드:**

```
/jikime:harness-plan create   → 새 Plans.md 생성 (대화형)
/jikime:harness-plan add      → 기존 Plans.md에 태스크/Phase 추가
/jikime:harness-plan update 1.2 cc:WIP   → 마커 변경
/jikime:harness-plan sync     → git 히스토리 기반 상태 추론 (경량)
```

**create 실행 흐름:**
1. 기술 스택 자동 감지 (package.json, go.mod 등)
2. 목표/기능 수집 (최대 3개 질문)
3. Phase 분류: Required(1) / Recommended(2) / Optional(3)
4. 각 태스크의 DoD 추론
5. 의존성 매핑
6. Plans.md 생성 (모든 태스크 `cc:TODO`)

#### harness-sync 스킬

Plans.md와 git 히스토리를 상세 동기화합니다.

**서브커맨드:**

```
/jikime:harness-sync sync    → 전체 동기화 (불일치 탐지 + 업데이트 제안)
/jikime:harness-sync retro   → 4항목 레트로스펙티브 생성
/jikime:harness-sync trace   → 특정 태스크 Agent Trace 분석
/jikime:harness-sync drift   → Plans.md vs 실제 구현 드리프트 감지
```

**레트로스펙티브 4항목:**

| 항목 | 내용 |
|------|------|
| 견적 정확도 | 계획 대비 완료율, 스코프 크리프 측정 |
| 블로킹 원인 | blocked 마커 분석, 평균 차단 시간 |
| 품질 마커 적중률 | cc:DONE → pm:OK 전환율, 재검토 비율 |
| 스코프 변동 | Plans.md 변경 이력, 태스크 추가/삭제 패턴 |

#### plans-watcher hook

Plans.md가 Write/Edit될 때마다 자동으로 실행됩니다.

```
PostToolUse (Write|Edit) → plans-watcher → 구조 검증
```

**검증 항목:**
- 마커 문법 검증 (`cc:TODO`, `cc:WIP`, `cc:DONE [hash]` 등)
- 모호한 DoD 감지 (Yes/No로 판정 불가능한 기준 경고)
- 의존성 참조 검증 (존재하지 않는 Task ID 참조 감지)
- 태스크 상태 요약 출력

**출력 예시:**
```
📋 Plans.md: 8 tasks [TODO:3 | WIP:1 | DONE:4]

⚠️ Plans.md validation warnings:
  - Line 12: Task 2.1 has vague DoD: `완료됨` — DoD must be Yes/No judgeable
```

---

### Phase 2: harness-work + 3 Agent Team ✅

> 상태: **완료** (v1.5.x)

**구현된 파일:**

| 파일 | 설명 |
|------|------|
| `templates/.claude/skills/jikime-harness-work/SKILL.md` | harness-work 스킬 |
| `templates/.claude/skills/jikime-harness-review/SKILL.md` | harness-review 스킬 |
| `templates/.claude/agents/jikime/harness-worker.md` | Worker Agent |
| `templates/.claude/agents/jikime/harness-reviewer.md` | Reviewer Agent |
| `templates/.claude/agents/jikime/harness-scaffolder.md` | Scaffolder Agent |

#### harness-work 스킬

Plans.md 태스크를 `cc:WIP → cc:DONE` 라이프사이클로 실행합니다.

```
/jikime:harness-work 1.2              → 태스크 1.2 단독 실행 (Solo)
/jikime:harness-work 1.2 1.3         → 병렬 실행 (Parallel)
/jikime:harness-work --auto           → TODO 태스크 자동 선택
/jikime:harness-work --mode breezing  → 4+ 태스크 Breezing Mode
```

**실행 흐름:**
```
Phase 0: Plans.md 검증 (의존성 완료 확인)
Phase 1: Scaffolder → Mode 선택 (Solo/Parallel/Breezing)
Phase 2: Scaffolder → 태스크 분석 + 스캐폴딩
Phase 3: Worker → 구현 + DoD 검증 (cc:WIP → cc:DONE)
Phase 4: Reviewer → 4관점 리뷰
Phase 5: 사용자 알림 (pm:REVIEW 요청)
```

**Breezing Mode:**

| 조건 | Mode | 설명 |
|------|------|------|
| 태스크 1개 | Solo | Worker 1명, 순차 실행 |
| 태스크 2–3개, 독립적 | Parallel | Worker 2–3명, 병렬 worktree |
| 태스크 4개 이상 | Breezing | Lead + Workers + Reviewer 팀 |

#### harness-review 스킬

`cc:DONE` 태스크에 4관점 코드 리뷰를 실행하고 `pm:REVIEW → pm:OK`를 관리합니다.

```
/jikime:harness-review 1.2           → 태스크 1.2 리뷰
/jikime:harness-review --all         → 모든 pm:REVIEW 태스크 일괄 리뷰
/jikime:harness-review --approve 1.2 → 직접 pm:OK 승인
```

#### 3 Agent Team

| Agent | 역할 | 권한 |
|-------|------|------|
| **harness-worker** | 구현 전담, DoD 검증, Plans.md 마커 업데이트 | Read/Write/Edit/Bash |
| **harness-reviewer** | 4관점 리뷰 (보안/성능/품질/DoD), 절대 코드 수정 안 함 | Read/Grep/Glob/Bash |
| **harness-scaffolder** | 태스크 분석, 스캐폴딩, Mode 선택, 상태 전환 관리 | Read/Write/Edit/Bash |

---

### Phase 3: 선언적 가드레일 엔진 ✅

> 상태: **완료** (v1.5.x)

**구현된 파일:**

| 파일 | 설명 |
|------|------|
| `cmd/hookscmd/guardrail_engine.go` | R01-R08 룰 평가 엔진 (PostToolUse hook) |
| `templates/.claude/settings.json` | `guardrail-engine` hook 등록 |

**Plans.md가 Write/Edit될 때마다 자동 실행됩니다.**

#### 규칙 목록

| 규칙 | 수준 | 내용 | 감지 방법 |
|------|------|------|----------|
| **R02** | error | cc:WIP 태스크 동시 2개 초과 | WIP 마커 카운트 |
| **R03** | error | 의존성 미완료 태스크를 WIP로 설정 | 의존성 그래프 검증 |
| **R05** | error | pm:REVIEW 없이 pm:OK 설정 | git log -S 검색 |
| **R06** | warn | blocked 태스크가 5일 이상 지속 | git log 타임스탬프 분석 |
| **R07** | warn | 최초 대비 태스크 수 30% 초과 증가 | git 첫 커밋 대비 현재 비교 |
| **R08** | warn | 이전 Phase 미완료 상태에서 이후 Phase 완료 | Phase 순서 역전 감지 |

> R01(Plans.md 필수), R04(DoD 검증), R09(Breezing Mode)는 각각 plans-watcher hook, Worker Agent, Scaffolder Agent에서 처리합니다.

#### 출력 예시

```
🚨 Harness Guardrail 위반 (2건):

[R02] WIP 동시성 초과: 3개 태스크가 cc:WIP 상태입니다 (1.1, 1.2, 1.3). 최대 2개 권장.
  → 완료되지 않은 태스크를 먼저 처리하거나 blocked 처리하세요.

[R03] 의존성 미완료: Task 1.2가 cc:WIP이지만, 의존 태스크 1.1는 cc:WIP 상태입니다.
  → 1.1를 먼저 완료하세요.

⚠️  Harness Guardrail 경고 (1건):

[R08] Phase 순서 역전: Phase 2 태스크가 완료됐지만 Phase 1의 Task 1.1는 cc:WIP 상태입니다.
  → Phase 1 태스크를 먼저 완료하는 것을 권장합니다.
```

---

### Phase 4: harness-setup + harness-release ✅

> 상태: **완료** (v1.5.x)

**구현된 파일:**

| 파일 | 설명 |
|------|------|
| `templates/.claude/skills/jikime-harness-setup/SKILL.md` | harness-setup 스킬 |
| `templates/.claude/skills/jikime-harness-release/SKILL.md` | harness-release 스킬 |

#### harness-setup 스킬

새 프로젝트 또는 기존 프로젝트에 Harness Engineering 워크플로우를 설정합니다.

```
/jikime:harness-setup           → 전체 설정 (대화형)
/jikime:harness-setup --check   → 현재 설정 상태만 진단
/jikime:harness-setup --reset   → 설정 초기화 후 재설정
```

**5단계 실행 흐름:**
1. 환경 진단 (git 저장소, jikime 바이너리, settings.json, Plans.md, hooks)
2. Plans.md 생성 안내 (미존재 시)
3. Hook 등록 검증 및 자동 추가 (plans-watcher, guardrail-engine)
4. Git 환경 준비 (worktree 지원, .gitignore 업데이트)
5. 설정 완료 보고

#### harness-release 스킬

Plans.md의 `pm:OK` 태스크를 기반으로 릴리스를 자동화합니다.

```
/jikime:harness-release                → 대화형 릴리스 (버전 자동 제안)
/jikime:harness-release --patch        → 패치 버전 (v1.0.0 → v1.0.1)
/jikime:harness-release --minor        → 마이너 버전 (v1.0.0 → v1.1.0)
/jikime:harness-release --major        → 메이저 버전 (v1.0.0 → v2.0.0)
/jikime:harness-release --dry-run      → 릴리스 미리보기 (변경사항 없음)
/jikime:harness-release --version v2.1.0 → 직접 버전 지정
```

**7단계 실행 흐름:**
1. pm:OK 태스크 수집 (전제 조건 검증: cc:WIP 없음, pm:REVIEW 없음, working directory 클린)
2. 버전 결정 (package.json, go.mod, VERSION 파일, git 태그 자동 감지 + 버전 범프 제안)
3. CHANGELOG.md 생성/업데이트 (커밋 프리픽스 → 섹션 매핑: feat→✨, fix→🐛 등)
4. 버전 파일 업데이트
5. Plans.md 릴리스 기록 추가
6. git 커밋 + 태그 생성 (사용자 확인 후 push)
7. 릴리스 완료 보고

---

## 사용 방법

### 기본 흐름

```
1. 프로젝트 설정    → /jikime:harness-setup
2. Plans.md 생성   → /jikime:harness-plan create
3. 태스크 구현     → /jikime:harness-work
4. 코드 리뷰       → /jikime:harness-review
5. 동기화          → /jikime:harness-sync
6. 릴리스          → /jikime:harness-release
```

### Step 1: 초기 설정 (최초 1회)

새 프로젝트에서 처음 시작할 때 실행합니다.

```
/jikime:harness-setup
```

git 저장소 확인, hook 등록 여부 검증, 문제가 있으면 자동 수정 제안까지 처리합니다.
현재 상태만 확인하고 싶다면:

```
/jikime:harness-setup --check
```

### Step 2: Plans.md 생성

구현할 기능/목표를 설명하면 Claude가 태스크로 분해해 줍니다.

```
/jikime:harness-plan create "JWT 인증 시스템 구현"
```

Claude가 최대 3가지를 물어본 뒤 Plans.md를 생성합니다:

```markdown
## Phase 1: 핵심 기능
| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 1.1  | JWT 토큰 생성 | jwt.sign() 동작, 만료 테스트 통과 | - | cc:TODO |
| 1.2  | 로그인 API | POST /auth/login 200 반환 | 1.1 | cc:TODO |
| 1.3  | 미들웨어 검증 | 유효/만료/위조 토큰 구분 | 1.2 | cc:TODO |
```

기존 Plans.md에 태스크를 추가하려면:

```
/jikime:harness-plan add "Phase 2: 소셜 로그인"
```

### Step 3: 태스크 구현

Plans.md가 준비됐으면 구현을 시작합니다.

```bash
# 단일 태스크
/jikime:harness-work 1.1

# 병렬 실행 (독립적인 태스크 2~3개)
/jikime:harness-work 1.1 1.3

# TODO 태스크 자동 선택
/jikime:harness-work --auto
```

내부 동작:
1. Scaffolder가 태스크 분석 + 실행 모드 결정 (Solo/Parallel/Breezing)
2. Worker가 격리된 worktree에서 구현
3. Plans.md 마커 자동 변경: `cc:TODO → cc:WIP → cc:DONE [abc1234]`
4. Reviewer가 4관점(보안/성능/품질/DoD) 자동 리뷰
5. 완료 시 `pm:REVIEW` 상태로 전환 + 사용자 알림

### Step 4: 리뷰 승인

Worker가 완료하면 Plans.md에 `pm:REVIEW` 상태가 됩니다.

```bash
# 리뷰 내용 보기
/jikime:harness-review 1.1

# 승인 (pm:OK로 변경)
/jikime:harness-review --approve 1.1

# 모든 REVIEW 태스크 일괄 처리
/jikime:harness-review --all
```

### Step 5: 동기화 & 레트로스펙티브

Plans.md와 실제 git 히스토리가 맞는지 확인합니다.

```bash
/jikime:harness-sync sync    # 불일치 감지 + 수정 제안
/jikime:harness-sync retro   # 4항목 레트로스펙티브 생성
/jikime:harness-sync drift   # Plans.md vs 실제 구현 차이 감지
```

레트로스펙티브 출력 예시:

```
📊 레트로스펙티브 v1.2.0

견적 정확도: 계획 8태스크 → 완료 8태스크 (100%)
블로킹 분석: blocked 0건 발생
품질 적중률: cc:DONE → pm:OK 전환 100%
스코프 변동: 태스크 추가 없음 (안정적)
```

### Step 6: 릴리스

모든 태스크가 `pm:OK` 상태가 되면 릴리스합니다.

```bash
# 대화형 릴리스 (버전 자동 제안)
/jikime:harness-release

# 미리보기만 (변경 없음)
/jikime:harness-release --dry-run

# 버전 직접 지정
/jikime:harness-release --minor   # v1.0.0 → v1.1.0
/jikime:harness-release --patch   # v1.0.0 → v1.0.1
```

Claude가 자동으로 CHANGELOG.md 생성, 버전 파일 업데이트, `git commit + tag` 생성 후 push 여부를 확인합니다.

---

### 실전 흐름 예시

```bash
# 1. 초기 설정
/jikime:harness-setup

# 2. 기획
/jikime:harness-plan create "사용자 인증 시스템"

# 3. 구현 (Plans.md 보면서 진행)
/jikime:harness-work 1.1
/jikime:harness-review --approve 1.1

/jikime:harness-work 1.2 1.3     # 병렬 실행
/jikime:harness-review --all

# 4. 마무리
/jikime:harness-sync retro
/jikime:harness-release --minor
```

### 상황별 빠른 참조

| 상황 | 명령어 |
|------|--------|
| 남은 태스크 확인 | Plans.md 열기 또는 `/jikime:harness-plan update` |
| 구현이 막혔을 때 | 해당 태스크를 `blocked:이유`로 수동 변경 |
| 태스크 추가 | `/jikime:harness-plan add` |
| git과 Plans.md가 불일치 | `/jikime:harness-sync sync` |
| 릴리스 전 내용 확인 | `/jikime:harness-release --dry-run` |

> **핵심**: Plans.md가 항상 단일 진실 원천(SSOT)입니다. Claude가 마커를 자동으로 업데이트하며 진행 상황을 추적합니다.

---

## 전체 워크플로우

```
사용자: /jikime:harness-plan create "JWT 인증 시스템 구현"
        ↓
Claude: Plans.md 생성 (Phase 1~3, 12개 태스크, 모두 cc:TODO)
        ↓
사용자: /jikime:harness-work 1.1
        ↓
Worker Agent: worktree 격리, 구현, 테스트, 커밋
              Plans.md: 1.1 → cc:WIP → cc:DONE [abc1234]
        ↓
Reviewer Agent: 4관점 리뷰 (보안/성능/품질/문서)
                Plans.md: 1.1 → pm:REVIEW
        ↓
사용자: 리뷰 확인 → 승인
        Plans.md: 1.1 → pm:OK
        ↓
(다음 태스크 반복...)
        ↓
사용자: /jikime:harness-sync retro
        ↓
Claude: 4항목 레트로스펙티브 생성
        Plans.md 하단에 추가
        ↓
사용자: /jikime:harness-release
        ↓
Claude: CHANGELOG 업데이트, 태그 생성, 릴리스 노트 작성
```

---

## 슬래시 커맨드 연결

| 커맨드 | 스킬 | 단계 |
|--------|------|------|
| `/jikime:1-plan` | harness-plan create | Plan |
| `/jikime:2-run` | harness-work | Work |
| `/jikime:3-sync` | harness-sync | Sync |
| `/jikime:harness-plan` | harness-plan | Plan 관리 |

---

## 관련 문서

- [하네스 엔지니어링 (jikime serve)](./harness-engineering.md) — GitHub Issue 자동화 데몬
- [훅 시스템](./hooks.md)
- [스킬 카탈로그](./skills-catalog.md)
- [POC-First 워크플로우](./poc-first.md)

---

Last Updated: 2026-03-15
Version: Phase 4 완료
