# J.A.R.V.I.S. - Intelligent Autonomous Orchestration

JikiME-ADK의 지능형 자율 오케스트레이션 시스템. Iron Man의 AI 비서에서 영감을 받은 선제적이고 적응적인 개발 자동화.

## Overview

J.A.R.V.I.S. (Just A Rather Very Intelligent System)는 단순 명령 실행이 아닌 **예측하고, 적응하고, 학습하는** 지능형 오케스트레이터입니다.

### 핵심 철학

```
"I'm not just following orders, sir. I'm anticipating your needs."
```

### 기존 오케스트레이터와의 차별점

| 기능 | 기존 오케스트레이터 | J.A.R.V.I.S. |
|------|--------------|--------------|
| 탐색 | 3개 에이전트 병렬 | 5개 에이전트 + 의존성 분석 |
| 계획 | 단일 전략 | 멀티 전략 비교 후 최적 선택 |
| 실행 | 순차/병렬 고정 | 상황 적응형 동적 전환 |
| 에러 처리 | 단순 재시도 | 자가 진단 + 대안 전략 피봇 |
| 학습 | 없음 | 세션 내 패턴 학습 |
| 예측 | 없음 | 다음 단계 선제 제안 |

## Architecture

### 시스템 구조

```
┌─────────────────────────────────────────────────────────────────┐
│                    J.A.R.V.I.S. System                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Phase 0: Proactive Intelligence Gathering                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐│
│  │ Explore  │ │ Research │ │ Quality  │ │ Security │ │  Perf  ││
│  │  Agent   │ │  Agent   │ │  Agent   │ │  Agent   │ │ Agent  ││
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └───┬────┘│
│       └────────────┴────────────┼────────────┴───────────┘     │
│                                 ▼                               │
│                    ┌────────────────────┐                       │
│                    │ Integration Engine │                       │
│                    │  + Dependency Map  │                       │
│                    │  + Risk Assessment │                       │
│                    └─────────┬──────────┘                       │
│                              ▼                                  │
│  Phase 1: Multi-Strategy Planning                               │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│  │ Strategy A   │ │ Strategy B   │ │ Strategy C   │            │
│  │ Conservative │ │  Balanced    │ │  Aggressive  │            │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘            │
│         └────────────────┼────────────────┘                    │
│                          ▼                                      │
│                ┌──────────────────┐                             │
│                │ Trade-off Matrix │                             │
│                │ Optimal Selection│                             │
│                └────────┬─────────┘                             │
│                         ▼                                       │
│  Phase 2: Adaptive DDD Implementation                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  WHILE (issues_exist AND iteration < max):              │   │
│  │    ├── Diagnostics (LSP + Tests + Coverage)             │   │
│  │    ├── Self-Assessment: "Is approach working?"          │   │
│  │    │   ├── YES → Continue                               │   │
│  │    │   └── NO  → Pivot Strategy                         │   │
│  │    ├── Expert Agent Delegation                          │   │
│  │    └── Verification                                     │   │
│  └─────────────────────────────────────────────────────────┘   │
│                         ▼                                       │
│  Phase 3: Completion & Prediction                               │
│  ┌──────────────┐ ┌──────────────────────┐                     │
│  │  Doc Sync    │ │ Predictive Suggest   │                     │
│  │ manager-docs │ │ "You might also..."  │                     │
│  └──────────────┘ └──────────────────────┘                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 관련 파일

| 파일 | 설명 |
|------|------|
| `templates/.claude/commands/jikime/jarvis.md` | J.A.R.V.I.S. 슬래시 커맨드 |
| `templates/CLAUDE.md` | Type B 유틸리티 명령어로 등록 |
| `docs/migration.md` | 마이그레이션 시스템 문서 |

## Migration Mode

### 개요

J.A.R.V.I.S. 마이그레이션 모드는 레거시 프로젝트를 현대 프레임워크로 전환하는 전체 과정을 자동화합니다.

### 마이그레이션 워크플로우

```
Phase 0: Source Discovery
  ├── 기술 스택 분석 (프레임워크, 라이브러리)
  ├── 아키텍처 패턴 파악
  ├── 복잡도 점수 계산
  └── 타겟 프레임워크 제안
        ↓
Phase 1: Detailed Analysis
  ├── as_is_spec.md 생성
  ├── 비즈니스 로직 문서화
  ├── 라우트/컴포넌트 매핑
  └── .migrate-config.yaml 생성
        ↓
