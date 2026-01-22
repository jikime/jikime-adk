# TDD & DDD 워크플로우 가이드

> JikiME-ADK의 테스트 주도 개발(TDD)과 도메인 주도 개발(DDD) 워크플로우에 대한 종합 가이드입니다.

## 개요

JikiME-ADK는 두 가지 핵심 개발 방법론을 지원합니다:

| 방법론 | 약어 | 사이클 | 적용 시점 |
|--------|------|--------|-----------|
| **Test-Driven Development** | TDD | RED → GREEN → REFACTOR | 새로운 기능 개발 |
| **Domain-Driven Development** | DDD | ANALYZE → PRESERVE → IMPROVE | 기존 코드 리팩토링 |

---

## TDD (Test-Driven Development)

### 핵심 사이클: RED → GREEN → REFACTOR

```
┌─────────────────────────────────────────────────────────┐
│                    TDD Cycle                            │
│                                                         │
│    ┌─────────┐     ┌─────────┐     ┌──────────┐        │
│    │  RED    │ ──▶ │  GREEN  │ ──▶ │ REFACTOR │        │
│    │ (실패)  │     │ (통과)  │     │ (개선)   │        │
│    └─────────┘     └─────────┘     └──────────┘        │
│         │                               │               │
│         └───────────────────────────────┘               │
│                    반복                                 │
└─────────────────────────────────────────────────────────┘
```

### TDD 단계별 설명

**1. RED Phase (실패하는 테스트 작성)**
- 구현하려는 기능의 테스트를 먼저 작성
- 테스트가 실패하는지 확인 (아직 구현이 없으므로)
- 명확한 기대값과 입력을 정의

```typescript
// RED: 테스트 먼저 작성
describe('Calculator', () => {
  it('should add two numbers', () => {
    const result = calculator.add(2, 3);
    expect(result).toBe(5); // 실패 - add 메서드 없음
  });
});
```

**2. GREEN Phase (최소한의 구현)**
- 테스트를 통과시키는 최소한의 코드 작성
- 완벽한 코드가 아니어도 됨
- 목표는 오직 테스트 통과

```typescript
// GREEN: 테스트 통과를 위한 최소 구현
class Calculator {
  add(a: number, b: number): number {
    return a + b; // 통과!
  }
}
```

**3. REFACTOR Phase (코드 개선)**
- 테스트가 통과하는 상태를 유지하면서 코드 개선
- 중복 제거, 가독성 향상, 성능 최적화
- 리팩토링 후 테스트 재실행으로 검증

### TDD 관련 리소스

| 리소스 | 경로 | 설명 |
|--------|------|------|
| TDD Skill | `.claude/skills/jikime-workflow-tdd/SKILL.md` | TDD 워크플로우 스킬 |
| Testing Skill | `.claude/skills/jikime-workflow-testing/SKILL.md` | 종합 테스팅 스킬 |

### TDD 핵심 원칙

**FIRST 원칙**:
- **F**ast: 테스트는 빠르게 실행되어야 함
- **I**ndependent: 테스트는 서로 독립적이어야 함
- **R**epeatable: 어떤 환경에서도 동일한 결과
- **S**elf-validating: 성공/실패를 명확히 판단
- **T**imely: 코드 작성 전에 테스트 작성

**AAA 패턴**:
```typescript
it('should validate user input', () => {
  // Arrange - 테스트 준비
  const input = { email: 'test@example.com' };

  // Act - 동작 실행
  const result = validateInput(input);

  // Assert - 결과 검증
  expect(result.isValid).toBe(true);
});
```

---

## DDD (Domain-Driven Development)

### 핵심 사이클: ANALYZE → PRESERVE → IMPROVE

```
┌─────────────────────────────────────────────────────────┐
│                    DDD Cycle                            │
│                                                         │
│    ┌─────────┐     ┌──────────┐     ┌─────────┐        │
│    │ ANALYZE │ ──▶ │ PRESERVE │ ──▶ │ IMPROVE │        │
│    │ (분석)  │     │ (보존)   │     │ (개선)  │        │
│    └─────────┘     └──────────┘     └─────────┘        │
│         │                               │               │
│         └───────────────────────────────┘               │
│              동작 보존 검증                              │
└─────────────────────────────────────────────────────────┘
```

