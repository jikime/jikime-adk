# JikiME-ADK Rules Reference

JikiME-ADK의 규칙 시스템 문서입니다.

---

## 개요

JikiME-ADK는 13개의 규칙 파일을 통해 일관된 개발 표준을 제공합니다:

### 규칙 맵

```
┌─────────────────────────────────────────────────────────────────┐
│                     JikiME-ADK Rules                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ Core Rules (필수) ───────────────────────────────────────┐  │
│  │                                                            │  │
│  │  core.md          HARD 규칙 (언어, 실행, 출력)             │  │
│  │  agents.md        에이전트 위임 규칙                       │  │
│  │  quality.md       품질 게이트                              │  │
│  │  interaction.md   사용자 상호작용                          │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ Development Rules (개발) ────────────────────────────────┐  │
│  │                                                            │  │
│  │  coding-style.md  코딩 스타일                              │  │
│  │  git-workflow.md  Git 워크플로우                           │  │
│  │  testing.md       테스트 가이드라인                         │  │
│  │  security.md      보안 가이드라인                          │  │
│  │  patterns.md      공통 패턴                                │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ System Rules (시스템) ───────────────────────────────────┐  │
│  │                                                            │  │
│  │  hooks.md         Hook 시스템                              │  │
│  │  performance.md   성능 최적화                              │  │
│  │  skills.md        스킬 발견/관리                           │  │
│  │  web-search.md    웹 검색 프로토콜                         │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Core Rules

### core.md - 핵심 규칙

**목적**: 반드시 준수해야 하는 HARD 규칙 정의

#### 언어 규칙

| 규칙 | 설명 |
|------|------|
| [HARD] Language-Aware Responses | 모든 사용자 응답은 `conversation_language`로 |
| [HARD] Internal Communication | 에이전트 간 통신은 영어 |
| [HARD] Code Comments | `code_comments` 설정 따름 (기본: 영어) |

#### 실행 규칙

| 규칙 | 설명 |
|------|------|
| [HARD] Parallel Execution | 의존성 없는 독립 도구 호출은 병렬 실행 |
| [HARD] No XML in User Responses | XML 태그는 사용자에게 표시 금지 |

#### 출력 형식 규칙

| 규칙 | 설명 |
|------|------|
| [HARD] Markdown Required | 사용자 응답에 항상 Markdown 사용 |
| [HARD] XML Reserved | XML 태그는 내부 에이전트 데이터 전송용으로만 |

#### 체크리스트

응답 전 확인:

- [ ] 응답이 사용자의 `conversation_language`로 작성됨
- [ ] 독립 작업이 병렬화됨
- [ ] 응답에 XML 태그 없음
- [ ] Markdown 형식 적용됨
- [ ] URL이 포함 전 검증됨

---

### agents.md - 에이전트 위임 규칙

**목적**: 에이전트 위임 시점과 방법 정의

#### 명령 유형별 규칙

##### Type A: Workflow Commands

**명령어**: `/jikime:0-project`, `/jikime:1-plan`, `/jikime:2-run`, `/jikime:3-sync`

- 에이전트 위임 **권장** (복잡한 작업에 전문성 필요 시)
- 직접 도구 사용 **허용** (단순 작업에)
- 사용자 상호작용은 Alfred의 `AskUserQuestion`만 사용

##### Type B: Utility Commands

**명령어**: `/jikime:jarvis`, `/jikime:fix`, `/jikime:loop`, `/jikime:test`

- [HARD] 모든 구현/수정 작업에 **에이전트 위임 필수**
- 직접 도구 접근은 진단(LSP, 테스트, 린터)에만 허용
- **모든** 코드 수정은 전문 에이전트에 위임
- auto compact 또는 세션 복구 후에도 적용

**이유**: 세션 컨텍스트 손실 시 품질 저하 방지

#### 선택 결정 트리

```
1. 읽기 전용 코드베이스 탐색?
   → Explore subagent 사용

2. 외부 문서/API 조사 필요?
   → WebSearch, WebFetch, Context7 MCP 도구 사용

3. 도메인 전문성 필요?
   → expert-[domain] subagent 사용

4. 워크플로우 조율 필요?
   → manager-[workflow] subagent 사용

5. 복잡한 다단계 작업?
   → manager-strategy subagent 사용
