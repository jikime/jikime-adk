# jikime serve — 자율 에이전트 오케스트레이션

> 버전: 1.5.0
> 경로: `internal/serve/`, `cmd/servecmd/`
> 역할: GitHub 이슈를 감지해 Claude Code를 자동 실행하는 장기 실행 서비스

---

## 개요

`jikime serve`는 Symphony 아키텍처에서 영감을 받은 자율 에이전트 오케스트레이션 서비스다.
GitHub 이슈를 태스크 큐로 삼아, 이슈가 생성되면 자동으로 워크스페이스를 만들고 Claude Code(`claude --print`)를 실행한다.

```
GitHub Issues (jikime-todo)
        │  polling
        ▼
  orchestrator.go
        │  goroutine per issue
        ▼
  workspace.go  →  git worktree / hook 실행
        │
        ▼
  runner.go  →  claude --print --output-format stream-json
        │  stdout NDJSON 스트림
        ▼
  summarizeEvent()  →  AgentEvent 방출
        │
        ▼
  /api/v1/state  →  웹챗 SSE 프록시
```

---

## 실행 방법

```bash
# WORKFLOW.md 생성 (초기 1회)
jikime serve init

# 서비스 시작
jikime serve                        # 현재 디렉터리 WORKFLOW.md 사용
jikime serve path/to/WORKFLOW.md    # 경로 지정
jikime serve --port 8080            # HTTP API 활성화
```

---

## 핵심 컴포넌트

### 1. WORKFLOW.md

오케스트레이터 설정 파일. YAML 헤더 + 프롬프트 템플릿으로 구성된다.

```yaml
---
tracker:
  kind: github
  project_slug: owner/repo
  active_states:    # Claude가 처리할 라벨
    - jikime-todo
  terminal_states:  # 완료 판단 라벨
    - jikime-done

polling:
  interval_ms: 15000    # 15초마다 GitHub 폴링

workspace:
  root: /tmp/jikime-myrepo   # 워크스페이스 루트

hooks:
  after_create: |   # 워크스페이스 생성 직후 실행
    git clone https://github.com/owner/repo .
  before_run: |     # claude 실행 전 실행
    git fetch origin && git reset --hard origin/main
  after_run: |      # claude 완료 후 실행
    git pull --ff-only
  timeout_ms: 60000

agent:
  max_concurrent_agents: 1   # 동시 실행 에이전트 수
  max_turns: 10
  max_retry_backoff_ms: 300000

claude:
  command: claude
  turn_timeout_ms: 3600000    # 1시간 턴 타임아웃
  stall_timeout_ms: 180000    # 3분 스탈 감지

server:
  port: 8001                  # HTTP API 포트
---

프롬프트 템플릿 ({{ issue.title }}, {{ issue.description }} 변수 사용)
```

**핫 리로드**: WORKFLOW.md 파일 변경 시 서비스 재시작 없이 자동 반영된다.

---

### 2. Orchestrator (`internal/serve/orchestrator/orchestrator.go`)

이슈 생명주기 전체를 관리하는 중앙 제어기.

**주요 상태:**
```
running  map[issueID] → RunningEntry   # 현재 실행 중인 이슈
retrying map[issueID] → RetryEntry     # 재시도 대기 중인 이슈
recentEvents map[issueID] → []LiveEvent  # 실시간 이벤트 누적 (최대 50개)
```

**폴링 루프:**
1. GitHub에서 `active_states` 라벨 이슈 목록 조회
2. 이미 실행 중이거나 재시도 대기 중인 이슈 제외
3. `max_concurrent_agents` 한도 내에서 신규 이슈 처리 시작
4. 실행 중인 이슈의 GitHub 상태 확인 → `terminal_states` 도달 시 중지

**실시간 이벤트 처리 (`HandleAgentEvent`):**
```go
func (o *Orchestrator) HandleAgentEvent(e serve.AgentEvent) {
    // AgentEventMessage 이벤트만 처리
    // Session.LastMessage 즉시 업데이트 (폴링 완료 전에도 반영)
    // recentEvents[]에 LiveEvent 추가 (최대 50개 슬라이딩 윈도우)
}
```

이 메서드 덕분에 3초 폴링 사이에도 `/api/v1/state`가 최신 작업 내용을 반환한다.

**재시도 정책:**
- 실패 시 지수 백오프(exponential backoff)로 재시도
- `max_retry_backoff_ms` 상한 적용
- 최대 재시도 횟수 초과 시 이슈를 `failed` 상태로 전환

---

### 3. Runner (`internal/serve/agent/runner.go`)

`claude` CLI를 headless 모드로 실행하고 stdout을 파싱하는 실행기.

**실행 명령:**
```bash
claude \
  --print \
  --output-format stream-json \
  --verbose \
  --dangerously-skip-permissions \
  "<prompt>"
```

**스트림 파싱 (`summarizeEvent`):**

Claude CLI는 NDJSON 형식으로 이벤트를 출력한다. 각 이벤트 타입별 처리:

