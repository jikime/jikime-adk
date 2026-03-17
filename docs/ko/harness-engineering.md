# 하네스 엔지니어링 (Harness Engineering)

> GitHub Issue 생성부터 PR 자동 머지까지 — 사람 개입 없이 Claude가 처리하는 완전 자동화 오케스트레이션.

---

## 왜 하네스 엔지니어링인가?

개발팀은 반복적인 작업(버그 수정, 소규모 기능 추가, 문서 업데이트)에 상당한 시간을 소비합니다. 이런 작업들은 명확히 정의되어 있고 테스트 가능하지만, 개발자가 직접 처리하기엔 너무 단조롭습니다.

**하네스 엔지니어링**은 이 문제를 해결합니다.

| 기존 방식 | 하네스 엔지니어링 |
|-----------|------------------|
| 개발자가 이슈를 읽고, 브랜치 만들고, 코드 작성 | Claude가 자동으로 처리 |
| PR 생성·리뷰 요청·머지를 수동으로 | 완전 자동화 파이프라인 |
| 야간·주말엔 작업 중단 | 24시간 무중단 실행 |
| 개발자 컨텍스트 전환 비용 발생 | 개발자는 고난도 작업에 집중 |
| CI 큐에 쌓이는 단순 PR들 | 자율 에이전트가 백그라운드 처리 |

### 어떤 작업에 적합한가?

```
✅ 적합
  - 버그 수정 (재현 단계가 명확한 경우)
  - 소규모 기능 추가 (단일 컴포넌트 범위)
  - 의존성 업데이트
  - 문서·주석 작성
  - 테스트 추가
  - 린팅·포매팅 수정
  - 타입 오류 수정

⚠️ 주의 (충분히 명세화 필요)
  - 중규모 리팩터링
  - 새 API 엔드포인트 추가
  - 다중 파일 변경

❌ 부적합
  - 시스템 아키텍처 결정
  - 복잡한 비즈니스 로직 설계
  - 데이터베이스 스키마 대규모 변경
```

---

## 개념

**하네스 엔지니어링(Harness Engineering)** 이란 AI 에이전트가 소프트웨어 작업을 자율적으로 수행하도록 *하네스(제어 프레임워크)* 를 구축하는 방법론입니다. 마치 말 마구(harness)가 말의 힘을 정밀한 움직임으로 유도하듯, 하네스 엔지니어링은 Claude의 능력을 구조화되고 안전하며 반복 가능한 워크플로우로 이끌어냅니다.

JiKiME-ADK에서는 **`jikime serve`** 명령어로 하네스 엔지니어링이 구현됩니다. GitHub Issues를 폴링하고, 격리된 워크스페이스를 생성하고, Claude를 헤드리스로 실행하여 이슈 할당부터 PR 머지까지 전체 라이프사이클을 자동 처리합니다.

```
사람이 GitHub Issue를 작성 (라벨: jikime-todo)
        ↓
jikime serve가 감지 (15초마다 폴링)
        ↓
격리된 워크스페이스에서 git clone
        ↓
Claude가 이슈를 읽고, 브랜치 생성, 코드 작성
        ↓
Claude가 PR 생성 → 자동 머지
        ↓
GitHub: Issue 자동 Close (상태: Done)
        ↓
워크스페이스 정리 (before_remove 훅)
```

---

## 아키텍처

```
┌──────────────────────────────────────────────────────────┐
│                      jikime serve                        │
│                                                          │
│  ┌──────────┐    ┌──────────────┐    ┌────────────────┐  │
│  │ Tracker  │───▶│ Orchestrator │───▶│  Agent Runner  │  │
│  │ (GitHub) │    │              │    │    (Claude)    │  │
│  └──────────┘    └──────┬───────┘    └────────────────┘  │
│                         │                                 │
│  ┌──────────┐    ┌──────▼───────┐    ┌────────────────┐  │
│  │ HTTP API │    │  Workspace   │    │     Hooks      │  │
│  │  :8888   │    │  Manager     │    │   lifecycle    │  │
│  └──────────┘    └──────────────┘    └────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

| 컴포넌트 | 역할 |
|----------|------|
| **Tracker** | GitHub API로 `active_states` 라벨이 붙은 이슈 폴링 |
| **Orchestrator** | 상태 머신: 디스패치 → 재시도(지수 백오프) → 터미널 상태 reconcile |
| **Agent Runner** | `claude --print --output-format stream-json` 헤드리스 실행, 토큰 집계 |
| **Workspace Manager** | 이슈별 디렉토리 생성·재사용·삭제, 라이프사이클 훅 실행 |
| **HTTP API** | `http://127.0.0.1:<port>`에서 실시간 상태 스냅샷 제공 |

