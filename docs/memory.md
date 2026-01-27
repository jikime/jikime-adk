# JikiME-ADK 메모리 시스템 설계

**버전**: 1.1.0
**최종 업데이트**: 2026-01-27
**상태**: 설계 단계 (Design Phase) - Clawdbot 코드 분석 및 Hook API 검증 완료

---

## 1. 개요

JikiME-ADK 메모리 시스템은 Claude Code 세션 간 컨텍스트 연속성을 제공하는 시스템이다.
Claude Code의 공식 Hooks API를 활용하여 대화 이력, 학습 내용, 프로젝트 지식을 자동으로 수집하고,
다음 세션에서 LLM에 주입하여 "기억하는 AI 어시스턴트"를 구현한다.

### 핵심 가치

- **세션 연속성**: 이전 세션의 맥락을 자동으로 이어받아 작업 효율 향상
- **프로젝트 지식 축적**: 파일 수정 이력, 의사결정 기록을 누적하여 프로젝트 이해도 향상
- **학습 기반 최적화**: 실패 패턴 학습으로 반복 오류 회피

---

## 2. 기존 시스템 분석

### 2.1 참조 시스템 비교

| 항목 | Clawdbot Memory | jikime-mem | jikime-adk Hooks |
|------|----------------|------------|------------------|
| **언어** | TypeScript | TypeScript/Bun | Go |
| **저장소** | 마크다운 파일 (원본) + SQLite (인덱스) | SQLite + Chroma | 파일 기반 |
| **벡터 검색** | sqlite-vec 내장 | Chroma 별도 서버 | 미구현 |
| **검색 방식** | Hybrid (0.7 벡터 + 0.3 텍스트) | 벡터 전용 | N/A |
| **세션 관리** | Two-Layer Memory | 단일 레이어 | Hook 기반 |
| **메모리 접근** | Tool 기반 (memory_search + memory_get) | Hook 기반 | Hook 기반 (additionalContext) |
| **임베딩 프로바이더** | OpenAI / Gemini / Local (3중 폴백) | Chroma 내장 | 미구현 |
| **압축 전 저장** | Pre-compaction Memory Flush | PreCompact 훅 | PreCompact 훅 |
| **재인덱싱** | Atomic Reindex (임시 DB → 스왑) | 없음 | 미구현 |

### 2.2 Clawdbot 메모리 시스템 코드 분석

> **분석 대상**: `clawdbot/src/memory/`, `clawdbot/src/agents/tools/`
> **GitHub Stars**: 32,600+ (2026-01 기준), MIT 라이선스

#### 2.2.1 핵심 아키텍처: "마크다운이 원본, DB는 인덱스"

Clawdbot의 가장 중요한 설계 원칙은 **마크다운 파일이 진실의 원천(source of truth)** 이라는 점이다.
SQLite DB는 검색용 파생 인덱스일 뿐이며, 언제든 원본 마크다운에서 재구축 가능하다.

```
~/clawd/                          # 에이전트 워크스페이스 (원본)
├── MEMORY.md                     # Layer 2: 장기 기억 (큐레이션된 지식)
└── memory/
    ├── 2026-01-26.md             # Layer 1: 일일 노트 (append-only)
    ├── 2026-01-25.md
    └── ...

~/.clawdbot/memory/               # 상태 디렉토리 (파생 인덱스)
├── main.sqlite                   # 벡터 인덱스 + FTS5 인덱스
└── work.sqlite                   # 에이전트별 격리
```

#### 2.2.2 Two-Layer Memory (2계층 메모리)

**Layer 1: Daily Logs (`memory/YYYY-MM-DD.md`)**
- Append-only 일일 노트. 에이전트가 기억할 내용을 시간순으로 기록
- 결정사항, 배포 이력, 사용자 선호도 등 즉시 기록

**Layer 2: Long-term Memory (`MEMORY.md`)**
- 큐레이션된 영구 지식. 중요한 이벤트, 결정, 교훈을 정리
- 사용자 선호도, 주요 결정사항, 핵심 연락처 등 구조화된 정보

#### 2.2.3 메모리 도구 (Tool 기반 접근)

Clawdbot은 **전용 memory_write 도구 없이** 표준 write/edit 도구로 마크다운 파일을 직접 수정한다.
읽기는 전용 검색 도구 2개를 제공한다:

| 도구 | 목적 | 핵심 파라미터 |
|------|------|-------------|
| `memory_search` | 시맨틱 검색 | `query`, `maxResults: 6`, `minScore: 0.35` |
| `memory_get` | 특정 라인 읽기 | `path`, `from` (라인번호), `lines` (줄 수) |

> **jikime-adk와의 차이**: Clawdbot은 Tool 기반으로 에이전트가 능동적으로 메모리를 검색하지만,
> jikime-adk는 Hook의 `additionalContext`를 통해 세션 시작 시 자동 주입하는 방식이다.

#### 2.2.4 인덱싱 파이프라인

```
파일 저장/변경 감지 (Chokidar, 1.5초 디바운스)
    ↓
청킹 (400토큰 단위, 80토큰 오버랩)
    ↓
임베딩 (OpenAI text-embedding-3-small → Gemini → Local 폴백)
    ↓
SQLite 저장 (chunks + chunks_vec + chunks_fts + embedding_cache)
```

**핵심 구현 파일**:
- `src/memory/manager.ts` (2178줄) — 메모리 인덱스 매니저 핵심 클래스
- `src/memory/internal.ts` — 청킹 파이프라인 (`chunkMarkdown()`)
- `src/memory/manager-search.ts` — 벡터/키워드 검색 구현
- `src/memory/hybrid.ts` — 하이브리드 결과 병합
- `src/memory/memory-schema.ts` — SQLite 스키마 정의

**SQLite 테이블 구조**:
| 테이블 | 용도 |
|--------|------|
| `chunks` | 인덱싱된 텍스트 청크 (id, path, start_line, end_line, hash, text, embedding) |
| `chunks_vec` | sqlite-vec 벡터 저장소 (코사인 유사도 검색) |
| `chunks_fts` | FTS5 전문 검색 (BM25 랭킹) |
| `embedding_cache` | 임베딩 캐시 (해시 기반, 동일 청크 재임베딩 방지) |
| `files` | 파일 추적 (path, hash, mtime, size) |
| `meta` | 설정 메타데이터 (model, provider, chunk 설정) |

