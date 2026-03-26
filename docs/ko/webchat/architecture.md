# 아키텍처

## 기술 스택

| 분류 | 기술 |
|---|---|
| 프레임워크 | Next.js 16 (App Router) |
| 언어 | TypeScript 5 |
| UI | React 19 + Tailwind CSS v4 + shadcn/ui |
| 에디터 | Monaco Editor |
| 터미널 | xterm.js + node-pty |
| WebSocket | ws (서버) + 브라우저 네이티브 (클라이언트) |
| Claude 연동 | @anthropic-ai/claude-agent-sdk |
| 서버 런타임 | Node.js 22 (tsx로 server.ts 직접 실행) |
| 패키지 매니저 | pnpm |

---

## 서버 구조 (`server.ts`)

Next.js의 내장 서버를 사용하지 않고, 커스텀 HTTP 서버를 직접 구성합니다. 하나의 포트(4000)에서 Next.js 요청, REST API, WebSocket을 모두 처리합니다.

```
HTTP :4000
├── /ws/terminal  → terminalWss (WebSocket)
│     node-pty 세션 관리 (PTY Session Store)
│
├── /ws/chat      → chatWss (WebSocket)
│     Claude Agent SDK query() 스트리밍
│
├── /api/ws/*     → handleCustomRoutes (REST)
│     ├── GET  /api/ws/health              헬스체크
│     ├── GET  /api/ws/projects            프로젝트 목록
│     ├── GET  /api/ws/session             세션 히스토리 조회
│     ├── GET  /api/ws/session-lookup      세션 ID → 프로젝트 경로 역조회
│     ├── DELETE /api/ws/session           세션 삭제
│     ├── GET  /api/ws/files               파일 트리
│     ├── GET  /api/ws/file                파일 내용 읽기
│     ├── POST /api/ws/file                파일 내용 저장
│     ├── DELETE /api/ws/project           프로젝트 삭제
│     ├── POST /api/ws/project             프로젝트 등록
│     └── POST /api/ws/git                 Git 작업
│
├── /api/ws/session/new (Next.js API Route)
│     POST → 새 세션 파일 생성 + sessionId 반환
│
├── /api/harness/* → handleHarnessRoutes
│     Harness Engineering 워크플로우 관리
│
└── /* → Next.js handle()                  페이지 / 정적 파일
```

> **참고**: `POST /api/ws/session/new`는 `server.ts` 커스텀 라우트가 아닌 Next.js App Router API Route(`src/app/api/ws/session/new/route.ts`)로 구현되어 있습니다. `tsx server.ts`는 watch 모드 없이 실행되므로 파일 수정 시 서버 재시작이 필요하지만, Next.js API Route는 개발 서버가 핫 리로드로 즉시 반영합니다.

### URL 라우팅

세션 선택 시 URL이 **세션 ID 기반**으로 변경됩니다.

| URL 패턴 | 설명 |
|---|---|
| `/` | 세션 미선택 (통합 안내 페이지 표시) |
| `/session/{sessionId}` | 특정 세션 활성화 |

- 프로젝트 선택은 URL에 반영되지 않습니다 (사이드바 상태만 변경).
- 세션 ID는 UUID v4 형식 (`~/.claude/projects/{project}/{uuid}.jsonl`).
- 브라우저 새로고침 시 `/api/ws/session-lookup?sessionId=...` 로 프로젝트 경로를 역조회합니다.

### Claude 경로 탐색 (`findClaudePath`)

서버 시작 시 Claude CLI 네이티브 바이너리를 다음 순서로 탐색합니다.

1. `CLAUDE_PATH` 환경변수 (최우선)
2. `which -a claude` — PATH 전체 탐색
3. 고정 경로 목록 직접 확인

각 후보 경로에 대해 파일 매직 바이트를 읽어 **네이티브 바이너리 여부**를 확인합니다.
- ELF(Linux): `7F 45 4C 46`
- Mach-O(macOS): `FEEDFACE / FEEDFACF / CEFAEDFE / CFFAEDFE`

npm 래퍼 스크립트(셸 스크립트)는 제외하고, 실제 네이티브 바이너리 경로만 반환합니다.

### 프로젝트 경로 디코딩 (`decodeProjectPath`)

Claude는 프로젝트 경로의 `/`를 `-`로 치환하여 `~/.claude/projects/` 하위 디렉터리 이름으로 사용합니다.

