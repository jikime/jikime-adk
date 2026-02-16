# Agent Teams - 병렬 팀 기반 개발

**Claude Code Agent Teams를 활용한 병렬 멀티 에이전트 오케스트레이션**

> Agent Teams는 Claude Code v2.1.32+의 실험적 기능으로, 복잡한 멀티 도메인 작업을 팀 기반으로 병렬 처리합니다.

---

## 개요

Agent Teams는 J.A.R.V.I.S./F.R.I.D.A.Y. 오케스트레이터가 여러 전문 에이전트를 팀으로 구성하여 병렬로 작업을 수행하는 기능입니다.

### 기존 Sub-Agent 방식 vs Agent Teams

| 구분 | Sub-Agent 방식 | Agent Teams 방식 |
|------|---------------|------------------|
| **실행** | 순차 실행 | 병렬 실행 |
| **통신** | Task() 호출/반환 | SendMessage로 실시간 협업 |
| **작업 관리** | 오케스트레이터가 직접 관리 | 공유 TaskList로 자율 분배 |
| **상태** | Stateless | 팀 세션 동안 상태 유지 |
| **적합한 경우** | 단일 도메인, 간단한 작업 | 멀티 도메인, 복잡한 작업 |

---

## 활성화 조건

### 필수 요구사항

1. **Claude Code 버전**: v2.1.32 이상
2. **환경 변수**: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
3. **설정 파일**: `.jikime/config/workflow.yaml`에서 `team.enabled: true`

```bash
# 환경 변수 설정
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1

# 또는 settings.json에 추가
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

### 자동 활성화 조건

다음 조건 중 하나라도 충족되면 자동으로 팀 모드가 활성화됩니다:

| 조건 | 기준값 | 설명 |
|------|--------|------|
| **도메인 수** | >= 3 | frontend, backend, database 등 |
| **파일 수** | >= 10 | 변경될 파일 개수 |
| **복잡도 점수** | >= 7 | 1-10 스케일 |

---

## 빠른 시작 가이드

### Step 1: 환경 설정

```bash
# 1. Claude Code 버전 확인 (v2.1.32 이상 필요)
claude --version

# 2. 환경 변수 설정 (터미널에서)
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1

# 3. 또는 ~/.claude/settings.json에 영구 설정
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

### Step 2: 프로젝트 설정 확인

```bash
# JikiME-ADK 초기화 (이미 되어있다면 생략)
jikime-adk init

# workflow.yaml에서 team 설정 확인
cat .jikime/config/workflow.yaml | grep -A 5 "team:"
```

### Step 3: 팀 모드 실행

```bash
# Plan Phase에서 팀 모드 사용
/jikime:1-plan "사용자 인증 시스템 구현" --team

# Run Phase에서 팀 모드 사용
/jikime:2-run SPEC-AUTH-001 --team
```

---

## 상세 사용 가이드

### 시나리오 1: 새로운 기능 개발 (Full Stack)

복잡한 기능을 백엔드, 프론트엔드, 테스트를 병렬로 개발하는 경우:

```bash
# 1. Plan Phase - 팀으로 요구사항 분석
/jikime:1-plan "OAuth 2.0 기반 소셜 로그인 구현 (Google, GitHub, Kakao)" --team

# 내부적으로 발생하는 일:
# - team-researcher: 기존 인증 코드 분석
# - team-analyst: 요구사항 및 엣지 케이스 도출
# - team-architect: 기술 설계 및 아키텍처 결정
# → 결과: SPEC-SOCIAL-LOGIN-001.md 생성

# 2. SPEC 승인 후 Run Phase - 팀으로 구현
/jikime:2-run SPEC-SOCIAL-LOGIN-001 --team

# 내부적으로 발생하는 일:
# - team-backend-dev: OAuth 프로바이더 연동, API 엔드포인트 구현
# - team-frontend-dev: 로그인 버튼, 콜백 페이지 구현
# - team-tester: E2E 테스트, 단위 테스트 작성
# - team-quality: TRUST 5 품질 검증

# 3. 완료 후 동기화
/jikime:3-sync SPEC-SOCIAL-LOGIN-001
```

### 시나리오 2: UI/UX 중심 기능 개발