#### 2.2.5 하이브리드 검색 구현

```
검색 쿼리
    ├─→ searchVector() → sqlite-vec vec_distance_cosine() → score = 1 - distance
    ├─→ searchKeyword() → FTS5 bm25() → score = 1 / (1 + rank)
    └─→ mergeHybridResults() → finalScore = 0.7 * vectorScore + 0.3 * textScore
         └─→ minScore(0.35) 필터링 → maxResults(6) 제한
```

#### 2.2.6 Pre-compaction Memory Flush

```
컨텍스트 75% 도달 (contextWindow - reserve - softThreshold)
    ↓
Silent Memory Flush Turn (사용자에게 보이지 않음)
    ↓
System: "Pre-compaction memory flush. Store durable memories now."
    ↓
에이전트가 memory/YYYY-MM-DD.md에 중요 내용 기록
    ↓
Compaction 안전하게 진행 (중요 정보 이미 디스크에 저장됨)
```

**설정값** (`src/auto-reply/reply/memory-flush.ts`):
- `enabled: true`
- `softThresholdTokens: 4000`
- `reserveTokensFloor: 20000`

#### 2.2.7 Clawdbot에서 차용할 핵심 패턴

| # | 패턴 | 설명 | jikime-adk 적용 방안 |
|---|------|------|---------------------|
| 1 | **Atomic Reindex** | 임시 DB로 재인덱싱 후 원자적 스왑, 손상 방지 | SQLite DB 스키마 변경 시 적용 |
| 2 | **Embedding Cache** | 해시 기반 캐시로 동일 청크 재임베딩 방지 | embedding_cache 테이블 도입 |
| 3 | **Provider Fallback** | OpenAI → Gemini → Local 자동 폴백 체인 | Go에서 복수 프로바이더 지원 |
| 4 | **Delta Tracking** | 바이트/메시지 수 변화량 추적, 불필요한 재인덱싱 방지 | 파일 해시 기반 변경 감지 |
| 5 | **Chunk Overlap** | 400토큰 청크 + 80토큰 오버랩으로 경계 정보 보존 | 동일 설정값 적용 |
| 6 | **Hybrid Search** | 0.7 벡터 + 0.3 BM25 가중치 병합 | Phase 2에서 구현 |

### 2.3 jikime-mem 플러그인 현황

기존 `jikime-mem` 플러그인 (TypeScript/Bun 기반 Claude Code 플러그인):

- **구조**: `.claude-plugin` 매니페스트 기반
- **저장소**: SQLite (메타데이터) + Chroma (벡터 임베딩)
- **한계**:
  - Chroma 서버 별도 실행 필요 (운영 복잡도 증가)
  - Bun 런타임 의존 (Go 기반 jikime-adk와 기술 스택 불일치)
  - 벡터 검색만 지원 (하이브리드 검색 미지원)

### 2.4 Clawdbot vs jikime-adk 설계 차이점

| 항목 | Clawdbot 방식 | jikime-adk 방식 | 이유 |
|------|-------------|----------------|------|
| **원본 저장소** | 마크다운 파일 | SQLite DB | Go 바이너리에서 마크다운 관리보다 DB가 효율적 |
| **메모리 접근** | Tool 기반 (에이전트가 능동 검색) | Hook 기반 (세션 시작 시 자동 주입) | Claude Code Hook API의 additionalContext 활용 |
| **쓰기** | write/edit 도구로 마크다운 직접 수정 | Hook 핸들러가 transcript 파싱 후 DB 저장 | 자동화 우선 (사용자 개입 최소화) |
| **임베딩** | OpenAI/Gemini/Local 3중 폴백 | Phase 2에서 외부 API 도입 | MVP는 FTS5 텍스트 검색으로 충분 |
| **인덱싱** | Chokidar 파일 감시 + 디바운스 | Hook 이벤트 트리거 기반 | Hook이 이미 이벤트를 제공하므로 파일 감시 불필요 |

---

## 3. Claude Code Hooks API 레퍼런스

> **출처**: [공식 Claude Code Hooks 문서](https://code.claude.com/docs/en/hooks)
> **검증일**: 2026-01-27 (공식 문서 직접 확인)

### 3.1 전체 Hook Event 목록 (12종)

| Hook Event | 타이밍 | 설명 | 매처 지원 |
|-----------|--------|------|----------|
| `SessionStart` | 세션 시작/재개 시 | source: `startup`, `resume`, `clear`, `compact` | O (source별) |
| `Setup` | 초기 설정 시 (`--init`, `--maintenance`) | `CLAUDE_ENV_FILE` 환경변수 접근 가능 | O (`init`, `maintenance`) |
| `UserPromptSubmit` | 사용자 프롬프트 제출 후 | prompt 내용 수신, 변환/차단 가능 | X |
| `PreToolUse` | 도구 실행 전 | tool_name, tool_input 수신, 차단/승인/수정 가능 | O (도구명 정규식) |
| `PermissionRequest` | 권한 요청 시 | 자동 승인/거부 가능 | O (도구명 정규식) |
| `PostToolUse` | 도구 실행 성공 후 | tool_result 포함 | O (도구명 정규식) |
| `PostToolUseFailure` | 도구 실행 실패 후 | error 정보 포함 | X |
| `SubagentStart` | 서브에이전트 시작 시 | agent_id, agent_type 정보 | X |
| `SubagentStop` | 서브에이전트 종료 시 | agent_transcript_path 포함 | X |
| `Stop` | 응답 완료 시 | stop_hook_active 플래그 | X |
| `PreCompact` | 컨텍스트 압축 전 | trigger: `manual`/`auto` | O (trigger별) |
| `SessionEnd` | 세션 종료 시 | reason: `exit`, `clear`, `logout` 등 | X |
| `Notification` | 알림 발생 시 | notification_type 포함 | O (알림 타입별) |

### 3.2 Hook 설정 구조

Hook은 `.claude/settings.json` (프로젝트), `~/.claude/settings.json` (사용자), 또는 관리 정책에서 설정한다.

```json
{
  "hooks": {
    "EventName": [
      {
        "matcher": "ToolPattern",
        "hooks": [
          {
            "type": "command",
            "command": "your-command-here",
            "timeout": 60
          }
        ]
      }
    ]
  }
}
```

**Hook 타입**:
- `"command"` — bash 명령 실행 (모든 이벤트)
- `"prompt"` — LLM(Haiku)에게 판단 위임 (`Stop`, `SubagentStop` 등에 유용)

**매처 패턴**:
- 정확한 도구명: `"Write"` → Write 도구만
- 정규식: `"Edit|Write"` 또는 `"Notebook.*"`
- 전체 매칭: `"*"` 또는 `""`
- MCP 도구: `"mcp__server__tool"` 패턴

### 3.3 Hook I/O 프로토콜

**입력**: stdin으로 JSON 전달

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../.claude/projects/.../abc.jsonl",
  "cwd": "/Users/project-dir",
  "permission_mode": "default",
  "hook_event_name": "SessionStart"
}
```

**공통 입력 필드**: `session_id`, `transcript_path`, `cwd`, `permission_mode`, `hook_event_name`

**출력 방식 2가지** (상호 배타적):

1. **Simple: Exit Code**
   - Exit `0` = 성공. stdout는 verbose 모드에서만 표시 (단, `UserPromptSubmit`와 `SessionStart`는 컨텍스트에 추가)
   - Exit `2` = 블로킹 에러. stderr만 에러 메시지로 사용
   - 기타 = 비차단 에러

2. **Advanced: JSON Output** (exit 0일 때만 처리)

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "여기에 메모리 컨텍스트 주입"
  }
}
```

