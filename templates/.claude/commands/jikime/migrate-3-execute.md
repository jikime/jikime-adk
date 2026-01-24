---
description: "[Step 3/4] Execute migration using DDD methodology. ANALYZE → PRESERVE → IMPROVE cycle."
argument-hint: '[--module name] [--resume] [--dry-run]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

# Migration Step 3: Execute

**Execution Phase**: Execute the actual migration using DDD methodology.

## CRITICAL: Input Sources

**All settings are read from `.migrate-config.yaml`.** No need to re-enter source/target.

### Required Inputs (from Previous Steps)

1. **`.migrate-config.yaml`** - source_path, target_framework, artifacts_dir, output_dir
2. **`{artifacts_dir}/migration_plan.md`** - Migration plan (Step 2 output)

### Input Loading Flow

```
Step 1: Read `.migrate-config.yaml`
        → Extract: source_path, target_framework, artifacts_dir, output_dir

Step 2: Read `{artifacts_dir}/migration_plan.md`
        → Extract: module list, migration order, skill conventions

Step 3: Execute migration per module following plan
        → DO NOT ask user for source/target again
```

### Error Handling

If `.migrate-config.yaml` or `migration_plan.md` is not found:
- Inform user that previous steps must be completed first
- Suggest running `/jikime:migrate-2-plan` before this command
- DO NOT attempt to guess source/target frameworks

[SOFT] Apply --ultrathink keyword for deep migration execution analysis
WHY: Migration execution requires systematic DDD cycle management, behavior preservation verification, and incremental transformation validation
IMPACT: Sequential thinking ensures each module transformation preserves existing behavior while achieving target framework conventions

## What This Command Does

### DDD Cycle: ANALYZE → PRESERVE → IMPROVE

1. **ANALYZE** - Understand existing code behavior (from as_is_spec.md)
2. **PRESERVE** - Preserve behavior with characterization tests
3. **IMPROVE** - Transform to new code (following migration_plan.md conventions)
4. **Repeat** - Repeat for each module

## Usage

```bash
# Execute migration (reads all config from .migrate-config.yaml)
/jikime:migrate-3-execute

# Migrate specific module only
/jikime:migrate-3-execute --module auth

# Resume interrupted migration
/jikime:migrate-3-execute --resume

# Preview what would be done
/jikime:migrate-3-execute --dry-run
```

## Options

| Option | Description |
|--------|-------------|
| `--module` | Migrate specific module only |
| `--resume` | Resume from last checkpoint (reads progress.yaml) |
| `--dry-run` | Show what would be done without writing files |

**Note**: `source` and `target` are read from `.migrate-config.yaml`. No need to specify them.

## Execution Flow

### Step 0: Load Configuration

```python
config = load(".migrate-config.yaml")
source_path = config["source_path"]
target_framework = config["target_framework"]
artifacts_dir = config["artifacts_dir"]
output_dir = config["output_dir"]

plan = load(f"{artifacts_dir}/migration_plan.md")
modules = extract_modules(plan)
```

### Step 1: Initialize Target Project (if not exists)

Based on `migration_plan.md` project initialization section:
- Run project creation command (from skill conventions)
- Install dependencies listed in plan
- Create directory structure per plan

### Step 2: Migrate Each Module

For each module in `migration_plan.md` order:

```python
for module in modules:
    # ANALYZE: Read source module from source_path
    source_code = read_module(source_path, module)

    # PRESERVE: Create characterization tests
    create_characterization_tests(source_code, module)

    # IMPROVE: Transform to target framework
    transform_module(source_code, module, target_framework)

    # Validate: Build and test
    validate_module(output_dir, module)

    # Track progress
    update_progress(artifacts_dir, module, "completed")
```

### Step 3: Quality Validation

After all modules are migrated:
- TypeScript compiles (if applicable)
- Lint passes
- Build succeeds
- Characterization tests pass

## Progress Tracking

Progress is saved to `{artifacts_dir}/progress.yaml`:

```yaml
project: my-vue-app
source_framework: vue3            # From config
target_framework: nextjs16        # From config
status: in_progress

modules:
  total: 15
  completed: 8
  in_progress: 1
  failed: 0
  pending: 6

current:
  module: UserProfile
  phase: IMPROVE
  iteration: 2
  started_at: "2026-01-23T10:30:00Z"

history:
  - module: auth
    status: completed
    duration: "5m"
  - module: users
    status: completed
    duration: "8m"
```

## Progress Display

```
╔══════════════════════════════════════════════════════════╗
║  Migration: {project_name}                               ║
║  Source: {source_framework} → Target: {target_framework} ║
║  Phase: IMPROVE                                          ║
║  Module: user-service                                    ║
║  Progress: [████████████░░░░░░░░] 60%                   ║
╚══════════════════════════════════════════════════════════╝
```

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Analysis | `Explore` | Source code understanding |
| Test Creation | `test-guide` | Characterization tests |
| Code Generation | `frontend` or `backend` | Target code creation |
| Validation | `debugger` | Build/test error fixing |

## Workflow (Data Flow)

```
/jikime:migrate-0-discover
        ↓ (.migrate-config.yaml created)
/jikime:migrate-1-analyze
        ↓ (config updated + as_is_spec.md)
/jikime:migrate-2-plan
        ↓ (migration_plan.md)
/jikime:migrate-3-execute  ← current
        │
        ├─ Reads: .migrate-config.yaml (source, target, paths)
        ├─ Reads: {artifacts_dir}/migration_plan.md (modules, order)
        ├─ Creates: {output_dir}/ (migrated project)
        ├─ Updates: {artifacts_dir}/progress.yaml
        │
        ↓
/jikime:migrate-4-verify
```

## Next Step

After execution, proceed to next step:
```bash
/jikime:migrate-4-verify
```

---

Version: 3.0.0
Changelog:
- v3.0.0: Removed redundant source/target options; Config-first approach; All settings from .migrate-config.yaml
- v2.1.0: Initial DDD-based execution command
Methodology: DDD (ANALYZE-PRESERVE-IMPROVE)
