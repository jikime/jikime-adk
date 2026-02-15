# JikiME-ADK Contexts Reference

Context system documentation for JikiME-ADK.

---

## Overview

Contexts define Claude's operational modes. Appropriate contexts are automatically loaded depending on the situation, or can be manually switched.

### Context Map

```
┌─────────────────────────────────────────────────────────────────┐
│                     JikiME-ADK Contexts                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ Development Contexts ────────────────────────────────────┐  │
│  │                                                            │  │
│  │  dev.md      Development Mode (Code First)                 │  │
│  │  planning.md Planning Mode (Think First)                   │  │
│  │  debug.md    Debug Mode (Investigation)                    │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ Review & Sync Contexts ──────────────────────────────────┐  │
│  │                                                            │  │
│  │  review.md   Review Mode (Quality Focus)                   │  │
│  │  sync.md     Sync Mode (Documentation First)               │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ Research Context ────────────────────────────────────────┐  │
│  │                                                            │  │
│  │  research.md Research Mode (Understanding First)           │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Auto-Load Mapping

| Command | Auto-Loaded Context |
|---------|---------------------|
| `/jikime:0-project` | research.md |
| `/jikime:1-plan` | planning.md |
| `/jikime:2-run` | dev.md |
| `/jikime:3-sync` | sync.md |
| `/jikime:build-fix` | debug.md |
| `/jikime:learn` | research.md |

---

## dev.md - Development Context

**Mode**: Active Development
**Focus**: Implementation, coding, building features
**Methodology**: DDD (ANALYZE → PRESERVE → IMPROVE)

### Core Principles

```
1. Get it working   → Make it work first
2. Get it right     → Then make it correct
3. Get it clean     → Finally make it clean
```

### Behavioral Rules

| DO | DON'T |
|----|-------|
| Code first, explain later | Over-engineer simple solutions |
| Working solution over perfection | Add unrequested features |
| Run tests after changes | Skip error handling |
| Keep commits atomic | console.log in production code |
| Follow existing code patterns | Ignore existing tests |

### DDD Cycle

Before modifying existing code:

```
ANALYZE   → Understand current behavior
PRESERVE  → Ensure existing behavior with tests
IMPROVE   → Implement changes incrementally
```

### Tool Priority

| Priority | Tool | Purpose |
|----------|------|---------|
| 1 | Edit | Modify existing files |
| 2 | Write | Create new files |
| 3 | Bash | Run tests, builds, commands |
| 4 | Grep/Glob | Find code patterns |
| 5 | Read | Understand before editing |

### Coding Standards

```yaml
files:
  max_lines: 400
  organization: by_feature

functions:
  max_lines: 50
  max_nesting: 4
  single_responsibility: true

error_handling:
  explicit: true
  user_friendly_messages: true

testing:
  write_tests: after_implementation
  coverage_target: 80%
