# Provider Router

Claude Code의 API 요청을 외부 LLM 프로바이더(OpenAI, Gemini, GLM, Ollama)로 라우팅하는 프록시 시스템입니다.

## 개요

```
Claude Code ─── ANTHROPIC_BASE_URL ──→ jikime router (localhost:8787/{provider})
                                            │
                                    ┌───────┴───────┐
                                    ▼               ▼
                              OpenAI-Compatible    Gemini API
                              (OpenAI/GLM/Ollama)
```

`ANTHROPIC_BASE_URL` 환경변수를 통해 Claude Code의 API 요청을 가로채고, URL 경로에 지정된 프로바이더의 형식으로 변환하여 전달합니다.

### URL 기반 라우팅

라우터는 URL 경로에서 프로바이더를 식별합니다:

```
http://localhost:8787/{provider}/v1/messages
                       ↑
                       프로바이더 이름 (openai, gemini, ollama 등)
```

이 방식을 통해 **여러 Claude Code 세션이 서로 다른 프로바이더를 동시에 사용**할 수 있습니다. 하나의 라우터 인스턴스가 모든 프로바이더 요청을 처리합니다.

## 지원 프로바이더

| 프로바이더 | 연결 방식 | 설명 |
|-----------|----------|------|
| OpenAI | 프록시 경유 | chat/completions 형식 변환 |
| Gemini | 프록시 경유 | generateContent 형식 변환 |
| GLM | 직접 연결 | Z.ai Anthropic 호환 엔드포인트 사용 |
| Ollama | 프록시 경유 | OpenAI 호환 모드 사용 |

## 설정

### 설정 파일

`~/.jikime/router.yaml`:

```yaml
router:
  port: 8787
  host: "127.0.0.1"

providers:
  openai:
    model: gpt-5.1
    base_url: https://api.openai.com/v1

  gemini:
    model: gemini-2.5-flash
    base_url: https://generativelanguage.googleapis.com

  glm:
    model: glm-4.7
    base_url: https://api.z.ai/api/paas/v4
    anthropic_url: https://api.z.ai/api/anthropic
    region: international

  ollama:
    model: llama3.1
    base_url: http://localhost:11434

# 시나리오 기반 라우팅 (선택)
scenarios:
  default: openai/gpt-5.1
  background: ollama/llama3.1
  think: openai/o1
  long_context: gemini/gemini-2.5-flash
  long_context_threshold: 60000
```

API 키는 `router.yaml`에 저장하지 않습니다. 환경변수에서 자동으로 읽습니다.

### API 키 설정

환경변수로 API 키를 설정합니다:

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="AI..."
export GLM_API_KEY="..."
```

Ollama는 로컬 실행이므로 API 키가 불필요합니다.

## 사용법

### 프로바이더 전환

```bash
# GLM으로 전환 (프록시 불필요, 직접 연결)
jikime router switch glm

# OpenAI로 전환 (프록시 자동 시작)
jikime router switch openai

# Gemini로 전환 (프록시 자동 시작)
jikime router switch gemini

# Ollama로 전환 (프록시 자동 시작)
jikime router switch ollama

# Claude 네이티브로 복원
jikime router switch claude
```

`switch` 명령은 현재 프로젝트의 `.claude/settings.local.json`을 자동 업데이트합니다.
프로젝트 디렉토리(`.git` 또는 `.claude` 존재) 안에서 실행해야 합니다.
변경 후 **Claude Code 재시작**이 필요합니다.

프로젝트별로 다른 프로바이더를 사용할 수 있습니다:

```bash
# 프로젝트 A: 비용 절감 (GLM)
cd ~/projects/project-a && jikime router switch glm

# 프로젝트 B: 고품질 (Claude 네이티브)
cd ~/projects/project-b && jikime router switch claude
```

### 모델 선택

`provider/model` 형식으로 특정 모델을 지정할 수 있습니다:

```bash
# 기본 모델 (router.yaml의 model 값 사용)
jikime router switch openai          # → gpt-5.1 (설정 기본값)

# 특정 모델 지정
jikime router switch openai/gpt-4o-mini
jikime router switch openai/o1
jikime router switch gemini/gemini-2.5-pro
jikime router switch glm/glm-4.7
jikime router switch ollama/deepseek-r1
jikime router switch ollama/llama3.1:70b
```

모델을 지정하면 `router.yaml`의 기본 모델 대신 해당 모델이 사용됩니다.

### 라우터 관리

```bash
# 라우터 시작 (포그라운드)
jikime router start