```

#### 컨텍스트 최적화

에이전트에 위임 시:

- **최소 컨텍스트** 전달 (spec_id, 핵심 요구사항 3개 이하, 200자 이하 아키텍처 요약)
- 배경 정보, 추론, 비필수 세부사항 **제외**
- 각 에이전트는 독립적인 200K 토큰 세션 보유

---

### quality.md - 품질 게이트

**목적**: 모든 작업에 대한 품질 검증 규칙과 체크리스트

#### HARD Rules 체크리스트

작업 완료 전 필수 확인:

- [ ] 전문성 필요 시 모든 구현 작업이 에이전트에 위임됨
- [ ] 사용자 응답이 `conversation_language`로 작성됨
- [ ] 독립 작업이 병렬 실행됨
- [ ] XML 태그가 사용자에게 표시되지 않음
- [ ] URL이 포함 전 검증됨 (WebSearch)
- [ ] WebSearch 사용 시 출처 명시됨

#### SOFT Rules 체크리스트

권장 모범 사례:

- [ ] 작업에 적절한 에이전트 선택
- [ ] 최소 컨텍스트가 에이전트에 전달됨
- [ ] 결과가 일관되게 통합됨
- [ ] 복잡한 작업에 에이전트 위임 (Type B 명령)

#### 위반 감지

| 위반 | 설명 |
|------|------|
| **에이전트 고려 없음** | 복잡한 구현 요청에 에이전트 위임 미고려 |
| **검증 생략** | 중요 변경에 품질 검증 생략 |
| **언어 불일치** | 사용자의 `conversation_language` 무시 |

#### DDD 품질 표준

Domain-Driven Development 사용 시:

- [ ] 리팩토링 전 기존 테스트 실행
- [ ] 커버리지 없는 코드에 특성화 테스트 생성
- [ ] ANALYZE-PRESERVE-IMPROVE 사이클로 동작 보존
- [ ] 변경이 점진적이고 검증됨

#### TRUST 5 Framework

| 원칙 | 설명 |
|------|------|
| **T**ested | 모든 코드에 적절한 테스트 커버리지 |
| **R**eadable | 코드가 자기 문서화되고 명확함 |
| **U**nified | 코드베이스 전체에 일관된 패턴 |
| **S**ecured | 보안 모범 사례 적용 |
| **T**rackable | 변경이 문서화되고 추적 가능 |

---

### interaction.md - 사용자 상호작용 규칙

**목적**: 사용자 상호작용과 AskUserQuestion 사용 규칙

#### 중요 제약

> Task()로 호출된 서브에이전트는 격리된 상태 없는 컨텍스트에서 작동하며 사용자와 직접 상호작용할 수 없습니다.

**Alfred만 AskUserQuestion 사용 가능** - 서브에이전트는 불가

#### 올바른 워크플로우 패턴

```
Step 1: Alfred가 AskUserQuestion으로 사용자 선호도 수집
        ↓
Step 2: Alfred가 사용자 선택을 프롬프트에 포함하여 Task() 호출
        ↓
Step 3: 서브에이전트가 제공된 매개변수로 실행 (사용자 상호작용 없음)
        ↓
Step 4: 서브에이전트가 결과와 함께 구조화된 응답 반환
        ↓
Step 5: Alfred가 에이전트 응답에 따라 다음 결정을 위해 AskUserQuestion 사용
```

#### AskUserQuestion 제약

| 제약 | 규칙 |
|------|------|
| 질문당 옵션 | 최대 4개 |
| 이모지 사용 | 질문 텍스트, 헤더, 옵션 라벨에 이모지 금지 |
| 언어 | 질문은 사용자의 `conversation_language`로 |

#### 명확화 규칙

- 사용자 의도가 불명확하면 진행 **전** AskUserQuestion으로 명확화
- 에이전트에 위임 **전** 필요한 모든 사용자 선호도 수집
- 확인 없이 사용자 선호도 가정 금지

---

## Development Rules

### coding-style.md - 코딩 스타일 규칙

**목적**: 일관되고 유지보수 가능한 코드를 위한 품질 및 스타일 가이드라인

#### 불변성 (CRITICAL)

항상 새 객체 생성, 절대 변경 금지:

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

#### 파일 구성

**많은 작은 파일 > 적은 큰 파일**

| 가이드라인 | 목표 |
|-----------|------|
| 파일당 줄 수 | 200-400 일반적, 800 최대 |
| 함수당 줄 수 | < 50줄 |
| 중첩 깊이 | < 4단계 |
| 응집도 | 높음 (단일 책임) |
| 결합도 | 낮음 (최소 의존성) |

**구성 원칙**: 타입이 아닌 기능/도메인으로 구성

```
# ❌ WRONG: 타입별
src/
├── components/
├── hooks/
├── services/
└── utils/

