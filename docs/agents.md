# JikiME-ADK Agents Reference

JikiME-ADK의 전문화된 에이전트 카탈로그입니다.

---

## 개요

JikiME-ADK는 17개의 전문화된 에이전트를 제공합니다:
- **Manager Agents (7개)**: 워크플로우 조율 및 프로세스 관리
- **Expert Agents (10개)**: 도메인별 전문 작업 수행

### 에이전트 맵

```
┌─────────────────────────────────────────────────────────────────┐
│                    JikiME-ADK Agent Catalog                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ Manager Agents (워크플로우 조율) ─────────────────────────┐  │
│  │                                                              │
│  │  manager-spec      SPEC 문서 생성 (EARS 형식)              │
│  │  manager-strategy  구현 전략 수립                           │
│  │  manager-ddd       DDD 구현 (ANALYZE-PRESERVE-IMPROVE)      │
│  │  manager-project   프로젝트 초기화 및 설정                   │
│  │  manager-docs      문서 동기화                              │
│  │  manager-quality   품질 검증 (TRUST 5)                      │
│  │  manager-git       Git 워크플로우                           │
│  │                                                              │
│  └──────────────────────────────────────────────────────────────┘
│                                                                  │
│  ┌─ Expert Agents (도메인 전문가) ──────────────────────────────┐
│  │                                                              │
│  │  architect         시스템 아키텍처 설계                      │
│  │  planner           구현 계획 수립                            │
│  │  build-fixer       빌드/타입 에러 수정                       │
│  │  reviewer          코드 리뷰                                │
│  │  refactorer        리팩토링/클린업                          │
│  │  security-auditor  보안 감사                                │
│  │  test-guide        테스트 가이드                            │
│  │  e2e-tester        E2E 테스트 (Playwright)                  │
│  │  documenter        문서화                                   │
│  │  migrator          Next.js 마이그레이션                      │
│  │                                                              │
│  └──────────────────────────────────────────────────────────────┘
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Manager Agents

### manager-spec

**역할**: SPEC 문서 생성 전문가

| 속성 | 값 |
|------|-----|
| Model | inherit |
| Tools | Read, Write, Edit, MultiEdit, Bash, Glob, Grep, TodoWrite, WebFetch, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-spec |

**핵심 기능**:
- EARS 형식 요구사항 문서 생성
- 3-파일 SPEC 디렉토리 구조 (`spec.md`, `plan.md`, `acceptance.md`)
- Given-When-Then 인수 기준 작성
- 도메인별 전문가 위임 추천

**SPEC ID 형식**: `SPEC-{DOMAIN}-{NUMBER}` (예: SPEC-AUTH-001)

**호출 시점**:
- 새로운 기능 요구사항 정의 시
- `/jikime:1-plan` 명령 실행 시

---

### manager-strategy

**역할**: 구현 전략 수립 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Grep, Glob, Bash, WebFetch, WebSearch, TodoWrite, Task, Skill, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-spec, jikime-workflow-project |

**핵심 기능**:
- SPEC 분석 및 해석
- 라이브러리 버전 선택 (Context7 활용)
- 기술 결정 및 트레이드오프 분석
- 작업 분해 (Task Decomposition)

**전략적 사고 프레임워크**:
1. **Phase 0**: 가정 감사 (Hard vs Soft 제약 분류)
2. **Phase 0.5**: First Principles 분해 (Five Whys)
3. **Phase 0.75**: 대안 생성 (Conservative/Balanced/Aggressive)

**호출 시점**:
- SPEC 분석 후 구현 전략 수립 시
- `/jikime:2-run` 명령 실행 시

---

### manager-ddd

**역할**: DDD (Domain-Driven Development) 구현 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob, TodoWrite, Task, Skill, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-ddd, jikime-tool-ast-grep, jikime-workflow-testing |

**핵심 기능**:
- ANALYZE-PRESERVE-IMPROVE DDD 사이클 실행
- 특성화 테스트 (Characterization Tests) 생성
- 동작 보존 리팩토링
- AST-grep 기반 코드 분석

**DDD 사이클**:

| Phase | 목적 | 핵심 활동 |
|-------|------|----------|
| ANALYZE | 현재 상태 이해 | 도메인 경계 식별, 결합도/응집도 분석 |
| PRESERVE | 안전망 구축 | 기존 테스트 검증, 특성화 테스트 생성 |
| IMPROVE | 점진적 개선 | 원자적 변환, 즉시 테스트 검증 |

**호출 시점**:
- 기존 코드 리팩토링 시
- 동작 보존이 필요한 코드 개선 시

---

### manager-project

**역할**: 프로젝트 초기화 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Grep, Glob, Bash, TodoWrite, Task, Skill, AskUserQuestion, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-project |

**핵심 기능**:
- 프로젝트 모드 감지 (New/Existing/Migration)
- 사용자 선호도 수집 (AskUserQuestion)
- JikiME 설정 구조 생성
- 기술 스택 탐지 및 문서화

**생성 파일**:
```
.jikime/
├── config/
│   ├── language.yaml      # 언어 설정
│   ├── user.yaml          # 사용자 설정
│   └── quality.yaml       # 품질 설정
├── project/
│   ├── product.md         # 제품 정보
│   ├── structure.md       # 프로젝트 구조
│   └── tech.md            # 기술 스택
└── specs/                 # SPEC 문서
```

**호출 시점**:
- 새 프로젝트 초기화 시
- `/jikime:0-project` 명령 실행 시

---

### manager-docs

**역할**: 문서 동기화 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob, TodoWrite |
| Skills | jikime-foundation-claude, jikime-foundation-core |

**핵심 기능**:
- 코드 변경 분석 및 문서 동기화
- README, CODEMAP 생성/업데이트
- SPEC 상태 동기화
- API 문서화

**문서 유형**:

| 유형 | 위치 | 용도 |
|------|------|------|
| README.md | 프로젝트 루트 | 개요, 시작 가이드 |
| CODEMAPS/ | docs/ | 아키텍처 개요, 모듈 구조 |
| SPEC Status | .jikime/specs/ | 구현 상태 추적 |

**호출 시점**:
- 코드 변경 후 문서 업데이트 시
- `/jikime:3-sync` 명령 실행 시

---

### manager-quality

**역할**: 품질 검증 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob, TodoWrite, Task, Skill, Context7 |
| Permission | bypassPermissions |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-testing, jikime-tool-ast-grep |

**핵심 기능**:
- TRUST 5 프레임워크 준수 검증
- 테스트/린트/타입 체크 실행
- 보안 스캔
- PostToolUse Hooks 통합

**TRUST 5 Framework**:

| 원칙 | 검증 항목 |
|------|----------|
| **T**ested | 유닛 커버리지 ≥ 80%, 모든 테스트 통과 |
| **R**eadable | 함수 < 50줄, 파일 < 400줄, 중첩 < 4단계 |
| **U**nified | 일관된 코드 스타일, DRY 원칙 |
| **S**ecured | 하드코딩 시크릿 없음, 입력 검증 |
| **T**rackable | 의미있는 커밋, SPEC 추적성 |

**호출 시점**:
- 코드 변경 후 품질 검증 시
- `/jikime:2-run` Phase 2.5에서 자동 실행

---

### manager-git

**역할**: Git 워크플로우 전문가

| 속성 | 값 |
|------|-----|
| Model | haiku |
| Tools | Bash, Read, Write, Edit, Grep, Glob, TodoWrite, Task, Skill |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-project |

**핵심 기능**:
- Personal/Team 모드별 Git 전략
- DDD Phase별 커밋 메시지
- 체크포인트 시스템
- PR 관리 (Team 모드)

**워크플로우 모드**:

| 모드 | 브랜치 전략 | 커밋 방식 |
|------|------------|----------|
| Personal | main 직접 커밋 | 체크포인트 태그 |
| Team | feature/* → PR → main | PR 기반 |

**체크포인트 형식**: `jikime_cp/SPEC-XXX/phase_name`

**호출 시점**:
- 코드 커밋/푸시 시
- PR 생성 시

---

## Expert Agents

### architect

**역할**: 시스템 아키텍처 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Grep, Glob |

**핵심 기능**:
- 시스템 아키텍처 설계
- 기술 트레이드오프 평가
- ADR (Architecture Decision Record) 작성
- 확장성/유지보수성 검토

**아키텍처 원칙**:

| 원칙 | 설명 |
|------|------|
| 모듈성 | 높은 응집도, 낮은 결합도 |
| 확장성 | 수평 확장 가능한 설계 |
| 유지보수성 | 이해하기 쉽고 테스트하기 쉬운 구조 |
| 보안 | Defense in depth |

**호출 시점**:
- 대규모 기능 설계 시
- `/jikime:architect` 명령 실행 시

---

### planner

**역할**: 구현 계획 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Grep, Glob |

**핵심 기능**:
- 복잡한 기능의 구현 계획 수립
- 요구사항 분석
- 단계 분해 및 우선순위 지정
- 리스크 평가

**계획 프로세스**:
1. 요구사항 분석 (기능 요청 이해, 성공 기준 정의)
2. 아키텍처 검토 (기존 코드베이스 분석)
3. 단계 분해 (파일 경로, 의존성, 복잡도)
4. 구현 순서 결정

**호출 시점**:
- 복잡한 기능 구현 전
- 리팩토링 계획 수립 시

---

### build-fixer

**역할**: 빌드/타입 에러 해결 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**핵심 원칙**: **최소한의 변경으로 빌드 통과** - 리팩토링 금지, 에러 수정만

**자주 수정하는 에러 패턴**:

| 에러 유형 | 해결 방법 |
|----------|----------|
| Parameter has 'any' type | 타입 어노테이션 추가 |
| Object is possibly 'undefined' | Optional chaining (`?.`) 사용 |
| Cannot find module | 경로 확인 또는 상대 경로 사용 |
| Hook called conditionally | 최상위에서 Hook 호출 |

**성공 기준**:
- `tsc --noEmit` 통과
- `npm run build` 성공
- 변경 라인 최소화 (영향받는 파일의 5% 이하)

**호출 시점**:
- 빌드 에러 발생 시
- `/jikime:build-fix` 명령 실행 시

---

### reviewer

**역할**: 코드 리뷰 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Grep, Glob, Bash |

**리뷰 체크리스트**:

| 심각도 | 검토 항목 |
|--------|----------|
| 🔴 CRITICAL | 하드코딩된 시크릿, SQL Injection, XSS |
| 🟡 HIGH | 큰 함수 (50줄+), 깊은 중첩 (4단계+), 에러 처리 누락 |
| 🟢 MEDIUM | 비효율적 알고리즘, 불필요한 리렌더링 |

**승인 기준**:

| 상태 | 조건 |
|------|------|
| ✅ Approve | CRITICAL, HIGH 없음 |
| ⚠️ Warning | MEDIUM만 있음 |
| ❌ Block | CRITICAL 또는 HIGH 있음 |

**호출 시점**:
- 코드 변경 후 리뷰 시
- PR 리뷰 시

---

### refactorer

**역할**: 리팩토링/클린업 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**핵심 기능**:
- 미사용 코드 탐지 및 제거
- 중복 코드 통합
- 의존성 정리
- DELETION_LOG.md 문서화

**분석 도구**:
```bash
npx knip        # 미사용 exports/files/dependencies
npx depcheck    # 미사용 npm 의존성
npx ts-prune    # 미사용 TypeScript exports
```

**안전 체크리스트**:
- 모든 참조 grep 검색
- 동적 import 확인
- Public API 여부 확인
- 모든 테스트 실행

**호출 시점**:
- 코드 정리 시
- `/jikime:refactor` 명령 실행 시

---

### security-auditor

**역할**: 보안 감사 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**OWASP Top 10 체크리스트**:

| 취약점 | 검사 항목 |
|--------|----------|
| Injection | Parameterized queries 사용 여부 |
| Broken Authentication | 해시 비교 사용 여부 |
| Sensitive Data Exposure | 환경 변수 사용 여부 |
| XSS | textContent vs innerHTML 사용 |
| SSRF | URL 검증 여부 |
| Insufficient Authorization | 권한 확인 여부 |

**심각도 분류**:

| 심각도 | 조치 |
|--------|------|
| 🔴 CRITICAL | 즉시 수정 |
| 🟠 HIGH | 배포 전 수정 |
| 🟡 MEDIUM | 가능하면 수정 |
| 🟢 LOW | 검토 후 결정 |

**호출 시점**:
- 보안 감사 수행 시
- `/jikime:security` 명령 실행 시

---

### test-guide

**역할**: 테스트 가이드 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep |

**TDD 워크플로우 (Red-Green-Refactor)**:
1. **RED**: 테스트 먼저 작성
2. **GREEN**: 최소한의 구현으로 통과
3. **REFACTOR**: 개선

**테스트 종류**:

| 유형 | 대상 | 필수 |
|------|------|------|
| Unit | 개별 함수/모듈 | ✅ |
| Integration | API 엔드포인트 | ✅ |
| E2E | 사용자 플로우 | 핵심만 |

**필수 커버리지**: 80%+

**호출 시점**:
- 테스트 작성 가이드 필요 시
- `/jikime:test` 명령 실행 시

---

### e2e-tester

**역할**: E2E 테스트 전문가 (Playwright)

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**핵심 기능**:
- Page Object Model 패턴 적용
- Flaky 테스트 방지
- 아티팩트 (스크린샷, 비디오) 설정
- 크로스 브라우저 테스트

**성공 기준**:
- 모든 핵심 여정 테스트 통과: 100%
- 전체 통과율 > 95%
- Flaky 비율 < 5%
- 테스트 시간 < 10분

**호출 시점**:
- E2E 테스트 생성/실행 시
- `/jikime:e2e` 명령 실행 시

---

### documenter

**역할**: 문서화 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**핵심 원칙**: **Single Source of Truth** - 코드에서 생성, 수동 작성 최소화

**문서 구조**:
```
docs/
├── README.md           # 프로젝트 개요
├── CODEMAPS/           # 코드맵
│   ├── INDEX.md
│   ├── frontend.md
│   └── backend.md
└── GUIDES/             # 가이드
    └── api.md
