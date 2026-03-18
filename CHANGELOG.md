# Changelog

All notable changes to JikiME-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.10] - 2026-03-18

### Added

- **Webchat — Resizable sidebar**: Left sidebar now uses `react-resizable-panels` v4; drag the handle to adjust width (default 256 px, min 160 px, max 480 px); collapse/expand via the Menu button; header logo width dynamically syncs to the actual panel size
- **Webchat — Code block copy button**: Hovering over a code block in chat messages reveals a copy button; shows a checkmark for 2 s after copying to clipboard
- **Webchat — Markdown export**: Download button in the chat header exports the full conversation as a `.md` file
- **Webchat — Harness progress badge**: Header displays a pulsing green badge with active agent count while a Harness Engineering workflow is running for the selected project
- **Webchat — Git Changes tab — vertical resizable split**: File list (top) and diff viewer (bottom) can be resized by dragging the separator; default 40 / 60 split
- **Webchat — Git Log tab — vertical resizable split**: Commit list (top) and diff viewer (bottom) follow the same resizable pattern
- **Webchat — Resizable handle icon**: Invisible separator line with a subtle pill icon indicating the drag area; pill is vertical `│` for left-right splits and horizontal `—` for top-bottom splits

### Changed

- **Webchat — Git diff viewer**: Light/dark theme-aware colors for added lines (`emerald`), removed lines (`red`), and hunk headers (`blue`); padding added above and below diff content for readability

### Fixed

- **Webchat — `react-resizable-panels` v4 API**: Corrected prop names (`orientation` instead of `direction`; `panelRef` instead of `ref`; `onResize` with `PanelSize` object instead of raw number); numeric size values are pixels in v4 — switched to explicit `"256px"` string format; `usePanelRef()` hook used for imperative collapse/expand
- **Webchat — Turbopack CSS build error**: Replaced arbitrary CSS variant selectors (`[aria-orientation=horizontal_&]`, `group-aria-[...]`) that Turbopack's CSS parser cannot handle with JSX conditional logic for orientation-aware handle icon sizing

## [1.6.6] - 2026-03-18

### Added

- **Webchat — Slash command menu**:
  - Typing `/` in the chat input opens an autocomplete dropdown listing all available commands
  - Keyboard navigation: `↑` / `↓` to move, `Enter` or `Tab` to select, `Esc` to close; mouse hover also updates the highlighted row
  - Commands are filtered in real-time as the user continues typing after `/`
  - Each entry shows an icon, command name, argument hint, description, and a `local` / `claude` type badge
  - **Direction A — client-side commands** (no server call):
    - `/clear` — clears the message list and starts a new Claude session
    - `/new` — alias for `/clear`
    - `/help` — renders a Markdown help table in the chat with all available commands
  - **Direction B — Claude Code passthrough commands**:
    - `/explain`, `/analyze`, `/implement`, `/improve`, `/fix`, `/test`, `/document`, `/build`, `/git`, `/design`
    - Selecting a passthrough command fills the input with `/<name> ` and waits for the user to add arguments before sending
    - Typing `/explain some topic` and pressing Enter sends the full text to Claude as a prompt
  - Client commands (`/clear`, `/new`, `/help`) are also intercepted when submitted directly via Enter without opening the menu
  - Hint text updated: `Enter 전송 · Shift+Enter 줄바꿈 · / 슬래시 커맨드`

## [1.6.5] - 2026-03-18

### Fixed

- **Webchat — Docker: terminal `execvp(3) failed` on shell spawn**:
  - `server.ts`: Shell detection now probes candidates in order (`$SHELL` → `/bin/bash` → `/bin/zsh` → `/bin/sh`) using `fs.existsSync`, preventing crash when `zsh` is absent in slim Docker images
  - `Dockerfile`: Added `zsh` to runner stage `apt-get install` for full shell support
  - `docker-compose.yml`: Added `SHELL=/bin/bash` environment variable as explicit fallback

- **Webchat — Docker: `cwd` ENOENT misreported as "Claude Code native binary not found"**:
  - `server.ts`: Validates `projectPath` with `fs.existsSync` before passing as `cwd` to the SDK; falls back to `os.homedir()` when the path does not exist
  - Prevents Node.js `spawn` ENOENT (caused by a non-existent working directory) from being incorrectly surfaced as a binary detection error by the SDK

