---
description: "[Step 4/4] Migration verification. Start dev servers, run Playwright E2E tests, compare behavior, generate final report."
argument-hint: '[--full] [--behavior] [--e2e] [--visual] [--performance] [--cross-browser] [--a11y] [--source-url URL] [--target-url URL] [--port N] [--source-port N] [--headless] [--threshold N] [--depth N]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

# Migration Step 4: Verify

**Verification Phase**: Validate migration success through automated Playwright-based testing.

[SOFT] Apply --ultrathink keyword for deep verification strategy analysis
WHY: Migration verification requires systematic planning of test execution order, server lifecycle management, and multi-dimensional quality assessment
IMPACT: Sequential thinking ensures comprehensive verification coverage with behavioral preservation validation

## CRITICAL: Input Sources

**Project settings are automatically read from `.migrate-config.yaml`.**

### Required Inputs (from Previous Steps)

1. **`.migrate-config.yaml`** - artifacts_dir, output_dir, source/target framework, verification settings
2. **`{artifacts_dir}/progress.yaml`** - Migration progress status (Step 3 output)
3. **`{output_dir}/`** - Migrated project (Step 3 output)

### Optional Inputs

4. **`{artifacts_dir}/as_is_spec.md`** - Route information for auto-discovery
5. **`{artifacts_dir}/migration_plan.md`** - User flow information for E2E scenarios

## What This Command Does

1. **Dev Server Setup** - Start source and target dev servers automatically
2. **Route Discovery** - Auto-discover testable routes from migration artifacts
3. **Characterization Tests** - Run behavior preservation tests
4. **Behavior Comparison** - Compare source/target outputs
5. **E2E Testing** - Validate full user flows with Playwright
6. **Visual Regression** - Screenshot comparison (source vs target)
7. **Performance Check** - Core Web Vitals and load time comparison
8. **Cross-Browser** - Chromium, Firefox, WebKit validation
9. **Accessibility** - WCAG compliance check with axe-core
10. **Final Report** - Comprehensive verification report with visual evidence

## Usage

```bash
# Verify current migration (reads all from config)
/jikime:migrate-4-verify

# Verify with all checks (visual + cross-browser + a11y + performance)
/jikime:migrate-4-verify --full

# Verify specific aspects
/jikime:migrate-4-verify --behavior
/jikime:migrate-4-verify --e2e
/jikime:migrate-4-verify --visual
/jikime:migrate-4-verify --performance
/jikime:migrate-4-verify --cross-browser
/jikime:migrate-4-verify --a11y

# Custom ports for dev servers
/jikime:migrate-4-verify --source-port 3000 --port 3001

# Compare live systems (skip dev server startup)
/jikime:migrate-4-verify --source-url http://old.local:3000 --target-url http://new.local:3001

# Visual regression with custom threshold
/jikime:migrate-4-verify --visual --threshold 3

# Full verification with custom depth
/jikime:migrate-4-verify --full --depth 5

# Capture migration patterns as a reusable skill
/jikime:migrate-4-verify --capture-skill
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--full` | Run ALL verification types | false |
| `--behavior` | Behavior comparison only | false |
| `--e2e` | E2E Playwright tests only | false |
| `--visual` | Visual regression (screenshot comparison) | false |
| `--performance` | Performance metrics comparison | false |
| `--cross-browser` | Cross-browser verification (Chromium, Firefox, WebKit) | false |
| `--a11y` | Accessibility (axe-core) checks | false |
| `--source-url` | Source system URL (skip source server startup) | auto |
| `--target-url` | Target system URL (skip target server startup) | auto |
| `--port` | Target dev server port | 3001 |
| `--source-port` | Source dev server port | 3000 |
| `--headless` | Run browsers in headless mode | true |
| `--threshold` | Visual diff threshold percentage | 5 |
| `--depth` | Navigation crawl depth for route discovery | 3 |
| `--capture-skill` | Generate migration skill from verified patterns | false |

**Note**: `--source-url` and `--target-url` are for comparing **already running instances**. When not provided, the command automatically starts dev servers.

---

## --capture-skill Option

After successful verification, this option captures migration patterns and creates a reusable skill for similar future migrations.

### Prerequisites

- Verification must pass (at least `--behavior` or `--full`)
- Migration artifacts must exist:
  - `{artifacts_dir}/as_is_spec.md` - Source analysis
  - `{artifacts_dir}/migration_plan.md` - Transformation rules
  - `{artifacts_dir}/progress.yaml` - Actual migration history

### Workflow

```
Step 1: Analyze Artifacts
  → Read as_is_spec.md (source patterns)
  → Read migration_plan.md (transformation rules)
  → Read progress.yaml (actual transformations applied)

Step 2: Extract Patterns
  → Identify recurring transformation patterns
  → Capture special case solutions
  → Document framework-specific mapping rules

Step 3: Invoke skill-builder Agent
  → Task(subagent_type="skill-builder", prompt="
      Create migration skill from verified patterns:
      - Source Framework: {source_framework}
      - Target Framework: {target_framework}
      - Artifacts Directory: {artifacts_dir}
      - Key Patterns: {extracted_patterns}
    ")

Step 4: Generate Skill Draft
  → Output: skills/jikime-migration-{source}-to-{target}/SKILL.md
  → Follows Progressive Disclosure format (Level 1/2/3)

Step 5: Request User Review
  → Display generated skill summary
  → Ask user to review and refine before finalizing
```

### Generated Skill Structure

```
skills/jikime-migration-{source}-to-{target}/
├── SKILL.md                 # Main skill definition (Level 1-2)
├── reference.md             # Detailed patterns (Level 3)
├── modules/
│   ├── components.md        # Component transformation rules
│   ├── routing.md           # Routing migration patterns
│   └── state.md             # State management patterns
└── examples/
    └── common-cases.md      # Real examples from this migration
```

### Example Usage

```bash
# After successful verification
/jikime:migrate-4-verify --full

# Capture patterns as reusable skill
/jikime:migrate-4-verify --capture-skill

# Combined: verify and capture in one command
/jikime:migrate-4-verify --full --capture-skill
```

### Generated Skill Frontmatter

```yaml
---
name: jikime-migration-{source}-to-{target}
description: Migration patterns from {source} to {target}
version: 1.0.0

progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000

triggers:
  keywords: ["{source}", "{target}", "migration", "convert"]
  phases: ["plan", "run"]
  agents: ["manager-ddd", "refactorer"]

metadata:
  source_framework: "{source}"
  target_framework: "{target}"
  generated_from: "{project_name}"
  generation_date: "{date}"
---
```

---

