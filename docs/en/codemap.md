# Codemap & Cleanup Commands

Command reference for architecture documentation and code cleanup.

## Overview

| Command | Purpose | Main Tools |
|---------|---------|------------|
| `/jikime:codemap` | AST analysis-based architecture documentation | ts-morph, madge |
| `/jikime:cleanup` | Dead code detection and safe removal | knip, depcheck, ts-prune |

---

## /jikime:codemap

**AST Analysis-Based Architecture Mapping**

| Item | Description |
|------|-------------|
| **Description** | Automatically generates architecture documentation from codebase |
| **Type** | Utility (Type B) |
| **Context** | sync.md |
| **Skill** | jikime-workflow-codemap |
| **Standalone Use** | ✅ High - Can be run independently anytime |

### Usage

```bash
# Generate full architecture map
/jikime:codemap all

# Generate specific areas only
/jikime:codemap frontend
/jikime:codemap backend
/jikime:codemap database
/jikime:codemap integrations

# Include AST analysis (TypeScript/JavaScript)
/jikime:codemap all --ast

# Generate dependency graph
/jikime:codemap all --deps

# Force regeneration
/jikime:codemap all --refresh

# JSON output (for automation)
/jikime:codemap all --json
```

### Options

| Option | Description |
|--------|-------------|
| `all` | Generate codemap for all areas |
| `frontend` | Frontend architecture only |
| `backend` | Backend/API architecture only |
| `database` | Database schema/models |
| `integrations` | External service integrations |
| `--ast` | Enable AST analysis with ts-morph |
| `--deps` | Generate dependency graph with madge |
| `--refresh` | Force regeneration ignoring cache |
| `--json` | JSON output for automation |

### Analysis Tools

#### 1. ts-morph (AST Analysis)

Structural analysis for TypeScript/JavaScript projects:

```typescript
// Extracted information
- All exported functions, classes, types
- import/export relationships
- Module dependencies
- Route definitions (Next.js, Express, etc.)
```

#### 2. madge (Dependency Graph)

```bash
# Generate SVG graph
npx madge --image docs/CODEMAPS/assets/dependency-graph.svg src/

# Detect circular dependencies
npx madge --circular src/
```

### Framework Detection

| Detection File | Framework | Codemap Focus |
|----------------|-----------|---------------|
| `next.config.*` | Next.js | App Router, API Routes, Pages |
| `vite.config.*` | Vite | Components, Modules |
| `angular.json` | Angular | Modules, Services, Components |
| `nuxt.config.*` | Nuxt | Pages, Plugins, Modules |
| `package.json` + express | Express | Routes, Middleware |
| `go.mod` | Go | Packages, Handlers |
| `Cargo.toml` | Rust | Crates, Modules |
| `pyproject.toml` | Python | Packages, Modules |

### Output Structure

```
docs/
├── CODEMAPS/
│   ├── INDEX.md              # Architecture overview
│   ├── frontend.md           # Frontend structure
│   ├── backend.md            # Backend/API structure
│   ├── database.md           # Database schema
│   ├── integrations.md       # External services
│   └── assets/
│       ├── dependency-graph.svg
│       └── architecture-diagram.svg
```

### Codemap File Format

```markdown
# [Area] Codemap

**Last Updated:** YYYY-MM-DD
**Version:** X.Y.Z
**Entry Points:** [List of main entry points]

## Overview
[Brief description of the area]

## Architecture
[ASCII diagram showing component relationships]

## Key Modules

| Module | Purpose | Exports | Dependencies |
|--------|---------|---------|--------------||
| ... | ... | ... | ... |

## Data Flow
[Description of data flow through this area]

## External Dependencies
- package@version - purpose
- ...

## Related Codemaps
- [Related Area](./related.md)
```

### Process

