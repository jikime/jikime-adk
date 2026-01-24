# JiKiME-ADK: 레거시의 가치를 지키고, 현대화의 길을 열다

**AI-Powered Agentic Development Kit for Legacy Modernization**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![License: Copyleft](https://img.shields.io/badge/License-Copyleft--3.0-blue.svg)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/jikime/jikime-adk)](https://github.com/jikime/jikime-adk/releases)

> **"레거시 코드에 담긴 본질과 가치를 끝까지 보존하면서, 이를 현대화된 코드로 안전하게 탈바꿈시킨다."**

---

## JiKiME-ADK란?

16,000개 이상의 홈페이지. 온맘닷컴의 방대한 레거시 코드를 현대화해야 하는 거대한 과제 앞에서 깊은 고민에 빠졌습니다. 이 수많은 코드는 단순히 낡은 과거가 아니라, 오랜 시간 쌓여온 **비즈니스의 가치 그 자체**였기 때문입니다.

이 소중한 자산들을 어떻게 하면 가장 안전하고 효율적으로 미래로 연결할 수 있을까? 그 고민의 결과로 **JiKiME-ADK**가 탄생했습니다.

### 'JiKiME(지키미)' - 레거시를 보존하며 미래를 수호하다

우리말로 읽으면 **'지키미'**가 되는 이 이름에는 "레거시 코드에 담긴 본질과 가치를 끝까지 보존(지키고)하면서, 이를 현대화된 코드로 안전하게 탈바꿈시켜 전체 시스템을 수호하겠다"는 의지를 담았습니다. 단순히 코드를 새로 쓰는 것이 아니라, 과거와 미래를 잇는 든든한 파수꾼 역할을 합니다.

### 듀얼 오케스트레이션: J.A.R.V.I.S. + F.R.I.D.A.Y.

아이언맨의 조력자들에게서 아이디어를 얻어 **듀얼 오케스트레이션** 체계를 도입했습니다:

| Orchestrator | 역할 | 설명 |
|---|---|---|
| **J.A.R.V.I.S.** | 개발 담당 | 새로운 아키텍처 설계와 표준 코드 생성을 담당하는 스마트한 조력자 |
| **F.R.I.D.A.Y.** | 마이그레이션 담당 | 레거시 코드를 분석하고 현대적 구조로 전환하는 마이그레이션 스페셜리스트 |

각 에이전트의 역할을 명확히 구분하여 개발과 마이그레이션 두 영역의 전문성을 극대화합니다.

### 영감과 뿌리

이 여정의 시작에서 명확한 이정표를 제시해준 것은 구스킴님의 **MOAI-ADK**였습니다. 에이전틱(Agentic) 워크플로우에 대한 깊은 통찰을 접하며 큰 영감을 얻었고, 그 단단한 철학적 기반 위에서 Golang을 활용해 마이그레이션에 특화된 새로운 ADK를 구축했습니다. 또한 **everything-claude-code**의 에이전트, 커맨드, 훅 구조를 참고하여 기능을 보강했습니다.

이는 단순한 카피가 아닌, 선배 개발자의 소중한 자산을 양분 삼아 피워낸 새로운 꽃이라고 생각합니다.

---

## 핵심 기능

- **SPEC-First DDD**: 명확한 명세 → ANALYZE-PRESERVE-IMPROVE 사이클로 동작 보존 개발
- **20+ 전문 에이전트**: Manager, Specialist, Builder 에이전트 자동 위임
- **레거시 마이그레이션**: Vue.js, React CRA, Angular 등 → Next.js 16 자동 전환
- **품질 보증**: TRUST 5 프레임워크 (Tested, Readable, Unified, Secured, Trackable)
- **Self-Update**: 바이너리 자동 업데이트 + 임베디드 템플릿 싱크

---

## 설치

### 방법 1: Install Script (권장)

```bash
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash
```

#### 옵션

```bash
# 글로벌 설치 (/usr/local/bin, sudo 필요)
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --global

# 특정 버전 설치
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --version 0.0.2
```

### 방법 2: go install

```bash
go install github.com/jikime/jikime-adk@latest
```

### 방법 3: 수동 다운로드

[GitHub Releases](https://github.com/jikime/jikime-adk/releases)에서 플랫폼에 맞는 바이너리를 다운로드합니다.

| Platform | Architecture | jikime-adk | jikime-wt |
|----------|-------------|------------|-----------|
| macOS | Apple Silicon | `jikime-adk-darwin-arm64` | `jikime-wt-darwin-arm64` |
| macOS | Intel | `jikime-adk-darwin-amd64` | `jikime-wt-darwin-amd64` |
| Linux | x86_64 | `jikime-adk-linux-amd64` | `jikime-wt-linux-amd64` |
| Linux | ARM64 | `jikime-adk-linux-arm64` | `jikime-wt-linux-arm64` |

### 설치 후 디렉토리 구성

```
~/.local/bin/
├── jikime-adk          # 메인 바이너리
├── jikime → jikime-adk # 심볼릭 링크 (단축 명령어)
└── jikime-wt           # Worktree 전용 바이너리
```

`jikime`는 `jikime-adk`의 심볼릭 링크로, 더 짧은 명령어를 제공합니다:

```bash
jikime init        # = jikime-adk init
jikime update      # = jikime-adk update
jikime-wt new auth # = jikime-adk worktree new auth
```

---

## 업데이트

```bash
# 업데이트 확인
jikime-adk update --check

# 업데이트 실행
jikime-adk update

# 템플릿 싱크 (프로젝트 템플릿을 최신 버전으로 동기화)
jikime-adk update --sync-templates
```

---

## 시작하기

### 프로젝트 초기화

```bash
# 프로젝트 디렉토리에서 실행
jikime-adk init
```

`init` 명령은 `.claude/`와 `.jikime/` 디렉토리에 에이전트, 스킬, 커맨드 템플릿을 설치합니다.

### Claude Code에서 사용

초기화 후 Claude Code에서 슬래시 명령어로 사용합니다:

```bash
# 프로젝트 분석 및 문서 생성
/jikime:0-project

# SPEC 정의 (개발 계획)
/jikime:1-plan "User authentication system"

# SPEC 구현 (DDD 사이클)
/jikime:2-run SPEC-AUTH-001

# 문서 동기화 & 완료 처리
/jikime:3-sync SPEC-AUTH-001
```

---

## 명령어 레퍼런스

### Workflow Commands (Type A)

핵심 개발 워크플로우를 구성하는 명령어입니다.

| 명령어 | 설명 |
|--------|------|
| `/jikime:0-project` | 프로젝트 초기화 및 문서 생성 |
| `/jikime:1-plan` | SPEC 정의 및 개발 브랜치 생성 |
| `/jikime:2-run` | DDD 기반 SPEC 구현 |
| `/jikime:3-sync` | 문서 동기화 및 SPEC 완료 처리 |

### Utility Commands (Type B)

빠른 실행과 자동화를 위한 명령어입니다.

| 명령어 | 설명 |
|--------|------|
| `/jikime:jarvis` | J.A.R.V.I.S. 자율 개발 오케스트레이션 |
| `/jikime:friday` | F.R.I.D.A.Y. 자율 마이그레이션 오케스트레이션 |
| `/jikime:test` | 단위/통합 테스트 실행 |
| `/jikime:loop` | LSP/AST-grep 피드백 기반 반복 개선 |

### Standalone Utilities

워크플로우와 독립적으로 사용할 수 있는 유틸리티입니다.

| 명령어 | 설명 |
|--------|------|
| `/jikime:architect` | 아키텍처 리뷰 및 설계, ADR 생성 |
| `/jikime:build-fix` | 빌드 에러 점진적 수정 |
| `/jikime:docs` | 문서 업데이트 및 생성 |
| `/jikime:e2e` | Playwright E2E 테스트 |
| `/jikime:learn` | 코드베이스 탐색 및 학습 |
| `/jikime:refactor` | DDD 방법론 리팩토링 |
| `/jikime:security` | OWASP Top 10 보안 감사 |

### Migration Commands

레거시 프로젝트를 현대화하는 마이그레이션 워크플로우입니다.

| 명령어 | 설명 |
|--------|------|
| `/jikime:friday` | 마이그레이션 통합 명령어 (전체 자동화) |
| `/jikime:migrate-0-discover` | Step 0: 소스 탐색 |
| `/jikime:migrate-1-analyze` | Step 1: 상세 분석 |
| `/jikime:migrate-2-plan` | Step 2: 마이그레이션 계획 수립 |
| `/jikime:migrate-3-execute` | Step 3: 마이그레이션 실행 |
| `/jikime:migrate-4-verify` | Step 4: 검증 |

#### 마이그레이션 대상 스택

| Source | Target |
|--------|--------|
| Vue.js, React CRA, Angular, Svelte 등 | Next.js 16 (App Router) |
| - | TypeScript 5.x + Tailwind CSS 4.x |
| - | shadcn/ui + Zustand |

---

## CLI 명령어

`jikime-adk` 바이너리가 제공하는 CLI 명령어입니다.

### 기본 명령어

| 명령어 | 설명 |
|--------|------|
| `jikime-adk init [path] [name]` | 프로젝트에 템플릿 설치 |
| `jikime-adk status` | 프로젝트 상태 및 설정 확인 |
| `jikime-adk doctor` | 시스템 진단 (의존성, 설정 검증) |
| `jikime-adk update` | 바이너리 자동 업데이트 |
| `jikime-adk statusline` | Claude Code 상태줄 렌더링 |
| `jikime-adk --version` | 버전 확인 |

### init

프로젝트에 에이전트, 스킬, 커맨드 템플릿을 설치합니다.

```bash
jikime-adk init [path] [project-name]
```

| 플래그 | 설명 |
|--------|------|
| `-y, --non-interactive` | 대화형 프롬프트 없이 기본값으로 진행 |
| `--mode <mode>` | 프로젝트 모드 설정 |
| `--locale <locale>` | 로케일 설정 |
| `--language <lang>` | 대화 언어 설정 |
| `--force` | 기존 파일 덮어쓰기 |

### doctor

시스템 환경을 진단하고 문제를 감지합니다.

```bash
jikime-adk doctor
```

| 플래그 | 설명 |
|--------|------|
| `-v, --verbose` | 상세 출력 |
| `--fix` | 발견된 문제 자동 수정 시도 |
| `--export` | 진단 결과 파일로 내보내기 |
| `--check-commands` | 명령어 가용성 검사 |

### update

바이너리 업데이트 및 템플릿 동기화를 수행합니다.

```bash
jikime-adk update
```

| 플래그 | 설명 |
|--------|------|
| `--check` | 새 버전 존재 여부만 확인 |
| `-f, --force` | 강제 업데이트 |
| `--skip-backup` | 백업 생성 건너뛰기 |
| `--sync-templates` | 프로젝트 템플릿을 최신 버전으로 동기화 |

### language

대화 언어를 관리합니다. (en, ko, ja, zh, es, fr, de, pt, it, ru)

| 서브커맨드 | 설명 |
|-----------|------|
| `language list` | 지원 언어 목록 |
| `language info` | 현재 언어 설정 정보 |
| `language set <lang>` | 언어 변경 |
| `language validate` | 언어 설정 유효성 검사 |

### worktree (별칭: wt) / jikime-wt

Git Worktree 기반 병렬 개발 환경을 관리합니다.

`jikime-wt`는 `jikime-adk worktree`의 독립 실행 바이너리로, 더 짧은 명령어를 제공합니다.

```bash
# 아래 두 명령은 동일합니다
jikime-adk worktree new feature/auth
jikime-wt new feature/auth
```

| 서브커맨드 | 설명 |
|-----------|------|
| `worktree new <branch>` | 새 worktree 생성 |
| `worktree list` | worktree 목록 |
| `worktree go <branch>` | worktree로 이동 |
| `worktree remove <branch>` | worktree 제거 |
| `worktree status` | 전체 worktree 상태 |
| `worktree sync` | worktree 간 동기화 |
| `worktree clean` | 불필요한 worktree 정리 |
| `worktree recover` | 깨진 worktree 복구 |
| `worktree done` | 작업 완료 및 정리 |
| `worktree config` | worktree 설정 관리 |

공통 플래그: `--repo <path>`, `--worktree-root <path>`

### tag

TAG System v2.0 - SPEC과 코드 간 추적성을 관리합니다.

| 서브커맨드 | 설명 |
|-----------|------|
| `tag validate` | 태그 형식 유효성 검사 |
| `tag scan` | 코드베이스 태그 스캔 |
| `tag linkage` | SPEC↔CODE 연결 상태 확인 |

### skill

스킬 시스템을 탐색하고 관리합니다.

| 서브커맨드 | 설명 |
|-----------|------|
| `skill list` | 등록된 스킬 목록 |
| `skill search <query>` | 스킬 검색 |
| `skill related <skill>` | 관련 스킬 탐색 |
| `skill info <skill>` | 스킬 상세 정보 |

### hooks

Claude Code 통합 훅을 관리합니다.

```bash
jikime-adk hooks <hook-name>
```

주요 훅: `session-start`, `pre-tool-security`, `pre-bash`, `pre-write`, `post-tool-formatter`, `post-tool-linter`, `post-tool-ast-grep`, `post-tool-lsp`, `stop-loop`, `orchestrator-route` 등

---

## 에이전트 카탈로그

### Manager Agents (8)

| Agent | 역할 |
|-------|------|
| manager-spec | SPEC 문서 생성, EARS 형식, 요구사항 분석 |
| manager-ddd | DDD 개발, ANALYZE-PRESERVE-IMPROVE 사이클 |
| manager-docs | 문서 생성, Nextra 통합 |
| manager-quality | 품질 게이트, TRUST 5 검증 |
| manager-project | 프로젝트 설정, 구조 관리 |
| manager-strategy | 시스템 설계, 아키텍처 결정 |
| manager-git | Git 운영, 브랜치 전략 |
| manager-claude-code | Claude Code 설정, 스킬/에이전트 관리 |

### Specialist Agents (14)

| Agent | 역할 |
|-------|------|
| architect | 시스템 설계, 컴포넌트 설계 |
| backend | API 개발, 서버사이드 로직 |
| frontend | React 컴포넌트, UI 구현 |
| security-auditor | 보안 분석, OWASP 컴플라이언스 |
| devops | CI/CD, 인프라, 배포 자동화 |
| optimizer | 성능 최적화, 프로파일링 |
| debugger | 디버깅, 에러 분석 |
| e2e-tester | E2E 테스트, 브라우저 테스팅 |
| test-guide | 테스트 전략, 커버리지 개선 |
| refactorer | 코드 리팩토링, 아키텍처 개선 |
| build-fixer | 빌드 에러 해결 |
| reviewer | 코드 리뷰, PR 리뷰 |
| documenter | API 문서, 코드 문서 생성 |
| planner | 태스크 계획, 분해, 추정 |

### Builder Agents (4)

| Agent | 역할 |
|-------|------|
| agent-builder | 새 에이전트 정의 생성 |
| command-builder | 새 슬래시 명령어 생성 |
| skill-builder | 새 스킬 정의 생성 |
| plugin-builder | 새 플러그인 패키지 생성 |

---

## 개발 방법론: DDD (Domain-Driven Development)

JiKiME-ADK는 모든 개발에 **ANALYZE-PRESERVE-IMPROVE** 사이클을 적용합니다:

```
┌─────────────┐
│   ANALYZE   │  ← 현재 동작 이해
└──────┬──────┘
       ↓
┌─────────────┐
│  PRESERVE   │  ← 특성화 테스트로 동작 보존
└──────┬──────┘
       ↓
┌─────────────┐
│   IMPROVE   │  ← 자신감 있게 변경
└──────┬──────┘
       ↓
    (반복)
```

- 기존 테스트를 먼저 실행한 후 리팩토링
- 커버리지가 없는 코드는 특성화 테스트 생성
- 동작 보존을 보장하면서 점진적으로 개선

---

## 프로젝트 구조

```
jikime-adk/
├── cmd/                    # CLI 명령어 구현
│   ├── initcmd/           # init - 프로젝트 초기화
│   ├── statuscmd/         # status - 프로젝트 상태
│   ├── doctorcmd/         # doctor - 시스템 진단
│   ├── updatecmd/         # update - 자동 업데이트
│   ├── languagecmd/       # language - 언어 관리
│   ├── worktreecmd/       # worktree - Git Worktree 관리
│   ├── tagcmd/            # tag - TAG System
│   ├── skillcmd/          # skill - 스킬 시스템
│   ├── hookscmd/          # hooks - Claude Code 훅
│   └── statuslinecmd/     # statusline - 상태줄 렌더링
├── templates/             # 임베디드 프로젝트 템플릿
│   ├── .claude/           # Claude Code 설정
│   │   ├── agents/        # 에이전트 정의
│   │   ├── commands/      # 슬래시 명령어
│   │   └── skills/        # 스킬 정의
│   └── .jikime/           # JikiME 설정
│       └── config/        # 프로젝트 설정
├── version/               # 버전 관리
├── install/               # 설치 스크립트
├── .github/               # CI/CD 워크플로우
│   ├── workflows/         # release.yml, deploy-install.yml
│   └── scripts/           # sync-versions.sh
└── docs/                  # 문서
```

---

## 감사의 말

JiKiME-ADK가 지금의 모습을 갖출 수 있었던 것은 아래 프로젝트들 덕분입니다:

- **[MOAI-ADK](https://github.com/modu-ai/moai-adk)** - 구스킴님의 에이전틱 워크플로우 철학과 구조적 영감
- **[everything-claude-code](https://github.com/anthropics/anthropic-cookbook)** - 에이전트, 커맨드, 훅 구조 참고

앞으로 JiKiME는 고유한 로직과 코드로 채워지며 계속 진화하겠지만, 그 뿌리에 닿아있는 영감은 오래도록 남을 것입니다.

---

## 앞으로의 약속

JiKiME-ADK의 발전 과정을 꾸준히 공유하겠습니다. 수많은 시행착오를 거쳐 레거시 마이그레이션의 해답이 될 수 있는 도구임을 증명해내겠습니다. 그리고 그 결실이 맺어지는 날, 더 많은 개발자분께 도움이 될 수 있도록 기꺼이 공개하겠습니다.

**레거시를 지키고 미래를 여는 JiKiME의 행보를 지켜봐 주십시오.**

---

## Links

- [GitHub Repository](https://github.com/jikime/jikime-adk)
- [Releases](https://github.com/jikime/jikime-adk/releases)
- [Install Script](https://jikime.github.io/jikime-adk/install.sh)

---

## License

Copyleft License (COPYLEFT-3.0) - See [LICENSE](./LICENSE) for details.
