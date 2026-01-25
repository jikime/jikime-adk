# JikiME-ADK Hooks System

Claude Code의 라이프사이클 이벤트와 통합되는 Go 기반 훅 시스템입니다.

## Overview

JikiME-ADK는 Go로 구현된 훅 시스템을 제공하여 Claude Code의 다양한 이벤트에 반응합니다. 모든 훅은 JSON stdin/stdout 프로토콜을 사용하여 Claude Code와 통신합니다.

### 장점

| 특성 | 설명 |
|------|------|
| **단일 바이너리** | 별도 런타임 불필요 (Python, Node.js 등) |
| **빠른 시작** | 밀리초 단위 실행 |
| **크로스 플랫폼** | macOS, Linux, Windows 지원 |
| **일관성** | jikime-adk CLI와 동일한 코드베이스 |

## Hook Categories

### 1. Session Hooks

세션 시작/종료 시 실행되는 훅입니다.

#### session-start

**목적**: 세션 시작 시 프로젝트 정보 및 환경 상태 표시

```bash
jikime hooks session-start
```

**기능**:
- 프로젝트 이름 및 버전 표시
- Git 브랜치 및 변경사항 상태
- Github-Flow 모드 및 Auto Branch 설정
- 최근 커밋 정보
- 대화 언어 설정
- 환경 검증 경고 (v1.1.0+)

**출력 예시**:
```
🚀 JikiME-ADK Session Started
   📦 Version: 1.0.0
   🔄 Changes: 3 file(s) modified
   🌿 Branch: feature/auth
   🔧 Github-Flow: personal | Auto Branch: Yes
   🔨 Last Commit: abc1234 - Add login feature (2 hours ago)
   🌐 Language: 한국어 (ko)
   👋 Welcome back, Anthony!

   ⚠️  Environment Warnings:
      - node_modules not found - run 'npm install' or equivalent
```

**환경 검증 (v1.1.0+)**:

프로젝트 타입을 자동 감지하고 필요한 도구를 확인합니다:

| 프로젝트 타입 | 감지 파일 | 검증 항목 |
|--------------|----------|----------|
| Node.js | `package.json` | node, npm/pnpm/yarn, node_modules |
| Python | `pyproject.toml`, `requirements.txt` | python3, .venv |
| Go | `go.mod` | go |
| Rust | `Cargo.toml` | cargo |

추가 검증:
- `.env.example` 존재 시 `.env` 파일 존재 확인
- Git 설치 여부 확인

#### session-end-cleanup

**목적**: 세션 종료 시 정리 작업 수행

```bash
jikime hooks session-end-cleanup
```

**기능**:
- 임시 파일 정리
- 데스크톱 알림 전송 (v1.1.0+)
- 세션 요약 생성

**데스크톱 알림 (v1.1.0+)**:

크로스 플랫폼 데스크톱 알림을 지원합니다:

| 플랫폼 | 구현 방식 |
|--------|----------|
| macOS | osascript (AppleScript) |
| Linux | notify-send |
| Windows | PowerShell Toast Notification |

알림 비활성화:
```bash
export JIKIME_NO_NOTIFY=1
```

#### session-end-rank

**목적**: 세션 품질 랭킹 및 통계 생성

```bash
jikime hooks session-end-rank
```

### 2. UserPromptSubmit Hooks

사용자 프롬프트 제출 시 실행되는 훅입니다.

#### user-prompt-submit

**목적**: 프롬프트 분석 및 에이전트 힌트 제공 (v1.1.0+)

```bash
jikime hooks user-prompt-submit
```

**기능**:

1. **에이전트 힌트 제공**:

   프롬프트 키워드를 분석하여 적합한 에이전트를 제안합니다:

   | 키워드 | 추천 에이전트 | 힌트 |
   |--------|--------------|------|
   | security, vulnerability, audit | security-auditor | 보안 분석 감지 |
   | performance, optimize, bottleneck | optimizer | 성능 최적화 감지 |
   | test, coverage, unit test | test-guide | 테스트 작업 감지 |
   | refactor, clean up, simplify | refactorer | 리팩토링 감지 |
   | debug, error, fix bug | debugger | 디버깅 감지 |
   | api, endpoint, backend | backend | 백엔드 작업 감지 |
   | component, ui, frontend | frontend | 프론트엔드 작업 감지 |
   | deploy, ci/cd, pipeline | devops | DevOps 작업 감지 |
   | architecture, design, structure | architect | 아키텍처 설계 감지 |
   | document, readme, guide | documenter | 문서화 작업 감지 |
   | database, schema, migration | backend | 데이터베이스 작업 감지 |
   | e2e, playwright, browser | e2e-tester | E2E 테스트 감지 |

2. **위험 패턴 경고**:

   잠재적으로 위험한 명령어 패턴을 감지하여 경고합니다:

   | 패턴 | 경고 메시지 |
   |------|-----------|
   | `rm -rf` | 파괴적 'rm -rf' 명령어 감지 |
   | `git push --force` | Force push 명령어 감지 |
   | `git reset --hard` | Hard reset 명령어 감지 |
   | `DROP TABLE`, `DROP DATABASE` | 데이터베이스 삭제 명령어 감지 |
   | `sudo rm` | 관리자 권한 삭제 명령어 감지 |
   | `chmod 777` | 전체 권한 설정 감지 |
   | `curl \| sh`, `wget \| sh` | 파이프 실행 패턴 감지 |
   | `::: force`, `--no-verify` | 강제 플래그 감지 |
   | `password =`, `secret =` | 잠재적 시크릿 노출 감지 |
   | `*` 와일드카드 삭제 | 와일드카드 삭제 패턴 감지 |

#### orchestrator-route

**목적**: 요청을 적절한 오케스트레이터로 라우팅

```bash
jikime hooks orchestrator-route
```

**기능**:
- 마이그레이션 키워드 감지 → F.R.I.D.A.Y. 활성화
- 개발 키워드 감지 → J.A.R.V.I.S. 활성화
- 오케스트레이터 상태 파일 관리

### 3. PreToolUse Hooks

도구 실행 전 검증을 수행하는 훅입니다.

#### pre-tool-security

**목적**: Write/Edit 도구 실행 전 보안 검증

```bash
jikime hooks pre-tool-security
```

**검증 항목**:
- 민감한 파일 수정 차단 (`.env`, `secrets/`, `~/.ssh/`)
- 하드코딩된 시크릿 패턴 탐지

#### pre-bash

**목적**: Bash 명령어 실행 전 검증

```bash
jikime hooks pre-bash
```

**검증 항목**:
- 위험 명령어 패턴 탐지
- 필요시 사용자 확인 요청

#### pre-write

**목적**: 파일 생성 전 경로 검증

```bash
jikime hooks pre-write
```

**검증 항목**:
- 문서 파일(`.md`, `.txt`) 생성 경로 제한
- 허용된 경로: `README.md`, `CLAUDE.md`, `docs/`, `.jikime/`, `.claude/`, `migrations/`, `SKILL.md`

### 4. PostToolUse Hooks

도구 실행 후 처리를 수행하는 훅입니다.

#### post-tool-formatter

**목적**: 코드 파일 자동 포맷팅

```bash
jikime hooks post-tool-formatter
```

**지원 포맷터**:
- Prettier (JS/TS/JSON/CSS/MD)
- Black (Python)
- gofmt (Go)
- rustfmt (Rust)

#### post-tool-linter

**목적**: 코드 파일 린트 검사

```bash
jikime hooks post-tool-linter
```

**지원 린터**:
- ESLint (JS/TS)
- Ruff (Python)
- golangci-lint (Go)
- Clippy (Rust)

#### post-tool-ast-grep

**목적**: AST 기반 코드 패턴 검사

```bash
jikime hooks post-tool-ast-grep
```

**검사 항목**:
- 보안 취약점 패턴
- 코드 품질 이슈
- 프로젝트별 커스텀 규칙

#### post-tool-lsp

**목적**: LSP 진단 수집 및 보고

```bash
jikime hooks post-tool-lsp
```

**기능**:
- 타입 에러 수집
- 린트 경고 수집
- 품질 게이트 검증

