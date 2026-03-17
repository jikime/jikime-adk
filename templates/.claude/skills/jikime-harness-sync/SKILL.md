---
name: jikime-harness-sync
description: Harness Engineering sync manager — Plans.md ↔ git history detailed sync, Agent Trace analysis, and 4-item retrospective generation
version: 1.0.0
category: harness
tags: ["harness", "sync", "retrospective", "plans.md", "git", "agent-trace", "harness-engineering"]
triggers:
  keywords:
    - "harness-sync"
    - "harness sync"
    - "sync plans"
    - "retrospective"
    - "레트로스펙티브"
    - "동기화"
    - "plans 동기화"
    - "harness 동기화"
  phases: ["sync", "review"]
  agents: ["orchestrator", "planner", "manager-strategy"]
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
  - WebSearch
---

# Harness Sync — Plans.md ↔ Git 상세 동기화 스킬

## Quick Reference

Plans.md와 git 히스토리를 상세 동기화하고 4항목 레트로스펙티브를 생성하는 Harness Engineering 핵심 스킬.

**서브커맨드:**
- `sync` — Plans.md ↔ git 상태 전체 동기화
- `retro` — 레트로스펙티브 생성 (완료 태스크 1개 이상 필요)
- `trace` — Agent Trace 분석 (특정 태스크 실행 이력 추적)
- `drift` — Plans.md vs 실제 구현 간 드리프트 감지

**마커 시스템 (참고):**

| 마커 | 의미 |
|------|------|
| `cc:TODO` | 미시작 |
| `cc:WIP` | 진행 중 |
| `cc:DONE [hash]` | 완료 + git hash |
| `pm:REVIEW` | 사용자 검토 중 |
| `pm:OK` | 검토 완료 |
| `blocked:<이유>` | 차단 |
| `cc:SKIP` | 건너뜀 |

---

## sync — 상세 동기화

### 실행 흐름

1. Plans.md 전체 파싱 (모든 Phase, 모든 태스크)
2. `git log --oneline -50` + `git log --all --format="%H %s"` 수집
3. 각 태스크별 커밋 연결:
   ```bash
   git log --oneline --all | grep -E "(1\.2|task-1\.2|feat.*1\.2)"
   ```
4. 상태 불일치 탐지:
   - `cc:WIP`이지만 관련 커밋 없음 → ⚠️ stale WIP
   - `cc:TODO`이지만 커밋에서 완료 추론 가능 → ✅ 업데이트 제안
   - `cc:DONE`이지만 hash가 존재하지 않음 → ❌ 유효하지 않은 hash
5. 변경 제안 목록 출력 (사용자 확인 후 적용)
6. Plans.md 업데이트

### 커밋 연결 전략

**우선순위 순서:**

| 우선순위 | 패턴 | 예시 |
|----------|------|------|
| 1 | 명시적 태스크 ID | `feat: [1.2] implement login` |
| 2 | Phase + 내용 키워드 | `feat(auth): add JWT` → 태스크 1.x 추론 |
| 3 | 파일 변경 패턴 | `auth.go`, `login.ts` → 태스크 도메인 매핑 |
| 4 | 타임스탬프 근접 | cc:WIP 설정 시점 이후 커밋 |

### 불일치 보고 형식

```
📋 Plans.md 동기화 결과

✅ 자동 완료 추론:
  1.3 → 커밋 abc1234 "feat(auth): add JWT login" 에서 완료 추론
  2.1 → 커밋 def5678 "feat(db): add user schema" 에서 완료 추론

⚠️ 검토 필요:
  1.2 → cc:WIP로 표시되어 있지만 관련 커밋 없음 (5일 경과)
  2.3 → cc:DONE [xyz0000] 이지만 hash 없음

❌ 블로킹:
  1.4 → blocked:API 응답 대기 중 (3일 경과)

변경사항을 적용할까요? (y/n)
```

---

## retro — 레트로스펙티브 생성

**전제 조건:** 완료된 태스크(`cc:DONE`) 1개 이상

### 4항목 레트로스펙티브

#### 1. 견적 정확도 (Estimation Accuracy)

```
계획된 태스크 수: N
완료된 태스크 수: M (cc:DONE + pm:OK)
건너뛴 태스크 수: K (cc:SKIP)
추가된 태스크 수: J (스코프 크리프)

견적 정확도: M / (N + J) * 100%
```

**분석 출력:**
```markdown
## 견적 정확도
- 계획: 8개 태스크
- 완료: 6개 (75%)
- 스킵: 1개 (범위 축소)
- 추가: 3개 (스코프 크리프 +37.5%)
- 순 정확도: 70%

💡 인사이트: Phase 2 태스크가 예상보다 2배 많이 추가됨.
   다음에는 Phase 2 견적에 버퍼 50% 추가 권장.
```

#### 2. 블로킹 원인 (Blocking Analysis)

```bash
# Plans.md에서 blocked 항목 추출
grep -E "blocked:" Plans.md
# git log에서 해당 기간 분석
git log --all --format="%ai %s" --after="2026-01-01"
```

**분석 출력:**
```markdown
## 블로킹 원인 분석
- 외부 API 대기: 2건 (평균 3.5일)
- 의존성 미완료: 1건 (1.5일)
- 환경 설정: 1건 (0.5일)

💡 인사이트: 외부 의존성이 주요 블로킹 원인.
   다음 계획에서 외부 API 태스크는 Phase 1 시작 시 병렬로 착수 권장.
```

#### 3. 품질 마커 적중률 (Quality Marker Hit Rate)

```
pm:OK 전환 비율: pm:OK 수 / cc:DONE 수 * 100%
재검토 비율: pm:REVIEW → pm:REVIEW (재검토 횟수) / 전체 REVIEW 수
```

