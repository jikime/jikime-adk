---
name: jikime-workflow-changelog
description: Automated changelog generation from Git commits. Transforms commit history into user-friendly release notes with proper categorization, breaking change detection, and multiple output formats.
version: 1.0.0
category: workflow

# Progressive Disclosure Configuration
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~3500

# Trigger Conditions for Level 2 Loading
triggers:
  keywords: ["changelog", "release notes", "version history", "what's new", "release", "CHANGELOG.md"]
  phases: ["sync"]
  agents: ["manager-git", "documenter", "scribe"]
  languages: []
---

# Changelog Generator

Transform Git commits into user-friendly changelogs automatically.

## Quick Reference

### Commit Categories

| Prefix | Category | User-Facing Label |
|--------|----------|-------------------|
| `feat:` | Features | ✨ New Features |
| `fix:` | Bug Fixes | 🐛 Bug Fixes |
| `perf:` | Performance | ⚡ Performance |
| `docs:` | Documentation | 📚 Documentation |
| `refactor:` | Improvements | 🔧 Improvements |
| `BREAKING CHANGE:` | Breaking | 💥 Breaking Changes |
| `security:` | Security | 🔒 Security |
| `deprecate:` | Deprecated | ⚠️ Deprecated |

### Quick Commands

```bash
# Generate changelog for latest tag
/jikime:changelog

# Generate for specific version range
/jikime:changelog --from v1.0.0 --to v2.0.0

# Include all commits (not just conventional)
/jikime:changelog --include-all
```

---

## Implementation Guide

### Step 1: Analyze Commits

Extract commits between versions:

```bash
# Get commits since last tag
git log $(git describe --tags --abbrev=0)..HEAD --oneline

# Get commits between two tags
git log v1.0.0..v2.0.0 --pretty=format:"%h|%s|%an|%ad" --date=short
```

### Step 2: Parse Conventional Commits

Parse commit messages following Conventional Commits spec:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Parsing Pattern:**
```javascript
const conventionalCommitRegex = /^(\w+)(?:\(([^)]+)\))?(!)?:\s*(.+)$/;
// Groups: [type, scope, breaking, description]
```

### Step 3: Categorize Changes

Group commits by category with priority ordering:

```
1. 💥 Breaking Changes (highest priority)
2. 🔒 Security
3. ✨ New Features
4. 🐛 Bug Fixes
5. ⚡ Performance
6. 🔧 Improvements
7. 📚 Documentation
8. ⚠️ Deprecated
```

### Step 4: Generate Output

**Markdown Format (CHANGELOG.md):**

```markdown
# Changelog

## [2.0.0] - 2026-02-06

### 💥 Breaking Changes

- **auth**: Remove legacy session-based authentication
  - Migration: Use JWT tokens instead of session cookies

### ✨ New Features

- **api**: Add rate limiting to all endpoints (#123)
- **ui**: New dark mode theme support

### 🐛 Bug Fixes

- **forms**: Fix validation not triggering on blur (#456)
- **export**: Resolve memory leak in large file exports

### 🔧 Improvements

- **deps**: Update React to v19
- **build**: Improve build time by 40%
```

**JSON Format (for programmatic use):**

```json
{
  "version": "2.0.0",
  "date": "2026-02-06",
  "sections": {
    "breaking": [
      {
        "scope": "auth",
        "message": "Remove legacy session-based authentication",
        "hash": "abc1234",
        "migration": "Use JWT tokens instead of session cookies"
      }
    ],
    "features": [...],
    "fixes": [...]
  }
}
```

---

## Workflow Integration

### Pre-Release Workflow

```
1. Ensure all commits follow Conventional Commits
2. Run changelog generation
3. Review and edit if needed
4. Commit CHANGELOG.md update
5. Create version tag
6. Push tag and trigger release
```

### Git Hooks Integration

Add to `.husky/commit-msg`:

```bash
#!/bin/sh
npx commitlint --edit $1
```

**commitlint.config.js:**
```javascript
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [2, 'always', [
      'feat', 'fix', 'docs', 'style', 'refactor',
      'perf', 'test', 'build', 'ci', 'chore', 'revert'
    ]]
  }
};
```

