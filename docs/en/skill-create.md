# skill-create Command

A universal skill generator for creating Claude Code skills.

## Overview

`skill-create` generates various types of specialized skills. It follows the official Claude Code Skills specification and applies the Progressive Disclosure pattern.

**Location**: `/templates/.claude/commands/jikime/skill-create.md`

---

## Usage

```bash
/jikime:skill-create --type <type> --name <name> [--enhance-only]
```

| Argument | Required | Options | Description |
|----------|----------|---------|-------------|
| `--type` | Yes | `lang`, `platform`, `domain`, `workflow`, `library`, `framework` | Skill type |
| `--name` | Yes | Any name | Skill name |
| `--enhance-only` | No | - | Enhance existing skill only (do not create new) |

### Examples

```bash
# Create Rust language expert skill
/jikime:skill-create --type lang --name rust

# Create Firebase platform skill
/jikime:skill-create --type platform --name firebase

# Create security domain skill
/jikime:skill-create --type domain --name security

# Create CI/CD workflow skill
/jikime:skill-create --type workflow --name ci-cd

# Create Prisma library skill
/jikime:skill-create --type library --name prisma

# Create Remix framework skill
/jikime:skill-create --type framework --name remix

# Enhance existing Python skill
/jikime:skill-create --type lang --name python --enhance-only
```

---

## Skill Type Generation Structures

### `lang` (Language Expert)

```
jikime-lang-{name}/
├── SKILL.md              # Main skill file
├── examples.md           # Production code examples
└── reference.md          # Complete API reference
```

**Purpose**: Programming language syntax, patterns, and best practices

**Examples**: `jikime-lang-typescript`, `jikime-lang-python`, `jikime-lang-rust`

### `platform` (Platform Integration)

```
jikime-platform-{name}/
├── SKILL.md              # Main skill file
├── setup.md              # Setup and configuration guide
└── reference.md          # API and integration reference
```

**Purpose**: Cloud platform and SaaS service integration

**Examples**: `jikime-platform-vercel`, `jikime-platform-supabase`, `jikime-platform-firebase`

### `domain` (Domain Expert)

```
jikime-domain-{name}/
├── SKILL.md              # Main skill file
├── patterns.md           # Domain-specific patterns
└── examples.md           # Implementation examples
```

**Purpose**: Specialized knowledge in specific technical areas

**Examples**: `jikime-domain-frontend`, `jikime-domain-backend`, `jikime-domain-security`

### `workflow` (Workflow)

```
jikime-workflow-{name}/
├── SKILL.md              # Main skill file
├── steps.md              # Workflow steps and phases
└── examples.md           # Workflow examples
```

**Purpose**: Development processes, automation, and CI/CD patterns

**Examples**: `jikime-workflow-tdd`, `jikime-workflow-ddd`, `jikime-workflow-ci-cd`

### `library` (Library Expert)

```
jikime-library-{name}/
├── SKILL.md              # Main skill file
├── examples.md           # Usage examples
└── reference.md          # API reference
```

**Purpose**: Usage of specific libraries/packages

**Examples**: `jikime-library-zod`, `jikime-library-prisma`, `jikime-library-shadcn`

### `framework` (Framework Expert)

```
jikime-framework-{name}/
├── SKILL.md              # Main skill file
├── patterns.md           # Framework patterns
└── upgrade.md            # Version upgrade guide
```

**Purpose**: Framework-specific conventions, routing, and components

**Examples**: `jikime-framework-nextjs`, `jikime-framework-remix`, `jikime-framework-nuxt`

---

## Execution Workflow

### Phase 1: Context7 Research

Query related documentation from Context7 based on skill type:

1. Library ID lookup: `resolve-library-id`
2. Documentation query: `query-docs`
3. Collect: API patterns, best practices, common pitfalls, version information

### Phase 2: Skill Discovery

Search for existing skills:
- Pattern: `jikime-{type}-{name}` or `jikime-{name}`
- Exists + `--enhance-only` → Prepare enhancement plan
- Exists + no flag → Ask user (enhance / create new / cancel)
- Does not exist → Create new

