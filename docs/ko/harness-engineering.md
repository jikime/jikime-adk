# 하네스 엔지니어링 (Harness Engineering)

> GitHub Issue 생성부터 PR 자동 머지까지 — 완전 자동화된 자율 에이전트 오케스트레이션.

## 개념

**하네스 엔지니어링(Harness Engineering)** 이란 AI 에이전트가 소프트웨어 작업을 자율적으로 수행하도록 *하네스(제어 프레임워크)* 를 구축하는 방법론입니다. 마치 말 마구(harness)가 말의 힘을 정밀한 움직임으로 유도하듯, 하네스 엔지니어링은 Claude의 능력을 구조화되고 안전하며 반복 가능한 워크플로우로 이끌어냅니다.

JiKiME-ADK에서는 **`jikime serve`** 명령어로 하네스 엔지니어링이 구현되어 있습니다. GitHub Issues를 폴링하고, 격리된 워크스페이스를 생성하고, Claude를 헤드리스로 실행하여 이슈 할당부터 PR 머지까지 전체 라이프사이클을 사람 개입 없이 처리하는 장기 실행 데몬입니다.

```
사람이 GitHub Issue를 작성
        ↓
jikime serve가 감지 (15초마다)
        ↓
Claude가 이슈를 읽고, 코드를 작성하고, PR을 생성
        ↓
PR 자동 머지
        ↓
Issue 자동 Close
```

---

## 아키텍처

```
┌─────────────────────────────────────────────────────┐
│                   jikime serve                      │
│                                                     │
│  ┌──────────┐    ┌─────────────┐    ┌───────────┐  │
│  │ Tracker  │───▶│Orchestrator │───▶│  Runner   │  │
│  │ (GitHub) │    │             │    │ (Claude)  │  │
│  └──────────┘    └──────┬──────┘    └───────────┘  │
│                         │                           │
│  ┌──────────┐    ┌──────▼──────┐    ┌───────────┐  │
│  │ HTTP API │    │  Workspace  │    │   Hooks   │  │
│  │  :8888   │    │  Manager    │    │ lifecycle │  │
│  └──────────┘    └─────────────┘    └───────────┘  │
└─────────────────────────────────────────────────────┘
```

| 컴포넌트 | 역할 |
|----------|------|
| **Tracker** | GitHub Issues에서 활성 상태 라벨 (예: `jikime-todo`) 폴링 |
| **Orchestrator** | 상태 머신: 디스패치, 백오프 재시도, 터미널 상태 reconcile |
| **Runner** | `claude --print --output-format stream-json` 헤드리스 실행 |
| **Workspace Manager** | 이슈별 디렉토리 생성 및 라이프사이클 훅 실행 |
| **HTTP API** | `http://127.0.0.1:<port>`에서 실시간 상태 대시보드 제공 |

---

## WORKFLOW.md

`jikime serve`를 사용하는 모든 프로젝트는 하나의 `WORKFLOW.md` 파일로 구성됩니다. YAML 프론트매터 블록과 프롬프트 템플릿으로 구성됩니다.

```yaml
---
tracker:
  kind: github
  project_slug: owner/repo
  # api_key: $GITHUB_TOKEN   # 생략하면 gh auth token 자동 사용
  active_states:
    - jikime-todo             # Claude가 작업할 이슈 상태
  terminal_states:
    - jikime-done             # 사람이 완료 처리
    - Done                    # GitHub 자동 close

polling:
  interval_ms: 15000          # 15초마다 폴링

workspace:
  root: /tmp/my-workspaces    # 이슈별 클론 디렉토리

hooks:
  after_create: |             # 워크스페이스 최초 생성 시 1회 실행
    git clone https://github.com/owner/repo.git .

  before_run: |               # Claude 세션 시작 전마다 실행
    git fetch origin
    git checkout main
    git reset --hard origin/main

  after_run: |                # Claude 세션 종료 후마다 실행
    echo "완료"

  timeout_ms: 60000           # 훅 타임아웃 (60초)

agent:
  max_concurrent_agents: 1    # 동시 Claude 세션 수
  max_turns: 5                # 세션당 최대 멀티턴 횟수
  max_retry_backoff_ms: 60000 # 재시도 최대 대기 시간

claude:
  stall_timeout_ms: 180000    # 3분간 출력 없으면 강제 종료

server:
  port: 8888                  # HTTP 상태 API 포트 (0 = 비활성)
---

당신은 GitHub 이슈를 처리하는 자율 소프트웨어 엔지니어입니다.

## Issue

**{{ issue.identifier }}**: {{ issue.title }}

{{ issue.description }}

## 지시사항

1. 이슈를 꼼꼼히 읽고 요청사항을 구현하세요.
2. 피처 브랜치 생성: `git checkout -b fix/issue-{{ issue.id }}`
3. 변경사항 작성
4. 커밋: `git add -A && git commit -m "fix: {{ issue.identifier }} - {{ issue.title }}"`
5. 푸시: `git push origin fix/issue-{{ issue.id }}`
6. PR 생성: `gh pr create --title "fix: {{ issue.title }}" --body "Closes #{{ issue.id }}" --base main --head fix/issue-{{ issue.id }}`
7. 머지: `gh pr merge --squash --delete-branch --admin`
```

