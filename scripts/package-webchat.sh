#!/usr/bin/env bash
#
# Package webchat source as tar.gz for GitHub Release.
#
# Usage:
#   ./scripts/package-webchat.sh              # Uses version from version/version.go
#   ./scripts/package-webchat.sh 1.7.0        # Explicit version
#
# Output:
#   dist/jikime-webchat-v{VERSION}.tar.gz
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEBCHAT_DIR="${PROJECT_ROOT}/webchat"
DIST_DIR="${PROJECT_ROOT}/dist"

# 버전 결정
if [ -n "${1:-}" ]; then
    VERSION="$1"
else
    VERSION=$(grep 'fallbackVersion' "${PROJECT_ROOT}/version/version.go" | sed -n 's/.*"\(.*\)".*/\1/p')
fi

ASSET_NAME="jikime-webchat-v${VERSION}.tar.gz"

echo "Packaging webchat v${VERSION}..."
echo "  Source: ${WEBCHAT_DIR}"
echo "  Output: ${DIST_DIR}/${ASSET_NAME}"

# webchat 디렉토리 확인
if [ ! -f "${WEBCHAT_DIR}/package.json" ]; then
    echo "ERROR: webchat/package.json not found"
    exit 1
fi

# dist 디렉토리 생성
mkdir -p "$DIST_DIR"

# tar.gz 생성 (node_modules, .next, .env* 제외)
# macOS BSD tar는 --transform 미지원 → -s 사용
tar -czf "${DIST_DIR}/${ASSET_NAME}" \
    -C "${WEBCHAT_DIR}" \
    --exclude='node_modules' \
    --exclude='.next' \
    --exclude='.env' \
    --exclude='.env.local' \
    --exclude='.env.*.local' \
    --exclude='tsconfig.tsbuildinfo' \
    .

# 크기 확인
SIZE=$(du -h "${DIST_DIR}/${ASSET_NAME}" | cut -f1)
echo ""
echo "Done! ${ASSET_NAME} (${SIZE})"
echo ""
echo "Upload to GitHub Release:"
echo "  gh release upload v${VERSION} ${DIST_DIR}/${ASSET_NAME}"
