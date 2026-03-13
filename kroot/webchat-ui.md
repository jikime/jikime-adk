# webchat-git — Next.js 웹챗 UI

> 버전: 1.5.0
> 경로: `webchat-git/`
> 역할: 사용자가 채팅으로 작업을 요청하고, jikime serve의 실행 상태를 실시간으로 확인하는 웹 인터페이스

---

## 개요

Next.js 14 App Router 기반의 웹챗 UI. 사용자가 채팅 메시지를 입력하면 GitHub 이슈가 생성되고, `jikime serve`가 이슈를 처리하는 동안 작업 내용이 채팅 화면에 실시간으로 표시된다.

---

## 주요 기능

### 1. 프로젝트 관리
- GitHub 토큰 + repo 정보로 프로젝트 등록
- 프로젝트별 독립 포트로 `jikime serve` 프로세스 격리
- 프로젝트 목록 사이드바에서 전환

### 2. 채팅 기반 이슈 생성
- 채팅 입력 → GitHub Issue 자동 생성 (`jikime-todo` 라벨)
- 이슈 제목 자동 추출 (72자 이내)
- `{{ }}` 이스케이프 처리 (jikime serve 템플릿 변수 충돌 방지)

### 3. 실시간 작업 표시 (LiveIssueCard)
- `jikime serve`가 처리 중인 이슈를 채팅 영역에 카드 형태로 표시
- 툴 호출 / 텍스트 응답 / 결과를 종류별로 구분 렌더링
- 작업 완료 시 완료 카드로 전환

### 4. 로그 패널
- `jikime serve` 운영 로그 실시간 표시 (접기/펼치기)
- 서버 운영 정보 (워크스페이스 생성, 훅 실행, 재시도, stderr) 포함
- Claude 작업 내용(`agent event`)은 채팅 영역에만 표시 — 로그 패널에서 제외

---

## 화면 구조

```
┌─────────────────────────────────────────────┐
│ 헤더                                         │
│ [프로젝트명] ● 2 작업 중  [이슈▼] [:8001] [중지] │
├─────────────────────────────────────────────┤
│ 이슈 목록 패널 (접기/펼치기, 기본 닫힘)         │
│ GitHub 이슈 목록 + live 오버레이               │
├─────────────────────────────────────────────┤
│ 로그 패널 (접기/펼치기, 기본 닫힘)              │
│ jikime serve 운영 로그                        │
├─────────────────────────────────────────────┤
│                                             │
│  채팅 메시지 영역 (flex-1, 스크롤)             │
│                                             │
│  [사용자 메시지 버블]                          │
│  [시스템 알림: "GitHub 이슈 #5 생성됨"]         │
│                                             │
│  ┌─ LiveIssueCard ──────────────────────┐   │
│  │ owner/repo#5  턴 3  1,234 tokens  ∨  │   │
│  │ 🔧 Bash                               │   │
│  │ $ git status                          │   │
│  │ ✓ 결과                                │   │
│  │ 작업 중 파일: src/components/...       │   │
│  └────────────────────────────────────────┘   │
│                                             │
│  ✅ owner/repo#4  작업 완료                  │
│                                             │
├─────────────────────────────────────────────┤
│ 입력창                                       │
│ [작업 내용 입력... (GitHub Issue로 생성됨)] [▶] │
│ Enter 전송 · Shift+Enter 줄바꿈               │
└─────────────────────────────────────────────┘
```

---

## 파일 구조

```
webchat-git/
├── app/
│   ├── page.tsx                  # 루트: ProjectList 렌더
│   ├── layout.tsx
│   └── api/
│       ├── projects/route.ts     # 프로젝트 CRUD + start/stop
│       ├── issue/route.ts        # GitHub 이슈 생성/조회
│       ├── state/route.ts        # /api/v1/state SSE 프록시
│       └── logs/route.ts         # jikime serve 로그 SSE
├── components/
│   ├── ProjectList.tsx           # 프로젝트 목록 사이드바
│   ├── ChatView.tsx              # 메인 채팅 + 실시간 표시
│   ├── IssueList.tsx             # GitHub 이슈 목록 + live 오버레이
│   └── MarkdownRenderer.tsx      # Markdown 렌더러
├── lib/
│   ├── store.ts                  # 프로젝트 데이터 영속성 (JSON 파일)
│   └── serve.ts                  # jikime serve 프로세스 관리
└── public/
```

---

## API 엔드포인트

### `GET /api/projects`
등록된 프로젝트 목록 반환.

### `POST /api/projects`
새 프로젝트 등록. 포트 자동 할당 (8001부터 순차).