```

### Transition Guide

- `@contexts/planning.md` → Before starting complex tasks
- `@contexts/debug.md` → When blocked by errors
- `@contexts/review.md` → Before committing

---

## planning.md - Planning Context

**Mode**: Strategic Planning & Design
**Focus**: Think before code, plan before act
**Principle**: Measure twice, cut once

### Core Principles

```
1. Understand scope   → Be clear about what needs to be done
2. Identify risks     → Identify potential issues in advance
3. Break into phases  → Decompose into smaller units
4. Get confirmation   → Approval required before proceeding
```

### Behavioral Rules

| DO | DON'T |
|----|-------|
| Thoroughly analyze requirements | Start coding before plan approval |
| Identify dependencies/blockers | Skip risk assessment |
| Consider multiple approaches | Underestimate complexity |
| Honestly estimate complexity | Ignore existing patterns |
| Document assumptions | Not recording assumptions |
| Wait for user confirmation | |

### Planning Process

```
┌─────────────────────────────────────────┐
│           Planning Workflow             │
├─────────────────────────────────────────┤
│  1. UNDERSTAND                          │
│     └─ What exactly needs to be done?   │
│              ↓                          │
│  2. ANALYZE                             │
│     └─ What exists? What's affected?    │
│              ↓                          │
│  3. DESIGN                              │
│     └─ How should we approach this?     │
│              ↓                          │
│  4. DECOMPOSE                           │
│     └─ Break into manageable phases     │
│              ↓                          │
│  5. ASSESS                              │
│     └─ Risks, dependencies, complexity  │
│              ↓                          │
│  6. PRESENT                             │
│     └─ Show plan, wait for approval     │
└─────────────────────────────────────────┘
```

### Complexity Estimation

| Level | Characteristics |
|-------|-----------------|
| **LOW** | Single file, < 100 lines, no dependencies |
| **MEDIUM** | 2-5 files, < 500 lines, some dependencies |
| **HIGH** | 5+ files, architecture changes, external dependencies |

### Risk Assessment Matrix

| Probability \ Impact | LOW | MEDIUM | HIGH |
|----------------------|-----|--------|------|
| **HIGH** | ⚠️ Monitor | 🔶 Plan B | 🔴 Blocker |
| **MEDIUM** | ✅ Accept | ⚠️ Monitor | 🔶 Plan B |
| **LOW** | ✅ Accept | ✅ Accept | ⚠️ Monitor |

### Transition Guide

- `@contexts/research.md` → When more understanding is needed first
- `@contexts/dev.md` → Ready to code after plan approval
- `@contexts/review.md` → Review plan before starting

---

## debug.md - Debug Context

**Mode**: Problem Investigation & Resolution
**Focus**: Root cause analysis, systematic debugging
**Principle**: Hypothesize → Test → Verify

### Core Principles

```
1. Reproduce first  → Confirm problem reproduction
2. Isolate the issue → Narrow down the cause
3. Find root cause  → Root cause, not surface symptoms
4. Verify the fix   → Confirm no recurrence after fix
```

### Behavioral Rules

| DO | DON'T |
|----|-------|
| Reproduce bug before investigating | Guess without evidence |
| Read error messages carefully | Apply fixes without understanding cause |
| Check recent changes (git log/diff) | Ignore stack traces |
| Binary search to isolate problems | Skip reproduction steps |
| Verify fix doesn't break other things | Fix symptoms instead of root cause |

### Debug Process

```
┌─────────────────────────────────────────┐
│           Debug Workflow                │
├─────────────────────────────────────────┤
│  1. REPRODUCE                           │
│     └─ Can we consistently trigger it?  │
│              ↓                          │
│  2. GATHER INFO                         │
│     └─ Error messages, logs, stack      │
│              ↓                          │
│  3. HYPOTHESIZE                         │
│     └─ What could cause this?           │
│              ↓                          │
│  4. ISOLATE                             │
│     └─ Binary search to narrow scope    │
│              ↓                          │
│  5. ROOT CAUSE                          │
│     └─ Why does this happen?            │
│              ↓                          │
│  6. FIX & VERIFY                        │
│     └─ Fix and confirm resolution       │
└─────────────────────────────────────────┘
```

### Common Bug Patterns

| Symptom | Expected Cause | How to Check |
|---------|----------------|--------------|
| "undefined is not..." | Null reference | Optional chaining, null checks |
| Infinite loop | Missing termination condition | Loop conditions, recursion base |
| Race condition | Async timing | await, Promise handling |
| Wrong data | Type mismatch | Input validation, type coercion |
| Silent failure | Swallowed error | try/catch, error handling |

### Quick Debugging Checklist

```markdown
□ Can it be reproduced consistently?
□ Have you read the full error message?
□ Have you checked the stack trace?
□ What changed recently?
□ Is the input data correct?
□ Are all dependencies loaded?
□ Is async/await handled correctly?
□ Are there null/undefined values?
```

### Transition Guide

- `@contexts/dev.md` → Ready to implement fix
- `@contexts/research.md` → When more system understanding is needed
- `@contexts/review.md` → Verify fix quality

---

## review.md - Review Context

**Mode**: Quality Analysis & PR Review
**Focus**: Security, maintainability, correctness
**Principle**: Suggest fixes, don't just criticize

### Core Principles

```
1. Read thoroughly   → Understand the full context
2. Prioritize issues → Categorize by severity
3. Suggest fixes     → Don't just point out problems, provide solutions
4. Be constructive   → Improvement suggestions, not criticism
```

### Behavioral Rules

| DO | DON'T |
|----|-------|
| Read all changes before commenting | Nitpick style with no readability impact |
| Prioritize by severity | Block PR for minor issues |
| Provide actionable fix suggestions | Criticize without alternatives |
| Check for security vulnerabilities | Ignore broader context |
| Verify test coverage for changes | Skip security checks |
| Acknowledge good patterns | |

### Review Checklist

#### Security (CRITICAL)
- No hardcoded secrets/credentials
- Input validation exists
- SQL Injection prevention
- XSS protection
- Authentication/authorization checks
- Sensitive data handling

#### Logic (HIGH)
- Edge case handling
- Error handling complete
- Null/undefined checks
- Race condition consideration
- Business logic correctness

#### Quality (MEDIUM)
- Code readability
- Function size (< 50 lines)
- Nesting depth (< 4 levels)
- DRY principle compliance
- Naming conventions

#### Testing (MEDIUM)
- Tests exist for new code
- Edge cases tested
- Existing tests still pass
- Coverage maintained

#### Performance (LOW)
- No obvious bottlenecks
- Efficient algorithms
- Memory considerations
- N+1 query prevention

### Severity Definitions

| Level | Impact | Required Action |
|-------|--------|-----------------|
| **CRITICAL** | Security breach, data loss | Block merge, fix immediately |
| **HIGH** | Bug, crash, logic error | Fix before merge |
| **MEDIUM** | Maintainability, quality | Fix recommended, can merge |
| **LOW** | Style, minor improvements | Optional, nice to have |

### Transition Guide

- `@contexts/dev.md` → Ready to fix issues
- `@contexts/research.md` → More understanding needed
- `@contexts/debug.md` → Investigate specific bug

---

## research.md - Research Context

**Mode**: Exploration & Investigation
**Focus**: Understanding before acting
**Principle**: Read widely, conclude carefully

### Core Principles

```
1. Understand first  → Understand the question exactly
2. Explore broadly   → Explore related code/documentation
3. Verify evidence   → Verify hypotheses with evidence
4. Summarize clearly → Organize findings
```

### Behavioral Rules

| DO | DON'T |
|----|-------|
| Read thoroughly before concluding | Write code before understanding is clear |
| Ask clarifying questions when uncertain | Draw conclusions without evidence |
| Document as you discover | Ignore edge cases or exceptions |
| Cross-reference multiple sources | Assume without verification |
| Record uncertainties and assumptions | Skip documentation review |

### Research Process

```
┌─────────────────────────────────────────┐
│           Research Workflow             │
├─────────────────────────────────────────┤
│  1. QUESTION                            │
│     └─ Clarify what we need to know     │
│              ↓                          │
│  2. EXPLORE                             │
│     └─ Search code, docs, patterns      │
│              ↓                          │
│  3. HYPOTHESIZE                         │
│     └─ Form initial understanding       │
│              ↓                          │
│  4. VERIFY                              │
│     └─ Test hypothesis with evidence    │
│              ↓                          │
│  5. DOCUMENT                            │
│     └─ Record findings and gaps         │
└─────────────────────────────────────────┘
```

### Tool Priority

| Priority | Tool | Purpose |
|----------|------|---------|
| 1 | Read | Deep exploration of specific files |
| 2 | Grep | Find patterns across codebase |
| 3 | Glob | Find file locations by pattern |
| 4 | Task (Explore) | Broad codebase questions |
| 5 | WebSearch | External documentation |
| 6 | WebFetch | URL verification, detailed information |

### Evidence Standards

| Claim Type | Required Evidence |
|------------|-------------------|
| "X does Y" | Code reference with line number |
| "Pattern is Z" | 3+ examples from codebase |
| "Best practice" | Official documentation or established convention |
| "Performance" | Benchmark or profiling data |

### Transition Guide

- `@contexts/planning.md` → Ready to plan implementation
- `@contexts/dev.md` → Ready to code
- `@contexts/debug.md` → Investigate a bug

---

## sync.md - Sync Context

**Mode**: Documentation Synchronization
**Focus**: Document generation, quality verification, git operations
**Methodology**: Sync → Verify → Commit

### Core Principles

```
1. Analyze changes  → Analyze changed code
2. Update docs      → Synchronize documentation
3. Verify quality   → Verify quality
4. Commit changes   → Commit changes
```

### Behavioral Rules

| DO | DON'T |
|----|-------|
| Analyze git changes before syncing | Regenerate unchanged documents |
| Update only affected documents | Skip quality verification |
| Verify link integrity after sync | Commit without review option |
| Follow TRUST 5 principles | Create unnecessary documents |
| Write meaningful commit messages | Break existing document links |
| Delegate to specialized agents | |

### Sync Phases

```
PHASE 0.5: Quality Verification
    ↓
