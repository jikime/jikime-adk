# 세션 메모리 전체 Flow

**버전**: 3.0.0
**최종 업데이트**: 2026-01-28
**상태**: 구현 완료

---

## 1. 세션 시작 (SessionStart)

```
Claude Code 시작
    │
    ▼
hooks session-start
    │  → findProjectRoot() → .jikime 디렉토리 탐색
    │  → loadConfig() → YAML 설정 4개 로드 (user, language, project, git-strategy)
    │  → orchestrator 상태 초기화 (state 파일 없을 때만)
    │  → getGitInfoParallel() → git branch/status/log 3개 병렬 실행
    │  → formatSessionOutput() → 세션 정보 포맷팅
    │
    ▼
hooks memory-load (no-op)
    │  → SessionStart hook은 빈 출력 반환
    │  → 컨텍스트 로딩은 memory_load MCP tool로 이관됨
    │
    ▼
Claude가 memory_load MCP tool 호출 (CLAUDE.md Memory-First 규칙)
    │
    ├─ source="startup" → MEMORY.md 읽기
    ├─ source="full"    → MEMORY.md + 오늘의 Daily Log 읽기
    │
    ▼
Claude에 프로젝트 컨텍스트 전달
```

---

## 2. 사용자 프롬프트 저장 (UserPromptSubmit)

```
사용자가 프롬프트 입력
    │
    ▼
hooks memory-prompt-save (UserPromptSubmit hook)
    │
    ├─ 1) stdin에서 프롬프트 텍스트 읽기
    │
    ├─ 2) Daily Log MD에 시간순 추가 (AppendDailyLog)
    │     .jikime/memory/YYYY-MM-DD.md
    │     → "- [HH:MM:SS] **User Prompt**: 프롬프트 내용"
    │
    ├─ 3) memories 테이블에 저장 (SaveIfNew)
    │     type: "user_prompt"
    │     content: 프롬프트 텍스트
    │     embedding: NULL (Deferred — SessionEnd에서 배치 생성)
    │
    └─ 4) JSON 응답 출력 (stdout)
```

---

## 3. 메모리 검색 (memory_search — Dual-Source Merged)

```
Claude: memory_search MCP tool 호출
    │  (Memory-First Reasoning: 과거 컨텍스트가 필요하다고 판단)
    │
    ▼
handleMemorySearch()
    │
    ├─ 1) chunks 테이블 검색 (SearchChunks)
    │     ├─ EmbedAndCache(query) → queryVec
    │     │     ├─ 캐시 히트? → 즉시 반환
    │     │     └─ 캐시 미스? → API 호출 → 캐시 저장
    │     │
    │     ├─ searchChunksVector(queryVec)
    │     │     SELECT embedding FROM chunks
    │     │       WHERE embedding IS NOT NULL LIMIT 1000
    │     │     → Go CosineSimilarity() brute-force
    │     │
    │     ├─ searchChunksFTS(query)
    │     │     FTS5 MATCH 또는 LIKE 폴백
    │     │     → BM25 scoring
    │     │
    │     └─ mergeChunkResults()
    │           finalScore = 0.7 × vecScore + 0.3 × textScore
    │
    ├─ 2) memories 테이블 검색 (SearchHybrid)
    │     ├─ searchVector(queryVec) — memories.embedding Cosine Similarity
    │     ├─ searchText(query) — memories FTS5/LIKE
    │     └─ mergeResults() — 0.7×vec + 0.3×text
    │
    └─ 3) 양쪽 결과 통합
          allResults = make([]memorySearchResult, 0)  ← nil이 아닌 빈 배열
          allResults = chunks결과 + memories결과
          → Score 내림차순 sort.Slice()
          → 상위 N건 반환 (기본 6건)

    │
    ▼
결과: [{path, start_line, end_line, snippet, score, source:"chunks"|"memory"}, ...]
    │  snippet = 최대 200자 미리보기 (토큰 절약)
    │
    ▼
(필요 시) memory_get MCP tool 호출 — snippet 확인 후 원본 MD에서 상세 읽기
    ├─ path: ".jikime/memory/2026-01-28.md"
    ├─ from: start_line
    ├─ lines: end_line - start_line
    └─ → 원본 MD 파일에서 해당 라인 범위 반환
```

---

## 4. 도구 사용 추적 (PostToolUse)

```
Claude: Edit("src/auth/jwt.go", ...)
    │
    ▼
hooks memory-track (PostToolUse, matcher: Edit|Write)
    │
    ├─ tool_name == "Edit" or "Write"?
    │     ├─ YES → JSONL 버퍼에 append
    │     │        .jikime/memory/track_buffer.jsonl:
    │     │        {"session_id":"abc","file":"src/auth/jwt.go","tool":"Edit","ts":"..."}
    │     │
    │     └─ NO → 즉시 반환
    │
    ▼
suppressOutput: true (DB 접근 없이 파일 append만 — 3초 내 완료)
```

---

## 5. 응답 완료 (Stop)

