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
│     ├── GET  /api/ws/health       헬스체크
│     ├── GET  /api/ws/projects     프로젝트 목록
│     ├── GET  /api/ws/sessions     세션 히스토리
│     ├── GET  /api/ws/files        파일 트리
│     ├── GET  /api/ws/file         파일 내용 읽기
│     ├── POST /api/ws/file         파일 내용 저장
│     ├── DELETE /api/ws/session    세션 삭제
│     ├── DELETE /api/ws/project    프로젝트 삭제
│     └── POST /api/ws/git          Git 작업
│
└── /* → Next.js handle()          페이지 / 정적 파일
```

### Claude 경로 탐색 (`findClaudePath`)

서버 시작 시 Claude CLI 네이티브 바이너리를 다음 순서로 탐색합니다.

1. `CLAUDE_PATH` 환경변수 (최우선)
2. `which -a claude` — PATH 전체 탐색
3. 고정 경로 목록 직접 확인

각 후보 경로에 대해 파일 매직 바이트를 읽어 **네이티브 바이너리 여부**를 확인합니다.
- ELF(Linux): `7F 45 4C 46`
- Mach-O(macOS): `FEEDFACE / FEEDFACF / CEFAEDFE / CFFAEDFE`

npm 래퍼 스크립트(셸 스크립트)는 제외하고, 실제 네이티브 바이너리 경로만 반환합니다. 단, 심링크의 경우 해소된 경로가 아닌 원본 심링크 경로를 반환합니다.

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
    └── ProjectContext     프로젝트 목록 · 활성 프로젝트 / 세션
```

**ServerContext** (`src/contexts/ServerContext.tsx`)
- 서버 목록을 `localStorage`에 저장/복원합니다.
- `getWsUrl(path)` / `getApiUrl(path)` — 활성 서버 기준 URL을 생성합니다.
- 로컬 서버(`__local__`)는 삭제 불가합니다.

**WebSocketContext** (`src/contexts/WebSocketContext.tsx`)
- 활성 서버가 바뀌면 기존 연결을 닫고 새 서버로 재연결합니다.
- `connectRef` 패턴으로 stale closure를 방지합니다. `ws.onclose` 내부에서 항상 최신 `connect` 함수를 참조합니다.
- 연결 실패 시 3초 후 자동 재연결합니다.

**ProjectContext** (`src/contexts/ProjectContext.tsx`)
- 활성 서버가 변경되면 프로젝트 목록을 초기화하고 새 서버에서 재조회합니다.

### 주요 컴포넌트

```
src/components/
├── sidebar/
│   └── Sidebar.tsx       서버 선택 · 프로젝트 목록 · 세션 목록 · 연결 상태 표시
├── chat/
│   └── ChatInterface.tsx 채팅 UI · 스트리밍 메시지 · 도구 승인 요청
├── shell/
│   └── ShellPanel.tsx    xterm.js 터미널 패널
├── file-tree/
│   └── FileTree.tsx      파일 트리 + Monaco 에디터
└── layout/
    └── AppLayout.tsx     react-resizable-panels 기반 레이아웃
```

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
├── server.ts                  커스텀 HTTP + WebSocket 서버
├── src/
│   ├── app/
│   │   ├── layout.tsx         Context Provider 트리
│   │   └── page.tsx           AppLayout 진입점
│   ├── components/
│   │   ├── chat/              채팅 UI
│   │   ├── file-tree/         파일 탐색기 + 에디터
│   │   ├── layout/            패널 레이아웃
│   │   ├── shell/             터미널
│   │   ├── sidebar/           사이드바
│   │   └── ui/                shadcn/ui 기본 컴포넌트
│   └── contexts/
│       ├── ProjectContext.tsx
│       ├── ServerContext.tsx
│       └── WebSocketContext.tsx
├── scripts/
│   ├── postinstall.js         node-pty 빌드 자동화
│   └── fix-pty-linux.sh       node-pty 수동 빌드 스크립트
├── docs/                      문서
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── package.json
```
