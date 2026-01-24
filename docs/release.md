# Release Guide

JikiME-ADK 릴리스 배포 방법, 순서, 관련 정보를 정리한 가이드.

## Overview

```
Developer                    GitHub Actions                GitHub
─────────                    ──────────────                ──────
1. 코드 변경                         │                        │
2. fallback 버전 업데이트              │                        │
3. commit & push main                │                        │
4. git tag v0.0.2 ─────────────► release.yml 트리거            │
5. git push --tags                   │                        │
                                     ├─► Build (4 바이너리)     │
                                     ├─► checksums.txt 생성    │
                                     └─► GitHub Release 생성 ──┤
                                                               ├─► Releases 탭
                                     deploy-install.yml ────────┤
                                     (release published 트리거)  ├─► GitHub Pages
                                                               │   install.sh
```

## Prerequisites

- Go 1.24+ 설치
- GitHub CLI (`gh`) 설치 (선택)
- Repository push 권한
- GitHub Pages 활성화 (Settings → Pages → Source: GitHub Actions)

## Release Process (Step by Step)

### Step 1: 코드 변경 및 빌드 확인

```bash
cd /path/to/jikime-adk

# 빌드 확인
go build ./...

# 로컬 테스트
go run . --version
```

### Step 2: fallbackVersion 업데이트

```bash
# 자동 스크립트 사용 (권장)
.github/scripts/sync-versions.sh 0.0.2

# 결과 확인
grep fallbackVersion version/version.go
#   const fallbackVersion = "0.0.2"
```

또는 수동으로 `version/version.go`의 `fallbackVersion` 상수를 직접 수정.

### Step 3: 변경사항 커밋 & 푸시

```bash
# 변경된 파일 스테이징
git add -A

# 커밋 (릴리스 내용 포함)
git commit -m "chore: bump version to 0.0.2

- fix: use embedded templates for sync-templates
- 기타 변경사항..."

# main 브랜치에 푸시
git push origin main
```

> **중요**: 반드시 코드를 먼저 푸시한 후에 태그를 생성해야 합니다.
> 태그를 먼저 푸시하면 Release 워크플로우가 이전 코드로 빌드됩니다.

### Step 4: 태그 생성 & 푸시

```bash
# 태그 생성 (v 접두사 필수)
git tag v0.0.2

# 태그 푸시 → release.yml 자동 트리거
git push --tags
```

### Step 5: 배포 확인

1. **Release Workflow**: GitHub → Actions → "Release" 워크플로우 성공 확인
2. **Releases 탭**: GitHub → Releases → `v0.0.2` 릴리스 생성 확인
3. **Assets 확인**: 8개 바이너리 (jikime-adk 4 + jikime-wt 4) + checksums.txt 존재 확인
4. **Install Script**: `deploy-install.yml` 워크플로우 성공 확인

```bash
# CLI로 확인
gh release view v0.0.2

# 설치 스크립트 확인
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | head -5
```

---

## Release Artifacts

태그 푸시 시 자동 생성되는 파일:

| File | Platform | Architecture |
|------|----------|-------------|
| `jikime-adk-darwin-amd64` | macOS | Intel |
| `jikime-adk-darwin-arm64` | macOS | Apple Silicon |
| `jikime-adk-linux-amd64` | Linux | x86_64 |
| `jikime-adk-linux-arm64` | Linux | ARM64 |
| `jikime-wt-darwin-amd64` | macOS | Intel |
| `jikime-wt-darwin-arm64` | macOS | Apple Silicon |
| `jikime-wt-linux-amd64` | Linux | x86_64 |
| `jikime-wt-linux-arm64` | Linux | ARM64 |
| `checksums.txt` | - | SHA256 체크섬 |

### Binary Naming Convention

```
jikime-adk-{GOOS}-{GOARCH}    # 메인 바이너리
jikime-wt-{GOOS}-{GOARCH}     # Worktree 전용 바이너리
```

### 설치 시 구성

```
~/.local/bin/
├── jikime-adk          # 메인 바이너리
├── jikime → jikime-adk # 심볼릭 링크 (단축 명령어)
└── jikime-wt           # Worktree 전용 독립 바이너리
```

### Build Flags

```bash
# jikime-adk (메인)
CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w -X 'jikime-adk/version.buildVersion=${VERSION}'" \
  -o "jikime-adk-${GOOS}-${GOARCH}" .

# jikime-wt (worktree 전용)
CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w -X 'jikime-adk/version.buildVersion=${VERSION}'" \
  -o "jikime-wt-${GOOS}-${GOARCH}" ./cmd/jikime-wt
```