### CI/CD Integration

**GitHub Actions:**

```yaml
name: Generate Changelog

on:
  push:
    tags:
      - 'v*'

jobs:
  changelog:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Generate Changelog
        run: |
          # Your changelog generation script

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          body_path: RELEASE_NOTES.md
```

---

## Advanced Patterns

### Breaking Change Detection

Detect breaking changes from:

1. `!` after type: `feat!: new API`
2. Footer: `BREAKING CHANGE: description`
3. Body contains "BREAKING"

```javascript
function isBreaking(commit) {
  return commit.type.endsWith('!') ||
         commit.footer?.includes('BREAKING CHANGE') ||
         commit.body?.includes('BREAKING');
}
```

### Scope Grouping

Group by scope for large projects:

```markdown
## [2.0.0] - 2026-02-06

### @myorg/api

#### ✨ Features
- Add GraphQL subscriptions

#### 🐛 Fixes
- Fix connection pooling

### @myorg/ui

#### ✨ Features
- New Button variants
```

### PR/Issue Linking

Extract references and create links:

```javascript
const issueRegex = /#(\d+)/g;
const prRegex = /\(#(\d+)\)/g;

function addLinks(message, repoUrl) {
  return message.replace(
    issueRegex,
    `[#$1](${repoUrl}/issues/$1)`
  );
}
```

### Multi-Language Support

Generate localized changelogs:

```markdown
<!-- CHANGELOG.md (English) -->
## [2.0.0] - 2026-02-06

### ✨ New Features
- Add dark mode support

<!-- CHANGELOG.ko.md (Korean) -->
## [2.0.0] - 2026-02-06

### ✨ 새로운 기능
- 다크 모드 지원 추가
```

---

## Configuration Options

### changelog.config.js

```javascript
module.exports = {
  // Output file
  output: 'CHANGELOG.md',

  // Include types (others go to "Other Changes")
  types: {
    feat: '✨ New Features',
    fix: '🐛 Bug Fixes',
    perf: '⚡ Performance',
    docs: '📚 Documentation',
    refactor: '🔧 Improvements',
    security: '🔒 Security',
    deprecate: '⚠️ Deprecated'
  },

  // Excluded types (won't appear in changelog)
  exclude: ['chore', 'ci', 'test', 'style'],

  // Scope aliases
  scopeAlias: {
    'ui': 'User Interface',
    'api': 'API',
    'db': 'Database'
  },

  // Repository URL for links
  repoUrl: 'https://github.com/org/repo',

  // Include commit hash
  includeHash: true,

  // Include author
  includeAuthor: false,

  // Date format
  dateFormat: 'YYYY-MM-DD'
};
```

---

## Works Well With

| Skill | Integration |
|-------|-------------|
| `jikime-workflow-git` | Git operations and tagging |
| `sc:git` | Commit message generation |
| `jikime-workflow-sync` | Documentation sync workflow |

---

## Tool Reference

### Popular Changelog Tools

| Tool | Language | Features |
|------|----------|----------|
| `conventional-changelog` | Node.js | Standard, customizable |
| `git-cliff` | Rust | Fast, highly configurable |
| `release-please` | Node.js | GitHub Action integration |
| `semantic-release` | Node.js | Full automation |

### Manual Generation Script

```bash
#!/bin/bash
# generate-changelog.sh

VERSION=$1
PREV_TAG=$(git describe --tags --abbrev=0 HEAD^)

echo "## [$VERSION] - $(date +%Y-%m-%d)"
echo ""

# Features
echo "### ✨ New Features"
git log $PREV_TAG..HEAD --oneline --grep="^feat" | sed 's/^[a-f0-9]* /- /'
echo ""

# Fixes
echo "### 🐛 Bug Fixes"
git log $PREV_TAG..HEAD --oneline --grep="^fix" | sed 's/^[a-f0-9]* /- /'
```

---

Version: 1.0.0
Source: Conventional Commits + Keep a Changelog standards
