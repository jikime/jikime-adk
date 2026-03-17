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
  const ptyDir  = path.join('node_modules', 'node-pty')
  // pnpm 경로도 탐색
  const ptyPnpm = path.join('node_modules', '.pnpm')

  const prebuild = path.join(ptyDir, 'prebuilds', `linux-x64`, 'pty.node')
  const release  = path.join(ptyDir, 'build', 'Release', 'pty.node')

  if (!fs.existsSync(prebuild) && !fs.existsSync(release)) {
    console.log('[postinstall] linux: node-pty 네이티브 바이너리 없음 → 소스 컴파일 시작...')
    console.log('[postinstall] 필요 패키지: python3, make, gcc, gcc-c++ (없으면 sudo dnf install -y python3 make gcc gcc-c++)')
    try {
      execSync('npm rebuild node-pty', { stdio: 'inherit' })
      console.log('[postinstall] linux: node-pty 컴파일 완료')
    } catch (e) {
      console.warn('[postinstall] linux: node-pty 컴파일 실패. 수동으로 실행하세요:')
      console.warn('  sudo dnf install -y python3 make gcc gcc-c++')
      console.warn('  pnpm rebuild node-pty   (또는 npm rebuild node-pty)')
    }
  } else {
    console.log('[postinstall] linux: node-pty 바이너리 확인됨')
  }
}
