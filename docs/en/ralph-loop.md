# Ralph Loop - Intelligent Iterative Code Improvement

JikiME-ADK's intelligent iterative code improvement system. Automated code fix loop based on LSP/AST-grep feedback.

## Overview

Ralph Loop was inspired by Claude Code's official plugin "Ralph Wiggum", but provides **diagnostic-based intelligent iteration** rather than simple repetition.

### Differentiation Points

| Existing Ralph (Simple) | JikiME Ralph (Intelligent) |
|------------------|----------------------|
| Simple prompt repetition | LSP/AST-grep feedback-based iteration |
| Zero errors = complete | Auto-continue based on errors + multiple conditions |
| No state | DiagnosticSnapshot history tracking |
| Fixed conditions | Adaptive completion conditions |
| Manual start required | Automatic detection and continuation |

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Ralph Loop System                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │ start-loop  │    │  stop-loop  │    │ cancel-loop │     │
│  │   (manual)  │    │   (auto)    │    │   (manual)  │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│         │                  │                  │             │
│         └──────────┬───────┴──────────┬──────┘             │
│                    │                  │                     │
│              ┌─────▼─────┐     ┌──────▼──────┐             │
│              │ LoopState │     │  Diagnostic  │             │
│              │   (.json) │     │  Snapshot    │             │
│              └───────────┘     └─────────────┘             │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              PostToolUse Hooks                       │   │
│  │  ┌────────────┐  ┌─────────────┐  ┌────────────┐   │   │
│  │  │ post-tool- │  │ post-tool-  │  │ post-tool- │   │   │
│  │  │    lsp     │  │  ast-grep   │  │   linter   │   │   │
│  │  └────────────┘  └─────────────┘  └────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Implementation File List

#### Go Hook Files

| File | Description |
|------|------|
| `cmd/hookscmd/loop_state.go` | LoopState, DiagnosticSnapshot, CompletionCriteria structs |
| `cmd/hookscmd/start_loop.go` | Loop start command (`jikime hooks start-loop`) |
| `cmd/hookscmd/stop_loop.go` | Completion condition evaluation and auto-continue logic |
| `cmd/hookscmd/cancel_loop.go` | Loop cancel command (`jikime hooks cancel-loop`) |
| `cmd/hookscmd/post_tool_lsp.go` | LSP result snapshot recording |
| `cmd/hookscmd/post_tool_ast_grep.go` | AST-grep result snapshot recording |

#### Skill/Command Files

| File | Description |
|------|------|
| `templates/.claude/skills/jikime-workflow-loop/SKILL.md` | Ralph Loop workflow skill |
| `templates/.claude/commands/jikime/loop.md` | `/jikime:loop` slash command |

#### Configuration Files

| File | Hook Registration |
|------|----------|
| `templates/.claude/settings.json` | Stop Hook, PostToolUse Hooks registration |

## Data Structures

### LoopState

Manages the overall state of a loop session.

```go
type LoopState struct {
    // Basic info
    Active           bool      `json:"active"`
    SessionID        string    `json:"session_id"`
    StartedAt        time.Time `json:"started_at"`
    UpdatedAt        time.Time `json:"updated_at"`

    // Iteration info
    Iteration        int       `json:"iteration"`
    MaxIterations    int       `json:"max_iterations"`

    // Task info
    TaskDescription  string    `json:"task_description"`
    TargetFiles      []string  `json:"target_files,omitempty"`

    // Completion criteria
    Criteria         CompletionCriteria `json:"completion_criteria"`

    // Diagnostic history
    Snapshots        []DiagnosticSnapshot `json:"snapshots"`

    // Final result
    CompletionReason string    `json:"completion_reason,omitempty"`
    FinalStatus      string    `json:"final_status,omitempty"`
}
```

### DiagnosticSnapshot

Captures diagnostic results at each iteration.

```go
type DiagnosticSnapshot struct {
    Iteration      int       `json:"iteration"`
    Timestamp      time.Time `json:"timestamp"`

    // LSP diagnostics
    ErrorCount     int       `json:"error_count"`
    WarningCount   int       `json:"warning_count"`
    InfoCount      int       `json:"info_count"`

    // AST-grep results
    SecurityIssues int       `json:"security_issues"`

    // Test results
    TestsPassed    bool      `json:"tests_passed"`
    TestsRun       int       `json:"tests_run"`
    TestsFailed    int       `json:"tests_failed"`

    // File details
    FileDetails    []FileDetail `json:"file_details,omitempty"`
}
```

### CompletionCriteria

Defines completion conditions.

