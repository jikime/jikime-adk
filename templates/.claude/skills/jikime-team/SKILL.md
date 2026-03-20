---
name: jikime-team
description: >
  Use this skill when the user asks to "jikime team을 사용해줘", "팀 만들어줘", "에이전트 생성해줘",
  "멀티 에이전트로 작업해줘", "병렬로 에이전트 실행해줘", "여러 에이전트 조율해줘",
  "작업을 여러 에이전트에게 분배해줘", "team status 확인해줘", "task 만들어줘",
  "inbox 확인해줘", or mentions "jikime team", "multi-agent coordination",
  "spawn agents", "team tasks", "agent inbox", "task board", "team leader",
  "team worker". Also trigger when the scope of work is large enough to benefit
  from splitting into parallel subtasks — for example "전체 코드베이스 리팩토링",
  "여러 기능 동시에 구현", "대규모 분석", "full-stack app 만들어줘".
  Provides comprehensive guidance for using the jikime team CLI to orchestrate
  multi-agent teams with task management, messaging, and monitoring.
license: Apache-2.0
user-invocable: false
metadata:
  version: "1.0.0"
  category: "team"
  updated: "2026-03-19"
  tags: "team, multi-agent, orchestration, tasks, inbox, board, spawn"
  related-skills: "jikime-team-leader, jikime-team-worker, jikime-team-reviewer"
---

# jikime team — Multi-Agent Orchestration

`jikime team`은 여러 Claude 에이전트를 팀으로 조율하는 CLI 도구예요.
리더 에이전트가 작업을 분해해서 워커에게 배분하고, 파일 기반 메시지 시스템으로 소통합니다.

모든 상태는 `~/.jikime/teams/<team-name>/` 에 JSON 파일로 저장돼요.

---

## 핵심 개념

| 개념 | 설명 |
|------|------|
| **Team** | 이름 있는 에이전트 그룹. 리더(1명) + 워커(N명) + 리뷰어(선택) |
| **Tasks** | 공유 작업 보드. `pending → in_progress → done` 상태 전환 |
| **Inbox** | 에이전트 간 파일 기반 메시지 큐 |
| **Board** | 터미널 칸반 대시보드 (tmux 기반) |
| **Spawn** | Claude Code 프로세스를 새 tmux 창에서 에이전트로 실행 |

---

## 빠른 시작

### 1. 팀 생성

```bash
jikime team init my-team --desc "프로젝트 연구팀"
```

### 2. 에이전트 스폰 (tmux 창에서 Claude 실행)

```bash
# 리더 에이전트 스폰
jikime team spawn my-team leader \
  --task "목표를 분석하고 작업을 분해해서 워커에게 배분하세요"

# 워커 에이전트 스폰
jikime team spawn my-team worker-1 \
  --task "팀의 작업을 수행하세요"

jikime team spawn my-team worker-2 \
  --task "팀의 작업을 수행하세요"
```

### 3. 작업 생성

```bash
# 리더가 작업 분해 후 생성
jikime team tasks create my-team "기능 A 구현" \
  --desc "로그인 API 개발" \
  --dod "단위 테스트 통과, PR 생성" \
  --priority 10

jikime team tasks create my-team "기능 B 구현" \
  --desc "회원가입 API 개발" \
  --dod "단위 테스트 통과" \
  --priority 8

# 의존성이 있는 작업 (A와 B가 완료되어야 시작 가능)
jikime team tasks create my-team "통합 테스트" \
  --depends-on <task-a-id>,<task-b-id>
```

### 4. 에이전트 화면 모니터링

```bash
# 모든 에이전트 창을 tmux 타일 뷰로 보기
jikime team board attach my-team
```

### 5. 작업 현황 확인

```bash
jikime team status my-team
jikime team tasks list my-team
jikime team tasks list my-team --status pending
jikime team tasks list my-team --status done
```

---

## 작업(Task) 라이프사이클

```
pending → in_progress (claim) → done (complete)
           ↓
         blocked  (의존성 미충족)
           ↓
         failed   (실패)
```

### 워커가 작업 처리하는 방법

```bash
# 1. 대기 중인 작업 목록 확인
jikime team tasks list my-team --status pending

# 2. 작업 클레임 (다른 에이전트가 먼저 가져가지 못하도록 atomic 락)
jikime team tasks claim my-team <task-id> --agent worker-1

# 3. 작업 세부 정보 읽기
jikime team tasks get my-team <task-id>

# 4. 작업 완료
jikime team tasks complete my-team <task-id> \
  --agent worker-1 \
  --result "로그인 API 구현 완료. 테스트 통과. PR #42 생성."
```

### 관리자 오버라이드

```bash
# 강제로 상태 변경 (소유권 체크 없이)
jikime team tasks update my-team <task-id> --status pending
jikime team tasks update my-team <task-id> --status done --result "수동 완료"
jikime team tasks update my-team <task-id> --status failed
```

---

## 메시지(Inbox)

```bash
# 특정 에이전트에게 메시지 전송
jikime team inbox send my-team worker-1 "task abc123에 집중해주세요"

# 전체 브로드캐스트
jikime team inbox broadcast my-team "새 제약: 모든 API 응답은 JSON이어야 함"

# 메시지 수신 (소비 — 읽으면 삭제됨)
jikime team inbox receive my-team
jikime team inbox receive my-team --agent worker-1

# 메시지 미리 보기 (삭제 안 됨)
jikime team inbox peek my-team --agent leader
```

