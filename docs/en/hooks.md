# JikiME-ADK Hooks System

A Go-based hook system that integrates with Claude Code's lifecycle events.

## Overview

JikiME-ADK provides a hook system implemented in Go that responds to various Claude Code events. All hooks communicate with Claude Code using the JSON stdin/stdout protocol.

### Advantages

| Characteristic | Description |
|----------------|-------------|
| **Single Binary** | No separate runtime required (Python, Node.js, etc.) |
| **Fast Startup** | Millisecond-level execution |
| **Cross-Platform** | Supports macOS, Linux, Windows |
| **Consistency** | Same codebase as jikime-adk CLI |

## Hook Categories

### 1. Session Hooks

Hooks that run at session start/end.

#### session-start

**Purpose**: Display project information and environment status at session start

```bash
jikime hooks session-start
```

**Features**:
- Display project name and version
- Git branch and changes status
- Github-Flow mode and Auto Branch settings
- Recent commit information
- Conversation language settings
- Environment validation warnings (v1.1.0+)

**Output Example**:
```
🚀 JikiME-ADK Session Started
   📦 Version: 1.0.0
   🔄 Changes: 3 file(s) modified
   🌿 Branch: feature/auth
   🔧 Github-Flow: personal | Auto Branch: Yes
   🔨 Last Commit: abc1234 - Add login feature (2 hours ago)
   🌐 Language: 한국어 (ko)
   👋 Welcome back, Anthony!

   ⚠️  Environment Warnings:
      - node_modules not found - run 'npm install' or equivalent
```

**Environment Validation (v1.1.0+)**:

Automatically detects project type and verifies required tools:

| Project Type | Detection Files | Validation Items |
|--------------|-----------------|------------------|
| Node.js | `package.json` | node, npm/pnpm/yarn, node_modules |
| Python | `pyproject.toml`, `requirements.txt` | python3, .venv |
| Go | `go.mod` | go |
| Rust | `Cargo.toml` | cargo |

Additional validations:
- Checks for `.env` file existence when `.env.example` exists
- Verifies Git installation

#### session-end-cleanup

**Purpose**: Perform cleanup tasks at session end

```bash
jikime hooks session-end-cleanup
```

**Features**:
- Temporary file cleanup
- Desktop notification sending (v1.1.0+)
- Session summary generation

**Desktop Notifications (v1.1.0+)**:

Supports cross-platform desktop notifications:

| Platform | Implementation |
|----------|----------------|
| macOS | osascript (AppleScript) |
| Linux | notify-send |
| Windows | PowerShell Toast Notification |

To disable notifications:
```bash
export JIKIME_NO_NOTIFY=1
```

### 2. UserPromptSubmit Hooks

Hooks that run when a user prompt is submitted.

#### user-prompt-submit

**Purpose**: Prompt analysis and agent hint provision (v1.1.0+)

```bash
jikime hooks user-prompt-submit
```

**Features**:

1. **Agent Hint Provision**:

   Analyzes prompt keywords and suggests appropriate agents:

   | Keyword | Recommended Agent | Hint |
   |---------|-------------------|------|
   | security, vulnerability, audit | security-auditor | Security analysis detected |
   | performance, optimize, bottleneck | optimizer | Performance optimization detected |
   | test, coverage, unit test | test-guide | Testing task detected |
   | refactor, clean up, simplify | refactorer | Refactoring detected |
   | debug, error, fix bug | debugger | Debugging detected |
   | api, endpoint, backend | backend | Backend task detected |
   | component, ui, frontend | frontend | Frontend task detected |
   | deploy, ci/cd, pipeline | devops | DevOps task detected |
   | architecture, design, structure | architect | Architecture design detected |
   | document, readme, guide | documenter | Documentation task detected |
   | database, schema, migration | backend | Database task detected |
   | e2e, playwright, browser | e2e-tester | E2E testing detected |

2. **Dangerous Pattern Warnings**:

   Detects and warns about potentially dangerous command patterns:

   | Pattern | Warning Message |
   |---------|-----------------|
   | `rm -rf` | Destructive 'rm -rf' command detected |
   | `git push --force` | Force push command detected |
   | `git reset --hard` | Hard reset command detected |
   | `DROP TABLE`, `DROP DATABASE` | Database deletion command detected |
   | `sudo rm` | Admin privilege deletion command detected |
   | `chmod 777` | Full permission setting detected |
   | `curl \| sh`, `wget \| sh` | Pipe execution pattern detected |
   | `::: force`, `--no-verify` | Force flag detected |
   | `password =`, `secret =` | Potential secret exposure detected |
   | `*` wildcard deletion | Wildcard deletion pattern detected |

#### orchestrator-route

**Purpose**: Route requests to the appropriate orchestrator

```bash
jikime hooks orchestrator-route
```

**Features**:
- Migration keyword detection → F.R.I.D.A.Y. activation
- Development keyword detection → J.A.R.V.I.S. activation
- Orchestrator state file management

### 3. PreToolUse Hooks

Hooks that perform validation before tool execution.

