# Ralph Loop - Intelligent Iterative Code Improvement

JikiME-ADK의 지능적 반복 코드 개선 시스템. LSP/AST-grep 피드백 기반의 자동화된 코드 수정 루프.

## Overview

Ralph Loop는 Claude Code의 공식 플러그인인 "Ralph Wiggum"에서 영감을 받아 구현되었지만, 단순 반복이 아닌 **진단 기반 지능적 반복**을 제공합니다.

### 차별화 포인트

| 기존 Ralph (단순) | JikiME Ralph (지능적) |
|------------------|----------------------|
| 단순 프롬프트 반복 | LSP/AST-grep 피드백 기반 반복 |
| 에러 0개 = 완료 | 에러 기반 자동 계속 + 다중 조건 |
| 상태 없음 | DiagnosticSnapshot 이력 추적 |
| 고정 조건 | 적응형 완료 조건 |
| 수동 시작 필수 | 자동 감지 및 계속 |

## Architecture

### 핵심 컴포넌트

```
┌─────────────────────────────────────────────────────────────┐
│                    Ralph Loop System                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │ start-loop  │    │  stop-loop  │    │ cancel-loop │     │
│  │   (수동)    │    │   (자동)    │    │   (수동)    │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│         │                  │                  │             │
│         └──────────┬───────┴──────────┬──────┘             │
│                    │                  │                     │
│              ┌─────▼─────┐     ┌──────▼──────┐             │
│              │ LoopState │     │  Diagnostic  │             │
│              │   (.json) │     │  Snapshot    │             │
│              └───────────┘     └─────────────┘             │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              PostToolUse Hooks                       │   │
│  │  ┌────────────┐  ┌─────────────┐  ┌────────────┐   │   │
│  │  │ post-tool- │  │ post-tool-  │  │ post-tool- │   │   │
│  │  │    lsp     │  │  ast-grep   │  │   linter   │   │   │
│  │  └────────────┘  └─────────────┘  └────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 구현 파일 목록

#### Go Hook 파일

| 파일 | 설명 |
|------|------|
| `cmd/hookscmd/loop_state.go` | LoopState, DiagnosticSnapshot, CompletionCriteria 구조체 |
| `cmd/hookscmd/start_loop.go` | 루프 시작 커맨드 (`jikime hooks start-loop`) |
| `cmd/hookscmd/stop_loop.go` | 완료 조건 판정 및 자동 계속 로직 |
| `cmd/hookscmd/cancel_loop.go` | 루프 취소 커맨드 (`jikime hooks cancel-loop`) |
| `cmd/hookscmd/post_tool_lsp.go` | LSP 결과 스냅샷 기록 |
| `cmd/hookscmd/post_tool_ast_grep.go` | AST-grep 결과 스냅샷 기록 |

#### Skill/Command 파일

| 파일 | 설명 |
|------|------|
| `templates/.claude/skills/jikime-workflow-loop/SKILL.md` | Ralph Loop 워크플로우 스킬 |
| `templates/.claude/commands/jikime/loop.md` | `/jikime:loop` 슬래시 커맨드 |

#### 설정 파일

| 파일 | Hook 등록 |
|------|----------|
| `templates/.claude/settings.json` | Stop Hook, PostToolUse Hooks 등록 |

## Data Structures

### LoopState

루프 세션의 전체 상태를 관리합니다.

```go
type LoopState struct {
    // Basic info
    Active           bool      `json:"active"`
    SessionID        string    `json:"session_id"`
    StartedAt        time.Time `json:"started_at"`
    UpdatedAt        time.Time `json:"updated_at"`

    // Iteration info
    Iteration        int       `json:"iteration"`
    MaxIterations    int       `json:"max_iterations"`

    // Task info
    TaskDescription  string    `json:"task_description"`
    TargetFiles      []string  `json:"target_files,omitempty"`

    // Completion criteria
    Criteria         CompletionCriteria `json:"completion_criteria"`

    // Diagnostic history
    Snapshots        []DiagnosticSnapshot `json:"snapshots"`

    // Final result
    CompletionReason string    `json:"completion_reason,omitempty"`
    FinalStatus      string    `json:"final_status,omitempty"`
}
```

### DiagnosticSnapshot

각 반복에서의 진단 결과를 캡처합니다.

```go
type DiagnosticSnapshot struct {
    Iteration      int       `json:"iteration"`
    Timestamp      time.Time `json:"timestamp"`

    // LSP diagnostics
    ErrorCount     int       `json:"error_count"`
    WarningCount   int       `json:"warning_count"`
    InfoCount      int       `json:"info_count"`

    // AST-grep results
    SecurityIssues int       `json:"security_issues"`

    // Test results
    TestsPassed    bool      `json:"tests_passed"`
    TestsRun       int       `json:"tests_run"`
    TestsFailed    int       `json:"tests_failed"`

    // File details
    FileDetails    []FileDetail `json:"file_details,omitempty"`
}
```

### CompletionCriteria

완료 조건을 정의합니다.

```go
type CompletionCriteria struct {
    ZeroErrors      bool `json:"zero_errors"`       // 에러 0개 필수
    ZeroWarnings    bool `json:"zero_warnings"`     // 경고 0개 필수
    ZeroSecurity    bool `json:"zero_security"`     // 보안이슈 0개 필수
    TestsPass       bool `json:"tests_pass"`        // 테스트 통과 필수
    StagnationLimit int  `json:"stagnation_limit"`  // 개선 없는 반복 한계
}
```

## Execution Flow

### 자동 동작 흐름 (기본)

```
사용자: "TypeScript 에러 수정해줘"
        │
        ▼
   Claude가 작업 수행 (Edit/Write)
        │
        ▼
   PostToolUse Hooks 자동 실행
   - post-tool-lsp → 스냅샷 기록
   - post-tool-ast-grep → 스냅샷 기록
        │
        ▼
   Claude 응답 완료 시도
        │
        ▼
   Stop Hook (stop-loop) 자동 실행
        │
        ▼
   ┌─────────────────────────────────────┐
   │ 1. 완료 마커 감지?                  │
   │    YES → exit 0 (강제 종료)         │
   │                                     │
   │ 2. 진단 수집 (ruff, tsc)            │
   │                                     │
   │ 3. 에러 > 0 또는 보안이슈 > 0?      │
   │    YES → exit 1 (자동 계속)         │
   │    NO  → exit 0 (정상 종료)         │
   └─────────────────────────────────────┘
        │
        ├── exit 0 → Claude 정상 종료
        │
        └── exit 1 → 피드백 재주입 → Claude 계속 작업
                     "Ralph Loop: AUTO-CONTINUE |
                      5 error(s) remaining |
                      Next: Fix the remaining errors"
