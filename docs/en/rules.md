# JikiME-ADK Rules Reference

This is the rules system documentation for JikiME-ADK.

---

## Overview

JikiME-ADK provides consistent development standards through 13 rule files:

### Rules Map

```
┌─────────────────────────────────────────────────────────────────┐
│                     JikiME-ADK Rules                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ Core Rules (Required) ──────────────────────────────────┐  │
│  │                                                            │  │
│  │  core.md          HARD rules (language, execution, output) │  │
│  │  agents.md        Agent delegation rules                   │  │
│  │  quality.md       Quality gates                            │  │
│  │  interaction.md   User interaction                         │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ Development Rules ──────────────────────────────────────┐  │
│  │                                                            │  │
│  │  coding-style.md  Coding style                             │  │
│  │  git-workflow.md  Git workflow                             │  │
│  │  testing.md       Testing guidelines                       │  │
│  │  security.md      Security guidelines                      │  │
│  │  patterns.md      Common patterns                          │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ System Rules ───────────────────────────────────────────┐  │
│  │                                                            │  │
│  │  hooks.md         Hook system                              │  │
│  │  performance.md   Performance optimization                 │  │
│  │  skills.md        Skill discovery/management               │  │
│  │  web-search.md    Web search protocol                      │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Core Rules

### core.md - Core Rules

**Purpose**: Defines HARD rules that must be followed

#### Language Rules

| Rule | Description |
|------|-------------|
| [HARD] Language-Aware Responses | All user responses in `conversation_language` |
| [HARD] Internal Communication | Agent-to-agent communication in English |
| [HARD] Code Comments | Follow `code_comments` setting (default: English) |

#### Execution Rules

| Rule | Description |
|------|-------------|
| [HARD] Parallel Execution | Independent tool calls without dependencies run in parallel |
| [HARD] No XML in User Responses | XML tags must not be shown to users |

#### Output Format Rules

| Rule | Description |
|------|-------------|
| [HARD] Markdown Required | Always use Markdown in user responses |
| [HARD] XML Reserved | XML tags only for internal agent data transfer |

#### Checklist

Verify before responding:

- [ ] Response is in user's `conversation_language`
- [ ] Independent tasks are parallelized
- [ ] No XML tags in response
- [ ] Markdown formatting applied
- [ ] URLs verified before inclusion

---

### agents.md - Agent Delegation Rules

**Purpose**: Defines when and how to delegate to agents

#### Rules by Command Type

##### Type A: Workflow Commands

**Commands**: `/jikime:0-project`, `/jikime:1-plan`, `/jikime:2-run`, `/jikime:3-sync`

- Agent delegation **recommended** (when expertise needed for complex tasks)
- Direct tool use **allowed** (for simple tasks)
- User interaction only via orchestrator's `AskUserQuestion`

##### Type B: Utility Commands

**Commands**: `/jikime:jarvis`, `/jikime:fix`, `/jikime:loop`, `/jikime:test`

- [HARD] **Agent delegation required** for all implementation/modification tasks
- Direct tool access allowed only for diagnostics (LSP, tests, linters)
- **All** code modifications delegated to specialist agents
- Applies even after auto compact or session recovery

**Reason**: Prevents quality degradation during session context loss

#### Selection Decision Tree

```
1. Read-only codebase exploration?
   → Use Explore subagent

2. Need external documentation/API research?
   → Use WebSearch, WebFetch, Context7 MCP tools

3. Need domain expertise?
   → Use specialist subagent (backend, frontend, debugger, etc.)

4. Need workflow coordination?
   → Use manager-[workflow] subagent

5. Complex multi-step task?
   → Use manager-strategy subagent