#### pre-tool-security

**Purpose**: Security validation before Write/Edit tool execution

```bash
jikime hooks pre-tool-security
```

**Validation Items**:
- Block modification of sensitive files (`.env`, `secrets/`, `~/.ssh/`)
- Hardcoded secret pattern detection

#### pre-write

**Purpose**: Path validation before file creation

```bash
jikime hooks pre-write
```

**Validation Items**:
- Restrict document file (`.md`, `.txt`) creation paths
- Allowed paths: `README.md`, `CLAUDE.md`, `docs/`, `.jikime/`, `.claude/`, `migrations/`, `SKILL.md`

### 4. PostToolUse Hooks

Hooks that perform processing after tool execution.

#### post-tool-formatter

**Purpose**: Automatic code file formatting

```bash
jikime hooks post-tool-formatter
```

**Supported Formatters**:
- Prettier (JS/TS/JSON/CSS/MD)
- Black (Python)
- gofmt (Go)
- rustfmt (Rust)

#### post-tool-linter

**Purpose**: Code file lint checking

```bash
jikime hooks post-tool-linter
```

**Supported Linters**:
- ESLint (JS/TS)
- Ruff (Python)
- golangci-lint (Go)
- Clippy (Rust)

#### post-tool-ast-grep

**Purpose**: AST-based code pattern checking

```bash
jikime hooks post-tool-ast-grep
```

**Inspection Items**:
- Security vulnerability patterns
- Code quality issues
- Project-specific custom rules

#### post-tool-lsp

**Purpose**: LSP diagnostic collection and reporting

```bash
jikime hooks post-tool-lsp
```

**Features**:
- Type error collection
- Lint warning collection
- Quality gate validation

#### post-bash

**Purpose**: Post-processing after Bash command execution

```bash
jikime hooks post-bash
```

### 5. Stop Hooks

Hooks that run when a session is interrupted/completed.

#### stop-loop

**Purpose**: Handle active loop termination

```bash
jikime hooks stop-loop
```

**Features**:
- Loop state file cleanup
- Final status report

#### stop-audit

**Purpose**: Generate session audit logs

```bash
jikime hooks stop-audit
```

### 6. Loop Control Hooks

Hooks for controlling repeated execution.

#### start-loop

**Purpose**: Start a new loop session

```bash
jikime hooks start-loop --task "Fix all errors" --max-iterations 10
```

#### cancel-loop

**Purpose**: Cancel an active loop

```bash
jikime hooks cancel-loop
```

### 7. Pre-Compact Hooks

Hooks that run before context compression.

#### pre-compact

**Purpose**: Preserve important state before compression

```bash
jikime hooks pre-compact
```

**Features**:
- Orchestrator state preservation
- Active task state saving

## Configuration

Hooks are configured in `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks session-start"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks session-end-cleanup",
            "timeout": 5000
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks user-prompt-submit",
            "timeout": 3000
          },
          {
            "type": "command",
            "command": "jikime hooks orchestrator-route",
            "timeout": 3000
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks pre-tool-security",
            "timeout": 5000
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks post-tool-formatter",
            "timeout": 30000
          },
          {
            "type": "command",
            "command": "jikime hooks post-tool-linter",
            "timeout": 60000
          },
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
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks post-bash",
            "timeout": 10000
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
          },
          {
            "type": "command",
            "command": "jikime hooks stop-audit",
            "timeout": 10000
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks pre-compact",
            "timeout": 5000
          }
        ]
      }
    ]
  }
}
```

## Hook Response Format

All hooks respond in JSON format:

```json
{
  "continue": true,
  "systemMessage": "Hook executed successfully",
  "performance": {
    "go_hook": true
  },
  "error_details": {
    "error": "Optional error message"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `continue` | boolean | Whether to continue execution |
| `systemMessage` | string | System message (optional) |
| `performance` | object | Performance-related metadata (optional) |
| `error_details` | object | Error details (optional) |

## Related Files

| File | Description |
|------|-------------|
| `cmd/hookscmd/hooks.go` | Hook command registration |
| `cmd/hookscmd/session_start.go` | SessionStart hook implementation |
| `cmd/hookscmd/session_end_*.go` | SessionEnd hook implementation |
| `cmd/hookscmd/user_prompt_submit.go` | UserPromptSubmit hook implementation |
| `cmd/hookscmd/orchestrator_route.go` | Orchestrator routing hook |
| `cmd/hookscmd/pre_*.go` | PreToolUse hook implementation |
| `cmd/hookscmd/post_*.go` | PostToolUse hook implementation |
| `cmd/hookscmd/stop_*.go` | Stop hook implementation |
| `cmd/hookscmd/loop_*.go` | Loop control hook implementation |
| `templates/.claude/settings.json` | Hook settings template |

---

Version: 1.1.0
Last Updated: 2026-01-25
Changelog:
- v1.1.0: Added environment validation, agent hints, dangerous pattern warnings, desktop notifications
- v1.0.0: Initial Go-based hook system implementation
