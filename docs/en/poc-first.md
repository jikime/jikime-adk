# POC-First Development Workflow

> A phase-based development approach for greenfield features in JikiME-ADK.

## Overview

POC-First is a structured workflow for building new features from scratch. Instead of writing tests first (TDD) or analyzing existing behavior (DDD), you get it working first, then systematically improve.

### When to Use

| Scenario | Workflow | Why |
|----------|----------|-----|
| Brand new feature | **POC-First** | No behavior to preserve |
| New API endpoint | **POC-First** | Greenfield implementation |
| New UI component | **POC-First** | Build first, test after |
| Refactoring existing code | DDD | Must preserve behavior |
| Bug fix | TDD | Regression test first |
| Legacy migration | DDD | Characterize before changing |

---

## 5-Phase Structure

```
┌─────────────────────────────────────────────────────────────┐
│                  POC-First Workflow                          │
│                                                             │
│  Phase 1        Phase 2        Phase 3        Phase 4       │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌─────────┐  │
│  │ Make It  │──▶│ Refactor │──▶│ Testing  │──▶│ Quality │  │
│  │  Work    │   │          │   │          │   │  Gates  │  │
│  │ (50-60%) │   │ (15-20%) │   │ (15-20%) │   │(10-15%) │  │
│  └──────────┘   └──────────┘   └──────────┘   └─────────┘  │
│       │                                            │        │
│       │              Phase 5                       │        │
│       │         ┌──────────────┐                   │        │
│       └────────▶│ PR Lifecycle │◀──────────────────┘        │
│                 └──────────────┘                            │
└─────────────────────────────────────────────────────────────┘
```

### Phase 1: Make It Work (50-60% of effort)

**Goal**: Get the core functionality working end-to-end.

- Focus ONLY on making it work
- Hardcoding is acceptable
- Skip edge cases and error handling
- No premature optimization
- Use console.log freely for debugging

**Done when**: The feature works for the happy path. Demo-able to stakeholders.

### Phase 2: Refactor (15-20% of effort)

**Goal**: Clean up code without changing behavior.

- Extract hardcoded values to constants/config
- Split large files (>400 lines)
- Apply naming conventions
- Remove debug statements
- Add proper types

**Done when**: Code is clean, well-organized, and still works.

### Phase 3: Testing (15-20% of effort)

**Goal**: Add comprehensive test coverage.

- Unit tests for business logic (80%+ coverage)
- Integration tests for API endpoints
- E2E tests for critical user flows
- Edge case and error scenario coverage

**Done when**: Test suite passes, coverage meets threshold (80%+).

### Phase 4: Quality Gates (10-15% of effort)

**Goal**: Production-ready quality.

- Zero linting errors and type errors
- Security audit clean
- Performance acceptable
- Documentation complete

**Quality Gate Checklist**:
- [ ] Build passes with zero errors
- [ ] All tests passing, coverage >= 80%
- [ ] Lint clean (zero errors/warnings)
- [ ] Type check clean
- [ ] No critical/high security vulnerabilities
- [ ] Key functions documented

### Phase 5: PR Lifecycle

**Goal**: Create PR and get it merged.

Uses the [PR Lifecycle Automation](./pr-lifecycle.md) workflow for automated PR management.

---

## Intent Classification

Before starting, the workflow automatically classifies your intent:

```
Is this NEW code (no existing behavior)?
  ├── YES → POC-First Workflow
  └── NO  → Is existing code being modified?
        ├── YES → DDD Workflow (ANALYZE-PRESERVE-IMPROVE)
        └── Regression test needed? → TDD Workflow (RED-GREEN-REFACTOR)
```

---

## Phase Transition Rules

1. **No skipping phases**: Phase 1 → 2 → 3 → 4 → 5 (strict order)
2. **Phase gate required**: Each phase ends with a [VERIFY] checkpoint
3. **No regression**: Previous phase's verification must still pass
4. **User confirmation**: Phase 1 completion requires user confirmation

---

## Usage

### Slash Command

```bash
# Full POC workflow
/jikime:poc "User authentication with JWT"

# Start from specific phase
/jikime:poc "Auth system" --phase 3

# Skip PR phase
/jikime:poc "Auth system" --skip-pr
```

### As a Skill

The workflow is also available as a skill for use by other commands and agents:

```
Skill("jikime-workflow-poc")
```

---

## Task Format Integration

All tasks within the POC workflow use the [Structured Task Format](./task-format.md) with Do/Files/Done when/Verify/Commit fields and [VERIFY] quality checkpoints.

---

## Comparison with Other Workflows

| Aspect | POC-First | TDD | DDD |
|--------|-----------|-----|-----|
| **Starting point** | Working code | Failing test | Existing behavior analysis |
| **Test timing** | Phase 3 (after working) | Before implementation | Before modification |
| **Best for** | Greenfield features | New functions with clear specs | Refactoring existing code |
| **Risk** | Technical debt (mitigated by phases) | Over-testing | Over-analysis |
| **Speed to demo** | Fastest | Moderate | Slowest |

---

## Related Documentation

- [Structured Task Format](./task-format.md) — Task decomposition structure
- [PR Lifecycle Automation](./pr-lifecycle.md) — Phase 5 automation
- [TDD & DDD Workflow](./tdd-ddd.md) — Alternative development methodologies
- [Ralph Loop](./ralph-loop.md) — Iterative improvement within phases