디자이너가 포함된 팀으로 UI 중심 기능을 개발하는 경우:

```bash
# design_implementation 패턴 사용
/jikime:2-run SPEC-DASHBOARD-001 --team --pattern=design_implementation

# 팀 구성:
# - team-designer: 디자인 토큰, 컴포넌트 스타일 정의
# - team-backend-dev: 대시보드 데이터 API
# - team-frontend-dev: React 컴포넌트 구현
# - team-tester: 비주얼 리그레션 테스트
```

### 시나리오 3: 복잡한 버그 조사

여러 가설을 병렬로 조사하는 경우:

```bash
# investigation 패턴 사용
/jikime:build-fix --team

# 팀 구성 (경쟁적 가설 조사):
# - hypothesis-1: 메모리 누수 가설 조사
# - hypothesis-2: 레이스 컨디션 가설 조사
# - hypothesis-3: API 타임아웃 가설 조사
# → 가장 먼저 원인을 찾은 가설이 채택됨
```

### 시나리오 4: 프로덕션 배포 전 품질 검증

품질 게이트가 중요한 경우:

```bash
# quality_gate 패턴 사용
/jikime:2-run SPEC-PAYMENT-001 --team --pattern=quality_gate

# 팀 구성:
# - team-backend-dev: 결제 로직 구현
# - team-frontend-dev: 결제 UI 구현
# - team-tester: 보안 테스트, 결제 플로우 테스트
# - team-quality: TRUST 5 + 보안 감사 + 성능 검증
```

### 팀 모드 vs Solo 모드 선택 가이드

| 상황 | 권장 모드 | 명령어 |
|------|----------|--------|
| 간단한 버그 수정 | Solo | `/jikime:2-run SPEC-001 --solo` |
| 단일 파일 수정 | Solo | `/jikime:2-run SPEC-001 --solo` |
| 멀티 도메인 기능 | Team | `/jikime:2-run SPEC-001 --team` |
| 10개 이상 파일 변경 | Team | 자동 감지 |
| UI + API + DB 동시 변경 | Team | `/jikime:2-run SPEC-001 --team` |
| 복잡한 리팩토링 | Team | `/jikime:2-run SPEC-001 --team` |

---

## 팀원 간 협업 패턴

### SendMessage 활용 예시

```javascript
// Backend → Frontend: API 준비 완료 알림
SendMessage(
  recipient: "frontend-dev",
  type: "api_ready",
  content: {
    endpoint: "POST /api/auth/oauth/callback",
    schema: {
      provider: "string",
      code: "string",
      state: "string"
    },
    response: {
      user: "User",
      accessToken: "string",
      refreshToken: "string"
    }
  }
)

// Frontend → Tester: 컴포넌트 준비 완료 알림
SendMessage(
  recipient: "tester",
  type: "component_ready",
  content: {
    component: "SocialLoginButton",
    path: "src/components/auth/SocialLoginButton.tsx",
    testable: true,
    props: ["provider", "onSuccess", "onError"]
  }
)

// Tester → Quality: 테스트 완료 알림
SendMessage(
  recipient: "quality",
  type: "tests_passed",
  content: {
    coverage: 87,
    passedTests: 45,
    failedTests: 0,
    duration: "2m 34s"
  }
)

// Quality → Team Lead: 품질 게이트 통과
SendMessage(
  recipient: "team-lead",
  type: "quality_gate_passed",
  content: {
    trust5Score: 4.2,
    securityIssues: 0,
    performanceScore: "A",
    recommendation: "Ready for production"
  }
)
```

### 파일 변경 요청 패턴

```javascript
// Frontend에서 Backend 타입 변경 요청
SendMessage(
  recipient: "backend-dev",
  type: "change_request",
  content: {
    file: "src/types/auth.ts",
    requestedChange: "Add 'profileImage' field to User type",
    reason: "Required for social login avatar display",
    priority: "high"
  }
)

// Backend 응답
SendMessage(
  recipient: "frontend-dev",
  type: "change_completed",
  content: {
    file: "src/types/auth.ts",
    change: "Added profileImage: string | null to User type",
    commitRef: "abc123"
  }
)
```

