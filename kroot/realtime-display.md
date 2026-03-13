# 실시간 작업 표시 — SSE + LiveIssueCard

> 버전: 1.5.0
> 관련 파일: `components/ChatView.tsx`, `app/api/state/route.ts`, `app/api/logs/route.ts`
> 역할: jikime serve가 Claude로 작업하는 내용을 채팅 화면에 실시간 표시

---

## 개요

`jikime serve`가 `claude --print --output-format stream-json`으로 작업하는 동안, 그 내용을 웹챗 채팅 화면에 실시간으로 스트리밍한다. 두 개의 독립적인 SSE 채널을 병렬로 사용해 즉각성과 신뢰성을 동시에 확보한다.

---

## 두 채널 구조

```
jikime serve 프로세스
│
├─ stdout/stderr (모든 출력)
│       │
│       ▼
│  lib/serve.ts: appendLog()
│  logBuffers[] + logEmitters EventEmitter
│       │
│       ▼
│  GET /api/logs (SSE)  ──────────────────▶  채널 A: Log SSE
│  실시간, 폴링 없음                           즉각 반영 (주 채널)
│
└─ /api/v1/state (HTTP, Go 서버)
        │  3초 폴링
        ▼
   GET /api/state (SSE)  ────────────────▶  채널 B: State SSE
   3초 지연                                  메타 정보 + 누락 보완
```

| 항목 | 채널 A (Log SSE) | 채널 B (State SSE) |
|---|---|---|
| 지연 | 즉시 | 최대 3초 |
| 내용 | Claude 작업 내용 | 전체 상태 스냅샷 |
| 중복 제거 | 불필요 (실시간 스트림) | Timestamp 기준 dedup |
| 역할 | 즉각 표시 (주) | 누락 보완 + 메타 (부) |

---

## 채널 A: Log SSE 상세

### 서버 사이드 (`app/api/logs/route.ts`)

```typescript
// 1. 버퍼에 쌓인 기존 로그 먼저 flush
getLogBuffer(projectId).forEach(send)

// 2. 이후 실시간 구독
const unsub = subscribeLog(projectId, send)
```

연결 즉시 이전 로그를 받고, 이후 신규 라인을 실시간으로 수신한다.

### 클라이언트 파싱 (`ChatView.tsx`)

Log SSE에서 `agent event` 라인만 필터링:

```
level=INFO msg="agent event" type=notification issue_identifier=owner/repo#5 message="🔧 Bash: git status"
```

**파싱 조건:**
```typescript
if (line.includes('msg="agent event"') && line.includes('type=notification'))
```

**`slogAttr()` 함수**: slog 텍스트 형식에서 `key="value"` 또는 `key=value` 추출

```typescript
function slogAttr(raw: string, key: string): string | null {
  // key="value with spaces"  → quoted regex
  // key=value_plain          → plain regex
}
```

**메시지 → ParsedEvent 변환:**

| 메시지 패턴 | ParsedEvent.kind | 표시 방식 |
|---|---|---|
| `🔧 ` 로 시작 | `tool_call` | 툴 이름 + 입력 코드 블록 |
| `✓ tool result received` | `tool_result` | 체크 아이콘 + 텍스트 |
| 그 외 | `text` | Markdown 렌더링 |

---

## 채널 B: State SSE 상세

### 서버 사이드 (`app/api/state/route.ts`)

```typescript
// 3초마다 jikime serve HTTP API 폴링
const poll = async () => {
  const res = await fetch(`http://127.0.0.1:${project.port}/api/v1/state`)
  if (res.ok) send({ type: 'state', data: await res.json() })
  setTimeout(poll, 3000)
}
```

### `RecentEvents` — 이벤트 손실 방지

`jikime serve`의 `/api/v1/state`는 `RecentEvents[]`를 포함한다. 이슈별 최신 이벤트를 최대 50개 누적하며, 3초 폴링 사이에 발생한 이벤트가 손실되지 않도록 한다.

```go
// orchestrator.go
const maxRecentEvents = 50

type LiveEvent struct {
    IssueIdentifier string
    Message         string
    Timestamp       time.Time
}