# ✅ CORRECT: 기능별
src/
├── auth/
│   ├── components/
│   ├── hooks/
│   └── services/
├── users/
└── products/
```

#### 에러 처리

항상 포괄적으로 에러 처리:

```typescript
try {
  const result = await riskyOperation()
  return result
} catch (error) {
  console.error('Operation failed:', error)
  throw new Error('Detailed user-friendly message')
}
```

#### 네이밍 컨벤션

| 타입 | 컨벤션 | 예시 |
|------|--------|------|
| 변수 | camelCase | `userName`, `isActive` |
| 함수 | camelCase, 동사 접두사 | `getUserById`, `validateInput` |
| 클래스 | PascalCase | `UserService`, `AuthController` |
| 상수 | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT`, `API_BASE_URL` |
| 파일 | kebab-case 또는 camelCase | `user-service.ts`, `userService.ts` |

#### 금지 패턴

| 패턴 | 이유 |
|------|------|
| `any` 타입 (TypeScript) | 타입 안전성 무효화 |
| 매직 넘버 | 명명된 상수 사용 |
| 깊은 중첩 | 함수로 추출 |
| God 객체/함수 | 책임별로 분리 |
| 주석 처리된 코드 | 삭제 (git 히스토리 사용) |

---

### git-workflow.md - Git 워크플로우 규칙

**목적**: 일관된 버전 관리를 위한 Git 컨벤션과 워크플로우 가이드라인

#### 커밋 메시지 형식

```
<type>: <description>

<optional body>
```

#### 커밋 타입

| 타입 | 설명 |
|------|------|
| `feat` | 새 기능 |
| `fix` | 버그 수정 |
| `refactor` | 코드 리팩토링 (동작 변경 없음) |
| `docs` | 문서 변경 |
| `test` | 테스트 추가/수정 |
| `chore` | 유지보수 작업 |
| `perf` | 성능 개선 |
| `ci` | CI/CD 변경 |

#### 브랜치 네이밍

```
<type>/<description>
```

| 타입 | 용도 |
|------|------|
| `feature/` | 새 기능 |
| `fix/` | 버그 수정 |
| `refactor/` | 코드 리팩토링 |
| `docs/` | 문서 |
| `chore/` | 유지보수 |

#### 금지 사항

| 사항 | 이유 |
|------|------|
| main/master에 Force push | 히스토리 파괴 |
| 시크릿 커밋 | 보안 위험 |
| 큰 단일 커밋 | 리뷰/되돌리기 어려움 |
| 기능 브랜치에 머지 커밋 | 히스토리 복잡 |
| 빌드 아티팩트 커밋 | 저장소 비대화 |

---

### testing.md - 테스트 가이드라인

**목적**: DDD 방법론을 적용한 테스트 모범 사례

#### 커버리지 목표

| 유형 | 목표 | 우선순위 |
|------|------|----------|
| 비즈니스 로직 | 90%+ | Critical |
| API 엔드포인트 | 80%+ | High |
| UI 컴포넌트 | 70%+ | Medium |
| 유틸리티 | 80%+ | Medium |
| **전체** | **80%+** | Required |

#### DDD 테스트 접근

##### ANALYZE → PRESERVE → IMPROVE

```
1. ANALYZE
   - 기존 테스트 실행
   - 테스트 커버리지 갭 식별
   - 현재 동작 이해

2. PRESERVE
   - 커버리지 없는 코드에 특성화 테스트 작성
   - 현재 동작을 기준선으로 캡처
   - 회귀 없음 보장

3. IMPROVE
   - 변경 구현
   - 각 변경 후 모든 테스트 실행
   - 새 기능에 테스트 추가
```

#### 좋은 테스트 원칙

| 원칙 | 설명 |
|------|------|
| **Fast** | 빠르게 실행, 자주 실행 유도 |
| **Isolated** | 테스트 간 의존성 없음 |
| **Repeatable** | 매번 같은 결과 |
| **Self-validating** | 명확한 통과/실패, 수동 확인 불필요 |
| **Timely** | 코드 변경과 가까이 작성 |

---

### security.md - 보안 가이드라인

**목적**: OWASP Top 10 및 산업 표준 기반 보안 모범 사례

#### OWASP Top 10 체크리스트

##### 1. Injection

```typescript
// ❌ CRITICAL: SQL Injection
const query = `SELECT * FROM users WHERE id = ${userId}`

// ✅ SAFE: Parameterized query
const { data } = await supabase.from('users').select('*').eq('id', userId)
```

##### 2. Broken Authentication

```typescript
// ❌ CRITICAL: 평문 비밀번호 비교
if (password === storedPassword) { /* login */ }

// ✅ SAFE: 해시 비교
const isValid = await bcrypt.compare(password, hashedPassword)
```

