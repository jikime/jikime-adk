# JikiME-ADK Worktree Management

Git Worktree 기반의 병렬 SPEC 개발을 위한 기능입니다.

## 개요

JikiME-ADK의 worktree 기능은 여러 SPEC을 동시에 개발할 수 있도록 Git worktree를 관리합니다. 각 SPEC은 독립적인 브랜치와 작업 디렉토리를 가지며, 메인 저장소와 동기화할 수 있습니다.

## 사용 흐름

### 구성 요소

| 구성요소 | 설명 |
|----------|------|
| **CLI 도구** | `jikime worktree` (Go로 구현된 실제 명령어) |
| **스킬** | `jikime-workflow-worktree` (Claude가 worktree 사용법을 아는 지식) |

### Claude Code 연동 방식

worktree는 **자동 실행이 아닌 요청 기반**으로 동작합니다.

```
1. 사용자 요청
   "SPEC-001용 worktree 만들어줘"
        ↓
2. Claude가 스킬 로드
   Skill("jikime-workflow-worktree") 자동 활성화
        ↓
3. Claude가 CLI 실행
   Bash("jikime worktree new SPEC-001")
        ↓
4. 결과 확인 및 안내
   "SPEC-001 worktree가 생성되었습니다. 경로: ~/jikime/worktrees/..."
```

### 실행 시점

| 시점 | 명령어 | 트리거 |
|------|--------|--------|
| SPEC 개발 시작 | `worktree new SPEC-001` | 사용자 요청 |
| 베이스 브랜치 변경 시 | `worktree sync SPEC-001` | 사용자 요청 |
| 개발 완료 후 | `worktree done SPEC-001` | 사용자 요청 |
| 정리 필요 시 | `worktree clean --stale` | 사용자 요청 |

### 직접 CLI 사용

터미널에서 직접 실행할 수도 있습니다:

```bash
# CLI 직접 사용
jikime worktree new SPEC-001
jikime worktree sync SPEC-001
jikime worktree done SPEC-001
```

## 주요 기능

| 기능 | 설명 |
|------|------|
| 자동 LLM 설정 복사 | `.claude/settings.local.json` 자동 감지 및 복사 |
| 다양한 Sync 전략 | merge, rebase, squash, fast-forward 지원 |
| 충돌 자동 해결 | 3단계 충돌 해결 전략 (ours → theirs → markers 제거) |
| 배치 작업 | `sync --all`, `clean --stale` 지원 |
| Registry 복구 | 디스크에서 worktree 자동 복구 |

## CLI 명령어

### worktree new - Worktree 생성

```bash
# 기본 생성
jikime worktree new SPEC-001

# 커스텀 브랜치명 지정
jikime worktree new SPEC-001 --branch feature/custom-branch

# 특정 브랜치에서 생성
jikime worktree new SPEC-001 --base develop

# LLM 설정 파일 지정
jikime worktree new SPEC-001 --llm-config ~/.claude/my-settings.json

# 강제 재생성
jikime worktree new SPEC-001 --force
```

**자동 LLM 설정 복사**: `--llm-config`를 지정하지 않아도 메인 저장소의 `.claude/settings.local.json`이 존재하면 자동으로 새 worktree에 복사됩니다.

### worktree sync - 베이스 브랜치와 동기화

```bash
# 기본 동기화 (merge)
jikime worktree sync SPEC-001

# Rebase 전략 사용
jikime worktree sync SPEC-001 --rebase

# Fast-forward만 허용
jikime worktree sync SPEC-001 --ff-only

# Squash 전략 (모든 커밋을 하나로)
jikime worktree sync SPEC-001 --squash

# 다른 베이스 브랜치 지정
jikime worktree sync SPEC-001 --base develop

# 충돌 자동 해결
jikime worktree sync SPEC-001 --auto-resolve

# 모든 worktree 동기화
jikime worktree sync --all
```

### worktree clean - Worktree 정리

```bash
# 병합된 브랜치의 worktree만 정리
jikime worktree clean --merged-only

# 오래된 worktree 정리 (기본 30일)
jikime worktree clean --stale

# 오래된 worktree 정리 (14일)
jikime worktree clean --stale --days 14

# 대화형 정리
jikime worktree clean --interactive

# 모든 worktree 정리
jikime worktree clean
```