**환경 변수**:
- `CLAUDE_PROJECT_DIR` → 모든 Hook에서 접근 가능
- `CLAUDE_ENV_FILE` → `SessionStart`와 `Setup`에서만 접근 가능
- `CLAUDE_CODE_REMOTE` → 원격(웹) 환경 여부 (`"true"` 또는 미설정)

**실행 특성**:
- 기본 타임아웃: 60초 (개별 설정 가능)
- 매칭된 모든 Hook은 **병렬 실행**
- 동일 명령은 자동 중복 제거

### 3.4 주요 Hook Input 스키마

**SessionStart Input**:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/session.jsonl",
  "cwd": "/project/dir",
  "permission_mode": "default",
  "hook_event_name": "SessionStart",
  "source": "startup",
  "model": "claude-sonnet-4-20250514"
}
```
- `source` 값: `startup` (새 세션), `resume` (재개), `clear` (/clear 후), `compact` (압축 후)

**PreCompact Input**:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/session.jsonl",
  "permission_mode": "default",
  "hook_event_name": "PreCompact",
  "trigger": "manual",
  "custom_instructions": ""
}
```
- `trigger` 값: `manual` (/compact 명령), `auto` (자동 압축)

**UserPromptSubmit Input**:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/session.jsonl",
  "permission_mode": "default",
  "hook_event_name": "UserPromptSubmit",
  "prompt": "Write a function..."
}
```

**PostToolUse Input**:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/session.jsonl",
  "permission_mode": "default",
  "hook_event_name": "PostToolUse",
  "tool_name": "Write",
  "tool_input": { "file_path": "/path/to/file.txt", "content": "..." },
  "tool_response": { "filePath": "/path/to/file.txt", "success": true },
  "tool_use_id": "toolu_01ABC123..."
}
```

**Stop Input**:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/session.jsonl",
  "permission_mode": "default",
  "hook_event_name": "Stop",
  "stop_hook_active": true
}
```
- `stop_hook_active`: Stop Hook으로 인해 이미 계속 중인 경우 `true` → 무한 루프 방지에 활용

**SubagentStop Input**:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/main-session.jsonl",
  "permission_mode": "default",
  "hook_event_name": "SubagentStop",
  "stop_hook_active": false,
  "agent_id": "def456",
  "agent_transcript_path": "/path/to/subagents/agent-def456.jsonl"
}
```

**SessionEnd Input**:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/session.jsonl",
  "permission_mode": "default",
  "hook_event_name": "SessionEnd",
  "reason": "exit"
}
```
- `reason` 값: `exit`, `clear`, `logout`, `prompt_input_exit`, `other`

### 3.5 Context 주입 메커니즘

다음 Hook들이 `additionalContext`를 통해 LLM에 컨텍스트를 주입할 수 있다:

| Hook | 주입 방식 | 비고 |
|------|----------|------|
| **SessionStart** | `hookSpecificOutput.additionalContext` | 세션 시작 시 메모리 로드 |
| **UserPromptSubmit** | `hookSpecificOutput.additionalContext` 또는 plain text stdout | 프롬프트별 추가 컨텍스트 |
| **Setup** | `hookSpecificOutput.additionalContext` | 초기 설정 시 |
| **PreToolUse** | `hookSpecificOutput.additionalContext` | 도구 실행 전 컨텍스트 |
| **PostToolUse** | `hookSpecificOutput.additionalContext` | 도구 실행 후 피드백 |

> **메모리 시스템 핵심**: `SessionStart`의 `additionalContext`가 세션 간 메모리를 LLM에 전달하는 **주요 메커니즘**이다.
> `UserPromptSubmit`는 프롬프트별 동적 메모리 검색에 활용할 수 있다 (Phase 2).

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "## 이전 세션 기억\n- SPEC-SERVICE-001 작업 중\n- Slack 채널 어댑터 구현 완료"
  }
}
```

### 3.6 Hook 설정 보안

- Hook 설정은 **세션 시작 시 스냅샷**으로 캡처됨
- 세션 중 외부 수정 시 경고 표시, `/hooks` 메뉴에서 검토 필요
- 기업 관리자는 `allowManagedHooksOnly`로 사용자/프로젝트/플러그인 Hook 차단 가능

---

## 4. 메모리 시스템용 Hook 설계

### 4.1 Hook Lifecycle과 메모리 매핑

> **출처**: Claude Code 공식 Hook Lifecycle 다이어그램