# 라우터 시작 (백그라운드 데몬)
jikime router start -d

# 포트 지정
jikime router start -d -p 9090

# 상태 확인
jikime router status

# 테스트 요청 전송 (프로바이더 지정 필수)
jikime router test openai
jikime router test gemini

# 라우터 중지
jikime router stop
```

## 동작 방식

### 프록시 모드 (OpenAI, Gemini, Ollama)

1. `switch` 명령이 라우터를 자동 시작 (라우터가 실행 중이면 재시작하지 않음)
2. `.claude/settings.local.json`에 `ANTHROPIC_BASE_URL=http://localhost:8787/{provider}` 설정
3. Claude Code가 로컬 라우터로 요청 전송 (URL 경로에 프로바이더 포함)
4. 라우터가 URL 경로에서 프로바이더를 식별하고 Anthropic 형식 → 프로바이더 형식으로 변환
5. 프로바이더 응답을 Anthropic 형식으로 역변환하여 반환

**멀티 세션 지원**: 하나의 라우터가 모든 프로바이더를 처리하므로, 여러 Claude Code 세션에서 서로 다른 프로바이더를 동시에 사용할 수 있습니다.

### 직접 모드 (GLM)

1. `switch` 명령이 `.claude/settings.local.json`에 직접 설정:
   - `ANTHROPIC_BASE_URL`: Z.ai의 Anthropic 호환 엔드포인트
   - `ANTHROPIC_API_KEY`: GLM API 키
   - `ANTHROPIC_DEFAULT_*_MODEL`: GLM 모델명
2. Claude Code가 Z.ai로 직접 요청 (프록시 불필요)

### 스트리밍

모든 프로바이더에서 SSE(Server-Sent Events) 스트리밍을 지원합니다:

- OpenAI: `chat/completions` SSE → Anthropic SSE
- Gemini: `streamGenerateContent?alt=sse` → Anthropic SSE
- Ollama: OpenAI 호환 SSE → Anthropic SSE

### Tool Use

Claude Code의 tool_use(파일 읽기/쓰기, 코드 실행 등)를 지원합니다:

- OpenAI/Ollama: `tool_calls` ↔ `tool_use` content block 변환
- Gemini: `functionCall`/`functionResponse` ↔ `tool_use`/`tool_result` 변환

### API 파라미터 변환

프로바이더별 API 호환성을 자동으로 처리합니다:

#### OpenAI

- **max_tokens 처리**: 최신 모델(gpt-5.x, o1, o3, o4 시리즈)은 `max_completion_tokens` 파라미터를 사용해야 하지만, 이 모델들은 제한값이 낮으면 400 에러를 반환하므로 출력 토큰 제한을 전송하지 않고 모델 기본값을 사용합니다. 기존 모델(gpt-4o, gpt-4 등)은 `max_tokens`를 정상 전송합니다.
- **tool_choice 변환**: `any` → `required`, `tool` → `function` 형식

#### Gemini

- **JSON Schema 정리**: Gemini가 지원하지 않는 스키마 필드(`exclusiveMinimum`, `additionalProperties`, `propertyNames`, `$schema`, `exclusiveMaximum`)를 재귀적으로 제거합니다. 중첩된 `properties`, `items`, `allOf`/`anyOf`/`oneOf` 내부까지 처리합니다.
- **system_instruction**: 시스템 메시지를 Gemini의 `system_instruction` 형식으로 변환

## Statusline 연동

라우터 상태가 Claude Code의 statusline에 자동 반영됩니다.

### 표시 형식

```
# 라우터 활성 시
🤖 openai/gpt-5.1

# 모델 지정 시
🤖 gemini/gemini-2.5-pro
🤖 glm/glm-4.7
🤖 ollama/deepseek-r1

# Claude 네이티브 사용 시
🤖 Claude Opus 4.5
```

### 동작 원리

`switch` 명령은 `~/.jikime/router-state.json`에 현재 상태를 기록합니다:

```json
{
  "provider": "openai",
  "model": "gpt-5.1",
  "mode": "proxy",
  "active": true
}
```

Statusline은 이 파일을 읽어 `provider/model` 형식으로 표시합니다.
`switch claude` 실행 시 상태 파일이 삭제되어 네이티브 모델명이 표시됩니다.

## 파일 구조