### worktree list - Worktree 목록 조회

```bash
jikime worktree list
```

### worktree status - Worktree 상태 확인

```bash
jikime worktree status SPEC-001
```

### worktree go - Worktree로 이동

```bash
# 디렉토리 경로 출력
jikime worktree go SPEC-001

# 실제 이동 (shell eval 사용)
eval $(jikime worktree go SPEC-001)
```

### worktree remove - Worktree 삭제

```bash
# 기본 삭제 (커밋되지 않은 변경 확인)
jikime worktree remove SPEC-001

# 강제 삭제
jikime worktree remove SPEC-001 --force
```

### worktree done - 작업 완료 및 병합

```bash
# 메인 브랜치로 병합
jikime worktree done SPEC-001

# 병합 후 푸시
jikime worktree done SPEC-001 --push

# 강제 병합
jikime worktree done SPEC-001 --force
```

### worktree recover - Registry 복구

```bash
jikime worktree recover
```

### worktree config - 설정 확인

```bash
jikime worktree config
jikime worktree config root
jikime worktree config registry
```

## 아키텍처

### 디렉토리 구조

```
~/jikime/worktrees/           # Worktree 루트 (우선순위 1)
~/worktrees/                  # Worktree 루트 (우선순위 2)
├── {project-name}/
│   ├── SPEC-001/
│   │   ├── .git             # Git worktree
│   │   ├── .claude/
│   │   │   └── settings.local.json  # 자동 복사된 LLM 설정
│   │   └── ...
│   └── SPEC-002/
└── .jikime-worktree-registry.json  # Registry 파일
```

### Registry 구조

```json
{
  "worktrees": {
    "project-name": {
      "SPEC-001": {
        "spec_id": "SPEC-001",
        "path": "/path/to/worktree",
        "branch": "feature/SPEC-001",
        "created_at": "2024-01-01T00:00:00Z",
        "last_accessed": "2024-01-02T00:00:00Z",
        "status": "active"
      }
    }
  }
}
```

## Sync 전략

| 전략 | 플래그 | 설명 |
|------|--------|------|
| Merge | (기본) | 히스토리 보존, 병합 커밋 생성 |
| Rebase | `--rebase` | 선형 히스토리, 커밋 재작성 |
| Squash | `--squash` | 모든 변경을 단일 커밋으로 병합 |
| Fast-forward | `--ff-only` | Fast-forward 가능할 때만 동기화 |

## 충돌 해결

`--auto-resolve` 플래그 사용 시 3단계 충돌 해결:

1. **Our Changes**: 현재 브랜치의 변경 사항 유지 (`git checkout --ours`)
2. **Their Changes**: 베이스 브랜치의 변경 사항 적용 (`git checkout --theirs`)
3. **Marker Removal**: Git 충돌 마커 제거 후 커밋

자동 해결 실패 시 수동 해결이 필요합니다.

## LLM 설정 자동 복사

worktree 생성 시 다음 순서로 LLM 설정을 처리합니다:

1. `--llm-config` 플래그로 명시적 지정 → 해당 파일 복사
2. 플래그 미지정 → 메인 저장소의 `.claude/settings.local.json` 자동 감지 및 복사
3. 파일이 없으면 복사하지 않음

복사 시 환경 변수(`${VAR_NAME}`)가 자동으로 치환됩니다.

## 모범 사례

1. **SPEC별 독립 개발**: 각 SPEC은 독립된 worktree에서 개발
2. **정기적 동기화**: `sync --all`로 모든 worktree를 주기적으로 동기화
3. **정리 자동화**: `clean --stale --days 14`로 오래된 worktree 정리
4. **브랜치 명명**: 기본 브랜치명 `feature/{SPEC-ID}` 사용 권장

## 문제 해결

### Registry 손상

```bash
jikime worktree recover
```

### Worktree 상태 불일치

```bash
git worktree prune
jikime worktree recover
```

### 충돌 해결 실패

```bash
cd /path/to/worktree
git merge --abort  # 또는 git rebase --abort
# 수동으로 충돌 해결
```

---

Version: 1.0.0
Last Updated: 2026-01-22