예) `/home/anthony/jikime-adk/webchat` → `-home-anthony-jikime-adk-webchat`

경로에 `-`가 포함된 디렉터리명(예: `jikime-adk`)이 있으면 단순 `-` → `/` 치환으로는 올바른 경로를 복원할 수 없습니다. 이를 해결하기 위해 **파일시스템 탐색 기반의 최장 일치 알고리즘**을 사용합니다.

```
인코딩: -home-anthony-jikime-adk-webchat
         └─ /  home / anthony / jikime-adk / webchat
                                 ↑
                     'jikime' 보다 'jikime-adk' 가 먼저 매칭
```

### root 환경 권한 처리

`--dangerously-skip-permissions` 옵션은 Claude CLI에서 root/sudo 환경 사용이 차단됩니다. 서버 시작 시 `process.getuid()` 로 root 여부를 확인하여 다음과 같이 처리합니다.

| 환경 | 요청 permissionMode | 실제 실행 |
|---|---|---|
| 일반 유저 | `bypassPermissions` | `bypassPermissions` + `allowDangerouslySkipPermissions: true` |
| root | `bypassPermissions` | `acceptEdits` (dangerous 옵션 제외) |
| 모두 | `default` / `acceptEdits` | 요청 그대로 |

---

## 클라이언트 구조 (`src/`)

### Context 레이어

```
ServerContext       서버 목록 · 활성 서버 관리 · URL 생성
    └── WebSocketContext   WebSocket 연결 관리 · 자동 재연결
    └── ProjectContext     프로젝트 목록 · 활성 프로젝트 / 세션 · URL 라우팅
    └── TeamContext        멀티 에이전트 팀 상태 · SSE 이벤트 구독
```

**ServerContext** (`src/contexts/ServerContext.tsx`)
- 서버 목록을 `localStorage`에 저장/복원합니다.
- `getWsUrl(path)` / `getApiUrl(path)` — 활성 서버 기준 URL을 생성합니다.
- 로컬 서버(`__local__`)는 삭제 불가합니다.

**WebSocketContext** (`src/contexts/WebSocketContext.tsx`)
- 활성 서버가 바뀌면 기존 연결을 닫고 새 서버로 재연결합니다.
- `connectRef` 패턴으로 stale closure를 방지합니다.
- 연결 실패 시 3초 후 자동 재연결합니다.

**ProjectContext** (`src/contexts/ProjectContext.tsx`)
- `activeProject`: 현재 선택된 프로젝트 (URL에 반영 안 됨, 사이드바 상태).
- `activeSessionId`: 현재 세션 ID (`/session/[id]` URL에서 관리).
- `navigateToSession(project, sessionId)`: 세션 선택 시 URL 이동 (`/session/{id}` 또는 `/`).
- 페이지 로드 시 URL의 세션 ID로 프로젝트를 역조회하여 상태를 복원합니다.

**TeamContext** (`src/contexts/TeamContext.tsx`)
- 팀 목록 · 활성 팀 · 에이전트 상태를 관리합니다.
- SSE(`/api/team/{name}/events`)로 실시간 이벤트를 구독합니다.

### 주요 컴포넌트

```
src/components/
├── sidebar/
│   └── Sidebar.tsx          서버 선택 · 프로젝트/세션 목록 · 새 대화 버튼
├── chat/
│   └── ChatInterface.tsx    채팅 UI · 스트리밍 메시지 · 도구 승인 요청
├── shell/
│   └── ShellPanel.tsx       xterm.js 터미널 패널
├── file-tree/
│   └── FileTree.tsx         파일 트리 + Monaco 에디터
├── git-panel/
│   └── GitPanel.tsx         Git 변경사항 · 로그 · 브랜치 · Issues
├── team/
│   ├── BoardPanel.tsx        Team 탭 진입점 · 팀 선택 · 칸반 대시보드
│   ├── TeamBoard.tsx         에이전트 상태 칸반 보드
│   ├── TeamCreateModal.tsx   팀 생성 모달
│   ├── TaskAddModal.tsx      태스크 추가 모달
│   └── TeamServeModal.tsx    팀 실행 모달
└── layout/
    └── AppLayout.tsx         탭 헤더 · 프로젝트 컨텍스트 배지 · 패널 레이아웃
```

### AppLayout 헤더 구성

```
[☰] [탭버튼들]   📁 프로젝트명   [⚡Harness]   [⚙️][🌐][🌙]
```