Claude Code의 Hook은 다음 순서로 실행된다. 메모리 시스템은 이 lifecycle의 특정 지점에 개입한다.

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ┌──────────────┐                                               │
│  │ SessionStart │ ◄───────────── [compact 후 재진입] ◄──────┐   │
│  │  (녹색)      │                                            │   │
│  └──────┬───────┘                                            │   │
│         │  ★ Phase 1: 메모리 로드 (additionalContext 주입)   │   │
│         ▼                                                    │   │
│  ┌──────────────────┐                                        │   │
│  │UserPromptSubmit  │                                        │   │
│  └──────┬───────────┘                                        │   │
│         │  ★ Phase 2: 프롬프트 기반 동적 메모리 검색          │   │
│         ▼                                                    │   │
│  ┌─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐       │   │
│  │          AGENTIC LOOP (반복 실행)                  │       │   │
│  │                                                    │       │   │
│  │  ┌─────────────┐                                   │       │   │
│  │  │ PreToolUse  │ ◄──────────────────────┐          │       │   │
│  │  └──────┬──────┘                        │          │       │   │
│  │         ▼                               │          │       │   │
│  │  ┌───────────────────┐                  │          │       │   │
│  │  │PermissionRequest  │                  │          │       │   │
│  │  └──────┬────────────┘                  │          │       │   │
│  │         ▼                               │          │       │   │
│  │  ┌───────────────┐                      │          │       │   │
│  │  │ [tool 실행]   │ (파란색)             │          │       │   │
│  │  └──────┬────────┘                      │          │       │   │
│  │         ▼                               │          │       │   │
│  │  ┌──────────────────────────┐           │          │       │   │
│  │  │PostToolUse/PostToolFail  │           │          │       │   │
│  │  └──────┬───────────────────┘           │          │       │   │
│  │         │  ★ Phase 2: 파일 수정 이력 기록│          │       │   │
│  │         ▼                               │          │       │   │
│  │  ┌──────────────────────────┐           │          │       │   │
│  │  │SubagentStart/SubagentStop│ ──────────┘          │       │   │
│  │  └──────────────────────────┘  (루프 복귀)         │       │   │
│  │         ★ Phase 3: 서브에이전트 결과 학습           │       │   │
│  └─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘       │   │
│         │                                                    │   │
│         ▼                                                    │   │
│  ┌──────────┐                                                │   │
│  │   Stop   │ (분홍색)                                       │   │
│  └──────┬───┘                                                │   │
│         │  ★ Phase 2: 태스크 완료 기록                        │   │
│         ▼                                                    │   │
│  ┌──────────────┐                                            │   │
│  │  PreCompact  │ ───────────────────────────────────────────┘   │
│  └──────┬───────┘                                                │
│         │  ★ Phase 1: Pre-compaction Flush (메모리 DB 저장)      │
│         ▼                                                        │
│  ┌──────────────┐                                                │
│  │  SessionEnd  │                                                │
│  └──────────────┘                                                │
│         ★ Phase 1: 세션 요약 생성 + 메모리 DB 최종 저장           │
│                                                                  │
│  ┌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┐                                            │
│  ╎ Notification    ╎  (비동기, 독립 실행 — 메모리 시스템 무관)    │
│  └╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┘                                            │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 Lifecycle 핵심 관찰 (메모리 설계 근거)

**1. PreCompact → SessionStart 순환 (점선 화살표)**

PreCompact 이후 컨텍스트가 압축되면, SessionStart(source="compact")가 다시 실행된다.
이것은 메모리 시스템의 **핵심 순환 경로**이다:

```
PreCompact 발동 → 메모리 DB에 중요 내용 저장 → 컨텍스트 압축
    → SessionStart(source="compact") → 메모리 DB에서 로드 → additionalContext 주입
```

따라서 PreCompact에서 반드시 현재 대화의 핵심 내용을 DB에 저장해야,
압축 후 SessionStart에서 다시 로드하여 컨텍스트 연속성을 유지할 수 있다.

**2. Agentic Loop 내부 순환**

SubagentStart/SubagentStop은 PreToolUse로 다시 돌아간다 (루프 내부 순환).
이 루프 안에서 PostToolUse는 **매 도구 실행마다** 반복 호출되므로,
Phase 2에서 파일 수정 이력 기록 시 **경량 처리**가 필수적이다 (DB 쓰기는 배치 처리).

**3. Stop은 루프 밖, PreCompact/SessionEnd 직전**

Stop → PreCompact → SessionEnd 순서이므로:
- Stop: 태스크 완료 여부 기록 (Phase 2)
- PreCompact: 메모리 플러시 (Phase 1 필수)
- SessionEnd: 세션 최종 요약 (Phase 1 필수)

**4. Notification은 비동기 독립**

메모리 시스템에서 Notification Hook은 사용하지 않는다.

### 4.3 Hook 분류 (메모리 관점)

#### Phase 1: 핵심 Hook (MVP 필수) — 3개

| Hook | 위치 | 메모리 역할 | 동작 |
|------|------|------------|------|
| **SessionStart** | 최상단 진입점 | 메모리 로드 | DB에서 검색 → `additionalContext` 주입 |
| **PreCompact** | Stop 이후, SessionEnd 이전 | 컨텍스트 보존 | transcript 파싱 → 중요 내용 DB 저장 |
| **SessionEnd** | 최하단 종료점 | 세션 저장 | 전체 transcript → 세션 요약 DB 저장 |

> **source별 SessionStart 동작 분기**:
> - `startup`: 이전 세션 요약 + 프로젝트 지식 로드
> - `resume`: 이전 세션 요약 로드 (이미 컨텍스트 있을 수 있음)
> - `compact`: PreCompact에서 저장한 내용 재로드 (**핵심 순환 경로**)
> - `clear`: 프로젝트 지식만 로드 (세션 기억은 초기화)

#### Phase 2: 보조 Hook (강화용) — 3개

