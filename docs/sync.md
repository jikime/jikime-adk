# JikiME-ADK Sync 워크플로우 구현 문서

## 개요

`/jikime:3-sync` 커맨드는 코드 변경사항을 문서와 동기화하고, 품질을 검증한 후, Git 작업을 수행하는 워크플로우입니다.

**핵심 철학**: "Sync to Verify to Commit"

## 구현된 파일 목록

### 1. 커맨드

| 파일 | 경로 | 설명 |
|------|------|------|
| `3-sync.md` | `templates/.claude/commands/jikime/` | 메인 Sync 커맨드 (Step 5/5) |

### 2. 컨텍스트

| 파일 | 경로 | 설명 |
|------|------|------|
| `sync.md` | `templates/.claude/contexts/` | Sync 모드 동작 규칙 정의 |

### 3. 에이전트

| 파일 | 경로 | 설명 |
|------|------|------|
| `manager-docs.md` | `templates/.claude/agents/jikime/` | 문서 동기화 전문 에이전트 |
| `manager-quality.md` | `templates/.claude/agents/jikime/` | 품질 검증 전문 에이전트 |
| `manager-git.md` | `templates/.claude/agents/jikime/` | Git 작업 전문 에이전트 |

### 4. 스킬

| 파일 | 경로 | 설명 |
|------|------|------|
| `SKILL.md` | `templates/.claude/skills/jikime-workflow-sync/` | Sync 워크플로우 스킬 |

---

## 아키텍처

### 워크플로우 다이어그램

```
┌─────────────────────────────────────────────────────────────┐
│                    /jikime:3-sync                           │
│                   (Main Entry Point)                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 0.5: Quality Verification                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Tests     │  │   Linter    │  │ Type Check  │         │
│  │   pytest    │  │   ruff      │  │   mypy      │         │
│  │   vitest    │  │   eslint    │  │   tsc       │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────┘
                      │ PASS
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 1: Analysis & Planning                   │
│                                                             │
│  git diff --name-only HEAD                                  │
│  git status --porcelain                                     │
│                     ↓                                       │
│  Documentation Mapping (변경 유형 → 문서 업데이트)           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 2: Execute Sync                          │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                  manager-docs                          │ │
│  │  • README.md 동기화                                    │ │
│  │  • CODEMAP 생성/업데이트                               │ │
│  │  • SPEC 상태 동기화                                    │ │
│  │  • API 문서화                                          │ │
│  └───────────────────────────────────────────────────────┘ │
│                          ↓                                  │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                 manager-quality                        │ │
│  │  • TRUST 5 검증                                        │ │
│  │  • 링크 무결성 확인                                    │ │
│  │  • 일관성 검사                                         │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Phase 3: Git Operations                        │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                  manager-git                           │ │
│  │  • 문서 파일 스테이징                                  │ │
│  │  • 커밋 생성 (HEREDOC 메시지)                          │ │
│  │  • PR 관리 (Team Mode)                                 │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 에이전트 협업 구조

```
┌──────────────────────────────────────────────────────────────┐
│                     Sync Orchestrator                        │
│                    (/jikime:3-sync)                          │
└──────────────────────────┬───────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
           ▼               ▼               ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ manager-docs │  │manager-quality│  │ manager-git  │
