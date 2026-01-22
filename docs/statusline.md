# Statusline

Claude Code의 터미널 하단에 세션 상태 정보를 표시하는 커스텀 렌더러입니다.

## 개요

`jikime statusline` 명령은 Claude Code가 statusline 기능을 사용할 때 호출하는 외부 렌더러입니다. Claude Code로부터 세션 컨텍스트를 JSON으로 받아 포맷된 상태 문자열을 반환합니다.

```
Claude Code → stdin (JSON) → jikime statusline → stdout (상태 문자열)
```

## 표시 정보

| 아이콘 | 항목 | 설명 |
|--------|------|------|
| 🤖 | Model | AI 모델명 (예: Opus 4.5) |
| 💰 | Context | 컨텍스트 윈도우 사용량 (예: 15K/200K) |
| 💬 | Style | 출력 스타일 이름 |
| 📁 | Directory | 현재 프로젝트 디렉토리 |
| 📊 | GitStatus | Git 변경사항 (+staged M수정 ?추적안됨) |
| 💾 | Memory | 메모리 사용량 |
| 🔀 | Branch | Git 브랜치명 |
| ⏱️ | Duration | 세션 지속 시간 |
| 🎯 | Task | 활성 작업 표시 |
| 📦 | Version | JikiME-ADK 버전 |
| 🔄 | Update | 업데이트 가능 여부 |

## 사용법

```bash
# 기본 사용 (extended 모드)
jikime statusline

# 특정 모드로 표시
jikime statusline --mode compact
jikime statusline --mode minimal

# 데모 보기
jikime statusline --demo

# Pretty 박스 형식으로 표시
jikime statusline --pretty
```

## 디스플레이 모드

### Extended (기본)

모든 정보를 표시합니다.

```
🤖 Opus 4.5 │ 💰 15K/200K │ 💬 Mr.Alfred │ 📁 jikime-adk │ 📊 +0 M5 ?5 │ 💾 128MB │ 🔀 main │ ⏱️ 45m │ 🎯 IMPLEMENT │ 📦 v2.0.0
```

### Compact

80자 이내로 핵심 정보만 표시합니다.

```
🤖 Opus 4.5 │ 💰 15K/200K │ 💬 Mr.Alfred │ 📁 jikime-adk │ 📊 +0 M5 ?5 │ 💾 128MB │ 🔀 main
```

### Minimal

40자 이내, 가장 핵심적인 정보만 표시합니다.

```
🤖 Opus 4.5 │ 💰 15K/200K
```

## Claude Code 설정

Claude Code의 `settings.json`에서 statusline을 활성화합니다:

```json
{
  "statusline": {
    "enabled": true,
    "command": "jikime statusline"
  }
}
```

## 설정 파일

`.jikime/config/statusline-config.yaml` 파일로 statusline을 커스터마이징할 수 있습니다:

```yaml
statusline:
  enabled: true
  mode: extended  # extended | compact | minimal
  refresh_interval_ms: 1000

  display:
    model: true
    version: true
    context_window: true
    output_style: true
    memory_usage: true
    todo_count: true
    branch: true
    git_status: true
    duration: true
    directory: true
    active_task: true
    update_indicator: true

  format:
    max_branch_length: 30
    truncate_with: "..."
    separator: " │ "
    icons:
      git: "🔀"
      git_status: "📊"
      model: "🤖"
      claude_version: "🤖"
      context_window: "💰"
      output_style: "💬"
      duration: "⏱️"
      update: "🔄"
      project: "📁"

  cache:
    git_ttl_seconds: 10
    update_ttl_seconds: 600
```

## 세션 컨텍스트 구조

Claude Code가 전달하는 JSON 구조:

```json
{
  "model": {
    "display_name": "Opus 4.5",
    "name": "claude-opus-4-5-20251101"
  },
  "version": "2.0.46",
  "cwd": "/path/to/project",
  "output_style": {
    "name": "Mr.Alfred"
  },
  "context_window": {
    "context_window_size": 200000,
    "total_input_tokens": 15000,
    "current_usage": {
      "input_tokens": 10000,
      "cache_creation_input_tokens": 3000,
      "cache_read_input_tokens": 2000
    }
  },
  "statusline": {
    "mode": "extended"
  }
}
```

## 환경 변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `JIKIME_STATUSLINE_MODE` | 디스플레이 모드 | `extended` |

## 업데이트 확인

statusline은 GitHub Releases API를 통해 JikiME-ADK의 새 버전을 확인합니다:

- 캐시 TTL: 10분 (설정 가능)
- 새 버전이 있으면 `🔄 x.x.x available` 표시

## 관련 파일

| 파일 | 설명 |
|------|------|
| `cmd/statuslinecmd/statusline.go` | Statusline 명령 구현 |
| `.jikime/config/statusline-config.yaml` | 설정 파일 |
| `~/.jikime/metrics/session.json` | 세션 시작 시간 (Duration 계산용) |
| `~/.jikime/state/active_task.json` | 활성 작업 정보 |
