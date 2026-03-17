#!/usr/bin/env node
const { execSync } = require('child_process')
const os = require('os')
const fs = require('fs')
const path = require('path')

const platform = os.platform()

// macOS: spawn-helper 실행 권한 부여
if (platform === 'darwin') {
  const spawnHelper = path.join('node_modules', 'node-pty', 'prebuilds', 'darwin-arm64', 'spawn-helper')
  if (fs.existsSync(spawnHelper)) {
    try {
      execSync(`chmod +x ${spawnHelper}`)
      console.log('[postinstall] darwin: spawn-helper +x OK')
    } catch { /* */ }
  }
}

// Linux: node-pty 네이티브 모듈 소스 컴파일
if (platform === 'linux') {
  // pnpm virtual store 우선, 없으면 일반 node_modules
  function findPtyDir() {
    // pnpm store
    const pnpmStore = path.join('node_modules', '.pnpm')
    if (fs.existsSync(pnpmStore)) {
      try {
        const entries = fs.readdirSync(pnpmStore)
        for (const entry of entries) {
          if (!entry.startsWith('node-pty@')) continue
          const candidate = path.join(pnpmStore, entry, 'node_modules', 'node-pty')
          if (fs.existsSync(candidate)) return candidate
        }
      } catch { /* */ }
    }
    // npm / yarn
    const direct = path.join('node_modules', 'node-pty')
    if (fs.existsSync(direct)) return direct
    return null
  }

  const ptyDir = findPtyDir()
  if (!ptyDir) {
    console.log('[postinstall] linux: node-pty 경로를 찾지 못함 — 건너뜀')
    process.exit(0)
  }

  const prebuild = path.join(ptyDir, 'prebuilds', 'linux-x64', 'pty.node')
  const release  = path.join(ptyDir, 'build', 'Release', 'pty.node')

  if (fs.existsSync(prebuild) || fs.existsSync(release)) {
    console.log('[postinstall] linux: node-pty 바이너리 확인됨')
    process.exit(0)
  }

  console.log('[postinstall] linux: node-pty 네이티브 바이너리 없음 → node-gyp 컴파일 시작...')
  console.log('[postinstall] 필요 패키지: python3, make, gcc, gcc-c++ (없으면 sudo dnf install -y python3 make gcc gcc-c++)')
  console.log(`[postinstall] 경로: ${ptyDir}`)

  try {
    // node-gyp rebuild를 node-pty 디렉터리 안에서 직접 실행
    execSync('npx --yes node-gyp rebuild', {
      stdio: 'inherit',
      cwd: ptyDir,
    })
    console.log('[postinstall] linux: node-pty 컴파일 완료')
  } catch (e) {
    console.warn('[postinstall] linux: node-pty 컴파일 실패. 수동으로 실행하세요:')
    console.warn('  sudo dnf install -y python3 make gcc gcc-c++')
    console.warn('  pnpm fix-pty   (또는 bash scripts/fix-pty-linux.sh)')
  }
}
