---
description: "[Step 1/4] 레거시 프로젝트 상세 분석. 컴포넌트, 라우팅, 상태 관리, 의존성 분석."
argument-hint: '"project-path" [--framework vue|react|angular|svelte|auto] [--artifacts-output path] [--whitepaper] [--whitepaper-output path] [--lang ko|en|ja|zh]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Glob, Grep
model: inherit
---

# Migration Step 1: Analyze

레거시 프로젝트를 상세 분석하여 마이그레이션을 준비합니다.

**Note**: 다음 단계에서 마이그레이션을 실행합니다:
- `/jikime:migrate-2-plan` - 마이그레이션 계획 수립
- `/jikime:migrate` - 전체 자동화 (Next.js 16)

## Purpose

This command performs deep analysis of the source project to understand:
- Framework type and version
- Component structure and hierarchy
- State management patterns
- Routing configuration
- Dependencies and their compatibility

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| project-path | Yes | Path to the legacy project root |
| --framework | No | Force framework detection (vue\|react\|angular\|svelte\|auto) |
| --artifacts-output | No | Migration artifacts output directory (default: `./migrations/{project}/`) |
| --whitepaper | No | Generate full whitepaper package for client delivery |
| --whitepaper-output | No | Whitepaper output directory (default: `./whitepaper/`) |
| --client | No | Client company name (used in whitepaper cover) |
| --target | No | Target framework hint for whitepaper (nextjs\|vite\|fastapi) |
| --lang | No | Whitepaper language (ko\|en\|ja\|zh). Default: user's conversation_language |

## Execution Flow

### Phase 1: Framework Detection

Use **Explore** agent to scan project structure:

```
Task(subagent_type="Explore", prompt="
Analyze the project at {project-path} to detect:
1. Primary framework (Vue, React, Angular, Svelte)
2. Framework version
3. Build tool (Webpack, Vite, CRA, Angular CLI)
4. Package manager (npm, yarn, pnpm)

Check these files:
- package.json (dependencies)
- Configuration files (vue.config.js, vite.config.ts, angular.json, svelte.config.js)
- File patterns (*.vue, *.tsx, *.component.ts, *.svelte)
")
```

### Phase 2: Component Analysis

Delegate to **expert-frontend**:

```
Task(subagent_type="expert-frontend", prompt="
Analyze all components in {project-path}:
1. Create component inventory with file paths
2. Identify component hierarchy (parent-child relationships)
3. Detect component patterns:
   - Stateful vs stateless
   - Container vs presentational
   - HOCs and render props
   - Composition patterns
4. List props and events for each component
5. Identify shared/reusable components
")
```

### Phase 3: Infrastructure Analysis

Analyze project infrastructure:

1. **Routing**:
   - Route definitions
   - Dynamic routes
   - Nested routes
   - Route guards/middleware

2. **State Management**:
   - Global state (Vuex, Pinia, Redux, NgRx, Svelte stores)
   - Local component state
   - Server state (React Query, Apollo, etc.)

3. **API Integration**:
   - API client configuration
   - Authentication patterns
   - Data fetching strategies

4. **Styling**:
   - CSS approach (modules, scoped, global)
   - Preprocessors (SCSS, Less)
   - CSS-in-JS libraries
   - UI framework (Vuetify, MUI, etc.)

### Phase 4: Generate AS_IS_SPEC

Create comprehensive analysis document:

```markdown
# AS-IS Specification: {project-name}

## Project Overview
- Framework: {detected-framework} {version}
- Build Tool: {build-tool}
- Package Manager: {package-manager}

## Component Inventory
| Component | Path | Type | State | Dependencies |
|-----------|------|------|-------|--------------|
| Header | src/components/Header.vue | Stateful | Local | NavLink, Logo |
| ... | ... | ... | ... | ... |

## Routing Structure
```mermaid
graph TD
    A[/] --> B[/dashboard]
    A --> C[/settings]
    B --> D[/dashboard/analytics]
    ...