---

## WORKFLOW.md — 핵심 설정 파일

`jikime serve`를 사용하는 모든 프로젝트는 하나의 `WORKFLOW.md` 파일로 구성됩니다. YAML 프론트매터(런타임 설정) + Markdown 본문(프롬프트 템플릿)으로 구성됩니다.

```yaml
---
tracker:
  kind: github
  # api_key: $GITHUB_TOKEN   # 생략하면 gh auth token 자동 사용
  project_slug: owner/repo   # GitHub "owner/repo" 형식
  active_states:
    - jikime-todo             # 이 라벨이 붙은 이슈를 Claude가 처리
  terminal_states:
    - jikime-done             # 사람이 완료 처리
    - Done                    # PR 머지 시 GitHub이 자동 Close

polling:
  interval_ms: 15000          # 15초마다 GitHub 폴링

workspace:
  root: /tmp/jikime-myrepo   # 이슈별 격리 디렉토리

hooks:
  after_create: |             # 워크스페이스 최초 생성 시 1회
    git clone https://github.com/owner/repo.git .
    echo "[after_create] cloned to $(pwd)"

  before_run: |               # Claude 세션 시작 전마다
    git fetch origin
    git checkout main
    git reset --hard origin/main
    echo "[before_run] synced to $(git rev-parse --short HEAD)"

  after_run: |                # Claude 세션 종료 후마다 (실패해도 무시)
    echo "[after_run] done"
    if [ -d "/path/to/local-project/.git" ]; then
      cd "/path/to/local-project" && git pull --ff-only 2>&1 \
        && echo "[after_run] local repo synced at $(git rev-parse --short HEAD)" \
        || echo "[after_run] git pull skipped (local changes or diverged branch)"
    fi

  timeout_ms: 60000           # 훅 타임아웃 (60초)

agent:
  max_concurrent_agents: 1    # 동시 Claude 세션 수
  max_turns: 5                # 세션당 최대 멀티턴 횟수
  max_retry_backoff_ms: 300000 # 재시도 최대 대기 시간 (5분)

claude:
  command: claude              # Claude CLI 실행 명령어
  turn_timeout_ms: 3600000    # 세션 최대 시간 (1시간)
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

### 프롬프트 템플릿 변수

<div v-pre>

| 변수 | 예시 | 설명 |
|------|------|------|
| `{{ issue.id }}` | `9` | GitHub Issue 번호 |
| `{{ issue.identifier }}` | `owner/repo#9` | 사람이 읽기 좋은 키 |
| `{{ issue.title }}` | `푸터 컴포넌트 추가` | Issue 제목 |
| `{{ issue.description }}` | *(이슈 본문 전체)* | Issue 설명 |
| `{{ issue.state }}` | `jikime-todo` | 현재 상태 |
| `{{ issue.url }}` | `https://github.com/...` | Issue URL |
| `{{ issue.branch_name }}` | `fix/footer` | 트래커 제공 브랜치명 |
| `{{ attempt }}` | `2` | 재시도 횟수 (첫 실행: 빈 문자열) |

> **Strict mode**: 템플릿에 정의되지 않은 `{{ 변수 }}`가 남아있으면 렌더링 오류로 실행이 중단됩니다.

</div>

---

## WORKFLOW.md 생성 방법

### 방법 1: CLI 마법사 — `jikime serve init` (권장)

```bash
cd my-project
jikime serve init
```

대화형 프롬프트가 5가지를 물어봅니다:

```
? GitHub repo slug (owner/repo)  › owner/my-repo   ← git remote에서 자동 감지
? Active label                   › jikime-todo
? Workspace root                 › /tmp/jikime-my-repo
? HTTP status API port           › 8888 (recommended)
? Max concurrent agents          › 1 (safe, recommended)
```

`.claude/` 디렉토리가 있으면 자동으로 **JiKiME-ADK 모드** (J.A.R.V.I.S. 에이전트 스택 사용),
없으면 **기본 모드** (표준 git/PR 워크플로우)로 생성됩니다.

생성 후 다음 단계를 안내해 줍니다:

