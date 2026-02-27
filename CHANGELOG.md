# Changelog

All notable changes to JikiME-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