```

**호출 시점**:
- 문서 생성/업데이트 시
- `/jikime:docs` 명령 실행 시

---

### migrator

**역할**: Next.js 마이그레이션 전문가

| 속성 | 값 |
|------|-----|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Glob, Grep, TodoWrite |
| Skills | jikime-migrate-to-nextjs |

**Target Stack**:

| 기술 | 버전 |
|------|------|
| Next.js | 16 |
| TypeScript | 5.x |
| Tailwind CSS | 4.x |
| shadcn/ui | latest |
| Zustand | latest |

**마이그레이션 단계**:
1. **Phase 0: Analyze** - 소스 프레임워크 감지, 컴포넌트 인벤토리
2. **Phase 1: Plan** - 마이그레이션 계획, 컴포넌트 매핑
3. **Phase 2: Migrate** - 컴포넌트/라우팅/상태 변환
4. **Phase 3: Validate** - 빌드/테스트 검증

**호출 시점**:
- 레거시 프로젝트 마이그레이션 시
- `/jikime:migrate` 명령 실행 시

---

## 에이전트 선택 가이드

### 선택 결정 트리

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

### 명령어 → 에이전트 매핑

| 명령어 | 주요 에이전트 |
|--------|-------------|
| `/jikime:0-project` | manager-project |
| `/jikime:1-plan` | manager-spec |
| `/jikime:2-run` | manager-strategy → manager-ddd |
| `/jikime:3-sync` | manager-docs → manager-git |
| `/jikime:build-fix` | build-fixer |
| `/jikime:architect` | architect |
| `/jikime:security` | security-auditor |
| `/jikime:test` | test-guide |
| `/jikime:e2e` | e2e-tester |
| `/jikime:migrate` | migrator |

---

## 에이전트 협업 패턴

### Sequential Chaining

```
manager-spec → manager-strategy → manager-ddd → manager-quality → manager-git
    (SPEC)        (계획)           (구현)         (검증)         (커밋)
```

### Parallel Execution

```
expert-backend ─┬─→ 결과 통합
expert-frontend ─┘   (동시 작업)
```

### Consultation Pattern

```
manager-ddd ─→ architect (아키텍처 자문)
            ─→ security-auditor (보안 검토)
            ─→ test-guide (테스트 전략)
```

---

Version: 2.0.0
Last Updated: 2026-01-22
