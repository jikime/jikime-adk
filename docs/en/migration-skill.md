# migration-skill Command

A Claude Code skill generator that creates framework migration skills.

## Overview

`migration-skill` generates specialized skills for migrating legacy projects to modern frameworks. It follows the official Claude Code Skills specification and integrates with the F.R.I.D.A.Y. migration orchestrator.

**Location**: `/templates/.claude/commands/jikime/migration-skill.md`

---

## Usage

```bash
/jikime:migration-skill --from <source> --to <target> [--enhance-only]
```

| Argument | Required | Options | Description |
|----------|----------|---------|-------------|
| `--from` | Yes | `cra`, `vue`, `angular`, `svelte`, `jquery`, `php` | Source framework |
| `--to` | Yes | `nextjs`, `nuxt`, `react`, `vue` | Target framework |
| `--enhance-only` | No | - | Enhance existing skill only (don't create new) |

### Examples

```bash
# Create migration skill from CRA to Next.js
/jikime:migration-skill --from cra --to nextjs

# Create migration skill from Vue to Nuxt
/jikime:migration-skill --from vue --to nuxt

# Enhance existing Angular→React skill
/jikime:migration-skill --from angular --to react --enhance-only
```

---

## Generated Structure

```
jikime-migration-{from}-to-{to}/
├── SKILL.md                    # Main skill file (required)
├── modules/
│   ├── {from}-patterns.md      # Detailed conversion patterns
│   ├── migration-scenarios.md  # Common migration scenarios
│   └── troubleshooting.md      # Troubleshooting guide
├── examples/
│   ├── before-after.md         # Before/After code comparison
│   └── sample-migration.md     # Complete migration example
└── scripts/
    └── analyze.sh              # Analysis script (optional)
```

---

## Execution Workflow

### Phase 1: Context7 Research

- Query target framework library ID
- Query migration documentation
- Collect Codemod CLI tools
- Collect incremental migration strategies
- Identify common pitfalls

### Phase 2: Skill Discovery

- Search for existing skills: `jikime-migration-to-{to}` or `jikime-migrate-{from}-to-{to}`
- If exists → Prepare enhancement plan
- If not exists → Create new from template

### Phase 3: Skill Structure Generation

- Create complete skill directory
- Generate patterns, scenarios, examples, and script files

### Phase 4: SKILL.md Template Generation

Generate with official frontmatter format:

```yaml
---
name: migrate-{from}-to-{to}
description: "{From} to {To} migration specialist..."
argument-hint: [source-path]
disable-model-invocation: false
user-invocable: true
allowed-tools: Read, Grep, Glob, Edit, Write
context: fork
agent: Explore
---
```

### Phase 5: Pattern Module Generation

Generate `modules/{from}-patterns.md`:
- Official migration tools & codemods
- Incremental migration strategies
- Pattern mapping tables
- Common pitfalls & solutions

---

## Supported Frameworks

### Source Frameworks

| Framework | Alias | Detection Pattern |
|-----------|-------|-------------------|
| Create React App | `cra` | `react-scripts` in package.json |
| Vue.js | `vue` | `vue` in package.json |
| Angular | `angular` | `@angular/core` in package.json |
| Svelte | `svelte` | `svelte` in package.json |
| jQuery | `jquery` | `jquery` in package.json or `$()` pattern |
| PHP/Laravel | `php` | `composer.json` exists |

### Target Frameworks

| Framework | Alias | Default Version |
|-----------|-------|-----------------|
| Next.js | `nextjs` | 16 (App Router) |
| Nuxt | `nuxt` | 3 |
| React | `react` | 19 |
| Vue | `vue` | 3.5 |

---

## SKILL.md Frontmatter

### Official Fields

| Field | Description |
|-------|-------------|
| `name` | Skill display name (becomes `/slash-command`) |
| `description` | Auto-load trigger keywords |
| `argument-hint` | Autocomplete hint |
| `disable-model-invocation` | Claude auto-invocation control |
| `user-invocable` | Whether to show in user menu |
| `allowed-tools` | Tools available without permission prompt |
| `model` | Model selection (opus, sonnet, haiku) |
| `context` | Execution context (fork, inline) |
| `agent` | Sub-agent type (Explore, Plan, etc.) |
| `hooks` | Skill-scoped hooks |

### Example

```yaml
---
name: migrate-cra-to-nextjs
description: "CRA to Next.js 16 migration specialist. Handles react-scripts removal, App Router migration, SSR/SSG patterns."
argument-hint: [source-path]
user-invocable: true
allowed-tools: Read, Grep, Glob, Edit, Write
context: fork
agent: Explore
---
```

---

## Dynamic Context Injection

Dynamically collect project information within a skill:

```markdown
### Current Dependencies
!`cat package.json 2>/dev/null | grep -A 20 '"dependencies"' || echo "No package.json"`

### Framework Detection
!`ls -la src/ 2>/dev/null | head -20 || echo "No src directory"`
```

### String Substitution

- `$ARGUMENTS`: All arguments passed during invocation
- `${CLAUDE_SESSION_ID}`: Current session ID

---

## Progressive Disclosure

Skills use a 3-level loading system:

| Level | Tokens | Content |
|-------|--------|---------|
| **Level 1** | ~100 | Metadata/frontmatter only |
| **Level 2** | ~5K | Full markdown body |
| **Level 3+** | Variable | Bundled reference files (on-demand loading) |

---

## Skill Storage Locations

| Location | Path | Scope |
|----------|------|-------|
| Enterprise | Managed Settings | All organization users |
| Personal | `~/.claude/skills/<name>/SKILL.md` | All projects |
| Project | `.claude/skills/<name>/SKILL.md` | Current project only |
| Plugin | `<plugin>/skills/<name>/SKILL.md` | When plugin is active |

---

## Quality Checklist

Verify before completing skill generation:

- [ ] Context7 latest documentation query completed
- [ ] Frontmatter complies with official spec
- [ ] Description specifies when to use
- [ ] Key migration patterns documented
- [ ] Dynamic context injection for project analysis
- [ ] Code example syntax is accurate
- [ ] Incremental migration strategy included
- [ ] Official tools/codemods documented
- [ ] Troubleshooting section included
- [ ] SKILL.md is under 500 lines
- [ ] Version and changelog updated
- [ ] Supporting files properly linked

---

## Related Commands

### Migration Workflow

| Command | Description |
|---------|-------------|
| `/jikime:migrate-0-discover` | Discover source project |
| `/jikime:migrate-1-analyze` | Detailed analysis |
| `/jikime:migrate-2-plan` | Migration planning |
| `/jikime:migrate-3-execute` | Execute migration |
| `/jikime:migrate-4-verify` | Verification |
| `/jikime:friday` | Full migration orchestration |

### Related Skills

| Skill | Description |
|-------|-------------|
| `jikime-migration-to-nextjs` | Legacy → Next.js 16 |
| `jikime-migration-angular-to-nextjs` | Angular → Next.js |
| `jikime-migration-jquery-to-react` | jQuery → React |
| `jikime-migration-patterns-auth` | Authentication migration patterns |
| `jikime-migration-ast-grep` | AST-based code transformation |

---

## Context7 Query Templates

```
Migration Guide: "{from} to {to} migration guide official"
Pattern Query: "{from} {pattern} equivalent in {to}"
Best Practices: "{to} performance best practices migration"
```

---

Version: 1.0.0
Last Updated: 2026-01-26
