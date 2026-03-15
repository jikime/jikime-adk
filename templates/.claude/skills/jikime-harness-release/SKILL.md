---
name: jikime-harness-release
description: Harness Engineering release manager — generates CHANGELOG from pm:OK tasks, bumps version, creates git tag, and writes release notes
version: 1.0.0
category: harness
tags: ["harness", "release", "changelog", "version", "tag", "plans.md", "pm:OK", "harness-engineering"]
triggers:
  keywords:
    - "harness-release"
    - "harness release"
    - "릴리스"
    - "release"
    - "버전 릴리스"
    - "changelog 생성"
    - "배포 준비"
    - "ship"
  phases: ["sync"]
  agents: ["orchestrator", "manager-git", "documenter"]
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
---

# Harness Release — 릴리스 자동화 스킬

## Quick Reference

Plans.md의 `pm:OK` 태스크를 기반으로 릴리스를 자동화합니다.
CHANGELOG 업데이트, 버전 범프, git 태그, 릴리스 노트 생성까지 처리합니다.

**사용법:**
```
/jikime:harness-release                → 대화형 릴리스 (버전 자동 제안)
/jikime:harness-release --patch        → 패치 버전 (v1.0.0 → v1.0.1)
/jikime:harness-release --minor        → 마이너 버전 (v1.0.0 → v1.1.0)
/jikime:harness-release --major        → 메이저 버전 (v1.0.0 → v2.0.0)
/jikime:harness-release --dry-run      → 릴리스 미리보기 (변경사항 없음)
/jikime:harness-release --version v2.1.0 → 직접 버전 지정
```

---

## 릴리스 전제 조건

```bash
# 1. Plans.md에 pm:OK 태스크가 1개 이상 존재
grep -c "pm:OK" Plans.md

# 2. cc:WIP 태스크가 없음 (진행 중인 작업 없음)
grep "cc:WIP" Plans.md | wc -l   # → 0이어야 함

# 3. pm:REVIEW 태스크가 없음 (검토 대기 없음)
grep "pm:REVIEW" Plans.md | wc -l  # → 0이어야 함 (또는 사용자 확인)

# 4. 워킹 디렉토리 클린
git status --porcelain | wc -l   # → 0이어야 함
```

전제 조건 미충족 시:
```
⚠️  릴리스 전제 조건 미충족:
  - cc:WIP 태스크 2개 존재 (Task 2.1, 2.2)
  - pm:REVIEW 태스크 1개 대기 중 (Task 1.3)

강제 진행하시겠습니까? (권장하지 않음)
```

---

## 실행 흐름

### Step 1: 릴리스 대상 수집

Plans.md에서 릴리스할 태스크 목록 추출:

```bash
# pm:OK 태스크 전체
grep "pm:OK" Plans.md

# 마지막 릴리스 이후 pm:OK가 된 태스크
# (git tag 기준)
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$LAST_TAG" ]; then
  git log ${LAST_TAG}..HEAD --oneline -- Plans.md
fi
```

**수집 결과:**
```
릴리스 대상 태스크 (pm:OK):
  ✅ 1.1 — JWT 토큰 생성 (commit: abc1234)
  ✅ 1.2 — 로그인 API (commit: def5678)
  ✅ 1.3 — 리프레시 토큰 (commit: ghi9012)
  ✅ 2.1 — 통합 테스트 추가 (commit: jkl3456)
```

### Step 2: 버전 결정

**현재 버전 감지:**

```bash
# package.json
node -p "require('./package.json').version" 2>/dev/null

# go.mod
grep "^module" go.mod | head -1

# version 파일
cat version/VERSION 2>/dev/null || cat VERSION 2>/dev/null

# git tag
git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
```

**버전 범프 자동 제안:**

태스크 분석으로 적절한 버전 변화 추천:

| 태스크 내용 분석 | 권장 버전 범프 |
|----------------|--------------|
| `BREAKING CHANGE` 포함 커밋 존재 | major |
| 신규 기능 (`feat:`) 커밋 존재 | minor |
| 버그수정/개선만 존재 | patch |

```
현재 버전: v1.2.0
권장 버전: v1.3.0 (minor — 신규 기능 3개 포함)

버전을 선택하세요:
1. v1.3.0 (권장: minor)
2. v1.2.1 (patch)
3. v2.0.0 (major)
4. 직접 입력
```

### Step 3: CHANGELOG.md 생성/업데이트

**CHANGELOG 항목 형식:**

```markdown
## [v1.3.0] — 2026-03-15

### ✨ 새 기능
- JWT 토큰 생성 및 검증 구현 ([1.1] abc1234)
- 로그인 API 엔드포인트 추가 ([1.2] def5678)
- 리프레시 토큰 지원 ([1.3] ghi9012)

### 🧪 테스트
- 통합 테스트 커버리지 80% 달성 ([2.1] jkl3456)

### 📋 Plans.md 참조
이 릴리스의 상세 구현 현황은 Plans.md의 Phase 1, Phase 2를 참조하세요.
```

