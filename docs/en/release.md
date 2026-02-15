# Release Guide

A guide covering JikiME-ADK release deployment methods, sequence, and related information.

## Overview

```
Developer                    GitHub Actions                GitHub
─────────                    ──────────────                ──────
1. Code changes                       │                        │
2. Update fallback version            │                        │
3. commit & push main                 │                        │
4. git tag v0.0.2 ─────────────► release.yml trigger           │
5. git push --tags                    │                        │
                                      ├─► Build (4 binaries)   │
                                      ├─► Generate checksums.txt │
                                      └─► Create GitHub Release ──┤
                                                                ├─► Releases tab
                                      deploy-install.yml ───────┤
                                      (release published trigger) ├─► GitHub Pages
                                                                │   install.sh
```

## Prerequisites

- Go 1.24+ installed
- GitHub CLI (`gh`) installed (optional)
- Repository push permissions
- GitHub Pages enabled (Settings → Pages → Source: GitHub Actions)

## Quick Release (Recommended)

The `release.sh` script performs everything from build to release trigger in one step.

```bash
# Basic usage
.github/scripts/release.sh 0.0.1

# Specify commit message
.github/scripts/release.sh 0.0.2 "feat: add jikime-wt binary support"

# Dry run (preview without actual execution)
.github/scripts/release.sh 0.0.1 --dry-run
```

### Script Execution Order

```
[0] Pre-flight checks   → Verify branch, tag duplicates, required files
[1] Build verification   → go build ./... + ./cmd/jikime-wt
[2] Update version       → Update fallbackVersion with sync-versions.sh
[3] Stage and commit     → git add -A && git commit
[4] Push to remote       → git push origin main
[5] Create and push tag  → git tag vX.Y.Z && git push origin vX.Y.Z
                           → GitHub Actions release.yml auto-triggered
```

### Options

| Option | Description |
|------|------|
| `--dry-run` | Display process without actual execution |
| `--skip-build` | Skip build verification step |
| `--force` | Proceed even without changes |

> **Note**: Grant execution permission on first use: `chmod +x .github/scripts/release.sh`

---

## Release Process (Manual, Step by Step)

Below is the manual process for what `release.sh` performs internally.

### Step 1: Code Changes and Build Verification

```bash
cd /path/to/jikime-adk

# Build verification
go build ./...

# Local test
go run . --version
```

### Step 2: Update fallbackVersion

```bash
# Use automated script (recommended)
.github/scripts/sync-versions.sh 0.0.2

# Verify result
grep fallbackVersion version/version.go
#   const fallbackVersion = "0.0.2"
```

Or manually edit the `fallbackVersion` constant in `version/version.go`.

### Step 3: Commit & Push Changes

```bash
# Stage changed files
git add -A

# Commit (include release content)
git commit -m "chore: bump version to 0.0.2

- fix: use embedded templates for sync-templates
- Other changes..."

# Push to main branch
git push origin main
```

> **Important**: You must push code first before creating the tag.
> If you push the tag first, the Release workflow will build with the previous code.

### Step 4: Create & Push Tag

```bash
# Create tag (v prefix required)
git tag v0.0.2

# Push tag → auto-triggers release.yml
git push --tags
```

### Step 5: Verify Deployment

1. **Release Workflow**: GitHub → Actions → Verify "Release" workflow succeeded
2. **Releases Tab**: GitHub → Releases → Verify `v0.0.2` release was created
3. **Assets Verification**: Verify 8 binaries (jikime-adk 4 + jikime-wt 4) + checksums.txt exist
4. **Install Script**: Verify `deploy-install.yml` workflow succeeded

```bash
# Verify with CLI
gh release view v0.0.2

# Verify install script
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | head -5
```

---

## Release Artifacts

Files automatically generated when a tag is pushed:

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
| `checksums.txt` | - | SHA256 checksums |

### Binary Naming Convention

```
jikime-adk-{GOOS}-{GOARCH}    # Main binary
jikime-wt-{GOOS}-{GOARCH}     # Worktree-dedicated binary
```

### Installation Structure

```
~/.local/bin/
├── jikime-adk          # Main binary
├── jikime → jikime-adk # Symbolic link (shortcut command)
└── jikime-wt           # Worktree-dedicated standalone binary
```