| Flag | Purpose |
|------|---------|
| `CGO_ENABLED=0` | Static binary (외부 C 라이브러리 의존 없음) |
| `-trimpath` | 빌드 경로 정보 제거 (보안) |
| `-s` | Symbol table 제거 (바이너리 크기 감소) |
| `-w` | DWARF debug info 제거 (바이너리 크기 감소) |
| `-X '...buildVersion=...'` | 버전 빌드 타임 인젝션 |

---

## CI/CD Workflows

### 1. Release Workflow (`.github/workflows/release.yml`)

**Trigger**: `push: tags: ['v*']`

**Jobs**:
1. **build**: 4개 플랫폼 매트릭스 빌드 (darwin/linux × amd64/arm64) × 2 바이너리 (jikime-adk + jikime-wt)
2. **release**: 아티팩트 수집 → checksums.txt 생성 → GitHub Release 발행

**Permissions**: `contents: write` (Release 생성에 필요)

### 2. Deploy Install Script (`.github/workflows/deploy-install.yml`)

**Triggers**:
- `release: types: [published]` - 릴리스 발행 시
- `push: paths: ['install/install.sh', '.github/workflows/deploy-install.yml']` - 스크립트 변경 시
- `workflow_dispatch` - 수동 트리거

**Result**: `https://jikime.github.io/jikime-adk/install.sh`

**Permissions**: `contents: read`, `pages: write`, `id-token: write`

---

## Version Management

### Version Flow

```
git tag v0.0.2 → release.yml → -ldflags 인젝션 → buildVersion = "0.0.2"
                                                        │
                                                        ▼
                                              version.String() → "0.0.2"
```

### fallbackVersion vs buildVersion

| | fallbackVersion | buildVersion |
|---|---|---|
| **위치** | `version/version.go` 상수 | 빌드 타임 `-ldflags` 인젝션 |
| **용도** | `go install` 또는 개발 빌드 시 사용 | Release 바이너리에서 사용 |
| **업데이트** | `sync-versions.sh` 스크립트 | CI에서 자동 인젝션 |
| **우선순위** | buildVersion이 비어있을 때만 사용 | 항상 우선 |

### sync-versions.sh 사용법

```bash
# 프로젝트 루트에서 실행
.github/scripts/sync-versions.sh <X.Y.Z>

# 예시
.github/scripts/sync-versions.sh 0.0.2
# Output: Updating fallbackVersion: 2.0.0 → 0.0.2
#         Successfully updated fallbackVersion to 0.0.2
```

**Validation**:
- Semver 형식 검증 (X.Y.Z)
- 프로젝트 루트 확인 (`version/version.go` 존재)
- macOS/Linux 호환 (sed -i.bak)

---

## Installation Methods

### 1. Install Script (권장)

```bash
# 기본 설치 (~/.local/bin)
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash

# 글로벌 설치 (/usr/local/bin, sudo 필요)
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --global

# 특정 버전 설치
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --version 0.0.2

# 컬러 출력 비활성화
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --no-color
```

**Install Script Features**:
- SHA256 체크섬 검증
- 플랫폼/아키텍처 자동 감지
- 3회 재시도 로직
- PATH 자동 확인 및 안내
- Cyberpunk 테마 출력

### 2. go install

```bash
go install github.com/jikime/jikime-adk@latest

# 특정 버전
go install github.com/jikime/jikime-adk@v0.0.2
```

> Note: `go install`은 `buildVersion`을 인젝션하지 않으므로 `fallbackVersion`이 사용됩니다.

### 3. 수동 다운로드

```bash
# GitHub Releases에서 직접 다운로드
curl -LO https://github.com/jikime/jikime-adk/releases/download/v0.0.2/jikime-adk-darwin-arm64
chmod +x jikime-adk-darwin-arm64
mv jikime-adk-darwin-arm64 ~/.local/bin/jikime-adk
```

---

## Update Mechanism

### Self-Update Command

```bash
# 업데이트 확인
jikime-adk update --check

# 업데이트 실행
jikime-adk update

# 템플릿 싱크 (임베디드 템플릿 사용)
jikime-adk update --sync-templates
```

### Update Types

| Installer Type | 감지 방법 | 업데이트 방식 |
|---|---|---|
| `binary` | 직접 다운로드 설치 | Atomic binary replace + checksum 검증 |
| `go install` | GOPATH/bin 하위 | `go install github.com/jikime/jikime-adk@latest` |
| `brew` | `brew list` 확인 | `brew upgrade jikime-adk` |

### Atomic Binary Update Flow

