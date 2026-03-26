# JiKiME Team — 멀티 에이전트 팀 오케스트레이션 완전 가이드

**Claude Code 에이전트들을 팀으로 구성하여 복잡한 작업을 병렬로 처리하는 시스템**

> **버전:** JiKiME-ADK v1.5.0+
> **최종 업데이트:** 2026-03-20

---

## 목차

1. [개요 및 아키텍처](#1-개요-및-아키텍처)
2. [CLI 명령어 전체 레퍼런스](#2-cli-명령어-전체-레퍼런스)
3. [웹 UI (Webchat) 기능](#3-웹-ui-webchat-기능)
4. [REST API 및 SSE 엔드포인트](#4-rest-api-및-sse-엔드포인트)
5. [파일 시스템 구조](#5-파일-시스템-구조)
6. [환경 변수](#6-환경-변수)
7. [팀 워크플로우 패턴](#7-팀-워크플로우-패턴)
8. [Git Worktree 격리](#8-git-worktree-격리)
9. [GitHub Issues / Harness 통합](#9-github-issues--harness-통합)
10. [템플릿 시스템](#10-템플릿-시스템)
11. [실전 예시](#11-실전-예시)
12. [트러블슈팅](#12-트러블슈팅)

---

## 1. 개요 및 아키텍처

### JiKiME Team이란?

JiKiME Team은 여러 Claude Code 에이전트 인스턴스를 하나의 팀으로 묶어, 복잡한 소프트웨어 개발 작업을 **병렬로** 처리하는 오케스트레이션 시스템입니다.

```
┌─────────────────────────────────────────────────────┐
│                   jikime team                        │
│                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐       │
│  │  leader  │◄──►│ worker-1 │    │ worker-2 │       │
│  │ (Claude) │    │ (Claude) │    │ (Claude) │       │
│  └──────────┘    └──────────┘    └──────────┘       │
│        │               │               │            │
│        └───────────────┴───────────────┘            │
│                        │                            │
│            ┌───────────▼──────────┐                 │
│            │   ~/.jikime/teams/   │                 │
│            │  tasks/ registry/    │                 │
│            │  inbox/ costs/       │                 │
│            └──────────────────────┘                 │
└─────────────────────────────────────────────────────┘
```

### 핵심 개념

| 개념 | 설명 |
|------|------|
| **Team** | 공유 작업 저장소를 가진 에이전트 그룹 |
| **Agent** | Claude Code CLI를 실행하는 독립 프로세스 (tmux 또는 subprocess) |
| **Task** | 에이전트가 처리하는 작업 단위 (pending → in_progress → done) |
| **Role** | 에이전트 역할: leader, worker, reviewer |
| **Workspace** | 에이전트별 격리된 git worktree |
| **Template** | 팀 구성 정의 (에이전트 수, 역할, 초기 작업) |
| **Budget** | 팀 전체 토큰 사용량 제한 |

### 스폰 백엔드

| 백엔드 | 특징 | 사용 시기 |
|--------|------|-----------|
| `tmux` (기본값) | 대화형, 터미널에서 실시간 모니터링 가능 | 개발 환경, 디버깅 |
| `subprocess` | 비대화형, 로그 파일로 출력 | CI/CD, 자동화 |

---

## 2. CLI 명령어 전체 레퍼런스

최상위 진입점: `jikime team` (별칭: `jikime t`)

### 2.1 팀 생성 및 관리

#### `jikime team create <team-name>`

빈 팀 작업공간을 생성합니다.

```bash
jikime team create <team-name> [플래그]

플래그:
  -w, --workers int        워커 에이전트 수 (0 = 무제한, 기본값: 0)
  -b, --backend string     스폰 백엔드: tmux 또는 subprocess (기본값: tmux)
      --budget int         토큰 예산 제한 (0 = 무제한)
      --timeout int        실행 타임아웃 (초, 0 = 무제한)
      --max-agents int     최대 동시 에이전트 수 (0 = 무제한)
  -t, --template string    초기 템플릿 이름

예시:
  jikime team create my-team
  jikime team create auth-team --workers 3 --budget 100000
  jikime team create api-team --template leader-worker --backend subprocess
```

**생성되는 디렉토리:**
```
~/.jikime/teams/<team-name>/
├── config.json      # 팀 설정
├── webchat.json     # 웹챗 메타데이터
├── tasks/           # 작업 파일들
├── inbox/           # 에이전트 메시지 수신함
├── registry/        # 에이전트 등록 정보
├── costs/           # 토큰 사용 기록
└── events/          # 이벤트 로그
```

---

#### `jikime team launch`

템플릿에서 전체 팀을 한 번에 생성하고 시작합니다.

```bash
jikime team launch [플래그]

플래그:
  -t, --template string    템플릿 이름 (필수)
      --name string        팀 이름 (지정 안 하면 자동 생성)
  -g, --goal string        에이전트 프롬프트에 주입할 목표
  -b, --backend string     스폰 백엔드 (기본값: tmux)
  -w, --worktree           에이전트별 격리된 git worktree 생성
      --budget int         토큰 예산 (템플릿 기본값 재정의)

예시:
  jikime team launch --template leader-worker \
    --goal "implement user authentication with JWT"

  jikime team launch --template leader-worker-reviewer \
    --name auth-team \
    --goal "redesign API layer" \
    --worktree \
    --budget 200000
```

**자동 처리 순서:**
1. 템플릿 로드 및 팀 디렉토리 구조 생성
2. 템플릿에 정의된 초기 작업 자동 생성
3. `--worktree` 플래그 시 각 에이전트마다 git worktree 생성
4. 각 에이전트 역할별 프롬프트 생성 (목표 주입)
5. 모든 에이전트 자동 스폰

---

#### `jikime team spawn <team-name>`

기존 팀에 새 에이전트를 추가합니다.

```bash
jikime team spawn <team-name> [플래그]

플래그:
  -r, --role string        에이전트 역할: leader, worker, reviewer (기본값: worker)
      --agent-id string    에이전트 ID (지정 안 하면 자동 생성: agent-XXXXXXXX)
  -b, --backend string     스폰 백엔드 (기본값: tmux)
      --worktree string    이 에이전트의 git worktree 경로
  -p, --prompt string      에이전트 초기 프롬프트
      --skip-permissions   --dangerously-skip-permissions 전달 (기본값: true)
      --resume             이전 Claude 세션 복구 시도

예시:
  jikime team spawn my-team --role leader --agent-id leader
  jikime team spawn my-team --role worker --agent-id worker-1
  jikime team spawn my-team --role worker \
    --worktree ~/.jikime/worktrees/my-team/worker-1 \
    --prompt "Focus on implementing the database layer"
```

**Tmux 백엔드 동작:**
- 세션명: `jikime-<team>-<agent-id>`
- 에이전트는 `JIKIME_AGENT_ID`, `JIKIME_TEAM_NAME` 환경변수로 자신을 식별
- `claude` 명령어 실행 후 대화형 세션 유지

---

#### `jikime team status <team-name>`

팀의 현재 상태를 조회합니다.

```bash
jikime team status <team-name> [--json]

예시:
  jikime team status my-team
  jikime team status my-team --json | jq .agents
```

**출력 예시:**
```
Team: my-team
Dir:  ~/.jikime/teams/my-team

Agents (3):
  ✅ worker-1 [worker] task:abc12345
  ✅ worker-2 [worker] task:def67890
  ❌ leader   [leader] task:-

Tasks (5): todo=2 wip=2 done=1 blocked=0

Tokens: 45,230 / 100,000 (45.2%)
```

---

#### `jikime team stop <team-name>`

팀의 모든 에이전트를 중지합니다.

```bash
jikime team stop my-team
```

---

#### `jikime team discover`

현재 머신에서 실행 중인 팀들을 탐색합니다.

```bash
# 모든 활성 팀 나열
jikime team discover list [--json]

# 기존 팀에 합류 (새 에이전트로)
jikime team discover join <team-name> \
  --role worker \
  --agent-id my-worker

# 합류 요청 승인 (리더로서)
jikime team discover approve <agent-id> --team my-team

# 합류 요청 거부
jikime team discover reject <agent-id>
```

---

### 2.2 작업(Task) 관리

#### `jikime team tasks create <team-name> <title>`

새 작업을 생성합니다.

```bash
jikime team tasks create <team-name> <title> [플래그]

플래그:
  -d, --desc string        작업 상세 설명
      --dod string         완료 정의 (Definition of Done)
      --depends-on string  쉼표로 구분된 의존 작업 ID
  -p, --priority int       우선순위 (높을수록 중요, 기본값: 0)
      --tags string        쉼표로 구분된 태그

예시:
  jikime team tasks create my-team "Implement login endpoint"

  jikime team tasks create my-team "Design database schema" \
    --desc "Create users, sessions, tokens tables" \
    --dod "All tables created with proper indexes and constraints" \
    --priority 3 \
    --tags "database,schema"

  jikime team tasks create my-team "Write unit tests" \
    --depends-on abc12345,def67890 \
    --priority 1
```

---

#### `jikime team tasks list <team-name>`

팀의 모든 작업을 나열합니다.

```bash
jikime team tasks list <team-name> [플래그]

플래그:
  -s, --status string    상태별 필터: pending|in_progress|done|blocked|failed
  -a, --agent string     에이전트 ID별 필터

예시:
  jikime team tasks list my-team
  jikime team tasks list my-team --status in_progress
  jikime team tasks list my-team --agent worker-1
  jikime team tasks list my-team --status done --agent worker-1
```

---

#### `jikime team tasks get <team-name> <task-id>`

특정 작업의 상세 정보를 조회합니다.

```bash
jikime team tasks get my-team abc12345
```

---

#### `jikime team tasks update <team-name> <task-id>`

작업 상태 또는 메타데이터를 업데이트합니다.

```bash
jikime team tasks update <team-name> <task-id> [플래그]

플래그:
  -t, --title string     새 제목
  -d, --desc string      새 설명
      --dod string       새 완료 정의
  -p, --priority int     새 우선순위
  -s, --status string    상태 전환: pending|in_progress|done|blocked|failed
  -a, --agent string     에이전트 ID (in_progress 전환 시 필수)
  -r, --result string    결과 요약 (done/failed 시 사용)

상태 전환 규칙:
  pending    → in_progress  에이전트가 claim (--agent 필수)
  in_progress→ done         에이전트가 완료 (--result 권장)
  in_progress→ failed       에이전트가 실패 보고
  any        → blocked      의존 작업 대기 중
  blocked    → pending      차단 해제

예시:
  # 작업 claim (시작)
  jikime team tasks update my-team abc123 \
    --status in_progress --agent worker-1

  # 작업 완료
  jikime team tasks update my-team abc123 \
    --status done --agent worker-1 \
    --result "Implemented with 95% test coverage"

  # 작업 실패 표시
  jikime team tasks update my-team abc123 \
    --status failed --agent worker-1 \
    --result "API returned 403, need credentials"
```

---

#### `jikime team tasks claim <team-name> <task-id>`

현재 에이전트가 작업을 claim합니다 (환경변수 자동 사용).

```bash
jikime team tasks claim <team-name> <task-id> [--agent <id>]

# 환경변수 사용 (에이전트 프로세스 내부에서)
export JIKIME_AGENT_ID=worker-1
jikime team tasks claim my-team abc12345

# 명시적 에이전트 ID
jikime team tasks claim my-team abc12345 --agent worker-1
```

---

#### `jikime team tasks complete <team-name> <task-id>`

작업을 완료 처리합니다.

```bash
jikime team tasks complete <team-name> <task-id> [플래그]

플래그:
  -a, --agent string     에이전트 ID (필수)
  -r, --result string    결과 요약

예시:
  jikime team tasks complete my-team abc12345 \
    --agent worker-1 \
    --result "All tests passing, ready for code review"
```

---

#### `jikime team tasks wait <team-name>`

모든 작업이 완료될 때까지 대기합니다 (CI/CD 통합에 유용).

```bash
jikime team tasks wait <team-name> [플래그]

플래그:
  -t, --timeout int     최대 대기 시간 (초, 0 = 무제한)
  -i, --interval int    폴링 간격 (초, 기본값: 5)

예시:
  jikime team tasks wait my-team --timeout 3600
  jikime team tasks wait my-team --interval 10

# 종료 코드:
  0 = 모든 작업 완료 (done 상태)
  1 = 타임아웃 초과
  2 = 실패한 작업 존재
```

**진행률 표시:**
```
tasks: 5/10 done | wip:3 pending:2 blocked:0
```

---

### 2.3 계획(Plan) 관리

#### `jikime team plan submit <team-name>`

워커가 리더에게 작업 계획을 제출합니다.

```bash
jikime team plan submit <team-name> [플래그]

플래그:
  -t, --title string     계획 제목 (기본값: "Plan")
  -b, --body string      계획 본문 (인라인 텍스트)
  -f, --file string      파일에서 계획 본문 읽기
  -a, --agent string     제출하는 에이전트 ID

예시:
  jikime team plan submit my-team \
    --title "Database Schema Design" \
    --body "Propose using PostgreSQL with 3 tables: users, sessions, tokens"

  jikime team plan submit my-team \
    --title "API Implementation Plan" \
    --file plan.md \
    --agent worker-1
```

---

#### `jikime team plan approve <plan-id>`

리더가 계획을 승인합니다.

```bash
jikime team plan approve <plan-id> [--reviewer <agent-id>]

예시:
  jikime team plan approve plan-abc12345
  jikime team plan approve plan-abc12345 --reviewer leader
```

---

#### `jikime team plan reject <plan-id>`

리더가 계획을 거부합니다.

```bash
jikime team plan reject <plan-id> [--reviewer <id>] [--reason <text>]

예시:
  jikime team plan reject plan-abc12345 \
    --reason "Need to consider distributed caching layer"
```

---

#### `jikime team plan list`

계획 목록을 조회합니다.

```bash
jikime team plan list [--team <team-name>]
```

---

### 2.4 보드(Board) 관리

#### `jikime team board show <team-name>`

팀 보드의 현재 스냅샷을 표시합니다.

```bash
jikime team board show <team-name> [--json]

예시:
  jikime team board show my-team
  jikime team board show my-team --json | jq .tasks
```

**출력 예시:**
```
╔══════════════════════════════════╗
║  Team Board: my-team             ║
╚══════════════════════════════════╝

Agents (3):
  ✅ worker-1  [active]   role:worker  task:abc12345
  ✅ worker-2  [active]   role:worker  task:def67890
  ❌ leader    [offline]  role:leader  task:-

Tasks (5 total):
  pending:2  in_progress:2  done:1  failed:0  blocked:0
```

---

#### `jikime team board live <team-name>`

터미널에서 실시간 갱신 보드를 표시합니다 (Ctrl+C로 종료).

```bash
jikime team board live <team-name> [--interval <seconds>]

플래그:
  -i, --interval int    갱신 간격 (초, 기본값: 3)
```

---

#### `jikime team board attach <team-name>`

모든 에이전트 창을 하나의 tmux 대시보드 세션으로 연결합니다.

```bash
jikime team board attach my-team

# tmux 세션에 연결 후:
# Ctrl-b n → 다음 에이전트 창
# Ctrl-b p → 이전 에이전트 창
# Ctrl-b d → 세션 분리
```

---

#### `jikime team board serve [team-name]`

웹 대시보드 HTTP 서버를 시작합니다.

```bash
jikime team board serve [team-name] [플래그]

플래그:
  -p, --port int         HTTP 포트 (기본값: 8080)
      --host string      바인드 주소 (기본값: 127.0.0.1)
  -i, --interval float   SSE 푸시 간격 (초, 기본값: 2.0)

예시:
  jikime team board serve my-team
  jikime team board serve --port 3000 --host 0.0.0.0
```

**제공 엔드포인트:**
- `GET /` → React SPA 대시보드
- `GET /api/overview` → 모든 팀 목록 (JSON)
- `GET /api/team/:name` → 특정 팀 스냅샷 (JSON)
- `GET /api/events/:name` → 실시간 SSE 스트림

---

#### `jikime team board overview`

모든 팀의 개요를 표시합니다.

```bash
jikime team board overview
```

---

### 2.5 예산(Budget) 관리

#### `jikime team budget show <team-name>`

토큰 사용량과 예산을 표시합니다.

```bash
jikime team budget show <team-name> [--agent <id>]

예시:
  jikime team budget show my-team
  jikime team budget show my-team --agent worker-1
```

**출력 예시:**
```
Budget for team 'my-team' (limit: 100,000 tokens)

AGENT          INPUT      OUTPUT     TOTAL      %
worker-1       12,000     3,400      15,400     15.4%
worker-2       10,500     2,800      13,300     13.3%
leader         8,000      2,100      10,100     10.1%

TOTAL          38,800 tokens used
Budget used: 38.8% (38,800 / 100,000)
```

---

#### `jikime team budget set <team-name> <tokens>`

팀의 토큰 예산을 설정합니다.

```bash
jikime team budget set my-team 200000
```

---

#### `jikime team budget report <team-name>`

에이전트가 토큰 사용량을 보고합니다 (Claude 훅에서 호출).

```bash
jikime team budget report <team-name> [플래그]

플래그:
  -a, --agent string       에이전트 ID (기본값: $JIKIME_AGENT_ID)
      --task string        작업 ID
      --model string       모델명 (예: claude-sonnet-4-6)
      --input-tokens int   입력 토큰 수
      --output-tokens int  출력 토큰 수

예시 (Claude 훅 스크립트):
  #!/bin/bash
  jikime team budget report "$JIKIME_TEAM_NAME" \
    --agent "$JIKIME_AGENT_ID" \
    --task "$JIKIME_TASK_ID" \
    --input-tokens 1234 \
    --output-tokens 567 \
    --model claude-sonnet-4-6
```

---

### 2.6 워크스페이스(Workspace) 관리

Git worktree 기반 에이전트 격리 작업공간을 관리합니다.

Worktree 경로: `~/.jikime/worktrees/<team-name>/<agent-id>/`
브랜치명: `jikime-<team-name>-<agent-id>`

#### `jikime team workspace list <team-name>`

팀의 모든 활성 worktree를 나열합니다.

```bash
jikime team workspace list my-team

# 출력:
# Workspaces for team 'my-team':
#   worker-1  ~/.jikime/worktrees/my-team/worker-1
#   worker-2  ~/.jikime/worktrees/my-team/worker-2
```

---

#### `jikime team workspace checkpoint <team-name>`

현재 워크스페이스 변경사항을 커밋합니다.

```bash
jikime team workspace checkpoint <team-name> [플래그]

플래그:
  -a, --agent string     에이전트 ID (기본값: $JIKIME_AGENT_ID)
  -m, --message string   커밋 메시지 (기본값: "checkpoint: <agent> <timestamp>")

예시:
  jikime team workspace checkpoint my-team --agent worker-1
  jikime team workspace checkpoint my-team \
    --agent worker-1 \
    --message "feat: implement login endpoint"
```

---

#### `jikime team workspace merge <team-name>`

에이전트 워크스페이스 브랜치를 메인 브랜치로 병합합니다.

```bash
jikime team workspace merge <team-name> [플래그]

플래그:
  -a, --agent string     에이전트 ID (기본값: $JIKIME_AGENT_ID)
  -t, --target string    대상 브랜치 (기본값: main)
      --cleanup          병합 후 worktree 제거

예시:
  jikime team workspace merge my-team --agent worker-1
  jikime team workspace merge my-team \
    --agent worker-1 \
    --target develop \
    --cleanup
```

---

#### `jikime team workspace cleanup <team-name>`

워크스페이스를 제거합니다.

```bash
jikime team workspace cleanup <team-name> [--agent <id>]

# 특정 에이전트 워크스페이스만 제거
jikime team workspace cleanup my-team --agent worker-1

# 팀 전체 워크스페이스 제거
jikime team workspace cleanup my-team
```

---

#### `jikime team workspace status <team-name>`

에이전트 워크스페이스의 git diff 상태를 표시합니다.

```bash
jikime team workspace status my-team --agent worker-1
```

---

### 2.7 기타 명령어

```bash
# 팀 구성 조회/수정
jikime team config show <team-name>
jikime team config set <team-name> <key> <value>
jikime team config get <team-name> <key>
jikime team config health         # ~/.jikime 디렉토리 상태 확인

# 수신함 (에이전트 간 메시지)
jikime team inbox <team-name>     # 메시지 목록 조회

# 에이전트 신원
jikime team identity <team-name>  # 에이전트 ID/팀 정보 조회

# 세션 관리 (상태 스냅샷)
jikime team session <team-name>   # 세션 목록 조회

# 생명주기 훅 (에이전트 종료 시 자동 호출)
jikime team lifecycle on-exit     # 에이전트 종료 정리

# 템플릿
jikime team template list         # 사용 가능한 템플릿 목록
```

---

## 3. 웹 UI (Webchat) 기능

Webchat(`http://localhost:<port>`)에서 Team 기능을 시각적으로 관리할 수 있습니다.

### 3.1 팀 탭 접근

상단 헤더의 **Team** 탭 버튼을 클릭하면 팀 대시보드가 열립니다.

```
[채팅] [터미널] [파일] [Git] [Team] ← 클릭
```

- **프로젝트 미선택**: FolderOpen 아이콘과 함께 안내 메시지가 표시됩니다.
  → 사이드바에서 프로젝트를 먼저 선택하거나 세션을 클릭하세요.
- **팀 미선택**: Users 아이콘과 함께 팀 선택 안내가 표시됩니다.
  → 상단 드롭다운에서 팀을 선택하거나 새 팀을 생성하세요.

### 3.2 팀 생성 (TeamCreateModal)

**New Team** 버튼을 클릭하여 팀을 생성합니다.

| 필드 | 설명 | 기본값 |
|------|------|--------|
| Team Name | 팀 고유 이름 (필수) | - |
| Template | 템플릿 선택 (기본/커스텀) | 없음 |
| Workers | 워커 에이전트 수 | 2 |
| Budget | 토큰 예산 | 0 (무제한) |

**템플릿 그룹:**
- **기본 템플릿**: `leader-worker`, `leader-worker-reviewer`, `parallel-workers`
- **커스텀 템플릿**: 사용자가 생성한 템플릿

---

### 3.3 칸반 보드 (TeamBoard)

5개 상태 컬럼으로 구성된 칸반 보드:

| 컬럼 | 상태 | 색상 |
|------|------|------|
| 대기 중 | `pending` | 회색 |
| 진행 중 | `in_progress` | 파란색 |
| 차단됨 | `blocked` | 주황색 |
| 완료 | `done` | 초록색 |
| 실패 | `failed` | 빨간색 |

**작업 카드에 표시되는 정보:**
- 작업 ID (앞 7자리)
- 상태 아이콘
- 제목
- 담당 에이전트 ID
- 우선순위 ★

---

### 3.4 에이전트 패널 (BoardPanel)

팀의 에이전트 목록을 표시합니다:

- 에이전트 역할 (leader/worker/reviewer)
- 현재 상태 (active/idle/offline)
- 현재 담당 작업 ID
- 미읽은 메시지 수
- **X 버튼**: 해당 에이전트 세션 즉시 종료

---

### 3.5 작업 추가

보드 상단의 **+ 버튼**으로 새 작업을 추가합니다:

| 필드 | 설명 |
|------|------|
| Title | 작업 제목 (필수) |
| Description | 상세 설명 |
| Priority | 우선순위 숫자 |
| Tags | 쉼표로 구분된 태그 |

---

### 3.6 실시간 업데이트

SSE(Server-Sent Events)를 통해 팀 상태가 실시간으로 갱신됩니다:
- 팀 변경 시 자동 업데이트 (기본 2초 간격)
- 에이전트 상태 변경 즉시 반영
- 작업 상태 변경 즉시 반영

---

### 3.7 팀 서브 실행 (TeamServeModal)

**▶️ 버튼** 클릭 후 "Board Server" 선택:
- 포트, 호스트, 갱신 간격 설정
- CLI 명령어 자동 생성 및 복사

---

### 3.8 템플릿 관리 (TemplateManagerModal)

⚙️ 아이콘 클릭:
- 커스텀 템플릿 목록 조회
- 새 템플릿 생성 (YAML 에디터)
- 기존 템플릿 편집/삭제

---

## 4. REST API 및 SSE 엔드포인트

웹챗 서버(`http://localhost:port`)가 제공하는 API입니다.

### 4.1 팀 관리 API

```http
# 팀 목록 조회
GET /api/team/list
GET /api/team/list?projectPath=/path/to/project

응답: {
  "teams": [
    {
      "name": "my-team",
      "config": { "budget": 100000, "workers": 3 },
      "taskCounts": { "pending": 2, "in_progress": 1, "done": 5 }
    }
  ]
}

---

# 팀 상세 조회
GET /api/team/:name

응답: {
  "config":      { "name": "my-team", "budget": 100000 },
  "agents":      [{ "id": "worker-1", "role": "worker", "status": "active" }],
  "tasks":       [{ "id": "abc123", "title": "...", "status": "in_progress" }],
  "cost":        { "total": 45230, "agents": { "worker-1": { "tokens": 15400 } } },
  "taskCounts":  { "pending": 2, "in_progress": 1, "done": 3 }
}

---

# 팀 생성
POST /api/team/create
Content-Type: application/json

{
  "name":        "my-team",
  "template":    "leader-worker",   // 선택
  "workers":     2,                 // 선택
  "budget":      100000,            // 선택
  "projectPath": "/path/to/project" // 선택
}

응답: 200 OK (성공) | 400/500 (오류)

---

# 팀 삭제
DELETE /api/team/:name

응답: 200 OK
```

### 4.2 작업 관리 API

```http
# 작업 목록 조회
GET /api/team/:name/tasks
GET /api/team/:name/tasks?status=in_progress&agent=worker-1

응답: { "tasks": [...] }

---

# 작업 생성
POST /api/team/:name/tasks
Content-Type: application/json

{ "title": "Implement login", "desc": "Create POST /auth/login" }

응답: { "task": { "id": "abc123", "title": "...", "status": "pending" } }

---

# 작업 업데이트
PATCH /api/team/:name/tasks/:id
Content-Type: application/json

{
  "status":   "in_progress",  // 선택
  "agent_id": "worker-1",     // 선택
  "result":   "Done"          // 선택
}

응답: { "task": { "id": "abc123", "status": "in_progress" } }
```

### 4.3 에이전트 관리 API

```http
# 에이전트 목록 조회
GET /api/team/:name/agents

응답: {
  "agents": [
    {
      "id":            "worker-1",
      "role":          "worker",
      "status":        "active",
      "current_task":  "abc12345",
      "tmux_session":  "jikime-my-team-worker-1",
      "pid":           12345
    }
  ]
}

---

# 에이전트 세션 종료
DELETE /api/team/:name/agents/:agentId

응답: 200 OK (tmux kill-session 실행)
```

### 4.4 메시지 API

```http
# 메시지 전송
POST /api/team/:name/inbox/send
Content-Type: application/json

{
  "to":   "worker-1",    // 특정 에이전트 또는 "broadcast"
  "body": "작업 우선순위를 높여주세요"
}

응답: { "message": { "id": "msg-xxx", "from": "leader", "to": "worker-1" } }
```

### 4.5 예산 API

```http
# 예산 요약 조회
GET /api/team/:name/budget

응답: {
  "total": 45230,
  "agents": {
    "worker-1": { "tokens": 15400 },
    "worker-2": { "tokens": 13300 }
  }
}
```

### 4.6 SSE 스트림 (실시간)

```http
GET /api/team/:name/events
Accept: text/event-stream

# 응답 형식 (2초마다 푸시):
data: {
  "type": "update",
  "time": "2026-03-20T10:00:00Z",
  "team": {
    "name":        "my-team",
    "leaderName":  "leader",
    "description": ""
  },
  "tasks":   [...],
  "agents":  [...],
  "members": [...],
  "taskSummary": {
    "pending":     2,
    "in_progress": 1,
    "done":        3,
    "failed":      0,
    "blocked":     0
  },
  "messages": [
    {
      "from":      "worker-1",
      "to":        "leader",
      "type":      "direct",
      "timestamp": "2026-03-20T09:59:30Z",
      "content":   "Task abc123 complete"
    }
  ]
}
```

### 4.7 템플릿 API

```http
# 템플릿 목록 조회
GET /api/template/list

응답: {
  "templates": [
    {
      "name":        "leader-worker",
      "description": "Leader coordinates, workers execute",
      "agents":      2
    }
  ]
}

---

# 특정 템플릿 조회
GET /api/template/:name
```

---

## 5. 파일 시스템 구조

```
~/.jikime/
├── teams/
│   └── <team-name>/
│       ├── config.json              # TeamConfig (팀 설정)
│       │   {
│       │     "name": "my-team",
│       │     "template": "leader-worker",
│       │     "budget": 100000,
│       │     "maxAgents": 0,
│       │     "timeoutSeconds": 0,
│       │     "createdAt": "..."
│       │   }
│       ├── webchat.json             # 웹챗 메타데이터
│       │   { "projectPath": "/path/to/project" }
│       ├── tasks/
│       │   └── <task-id>.json       # Task 파일
│       │       {
│       │         "id": "abc123",
│       │         "title": "Implement login",
│       │         "status": "in_progress",
│       │         "agent_id": "worker-1",
│       │         "priority": 2,
│       │         "dod": "...",
│       │         "created_at": "...",
│       │         "claimed_at": "..."
│       │       }
│       ├── inbox/
│       │   ├── <agent-id>/
│       │   │   └── <msg-id>.json    # Message 파일
│       │   └── event-log.jsonl      # 전체 이벤트 로그 (JSON Lines)
│       ├── registry/
│       │   └── <agent-id>.json      # AgentInfo 파일
│       │       {
│       │         "id": "worker-1",
│       │         "role": "worker",
│       │         "status": "active",
│       │         "pid": 12345,
│       │         "tmux_session": "jikime-my-team-worker-1",
│       │         "current_task": "abc123",
│       │         "last_heartbeat": "...",
│       │         "joined_at": "..."
│       │       }
│       ├── costs/
│       │   └── <agent-id>-<timestamp>.json  # CostEvent 파일
│       └── events/
│           └── (이벤트 기록)
│
├── sessions/
│   └── <team-name>/
│       └── <session-id>.json        # 팀 상태 스냅샷
│
├── plans/
│   └── <plan-id>.json               # Plan 파일
│
├── worktrees/
│   └── <team-name>/
│       └── <agent-id>/              # Git worktree 루트
│           ├── .git                 # worktree 링크
│           └── (프로젝트 파일들)
│
├── templates/
│   ├── leader-worker.yaml
│   ├── leader-worker-reviewer.yaml
│   └── parallel-workers.yaml
│
└── logs/
    └── <team-name>/
        └── <agent-id>.log           # subprocess 백엔드 로그
```

---

## 6. 환경 변수

에이전트가 스폰될 때 자동으로 주입되는 환경변수:

| 변수 | 설명 | 예시 |
|------|------|------|
| `JIKIME_AGENT_ID` | 에이전트 고유 ID | `worker-1`, `agent-a1b2c3d4` |
| `JIKIME_TEAM_NAME` | 팀 이름 | `my-team` |
| `JIKIME_ROLE` | 에이전트 역할 | `leader`, `worker`, `reviewer` |
| `JIKIME_DATA_DIR` | 데이터 디렉토리 경로 | `~/.jikime` |
| `JIKIME_WORKTREE_PATH` | Git worktree 경로 (설정된 경우) | `~/.jikime/worktrees/...` |
| `JIKIME_SPAWN_TIME` | 스폰 시각 (ISO 8601) | `2026-03-20T10:00:00Z` |

에이전트 CLI 명령어 내에서 이 변수들을 활용합니다:

```bash
# 에이전트 프로세스 내부에서 CLI 사용 예시
jikime team tasks claim "$JIKIME_TEAM_NAME" <task-id> --agent "$JIKIME_AGENT_ID"
jikime team tasks complete "$JIKIME_TEAM_NAME" <task-id> --agent "$JIKIME_AGENT_ID"
jikime team budget report "$JIKIME_TEAM_NAME" --agent "$JIKIME_AGENT_ID" --input-tokens 1234
```

**외부 설정:**

```bash
# 데이터 디렉토리 재정의
export JIKIME_DATA_DIR=/custom/path

# Claude 바이너리 경로 재정의
export CLAUDE_PATH=/usr/local/bin/claude
```

---

## 7. 팀 워크플로우 패턴

### 7.1 기본 패턴: Leader-Worker

```
Leader
  ├── 전체 목표 분석
  ├── 작업 분해 → tasks 생성
  ├── Worker들이 claim하고 처리
  ├── 결과 검토 및 통합
  └── 완료 보고
```

```bash
# 1. 팀 생성
jikime team create auth-team --budget 100000

# 2. 초기 작업 생성
jikime team tasks create auth-team "Design API schema" --priority 3
jikime team tasks create auth-team "Implement login endpoint" --priority 2
jikime team tasks create auth-team "Write unit tests" --priority 1

# 3. 에이전트 스폰
jikime team spawn auth-team --role leader --agent-id leader
jikime team spawn auth-team --role worker --agent-id worker-1
jikime team spawn auth-team --role worker --agent-id worker-2

# 4. 모니터링
jikime team board live auth-team
```

---

### 7.2 고급 패턴: 계획 승인 워크플로우

```bash
# Worker 관점 (Claude 에이전트 내부)
jikime team plan submit "$JIKIME_TEAM_NAME" \
  --title "Database Implementation Plan" \
  --file implementation-plan.md

# Leader 관점 (계획 검토)
jikime team plan list --team auth-team
jikime team plan approve plan-abc12345
# 또는
jikime team plan reject plan-abc12345 --reason "Need distributed approach"
```

---

### 7.3 Git Worktree 격리 패턴

```bash
# 워크트리 격리로 팀 시작
jikime team launch --template leader-worker \
  --goal "implement new feature" \
  --name feature-team \
  --worktree

# 각 에이전트:
# - 독립적인 git 브랜치에서 작업
# - 브랜치명: jikime-feature-team-<agent-id>
# - 경로: ~/.jikime/worktrees/feature-team/<agent-id>/

# 작업 완료 후 병합
jikime team workspace merge feature-team --agent worker-1 --target develop
jikime team workspace merge feature-team --agent worker-2 --target develop
```

---

### 7.4 CI/CD 통합 패턴

```bash
#!/bin/bash
# ci-team.sh

# 팀 생성 및 시작
jikime team launch --template leader-worker \
  --goal "Fix security vulnerabilities in $PR_TITLE" \
  --name "ci-$PR_NUMBER" \
  --backend subprocess \
  --budget 50000

# 모든 작업 완료 대기 (최대 1시간)
jikime team tasks wait "ci-$PR_NUMBER" --timeout 3600
EXIT_CODE=$?

# 예산 보고서 출력
jikime team budget show "ci-$PR_NUMBER"

# 정리
jikime team stop "ci-$PR_NUMBER"

exit $EXIT_CODE
```

---

## 8. Git Worktree 격리

### 작동 원리

```
메인 레포지토리
├── .git/
├── src/
└── ...
    │
    └── (worktree 링크)
        │
        ├── ~/.jikime/worktrees/my-team/worker-1/   ← worker-1 작업공간
        │   ├── .git    (→ 메인 .git/worktrees/worker-1)
        │   ├── src/    (브랜치: jikime-my-team-worker-1)
        │   └── ...
        │
        └── ~/.jikime/worktrees/my-team/worker-2/   ← worker-2 작업공간
            ├── .git    (→ 메인 .git/worktrees/worker-2)
            ├── src/    (브랜치: jikime-my-team-worker-2)
            └── ...
```

### 브랜치 명명 규칙

```
jikime-<team-name>-<agent-id>

예시:
  jikime-auth-team-worker-1
  jikime-auth-team-worker-2
  jikime-auth-team-leader
```

### Worktree 사용 이점

1. **병렬 코드 작성**: 여러 에이전트가 동시에 다른 파일/기능 작업
2. **충돌 방지**: 각 에이전트가 독립 브랜치에서 작업
3. **진행 보존**: 각 에이전트의 변경사항이 즉시 커밋됨
4. **개별 병합**: 각 작업별로 메인 브랜치에 독립적으로 병합

### Claude 세션 경로

tmux로 에이전트를 `-c <worktreePath>` 옵션으로 시작하면 Claude 세션이 올바른 프로젝트 경로에 저장됩니다:

```
~/.claude/projects/<worktree-path-hash>/
```

---

## 9. GitHub Issues / Harness 통합

Webchat의 **Git 탭 → Issues** 섹션에서 GitHub Issues를 AI가 자동으로 처리합니다.

### 9.1 설정 요구사항

1. **GitHub PAT (Personal Access Token)**: 사이드바 설정 → Git PAT 입력
2. **WORKFLOW.md**: 프로젝트 루트에 Harness 설정 파일 필요

### 9.2 GitHub Issues 레이블 시스템

| 레이블 | 의미 | AI 동작 |
|--------|------|---------|
| `jikime-todo` | AI가 처리해야 할 이슈 | 자동 감지 및 처리 |
| `jikime-done` | AI가 처리 완료한 이슈 | 자동으로 추가됨 |

### 9.3 Harness (자동 폴링)

WORKFLOW.md를 설정하면 Harness가 주기적으로 `jikime-todo` 레이블 이슈를 감지하여 자동 처리합니다.

```bash
# Harness 상태 확인 API
GET /api/harness/status?projectPath=/path/to/project

# Harness 시작
POST /api/harness/start
{ "projectPath": "/path/to/project" }

# Harness 중지
DELETE /api/harness/stop?projectPath=/path/to/project
```

**Webchat에서:**
1. Issues 탭 → ⚡ **Start** 버튼
2. Harness가 실행 중이면 `🔵 running` 배너 표시
3. **Stop** 버튼으로 중지

### 9.4 이슈 수동 처리

특정 이슈를 즉시 AI로 처리:

1. Issues 목록에서 이슈 클릭
2. 우측 패널의 **▶ Process with AI** 버튼 클릭
3. 처리 로그가 실시간으로 표시됨:
   - 🚀 시작 메시지 (상태 표시)
   - 🔧 ToolName (도구 사용 — 파일 경로/명령어 포함)
   - Claude 텍스트 (마크다운 렌더링)
   - ✅ 완료 / ❌ 오류 (상태 표시)

### 9.5 처리 로그 포맷

Issues 탭의 처리 로그는 Chat 탭과 동일한 형식으로 표시됩니다:

```
🚀 이슈 #42 처리 시작: Add rate limiting
     ↓ 상태 메시지 (작은 회색 텍스트)

🔧 Read: src/middleware/auth.js
     ↓ 도구 버블 (파일 경로 표시)

Claude의 텍스트 응답...
     ↓ 오렌지색 "C" 아바타 + 마크다운 렌더링

🔧 Edit: src/middleware/rate-limit.js
     ↓ 도구 버블 (편집 파일 표시)

✅ 이슈 #42 처리 완료
     ↓ 상태 메시지
```

---

## 10. 템플릿 시스템

### 10.1 기본 제공 템플릿

#### `leader-worker`

리더 1명 + 워커 N명 구조.

```yaml
name: leader-worker
description: "Leader coordinates tasks, workers execute"
agents:
  - id: leader
    role: leader
    auto_spawn: true
    task: |
      You are the team leader. Analyze the goal, create tasks,
      and coordinate workers.
  - id: worker-1
    role: worker
    auto_spawn: true
    task: |
      You are a worker. Check available tasks, claim one, and complete it.
tasks:
  - subject: "Analyze requirements"
    description: "Break down the goal into concrete tasks"
    owner: leader
default_budget: 100000
```

#### `leader-worker-reviewer`

리더 + 워커 + 코드 리뷰어 구조.

#### `parallel-workers`

리더 없이 워커들이 병렬로 동작하는 구조.

---

### 10.2 커스텀 템플릿 생성

Webchat **⚙️ → Template Manager** 또는 직접 파일 생성:

```yaml
# ~/.jikime/templates/my-custom-template.yaml
name: my-custom-template
description: "My specialized team structure"
version: "1.0.0"

agents:
  - id: architect
    role: leader
    description: "System design and task decomposition"
    auto_spawn: true
    task: |
      You are the system architect. Design the solution and
      create specific tasks for each team member.

  - id: backend-dev
    role: worker
    description: "Backend development specialist"
    auto_spawn: true
    task: |
      You are a backend developer. Focus on API implementation,
      database design, and server-side logic.

  - id: frontend-dev
    role: worker
    description: "Frontend development specialist"
    auto_spawn: true
    task: |
      You are a frontend developer. Focus on UI components,
      state management, and user experience.

  - id: qa-engineer
    role: reviewer
    description: "Quality assurance and testing"
    auto_spawn: true
    task: |
      You are a QA engineer. Write tests, review code quality,
      and ensure all requirements are met.

tasks:
  - subject: "Kickoff meeting"
    description: "Architect defines the implementation plan"
    owner: architect
  - subject: "Project setup"
    description: "Setup project structure and dependencies"

default_budget: 200000
default_max_agents: 4
```

---

## 11. 실전 예시

### 11.1 SaaS 기능 개발

```bash
# 새 결제 기능 개발
jikime team launch \
  --template leader-worker-reviewer \
  --goal "Implement Stripe payment integration with webhook support" \
  --name payment-team \
  --worktree \
  --budget 150000

# 진행 상황 실시간 모니터링
jikime team board serve payment-team --port 8080
# 브라우저: http://localhost:8080

# 수동 작업 추가 (실행 중에도 가능)
jikime team tasks create payment-team "Add retry logic for failed payments" \
  --priority 2 \
  --depends-on abc12345

# 완료 대기
jikime team tasks wait payment-team --timeout 7200

# 결과 확인
jikime team budget show payment-team
jikime team board show payment-team

# 워크트리 병합
jikime team workspace merge payment-team --agent worker-1 --target main --cleanup
jikime team workspace merge payment-team --agent reviewer --target main --cleanup
```

---

### 11.2 버그 수정 팀

```bash
# 크리티컬 버그 수정
jikime team create bugfix-team --budget 50000

# 버그 분석 작업 생성
jikime team tasks create bugfix-team \
  "Investigate memory leak in connection pool" \
  --priority 5 \
  --dod "Root cause identified and fixed, no memory growth over 1 hour"

# 에이전트 스폰
jikime team spawn bugfix-team \
  --role worker \
  --agent-id debugger \
  --prompt "You are a senior debugging engineer. Investigate the memory leak in src/db/pool.js"

# 모니터링
jikime team board live bugfix-team
```

---

### 11.3 코드 리뷰 자동화

```bash
# PR 코드 리뷰 팀
jikime team launch \
  --template parallel-workers \
  --goal "Review PR #$PR_NUMBER: security, performance, code quality" \
  --name "review-pr-$PR_NUMBER" \
  --backend subprocess

# 완료 대기 후 결과 수집
jikime team tasks wait "review-pr-$PR_NUMBER" --timeout 1800
jikime team board show "review-pr-$PR_NUMBER" --json > review-results.json
```

---

### 11.4 에이전트 프롬프트에서 CLI 활용

에이전트(Claude Code)는 JIKIME CLI를 직접 호출하여 팀과 상호작용합니다:

```bash
# Claude 에이전트의 시스템 프롬프트 예시
You are a worker agent in team $JIKIME_TEAM_NAME.
Your agent ID is $JIKIME_AGENT_ID.

## Workflow:
1. Check available tasks:
   $ jikime team tasks list $JIKIME_TEAM_NAME --status pending

2. Claim a task:
   $ jikime team tasks claim $JIKIME_TEAM_NAME <task-id> --agent $JIKIME_AGENT_ID

3. Work on the task using your coding tools

4. Commit your progress (if using worktree):
   $ jikime team workspace checkpoint $JIKIME_TEAM_NAME --agent $JIKIME_AGENT_ID

5. Complete the task:
   $ jikime team tasks complete $JIKIME_TEAM_NAME <task-id> \
     --agent $JIKIME_AGENT_ID \
     --result "What was accomplished"

6. Repeat from step 1 until no tasks remain
```

---

## 12. 트러블슈팅

### tmux 세션을 찾을 수 없을 때

```bash
# 모든 jikime tmux 세션 확인
tmux ls | grep jikime

# 수동으로 세션에 연결
tmux attach -t jikime-my-team-worker-1

# 세션 강제 종료
tmux kill-session -t jikime-my-team-worker-1
```

---

### 에이전트가 offline으로 표시될 때

```bash
# 에이전트 프로세스 확인
jikime team status my-team

# registry에서 에이전트 정보 확인
cat ~/.jikime/teams/my-team/registry/worker-1.json

# 에이전트 재스폰
jikime team spawn my-team --role worker --agent-id worker-1
```

---

### worktree 생성 오류

```bash
# 기존 worktree 목록 확인
git worktree list

# 손상된 worktree 정리
git worktree prune

# 수동으로 worktree 제거
jikime team workspace cleanup my-team --agent worker-1
```

---

### 예산 초과 시

```bash
# 현재 예산 상태 확인
jikime team budget show my-team

# 예산 증가
jikime team budget set my-team 200000

# 특정 에이전트 중지
tmux kill-session -t jikime-my-team-worker-2
```

---

### 작업이 stuck (in_progress) 상태일 때

```bash
# 작업을 pending으로 복구
jikime team tasks update my-team <task-id> --status pending

# 다른 에이전트가 재claim
jikime team tasks claim my-team <task-id> --agent worker-2
```

---

### 웹챗에서 팀이 표시되지 않을 때

1. 팀이 현재 프로젝트 경로와 연결되어 있는지 확인
2. `~/.jikime/teams/<team-name>/webchat.json`에 `projectPath` 설정 확인
3. API 직접 확인: `curl http://localhost:3000/api/team/list?projectPath=/your/project`

---

## 관련 문서

- [agents.md](agents.md) — Claude Code 에이전트 일반 가이드
- [agents-team.md](agents-team.md) — Claude Code Agent Teams (실험적 기능)
- [worktree.md](worktree.md) — Git Worktree 워크플로우
- [harness-workflow.md](harness-workflow.md) — Harness Engineering 워크플로우
- [hooks.md](hooks.md) — Claude Code 훅 설정
- [webchat/usage.md](webchat/usage.md) — Webchat 사용 가이드
