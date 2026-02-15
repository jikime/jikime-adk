# JikiME-ADK 메모리 시스템

**버전**: 3.0.0
**최종 업데이트**: 2026-01-28
**상태**: 구현 완료 (Implementation Complete)

---

## 1. 개요

JikiME-ADK 메모리 시스템은 Claude Code 세션 간 컨텍스트 연속성을 제공한다.
2-Layer Memory Architecture를 Go 네이티브로 구현하며, 두 가지 저장소를 사용한다:

- **MD 파일**: 사람이 읽을 수 있는 일일 로그 (human-readable logs)
- **memories 테이블**: 검색/임베딩의 단일 소스 (single source of truth for search)

### 핵심 원칙

- **memories = Single Source for Search**: 세션 데이터(user_prompt, assistant_response, tool_usage)는 memories 테이블에 저장되며, 검색과 임베딩의 단일 소스
- **MD = Human-Readable Logs**: `.jikime/memory/*.md` 파일은 시간순 기록용 로그
- **chunks = MD Index**: MD 파일을 청크로 분할한 파생 인덱스
- **Dual-Source Search**: memory_search는 chunks 테이블과 memories 테이블을 **모두** 검색하여 병합
- **Hybrid Scoring**: 0.7 × vector + 0.3 × text (BM25) 가중 스코어링
- **Deferred Embedding**: 세션 중에는 텍스트만 저장, SessionEnd에서 백그라운드로 배치 임베딩

---

## 2. 아키텍처

### 2.1 2-Layer Memory

```
<projectDir>/
└── .jikime/
    └── memory/
        ├── MEMORY.md              # Layer 2: 장기 지식 (큐레이션된 프로젝트 지식)
        ├── 2026-01-28.md          # Layer 1: 일일 로그 (시간순 append)
        ├── 2026-01-27.md
        ├── memory.db              # SQLite (memories + chunks + embedding_cache)
        └── ...
```

**Layer 1: Daily Logs** (`YYYY-MM-DD.md`)
- 시간순(chronological) append-only 기록
- 세션 데이터 (user_prompt, assistant_response, tool_usage): 타임스탬프와 함께 파일 끝에 순서대로 추가
- 구조적 메모리 (decision, learning, error_fix): `## Section` 헤딩 아래 그룹핑
- Hook과 MCP tool에서 자동 기록

**Layer 2: MEMORY.md**
- 큐레이션된 영구 지식
- 프로젝트 아키텍처, 컨벤션, 주요 결정사항
- 사용자 또는 Claude가 수동 편집

### 2.2 데이터 흐름

```
쓰기 (Write)                              읽기 (Read)
──────────────                            ──────────────
UserPromptSubmit hook ──┐
  (memory_prompt_save)  │                 memory_search MCP tool
                        ├─→ memories DB       │
Stop hook ──────────────┤    + Daily MD   ┌───┴───┐
  (memory_complete)     │                 │       │
                        │                 ▼       ▼
memory_save MCP tool ───┘            chunks    memories
  (decision/learning)    │           (MD index) (session data)
                         ▼               │       │
                   Daily Log MD          └───┬───┘
                   (.jikime/memory/          ▼
                    YYYY-MM-DD.md)     Score 병합 → 상위 N건
                         │
                    File Watcher
                    → Indexer → chunks DB

SessionEnd hook ────→ embed-backfill (백그라운드 프로세스)
  (memory_save)          → memories 테이블 배치 임베딩
```

### 2.3 검색 흐름 (Dual-Source Merged Search)

```
memory_search(query)
    │
    ├─ 1) chunks 테이블 검색 (SearchChunks)
    │     ├─ EmbedAndCache(query) → queryVec
    │     ├─ searchChunksVector(queryVec) — Cosine Similarity brute-force
    │     ├─ searchChunksFTS(query) — FTS5 MATCH (BM25)
    │     └─ mergeChunkResults() — 0.7×vec + 0.3×text
    │
    ├─ 2) memories 테이블 검색 (SearchHybrid)
    │     ├─ searchVector(queryVec) — memories.embedding Cosine Similarity
    │     ├─ searchText(query) — memories FTS5/LIKE
    │     └─ mergeResults() — 0.7×vec + 0.3×text
    │
    └─ 3) 양쪽 결과 통합
          allResults = chunks결과 + memories결과
          → Score 내림차순 정렬
          → 상위 N건 반환
```

