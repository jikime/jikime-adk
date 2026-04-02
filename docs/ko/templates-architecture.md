## 1. Commands (34개)

### Commands란?

Commands는 Markdown 파일로 정의되는 **슬래시 커맨드**이다. 사용자가 `/커맨드이름`으로 호출하면, Claude Code가 해당 파일의 내용을 프롬프트로 주입하여 실행한다.

- **핵심 역할**: **"무엇을 할 것인가"** — 사용자의 의도를 구조화된 워크플로우로 변환하고, 에이전트에 작업을 위임한다

# 커맨드 본문 (Phase별 실행 지시문)

### 1-1. Type A: 워크플로우 커맨드 (핵심 4단계)

| 커맨드 | 기능 | 위임 에이전트 체인 |
|--------|------|-------------------|
| `/jikime:0-project` | 프로젝트 초기화, 코드베이스 분석, 문서 생성 | manager-project → Explore → manager-docs |
| `/jikime:1-plan` | SPEC 문서 작성, EARS 요구사항 정의, 브랜치/워크트리 생성 | Explore → manager-spec → manager-git |
| `/jikime:2-run` | DDD(ANALYZE-PRESERVE-IMPROVE) 구현, 품질 검증 | manager-strategy → manager-ddd → manager-quality → manager-git |
| `/jikime:3-sync` | 문서 동기화, SPEC 완료 처리, PR 머지 | manager-quality → manager-docs → manager-git |

0-project → 1-plan → 2-run → 3-sync
(초기화)    (설계)    (구현)    (동기화)

### 1-2. Type B: J.A.R.V.I.S. 유틸리티 커맨드 (19개)

| 커맨드 | 기능 |
|--------|------|
| `jarvis` | 자율 개발 오케스트레이션 (모든 작업 자동 분해·위임) |
| `build-fix` | 빌드/타입 에러 점진적 수정 |
| `cleanup` | 데드코드 탐지 및 안전 제거, DELETION_LOG 추적 |
| `codemap` | AST 분석 기반 코드베이스 아키텍처 맵 생성 |
| `verify` | 통합 품질 검증 (빌드, 타입, 린트, 테스트, 보안, 브라우저) |
| `test` | 테스트 실행 및 커버리지 확인 |
| `loop` | 자율 반복 개선 루프 (Ralph Loop) |
| `eval` | Eval 기반 개발 (pass@k 메트릭) |
| `e2e` | Playwright E2E 테스트 생성·실행 |
| `architect` | 아키텍처 리뷰, 시스템 설계, ADR 작성 |
| `docs` | README, API 문서, 코드 주석 동기화 |
| `learn` | 코드베이스 탐색 및 학습 |
| `refactor` | DDD 방법론 기반 코드 리팩토링 |
| `security` | OWASP Top 10, 의존성 스캔, 시크릿 탐지 |
| `poc` | POC 우선 개발 (Make It Work → Refactor → Test → Quality → PR) |
| `pr-lifecycle` | PR 생성 → CI 모니터링 → 리뷰 해결 → 머지 자동화 |
| `github` | GitHub Issue 병렬 수정 + PR 리뷰 (워크트리 격리) |
| `harness` | `WORKFLOW.md` 생성 (jikime serve 자동화) |
| `perspective` | 다관점 병렬 분석 (아키텍처, 보안, 성능, 테스트) |

### 1-3. Type B: F.R.I.D.A.Y. 마이그레이션 커맨드 (7개)

| 커맨드 | 기능 |
|--------|------|
| `friday` | 마이그레이션 오케스트레이션 (레거시 → 모던 프레임워크) |
| `migrate-0-discover` | [0/4] 소스 프로젝트 기술 스택·아키텍처·복잡도 탐색 |
| `migrate-1-analyze` | [1/4] 컴포넌트, 라우팅, 상태관리, 의존성 상세 분석 |
| `migrate-2-plan` | [2/4] 마이그레이션 계획 수립 (단계 정의, 공수 추정, 리스크) |
| `migrate-3-execute` | [3/4] DDD 방법론으로 마이그레이션 실행 |
| `migrate-4-verify` | [4/4] 마이그레이션 검증 |
| `smart-rebuild` | 스크린샷 기반 AI 레거시 사이트 리빌딩 |

### 1-4. Type C: 생성기 커맨드 (2개)

| 커맨드 | 기능 |
|--------|------|
| `skill-create` | Claude Code 스킬 생성 (Progressive Disclosure 구조) |
| `migration-skill` | 마이그레이션 전용 스킬 생성 |

---

## 2. Agents (68개)

### Agents란?