│              │  │              │  │              │
│ Tools:       │  │ Tools:       │  │ Tools:       │
│ • Read       │  │ • Read       │  │ • Bash       │
│ • Write      │  │ • Bash       │  │ • Read       │
│ • Edit       │  │ • Grep       │  │ • Grep       │
│ • Bash       │  │ • Glob       │  │ • TodoWrite  │
│ • Grep       │  │ • TodoWrite  │  │              │
│ • Glob       │  │              │  │              │
│ • TodoWrite  │  │              │  │              │
└──────────────┘  └──────────────┘  └──────────────┘
```

---

## 각 컴포넌트 상세

### 1. `/jikime:3-sync` 커맨드

**위치**: `templates/.claude/commands/jikime/3-sync.md`

**핵심 설정**:
```yaml
---
description: "[Step 5/5] Sync docs, verify quality, commit changes."
context: sync
---
```

**실행 모드**:

| 모드 | 플래그 | 설명 |
|------|--------|------|
| Auto | (기본) | 변경된 파일만 동기화 |
| Full | `--full` | 전체 문서 재생성 |
| Status | `--status` | 상태 확인만 (읽기 전용) |

**옵션**:

| 옵션 | 설명 |
|------|------|
| `--skip-quality` | Phase 0.5 품질 검증 건너뛰기 |
| `--commit` | 변경사항 자동 커밋 |
| `--worktree` | Worktree 환경에서 실행 |

---

### 2. `sync.md` 컨텍스트

**위치**: `templates/.claude/contexts/sync.md`

**모드 정의**:
```yaml
Mode: Documentation Synchronization
Focus: Document generation, quality verification, git operations
Methodology: Sync to Verify to Commit
```

**동작 규칙**:

| DO | DON'T |
|----|-------|
| git 변경 분석 후 동기화 | 변경 없는 문서 재생성 |
| 영향받는 문서만 업데이트 | 품질 검증 건너뛰기 |
| 링크 무결성 검증 | 리뷰 없이 커밋 |
| TRUST 5 원칙 준수 | 불필요한 문서 생성 |
| 의미 있는 커밋 메시지 | 기존 링크 파괴 |

---

### 3. manager-docs 에이전트

**위치**: `templates/.claude/agents/jikime/manager-docs.md`

**역할**: Technical Writer & Documentation Architect

**담당 문서 유형**:

1. **README.md**
   - 프로젝트 개요
   - Quick Start 가이드
   - 아키텍처 참조

2. **CODEMAPS**
   ```
   docs/CODEMAPS/
   ├── INDEX.md       # 아키텍처 개요
   ├── frontend.md    # 프론트엔드 구조
   ├── backend.md     # 백엔드 구조
   └── database.md    # 데이터베이스 스키마
   ```

3. **SPEC 상태 동기화**
   - Status: Planning → In Progress → Completed
   - Progress 퍼센트
   - Last Updated 타임스탬프

4. **API 문서**
   - 엔드포인트 설명
   - 파라미터/응답 스키마

---

### 4. manager-quality 에이전트

**위치**: `templates/.claude/agents/jikime/manager-quality.md`

**역할**: Quality Assurance Architect

**TRUST 5 프레임워크**:

| 원칙 | 검증 항목 |
|------|----------|
| **T**ested | 테스트 커버리지 ≥80%, 모든 테스트 통과 |
| **R**eadable | 함수 <50줄, 파일 <400줄, 중첩 <4 |
| **U**nified | 일관된 스타일, 표준 패턴, 중복 없음 |
| **S**ecured | 하드코딩 시크릿 없음, 입력 검증, 인젝션 방지 |
| **T**rackable | 의미 있는 커밋, SPEC 추적, 타임스탬프 |

**승인 기준**:

| 상태 | 조건 |
|------|------|
| APPROVE | CRITICAL/HIGH 이슈 없음 |
| WARNING | MEDIUM 이슈만 있음 |
| BLOCK | CRITICAL/HIGH 이슈 있음 |

---

### 5. manager-git 에이전트

**위치**: `templates/.claude/agents/jikime/manager-git.md`

**역할**: Version Control Specialist

**커밋 메시지 템플릿**:
```
docs: sync documentation with code changes

Synchronized:
- [업데이트된 문서 목록]

SPEC updates:
- [SPEC 상태 변경 사항]

Quality verification:
- Tests: PASS
- Linter: PASS

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

**워크플로우 모드**:

| 모드 | 설명 |
|------|------|
| Personal | 단일 브랜치 직접 커밋 |
| Team | PR 기반 브랜치 워크플로우 |

**안전 규칙**:

❌ **금지 사항**:
- `git push --force` (main/master)
- `git reset --hard` (확인 없이)
- `git checkout .` (모든 변경 폐기)
- 시크릿/자격 증명 커밋
- pre-commit 훅 건너뛰기

✅ **필수 사항**:
- 커밋 전 변경사항 리뷰
- 의미 있는 커밋 메시지
- 푸시 전 테스트
- amend 대신 새 커밋 (요청 시 제외)

---

### 6. jikime-workflow-sync 스킬

**위치**: `templates/.claude/skills/jikime-workflow-sync/SKILL.md`

**트리거**:
```yaml
triggers:
  keywords: ["sync", "동기화", "documentation", "문서", "CODEMAP", "README", "commit"]
  phases: ["sync"]
  agents: ["manager-docs", "manager-quality", "manager-git"]
```

