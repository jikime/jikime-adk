#!/usr/bin/env bash
#
# release.sh - Build, commit, tag, and push for release
#
# Usage:
#   .github/scripts/release.sh <version> [commit-message]
#
# Examples:
#   .github/scripts/release.sh 0.0.1
#   .github/scripts/release.sh 0.0.2 "feat: add jikime-wt binary"
#
# This script performs the complete release process:
#   1. Verify build (jikime-adk + jikime-wt)
#   2. Update fallbackVersion
#   3. Stage and commit all changes
#   4. Push to main
#   5. Create and push tag → triggers GitHub Actions release workflow

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
MAGENTA='\033[0;35m'
DIM='\033[2m'
BOLD='\033[1m'
NC='\033[0m'

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

info()    { echo -e "  ${CYAN}$1${NC}"; }
success() { echo -e "  ${GREEN}$1${NC}"; }
warn()    { echo -e "  ${YELLOW}$1${NC}"; }
error()   { echo -e "  ${RED}Error: $1${NC}" >&2; exit 1; }
step()    { echo -e "\n${MAGENTA}${BOLD}[$1]${NC} $2"; }
dim()     { echo -e "  ${DIM}$1${NC}"; }

usage() {
    echo "Usage: $0 <version> [commit-message]"
    echo ""
    echo "  version          Semantic version (X.Y.Z format, without 'v' prefix)"
    echo "  commit-message   Optional commit message body (default: auto-generated)"
    echo ""
    echo "Options:"
    echo "  --dry-run        Show what would be done without executing"
    echo "  --skip-build     Skip build verification step"
    echo "  --force          Force even if working tree has no changes"
    echo "  --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 0.0.1"
    echo "  $0 0.0.2 \"feat: add worktree management\""
    echo "  $0 0.0.3 --dry-run"
    exit 0
}

# ─── Argument Parsing ─────────────────────────────────────────────

VERSION=""
COMMIT_MSG=""
DRY_RUN=false
SKIP_BUILD=false
FORCE=false

while [ $# -gt 0 ]; do
    case "$1" in
        --dry-run)   DRY_RUN=true; shift ;;
        --skip-build) SKIP_BUILD=true; shift ;;
        --force)     FORCE=true; shift ;;
        --help|-h)   usage ;;
        -*)          error "Unknown option: $1. Use --help for usage." ;;
        *)
            if [ -z "$VERSION" ]; then
                VERSION="$1"
            elif [ -z "$COMMIT_MSG" ]; then
                COMMIT_MSG="$1"
            else
                error "Too many arguments. Use --help for usage."
            fi
            shift
            ;;
    esac
done

if [ -z "$VERSION" ]; then
    usage
fi

# Validate semver format
if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    error "Invalid version format: '$VERSION'. Expected X.Y.Z (e.g., 0.0.1)"
fi

# Set default commit message
if [ -z "$COMMIT_MSG" ]; then
    COMMIT_MSG="chore: release v${VERSION}"
fi

# ─── Pre-flight Checks ───────────────────────────────────────────

cd "$PROJECT_ROOT"

echo ""
echo -e "${MAGENTA}╔════════════════════════════════════════╗${NC}"
echo -e "${MAGENTA}║${NC}  ${CYAN}${BOLD}JikiME-ADK Release${NC} ${DIM}v${VERSION}${NC}        ${MAGENTA}║${NC}"
echo -e "${MAGENTA}╚════════════════════════════════════════╝${NC}"

if [ "$DRY_RUN" = true ]; then
    warn "DRY RUN MODE - No changes will be made"
fi

step "0" "Pre-flight checks"

# Check we're on main branch
CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
if [ "$CURRENT_BRANCH" != "main" ]; then
    warn "Current branch: $CURRENT_BRANCH (expected: main)"
    read -rp "  Continue anyway? [y/N] " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        echo "  Aborted."
        exit 0
    fi
fi