#### post-bash

**목적**: Bash 명령어 실행 후 처리

```bash
jikime hooks post-bash
```

### 5. Stop Hooks

세션 중단/완료 시 실행되는 훅입니다.

#### stop-loop

**목적**: 활성 루프 종료 처리

```bash
jikime hooks stop-loop
```

**기능**:
- 루프 상태 파일 정리
- 최종 상태 보고

#### stop-audit

**목적**: 세션 감사 로그 생성

```bash
jikime hooks stop-audit
```

### 6. Loop Control Hooks

반복 실행 제어를 위한 훅입니다.

#### start-loop

**목적**: 새 루프 세션 시작

```bash
jikime hooks start-loop --task "Fix all errors" --max-iterations 10
```

#### cancel-loop

**목적**: 활성 루프 취소

```bash
jikime hooks cancel-loop
```

### 7. Pre-Compact Hooks

컨텍스트 압축 전 실행되는 훅입니다.

#### pre-compact

**목적**: 압축 전 중요 상태 보존

```bash
jikime hooks pre-compact
```

**기능**:
- 오케스트레이터 상태 보존
- 활성 작업 상태 저장

## Configuration

훅은 `.claude/settings.json`에서 설정합니다:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks session-start"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks session-end-cleanup",
            "timeout": 5000
          },
          {
            "type": "command",
            "command": "jikime hooks session-end-rank",
            "timeout": 5000
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks user-prompt-submit",
            "timeout": 3000
          },
          {
            "type": "command",
            "command": "jikime hooks orchestrator-route",
            "timeout": 3000
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks pre-tool-security",
            "timeout": 5000
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks pre-bash",
            "timeout": 5000
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks post-tool-formatter",
            "timeout": 30000
          },
          {
            "type": "command",
            "command": "jikime hooks post-tool-linter",
            "timeout": 60000
          },
          {
            "type": "command",
            "command": "jikime hooks post-tool-ast-grep",
            "timeout": 30000
          },
          {
            "type": "command",
            "command": "jikime hooks post-tool-lsp",
            "timeout": 30000
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks post-bash",
            "timeout": 10000
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks stop-loop",
            "timeout": 10000
          },
          {
            "type": "command",
            "command": "jikime hooks stop-audit",
            "timeout": 10000
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks pre-compact",
            "timeout": 5000
          }
        ]
      }
    ]
  }
}
```

## Hook Response Format

모든 훅은 JSON 형식으로 응답합니다:

```json
{
  "continue": true,
  "systemMessage": "Hook executed successfully",
  "performance": {
    "go_hook": true
  },
  "error_details": {
    "error": "Optional error message"
  }
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `continue` | boolean | 실행 계속 여부 |
| `systemMessage` | string | 시스템 메시지 (선택) |
| `performance` | object | 성능 관련 메타데이터 (선택) |
| `error_details` | object | 에러 상세 정보 (선택) |

## Related Files

| 파일 | 설명 |
|------|------|
| `cmd/hookscmd/hooks.go` | 훅 명령어 등록 |
| `cmd/hookscmd/session_start.go` | SessionStart 훅 구현 |
| `cmd/hookscmd/session_end_*.go` | SessionEnd 훅 구현 |
| `cmd/hookscmd/user_prompt_submit.go` | UserPromptSubmit 훅 구현 |
| `cmd/hookscmd/orchestrator_route.go` | Orchestrator 라우팅 훅 |
| `cmd/hookscmd/pre_*.go` | PreToolUse 훅 구현 |
| `cmd/hookscmd/post_*.go` | PostToolUse 훅 구현 |
| `cmd/hookscmd/stop_*.go` | Stop 훅 구현 |
| `cmd/hookscmd/loop_*.go` | Loop 제어 훅 구현 |
| `templates/.claude/settings.json` | 훅 설정 템플릿 |

---

Version: 1.1.0
Last Updated: 2026-01-25
Changelog:
- v1.1.0: 환경 검증, 에이전트 힌트, 위험 패턴 경고, 데스크톱 알림 추가
- v1.0.0: 초기 Go 기반 훅 시스템 구현