Agents는 Markdown 파일로 정의되며, Claude Code의 **Agent 도구**로 호출되는 독립 실행 전문가이다. 각 에이전트는 메인 대화와 별도의 200K 토큰 컨텍스트에서 실행된다.

- **핵심 역할**: **"누가 할 것인가"** — 특정 도메인 전문 지식을 가진 독립 실행 단위

# 에이전트 역할 및 실행 지침

**격리 원칙**: 에이전트는 사용자와 직접 대화할 수 없다. 사용자 상호작용은 반드시 커맨드/오케스트레이터에서 처리한다.

### 2-1. Manager 에이전트 (12개) — 워크플로우 조율

| 에이전트 | 역할 | 사용 스킬 |
|----------|------|----------|
| `manager-spec` | SPEC 문서 생성, EARS 포맷, 요구사항 분석 | jikime-foundation-claude, jikime-workflow-spec |
| `manager-ddd` | DDD ANALYZE-PRESERVE-IMPROVE 사이클 실행 | jikime-workflow-ddd, jikime-tool-ast-grep |
| `manager-strategy` | 시스템 설계, 아키텍처 결정, 트레이드오프 분석 | jikime-workflow-spec, jikime-workflow-project |
| `manager-quality` | TRUST 5 품질 검증, 코드 리뷰 | jikime-workflow-testing, jikime-tool-ast-grep |
| `manager-docs` | 문서 생성, 마크다운 최적화 | jikime-foundation-core |
| `manager-git` | Git 커밋, 브랜치, PR 관리 | jikime-workflow-project |
| `manager-project` | 프로젝트 설정, 구조 관리, 초기화 | jikime-workflow-project |
| `manager-claude-code` | Claude Code 설정, 스킬·에이전트·커맨드 관리 | jikime-foundation-claude |
| `manager-database` | DB 스키마 설계, 쿼리 최적화 | — |
| `manager-dependency` | 패키지 업데이트, 취약점 수정 | — |
| `manager-data` | 데이터 파이프라인, ETL, 데이터 모델링 | — |
| `manager-context` | 컨텍스트 윈도우 최적화, 세션 상태 관리 | — |

### 2-2. Specialist 에이전트 (20개) — 도메인 전문가

| 에이전트 | 역할 | 사용 스킬 |
|----------|------|----------|
| `backend` | API 개발, 서버 로직, DB 통합 | jikime-domain-backend, jikime-domain-database |
| `frontend` | React 컴포넌트, UI 구현 | jikime-domain-frontend, jikime-library-shadcn |
| `fullstack` | 엔드투엔드 기능 개발 (DB → API → UI) | — |
| `architect` | 시스템 설계, 컴포넌트 설계 | — |
| `debugger` | 디버깅, 에러 분석, 근본 원인 추적 | jikime-lang-typescript, jikime-lang-python |
| `security-auditor` | 보안 분석, OWASP 준수 | — |
| `devops` | CI/CD, 인프라, 배포 자동화 | jikime-platform-vercel |
| `optimizer` | 성능 최적화, 프로파일링 | jikime-lang-typescript |
| `e2e-tester` | E2E 테스트 실행, 브라우저 테스트 | — |
| `test-guide` | 테스트 전략, 테스트 작성 가이드 | — |
| `refactorer` | 코드 리팩토링, 아키텍처 개선 | — |
| `build-fixer` | 빌드 에러 해결, 컴파일 수정 | — |
| `reviewer` | 코드 리뷰, PR 리뷰 | — |
| `documenter` | API 문서, 코드 문서 생성 | — |
| `planner` | 작업 계획, 분해, 추정 | — |
| `migrator` | 레거시 현대화, 프레임워크 마이그레이션 | — |
| `analyst` | 기술 리서치, 경쟁 분석 | — |
| `explorer` | 코드베이스 검색, 구현 탐색 | — |
| `designer-ui` | UI 디자인 시스템, 접근성 | jikime-domain-uiux, jikime-design-tools |
| `coordinator` | 멀티 에이전트 조율, 결과 집계 | — |

### 2-3. Language/Framework Specialist (18개)

| 에이전트 | 전문 영역 |
|----------|----------|
| `specialist-typescript` | TypeScript 5.0+, 고급 타입 패턴 |
| `specialist-javascript` | ES2023+, Node.js 20+ |
| `specialist-python` | Python 3.11+, FastAPI, Django |
| `specialist-java` | Java 21+, Spring Boot |
| `specialist-go` | Go, Fiber/Gin, GORM |
| `specialist-php` | PHP 8.3+, Laravel, Symfony |
| `specialist-rust` | Rust 2021, 메모리 안전성 |
| `specialist-nextjs` | Next.js App Router, RSC |
| `specialist-angular` | Angular 15+, NgRx, RxJS |
| `specialist-vue` | Vue 3, Composition API, Nuxt 3 |
| `specialist-spring` | Spring 생태계 |
| `specialist-postgres` | PostgreSQL, pgvector, RLS |
| `specialist-sql` | 멀티DB SQL 최적화 |
| `specialist-graphql` | GraphQL 스키마, Apollo Federation |
| `specialist-api` | REST/GraphQL API 설계 |
| `specialist-microservices` | 분산 시스템, Kubernetes |
| `specialist-mobile` | React Native, Flutter |
| `specialist-electron` | Electron 데스크탑 앱 |
| `specialist-websocket` | WebSocket, Socket.IO |

