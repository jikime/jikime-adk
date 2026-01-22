---
description: "Execute iterative improvement loop with LSP/AST-grep feedback (Ralph Loop)"
context: debug
---

# Ralph Loop Command

**Context**: @.claude/contexts/debug.md (Auto-loaded)

Execute iterative code improvement with intelligent feedback from LSP and AST-grep diagnostics.

## EXECUTION PROTOCOL (MUST FOLLOW)

When this command is invoked, Claude MUST execute these steps in order:

### Step 1: Initialize Loop (REQUIRED)

```bash
# Parse user arguments and start the loop
jikime hooks start-loop --task "$ARGUMENTS" [options]
```

**Argument parsing:**
- `$ARGUMENTS` = User's task description (e.g., "Fix all TypeScript errors")
- `--cancel` flag → Run `jikime hooks cancel-loop` instead and stop
- `--max-iterations N` → Pass to start-loop
- `--zero-errors` → Pass to start-loop (default: true)
- `--zero-warnings` → Pass to start-loop
- `--zero-security` → Pass to start-loop
- `--tests-pass` → Pass to start-loop

### Step 2: Load Skill

```
Skill("jikime-workflow-loop")
```

### Step 3: Execute Task

Work on the task described in `$ARGUMENTS`. The Stop Hook (`jikime hooks stop-loop`) will automatically:
- Collect diagnostic snapshots after each Edit/Write
- Check completion conditions when Claude tries to respond
- Re-inject feedback if more work is needed
- Signal completion when all conditions are satisfied

### Step 4: Completion Markers

When the task is truly complete, output one of these markers:
- `<jikime:done />`
- `<jikime:complete />`

---

## Usage

```bash
# Basic usage (fix all errors)
/jikime:loop "Fix all TypeScript errors"

# With options
/jikime:loop "Remove security vulnerabilities" --max-iterations 5 --zero-security

# Fix specific directory
/jikime:loop @src/services/ "Fix all lint errors" --zero-warnings

# Until tests pass
/jikime:loop "Fix failing tests" --tests-pass --max-iterations 10

# Cancel active loop
/jikime:loop --cancel
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--max-iterations` | Maximum number of iterations | 10 |
| `--zero-errors` | Require zero errors for completion | true |
| `--zero-warnings` | Require zero warnings for completion | false |
| `--zero-security` | Require zero security issues | false |
| `--tests-pass` | Require all tests to pass | false |
| `--stagnation-limit` | Max iterations without improvement | 3 |
| `--cancel` | Cancel active loop | - |

## Process

```
1. Initialize Loop
   jikime hooks start-loop --task "..." --options ...
        |
2. Load Skill
   Skill("jikime-workflow-loop")
        |
3. Execute Iteration
   - Analyze current state
   - Fix issues one by one
   - Collect LSP/AST-grep feedback
        |
4. Stop Hook Evaluation
   - Check completion criteria
   - Calculate improvement rate
   - Decide: Continue or Complete
        |
5. (Continue) Re-inject prompt with feedback
        |
6. (Complete) Generate final report
```

## Completion Markers

When task is complete, output one of these markers:

- `<jikime:done />` - Task completed successfully
- `<jikime:complete />` - All goals achieved
- `<promise>DONE</promise>` - Alternative completion marker

## Example Output

```markdown
## Ralph Loop Report

### Session Summary
- Task: "Fix all TypeScript errors"
- Iterations: 5
- Duration: 3m 42s

### Progress
| Iteration | Errors | Warnings | Improvement |
|-----------|--------|----------|-------------|
| 1 (Initial) | 12 | 28 | - |
| 2 | 8 | 24 | 22% |
| 3 | 4 | 18 | 45% |
| 4 | 1 | 12 | 68% |
| 5 (Final) | 0 | 8 | 80% |

### Result
- Status: COMPLETE
- Reason: All errors fixed
- Remaining: 8 warnings (non-blocking)
```

## Skill Integration

This command automatically loads:

```
Skill("jikime-workflow-loop")
```

## Safety Features

1. **Max Iterations**: Default 10, prevents infinite loops
2. **Stagnation Detection**: Stops if no improvement in 3 iterations
3. **Manual Cancel**: Use `/jikime:loop --cancel` anytime
4. **Cost Awareness**: Each iteration consumes tokens

## Best Practices

1. **Be Specific**: "Fix TypeScript errors in src/api/" is better than "Fix errors"
2. **Start Small**: Test with `--max-iterations 3` first
3. **One Goal**: Focus on one type of issue per loop
4. **Use Markers**: Output `<jikime:done />` when truly complete

## Troubleshooting

### Loop Not Starting
- Check if another loop is active: `jikime hooks cancel-loop`
- Verify CLI is installed: `jikime --version`

### Loop Not Stopping
- Output completion marker: `<jikime:done />`
- Check max iterations setting
- Use cancel command: `/jikime:loop --cancel`

### No Progress
- Check if diagnostics are running (LSP, ruff, tsc)
- Verify file types are supported
- Review stagnation limit setting