### DDD 단계별 설명

**1. ANALYZE Phase (현재 코드 분석)**
- 도메인 경계 식별
- 결합도/응집도 지표 계산
- AST 구조 분석
- 기술 부채 평가

```typescript
// ANALYZE: 코드 분석
// - 의존성 그래프 매핑
// - 코드 스멜 감지 (God Class, Feature Envy 등)
// - 리팩토링 대상 우선순위 지정
```

**2. PRESERVE Phase (동작 보존 테스트 생성)**
- 특성화 테스트(Characterization Test) 생성
- 현재 동작을 "골든 스탠다드"로 캡처
- 테스트 안전망 구축

```typescript
// PRESERVE: 특성화 테스트
describe('ExistingBehavior', () => {
  it('should preserve current calculation logic', () => {
    // 현재 동작을 그대로 캡처
    const result = existingFunction(input);
    expect(result).toMatchSnapshot();
  });
});
```

**3. IMPROVE Phase (안전한 리팩토링)**
- 테스트가 통과하는 상태 유지
- 점진적이고 작은 변경
- 각 변경 후 테스트 실행

### DDD 관련 리소스

| 리소스 | 경로 | 설명 |
|--------|------|------|
| DDD Skill | `.claude/skills/jikime-workflow-ddd/SKILL.md` | DDD 워크플로우 스킬 |
| DDD Agent | `.claude/agents/jikime/manager-ddd.md` | DDD 전문 에이전트 |
| Context7 Module | `.claude/skills/jikime-workflow-testing/modules/ddd-context7.md` | Context7 통합 모듈 |

### DDD 리팩토링 전략

| 전략 | 적용 시점 | 설명 |
|------|-----------|------|
| Extract Method | 긴 메서드, 중복 코드 | 코드 블록을 별도 메서드로 추출 |
| Extract Class | 다중 책임 클래스 | 책임 분리하여 새 클래스 생성 |
| Move Method | Feature Envy | 메서드를 적절한 클래스로 이동 |
| Inline | 불필요한 간접 참조 | 과도한 추상화 제거 |
| Rename | 명확성 부족 | AST-grep으로 안전한 이름 변경 |

---

## TDD vs DDD: 언제 무엇을 사용할까?

### 결정 흐름도

```
                    ┌─────────────────────────┐
                    │ 변경하려는 코드가       │
                    │ 이미 존재하는가?        │
                    └───────────┬─────────────┘
                                │
                    ┌───────────┴───────────┐
                    │                       │
                   YES                      NO
                    │                       │
                    ▼                       ▼
            ┌───────────────┐       ┌───────────────┐
            │   DDD 사용    │       │   TDD 사용    │
            │               │       │               │
            │ ANALYZE       │       │ RED           │
            │ PRESERVE      │       │ GREEN         │
            │ IMPROVE       │       │ REFACTOR      │
            └───────────────┘       └───────────────┘
```

### 상세 비교표

| 측면 | TDD | DDD |
|------|-----|-----|
| **목적** | 새 기능 개발 | 기존 코드 개선 |
| **시작점** | 테스트 작성 | 코드 분석 |
| **테스트 역할** | 요구사항 정의 | 동작 보존 검증 |
| **변경 범위** | 새 코드 추가 | 구조 변경 |
| **리스크** | 낮음 (새 코드) | 높음 (기존 동작 영향) |
| **성공 기준** | 모든 테스트 통과 | 기존 테스트 + 새 테스트 통과 |

### 사용 시나리오

**TDD를 사용해야 할 때:**
- 처음부터 새 기능을 개발할 때
- 명확한 요구사항이 있을 때
- 보존할 기존 코드가 없을 때
- API 계약을 먼저 정의하고 싶을 때

**DDD를 사용해야 할 때:**
- 레거시 코드를 리팩토링할 때
- 기술 부채를 줄이고 싶을 때
- 기존 동작을 유지하면서 구조를 개선할 때
- 테스트 없는 코드에 테스트를 추가할 때