### Build Flags

```bash
# jikime-adk (main)
CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w -X 'jikime-adk/version.buildVersion=${VERSION}'" \
  -o "jikime-adk-${GOOS}-${GOARCH}" .

# jikime-wt (worktree-dedicated)
CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w -X 'jikime-adk/version.buildVersion=${VERSION}'" \
  -o "jikime-wt-${GOOS}-${GOARCH}" ./cmd/jikime-wt
```

| Flag | Purpose |
|------|---------|
| `CGO_ENABLED=0` | Static binary (no external C library dependencies) |
| `-trimpath` | Remove build path information (security) |
| `-s` | Remove symbol table (reduce binary size) |
| `-w` | Remove DWARF debug info (reduce binary size) |
| `-X '...buildVersion=...'` | Version build-time injection |

---

## CI/CD Workflows

### 1. Release Workflow (`.github/workflows/release.yml`)

**Trigger**: `push: tags: ['v*']`

**Jobs**:
1. **build**: 4-platform matrix build (darwin/linux × amd64/arm64) × 2 binaries (jikime-adk + jikime-wt)
2. **release**: Collect artifacts → Generate checksums.txt → Publish GitHub Release

**Permissions**: `contents: write` (required for Release creation)

### 2. Deploy Install Script (`.github/workflows/deploy-install.yml`)

**Triggers**:
- `release: types: [published]` - On release publish
- `push: paths: ['install/install.sh', '.github/workflows/deploy-install.yml']` - On script changes
- `workflow_dispatch` - Manual trigger

**Result**: `https://jikime.github.io/jikime-adk/install.sh`

**Permissions**: `contents: read`, `pages: write`, `id-token: write`

---

## Version Management

### Version Flow

```
git tag v0.0.2 → release.yml → -ldflags injection → buildVersion = "0.0.2"
                                                        │
                                                        ▼
                                              version.String() → "0.0.2"
```

### fallbackVersion vs buildVersion

| | fallbackVersion | buildVersion |
|---|---|---|
| **Location** | `version/version.go` constant | Build-time `-ldflags` injection |
| **Purpose** | Used for `go install` or development builds | Used in Release binaries |
| **Update** | `sync-versions.sh` script | Auto-injected in CI |
| **Priority** | Used only when buildVersion is empty | Always takes precedence |

### sync-versions.sh Usage

```bash
# Run from project root
.github/scripts/sync-versions.sh <X.Y.Z>

# Example
.github/scripts/sync-versions.sh 0.0.2
# Output: Updating fallbackVersion: 2.0.0 → 0.0.2
#         Successfully updated fallbackVersion to 0.0.2
```

**Validation**:
- Semver format validation (X.Y.Z)
- Project root verification (`version/version.go` exists)
- macOS/Linux compatible (sed -i.bak)

---

## Installation Methods

### 1. Install Script (Recommended)

```bash
# Default installation (~/.local/bin)
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash

# Global installation (/usr/local/bin, requires sudo)
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --global

# Install specific version
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --version 0.0.2

# Disable color output
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash -s -- --no-color
```

**Install Script Features**:
- SHA256 checksum verification
- Auto-detect platform/architecture
- 3-retry logic
- Automatic PATH verification and guidance
- Cyberpunk theme output

### 2. go install

```bash
go install github.com/jikime/jikime-adk@latest

# Specific version
go install github.com/jikime/jikime-adk@v0.0.2
```

> Note: `go install` does not inject `buildVersion`, so `fallbackVersion` is used.

### 3. Manual Download

```bash
# Download directly from GitHub Releases
curl -LO https://github.com/jikime/jikime-adk/releases/download/v0.0.2/jikime-adk-darwin-arm64
chmod +x jikime-adk-darwin-arm64
mv jikime-adk-darwin-arm64 ~/.local/bin/jikime-adk
```

---

## Update Mechanism

### Self-Update Command

```bash
# Check for updates
jikime-adk update --check

# Execute update
jikime-adk update

# Sync templates (uses embedded templates)
jikime-adk update --sync-templates
```

### Update Types