```

#### Context Optimization

When delegating to agents:

- Pass **minimal context** (spec_id, 3 or fewer key requirements, architecture summary under 200 chars)
- **Exclude** background info, reasoning, non-essential details
- Each agent has independent 200K token session

---

### quality.md - Quality Gates

**Purpose**: Quality verification rules and checklists for all tasks

#### HARD Rules Checklist

Required verification before task completion:

- [ ] All implementation tasks delegated to agents when expertise needed
- [ ] User response in `conversation_language`
- [ ] Independent tasks run in parallel
- [ ] XML tags not shown to users
- [ ] URLs verified before inclusion (WebSearch)
- [ ] Sources cited when using WebSearch

#### SOFT Rules Checklist

Recommended best practices:

- [ ] Appropriate agent selected for task
- [ ] Minimal context passed to agents
- [ ] Results integrated consistently
- [ ] Complex tasks delegated to agents (Type B commands)

#### Violation Detection

| Violation | Description |
|-----------|-------------|
| **No Agent Consideration** | Not considering agent delegation for complex implementation requests |
| **Skipped Verification** | Skipping quality verification for important changes |
| **Language Mismatch** | Ignoring user's `conversation_language` |

#### DDD Quality Standards

When using Domain-Driven Development:

- [ ] Run existing tests before refactoring
- [ ] Create characterization tests for code without coverage
- [ ] Preserve behavior with ANALYZE-PRESERVE-IMPROVE cycle
- [ ] Changes are incremental and verified

#### TRUST 5 Framework

| Principle | Description |
|-----------|-------------|
| **T**ested | Appropriate test coverage for all code |
| **R**eadable | Code is self-documenting and clear |
| **U**nified | Consistent patterns across codebase |
| **S**ecured | Security best practices applied |
| **T**rackable | Changes documented and traceable |

---

### interaction.md - User Interaction Rules

**Purpose**: User interaction and AskUserQuestion usage rules

#### Important Constraint

> Subagents called via Task() operate in isolated stateless contexts and cannot interact directly with users.

**Only orchestrator can use AskUserQuestion** - subagents cannot

#### Correct Workflow Pattern

```
Step 1: Orchestrator collects user preferences via AskUserQuestion
        ↓
Step 2: Orchestrator calls Task() with user choices included in prompt
        ↓
Step 3: Subagent executes with provided parameters (no user interaction)
        ↓
Step 4: Subagent returns structured response with results
        ↓
Step 5: Orchestrator uses AskUserQuestion for next decision based on agent response
```

#### AskUserQuestion Constraints

| Constraint | Rule |
|------------|------|
| Options per question | Maximum 4 |
| Emoji usage | No emojis in question text, headers, option labels |
| Language | Questions in user's `conversation_language` |

#### Clarification Rules

- If user intent is unclear, clarify **before** proceeding via AskUserQuestion
- Collect all necessary user preferences **before** delegating to agents
- Do not assume user preferences without confirmation

---

## Development Rules

### coding-style.md - Coding Style Rules

**Purpose**: Quality and style guidelines for consistent, maintainable code

#### Immutability (CRITICAL)

Always create new objects, never mutate:

```javascript
// ❌ WRONG: Mutation
function updateUser(user, name) {
  user.name = name  // MUTATION!
  return user
}

// ✅ CORRECT: Immutability
function updateUser(user, name) {
  return { ...user, name }
}
```

#### File Organization

**Many small files > Few large files**

| Guideline | Target |
|-----------|--------|
| Lines per file | 200-400 typical, 800 max |
| Lines per function | < 50 lines |
| Nesting depth | < 4 levels |
| Cohesion | High (single responsibility) |
| Coupling | Low (minimal dependencies) |

**Organization Principle**: Organize by feature/domain, not by type

```
# ❌ WRONG: By type
src/
├── components/
├── hooks/
├── services/
└── utils/

