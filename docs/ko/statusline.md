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
| ▰▱ | Progress | 컨텍스트 윈도우 사용량 (Progress Bar) |
| 💵 | Cost | 예상 토큰 비용 (예: $0.23) |
| 💬 | OutputStyle | 응답 스타일/페르소나 (예: J.A.R.V.I.S.) |
| 📁 | Directory | 현재 프로젝트 디렉토리 |
| 🔀 | Branch | Git 브랜치 및 상태 |
| 💾 | Memory | 메모리 사용량 |
| ⚡ | CPU | 시스템 CPU 부하 |
| 💿 | Disk | 디스크 사용량 |
| 🌐 | Network | API 응답 지연 |
| 🌤️ | Weather | 현재 날씨 (선택적) |
| ⏱️ | Duration | 세션 지속 시간 |
| 🎯 | Task | 활성 작업 표시 |
| 📦 | Version | JikiME-ADK 버전 |
| 🔄 | Update | 업데이트 가능 여부 |

## 사용법

```bash
# 기본 사용 (extended 모드, progress bar 포함)
jikime statusline

# 특정 모드로 표시
jikime statusline --mode compact
jikime statusline --mode minimal
jikime statusline --mode geek

# 데모 보기
jikime statusline --demo

# Pretty 박스 형식으로 표시
jikime statusline --pretty
```

## 디스플레이 모드

### Minimal

모델과 컨텍스트만 표시합니다.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 7%
```

### Compact

핵심 정보를 압축해서 표시합니다.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 15K/200K 7% ┃ 💵 $0.23 ┃ 💬 J.A.R.V.I.S. ┃ 🔀 main +0 M5 ?5 ┃ 💾 128MB ┃ ⚡ 45% ┃ ☀️ +12°C
```

### Extended (기본)

균형 잡힌 정보를 progress bar와 함께 표시합니다.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 15K/200K ┃ 💵 $0.23 ┃ 💬 J.A.R.V.I.S. ┃ 📁 jikime-adk ┃ 🔀 main +0 M5 ?5 ┃ 💾 128MB ┃ ⚡ 45% ┃ ☀️ +12°C ┃ ⏱️ 45m ┃ 🎯 IMPLEMENT ┃ 📦 v2.0.0
```

### Geek (전체 기능)

모든 기능을 포함한 개발자 모드입니다. 색상 코딩된 progress bar와 모든 시스템 정보를 표시합니다.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 15K/200K (7%) ┃ 💵 $0.23 ┃ 💬 J.A.R.V.I.S. ┃ 📁 jikime-adk ┃ 🔀 main +0 M5 ?5 ┃ 💾 128MB ┃ ⚡ 45% ┃ 💿 120GB (65%) ┃ 🌐 120ms ┃ ☀️ +12°C ┃ ⏱️ 45m ┃ 📦 v2.0.0
```

## Progress Bar

컨텍스트 사용량을 시각적으로 표시합니다:

```
▱▱▱▱▱▱▱▱▱▱ = 0%
▰▰▰▰▰▱▱▱▱▱ = 50%
▰▰▰▰▰▰▰▰▰▰ = 100%
```

색상 코딩 (Geek 모드):
- 🟢 초록: 0-49% (정상)
- 🟡 노랑: 50-79% (주의)
- 🔴 빨강: 80-100% (경고)

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
  mode: extended  # extended | compact | minimal | geek
  refresh_interval_ms: 1000

  display:
    model: true
    version: true
    context_window: true
    output_style: true
    memory_usage: true
    branch: true
    git_status: true
    duration: true
    directory: true
    active_task: true
    update_indicator: true
    # New features
    token_cost: true
    cpu_load: true
    disk_usage: false    # opt-in
    network_latency: false # opt-in
    weather: false       # opt-in
    progress_bar: true

  weather:
    enabled: false
    location: ""  # Empty = auto-detect by IP
    unit: "celsius"  # celsius | fahrenheit

  token_cost:
    input_price_per_mtok: 15.0   # $15 per 1M input tokens
    output_price_per_mtok: 75.0  # $75 per 1M output tokens

  format:
    max_branch_length: 30
    truncate_with: "..."
    separator: " ┃ "

  cache:
    git_ttl_seconds: 10
    update_ttl_seconds: 600
```

## 새로운 기능

### Token Cost (💵)

현재 세션의 예상 API 비용을 표시합니다. Claude Opus 기준으로 계산됩니다:
- Input: $15 / 1M tokens
- Output: $75 / 1M tokens

### System Status

시스템 상태 정보를 표시합니다:
- **CPU Load (⚡)**: `uptime` 명령어로 시스템 부하 측정
- **Disk Usage (💿)**: `df` 명령어로 현재 디렉토리 디스크 사용량
- **Network Latency (🌐)**: Anthropic API 응답 지연 시간 (60초 캐시)

### Weather (🌤️)

현재 날씨를 표시합니다 (wttr.in API 사용, 30분 캐시):
- 위치 자동 감지 또는 수동 설정
- 섭씨/화씨 단위 선택 가능

## 환경 변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `JIKIME_STATUSLINE_MODE` | 디스플레이 모드 | `extended` |

## 관련 파일

| 파일 | 설명 |
|------|------|
| `cmd/statuslinecmd/statusline.go` | Statusline 명령 구현 |
| `.jikime/config/statusline-config.yaml` | 설정 파일 |
| `~/.jikime/metrics/session.json` | 세션 시작 시간 (Duration 계산용) |
| `~/.jikime/state/active_task.json` | 활성 작업 정보 |