```
Phase 1: Discovery
    ↓
  Detect framework, language, project type
    ↓
  Identify entry points and core files
    ↓
Phase 2: Analysis
    ↓
  AST parsing (ts-morph for TS/JS)
    ↓
  Dependency graph (madge)
    ↓
  Pattern recognition (MVC, Clean, etc.)
    ↓
Phase 3: Generation
    ↓
  Generate structured codemap
    ↓
  Create ASCII diagrams
    ↓
  Build relationship tables
    ↓
Phase 4: Validation
    ↓
  Verify path existence
    ↓
  Validate link targets
    ↓
  Report coverage statistics
```

---

## /jikime:cleanup

**Dead Code Detection and Safe Removal**

| Item | Description |
|------|-------------|
| **Description** | Comprehensive dead code analysis with DELETION_LOG tracking |
| **Type** | Utility (Type B) |
| **Context** | dev.md |
| **Agent** | refactorer |
| **Standalone Use** | ✅ High - Can be run independently anytime |

### Usage

```bash
# Dead code scan (analysis only, no changes)
/jikime:cleanup scan

# Remove SAFE items only (low risk)
/jikime:cleanup remove --safe

# Include CAREFUL items (medium risk, confirmation required)
/jikime:cleanup remove --careful

# Target specific categories
/jikime:cleanup remove --deps      # Unused npm dependencies
/jikime:cleanup remove --exports   # Unused exports
/jikime:cleanup remove --files     # Unused files

# Dry run (shows what would be removed)
/jikime:cleanup scan --dry-run

# Check deletion history
/jikime:cleanup log

# Generate comprehensive cleanup report
/jikime:cleanup report
```

### Options

| Option | Description |
|--------|-------------|
| `scan` | Analyze codebase for dead code (no changes) |
| `remove` | Remove detected dead code |
| `report` | Generate comprehensive cleanup report |
| `log` | Check DELETION_LOG.md history |
| `--safe` | Remove low-risk items only |
| `--careful` | Include medium-risk items (verification required) |
| `--deps` | Target unused dependencies |
| `--exports` | Target unused exports |
| `--files` | Target unused files |
| `--dry-run` | Show what would be removed |

### Analysis Tools

#### 1. knip - Comprehensive Dead Code Detection

```bash
# Installation
npm install -D knip

# Full analysis
npx knip

# JSON report
npx knip --reporter json > .jikime/cleanup/knip-report.json
```

**Detection Items**:
- Unused files
- Unused exports
- Unused dependencies
- Unused devDependencies
- Unused types

#### 2. depcheck - Dependency Analysis

```bash
# Installation
npm install -D depcheck

# Analysis
npx depcheck

# JSON report
npx depcheck --json > .jikime/cleanup/depcheck-report.json
```

**Detection Items**:
- Unused dependencies
- Missing dependencies
- Phantom dependencies

#### 3. ts-prune - TypeScript Export Analysis

```bash
# Installation
npm install -D ts-prune

# Analysis
npx ts-prune

# Filtering
npx ts-prune | grep -v "used in module"
```

**Detection Items**:
- Unused exports
- Unused types
- Dead code paths

#### 4. ESLint - Unused Directives

```bash
# Check unused eslint-disable comments
npx eslint . --report-unused-disable-directives
```

### Risk Classification System

#### SAFE (Automatically Removable)

| Category | Risk | Detection Method | Verification |
|----------|------|------------------|--------------|
| Unused npm deps | Low | depcheck | No imports confirmed |
| Unused devDeps | Low | depcheck | Not used in scripts |
| Commented code | Low | Regex patterns | Visual confirmation |
| Unused imports | Low | ESLint + knip | No references |
| Unused eslint-disable | Low | ESLint report | Directive check |

#### CAREFUL (Confirmation Required)

| Category | Risk | Detection Method | Verification |
|----------|------|------------------|--------------|
| Unused exports | Medium | ts-prune + knip | Grep + git history |
| Unused files | Medium | knip | Check dynamic imports |
| Unused types | Medium | ts-prune | Check type inference |
| Dead branches | Medium | Coverage report | Runtime tests |

#### RISKY (Manual Review Required)

