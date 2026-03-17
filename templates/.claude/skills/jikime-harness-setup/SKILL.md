---
name: jikime-harness-setup
description: Harness Engineering project initializer — validates environment, registers hooks, and prepares Plans.md workspace for the full Plan→Work→Review→Ship workflow
version: 1.0.0
category: harness
tags: ["harness", "setup", "init", "initialization", "plans.md", "hooks", "harness-engineering"]
triggers:
  keywords:
    - "harness-setup"
    - "harness setup"
    - "harness init"
    - "하네스 설정"
    - "하네스 초기화"
    - "워크플로우 설정"
    - "setup harness"
  phases: ["plan"]
  agents: ["orchestrator", "manager-project"]
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
---

# Harness Setup — 프로젝트 초기화 스킬

## Quick Reference

새 프로젝트 또는 기존 프로젝트에 Harness Engineering 워크플로우를 설정합니다.
Plans.md 생성, hooks 검증, git 환경 준비를 한 번에 처리합니다.

**사용법:**
```
/jikime:harness-setup           → 전체 설정 (대화형)
/jikime:harness-setup --check   → 현재 설정 상태만 진단
/jikime:harness-setup --reset   → 설정 초기화 후 재설정
```

---

## 실행 흐름

### Step 1: 환경 진단 (Diagnostic)

```bash
# 1. git 저장소 확인
git rev-parse --git-dir 2>/dev/null || echo "NOT_GIT"

# 2. jikime 바이너리 확인
which jikime || echo "NOT_FOUND"

# 3. .claude/settings.json 확인
ls .claude/settings.json 2>/dev/null || echo "NOT_FOUND"

# 4. Plans.md 존재 여부
ls Plans.md 2>/dev/null || echo "NOT_FOUND"

# 5. hook 등록 상태 확인
grep -c "plans-watcher\|guardrail-engine" .claude/settings.json 2>/dev/null || echo "0"
```

**진단 결과 출력:**

```
🔍 Harness Engineering 환경 진단

✅ git 저장소: 감지됨 (main 브랜치)
✅ jikime 바이너리: v1.5.0
⚠️  .claude/settings.json: 있음 (hooks 미등록)
❌ Plans.md: 없음
✅ worktree 지원: 사용 가능

설정이 필요한 항목: 2개
계속하시겠습니까? (y/n)
```

### Step 2: Plans.md 생성 (미존재 시)

Plans.md가 없으면 자동으로 `harness-plan create` 흐름을 안내:

```
Plans.md가 없습니다.
harness-plan create로 Plans.md를 생성하거나,
기존 프로젝트 구조를 분석해서 자동 생성할 수 있습니다.

1. 자동 분석 후 Plans.md 생성
2. 빈 Plans.md 템플릿 생성
3. 건너뜀 (나중에 /jikime:harness-plan create 사용)
```

**자동 분석 시:**

```bash
# 기술 스택 감지
ls package.json go.mod pyproject.toml Cargo.toml 2>/dev/null

# 최근 이슈/PR 분석 (GitHub 연결 시)
gh issue list --state open --limit 10 2>/dev/null

# git log로 최근 작업 파악
git log --oneline -10
```

생성된 Plans.md는 프로젝트 컨텍스트 기반으로 Phase 1~3 자동 구성.

### Step 3: Hook 등록 검증 및 수정

`.claude/settings.json`의 PostToolUse 섹션에 두 hook이 등록됐는지 확인:

**필수 hooks:**
```json
{
  "type": "command",
  "command": "jikime hooks plans-watcher",
  "timeout": 5000
},
{
  "type": "command",
  "command": "jikime hooks guardrail-engine",
  "timeout": 10000
}
```

미등록 시 자동으로 추가:
```
⚠️  plans-watcher hook이 등록되지 않았습니다.
    .claude/settings.json에 자동 추가하시겠습니까? (y/n)
```

### Step 4: Git 환경 준비

```bash
# worktree 지원 확인
git worktree list

# harness 작업 브랜치 네임스페이스 예약 확인
# (기존 브랜치와 충돌 없는지)
git branch -a | grep "harness/" || echo "clean"

# .gitignore에 임시 파일 추가
grep -q "harness-tmp" .gitignore 2>/dev/null || \
  echo "\n# Harness Engineering\n.harness-tmp/" >> .gitignore
```

### Step 5: 설정 완료 보고

```
✅ Harness Engineering 설정 완료

환경:
  git 브랜치:    main
  Plans.md:     생성됨 (8개 태스크)
  hooks:        plans-watcher ✅ guardrail-engine ✅
  worktree:     준비됨

시작하려면:
  /jikime:harness-plan create    → 새 Plans.md 생성
  /jikime:harness-work --auto    → TODO 태스크 자동 실행
  /jikime:harness-review --all   → 리뷰 대기 태스크 처리
```

---

## --check 모드 (진단만)

```
/jikime:harness-setup --check

Harness Engineering 설정 상태:

✅ git 저장소
✅ jikime v1.5.0
✅ .claude/settings.json
✅ plans-watcher hook
✅ guardrail-engine hook
✅ Plans.md (12개 태스크)
  └─ TODO: 4 | WIP: 0 | DONE: 6 | OK: 2

워크플로우 상태: 정상 ✅
```

---

## 설정 파일 구조

```
프로젝트 루트/
├── Plans.md                    ← 태스크 관리 (SSOT)
├── .claude/
│   ├── settings.json           ← plans-watcher, guardrail-engine hooks
│   └── skills/
│       ├── jikime-harness-plan/
│       ├── jikime-harness-work/
│       ├── jikime-harness-review/
│       ├── jikime-harness-sync/
│       └── jikime-harness-release/
└── .gitignore                  ← .harness-tmp/ 추가
```

---

## 통합 포인트

| 스킬/커맨드 | 연관 방식 |
|-------------|-----------|
| `jikime-harness-plan` | setup 완료 후 가장 먼저 실행 |
| `plans-watcher hook` | setup이 등록 여부 검증 |
| `guardrail-engine hook` | setup이 등록 여부 검증 |
| `/jikime:0-project` | 프로젝트 초기화 → harness-setup 호출 |

---

Version: 1.0.0
Status: Active
Last Updated: 2026-03-15