---

## Context7 통합

### AI 기반 테스트 생성

DDD 워크플로우는 Context7 MCP를 통해 최신 테스팅 패턴에 접근할 수 있습니다:

```typescript
// Context7을 통한 패턴 로드
// mcp__context7__resolve-library-id: "vitest testing patterns"
// mcp__context7__query-docs: "mocking best practices", libraryId: resolved_id
```

### 지원 언어별 테스팅 쿼리

| 언어 | Context7 쿼리 예시 |
|------|-------------------|
| TypeScript | `"vitest typescript testing patterns"` |
| Python | `"pytest best practices"` |
| Go | `"go testing patterns"` |
| Rust | `"rust testing cargo"` |

---

## 에이전트 및 스킬 구조

### 관련 에이전트

```
┌─────────────────────────────────────────────────────────┐
│                   Agent Hierarchy                        │
│                                                         │
│  ┌─────────────┐     ┌─────────────┐                   │
│  │ manager-ddd │     │ expert-     │                   │
│  │             │◀───▶│ testing     │                   │
│  │ DDD 사이클  │     │ 테스트 생성  │                   │
│  │ 오케스트레이션│     │             │                   │
│  └─────────────┘     └─────────────┘                   │
│         │                   │                           │
│         └─────────┬─────────┘                           │
│                   │                                     │
│                   ▼                                     │
│         ┌─────────────────┐                            │
│         │ expert-         │                            │
│         │ refactoring     │                            │
│         │ 코드 리팩토링     │                            │
│         └─────────────────┘                            │
└─────────────────────────────────────────────────────────┘
```

### 관련 스킬

```yaml
# DDD/TDD 관련 스킬
jikime-workflow-tdd:     # TDD RED-GREEN-REFACTOR 사이클
jikime-workflow-ddd:     # DDD ANALYZE-PRESERVE-IMPROVE 사이클
jikime-workflow-testing: # 종합 테스팅 워크플로우
  modules:
    - ddd-context7.md    # Context7 통합
    - vitest.md          # Vitest 테스팅
    - playwright.md      # E2E 테스팅
```

---

## 품질 지표

### DDD 성공 기준

**동작 보존 (필수)**:
- 모든 기존 테스트 통과: 100%
- 모든 특성화 테스트 통과: 100%
- API 계약 변경 없음
- 성능 범위 유지

**구조 개선 (목표)**:

| 지표 | Before | After | 변화 |
|------|--------|-------|------|
| 결합도 (Ce) | - | - | 감소 |
| 응집도 | - | - | 증가 |
| 복잡도 | - | - | 감소 |
| 기술 부채 | - | - | 감소 |

### TDD 성공 기준

- 모든 신규 테스트 통과
- 테스트 커버리지 목표 달성
- FIRST 원칙 준수
- 코드 품질 기준 충족

---

## 명령어 사용법

### /jikime:2-run with DDD

```bash
# DDD 모드로 SPEC 실행
/jikime:2-run SPEC-001
# → manager-ddd 에이전트가 ANALYZE-PRESERVE-IMPROVE 사이클 실행
```

### 직접 DDD 요청

```bash
# 특정 코드 리팩토링 요청
"@src/services/user.ts 파일을 DDD 방식으로 리팩토링해줘"
# → manager-ddd 에이전트 자동 활성화
```

---

## 고급 기능

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

## 트러블슈팅

### 일반적인 문제

**1. 특성화 테스트가 불안정할 때**
- 비결정성 원인 확인 (시간, 랜덤, 외부 상태)
- 외부 의존성 목(Mock) 처리
- 테스트 격리 강화

**2. 리팩토링 후 테스트 실패**
- 즉시 롤백
- 원인 분석
- 더 작은 변환 단계로 재시도

**3. Context7 연결 문제**
- MCP 설정 확인
- 네트워크 연결 확인
- 기본 패턴으로 폴백

---

## 관련 문서

- [Sync 워크플로우 가이드](./sync.md)
- [SPEC 시스템 가이드](./spec.md)
- [품질 게이트 가이드](./quality.md)

---

Version: 1.0.0
Last Updated: 2026-01-22