```
Claude 응답 완료
    │
    ▼
hooks memory-complete (Stop hook)
    │
    ├─ 1) transcript에서 마지막 assistant 메시지 추출
    │
    ├─ 2) Daily Log MD에 시간순 추가 (AppendDailyLog)
    │     → "- [HH:MM:SS] **Assistant Response**: 응답 내용 (truncated)"
    │
    ├─ 3) memories 테이블에 저장 (SaveIfNew)
    │     type: "assistant_response"
    │     embedding: NULL (Deferred)
    │
    ├─ 4) FlushTrack() — track_buffer.jsonl 읽기 + 삭제
    │     [jwt.go, middleware.go, auth_test.go, ...]
    │
    ├─ 5) 파일 중복 제거 (dedupe)
    │     jwt.go, middleware.go, auth_test.go (3건)
    │
    ├─ 6) Daily Log MD에 시간순 추가 (AppendDailyLog)
    │     → "- [HH:MM:SS] **Tool Usage**: Files modified: jwt.go, middleware.go, auth_test.go"
    │
    └─ 7) memories 테이블에 저장 (SaveIfNew)
          type: "tool_usage"
          embedding: NULL (Deferred)
```

---

## 6. 메모리 저장 (memory_save MCP tool)

```
Claude: "이 결정사항을 기억해둬"
    │
    ▼
memory_save MCP tool 호출
    │
    ├─ type: "decision"
    ├─ content: "JWT 토큰 캐싱 결정..."
    ├─ metadata: "{...}" (선택)
    │
    ├─ 1) Daily Log MD에 추가 (AppendDailyLog)
    │     .jikime/memory/YYYY-MM-DD.md
    │     → ## Decision 섹션에 "- content" append
    │
    ├─ 2) SQLite 인덱싱 (IndexFile)
    │     → 청킹 (~400 token)
    │     → 임베딩 생성 (OpenAI/Gemini)
    │     → chunks 테이블에 저장
    │
    └─ 3) File Watcher가 감지 (백업)
          → 500ms debounce 후 재인덱싱
          → save에서 이미 인덱싱했으므로 중복 시 스킵
```

---

## 7. 컨텍스트 압축 (PreCompact)

```
Claude Code: 컨텍스트 한계 도달 (75%+)
    │
    ▼
hooks memory-flush (PreCompact hook)
    │
    ├─ 1) Transcript JSONL 파싱 (ParseTranscript)
    │
    ├─ 2) 중요 콘텐츠 추출 (Extract)
    │     → decision, learning, error_fix 등
    │
    ├─ 3) Daily Log MD에 저장 (AppendDailyLog)
    │     → 각 추출 항목을 타입에 따라:
    │       구조적 메모리 → ## Section에 append
    │       세션 데이터 → 시간순 파일 끝에 append
    │
    ├─ 4) SQLite 인덱싱 (IndexFile + 임베딩)
    │
    ├─ 5) FlushTrack() — track_buffer 안전망
    │
    ▼
Claude Code: Compaction 실행
    → 오래된 대화 요약/압축
    → 중요 정보는 MD + SQLite에 보존됨
```

---

## 8. 세션 종료 (SessionEnd)

세션 종료 시 **요약 생성/저장을 하지 않는다**.
텍스트 데이터는 UserPromptSubmit(memory-prompt-save)과 Stop(memory-complete)에서 이미 저장됨.
SessionEnd에서는 **백그라운드 임베딩만 트리거**한다.

```
Claude Code: /exit 또는 세션 종료
    │
    ▼
hooks memory-save (SessionEnd hook)
    │
    ├─ 1) stdin에서 session_id, cwd 읽기
    │     memorySaveInput { session_id, cwd }
    │
    ├─ 2) projectDir 결정 (cwd 또는 os.Getwd())
    │
    └─ 3) spawnEmbedBackfill(projectDir, sessionID)
         │
         ├─ os.Executable() → 현재 바이너리 경로
         ├─ exec.Command(exe, "hooks", "embed-backfill",
         │     "--project-dir", projectDir,
         │     "--session-id", sessionID)
         ├─ cmd.Stdin = nil
         ├─ cmd.Stdout = nil
         ├─ cmd.Stderr = os.Stderr  ← 디버그 로그
         ├─ cmd.Start()
         └─ cmd.Process.Release()  ← 디태치 (부모 종료해도 계속 실행)
              │
              ▼
         embed-backfill (독립 백그라운드 프로세스)
              │
              ├─ context.WithTimeout(30초)
              ├─ memory.NewStore(projectDir) → SQLite 열기
              ├─ memory.LoadEmbeddingConfig() → 환경변수 읽기
              ├─ memory.NewEmbeddingProvider(cfg)
              │
              └─ store.BackfillMemoryEmbeddings(ctx, provider, projectDir, sessionID)
                   │
                   ├─ SELECT id, content FROM memories
                   │    WHERE embedding IS NULL
                   │      AND session_id = ? AND project_dir = ?
                   │
                   ├─ for each unembedded memory:
                   │     embedding = provider.Embed(content)
                   │     UPDATE memories SET embedding = ? WHERE id = ?
                   │
                   └─ return count, nil
```