- **프로젝트 컨텍스트 배지**: `activeProject.name`을 항상 동일한 위치에 표시. hover 시 전체 경로 tooltip.
- **탭 패널 헤더**: 각 패널 상단에 탭 아이콘 + 타이틀을 일관된 스타일로 표시. 각 탭의 특화 정보(브랜치명, 연결 상태 등)는 같은 헤더에 추가.

### WebSocket 메시지 프로토콜

**채팅 (`/ws/chat`)**

| 방향 | type | 내용 |
|---|---|---|
| 클라→서버 | `chat` | `{ sessionId, projectPath, prompt, model, permissionMode }` |
| 클라→서버 | `abort` | 응답 중단 요청 |
| 클라→서버 | `permission_response` | `{ requestId, allow }` |
| 서버→클라 | `text` | 스트리밍 텍스트 조각 |
| 서버→클라 | `tool_call` | 도구 호출 정보 |
| 서버→클라 | `tool_result` | 도구 실행 결과 |
| 서버→클라 | `thinking` | thinking 블록 |
| 서버→클라 | `done` | 응답 완료 `{ sessionId }` |
| 서버→클라 | `aborted` | 중단 완료 |
| 서버→클라 | `error` | 오류 메시지 |
| 서버→클라 | `token_budget` | `{ used, total }` |
| 서버→클라 | `permission_request` | `{ requestId, toolName, input }` |

**터미널 (`/ws/terminal`)**

| 방향 | type | 내용 |
|---|---|---|
| 클라→서버 | `init` | `{ sessionId, cwd, cols, rows }` |
| 클라→서버 | `input` | `{ data }` 키 입력 |
| 클라→서버 | `resize` | `{ cols, rows }` |
| 서버→클라 | `ready` | `{ sessionId }` |
| 서버→클라 | `output` | `{ data }` PTY 출력 |
| 서버→클라 | `exit` | PTY 프로세스 종료 |

---

## Docker 구조

```
Dockerfile (멀티스테이지)
├── Stage 1: builder
│   ├── python3, gcc, g++ (node-pty 컴파일용)
│   ├── pnpm install
│   └── next build
└── Stage 2: runner
    ├── 빌드 결과물만 복사
    ├── Claude CLI 설치 (npm -g @anthropic-ai/claude-code)
    └── tsx server.ts

docker-compose.yml
├── ports: 4000:4000
├── volumes: claude_data → /root/.claude  (인증 · 세션 영속)
└── healthcheck: GET /api/ws/health
```

---

## 파일 구조

```
webchat/
├── server.ts                    커스텀 HTTP + WebSocket 서버
├── harness.ts                   Harness Engineering 라우트 핸들러
├── src/
│   ├── app/
│   │   ├── (app)/
│   │   │   └── session/
│   │   │       └── [sessionId]/
│   │   │           └── page.tsx   세션 URL 라우트 (null 반환, AppLayout이 처리)
│   │   ├── api/
│   │   │   ├── ws/
│   │   │   │   ├── project/route.ts
│   │   │   │   └── session/
│   │   │   │       └── new/route.ts  새 세션 파일 생성 API
│   │   │   └── team/              멀티 에이전트 팀 API
│   │   ├── layout.tsx             Context Provider 트리
│   │   └── page.tsx               AppLayout 진입점
│   ├── components/
│   │   ├── chat/                  채팅 UI
│   │   ├── file-tree/             파일 탐색기 + 에디터
│   │   ├── git-panel/             Git 패널
│   │   ├── layout/                패널 레이아웃 (AppLayout)
│   │   ├── shell/                 터미널
│   │   ├── sidebar/               사이드바
│   │   ├── team/                  Team 탭 컴포넌트
│   │   └── ui/                    shadcn/ui 기본 컴포넌트
│   ├── contexts/
│   │   ├── ProjectContext.tsx     세션 URL 라우팅 포함
│   │   ├── ServerContext.tsx
│   │   ├── TeamContext.tsx        멀티 에이전트 팀 상태
│   │   └── WebSocketContext.tsx
│   ├── i18n/
│   │   ├── index.ts               타입 정의
│   │   └── locales/               ko.ts, en.ts, ja.ts, zh.ts
│   └── lib/
│       └── team-store.ts          팀 데이터 TTL 캐시
├── scripts/
│   ├── postinstall.js             node-pty 빌드 자동화
│   └── fix-pty-linux.sh           node-pty 수동 빌드 스크립트
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── package.json
```