### Phase 3: SKILL.md Template Generation

Generate with official frontmatter format:

```yaml
---
name: jikime-{type}-{name}
description: "{Name} {type} specialist covering..."
version: 1.0.0
tags: ["{type}", "{name}"]
triggers:
  keywords: ["{name}"]
  phases: ["run"]
  agents: [relevant agents]
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000
user-invocable: false
allowed-tools: [Read, Grep, Glob, Context7 MCP tools]
---
```

### Phase 4: Supporting Files Generation

Generate supporting files by type:

| Type | Generated Files |
|------|-----------------|
| lang | examples.md + reference.md |
| platform | setup.md + reference.md |
| domain | patterns.md + examples.md |
| workflow | steps.md + examples.md |
| library | examples.md + reference.md |
| framework | patterns.md + upgrade.md |

### Phase 5: Progressive Disclosure Integration

Reference supporting files in SKILL.md:

```markdown
## Advanced Patterns

For comprehensive documentation, see:

- examples.md for production-ready code examples
- reference.md for complete API reference
```

---

## SKILL.md Frontmatter

### Official Fields

| Field | Description |
|-------|-------------|
| `name` | Skill name (`jikime-{type}-{name}`) |
| `description` | Description including when to use |
| `version` | Semantic version |
| `tags` | Classification tags |
| `triggers` | Level 2 loading trigger conditions |
| `progressive_disclosure` | Progressive Disclosure settings |
| `user-invocable` | Whether user can invoke directly |
| `allowed-tools` | List of allowed tools |

---

## Progressive Disclosure

Skills use a 3-level loading system:

| Level | Tokens | Content |
|-------|--------|---------|
| **Level 1** | ~100 | Metadata/frontmatter only |
| **Level 2** | ~5K | Full SKILL.md body |
| **Level 3+** | Variable | Bundled reference files (on-demand loading) |

---

## Skill Storage Locations

| Location | Path | Scope |
|----------|------|-------|
| Personal | `~/.claude/skills/<name>/SKILL.md` | All projects |
| Project | `.claude/skills/<name>/SKILL.md` | Current project only |
| Plugin | `<plugin>/skills/<name>/SKILL.md` | When plugin is activated |

**Default**: Project level `.claude/skills/jikime-{type}-{name}/`

---

## Quality Checklist

Verify before completing skill creation:

- [ ] Context7 latest documentation query completed
- [ ] Frontmatter complies with official spec
- [ ] Description specifies when to use ("Use when...")
- [ ] Progressive Disclosure settings completed
- [ ] SKILL.md is under 500 lines
- [ ] Supporting files properly linked
- [ ] Code examples have correct syntax
- [ ] Context7 library mapping documented
- [ ] Related skills identified
- [ ] Troubleshooting section included
- [ ] Version and changelog updated

---

## Related Commands

| Command | Description |
|---------|-------------|
| `/jikime:migration-skill` | Create migration-specific skill |
| `jikime-adk skill list` | List all skills |
| `jikime-adk skill info <name>` | View skill details |
| `jikime-adk skill search <keyword>` | Search skills |

---

## Differences from migration-skill

| Item | skill-create | migration-skill |
|------|--------------|-----------------|
| **Purpose** | General-purpose skill creation | Migration-specific |
| **Types** | 6 types (lang, platform, domain, workflow, library, framework) | Migration only |
| **Generation Structure** | Varies by type | Fixed: modules/, examples/, scripts/ |
| **Context7 Query** | Type-specific API/patterns | Migration guides |
| **Workflow Integration** | General development | F.R.I.D.A.Y. migration |

---

## Context7 Query Templates

```
# Language skill
Query: "{name} language features best practices"

# Platform skill
Query: "{name} SDK API integration guide"

# Domain skill
Query: "{name} architecture patterns best practices"

# Workflow skill
Query: "{name} workflow automation CI/CD"

# Library skill
Query: "{name} library API usage examples"

# Framework skill
Query: "{name} framework conventions routing"
```

---

Version: 1.0.0
Last Updated: 2026-01-26