**왜 이 구조인가**:
- settings.json의 SessionEnd hook 타임아웃: 5000ms (5초)
- 임베딩 API 1회 호출: ~200-500ms, 세션 중 축적된 memories: ~10-50건
- 총 소요시간: ~2-25초 → 5초 타임아웃 초과 가능
- `cmd.Process.Release()`로 디태치하면 Claude Code 종료 후에도 30초까지 계속 실행

---

## 9. 임베딩 Fallback 체인

```
LoadEmbeddingConfig()
    │
    ├─ JIKIME_EMBEDDING_PROVIDER 환경변수?
    │     ├─ "openai" → OpenAI (text-embedding-3-small, 1536 dims)
    │     ├─ "gemini" → Gemini (text-embedding-004, 768 dims)
    │     ├─ "none"   → 임베딩 비활성화
    │     └─ "auto" 또는 "" → 아래 자동 감지
    │
    ├─ auto 감지:
    │     ├─ OPENAI_API_KEY 있음? → OpenAI 사용
    │     ├─ GEMINI_API_KEY 있음? → Gemini 사용
    │     └─ 둘 다 없음?         → provider=nil
    │
    ▼
provider == nil 일 때:
    ├─ 임베딩 생성 스킵
    ├─ 벡터 검색 스킵
    └─ FTS5 텍스트 검색만 사용 (또는 LIKE 폴백)
```

---

## 10. File Watcher (자동 인덱싱)

```
MCP 서버 시작 (jikime-adk mcp serve)
    │
    ├─ Store 열기
    ├─ Embedding Provider 초기화
    ├─ WatchMemoryFiles() 고루틴 시작
    │     │
    │     ├─ fsnotify.NewWatcher()
    │     ├─ .jikime/memory/ 디렉토리 감시
    │     │
    │     └─ 이벤트 루프:
    │           ├─ Create/Write .md 파일 감지
    │           ├─ 500ms debounce (파일별 timer)
    │           └─ IndexFile() 호출
    │                 ├─ 청킹
    │                 ├─ 임베딩 생성
    │                 └─ SQLite chunks 테이블 저장
    │
    └─ MCP 서버 실행 (STDIO transport)
        └─ 서버 종료 시 cancel() → watcher 정리
```

---

## 11. 타임라인 요약

```
세션 시작 ──→ session-start hook (git info + config + orchestrator)
              memory-load hook (no-op)
              Claude가 memory_load MCP tool 호출 (Memory-First)
    │
사용자 입력 ──→ memory-prompt-save hook (user_prompt → Daily MD + memories DB)
              Claude가 memory_search MCP tool 호출 (필요 시)
              → memory_get으로 원본 MD 상세 읽기
    │
Claude 도구 ──→ memory-track hook (파일 수정 기록, 반복)
    │
Claude 저장 ──→ memory_save MCP tool (구조적 메모리 → Daily MD + chunks 인덱싱)
    │
응답 완료 ──→ memory-complete hook (assistant_response + tool_usage → Daily MD + memories DB)
    │
컨텍스트 압축 ──→ memory-flush hook (트랜스크립트 → Daily MD + 인덱싱)
    │
세션 종료 ──→ memory-save hook (백그라운드 embed-backfill 실행)
              → memories 테이블 배치 임베딩 (30초 타임아웃)
```

---

## 12. DB 스키마 (v3 — 2-Layer Architecture)

```
memories (세션 데이터의 단일 소스 — 검색/임베딩 대상)
├─ id TEXT PK
├─ session_id TEXT              ← 세션 식별자
├─ project_dir TEXT             ← 프로젝트 경로
├─ type TEXT                    ← user_prompt, assistant_response, tool_usage, decision, ...
├─ content TEXT                 ← 메모리 텍스트
├─ content_hash TEXT            ← SHA256 (중복 방지)
├─ metadata TEXT                ← JSON 메타데이터
├─ embedding BLOB               ← Deferred: SessionEnd에서 배치 생성
├─ created_at DATETIME
├─ accessed_at DATETIME
└─ access_count INTEGER

chunks (MD 파일에서 파생된 인덱스)
├─ id INTEGER PK AUTOINCREMENT
├─ path TEXT                    ← MD 파일 상대 경로
├─ start_line INTEGER
├─ end_line INTEGER
├─ text TEXT
├─ hash TEXT                    ← SHA256 (변경 감지)
├─ heading TEXT                 ← 마크다운 헤딩
└─ embedding BLOB               ← float32 little-endian

files (인덱싱 상태 추적)
├─ path TEXT PK
├─ hash TEXT
├─ mtime INTEGER
├─ size INTEGER
└─ indexed_at DATETIME

embedding_cache (임베딩 캐시)
├─ content_hash TEXT  ─┐
├─ provider TEXT       ├─ 복합 PK
├─ model TEXT         ─┘
├─ embedding BLOB
├─ dims INTEGER
└─ created_at DATETIME
```
