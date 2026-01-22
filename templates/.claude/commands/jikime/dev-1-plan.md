---
description: "[Step 1/4] Create SPEC document with EARS format. Analyze requirements, create 3-file structure, WAIT for confirmation."
context: planning
---

# Development Step 1: Plan (SPEC Creation)

**Context**: @.claude/contexts/planning.md (Auto-loaded)

**계획 단계**: 요구사항을 분석하고 EARS 형식의 SPEC 문서를 생성합니다.

**Note**: 이 단계에서 사용자 확인을 받기 전까지 코드를 작성하지 않습니다.

## Usage

```bash
# Create new SPEC document
/jikime:dev-1-plan Add user authentication

# Create SPEC with domain specification
/jikime:dev-1-plan --domain AUTH Add JWT-based login

# Create SPEC with context reference
/jikime:dev-1-plan @src/services/ Add caching layer

# Use existing SPEC document
/jikime:dev-1-plan SPEC-AUTH-001
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `[description]` | Feature or task description | Required |
| `@path` | Reference existing code | - |
| `--domain` | Specify SPEC domain (AUTH, USER, API, etc.) | Auto-detect |
| `[SPEC-ID]` | Use existing SPEC document | - |

## Process

```
1. Analyze Request
   - Understand requirements
   - Identify domain and affected areas
   - Detect technology keywords
        ↓
2. Delegate to manager-spec Agent
   - Use the manager-spec subagent to create SPEC document
   - Manager-spec creates 3-file structure in .jikime/specs/
        ↓
3. Create SPEC Directory Structure
   .jikime/specs/SPEC-{DOMAIN}-{NUMBER}/
   ├── spec.md         (EARS requirements)
   ├── plan.md         (Implementation plan)
   └── acceptance.md   (Acceptance criteria)
        ↓
4. Present & WAIT
   - Show SPEC summary to user
   - Wait for confirmation
        ↓
5. User Response
   - "yes" → Proceed to dev-2-implement
   - "modify: [changes]" → Revise SPEC
   - "no" → Cancel
```

## SPEC ID Format

- **Pattern**: `SPEC-{DOMAIN}-{NUMBER}`
- **Domain**: Uppercase letters indicating feature area
- **Number**: 3-digit zero-padded sequential number

### Domain Examples

| Domain | Description |
|--------|-------------|
| AUTH | Authentication, authorization |
| USER | User management |
| API | API endpoints |
| DB | Database operations |
| UI | User interface |
| PERF | Performance optimization |
| SEC | Security features |

## 3-File Structure

### spec.md (Requirements)

EARS format specification containing:
- Metadata (ID, Title, Status, Priority)
- Environment and Assumptions
- Requirements (Ubiquitous, Event-driven, State-driven, Unwanted, Optional)
- Specifications and Traceability

### plan.md (Implementation Plan)

- Milestones by priority (Primary, Secondary, Optional)
- Technical approach and architecture
- Implementation phases
- Risks and mitigations

### acceptance.md (Acceptance Criteria)

- Success criteria
- Test scenarios (Given-When-Then format)
- Quality gates
- Definition of Done

## Output Example

```markdown
## SPEC Creation Complete: SPEC-AUTH-001

**Status**: SUCCESS
**Title**: JWT-based User Authentication

### Created Files
- `.jikime/specs/SPEC-AUTH-001/spec.md` (EARS format)
- `.jikime/specs/SPEC-AUTH-001/plan.md` (Implementation plan)
- `.jikime/specs/SPEC-AUTH-001/acceptance.md` (Acceptance criteria)

### Requirements Summary
- Ubiquitous: 3 requirements
- Event-driven: 2 requirements
- State-driven: 1 requirement
- Unwanted: 2 requirements

### Quality Verification
- EARS Syntax: PASS
- Completeness: 100%
- Traceability Tags: Applied

---

**WAITING FOR CONFIRMATION**

Proceed to implementation? (yes/no/modify)
- "yes" → Run /jikime:dev-2-implement SPEC-AUTH-001
- "modify: [changes]" → Revise SPEC
- "no" → Cancel
```

## Agent Delegation

[HARD] This command MUST delegate SPEC creation to manager-spec agent:

```
Use the manager-spec subagent to create SPEC document for: {description}

Context:
- Domain: {detected or specified domain}
- Referenced code: {if @path provided}
- Requirements: {user description}

Expected output:
- 3-file SPEC structure in .jikime/specs/
- EARS-formatted requirements
- Implementation plan with milestones
- Acceptance criteria with test scenarios
```

## Critical Rules

1. **MUST WAIT** - 사용자 확인 전 코드 작성 금지
2. **3-File Structure** - 반드시 3파일 구조로 생성
3. **EARS Format** - 요구사항은 EARS 패턴 사용
4. **No Flat Files** - `.jikime/specs/SPEC-*.md` 단일 파일 금지
5. **Delegate to Agent** - manager-spec 에이전트에 위임

## Workflow

```
/jikime:dev-0-init   (Optional: Project initialization)
        ↓
/jikime:dev-1-plan  ← Current (SPEC Creation)
        ↓
/jikime:dev-2-implement (DDD Implementation)
        ↓
/jikime:dev-3-test (Testing)
        ↓
/jikime:dev-4-review (Code Review)
```

## Next Step

승인 후 다음 단계로:
```bash
/jikime:dev-2-implement SPEC-AUTH-001
```

---

Version: 3.0.0 (3-File SPEC Structure)
