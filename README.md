<p align="center">
  <img src="./assets/images/jikime-hero.webp" alt="JiKiME-ADK Hero Image" width="800">
</p>

# JiKiME-ADK: 레거시의 가치를 지키고, 현대화의 길을 열다

**AI-Powered Agentic Development Kit for Legacy Modernization**

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go" alt="Go"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Copyleft--3.0-blue.svg" alt="License: Copyleft"></a>
  <a href="https://github.com/jikime/jikime-adk/releases"><img src="https://img.shields.io/github/v/release/jikime/jikime-adk" alt="Release"></a>
</p>

> **"레거시 코드에 담긴 본질과 가치를 끝까지 보존하면서, 이를 현대화된 코드로 안전하게 탈바꿈시킨다."**

---

## JiKiME-ADK란?

16,000개 이상의 홈페이지. 온맘닷컴의 방대한 레거시 코드를 현대화해야 하는 거대한 과제 앞에서 깊은 고민에 빠졌습니다. 이 수많은 코드는 단순히 낡은 과거가 아니라, 오랜 시간 쌓여온 **비즈니스의 가치 그 자체**였기 때문입니다.

우리말로 읽으면 **'지키미'**가 되는 이 이름에는 "레거시 코드에 담긴 본질과 가치를 끝까지 보존(지키고)하면서, 이를 현대화된 코드로 안전하게 탈바꿈시켜 전체 시스템을 수호하겠다"는 의지를 담았습니다.

### 듀얼 오케스트레이션: J.A.R.V.I.S. + F.R.I.D.A.Y.

| Orchestrator | 역할 | 설명 |
|---|---|---|
| **J.A.R.V.I.S.** | 개발 담당 | 새로운 아키텍처 설계와 표준 코드 생성 |
| **F.R.I.D.A.Y.** | 마이그레이션 담당 | 레거시 코드 분석 및 현대적 구조로 전환 |

> 자세한 내용: [J.A.R.V.I.S. 문서](./docs/jarvis.md) | [F.R.I.D.A.Y. 문서](./docs/friday.md)

---

## 핵심 기능

| 기능 | 설명 | 문서 |
|------|------|------|
| **SPEC-First DDD** | ANALYZE-PRESERVE-IMPROVE 사이클로 동작 보존 개발 | [DDD 문서](./docs/tdd-ddd.md) |
| **26개 전문 에이전트** | Manager 8, Specialist 14, Builder 4 에이전트 자동 위임 | [에이전트 카탈로그](./docs/agents.md) |
| **레거시 마이그레이션** | Vue.js, React CRA, Angular 등 → Next.js 16 자동 전환 | [마이그레이션 가이드](./docs/migration.md) |
| **60개 스킬 시스템** | Progressive Disclosure 기반 지식 로딩 | [스킬 카탈로그](./docs/skills-catalog.md) |
| **품질 보증** | TRUST 5 프레임워크 + LSP 품질 게이트 | [품질 가이드](./docs/rules.md) |
| **LLM 프로바이더 라우터** | OpenAI, Gemini, GLM, Ollama 전환 | [라우터 문서](./docs/provider-router.md) |

---

## 설치

### 방법 1: Install Script (권장)

```bash
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash
```

### 방법 2: go install

```bash
go install github.com/jikime/jikime-adk@latest
```

### 방법 3: 수동 다운로드