# Check if tag already exists
if git tag -l "v${VERSION}" | grep -q "v${VERSION}"; then
    error "Tag v${VERSION} already exists. Delete it first or use a different version."
fi

# Check for required files
if [ ! -f "version/version.go" ]; then
    error "version/version.go not found. Run from project root."
fi

if [ ! -f "go.mod" ]; then
    error "go.mod not found. Run from project root."
fi

success "Pre-flight checks passed"
dim "Branch: $CURRENT_BRANCH"
dim "Target: v${VERSION}"

# ─── Step 1: Build Verification ──────────────────────────────────

step "1" "Build verification"

if [ "$SKIP_BUILD" = true ]; then
    warn "Skipping build verification (--skip-build)"
else
    info "Building jikime-adk..."
    if ! go build ./... 2>&1; then
        error "Build failed. Fix errors before releasing."
    fi
    success "jikime-adk build OK"

    info "Building jikime-wt..."
    if ! go build ./cmd/jikime-wt 2>&1; then
        error "jikime-wt build failed. Fix errors before releasing."
    fi
    success "jikime-wt build OK"

    # Clean up build artifacts
    rm -f jikime-adk jikime-wt
fi

# ─── Step 2: Update fallbackVersion ──────────────────────────────

step "2" "Update fallbackVersion"

CURRENT_VERSION=$(grep -o 'fallbackVersion = "[^"]*"' version/version.go | grep -o '"[^"]*"' | tr -d '"')
info "Current: $CURRENT_VERSION → New: $VERSION"

if [ "$DRY_RUN" = true ]; then
    dim "Would update version/version.go"
else
    "$SCRIPT_DIR/sync-versions.sh" "$VERSION"
fi

# ─── Step 3: Stage and Commit ─────────────────────────────────────

step "3" "Stage and commit"

if [ "$DRY_RUN" = true ]; then
    dim "Would stage all changes and commit:"
    dim "  Message: $COMMIT_MSG"
else
    # Check if there are changes to commit
    if git diff --quiet && git diff --cached --quiet && [ "$FORCE" != true ]; then
        # Only version/version.go might have changed
        if git diff --quiet version/version.go; then
            warn "No changes detected (use --force to commit anyway)"
        fi
    fi

    # Stage all changes
    git add -A

    # Check if there's anything to commit
    if git diff --cached --quiet; then
        if [ "$FORCE" = true ]; then
            warn "No changes to commit, but continuing (--force)"
        else
            warn "Nothing to commit, skipping..."
        fi
    else
        git commit -m "$COMMIT_MSG"
        success "Committed: $COMMIT_MSG"
    fi
fi

# ─── Step 4: Push to main ─────────────────────────────────────────

step "4" "Push to remote"

if [ "$DRY_RUN" = true ]; then
    dim "Would push to origin/$CURRENT_BRANCH"
else
    info "Pushing to origin/$CURRENT_BRANCH..."
    git push origin "$CURRENT_BRANCH"
    success "Pushed to origin/$CURRENT_BRANCH"
fi

# ─── Step 5: Create and Push Tag ──────────────────────────────────

step "5" "Create and push tag"

if [ "$DRY_RUN" = true ]; then
    dim "Would create tag: v${VERSION}"
    dim "Would push tag to origin"
else
    info "Creating tag v${VERSION}..."
    git tag "v${VERSION}"
    success "Tag created: v${VERSION}"

    info "Pushing tag (triggers release workflow)..."
    git push origin "v${VERSION}"
    success "Tag pushed: v${VERSION}"
fi

# ─── Done ─────────────────────────────────────────────────────────

echo ""
echo -e "  ${GREEN}${BOLD}Release v${VERSION} initiated!${NC}"
echo ""
dim "Next steps:"
dim "  1. Check GitHub Actions: https://github.com/jikime/jikime-adk/actions"
dim "  2. Verify release: gh release view v${VERSION}"
dim "  3. Test install: curl -fsSL https://jikime.github.io/jikime-adk/install.sh | bash"
echo ""