PHASE 1: Analysis & Planning
    ↓
PHASE 2: Execute Sync
    ↓
PHASE 3: Git Operations
```

### Agent Delegation

```yaml
manager-docs:
  purpose: Document generation and updates
  tasks:
    - README sync
    - CODEMAP update
    - SPEC status sync
    - API documentation

manager-quality:
  purpose: Quality verification
  tasks:
    - TRUST 5 compliance check
    - Link integrity verification
    - Consistency check

manager-git:
  purpose: Git operations
  tasks:
    - Stage document files
    - Create commits
    - PR management (Team mode)
```

### TRUST 5 Checklist

```
T - Tested: All links work
R - Readable: Clear structure, proper formatting
U - Unified: Consistent terminology
S - Secured: No sensitive data exposure
T - Trackable: Version info, timestamps
```

### Transition Guide

- `@contexts/dev.md` → Continue development
- `@contexts/review.md` → Code review before sync
- `@contexts/planning.md` → Plan next feature

---

## Context Switching Guide

### Manual Switching

```bash
@.claude/contexts/dev.md Implement in this mode
@.claude/contexts/debug.md Analyze this error
@.claude/contexts/planning.md Plan this feature
```

### Context Flow

```
          ┌─────────────┐
          │  research   │  ← Starting point (understanding needed)
          └──────┬──────┘
                 │
                 ▼
          ┌─────────────┐
          │  planning   │  ← Plan development
          └──────┬──────┘
                 │
                 ▼
          ┌─────────────┐
          │    dev      │  ← Implementation
          └──────┬──────┘
                 │
        ┌────────┴────────┐
        │                 │
        ▼                 ▼
  ┌──────────┐     ┌──────────┐
  │  debug   │     │  review  │
  └──────────┘     └────┬─────┘
        │                │
        └────────┬───────┘
                 │
                 ▼
          ┌─────────────┐
          │    sync     │  ← Completion and documentation
          └─────────────┘
```

### Context Comparison

| Context | Purpose | Code Writing | Primary Tools |
|---------|---------|--------------|---------------|
| research | Understanding | ❌ | Read, Grep, WebSearch |
| planning | Planning | ❌ | Read, Grep, AskUser |
| dev | Implementation | ✅ | Edit, Write, Bash |
| debug | Investigation | Fixes only | Read, Grep, LSP |
| review | Verification | ❌ | Read, Grep, git diff |
| sync | Documentation | Docs only | Task, Write, Bash |

---

Version: 1.0.0
Last Updated: 2026-01-22