**커밋 메시지 → CHANGELOG 카테고리 매핑:**

| 커밋 프리픽스 | CHANGELOG 섹션 |
|-------------|---------------|
| `feat:` | ✨ 새 기능 |
| `fix:` | 🐛 버그 수정 |
| `perf:` | ⚡ 성능 개선 |
| `test:` | 🧪 테스트 |
| `docs:` | 📚 문서 |
| `refactor:` | 🔧 개선 |
| `security:` | 🔒 보안 |
| `BREAKING CHANGE` | 💥 주요 변경 (호환성 깨짐) |
| `chore(plans):` | 건너뜀 (Plans.md 관리 커밋) |

**기존 CHANGELOG.md가 있으면 맨 위에 추가, 없으면 신규 생성.**

### Step 4: 버전 파일 업데이트

감지된 버전 파일 유형에 따라 자동 업데이트:

```bash
# package.json
npm version ${NEW_VERSION} --no-git-tag-version 2>/dev/null

# version/VERSION 또는 VERSION
echo "${NEW_VERSION}" > version/VERSION

# go: 태그만 사용 (go.mod 수정 불필요)

# Cargo.toml
sed -i '' "s/^version = .*/version = \"${NEW_VERSION_BARE}\"/" Cargo.toml
```

### Step 5: Plans.md 릴리스 기록

Plans.md 하단에 릴리스 메타데이터 추가:

```markdown
---

## Release: v1.3.0 — 2026-03-15

| 항목 | 내용 |
|------|------|
| **버전** | v1.3.0 |
| **릴리스 태스크** | 4개 (Task 1.1, 1.2, 1.3, 2.1) |
| **git 태그** | v1.3.0 (commit: xyz7890) |
| **CHANGELOG** | [CHANGELOG.md](./CHANGELOG.md) |

릴리스된 태스크는 pm:OK 상태로 유지됩니다.
다음 개발 사이클을 시작하려면 /jikime:harness-plan create로 새 Plans.md를 생성하거나,
새 Phase를 추가하세요.
```

### Step 6: Git 커밋 및 태그

```bash
# 변경된 파일 커밋
git add CHANGELOG.md Plans.md package.json   # (해당하는 것만)
git commit -m "chore(release): v${NEW_VERSION}

Release v${NEW_VERSION} — ${TASK_COUNT}개 태스크 완료.

Included tasks:
$(for task in ${TASK_LIST}; do echo "  - [${task}] ${TASK_CONTENT[$task]}"; done)

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"

# 태그 생성
git tag -a "v${NEW_VERSION}" -m "Release v${NEW_VERSION}"

# 푸시 (사용자 확인 후)
```

**푸시 전 반드시 확인:**
```
태그 v1.3.0을 원격에 푸시하시겠습니까?
  git push origin main
  git push origin v1.3.0

(y/n)
```

### Step 7: 릴리스 완료 보고

```
✅ Harness Release v1.3.0 완료

릴리스 내역:
  버전:     v1.3.0
  태스크:   4개 (모두 pm:OK)
  커밋:     xyz7890
  태그:     v1.3.0 ✅
  CHANGELOG: 업데이트됨 ✅

다음 단계:
  /jikime:harness-sync retro  → 레트로스펙티브 생성
  /jikime:harness-plan create → 다음 개발 사이클 시작
  gh release create v1.3.0   → GitHub Release 생성 (선택)
```

---

## --dry-run 모드

실제 변경 없이 릴리스 내용 미리보기:

```
🔍 Harness Release 미리보기 (dry-run)

버전: v1.2.0 → v1.3.0

CHANGELOG 추가 예정:
---
## [v1.3.0] — 2026-03-15
### ✨ 새 기능
- JWT 토큰 생성 ([1.1] abc1234)
...
---

변경 예정 파일:
  CHANGELOG.md (추가)
  Plans.md (릴리스 기록 추가)
  package.json (version: "1.3.0")

git 태그: v1.3.0

실제 릴리스를 진행하려면 --dry-run 없이 실행하세요.
```

---

## 릴리스 후 Plans.md 상태

릴리스 후 Plans.md는 그대로 유지됩니다:

```
| 1.1 | JWT 토큰 생성 | ... | pm:OK |   ← 유지
| 1.2 | 로그인 API    | ... | pm:OK |   ← 유지
```

다음 개발 사이클:
- **계속 추가**: `harness-plan add`로 새 태스크 추가
- **새 시작**: 새 Plans.md 생성 (이전 Plans.md는 `Plans-v1.3.0.md`로 보관)

---

## 통합 포인트

| 스킬/커맨드 | 연관 방식 |
|-------------|-----------|
| `jikime-harness-sync` | 릴리스 전 sync retro 권장 |
| `jikime-harness-plan` | 릴리스 후 다음 사이클 시작 |
| `jikime-workflow-changelog` | CHANGELOG 생성 패턴 공유 |
| `/jikime:3-sync` | harness-release 진입점 중 하나 |

---

Version: 1.0.0
Status: Active
Last Updated: 2026-03-15