```
cmd/routercmd/
├── router.go          # 상위 명령
├── start.go           # jikime router start
├── stop.go            # jikime router stop
├── status.go          # jikime router status
├── switch.go          # jikime router switch
└── test.go            # jikime router test

internal/router/
├── config.go          # 설정 로더
├── server.go          # HTTP 프록시 서버
├── handler.go         # /v1/messages 핸들러
├── stream.go          # SSE 유틸리티
├── types/
│   └── types.go       # Anthropic API 타입
└── provider/
    ├── provider.go    # Provider 인터페이스
    ├── openai.go      # OpenAI 프로바이더
    ├── gemini.go      # Gemini 프로바이더
    ├── glm.go         # GLM 프로바이더 (OpenAI 래퍼)
    └── ollama.go      # Ollama 프로바이더 (OpenAI 래퍼)
```

## 디버깅

라우터의 요청/응답을 확인하려면 포그라운드 모드로 실행합니다:

```bash
# 기존 라우터 중지
jikime router stop

# 포그라운드로 실행 (로그가 터미널에 출력됨)
jikime router start
```

로그 예시:
```
[router] 2026/01/25 01:30:00 Starting on 127.0.0.1:8787 (providers: openai, gemini, glm, ollama)
[router] 2026/01/25 01:30:05 -> openai/gpt-5.1 (stream=true)
[router] 2026/01/25 01:30:10 <- Provider error (400): {...}
```

## 트러블슈팅

### API 키 오류

```
Error: API key environment variable not set for 'openai'
```

해당 프로바이더의 환경변수가 설정되어 있는지 확인:

```bash
echo $OPENAI_API_KEY
```

### 라우터 연결 실패

```
Error: router is not running
```

라우터를 먼저 시작하거나 `switch`를 사용하세요 (자동 시작됨):

```bash
jikime router switch openai
```

### Claude Code에 적용 안 됨

`switch` 후 반드시 Claude Code를 재시작해야 합니다.
`.claude/settings.local.json`은 세션 시작 시에만 읽힙니다.

### Gemini 400 에러 (스키마 관련)

```
Unsupported value for 'properties.exclusiveMinimum'
```

Gemini가 지원하지 않는 JSON Schema 필드를 사용하는 tool이 있을 때 발생합니다.
라우터가 자동으로 제거하므로, 최신 바이너리로 업데이트하세요:

```bash
jikime-adk update
```

### Ollama 연결 오류

Ollama가 로컬에서 실행 중인지 확인:

```bash
ollama list
# 또는
curl http://localhost:11434/api/tags
```

### GLM 리전 설정

중국 내 사용 시 `router.yaml`에서 리전 변경:

```yaml
glm:
  region: china  # base_url이 open.bigmodel.cn으로 전환됨
```

## 기술 상세

### API 키 관리

API 키는 `router.yaml`에 절대 저장되지 않습니다 (`yaml:"-"` 태그 적용).
환경변수에서만 읽어오며, 메모리에서만 사용됩니다:

| 프로바이더 | 환경변수 |
|-----------|----------|
| OpenAI | `OPENAI_API_KEY` |
| Gemini | `GEMINI_API_KEY` |
| GLM | `GLM_API_KEY` |
| Ollama | 불필요 (로컬) |

### 지원 모델별 토큰 처리

| 모델 그룹 | max_tokens 처리 | 이유 |
|-----------|----------------|------|
| gpt-4o, gpt-4, gpt-3.5 | `max_tokens` 전송 | 레거시 API 호환 |
| gpt-5.x | 미전송 (모델 기본값) | 제한값 낮으면 400 에러 반환 |
| o1, o3, o4 시리즈 | 미전송 (모델 기본값) | `max_completion_tokens` 필수이나 제한 불필요 |

### URL 기반 라우팅 동작

`switch` 명령 실행 시:
1. 라우터가 실행 중이 아니면 백그라운드에서 시작
2. `.claude/settings.local.json`에 `ANTHROPIC_BASE_URL=http://localhost:8787/{provider}` 설정

**멀티 세션 지원**: 라우터는 URL 경로에서 프로바이더를 식별하므로, 한 번 시작된 라우터가 모든 프로바이더 요청을 처리합니다. 프로바이더를 전환해도 라우터를 재시작할 필요가 없습니다.

예시:
- 세션 A: `ANTHROPIC_BASE_URL=http://localhost:8787/openai` → OpenAI 사용
- 세션 B: `ANTHROPIC_BASE_URL=http://localhost:8787/gemini` → Gemini 사용
- 동일한 라우터 인스턴스가 두 세션 모두 처리