**분석 출력:**
```markdown
## 품질 마커 적중률
- cc:DONE → pm:OK 전환: 5/6 (83%)
- 재검토 없이 1회 통과: 4/5 (80%)
- 재검토 발생: 1건 (DoD 기준 불명확)

💡 인사이트: DoD 기준이 명확한 태스크는 100% 1회 통과.
   모호한 DoD 태스크에서만 재검토 발생.
   다음에는 DoD에 구체적인 테스트 커맨드 포함 권장.
```

#### 4. 스코프 변동 (Scope Drift)

```bash
# git log에서 태스크 추가 이력 분석
git log --all --format="%s" | grep -E "(Plans\.md|태스크 추가|add task)"
# Plans.md 변경 이력
git log --all --follow Plans.md --format="%ai %s"
```

**분석 출력:**
```markdown
## 스코프 변동 분석
- 최초 계획: 8개 태스크 (2026-03-01)
- 현재: 11개 태스크 (+37.5%)
- 변동 시점:
  - 2026-03-05: 태스크 2.4, 2.5 추가 (요구사항 변경)
  - 2026-03-10: 태스크 3.1 추가 (사용자 피드백)

💡 인사이트: 첫 주에 요구사항이 크게 변동.
   다음에는 Phase 1 완료 후 Phase 2/3 태스크 확정하는 Rolling Planning 권장.
```

### 레트로스펙티브 출력 형식

Plans.md 하단에 자동 추가:

```markdown
---

## Retrospective — [날짜]

### 요약

| 항목 | 결과 |
|------|------|
| 견적 정확도 | 70% |
| 블로킹 원인 (주요) | 외부 API 대기 |
| 품질 마커 적중률 | 83% |
| 스코프 변동 | +37.5% |

### 상세 분석

[4항목 상세 내용]

### 다음 계획을 위한 액션 아이템

1. Phase 2 견적에 50% 버퍼 추가
2. 외부 API 태스크를 Phase 1 시작 시 병렬 착수
3. DoD에 구체적인 테스트 커맨드 포함
4. Rolling Planning 적용 (Phase 1 후 Phase 2/3 확정)
```

---

## trace — Agent Trace 분석

특정 태스크의 실행 이력을 Agent 관점에서 추적합니다.

### 실행 흐름

1. 태스크 ID로 관련 커밋 수집
2. 각 커밋의 변경 파일 목록 분석
3. 실행 패턴 감지:
   - Worker → Reviewer → commit 순서
   - 재시도 패턴 (동일 파일 반복 수정)
   - 에러 복구 패턴 (롤백 후 재구현)
4. Agent 효율성 메트릭 계산

### Trace 출력 형식

```
🔍 태스크 1.3 Agent Trace

커밋 이력:
  abc1234 (2026-03-05 14:23) feat(auth): add JWT token generation
    변경: auth/jwt.go (+45 -0), auth/jwt_test.go (+30 -0)

  def5678 (2026-03-05 15:10) fix(auth): handle token expiry edge case
    변경: auth/jwt.go (+8 -2), auth/jwt_test.go (+12 -0)

  ghi9012 (2026-03-05 15:45) chore(plans): update 1.3 → cc:DONE [ghi9012]
    변경: Plans.md (+1 -1)

패턴 감지:
  ✅ 순차 실행 (구현 → 테스트 → 완료 마킹)
  ⚠️ 1회 수정 발생 (edge case 처리 누락)
  ✅ 테스트 커버리지 동반 (모든 커밋에 _test.go 포함)

효율성: 3커밋 / 1태스크 (평균: 2.5)
소요 시간: 82분 (계획 대비 +27분)
```

---

## drift — 드리프트 감지

Plans.md에 정의된 태스크와 실제 구현된 파일 간의 불일치를 감지합니다.

### 실행 흐름

1. Plans.md의 모든 `cc:DONE` 태스크와 연결된 커밋 수집
2. 각 커밋의 변경 파일 목록 추출
3. 파일 존재 여부 검증:
   ```bash
   git show --name-only [hash]
   ls -la [변경된 파일들]
   ```
4. Plans.md의 태스크 설명 vs 실제 변경 파일 도메인 비교
5. 드리프트 보고

### 드리프트 유형

| 유형 | 설명 | 예시 |
|------|------|------|
| Missing File | 커밋에서 변경됐지만 파일 삭제됨 | 리팩토링으로 파일 제거 |
| Scope Drift | 태스크 설명 도메인과 다른 파일 변경 | auth 태스크에서 DB 파일 수정 |
| Undocumented | 커밋은 있지만 Plans.md 태스크 없음 | 긴급 핫픽스 |
| Orphan Task | Plans.md에 있지만 관련 커밋 없음 | 스펙만 있고 구현 없는 태스크 |

---

## 통합 포인트

| 스킬/커맨드 | 연관 방식 |
|-------------|-----------|
| `jikime-harness-plan` | Plans.md 생성/수정의 경량 sync 포함 |
| `jikime-harness-work` | cc:WIP → cc:DONE 자동화 후 sync 실행 권장 |
| `jikime-harness-review` | pm:REVIEW → pm:OK 후 sync + retro 실행 |
| `/jikime:3-sync` | harness-sync sync 진입점 |
| `/jikime:harness` | WORKFLOW.md에서 sync 단계 자동 실행 |

---

## 품질 기준

```
✅ 모든 cc:DONE 태스크에 유효한 git hash 연결 확인
✅ 블로킹 태스크의 원인이 3일 이상이면 자동 경고
✅ 레트로스펙티브는 완료 태스크 1개 이상일 때만 생성
✅ drift 감지는 Plans.md 변경 전 항상 실행 권장
```

---

Version: 1.0.0
Status: Active
Last Updated: 2026-03-15