```

## State Management
- Pattern: {Vuex|Pinia|Redux|...}
- Stores: {list of stores/slices}
- Global State Shape: {structure}

## Dependencies Analysis
| Package | Version | Migration Notes |
|---------|---------|-----------------|
| vue-router | 4.x | Replace with Next.js App Router |
| ... | ... | ... |

## Special Patterns
- {Pattern 1}: {description and location}
- {Pattern 2}: {description and location}

## Risk Assessment
| Risk | Severity | Mitigation |
|------|----------|------------|
| {risk} | High/Medium/Low | {mitigation strategy} |
```

## Output

- **Directory**: `{--artifacts-output}` or `./migrations/{project-name}/` (default)
- **File**: `as_is_spec.md`
- **Format**: Markdown with Mermaid diagrams

---

## Configuration File

After analysis, the command automatically creates/updates `.migrate-config.yaml` in the project root to track artifacts location for subsequent commands.

### File Structure

```yaml
# .migrate-config.yaml
version: "1.0"

projects:
  my-vue-app:
    source_path: ./my-vue-app
    artifacts_dir: ./migrations/my-vue-app
    target_framework: nextjs16
    created_at: "2026-01-20T15:30:00Z"

  # Multiple projects can be tracked
  my-react-app:
    source_path: ./my-react-app
    artifacts_dir: ./custom/react-migration
    target_framework: nextjs16
    created_at: "2026-01-20T16:00:00Z"
```

### Purpose

- **Automatic Path Resolution**: `migrate-to-nextjs plan/skill/run` commands automatically read this file to find artifacts
- **Multi-Project Support**: Track multiple migration projects in a single workspace
- **No Repeated Flags**: Users don't need to specify `--artifacts-output` for every command

### Update Logic

```python
# On migrate-analyze completion:
config = load_or_create(".migrate-config.yaml")
config["projects"][project_name] = {
    "source_path": source_path,
    "artifacts_dir": artifacts_output or f"./migrations/{project_name}",
    "target_framework": target or "nextjs16",
    "created_at": datetime.now().isoformat()
}
save(config)
```

---

## Whitepaper Generation (--whitepaper)

When `--whitepaper` flag is provided, generate a comprehensive client-deliverable package.

### Language Selection (--lang)

The whitepaper can be generated in different languages:

| Language | Code | Description |
|----------|------|-------------|
| Korean | `ko` | 한국어 백서 |
| English | `en` | English whitepaper |
| Japanese | `ja` | 日本語ホワイトペーパー |
| Chinese | `zh` | 中文白皮书 |

**Default Behavior**:
- If `--lang` is not specified, uses user's `conversation_language` from `.jikime/config/language.yaml`
- Templates from `.claude/skills/jikime-migrate-to-nextjs/templates/pre-migration/` are used as structure reference
- Content is generated in the specified language by the delegated agents

**Language Application**:
- All document titles and headings are translated
- Technical terms remain in English (React, Next.js, API, etc.)
- Mermaid diagrams use English labels for compatibility
- Client name and project name remain as provided

### Whitepaper Output Structure

**Output Directory**: `{--whitepaper-output}` or `./whitepaper/` (default)

```
{whitepaper-output}/               # Whitepaper 패키지 (default: ./whitepaper/)
    ├── 00_cover.md                # 표지 및 목차
    ├── 01_executive_summary.md    # 경영진 요약
    ├── 02_feasibility_report.md   # 타당성 보고서
    ├── 03_architecture_report.md  # 아키텍처 보고서
    ├── 04_complexity_matrix.md    # 복잡도 매트릭스
    ├── 05_migration_roadmap.md    # 마이그레이션 로드맵
    ├── 06_baseline_report.md      # 보안/성능 기준선
    └── assets/
        └── diagrams/              # 다이어그램 자산
```

### Phase 5: Whitepaper Generation (if --whitepaper)

#### 5.1 Cover Page (00_cover.md)