```

### 명시적 루프 흐름 (고급)

```bash
# 1. 루프 시작 (옵션 지정 가능)
jikime hooks start-loop --task "Fix all errors" --max-iterations 10

# 2. Claude 작업 수행
# ... (PostToolUse hooks가 스냅샷 수집)

# 3. Stop Hook이 완료 조건 평가
# - ZeroErrors 체크
# - Stagnation 감지
# - MaxIterations 체크

# 4. 완료 또는 계속
# exit 0: 완료 / exit 1: 계속

# 5. 취소 (필요시)
jikime hooks cancel-loop
```

## Auto-Loop Mechanism

### 핵심 로직 (stop_loop.go)

```go
func runStopLoop(cmd *cobra.Command, args []string) error {
    // 1. 완료 마커 체크 (최우선)
    if checkCompletionPromise(conversationText) {
        return nil // exit 0 - 완료
    }

    // 2. 진단 수집
    currentSnapshot := collectCurrentDiagnostics()

    // 3. 루프 상태 확인
    state := LoadEnhancedLoopState()

    // 4. AUTO-LOOP: 루프가 비활성이어도 에러가 있으면 자동 계속
    if !state.Active {
        if currentSnapshot.ErrorCount > 0 || currentSnapshot.SecurityIssues > 0 {
            // 피드백 출력 후 계속
            os.Exit(1) // exit 1 - 계속
        }
        return nil // exit 0 - 에러 없음
    }

    // 5. 명시적 루프 로직 (생략)
    // ...
}
```

### 진단 수집 방법

```go
func collectCurrentDiagnostics() DiagnosticSnapshot {
    snapshot := DiagnosticSnapshot{}

    // Python: ruff check
    if _, err := exec.LookPath("ruff"); err == nil {
        cmd := exec.Command("ruff", "check", "--output-format=json", ".")
        // E*, F* 코드 → ErrorCount
        // 그 외 → WarningCount
    }

    // TypeScript: tsc --noEmit
    if _, err := exec.LookPath("tsc"); err == nil {
        cmd := exec.Command("tsc", "--noEmit", "--pretty", "false")
        // "error" 포함 라인 → ErrorCount
    }

    // 테스트: pytest / npm test
    snapshot.TestsPassed, _ = checkTests()

    return snapshot
}
```

## Usage

### 자동 모드 (기본)

특별한 명령어 없이 일반 프롬프트를 사용하면 자동으로 동작합니다.

```
사용자: TypeScript 에러 모두 수정해줘
        ↓
Claude: (작업 수행)
        ↓
Stop Hook: 에러 5개 감지 → 자동 계속
        ↓
Claude: (계속 작업)
        ↓
Stop Hook: 에러 0개 → 자동 종료
```

### 명시적 커맨드

```bash
# 기본 사용
/jikime:loop "Fix all TypeScript errors"

# 옵션 지정
/jikime:loop "Remove security vulnerabilities" --max-iterations 5 --zero-security

# 특정 디렉토리
/jikime:loop @src/services/ "Fix all lint errors" --zero-warnings

# 테스트 통과까지
/jikime:loop "Fix failing tests" --tests-pass --max-iterations 10

# 취소
/jikime:loop --cancel
```

### CLI 직접 실행

```bash
# 루프 시작
jikime hooks start-loop \
  --task "Fix all errors" \
  --max-iterations 10 \
  --zero-errors \
  --tests-pass