| Hook | 위치 | 메모리 역할 | 동작 |
|------|------|------------|------|
| **UserPromptSubmit** | SessionStart 직후 | 동적 메모리 검색 | 프롬프트 키워드 → 관련 메모리 검색 → `additionalContext` 주입 |
| **PostToolUse** | Agentic Loop 내부 | 파일 수정 추적 | Edit/Write 매처 → 수정 파일 경로 배치 기록 |
| **Stop** | Agentic Loop 직후 | 태스크 완료 기록 | stop_hook_active 확인 → 작업 완료 시점 기록 |

#### Phase 3: 선택적 Hook (고급 기능) — 2개

| Hook | 위치 | 메모리 역할 | 동작 |
|------|------|------------|------|
| **SubagentStop** | Agentic Loop 내부 | 에이전트 결과 학습 | agent_transcript_path → 서브에이전트 결과 저장 |
| **PostToolUseFailure** | Agentic Loop 내부 | 실패 패턴 학습 | 반복 실패 패턴 기록 → 다음 세션 회피 |

### 4.4 메모리 데이터 흐름 (Lifecycle 기반)

```
═══════════════════════════════════════════════════════════════
 Phase 1 (MVP): 저장-로드 순환
═══════════════════════════════════════════════════════════════

  [세션 A]                          [세션 B]
  ─────────                         ─────────
  ... 대화 진행 ...
        │
  PreCompact ──→ DB 저장             SessionStart(startup)
        │       (transcript 파싱)         │
  SessionEnd ──→ DB 저장             DB 검색 ──→ additionalContext
                (세션 요약)               │
                                    "이전 세션에서 X 작업함"

═══════════════════════════════════════════════════════════════
 Phase 1 (MVP): 컴팩션 내 순환 (동일 세션)
═══════════════════════════════════════════════════════════════

  ... 대화가 길어짐 ...
        │
  PreCompact(auto) ──→ DB 저장 (현재 대화 핵심 내용)
        │
  [컨텍스트 압축 발생]
        │
  SessionStart(compact) ──→ DB에서 방금 저장한 내용 재로드
        │                   → additionalContext로 주입
  ... 대화 계속 (컨텍스트 유지됨) ...

═══════════════════════════════════════════════════════════════
 Phase 2 (강화): 실시간 추적
═══════════════════════════════════════════════════════════════

  UserPromptSubmit ──→ 키워드 추출 → DB 검색 → 관련 기억 주입
        │
  Agentic Loop:
  │  PostToolUse(Edit) ──→ 파일 수정 이력 배치 기록
  │  PostToolUse(Write) ──→ 파일 생성 이력 배치 기록
  │  (... 루프 반복 ...)
        │
  Stop ──→ 태스크 완료 시점 + 수정 파일 목록 DB 저장
```

---

## 5. 구현 전략

### 5.1 기술 스택 결정: Go 네이티브

jikime-adk가 Go 기반이므로, 메모리 시스템도 Go 네이티브로 구현한다.

**이유**:
- jikime-adk CLI와 단일 바이너리로 배포 가능
- 별도 런타임(Bun, Node.js) 불필요
- 기존 Hook 인프라 (`internal/hooks/`) 재활용
- 크로스 플랫폼 배포 용이

**벡터 검색 옵션** (Clawdbot 분석 기반):
- **Phase 1**: SQLite + FTS5 (전문 검색) — Clawdbot의 `chunks_fts` 테이블과 동일 방식
- **Phase 2**: Go 내장 코사인 유사도 (Clawdbot의 sqlite-vec 폴백 경로와 유사)
- **Phase 2+**: 외부 임베딩 API (OpenAI `text-embedding-3-small`) + 로컬 코사인 유사도
- **Go 라이브러리**: `github.com/mattn/go-sqlite3` (CGo) 또는 `modernc.org/sqlite` (Pure Go)

### 5.2 저장소 설계

Clawdbot과 달리 **SQLite DB가 원본 저장소**이다 (마크다운 파일 관리 불필요).
Hook 이벤트가 트리거할 때마다 DB에 직접 저장하므로 파일 감시(Chokidar) 불필요.

```
~/.jikime/memory/
├── memory.db              # SQLite 메인 DB (원본 + 인덱스 통합)
└── sessions/              # 세션별 요약 캐시
    ├── {session_id}.json
    └── ...
```

**SQLite 스키마** (Clawdbot 분석 반영):

```sql
-- 스키마 메타데이터 (Clawdbot의 meta 테이블 차용)
CREATE TABLE meta (
    key TEXT PRIMARY KEY,
    value TEXT
);
-- 저장값: schema_version, embedding_model, embedding_provider, chunk_tokens, chunk_overlap

-- 메모리 항목
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    project_dir TEXT NOT NULL,
    type TEXT NOT NULL,           -- 'session_summary', 'decision', 'learning', 'tool_usage'
    content TEXT NOT NULL,
    content_hash TEXT NOT NULL,   -- SHA256 해시 (Clawdbot의 중복 감지 패턴)
    metadata TEXT,                -- JSON: tags, importance, related_files
    embedding BLOB,              -- 벡터 임베딩 (Phase 2)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    accessed_at DATETIME,
    access_count INTEGER DEFAULT 0
);

-- 메모리 전문 검색 (Clawdbot의 chunks_fts 패턴)
CREATE VIRTUAL TABLE memories_fts USING fts5(
    content,
    id UNINDEXED,
    project_dir UNINDEXED,
    type UNINDEXED
);

-- 임베딩 캐시 (Clawdbot의 embedding_cache 패턴 — Phase 2)
CREATE TABLE embedding_cache (
    content_hash TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    embedding BLOB NOT NULL,
    dims INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 프로젝트 지식
CREATE TABLE project_knowledge (
    id TEXT PRIMARY KEY,
    project_dir TEXT NOT NULL,
    file_path TEXT,
    knowledge_type TEXT,         -- 'architecture', 'pattern', 'convention', 'decision'
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

-- 세션 이력
CREATE TABLE session_history (
    session_id TEXT PRIMARY KEY,
    project_dir TEXT NOT NULL,
    started_at DATETIME,
    ended_at DATETIME,
    summary TEXT,
    topics TEXT,                  -- JSON array of topics
    files_modified TEXT,          -- JSON array of file paths
    model TEXT
);

-- 인덱스
CREATE INDEX idx_memories_project ON memories(project_dir);
CREATE INDEX idx_memories_type ON memories(type);
CREATE INDEX idx_memories_created ON memories(created_at DESC);
CREATE INDEX idx_memories_hash ON memories(content_hash);
CREATE INDEX idx_knowledge_project ON project_knowledge(project_dir);
CREATE INDEX idx_sessions_project ON session_history(project_dir);
```