### 2-4. Harness 에이전트 (3개) — 자율 작업 실행

| 에이전트 | 역할 | 사용 스킬 |
|----------|------|----------|
| `harness-worker` | Plans.md 태스크를 워크트리에서 구현 (cc:WIP → cc:DONE) | jikime-harness-plan, jikime-lang-* |
| `harness-scaffolder` | Plans.md 태스크 분석, 구현 구조 스캐폴딩 | jikime-harness-plan |
| `harness-reviewer` | 4관점 코드 리뷰 (pm:REVIEW → pm:OK) | jikime-foundation-quality |

### 2-5. Builder 에이전트 (4개) — 확장 생성

| 에이전트 | 역할 | 사용 스킬 |
|----------|------|----------|
| `agent-builder` | 새 에이전트 정의 생성 | jikime-foundation-claude |
| `command-builder` | 새 슬래시 커맨드 생성 | jikime-foundation-claude |
| `skill-builder` | 새 스킬 정의 생성 | jikime-foundation-claude |
| `plugin-builder` | 새 플러그인 패키지 생성 | jikime-foundation-claude |

### 2-6. Orchestration 에이전트 (3개)

| 에이전트 | 역할 |
|----------|------|
| `orchestrator` | 워크플로우 파이프라인 조율, 프로세스 자동화 |
| `coordinator` | 멀티 에이전트 조율, 작업 분배, 결과 집계 |
| `dispatcher` | 작업 큐 관리, 로드 밸런싱, 우선순위 스케줄링 |

### 2-7. Team 에이전트 (8개) — 실험적 병렬 실행

| 에이전트 | Phase | 권한 | 역할 |
|----------|-------|------|------|
| `team-researcher` | plan | read-only | 코드베이스 탐색·리서치 |
| `team-analyst` | plan | read-only | 요구사항 분석, 엣지케이스 식별 |
| `team-architect` | plan | read-only | 기술 설계, 대안 평가 |
| `team-backend-dev` | run | acceptEdits | 서버 구현 (src/api/**, src/services/**) |
| `team-frontend-dev` | run | acceptEdits | UI 구현 (src/components/**, src/pages/**) |
| `team-designer` | run | acceptEdits | UI/UX 디자인, 디자인 토큰 |
| `team-tester` | run | acceptEdits | 테스트 작성 (tests/**, **/*.test.*) |
| `team-quality` | run | read-only | TRUST 5 품질 검증 |

---

## 3. Skills (87개, 14개 도메인)

### Skills란?

Skills는 폴더별(`skill-name/SKILL.md`)로 정의되며, Claude Code가 **사용자 요청이나 에이전트 실행 시 자동 트리거**하는 지식 모듈이다.

- **핵심 역할**: **"어떻게 할 것인가"** — 기술·도메인·워크플로우에 대한 지식과 패턴을 담은 가이드북

# 스킬 본문 (실행 가이드, 패턴, 규칙)

**Progressive Disclosure**: 3단계 로딩으로 컨텍스트를 절약한다.

| 단계 | 로딩 시점 | 크기 |
|------|----------|------|
| Level 1: Metadata | 항상 (name + description) | ~100토큰 |
| Level 2: 본문 | 스킬 트리거 시 | <500줄 |
| Level 3: References | 필요 시 | 무제한 |

**에이전트와의 관계**: 에이전트의 `skills` 필드에 스킬을 지정하면, 에이전트 실행 시 자동 로딩된다. 1:N, N:1 관계 모두 가능.

### 도메인별 스킬 목록

