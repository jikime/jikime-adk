#!/usr/bin/env bash
#
# sync-versions.sh - Update fallbackVersion in version/version.go
#
# Usage:
#   .github/scripts/sync-versions.sh <version>
#
# Examples:
#   .github/scripts/sync-versions.sh 2.1.0
#   .github/scripts/sync-versions.sh 3.0.0
#
# This script updates the fallbackVersion constant in version/version.go
# to match the specified semver version. Used during release process.

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

readonly VERSION_FILE="version/version.go"

usage() {
    echo "Usage: $0 <version>"
    echo ""
    echo "  version    Semantic version (X.Y.Z format)"
    echo ""
    echo "Examples:"
    echo "  $0 2.1.0"
    echo "  $0 3.0.0"
    exit 1
}

error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

info() {
    echo -e "${CYAN}$1${NC}"
}

success() {
    echo -e "${GREEN}$1${NC}"
}

# Validate arguments
if [ $# -ne 1 ]; then
    usage
fi

VERSION="$1"

# Validate semver format (X.Y.Z)
if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    error "Invalid version format: '$VERSION'. Expected X.Y.Z (e.g., 2.1.0)"
fi

# Ensure we're in the project root
if [ ! -f "$VERSION_FILE" ]; then
    error "'$VERSION_FILE' not found. Run this script from the project root."
fi

# Get current version
CURRENT=$(grep -o 'fallbackVersion = "[^"]*"' "$VERSION_FILE" | grep -o '"[^"]*"' | tr -d '"')
if [ -z "$CURRENT" ]; then
    error "Could not detect current fallbackVersion in $VERSION_FILE"
fi

info "Updating fallbackVersion: $CURRENT → $VERSION"

# Update version (macOS/Linux compatible sed)
sed -i.bak "s/fallbackVersion = \"$CURRENT\"/fallbackVersion = \"$VERSION\"/" "$VERSION_FILE"
rm -f "${VERSION_FILE}.bak"

# Verify the change
NEW=$(grep -o 'fallbackVersion = "[^"]*"' "$VERSION_FILE" | grep -o '"[^"]*"' | tr -d '"')
if [ "$NEW" != "$VERSION" ]; then
    error "Version update failed. Expected '$VERSION', got '$NEW'"
fi

success "Successfully updated fallbackVersion to $VERSION"
echo ""
echo "  $VERSION_FILE:"
grep "fallbackVersion" "$VERSION_FILE" | sed 's/^/    /'
