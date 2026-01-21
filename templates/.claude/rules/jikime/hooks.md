# Hooks System

Claude Code hooks for automated workflows and quality enforcement.

## Hook Types

| Type | When | Purpose |
|------|------|---------|
| **PreToolUse** | Before tool execution | Validation, modification, blocking |
| **PostToolUse** | After tool execution | Auto-format, checks, logging |
| **Notification** | On specific events | Alerts, status updates |
| **Stop** | Session ends | Final verification |

## Recommended Hooks

### PreToolUse Hooks

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "command": "echo 'Long command detected' && exit 1",
        "condition": "npm|pnpm|yarn|cargo build|cargo test",
        "message": "Consider using tmux for long-running commands"
      },
      {
        "matcher": "Write",
        "condition": "\\.md$|\\.txt$",
        "command": "echo 'Blocked' && exit 1",
        "message": "Do not create documentation files unless explicitly requested"
      }
    ]
  }
}
```

### PostToolUse Hooks

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "condition": "\\.(ts|tsx|js|jsx)$",
        "command": "npx prettier --write $FILE"
      },
      {
        "matcher": "Edit|Write",
        "condition": "\\.(ts|tsx)$",
        "command": "npx tsc --noEmit $FILE 2>&1 | head -20"
      },
      {
        "matcher": "Edit|Write",
        "condition": "console\\.log",
        "message": "Warning: console.log detected in $FILE"
      }
    ]
  }
}
```

### Stop Hooks

```json
{
  "hooks": {
    "Stop": [
      {
        "command": "git diff --cached --name-only | xargs grep -l 'console.log' 2>/dev/null",
        "message": "Check for console.log in staged files before commit"
      }
    ]
  }
}
```

## TodoWrite Best Practices

### When to Use

Use TodoWrite for:
- Multi-step tasks (3+ steps)
- Complex implementations
- Tracking progress
- Validating understanding

### What It Reveals

Good todo list shows:
- Correct step order
- Appropriate granularity
- Complete coverage
- Accurate understanding

Bad todo list signals:
- Out of order steps
- Missing items
- Unnecessary items
- Wrong granularity
- Misinterpreted requirements

### Example

```markdown
## Good Todo List
- [ ] Analyze existing auth flow
- [ ] Create characterization tests
- [ ] Implement JWT validation
- [ ] Update existing tests
- [ ] Run full test suite

## Bad Todo List
- [ ] Do auth stuff
- [ ] Write tests
- [ ] Done
```

## Permission Management

### Auto-Accept (Use with Caution)

```json
{
  "permissions": {
    "allow": [
      "Read",
      "Glob",
      "Grep",
      "LSP"
    ],
    "deny": [
      "Bash:rm -rf",
      "Bash:sudo"
    ]
  }
}
```

### Safety Rules

| Level | Actions |
|-------|---------|
| **Safe** | Read, Glob, Grep, LSP - Can auto-accept |
| **Review** | Edit, Write, Bash - Review before accept |
| **Block** | rm -rf, sudo, force push - Always block |

## Integration with DDD

Hooks support DDD workflow:

```
PreToolUse:
  - Before Edit → Check existing tests

PostToolUse:
  - After Edit → Run related tests

Stop:
  - Session end → Verify test coverage
```

---

Version: 1.0.0
Source: Adapted from everything-claude-code + jikime enhancements
