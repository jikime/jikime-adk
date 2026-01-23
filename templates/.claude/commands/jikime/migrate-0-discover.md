---
description: "[Step 0/4] Source project discovery. Identify tech stack, architecture, and migration complexity."
argument-hint: '@<source-path> [--target nextjs|fastapi|go|flutter] [--quick]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Glob, Grep
model: inherit
---

# Migration Step 0: Discover

**Starting Phase**: Explore source code and perform initial analysis.

## What This Command Does

1. **Technology Detection** - Identify languages, frameworks, and libraries
2. **Architecture Analysis** - Understand structure, patterns, and dependencies
3. **Complexity Assessment** - Evaluate migration difficulty
4. **Config Initialization** - Create initial `.migrate-config.yaml`
5. **Target Recommendation** - Recommend suitable target frameworks

## Usage

```bash
# Discover source codebase
/jikime:migrate-0-discover @./legacy-app/

# Discover with target already decided
/jikime:migrate-0-discover @./legacy-app/ --target nextjs

# Quick discovery (overview only)
/jikime:migrate-0-discover @./legacy-app/ --quick
```

## Options

| Option | Description |
|--------|-------------|
| `@path` | Source code path to analyze (required) |
| `--target` | Target framework if already decided (optional) |
| `--quick` | Quick overview without deep analysis |

## Execution Flow

### Step 1: Analyze Source

Explore the source project to detect:
- Primary language and framework
- Framework version
- Build tools and package manager
- File count and complexity

### Step 2: Create `.migrate-config.yaml`

After discovery, **automatically create** the config file:

```yaml
# .migrate-config.yaml (created by Step 0)
version: "1.0"
project_name: legacy-app          # Derived from @path
source_path: ./legacy-app         # From @path argument
source_framework: laravel8        # Detected framework
target_framework: nextjs16        # From --target or "pending"
artifacts_dir: ./migrations/legacy-app  # Default artifacts location
output_dir: ./migrations/legacy-app/out # Default output location
created_at: "2026-01-23T10:00:00Z"
```

**If `--target` is not specified**: Set `target_framework: pending` and recommend targets in the report.

### Step 3: Generate Report

```markdown
# Discovery Report: {project_name}

## Source Overview
- **Language**: PHP 7.4
- **Framework**: Laravel 8
- **Database**: MySQL 5.7
- **Frontend**: jQuery + Blade

## Complexity Score: 7/10 (Medium-High)

## Recommended Targets
Based on analysis, suitable migration targets:
1. **Next.js 16 + Prisma** (Recommended) - Modern full-stack
2. **FastAPI + React** - Python backend option
3. **Go + htmx** - Lightweight option

## Config Created
`.migrate-config.yaml` has been initialized.
Target framework: {--target or "pending - specify in next step"}

## Next Step
Run `/jikime:migrate-1-analyze` to perform deep analysis.
(Source path and target are already saved in .migrate-config.yaml)
```

## Config File Purpose

`.migrate-config.yaml` is the **single source of truth** for all subsequent steps:

| Field | Set by | Used by |
|-------|--------|---------|
| `source_path` | Step 0 | Step 1, 3 |
| `source_framework` | Step 0 | Step 1, 2, 3 |
| `target_framework` | Step 0 or 1 | Step 2, 3 |
| `artifacts_dir` | Step 0 (default) or Step 1 | Step 2, 3, 4 |
| `output_dir` | Step 0 (default) or Step 3 | Step 3, 4 |

**Users never need to re-enter these values** in subsequent steps.

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Exploration | `Explore` | File structure and tech detection |
| Architecture | `Explore` | Pattern identification |

## Workflow (Data Flow)

```
/jikime:migrate-0-discover @./src/ --target nextjs  ← current
        │
        ├─ Creates: .migrate-config.yaml
        │   (source_path, source_framework, target_framework, artifacts_dir)
        │
        ↓
/jikime:migrate-1-analyze
        │ (reads config → no path re-entry needed)
        ├─ Updates: .migrate-config.yaml (enriches with details)
        ├─ Creates: {artifacts_dir}/as_is_spec.md
        ↓
/jikime:migrate-2-plan
        │ (reads config + as_is_spec.md)
        ├─ Creates: {artifacts_dir}/migration_plan.md
        ↓
/jikime:migrate-3-execute
        │ (reads config + plan)
        ├─ Creates: {output_dir}/ (migrated project)
        ├─ Updates: {artifacts_dir}/progress.yaml
        ↓
/jikime:migrate-4-verify
        │ (reads config + progress)
        ├─ Creates: {artifacts_dir}/verification_report.md
```

## Next Step

After discovery, proceed to next step:
```bash
/jikime:migrate-1-analyze
```

---

Version: 3.0.0
Changelog:
- v3.0.0: Added .migrate-config.yaml creation; Added --target option; Defined data flow across steps
- v2.1.0: Initial structured discover command