# ✅ CORRECT: By feature
src/
├── auth/
│   ├── components/
│   ├── hooks/
│   └── services/
├── users/
└── products/
```

#### Error Handling

Always handle errors comprehensively:

```typescript
try {
  const result = await riskyOperation()
  return result
} catch (error) {
  console.error('Operation failed:', error)
  throw new Error('Detailed user-friendly message')
}
```

#### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Variables | camelCase | `userName`, `isActive` |
| Functions | camelCase, verb prefix | `getUserById`, `validateInput` |
| Classes | PascalCase | `UserService`, `AuthController` |
| Constants | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT`, `API_BASE_URL` |
| Files | kebab-case or camelCase | `user-service.ts`, `userService.ts` |

#### Prohibited Patterns

| Pattern | Reason |
|---------|--------|
| `any` type (TypeScript) | Defeats type safety |
| Magic numbers | Use named constants |
| Deep nesting | Extract to functions |
| God objects/functions | Split by responsibility |
| Commented-out code | Delete (use git history) |

---

### git-workflow.md - Git Workflow Rules

**Purpose**: Git conventions and workflow guidelines for consistent version control

#### Commit Message Format

```
<type>: <description>

<optional body>
```

#### Commit Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code refactoring (no behavior change) |
| `docs` | Documentation changes |
| `test` | Adding/modifying tests |
| `chore` | Maintenance tasks |
| `perf` | Performance improvements |
| `ci` | CI/CD changes |

#### Branch Naming

```
<type>/<description>
```

| Type | Purpose |
|------|---------|
| `feature/` | New feature |
| `fix/` | Bug fix |
| `refactor/` | Code refactoring |
| `docs/` | Documentation |
| `chore/` | Maintenance |

#### Prohibited Actions

| Action | Reason |
|--------|--------|
| Force push to main/master | Destroys history |
| Committing secrets | Security risk |
| Large single commits | Hard to review/revert |
| Merge commits on feature branches | Complicates history |
| Committing build artifacts | Bloats repository |

---

### testing.md - Testing Guidelines

**Purpose**: Testing best practices applying DDD methodology

#### Coverage Targets

| Type | Target | Priority |
|------|--------|----------|
| Business logic | 90%+ | Critical |
| API endpoints | 80%+ | High |
| UI components | 70%+ | Medium |
| Utilities | 80%+ | Medium |
| **Overall** | **80%+** | Required |

#### DDD Testing Approach

##### ANALYZE → PRESERVE → IMPROVE

```
1. ANALYZE
   - Run existing tests
   - Identify test coverage gaps
   - Understand current behavior

2. PRESERVE
   - Write characterization tests for code without coverage
   - Capture current behavior as baseline
   - Ensure no regressions

3. IMPROVE
   - Implement changes
   - Run all tests after each change
   - Add tests for new features
```

#### Good Test Principles

| Principle | Description |
|-----------|-------------|
| **Fast** | Run quickly, encourage frequent execution |
| **Isolated** | No dependencies between tests |
| **Repeatable** | Same results every time |
| **Self-validating** | Clear pass/fail, no manual verification |
| **Timely** | Written close to code changes |

---

### security.md - Security Guidelines

**Purpose**: Security best practices based on OWASP Top 10 and industry standards

#### OWASP Top 10 Checklist

##### 1. Injection

```typescript
// ❌ CRITICAL: SQL Injection
const query = `SELECT * FROM users WHERE id = ${userId}`

// ✅ SAFE: Parameterized query
const { data } = await supabase.from('users').select('*').eq('id', userId)
```

##### 2. Broken Authentication

```typescript
// ❌ CRITICAL: Plaintext password comparison
if (password === storedPassword) { /* login */ }

// ✅ SAFE: Hash comparison
const isValid = await bcrypt.compare(password, hashedPassword)
```

##### 3. Sensitive Data Exposure

```typescript
// ❌ CRITICAL: Hardcoded secret
const apiKey = "sk-proj-xxxxx"

// ✅ SAFE: Environment variable
const apiKey = process.env.OPENAI_API_KEY
```

##### 4. XSS

```typescript
// ❌ HIGH: XSS vulnerability
element.innerHTML = userInput

// ✅ SAFE: Use textContent
element.textContent = userInput
```