---

## 모니터링

```bash
# 팀 전체 현황 (에이전트 + 작업 요약)
jikime team status my-team

# 작업 보드 보기
jikime team board show my-team

# tmux 타일 뷰 (모든 에이전트 창을 한 화면에)
jikime team board attach my-team

# 모든 팀 개요
jikime team board overview

# 완료될 때까지 대기
jikime team tasks wait my-team
jikime team tasks wait my-team --timeout 300 --interval 10
```

---

## 수명 주기 (Lifecycle)

```bash
# 팀 종료 신호 전송
jikime team lifecycle shutdown my-team

# 에이전트 레지스트리 상태 확인
jikime team status my-team

# 팀 데이터 완전 삭제
jikime team lifecycle destroy my-team
```

---

## 워크스페이스 (Git Worktree 격리)

에이전트별 격리된 git worktree를 제공해요 (충돌 방지):

```bash
# 워크스페이스 설정
jikime team workspace setup my-team worker-1

# 작업 후 체크포인트
jikime team workspace checkpoint my-team --agent worker-1

# 메인 브랜치로 병합
jikime team workspace merge my-team --agent worker-1 --target main
```

---

## 템플릿으로 팀 실행

사전 정의된 팀 구성을 사용하면 더 쉬워요:

```bash
# 사용 가능한 템플릿 목록
jikime template list

# 템플릿으로 팀 생성 및 스폰
jikime team serve hedge-fund --goal "AI 반도체 섹터 투자 연구"

# 커스텀 변수 전달
jikime team serve my-template \
  --var goal="분석 목표" \
  --var team_name="my-research-team"
```

---

## 명령어 전체 참조

| 명령 그룹 | 주요 명령 |
|-----------|-----------|
| `team init` | 팀 초기화 |
| `team spawn` | 에이전트 스폰 (tmux/subprocess) |
| `team status` | 팀 상태 확인 |
| `team tasks create/list/get/claim/complete/update` | 작업 관리 |
| `team tasks wait` | 완료 대기 |
| `team inbox send/broadcast/receive/peek` | 메시지 관리 |
| `team board show/attach/overview` | 모니터링 대시보드 |
| `team workspace setup/checkpoint/merge` | 워크스페이스 관리 |
| `team lifecycle shutdown/destroy` | 팀 수명 주기 |
| `team serve` | 템플릿으로 전체 팀 실행 |

---

## Claude 인터랙티브 모드에서 사용하는 방법

Claude가 직접 팀 오케스트레이터 역할을 할 때:

1. **목표 분석**: 사용자의 목표를 이해하고 독립 작업으로 분해
2. **팀 초기화**: `jikime team init <team-name>` 실행
3. **작업 생성**: `jikime team tasks create` 로 각 서브태스크 생성
4. **에이전트 스폰**: `jikime team spawn` 으로 Claude 에이전트 실행
5. **모니터링**: `jikime team board attach` 또는 `jikime team status` 로 진행 추적
6. **조율**: inbox를 통해 에이전트에게 지시 전달
7. **완료**: 모든 작업 완료 후 `jikime team lifecycle shutdown` 실행

### 예시: Claude가 혼자 오케스트레이터로 동작

```bash
# Claude가 리더 역할로 직접 팀 조율
jikime team init research-team --desc "리서치 프로젝트"

# 작업 생성
jikime team tasks create research-team "기술 분석" --desc "기술 스택 조사"
jikime team tasks create research-team "시장 분석" --desc "경쟁사 조사"

# 에이전트 스폰 (별도 Claude 프로세스)
jikime team spawn research-team tech-analyst \
  --task "기술 스택을 분석하고 결과를 tasks complete로 보고하세요"

jikime team spawn research-team market-analyst \
  --task "시장과 경쟁사를 분석하고 결과를 tasks complete로 보고하세요"

# 완료 대기
jikime team tasks wait research-team

# 결과 수집
jikime team tasks list research-team --status done
```

---

## 에이전트 환경 변수

스폰된 에이전트에 자동으로 설정되는 환경 변수:

```
JIKIME_TEAM_NAME   — 팀 이름
JIKIME_AGENT_ID    — 에이전트 ID
JIKIME_ROLE        — leader / worker / reviewer
JIKIME_DATA_DIR    — 데이터 루트 (~/.jikime)
JIKIME_PLAN_GATE   — "1"이면 리더 승인 후 작업 시작
JIKIME_WORKTREE_PATH — (격리 모드) git worktree 경로
```

---

## 중요 참고사항

- `inbox receive`는 메시지를 **소비**(삭제)합니다. 비파괴 읽기는 `inbox peek` 사용
- `tasks claim`은 **원자적** — 두 에이전트가 동시에 같은 작업을 가져가지 못함
- `tasks complete`은 해당 작업에 **의존하는 blocked 작업을 자동으로 pending으로 전환**
- `board attach`는 현재 tmux 세션에 있어야 작동 (tmux 외부에서는 새 세션 생성)
- 모든 파일 쓰기는 **tmp+rename 원자적 방식**으로 데이터 손상 방지
