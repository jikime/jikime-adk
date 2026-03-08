# Auto-Memory

Claude Code의 네이티브 메모리 시스템을 활용한 크로스 세션 프로젝트 컨텍스트 주입 기능입니다.

## 개요

Auto-Memory는 세션 시작 시마다 프로젝트별 메모리 파일을 자동으로 찾아서 Claude의 컨텍스트에 주입합니다. 별도의 수동 컨텍스트 로딩 없이 Claude가 프로젝트에 대한 지속적인 지식을 유지할 수 있게 해줍니다.

```
세션 시작
    ↓
jikime-adk가 ~/.claude/projects/{hash}/memory/*.md 읽기
    ↓
systemMessage에 내용 주입
    ↓
Claude가 처음부터 프로젝트 컨텍스트를 알고 있는 상태로 시작
```

## 동작 원리

### 경로 탐색

Claude Code는 프로젝트 메모리를 다음 위치에 저장합니다:

```
~/.claude/projects/{path-hash}/memory/
```

`{path-hash}`는 프로젝트 경로에서 `/`를 `-`로 치환한 값입니다:

```
/Users/foo/myproject  →  -Users-foo-myproject
```

jikime-adk는 `os.Getwd()` 대신 Claude Code의 stdin 페이로드에서 `cwd` 값을 가져와 올바른 해시를 계산합니다. 이렇게 하면 Claude Code가 내부적으로 사용하는 경로와 항상 일치합니다.

### 세션 시작 흐름

```
Claude Code 세션 시작
    ↓
stdin으로 JSON 전송: {"cwd": "/your/project", "session_id": "..."}
    ↓
jikime hooks session-start가 cwd 읽기
    ↓
ensureMemoryDir() — 디렉토리 없으면 자동 생성
    ↓
discoverAutoMemory() — 모든 .md 파일 읽기
    ↓
formatMemorySection() — systemMessage 섹션 구성
    ↓
반환: {"continue": true, "systemMessage": "...Auto-Memory Loaded..."}
```

### 출력 예시

```
🚀 JikiME-ADK Session Started
   📦 Version: 1.0.0
   🔄 Changes: 5 file(s) modified
   🌿 Branch: master
   ...

---
📚 **Auto-Memory Loaded**
   📁 Path: /Users/foo/.claude/projects/-Users-foo-myproject/memory
   📄 Files: 2 (3104 bytes)

### MEMORY.md
# 프로젝트 메모리

## 아키텍처
- 레거시 PHP에서 Next.js 16 마이그레이션
...

### lessons.md
## 배운 점
- 커밋 전 항상 `pnpm build` 실행
...
---
```

## 메모리 파일 컨벤션

### 우선순위 순서

파일은 다음 순서로 로드되어 표시됩니다:

| 우선순위 | 파일명 | 최대 길이 | 용도 |
|---------|--------|----------|------|
| 1순위 | `MEMORY.md` | 800자 | 주요 프로젝트 메모리 |
| 2순위 | `lessons.md` | 800자 | 배운 교훈 |
| 3순위 | `context.md` | 800자 | 현재 컨텍스트 |
| 나머지 | `*.md` | 400자 | 주제별 노트 |

### MEMORY.md 권장 구조

```markdown
# 프로젝트 메모리

## 아키텍처
- 기술 스택 및 구조 간략 설명

## 주요 결정사항
- 중요한 결정과 그 이유

## 패턴 & 컨벤션
- 코드 패턴, 네이밍 컨벤션

## 최근 작업
- 최근 변경사항 요약
```

## 메모리 파일은 누가 만드나요?

**Claude가 직접 작성합니다** — `Write`와 `Edit` 툴을 사용해서요.

Claude Code의 시스템 프롬프트가 Claude에게 중요한 정보를 메모리 디렉토리에 저장하도록 지시합니다. 다음과 같이 명시적으로 요청할 수도 있습니다:

```
"다음 세션을 위해 현재 프로젝트 구조를 기억해줘"
"API 설계에 대해 논의한 내용을 메모리에 저장해줘"
"오늘 결정한 사항들로 MEMORY.md를 업데이트해줘"
```

직접 파일을 작성할 수도 있습니다:

```bash
# 직접 생성 또는 편집
vim ~/.claude/projects/{hash}/memory/MEMORY.md
```

## 설정

별도 설정 불필요합니다. Auto-Memory는 다음 조건이 충족되면 자동으로 활성화됩니다:

1. jikime-adk 설치 완료 (`go install .` 또는 설치 스크립트)
2. `.claude/settings.json`에 `SessionStart` 훅 등록
3. jikime-adk가 초기화된 프로젝트에서 Claude Code 세션 시작

### 훅 등록 확인

```bash
cat .claude/settings.json | grep -A5 "SessionStart"
```

예상 출력:
```json
"SessionStart": [
  {
    "hooks": [
      {
        "type": "command",
        "command": "jikime hooks session-start"
      }
    ]
  }
]
```

## 테스트

### CLI 테스트

```bash
echo '{"cwd":"/your/project/path"}' | jikime-adk hooks session-start | python3 -m json.tool
```

memory 디렉토리에 `.md` 파일이 있을 때 `systemMessage`에 `Auto-Memory Loaded`가 포함되어 있는지 확인합니다.

### 메모리 디렉토리 확인

```bash
# 프로젝트 해시 찾기
PROJECT_HASH=$(echo "/your/project/path" | sed 's|/|-|g')
ls ~/.claude/projects/${PROJECT_HASH}/memory/
```

### 엔드투엔드 테스트

1. 테스트 메모리 파일 생성:
   ```bash
   echo "# 테스트" > ~/.claude/projects/{hash}/memory/MEMORY.md
   ```
2. 해당 프로젝트에서 새 Claude Code 세션 시작
3. Claude에게 질문: *"이번 세션 시스템 메시지에 어떤 내용이 있어?"*
4. MEMORY.md 내용이 응답에 나타나는지 확인

## 트러블슈팅

### 메모리가 표시되지 않을 때

| 원인 | 해결방법 |
|------|---------|
| 메모리 디렉토리가 비어있음 | Claude에게 MEMORY.md에 내용 작성 요청 |
| 잘못된 프로젝트 해시 | Claude Code의 `cwd`가 예상 경로와 일치하는지 확인 |
| 구버전 바이너리 (< 1.0.0) | `go install github.com/jikime/jikime-adk@latest` 실행 |
| 훅 미등록 | `jikime-adk init` 실행하여 템플릿 재설치 |

### 내용이 잘림

800자(MEMORY.md) 또는 400자(기타)를 초과하는 파일은 잘립니다. 메모리 파일을 간결하게 유지하거나 여러 주제 파일로 분리하세요.

## 관련 문서

- [훅 시스템](./hooks.md) — 전체 훅 시스템 문서
- [Session Start 훅](./hooks.md#session-start) — session-start 훅 상세
