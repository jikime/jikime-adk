# Codemap & Cleanup Commands

아키텍처 문서화와 코드 정리를 위한 명령어 레퍼런스입니다.

## Overview

| 명령어 | 목적 | 주요 도구 |
|--------|------|-----------|
| `/jikime:codemap` | AST 분석 기반 아키텍처 문서화 | ts-morph, madge |
| `/jikime:cleanup` | Dead code 탐지 및 안전한 제거 | knip, depcheck, ts-prune |

---

## /jikime:codemap

**AST 분석 기반 아키텍처 매핑**

| 항목 | 내용 |
|------|------|
| **설명** | 코드베이스에서 아키텍처 문서를 자동 생성 |
| **Type** | Utility (Type B) |
| **Context** | sync.md |
| **Skill** | jikime-workflow-codemap |
| **단독 사용** | ✅ 높음 - 언제든 독립 실행 가능 |

### Usage

```bash
# 전체 아키텍처 맵 생성
/jikime:codemap all

# 특정 영역만 생성
/jikime:codemap frontend
/jikime:codemap backend
/jikime:codemap database
/jikime:codemap integrations

# AST 분석 포함 (TypeScript/JavaScript)
/jikime:codemap all --ast

# 의존성 그래프 생성
/jikime:codemap all --deps

# 강제 재생성
/jikime:codemap all --refresh

# JSON 출력 (자동화용)
/jikime:codemap all --json
```

### Options

| Option | 설명 |
|--------|------|
| `all` | 모든 영역의 codemap 생성 |
| `frontend` | 프론트엔드 아키텍처만 |
| `backend` | 백엔드/API 아키텍처만 |
| `database` | 데이터베이스 스키마/모델 |
| `integrations` | 외부 서비스 연동 |
| `--ast` | ts-morph로 AST 분석 활성화 |
| `--deps` | madge로 의존성 그래프 생성 |
| `--refresh` | 캐시 무시하고 강제 재생성 |
| `--json` | 자동화를 위한 JSON 출력 |

### 분석 도구

#### 1. ts-morph (AST 분석)

TypeScript/JavaScript 프로젝트의 구조적 분석:

```typescript
// 추출 정보
- 모든 exported 함수, 클래스, 타입
- import/export 관계
- 모듈 의존성
- 라우트 정의 (Next.js, Express 등)
```

#### 2. madge (의존성 그래프)

```bash
# SVG 그래프 생성
npx madge --image docs/CODEMAPS/assets/dependency-graph.svg src/

# 순환 의존성 탐지
npx madge --circular src/
```

### 프레임워크 감지

| 감지 파일 | 프레임워크 | Codemap 초점 |
|-----------|-----------|--------------|
| `next.config.*` | Next.js | App Router, API Routes, Pages |
| `vite.config.*` | Vite | Components, Modules |
| `angular.json` | Angular | Modules, Services, Components |
| `nuxt.config.*` | Nuxt | Pages, Plugins, Modules |
| `package.json` + express | Express | Routes, Middleware |
| `go.mod` | Go | Packages, Handlers |
| `Cargo.toml` | Rust | Crates, Modules |
| `pyproject.toml` | Python | Packages, Modules |

### Output Structure

```
docs/
├── CODEMAPS/
│   ├── INDEX.md              # 아키텍처 개요
│   ├── frontend.md           # 프론트엔드 구조
│   ├── backend.md            # 백엔드/API 구조
│   ├── database.md           # 데이터베이스 스키마
│   ├── integrations.md       # 외부 서비스
│   └── assets/
│       ├── dependency-graph.svg
│       └── architecture-diagram.svg
```

### Codemap 파일 형식

```markdown
# [Area] Codemap

**Last Updated:** YYYY-MM-DD
**Version:** X.Y.Z
**Entry Points:** [주요 진입점 목록]

## Overview
[영역에 대한 간략한 설명]

## Architecture
[컴포넌트 관계를 보여주는 ASCII 다이어그램]

## Key Modules

| Module | Purpose | Exports | Dependencies |
|--------|---------|---------|--------------||
| ... | ... | ... | ... |

## Data Flow
[이 영역을 통한 데이터 흐름 설명]

## External Dependencies
- package@version - 용도
- ...

## Related Codemaps
- [Related Area](./related.md)
```

### Process

