# Structured Task Format

> Systematic task decomposition with 5-field structure and quality checkpoints.

## Overview

Every implementation task in JikiME-ADK can be decomposed into structured units with clear acceptance criteria. This format ensures nothing is forgotten, progress is trackable, and quality is verified at regular intervals.

---

## Task Structure (5 Fields)

Each task follows this mandatory format:

```
### Task N: [Title]

Do:        What to implement (specific action)
Files:     Which files to create or modify
Done when: Measurable acceptance criteria
Verify:    How to verify completion (test command, manual check)
Commit:    Commit message when task is complete
```

### Field Descriptions

| Field | Required | Description |
|-------|----------|-------------|
| **Do** | Yes | Single, actionable implementation step |
| **Files** | Yes | Explicit file paths (create/modify/delete) |
| **Done when** | Yes | Measurable criteria (test passes, build succeeds, output matches) |
| **Verify** | Yes | Concrete verification command or check |
| **Commit** | Yes | Conventional commit message (feat/fix/refactor/test/docs) |

### Example

```
### Task 1: Create user authentication API

Do:        Implement POST /api/auth/login endpoint with JWT token generation
Files:     src/api/auth/login.ts (create), src/types/auth.ts (create)
Done when: POST /api/auth/login returns 200 with valid JWT for correct credentials, 401 for invalid
Verify:    npm test -- --grep "auth login" passes
Commit:    feat(auth): add login endpoint with JWT generation
```

---

## Quality Checkpoints ([VERIFY])

Insert a `[VERIFY]` checkpoint task after every 2-3 implementation tasks.

### Format

```
### Task N: [VERIFY] Quality Checkpoint

Do:        Run full verification suite
Files:     (none - verification only)
Done when: All checks pass with zero errors
Verify:
  - Build passes (zero errors)
  - Tests pass (all green)
  - Lint clean (zero warnings)
  - Manual: feature works as expected
Commit:    (no commit - checkpoint only)
```

### Checkpoint Rules

| Rule | Description |
|------|-------------|
| **Frequency** | Insert [VERIFY] after every 2-3 tasks |
| **Blocking** | Do NOT proceed if checkpoint fails |
| **Fix First** | Create fix tasks before continuing |
| **No Skip** | [VERIFY] tasks cannot be skipped or deferred |
| **Final** | Always insert [VERIFY] as the last task |

### Insertion Pattern

```
Task 1: Implementation
Task 2: Implementation
Task 3: Implementation
Task 4: [VERIFY] Quality Checkpoint     ← after 3 tasks
Task 5: Implementation
Task 6: Implementation
Task 7: [VERIFY] Quality Checkpoint     ← after 2 tasks
...
Task N: [VERIFY] Final Checkpoint       ← always at the end
```

---

## DDD Task Variant

When working with existing code (DDD workflow), tasks use ANALYZE/PRESERVE/IMPROVE prefixes:

```
Task 1: [ANALYZE] Understand current auth flow
Task 2: [PRESERVE] Add characterization tests
Task 3: [IMPROVE] Refactor to JWT-based auth
Task 4: [VERIFY] Quality Checkpoint
```

---

## Task Sizing Guidelines

| Size | Description | Example |
|------|-------------|---------|
| **XS** | Single function/method | Add validation helper |
| **S** | Single file change | Create API endpoint |
| **M** | 2-3 file changes | Feature with tests |
| **L** | 4+ file changes | **Split into smaller tasks** |

**Rule**: If a task touches more than 3 files, split it into smaller tasks.

---

## TodoWrite Integration

All tasks are tracked using TodoWrite for real-time progress:

```
Task discovered  → TodoWrite: add with "pending" status
Task started     → TodoWrite: change to "in_progress"
Task completed   → TodoWrite: change to "completed"
[VERIFY] failed  → TodoWrite: add fix tasks with "pending"
```

---

## Integration with Workflows

| Workflow | How Task Format is Used |
|----------|------------------------|
| **POC-First** | Each phase generates tasks in this format |
| **TDD** | RED-GREEN-REFACTOR maps to task sequences |
| **DDD** | ANALYZE-PRESERVE-IMPROVE maps to task prefixes |
| **Ralph Loop** | Loop iterations track progress via TodoWrite |

---

## Related Documentation

- [POC-First Workflow](./poc-first.md) — Phase-based greenfield development
- [PR Lifecycle Automation](./pr-lifecycle.md) — PR management automation
- [TDD & DDD Workflow](./tdd-ddd.md) — Alternative development methodologies