#### Secret Management

##### Environment Variables

```typescript
// ❌ NEVER: Hardcoded secret
const apiKey = 'sk-proj-xxxxx'

// ✅ ALWAYS: Environment variable
const apiKey = process.env.API_KEY

if (!apiKey) {
  throw new Error('API_KEY environment variable not set')
}
```

#### Security Checklist

Before every commit:

- [ ] No hardcoded secrets
- [ ] All user input validated
- [ ] SQL Injection prevented
- [ ] XSS prevented
- [ ] CSRF protection (if applicable)
- [ ] Authentication verified
- [ ] Authorization checked
- [ ] No sensitive data logging
- [ ] Error messages don't leak information
- [ ] Dependencies up to date

---

### patterns.md - Common Patterns

**Purpose**: Reusable code patterns for consistent implementation

#### API Patterns

##### Response Format

```typescript
interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: unknown
  }
  meta?: {
    total: number
    page: number
    limit: number
    hasMore: boolean
  }
}
```

#### Repository Pattern

```typescript
interface Repository<T, ID = string> {
  findAll(filters?: Filters): Promise<T[]>
  findById(id: ID): Promise<T | null>
  create(data: CreateDto<T>): Promise<T>
  update(id: ID, data: UpdateDto<T>): Promise<T>
  delete(id: ID): Promise<void>
  exists(id: ID): Promise<boolean>
}
```

#### Validation Pattern

```typescript
import { z } from 'zod'

const userSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8).max(100),
  name: z.string().min(1).max(50).optional()
})

type UserInput = z.infer<typeof userSchema>
```

#### Pattern Selection Guide

| Scenario | Pattern |
|----------|---------|
| Data access | Repository |
| Business logic | Service |
| Object creation | Factory |
| State management | Custom Hook |
| Complex components | Compound |
| Input validation | Zod Schema |

---

## System Rules

### hooks.md - Hook System

**Purpose**: Claude Code hooks for automated workflows and quality enforcement

#### Hook Types

| Type | Timing | Purpose |
|------|--------|---------|
| **PreToolUse** | Before tool execution | Validation, modification, blocking |
| **PostToolUse** | After tool execution | Auto-format, checks, logging |
| **Notification** | On specific events | Notifications, status updates |
| **Stop** | On session end | Final verification |

#### Recommended PostToolUse Hooks

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
      }
    ]
  }
}
```

#### Permission Management

| Level | Action |
|-------|--------|
| **Safe** | Read, Glob, Grep, LSP - Can auto-accept |
| **Review** | Edit, Write, Bash - Review before accepting |
| **Block** | rm -rf, sudo, force push - Always block |

---

### performance.md - Performance Optimization

**Purpose**: Efficient Claude Code usage and code performance guidelines

#### Model Selection Strategy

##### Haiku (Fast, Cost-efficient)

**Use for**: Simple code generation, formatting, simple Q&A, workers in multi-agent workflows

**Characteristics**: 90% of Sonnet capability, 3x cost savings, fastest response

##### Sonnet (Balanced, Recommended)

**Use for**: Main development tasks, complex coding, multi-agent workflow coordination

**Characteristics**: Optimal balance of capability and cost, strong coding performance

##### Opus (Maximum Capability)

**Use for**: Complex architecture decisions, deep reasoning, research analysis

**Characteristics**: Maximum reasoning depth, best for complex problems, higher cost and latency

#### Context Window Management

| Zone | Context | Recommendation |
|------|---------|----------------|
| Critical | 80-100% | Avoid large refactoring, complex debugging |
| Safe | 0-60% | Single file edits, independent utility creation |

#### Algorithm Complexity Targets

```
- General operations: O(n) or less
- Sorting: O(n log n) acceptable
- Avoid O(n²) in hot paths
- No O(2^n) without explicit approval
```

---

### skills.md - Skill Discovery and Management

**Purpose**: Skill discovery, loading, and utilization rules

#### Skill Discovery Commands

```bash
# List all available skills
jikime-adk skill list

