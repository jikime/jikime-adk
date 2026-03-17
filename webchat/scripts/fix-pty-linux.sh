#!/usr/bin/env bash
# node-pty Rocky Linux / RHEL 계열 수동 빌드 스크립트
# 사용: bash scripts/fix-pty-linux.sh

set -e
cd "$(dirname "$0")/.."

echo "=== node-pty Linux 빌드 ==="

# ── 1. 빌드 도구 확인 ───────────────────────────────────────────
MISSING=()
command -v python3 &>/dev/null || MISSING+=(python3)
command -v make    &>/dev/null || MISSING+=(make)
command -v gcc     &>/dev/null || MISSING+=(gcc)
command -v g++     &>/dev/null || MISSING+=(gcc-c++)

if [ ${#MISSING[@]} -gt 0 ]; then
  echo "[!] 누락된 빌드 도구: ${MISSING[*]}"
  echo "[!] 아래 명령으로 설치 후 다시 실행하세요:"
  echo "    sudo dnf install -y python3 make gcc gcc-c++"
  exit 1
fi
echo "[ok] 빌드 도구 확인됨"

# ── 2. node-pty 경로 탐색 ────────────────────────────────────────
# pnpm virtual store 우선, 없으면 일반 node_modules
PTY_DIR=""

# pnpm store
if [ -d "node_modules/.pnpm" ]; then
  PTY_DIR=$(find node_modules/.pnpm -maxdepth 3 -name "node-pty" -type d 2>/dev/null | grep "node_modules/node-pty$" | head -1)
fi

# npm / yarn
if [ -z "$PTY_DIR" ] && [ -d "node_modules/node-pty" ]; then
  PTY_DIR="node_modules/node-pty"
fi

if [ -z "$PTY_DIR" ]; then
  echo "[!] node-pty 경로를 찾지 못했습니다."
  echo "[!] pnpm install 을 먼저 실행하세요."
  exit 1
fi
echo "[ok] node-pty 경로: $PTY_DIR"

# ── 3. 이미 빌드됐는지 확인 ─────────────────────────────────────
RELEASE="$PTY_DIR/build/Release/pty.node"
PREBUILD="$PTY_DIR/prebuilds/linux-x64/pty.node"
if [ -f "$RELEASE" ] || [ -f "$PREBUILD" ]; then
  echo "[ok] pty.node 이미 존재 — 재빌드 건너뜀"
  echo "    서버를 재시작하면 터미널이 활성화됩니다."
  exit 0
fi

# ── 4. node-gyp 빌드 ────────────────────────────────────────────
echo "[..] node-gyp rebuild 시작 (잠시 걸릴 수 있습니다)"
cd "$PTY_DIR"

# node-gyp가 없으면 로컬 설치
if ! command -v node-gyp &>/dev/null; then
  npx node-gyp rebuild
else
  node-gyp rebuild
fi

echo ""
if [ -f "build/Release/pty.node" ]; then
  echo "[ok] 빌드 성공: build/Release/pty.node"
  echo "[ok] 서버를 재시작하면 터미널이 활성화됩니다."
else
  echo "[!] 빌드는 완료됐지만 pty.node 파일을 확인할 수 없습니다."
  echo "    build/ 디렉터리를 직접 확인해 보세요."
fi