- **Webchat — Docker: file tree resolves to wrong container path**:
  - `server.ts`: `/api/ws/files` now validates the requested path and falls back to `os.homedir()` when not found; response includes `{ path, tree }` so the UI shows the actual resolved path
  - `FileTree.tsx`: Handles both legacy array response and new `{ path, tree }` shape; header now displays the real resolved path with a tooltip

- **Webchat — Session history silently fails on parse error**:
  - `server.ts`: `/api/ws/session` parse errors now log to server console and return `[]` (200) instead of crashing with 500, so the client always receives valid JSON
  - Added server-side logging for session file lookup (`[session] jsonl not found`) and successful loads (`[session] loaded N messages`)

## [1.6.4] - 2026-03-18

### Changed

- **Webchat — UI 레이아웃 & 테마 전면 개선**:
  - 테마 파일(`globals.css`) 전면 교체 — warm 계열 oklch 색상 시스템 적용, shadow 변수 추가
  - 헤더: 라이트 모드 배경 `#cb6441` 적용, 모든 텍스트·아이콘·탭 버튼 흰색 계열로 변경, hover 배경 제거
  - 헤더: collapse 아이콘 → `Menu`(햄버거) 아이콘으로 변경, 설정 버튼 헤더 우측으로 이전
  - 사이드바: 라운드(`rounded-lg`) + 여백(`p-2`) 적용, 내부 섹션 구분선 전체 제거
  - 사이드바 하단 설정 버튼 제거 → `SettingsModal` export 후 헤더 우측 ⚙️ 버튼으로 이전
  - Body 영역: 라이트 `bg-muted` / 다크 `bg-muted` 배경 적용으로 사이드바·메인 패널 레이어 구분
  - 각 탭 패널: `bg-white` 배경 + `rounded-lg`, 헤더 `bg-white dark:bg-accent` 로 구분감 부여
  - 채팅 입력 폼: 입력 박스 `bg-white shadow-sm`, 외부 컨테이너 `bg-white`, 힌트 텍스트 진하게

## [1.6.3] - 2026-03-17

### Changed

- **Webchat — Shadcn UI 컴포넌트 전면 적용**:
  - 새 컴포넌트 설치: `Textarea`, `Checkbox`, `Label`, `Switch`, `AlertDialog`, `Dialog`
  - `Sidebar.tsx`: `<input>` → `<Input>`, `<input type="checkbox">` + `<label>` → `<Checkbox>` + `<Label>`, 저장/취소 버튼 → `<Button>`, 설정 오버레이 → `<Dialog>`, 삭제 확인 커스텀 div → `<AlertDialog>`, 서버 편집/삭제/추가 버튼 → `<Button variant="ghost/outline">`, Git PAT `<input>` → `<Input>`, 설정 버튼 → `<Button variant="ghost">`
  - `GitPanel.tsx`: 파일 선택 `<input type="checkbox">` → `<Checkbox>`, 커밋 `<input>` → `<Input>`, `alert()` 에러 처리 → `<AlertDialog>`
  - `ChatInterface.tsx`: `<textarea>` → `<Textarea>`, 확장 사고 토글 → `<Switch>`, 첨부/마이크/전송/중단/권한 버튼 → `<Button>`

## [1.6.2] - 2026-03-17

### Added

- **Webchat — Git PAT push/pull authentication**:
  - `server.ts`: Added `push` and `pull` git actions with PAT-based HTTPS auth using URL injection (`https://oauth2:TOKEN@host/...`) — reliably bypasses git credential helpers
  - PAT masked in error messages to prevent token leakage
  - SSH remotes fall through without PAT (no credential helper needed)
  - `Sidebar.tsx`: Added `gitPat` field to `AppSettings`; Settings modal gains a **Git PAT Auth** section with password input and eye-toggle (show/hide), saved to localStorage
  - `GitPanel.tsx`: Added **Pull** (sky) and **Push** (emerald) outline buttons to git header with theme-responsive colors and loading animations
  - i18n: Added `gitPatTitle`, `gitPatPlaceholder`, `gitPatDesc` keys to all 4 locales (ko/en/zh/ja)

### Fixed