Phase 2: Migration Planning
  ├── 마이그레이션 전략 비교 (점진적 vs 빅뱅)
  ├── 모듈별 우선순위 결정
  ├── 리스크 평가
  └── migration_plan.md 생성
        ↓
Phase 3: DDD Execution
  ├── 모듈별 ANALYZE-PRESERVE-IMPROVE
  ├── 전문가 에이전트 위임
  ├── 자가 진단 및 피봇
  └── 코드 마이그레이션
        ↓
Phase 4: Verification
  ├── 동작 비교 테스트
  ├── E2E 테스트
  ├── 성능 비교
  └── 검증 보고서 생성
```

### 마이그레이션 에이전트 위임

| Phase | 에이전트 | 역할 |
|-------|---------|------|
| Phase 0 | Explore, Research | 소스 코드 분석, 기술 조사 |
| Phase 1 | manager-spec | as_is_spec.md 생성 |
| Phase 2 | manager-strategy | 마이그레이션 전략 수립 |
| Phase 3 | backend, frontend | 코드 마이그레이션 |
| Phase 4 | test-guide | 검증 및 테스트 |

### 산출물 흐름

```
./legacy-project/
    │
    ▼ (Phase 0-1)
./migrations/{project}/
    ├── as_is_spec.md
    └── .migrate-config.yaml
    │
    ▼ (Phase 2)
./migrations/{project}/
    └── migration_plan.md
    │
    ▼ (Phase 3)
./migrations/{project}/out/
    └── (마이그레이션된 프로젝트)
    │
    ▼ (Phase 4)
검증 보고서 + 완료
```

### 지원 마이그레이션

| 소스 | 타겟 옵션 |
|------|----------|
| Vue 2/3 | Next.js (App Router) |
| React (CRA) | Next.js (App Router) |
| Angular | Next.js, SvelteKit |
| jQuery | React, Vue, Svelte |
| PHP | Next.js, FastAPI, Go |

## Usage

### 기본 사용법

```bash
# 지능형 자율 실행 (개발 모드 - 기본)
/jikime:jarvis "Add JWT authentication"

# 안전 전략 (보수적 접근)
/jikime:jarvis "Refactor payment module" --strategy safe

# 빠른 전략 (공격적 접근)
/jikime:jarvis "Fix typo in README" --strategy fast

# 자동 루프 활성화
/jikime:jarvis "Implement user dashboard" --loop --max 20

# 이전 작업 재개
/jikime:jarvis resume SPEC-AUTH-001

# === 마이그레이션 모드 ===
# 레거시 프로젝트를 현대 프레임워크로 마이그레이션
/jikime:jarvis "Migrate Vue app to Next.js" --mode migrate

# 자동 모드 감지 (.migrate-config.yaml 확인)
/jikime:jarvis "Continue migration" --mode auto

# 안전 전략으로 마이그레이션
/jikime:jarvis @./legacy-app/ "Migrate to Next.js 16" --mode migrate --strategy safe
```

### 명령어 옵션

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--mode` | 워크플로우 모드: dev, migrate, auto | auto |
| `--strategy` | 실행 전략: auto, safe, fast | auto |
| `--loop` | 에러 자동 수정 반복 활성화 | config |
| `--max N` | 최대 반복 횟수 | 50 |
| `--branch` | 피처 브랜치 자동 생성 | config |
| `--pr` | 완료 시 PR 자동 생성 | config |
| `--resume SPEC` | 이전 작업 재개 | - |

### 모드 자동 감지 (--mode auto)

J.A.R.V.I.S.는 다음 기준으로 워크플로우 모드를 자동 감지합니다:

| 지표 | 감지 모드 |
|------|----------|
| `.migrate-config.yaml` 존재 | migrate |
| `migrations/*/as_is_spec.md` 존재 | migrate |
| 키워드: "migrate", "migration", "convert" | migrate |
| `spec.md` 또는 `.jikime/project/` 존재 | dev |
| 기본값 | dev |

## Intelligence Features

### 1. Proactive Intelligence Gathering (Phase 0)

5개의 전문 에이전트가 **동시에** 분석을 수행합니다:

| 에이전트 | 역할 | 출력 |
|----------|------|------|
| **Explore Agent** | 코드베이스 구조, 아키텍처 패턴 분석 | 관련 파일 목록, 의존성 맵 |
| **Research Agent** | 외부 문서, 라이브러리 베스트 프랙티스 | 구현 패턴, API 레퍼런스 |
| **Quality Agent** | 테스트 커버리지, 코드 품질 베이스라인 | 품질 메트릭, 기술 부채 평가 |
| **Security Agent** | 잠재적 보안 영향 사전 스캔 | 보안 고려사항, OWASP 체크리스트 |
| **Performance Agent** | 성능 영향 예측 분석 | 병목 위험, 최적화 기회 |

