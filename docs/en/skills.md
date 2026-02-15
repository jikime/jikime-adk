# JikiME-ADK Skill System

Documentation for the skill system structure and management tools in JikiME-ADK.

## Overview

Skills are modules that provide Claude Code with specialized knowledge for specific domains or tasks. They use tokens efficiently through the Progressive Disclosure pattern.

## Skill Structure

```
templates/.claude/skills/
├── _template/                    # Skill template
│   ├── SKILL.md
│   └── tests/
│       ├── README.md
│       └── examples.yaml
├── jikime-lang-typescript/       # Language skill
│   └── SKILL.md
├── jikime-domain-frontend/       # Domain skill
│   └── SKILL.md
├── jikime-workflow-spec/         # Workflow skill
│   └── SKILL.md
└── ...
```

### Skill Naming Convention

```
jikime-{domain}-{name}
```

| Domain | Description | Example |
|--------|-------------|---------|
| `lang` | Programming language | `jikime-lang-typescript` |
| `domain` | Development domain | `jikime-domain-frontend` |
| `workflow` | Workflow | `jikime-workflow-spec` |
| `platform` | Platform/Service | `jikime-platform-vercel` |
| `framework` | Framework | `jikime-framework-nextjs@16` |
| `library` | Library | `jikime-library-zod` |
| `foundation` | Core foundation | `jikime-foundation-core` |
| `marketing` | Marketing | `jikime-marketing-seo` |
| `tool` | Tool | `jikime-tool-ast-grep` |
| `migration` | Migration | `jikime-migration-to-nextjs` |

### SKILL.md Structure

```yaml
---
name: jikime-example-skill
description: Brief description of the skill
version: 1.0.0
tags:
  - example
  - tutorial

# Progressive Disclosure settings
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000

# Trigger conditions
triggers:
  keywords: ["example", "example"]
  phases: ["plan", "run"]
  agents: ["manager-spec"]
  languages: ["typescript"]
---

# Skill body

Write the detailed content of the skill here.
```

## Skill CLI (Discovery)

You can explore and search for skills using the `jikime-adk skill` command.

### List Skills (list)

```bash
# List all skills
jikime-adk skill list

# Filter by tag
jikime-adk skill list --tag framework

# Filter by phase (plan, run, sync)
jikime-adk skill list --phase run

# Filter by agent
jikime-adk skill list --agent frontend

# Filter by language
jikime-adk skill list --language typescript

# Specify output format
jikime-adk skill list --format json      # JSON format
jikime-adk skill list --format compact   # Compact format
jikime-adk skill list --format table     # Table format (default)
```

### Search Skills (search)

```bash
# Search by text
jikime-adk skill search nextjs
jikime-adk skill search "react components" --limit 5

# Search by tags (comma-separated)
jikime-adk skill search --tags framework,nextjs

# Combined filtering
jikime-adk skill search --phases run --languages typescript
jikime-adk skill search --agents frontend,backend --limit 10
```

**Search result example:**

```
Search Results (3 found):
--------------------------------------------------------------------------------
1. jikime-migration-to-nextjs (score: 95.0)
   Legacy to Next.js 16 migration workflow specialist...
   Tags: migration, nextjs, react, vue, angular
   Keywords: migrate, migration, nextjs

2. jikime-framework-nextjs@16 (score: 95.0)
   Next.js 16 upgrade guide with breaking changes...
   Tags: framework, nextjs, version, use-cache
   Keywords: nextjs 16, next.js 16
```

### Find Related Skills (related)

Finds related skills based on shared tags, phases, agents, and languages.

```bash
# Find related skills
jikime-adk skill related jikime-lang-typescript

# Limit number of results
jikime-adk skill related jikime-platform-vercel --limit 5
```

### Skill Details (info)

```bash
# View metadata
jikime-adk skill info jikime-lang-typescript

# Include markdown body
jikime-adk skill info jikime-platform-vercel --body
```

### CLI Options Summary

| Command | Main Options | Description |
|---------|--------------|-------------|
| `list` | `--tag`, `--phase`, `--agent`, `--language`, `--format` | List and filter skills |
| `search` | `--tags`, `--phases`, `--agents`, `--languages`, `--limit` | Search skills |
| `related` | `--limit` | Find related skills |
| `info` | `--body` | Skill details |

---

## Skill Management Tools (Scripts)

### 1. Generate Skill Catalog

Scans metadata from all skills and generates a catalog.

```bash
python3 scripts/generate_skill_catalog.py
```

**Generated files:**
- `skills-catalog.yaml` - Machine-readable catalog
- `docs/skills-catalog.md` - Documentation catalog

**When should you run it?**

| Situation | Regeneration Required |
|-----------|----------------------|
| Adding a new skill | Yes |
| Modifying skill metadata (frontmatter) | Yes |
| Modifying only skill body | No |
| Deleting a skill | Yes |

### 2. Validate Skill Metadata

Validates that all skill frontmatter conforms to the schema.

```bash
# Validate all skills
python3 scripts/validate_skills.py

# Validate specific skill only
python3 scripts/validate_skills.py --skill jikime-marketing-seo

# Verbose output
python3 scripts/validate_skills.py --verbose
```

**Validation items:**
- Required fields: `name`, `description`, `version`
- Name pattern: `jikime-{domain}-{name}` format
- Version format: semver (framework skills are exceptions)
- Trigger settings: phases, keywords validity

