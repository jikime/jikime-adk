# Architecture

## Tech Stack

| Category | Technology |
|---|---|
| Framework | Next.js 16 (App Router) |
| Language | TypeScript 5 |
| UI | React 19 + Tailwind CSS v4 + shadcn/ui |
| Editor | Monaco Editor |
| Terminal | xterm.js + node-pty |
| WebSocket | ws (server) + browser native (client) |
| Claude Integration | @anthropic-ai/claude-agent-sdk |
| Server Runtime | Node.js 22 (tsx runs server.ts directly) |
| Package Manager | pnpm |

---

## Server Structure (`server.ts`)

A custom HTTP server handles Next.js requests, REST API, and WebSocket on a single port (4000).

```
HTTP :4000
+-- /ws/terminal  -> terminalWss (WebSocket)
|     node-pty session management
|
+-- /ws/chat      -> chatWss (WebSocket)
|     Claude Agent SDK query() streaming
|
+-- /api/ws/*     -> handleCustomRoutes (REST)
|     GET  /api/ws/health, /api/ws/projects, /api/ws/session
|     POST /api/ws/file, /api/ws/git
|
+-- /api/harness/* -> handleHarnessRoutes
|     Harness Engineering workflow management
|
+-- /api/team/*   -> handleTeamRoutes (team.ts)
|     Multi-agent team orchestration API
|
+-- /*            -> Next.js handle()
```

### JikiME-ADK Integration

The webchat server calls the `jikime` CLI binary via `execFile()` for team operations:

| File | CLI Calls |
|---|---|
| `team.ts` | `jikime team create/stop/tasks/inbox` |
| `server.ts` | Reads `.claude/commands/jikime/*.md`, manages `jikime-todo`/`jikime-done` labels |
| `harness.ts` | Node.js port of `jikime serve` (WORKFLOW.md-based automation) |

> **Requirement**: The `jikime` binary must be installed and available in PATH for team features to work.

### Claude Path Discovery (`findClaudePath`)

At startup, the server searches for the Claude CLI native binary:
1. `CLAUDE_PATH` env var (highest priority)
2. `which -a claude` — full PATH search
3. Fixed path list

Validates each candidate by reading file magic bytes (ELF for Linux, Mach-O for macOS).

---

## Client Structure (`src/`)

### Context Layer

```
ServerContext       Server list, active server, URL generation
    +-- WebSocketContext   WebSocket connection, auto-reconnect
    +-- ProjectContext     Project list, active project/session, URL routing
    +-- TeamContext        Multi-agent team state, SSE event subscription
```

### Key Components

```
src/components/
+-- sidebar/        Server selection, project/session list
+-- chat/           Chat UI, streaming messages, tool approval
+-- shell/          xterm.js terminal panel
+-- file-tree/      File tree + Monaco editor
+-- git-panel/      Git changes, log, branches, Issues
+-- team/           Team tab (kanban board, create/serve modals)
+-- layout/         Tab header, panel layout (AppLayout)
```

### WebSocket Protocol

**Chat (`/ws/chat`)**

| Direction | Type | Content |
|---|---|---|
| Client->Server | `chat` | `{ sessionId, projectPath, prompt, model, permissionMode }` |
| Client->Server | `abort` | Abort response |
| Client->Server | `permission_response` | `{ requestId, allow }` |
| Server->Client | `text` | Streaming text chunk |
| Server->Client | `tool_call` / `tool_result` | Tool execution info |
| Server->Client | `done` | Response complete |
| Server->Client | `permission_request` | `{ requestId, toolName, input }` |

**Terminal (`/ws/terminal`)**

| Direction | Type | Content |
|---|---|---|
| Client->Server | `init` | `{ sessionId, cwd, cols, rows }` |
| Client->Server | `input` | `{ data }` keystroke |
| Server->Client | `output` | `{ data }` PTY output |

---

## Docker Structure

```
Dockerfile (multi-stage)
+-- Stage 1: builder (python3, gcc, pnpm install, next build)
+-- Stage 2: runner (build artifacts + Claude CLI)

docker-compose.yml
+-- ports: 4000:4000
+-- volumes: claude_data -> /root/.claude
+-- healthcheck: GET /api/ws/health
```

---

## File Structure

```
webchat/
+-- server.ts                    Custom HTTP + WebSocket server
+-- harness.ts                   Harness Engineering route handler
+-- team.ts                      Multi-agent team API + SSE
+-- src/
|   +-- app/                     Next.js App Router pages + API routes
|   +-- components/              UI components (chat, shell, git, team, etc.)
|   +-- contexts/                React contexts (Server, WebSocket, Project, Team)
|   +-- i18n/                    Localization (ko, en, ja, zh)
|   +-- lib/                     Utilities (team-store, etc.)
+-- scripts/                     Build helpers (postinstall, fix-pty)
+-- Dockerfile / docker-compose.yml
+-- package.json
```
