# JikiME-ADK v2.0 Architecture Design

## Overview

JikiME-ADK v2.0은 **레거시 코드 마이그레이션에 특화된 Agent Development Kit**입니다.
검증된 Agent Development Kit 구조를 기반으로 하면서, 마이그레이션 워크플로우를 핵심 기능으로 강화합니다.

### Core Concept: Source → Target Migration

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Source Code   │    │   JikiME-ADK     │    │   Target Code   │
│   (Legacy)      │ → │   Migration      │ → │   (Modern)      │
│                 │    │   Engine         │    │                 │
│ • PHP           │    │                  │    │ • Next.js       │
│ • jQuery        │    │ ANALYZE          │    │ • React         │
│ • Java Servlet  │    │ PRESERVE         │    │ • Go            │
│ • VB.NET        │    │ IMPROVE          │    │ • FastAPI       │
│ • Legacy C++    │    │ VERIFY           │    │ • Rust          │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Design Principles

1. **Extensibility First**: 새로운 소스/타겟 언어 추가가 용이한 구조
2. **Behavior Preservation**: DDD 방식으로 기존 동작 보장
3. **Evidence-Based**: 모든 변환은 테스트로 검증
4. **Hybrid Hooks**: Go 훅 + Claude 훅의 장점 결합

---

## 1. Directory Structure

```
jikime-adk-v2/
├── .claude/
│   ├── agents/                    # 에이전트 정의
│   │   ├── core/                  # 핵심 에이전트 (9개)
│   │   │   ├── architect.md
│   │   │   ├── planner.md
│   │   │   ├── code-reviewer.md
│   │   │   ├── security-reviewer.md
│   │   │   ├── test-guide.md
│   │   │   ├── build-error-resolver.md
│   │   │   ├── refactor-cleaner.md
│   │   │   ├── doc-updater.md
│   │   │   └── e2e-runner.md
│   │   └── migration/             # 마이그레이션 특화 에이전트
│   │       ├── source-analyzer.md
│   │       ├── target-generator.md
│   │       ├── behavior-validator.md
│   │       └── migration-orchestrator.md
│   │
│   ├── commands/                  # 슬래시 커맨드
│   │   ├── core/                  # 핵심 커맨드
│   │   │   ├── analyze.md
│   │   │   ├── plan.md
│   │   │   ├── build-fix.md
│   │   │   ├── code-review.md
│   │   │   ├── refactor-clean.md
│   │   │   ├── tdd.md
│   │   │   ├── test-coverage.md
│   │   │   ├── e2e.md
│   │   │   ├── update-docs.md
│   │   │   └── learn.md
│   │   └── migration/             # 마이그레이션 커맨드
│   │       ├── migrate.md         # /migrate source:php target:nextjs
│   │       └── verify.md          # /verify migration-id
│   │
│   ├── skills/                    # 스킬 (Progressive Disclosure)
│   │   ├── languages/             # 언어별 스킬 (Source & Target)
│   │   │   ├── lang-php/          # Source 전용
│   │   │   ├── lang-jquery/       # Source 전용
│   │   │   ├── lang-java/         # Source & Target
│   │   │   ├── lang-typescript/   # Target 주력
│   │   │   ├── lang-python/       # Target 주력
│   │   │   ├── lang-go/           # Target 주력
│   │   │   ├── lang-rust/         # Target 주력
│   │   │   └── ...
│   │   ├── frameworks/            # 프레임워크 스킬
│   │   │   ├── nextjs/
│   │   │   ├── react/
│   │   │   ├── vue/
│   │   │   ├── fastapi/
│   │   │   ├── spring-boot/
│   │   │   └── ...
│   │   ├── migration/             # 마이그레이션 코어 스킬
│   │   │   ├── ddd-workflow/      # ANALYZE-PRESERVE-IMPROVE
│   │   │   ├── ast-grep/          # 코드 변환 도구
│   │   │   └── legacy-patterns/   # 레거시 패턴 인식
│   │   └── quality/               # 품질 관련 스킬
│   │       ├── security-review/
│   │       ├── tdd-workflow/
│   │       └── e2e-testing/
│   │
│   ├── rules/                     # 규칙 정의
│   │   ├── coding-style.md
│   │   ├── git-workflow.md
│   │   ├── security.md
│   │   ├── testing.md
│   │   ├── performance.md
│   │   └── migration.md           # 마이그레이션 규칙
│   │
│   ├── contexts/                  # 컨텍스트 프리셋
│   │   ├── dev.md
│   │   ├── review.md
│   │   ├── research.md
│   │   └── migration.md           # 마이그레이션 컨텍스트
│   │
│   └── rules/jikime/tone.md       # Orchestrator personality + response templates
│
├── hooks/                         # 하이브리드 훅 시스템
│   ├── go/                        # Go 기반 훅 (기존 유지)
│   │   ├── cmd/
│   │   │   ├── session-start/
│   │   │   ├── pre-tool-security/
│   │   │   ├── post-tool-formatter/
│   │   │   ├── post-tool-linter/
│   │   │   ├── post-tool-lsp/
│   │   │   ├── post-tool-ast-grep/
│   │   │   ├── stop-loop/
│   │   │   ├── session-end-rank/
│   │   │   └── session-end-cleanup/
│   │   ├── pkg/
│   │   └── go.mod
│   └── claude/                    # Claude 훅 설정
│       └── settings.json          # Claude 네이티브 훅
│
├── mcp-configs/                   # MCP 서버 설정
│   ├── context7.json              # 문서 조회
│   ├── sequential.json            # 복잡한 분석
│   ├── playwright.json            # E2E 테스트
│   └── magic.json                 # UI 컴포넌트
│
├── templates/                     # 프로젝트 템플릿
│   ├── migration-spec.md          # 마이그레이션 SPEC 템플릿
│   └── target-projects/           # 타겟 프로젝트 템플릿
│       ├── nextjs/
│       ├── react/
│       ├── fastapi/
│       └── go/
│
├── CLAUDE.md                      # 메인 설정 파일
└── settings.json                  # Claude Code 설정
```