| Category | Risk | Detection Method | Verification |
|----------|------|------------------|--------------|
| Public API | High | API tests | Integration tests |
| Shared utilities | High | Cross-project search | Stakeholder review |
| Dynamic imports | High | String pattern search | Runtime tests |
| Reflection code | High | Pattern analysis | Full test suite |

### DDD-Aligned Workflow

```
Phase 1: ANALYZE
    └─ Run all detection tools in parallel
    └─ Aggregate results with risk classification
    └─ Check test coverage for affected code
    └─ Review git history for context
         ↓
Phase 2: PRESERVE
    └─ Ensure characterization tests exist for affected code
    └─ Create backup branch: cleanup/YYYY-MM-DD-HHMM
    └─ Document current behavior if no tests exist
         ↓
Phase 3: IMPROVE
    └─ Remove by category (safest first):
        a. Unused npm dependencies
        b. Unused devDependencies
        c. Unused imports
        d. Unused exports
        e. Unused files
    └─ After each category:
        - Run build
        - Run full test suite
        - Commit if passing
        - Update DELETION_LOG.md
```

### DELETION_LOG.md Format

All deletions are tracked in `docs/DELETION_LOG.md`:

```markdown
# Code Deletion Log

Audit trail for code cleanup operations.

---

## [YYYY-MM-DD HH:MM] Cleanup Session

**Operator**: J.A.R.V.I.S. / refactorer agent
**Branch**: cleanup/YYYY-MM-DD-HHMM
**Commit**: abc123def
**Tools**: knip v5.x, depcheck v1.x, ts-prune v0.x

### Summary

| Category | Items | Lines | Size Impact |
|----------|-------|-------|-------------|
| Dependencies | 5 | - | -120 KB |
| DevDependencies | 3 | - | -45 KB |
| Files | 12 | 1,450 | -45 KB |
| Exports | 23 | 89 | - |
| Imports | 45 | 45 | - |
| **Total** | **88** | **1,584** | **-210 KB** |

### Dependencies Removed

| Package | Version | Reason | Alternative |
|---------|---------|--------|-------------|
| lodash | 4.17.21 | Not imported | Use native methods |
| moment | 2.29.4 | Deprecated | date-fns already used |

### Files Deleted

| Path | Lines | Last Modified | Replaced By |
|------|-------|---------------|-------------|
| src/utils/old-helpers.ts | 120 | 2023-08-15 | N/A (unused) |
| src/components/LegacyButton.tsx | 85 | 2023-09-01 | Button.tsx |

### Verification Results

- [x] TypeScript compiles: `npx tsc --noEmit`
- [x] Build succeeds: `npm run build`
- [x] Tests pass: 47/47 (100%)
- [x] No lint errors: `npm run lint`
- [x] Bundle size verified

### Recovery Instructions

```bash
# If issues occur after this cleanup:
git log --oneline | head -5  # Find cleanup commit
git revert <commit-sha>       # Revert specific commit
npm install                   # Reinstall dependencies
npm run build && npm test     # Verify recovery
```
```

### Protected Items

Manage items that should not be removed in `.jikime/cleanup/protected.yaml`:

```yaml
# Items that should not be removed
protected:
  dependencies:
    - "@types/*"  # Type definitions
    - "eslint-*"  # Linting infrastructure

  files:
    - "src/polyfills/*"  # Browser compatibility
    - "src/lib/dynamic-*"  # Dynamic import targets

  exports:
    - "src/api/public.ts:*"  # Public API
    - "src/sdk/index.ts:*"   # SDK exports

  patterns:
    - "**/index.ts"  # Barrel files (may appear unused)
    - "**/__tests__/*"  # Test utilities
```

### Safety Checklist

**Pre-removal Verification**:
- [ ] All detection tools run
- [ ] Risk classification complete
- [ ] Backup branch created
- [ ] Characterization tests exist (or created)
- [ ] Git history reviewed for context
- [ ] Dynamic import patterns checked
- [ ] Public API impact assessed