### 2. Multi-Strategy Planning (Phase 1)

모든 작업에 대해 2-3개의 접근 전략을 생성하고 비교합니다:

#### 전략 유형

| 전략 | 리스크 | 속도 | 되돌리기 | 테스트 커버리지 |
|------|--------|------|----------|----------------|
| **Conservative** | 낮음 | 느림 | 쉬움 | 100% |
| **Balanced** | 중간 | 중간 | 중간 | 85% |
| **Aggressive** | 높음 | 빠름 | 어려움 | 70% |

#### 자동 전략 선택 알고리즘

```
IF risk_score > 70:
    SELECT Conservative (안전 우선)
ELIF risk_score > 40:
    SELECT Balanced (균형)
ELSE:
    SELECT Aggressive (속도 우선)

OVERRIDE: --strategy 플래그로 수동 지정 가능
```

### 3. Adaptive Execution (Phase 2)

#### 자가 진단 루프

매 반복마다 J.A.R.V.I.S.는 스스로 질문합니다:

1. **"현재 접근법이 진전을 보이고 있는가?"**
   - 에러 수가 줄어들고 있는가?
   - 테스트 통과율이 개선되고 있는가?

2. **"다른 전략으로 전환해야 하는가?"**
   - 트리거: 3회 연속 개선 없음
   - 행동: 대안 전략으로 피봇

3. **"이전에 본 패턴인가?"**
   - 세션 내 유사 에러 패턴 확인
   - 학습된 해결책 즉시 적용

#### 피봇 결정 트리

```
IF no_progress_count >= 3:
    IF current_strategy == "aggressive":
        PIVOT → "balanced"
    ELIF current_strategy == "balanced":
        PIVOT → "conservative"
    ELSE:
        REQUEST → user_intervention
```

### 4. Predictive Suggestions (Phase 3)

완료된 작업을 기반으로 다음 단계를 예측하고 제안합니다:

```markdown
## Completed: JWT Authentication

### Predictive Suggestions

Based on this implementation, you might also want to:

1. **Add refresh token mechanism** - JWT 토큰 만료 시 세션 연장
2. **Implement rate limiting** - 인증 엔드포인트 무차별 공격 방지
3. **Add password reset flow** - 인증과 짝을 이루는 일반적인 기능
4. **Set up audit logging** - 보안을 위한 인증 이벤트 추적

Would you like me to start any of these?
```

## Strategy Details

### auto (기본값)

J.A.R.V.I.S.가 작업 복잡도를 분석하여 최적 전략을 자동 선택:

| 작업 유형 | 분석 결과 | 선택 전략 |
|-----------|----------|----------|
| 단순 (단일 도메인) | 리스크 < 40 | 직접 전문가 위임 |
| 중간 (2-3 도메인) | 리스크 40-70 | 순차 워크플로우 |
| 복잡 (4+ 도메인) | 리스크 > 70 | 전체 병렬 오케스트레이션 |

### safe (보수적)

최대한의 검증과 안전장치를 적용:

- 더 많은 사용자 체크포인트
- 더 작은 점진적 변경
- 각 단계마다 종합 테스트
- 모든 Phase에 롤백 포인트

### fast (공격적)

단순하거나 긴급한 작업을 위한 빠른 실행:

- 최소한의 체크포인트
- 가능한 모든 것 병렬화
- 선택적 검증 건너뛰기
- 빠른 완료 우선

## Agent Delegation

### 위임 규칙

[HARD] 모든 구현 작업은 반드시 전문가 에이전트에 위임:

| 작업 유형 | 담당 에이전트 |
|-----------|--------------|
| 백엔드 로직 | backend |
| 프론트엔드 컴포넌트 | frontend |
| 테스트 생성 | test-guide |
| 버그 수정 | debugger |
| 리팩토링 | refactorer |
| 보안 수정 | security-auditor |
| 성능 최적화 | optimizer |

## Output Format

### 실행 중

```markdown
## J.A.R.V.I.S.: Phase 2 (Iteration 3/50)

### Strategy: Balanced (auto-selected)
### Risk Score: 45/100

### Current Status
- [x] User model created
- [x] JWT token generation
- [ ] Login endpoint ← in progress
- [ ] Token validation middleware

### Self-Assessment
- Progress: YES (2 errors → 1 error)
- Pivot needed: NO
- Confidence: 85%

### Active Issues
- ERROR: src/auth/login.ts:45 - undefined 'hashPassword'

Fixing...
```