```
Phase 1: Discovery
    ↓
  프레임워크, 언어, 프로젝트 타입 감지
    ↓
  진입점과 핵심 파일 식별
    ↓
Phase 2: Analysis
    ↓
  AST 파싱 (ts-morph for TS/JS)
    ↓
  의존성 그래프 (madge)
    ↓
  패턴 인식 (MVC, Clean 등)
    ↓
Phase 3: Generation
    ↓
  구조화된 codemap 생성
    ↓
  ASCII 다이어그램 생성
    ↓
  관계 테이블 빌드
    ↓
Phase 4: Validation
    ↓
  경로 존재 확인
    ↓
  링크 타겟 검증
    ↓
  커버리지 통계 리포트
```

---

## /jikime:cleanup

**Dead Code 탐지 및 안전한 제거**

| 항목 | 내용 |
|------|------|
| **설명** | 종합적인 dead code 분석과 DELETION_LOG 추적 |
| **Type** | Utility (Type B) |
| **Context** | dev.md |
| **Agent** | refactorer |
| **단독 사용** | ✅ 높음 - 언제든 독립 실행 가능 |

### Usage

```bash
# Dead code 스캔 (분석만, 변경 없음)
/jikime:cleanup scan

# Safe 항목만 제거 (저위험)
/jikime:cleanup remove --safe

# Careful 항목 포함 제거 (중위험, 확인 필요)
/jikime:cleanup remove --careful

# 특정 카테고리만 대상
/jikime:cleanup remove --deps      # 미사용 npm 의존성
/jikime:cleanup remove --exports   # 미사용 exports
/jikime:cleanup remove --files     # 미사용 파일

# Dry run (무엇이 제거될지 보여줌)
/jikime:cleanup scan --dry-run

# 삭제 기록 확인
/jikime:cleanup log

# 종합 정리 리포트 생성
/jikime:cleanup report
```

### Options

| Option | 설명 |
|--------|------|
| `scan` | 코드베이스의 dead code 분석 (변경 없음) |
| `remove` | 감지된 dead code 제거 |
| `report` | 종합 정리 리포트 생성 |
| `log` | DELETION_LOG.md 기록 확인 |
| `--safe` | 저위험 항목만 제거 |
| `--careful` | 중위험 항목 포함 (검증 필요) |
| `--deps` | 미사용 의존성 대상 |
| `--exports` | 미사용 exports 대상 |
| `--files` | 미사용 파일 대상 |
| `--dry-run` | 무엇이 제거될지 보여줌 |

### 분석 도구

#### 1. knip - 종합 Dead Code 탐지

```bash
# 설치
npm install -D knip

# 전체 분석
npx knip

# JSON 리포트
npx knip --reporter json > .jikime/cleanup/knip-report.json
```

**탐지 항목**:
- 미사용 파일
- 미사용 exports
- 미사용 dependencies
- 미사용 devDependencies
- 미사용 types

#### 2. depcheck - 의존성 분석

```bash
# 설치
npm install -D depcheck

# 분석
npx depcheck

# JSON 리포트
npx depcheck --json > .jikime/cleanup/depcheck-report.json
```

**탐지 항목**:
- 미사용 dependencies
- 누락된 dependencies
- Phantom dependencies

#### 3. ts-prune - TypeScript Export 분석

```bash
# 설치
npm install -D ts-prune

# 분석
npx ts-prune

# 필터링
npx ts-prune | grep -v "used in module"
```

**탐지 항목**:
- 미사용 exports
- 미사용 types
- Dead code paths

#### 4. ESLint - 미사용 지시문

```bash
# 미사용 eslint-disable 코멘트 확인
npx eslint . --report-unused-disable-directives
```

### 위험 분류 시스템

#### SAFE (자동 제거 가능)

| 카테고리 | 위험 | 탐지 방법 | 검증 |
|----------|------|-----------|------|
| 미사용 npm deps | 낮음 | depcheck | import 없음 확인 |
| 미사용 devDeps | 낮음 | depcheck | 스크립트에서 사용 없음 |
| 주석 처리된 코드 | 낮음 | Regex 패턴 | 시각적 확인 |
| 미사용 imports | 낮음 | ESLint + knip | 참조 없음 |
| 미사용 eslint-disable | 낮음 | ESLint 리포트 | 지시문 확인 |

#### CAREFUL (확인 필요)

| 카테고리 | 위험 | 탐지 방법 | 검증 |
|----------|------|-----------|------|
| 미사용 exports | 중간 | ts-prune + knip | Grep + git history |
| 미사용 파일 | 중간 | knip | 동적 import 확인 |
| 미사용 types | 중간 | ts-prune | 타입 추론 확인 |
| Dead branches | 중간 | Coverage report | 런타임 테스트 |