---

## 트러블슈팅

### 문제 1: 팀 모드가 활성화되지 않음

**증상**: `--team` 플래그를 사용해도 순차 실행됨

**해결 방법**:
```bash
# 1. Claude Code 버전 확인
claude --version  # v2.1.32 이상이어야 함

# 2. 환경 변수 확인
echo $CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS  # "1"이어야 함

# 3. 설정 파일 확인
cat .jikime/config/workflow.yaml | grep "team:" -A 3
# enabled: true 확인

# 4. 강제 활성화
/jikime:2-run SPEC-001 --team --force
```

### 문제 2: 팀원 간 통신 실패

**증상**: SendMessage가 전달되지 않음

**해결 방법**:
```bash
# 1. 팀 세션 상태 확인
# 팀 이름이 올바른지 확인

# 2. 수신자 이름 확인 (정확한 이름 사용)
# ✅ "backend-dev"
# ❌ "team-backend-dev"
# ❌ "backendDev"

# 3. 팀 재시작
/jikime:team-restart
```

### 문제 3: 파일 충돌 발생

**증상**: 두 팀원이 같은 파일을 수정하려 함

**해결 방법**:
```javascript
// 1. 직접 수정하지 말고 소유자에게 요청
SendMessage(
  recipient: "backend-dev",  // 파일 소유자
  type: "change_request",
  content: {
    file: "src/types/user.ts",
    requestedChange: "Add socialProvider field",
    reason: "Needed for OAuth integration"
  }
)

// 2. 공유 파일은 조율 후 수정
// src/types/**, src/utils/** 등은 SendMessage로 조율 필요
```

### 문제 4: 팀원이 유휴 상태로 대기

**증상**: 팀원이 작업 없이 대기 중

**해결 방법**:
```javascript
// TaskList에서 할당되지 않은 작업 확인
TaskList()

// 유휴 팀원에게 작업 할당
TaskCreate(
  subject: "Implement error handling for OAuth callback",
  description: "Add try-catch and error UI for failed OAuth",
  owner: "frontend-dev"  // 유휴 팀원 지정
)
```

### 문제 5: 품질 게이트 실패

**증상**: team-quality가 작업을 거부함

**해결 방법**:
```bash
# 1. 품질 리포트 확인
# team-quality가 보낸 메시지에서 실패 원인 확인

# 2. 일반적인 실패 원인:
# - 테스트 커버리지 미달 (80% 이상 필요)
# - 린트 에러 존재
# - 보안 취약점 발견
# - 성능 기준 미달

# 3. 해당 팀원에게 수정 요청
SendMessage(
  recipient: "tester",
  type: "coverage_required",
  content: {
    current: 72,
    required: 80,
    uncoveredFiles: ["src/services/oauth.ts"]
  }
)
```

---

## Best Practices

### 1. 팀 구성 최적화

```yaml
# 프로젝트 규모별 권장 팀 구성
small_project:  # < 5개 파일
  mode: solo

medium_project:  # 5-20개 파일
  mode: team
  pattern: implementation
  roles: [backend-dev, frontend-dev, tester]

large_project:  # > 20개 파일
  mode: team
  pattern: quality_gate
  roles: [backend-dev, frontend-dev, designer, tester, quality]
```

### 2. 효과적인 파일 소유권 설정

```yaml
# 명확한 경계 설정
file_ownership:
  team-backend-dev:
    - "src/api/**"
    - "src/services/**"
    - "src/repositories/**"
    - "prisma/**"

  team-frontend-dev:
    - "src/components/**"
    - "src/pages/**"
    - "src/hooks/**"
    - "src/stores/**"

  # 공유 파일은 최소화
  shared:
    - "src/types/**"  # 타입 정의만 공유
```

### 3. 통신 최소화

```javascript
// ❌ 나쁜 예: 너무 잦은 통신
SendMessage(recipient: "frontend-dev", type: "progress", content: "Started task")
SendMessage(recipient: "frontend-dev", type: "progress", content: "50% done")
SendMessage(recipient: "frontend-dev", type: "progress", content: "Almost done")

// ✅ 좋은 예: 의미 있는 통신만
SendMessage(
  recipient: "frontend-dev",
  type: "api_ready",
  content: { endpoint: "/api/users", schema: {...} }
)
```

