/**
 * jikime serve 프로세스 관리
 * 서버사이드 전용
 */
import { spawn } from 'child_process'
import { writeFile, mkdir } from 'fs/promises'
import { join } from 'path'
import { EventEmitter } from 'events'
import type { Project } from './store'

// Next.js 재시작 시에도 PID 참조 유지용 메모리 맵
const procs = new Map<string, { pid: number }>()

// ─── 로그 버퍼 + 이벤트 ──────────────────────────────────────────────────────

const MAX_LOG_LINES = 300
const logBuffers = new Map<string, string[]>()
const logEmitters = new Map<string, EventEmitter>()

function getEmitter(projectId: string): EventEmitter {
  if (!logEmitters.has(projectId)) {
    const ee = new EventEmitter()
    ee.setMaxListeners(50)
    logEmitters.set(projectId, ee)
  }
  return logEmitters.get(projectId)!
}

function appendLog(projectId: string, line: string) {
  if (!line.trim()) return
  if (!logBuffers.has(projectId)) logBuffers.set(projectId, [])
  const buf = logBuffers.get(projectId)!
  buf.push(line)
  if (buf.length > MAX_LOG_LINES) buf.shift()
  getEmitter(projectId).emit('log', line)
}

/** 버퍼에 쌓인 로그 반환 */
export function getLogBuffer(projectId: string): string[] {
  return logBuffers.get(projectId) ?? []
}

/** 로그 이벤트 구독 */
export function subscribeLog(projectId: string, cb: (line: string) => void): () => void {
  const ee = getEmitter(projectId)
  ee.on('log', cb)
  return () => ee.off('log', cb)
}

// ─── WORKFLOW.md 템플릿 ───────────────────────────────────────────────────────

function buildWorkflow(p: Project): string {
  const [owner, repo] = p.repo.split('/')
  return `---
tracker:
  kind: github
  # api_key 는 GITHUB_TOKEN 환경변수로 전달
  project_slug: ${p.repo}
  active_states:
    - jikime-todo
  terminal_states:
    - jikime-done
    - Done

polling:
  interval_ms: 15000

workspace:
  root: /tmp/jikime-${repo}

hooks:
  after_create: |
    git clone https://github.com/${owner}/${repo}.git .
    echo "[after_create] cloned repo to $(pwd)"

  before_run: |
    echo "[before_run] syncing to latest main..."
    git fetch origin
    git checkout main
    git reset --hard origin/main
    echo "[before_run] ready at $(git rev-parse --short HEAD)"

  after_run: |
    echo "[after_run] done"
    if [ -d "${p.cwd}/.git" ]; then
      cd "${p.cwd}" && git pull --ff-only 2>&1 \\
        && echo "[after_run] local repo synced at $(git rev-parse --short HEAD)" \\
        || echo "[after_run] git pull skipped (local changes or diverged branch)"
    fi

  timeout_ms: 60000

agent:
  max_concurrent_agents: 1
  max_turns: 10
  max_retry_backoff_ms: 300000

claude:
  command: claude
  turn_timeout_ms: 3600000
  stall_timeout_ms: 180000

server:
  port: ${p.port}
---

You are an autonomous software engineer working on a GitHub issue.

Repository: https://github.com/${owner}/${repo}

## Issue

**{{ issue.identifier }}**: {{ issue.title }}

{{ issue.description }}

## Instructions

1. Read the issue carefully and implement what is requested.
2. Create a feature branch: \`git checkout -b fix/issue-{{ issue.id }}\`
3. Make your changes using the available file tools.
4. Commit: \`git add -A && git commit -m "fix: {{ issue.identifier }} - {{ issue.title }}"\`
5. Push the branch: \`git push origin fix/issue-{{ issue.id }}\`
6. Create a pull request:
   \`gh pr create --title "fix: {{ issue.title }}" --body "Closes #{{ issue.id }}" --base main --head fix/issue-{{ issue.id }}\`
7. Merge the pull request and delete the branch:
   \`gh pr merge --squash --delete-branch --admin\`

Work in the current directory. The repository has already been cloned here.
`
}

// ─── 퍼블릭 API ──────────────────────────────────────────────────────────────

/** WORKFLOW.md 생성 후 jikime serve 시작 */
export async function startServe(p: Project): Promise<number> {
  await mkdir(p.cwd, { recursive: true })

  const workflowPath = join(p.cwd, 'WORKFLOW.md')
  await writeFile(workflowPath, buildWorkflow(p), 'utf8')

  // 이전 로그 초기화
  logBuffers.set(p.id, [])

  const jikim = findJikime()

  const proc = spawn(jikim, ['serve'], {
    cwd: p.cwd,
    env: { ...process.env, GITHUB_TOKEN: p.token },
    stdio: ['ignore', 'pipe', 'pipe'],
    detached: false,
  })

  if (!proc.pid) throw new Error('jikime serve 시작 실패')

  procs.set(p.id, { pid: proc.pid })

  // stdout / stderr 모두 캡처
  const handleChunk = (chunk: Buffer) => {
    const text = chunk.toString()
    text.split('\n').forEach(line => appendLog(p.id, line))
  }
  proc.stdout?.on('data', handleChunk)
  proc.stderr?.on('data', handleChunk)

  proc.on('exit', (code) => {
    appendLog(p.id, `[serve] 프로세스 종료 (exit code: ${code ?? 'unknown'})`)
    procs.delete(p.id)
  })

  return proc.pid
}

/** jikime serve 중지 */
export function stopServe(projectId: string, pid: number): void {
  try { process.kill(pid, 'SIGTERM') } catch { /* already dead */ }
  procs.delete(projectId)
}

/** pid로 실행 중 여부 확인 */
export function isRunning(pid: number): boolean {
  try { process.kill(pid, 0); return true } catch { return false }
}

/** 포트로 jikime serve 응답 여부 확인 (pid 없을 때 fallback) */
export async function isPortRunning(port: number): Promise<boolean> {
  try {
    const res = await fetch(`http://127.0.0.1:${port}/api/v1/state`, {
      signal: AbortSignal.timeout(1500),
    })
    return res.ok
  } catch {
    return false
  }
}

/** 포트를 점유한 프로세스 pid 조회 (macOS/Linux: lsof) */
export function getPidFromPort(port: number): number | null {
  try {
    const { execSync } = require('child_process') as typeof import('child_process')
    const out = execSync(`lsof -ti tcp:${port}`, { timeout: 3000 }).toString().trim()
    const pid = parseInt(out.split('\n')[0], 10)
    return isNaN(pid) ? null : pid
  } catch {
    return null
  }
}

function findJikime(): string {
  const candidates = [
    join(process.env.HOME ?? '~', 'go', 'bin', 'jikime'),
    '/usr/local/bin/jikime',
    'jikime',
  ]
  for (const c of candidates) {
    try {
      require('child_process').execSync(`${c} version`, { stdio: 'pipe', timeout: 3000 })
      return c
    } catch { /* continue */ }
  }
  return 'jikime'
}