#### RISKY (수동 리뷰 필요)

| 카테고리 | 위험 | 탐지 방법 | 검증 |
|----------|------|-----------|------|
| Public API | 높음 | API tests | 통합 테스트 |
| 공유 유틸리티 | 높음 | 크로스 프로젝트 검색 | 이해관계자 리뷰 |
| 동적 imports | 높음 | String 패턴 검색 | 런타임 테스트 |
| Reflection 코드 | 높음 | 패턴 분석 | 전체 테스트 스위트 |

### DDD-Aligned Workflow

```
Phase 1: ANALYZE
    └─ 모든 탐지 도구 병렬 실행
    └─ 위험 분류와 함께 결과 집계
    └─ 영향받는 코드의 테스트 커버리지 확인
    └─ 컨텍스트를 위한 git history 리뷰
         ↓
Phase 2: PRESERVE
    └─ 영향받는 코드에 특성화 테스트 존재 확인
    └─ 백업 브랜치 생성: cleanup/YYYY-MM-DD-HHMM
    └─ 테스트 없으면 현재 동작 문서화
         ↓
Phase 3: IMPROVE
    └─ 카테고리별 제거 (가장 안전한 것부터):
        a. 미사용 npm dependencies
        b. 미사용 devDependencies
        c. 미사용 imports
        d. 미사용 exports
        e. 미사용 files
    └─ 각 카테고리 후:
        - 빌드 실행
        - 전체 테스트 스위트 실행
        - 통과하면 커밋
        - DELETION_LOG.md 업데이트
```

### DELETION_LOG.md 형식

모든 삭제는 `docs/DELETION_LOG.md`에 추적됩니다:

```markdown
# Code Deletion Log

코드 정리 작업의 감사 추적 기록.

---

## [YYYY-MM-DD HH:MM] Cleanup Session

**Operator**: J.A.R.V.I.S. / refactorer agent
**Branch**: cleanup/YYYY-MM-DD-HHMM
**Commit**: abc123def
**Tools**: knip v5.x, depcheck v1.x, ts-prune v0.x

### Summary

| Category | Items | Lines | Size Impact |
|----------|-------|-------|-------------|
| Dependencies | 5 | - | -120 KB |
| DevDependencies | 3 | - | -45 KB |
| Files | 12 | 1,450 | -45 KB |
| Exports | 23 | 89 | - |
| Imports | 45 | 45 | - |
| **Total** | **88** | **1,584** | **-210 KB** |

### Dependencies Removed

| Package | Version | Reason | Alternative |
|---------|---------|--------|-------------|
| lodash | 4.17.21 | Not imported | Use native methods |
| moment | 2.29.4 | Deprecated | date-fns already used |

### Files Deleted

| Path | Lines | Last Modified | Replaced By |
|------|-------|---------------|-------------|
| src/utils/old-helpers.ts | 120 | 2023-08-15 | N/A (unused) |
| src/components/LegacyButton.tsx | 85 | 2023-09-01 | Button.tsx |

### Verification Results

- [x] TypeScript compiles: `npx tsc --noEmit`
- [x] Build succeeds: `npm run build`
- [x] Tests pass: 47/47 (100%)
- [x] No lint errors: `npm run lint`
- [x] Bundle size verified

### Recovery Instructions

```bash
# 이 정리 후 문제 발생 시:
git log --oneline | head -5  # 정리 커밋 찾기
git revert <commit-sha>       # 특정 커밋 되돌리기
npm install                   # 의존성 재설치
npm run build && npm test     # 복구 확인
```
```

### Protected Items

`.jikime/cleanup/protected.yaml`에서 제거 금지 항목 관리:

```yaml
# 제거하면 안 되는 항목
protected:
  dependencies:
    - "@types/*"  # 타입 정의
    - "eslint-*"  # 린팅 인프라

  files:
    - "src/polyfills/*"  # 브라우저 호환성
    - "src/lib/dynamic-*"  # 동적 import 대상

  exports:
    - "src/api/public.ts:*"  # Public API
    - "src/sdk/index.ts:*"   # SDK exports

  patterns:
    - "**/index.ts"  # Barrel files (미사용으로 보일 수 있음)
    - "**/__tests__/*"  # 테스트 유틸리티
```

### Safety Checklist

**제거 전 검증**:
- [ ] 모든 탐지 도구 실행됨
- [ ] 위험 분류 완료
- [ ] 백업 브랜치 생성됨
- [ ] 특성화 테스트 존재 (또는 생성됨)
- [ ] 컨텍스트를 위한 git history 리뷰됨
- [ ] 동적 import 패턴 확인됨
- [ ] Public API 영향 평가됨