// HandleAgentEvent()에서 추가, 이슈 완료 시 삭제
```

### 클라이언트 Merge 로직

```typescript
// 이미 수신한 이벤트는 Timestamp로 중복 제거
const seenTs = new Set(existing?.events.map(e => e._ts) ?? [])

for (const ev of r.RecentEvents) {
  if (seenTs.has(ev.Timestamp)) continue  // 중복 스킵
  // ParsedEvent 생성 후 추가
}
```

---

## LiveIssueCard 컴포넌트

작업 중인 이슈를 채팅 영역에 카드 형태로 표시한다.

### 데이터 구조

```typescript
interface LiveIssue {
  identifier: string     // "owner/repo#N"
  turnCount: number      // 현재 턴 수
  lastEvent: string      // 마지막 이벤트 타입
  lastMessage: string    // 마지막 메시지 (중복 감지용)
  events: ParsedEvent[]  // 누적된 이벤트 목록
  tokens: number         // 총 토큰 사용량
  startedAt: string
}
```

### 이벤트 렌더링 (`EventRow`)

**`tool_call`** — 툴 호출:
```
🔧 Bash
$ git status --porcelain
```

**`tool_result`** — 툴 결과:
```
✓ 결과
M  src/components/Button.tsx
```

**`text` / `raw`** — 텍스트:
```
분석한 결과, Button 컴포넌트의 스타일을 수정해야 합니다...
```

### 완료 감지

State SSE에서 `running` 목록에서 이슈 ID가 사라지면 `DoneCard`로 전환:

```typescript
prevRunningRef.current.forEach(id => {
  if (!currentIds.has(id)) {
    setLiveIssues(prev => { const n = new Map(prev); n.delete(id); return n })
    setDoneIssues(prev => [...prev, id])
  }
})
```

---

## 로그 패널 vs 채팅 영역 역할 분리

| 항목 | 로그 패널 (하단) | 채팅 LiveIssueCard |
|---|---|---|
| 데이터 소스 | Log SSE 전체 | Log SSE `agent event` only |
| 내용 | 서버 운영 정보 | Claude 작업 내용 |
| 예시 | `• workspace created — owner/repo#5` | `🔧 Bash: git status` |
| 대상 | 개발자/디버깅 | 일반 사용자 |

`agent event` 라인은 `formatLogLine()`에서 `return null`로 로그 패널에서 제외한다.

### `formatLogLine()` 처리 규칙

```
slog 라인 (msg= 포함)
  ├─ msg="agent event"   → null (채팅에서만 표시)
  ├─ msg="claude stderr" → "⚠️ stderr: {line 속성값}"
  └─ 그 외               → "• {msg} — {identifier} {hook} {error} {addr}"

훅 스크립트 출력 ([로 시작)
  └─ "  [after_create] cloned repo..."

배너/설정/기타 출력
  └─ "  {텍스트 그대로}"
```

---

## SSE 연결 상태 표시

헤더의 작은 점(●)으로 SSE 연결 상태를 표시한다:

```typescript
<span className={cn(
  'w-1.5 h-1.5 rounded-full shrink-0',
  sseConnected ? 'bg-emerald-500' : 'bg-zinc-600'
)} />
```

- 🟢 초록: `/api/state` SSE 정상 연결
- ⚫ 회색: 연결 끊김 또는 서비스 중지

---

## 이슈 생명주기 시각화

```
사용자 입력 전송
    │
    ▼
[시스템 메시지] "GitHub 이슈 #5 생성됨 — jikime serve가 작업을 시작해요"
    │
    │  (jikime serve가 이슈 감지, claude 실행 시작)
    │
    ▼
[LiveIssueCard] owner/repo#5  턴 0
  ⏳ 작업 준비 중...
    │
    │  (Log SSE agent event 수신)
    │
    ▼
[LiveIssueCard] owner/repo#5  턴 1
  🔧 Read
  src/components/Button.tsx
  🔧 Bash
  $ git diff HEAD
  ✓ 결과
  ...
    │
    │  (running 목록에서 사라짐)
    │
    ▼
[DoneCard] ✅ owner/repo#5  작업 완료
```