**Post-removal Verification**:
- [ ] TypeScript compiles without errors
- [ ] Build succeeds
- [ ] All tests pass
- [ ] No console errors
- [ ] Bundle size measured
- [ ] DELETION_LOG.md updated
- [ ] Commit message detailed

---

## TRUST 5 Integration

| Principle | Codemap | Cleanup |
|-----------|---------|---------|
| **T**ested | Verify generated document paths | Run tests after each removal |
| **R**eadable | Clear structure with ASCII diagrams | Remove noise, improve signal/noise ratio |
| **U**nified | Consistent document format | Consolidate duplicates |
| **S**ecured | Prevent sensitive information exposure | Remove unused deps with vulnerabilities |
| **T**rackable | Timestamp and version control | DELETION_LOG.md audit trail |

---

## J.A.R.V.I.S. / F.R.I.D.A.Y. Output Format

### J.A.R.V.I.S. (Development)

```markdown
## J.A.R.V.I.S.: Codemap Generation Complete

### Generated Files
| File | Lines | Modules Documented |
|------|-------|-------------------|
| docs/CODEMAPS/INDEX.md | 120 | 5 entry points |
| docs/CODEMAPS/frontend.md | 85 | 12 components |
| docs/CODEMAPS/backend.md | 95 | 8 endpoints |

### Coverage
- Files analyzed: 47
- Modules documented: 25
- Dependencies mapped: 32
- Circular dependencies: 0

### Predictive Suggestions
- Consider documenting workers/ directory
- API rate limiting not documented
```

```markdown
## J.A.R.V.I.S.: Cleanup Scan Complete

### Dead Code Summary
| Category | Found | Risk | Action |
|----------|-------|------|--------|
| Dependencies | 5 | SAFE | Auto-remove |
| Exports | 23 | CAREFUL | Review |
| Files | 12 | CAREFUL | Review |
| Dynamic refs | 2 | RISKY | Skip |

### Recommended Actions

**Immediate (SAFE)**:
1. Remove 5 unused dependencies (-120 KB)
2. Remove 15 unused imports

**Review Required (CAREFUL)**:
1. 12 files appear unused but check git history
2. 23 exports not directly referenced

### Estimated Impact
- Bundle size: -165 KB (~8% reduction)
- Lines of code: -1,539
- Files: -12

Proceed with --safe removal? Use: /jikime:cleanup remove --safe
```

### F.R.I.D.A.Y. (Migration)

```markdown
## F.R.I.D.A.Y.: Migration Cleanup

### Legacy Code Status
| Module | Dead Code | Migrated | Clean |
|--------|-----------|----------|-------|
| Auth | 5 items | Yes | No |
| Users | 0 items | Yes | Yes |
| Products | 12 items | Yes | No |

### Migration-Safe Removal
Only removing code that has been:
- Fully migrated to target framework
- Verified by characterization tests
- Not referenced in migration artifacts
```

---

## Command Comparison

| Situation | Recommended Command |
|-----------|---------------------|
| Need architecture documentation | `/jikime:codemap all` |
| Visualize dependency graph | `/jikime:codemap all --deps` |
| Assess dead code status | `/jikime:cleanup scan` |
| Safe cleanup operation | `/jikime:cleanup remove --safe` |
| Check cleanup history | `/jikime:cleanup log` |
| Cleanup before refactoring | `/jikime:cleanup scan` → `/jikime:refactor` |

---

## Related Commands

- `/jikime:docs` - Documentation update and synchronization
- `/jikime:refactor` - DDD-based code refactoring
- `/jikime:3-sync` - SPEC completion and documentation sync
- `/jikime:learn` - Codebase exploration and learning

---

## Related Skills

- `jikime-workflow-codemap` - Codemap generation workflow
- `jikime-workflow-ddd` - DDD methodology (ANALYZE-PRESERVE-IMPROVE)
- `jikime-foundation-quality` - TRUST 5 quality framework

---

Version: 1.0.0
Last Updated: 2026-01-25
Integration: AST analysis (ts-morph), Dependency graphs (madge), Dead code detection (knip, depcheck, ts-prune)