---

## 3. MCP Tools

MCP 서버(`jikime-adk mcp serve`)가 제공하는 6개 도구:

| 도구 | 설명 |
|------|------|
| `memory_search` | Dual-source 하이브리드 검색. chunks + memories 양쪽 모두 검색 후 병합 |
| `memory_get` | MD 원본 파일에서 특정 라인 범위 읽기 |
| `memory_load` | MEMORY.md + 오늘의 Daily Log 온디맨드 로드 |
| `memory_save` | Daily Log MD에 구조적 메모리 저장 + 자동 인덱싱 |
| `memory_stats` | 메모리 DB 통계 (메모리 수, 청크 수, 파일 수, DB 크기) |
| `memory_reindex` | 전체 MD 파일 재인덱싱 |

### 3.1 memory_search

chunks 테이블(MD 인덱스)과 memories 테이블(세션 데이터)을 **모두 검색**하여 스코어 기준으로 병합 반환한다.

```json
// Input
{
  "query": "인증 시스템 JWT",
  "maxResults": 6,
  "minScore": 0.35,
  "type": "decision"
}

// Output
{
  "results": [
    {
      "path": ".jikime/memory/2026-01-27.md",
      "start_line": 15,
      "end_line": 22,
      "heading": "## Decision",
      "snippet": "JWT 토큰 캐싱으로 인증 지연 해소 결정...",
      "score": 0.82,
      "source": "chunks"
    },
    {
      "snippet": "API 인증에 JWT 방식 채택",
      "score": 0.75,
      "source": "memory"
    }
  ],
  "count": 2,
  "provider": "openai",
  "model": "text-embedding-3-small"
}
```

**Note**:
- `snippet`은 최대 200자 미리보기이다. 전체 내용이 필요하면 `memory_get`을 사용한다.
- 결과가 0건일 때 `results`는 빈 배열(`[]`)로 반환된다 (`null` 아님).

### 3.2 memory_get

`memory_search` 결과의 `path`, `start_line`, `end_line`을 사용하여 원본 MD에서 상세 읽기:

```json
// Input
{
  "path": ".jikime/memory/2026-01-27.md",
  "from": 15,
  "lines": 8
}

// Output
{
  "path": ".jikime/memory/2026-01-27.md",
  "start_line": 15,
  "end_line": 22,
  "content": "## Decision\n\n- JWT 토큰 캐싱 결정...\n..."
}
```

### 3.3 memory_load

세션 시작 시 또는 프로젝트 컨텍스트가 필요할 때 호출:

```json
// Input
{ "source": "full" }    // "startup" = MEMORY.md만, "full" = MEMORY.md + 오늘의 Daily Log

// Output
{
  "content": "# Project Knowledge\n...\n---\n\n# 2026-01-28\n...",
  "files": [".jikime/memory/MEMORY.md", ".jikime/memory/2026-01-28.md"]
}
```

### 3.4 memory_save

구조적 메모리(decision, learning, error_fix, tool_usage)를 Daily Log MD에 저장하고 자동 인덱싱:

```json
// Input
{
  "type": "decision",
  "content": "API 인증에 JWT 방식 채택",
  "metadata": "{\"context\": \"SPEC-001\"}"
}

// Output
{
  "id": ".jikime/memory/2026-01-28.md",
  "saved": true,
  "message": "memory saved to .jikime/memory/2026-01-28.md"
}
```

---

## 4. Hooks

Claude Code의 Hook 이벤트에 연결되어 자동으로 메모리를 수집/보존한다.

### 4.1 Hook 매핑