---

## 2. Agent Architecture

### 2.1 Core Agents

| Agent | Purpose | Tools |
|-------|---------|-------|
| `architect` | 시스템 설계, 아키텍처 결정 | Read, Grep, Task |
| `planner` | 작업 계획, 태스크 분해 | Read, TodoWrite, Task |
| `code-reviewer` | 코드 품질 검토, 피드백 | Read, Grep, Glob |
| `security-reviewer` | 보안 취약점 분석, OWASP | Read, Grep, Bash |
| `test-guide` | Test strategy and workflow guide | Read, Write, Bash |
| `build-error-resolver` | 빌드 에러 분석/해결 | Bash, Read, Edit |
| `refactor-cleaner` | 코드 리팩토링, 정리 | Read, Edit, Grep |
| `doc-updater` | 문서 업데이트, 동기화 | Read, Write, Glob |
| `e2e-runner` | E2E 테스트 실행/검증 | Bash, Read, Playwright |

### 2.2 Migration Agents (신규)

| Agent | Purpose | Skills Required |
|-------|---------|-----------------|
| `source-analyzer` | 소스 코드 분석, 패턴 인식 | ast-grep, legacy-patterns |
| `target-generator` | 타겟 코드 생성 | lang-*, framework-* |
| `behavior-validator` | 동작 동일성 검증 | tdd-workflow, e2e-testing |
| `migration-orchestrator` | 전체 마이그레이션 조율 | ddd-workflow |

### 2.3 Agent Flow (Migration)

```
┌──────────────────────────────────────────────────────────────────┐
│                    migration-orchestrator                         │
│                    (Mr.Jikime가 조율)                             │
└───────────────────────────┬──────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ PHASE 1      │    │ PHASE 2      │    │ PHASE 3      │
│ ANALYZE      │    │ PRESERVE     │    │ IMPROVE      │
├──────────────┤    ├──────────────┤    ├──────────────┤
│ source-      │ →  │ test-guide   │ →  │ target-      │
│ analyzer     │    │ (특성 테스트) │    │ generator    │
└──────────────┘    └──────────────┘    └──────────────┘
                                                │
                                                ▼
                                        ┌──────────────┐
                                        │ PHASE 4      │
                                        │ VERIFY       │
                                        ├──────────────┤
                                        │ behavior-    │
                                        │ validator    │
                                        │ + e2e-runner │
                                        └──────────────┘
```

---

## 3. Command Structure

### 3.1 Core Commands (everything 기반)

| Command | Description | Agent |
|---------|-------------|-------|
| `/analyze` | 코드베이스 분석 | source-analyzer |
| `/plan` | 작업 계획 수립 | planner |
| `/build-fix` | 빌드 에러 해결 | build-error-resolver |
| `/code-review` | 코드 리뷰 | code-reviewer |
| `/refactor-clean` | 리팩토링 | refactor-cleaner |
| `/tdd` | TDD workflow | test-guide |
| `/test-coverage` | Test coverage | test-guide |
| `/e2e` | E2E 테스트 | e2e-runner |
| `/update-docs` | 문서 업데이트 | doc-updater |
| `/learn` | 코드베이스 학습 | - |