| 이벤트 타입 | 처리 결과 |
|---|---|
| `assistant` + `tool_use` | `"🔧 {ToolName}: {input 요약}"` |
| `assistant` + `text` | Claude 텍스트 응답 (최대 500자) |
| `result` | 최종 답변 (최대 500자) |
| `user` + `tool_result` | `"✓ tool result received"` |
| 그 외 | 빈 문자열 → 이벤트 생략 |

**툴 입력 요약 (`summarizeToolInput`):**
`command`, `file_path`, `path`, `pattern`, `query` 순서로 주요 입력값 추출 (최대 200자).

**타임아웃 / 스탈 감지:**
- `turn_timeout_ms`: 전체 턴 타임아웃 (기본 1시간)
- `stall_timeout_ms`: 출력 없음 감지 타임아웃 (기본 3분), 5초 주기 체크

**AgentEvent 구조:**
```go
type AgentEvent struct {
    Type            AgentEventType   // session_started / notification / turn_completed / ...
    IssueID         string           // 내부 UUID
    IssueIdentifier string           // "owner/repo#N" 형식 (로그, UI 표시용)
    Message         string           // 요약된 작업 내용
    Tokens          *TokenUsage
    Timestamp       time.Time
}
```

---

### 4. Workspace Manager (`internal/serve/workspace/`)

이슈별 격리된 작업 디렉터리를 관리한다.

**생명주기:**
```
issue 감지
    │ after_create 훅 실행 (git clone 등)
    ▼
워크스페이스 준비
    │ before_run 훅 실행 (git reset 등)
    ▼
claude 실행
    │ after_run 훅 실행 (git pull 등)
    ▼
완료 또는 재시도
    │ before_remove 훅 실행 (정리)
    ▼
워크스페이스 삭제
```

---

### 5. Tracker (`internal/serve/tracker/`)

GitHub REST API를 통해 이슈 목록 조회 및 상태 변경을 담당한다.

- `active_states`: Claude가 처리할 라벨 목록
- `terminal_states`: 완료로 판단할 라벨 목록
- 이슈 식별자 형식: `owner/repo#N`

---

## HTTP API

`--port` 또는 `server.port` 설정 시 활성화된다.

### `GET /`
사람이 읽을 수 있는 대시보드 (텍스트)

```
jikime serve — 2026-03-11T10:00:00Z

Running (2):
  owner/repo#5      turns=3   🔧 Bash: git status
  owner/repo#6      turns=1   ✓ tool result received

Retrying (1):
  owner/repo#4      attempt=2  due=10:05:00  error=turn_failed: exit code 1

Tokens: input=15234 output=8921 total=24155 runtime=125.3s
```

### `GET /api/v1/state`
JSON 형식의 전체 상태. `RecentEvents` 포함.

```json
{
  "generated_at": "2026-03-11T10:00:00Z",
  "counts": { "running": 2, "retrying": 1 },
  "running": [
    {
      "IssueID": "uuid",
      "IssueIdentifier": "owner/repo#5",
      "TurnCount": 3,
      "LastMessage": "🔧 Bash: git status",
      "RecentEvents": [
        { "IssueIdentifier": "owner/repo#5", "Message": "🔧 Bash: ls -la", "Timestamp": "..." },
        { "IssueIdentifier": "owner/repo#5", "Message": "✓ tool result received", "Timestamp": "..." }
      ]
    }
  ],
  "retrying": [...],
  "jikime_totals": { "InputTokens": 15234, "OutputTokens": 8921, "TotalTokens": 24155, "SecondsRunning": 125.3 }
}
```

**`RecentEvents`**: 이슈별 최신 에이전트 이벤트 최대 50개 누적. 3초 폴링 간격 사이의 이벤트 손실을 방지한다.

### `POST /api/v1/refresh`
즉시 폴링 실행 요청 (비동기).

---

## 로그 출력 형식

`jikime serve`는 Go `slog` 텍스트 핸들러를 사용한다.

```
time=2026-03-11T10:00:00Z level=INFO msg="agent event" type=notification issue_id=uuid issue_identifier=owner/repo#5 message="🔧 Bash: git status"
time=2026-03-11T10:00:01Z level=WARN msg="claude stderr" issue_id=uuid line="Warning: ..."
time=2026-03-11T10:00:05Z level=INFO msg="workspace created" issue_identifier=owner/repo#5 path=/tmp/jikime-repo/owner-repo-5
```

`issue_identifier` 필드 덕분에 웹챗 로그 패널에서 이슈별 필터링이 가능하다.

---

## 설계 결정 사항

| 결정 | 이유 |
|---|---|
| GitHub Issues를 태스크 큐로 사용 | 분산 충돌 방지, 감사 로그 제공 |
| `--output-format stream-json` | 토큰 사용량 추적, 구조화된 이벤트 파싱 |
| `RecentEvents` 슬라이딩 윈도우 | 3초 폴링 간격 사이 이벤트 손실 방지 |
| `CLAUDECODE=` 환경변수 제거 | 중첩 Claude Code 세션 내 실행 허용 |
| 프로젝트별 독립 포트 | 프로세스 완전 격리, 장애 전파 차단 |