### 5.3 Hook 설정 (settings.json)

Claude Code의 `.claude/settings.json`에 등록할 Hook 설정:

**Phase 1 (MVP)**:
```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hook memory-load",
            "timeout": 10
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hook memory-flush",
            "timeout": 15
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hook memory-save",
            "timeout": 15
          }
        ]
      }
    ]
  }
}
```

**Phase 2 (강화)** — 추가 Hook:
```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hook memory-search",
            "timeout": 5
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hook memory-track",
            "timeout": 3
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jikime-adk hook memory-complete",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

> **타임아웃 설계 근거**:
> - SessionStart(10초): DB 읽기 + JSON 직렬화. 너무 길면 세션 시작 지연
> - PreCompact(15초): transcript 파싱 + DB 쓰기. 데이터 손실 방지를 위해 여유 확보
> - PostToolUse(3초): Agentic Loop 내부에서 매번 실행. **경량 필수**
> - Stop(5초): 배치 기록 플러시. 중간 수준

### 5.4 구현 단계

#### Phase 1: MVP (핵심 3 Hook)

**목표**: 기본 세션 간 컨텍스트 연속성 확보

- `SessionStart` Hook → source별 분기 처리 + `additionalContext` 주입
  - `startup`: 이전 세션 요약 + 프로젝트 지식
  - `compact`: PreCompact에서 저장한 내용 재로드 (순환 경로)
  - `resume`: 이전 세션 요약
  - `clear`: 프로젝트 지식만
- `PreCompact` Hook → transcript 파싱 → 핵심 내용 DB 저장 (Pre-compaction Flush)
- `SessionEnd` Hook → 전체 transcript → 세션 요약 + 수정 파일 목록 DB 저장
- SQLite 기반 텍스트 검색 (FTS5)
- `jikime-adk memory` CLI 서브커맨드 (검색, 목록, 삭제)
- `jikime-adk hook` CLI 서브커맨드 (Hook 핸들러 엔트리포인트)

**CLI 인터페이스**:
```bash
# 메모리 관리
jikime-adk memory search "인증 시스템"     # 메모리 검색
jikime-adk memory list --project .         # 프로젝트별 메모리 목록
jikime-adk memory show <id>               # 메모리 상세 보기
jikime-adk memory delete <id>             # 메모리 삭제
jikime-adk memory stats                   # 메모리 통계
jikime-adk memory gc                      # 오래된 메모리 정리

# Hook 핸들러 (Claude Code가 호출)
jikime-adk hook memory-load               # SessionStart → stdin JSON → stdout additionalContext
jikime-adk hook memory-flush              # PreCompact → stdin JSON → DB 저장
jikime-adk hook memory-save               # SessionEnd → stdin JSON → DB 저장
jikime-adk hook memory-search             # UserPromptSubmit → stdin JSON → stdout additionalContext (Phase 2)
jikime-adk hook memory-track              # PostToolUse → stdin JSON → 파일 이력 기록 (Phase 2)
jikime-adk hook memory-complete           # Stop → stdin JSON → 태스크 완료 기록 (Phase 2)
```

#### Phase 2: 강화 (보조 3 Hook)

**목표**: 사용자 의도 추적 + 프로젝트 지식 구축

- `UserPromptSubmit` Hook → 프롬프트 키워드 → DB 검색 → `additionalContext` 동적 주입
- `PostToolUse` Hook → Edit/Write 매처 → 파일 수정 이력 배치 기록 (인메모리 버퍼 → Stop에서 플러시)
- `Stop` Hook → 태스크 완료 기록 + PostToolUse 배치 플러시
- 벡터 임베딩 추가 (하이브리드 검색: 0.7 벡터 + 0.3 텍스트)
- 메모리 중요도 자동 산정 (접근 빈도, 최신성, 관련성)

> **PostToolUse 성능 고려**: Agentic Loop 내부에서 매 도구 실행마다 호출됨.
> DB 직접 쓰기 대신 **인메모리 버퍼 + 파일 캐시**에 누적하고,
> Stop 또는 PreCompact 시점에 배치로 DB에 기록한다.

#### Phase 3: 고급 (선택 Hook + 최적화)

**목표**: 학습 기반 최적화 + 지능형 메모리 관리

- `SubagentStop` Hook → `agent_transcript_path`에서 서브에이전트 결과 학습
- `PostToolUseFailure` Hook → 실패 패턴 학습
- 자동 메모리 압축 (오래된 기억 요약 병합)
- 메모리 관련성 스코어링 (TF-IDF + 시간 감쇠)
- 프로젝트 간 지식 공유 (공통 패턴 추출)

---

## 6. 아키텍처

### 6.1 패키지 구조 (Go)

```
internal/
├── memory/
│   ├── store.go           # 메모리 저장소 인터페이스 + SQLite 구현
│   ├── schema.go          # SQLite 스키마 정의 + 마이그레이션 (← Clawdbot memory-schema.ts)
│   ├── search.go          # 검색 엔진: FTS5 + 벡터 (← Clawdbot manager-search.ts)
│   ├── hybrid.go          # 하이브리드 검색 병합 (← Clawdbot hybrid.ts, Phase 2)
│   ├── extractor.go       # transcript에서 기억 추출
│   ├── injector.go        # additionalContext 생성
│   ├── transcript.go      # JSONL transcript 파서
│   ├── embedding.go       # 임베딩 프로바이더 인터페이스 + 폴백 (Phase 2)
│   ├── gc.go              # 가비지 컬렉션 (오래된 메모리 정리)
│   └── types.go           # 타입 정의
├── hooks/
│   ├── memory_hooks.go    # 메모리 관련 Hook 핸들러 (SessionStart, PreCompact, SessionEnd)
│   └── ...                # 기존 Hook 핸들러
└── cmd/
    └── memory.go          # CLI 서브커맨드
