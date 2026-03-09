# Changelog

All notable changes to JikiME-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