[GitHub Releases](https://github.com/jikime/jikime-adk/releases)에서 플랫폼에 맞는 바이너리를 다운로드합니다.

---

## 시작하기

### 1. 프로젝트 초기화

```bash
jikime-adk init
```

### 2. Claude Code에서 사용

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

## 명령어 요약

### Claude Code 슬래시 명령어

| 유형 | 명령어 | 설명 |
|------|--------|------|
| **Workflow** | `/jikime:0-project` ~ `/jikime:3-sync` | 핵심 개발 워크플로우 |
| **J.A.R.V.I.S.** | `/jikime:jarvis`, `/jikime:test`, `/jikime:loop` | 자율 개발 오케스트레이션 |
| **F.R.I.D.A.Y.** | `/jikime:friday`, `/jikime:migrate-*` | 자율 마이그레이션 |
| **Utility** | `/jikime:build-fix`, `/jikime:browser-verify` | 빌드/런타임 에러 수정 |

> 전체 명령어 목록: [명령어 레퍼런스](./docs/commands.md)

### CLI 명령어

| 명령어 | 설명 |
|--------|------|
| `jikime init` | 프로젝트에 템플릿 설치 |
| `jikime update` | 바이너리 자동 업데이트 |
| `jikime doctor` | 시스템 진단 |
| `jikime router switch <provider>` | LLM 프로바이더 전환 |
| `jikime worktree new <branch>` | Git Worktree 생성 |
| `jikime skill list` | 스킬 목록 조회 |

> CLI 상세 옵션: [CLI 문서](./docs/commands.md#cli-명령어)

---

## 에이전트 카탈로그

JiKiME-ADK는 **26개의 전문 에이전트**를 제공합니다:

| 유형 | 수량 | 대표 에이전트 |
|------|------|--------------|
| **Manager** | 8 | manager-spec, manager-ddd, manager-quality |
| **Specialist** | 14 | backend, frontend, security-auditor, optimizer |
| **Builder** | 4 | agent-builder, command-builder, skill-builder |

> 전체 에이전트 목록: [에이전트 카탈로그](./docs/agents.md)

---

## 개발 방법론: DDD

모든 개발에 **ANALYZE-PRESERVE-IMPROVE** 사이클을 적용합니다:

```
ANALYZE   →  현재 동작 이해
    ↓
PRESERVE  →  특성화 테스트로 동작 보존
    ↓
IMPROVE   →  자신감 있게 변경 → (반복)
```

> 상세 내용: [DDD 방법론 문서](./docs/tdd-ddd.md)

---

## 문서

| 문서 | 설명 |
|------|------|
| [에이전트 카탈로그](./docs/agents.md) | 26개 에이전트 상세 역할 |
| [명령어 레퍼런스](./docs/commands.md) | 슬래시 명령어 및 CLI 전체 목록 |
| [스킬 카탈로그](./docs/skills-catalog.md) | 60개 스킬 분류 및 설명 |
| [마이그레이션 가이드](./docs/migration.md) | F.R.I.D.A.Y. 마이그레이션 워크플로우 |
| [DDD 방법론](./docs/tdd-ddd.md) | ANALYZE-PRESERVE-IMPROVE 사이클 |
| [품질 규칙](./docs/rules.md) | TRUST 5, 코딩 스타일, 보안 가이드 |
| [Worktree 관리](./docs/worktree.md) | Git Worktree 병렬 개발 |
| [LLM 라우터](./docs/provider-router.md) | 외부 LLM 프로바이더 연동 |
| [Hooks 시스템](./docs/hooks.md) | Claude Code 훅 설정 |
| [Ralph Loop](./docs/ralph-loop.md) | LSP/AST-grep 피드백 루프 |
| [Statusline](./docs/statusline.md) | Claude Code 상태줄 커스터마이징 |
| [Codemap](./docs/codemap.md) | AST 기반 아키텍처 맵 |

---

## 프로젝트 구조

```
jikime-adk/
├── cmd/                    # CLI 명령어 구현
├── internal/               # 내부 패키지 (라우터 엔진 등)
├── templates/              # 임베디드 프로젝트 템플릿
│   ├── .claude/            # 에이전트, 커맨드, 스킬
│   └── .jikime/            # 설정 파일
├── docs/                   # 문서
└── scripts/                # 자동화 스크립트
```

---

## 감사의 말

- **[MoAI-ADK](https://github.com/modu-ai/moai-adk)** - Goos.Kim님의 에이전틱 워크플로우 철학과 구조적 영감
- **[everything-claude-code](https://github.com/anthropics/anthropic-cookbook)** - 에이전트, 커맨드, 훅 구조 참고

---

## Links

- [GitHub Repository](https://github.com/jikime/jikime-adk)
- [Releases](https://github.com/jikime/jikime-adk/releases)
- [Install Script](https://jikime.github.io/jikime-adk/install.sh)

---

## License

Copyleft License (COPYLEFT-3.0) - See [LICENSE](./LICENSE) for details.