```

> **Clawdbot 대응 관계**: `store.go` ← `manager.ts`, `search.go` ← `manager-search.ts`,
> `schema.go` ← `memory-schema.ts`, `hybrid.go` ← `hybrid.ts`, `transcript.go` ← `session-files.ts`

### 6.2 Hook 핸들러 설계 (Lifecycle 기반)

#### Phase 1: SessionStart — `jikime-adk hook memory-load`

```go
// SessionStart Hook: 메모리 로드 + additionalContext 주입
// Lifecycle 위치: 최상단 진입점 (PreCompact 후 재진입 포함)
func HandleMemoryLoad(input HookInput) (*HookOutput, error) {
    store := memory.NewStore(input.CWD)

    // source별 분기 (lifecycle의 PreCompact → SessionStart 순환 대응)
    switch input.Source {
    case "startup":
        // 새 세션: 이전 세션 요약 + 프로젝트 지식 + 최근 메모리
        lastSession, _ := store.GetLastSession(input.CWD)
        knowledge, _ := store.GetProjectKnowledge(input.CWD)
        memories, _ := store.SearchRecent(input.CWD, 10)
        context := memory.BuildStartupContext(lastSession, knowledge, memories)
        return outputWithContext("SessionStart", context), nil

    case "compact":
        // 컴팩션 후 재진입: PreCompact에서 저장한 내용 재로드 (핵심 순환 경로)
        memories, _ := store.GetBySession(input.SessionID)
        knowledge, _ := store.GetProjectKnowledge(input.CWD)
        context := memory.BuildCompactContext(memories, knowledge)
        return outputWithContext("SessionStart", context), nil

    case "resume":
        // 세션 재개: 해당 세션 요약 로드
        session, _ := store.GetSession(input.SessionID)
        context := memory.BuildResumeContext(session)
        return outputWithContext("SessionStart", context), nil

    case "clear":
        // /clear 후: 프로젝트 지식만 (세션 기억 초기화)
        knowledge, _ := store.GetProjectKnowledge(input.CWD)
        context := memory.BuildClearContext(knowledge)
        return outputWithContext("SessionStart", context), nil
    }

    return &HookOutput{}, nil
}
```

#### Phase 1: PreCompact — `jikime-adk hook memory-flush`

```go
// PreCompact Hook: Pre-compaction Flush (컨텍스트 보존)
// Lifecycle 위치: Stop 이후, SessionEnd 이전 → SessionStart(compact)로 순환
// 핵심: 여기서 저장한 내용이 컴팩션 후 SessionStart(compact)에서 재로드됨
func HandleMemoryFlush(input HookInput) (*HookOutput, error) {
    store := memory.NewStore(input.CWD)

    // 1. transcript 파싱 (JSONL → 구조화된 대화)
    transcript, err := memory.ParseTranscript(input.TranscriptPath)
    if err != nil {
        return nil, err
    }

    // 2. 중요 내용 추출 (결정사항, 학습, 에러 해결, 작업 진행 상황)
    extracted := memory.Extract(transcript, memory.ExtractOptions{
        SessionID:  input.SessionID,
        ProjectDir: input.CWD,
        Trigger:    input.Trigger, // "manual" or "auto"
    })

    // 3. 메모리 DB에 저장 (content_hash로 중복 방지)
    for _, item := range extracted {
        store.SaveIfNew(item) // hash 기반 중복 체크
    }

    // 4. Phase 2: PostToolUse 배치 버퍼 플러시 (있다면)
    // flushToolUsageBuffer(store, input.SessionID)

    return &HookOutput{}, nil // PreCompact는 additionalContext 반환 불필요
}
```

#### Phase 1: SessionEnd — `jikime-adk hook memory-save`

```go
// SessionEnd Hook: 세션 최종 요약 저장
// Lifecycle 위치: 최하단 종료점 (PreCompact 이후)
func HandleMemorySave(input HookInput) (*HookOutput, error) {
    store := memory.NewStore(input.CWD)

    // 1. 전체 transcript 파싱
    transcript, err := memory.ParseTranscript(input.TranscriptPath)
    if err != nil {
        return nil, err
    }

    // 2. 세션 요약 생성 (토픽, 수정 파일, 핵심 결정사항)
    summary := memory.Summarize(transcript)

    // 3. 세션 이력 저장
    store.SaveSession(memory.SessionRecord{
        SessionID:     input.SessionID,
        ProjectDir:    input.CWD,
        Summary:       summary.Text,
        Topics:        summary.Topics,
        FilesModified: summary.Files,
        EndReason:     input.Reason, // "exit", "clear", "logout" 등
        Model:         input.Model,
    })

    return &HookOutput{}, nil // SessionEnd는 additionalContext 반환 불필요
}
```

#### Phase 2: UserPromptSubmit — `jikime-adk hook memory-search`

```go
// UserPromptSubmit Hook: 프롬프트 기반 동적 메모리 검색
// Lifecycle 위치: SessionStart 직후, Agentic Loop 진입 전
func HandleMemorySearch(input HookInput) (*HookOutput, error) {
    store := memory.NewStore(input.CWD)

    // 1. 프롬프트에서 키워드 추출
    keywords := memory.ExtractKeywords(input.Prompt)
    if len(keywords) == 0 {
        return &HookOutput{}, nil // 키워드 없으면 스킵
    }

    // 2. 관련 메모리 검색 (FTS5 또는 하이브리드)
    results, _ := store.Search(memory.SearchQuery{
        ProjectDir: input.CWD,
        Query:      strings.Join(keywords, " "),
        Limit:      5,
        MinScore:   0.35,
    })

    if len(results) == 0 {
        return &HookOutput{}, nil
    }

    // 3. additionalContext로 관련 기억 주입
    context := memory.BuildSearchContext(results)
    return outputWithContext("UserPromptSubmit", context), nil
}
```

#### Phase 2: PostToolUse — `jikime-adk hook memory-track`

```go
// PostToolUse Hook: 파일 수정 이력 추적
// Lifecycle 위치: Agentic Loop 내부 (매 도구 실행마다 호출)
// 주의: 경량 처리 필수 (타임아웃 3초). DB 쓰기 대신 파일 버퍼에 누적
func HandleMemoryTrack(input HookInput) (*HookOutput, error) {
    // Edit/Write 매처로 이미 필터링됨 (settings.json)
    filePath := input.ToolInput.FilePath
    if filePath == "" {
        return &HookOutput{}, nil
    }

    // 파일 캐시에 누적 (DB 쓰기 아님 — Stop/PreCompact에서 플러시)
    bufferFile := filepath.Join(os.TempDir(), "jikime-memory-buffer-"+input.SessionID+".jsonl")
    entry := fmt.Sprintf(`{"tool":"%s","file":"%s","ts":"%s"}`,
        input.ToolName, filePath, time.Now().Format(time.RFC3339))
    appendToFile(bufferFile, entry)

    return &HookOutput{}, nil
}
```

### 6.3 Context 주입 포맷

SessionStart에서 LLM에 주입하는 컨텍스트 형식:

```markdown
## Session Memory

