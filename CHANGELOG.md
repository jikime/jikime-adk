# Changelog

All notable changes to JikiME-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.1] - 2026-02-24

### Changed

- **Token Performance Optimization**: Reduced session startup token load by ~40-50% (~31K tokens saved)
  - Moved 5 Smart Rebuild rule files (~25K tokens) from `rules/jikime/` to `skills/jikime-migration-smart-rebuild/rules/` for on-demand loading
  - Moved 4 conditional rule files (~6K tokens) to appropriate skills: `hooks.md` and `mcp-integration.md` → `jikime-foundation-claude`, `skills.md` and `web-search.md` → `jikime-foundation-core`
  - Removed 14 duplicate `@.claude/rules/jikime/` references from `CLAUDE.md` to prevent double-loading (rules directory is auto-loaded by Claude Code)
  - Rules directory reduced from 20 files to 11 always-needed core files
- **`core.md`** (v1.0.0 → v1.1.0): Consolidated HARD rules from moved files (Web Search anti-hallucination policy, MCP `.pen` file encryption rules)
- **`CLAUDE.md`**: Replaced explicit `@`-references to rules with auto-load documentation; removed redundant `@.claude/rules/` and `@.claude/contexts/` references
- **`jikime-migration-smart-rebuild` SKILL.md**: Updated Files section to reflect 5 new rule files relocated from global rules

### Performance Impact

| Metric | Before | After |
|--------|--------|-------|
| Rules files at startup | 20 (~55K tokens) | 11 (~24K tokens) |
| CLAUDE.md @rules/ refs | 14 (double-load) | 0 (auto-load only) |
| Smart Rebuild token cost | Always loaded | On-demand only |
| Estimated startup savings | — | ~31K tokens (~40-50%) |

## [0.9.0] - 2026-02-24

### Added

- **Pencil MCP Integration**: Full design tool integration for visual prototyping and design-to-code workflows
  - 14 Pencil MCP tool permissions added to `settings.json` (`batch_design`, `batch_get`, `get_editor_state`, `get_guidelines`, `get_screenshot`, `get_style_guide`, `get_style_guide_tags`, `get_variables`, `set_variables`, `open_document`, `snapshot_layout`, `find_empty_space_on_canvas`, `search_all_unique_properties`, `replace_all_matching_properties`)
  - New `jikime-design-tools` skill with Progressive Disclosure support (SKILL.md + 3 reference docs)
  - Pencil renderer reference (`pencil-renderer.md`): batch_design operations (Insert/Copy/Replace/Update/Delete/Move/Generate), Nova style tokens, workflow patterns
  - Design-to-code export reference (`pencil-code.md`): React/Tailwind export config, component generation patterns, UI kit support (Shadcn, Halo, Lunaris, Nitro)
  - Figma vs Pencil comparison guide (`comparison.md`): feature matrix, decision guide, hybrid workflow patterns
- **MCP Integration Rules** (`mcp-integration.md`): Centralized rules for all MCP servers (Context7, Sequential Thinking, jikime-memory, Pencil) with tool reference tables, activation patterns, and error handling

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

[0.9.1]: https://github.com/user/jikime-adk/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/user/jikime-adk/compare/v0.8.3...v0.9.0
[0.8.3]: https://github.com/user/jikime-adk/compare/v0.8.2...v0.8.3
[0.8.2]: https://github.com/user/jikime-adk/compare/v0.8.1...v0.8.2
[0.8.1]: https://github.com/user/jikime-adk/compare/v0.8.0...v0.8.1
[0.8.0]: https://github.com/user/jikime-adk/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/user/jikime-adk/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/user/jikime-adk/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/user/jikime-adk/compare/v0.5.2...v0.6.0