```
✓ WORKFLOW.md created

Configuration:
  Repo:    owner/my-repo
  Label:   jikime-todo
  Mode:    JiKiME-ADK (J.A.R.V.I.S. agent stack)
  Port:    8888

Next steps:
  1. Create GitHub labels:
     gh label create "jikime-todo" --repo owner/my-repo ...
     gh label create "jikime-done" --repo owner/my-repo ...

  2. Start the service:
     jikime serve WORKFLOW.md
```

### 방법 2: Claude Code 슬래시 커맨드 — `/jikime:harness`

JiKiME-ADK가 설치된 프로젝트에서 Claude Code 세션 안에서 실행합니다:

```
/jikime:harness
/jikime:harness --port 9999 --label ai-todo
/jikime:harness --basic --output my-workflow.md
```

Claude가 프로젝트를 분석해서 최적화된 WORKFLOW.md를 자동 생성합니다:
- Git remote에서 `owner/repo` 슬러그 자동 감지
- `.claude/` 디렉토리 유무로 모드 결정
- 기술 스택(package.json / go.mod / requirements.txt 등) 감지 → 전문 에이전트 선택

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--basic` | 꺼짐 | JiKiME-ADK 무시, 기본 모드 강제 |
| `--port N` | `8888` | HTTP API 포트 |
| `--label LABEL` | `jikime-todo` | 활성 라벨명 |
| `--output PATH` | `WORKFLOW.md` | 출력 파일 경로 |

### 방법 3: 예제 파일 복사

```bash
# JiKiME-ADK 설치 후
cp $(jikime --templates)/WORKFLOW.md.example ./WORKFLOW.md
# 편집
vim WORKFLOW.md
```

---

## 실행 방법 (Quick Start)

### 전체 설정 흐름

```bash
# 1. WORKFLOW.md 생성
cd my-project
jikime serve init

# 2. GitHub 라벨 생성
gh label create "jikime-todo" --repo owner/repo \
  --description "AI 에이전트 작업 대기" --color "0e8a16"

gh label create "jikime-done" --repo owner/repo \
  --description "AI 에이전트 완료" --color "6f42c1"

# 3. GitHub 인증 확인
gh auth login
gh auth status

# 4. 서비스 시작
jikime serve WORKFLOW.md

# 포트 명시 (WORKFLOW.md 설정보다 우선)
jikime serve --port 8888 WORKFLOW.md
```

### 이슈 할당

```bash
# 기존 이슈에 라벨 추가
gh issue edit 42 --repo owner/repo --add-label "jikime-todo"

# 새 이슈 바로 생성
gh issue create --repo owner/repo \
  --title "다크 모드 토글 추가" \
  --label "jikime-todo" \
  --body "헤더에 다크/라이트 모드 전환 버튼을 추가해 주세요.

## 요구사항
- 로컬스토리지에 설정 저장
- 기본값: 시스템 설정 따름
- CSS 변수 기반 구현"
```

> **Tip**: Issue 설명이 구체적일수록 Claude의 구현 품질이 높아집니다. 재현 단계, 기대 동작, 파일 경로 등을 포함하세요.

---

## 전체 실행 플로우 (상세)

```
┌─────────────────────────────────────────────────────────────────┐
│  1. POLL (매 15초)                                              │
│     GitHub API → active_states 라벨 이슈 수집                   │
│     정렬: 우선순위 오름차순 → created_at 오래된 것 먼저          │
│           → identifier 사전순 (동점 처리)                       │
└──────────────────────────────┬──────────────────────────────────┘
                               │ 새 이슈 발견