### 4. 품질 게이트 활용

```yaml
# 중요한 기능에는 항상 quality 팀원 포함
critical_features:
  - payment
  - authentication
  - user_data

pattern_for_critical: quality_gate
```

### 5. Fallback 대비

```yaml
# Sub-Agent 폴백 설정
workflow:
  team:
    fallback:
      enabled: true
      log_level: "warn"
      preserve_progress: true  # 진행 상황 보존
```

---

## Team 에이전트 목록

### Plan Phase 에이전트 (읽기 전용)

| 에이전트 | 모델 | 역할 | 스킬 |
|----------|------|------|------|
| **team-researcher** | haiku | 코드베이스 탐색, 아키텍처 분석 | jikime-foundation-philosopher |
| **team-analyst** | inherit | 요구사항 분석, 엣지 케이스 도출 | jikime-workflow-spec |
| **team-architect** | inherit | 기술 설계, 대안 평가 | jikime-domain-architecture |

### Run Phase 에이전트 (구현 권한)

| 에이전트 | 모델 | 역할 | 파일 소유권 |
|----------|------|------|-------------|
| **team-backend-dev** | inherit | API, 서비스, 비즈니스 로직 구현 | `src/api/**`, `src/services/**` |
| **team-frontend-dev** | inherit | UI 컴포넌트, 페이지 구현 | `src/components/**`, `src/pages/**` |
| **team-designer** | inherit | UI/UX 설계, 디자인 토큰 | `design/**`, `src/styles/tokens/**` |
| **team-tester** | inherit | 테스트 작성, 커버리지 검증 | `tests/**`, `**/*.test.*` |
| **team-quality** | inherit (읽기 전용) | TRUST 5 검증, 품질 게이트 | - |

---

## 팀 구성 패턴

### 1. plan_research (Plan Phase)

SPEC 문서 생성을 위한 병렬 리서치 팀

```yaml
roles:
  - researcher  # 코드베이스 탐색
  - analyst     # 요구사항 분석
  - architect   # 기술 설계
```

**사용 시점**: `/jikime:1-plan --team` 또는 복잡도 자동 감지

### 2. implementation (Run Phase)

기능 구현을 위한 개발 팀

```yaml
roles:
  - backend-dev   # 서버 사이드
  - frontend-dev  # 클라이언트 사이드
  - tester        # 테스트
```

**사용 시점**: `/jikime:2-run SPEC-001 --team`

### 3. design_implementation

UI/UX가 중요한 기능 구현

```yaml
roles:
  - designer      # UI/UX 설계
  - backend-dev
  - frontend-dev
  - tester
```

### 4. quality_gate

품질 검증이 중요한 프로덕션 배포

```yaml
roles:
  - backend-dev
  - frontend-dev
  - tester
  - quality       # TRUST 5 검증
```

### 5. investigation

복잡한 버그의 경쟁적 가설 조사

```yaml
roles:
  - hypothesis-1
  - hypothesis-2
  - hypothesis-3
model: haiku  # 빠르고 저렴한 모델
```

**사용 시점**: `/jikime:build-fix --team`

---

## 워크플로우

### Team Plan Workflow

```
┌─────────────────────────────────────────────────────────┐
│  Phase 0: TeamCreate("jikime-plan-{feature}")           │
│  ↓                                                       │
│  Phase 1: 병렬 Spawn                                     │
│  ├─ Task(team-researcher) ──┐                           │
│  ├─ Task(team-analyst) ─────┼─→ 병렬 실행               │
│  └─ Task(team-architect) ───┘                           │
│  ↓                                                       │
│  Phase 2: 모니터링 & 조정                                │
│  (SendMessage로 실시간 협업)                             │
│  ↓                                                       │
│  Phase 3: 결과 통합 → SPEC 문서 생성                    │
│  ↓                                                       │
│  Phase 4: 사용자 승인 (AskUserQuestion)                 │
│  ↓                                                       │
│  Phase 5: TeamDelete + /clear                           │
└─────────────────────────────────────────────────────────┘
```

### Team Run Workflow