# Filter by tag, phase, agent, language
jikime-adk skill list --tag framework
jikime-adk skill list --language typescript

# Search skills by keyword
jikime-adk skill search <keyword>

# Find related skills
jikime-adk skill related <skill-name>

# Detailed skill info
jikime-adk skill info <skill-name> --body
```

#### Skill Loading Rules

##### Automatic Loading (Triggers)

```yaml
triggers:
  keywords: ["react", "component"]     # When included in user input
  phases: ["run"]                      # Current development phase
  agents: ["frontend"]                 # Agent in use
  languages: ["typescript"]            # Project language
```

##### Progressive Disclosure

| Level | Content | Tokens | Load Timing |
|-------|---------|--------|-------------|
| **Level 1** | Metadata only | ~100 | Agent initialization |
| **Level 2** | Full body | ~5K | Trigger condition match |
| **Level 3+** | Bundle files | Variable | When Claude needs it |

#### Skill Categories

| Category | Prefix | Example |
|----------|--------|---------|
| Language | `jikime-lang-*` | jikime-lang-typescript, jikime-lang-python |
| Platform | `jikime-platform-*` | jikime-platform-vercel, jikime-platform-supabase |
| Domain | `jikime-domain-*` | jikime-domain-frontend, jikime-domain-backend |
| Workflow | `jikime-workflow-*` | jikime-workflow-spec, jikime-workflow-ddd |
| Foundation | `jikime-foundation-*` | jikime-foundation-claude, jikime-foundation-core |

---

### web-search.md - Web Search Protocol

**Purpose**: Anti-hallucination policy and URL verification rules

#### HARD Rules

| Rule | Description |
|------|-------------|
| [HARD] URL Verification | All URLs verified via WebFetch before inclusion |
| [HARD] Uncertainty Disclosure | Unverified information marked as uncertain |
| [HARD] Source Attribution | Include actual sources for all web search results |

#### Execution Steps

```
1. Initial Search
   → Use WebSearch with specific, targeted queries

2. URL Validation
   → Verify each URL via WebFetch before inclusion

3. Response Construction
   → Include only verified URLs with actual search sources
```

#### Prohibited Actions

| Action | Reason |
|--------|--------|
| Creating URLs not found in search | Generates false information |
| Presenting uncertain info as fact | Misleads users |
| Omitting "Sources:" section | Hides information sources |

#### Response Format

Always include when using WebSearch:

```markdown
## Answer

[Response with verified information]

## Sources

- [Source Title 1](https://verified-url-1.com)
- [Source Title 2](https://verified-url-2.com)
```

---

## Rule Priority

### HARD vs SOFT Rules

| Type | Application | Example |
|------|-------------|---------|
| **HARD** | Required, no exceptions | Language rules, parallel execution, XML prohibition |
| **SOFT** | Recommended, flexible by situation | Agent delegation, quality checks |

### Actions on Violation

| Violation Level | Action |
|-----------------|--------|
| HARD Rule violation | Immediate correction required |
| SOFT Rule violation | Consider recommendation |
| Security violation | Stop work, immediate correction |

---

## Rules Reference Table

| Rule File | Main Content | Applies To |
|-----------|--------------|------------|
| core.md | HARD rules | All tasks |
| agents.md | Agent delegation | Command execution |
| quality.md | Quality gates | Code changes |
| interaction.md | User interaction | Orchestrator responses |
| coding-style.md | Coding style | Code writing |
| git-workflow.md | Git conventions | Version control |
| testing.md | Testing guide | Test writing |
| security.md | Security guide | All code |
| patterns.md | Common patterns | Implementation |
| hooks.md | Hook system | Automation |
| performance.md | Performance optimization | Efficiency |
| skills.md | Skill management | Skill loading |
| web-search.md | Web search | Information retrieval |

---

Version: 1.0.0
Last Updated: 2026-01-22
