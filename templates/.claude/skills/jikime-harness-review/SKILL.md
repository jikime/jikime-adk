---
name: jikime-harness-review
description: Harness Engineering review orchestrator — manages pm:REVIEW → pm:OK flow with 4-perspective code review and user approval gate
version: 1.0.0
category: harness
tags: ["harness", "review", "pm:REVIEW", "pm:OK", "code-review", "4-perspective", "harness-engineering"]
triggers:
  keywords:
    - "harness-review"
    - "harness review"
    - "리뷰 승인"
    - "코드 리뷰"
    - "pm:REVIEW"
    - "pm:OK"
    - "review task"
    - "approve task"
  phases: ["review"]
  agents: ["orchestrator", "harness-reviewer"]
  languages: []
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~4000
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

# Harness Review — 리뷰 오케스트레이터

## Quick Reference

`cc:DONE` 태스크에 대한 4관점 코드 리뷰를 실행하고, `pm:REVIEW → pm:OK` 전환을 관리합니다.

**사용법:**
```
/jikime:harness-review 1.2          → 태스크 1.2 리뷰
/jikime:harness-review 1.2 1.3      → 복수 태스크 순차 리뷰
/jikime:harness-review --all        → Plans.md의 모든 pm:REVIEW 태스크 리뷰
/jikime:harness-review --approve 1.2 → 리뷰 없이 pm:OK로 직접 승인 (긴급)
```

---

## 4관점 리뷰 프레임워크

Reviewer Agent가 4개 관점에서 독립적으로 평가합니다:

| 관점 | 중점 항목 | Verdict 기준 |
|------|----------|-------------|
| 🛡️ **보안** | 하드코딩 시크릿, 인젝션, XSS | CRITICAL → block |
| ⚡ **성능** | O(n²), N+1, 불필요한 재렌더링 | HIGH → warn |
| 🏗️ **코드 품질** | SRP, 함수 크기, 중첩 깊이 | 종합 평가 |
| ✅ **DoD 준수** | 완료 기준 실제 충족 여부 | 미충족 → block |

---

## 실행 흐름

### Phase 0: 리뷰 대상 확인

```bash
# Plans.md에서 cc:DONE 태스크 확인
grep "cc:DONE" Plans.md

# pm:REVIEW 이미 표시된 태스크 확인
grep "pm:REVIEW" Plans.md
```

### Phase 1: Reviewer Agent 실행

**Use the harness-reviewer subagent**:

```
Reviewer: review task ${TASK_ID}
  - commit_hash: ${HASH from cc:DONE marker}
  - task_dod: ${DoD from Plans.md}
  → 4관점 리뷰 실행
  → verdict: approve | warn | block
```

### Phase 2: Verdict 처리

#### approve — 자동 pm:REVIEW 전환

```
Plans.md: cc:DONE [abc1234] → pm:REVIEW
commit: "chore(plans): task 1.2 → pm:REVIEW (reviewer approved)"
```

사용자에게 리뷰 결과 요약 제공:
```
✅ Task 1.2 리뷰 완료 — Approve

🛡️ 보안: PASS
⚡ 성능: PASS
🏗️ 품질: PASS (minor: 함수 1개 55줄, 허용 범위)
✅ DoD: PASS (모든 기준 충족)

Plans.md가 pm:REVIEW로 업데이트됐습니다.
최종 승인하려면: pm:OK 로 변경하거나 아래 버튼을 확인하세요.
```

#### warn — 경고와 함께 pm:REVIEW 전환

경고 내용을 사용자가 확인 후 결정:

```
⚠️ Task 1.2 리뷰 완료 — Warning

🛡️ 보안: PASS
⚡ 성능: WARN — getUserList()가 O(n²), 사용자 1000명 이상 시 문제 가능
🏗️ 품질: PASS
✅ DoD: PASS

경고 사항이 있습니다. 진행하시겠습니까?
1. pm:REVIEW로 전환 (경고 인지 후 진행)
2. Worker에게 수정 요청
```

#### block — Worker 재구현 요청

```
❌ Task 1.2 리뷰 차단

🛡️ 보안: BLOCK — API 키 하드코딩 발견 (src/api/client.ts:42)
   const apiKey = "sk-abc123"  → 환경변수로 이동 필요

Worker에게 수정 피드백을 전달합니다.
(재시도 1/2)
```

Worker에게 블록 피드백 전달 → 재구현 → 재리뷰 (최대 2회).

2회 후에도 block이면:
```
Plans.md: cc:DONE → blocked:review-rejected
사용자에게 수동 개입 요청
```

### Phase 3: 사용자 최종 승인 (pm:OK)

`pm:REVIEW` 상태에서 사용자가 직접 확인 후:

**옵션 A: 슬래시 커맨드로 승인**
```
/jikime:harness-review --approve 1.2
→ Plans.md: pm:REVIEW → pm:OK
→ commit: "chore(plans): task 1.2 → pm:OK"
```

**옵션 B: Plans.md 직접 수정**
```
| 1.2  | ... | pm:OK |   ← 직접 변경
plans-watcher hook이 자동 감지하여 확인
```

---

## pm:REVIEW 태스크 일괄 처리

```
/jikime:harness-review --all

Plans.md의 모든 pm:REVIEW 태스크:
  1.2 (commit: abc1234) — 리뷰 대기
  1.3 (commit: def5678) — 리뷰 대기
  2.1 (commit: ghi9012) — 리뷰 대기

순차적으로 리뷰합니다...
```

---

## Plans.md 마커 전환 (이 스킬 범위)

```
cc:DONE [hash]
  ↓ (Reviewer approve)
pm:REVIEW          ← harness-review가 자동 전환
  ↓ (사용자 확인)
pm:OK              ← --approve 또는 직접 편집
```

---

## 리뷰 결과 커밋 형식

```bash
# approve/warn 시
git commit -m "chore(plans): task ${TASK_ID} → pm:REVIEW

Review verdict: ${VERDICT}
Reviewer findings:
- Security: ${SECURITY_RESULT}
- Performance: ${PERF_RESULT}
- Quality: ${QUALITY_RESULT}
- DoD: ${DOD_RESULT}"

# pm:OK 승인 시
git commit -m "chore(plans): task ${TASK_ID} → pm:OK

User approved after review."
```

---

## 에러 처리

| 상황 | 처리 |
|------|------|
| 태스크가 cc:DONE 아님 | 에러: "harness-work 먼저 실행하세요" |
| git hash 유효하지 않음 | 경고 후 현재 HEAD 기준 리뷰 |
| Reviewer block 2회 반복 | `blocked:review-rejected` + 사용자 개입 요청 |
| Plans.md 파싱 오류 | 실행 중단 |

---

## 통합 포인트

| 스킬/커맨드 | 연관 방식 |
|-------------|-----------|
| `jikime-harness-work` | cc:DONE 생성 → harness-review 호출 |
| `jikime-harness-sync` | pm:OK 후 sync retro 권장 |
| `/jikime:3-sync` | 모든 pm:OK 후 최종 동기화 |

---

Version: 1.0.0
Status: Active
Last Updated: 2026-03-15