┌──────────────────────────────▼──────────────────────────────────┐
│  2. DISPATCH                                                     │
│     ✓ running map에 없음 확인                                    │
│     ✓ claimed set에 없음 확인                                    │
│     ✓ 동시 슬롯 여유 있음 확인 (max_concurrent_agents)           │
│     → claimed 표시 후 워커 고루틴 생성                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  3. WORKSPACE SETUP                                              │
│     경로: <workspace.root>/<sanitized_identifier>/              │
│     예: /tmp/jikime-myrepo/owner_repo_42/                       │
│                                                                  │
│     [최초 생성 시]  after_create 훅 실행                         │
│       → git clone https://github.com/owner/repo.git .           │
│                                                                  │
│     [매 세션 전]    before_run 훅 실행                           │
│       → git fetch origin                                         │
│       → git checkout main                                        │
│       → git reset --hard origin/main   ← 항상 최신 main 기준    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  4. PROMPT RENDERING                                             │
│     WORKFLOW.md 본문에 issue 필드 치환                           │
│     {{ issue.id }} → "42"                                        │
│     {{ issue.title }} → "다크 모드 토글 추가"                    │
│     미정의 변수 → template_render_error 발생 → 이슈 재시도       │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  5. CLAUDE EXECUTION                                             │
│     실행 위치: <workspace_path>/ (소스 저장소와 완전 분리)        │
│                                                                  │
│     claude --print \                                             │
│            --output-format stream-json \                         │
│            --verbose \                                           │
│            --dangerously-skip-permissions \                      │
│            --max-turns <max_turns> \                             │
│            "렌더링된 프롬프트"                                   │
│                                                                  │
│     [스톨 감지] stall_timeout_ms 동안 출력 없으면 강제 종료       │
│     [턴 타임아웃] turn_timeout_ms 초과 시 세션 종료               │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  6. GIT FLOW (Claude가 실행)                                     │
│     git checkout -b fix/issue-42                                 │
│     (코드 작성 / 파일 수정)                                      │
│     git add -A                                                   │
│     git commit -m "fix: owner/repo#42 - 다크 모드 토글 추가"    │
│     git push origin fix/issue-42                                 │
│                                                                  │
│     gh pr create \                                               │
│       --title "fix: 다크 모드 토글 추가" \                       │
│       --body "Closes #42" \                                      │
│       --base main \                                              │
│       --head fix/issue-42                                        │
│                                                                  │
│     gh pr merge --squash --delete-branch --admin                 │
└──────────────────────────────┬──────────────────────────────────┘
                               │ PR 머지
┌──────────────────────────────▼──────────────────────────────────┐
│  7. AUTO-CLOSE                                                   │
│     GitHub: "Closes #42" 감지 → Issue #42 자동 Close            │
│     Issue 상태: Done (terminal_states에 포함)                    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  8. RECONCILE & CLEANUP                                          │
│     다음 폴 틱에서 Issue #42가 terminal 상태임을 감지            │
│     after_run 훅 실행 (실패해도 무시)                            │
│     before_remove 훅 실행 (실패해도 무시)                        │
│     워크스페이스 디렉토리 삭제                                   │
│     claimed set에서 제거                                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## Git과의 동작 방식

### 브랜치 전략

하네스 엔지니어링은 **브랜치 격리**를 핵심 안전 장치로 사용합니다.

```
main ──●────────────────●────────────────●──▶
        │                │                │
        └─ fix/issue-42  └─ fix/issue-43  └─ fix/issue-44
           (Claude #1)      (Claude #2)      (Claude #3)
```

- 에이전트는 절대 `main`에 직접 커밋하지 않습니다.
- 각 이슈는 전용 `fix/issue-N` 브랜치에서 작업합니다.
- PR은 squash merge 후 브랜치가 자동 삭제됩니다.

### 워크스페이스 격리

```
/tmp/jikime-myrepo/
  owner_repo_42/    ← Issue #42 전용 (독립 git 저장소)
  owner_repo_43/    ← Issue #43 전용 (독립 git 저장소)
  owner_repo_44/    ← Issue #44 전용 (독립 git 저장소)
```

각 워크스페이스는 **완전히 독립적인 git 저장소**입니다. 에이전트들이 동시에 실행되어도 서로의 파일에 영향을 주지 않습니다.

### 충돌 방지 메커니즘

| 위험 요소 | 방어 방법 |
|-----------|-----------|
| 두 에이전트가 같은 브랜치에 푸시 | 이슈 ID 기반 전용 브랜치 (`fix/issue-N`) |
| 재시도 시 오래된 코드 기반 | `before_run`: `git reset --hard origin/main` |
| 에이전트와 개발자 브랜치 충돌 | 네이밍 규칙으로 완전 분리 (`fix/issue-*`) |
| 여러 에이전트가 main 경쟁 | `max_concurrent_agents: 1` (기본값) |
| 소스 저장소 오염 | 에이전트는 `workspace.root` 아래에서만 실행 |

### 개발자와 에이전트의 공존

```
✅ 개발자:  feature/my-feature-branch → PR → main
✅ 에이전트: fix/issue-42             → PR → main
→ main은 항상 clean, 충돌 없음
```