### 완료

```markdown
## J.A.R.V.I.S.: COMPLETE

### Summary
- SPEC: SPEC-AUTH-001
- Strategy Used: Balanced
- Files Modified: 12
- Tests: 34/34 passing
- Coverage: 92%
- Iterations: 5
- Self-Corrections: 1 (pivoted from aggressive at iteration 3)

### Changes Implemented
- JWT token generation and validation
- Login/logout endpoints
- Token refresh mechanism
- Authentication middleware
- Comprehensive test suite

### Predictive Suggestions
1. Add rate limiting to auth endpoints
2. Implement password reset flow
3. Set up audit logging

<jikime>DONE</jikime>
```

## Workflow Integration

### 개발 워크플로우와의 관계

J.A.R.V.I.S.는 개별 워크플로우 명령어들을 **통합 자동화**합니다:

```
개발 모드 (--mode dev):
  개별:  /jikime:0-project → /jikime:1-plan → /jikime:2-run → /jikime:3-sync
  통합:  /jikime:jarvis "task" → 전체 자동 실행

마이그레이션 모드 (--mode migrate):
  개별:  /jikime:migrate-0-discover → 1-analyze → 2-plan → 3-execute → 4-verify
  통합:  /jikime:jarvis "migrate to X" --mode migrate → 전체 자동 실행
```

### 명령어 체계

| 타입 | 명령어 | 용도 |
|------|--------|------|
| **Workflow (Type A)** | 0-project, 1-plan, 2-run, 3-sync | 개발 단계별 세밀한 제어 |
| **Migration** | migrate-0 ~ migrate-4 | 마이그레이션 단계별 제어 |
| **Utility (Type B)** | **jarvis**, test, loop, fix | 빠른 실행 및 자동화 |

### 워크플로우 모드 비교

| 측면 | 개발 모드 (dev) | 마이그레이션 모드 (migrate) |
|------|----------------|---------------------------|
| **목적** | 새 기능 구현, 개선 | 레거시 → 현대 프레임워크 전환 |
| **입력** | 작업 설명, SPEC | 레거시 소스 코드 |
| **단계** | 4단계 (0-project ~ 3-sync) | 5단계 (0-discover ~ 4-verify) |
| **산출물** | 코드, 문서 | 마이그레이션된 프로젝트, 검증 보고서 |
| **방법론** | DDD | DDD + 동작 비교 검증 |

## Limitations & Safety

### 제한사항

- 최대 3회 전략 피봇 (이후 사용자 개입 요청)
- 중요 작업(마이그레이션, 삭제) 중에는 피봇 금지
- 세션 내 학습만 지원 (세션 간 학습 미지원)

### 안전장치

- [HARD] 모든 구현은 전문가 에이전트에 위임
- [HARD] SPEC 생성 전 사용자 확인 필수
- [HARD] 완료 마커 필수: `<jikime>DONE</jikime>`
- 각 Phase에 롤백 포인트 생성

## Best Practices

### 언제 J.A.R.V.I.S.를 사용하나요?

**적합한 경우:**
- 새로운 기능 구현 (여러 도메인에 걸친)
- 대규모 리팩토링
- 복잡한 버그 수정
- 전체 워크플로우 자동화가 필요할 때

**개별 명령어가 나은 경우:**
- 단일 파일 수정
- 특정 단계만 실행하고 싶을 때
- 세밀한 제어가 필요할 때

### 권장 사용 패턴

```bash
# 복잡한 새 기능
/jikime:jarvis "Implement payment processing system"

# 안전이 중요한 리팩토링
/jikime:jarvis "Refactor database layer" --strategy safe

# 간단한 수정
/jikime:jarvis "Add validation to login form" --strategy fast

# 긴 작업에 대한 재개 지원
/jikime:jarvis "Complex feature" --loop --max 30
# ... 중단 후 ...
/jikime:jarvis resume SPEC-XXX
```

---

Version: 2.0.0
Last Updated: 2026-01-23
Codename: J.A.R.V.I.S. (Just A Rather Very Intelligent System)
Inspiration: Iron Man's AI Assistant
Changelog:
- v2.0.0: Added Migration Mode (--mode migrate), unified workflow orchestration
- v1.0.0: Initial release with Development Mode
