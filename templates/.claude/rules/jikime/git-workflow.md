# Git Workflow Rules

Git conventions and workflow guidelines for consistent version control.

## Branching Strategies

### GitHub Flow (Recommended for most projects)

```
main ──●────●────●────●────●── (always deployable)
        \          /
feature  └──●──●──┘
```

- `main` is always deployable
- Feature branches from main
- PR + review + merge
- Deploy after merge

### Git Flow (For release-based projects)

```
main     ──●─────────────●────── (releases only)
            \           /
release      └────●────┘
                 /
develop  ──●──●────●──●──●──
            \     /
feature      └──●┘
```

Use Git Flow when:
- Multiple releases in production
- Formal release cycle required
- Hotfix branches needed

## Commit Message Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Commit Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code refactoring (no behavior change) |
| `docs` | Documentation changes |
| `test` | Adding or updating tests |
| `chore` | Maintenance tasks |
| `perf` | Performance improvements |
| `ci` | CI/CD changes |
| `style` | Formatting, no logic change |

### Examples

```bash
# Good - Simple
feat: add user authentication with JWT
fix: resolve memory leak in WebSocket handler
refactor: extract validation logic to separate module
docs: update API documentation for v2 endpoints

# Good - With scope and body
feat(auth): add OAuth2 login support

Implements Google and GitHub OAuth providers.
Closes #123

BREAKING CHANGE: Session tokens now expire after 24h

# Good - Bug fix with context
fix(api): handle null response from payment gateway

Previously caused 500 error when gateway returned null.
Now returns appropriate error message to user.

# Bad
update code
fixed bug
WIP
asdf
```

## Pull Request Workflow

When creating PRs:

1. **Analyze full commit history** (not just latest commit)
2. Use `git diff [base-branch]...HEAD` to see all changes
3. Draft comprehensive PR summary
4. Include test plan with TODOs
5. Push with `-u` flag if new branch

### PR Size Guidelines

| Size | Lines Changed | Review Guidance |
|------|---------------|-----------------|
| **XS** | < 50 | Quick review, minimal context needed |
| **S** | 50-200 | Standard review, single reviewer |
| **M** | 200-500 | Thorough review, allocate focused time |
| **L** | 500+ | **Split if possible**, multiple reviewers |

**Best Practice**: Aim for XS-S PRs. Large PRs are harder to review and more likely to contain bugs.

### PR Description Template

```markdown
## Summary
- Brief description of changes
- Key features or fixes

## Changes
- [Change 1]
- [Change 2]

## Test Plan
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Screenshots (if UI changes)
[Before/After screenshots]

## Related Issues
- Closes #123
```

### PR Checklist

```
- [ ] Code follows project conventions
- [ ] Tests added/updated for changes
- [ ] All tests pass locally
- [ ] No merge conflicts with main
- [ ] Documentation updated if needed
- [ ] No security vulnerabilities introduced
- [ ] PR description explains the "why"
```

## Feature Implementation Workflow

### Feature Development Progress

Copy this checklist when starting a new feature:

```
Feature Development Progress:
- [ ] Step 1: Create feature branch from main
- [ ] Step 2: Make changes with atomic commits
- [ ] Step 3: Write/update tests
- [ ] Step 4: Rebase on latest main
- [ ] Step 5: Push and create PR
- [ ] Step 6: Address review feedback
- [ ] Step 7: Merge after approval
```

### 1. Plan First

Use **manager-spec** or **manager-strategy** subagent:
- Create implementation plan
- Identify dependencies and risks
- Break down into phases

### 2. DDD Approach (JikiME Standard)

Use **manager-ddd** subagent:
- **ANALYZE**: Understand existing behavior
- **PRESERVE**: Write characterization tests
- **IMPROVE**: Implement with test validation

### 3. Code Review

Use **manager-quality** subagent:
- Address CRITICAL and HIGH issues immediately
- Fix MEDIUM issues when possible
- Document LOW issues for future

### 4. Commit & Push

- Detailed commit messages
- Follow conventional commits format
- Reference related issues

### Commit Validation

Before pushing, validate commits:

```
Commit Validation:
- [ ] Each commit has a clear, descriptive message
- [ ] Commit type matches the change (feat, fix, etc.)
- [ ] No WIP or temporary commits
- [ ] No secrets or credentials committed
- [ ] Changes are atomic (one logical change per commit)
```

If validation fails, use `git rebase -i` to clean up commit history before pushing.

## Branch Naming

```
<type>/<description>
```

| Type | Usage |
|------|-------|
| `feature/` | New features |
| `fix/` | Bug fixes |
| `refactor/` | Code refactoring |
| `docs/` | Documentation |
| `chore/` | Maintenance |

### Examples

```
feature/user-authentication
feature/AUTH-123-oauth-login
fix/memory-leak-websocket
fix/BUG-456-null-pointer
refactor/validation-module
chore/TECH-789-upgrade-deps
docs/api-v2-endpoints
```

## Common Git Commands

### Daily Workflow

```bash
# Start new feature
git checkout main
git pull
git checkout -b feature/TICKET-123-description

# Commit changes
git add -p  # Stage interactively (review each change)
git commit -m "feat: description"

# Keep up with main
git fetch origin main
git rebase origin/main

# Push and create PR
git push -u origin HEAD
```

### Fixing Mistakes

```bash
# Amend last commit (before push)
git commit --amend

# Undo last commit (keep changes staged)
git reset --soft HEAD~1

# Undo last commit (keep changes unstaged)
git reset HEAD~1

# Undo last commit (discard changes) - DANGEROUS
git reset --hard HEAD~1

# Revert a pushed commit (safe for shared branches)
git revert <commit-hash>

# Interactive rebase to clean up commits
git rebase -i HEAD~3
```

### Advanced Operations

```bash
# Cherry-pick specific commit to current branch
git cherry-pick <commit-hash>

# Find which commit broke something (binary search)
git bisect start
git bisect bad HEAD
git bisect good <known-good-commit>
# Git will checkout commits, test and mark:
git bisect good  # or
git bisect bad
# When done:
git bisect reset

# Stash work in progress
git stash push -m "WIP: feature description"
git stash list
git stash pop  # Apply and remove
git stash apply  # Apply and keep

# View stash contents
git stash show -p stash@{0}
```

### Conflict Resolution

```bash
# During rebase, if conflicts occur:
# 1. Fix conflicts in files
# 2. Stage resolved files
git add <resolved-files>
# 3. Continue rebase
git rebase --continue

# Abort if needed
git rebase --abort

# After merge with conflicts:
# 1. Fix conflicts
# 2. Complete merge
git add <resolved-files>
git commit
```

## Git Checklist

Before pushing:

- [ ] Commit messages follow conventional format
- [ ] Branch name is descriptive
- [ ] No sensitive data in commits (secrets, API keys)
- [ ] Tests pass locally
- [ ] Code is properly formatted

Before merging PR:

- [ ] All CI checks pass
- [ ] Code review completed
- [ ] Conflicts resolved
- [ ] Documentation updated if needed

## Prohibited Practices

| Practice | Reason |
|----------|--------|
| Force push to main/master | Destroys history |
| Committing secrets | Security risk |
| Large monolithic commits | Hard to review/revert |
| Merge commits in feature branches | Clutters history |
| Committing build artifacts | Bloats repository |

---

Version: 2.0.0
Last Updated: 2026-01-25
Source: JikiME-ADK git-workflow rules (Enhanced with branching strategies, advanced commands, and PR guidelines)