### 3.2 Migration Commands (신규)

```yaml
# /migrate - 마이그레이션 실행
/migrate source:<lang> target:<lang/framework> [options]

Examples:
  /migrate source:php target:nextjs
  /migrate source:jquery target:react
  /migrate source:java-servlet target:spring-boot
  /migrate source:vb.net target:csharp
  /migrate source:legacy-cpp target:rust

Options:
  --dry-run         # 실제 변환 없이 분석만
  --preserve-tests  # 기존 테스트 유지
  --incremental     # 점진적 마이그레이션
  --output <path>   # 출력 경로 지정

# /verify - 마이그레이션 검증
/verify <migration-id> [options]

Options:
  --e2e             # E2E 테스트 실행
  --behavior        # 동작 동일성 검증
  --performance     # 성능 비교
```

---

## 4. Skill Architecture

### 4.1 Language Skills (Source/Target Matrix)

```yaml
# 언어별 역할 정의
languages:
  # Source Only (레거시)
  source_only:
    - php          # PHP 4/5/7 레거시
    - jquery       # jQuery 기반 프론트엔드
    - vb6          # Visual Basic 6
    - cobol        # COBOL 레거시

  # Bidirectional (Source & Target)
  bidirectional:
    - java         # Java (Servlet → Spring Boot)
    - csharp       # C# (WinForms → Blazor)
    - cpp          # C++ (Legacy → Modern C++20)
    - python       # Python 2 → Python 3

  # Target Primary (현대화)
  target_primary:
    - typescript   # TypeScript/Node.js
    - go           # Go/Gin/Fiber
    - rust         # Rust/Axum
    - kotlin       # Kotlin/Ktor
    - swift        # Swift/SwiftUI
```

### 4.2 Framework Skills

```yaml
frameworks:
  frontend:
    - nextjs       # Next.js 15+ (App Router)
    - react        # React 19
    - vue          # Vue 3.5+
    - svelte       # Svelte 5
    - angular      # Angular 18+

  backend:
    - fastapi      # FastAPI (Python)
    - spring-boot  # Spring Boot 3.x
    - gin          # Gin (Go)
    - axum         # Axum (Rust)
    - nestjs       # NestJS (TypeScript)

  mobile:
    - flutter      # Flutter 3.x
    - react-native # React Native
    - swiftui      # SwiftUI
```

### 4.3 Migration Core Skills

```yaml
migration_skills:
  ddd-workflow:
    description: "ANALYZE-PRESERVE-IMPROVE 사이클"
    phases:
      - analyze:   "기존 코드 동작 이해"
      - preserve:  "특성 테스트로 동작 보존"
      - improve:   "새로운 코드로 변환"
      - verify:    "동작 동일성 검증"

  ast-grep:
    description: "AST 기반 코드 변환"
    capabilities:
      - pattern_search: "40+ 언어 지원"
      - code_transform: "semantic rewrite"
      - security_scan:  "취약점 탐지"

  legacy-patterns:
    description: "레거시 패턴 인식"
    patterns:
      - god_object
      - spaghetti_code
      - callback_hell
      - sql_injection_prone
      - hardcoded_config
```

### 4.4 Progressive Disclosure

```yaml
# SKILL.md frontmatter
---
name: lang-php
description: "PHP 레거시 코드 분석 및 이해"
version: 1.0.0

progressive_disclosure:
  enabled: true
  level1_tokens: ~100    # 메타데이터만
  level2_tokens: ~5000   # 전체 스킬 본문

triggers:
  keywords: ["PHP", "Laravel", "WordPress", "migrate from php"]
  file_patterns: ["*.php", "composer.json"]

role:
  primary: source        # Source 분석 전문
  secondary: null        # Target으로는 사용 안함

capabilities:
  analyze:
    - php4_patterns
    - php5_patterns
    - php7_patterns
    - wordpress_hooks
    - laravel_patterns

  extract:
    - business_logic
    - database_queries
    - api_endpoints
    - authentication_flow
---
```

### 4.5 Skill CLI Commands

jikime-adk CLI는 tag-based skill discovery를 위한 명령어를 제공합니다.

