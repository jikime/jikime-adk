# JikiME-ADK Migration System

## Overview

JikiME-ADK Migration System provides intelligent framework migration and version upgrade capabilities. The system leverages Mr.Jikime orchestrator to automatically discover and execute appropriate skills based on source and target specifications.

## Command Structure

```bash
# Framework migration
jikime migration --source react --target nextjs

# Version upgrade (source == target)
jikime migration --source nextjs@15 --target nextjs@16

# Specific domain focus
jikime migration --source react --target nextjs --focus auth,api
```

## Architecture

### Three-Layer Skills Structure

| Layer | Role | Example |
|-------|------|---------|
| **L1: Migration Skills** | Framework transition strategies | `jikime-migrate-react-to-nextjs` |
| **L2: Version Skills** | Version-specific guides | `jikime-nextjs@16`, `jikime-react@19` |
| **L3: Domain Skills** | Domain-specific patterns | `jikime-migration-patterns-auth` |

### Skills Naming Convention

```
Migration:      jikime-migrate-{source}-to-{target}
Version Guide:  jikime-{framework}@{version}
Domain Pattern: jikime-migration-patterns-{domain}
```

**Examples:**
- `jikime-migrate-react-to-nextjs` - React to Next.js transition
- `jikime-migrate-vue-to-nuxt` - Vue to Nuxt transition
- `jikime-nextjs@16` - Next.js 16 specific guide
- `jikime-migration-patterns-auth` - Authentication migration patterns

## Skills Metadata

Skills use frontmatter metadata for automatic discovery:

```yaml
---
name: jikime-migrate-react-to-nextjs
type: migration
source: react
target: nextjs
supported_versions:
  source: ["17", "18", "19"]
  target: ["14", "15", "16"]
domains: [routing, state, api, auth, ssr]
requires_mcp: [context7, webfetch]
---
```

### Metadata Fields

| Field | Description |
|-------|-------------|
| `type` | Skill type: `migration`, `version`, `domain` |
| `source` | Source framework/library |
| `target` | Target framework/library |
| `supported_versions` | Compatible version ranges |
| `domains` | Covered migration domains |
| `requires_mcp` | Required MCP servers |

## Orchestration Flow

```
User Command
    │
    ▼
┌─────────────────────────────────────┐
│ Mr.Jikime Orchestrator              │
│  1. Parse source/target             │
│  2. Detect upgrade mode             │
│  3. Search matching skills          │
│  4. Identify domain skills          │
│  5. Fetch latest info (MCP)         │
│  6. Create execution plan           │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│ Execution Phases                    │
│  Phase 1: Analyze (current code)    │
│  Phase 2: Plan (migration strategy) │
│  Phase 3: Execute (transformations) │
│  Phase 4: Validate (tests, build)   │
└─────────────────────────────────────┘
```

## Version Upgrade Mode

When `source == target`, the system enters version upgrade mode:

```bash
jikime migration --source nextjs@15 --target nextjs@16
```

The orchestrator automatically:
1. Loads `jikime-nextjs@15` (current version understanding)
2. Loads `jikime-nextjs@16` (target version guide)
3. Analyzes breaking changes
4. References official migration guide via Context7

## MCP Integration

| MCP Server | Purpose |
|------------|---------|
| **Context7** | Official docs, migration guides, API changes |
| **WebFetch** | Latest release notes, breaking changes |
| **Sequential** | Complex migration analysis |

Skills specify required MCP servers in `requires_mcp` metadata for automatic activation.

## Migration Domains

| Domain | Description |
|--------|-------------|
| `routing` | Route structure, navigation patterns |
| `state` | State management migration |
| `api` | API integration, data fetching |
| `auth` | Authentication/authorization |
| `ssr` | Server-side rendering |
| `styling` | CSS, styling solutions |
| `testing` | Test migration |

## Best Practices

### For Skill Authors

1. **Use metadata** - Always include complete frontmatter for discovery
2. **Reference latest docs** - Include Context7/WebFetch instructions
3. **Cover breaking changes** - Document known issues and solutions
4. **Provide examples** - Include before/after code examples
5. **Version specificity** - Be explicit about version compatibility

### For Users

1. **Analyze first** - Run with `--dry-run` to see the plan
2. **Focus domains** - Use `--focus` for incremental migration
3. **Backup code** - Ensure git commits before migration
4. **Test incrementally** - Validate after each phase

## Supported Migrations

### Framework Migrations (Planned)

| Source | Target | Skill |
|--------|--------|-------|
| React | Next.js | `jikime-migrate-react-to-nextjs` |
| Vue | Nuxt | `jikime-migrate-vue-to-nuxt` |
| Angular | - | TBD |
| Svelte | SvelteKit | `jikime-migrate-svelte-to-sveltekit` |

### Version Upgrades (Planned)

| Framework | Versions | Skill |
|-----------|----------|-------|
| Next.js | 14 → 15 → 16 | `jikime-nextjs@{version}` |
| React | 17 → 18 → 19 | `jikime-react@{version}` |
| Vue | 2 → 3 | `jikime-vue@{version}` |

## Implementation Notes

### Considerations

- **Granularity balance** - Too fine-grained skills increase maintenance burden
- **Version explosion** - Limit to major versions only
- **Progressive Disclosure** - Load skills on-demand

### Recommendations

- L1 Migration Skills: Keep generic and comprehensive
- L2 Version Skills: Major versions only (14, 15, 16)
- L3 Domain Skills: Create only for significant breaking changes

---

Version: 1.0.0
Last Updated: 2025-01-22