### 템플릿 변수

| 변수 | 예시 |
|------|------|
| `{{ issue.id }}` | `9` |
| `{{ issue.identifier }}` | `owner/repo#9` |
| `{{ issue.title }}` | `푸터 컴포넌트 추가` |
| `{{ issue.description }}` | 이슈 본문 전체 |
| `{{ issue.state }}` | `jikime-todo` |
| `{{ issue.url }}` | `https://github.com/...` |
| `{{ attempt }}` | `2` (재시도 횟수) |

---

## 전체 플로우

```
1. POLL (폴링) ───────────────────────────────────────────────
   jikime serve가 15초마다 GitHub 폴링
   jikime-todo 라벨이 있는 이슈 수집
   우선순위 → created_at → identifier 순으로 정렬

2. DISPATCH (디스패치) ───────────────────────────────────────
   Orchestrator: 실행 중이지 않음, claimed 아님 확인
   이슈를 claimed로 표시
   워커 고루틴 생성

3. WORKSPACE SETUP (워크스페이스 준비) ───────────────────────
   /tmp/workspaces/owner_repo_9/ 디렉토리 생성
   after_create 훅 실행: git clone ...
   (재시도 시: before_run에서 최신 main으로 동기화)

4. CLAUDE 실행 ───────────────────────────────────────────────
   이슈 데이터로 프롬프트 템플릿 렌더링
   실행: claude --print --output-format stream-json \
         --verbose --dangerously-skip-permissions \
         "렌더링된 프롬프트"
   출력 스트리밍 (스톨 감지: 3분 타임아웃)

5. BRANCH + PR + MERGE ──────────────────────────────────────
   Claude: git checkout -b fix/issue-9
   Claude: (코드 작성)
   Claude: git push origin fix/issue-9
   Claude: gh pr create --body "Closes #9"
   Claude: gh pr merge --squash --delete-branch --admin

6. 자동 Close ────────────────────────────────────────────────
   GitHub: PR 머지 → Issue #9 closed (상태: Done)

7. RECONCILE (정리) ──────────────────────────────────────────
   jikime serve: 이슈가 터미널 상태(Done)임을 감지
   after_run 훅 실행
   워크스페이스 삭제
   claim 해제
```

---

## 기능 상세

### 워크스페이스 격리

각 이슈는 `workspace.root` 아래에 고유한 디렉토리를 가집니다:

```
/tmp/my-workspaces/
  owner_repo_7/    ← 이슈 #7
  owner_repo_9/    ← 이슈 #9
  owner_repo_11/   ← 이슈 #11
```

- 최초 생성 시 신선한 `git clone`
- `before_run`에서 Claude 시작 전 항상 최신 `origin/main` 동기화
- 격리됨: 여러 이슈가 동시에 실행되어도 서로 간섭 없음

### 브랜치 전략 및 충돌 방지

| 위험 요소 | 방어 방법 |
|-----------|-----------|
| 여러 에이전트가 같은 브랜치에 푸시 | 이슈마다 전용 `fix/issue-N` 브랜치 |
| 재시도 시 오래된 워크스페이스 | `before_run`: `git reset --hard origin/main` |
| 사람의 푸시와 충돌 | 브랜치 격리 — 사람과 에이전트가 같은 브랜치를 절대 공유하지 않음 |
| 에이전트 두 개가 main에서 경쟁 | `max_concurrent_agents: 1` (기본값) |

### 라이프사이클 훅

| 훅 | 실행 시점 | 주요 용도 |
|----|-----------|-----------|
| `after_create` | 워크스페이스 최초 생성 시 | `git clone` |
| `before_run` | 모든 Claude 세션 시작 전 | `git fetch && git reset --hard origin/main` |
| `after_run` | 모든 Claude 세션 종료 후 | 로컬 저장소 `git pull` 동기화 |
| `before_remove` | 워크스페이스 삭제 전 | 아티팩트 아카이브 |

### 지수 백오프 재시도

실패한 세션은 자동으로 재시도됩니다:

```
시도 1회 → 10초 대기
시도 2회 → 20초 대기
시도 3회 → 40초 대기
시도 4회 → 60초 (max_retry_backoff_ms로 상한 고정)
```

공식: `min(10000 × 2^(시도횟수-1), max_retry_backoff_ms)`

이슈가 터미널 상태로 전환되면 재시도가 취소됩니다.

### 토큰 집계