`before_run`의 `git reset --hard origin/main`은 에이전트 전용 워크스페이스(`/tmp/...`) 내부에서만 실행됩니다. 개발자의 로컬 저장소는 완전히 별개이며 영향받지 않습니다.

---

## 상태 확인 방법

### 터미널 로그

`jikime serve` 실행 시 구조화된 로그가 stderr에 출력됩니다:

```
  ╔══════════════════════════════════════╗
  ║   jikime serve — Agent Orchestrator  ║
  ║   Powered by Claude Code + Symphony  ║
  ╚══════════════════════════════════════╝

  Workflow:    /my-project/WORKFLOW.md
  Tracker:     github / owner/repo
  Workspace:   /tmp/jikime-myrepo
  Concurrency: 1 agents
  Poll:        every 15000ms
  HTTP API:    http://127.0.0.1:8888

time=2026-03-09T10:00:00 level=INFO msg="polling..."
time=2026-03-09T10:00:00 level=INFO msg="dispatching issue" issue_id=42
time=2026-03-09T10:00:01 level=INFO msg="agent event" type=session_started issue_id=42
time=2026-03-09T10:00:45 level=INFO msg="agent event" type=turn_completed issue_id=42
```

### HTTP 상태 API

`server.port`가 설정되면 웹 대시보드를 사용할 수 있습니다:

**텍스트 대시보드** — 사람이 읽기 좋은 형태:

```bash
curl http://127.0.0.1:8888/

# 출력 예:
# jikime serve — 2026-03-09T10:05:00Z
#
# Running (1):
#   owner/repo#42        turns=3   Implementing dark mode toggle...
#
# Retrying (0):
#
# Tokens: input=8420 output=2140 total=10560 runtime=42.3s
```

**JSON 스냅샷** — 프로그래밍적 상태 확인:

```bash
curl -s http://127.0.0.1:8888/api/v1/state | jq .

# {
#   "generated_at": "2026-03-09T10:05:00Z",
#   "counts": { "running": 1, "retrying": 0 },
#   "running": [
#     {
#       "IssueIdentifier": "owner/repo#42",
#       "TurnCount": 3,
#       "LastMessage": "Implementing dark mode toggle..."
#     }
#   ],
#   "jikime_totals": {
#     "InputTokens": 8420,
#     "OutputTokens": 2140,
#     "TotalTokens": 10560,
#     "SecondsRunning": 42.3
#   }
# }
```

**즉시 폴링 트리거** — 15초 기다리기 싫을 때:

```bash
curl -s -X POST http://127.0.0.1:8888/api/v1/refresh
```

**라이브 모니터링**:

```bash
# 3초마다 대시보드 갱신
watch -n 3 'curl -s http://127.0.0.1:8888/'

# 실행 중인 이슈만 추적
watch -n 5 'curl -s http://127.0.0.1:8888/api/v1/state | jq ".running[].IssueIdentifier"'

# 토큰 사용량 모니터링
watch -n 10 'curl -s http://127.0.0.1:8888/api/v1/state | jq ".jikime_totals"'
```

### WORKFLOW.md 핫 리로드

`jikime serve` 실행 중에 `WORKFLOW.md`를 수정하면 **재시작 없이** 자동 반영됩니다:

```bash
# 실행 중에 동시 세션 수를 늘리기
vim WORKFLOW.md   # max_concurrent_agents: 3으로 변경
# → jikime serve가 즉시 반영
# → "WORKFLOW.md reloaded" 로그 출력
```

적용되는 변경: 폴링 간격, 동시성 제한, active/terminal 상태, 훅, 프롬프트 템플릿.
실행 중인 세션은 영향받지 않으며 다음 디스패치부터 적용됩니다.

---

## 기능 상세

### 워크스페이스 격리 (Safety Invariants)

Symphony SPEC §9.5의 안전 불변식을 준수합니다:

1. **에이전트는 반드시 이슈 전용 워크스페이스 경로에서만 실행**: `cwd == workspace_path` 검증
2. **워크스페이스 경로는 반드시 workspace root 내부**: 경로 탈출(path traversal) 방지
3. **워크스페이스 키 sanitization**: `[A-Za-z0-9._-]` 이외 문자는 `_`으로 치환

### 라이프사이클 훅