| 도메인 | 개수 | 주요 스킬 | 역할 |
|--------|------|----------|------|
| **foundation** | 6 | claude, core, quality, context, philosopher, thinking | 오케스트레이션 패턴, TRUST 5, SPEC 시스템, 토큰 최적화, 사고 프레임워크 |
| **domain** | 6 | architecture, backend, database, frontend, uiux, design-tools | 도메인별 전문 지식 (API 설계, DB 패턴, UI 컴포넌트) |
| **workflow** | 23 | ddd, tdd, spec, verify, loop, poc, symphony, team, worktree 등 | 개발 워크플로우 (DDD 사이클, TDD, 품질검증, 반복루프, 문서동기화) |
| **lang** | 7 | typescript, javascript, python, java, go, php, flutter | 언어별 최신 패턴·관용구·프레임워크 가이드 |
| **library** | 6 | shadcn, zod, mermaid, streamdown, vercel-ai-sdk, tds-react-native | 라이브러리 사용 패턴·통합 가이드 |
| **marketing** | 10 | seo, copywriting, pricing, analytics, ab-test, launch 등 | 마케팅 전략·분석·최적화 |
| **migration** | 6 | to-nextjs, jquery-to-react, angular-to-nextjs, smart-rebuild 등 | 프레임워크 마이그레이션 전략·실행 |
| **platform** | 4 | supabase, vercel, clerk, vercel-react | 플랫폼별 통합·최적화 |
| **tool** | 3 | ast-grep, agent-browser, mcp-builder | 도구 활용 가이드 (AST 변환, 브라우저 자동화, MCP 서버 개발) |
| **harness** | 6 | plan, work, review, setup, sync, release | Harness Engineering 자동화 (Plans.md 관리, 작업 실행, 리뷰, 릴리즈) |
| **team** | 5 | team, leader, worker, reviewer, swarm | 멀티 에이전트 팀 조율·역할 정의 |
| **framework** | 3 | nextjs@14, nextjs@15, nextjs@16 | Next.js 버전별 마이그레이션·패턴 |
| **mobile** | 1 | react-native | React Native + Expo 모바일 개발 |
| **revfactory** | 1 | revfactory-harness | 하네스 설계 메타 스킬 (에이전트 팀·스킬 아키텍처 생성) |

---

## 4. 연관관계 다이어그램

### 4-1. 핵심 워크플로우 체인

```
/jikime:0-project ─── manager-project ─── jikime-workflow-project
                  ├── Explore
                  └── manager-docs

/jikime:1-plan ────── manager-spec ─────── jikime-workflow-spec
                  ├── Explore                jikime-foundation-core
                  └── manager-git ────────── jikime-workflow-project

/jikime:2-run ─────── manager-strategy ──── jikime-workflow-spec
                  ├── manager-ddd ─────────── jikime-workflow-ddd
                  │                           jikime-tool-ast-grep
                  ├── manager-quality ─────── jikime-workflow-testing
                  └── manager-git              jikime-foundation-quality

/jikime:3-sync ────── manager-quality
                  ├── manager-docs
                  └── manager-git
```

### 4-2. Harness Engineering 체인

```
/jikime:harness ──────→ WORKFLOW.md 생성
                           ↓
jikime serve ──────────→ GitHub Issue 감지
                           ↓
jikime-harness-setup ──→ 환경 준비
jikime-harness-plan ───→ Plans.md 생성/관리
                           ↓
jikime-harness-work ───→ harness-worker (워크트리에서 구현)
                       → harness-scaffolder (구조 스캐폴딩)
                           ↓
jikime-harness-review ─→ harness-reviewer (4관점 코드 리뷰)
                           ↓
jikime-harness-release → CHANGELOG, 버전 태그, 릴리즈 노트
jikime-harness-sync ───→ Plans.md ↔ git 동기화, 회고
```

### 4-3. 팀 모드 체인

```
/jikime:team ──────→ TeamCreate
                       ↓
Plan Phase:    team-researcher + team-analyst + team-architect
                       ↓ (findings → design)
Run Phase:     team-backend-dev + team-frontend-dev + team-designer
               + team-tester (병렬 실행, 파일 소유권으로 충돌 방지)
                       ↓
Quality Gate:  team-quality (TRUST 5 검증)
```

### 4-4. Foundation 스킬 의존 관계

```
jikime-foundation-claude ←── 거의 모든 Manager/Builder 에이전트가 참조
jikime-foundation-core   ←── SPEC 시스템, 위임 패턴, Progressive Disclosure
jikime-foundation-quality ←── harness-reviewer, team-quality
jikime-foundation-philosopher ←── team-researcher, team-analyst, team-architect
jikime-foundation-thinking ←── (독립적 사고 프레임워크)
jikime-foundation-context  ←── (토큰 최적화, 세션 관리)
```

---

## 5. 수치 요약

| 구성 요소 | 개수 |
|-----------|------|
| 슬래시 커맨드 | 34개 (워크플로우 4 + 유틸리티 26 + 생성기 2 + 팀 2) |
| 에이전트 | 68개 (매니저 12 + 전문가 20 + 언어 18 + 하네스 3 + 빌더 4 + 오케스트레이션 3 + 팀 8) |
| 스킬 | 87개 (14개 도메인) |

---