```
1. GitHub API로 최신 릴리스 확인
2. 플랫폼 매칭 asset 다운로드 (temp dir)
3. checksums.txt 다운로드 & SHA256 검증
4. 현재 바이너리 → .bak 이름 변경
5. 새 바이너리 → 현재 위치에 복사
6. --version 실행으로 검증
7. 성공 시 .bak 삭제 / 실패 시 .bak → 원래 위치 롤백
```

---

## Troubleshooting

### Release Workflow가 트리거되지 않음

**원인**: 태그를 코드보다 먼저 푸시했거나, workflow 파일이 main에 없음

**해결**:
```bash
# 태그 삭제 후 재생성
git tag -d v0.0.2
git push origin :refs/tags/v0.0.2

# 코드 먼저 푸시
git push origin main

# 태그 재생성
git tag v0.0.2
git push --tags
```

### Release 탭에 아무것도 없음

**원인**: 워크플로우 파일이 태그 푸시 시점에 main 브랜치에 없었음

**해결**: `.github/workflows/release.yml`이 main에 존재하는지 확인 후 태그 재생성

### Deploy Install Script 실패 ("Repository not found")

**원인**: `permissions`에 `contents: read`가 빠짐

**해결**: `deploy-install.yml`에 아래 권한 확인:
```yaml
permissions:
  contents: read
  pages: write
  id-token: write
```

### Install Script 실행 시 "Failed to fetch latest version"

**원인**: Repository가 private이거나, GitHub API rate limit

**해결**:
- Repository를 public으로 설정
- 또는 `GITHUB_TOKEN` 환경변수 설정

### `jikime-adk update --sync-templates`가 "not found"

**원인**: 이전 버전에서 외부 경로를 탐색하는 로직 사용

**해결**: v0.0.2+ 로 업데이트 (임베디드 템플릿 사용)

### 버전이 올바르게 표시되지 않음

```bash
# 현재 버전 확인
jikime-adk --version

# go install로 설치한 경우 fallbackVersion이 표시됨
# Release 바이너리에서는 buildVersion이 표시됨
```

---

## Release 초기화 (태그/릴리스 전체 삭제 후 재시작)

기존 릴리스와 태그를 모두 삭제하고 처음부터 다시 시작할 때 사용합니다.

### Step 1: GitHub Releases 삭제

```bash
# 현재 릴리스 목록 확인
gh release list

# 각 릴리스 삭제
gh release delete v0.0.1 --yes
gh release delete v0.0.2 --yes
# ... 존재하는 모든 릴리스에 대해 반복
```

### Step 2: 원격 태그 삭제

```bash
# 개별 삭제
git push origin --delete v0.0.1
git push origin --delete v0.0.2

# 또는 모든 원격 태그 한번에 삭제
git tag -l | xargs -I {} git push origin --delete {}
```

### Step 3: 로컬 태그 삭제

```bash
# 개별 삭제
git tag -d v0.0.1
git tag -d v0.0.2

# 또는 모든 로컬 태그 한번에 삭제
git tag -l | xargs git tag -d
```

### Step 4: 확인

```bash
git tag -l          # 로컬 태그 없음 확인
gh release list     # 릴리스 없음 확인
```

### Step 5: 새로 시작

```bash
# 버전 설정
.github/scripts/sync-versions.sh 0.0.1

# 커밋 & 푸시
git add version/version.go
git commit -m "chore: set version to 0.0.1"
git push origin main

# 태그 생성 & 릴리스 트리거
git tag v0.0.1
git push --tags
```

---

## Quick Reference

### 릴리스 체크리스트

- [ ] 코드 변경 완료 및 빌드 확인 (`go build ./...`)
- [ ] `sync-versions.sh`로 fallbackVersion 업데이트
- [ ] 모든 변경사항 커밋
- [ ] `main` 브랜치에 푸시
- [ ] `git tag vX.Y.Z` 생성
- [ ] `git push --tags`로 태그 푸시
- [ ] GitHub Actions → Release 워크플로우 성공 확인
- [ ] GitHub Releases 탭에서 바이너리 8개 (jikime-adk 4 + jikime-wt 4) + checksums.txt 확인
- [ ] Deploy Install Script 워크플로우 성공 확인
- [ ] `curl ... | bash`로 설치 테스트

### 핵심 명령어 요약

```bash
# 버전 업데이트
.github/scripts/sync-versions.sh X.Y.Z

# 커밋 & 푸시
git add -A && git commit -m "chore: release vX.Y.Z" && git push origin main

# 태그 & 릴리스 트리거
git tag vX.Y.Z && git push --tags

# 릴리스 확인
gh release view vX.Y.Z

# 설치 테스트
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash
```

---

Version: 1.0.0
Last Updated: 2026-01-24