### Last Session
- Date: 2026-01-27
- Summary: SPEC-SERVICE-001 기업용 AI 커뮤니케이션 허브 설계 완료
- Files: src/gateway/server.ts, src/channels/slack/adapter.ts

### Key Decisions
1. Gateway 인증에 API Key 방식 채택 (JWT 대비 운영 단순성)
2. 채널 어댑터 패턴으로 Slack/Teams 통합 (공통 인터페이스)

### Recent Topics
- Docker Compose 배포 설정
- Slack Bolt SDK 통합
- 감사 로그 시스템 설계

### Project Knowledge
- Architecture: Gateway → Channel Adapter → AI Agent 3계층
- Test Framework: Vitest + Playwright
- Build System: TypeScript + Vite
```

---

## 7. jikime-mem 플러그인과의 관계

### 7.1 마이그레이션 전략

기존 `jikime-mem` 플러그인에서 Go 네이티브 메모리 시스템으로 전환:

| 단계 | 작업 | 설명 |
|------|------|------|
| 1 | 병행 운영 | jikime-mem과 Go 네이티브 메모리 동시 운영 |
| 2 | 데이터 마이그레이션 | jikime-mem SQLite → Go 메모리 DB로 데이터 이관 |
| 3 | jikime-mem 폐기 | Chroma 의존성 제거, Go 네이티브만 유지 |

### 7.2 Go 네이티브 장점

- **단일 바이너리**: jikime-adk와 함께 배포, 별도 설치 불필요
- **런타임 독립**: Bun/Node.js 런타임 불필요
- **Chroma 제거**: 별도 벡터 DB 서버 운영 불필요
- **Hook 통합**: 기존 jikime-adk Hook 인프라 재활용
- **성능**: Go 네이티브 성능으로 Hook 실행 지연 최소화

---

## 8. 향후 고려사항

### 8.1 메모리 크기 관리

- `additionalContext` 크기 제한 (추정 ~4K 토큰 이내 권장)
- 중요도 기반 우선순위 정렬로 제한 내 최적 정보 선택
- 오래된 메모리 자동 압축/병합

### 8.2 프라이버시

- 메모리 DB는 로컬 파일시스템에만 저장
- 민감 정보 (API Key, 비밀번호) 자동 필터링
- 프로젝트별 격리 (다른 프로젝트 메모리 접근 불가)

### 8.3 확장 가능성

- **MCP 서버 통합**: 메모리를 MCP 서버로 노출하여 다른 도구에서 접근
- **팀 메모리**: 팀 공유 메모리 DB (선택적, 서버 기반)
- **LLM 기반 요약**: 대화 요약에 LLM API 활용 (고급)

### 8.4 Clawdbot에서 추가 검토할 패턴

| 패턴 | 설명 | 우선순위 | 적용 시기 |
|------|------|---------|----------|
| **Atomic Reindex** | DB 스키마 변경 시 임시 DB로 재인덱싱 후 원자적 스왑 | 중 | Phase 2 (임베딩 모델 변경 시) |
| **Session Transcript Delta** | 트랜스크립트 변경량(바이트/메시지 수)을 추적하여 불필요한 재파싱 방지 | 중 | Phase 2 |
| **Memory Flush Threshold** | `softThresholdTokens: 4000` + `reserveTokensFloor: 20000` 설정 | 높 | Phase 1 (PreCompact 구현 시) |
| **Multi-Agent Memory Isolation** | 에이전트별 독립 메모리 DB (Clawdbot: `{agentId}.sqlite`) | 낮 | Phase 3 |
| **Prompt-Based Hook** | `type: "prompt"`로 LLM(Haiku)에게 Stop 판단 위임 | 낮 | Phase 3 |
| **Plugin Hook Integration** | 플러그인 `hooks.json` + `${CLAUDE_PLUGIN_ROOT}` 환경변수 | 낮 | Phase 3 |

### 8.5 Hook API 활용 시 주의사항 (공식 문서 기반)

1. **타임아웃**: 기본 60초. 메모리 DB 접근이 느려지면 Hook이 실패할 수 있으므로 경량 쿼리 유지
2. **병렬 실행**: 매칭된 모든 Hook이 병렬 실행됨. 메모리 Hook 간 순서 보장 없음
3. **보안 스냅샷**: Hook 설정은 세션 시작 시 스냅샷. 세션 중 수정 시 `/hooks` 메뉴에서 검토 필요
4. **exit code 2**: `SessionStart`, `SessionEnd`, `PreCompact`에서는 블로킹이 아닌 사용자에게 stderr 표시만
5. **stop_hook_active**: `Stop` Hook에서 이 플래그를 확인하지 않으면 무한 루프 위험
6. **transcript_path**: JSONL 형식. 모든 Hook에서 접근 가능하며 대화 이력 파싱의 핵심 입력

---

## 관련 문서

- [Hooks 시스템](./hooks.md) - jikime-adk Hook 구현 상세
- [Commands](./commands.md) - CLI 명령어 레퍼런스
- [Claude Code Hooks 공식 문서](https://code.claude.com/docs/en/hooks) - 공식 API 레퍼런스
- [Clawdbot GitHub](https://github.com/psteinroe/clawdbot) - 참조 메모리 시스템 원본
