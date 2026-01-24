# Deployment & Build Guide

JikiME-ADK 빌드, 버전 관리, 배포에 대한 가이드.

## Version Management

### Architecture

JikiME-ADK는 Go의 `-ldflags` 빌드 타임 인젝션 패턴을 사용하여 버전을 관리합니다.

```
version/version.go
├── var buildVersion string       ← 빌드 시 -ldflags로 주입
├── const fallbackVersion = "2.0.0"  ← 주입되지 않았을 때 기본값
└── func String() string          ← buildVersion 우선, 없으면 fallback 반환
```

### Source Code

```go
// version/version.go
package version

// buildVersion is injected at build time via -ldflags
var buildVersion string

const fallbackVersion = "2.0.0"

func String() string {
    if buildVersion != "" {
        return buildVersion
    }
    return fallbackVersion
}
```

### Version Flow

```
go build -ldflags "-X 'jikime-adk/version.buildVersion=1.2.3'"
         │
         ▼
  buildVersion = "1.2.3"  (주입됨)
         │
         ▼
  version.String() → "1.2.3"
```

주입하지 않은 경우:

```
go build ./...
         │
         ▼
  buildVersion = ""  (빈 문자열)
         │
         ▼
  version.String() → "2.0.0"  (fallback)
```

## Build Commands

### Development Build (Local)

```bash
# 기본 빌드 (fallback 버전 사용)
go build -o jikime-adk ./...

# 또는 특정 디렉토리 지정
go build -o jikime-adk .
```

### Release Build (Version Injection)

```bash
# 수동 버전 지정
go build -ldflags "-X 'jikime-adk/version.buildVersion=1.0.0'" -o jikime-adk .

# Git 태그 기반 자동 버전
go build -ldflags "-X 'jikime-adk/version.buildVersion=$(git describe --tags --always)'" -o jikime-adk .

# Git 태그 + dirty flag
go build -ldflags "-X 'jikime-adk/version.buildVersion=$(git describe --tags --always --dirty)'" -o jikime-adk .
```

### Cross-Platform Build

```bash
# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags "-X 'jikime-adk/version.buildVersion=1.0.0'" -o jikime-adk-darwin-arm64 .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'jikime-adk/version.buildVersion=1.0.0'" -o jikime-adk-darwin-amd64 .

# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -ldflags "-X 'jikime-adk/version.buildVersion=1.0.0'" -o jikime-adk-linux-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags "-X 'jikime-adk/version.buildVersion=1.0.0'" -o jikime-adk.exe .
```

### Optimized Release Build

```bash
# Strip debug info + version injection
go build \
  -ldflags "-s -w -X 'jikime-adk/version.buildVersion=1.0.0'" \
  -trimpath \
  -o jikime-adk .
```

| Flag | Purpose |
|------|---------|
| `-s` | Symbol table 제거 |
| `-w` | DWARF debug info 제거 |
| `-trimpath` | 빌드 경로 정보 제거 |

## CI/CD Integration

### Current CI Pipeline

현재 `.github/workflows/ci-universal.yml`에서 Go 빌드는 기본 빌드만 수행합니다:

```yaml
- name: Build
  run: go build -v ./...
```

### Recommended CI Release Build

릴리스 시 버전 인젝션을 추가하려면:

```yaml
- name: Build Release
  run: |
    VERSION=$(git describe --tags --always --dirty)
    go build -ldflags "-s -w -X 'jikime-adk/version.buildVersion=${VERSION}'" \
      -trimpath -o jikime-adk .

- name: Upload Artifact
  uses: actions/upload-artifact@v4
  with:
    name: jikime-adk-${{ runner.os }}
    path: jikime-adk*
```

### GitHub Release Workflow (Example)

```yaml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [darwin, linux, windows]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          EXT=""
          if [ "$GOOS" = "windows" ]; then EXT=".exe"; fi
          go build -ldflags "-s -w -X 'jikime-adk/version.buildVersion=${VERSION}'" \
            -trimpath -o "jikime-adk-${GOOS}-${GOARCH}${EXT}" .

      - name: Upload Release Asset
        uses: softprops/action-gh-release@v2
        with:
          files: jikime-adk-*
```

## Version Display

### Banner Integration

`jikime init` 실행 시 버전이 배너에 표시됩니다:

```go
// cmd/initcmd/init.go
banner.PrintIntro(version.String())
```

배너에서 버전은 다음과 같이 렌더링됩니다:

```go
// cmd/banner/banner.go
versionStyle := lipgloss.NewStyle().
    Foreground(dimCyan).
    Italic(true)
fmt.Println(versionStyle.Render(fmt.Sprintf("    v%s", version)))
```

### Fallback Version Update

새 메이저/마이너 릴리스 시 `fallbackVersion`을 업데이트해야 합니다:

```go
// version/version.go
const fallbackVersion = "2.0.0"  // ← 릴리스에 맞춰 업데이트
```

## Module Information

| Property | Value |
|----------|-------|
| Module | `jikime-adk` |
| Go Version | 1.24.0 |
| Version Package | `jikime-adk/version` |
| ldflags Variable | `jikime-adk/version.buildVersion` |
| Fallback Version | `2.0.0` |

---

Version: 1.0.0
Last Updated: 2026-01-24