- **Webchat — Chat light-mode color clashes**:
  - Tool call trigger: `text-amber-400` → `dark:text-amber-400 text-amber-600`
  - Tool result box: `bg-emerald-950/30 text-emerald-300` → `dark:` variants + `bg-emerald-50 text-emerald-800` in light
  - Thinking trigger: `text-purple-400` → `dark:text-purple-400 text-purple-600`
  - Thinking content box: `bg-purple-950/20 text-purple-200` → `dark:` variants + `bg-purple-50 text-purple-800` in light
  - Error message box: `bg-red-950/50 text-red-200` → `dark:` variants + `bg-red-50 text-red-800` in light
  - Inline code: `prose-code:text-amber-300` → `dark:prose-code:text-amber-300 prose-code:text-amber-700`
  - Links: `prose-a:text-blue-400` → `dark:prose-a:text-blue-400 prose-a:text-blue-600`

## [1.6.1] - 2026-03-17

### Fixed

- **Webchat — Light/Dark theme color consistency**:
  - Replaced all hardcoded `text-blue-400` / `text-blue-300` with `dark:text-blue-400 text-blue-600` / `dark:text-blue-300 text-blue-700` across `AppLayout`, `Sidebar`, `ChatInterface` so colors are visible in both light and dark themes
  - **Language dropdown z-index**: Header raised to `z-20 relative` so the dropdown paints above `z-10` tab panels
  - **Send button**: Changed from hardcoded `bg-orange-500 text-white` to `dark:bg-blue-600 bg-blue-700` with `disabled:text-muted-foreground` so the icon is visible when disabled in light mode
  - **Tab icons**: Added colored icons to each tab button — Chat (`MessageSquare`/sky), Terminal (`SquareTerminal`/emerald), Files (`FolderOpen`/amber), Git (`GitBranch`/purple)
  - **Settings modal**: Selected model and permission-mode items fixed from `text-blue-300` to `dark:text-blue-300 text-blue-700`; description text opacity raised for light mode readability
  - **Server picker**: Server icon, Globe icon, Check icon, connection status text, and URL text all updated with theme-responsive color variants
  - **Active tab**: Dark/light responsive blue highlight (`dark:bg-blue-500/25 dark:text-blue-200` / `bg-blue-600/15 text-blue-700`)

## [1.6.0] - 2026-03-17

### Added

- **Webchat — Claude Code Web Interface** (`webchat/`):
  - Full-featured browser-based UI for Claude Code, built with React + TypeScript + Tailwind CSS
  - Real-time streaming via WebSocket with token budget display and permission request handling
  - Multi-project sidebar: lists all Claude Code projects with open/closed folder icons and per-project session lists
  - Sessions sorted by most recent modification time (newest first)
  - **"New Chat" bug fix**: StrictMode-safe `serverAssignedIdRef` pattern prevents history reload on server-assigned sessions; eliminates page-refresh effect after sending the first message in a new conversation
  - **ThinkingIndicator**: Animated orbital ring component shown while Claude processes a response — triple-ring SVG animation with glowing core sphere and bounce-dot label
  - Docker support: `Dockerfile` + `docker-compose.yml` for containerized deployment
  - Remote server connection support via SSH tunnel documentation

- **VitePress Documentation Site** (`docs/`):
  - GitHub Pages site at `https://jikime.github.io/jikime-adk/`
  - Bilingual (Korean `/ko/`, English `/en/`) with full sidebar navigation
  - Sections: Getting Started, Skills System, AI Agents, Workflows, Migration, System Reference, Webchat
  - Local search provider, GitHub social link, MIT footer
  - Automatic deployment via GitHub Actions on `docs/**` push to `main`
  - Webchat docs migrated from `webchat/docs/` → `docs/ko/webchat/` (git history preserved)

- **README docs badge and links**:
  - Added `Docs` shield badge linking to the GitHub Pages site in both `README.md` and `README.ko.md`
  - Added documentation site URL to the Links section of both READMEs

### Fixed

- **VitePress locale routing** (`docs/.vitepress/config.ts`):
  - Changed `locales.root` → `locales.ko` so that files under `docs/ko/` are correctly served at `/ko/` URLs

---

## [1.4.3] - 2026-03-09

### Added

- **Harness Engineering test flow guide** (`docs/ko/harness-test-flow.md`, `docs/en/harness-test-flow.md`):
  - Step-by-step guide (Step 0–7) for applying Harness Engineering to a new project from scratch
  - Covers: jikime install → GitHub repo creation → `jikime serve init` → label creation → issue creation → `jikime serve` → monitoring → PR verification
  - Includes interactive wizard input reference table, expected log output, common issues & fixes
  - Test issue example: Next.js 16 + Tailwind CSS 4 + shadcn/ui simple app