**허용된 도구**:
- Read, Write, Edit
- Bash, Grep, Glob
- TodoWrite

---

## 사용 예시

### 기본 사용

```bash
# 변경된 파일만 동기화 (Auto 모드)
/jikime:3-sync

# 전체 문서 재생성 (Full 모드)
/jikime:3-sync --full

# 상태 확인만 (Status 모드)
/jikime:3-sync --status
```

### 고급 사용

```bash
# 품질 검증 건너뛰기
/jikime:3-sync --skip-quality

# 자동 커밋 포함
/jikime:3-sync --commit

# Worktree 환경에서 실행
/jikime:3-sync --worktree

# 조합 사용
/jikime:3-sync --full --commit
```

### 워크플로우 통합

```bash
# 전체 개발 워크플로우
/jikime:dev-0-init          # Step 1: 프로젝트 초기화
/jikime:dev-1-plan          # Step 2: SPEC 기반 계획
/jikime:dev-2-implement     # Step 3: 구현
/jikime:dev-3-test          # Step 4: 테스트
/jikime:3-sync              # Step 5: 문서 동기화 & 커밋
```

---

## moai-adk 대비 개선사항

| 항목 | moai-adk | jikime-adk-v2 |
|------|----------|---------------|
| 커맨드 파일 크기 | ~1400줄 | ~250줄 |
| 실행 모드 | 4개 (auto, force, status, project) | 3개 (auto, full, status) |
| Context 로딩 | 수동 참조 | 자동 로딩 (`context: sync`) |
| 에이전트 구조 | 단일 파일 내 정의 | 분리된 에이전트 파일 |
| 스킬 시스템 | 없음 | Progressive Disclosure 통합 |
| CLI 통합 | Python (moai) | Go (jikime) |

### 핵심 개선 포인트

1. **간결성**: 불필요한 중복 제거, 핵심 로직에 집중
2. **모듈화**: 에이전트별 책임 분리로 유지보수성 향상
3. **자동화**: Context 자동 로딩으로 사용자 경험 개선
4. **확장성**: 스킬 시스템 통합으로 Progressive Disclosure 지원
5. **성능**: Go 기반 CLI로 빠른 실행

---

## 연관 파일

### 통합 포인트

```
templates/.claude/
├── commands/jikime/
│   ├── 3-sync.md           ← 메인 커맨드
│   ├── dev-0-init.md       ← Step 1
│   ├── dev-1-plan.md       ← Step 2
│   ├── dev-2-implement.md  ← Step 3
│   └── dev-3-test.md       ← Step 4
├── contexts/
│   ├── sync.md             ← Sync 컨텍스트
│   ├── dev.md              ← Dev 컨텍스트
│   └── review.md           ← Review 컨텍스트
├── agents/jikime/
│   ├── manager-docs.md     ← 문서 에이전트
│   ├── manager-quality.md  ← 품질 에이전트
│   └── manager-git.md      ← Git 에이전트
└── skills/
    └── jikime-workflow-sync/
        └── SKILL.md        ← Sync 스킬
```

---

## 트러블슈팅

### 문제: 문서가 동기화되지 않음

**원인**: git diff에서 변경 감지 실패

**해결**:
```bash
# 변경사항 확인
git status --porcelain
git diff --name-only HEAD

# 전체 재생성 모드 사용
/jikime:3-sync --full
```

### 문제: 품질 검증 실패

**원인**: 테스트/린터/타입체크 오류

**해결**:
```bash
# 구체적인 실패 내용 확인 후 수정
# 또는 건너뛰기 (권장하지 않음)
/jikime:3-sync --skip-quality
```

### 문제: 링크 무결성 실패

**원인**: 이동/삭제된 파일 참조

**해결**:
1. 깨진 링크 위치 확인
2. 참조 업데이트
3. 다시 동기화

---

## 버전 정보

- **Version**: 1.0.0
- **Created**: 2026-01-22
- **Last Updated**: 2026-01-22
- **Author**: Claude Opus 4.5

---

## 관련 문서

- [Worktree 워크플로우](./worktree.md)
- [Ralph Loop](./ralph-loop.md)
- [Migration 가이드](./migration.md)