##### 3. Sensitive Data Exposure

```typescript
// ❌ CRITICAL: 하드코딩된 시크릿
const apiKey = "sk-proj-xxxxx"

// ✅ SAFE: 환경 변수
const apiKey = process.env.OPENAI_API_KEY
```

##### 4. XSS

```typescript
// ❌ HIGH: XSS 취약점
element.innerHTML = userInput

// ✅ SAFE: textContent 사용
element.textContent = userInput
```

#### 시크릿 관리

##### 환경 변수

```typescript
// ❌ NEVER: 하드코딩된 시크릿
const apiKey = 'sk-proj-xxxxx'

// ✅ ALWAYS: 환경 변수
const apiKey = process.env.API_KEY

if (!apiKey) {
  throw new Error('API_KEY environment variable not set')
}
```

#### 보안 체크리스트

모든 커밋 전:

- [ ] 하드코딩된 시크릿 없음
- [ ] 모든 사용자 입력 검증됨
- [ ] SQL Injection 방지
- [ ] XSS 방지
- [ ] CSRF 보호 (해당 시)
- [ ] 인증 검증
- [ ] 권한 확인
- [ ] 민감 데이터 로깅 안 함
- [ ] 에러 메시지가 정보 누출 안 함
- [ ] 의존성 최신화

---

### patterns.md - 공통 패턴

**목적**: 일관된 구현을 위한 재사용 가능한 코드 패턴

#### API 패턴

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

#### 패턴 선택 가이드

| 시나리오 | 패턴 |
|----------|------|
| 데이터 접근 | Repository |
| 비즈니스 로직 | Service |
| 객체 생성 | Factory |
| 상태 관리 | Custom Hook |
| 복잡한 컴포넌트 | Compound |
| 입력 검증 | Zod Schema |

---

## System Rules

### hooks.md - Hook 시스템

**목적**: 자동화된 워크플로우와 품질 강화를 위한 Claude Code hooks

#### Hook 유형

| 유형 | 시점 | 목적 |
|------|------|------|
| **PreToolUse** | 도구 실행 전 | 검증, 수정, 차단 |
| **PostToolUse** | 도구 실행 후 | 자동 포맷, 체크, 로깅 |
| **Notification** | 특정 이벤트 발생 시 | 알림, 상태 업데이트 |
| **Stop** | 세션 종료 시 | 최종 검증 |

#### 권장 PostToolUse Hooks

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

#### 권한 관리

| 수준 | 액션 |
|------|------|
| **Safe** | Read, Glob, Grep, LSP - 자동 수락 가능 |
| **Review** | Edit, Write, Bash - 수락 전 검토 |
| **Block** | rm -rf, sudo, force push - 항상 차단 |

---

### performance.md - 성능 최적화

**목적**: 효율적인 Claude Code 사용과 코드 성능 가이드라인

#### 모델 선택 전략

##### Haiku (빠름, 비용 효율)

**사용 대상**: 단순 코드 생성, 포맷팅, 간단한 Q&A, 다중 에이전트 워크플로우의 워커

**특성**: Sonnet 능력의 90%, 3배 비용 절감, 가장 빠른 응답

##### Sonnet (균형, 권장)

**사용 대상**: 주요 개발 작업, 복잡한 코딩, 다중 에이전트 워크플로우 조율

**특성**: 능력과 비용의 최적 균형, 강력한 코딩 성능

##### Opus (최대 능력)

**사용 대상**: 복잡한 아키텍처 결정, 깊은 추론, 연구 분석

**특성**: 최대 추론 깊이, 복잡한 문제에 최적, 더 높은 비용과 지연

#### 컨텍스트 윈도우 관리

| 영역 | 컨텍스트 | 권장 |
|------|----------|------|
| Critical | 80-100% | 대규모 리팩토링, 복잡한 디버깅 피하기 |
| Safe | 0-60% | 단일 파일 편집, 독립 유틸리티 생성 |

#### 알고리즘 복잡도 목표

```
- 일반 작업: O(n) 이하
- 정렬: O(n log n) 허용
- 핫 패스에서 O(n²) 피하기
- 명시적 승인 없이 O(2^n) 금지
```

---

### skills.md - 스킬 발견 및 관리

**목적**: 스킬 발견, 로딩, 활용 규칙

#### 스킬 발견 명령