# 루프 취소
jikime hooks cancel-loop
```

## Completion Markers

Claude가 작업 완료를 선언할 때 사용하는 마커입니다.

```
<jikime>DONE</jikime>
<jikime>COMPLETE</jikime>
<jikime:done />
<jikime:complete />
```

이 마커 중 하나가 감지되면 에러 유무와 관계없이 **즉시 종료**됩니다.

## Completion Criteria

### 옵션

| 옵션 | 플래그 | 기본값 | 설명 |
|------|--------|--------|------|
| Zero Errors | `--zero-errors` | true | 에러 0개 필수 |
| Zero Warnings | `--zero-warnings` | false | 경고 0개 필수 |
| Zero Security | `--zero-security` | false | 보안이슈 0개 필수 |
| Tests Pass | `--tests-pass` | false | 테스트 통과 필수 |
| Max Iterations | `--max-iterations` | 10 | 최대 반복 횟수 |
| Stagnation Limit | `--stagnation-limit` | 3 | 개선 없는 반복 한계 |

### 종료 조건

1. **완료 마커 감지**: 즉시 종료 (exit 0)
2. **모든 조건 충족**: 에러 0개 등 → 종료 (exit 0)
3. **최대 반복 도달**: MaxIterations 초과 → 종료 (exit 0)
4. **정체 감지**: N회 연속 개선 없음 → 종료 (exit 0)
5. **에러 없음 (자동 모드)**: 에러 0개 + 보안이슈 0개 → 종료 (exit 0)

## Safety Features

### 1. 최대 반복 제한

```go
if state.Iteration >= state.MaxIterations {
    state.FinalStatus = "STOPPED"
    state.CompletionReason = "Max iterations reached"
    os.Exit(0)
}
```

### 2. 정체 감지

```go
func (s *LoopState) IsStagnant() bool {
    // 최근 N회 반복에서 개선이 없으면 정체로 판단
    recent := s.Snapshots[len(s.Snapshots)-limit:]
    // 이슈 수가 감소하지 않으면 true
}
```

### 3. 비활성화

```bash
# 환경변수로 비활성화
export JIKIME_DISABLE_LOOP_CONTROLLER=1
```

### 4. 수동 취소

```bash
jikime hooks cancel-loop
# 또는
/jikime:loop --cancel
```

## Configuration

### settings.json Hook 등록

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
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
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime hooks stop-loop",
            "timeout": 10000
          }
        ]
      }
    ]
  }
}
```

## Feedback Messages

### AUTO-CONTINUE (자동 계속)

```
Ralph Loop: AUTO-CONTINUE | Issues detected - continuing automatically |
5 error(s) remaining | 2 security issue(s) remaining |
Next: Fix the remaining errors | Output <jikime:done /> when complete
```

### CONTINUE (명시적 루프)

```
Ralph Loop: CONTINUE | Iteration: 3/10 |
Current: 5 error(s), 12 warning(s), 0 security issue(s) |
Progress: 45% improvement | Next: Fix 5 remaining error(s)
```

### COMPLETE (완료)

```
Ralph Loop: COMPLETE - All conditions satisfied |
Session: loop-1705912345 | Iterations: 5 |
Total improvement: 100% | Initial: 12 errors, 28 warnings |
Final: 0 errors, 8 warnings
```

## Integration with DDD

Ralph Loop는 Domain-Driven Development 워크플로우와 통합됩니다.

```
┌─────────────────────────────────────────────────────────────┐
│                   DDD + Ralph Loop                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ANALYZE: 루프가 진단 데이터 수집                           │
│     ↓                                                       │
│  PRESERVE: 각 반복에서 기존 동작 검증                       │
│     ↓                                                       │
│  IMPROVE: 측정 가능한 진행률로 점진적 수정                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Troubleshooting

### 루프가 시작되지 않음

```bash
# 다른 루프가 활성 상태인지 확인
jikime hooks cancel-loop

# CLI 설치 확인
jikime --version
```

### 루프가 멈추지 않음

```bash
# 완료 마커 출력
# Claude에게 "<jikime:done />" 출력하도록 요청

# 또는 강제 취소
/jikime:loop --cancel
```

### 진단이 수집되지 않음

```bash
# 도구 설치 확인
which ruff
which tsc

# 프로젝트 설정 확인
ls tsconfig.json
ls pyproject.toml
```

### Hook이 동작하지 않음

```bash
# settings.json 확인
cat .claude/settings.json | jq '.hooks.Stop'

# 수동 테스트
echo '{"messages":[]}' | jikime hooks stop-loop
```

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-01-22 | 초기 구현 - 에러 기반 자동 계속 |

## References

- [Ralph Wiggum Plugin](https://github.com/anthropics/claude-code-plugins) - 원본 영감
- [Claude Code Hooks](https://docs.anthropic.com/claude-code/hooks) - Hook 시스템 문서
- [JikiME-ADK](https://github.com/jikime/jikime-adk) - 프로젝트 저장소