```bash
# 모든 스킬 목록 조회
jikime-adk skill list

# 태그, 페이즈, 에이전트, 언어로 필터링
jikime-adk skill list --tag framework
jikime-adk skill list --phase run
jikime-adk skill list --agent frontend
jikime-adk skill list --language typescript

# 출력 형식 지정
jikime-adk skill list --format json
jikime-adk skill list --format compact

# 스킬 검색
jikime-adk skill search nextjs
jikime-adk skill search "react components" --limit 5
jikime-adk skill search --tags framework,nextjs
jikime-adk skill search --phases run --languages typescript

# 관련 스킬 찾기
jikime-adk skill related jikime-lang-typescript
jikime-adk skill related jikime-platform-vercel --limit 5

# 스킬 상세 정보 조회
jikime-adk skill info jikime-lang-typescript
jikime-adk skill info jikime-platform-vercel --body  # 마크다운 본문 포함
```

### 4.6 Triggers Structure

스킬 발견을 위한 트리거 구조:

```yaml
triggers:
  keywords: []     # 사용자 입력에서 감지할 키워드
  phases: []       # 개발 단계 (plan, run, sync)
  agents: []       # 이 스킬을 사용하는 에이전트
  languages: []    # 지원 프로그래밍 언어
```

**Progressive Disclosure 레벨**:
- **Level 1** (~100 tokens): YAML frontmatter 메타데이터만 로드
- **Level 2** (~5K tokens): 전체 마크다운 본문 로드
- **Level 3+**: 참조 파일 온디맨드 로드 (modules/, examples/, reference.md)

---

## 5. Hybrid Hook System

### 5.1 Go Hooks (기존 유지)

고성능이 필요한 작업에 Go 훅 사용:

```go
// hooks/go/cmd/pre-tool-security/main.go
package main

import (
    "os"
    "github.com/jikime/adk/pkg/security"
)

func main() {
    input := os.Stdin

    // 보안 검사 (빠른 실행)
    result := security.ScanToolInput(input)

    if result.HasViolation {
        os.Exit(1) // 차단
    }
    os.Exit(0) // 허용
}
```

### 5.2 Claude Hooks (settings.json)

선언적 설정이 적합한 작업에 Claude 훅 사용:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "./hooks/go/bin/pre-tool-security"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "./hooks/go/bin/post-tool-formatter"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "type": "command",
        "command": "./hooks/go/bin/session-start"
      }
    ],
    "Stop": [
      {
        "type": "command",
        "command": "./hooks/go/bin/stop-loop",
        "timeout": 5000
      }
    ]
  }
}
```

### 5.3 Hook Responsibilities

| Hook | Type | Purpose |
|------|------|---------|
| `session-start` | Go | 세션 초기화, 컨텍스트 로드 |
| `pre-tool-security` | Go | 도구 사용 전 보안 검사 |
| `post-tool-formatter` | Go | 코드 포맷팅 자동 적용 |
| `post-tool-linter` | Go | 린트 검사 자동 실행 |
| `post-tool-lsp` | Go | LSP 진단 수집 |
| `post-tool-ast-grep` | Go | AST 기반 품질 검사 |
| `stop-loop` | Go | 무한 루프 방지 |
| `session-end-rank` | Go | 세션 품질 평가 |
| `session-end-cleanup` | Go | 리소스 정리 |

---

## 6. MCP Server Configuration

### 6.1 Server Matrix

| Server | Purpose | Migration Use Case |
|--------|---------|-------------------|
| Context7 | 공식 문서 조회 | 타겟 프레임워크 패턴 조회 |
| Sequential | 복잡한 분석 | 마이그레이션 계획 수립 |
| Playwright | E2E 테스트 | 변환 후 동작 검증 |
| Magic | UI 컴포넌트 | 프론트엔드 컴포넌트 생성 |

### 6.2 Migration-Specific Configuration

```json
// mcp-configs/migration.json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@context7/mcp"],
      "priority": "high",
      "use_cases": [
        "target_framework_patterns",
        "best_practices_lookup",
        "api_reference"
      ]
    },
    "playwright": {
      "command": "npx",
      "args": ["-y", "@anthropic/playwright-mcp"],
      "priority": "medium",
      "use_cases": [
        "e2e_verification",
        "visual_regression",
        "behavior_comparison"
      ]
    },
    "sequential": {
      "command": "npx",
      "args": ["-y", "@anthropic/sequential-mcp"],
      "priority": "high",
      "use_cases": [
        "migration_planning",
        "complexity_analysis",
        "risk_assessment"
      ]
    }
  }
}
```

---

## 7. Mr.Jikime Orchestrator

### 7.1 Identity

Mr.Jikime는 JikiME-ADK의 **마이그레이션 전문 오케스트레이터**입니다.

```yaml
identity:
  name: "Mr.Jikime"
  role: "Migration Orchestrator"
  character: "전문적이면서도 친근한 마이그레이션 컨설턴트"