```go
type CompletionCriteria struct {
    ZeroErrors      bool `json:"zero_errors"`       // Zero errors required
    ZeroWarnings    bool `json:"zero_warnings"`     // Zero warnings required
    ZeroSecurity    bool `json:"zero_security"`     // Zero security issues required
    TestsPass       bool `json:"tests_pass"`        // Tests must pass
    StagnationLimit int  `json:"stagnation_limit"`  // Limit for iterations without improvement
}
```

## Execution Flow

### Automatic Operation Flow (Default)

```
User: "Fix the TypeScript errors"
        │
        ▼
   Claude performs work (Edit/Write)
        │
        ▼
   PostToolUse Hooks auto-execute
   - post-tool-lsp → record snapshot
   - post-tool-ast-grep → record snapshot
        │
        ▼
   Claude attempts to complete response
        │
        ▼
   Stop Hook (stop-loop) auto-executes
        │
        ▼
   ┌─────────────────────────────────────┐
   │ 1. Completion marker detected?      │
   │    YES → exit 0 (force terminate)   │
   │                                     │
   │ 2. Collect diagnostics (ruff, tsc)  │
   │                                     │
   │ 3. Errors > 0 or security issues > 0?│
   │    YES → exit 1 (auto-continue)     │
   │    NO  → exit 0 (normal terminate)  │
   └─────────────────────────────────────┘
        │
        ├── exit 0 → Claude terminates normally
        │
        └── exit 1 → Feedback re-injection → Claude continues work
                     "Ralph Loop: AUTO-CONTINUE |
                      5 error(s) remaining |
                      Next: Fix the remaining errors"
```

### Explicit Loop Flow (Advanced)

```bash
# 1. Start loop (options can be specified)
jikime hooks start-loop --task "Fix all errors" --max-iterations 10

# 2. Claude performs work
# ... (PostToolUse hooks collect snapshots)

# 3. Stop Hook evaluates completion conditions
# - ZeroErrors check
# - Stagnation detection
# - MaxIterations check

# 4. Complete or continue
# exit 0: complete / exit 1: continue

# 5. Cancel (if needed)
jikime hooks cancel-loop
```

## Auto-Loop Mechanism

### Core Logic (stop_loop.go)

```go
func runStopLoop(cmd *cobra.Command, args []string) error {
    // 1. Check completion marker (highest priority)
    if checkCompletionPromise(conversationText) {
        return nil // exit 0 - complete
    }

    // 2. Collect diagnostics
    currentSnapshot := collectCurrentDiagnostics()

    // 3. Check loop state
    state := LoadEnhancedLoopState()

    // 4. AUTO-LOOP: Auto-continue if errors exist even when loop is inactive
    if !state.Active {
        if currentSnapshot.ErrorCount > 0 || currentSnapshot.SecurityIssues > 0 {
            // Output feedback and continue
            os.Exit(1) // exit 1 - continue
        }
        return nil // exit 0 - no errors
    }

    // 5. Explicit loop logic (omitted)
    // ...
}
```

### Diagnostic Collection Method

```go
func collectCurrentDiagnostics() DiagnosticSnapshot {
    snapshot := DiagnosticSnapshot{}

    // Python: ruff check
    if _, err := exec.LookPath("ruff"); err == nil {
        cmd := exec.Command("ruff", "check", "--output-format=json", ".")
        // E*, F* codes → ErrorCount
        // Others → WarningCount
    }

    // TypeScript: tsc --noEmit
    if _, err := exec.LookPath("tsc"); err == nil {
        cmd := exec.Command("tsc", "--noEmit", "--pretty", "false")
        // Lines containing "error" → ErrorCount
    }

    // Tests: pytest / npm test
    snapshot.TestsPassed, _ = checkTests()

    return snapshot
}
```

## Usage

### Automatic Mode (Default)

Works automatically with regular prompts without special commands.

```
User: Fix all TypeScript errors
        ↓
Claude: (performs work)
        ↓
Stop Hook: 5 errors detected → auto-continue
        ↓
Claude: (continues work)
        ↓
Stop Hook: 0 errors → auto-terminate
```

### Explicit Commands

```bash
# Basic usage
/jikime:loop "Fix all TypeScript errors"

# Specify options
/jikime:loop "Remove security vulnerabilities" --max-iterations 5 --zero-security

# Specific directory
/jikime:loop @src/services/ "Fix all lint errors" --zero-warnings

# Until tests pass
/jikime:loop "Fix failing tests" --tests-pass --max-iterations 10

# Cancel
/jikime:loop --cancel
```

### Direct CLI Execution

```bash
# Start loop
jikime hooks start-loop \
  --task "Fix all errors" \
  --max-iterations 10 \
  --zero-errors \
  --tests-pass

# Cancel loop
jikime hooks cancel-loop
```