### Fixed

- **`after_run` hook missing local `git pull`** (`cmd/servecmd/init.go`):
  - Added `WorkDir` field to `workflowParams` struct — captures the directory where `jikime serve init` is run
  - Both Basic and JiKiME-ADK templates now include `git pull --ff-only` targeting the local project directory in `after_run`
  - Uses `--ff-only` to safely skip if local uncommitted changes exist
  - Logs sync result: `[after_run] local repo synced at <sha>` or `[after_run] git pull skipped`
- **HTTP API endpoint correction in docs** (`docs/ko/harness-test-flow.md`, `docs/en/harness-test-flow.md`):
  - Removed non-existent `/status` and `/health` endpoints
  - Corrected to actual implemented endpoints: `GET /`, `GET /api/v1/state`, `POST /api/v1/refresh`
- **`after_run` examples updated in docs** (`docs/ko/harness-engineering.md`, `docs/en/harness-engineering.md`, `templates/.claude/commands/jikime/harness.md`):
  - All `after_run` hook examples now include the local `git pull` pattern

---

## [1.4.2] - 2026-03-09

### Fixed

- **WORKFLOW.md template — SPEC alignment** (`cmd/servecmd/init.go`, `templates/.claude/commands/jikime/harness.md`):
  - Added `claude.command: claude` — explicit Claude CLI command field (maps to Symphony SPEC `codex.command`)
  - Added `claude.turn_timeout_ms: 3600000` — session max duration 1 hour (maps to `codex.turn_timeout_ms`)
  - Fixed `agent.max_retry_backoff_ms`: `60000` → `300000` (aligns with SPEC §6.4 default of 5 minutes)
- **Harness Engineering docs** (`docs/en/harness-engineering.md`, `docs/ko/harness-engineering.md`):
  - Added "Why Harness Engineering?" section with use-case suitability matrix
  - Added `jikime serve init` and `/jikime:harness` to WORKFLOW.md creation guide
  - Expanded Git flow section: branch strategy diagram, workspace isolation, conflict prevention table
  - Added complete execution flow diagram (8-step sequence)
  - Added detailed monitoring guide: terminal logs, HTTP API examples, live monitoring commands
  - Updated all WORKFLOW.md examples with `claude.command`, `claude.turn_timeout_ms`, corrected `max_retry_backoff_ms`
  - Updated configuration reference table with all keys including `claude.command` and `claude.turn_timeout_ms`

---

## [1.4.1] - 2026-03-09

### Added

- **`jikime serve init`**: Interactive CLI wizard to generate `WORKFLOW.md`
  - Auto-detects GitHub remote URL (SSH and HTTPS formats) → suggests `owner/repo` slug
  - Auto-detects `.claude/` directory → generates JiKiME-ADK mode or Basic mode prompt
  - 5 interactive prompts: repo slug, active label, workspace root, HTTP port, max concurrent agents
  - JiKiME-ADK mode: uses `jarvis` sub-agent with full agent stack
  - Basic mode: standard git/PR workflow
  - Prints GitHub label creation commands and `jikime serve` startup command on completion
- **`/jikime:harness`** slash command: Claude Code-powered `WORKFLOW.md` generation
  - Analyzes project context (git remote, .claude/, tech stack)
  - Tech stack → specialist agent mapping (Next.js, Go, Python, Java, etc.)
  - Argument flags: `--basic`, `--port N`, `--label LABEL`, `--output PATH`
  - Guides GitHub label creation and service startup
- **Improved error message**: `jikime serve` without WORKFLOW.md now suggests `jikime serve init`

---

## [1.4.0] - 2026-03-09

### Added

