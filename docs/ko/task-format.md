# 구조화된 태스크 포맷

> 5필드 구조와 품질 체크포인트를 사용한 체계적 태스크 분해 방법입니다.

## 개요

JikiME-ADK의 모든 구현 태스크는 명확한 완료 기준을 가진 구조화된 단위로 분해할 수 있습니다. 이 포맷은 누락 방지, 진행 상황 추적, 정기적 품질 검증을 보장합니다.

---

## 태스크 구조 (5필드)

각 태스크는 다음 필수 형식을 따릅니다:

```
### Task N: [제목]

Do:        무엇을 구현할 것인가 (구체적 행동)
Files:     어떤 파일을 생성/수정할 것인가
Done when: 측정 가능한 완료 기준
Verify:    완료를 어떻게 검증할 것인가 (테스트 명령, 수동 확인)
Commit:    완료 시 커밋 메시지
```

### 필드 설명

| 필드 | 필수 | 설명 |
|------|------|------|
| **Do** | 예 | 단일, 실행 가능한 구현 단계 |
| **Files** | 예 | 명시적 파일 경로 (생성/수정/삭제) |
| **Done when** | 예 | 측정 가능한 기준 (테스트 통과, 빌드 성공, 출력 일치) |
| **Verify** | 예 | 구체적 검증 명령 또는 확인 방법 |
| **Commit** | 예 | 컨벤셔널 커밋 메시지 (feat/fix/refactor/test/docs) |

### 예시

```
### Task 1: 사용자 인증 API 생성

Do:        POST /api/auth/login 엔드포인트를 JWT 토큰 생성과 함께 구현
Files:     src/api/auth/login.ts (생성), src/types/auth.ts (생성)
Done when: POST /api/auth/login이 올바른 자격증명에 200+JWT, 잘못된 경우 401 반환
Verify:    npm test -- --grep "auth login" 통과
Commit:    feat(auth): add login endpoint with JWT generation
```

---

## 품질 체크포인트 ([VERIFY])

구현 태스크 2-3개마다 `[VERIFY]` 체크포인트 태스크를 삽입합니다.

### 형식

```
### Task N: [VERIFY] 품질 체크포인트

Do:        전체 검증 스위트 실행
Files:     (없음 - 검증만)
Done when: 모든 검사가 에러 없이 통과
Verify:
  - 빌드 통과 (에러 제로)
  - 테스트 통과 (전체 그린)
  - 린트 클린 (경고 제로)
  - 수동: 기능이 예상대로 동작
Commit:    (커밋 없음 - 체크포인트만)
```

### 체크포인트 규칙

| 규칙 | 설명 |
|------|------|
| **빈도** | 태스크 2-3개마다 [VERIFY] 삽입 |
| **블로킹** | 체크포인트 실패 시 다음 태스크 진행 불가 |
| **수정 우선** | 계속하기 전에 수정 태스크 생성 |
| **생략 불가** | [VERIFY] 태스크는 건너뛸 수 없음 |
| **마지막** | 항상 마지막에 [VERIFY] 삽입 |

### 삽입 패턴

```
Task 1: 구현
Task 2: 구현
Task 3: 구현
Task 4: [VERIFY] 품질 체크포인트     ← 3개 후
Task 5: 구현
Task 6: 구현
Task 7: [VERIFY] 품질 체크포인트     ← 2개 후
...
Task N: [VERIFY] 최종 체크포인트     ← 항상 마지막에
```

---

## DDD 태스크 변형

기존 코드 작업 시 (DDD 워크플로우) ANALYZE/PRESERVE/IMPROVE 접두사 사용:

```
Task 1: [ANALYZE] 현재 인증 흐름 파악
Task 2: [PRESERVE] 특성화 테스트 추가
Task 3: [IMPROVE] JWT 기반 인증으로 리팩토링
Task 4: [VERIFY] 품질 체크포인트
```

---

## 태스크 크기 가이드라인

| 크기 | 설명 | 예시 |
|------|------|------|
| **XS** | 단일 함수/메서드 | 유효성 검증 헬퍼 추가 |
| **S** | 단일 파일 변경 | API 엔드포인트 생성 |
| **M** | 2-3개 파일 변경 | 기능 + 테스트 |
| **L** | 4개+ 파일 변경 | **더 작은 태스크로 분할** |

**규칙**: 태스크가 3개 이상의 파일을 수정하면, 더 작은 태스크로 분할합니다.

---

## TodoWrite 연동

모든 태스크는 TodoWrite를 통해 실시간 추적됩니다:

```
태스크 발견   → TodoWrite: "pending" 상태로 추가
태스크 시작   → TodoWrite: "in_progress"로 변경
태스크 완료   → TodoWrite: "completed"로 변경
[VERIFY] 실패 → TodoWrite: 수정 태스크를 "pending"으로 추가
```

---

## 워크플로우와의 연동

| 워크플로우 | 태스크 포맷 활용 방법 |
|-----------|---------------------|
| **POC-First** | 각 단계에서 이 형식으로 태스크 생성 |
| **TDD** | RED-GREEN-REFACTOR가 태스크 시퀀스로 매핑 |
| **DDD** | ANALYZE-PRESERVE-IMPROVE가 태스크 접두사로 매핑 |
| **Ralph Loop** | 루프 반복에서 TodoWrite로 진행 추적 |

---

## 관련 문서

- [POC-First 워크플로우](./poc-first.md) — 단계별 Greenfield 개발
- [PR 라이프사이클 자동화](./pr-lifecycle.md) — PR 관리 자동화
- [TDD & DDD 워크플로우](./tdd-ddd.md) — 대안 개발 방법론