| Hook 이벤트 | 명령 | 역할 |
|-------------|------|------|
| **UserPromptSubmit** | `jikime-adk hooks memory-prompt-save` | 사용자 프롬프트를 Daily Log + memories 테이블에 저장 |
| **PostToolUse** (Edit\|Write) | `jikime-adk hooks memory-track` | 파일 수정 기록을 track_buffer에 추가 |
| **Stop** | `jikime-adk hooks memory-complete` | assistant_response + tool_usage를 Daily Log + memories 테이블에 저장 |
| **PreCompact** | `jikime-adk hooks memory-flush` | 컨텍스트 압축 전 중요 정보를 Daily Log에 보존 |
| **SessionEnd** | `jikime-adk hooks memory-save` | 백그라운드 embed-backfill 프로세스 실행 |

### 4.2 세션 중 데이터 저장 흐름

```
사용자 프롬프트 입력
    │
    ▼
UserPromptSubmit hook (memory-prompt-save)
    ├─ Daily Log MD에 시간순 추가: "- [HH:MM:SS] **User Prompt**: ..."
    └─ memories 테이블에 저장 (type: user_prompt, 임베딩 없이 텍스트만)

    ... Claude 작업 중 ...

PostToolUse hook (memory-track) × N회
    └─ track_buffer.jsonl에 파일명 append (Edit|Write 시마다)

Claude 응답 완료
    │
    ▼
Stop hook (memory-complete)
    ├─ transcript에서 마지막 assistant 메시지 추출
    ├─ Daily Log MD에 시간순 추가: "- [HH:MM:SS] **Assistant Response**: ..."
    ├─ memories 테이블에 저장 (type: assistant_response)
    ├─ track_buffer 플러시 → 수정 파일 목록 추출
    ├─ Daily Log MD에 시간순 추가: "- [HH:MM:SS] **Tool Usage**: Files modified: ..."
    └─ memories 테이블에 저장 (type: tool_usage)
```

### 4.3 세션 종료 (SessionEnd)

세션 종료 시 **요약 생성/저장 없이** 백그라운드 임베딩만 트리거한다.
(텍스트 데이터는 UserPromptSubmit과 Stop 훅에서 이미 저장됨)

```
Claude Code: /exit 또는 세션 종료
    │
    ▼
hooks memory-save (SessionEnd hook)
    │
    ├─ stdin에서 session_id, cwd 읽기
    │
    └─ spawnEmbedBackfill()
         ├─ os.Executable() → 현재 바이너리 경로
         ├─ exec.Command(exe, "hooks", "embed-backfill",
         │     "--project-dir", projectDir, "--session-id", sessionID)
         ├─ cmd.Start()
         └─ cmd.Process.Release() → 디태치 (부모 프로세스 종료해도 계속 실행)
              │
              ▼
         embed-backfill (백그라운드 프로세스, 30초 타임아웃)
              ├─ Store 열기
              ├─ Embedding Provider 초기화
              └─ BackfillMemoryEmbeddings(ctx, provider, projectDir, sessionID)
                   → memories 테이블에서 embedding IS NULL인 레코드 조회
                   → 배치 임베딩 생성 + 저장
```

**왜 백그라운드인가**: settings.json의 SessionEnd hook 타임아웃이 5초이다.
임베딩 API 호출이 5초를 초과하면 프로세스가 kill되어 임베딩이 유실된다.
디태치된 백그라운드 프로세스는 Claude Code 종료 후에도 30초까지 계속 실행된다.

### 4.4 컨텍스트 압축 (PreCompact)

```
Claude Code: 컨텍스트 한계 도달 (75%+)
    │
    ▼
hooks memory-flush (PreCompact hook)
    ├─ 1) Transcript JSONL 파싱
    ├─ 2) 중요 정보 추출 (decisions, learnings, error_fixes)
    ├─ 3) Daily Log MD에 저장
    ├─ 4) SQLite 인덱싱 (임베딩 포함)
    └─ 5) Track buffer 플러시 (안전망)
```

---

## 5. Daily Log 형식

`.jikime/memory/YYYY-MM-DD.md`:

세션 데이터는 **시간순(chronological)**으로 파일 끝에 추가되며, 구조적 메모리는 **섹션별 그룹핑**으로 저장된다.