```
┌─────────────────────────────────────────────────────────┐
│  Phase 0: TeamCreate + Task 분해 (파일 소유권 할당)     │
│  ↓                                                       │
│  Phase 1: 병렬 구현 팀 Spawn                            │
│  ├─ backend-dev (src/api/**)                            │
│  ├─ frontend-dev (src/components/**)                    │
│  ├─ tester (tests/**)                                   │
│  └─ quality (읽기 전용)                                 │
│  ↓                                                       │
│  Phase 2: 병렬 구현                                      │
│  - SendMessage("api_ready") → frontend 작업 시작        │
│  - SendMessage("component_ready") → tester 작업 시작    │
│  ↓                                                       │
│  Phase 3: 품질 게이트                                    │
│  - team-quality가 TRUST 5 검증                          │
│  - 통과 시 완료, 실패 시 수정 요청                      │
│  ↓                                                       │
│  Phase 4: TeamDelete                                    │
└─────────────────────────────────────────────────────────┘
```

---

## Team API 레퍼런스

### TeamCreate

팀 세션을 초기화합니다.

```javascript
TeamCreate(team_name: "jikime-plan-auth-feature")
```

### Task (팀 모드)

팀원을 생성합니다. 반드시 `team_name`과 `name` 파라미터가 필요합니다.

```javascript
Task(
  subagent_type: "team-backend-dev",
  team_name: "jikime-run-spec-001",
  name: "backend-dev",
  prompt: "..."
)
```

### SendMessage

팀원 간 또는 팀 리드에게 메시지를 보냅니다.

```javascript
// API 준비 알림
SendMessage(
  recipient: "frontend-dev",
  type: "api_ready",
  content: {
    endpoint: "POST /api/auth/login",
    schema: { email: "string", password: "string" }
  }
)

// 버그 리포트
SendMessage(
  recipient: "backend-dev",
  type: "bug_report",
  content: {
    test: "auth.test.ts:45",
    expected: "200 OK",
    actual: "500 Error"
  }
)

// 셧다운 요청
SendMessage(
  type: "shutdown_request",
  recipient: "researcher",
  content: "Plan phase complete"
)
```

### TaskCreate/Update/List/Get

공유 작업 목록을 관리합니다.

```javascript
// 작업 생성
TaskCreate(
  subject: "Implement login API",
  description: "...",
  owner: "backend-dev"
)

// 상태 업데이트
TaskUpdate(taskId: "1", status: "in_progress")
TaskUpdate(taskId: "1", status: "completed")

// 작업 목록 조회
TaskList()  // 모든 팀원이 볼 수 있음

// 상세 조회
TaskGet(taskId: "1")
```

### TeamDelete

팀 세션을 종료합니다. **모든 팀원이 셧다운된 후 호출해야 합니다.**

```javascript
// 먼저 모든 팀원에게 셧다운 요청
SendMessage(type: "shutdown_request", recipient: "all", ...)

// 팀원 종료 확인 후
TeamDelete(team_name: "jikime-run-spec-001")
```

---

## 파일 소유권

팀 모드에서 파일 충돌을 방지하기 위해 각 팀원은 특정 파일 패턴을 소유합니다.

```yaml
file_ownership:
  team-backend-dev:
    - "src/api/**"
    - "src/services/**"
    - "src/repositories/**"
    - "src/models/**"
    - "src/middleware/**"
    - "prisma/migrations/**"

  team-frontend-dev:
    - "src/components/**"
    - "src/pages/**"
    - "src/app/**"
    - "src/hooks/**"
    - "src/stores/**"
    - "src/styles/**"

  team-designer:
    - "design/**"
    - "src/styles/tokens/**"

  team-tester:
    - "tests/**"
    - "__tests__/**"
    - "**/*.test.*"
    - "**/*.spec.*"
    - "cypress/**"
    - "playwright/**"

  shared:  # SendMessage로 조정 필요
    - "src/types/**"
    - "src/utils/**"
    - "src/lib/**"
```

### 충돌 해결

자신이 소유하지 않은 파일을 수정해야 할 경우:

