# TDD & DDD Workflow Guide

> A comprehensive guide to Test-Driven Development (TDD) and Domain-Driven Development (DDD) workflows in JikiME-ADK.

## Overview

JikiME-ADK supports two core development methodologies:

| Methodology | Abbreviation | Cycle | When to Apply |
|-------------|--------------|-------|---------------|
| **Test-Driven Development** | TDD | RED → GREEN → REFACTOR | New feature development |
| **Domain-Driven Development** | DDD | ANALYZE → PRESERVE → IMPROVE | Refactoring existing code |

---

## TDD (Test-Driven Development)

### Core Cycle: RED → GREEN → REFACTOR

```
┌─────────────────────────────────────────────────────────┐
│                    TDD Cycle                            │
│                                                         │
│    ┌─────────┐     ┌─────────┐     ┌──────────┐        │
│    │  RED    │ ──▶ │  GREEN  │ ──▶ │ REFACTOR │        │
│    │ (Fail)  │     │ (Pass)  │     │ (Improve)│        │
│    └─────────┘     └─────────┘     └──────────┘        │
│         │                               │               │
│         └───────────────────────────────┘               │
│                    Repeat                               │
└─────────────────────────────────────────────────────────┘
```

### TDD Step-by-Step Explanation

**1. RED Phase (Write a Failing Test)**
- Write the test first for the feature you want to implement
- Verify that the test fails (since there's no implementation yet)
- Define clear expected values and inputs

```typescript
// RED: Write test first
describe('Calculator', () => {
  it('should add two numbers', () => {
    const result = calculator.add(2, 3);
    expect(result).toBe(5); // Fails - add method doesn't exist
  });
});
```

**2. GREEN Phase (Minimal Implementation)**
- Write the minimum code to make the test pass
- It doesn't need to be perfect code
- The only goal is to pass the test

```typescript
// GREEN: Minimal implementation to pass the test
class Calculator {
  add(a: number, b: number): number {
    return a + b; // Passes!
  }
}
```

**3. REFACTOR Phase (Code Improvement)**
- Improve the code while maintaining passing tests
- Remove duplication, improve readability, optimize performance
- Run tests after refactoring to verify

### TDD Related Resources

| Resource | Path | Description |
|----------|------|-------------|
| TDD Skill | `.claude/skills/jikime-workflow-tdd/SKILL.md` | TDD workflow skill |
| Testing Skill | `.claude/skills/jikime-workflow-testing/SKILL.md` | Comprehensive testing skill |

### TDD Core Principles

**FIRST Principles**:
- **F**ast: Tests should run quickly
- **I**ndependent: Tests should be independent of each other
- **R**epeatable: Same results in any environment
- **S**elf-validating: Clear pass/fail determination
- **T**imely: Write tests before code

**AAA Pattern**:
```typescript
it('should validate user input', () => {
  // Arrange - Test setup
  const input = { email: 'test@example.com' };

  // Act - Execute action
  const result = validateInput(input);

  // Assert - Verify result
  expect(result.isValid).toBe(true);
});
```

---

## DDD (Domain-Driven Development)

### Core Cycle: ANALYZE → PRESERVE → IMPROVE

```
┌─────────────────────────────────────────────────────────┐
│                    DDD Cycle                            │
│                                                         │
│    ┌─────────┐     ┌──────────┐     ┌─────────┐        │
│    │ ANALYZE │ ──▶ │ PRESERVE │ ──▶ │ IMPROVE │        │
│    │(Analysis)│     │(Preserve)│     │(Improve)│        │
│    └─────────┘     └──────────┘     └─────────┘        │
│         │                               │               │
│         └───────────────────────────────┘               │
│              Behavior Preservation Verification         │
└─────────────────────────────────────────────────────────┘
```

### DDD Step-by-Step Explanation

**1. ANALYZE Phase (Current Code Analysis)**
- Identify domain boundaries
- Calculate coupling/cohesion metrics
- Analyze AST structure
- Assess technical debt

```typescript
// ANALYZE: Code analysis
// - Dependency graph mapping
// - Code smell detection (God Class, Feature Envy, etc.)
// - Prioritize refactoring targets
```

**2. PRESERVE Phase (Behavior Preservation Test Generation)**
- Generate characterization tests
- Capture current behavior as "golden standard"
- Build test safety net

```typescript
// PRESERVE: Characterization test
describe('ExistingBehavior', () => {
  it('should preserve current calculation logic', () => {
    // Capture current behavior as-is
    const result = existingFunction(input);
    expect(result).toMatchSnapshot();
  });
});
```

**3. IMPROVE Phase (Safe Refactoring)**
- Maintain passing test state
- Make gradual, small changes
- Run tests after each change

### DDD Related Resources

| Resource | Path | Description |
|----------|------|-------------|
| DDD Skill | `.claude/skills/jikime-workflow-ddd/SKILL.md` | DDD workflow skill |
| DDD Agent | `.claude/agents/jikime/manager-ddd.md` | DDD specialist agent |
| Context7 Module | `.claude/skills/jikime-workflow-testing/modules/ddd-context7.md` | Context7 integration module |

### DDD Refactoring Strategies

| Strategy | When to Apply | Description |
|----------|---------------|-------------|
| Extract Method | Long methods, duplicate code | Extract code block into separate method |
| Extract Class | Classes with multiple responsibilities | Separate responsibilities into new class |
| Move Method | Feature Envy | Move method to appropriate class |
| Inline | Unnecessary indirection | Remove excessive abstraction |
| Rename | Lack of clarity | Safe renaming with AST-grep |

---

## TDD vs DDD: When to Use Which?

### Decision Flowchart

```
                    ┌─────────────────────────┐
                    │ Does the code you want  │
                    │ to change already exist?│
                    └───────────┬─────────────┘
                                │
                    ┌───────────┴───────────┐
                    │                       │
                   YES                      NO
                    │                       │
                    ▼                       ▼
            ┌───────────────┐       ┌───────────────┐
            │   Use DDD     │       │   Use TDD     │
            │               │       │               │
            │ ANALYZE       │       │ RED           │
            │ PRESERVE      │       │ GREEN         │
            │ IMPROVE       │       │ REFACTOR      │
            └───────────────┘       └───────────────┘
```

### Detailed Comparison Table

| Aspect | TDD | DDD |
|--------|-----|-----|
| **Purpose** | New feature development | Existing code improvement |
| **Starting Point** | Writing tests | Code analysis |
| **Test Role** | Requirements definition | Behavior preservation verification |
| **Scope of Change** | Adding new code | Structural changes |
| **Risk** | Low (new code) | High (affects existing behavior) |
| **Success Criteria** | All tests pass | Existing tests + new tests pass |

### Usage Scenarios

**When to use TDD:**
- Developing new features from scratch
- When you have clear requirements
- When there's no existing code to preserve
- When you want to define API contracts first

**When to use DDD:**
- Refactoring legacy code
- When you want to reduce technical debt
- When improving structure while maintaining existing behavior
- When adding tests to untested code

---

## Context7 Integration

### AI-Based Test Generation

The DDD workflow can access the latest testing patterns through Context7 MCP:

```typescript
// Load patterns through Context7
// mcp__context7__resolve-library-id: "vitest testing patterns"
// mcp__context7__query-docs: "mocking best practices", libraryId: resolved_id
```

### Testing Queries by Supported Language

| Language | Context7 Query Example |
|----------|------------------------|
| TypeScript | `"vitest typescript testing patterns"` |
| Python | `"pytest best practices"` |
| Go | `"go testing patterns"` |
| Rust | `"rust testing cargo"` |

---

## Agents and Skills Structure

### Related Agents

```
┌─────────────────────────────────────────────────────────┐
│                   Agent Hierarchy                        │
│                                                         │
│  ┌─────────────┐     ┌─────────────┐                   │
│  │ manager-ddd │     │ test-guide  │                   │
│  │             │◀───▶│             │                   │
│  │ DDD Cycle   │     │ Test        │                   │
│  │Orchestration│     │ Generation  │                   │
│  └─────────────┘     └─────────────┘                   │
│         │                   │                           │
│         └─────────┬─────────┘                           │
│                   │                                     │
│                   ▼                                     │
│         ┌─────────────────┐                            │
│         │ refactorer      │                            │
│         │                 │                            │
│         │ Code Refactoring│                            │
│         └─────────────────┘                            │
└─────────────────────────────────────────────────────────┘
```

### Related Skills

```yaml
# DDD/TDD Related Skills
jikime-workflow-tdd:     # TDD RED-GREEN-REFACTOR cycle
jikime-workflow-ddd:     # DDD ANALYZE-PRESERVE-IMPROVE cycle
jikime-workflow-testing: # Comprehensive testing workflow
  modules:
    - ddd-context7.md    # Context7 integration
    - vitest.md          # Vitest testing
    - playwright.md      # E2E testing
```

---

## Quality Metrics

### DDD Success Criteria

**Behavior Preservation (Required)**:
- All existing tests pass: 100%
- All characterization tests pass: 100%
- No API contract changes
- Performance range maintained

**Structural Improvement (Goals)**:

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Coupling (Ce) | - | - | Decrease |
| Cohesion | - | - | Increase |
| Complexity | - | - | Decrease |
| Technical Debt | - | - | Decrease |

### TDD Success Criteria

- All new tests pass
- Test coverage goals achieved
- FIRST principles followed
- Code quality standards met

---

## Command Usage

### /jikime:2-run with DDD

```bash
# Execute SPEC in DDD mode
/jikime:2-run SPEC-001
# → manager-ddd agent executes ANALYZE-PRESERVE-IMPROVE cycle
```

### Direct DDD Request

```bash
# Request specific code refactoring
"Refactor @src/services/user.ts using DDD approach"
# → manager-ddd agent automatically activated
```

---

## Advanced Features

### Property-Based Testing

```typescript
import * as fc from 'fast-check';

describe('Addition Properties', () => {
  it('should be commutative', () => {
    fc.assert(fc.property(
      fc.integer(), fc.integer(),
      (a, b) => add(a, b) === add(b, a)
    ));
  });
});
```

### Mutation Testing

```bash
# TypeScript with Stryker
npx stryker run

# Python with mutmut
mutmut run
```

### Continuous Testing

```json
// package.json
{
  "scripts": {
    "test:watch": "vitest --watch",
    "test:coverage": "vitest --coverage"
  }
}
```

---

## Troubleshooting

### Common Issues

**1. When Characterization Tests are Unstable**
- Check for non-deterministic causes (time, random, external state)
- Mock external dependencies
- Strengthen test isolation

**2. Test Failures After Refactoring**
- Rollback immediately
- Analyze the cause
- Retry with smaller transformation steps

**3. Context7 Connection Issues**
- Check MCP settings
- Verify network connection
- Fall back to default patterns

---

## Related Documentation

- [Sync Workflow Guide](./sync.md)
- [SPEC System Guide](./spec.md)
- [Quality Gate Guide](./quality.md)

---

Version: 1.0.0
Last Updated: 2026-01-22