### 3. Test Skills

Validates skill test examples and trigger settings.

```bash
# Test all skills
python3 scripts/test_skills.py

# Test specific skill only
python3 scripts/test_skills.py --skill jikime-marketing-seo

# Verbose output
python3 scripts/test_skills.py --verbose
```

## Adding a New Skill

### 1. Copy Template

```bash
cp -r templates/.claude/skills/_template templates/.claude/skills/jikime-{domain}-{name}
```

### 2. Write SKILL.md

Refer to `_template/SKILL.md` to write the frontmatter and body.

### 3. Validate

```bash
python3 scripts/validate_skills.py --skill jikime-{domain}-{name}
```

### 4. Update Catalog

```bash
python3 scripts/generate_skill_catalog.py
```

## Writing Skill Tests

### Test File Structure

```
skills/jikime-example/
├── SKILL.md
└── tests/
    └── examples.yaml
```

### examples.yaml Format

```yaml
# Skill test examples
name: jikime-example-skill
version: 1.0.0

# Trigger keywords (same as SKILL.md or defined for testing)
keywords:
  - example
  - example

# Test cases (test_N_name, test_N_input, test_N_expected format)
test_1_name: Basic test
test_1_input: Show me an example
test_1_expected: Example explanation, code sample

test_2_name: English input
test_2_input: Show me an example
test_2_expected: Example explanation

# Trigger validation
should_trigger:
  - Write example code
  - example usage

should_not_trigger:
  - Unrelated topic
  - Other skill keywords
```

### Test Validation Items

| Validation Type | Description |
|-----------------|-------------|
| Test structure | Whether both `test_N_input` and `test_N_expected` exist |
| Trigger check | Whether input triggers keywords |
| should_trigger | Whether inputs that should trigger actually trigger |
| should_not_trigger | Whether inputs that should not trigger do not trigger |

## Related Files

| File | Description |
|------|-------------|
| `scripts/generate_skill_catalog.py` | Catalog generation script |
| `scripts/validate_skills.py` | Metadata validation script |
| `scripts/test_skills.py` | Test execution script |
| `schemas/skill-frontmatter.schema.json` | Frontmatter JSON schema |
| `skills-catalog.yaml` | Generated catalog (YAML) |
| `docs/skills-catalog.md` | Generated catalog (Markdown) |

## Version Management Policy

### Semantic Versioning (SemVer)

Skills follow [Semantic Versioning](https://semver.org/).

```
MAJOR.MINOR.PATCH
```

| Version Change | When to increment? | Example |
|----------------|-------------------|---------|
| **MAJOR** | Breaking change (compatibility broken) | Major changes/deletions to trigger keywords, removal of required sections |
| **MINOR** | New feature added (backward compatible) | Adding new patterns, expanding examples, adding sections |
| **PATCH** | Bug fix, typo correction | Typo fixes, link fixes, clarifying descriptions |

### Exception: Framework Version Skills

`jikime-framework-*` skills use the target framework version in the version field.

```yaml
# jikime-framework-nextjs@16/SKILL.md
name: jikime-framework-nextjs@16
version: "16"  # Means Next.js version 16
```

### Version Update Guidelines

#### 1. PATCH Update (1.0.0 -> 1.0.1)

```yaml
# Before change
version: 1.0.0

# After change
version: 1.0.1
```

**Applicable cases:**
- Typo fixes
- Improving description text
- Fixing broken links
- Fixing code example errors

#### 2. MINOR Update (1.0.0 -> 1.1.0)

```yaml
# Before change
version: 1.0.0

# After change
version: 1.1.0
```

**Applicable cases:**
- Adding new patterns/examples
- Adding new sections
- Adding trigger keywords (keeping existing ones)
- Adding Works Well With skills

#### 3. MAJOR Update (1.0.0 -> 2.0.0)

```yaml
# Before change
version: 1.0.0

# After change
version: 2.0.0
```

**Applicable cases:**
- Major changes/deletions to trigger keywords
- Structural changes to required sections
- Changes to skill purpose/scope
- Merging/splitting with other skills

### Inter-Skill Dependencies

#### Works Well With Section

Relationships between skills are documented in the `## Works Well With` section.

```markdown
## Works Well With

- **jikime-lang-typescript**: TypeScript type definition patterns
- **jikime-platform-vercel**: Deployment optimization
```

**Notes:**
- Dependencies are **documentation references**, not automatically loaded
- Required skills must be explicitly loaded with `Skill("skill-name")`
- Be careful of circular dependencies (A -> B -> A)

#### Dependency Management Principles

| Principle | Description |
|-----------|-------------|
| **Explicit loading** | Load only required skills explicitly (token efficiency) |
| **Loose coupling** | Design to work independently without other skills |
| **Documentation** | Related skills are specified in Works Well With |

### Progressive Disclosure Token Guidelines

| Token Range | Recommended Use |
|-------------|-----------------|
| ~2000 | General skills (lang, library) |
| ~3000-5000 | Complex skills (foundation, workflow) |
| ~5000+ | Large skills (consider splitting) |

**Recommendations:**
- Keep to ~2000-3000 tokens if possible
- Consider skill splitting if 5000+ tokens
- Quick Reference section should contain only essential content

## Reference

- All scripts use only Python standard library (no external dependencies)
- Skill catalog currently includes 59 skills
- Progressive Disclosure improves token efficiency by 67%+