```javascript
// ❌ 직접 수정하지 않음

// ✅ 파일 소유자에게 요청
SendMessage(
  recipient: "backend-dev",
  type: "change_request",
  content: {
    file: "src/types/user.ts",
    requested_change: "Add 'refreshToken' field",
    reason: "Needed for token refresh flow"
  }
)
```

---

## Hook 이벤트

### TeammateIdle

팀원이 작업을 완료하고 유휴 상태가 될 때 호출됩니다.

```yaml
TeammateIdle:
  exit_code_0: "유휴 상태 수락 - 더 이상 작업 없음"
  exit_code_2: "유휴 거부 - TaskList에서 추가 작업 할당"
```

### TaskCompleted

팀원이 작업을 완료로 표시할 때 호출됩니다.

```yaml
TaskCompleted:
  exit_code_0: "완료 수락 - 작업 종료"
  exit_code_2: "완료 거부 - 추가 작업 필요"
  validation:
    - 테스트 통과
    - 커버리지 목표 달성
    - 린트 에러 없음
```

---

## 사용 예시

### Plan Phase에서 팀 모드 사용

```bash
# 명시적 팀 모드
/jikime:1-plan "사용자 인증 시스템 구현" --team

# 자동 감지 (복잡도가 높으면 자동 활성화)
/jikime:1-plan "복잡한 멀티 도메인 기능"
```

### Run Phase에서 팀 모드 사용

```bash
# 명시적 팀 모드
/jikime:2-run SPEC-001 --team

# 디자이너 포함 팀
/jikime:2-run SPEC-001 --team --pattern=design_implementation
```

### 디버깅에서 팀 모드 사용

```bash
# 경쟁적 가설 조사
/jikime:build-fix --team

# 여러 가설을 병렬로 조사
```

### 강제 Sub-Agent 모드

```bash
# 팀 모드 비활성화 (단순 작업)
/jikime:2-run SPEC-001 --solo
```

---

## 설정 파일

### .jikime/config/workflow.yaml

```yaml
workflow:
  execution_mode: "auto"  # auto | subagent | team

  team:
    enabled: true
    max_teammates: 10
    default_model: "inherit"
    require_plan_approval: true
    delegate_mode: true
    teammate_display: "auto"

    auto_selection:
      min_domains_for_team: 3
      min_files_for_team: 10
      min_complexity_score: 7

    file_ownership:
      team-backend-dev:
        - "src/api/**"
        - "src/services/**"
      team-frontend-dev:
        - "src/components/**"
        - "src/pages/**"
      team-tester:
        - "tests/**"

    patterns:
      plan_research:
        roles: [researcher, analyst, architect]
      implementation:
        roles: [backend-dev, frontend-dev, tester]

    hooks:
      teammate_idle:
        enabled: true
        validate_work: true
      task_completed:
        enabled: true
        require_quality_check: true
```

---

## Fallback 동작

팀 모드가 실패하거나 요구사항이 충족되지 않을 경우:

1. **경고 로그** 출력
2. **Sub-Agent 모드**로 자동 전환
3. **마지막 완료 작업**부터 이어서 실행
4. **데이터 손실 없음**

### Fallback 트리거 조건

- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` 환경 변수 미설정
- `workflow.team.enabled: false`
- TeamCreate 실패
- 팀원 Spawn 실패
- 네트워크 오류

---

## 관련 문서

- [J.A.R.V.I.S. 오케스트레이터](./jarvis.md)
- [F.R.I.D.A.Y. 오케스트레이터](./friday.md)
- [에이전트 카탈로그](./agents.md)
- [SPEC 워크플로우](./spec-workflow.md)
- [DDD 개발 방법론](./tdd-ddd.md)

---

## 버전 정보

| 항목 | 값 |
|------|-----|
| **문서 버전** | 1.1.0 |
| **필요 Claude Code** | v2.1.32+ |
| **상태** | Experimental |
| **최종 업데이트** | 2026-02-15 |

---

## 변경 이력

| 버전 | 날짜 | 변경 내용 |
|------|------|----------|
| 1.1.0 | 2026-02-15 | 빠른 시작 가이드, 상세 사용 가이드, 트러블슈팅, Best Practices 추가 |
| 1.0.0 | 2026-02-14 | 최초 문서 작성 |