```markdown
# 2026-01-28

- [10:30:15] **User Prompt**: fsnotify 기반 watcher 구현해줘
- [10:30:45] **Assistant Response**: watcher.go 파일을 생성하겠습니다...
- [10:31:02] **Tool Usage**: Files modified: internal/memory/watcher.go
- [10:35:20] **User Prompt**: 빌드 확인해줘
- [10:35:35] **Assistant Response**: 빌드 성공했습니다

## Decision

- JWT 토큰 캐싱으로 인증 지연 해소 결정. Redis TTL 5분 설정.

## Learning

- Go 1.16+ embed 패턴으로 정적 파일 번들링 가능.

## Error Fix

- auth middleware에서 user가 nil일 때 패닉 발생. nil 체크 추가로 해결.
```

**저장 규칙**:
- `user_prompt`, `assistant_response`, `tool_usage` → 시간순, `- [HH:MM:SS] **Type**: content`
- `decision`, `learning`, `error_fix`, `session_summary` → `## Section` 아래 `- content`

---

## 6. 인덱싱 파이프라인

### 6.1 청킹

MD 파일을 검색 가능한 청크로 분할:

| 설정 | 값 |
|------|-----|
| Max Tokens | ~400 tokens/chunk |
| Overlap | ~80 tokens |
| Min Chunk Size | 50 bytes |
| 분할 방식 | Heading-aware (마크다운 헤딩 경계 우선) |

### 6.2 임베딩

| 프로바이더 | 모델 | 차원 | 환경 변수 |
|-----------|------|------|----------|
| OpenAI | `text-embedding-3-small` | 1536 | `OPENAI_API_KEY` |
| Gemini | `text-embedding-004` | 768 | `GEMINI_API_KEY` |
| (없음) | — | — | 텍스트 전용 검색 폴백 |

**Auto-detect 순서**: `JIKIME_EMBEDDING_PROVIDER` 환경 변수가 비어있으면 `auto`:
1. `OPENAI_API_KEY` → OpenAI
2. `GEMINI_API_KEY` → Gemini
3. 둘 다 없음 → `nil` (FTS5 텍스트 검색만 사용)

**캐싱**: `embedding_cache` 테이블에 `(content_hash, provider, model)` 기반 캐시.
동일 텍스트는 재임베딩하지 않음.

**Deferred Embedding**: 세션 중에는 텍스트만 memories 테이블에 저장 (임베딩 없음).
SessionEnd 시 백그라운드 프로세스(`embed-backfill`)가 해당 세션의 미임베딩 레코드를 배치 처리.

### 6.3 File Watcher (자동 인덱싱)

MCP 서버 실행 중 `.jikime/memory/` 디렉토리의 MD 파일 변경을 감지하여 자동 인덱싱:

- **라이브러리**: `github.com/fsnotify/fsnotify`
- **이벤트**: Create, Write (`.md` 파일만)
- **Debounce**: 500ms (빠른 연속 쓰기 시 마지막 1회만 인덱싱)
- **동작**: `IndexFile()` 호출 (청킹 + 임베딩 + SQLite 저장)

### 6.4 SQLite 스키마

```sql
-- 메모리 (세션 데이터의 단일 소스)
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    project_dir TEXT NOT NULL,
    type TEXT NOT NULL,           -- user_prompt, assistant_response, tool_usage, decision, ...
    content TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    metadata TEXT,
    embedding BLOB,               -- Deferred: SessionEnd에서 배치 생성
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    accessed_at DATETIME,
    access_count INTEGER DEFAULT 0
);

-- 청크 인덱스 (MD 파일에서 파생)
CREATE TABLE chunks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL,           -- MD 파일 상대 경로
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    text TEXT NOT NULL,
    hash TEXT NOT NULL,           -- SHA256 (변경 감지)
    heading TEXT DEFAULT '',      -- 마크다운 헤딩
    embedding BLOB               -- float32 little-endian
);

-- 파일 추적 (인덱싱 상태)
CREATE TABLE files (
    path TEXT PRIMARY KEY,
    hash TEXT NOT NULL,
    mtime INTEGER NOT NULL,
    size INTEGER NOT NULL,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 임베딩 캐시
CREATE TABLE embedding_cache (
    content_hash TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    embedding BLOB NOT NULL,
    dims INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_hash, provider, model)
);
```