## Completion Markers

Markers used when Claude declares work completion.

```
<jikime>DONE</jikime>
<jikime>COMPLETE</jikime>
<jikime:done />
<jikime:complete />
```

When one of these markers is detected, it **terminates immediately** regardless of error presence.

## Completion Criteria

### Options

| Option | Flag | Default | Description |
|------|--------|--------|------|
| Zero Errors | `--zero-errors` | true | Zero errors required |
| Zero Warnings | `--zero-warnings` | false | Zero warnings required |
| Zero Security | `--zero-security` | false | Zero security issues required |
| Tests Pass | `--tests-pass` | false | Tests must pass |
| Max Iterations | `--max-iterations` | 10 | Maximum iteration count |
| Stagnation Limit | `--stagnation-limit` | 3 | Limit for iterations without improvement |

### Termination Conditions

1. **Completion marker detected**: Immediate termination (exit 0)
2. **All conditions satisfied**: Zero errors, etc. → terminate (exit 0)
3. **Maximum iterations reached**: MaxIterations exceeded → terminate (exit 0)
4. **Stagnation detected**: N consecutive iterations without improvement → terminate (exit 0)
5. **No errors (auto mode)**: Zero errors + zero security issues → terminate (exit 0)

## Safety Features

### 1. Maximum Iteration Limit

```go
if state.Iteration >= state.MaxIterations {
    state.FinalStatus = "STOPPED"
    state.CompletionReason = "Max iterations reached"
    os.Exit(0)
}
```

### 2. Stagnation Detection

```go
func (s *LoopState) IsStagnant() bool {
    // Determine stagnation if no improvement in recent N iterations
    recent := s.Snapshots[len(s.Snapshots)-limit:]
    // Return true if issue count does not decrease
}
```

### 3. Disable

```bash
# Disable via environment variable
export JIKIME_DISABLE_LOOP_CONTROLLER=1
```

### 4. Manual Cancel

```bash
jikime hooks cancel-loop
# or
/jikime:loop --cancel
```

## Configuration

### settings.json Hook Registration

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks post-tool-ast-grep",
            "timeout": 30000
          },
          {
            "type": "command",
            "command": "jikime hooks post-tool-lsp",
            "timeout": 30000
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks stop-loop",
            "timeout": 10000
          }
        ]
      }
    ]
  }
}
```

## Feedback Messages

### AUTO-CONTINUE (Auto-continue)

```
Ralph Loop: AUTO-CONTINUE | Issues detected - continuing automatically |
5 error(s) remaining | 2 security issue(s) remaining |
Next: Fix the remaining errors | Output <jikime:done /> when complete
```

### CONTINUE (Explicit loop)

```
Ralph Loop: CONTINUE | Iteration: 3/10 |
Current: 5 error(s), 12 warning(s), 0 security issue(s) |
Progress: 45% improvement | Next: Fix 5 remaining error(s)
```

### COMPLETE (Complete)

```
Ralph Loop: COMPLETE - All conditions satisfied |
Session: loop-1705912345 | Iterations: 5 |
Total improvement: 100% | Initial: 12 errors, 28 warnings |
Final: 0 errors, 8 warnings
```

## Integration with DDD

Ralph Loop integrates with Domain-Driven Development workflows.

```
┌─────────────────────────────────────────────────────────────┐
│                   DDD + Ralph Loop                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ANALYZE: Loop collects diagnostic data                     │
│     ↓                                                       │
│  PRESERVE: Verify existing behavior at each iteration       │
│     ↓                                                       │
│  IMPROVE: Incremental fixes with measurable progress        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Troubleshooting

### Loop doesn't start

```bash
# Check if another loop is active
jikime hooks cancel-loop

# Verify CLI installation
jikime --version
```

### Loop doesn't stop

```bash
# Output completion marker
# Request Claude to output "<jikime:done />"

# Or force cancel
/jikime:loop --cancel
```

### Diagnostics not collected

```bash
# Verify tool installation
which ruff
which tsc

# Check project configuration
ls tsconfig.json
ls pyproject.toml
```

### Hook not working

```bash
# Check settings.json
cat .claude/settings.json | jq '.hooks.Stop'

# Manual test
echo '{"messages":[]}' | jikime hooks stop-loop
```

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-01-22 | Initial implementation - error-based auto-continue |

## References

- [Ralph Wiggum Plugin](https://github.com/anthropics/claude-code-plugins) - Original inspiration
- [Claude Code Hooks](https://docs.anthropic.com/claude-code/hooks) - Hook system documentation
- [JikiME-ADK](https://github.com/jikime/jikime-adk) - Project repository
