# Deployment & Build Guide

A guide to JikiME-ADK build, version management, and deployment.

## Version Management

### Architecture

JikiME-ADK manages versions using Go's `-ldflags` build-time injection pattern.

```
version/version.go
├── var buildVersion string       ← Injected at build time via -ldflags
├── const fallbackVersion = "2.0.0"  ← Default value when not injected
└── func String() string          ← Returns buildVersion first, fallback if empty
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
  buildVersion = "1.2.3"  (injected)
         │
         ▼
  version.String() → "1.2.3"
```

When not injected:

```
go build ./...
         │
         ▼
  buildVersion = ""  (empty string)
         │
         ▼
  version.String() → "2.0.0"  (fallback)
```

## Build Commands

### Development Build (Local)

```bash
# Basic build (uses fallback version)
go build -o jikime-adk ./...

# Or specify a particular directory
go build -o jikime-adk .
```

### Release Build (Version Injection)

```bash
# Manual version specification
go build -ldflags "-X 'jikime-adk/version.buildVersion=1.0.0'" -o jikime-adk .

# Automatic version based on Git tag
go build -ldflags "-X 'jikime-adk/version.buildVersion=$(git describe --tags --always)'" -o jikime-adk .

# Git tag + dirty flag
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
| `-s` | Remove symbol table |
| `-w` | Remove DWARF debug info |
| `-trimpath` | Remove build path information |

## CI/CD Integration

### Current CI Pipeline

Currently, Go build in `.github/workflows/ci-universal.yml` performs only a basic build:

```yaml
- name: Build
  run: go build -v ./...
```

### Recommended CI Release Build

To add version injection during release:

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

When running `jikime init`, the version is displayed in the banner:

```go
// cmd/initcmd/init.go
banner.PrintIntro(version.String())
```

The version is rendered in the banner as follows:

```go
// cmd/banner/banner.go
versionStyle := lipgloss.NewStyle().
    Foreground(dimCyan).
    Italic(true)
fmt.Println(versionStyle.Render(fmt.Sprintf("    v%s", version)))
```

### Fallback Version Update

The `fallbackVersion` should be updated for new major/minor releases:

```go
// version/version.go
const fallbackVersion = "2.0.0"  // ← Update to match the release
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
