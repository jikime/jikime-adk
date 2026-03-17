---
name: jikime-harness-work
description: Harness Engineering work orchestrator — executes Plans.md tasks through the cc:WIP → cc:DONE lifecycle with Worker, Reviewer, and Scaffolder agents
version: 1.0.0
category: harness
tags: ["harness", "work", "implementation", "cc:WIP", "cc:DONE", "worker", "plans.md", "harness-engineering"]
triggers:
  keywords:
    - "harness-work"
    - "harness work"
    - "태스크 구현"
    - "작업 시작"
    - "harness 구현"
    - "cc:WIP"
    - "implement task"
  phases: ["run"]
  agents: ["orchestrator", "harness-worker", "harness-reviewer", "harness-scaffolder"]
  languages: []
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000
user-invocable: true
context: fork
agent: general-purpose
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
  - TodoWrite
  - Task
---

# Harness Work — 태스크 구현 오케스트레이터

## Quick Reference

Plans.md의 태스크를 `cc:WIP → cc:DONE` 라이프사이클로 실행합니다. Scaffolder(분석) → Worker(구현) → Reviewer(검증) 3-에이전트 팀을 오케스트레이션합니다.

**사용법:**
```
/jikime:harness-work 1.2              → 태스크 1.2 단독 실행
/jikime:harness-work 1.2 1.3         → 태스크 1.2, 1.3 병렬 실행
/jikime:harness-work --auto           → Plans.md에서 TODO 태스크 자동 선택
/jikime:harness-work --mode breezing  → 4+ 태스크 Breezing Mode 강제
```

---

## 실행 흐름

### Phase 0: Plans.md 로드 및 검증

```bash
# Plans.md 존재 확인
ls Plans.md || exit 1

# 태스크 파싱
grep "cc:TODO\|cc:WIP" Plans.md
```

**검증 항목:**
- Plans.md 존재 여부
- 지정된 태스크 ID 유효성
- 의존성 태스크 완료 여부 (cc:DONE 또는 pm:OK)

```
의존성 미완료 시:
⚠️ Task 1.2는 Task 1.1 (현재: cc:TODO) 완료 후 실행 가능합니다.
   1.1을 먼저 실행하시겠습니까?
```

### Phase 1: Mode 선택 (Scaffolder)

**Use the harness-scaffolder subagent** to select execution mode:

```
Scaffolder: select-mode 분석
  → TODO 태스크 수 카운트
  → 의존성 그래프 분석
  → 병렬 실행 가능 태스크 그룹 도출
```

| 조건 | Mode |
|------|------|
| 태스크 1개 | Solo |
| 태스크 2–3개, 독립적 | Parallel |
| 태스크 4개 이상 | Breezing |

### Phase 2: 분석 및 스캐폴딩 (Scaffolder)

**Use the harness-scaffolder subagent** for each task:

```
Scaffolder: analyze task ${TASK_ID}
  → 관련 파일 탐색
  → 코드베이스 패턴 파악
  → 구현 접근법 제안
  → 필요 시 파일 스텁 생성
```

### Phase 3: 구현 (Worker)

#### Solo Mode

**Use the harness-worker subagent**:

```
Worker: implement task 1.2
  → Plans.md: cc:TODO → cc:WIP
  → 구현
  → DoD 검증
  → Plans.md: cc:WIP → cc:DONE [hash]
```

#### Parallel Mode

두 태스크를 동시에 처리 (단, 같은 파일 충돌 없을 때):

```
Worker A: implement task 1.2 (worktree: harness/task-1.2)
Worker B: implement task 1.3 (worktree: harness/task-1.3)
  → 동시 실행
  → 각각 cc:WIP → cc:DONE
```

**주의:** 같은 파일을 수정하는 태스크는 순차 실행

#### Breezing Mode (4+ 태스크)

```
Scaffolder: 태스크를 독립 그룹으로 분류
  Group A: [1.1, 1.2] (auth 도메인)
  Group B: [1.3, 1.4] (DB 도메인)

Worker A: Group A 순차 실행
Worker B: Group B 순차 실행
  → 병렬 처리
```

### Phase 4: 리뷰 (Reviewer)

각 cc:DONE 태스크에 대해:

**Use the harness-reviewer subagent**:

```
Reviewer: review task 1.2 (commit: abc1234)
  → 4관점 리뷰 (보안/성능/품질/DoD)
  → verdict: approve | warn | block
```

**결과 처리:**

| Verdict | Action |
|---------|--------|
| `approve` | Plans.md: cc:DONE → pm:REVIEW 요청 |
| `warn` | 경고 목록과 함께 pm:REVIEW 요청 (사용자가 판단) |
| `block` | Worker에게 피드백 → 재구현 (최대 2회) |

### Phase 5: 사용자 알림

```
✅ 태스크 1.2 완료

구현: abc1234 (feat: [1.2] JWT token generation)
리뷰: ✅ Approve (보안/성능/품질/DoD 모두 통과)

Plans.md에 pm:REVIEW가 표시됐습니다.
/jikime:harness-review 1.2 로 최종 승인하거나 직접 검토 후 pm:OK로 변경하세요.
```

---

## Worktree 격리

병렬 실행 시 각 태스크는 독립 git worktree에서 실행됩니다:

```bash
# Worker가 시작 전 자동으로 실행
git worktree add ../harness-task-1.2 -b harness/task-1.2

# 작업 완료 후
git worktree remove ../harness-task-1.2
```

**명명 규칙:** `harness/task-{TASK_ID}` (예: `harness/task-1.2`)

---

## Plans.md 마커 전환 순서

```
cc:TODO
  ↓ (Worker 시작)
cc:WIP
  ↓ (Worker 구현 완료 + DoD 검증)
cc:DONE [hash]
  ↓ (Reviewer approve/warn)
pm:REVIEW
  ↓ (사용자 승인 또는 /jikime:harness-review)
pm:OK
```

---

## 에러 처리

| 상황 | 처리 |
|------|------|
| 의존성 미완료 | 경고 후 사용자 확인 요청 |
| DoD 미충족 (3회 시도) | `blocked:<이유>` 마킹 + 사용자 알림 |
| Reviewer block (2회 재시도) | `blocked:review-rejected` 마킹 |
| Worktree 충돌 | 순차 실행으로 자동 전환 |
| Plans.md 파싱 오류 | 실행 중단 + 오류 상세 출력 |

---

## 통합 포인트

| 스킬/커맨드 | 연관 방식 |
|-------------|-----------|
| `jikime-harness-plan` | Plans.md 생성 → harness-work 실행 |
| `jikime-harness-review` | pm:REVIEW → pm:OK 최종 승인 |
| `jikime-harness-sync` | 완료 후 동기화 및 레트로스펙티브 |
| `/jikime:2-run` | harness-work 기반 실행 |

---

Version: 1.0.0
Status: Active
Last Updated: 2026-03-15