`--output-format stream-json`을 통해 매 세션의 토큰 사용량이 캡처되고 누적됩니다:

```json
{
  "jikime_totals": {
    "InputTokens": 12840,
    "OutputTokens": 3210,
    "TotalTokens": 16050,
    "SecondsRunning": 342.5
  }
}
```

### HTTP 상태 API

`server.port`가 설정되면 상태 API를 사용할 수 있습니다:

| 엔드포인트 | 메서드 | 설명 |
|------------|--------|------|
| `/` | GET | 사람이 읽기 좋은 텍스트 대시보드 |
| `/api/v1/state` | GET | JSON 상태 스냅샷 |
| `/api/v1/refresh` | POST | 즉시 폴링 트리거 |

```bash
# 라이브 대시보드 (3초마다 갱신)
watch -n 3 'curl -s http://127.0.0.1:8888/'

# JSON 상태 확인
curl -s http://127.0.0.1:8888/api/v1/state | jq .

# 즉시 폴링 트리거 (15초 기다리기 싫을 때)
curl -s -X POST http://127.0.0.1:8888/api/v1/refresh
```

### WORKFLOW.md 핫 리로드

`jikime serve` 실행 중에 `WORKFLOW.md`를 편집하면 다음 틱에 자동으로 반영됩니다. 재시작 불필요. `fsnotify`로 파일 변경을 감시합니다.

---

## 사용 가이드

### 1. 설치

```bash
go install github.com/jikime/jikime-adk@latest
```

### 2. WORKFLOW.md 생성

```bash
# 예제 파일 복사
cp WORKFLOW.md.example my-project/WORKFLOW.md

# 프로젝트에 맞게 편집
vim my-project/WORKFLOW.md
```

### 3. GitHub 라벨 생성

```bash
gh label create "jikime-todo" --repo owner/repo \
  --description "AI 에이전트 작업 대기" --color "0e8a16"

gh label create "jikime-done" --repo owner/repo \
  --description "AI 에이전트 완료" --color "6f42c1"
```

### 4. 인증

```bash
gh auth login    # jikime serve가 gh auth token을 자동으로 사용
```

### 5. 서비스 시작

```bash
jikime serve my-project/WORKFLOW.md

# 포트 명시
jikime serve --port 8888 my-project/WORKFLOW.md
```

### 6. 이슈 생성

```bash
gh issue create --repo owner/repo \
  --title "다크 모드 토글 추가" \
  --label "jikime-todo" \
  --body "다크/라이트 모드 전환 버튼을 추가해 주세요..."
```

### 7. 모니터링

```bash
# 터미널 대시보드
curl -s http://127.0.0.1:8888/

# JSON으로 실행 중인 작업 확인
curl -s http://127.0.0.1:8888/api/v1/state | jq '.running'
```

---

## 개발자 가이드라인

### main 브랜치에서 작업해도 되나요?

`before_run`의 `git reset --hard origin/main`은 **에이전트 전용 임시 워크스페이스** (`workspace.root`, 예: `/tmp/...`) 내부에서만 실행됩니다. 개발자의 로컬 개발 디렉토리는 완전히 별개이며 전혀 영향받지 않습니다.

다만, 개발자도 피처 브랜치를 사용하는 것을 권장합니다:

```
✅ 개발자: feature/my-work 브랜치 → PR → main에 머지
✅ 에이전트: fix/issue-N 브랜치   → PR → main에 머지
→ main은 항상 clean 상태 유지, 충돌 없음
```

### 오래 걸리는 작업 처리

많은 단계가 필요한 작업 (예: 새 프레임워크 설치)의 경우 타임아웃을 늘리세요:

```yaml
agent:
  max_turns: 10

claude:
  stall_timeout_ms: 300000   # 5분

hooks:
  timeout_ms: 120000         # npm install 등을 위해 2분
```

---

## 설정 레퍼런스

### 기본값

| 키 | 기본값 | 설명 |
|----|--------|------|
| `polling.interval_ms` | `30000` | 폴링 간격 |
| `agent.max_concurrent_agents` | `10` | 동시 세션 수 |
| `agent.max_turns` | `20` | 세션당 턴 수 |
| `agent.max_retry_backoff_ms` | `300000` | 최대 재시도 대기 |
| `claude.stall_timeout_ms` | `300000` | 스톨 종료 타임아웃 |
| `hooks.timeout_ms` | `60000` | 훅 실행 타임아웃 |
| `server.port` | `0` | HTTP API (비활성) |

### CLI 플래그

```bash
jikime serve [WORKFLOW.md] [flags]

Flags:
  -p, --port int   HTTP API 서버 포트 (0 = 비활성)
```

### 관련 문서

- [PR 라이프사이클 자동화](./pr-lifecycle.md)
- [구조화된 태스크 포맷](./task-format.md)
- [훅 설정 가이드](./hooks.md)