| 훅 | 실행 시점 | 실패 시 동작 | 주요 용도 |
|----|-----------|-------------|-----------|
| `after_create` | 워크스페이스 최초 생성 시 (1회) | **치명적** — 이슈 실행 중단 | `git clone` |
| `before_run` | 모든 Claude 세션 시작 전 | **치명적** — 해당 시도 중단 | `git reset --hard origin/main` |
| `after_run` | 모든 Claude 세션 종료 후 | **무시** (로그만) | 아티팩트 수집, 알림 |
| `before_remove` | 워크스페이스 삭제 전 | **무시** (로그만) | 백업, 아카이브 |

모든 훅은 `hooks.timeout_ms` (기본 60초) 내에 완료되어야 합니다.

### 지수 백오프 재시도

실패한 세션은 자동으로 재시도됩니다:

```
공식: min(10000 × 2^(시도횟수-1), max_retry_backoff_ms)

시도 1 → 10,000ms (10초)
시도 2 → 20,000ms (20초)
시도 3 → 40,000ms (40초)
시도 4 → 80,000ms (80초)
...
최대 → max_retry_backoff_ms (기본 300,000ms = 5분)
```

이슈가 `terminal_states`로 전환되면 재시도가 자동 취소됩니다.

정상 완료 후에도 이슈가 아직 active 상태면 1초 후 continuation retry를 스케줄해 상태를 재확인합니다.

### 토큰 집계

`--output-format stream-json`으로 매 세션의 토큰 사용량이 실시간 캡처·누적됩니다:

```bash
curl -s http://127.0.0.1:8888/api/v1/state | jq '.jikime_totals'
# {
#   "InputTokens": 45820,
#   "OutputTokens": 12340,
#   "TotalTokens": 58160,
#   "SecondsRunning": 1842.5
# }
```

---

## 설정 레퍼런스

### 전체 설정 키

| 키 | 타입 | 기본값 | 설명 |
|----|------|--------|------|
| `tracker.kind` | string | — | `"github"` 또는 `"linear"` (필수) |
| `tracker.api_key` | string | `$GITHUB_TOKEN` or `gh auth token` | GitHub 토큰 |
| `tracker.project_slug` | string | — | `"owner/repo"` (필수) |
| `tracker.active_states` | list | `["Todo", "In Progress"]` | 처리할 이슈 상태 |
| `tracker.terminal_states` | list | `["Closed", "Cancelled", "Done"]` | 완료 이슈 상태 |
| `polling.interval_ms` | int | `30000` | 폴링 간격 (ms) |
| `workspace.root` | path | `/tmp/jikime_workspaces` | 워크스페이스 루트 |
| `hooks.after_create` | script | — | 최초 생성 훅 |
| `hooks.before_run` | script | — | 세션 시작 전 훅 |
| `hooks.after_run` | script | — | 세션 종료 후 훅 |
| `hooks.before_remove` | script | — | 삭제 전 훅 |
| `hooks.timeout_ms` | int | `60000` | 훅 타임아웃 |
| `agent.max_concurrent_agents` | int | `10` | 동시 세션 수 |
| `agent.max_turns` | int | `20` | 세션당 최대 턴 |
| `agent.max_retry_backoff_ms` | int | `300000` | 최대 재시도 대기 |
| `claude.command` | string | `"claude"` | Claude CLI 실행 명령어 |
| `claude.turn_timeout_ms` | int | `3600000` | 세션 최대 시간 (1시간) |
| `claude.stall_timeout_ms` | int | `300000` | 스톨 종료 타임아웃 |
| `server.port` | int | `0` (비활성) | HTTP API 포트 |

### CLI 플래그

```bash
jikime serve [WORKFLOW.md] [flags]

플래그:
  -p, --port int   HTTP API 서버 포트 (0 = 비활성, WORKFLOW.md 설정보다 우선)

서브커맨드:
  init             대화형 마법사로 WORKFLOW.md 생성
```

### 오래 걸리는 작업 처리

많은 단계가 필요한 작업 (예: 새 프레임워크 설치, 대규모 리팩터링):

```yaml
agent:
  max_turns: 15           # 더 많은 멀티턴 허용

claude:
  turn_timeout_ms: 7200000  # 2시간 (기본 1시간에서 늘림)
  stall_timeout_ms: 600000  # 10분 스톨 감지 (기본 5분에서 늘림)

hooks:
  timeout_ms: 180000      # npm install / pip install 등을 위해 3분
```

---

## 관련 문서

- [PR 라이프사이클 자동화](./pr-lifecycle.md)
- [구조화된 태스크 포맷](./task-format.md)
- [POC-First 워크플로우](./poc.md)