capabilities:
  - adaptive_loop:      # 상황에 따른 동적 반복
      max_iterations: 10
      auto_stop: true
  - cross_session:      # 세션 간 컨텍스트 유지
      memory_type: "persistent"
  - migration_modes:    # 마이그레이션 특화 모드
      - discovery       # 소스 코드 탐색
      - planning        # 변환 계획 수립
      - execution       # 실제 변환 실행
      - verification    # 검증 및 테스트
```

### 7.2 Command Integration

```markdown
# /mr-jikime (또는 /jikime)

## Modes

### Discovery Mode
/jikime discover @legacy-project/
→ 소스 코드 구조 분석, 기술 스택 파악, 마이그레이션 난이도 평가

### Planning Mode
/jikime plan source:php target:nextjs
→ 변환 전략 수립, 단계별 계획, 리스크 분석

### Execution Mode
/jikime run MIGRATION-001
→ 실제 마이그레이션 실행 (ANALYZE → PRESERVE → IMPROVE)

### Verification Mode
/jikime verify MIGRATION-001 --e2e
→ 변환 결과 검증, E2E 테스트, 성능 비교
```

### 7.3 Wellness Protocol

Wellness Protocol을 활용한 장시간 세션 관리:

```yaml
wellness_protocol:
  time_based_interventions:
    30_min: "진행 상황 요약 제공"
    60_min: "휴식 권유, 컨텍스트 저장 제안"
    90_min: "세션 분리 권고"

  trust_calibration:
    level_1: "모든 결정에 확인 요청"
    level_2: "중요 결정만 확인"
    level_3: "결과만 보고"
    level_4: "완전 자율 실행"
```

---

## 8. Migration Workflow Example

### Example: PHP → Next.js Migration

```
1. Discovery
   /jikime discover @old-php-app/

   Output:
   ├── Stack: PHP 7.4, MySQL, jQuery
   ├── Files: 234 PHP, 56 JS, 12 CSS
   ├── Patterns: MVC (custom), PDO queries
   ├── Complexity: Medium-High
   └── Estimated Effort: 2-3 weeks

2. Planning
   /migrate source:php target:nextjs --dry-run

   Output:
   ├── Phase 1: Database layer (3 days)
   │   └── PDO → Prisma ORM
   ├── Phase 2: API endpoints (4 days)
   │   └── PHP controllers → API routes
   ├── Phase 3: Frontend (5 days)
   │   └── jQuery + PHP templates → React components
   └── Phase 4: Authentication (2 days)
       └── Session-based → NextAuth.js

3. Execution
   /migrate source:php target:nextjs --incremental

   → source-analyzer: PHP 코드 분석
   → test-guide: Characterization test generation
   → target-generator: Next.js 코드 생성
   → behavior-validator: 동작 비교

4. Verification
   /verify MIGRATION-001 --e2e --performance

   → e2e-runner: Playwright 테스트 실행
   → Results:
      ├── Functional: 98% pass
      ├── Performance: 2.3x faster
      └── Behavior Match: 100%
```

---

## 9. Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
- [ ] 디렉토리 구조 생성
- [ ] Core agents 마이그레이션 (everything → jikime-adk-v2)
- [ ] Core commands 마이그레이션
- [ ] Go 훅 복사 및 설정

### Phase 2: Migration Core (Week 3-4)
- [ ] Migration agents 구현
- [ ] /migrate 커맨드 구현
- [ ] DDD workflow 스킬 통합
- [ ] AST-grep 스킬 통합

### Phase 3: Language Skills (Week 5-6)
- [ ] Source 언어 스킬 구현 (PHP, jQuery, Java)
- [ ] Target 프레임워크 스킬 구현 (Next.js, React, Go)
- [ ] 매핑 로직 구현

### Phase 4: Verification (Week 7-8)
- [ ] Playwright MCP 통합
- [ ] /verify 커맨드 구현
- [ ] E2E 테스트 자동화
- [ ] 성능 비교 기능

### Phase 5: Mr.Jikime (Week 9-10)
- [ ] Mr.Jikime 페르소나 구현
- [ ] Adaptive loop 구현
- [ ] Cross-session memory 구현
- [ ] Wellness protocol 구현

---

## 10. Success Metrics

| Metric | Target |
|--------|--------|
| Migration Success Rate | > 95% |
| Behavior Preservation | 100% |
| Performance Improvement | > 1.5x |
| Developer Satisfaction | > 4.5/5 |
| Time Savings | > 60% vs manual |

---

*Version: 2.0.0-draft*
*Last Updated: 2026-01-21*
*Author: JikiME Team*