- **`jikime serve` — Autonomous Agent Orchestration (Harness Engineering)**: Long-running daemon that polls GitHub Issues and dispatches Claude Code agents automatically — GitHub Issue → agent → PR → auto-merge, fully automated
  - **GitHub Issues tracker**: Polls open issues with configured labels (`active_states`); closed/labelled issues trigger terminal state reconciliation
  - **Per-issue workspace isolation**: Dedicated directories per issue under `workspace.root`; path containment safety invariants enforced
  - **Claude headless agent runner**: Executes `claude --print --output-format stream-json --verbose --dangerously-skip-permissions` in workspace; stall detection (configurable, default 5 min), turn timeout (default 1 hour)
  - **Token tracking**: Parses NDJSON stream-json events (`type: "result"`) to extract `total_input_tokens` / `total_output_tokens`; accumulated per session and in `jikime_totals` API field
  - **Orchestrator state machine**: Dispatch eligibility checks (claimed/running/slots), exponential backoff retry (`min(10s × 2^n, max_retry_backoff_ms)`), reconciliation on each tick
  - **TerminalState lifecycle**: `reconcile()` sets `TerminalState` signal on running entry; `runWorker()` detects it to run `after_run` hook and workspace cleanup in correct order — prevents `after_run hook failed: no such file` race condition
  - **PR-based workflow**: Each issue gets its own `fix/issue-N` branch; `before_run` syncs to `origin/main`; agent creates PR with `Closes #N` and auto-merges via `gh pr merge --squash --delete-branch --admin`
  - **WORKFLOW.md hot-reload**: `fsnotify` watches for changes; config re-applied without restart
  - **4 workspace lifecycle hooks**: `after_create`, `before_run`, `after_run`, `before_remove`; `hooks.timeout_ms` enforced
  - **HTTP status API** (optional): `--port` flag enables `GET /`, `GET /api/v1/state`, `POST /api/v1/refresh` on `127.0.0.1`
  - **Structured logging**: `log/slog` with `issue_id`, `issue_identifier`, `session_id`; stderr logged at Warn level for visibility
- **Harness Engineering documentation**: Concept, architecture, complete flow diagram, features, usage guide, developer guidelines, configuration reference
  - `docs/en/harness-engineering.md` (English)
  - `docs/ko/harness-engineering.md` (Korean)
- **`WORKFLOW.md.example`**: Full annotated example template (GitHub tracker, hooks, prompt with `{{ issue.identifier }}` variables, PR-based workflow)
- **`jikime-workflow-symphony` skill**: Setup guide, config reference, GitHub label mapping, orchestration flow diagram, troubleshooting table
- **Skills catalog**: 74 → 75 total skills, workflow 22 → 23

---

## [1.3.0] - 2026-03-09

### Added