| Installer Type | Detection Method | Update Method |
|---|---|---|
| `binary` | Directly downloaded installation | Atomic binary replace + checksum verification |
| `go install` | Under GOPATH/bin | `go install github.com/jikime/jikime-adk@latest` |
| `brew` | `brew list` check | `brew upgrade jikime-adk` |

### Atomic Binary Update Flow

```
1. Check latest release via GitHub API
2. Download platform-matching asset (temp dir)
3. Download checksums.txt & verify SHA256
4. Rename current binary → .bak
5. Copy new binary → current location
6. Verify by running --version
7. On success, delete .bak / On failure, rollback .bak → original location
```

---

## Troubleshooting

### Release Workflow Not Triggered

**Cause**: Tag was pushed before code, or workflow file is not in main

**Solution**:
```bash
# Delete and recreate tag
git tag -d v0.0.2
git push origin :refs/tags/v0.0.2

# Push code first
git push origin main

# Recreate tag
git tag v0.0.2
git push --tags
```

### Nothing in Releases Tab

**Cause**: Workflow file was not in main branch at the time of tag push

**Solution**: Verify `.github/workflows/release.yml` exists in main, then recreate tag

### Deploy Install Script Failed ("Repository not found")

**Cause**: `contents: read` missing in `permissions`

**Solution**: Verify the following permissions in `deploy-install.yml`:
```yaml
permissions:
  contents: read
  pages: write
  id-token: write
```

### "Failed to fetch latest version" When Running Install Script

**Cause**: Repository is private, or GitHub API rate limit

**Solution**:
- Set repository to public
- Or set `GITHUB_TOKEN` environment variable

### `jikime-adk update --sync-templates` Shows "not found"

**Cause**: Previous version used external path lookup logic

**Solution**: Update to v0.0.2+ (uses embedded templates)

### Version Not Displaying Correctly

```bash
# Check current version
jikime-adk --version

# When installed via go install, fallbackVersion is displayed
# In Release binaries, buildVersion is displayed
```

---

## Release Initialization (Delete All Tags/Releases and Restart)

Use this when deleting all existing releases and tags to start fresh.

### Step 1: Delete GitHub Releases

```bash
# List current releases
gh release list

# Delete each release
gh release delete v0.0.1 --yes
gh release delete v0.0.2 --yes
# ... repeat for all existing releases
```

### Step 2: Delete Remote Tags

```bash
# Delete individually
git push origin --delete v0.0.1
git push origin --delete v0.0.2

# Or delete all remote tags at once
git tag -l | xargs -I {} git push origin --delete {}
```

### Step 3: Delete Local Tags

```bash
# Delete individually
git tag -d v0.0.1
git tag -d v0.0.2

# Or delete all local tags at once
git tag -l | xargs git tag -d
```

### Step 4: Verify

```bash
git tag -l          # Verify no local tags
gh release list     # Verify no releases
```

### Step 5: Start Fresh

```bash
# Use release.sh (recommended)
.github/scripts/release.sh 0.0.1

# Or manually:
.github/scripts/sync-versions.sh 0.0.1
git add version/version.go
git commit -m "chore: set version to 0.0.1"
git push origin main
git tag v0.0.1
git push --tags
```

---

## Quick Reference

### Release Checklist

- [ ] Code changes complete and build verified (`go build ./...`)
- [ ] fallbackVersion updated with `sync-versions.sh`
- [ ] All changes committed
- [ ] Pushed to `main` branch
- [ ] `git tag vX.Y.Z` created
- [ ] Tag pushed with `git push --tags`
- [ ] GitHub Actions → Release workflow succeeded
- [ ] Verified 8 binaries (jikime-adk 4 + jikime-wt 4) + checksums.txt in GitHub Releases tab
- [ ] Deploy Install Script workflow succeeded
- [ ] Installation tested with `curl ... | bash`

### Key Commands Summary

```bash
# One-click release (recommended)
.github/scripts/release.sh X.Y.Z

# Or specify commit message
.github/scripts/release.sh X.Y.Z "feat: add new feature"

# Dry run (preview only)
.github/scripts/release.sh X.Y.Z --dry-run

# Verify release
gh release view vX.Y.Z

# Test installation
curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash
```

---

Version: 1.0.0
Last Updated: 2026-01-24