### `PATCH /api/projects`
```json
{ "id": "proj-xxx", "action": "start" | "stop" }
```
`start`: WORKFLOW.md 생성 → `jikime serve` 프로세스 spawn
`stop`: SIGTERM으로 프로세스 종료

### `GET /api/issue?projectId=xxx`
최근 GitHub 이슈 5개 반환 (PR 제외).

### `POST /api/issue`
```json
{ "projectId": "...", "title": "...", "body": "..." }
```
`jikime-todo` 라벨로 GitHub 이슈 생성. 라벨 없으면 자동 생성.

### `GET /api/state?projectId=xxx` (SSE)
`jikime serve`의 `/api/v1/state`를 3초마다 폴링해 SSE로 변환.

**이벤트 형식:**
```
data: {"type":"state","data":{...ServeState...}}\n\n
data: {"type":"status","running":false}\n\n
```

### `POST /api/state?projectId=xxx`
`/api/v1/refresh` 호출 → 즉시 폴링 트리거.

### `GET /api/logs?projectId=xxx` (SSE)
`jikime serve` stdout/stderr 실시간 스트리밍.

**이벤트 형식:**
```
data: {"line":"time=... level=INFO msg=\"workspace created\" ..."}\n\n
```

---

## 데이터 흐름: 채팅 메시지 전송

```
사용자 입력
    │ Enter 키
    ▼
POST /api/issue
    │ GitHub 이슈 생성 (jikime-todo 라벨)
    ▼
messages 상태 업데이트
    │ "GitHub 이슈 #N 생성됨" 시스템 메시지 추가
    ▼
POST /api/state  (refresh 트리거)
    │ jikime serve에 즉시 폴링 요청
    ▼
fetchGithubIssues() 갱신
```

---

## 데이터 흐름: 실시간 상태 표시

**채널 A — Log SSE (즉시, 주 채널):**
```
jikime serve stdout/stderr
    → /api/logs SSE
    → ChatView.tsx onmessage
    → msg="agent event" && type=notification 감지
    → issue_identifier, message 파싱
    → liveIssues Map 업데이트
    → LiveIssueCard 즉시 렌더
```

**채널 B — State SSE (3초 폴링, 보완 채널):**
```
jikime serve /api/v1/state
    → /api/state SSE (3초 폴링)
    → ChatView.tsx onmessage
    → RecentEvents[] 신규 항목 merge (Timestamp 기준 중복 제거)
    → turnCount, tokens 등 메타 정보 업데이트
    → LiveIssueCard 보완 렌더
```

두 채널이 동시에 동작하므로 Log SSE로 즉각 표시, State SSE로 누락 이벤트 보완한다.

---

## 프로세스 관리 (`lib/serve.ts`)

`jikime serve` 프로세스를 Next.js 서버 메모리에서 직접 관리한다.

```typescript
// 프로세스 시작
spawn('jikime', ['serve'], {
  cwd: p.cwd,                            // WORKFLOW.md가 있는 경로
  env: { ...process.env, GITHUB_TOKEN: p.token },
  stdio: ['ignore', 'pipe', 'pipe'],
})

// stdout/stderr 모두 캡처 → logBuffers + logEmitters
proc.stdout?.on('data', handleChunk)
proc.stderr?.on('data', handleChunk)
```

**로그 버퍼:**
- 최근 300줄 인메모리 유지 (`MAX_LOG_LINES = 300`)
- `/api/logs` SSE 구독 시 버퍼 먼저 flush → 이후 실시간 구독

---

## 프로젝트 영속성 (`lib/store.ts`)

프로젝트 정보를 JSON 파일로 저장한다.

```typescript
interface Project {
  id: string
  name: string
  repo: string       // "owner/repo"
  cwd: string        // WORKFLOW.md + workspace 경로
  port: number       // jikime serve HTTP API 포트
  pid: number | null // 실행 중인 PID
  token: string      // GitHub Personal Access Token
  status: 'running' | 'stopped'
}
```

**프로세스 생존 확인 우선순위:**
1. `pid`로 `kill(pid, 0)` 시도
2. 실패 시 `port`로 `/api/v1/state` HTTP 요청 fallback

---

## GitHub 토큰 권한 요구사항

`jikime serve`가 이슈 처리 중 PR 생성 및 병합을 수행하려면:

**Classic PAT:**
- `repo` 스코프 전체

**Fine-grained PAT:**
- `Contents: Read and write`
- `Pull requests: Read and write`
- `Issues: Read and write`
- `Metadata: Read-only` (자동 부여)