```bash
# 모든 가용 스킬 목록
jikime-adk skill list

# 태그, 단계, 에이전트, 언어로 필터링
jikime-adk skill list --tag framework
jikime-adk skill list --language typescript

# 키워드로 스킬 검색
jikime-adk skill search <keyword>

# 관련 스킬 찾기
jikime-adk skill related <skill-name>

# 상세 스킬 정보
jikime-adk skill info <skill-name> --body
```

#### 스킬 로딩 규칙

##### 자동 로딩 (트리거)

```yaml
triggers:
  keywords: ["react", "component"]     # 사용자 입력에 포함 시
  phases: ["run"]                      # 현재 개발 단계
  agents: ["expert-frontend"]          # 사용 중인 에이전트
  languages: ["typescript"]            # 프로젝트 언어
```

##### Progressive Disclosure

| 레벨 | 내용 | 토큰 | 로드 시점 |
|------|------|------|----------|
| **Level 1** | 메타데이터만 | ~100 | 에이전트 초기화 |
| **Level 2** | 전체 본문 | ~5K | 트리거 조건 일치 |
| **Level 3+** | 번들 파일 | 가변 | Claude가 필요 시 |

#### 스킬 카테고리

| 카테고리 | 접두사 | 예시 |
|----------|--------|------|
| Language | `jikime-lang-*` | jikime-lang-typescript, jikime-lang-python |
| Platform | `jikime-platform-*` | jikime-platform-vercel, jikime-platform-supabase |
| Domain | `jikime-domain-*` | jikime-domain-frontend, jikime-domain-backend |
| Workflow | `jikime-workflow-*` | jikime-workflow-spec, jikime-workflow-ddd |
| Foundation | `jikime-foundation-*` | jikime-foundation-claude, jikime-foundation-core |

---

### web-search.md - 웹 검색 프로토콜

**목적**: Anti-hallucination 정책 및 URL 검증 규칙

#### HARD 규칙

| 규칙 | 설명 |
|------|------|
| [HARD] URL Verification | 모든 URL은 포함 전 WebFetch로 검증 |
| [HARD] Uncertainty Disclosure | 미검증 정보는 불확실함으로 표시 |
| [HARD] Source Attribution | 모든 웹 검색 결과에 실제 출처 포함 |

#### 실행 단계

```
1. Initial Search
   → 구체적, 타겟팅된 쿼리로 WebSearch 사용

2. URL Validation
   → 포함 전 각 URL을 WebFetch로 검증

3. Response Construction
   → 실제 검색 출처와 함께 검증된 URL만 포함
```

#### 금지 사항

| 사항 | 이유 |
|------|------|
| 검색에서 찾지 않은 URL 생성 | 허위 정보 생성 |
| 불확실한 정보를 사실로 제시 | 사용자 오도 |
| "Sources:" 섹션 생략 | 정보 출처 숨김 |

#### 응답 형식

WebSearch 사용 시 항상 포함:

```markdown
## Answer

[검증된 정보가 포함된 응답]

## Sources

- [Source Title 1](https://verified-url-1.com)
- [Source Title 2](https://verified-url-2.com)
```

---

## 규칙 우선순위

### HARD vs SOFT Rules

| 유형 | 적용 | 예시 |
|------|------|------|
| **HARD** | 필수, 예외 없음 | 언어 규칙, 병렬 실행, XML 금지 |
| **SOFT** | 권장, 상황에 따라 유연 | 에이전트 위임, 품질 체크 |

### 위반 시 조치

| 위반 수준 | 조치 |
|-----------|------|
| HARD Rule 위반 | 즉시 수정 필요 |
| SOFT Rule 위반 | 권장 사항 고려 |
| 보안 위반 | 작업 중단, 즉시 수정 |

---

## 규칙 참조 표

| 규칙 파일 | 주요 내용 | 적용 대상 |
|-----------|----------|----------|
| core.md | HARD 규칙 | 모든 작업 |
| agents.md | 에이전트 위임 | 명령 실행 |
| quality.md | 품질 게이트 | 코드 변경 |
| interaction.md | 사용자 상호작용 | Alfred 응답 |
| coding-style.md | 코딩 스타일 | 코드 작성 |
| git-workflow.md | Git 컨벤션 | 버전 관리 |
| testing.md | 테스트 가이드 | 테스트 작성 |
| security.md | 보안 가이드 | 모든 코드 |
| patterns.md | 공통 패턴 | 구현 |
| hooks.md | Hook 시스템 | 자동화 |
| performance.md | 성능 최적화 | 효율성 |
| skills.md | 스킬 관리 | 스킬 로딩 |
| web-search.md | 웹 검색 | 정보 검색 |

---

Version: 1.0.0
Last Updated: 2026-01-22