- **Context Search Protocol — Duplicate Prevention**: Added "When NOT to Search" section to CLAUDE.md Section 15 — skip search when SPEC/documents already loaded in session; updated Search Process with duplicate check as first step
- **Team Agent Worktree Isolation**: Added `isolation: worktree` and `background: true` to all run-phase team agents (team-backend-dev, team-frontend-dev, team-tester, team-designer) — enables parallel execution without file conflicts via Claude Code v2.1.49+ worktree feature; read-only agents unchanged
- **Boris Cherny Best Practices** (Claude Code creator's internal best practices):
  - **Lessons Protocol** (`core.md`): Capture learnings from user corrections/agent failures in `lessons.md` — categorized entries (architecture/testing/naming/workflow/security/performance), max 50 active lessons, archive to `lessons-archive.md`, SUPERSEDED tagging
  - **Re-planning Gate** (CLAUDE.md Section 17): Detect stuck implementation — triggers after 3+ iterations with zero SPEC acceptance criteria progress; 4 recovery options via AskUserQuestion; progress tracked in `progress.md`
  - **Pre-submission Self-Review** (CLAUDE.md Section 18): Self-review before completion marker — "Is there a simpler approach?"; skipped for <50 line changes, bug fixes with repro test, annotation-approved changes
- **DDD Project-Scale-Aware Test Strategy** (`manager-ddd.md` STEP 1.5): LARGE_SCALE classification (test files > 500 OR source lines > 50,000) — targeted test execution per changed package in PRESERVE/IMPROVE phases; STEP 5 Final Verification always runs full suite
- **Agent Memory Fields**: Added persistent cross-session learning to 6 agents — `memory: user` for debugger, agent-builder, skill-builder (cross-project patterns); `memory: project` for backend, frontend, manager-quality (project-specific patterns)
- **`jikime-foundation-thinking` Skill** (new): Structured thinking toolkit with 3 modules — Critical Evaluation (7-step proposal assessment), Diverge-Converge (5-phase brainstorming 20-50 ideas → 3-5 solutions), Deep Questioning (6-layer progressive inquiry); use sequentially for complex decisions
- **Skills catalog**: 73 → 74 total skills, foundation 5 → 6 skills

## [1.1.0] - 2026-03-09

### Added

- **acceptEdits Default Permission Mode**: Changed `defaultMode` from `"default"` to `"acceptEdits"` in settings.json template — reduces repetitive permission prompts during agent workflows
- **Context Search Protocol** (CLAUDE.md Section 15): J.A.R.V.I.S./F.R.I.D.A.Y. search previous Claude Code sessions when context is needed — user confirmation, 5K token budget, 30-day lookback
- **Research-Plan-Annotate Cycle** (CLAUDE.md Section 16): Phase 0.5 deep research producing `research.md` artifact + Phase 1.5 annotation cycle (1-6 iterations) before implementation — catches architectural misunderstandings early
- **New Lifecycle Hooks** (5 hooks):
  - `PostToolUseFailure` — logs tool execution failures for diagnostics
  - `Notification` — desktop notifications via osascript (macOS) / notify-send (Linux)
  - `PermissionRequest` — policy-based permission decisions (default: ask)
  - `SubagentStart` — logs sub-agent startup with timestamp
  - `SubagentStop` — logs sub-agent completion with timestamp
- **GitHub Workflow** (`/jikime:github`): Parallel issue fixing + PR review via worktree isolation
  - Issues mode: parallel worktree-isolated agents (max 3), complexity scoring, branch naming conventions
  - PR mode: parallel review (verifier + security + quality reviewers)
  - `--solo` flag for sequential fallback
- **Context Memory in Commits**: Structured `## Context (AI-Developer Memory)` section in git commits — captures Decision, Constraint, Gotcha, Pattern, Risk, UserPref across sessions

## [1.0.0] - 2026-02-28

### Added

- **Auto-Memory Integration**: Automatic cross-session project context injection via Claude Code's native memory system
  - Discovers `~/.claude/projects/{path-hash}/memory/` directory at session start
  - Reads all `.md` files and injects content into systemMessage
  - Priority ordering: `MEMORY.md` → `lessons.md` → `context.md` → other files
  - Auto-creates memory directory on first session (enables Claude Code's auto-memory persistence)
  - Uses Claude Code's actual CWD from stdin payload for correct path hash computation
  - Truncation: 800 chars for priority files, 400 chars for others

## [0.9.3] - 2026-02-27

### Added

- **POC-First Workflow** (`jikime-workflow-poc`): Phase-based greenfield development workflow
  - Intent Classification: Greenfield → POC-First, Existing code → DDD, Test-first → TDD
  - 5-Phase structure: Make It Work (50-60%) → Refactor (15-20%) → Testing (15-20%) → Quality Gates (10-15%) → PR Lifecycle
  - Phase transition rules with [VERIFY] checkpoints between phases
  - New `/jikime:poc` slash command for workflow execution
- **Structured Task Format** (`jikime-workflow-task-format`): Systematic task decomposition
  - Do/Files/Done when/Verify/Commit 5-field task structure
  - [VERIFY] quality checkpoints inserted every 2-3 tasks
  - DDD task format variant (ANALYZE/PRESERVE/IMPROVE prefixes)
  - TodoWrite integration for real-time progress tracking
- **PR Lifecycle Automation** (`jikime-workflow-pr-lifecycle`): End-to-end PR management
  - Automated PR creation with structured description (`gh pr create`)
  - CI monitoring loop with failure diagnosis and auto-fix (max 5 retry cycles, 10 min timeout)
  - Review comment resolution loop with batch processing (max 3 cycles)
  - New `/jikime:pr-lifecycle` slash command for workflow execution
- **Skills catalog**: 70 → 73 total skills, 19 → 22 workflow skills

---

## [0.9.2] - 2026-02-27

### Removed

- **jikime-memory MCP system**: Removed the SQLite + vector embedding based cross-session memory system in its entirety

  **Go source** (37 files deleted, 4 files edited):
  - Deleted `internal/memory/` package (21 files) — store, search, hybrid, indexer, extractor, chunker, embedding, schema, types
  - Deleted `cmd/memorycmd/` package (9 files) — `memory search/list/show/delete/stats/gc/index/migrate` CLI commands
  - Deleted 7 memory hook files from `cmd/hookscmd/` — `memory_load`, `memory_flush`, `memory_save`, `memory_search`, `memory_track`, `memory_complete`, `embed_backfill`
  - `cmd/mcpcmd/serve.go`: Removed all memory types, handlers, and 6 MCP tool registrations (`memory_search`, `memory_get`, `memory_load`, `memory_save`, `memory_stats`, `memory_reindex`)
  - `cmd/hookscmd/hooks.go`: Removed 7 memory hook registrations
  - `cmd/hookscmd/user_prompt_submit.go`: Removed prompt-to-memory-store save logic
  - `cmd/root.go`: Removed `memorycmd` import and `memory` subcommand registration

  **Templates** (config):
  - `templates/.mcp.json`: Removed `jikime-memory` MCP server entry
  - `templates/.claude/settings.json`: Removed 5 hooks (`memory-save`, `memory-flush`, `memory-track`, `memory-complete`, `memory-prompt-save`) and `mcp__jikime-memory__*` permission; fixed resulting trailing comma (JSON parse error)

  **Templates** (instructions & skills):
  - `templates/CLAUDE.md`: Removed Section 14 "Project Memory (jikime-memory MCP)" — 5 HARD Rules and core principle
  - `templates/.claude/skills/jikime-foundation-claude/reference/mcp-integration-rules.md`: Removed `jikime-memory` rows from Available MCP Servers and Server Activation Patterns tables
  - Deleted `templates/.claude/skills/jikime-foundation-claude/reference/jikime-memory-guide.md`

  **Documentation**:
  - Deleted `docs/en/memory.md`, `docs/en/session-memory-flow.md`, `docs/ko/memory.md`, `docs/ko/session-memory-flow.md`
  - `README.md` / `README.ko.md`: Removed Session Memory feature row, "Session Memory System" section, CLI command table rows (`jikime memory search`, `jikime memory stats`), docs table rows, and Project Structure entries for `memorycmd/` and `memory/`

---

## [0.9.1] - 2026-02-24

### Changed

- **Token Performance Optimization Phase 1**: Reduced session startup token load by ~40-50% (~31K tokens saved)
  - Moved 5 Smart Rebuild rule files (~25K tokens) from `rules/jikime/` to `skills/jikime-migration-smart-rebuild/rules/` for on-demand loading
  - Moved 4 conditional rule files (~6K tokens) to appropriate skills: `hooks.md` and `mcp-integration.md` → `jikime-foundation-claude`, `skills.md` and `web-search.md` → `jikime-foundation-core`
  - Removed 14 duplicate `@.claude/rules/jikime/` references from `CLAUDE.md` to prevent double-loading (rules directory is auto-loaded by Claude Code)
  - Rules directory reduced from 20 files to 11 always-needed core files
- **Token Performance Optimization Phase 2**: Additional ~5,900 tokens saved
  - Trimmed 6 verbose rules files by 64% (1,501 → 546 lines): `patterns.md`, `tone.md`, `git-workflow.md`, `testing.md`, `performance.md`, `security.md`
  - Condensed CLAUDE.md by 42% (1,080 → 630 lines): removed duplicate content in sections 6/7/9, moved detailed reference content from sections 11/11.1/12/14/15 to skills
  - New reference files: `sequential-thinking-guide.md` in `jikime-foundation-claude/reference/`
- **`core.md`** (v1.0.0 → v1.1.0): Consolidated HARD rules from moved files (Web Search anti-hallucination policy, MCP `.pen` file encryption rules)
- **`CLAUDE.md`** (v12.0.0 → v13.0.0): Replaced explicit `@`-references with auto-load; condensed reference sections to skill pointers
- **`jikime-migration-smart-rebuild` SKILL.md**: Updated Files section to reflect 5 new rule files relocated from global rules

### Performance Impact

| Metric | Before | After |
|--------|--------|-------|
| Rules files at startup | 20 (~55K tokens) | 11 (~24K tokens) |
| Rules file lines | 2,539 lines | 1,584 lines (-38%) |
| CLAUDE.md lines | 1,080 lines | 630 lines (-42%) |
| CLAUDE.md @rules/ refs | 14 (double-load) | 0 (auto-load only) |
| Smart Rebuild token cost | Always loaded | On-demand only |
| Estimated total savings | — | ~37K tokens (~55-60%) |

## [0.9.0] - 2026-02-24

### Added

- **Pencil MCP Integration**: Full design tool integration for visual prototyping and design-to-code workflows
  - 14 Pencil MCP tool permissions added to `settings.json` (`batch_design`, `batch_get`, `get_editor_state`, `get_guidelines`, `get_screenshot`, `get_style_guide`, `get_style_guide_tags`, `get_variables`, `set_variables`, `open_document`, `snapshot_layout`, `find_empty_space_on_canvas`, `search_all_unique_properties`, `replace_all_matching_properties`)
  - New `jikime-design-tools` skill with Progressive Disclosure support (SKILL.md + 3 reference docs)
  - Pencil renderer reference (`pencil-renderer.md`): batch_design operations (Insert/Copy/Replace/Update/Delete/Move/Generate), Nova style tokens, workflow patterns
  - Design-to-code export reference (`pencil-code.md`): React/Tailwind export config, component generation patterns, UI kit support (Shadcn, Halo, Lunaris, Nitro)
  - Figma vs Pencil comparison guide (`comparison.md`): feature matrix, decision guide, hybrid workflow patterns
- **MCP Integration Rules** (`mcp-integration.md`): Centralized rules for all MCP servers (Context7, Sequential Thinking, Pencil) with tool reference tables, activation patterns, and error handling

### Changed

- **`designer-ui` agent** (v2.0.0 → v3.0.0): Added Pencil MCP tools, `jikime-design-tools` skill, and complete Pencil workflow with HARD RULES
- **`team-designer` agent** (v1.0.0 → v2.0.0): Added 14 Pencil MCP tools, `jikime-design-tools` skill, comprehensive Pencil MCP workflow (5-step design process), file management guidelines
- **`skills-catalog.yaml`**: Total skills 69 → 70, registered `jikime-design-tools` in domain category
- **`CLAUDE.md`**: Added `mcp-integration.md` to rules reference list

### Design Tool HARD RULES

- `.pen` files are **encrypted** — NEVER use Read/Grep/Glob to access their contents
- ALWAYS use Pencil MCP tools for `.pen` file operations
- ALWAYS call `get_editor_state()` before any Pencil operation
- ALWAYS validate with `get_screenshot()` after design changes
- Maximum 25 operations per `batch_design` call

## [0.8.3] - 2026-02-24

### Fixed

- Restore accidentally deleted `release.yml` workflow
- Remove unnecessary files and dead code

### Added

- Site-flow CLI bootstrap authentication module (`SPEC-MIGRATION-001`)
- Site-flow API client library and migration command integration (`SPEC-MIGRATION-001`)

## [0.8.2] - 2026-02-16

### Fixed

- Memory timestamp sorting and project root detection consistency

## [0.8.1] - 2026-02-15

### Added

- Internationalization (i18n) support with English/Korean documentation structure

## [0.8.0] - 2026-02-15

### Added

- Agent Teams parallel multi-agent system with experimental support
- 8 team agents: researcher, analyst, architect, backend-dev, designer, frontend-dev, tester, quality
- Team APIs: TeamCreate, SendMessage, TaskCreate/Update/List/Get, TeamDelete
- File ownership strategy for parallel execution conflict prevention
- Team hook events (TeammateIdle, TaskCompleted)

## [0.7.0] - 2026-02-09

### Added

- Smart Rebuild: AI-powered legacy site rebuilding workflow (screenshot-based)
  - Phase 1: Playwright capture with lazy capture strategy
  - Phase 2: Source analysis with static/dynamic classification
  - Phase 3: Frontend generation with HITL (Human-in-the-Loop) refinement
  - Phase G: Backend integration (Spring Boot, FastAPI, Go Fiber, NestJS)
- `jikime-migration-smart-rebuild` skill
- `/jikime:smart-rebuild` Claude Code command
- `jikime-library-streamdown` skill
- DB schema extraction support
- Multiple agent additions and improvements

## [0.6.1] - 2026-02-08

### Added

- `jikime-library-tds-react-native` skill (Toss Design System for React Native)

## [0.6.0] - 2026-02-07

### Added

- Architecture pattern expansion features

[0.9.3]: https://github.com/user/jikime-adk/compare/v0.9.2...v0.9.3
[0.9.2]: https://github.com/user/jikime-adk/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/user/jikime-adk/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/user/jikime-adk/compare/v0.8.3...v0.9.0
[0.8.3]: https://github.com/user/jikime-adk/compare/v0.8.2...v0.8.3
[0.8.2]: https://github.com/user/jikime-adk/compare/v0.8.1...v0.8.2
[0.8.1]: https://github.com/user/jikime-adk/compare/v0.8.0...v0.8.1
[0.8.0]: https://github.com/user/jikime-adk/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/user/jikime-adk/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/user/jikime-adk/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/user/jikime-adk/compare/v0.5.2...v0.6.0