**제거 후 검증**:
- [ ] TypeScript 에러 없이 컴파일
- [ ] 빌드 성공
- [ ] 모든 테스트 통과
- [ ] 콘솔 에러 없음
- [ ] 번들 사이즈 측정됨
- [ ] DELETION_LOG.md 업데이트됨
- [ ] 커밋 메시지 상세함

---

## TRUST 5 Integration

| 원칙 | Codemap | Cleanup |
|------|---------|---------|
| **T**ested | 생성된 문서 경로 검증 | 각 제거 후 테스트 실행 |
| **R**eadable | 명확한 구조와 ASCII 다이어그램 | 노이즈 제거, 신호/잡음비 개선 |
| **U**nified | 일관된 문서 형식 | 중복 통합 |
| **S**ecured | 민감 정보 노출 방지 | 취약점 있는 미사용 deps 제거 |
| **T**rackable | 타임스탬프와 버전 관리 | DELETION_LOG.md 감사 추적 |

---

## J.A.R.V.I.S. / F.R.I.D.A.Y. 출력 형식

### J.A.R.V.I.S. (개발)

```markdown
## J.A.R.V.I.S.: Codemap Generation Complete

### Generated Files
| File | Lines | Modules Documented |
|------|-------|-------------------|
| docs/CODEMAPS/INDEX.md | 120 | 5 entry points |
| docs/CODEMAPS/frontend.md | 85 | 12 components |
| docs/CODEMAPS/backend.md | 95 | 8 endpoints |

### Coverage
- Files analyzed: 47
- Modules documented: 25
- Dependencies mapped: 32
- Circular dependencies: 0

### Predictive Suggestions
- Consider documenting workers/ directory
- API rate limiting not documented
```

```markdown
## J.A.R.V.I.S.: Cleanup Scan Complete

### Dead Code Summary
| Category | Found | Risk | Action |
|----------|-------|------|--------|
| Dependencies | 5 | SAFE | Auto-remove |
| Exports | 23 | CAREFUL | Review |
| Files | 12 | CAREFUL | Review |
| Dynamic refs | 2 | RISKY | Skip |

### Recommended Actions

**Immediate (SAFE)**:
1. Remove 5 unused dependencies (-120 KB)
2. Remove 15 unused imports

**Review Required (CAREFUL)**:
1. 12 files appear unused but check git history
2. 23 exports not directly referenced

### Estimated Impact
- Bundle size: -165 KB (~8% reduction)
- Lines of code: -1,539
- Files: -12

Proceed with --safe removal? Use: /jikime:cleanup remove --safe
```

### F.R.I.D.A.Y. (마이그레이션)

```markdown
## F.R.I.D.A.Y.: Migration Cleanup

### Legacy Code Status
| Module | Dead Code | Migrated | Clean |
|--------|-----------|----------|-------|
| Auth | 5 items | Yes | No |
| Users | 0 items | Yes | Yes |
| Products | 12 items | Yes | No |

### Migration-Safe Removal
Only removing code that has been:
- Fully migrated to target framework
- Verified by characterization tests
- Not referenced in migration artifacts
```

---

## Command Comparison

| 상황 | 권장 명령어 |
|------|------------|
| 아키텍처 문서화 필요 | `/jikime:codemap all` |
| 의존성 그래프 시각화 | `/jikime:codemap all --deps` |
| Dead code 현황 파악 | `/jikime:cleanup scan` |
| 안전한 정리 작업 | `/jikime:cleanup remove --safe` |
| 정리 기록 확인 | `/jikime:cleanup log` |
| 리팩토링 전 정리 | `/jikime:cleanup scan` → `/jikime:refactor` |

---

## Related Commands

- `/jikime:docs` - 문서 업데이트 및 동기화
- `/jikime:refactor` - DDD 기반 코드 리팩토링
- `/jikime:3-sync` - SPEC 완료 및 문서 동기화
- `/jikime:learn` - 코드베이스 탐색 및 학습

---

## Related Skills

- `jikime-workflow-codemap` - Codemap 생성 워크플로우
- `jikime-workflow-ddd` - DDD 방법론 (ANALYZE-PRESERVE-IMPROVE)
- `jikime-foundation-quality` - TRUST 5 품질 프레임워크

---

Version: 1.0.0
Last Updated: 2026-01-25
Integration: AST analysis (ts-morph), Dependency graphs (madge), Dead code detection (knip, depcheck, ts-prune)