### 6.5 벡터 검색 방식

**Brute-force Cosine Similarity** (Go 네이티브):

```
SELECT id, ..., embedding FROM chunks/memories
  WHERE embedding IS NOT NULL LIMIT 1000
→ DecodeEmbedding(BLOB → []float32)
→ CosineSimilarity(queryVec, rowVec) for each row
→ score > 0 인 것만 수집
```

**sqlite-vec 미사용 이유**: `modernc.org/sqlite`(pure Go)는 C 확장 로드 불가.
현재 메모리 규모(~1000 chunks)에서 brute-force로 충분.

---

## 7. Claude Code 연동

### 7.1 CLAUDE.md Memory-First Reasoning

CLAUDE.md에 Section 14로 추가된 Memory-First Reasoning 지시:

```
모든 사용자 메시지에 대해:
  1. 현재 세션 컨텍스트만으로 충분히 답할 수 있는가?
     → YES: 직접 응답
     → NO 또는 불확실: memory_search 먼저 호출
  2. 새 세션의 첫 메시지인가?
     → YES: memory_load(source: "full") 먼저 호출
```

키워드 기반이 아닌 **추론 기반** 판단. 과거 컨텍스트가 도움이 될 수 있다면 항상 검색.
불필요한 검색(false positive)은 허용, 검색 누락(false negative)은 불허.

### 7.2 MCP 서버 설정

`.mcp.json` (프로젝트 루트):

```json
{
  "mcpServers": {
    "jikime-memory": {
      "command": "jikime-adk",
      "args": ["mcp", "serve"],
      "description": "Project memory search and management (hybrid vector + text)"
    }
  }
}
```

### 7.3 임베딩 환경 변수

셸 프로파일(`~/.zshrc` 등)에 설정:

```bash
# OpenAI (권장)
export OPENAI_API_KEY="sk-..."

# 또는 Gemini
export GEMINI_API_KEY="AIza..."

# 선택: 프로바이더 직접 지정
export JIKIME_EMBEDDING_PROVIDER="openai"   # openai, gemini, auto, none
export JIKIME_EMBEDDING_MODEL="text-embedding-3-small"
export JIKIME_EMBEDDING_BASE_URL=""
```

---

## 8. CLI 명령어

```bash
# 메모리 검색
jikime-adk memory search "인증 시스템"

# 메모리 목록
jikime-adk memory list --project .

# 메모리 상세 보기
jikime-adk memory show <id>

# 메모리 삭제
jikime-adk memory delete <id>

# 메모리 통계
jikime-adk memory stats

# 오래된 메모리 정리
jikime-adk memory gc

# MD 파일 인덱싱
jikime-adk memory index              # 전체 인덱싱
jikime-adk memory index --file path  # 단일 파일

# 백그라운드 임베딩 (내부 사용, SessionEnd에서 자동 호출)
jikime-adk hooks embed-backfill --project-dir . --session-id abc123
```

---

## 9. Claude Code 내장 기능 (구현 불필요)

다음 기능은 Claude Code 런타임이 자체 처리하며, jikime-adk에서 구현하지 않는다:

| 기능 | 설명 |
|------|------|
| **Compaction** | 컨텍스트 윈도우 한계 시 오래된 대화 요약/압축 |
| **Pruning** | 오래된 tool 결과(exec 출력 등) 인메모리 축소 |
| **Cache-TTL Pruning** | 5분 캐시 만료 후 재캐싱 비용 최적화를 위한 tool 결과 정리 |

jikime-adk의 역할은 compaction **전에** 중요 정보를 MD 파일에 저장하는 것이다 (`memory-flush` hook).

---

## 10. 구현 파일 구조