```markdown
# Migration Assessment Whitepaper

## {project-name} → {target-framework} Migration

**Prepared for**: {--client or "Client Company"}
**Prepared by**: JikiME Migration Team
**Date**: {current-date}
**Version**: 1.0

---

## Table of Contents

1. Executive Summary
2. Feasibility Report
3. Architecture Report
4. Complexity Matrix
5. Migration Roadmap
6. Security & Performance Baseline

---

**Confidentiality Notice**: This document contains proprietary information...
```

#### 5.2 Executive Summary (01_executive_summary.md)

Delegate to **manager-docs** with business focus:

```
Task(subagent_type="manager-docs", prompt="
Create an executive summary for non-technical stakeholders:

Based on: {as_is_spec.md}

Include:
1. Project Overview (1 paragraph, no technical jargon)
2. Current System Summary (bullet points)
3. Why Migration is Needed (business benefits)
4. Expected Outcomes (measurable improvements)
5. Timeline Overview (high-level phases)
6. Investment Summary (effort estimation)
7. Key Risks & Mitigations (top 3)
8. Recommendation (clear Go/No-Go)

Tone: Professional, confident, accessible to C-level executives
Length: 2-3 pages maximum
")
```

#### 5.3 Feasibility Report (02_feasibility_report.md)

Delegate to **manager-strategy**:

```
Task(subagent_type="manager-strategy", prompt="
Create a migration feasibility report:

Based on: {as_is_spec.md}

Include:
1. Technical Feasibility Assessment
   - Framework compatibility score (1-10)
   - Dependency migration complexity
   - API compatibility analysis

2. Cost-Benefit Analysis
   - Estimated effort (person-days)
   - Expected ROI timeline
   - Maintenance cost comparison (before/after)

3. Risk Matrix
   | Risk | Probability | Impact | Score | Mitigation |
   |------|-------------|--------|-------|------------|

4. Alternative Analysis
   - Option A: Full Migration
   - Option B: Partial Migration
   - Option C: Refactoring Only
   - Option D: Maintain Status Quo

5. Go/No-Go Recommendation with justification
")
```

#### 5.4 Architecture Report (03_architecture_report.md)

Delegate to **expert-frontend** and **manager-strategy**:

```
Task(subagent_type="expert-frontend", prompt="
Create architecture comparison report:

Based on: {as_is_spec.md}
Target: {target-framework}

Include:
1. AS-IS Architecture Diagram (Mermaid)
   - Component hierarchy
   - Data flow
   - External integrations

2. TO-BE Architecture Diagram (Mermaid)
   - Proposed Next.js/target structure
   - App Router layout
   - Server/Client component split

3. Technology Stack Comparison Table
   | Category | Current | Target | Migration Effort |
   |----------|---------|--------|------------------|

4. Dependency Compatibility Matrix
   | Package | Current | Target Equivalent | Breaking Changes |
   |---------|---------|-------------------|------------------|

5. Architecture Improvement Points
   - Performance gains
   - Developer experience
   - Scalability improvements
")
```

#### 5.5 Complexity Matrix (04_complexity_matrix.md)

```
Task(subagent_type="expert-frontend", prompt="
Create component complexity matrix:

Based on: {as_is_spec.md}

Include:
1. Complexity Scoring Criteria
   - Lines of Code (1-5)
   - Dependencies (1-5)
   - State Complexity (1-5)
   - UI Complexity (1-5)
   - Overall Score (weighted average)

2. Component Complexity Table
   | Component | LOC | Deps | State | UI | Score | Effort (hours) | Priority |
   |-----------|-----|------|-------|-----|-------|----------------|----------|

3. Dependency Graph (Mermaid)
   - Component dependencies visualization
   - Critical path identification

4. Work Breakdown Structure (WBS)
   - Phase 1: Foundation (which components)
   - Phase 2: Core Features
   - Phase 3: Advanced Features
   - Phase 4: Polish & Optimization

5. Effort Summary
   - Total estimated hours
   - Recommended team size
   - Parallel work opportunities
")
```

#### 5.6 Migration Roadmap (05_migration_roadmap.md)

