# Web Chat + jikime-adk 아키텍처 설계

> 날짜: 2026-03-10
> 목적: Next.js + AI SDK 웹챗에서 jikime-adk(Claude Code CLI)로 작업 요청 → 진행상황 UI 표시 → 결과물 Target API 저장

---

## 요구사항

- 웹챗(Next.js + Vercel AI SDK)으로 작업 요청
- Claude Code CLI (jikime-adk) 가 실제 작업 수행
- 진행상황을 Web UI에서 실시간 확인
- 작업 결과물(산출물)을 특정 API에 저장
- Claude API SDK 미사용 (토큰 비용 과다) → Claude Code CLI 직접 설치 방식
- 여러 개발자가 동시에 작업 → 충돌 방지 필수
- 모든 프로젝트는 GitHub 사용 중

---

## 접근 방식 비교

### Option A: Claude Code SDK + SSE
```
Next.js Web Chat → POST /api/task
    → spawn Claude CLI subprocess
    → SSE 스트리밍 → 브라우저
    → 완료 시 Target API 저장
```
- 실시간성 높음
- 다중 개발자 git push 시 충돌 위험 존재

### Option B: jikime serve harness + GitHub ✅ 채택
```
Web Chat → GitHub Issue 생성 (jikime-todo 라벨)
    → jikime serve polling
    → Claude Code 작업 실행
    → 로그 → /api/v1/state
    → 완료 시 jikime-done 라벨
    → Target API 저장
```
- GitHub이 충돌 중재자 역할
- polling 지연(30~60초) 있지만 안정성 우선
- 감사 로그(audit trail) 완비

### Option C: WebSocket + Job Queue
```
Web Chat → BullMQ + Redis → Worker → Claude CLI
    → WebSocket 실시간 스트림
```
- 인프라 복잡도 높음

---

## 채택 아키텍처: GitHub Issue 기반

### 선택 이유

**충돌 문제 해결:**
```
Headless CLI 방식의 문제:
  개발자 A 작업 중 (Claude CLI 실행)
      ↓
  개발자 B가 git push
      ↓
  A의 Claude가 stale 상태에서 계속 작업
      ↓
  merge conflict 또는 덮어쓰기 💥

GitHub Issue 방식:
  Issue 기준으로 작업 → PR 생성 → 리뷰 후 merge
  → 충돌이 관리 가능한 구조
```

---

## 전체 플로우

```
1. UI에서 프로젝트 생성
   → 데몬이 특정 경로에 폴더 생성
   → GitHub 토큰 입력
   → git init + GitHub repo 생성
   → 라벨 생성: jikime-todo, jikime-done

2. jikime serve 실행 (프로젝트별 독립 포트)
   jikime serve init
   jikime serve --port {port}  ← 프로젝트마다 다른 포트

3. Web Chat에서 Issue 생성
   → GitHub Issue (jikime-todo 라벨)

4. jikime serve가 GitHub polling
   → Issue 감지 → Claude Code 실행
   → 진행 로그 → /api/v1/state 업데이트

5. Next.js Server가 /api/v1/state polling
   → SSE 변환 → 브라우저 진행상황 표시
   (브라우저가 내부 포트 직접 접근 X → 보안/CORS 문제 방지)

6. 작업 완료 (jikime-done 라벨 감지)
   → 결과물 파싱 (PR diff / commit message)
   → 요약 생성
   → Target API POST
```

---

## 프로젝트별 독립 프로세스 구조

프로젝트마다 독립 프로세스 + 독립 포트로 완전 격리:

```
Project A → jikime serve --port 8001 (PID 1234)
Project B → jikime serve --port 8002 (PID 1235)
Project C → jikime serve --port 8003 (PID 1236)
```

→ A 작업이 B에 영향을 줄 수 없음

---

## 데몬 레지스트리 (Next.js 백엔드)

```
DB (or JSON file)
┌─────────────────────────────────────────────────┐
│ project_id │ repo         │ port │ pid    │ status │
│ proj-001   │ owner/repo-a │ 8001 │ 1234   │ running │
│ proj-002   │ owner/repo-b │ 8002 │ 1235   │ running │
│ proj-003   │ owner/repo-c │ 8003 │ 1236   │ running │
└─────────────────────────────────────────────────┘
```

**프로젝트 생성 시:**
1. 사용 가능한 포트 자동 할당
2. `jikime serve --port {port}` 실행
3. PID + 포트 DB 저장

**서버 재시작 시:**
- DB 읽어서 죽은 프로세스 자동 재시작

---

## PM2로 프로세스 관리

```bash
# 프로젝트 생성 시
pm2 start "jikime serve --port 8001" \
  --name "proj-001" \
  --cwd /projects/proj-001

# 목록 확인
pm2 list

# 서버 재시작 후 자동 복구
pm2 save
pm2 startup
```

Next.js에서 `pm2` CLI를 `child_process`로 호출하여 프로세스 관리.

---

## 시스템 구성도

```
┌─────────────────────────────────────────┐
│  Web Browser                            │
│  Next.js Web Chat (Vercel AI SDK)       │
│  - 프로젝트 관리 UI                      │
│  - Web Chat (Issue 생성)                 │
│  - 진행상황 표시 (SSE)                   │
└──────────────┬──────────────────────────┘
               │ HTTP / SSE
               ▼
┌─────────────────────────────────────────┐
│  Next.js Server                         │
│  - 프로젝트 생성 API                     │
│  - GitHub Issue 생성 API                │
│  - /api/v1/state polling → SSE 변환     │
│  - PM2로 jikime serve 관리              │
│  - 완료 감지 → Target API POST          │
└──────┬──────────────────┬───────────────┘
       │                  │
       ▼                  ▼
┌─────────────┐   ┌───────────────────────┐
│  GitHub     │   │  jikime serve         │
│  - Issues   │   │  (프로젝트별 독립 포트) │
│  - Labels   │←→ │  - GitHub polling     │
│  - PRs      │   │  - Claude Code 실행   │
│  - Commits  │   │  - /api/v1/state      │
└─────────────┘   └───────────────────────┘
                           │ 완료
                           ▼
                  ┌───────────────────────┐
                  │  Target API           │
                  │  (결과 요약 저장)      │
                  └───────────────────────┘
```

---

## 미결 사항

- [ ] Target API 형태 확정 (외부 서비스 / 자체 API)
- [ ] 포트 범위 정책 (예: 8001~8999)
- [ ] 프로젝트 폴더 경로 규칙
- [ ] GitHub polling 주기 설정
- [ ] 결과물 요약 포맷 정의