```
internal/memory/
├── store.go            # SQLiteStore (DB 연결, CRUD)
├── schema.go           # 테이블 생성, 마이그레이션
├── search.go           # FTS5/LIKE 텍스트 검색
├── hybrid.go           # 하이브리드 검색 병합 (memories 테이블)
├── chunk_search.go     # 청크 하이브리드 검색 (chunks 테이블)
├── chunker.go          # MD → 청크 분할 (heading-aware)
├── indexer.go          # 인덱싱 파이프라인 (청킹 + 임베딩 + 저장)
├── watcher.go          # fsnotify 파일 감시 (자동 인덱싱)
├── embedding.go        # 임베딩 프로바이더 (OpenAI, Gemini)
├── embedding_cache.go  # 임베딩 캐시 (hash 기반) + BackfillMemoryEmbeddings
├── mdwriter.go         # Daily Log MD 파일 읽기/쓰기 (시간순 + 섹션별)
├── extractor.go        # Transcript에서 메모리 추출
├── injector.go         # additionalContext 생성
├── transcript.go       # JSONL transcript 파서
├── gc.go               # 가비지 컬렉션
├── track.go            # PostToolUse 파일 수정 추적 버퍼
└── types.go            # 타입 정의, 상수

cmd/hookscmd/
├── memory_prompt_save.go  # UserPromptSubmit hook (user_prompt 저장)
├── memory_complete.go     # Stop hook (assistant_response + tool_usage 저장)
├── memory_save.go         # SessionEnd hook (백그라운드 embed-backfill 실행)
├── memory_flush.go        # PreCompact hook
├── memory_track.go        # PostToolUse hook
├── memory_load.go         # SessionStart hook (no-op)
├── embed_backfill.go      # 백그라운드 임베딩 서브커맨드 (hidden)
└── hooks.go               # Hook 명령 등록

cmd/mcpcmd/
├── mcp.go              # MCP 명령 그룹
└── serve.go            # MCP 서버 (6개 tool + file watcher)

cmd/memorycmd/
├── memory.go           # CLI 명령 그룹
├── search.go           # memory search
├── list.go             # memory list
├── show.go             # memory show
├── delete.go           # memory delete
├── stats.go            # memory stats
├── gc.go               # memory gc
├── index.go            # memory index
└── migrate.go          # memory migrate (v2→v3)
```

---

## 11. 설계 결정 및 제약사항

### 11.1 주요 설계 결정

| 결정 | 이유 |
|------|------|
| **memories를 검색 단일 소스로** | MD 파일은 사람용 로그, 검색은 memories 테이블에서 수행 |
| **Dual-source merged search** | chunks(MD 인덱스)와 memories(세션 데이터) 모두 검색하여 누락 방지 |
| **Deferred embedding** | 세션 중 API 호출 없이 텍스트만 저장, SessionEnd에서 배치 처리 |
| **백그라운드 embed-backfill** | SessionEnd hook의 5초 타임아웃 제약 우회 |
| **시간순 Daily Log** | user_prompt/assistant_response를 대화 흐름대로 기록 |
| **Memory-First Reasoning** | 키워드 트리거가 아닌 추론 기반 메모리 사용 판단 |
| **validateEnvironment 제거** | 세션 시작 성능 최적화 (불필요한 exec.LookPath 제거) |

### 11.2 sqlite-vec 미구현

`modernc.org/sqlite`(pure Go)는 C 확장을 로드할 수 없어 sqlite-vec 사용 불가.
`mattn/go-sqlite3`(CGo)로 변경하면 가능하나, 크로스플랫폼 이식성이 깨진다.
현재 메모리 규모(~1000 chunks + memories)에서 brute-force 검색으로 충분.

### 11.3 데이터 용량 추정

| 기간 | 청크 수 | 임베딩 크기 | DB 크기 |
|------|---------|------------|---------|
| 1개월 | ~150-450 | ~1-3MB | ~3-8MB |
| 6개월 | ~900-2,700 | ~5-16MB | ~18-50MB |
| 1년 | ~1,800-5,400 | ~11-32MB | ~35-100MB |

---

## 관련 문서

- [Hooks 시스템](./hooks.md) - Hook 구현 상세
- [세션 메모리 흐름](./session-memory-flow.md) - 전체 데이터 흐름
- [Commands](./commands.md) - CLI 명령어 레퍼런스