```
Task(subagent_type="manager-strategy", prompt="
Create detailed migration roadmap:

Based on: {as_is_spec.md}, {complexity_matrix}

Include:
1. Phase Overview
   | Phase | Duration | Deliverables | Team Size |
   |-------|----------|--------------|-----------|

2. Detailed Timeline (Mermaid Gantt)
   ```mermaid
   gantt
       title Migration Roadmap
       dateFormat  YYYY-MM-DD
       section Phase 1
       ...
   ```

3. Milestone Definitions
   - M1: Project Setup Complete
   - M2: Core Components Migrated
   - M3: Feature Parity Achieved
   - M4: Testing Complete
   - M5: Production Ready

4. Quality Gates per Phase
   | Phase | Entry Criteria | Exit Criteria | Validation |
   |-------|----------------|---------------|------------|

5. Rollback Plan
   - Rollback triggers
   - Rollback procedure
   - Data preservation strategy

6. Resource Allocation
   - Team composition
   - Skill requirements
   - Training needs
")
```

#### 5.7 Baseline Report (06_baseline_report.md)

```
Task(subagent_type="expert-security", prompt="
Create security and performance baseline report:

Based on: {as_is_spec.md}

Include:
1. Security Assessment
   - Current vulnerability scan results
   - Dependency security audit
   - Authentication/Authorization review
   - Data handling practices

2. Security Improvements (Post-Migration)
   | Area | Current State | Target State | Improvement |
   |------|---------------|--------------|-------------|

3. Performance Baseline
   - Current bundle size
   - Load time metrics (estimated)
   - Lighthouse score estimation
   - Core Web Vitals targets

4. Performance Targets (Post-Migration)
   | Metric | Current | Target | Improvement |
   |--------|---------|--------|-------------|
   | Bundle Size | X KB | Y KB | Z% |
   | LCP | X s | <2.5s | ... |
   | FID | X ms | <100ms | ... |

5. Monitoring Plan
   - Metrics to track
   - Alerting thresholds
   - Reporting frequency
")
```

### Whitepaper Example Usage

```bash
# Generate Korean whitepaper (default if conversation_language is ko)
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "ABC Corp" --target nextjs

# Generate English whitepaper explicitly
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "ABC Corp" --target nextjs --lang en

# Generate Japanese whitepaper
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "株式会社ABC" --target nextjs --lang ja

# Custom output directory
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "ABC Corp" --whitepaper-output ./docs/pre-migration

# Basic analysis without whitepaper
/jikime:migrate-analyze "./my-vue-app"
```

### Whitepaper Quality Checklist

- [ ] All 7 documents generated (cover + 6 reports)
- [ ] Client name appears on cover page
- [ ] All Mermaid diagrams render correctly
- [ ] Effort estimates are realistic and justified
- [ ] Risks have corresponding mitigations
- [ ] Executive summary is non-technical
- [ ] Roadmap includes clear milestones
- [ ] No placeholder text remains

---

## Quality Checklist

Before completing:

- [ ] Framework correctly identified with version
- [ ] All components cataloged
- [ ] Routing structure mapped
- [ ] State management analyzed
- [ ] Dependencies reviewed for compatibility
- [ ] Risks identified

## Example Usage

```bash
# Auto-detect framework
/jikime:migrate-analyze "./my-vue-app"

# Force Vue detection
/jikime:migrate-analyze "./legacy-project" --framework vue

# Analyze with explicit path
/jikime:migrate-analyze "/Users/dev/projects/old-react-app"
```

## Error Handling

| Error | Action |
|-------|--------|
| Path not found | Ask user to verify path |
| No framework detected | Ask user to specify --framework |
| Multiple frameworks | Ask user to select primary |
| Permission denied | Request necessary permissions |

---

Version: 1.5.0
Changelog:
- v1.5.0: Added .migrate-config.yaml auto-generation for cross-command artifact path resolution
- v1.4.0: Added --artifacts-output option; Changed default artifacts path from .claude/skills/ to ./migrations/
- v1.3.0: Added --whitepaper-output option for custom whitepaper output directory
- v1.2.0: Added --lang option for multi-language whitepaper generation (ko|en|ja|zh)
- v1.1.0: Added --whitepaper option for client-deliverable package generation
