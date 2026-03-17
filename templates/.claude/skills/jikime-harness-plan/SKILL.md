---
name: jikime-harness-plan
description: Harness Engineering plan manager — creates and manages Plans.md with structured task tracking, DoD criteria, and dependency mapping
version: 1.0.0
category: harness
tags: ["harness", "plan", "plans.md", "task-management", "dod", "workflow", "harness-engineering"]
triggers:
  keywords:
    - "harness-plan"
    - "harness plan"
    - "Plans.md"
    - "plans.md"
    - "계획 생성"
    - "태스크 추가"
    - "작업 계획"
    - "harness 계획"
  phases: ["plan"]
  agents: ["orchestrator", "planner", "manager-strategy"]
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
  - WebSearch
---

# Harness Plan — Plans.md 관리 스킬

## Quick Reference

Plans.md를 생성하고 관리하는 Harness Engineering 핵심 스킬. 아이디어를 실행 가능한 구조화된 태스크로 변환합니다.

**서브커맨드:**
- `create` — 새 Plans.md 생성 (대화형 또는 자동)
- `add` — 기존 Plans.md에 태스크/Phase 추가
- `update` — 태스크 마커 변경
- `sync` — Plans.md ↔ git 히스토리 동기화 (경량)

**마커 시스템:**

| 마커 | 의미 | 사용 시점 |
|------|------|-----------|
| `cc:TODO` | 미시작 | 기본값 |
| `cc:WIP` | 진행 중 | Worker가 시작할 때 |
| `cc:DONE [hash]` | 완료 + git hash | Worker가 커밋 후 |
| `pm:REVIEW` | 사용자 검토 중 | 사용자에게 확인 요청 시 |
| `pm:OK` | 검토 완료 | 사용자가 승인 후 |
| `blocked:<이유>` | 차단 | 의존성/외부 차단 발생 시 |
| `cc:SKIP` | 건너뜀 | 범위 변경으로 불필요해진 경우 |

---

## Plans.md 포맷

```markdown
# Plans.md

## Overview

| 항목 | 내용 |
|------|------|
| **목표** | [프로젝트/기능 목표] |
| **마일스톤** | [완료 기준 날짜 또는 이벤트] |
| **담당** | [사용자명 / Claude] |
| **생성일** | [YYYY-MM-DD] |

---

## Phase 1: [Phase 이름]

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 1.1  | [태스크 설명] | [Yes/No 판정 가능한 완료 기준] | - | cc:TODO |
| 1.2  | [태스크 설명] | [완료 기준] | 1.1 | cc:TODO |

## Phase 2: [Phase 이름]

| Task | 내용 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 2.1  | [태스크 설명] | [완료 기준] | 1.2 | cc:TODO |
```

**DoD (Definition of Done) 작성 규칙:**
- ✅ "테스트 통과" (Yes/No 판정 가능)
- ✅ "lint 에러 0" (측정 가능)
- ✅ "API 엔드포인트 동작 확인" (검증 가능)
- ❌ "코드 품질 향상" (모호함)
- ❌ "잘 작동함" (주관적)

---

## 서브커맨드 상세

### create — 새 Plans.md 생성

**실행 흐름:**
1. 기존 Plans.md 확인 → 있으면 덮어쓰기 여부 확인
2. 사용자 요구사항 수집 (최대 3개 질문):
   - 목표/기능 설명
   - 기술 스택 (자동 감지 가능)
   - 마일스톤/기한
3. 기술 조사 (WebSearch로 패턴/라이브러리 확인)
4. 기능 목록 추출 및 Phase 분류:
   - Phase 1: 핵심 기능 (Required)
   - Phase 2: 권장 기능 (Recommended)
   - Phase 3: 선택 기능 (Optional)
5. 각 태스크의 DoD 추론 (테스트 가능한 기준)
6. 의존성 매핑 (Depends 컬럼)
7. Plans.md 생성 (모든 태스크 `cc:TODO`)

**자동 감지:**
```bash
# 기술 스택 감지
ls package.json go.mod pyproject.toml Cargo.toml 2>/dev/null
git log --oneline -5  # 최근 커밋으로 컨텍스트 파악
```

**Phase 분류 기준:**
| Phase | 기준 | 예시 |
|-------|------|------|
| 1 | 없으면 동작 불가 | 데이터베이스 연결, 핵심 API |
| 2 | 없어도 동작, 품질 향상 | 에러 처리, 로깅, 성능 최적화 |
| 3 | 추가 가치 제공 | 관리자 대시보드, 고급 필터 |

---

### add — 태스크/Phase 추가

**실행 흐름:**
1. 기존 Plans.md 로드 및 파싱
2. 추가 위치 결정 (기존 Phase에 추가 or 새 Phase)
3. 태스크 번호 자동 할당 (연속)
4. DoD 추론
5. 의존성 분석 (기존 태스크와 관계)
6. Plans.md 업데이트

**스코프 크리프 감지:**
새 태스크 추가 시 자동으로 확인:
- 현재 Phase 1 태스크 수 vs 추가 후 수
- 10개 이상이면 Phase 분리 권장
- 완전히 다른 도메인이면 새 Phase 권장

---

### update — 마커 변경

**사용 방법:**
```
/jikime:harness-plan update 1.2 cc:WIP
/jikime:harness-plan update 1.2 cc:DONE abc1234
/jikime:harness-plan update 1.3 blocked:API 응답 대기 중
```

**실행 흐름:**
1. Plans.md에서 태스크 ID로 해당 행 찾기
2. Status 컬럼 업데이트
3. 변경 내역 git에 커밋 (선택):
   ```
   chore(plans): update 1.2 → cc:DONE [abc1234]
   ```

---

### sync (경량) — git 히스토리 기반 상태 추론

**실행 흐름:**
1. Plans.md의 `cc:TODO`/`cc:WIP` 태스크 목록 추출
2. `git log --oneline -20` 분석
3. 커밋 메시지에서 태스크 참조 탐지
4. 추론된 상태 변경 제안 (사용자 확인 후 적용)
5. 불일치 보고:
   ```
   ⚠️ 1.2 → cc:WIP로 표시되어 있지만 관련 커밋 없음
   ✅ 1.3 → 커밋 abc1234에서 완료됨으로 추론
   ```

> 💡 더 상세한 동기화 + 레트로스펙티브는 `jikime-harness-sync` 스킬 사용

---

## 품질 기준

**DoD 검증 (태스크 완료 판정):**
```
✅ 모든 DoD 항목이 Yes/No로 판정 가능한가?
✅ 테스트/검증 방법이 명시되어 있는가?
✅ 의존성이 올바르게 매핑되어 있는가?
✅ Phase 순서가 논리적인가? (Phase 2가 Phase 1에 의존)
```

**완료 판정 기준:**
- `cc:DONE`: Worker가 구현 + 테스트 통과 + 커밋 완료
- `pm:OK`: 사용자가 변경사항 확인 후 승인

---

## 통합 포인트

| 스킬/커맨드 | 연관 방식 |
|-------------|-----------|
| `jikime-harness-sync` | 상세 동기화 + 레트로스펙티브 |
| `jikime-harness-work` | cc:WIP → cc:DONE 자동화 |
| `jikime-harness-review` | pm:REVIEW → pm:OK 자동화 |
| `/jikime:1-plan` | harness-plan create 진입점 |
| `/jikime:3-sync` | harness-sync 진입점 |

---

Version: 1.0.0
Status: Active
Last Updated: 2026-03-15
