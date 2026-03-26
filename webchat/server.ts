import { createServer } from 'http'
import next from 'next'
import { WebSocketServer, WebSocket } from 'ws'
import * as fs from 'fs'
import * as path from 'path'
import * as crypto from 'crypto'
import { execFile, execFileSync } from 'child_process'
import * as os from 'os'
import { handleHarnessRoutes } from './harness'

// node-pty 지연 로딩 — Linux에서 컴파일 안 된 경우에도 서버가 정상 기동
// 터미널 기능만 비활성화되고 채팅/파일/Git은 정상 동작
let pty: typeof import('node-pty') | null = null
try {
  pty = require('node-pty')
} catch (e) {
  console.warn('[server] node-pty 로드 실패 — 터미널 기능 비활성화됨')
  console.warn('[server] 해결: sudo dnf install -y python3 make gcc gcc-c++ && pnpm rebuild node-pty')
}

const dev = process.env.NODE_ENV !== 'production'
const hostname = process.env.HOSTNAME || 'localhost'
const port = parseInt(process.env.PORT || '4000', 10)

// ── 프로세스 레벨 에러 핸들러 (서버 크래시 방지) ──────────────────
process.on('uncaughtException', (err) => {
  console.error('[uncaughtException]', err)
})
process.on('unhandledRejection', (reason) => {
  console.error('[unhandledRejection]', reason)
})

const app = next({ dev, hostname, port })
const handle = app.getRequestHandler()

// ── Claude CLI 경로 탐색 ────────────────────────────────────────
// 파일 첫 4바이트로 네이티브 바이너리 여부 확인
// ELF(Linux): 7F 45 4C 46 / Mach-O(macOS): FE ED FA CE/CF 또는 CE/CF FA ED FE
function isNativeBinary(filePath: string): boolean {
  try {
    const buf = Buffer.alloc(4)
    const fd = fs.openSync(filePath, 'r')
    fs.readSync(fd, buf, 0, 4, 0)
    fs.closeSync(fd)
    // ELF
    if (buf[0] === 0x7f && buf[1] === 0x45 && buf[2] === 0x4c && buf[3] === 0x46) return true
    // Mach-O (LE/BE, 32/64bit)
    const magic = buf.readUInt32LE(0)
    if ([0xFEEDFACE, 0xCEFAEDFE, 0xFEEDFACF, 0xCFFAEDFE].includes(magic)) return true
    return false
  } catch { return false }
}

function findClaudePath(): string | undefined {
  // 1. 환경변수로 명시된 경우 최우선 (네이티브 여부 검사 없이 신뢰)
  if (process.env.CLAUDE_PATH) {
    console.log(`[server] claude 경로 (CLAUDE_PATH): ${process.env.CLAUDE_PATH}`)
    return process.env.CLAUDE_PATH
  }

  const check = (p: string): string | undefined => {
    try {
      // symlink는 해소하지 않음 — 래퍼/심링크 경로가 올바른 진입점
      // 단, isNativeBinary 확인은 실제 파일로
      const real = fs.realpathSync(p)
      if (isNativeBinary(real)) {
        console.log(`[server] claude 바이너리: ${p}`)
        return p  // 원본 경로(래퍼/심링크) 반환
      }
    } catch { /* */ }
    return undefined
  }

  // 2. PATH 환경변수 직접 탐색 — which 셸 호출 불필요, 인젝션 위험 없음
  for (const dir of (process.env.PATH ?? '').split(path.delimiter)) {
    if (!dir) continue
    const found = check(path.join(dir, 'claude'))
    if (found) return found
  }

  // 3. 흔한 네이티브 설치 경로 직접 확인
  const guesses = [
    `${os.homedir()}/.local/bin/claude`,           // Linux npm install 기본 경로
    '/usr/local/bin/claude',
    '/usr/bin/claude',
    `${os.homedir()}/.claude/local/claude`,
    `${os.homedir()}/.local/share/claude/claude`,
  ]
  for (const p of guesses) {
    if (fs.existsSync(p)) {
      const found = check(p)
      if (found) return found
    }
  }

  console.warn('[server] claude 네이티브 바이너리를 찾지 못했습니다.')
  console.warn('[server] 해결: CLAUDE_PATH=/path/to/claude pnpm dev  (네이티브 설치 경로 직접 지정)')
  return undefined
}

const CLAUDE_PATH = findClaudePath()

// 시작 시 claude 동작 확인
if (CLAUDE_PATH) {
  try {
    // execFileSync: 셸 미사용 — CLAUDE_PATH 경로에 메타문자 포함 시 인젝션 방지
    const ver = execFileSync(CLAUDE_PATH, ['--version'], { encoding: 'utf8', timeout: 5000 }).trim()
    console.log(`[server] claude 버전: ${ver}`)
  } catch (e: unknown) {
    const err = e as { stderr?: string; stdout?: string; message?: string }
    console.warn(`[server] claude --version 실패: ${err.stderr || err.stdout || err.message}`)
  }
}

// ── PTY Session Store ──────────────────────────────────────────
const MAX_PTY_SESSIONS = 50  // 무제한 터미널 세션 생성으로 인한 fd/메모리 고갈 방지

interface PtySession {
  pty: import('node-pty').IPty
  clients: Set<WebSocket>
  created: number
  lastActivity: number
  _lastInputMs: number  // rate limit 추적
}
const ptySessions = new Map<string, PtySession>()

function getOrCreatePtySession(sessionId: string, cwd: string, cols: number, rows: number): PtySession {
  if (!pty) throw new Error('node-pty unavailable')
  if (ptySessions.has(sessionId)) {
    return ptySessions.get(sessionId)!
  }
  if (ptySessions.size >= MAX_PTY_SESSIONS) {
    throw new Error(`PTY session limit reached (max ${MAX_PTY_SESSIONS})`)
  }

  let ptyProcess: import('node-pty').IPty

  // tmux: prefix → 기존 tmux 세션을 linked-session 으로 뷰어 생성
  if (sessionId.startsWith('tmux:')) {
    const tmuxSession = sessionId.slice(5) // "tmux:jikime-lotto-team-leader" → "jikime-lotto-team-leader"
    // tmux 세션명 검증 — 특수문자(:, {, }, .) 포함 시 tmux 오동작 방지
    if (!/^[a-zA-Z0-9_-]{1,128}$/.test(tmuxSession)) {
      throw new Error(`Invalid tmux session name: ${tmuxSession.slice(0, 40)}`)
    }

    // ── new-session -t 방식 (linked session) ────────────────────────────
    // attach-session 은 모든 클라이언트 중 가장 작은 크기로 윈도우를 제약해
    // 기존 세션이 더 넓은 경우 xterm 에서 줄이 겹쳐 보임.
    // new-session -t <원본세션> 은 같은 window group 을 공유하되
    // 각 세션이 독립적인 크기를 가지므로 줄겹침이 발생하지 않음.
    const linkedName = `web-${Date.now().toString(36)}`
    ptyProcess = pty.spawn('tmux', [
      'new-session',
      '-t', tmuxSession,  // 원본 세션의 window group 에 링크
      '-s', linkedName,   // 웹 뷰어 전용 세션 이름
      '-x', String(cols), // 이 세션만의 너비 — 원본과 독립
      '-y', String(rows), // 이 세션만의 높이
    ], {
      name:              'xterm-256color',
      cols,
      rows,
      cwd:               os.homedir(),
      handleFlowControl: true,
      flowControlPause:  '\x13', // XOFF
      flowControlResume: '\x11', // XON
      env: {
        ...process.env,
        TERM:      'xterm-256color',
        COLORTERM: 'truecolor',
        LANG:      'en_US.UTF-8',
        LC_ALL:    'en_US.UTF-8',
      },
    })

    // PTY 종료 시 linked session 정리 (원본 세션은 그대로 유지)
    ptyProcess.onExit(() => {
      // execFileSync: 셸 없이 tmux 직접 실행 — 향후 linkedName 변경 시 인젝션 방지
      try { execFileSync('tmux', ['kill-session', '-t', linkedName], { stdio: 'ignore' }) } catch { /* 이미 없으면 무시 */ }
    })
  } else {
    // SHELL 환경변수 → 시스템에 실제 존재하는 셸 순서로 폴백
    // Docker(node:slim)에는 zsh 없이 bash/sh만 있음
    const shellCandidates = [
      process.env.SHELL,
      '/bin/bash',
      '/bin/zsh',
      '/bin/sh',
    ]
    const shell = shellCandidates.find(s => s && fs.existsSync(s)) ?? '/bin/sh'
    ptyProcess = pty.spawn(shell, [], {
      name: 'xterm-256color',
      cols,
      rows,
      cwd: fs.existsSync(cwd) ? cwd : os.homedir(),
      env: { ...process.env, TERM: 'xterm-256color', COLORTERM: 'truecolor' },
    })
  }
  const session: PtySession = {
    pty: ptyProcess,
    clients: new Set(),
    created: Date.now(),
    lastActivity: Date.now(),
    _lastInputMs: 0,
  }
  ptyProcess.onData((data: string) => {
    session.lastActivity = Date.now()
    for (const client of session.clients) {
      if (client.readyState === WebSocket.OPEN) {
        client.send(JSON.stringify({ type: 'output', data }))
      }
    }
  })
  ptyProcess.onExit(() => {
    for (const client of session.clients) {
      if (client.readyState === WebSocket.OPEN) {
        client.send(JSON.stringify({ type: 'exit' }))
      }
    }
    ptySessions.delete(sessionId)
  })
  ptySessions.set(sessionId, session)
  return session
}

// ── Permission Store ────────────────────────────────────────────
interface PermissionDecision {
  allow: boolean
  alwaysAllow?: boolean
}
const pendingPermissions = new Map<string, (d: PermissionDecision) => void>()

// ── Claude Chat Session Store ──────────────────────────────────
interface ChatSession {
  ws: WebSocket
  queryInstance: { interrupt?: () => Promise<void> } | null
  claudeSessionId: string | null
}
const chatSessions = new Map<string, ChatSession>()

// ── GitHub Issue Processor Store ────────────────────────────────
interface IssueProcessor {
  status: 'running' | 'done' | 'error'
  events: string[]
  interrupt: (() => Promise<void>) | null
  sseClients: Set<import('http').ServerResponse>
  completedAt?: number
}
const issueProcessors = new Map<string, IssueProcessor>()
// done/error 상태의 processor를 10분 후 자동 삭제 — 메모리 누수 방지
setInterval(() => {
  const now = Date.now()
  for (const [key, proc] of issueProcessors) {
    if (proc.status !== 'running' && proc.completedAt && now - proc.completedAt > 10 * 60 * 1000) {
      issueProcessors.delete(key)
    }
  }
}, 60 * 1000).unref()

// ── GitHub Poller Store (자동 폴링 루프) ──────────────────────────
interface PollerEvent {
  type: 'tick' | 'issue_found' | 'issue_done' | 'error'
  lastCheck?: string
  activeCount?: number
  issueNumber?: number
  issueTitle?: string
  status?: string
  message?: string
}
interface ProjectPoller {
  projectPath: string
  token: string
  owner: string
  repo: string
  intervalMs: number
  maxConcurrent: number
  model: string
  timer: ReturnType<typeof setInterval> | null
  status: 'running' | 'stopped'
  lastCheck: string | null
  activeIssues: Set<number>
  sseClients: Set<import('http').ServerResponse>
}
const projectPollers = new Map<string, ProjectPoller>()  // key: projectPath

// ── Token Budget Extractor ──────────────────────────────────────
function extractTokenBudget(resultMsg: Record<string, unknown>): { used: number; total: number } | null {
  const modelUsage = resultMsg.modelUsage as Record<string, Record<string, number>> | undefined
  if (!modelUsage) return null
  const modelKey = Object.keys(modelUsage)[0]
  if (!modelKey) return null
  const m = modelUsage[modelKey]
  const input    = m.cumulativeInputTokens          ?? m.inputTokens          ?? 0
  const output   = m.cumulativeOutputTokens         ?? m.outputTokens         ?? 0
  const cacheR   = m.cumulativeCacheReadInputTokens ?? m.cacheReadInputTokens ?? 0
  const cacheC   = m.cumulativeCacheCreationInputTokens ?? m.cacheCreationInputTokens ?? 0
  return {
    used:  input + output + cacheR + cacheC,
    total: parseInt(process.env.CONTEXT_WINDOW || '160000', 10),
  }
}

async function handleClaudeMessage(
  ws: WebSocket,
  wsKey: string,
  claudeSessionId: string | null,
  projectPath: string,
  prompt: string,
  model = 'claude-sonnet-4-6',
  permissionMode = 'bypassPermissions',
  extendedThinking = false,
) {
  // cwd 존재 여부 검증 — 없으면 홈 디렉터리로 폴백
  // (Node.js spawn은 cwd가 없으면 ENOENT를 던지는데, SDK가 이를 바이너리 오류로 잘못 해석함)
  const effectiveCwd = fs.existsSync(projectPath) ? projectPath : os.homedir()
  if (effectiveCwd !== projectPath) {
    console.warn(`[chat] cwd does not exist: ${projectPath} — falling back to ${effectiveCwd}`)
  }

  console.log(`[chat] start — model=${model} mode=${permissionMode} thinking=${extendedThinking} cwd=${effectiveCwd}`)
  const { query } = await import('@anthropic-ai/claude-agent-sdk')

  let capturedSessionId = claudeSessionId

  // root 환경에서는 --dangerously-skip-permissions 사용 불가 → acceptEdits 로 대체
  const isRoot = process.getuid?.() === 0
  const effectivePermissionMode =
    isRoot && permissionMode === 'bypassPermissions' ? 'acceptEdits' : permissionMode

  const options: Record<string, unknown> = {
    cwd: effectiveCwd,
    permissionMode: effectivePermissionMode,
    model,
    // user + project settings 로드 — 슬래시 커맨드(/jikime:*, /sc:* 등) 인식에 필수
    settingSources: ['user', 'project'],
    ...(!isRoot && permissionMode === 'bypassPermissions' && { allowDangerouslySkipPermissions: true }),
    ...(CLAUDE_PATH && { pathToClaudeCodeExecutable: CLAUDE_PATH }),
    ...(extendedThinking && { thinking: { type: 'enabled', budget_tokens: 10000 } }),
  }
  if (claudeSessionId) options.resume = claudeSessionId

  // canUseTool — 권한 확인 모드일 때 브라우저에 승인 요청
  if (effectivePermissionMode !== 'bypassPermissions') {
    options.canUseTool = async (toolName: string, input: unknown) => {
      const requestId = `perm-${Date.now()}-${Math.random().toString(36).slice(2)}`
      ws.send(JSON.stringify({ type: 'permission_request', requestId, toolName, input }))

      return new Promise<{ behavior: string; updatedInput?: unknown; message?: string }>((resolve) => {
        // 30초 타임아웃 → 허용으로 처리 (Map에서 즉시 삭제 보장 — 메모리 누수 방지)
        const timer = setTimeout(() => {
          pendingPermissions.delete(requestId)  // resolver 호출 전 선제 삭제
          resolve({ behavior: 'allow', updatedInput: input })
        }, 30000)

        pendingPermissions.set(requestId, ({ allow }) => {
          clearTimeout(timer)
          pendingPermissions.delete(requestId)  // timer와 resolver 양쪽에서 delete 보장
          if (allow) {
            resolve({ behavior: 'allow', updatedInput: input })
          } else {
            resolve({ behavior: 'deny', message: '사용자가 거부했습니다' })
          }
        })
      })
    }
  }

  const queryInstance = query({ prompt, options })

  // 세션에 queryInstance 저장 (abort/interrupt 용)
  const sess = chatSessions.get(wsKey)
  if (sess) {
    sess.queryInstance = queryInstance as { interrupt?: () => Promise<void> }
    sess.claudeSessionId = claudeSessionId
  }

  try {
    for await (const event of queryInstance) {
      const e = event as Record<string, unknown>

      // session_id 최초 캡처
      if (e.session_id && !capturedSessionId) {
        capturedSessionId = e.session_id as string
        const s = chatSessions.get(wsKey)
        if (s) s.claudeSessionId = capturedSessionId
        ws.send(JSON.stringify({ type: 'session_id', sessionId: capturedSessionId }))
      }

      // assistant 블록 — text / tool_use / thinking
      if (e.type === 'assistant') {
        const content = ((e.message as Record<string, unknown>)?.content ?? []) as Record<string, unknown>[]
        for (const block of content) {
          if (block.type === 'text' && block.text) {
            ws.send(JSON.stringify({ type: 'text', text: block.text }))
          } else if (block.type === 'tool_use') {
            ws.send(JSON.stringify({ type: 'tool_call', name: block.name, input: block.input }))
          } else if (block.type === 'thinking' && block.thinking) {
            ws.send(JSON.stringify({ type: 'thinking', text: block.thinking }))
          }
        }
      }

      // user 블록 — tool_result
      if (e.type === 'user') {
        const content = ((e.message as Record<string, unknown>)?.content ?? []) as Record<string, unknown>[]
        for (const block of content) {
          if (block.type === 'tool_result') {
            ws.send(JSON.stringify({ type: 'tool_result', content: block.content }))
          }
        }
      }

      // result — 토큰 사용량 + 완료
      if (e.type === 'result') {
        const tokenBudget = extractTokenBudget(e)
        if (tokenBudget) {
          ws.send(JSON.stringify({ type: 'token_budget', used: tokenBudget.used, total: tokenBudget.total }))
        }
        ws.send(JSON.stringify({ type: 'done', sessionId: capturedSessionId }))
        return
      }
    }

    ws.send(JSON.stringify({ type: 'done', sessionId: capturedSessionId }))
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    console.error('[chat] error:', msg)

    // exit code 1 — claude CLI 실행 실패
    if (msg.includes('exit') && msg.includes('code 1')) {
      // 실제 오류 확인을 위해 직접 실행 (root 여부에 따라 옵션 분기)
      const isRootCheck = process.getuid?.() === 0
      let detail = ''
      try {
        const debugArgArray = isRootCheck
          ? ['--output-format', 'stream-json', '-p', 'ping']
          : ['--dangerously-skip-permissions', '--output-format', 'stream-json', '-p', 'ping']
        // execFileSync: 셸 미사용 — CLAUDE_PATH 경로 인젝션 방지
        execFileSync(CLAUDE_PATH ?? 'claude', debugArgArray, {
          encoding: 'utf8', timeout: 10000,
        })
      } catch (e: unknown) {
        const ce = e as { stderr?: string; stdout?: string; message?: string }
        detail = (ce.stderr || ce.stdout || ce.message || '').trim()
      }
      const hint = [
        `claude 종료 코드 1 (경로: ${CLAUDE_PATH ?? 'claude'})`,
        detail ? `오류: ${detail}` : '원격 서버에서 직접 확인: claude --output-format stream-json -p "hello"',
      ].join('\n')
      console.error('[chat]', hint)
      ws.send(JSON.stringify({ type: 'error', message: hint }))
    } else if (msg.includes('interrupt') || msg.includes('abort') || msg.includes('cancel')) {
      ws.send(JSON.stringify({ type: 'aborted' }))
    } else {
      ws.send(JSON.stringify({ type: 'error', message: msg }))
    }
  } finally {
    console.log('[chat] done —', wsKey)
    const s = chatSessions.get(wsKey)
    if (s) s.queryInstance = null
  }
}

// ── Project Discovery ──────────────────────────────────────────

// 파일시스템을 탐색해 인코딩된 경로를 실제 경로로 복원
// Claude는 경로의 '/'를 '-'로 인코딩 → 경로 내에 '-'가 있으면 단순 치환 불가
// 예) -home-anthony-jikime-adk-webchat → /home/anthony/jikime-adk/webchat
// 파일시스템에서 실제 경로를 찾지 못하면 null 반환 (잘못된 치환 방지)
function decodeProjectPath(encoded: string): string | null {
  if (!encoded.startsWith('-')) return encoded.replace(/-/g, '/')

  // 파일시스템 탐색으로 올바른 경로 복원
  function find(remaining: string, dir: string): string | null {
    if (!remaining) return dir
    if (!remaining.startsWith('-')) return null

    const sub = remaining.slice(1) // leading '-'(= '/') 제거
    let entries: string[]
    try {
      entries = fs.readdirSync(dir)
    } catch { return null }

    // 긴 이름 우선 정렬: 'jikime-adk'(10)가 'jikime'(6)보다 먼저 시도됨
    entries.sort((a, b) => b.length - a.length)

    for (const entry of entries) {
      if (sub === entry || sub.startsWith(entry + '-')) {
        const result = find(sub.slice(entry.length), path.join(dir, entry))
        if (result !== null) return result
      }
    }
    return null
  }

  return find(encoded, '/')
}

function discoverProjects(): Array<{ id: string; name: string; path: string; sessions: string[] }> {
  const claudeDir = path.join(os.homedir(), '.claude', 'projects')
  const projects: Array<{ id: string; name: string; path: string; sessions: string[] }> = []

  try {
    if (!fs.existsSync(claudeDir)) return projects
    const entries = fs.readdirSync(claudeDir)
    for (const entry of entries) {
      const fullPath = path.join(claudeDir, entry)
      if (!fs.statSync(fullPath).isDirectory()) continue

      const sessions: string[] = []
      try {
        const files = fs.readdirSync(fullPath).filter(f => f.endsWith('.jsonl'))
        // mtime을 sort 전에 O(n)으로 사전 계산 — sort 내부에서 반복 statSync 호출 방지
        const withMtime = files.map(f => {
          try { return { f, mtime: fs.statSync(path.join(fullPath, f)).mtimeMs } }
          catch { return { f, mtime: 0 } }
        })
        withMtime.sort((a, b) => b.mtime - a.mtime)
        for (const { f } of withMtime) sessions.push(f.replace('.jsonl', ''))
      } catch { /* */ }

      // 1차: 파일시스템 탐색으로 경로 복원
      // 2차: _webchat_path 파일에 저장된 원본 경로 사용 (경로가 존재하지 않아도 등록 가능)
      let actualPath = decodeProjectPath(entry)
      if (actualPath === null) {
        try {
          const metaFile = path.join(fullPath, '_webchat_path')
          if (fs.existsSync(metaFile)) {
            actualPath = fs.readFileSync(metaFile, 'utf8').trim()
          }
        } catch { /* */ }
      }
      if (actualPath === null) continue  // 경로 복원 실패 시 건너뜀

      projects.push({
        id: entry,
        name: path.basename(actualPath) || entry,
        path: actualPath,
        sessions,
      })
    }
  } catch { /* */ }

  return projects
}

// ── File Tree ──────────────────────────────────────────────────
interface FileNode {
  name: string
  path: string
  type: 'file' | 'directory'
  size?: number
  modified?: number
  children?: FileNode[]
}

function getFileTree(dirPath: string, depth = 0, maxDepth = 3): FileNode[] {
  if (depth >= maxDepth) return []
  const ignored = new Set(['node_modules', '.git', '.next', 'dist', 'build', '__pycache__', '.DS_Store'])

  try {
    const entries = fs.readdirSync(dirPath, { withFileTypes: true })
    const nodes: FileNode[] = []

    for (const entry of entries) {
      if (ignored.has(entry.name) || entry.name.startsWith('.')) continue
      const fullPath = path.join(dirPath, entry.name)
      try {
        const stat = fs.statSync(fullPath)
        if (entry.isDirectory()) {
          nodes.push({
            name: entry.name,
            path: fullPath,
            type: 'directory',
            modified: stat.mtimeMs,
            children: getFileTree(fullPath, depth + 1, maxDepth),
          })
        } else {
          nodes.push({
            name: entry.name,
            path: fullPath,
            type: 'file',
            size: stat.size,
            modified: stat.mtimeMs,
          })
        }
      } catch { /* */ }
    }

    return nodes.sort((a, b) => {
      if (a.type !== b.type) return a.type === 'directory' ? -1 : 1
      return a.name.localeCompare(b.name)
    })
  } catch { return [] }
}

// ── Session History Parser ─────────────────────────────────────
interface HistoryMessage {
  id: string
  role: 'user' | 'assistant'
  text: string
  status: 'done'
  thinking?: string
  toolCalls?: { name: string; input: unknown; result?: unknown }[]
}

function findSessionJsonl(projectPath: string, sessionId: string): string | null {
  const claudeDir = path.join(os.homedir(), '.claude', 'projects')
  // Claude encodes path by replacing '/' with '-'
  const encoded = projectPath.replace(/\//g, '-')
  const direct = path.join(claudeDir, encoded, `${sessionId}.jsonl`)
  if (fs.existsSync(direct)) return direct
  // Fallback: search all project dirs
  try {
    for (const entry of fs.readdirSync(claudeDir)) {
      const p = path.join(claudeDir, entry, `${sessionId}.jsonl`)
      if (fs.existsSync(p)) return p
    }
  } catch { /* */ }
  return null
}

const MAX_SESSION_FILE_SIZE = 50 * 1024 * 1024 // 50 MB

// mtime 기반 파싱 캐시 — 동일 파일 반복 파싱 방지 (60s TTL + mtime 불일치 시 무효화)
const _sessionHistoryCache = new Map<string, { messages: HistoryMessage[]; mtime: number; ts: number }>()
const SESSION_CACHE_TTL_MS = 60_000

function parseSessionHistory(jsonlPath: string): HistoryMessage[] {
  // 거대 세션 파일 로딩 방지 — 50 MB 초과 시 빈 배열 반환
  let stat: import('fs').Stats
  try {
    stat = fs.statSync(jsonlPath)
    if (stat.size > MAX_SESSION_FILE_SIZE) {
      console.warn(`[history] session file too large (${(stat.size / 1024 / 1024).toFixed(1)} MB), skipping: ${jsonlPath}`)
      return []
    }
  } catch { return [] }

  // 캐시 히트: TTL 내이고 mtime 동일 시 재파싱 생략
  const now    = Date.now()
  const cached = _sessionHistoryCache.get(jsonlPath)
  if (cached && now - cached.ts < SESSION_CACHE_TTL_MS && cached.mtime === stat.mtimeMs) {
    return cached.messages
  }

  const lines = fs.readFileSync(jsonlPath, 'utf8').split('\n').filter(Boolean)
  const messages: HistoryMessage[] = []

  for (const line of lines) {
    let event: Record<string, unknown>
    try { event = JSON.parse(line) } catch { continue }

    // file-history-snapshot / last-prompt 등 건너뜀
    if (event.type !== 'user' && event.type !== 'assistant') continue
    // isMeta: true — 내부 명령 메시지 건너뜀
    if (event.isMeta === true) continue

    const msg = event.message as {
      role?: string
      content?: string | Array<Record<string, unknown>>
    } | undefined
    if (!msg?.content) continue

    if (msg.role === 'user') {
      // content 가 문자열인 경우 (신규 포맷)
      if (typeof msg.content === 'string') {
        const text = msg.content.trim()
        // XML 명령(<command-name>, <local-command-stdout> 등) 건너뜀
        if (!text || text.startsWith('<')) continue
        messages.push({ id: `h-u-${messages.length}`, role: 'user', text, status: 'done' })

      // content 가 배열인 경우 (tool_result / text 블록 혼합)
      } else if (Array.isArray(msg.content)) {
        const content = msg.content
        const toolResults = content.filter(b => b.type === 'tool_result')
        const textBlocks  = content.filter(b => b.type === 'text')

        // tool_result → 앞 assistant 메시지의 마지막 tool call에 결과 연결
        if (toolResults.length > 0) {
          const lastAsst = [...messages].reverse().find(m => m.role === 'assistant')
          if (lastAsst?.toolCalls) {
            toolResults.forEach((tr, i) => {
              const tc = lastAsst.toolCalls![lastAsst.toolCalls!.length - toolResults.length + i]
              if (tc) tc.result = tr.content
            })
          }
        }

        if (textBlocks.length > 0) {
          const text = textBlocks.map(b => b.text as string).join('\n').trim()
          if (text && !text.startsWith('<')) {
            messages.push({ id: `h-u-${messages.length}`, role: 'user', text, status: 'done' })
          }
        }
      }

    } else if (msg.role === 'assistant') {
      if (!Array.isArray(msg.content)) continue
      const content = msg.content
      const textBlocks    = content.filter(b => b.type === 'text')
      const toolUseBlocks = content.filter(b => b.type === 'tool_use')
      const thinkBlocks   = content.filter(b => b.type === 'thinking')

      const text     = textBlocks.map(b => b.text as string).join('\n').trim()
      const thinking = thinkBlocks.map(b => b.thinking as string).join('\n').trim() || undefined
      const toolCalls = toolUseBlocks.length > 0
        ? toolUseBlocks.map(b => ({ name: b.name as string, input: b.input }))
        : undefined

      if (text || toolCalls || thinking) {
        messages.push({ id: `h-a-${messages.length}`, role: 'assistant', text, status: 'done', thinking, toolCalls })
      }
    }
  }
  // 결과 캐시 저장
  _sessionHistoryCache.set(jsonlPath, { messages, mtime: stat.mtimeMs, ts: now })
  return messages
}

// ── File Path Safety ───────────────────────────────────────────
// 민감한 경로 차단 — ~/.ssh, ~/.aws, .env 파일 등
const SENSITIVE_PATH_RE = /[/\\](\.ssh|\.gnupg|\.aws|\.netrc|\.npmrc|\.env[^/\\]*|\.config[/\\]gh)[/\\$]/i

function isSafeFilePath(p: string): boolean {
  if (!p || !path.isAbsolute(p)) return false
  const normalized = path.normalize(p)
  // path.normalize 후에도 '..' 잔존 시 거부
  if (normalized.includes('..')) return false
  if (SENSITIVE_PATH_RE.test(normalized)) return false
  return true
}

// ── Git Operations ─────────────────────────────────────────────
const ALLOWED_GIT_CUSTOM = new Set([
  'stash', 'reset', 'rebase', 'cherry-pick', 'merge', 'tag',
  'fetch', 'remote', 'show', 'describe', 'blame', 'shortlog',
  'rev-parse', 'ls-files', 'clean',
])

function runGitAsync(cwd: string, args: string[]): Promise<string> {
  return new Promise((resolve, reject) => {
    execFile('git', args, { cwd, encoding: 'utf8', timeout: 60000 }, (err, stdout, stderr) => {
      if (err) reject(new Error(stderr || err.message))
      else resolve(stdout)
    })
  })
}

// 디렉터리별 git 동시 실행 세마포어 — 무제한 git 프로세스 생성 방지
const _gitQueues = new Map<string, Array<() => void>>()
const MAX_CONCURRENT_GIT = 3
const _gitActive = new Map<string, number>()

async function runGitSemaphore(cwd: string, args: string[]): Promise<string> {
  const active = _gitActive.get(cwd) ?? 0
  if (active >= MAX_CONCURRENT_GIT) {
    // 대기열에 등록 후 순서 대기
    await new Promise<void>(resolve => {
      if (!_gitQueues.has(cwd)) _gitQueues.set(cwd, [])
      _gitQueues.get(cwd)!.push(resolve)
    })
  }
  _gitActive.set(cwd, (_gitActive.get(cwd) ?? 0) + 1)
  try {
    return await runGitAsync(cwd, args)
  } finally {
    _gitActive.set(cwd, (_gitActive.get(cwd) ?? 1) - 1)
    const next = _gitQueues.get(cwd)?.shift()
    if (next) next()
    else if ((_gitActive.get(cwd) ?? 0) === 0) { _gitQueues.delete(cwd); _gitActive.delete(cwd) }
  }
}

function runGit(cwd: string, args: string[]): string {
  try {
    // execFileSync: 셸 없이 git 직접 실행 — message/branch/files 등 사용자 입력의 인젝션 방지
    return execFileSync('git', args, { cwd, encoding: 'utf8', timeout: 30000 })
  } catch (e: unknown) {
    throw new Error((e as { stderr?: string; message?: string }).stderr || (e as Error).message)
  }
}

function runGitWithPat(cwd: string, action: 'push' | 'pull', pat: string): string {
  // 리모트 URL 확인 — execFileSync: 셸 없이 실행
  let remoteUrl: string
  try {
    remoteUrl = execFileSync('git', ['remote', 'get-url', 'origin'], { cwd, encoding: 'utf8', timeout: 3000 }).trim()
  } catch {
    throw new Error('원격 저장소(origin)가 설정되지 않았습니다.')
  }

  // SSH 리모트는 PAT 불필요 — 그냥 실행
  if (!remoteUrl.startsWith('https://')) {
    return runGit(cwd, [action])
  }

  // 이미 인증 정보가 URL에 포함된 경우 제거 후 재삽입
  // https://oauth2:TOKEN@github.com/user/repo.git
  const authUrl = remoteUrl.replace(/^https:\/\/([^@]*@)?/, `https://oauth2:${pat}@`)

  try {
    let result: string
    if (action === 'push') {
      // execFileSync: authUrl(PAT 포함) + branch를 셸 문자열 없이 배열로 전달 — 인젝션 방지
      const branch = execFileSync('git', ['rev-parse', '--abbrev-ref', 'HEAD'], { cwd, encoding: 'utf8', timeout: 3000 }).trim()
      result = execFileSync('git', ['push', authUrl, `HEAD:refs/heads/${branch}`], { cwd, encoding: 'utf8', timeout: 60000 })
    } else {
      result = execFileSync('git', ['pull', authUrl], { cwd, encoding: 'utf8', timeout: 60000 })
    }
    return result
  } catch (e: unknown) {
    const raw = (e as { stderr?: string; message?: string }).stderr || (e as Error).message || ''
    // PAT가 에러 메시지에 노출되지 않도록 마스킹
    const escaped = pat.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    throw new Error(raw.replace(new RegExp(escaped, 'g'), '***'))
  }
}

// ── GitHub API Helpers ──────────────────────────────────────────

/** git remote origin URL에서 owner/repo 자동 감지 */
function detectGitHubRepo(cwd: string): { owner: string; repo: string } | null {
  try {
    const remoteUrl = execFileSync('git', ['remote', 'get-url', 'origin'], { cwd, encoding: 'utf8', timeout: 3000 }).trim()
    const m = remoteUrl.match(/github\.com[/:]([\w.-]+)\/([\w.-]+?)(?:\.git)?$/)
    if (m) return { owner: m[1], repo: m[2] }
  } catch { /* not a git repo or no origin */ }
  return null
}

/** GitHub REST API 호출 (https 모듈 사용) */
function githubApiRequest(
  path: string,
  token: string,
  options: { method?: string; body?: unknown } = {}
): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const https = require('https') as typeof import('https')
    const bodyStr = options.body ? JSON.stringify(options.body) : undefined
    const req = https.request({
      hostname: 'api.github.com',
      path,
      method: options.method ?? 'GET',
      headers: {
        Authorization:           `Bearer ${token}`,
        Accept:                  'application/vnd.github+json',
        'X-GitHub-Api-Version':  '2022-11-28',
        'User-Agent':            'JiKiME-ADK-webchat',
        ...(bodyStr && {
          'Content-Type':   'application/json',
          'Content-Length': Buffer.byteLength(bodyStr).toString(),
        }),
      },
    }, (ghRes) => {
      let data = ''
      ghRes.on('data', (c: Buffer) => { data += c })
      ghRes.on('end', () => {
        try {
          const parsed: unknown = data ? JSON.parse(data) : {}
          const statusCode = ghRes.statusCode ?? 0
          if (statusCode >= 400) {
            const rawMsg = String((parsed as { message?: string })?.message ?? data)
            // 응답 본문에 토큰이 포함될 경우 마스킹 — 로그/클라이언트 노출 방지
            const msg = rawMsg
              .replace(/ghp_[A-Za-z0-9]{36}/g, 'ghp_***')
              .replace(/Bearer\s+\S+/gi, 'Bearer ***')
              .replace(/token\s+[A-Za-z0-9_-]{20,}/gi, 'token ***')
              .slice(0, 200)
            reject(new Error(`GitHub API ${statusCode}: ${msg}`))
          } else {
            resolve(parsed)
          }
        } catch { resolve(data) }
      })
    })
    req.on('error', reject)
    if (bodyStr) req.write(bodyStr)
    req.end()
  })
}

/** jikime-todo / jikime-done 라벨이 없으면 자동 생성 */
async function ensureIssueLabels(owner: string, repo: string, token: string): Promise<void> {
  const labels = [
    { name: 'jikime-todo', color: '0075ca', description: 'JiKiME: pending task' },
    { name: 'jikime-done', color: '0e8a16', description: 'JiKiME: completed task' },
  ]
  for (const label of labels) {
    try {
      await githubApiRequest(`/repos/${owner}/${repo}/labels`, token, { method: 'POST', body: label })
    } catch (e: unknown) {
      // 422 = 이미 존재 → 무시
      if (!(e as Error).message?.includes('422')) throw e
    }
  }
}

/** ADK로 GitHub 이슈 처리 (비동기, 백그라운드 실행) */
async function processIssueWithADK(
  issueKey: string,
  projectPath: string,
  token: string,
  issueNumber: number,
  issueTitle: string,
  issueBody: string,
  owner: string,
  repo: string,
  model: string,
): Promise<void> {
  const processor = issueProcessors.get(issueKey)
  if (!processor) return

  const emit = (message: string) => {
    processor.events.push(message)
    // 이벤트 배열 무한 누적 방지 — 최근 2500개만 유지
    if (processor.events.length > 5000) processor.events = processor.events.slice(-2500)
    const payload = `data: ${JSON.stringify({ type: 'event', message })}\n\n`
    for (const client of processor.sseClients) {
      try { client.write(payload) } catch { /* client disconnected */ }
    }
  }
  const sendDone = (status: 'done' | 'error') => {
    processor.status = status
    processor.completedAt = Date.now()
    const payload = `data: ${JSON.stringify({ type: 'done', status })}\n\n`
    for (const client of processor.sseClients) {
      try { client.write(payload); client.end() } catch { /* */ }
    }
  }

  try {
    emit(`🚀 이슈 #${issueNumber} 처리 시작: ${issueTitle}`)
    const { query } = await import('@anthropic-ai/claude-agent-sdk')
    const effectiveCwd = fs.existsSync(projectPath) ? projectPath : os.homedir()
    const isRoot = process.getuid?.() === 0

    const prompt = [
      `You are an autonomous software engineer working on a GitHub issue.`,
      `Repository: https://github.com/${owner}/${repo}`,
      ``,
      `## Issue #${issueNumber}: ${issueTitle}`,
      ``,
      issueBody || '(no description)',
      ``,
      `## Instructions`,
      `1. Analyze the issue and understand what needs to be done.`,
      `2. Implement the necessary changes in the repository.`,
      `3. Commit your changes with a clear message referencing the issue.`,
      `4. Push the branch and create a pull request if appropriate.`,
      `Work autonomously. Use the available tools to read files, make edits, and run commands.`,
    ].join('\n')

    const options: Record<string, unknown> = {
      cwd: effectiveCwd,
      permissionMode: isRoot ? 'acceptEdits' : 'bypassPermissions',
      model,
      settingSources: ['user', 'project'],
      ...(!isRoot && { allowDangerouslySkipPermissions: true }),
      ...(CLAUDE_PATH && { pathToClaudeCodeExecutable: CLAUDE_PATH }),
    }

    const queryInstance = query({ prompt, options })
    processor.interrupt = queryInstance.interrupt?.bind(queryInstance) ?? null

    for await (const event of queryInstance) {
      if (processor.status !== 'running') break
      const e = event as Record<string, unknown>
      if (e.type === 'assistant') {
        const content = (e.message as { content?: unknown[] } | undefined)?.content ?? []
        for (const block of content as Record<string, unknown>[]) {
          if (block.type === 'text' && typeof block.text === 'string') {
            emit(block.text.slice(0, 300))
          } else if (block.type === 'tool_use' && typeof block.name === 'string') {
            const inputJson = block.input ? JSON.stringify(block.input) : '{}'
            emit(`🔧 ${block.name}: ${inputJson}`)
          }
        }
      }
    }

    // 완료 시 라벨 교체: jikime-todo → jikime-done
    await githubApiRequest(`/repos/${owner}/${repo}/issues/${issueNumber}/labels`, token, {
      method: 'POST',
      body: { labels: ['jikime-done'] },
    }).catch(() => {})
    await githubApiRequest(
      `/repos/${owner}/${repo}/issues/${issueNumber}/labels/jikime-todo`,
      token,
      { method: 'DELETE' },
    ).catch(() => {})

    emit(`✅ 이슈 #${issueNumber} 처리 완료`)
    sendDone('done')
  } catch (err: unknown) {
    const msg = (err as Error).message ?? String(err)
    emit(`❌ 오류: ${msg}`)
    sendDone('error')
  }
}

// ── GitHub Poller Functions ────────────────────────────────────

function broadcastPollerEvent(poller: ProjectPoller, data: PollerEvent) {
  const payload = `data: ${JSON.stringify(data)}\n\n`
  for (const client of poller.sseClients) {
    try { client.write(payload) } catch { poller.sseClients.delete(client) }
  }
}

async function pollOnce(poller: ProjectPoller): Promise<void> {
  poller.lastCheck = new Date().toISOString()
  try {
    const issues = await githubApiRequest(
      `/repos/${poller.owner}/${poller.repo}/issues?state=open&per_page=50&labels=jikime-todo`,
      poller.token,
    ) as Array<Record<string, unknown>>

    for (const issue of issues) {
      if (poller.status !== 'running') break

      const number = issue.number as number
      const issueKey = `${poller.owner}/${poller.repo}#${number}`

      // 이미 처리 중이면 skip
      const existing = issueProcessors.get(issueKey)
      if (existing?.status === 'running') continue
      if (poller.activeIssues.has(number)) continue

      // maxConcurrent 제한
      if (poller.activeIssues.size >= poller.maxConcurrent) continue

      poller.activeIssues.add(number)
      broadcastPollerEvent(poller, {
        type: 'issue_found',
        issueNumber: number,
        issueTitle: issue.title as string,
        activeCount: poller.activeIssues.size,
      })

      // IssueProcessor 등록 — Map 크기 제한 500 (신규 키만 체크)
      if (!issueProcessors.has(issueKey) && issueProcessors.size >= 500) {
        console.warn('[poller] issueProcessors 한도 초과 — 이슈 처리 건너뜀:', issueKey)
        poller.activeIssues.delete(number)
        continue
      }
      const processor: IssueProcessor = {
        status: 'running',
        events: [],
        interrupt: null,
        sseClients: new Set(),
      }
      issueProcessors.set(issueKey, processor)

      // ADK 처리 (비동기)
      processIssueWithADK(
        issueKey,
        poller.projectPath,
        poller.token,
        number,
        issue.title as string,
        (issue.body as string) ?? '',
        poller.owner,
        poller.repo,
        poller.model,
      ).then(() => {
        poller.activeIssues.delete(number)
        broadcastPollerEvent(poller, {
          type: 'issue_done',
          issueNumber: number,
          status: issueProcessors.get(issueKey)?.status ?? 'done',
          activeCount: poller.activeIssues.size,
        })
      }).catch(() => {
        poller.activeIssues.delete(number)
        broadcastPollerEvent(poller, {
          type: 'issue_done',
          issueNumber: number,
          status: 'error',
          activeCount: poller.activeIssues.size,
        })
      })
    }

    broadcastPollerEvent(poller, {
      type: 'tick',
      lastCheck: poller.lastCheck,
      activeCount: poller.activeIssues.size,
    })
  } catch (err: unknown) {
    broadcastPollerEvent(poller, { type: 'error', message: (err as Error).message })
  }
}

function startProjectPoller(
  projectPath: string,
  token: string,
  owner: string,
  repo: string,
  intervalMs = 15000,
  maxConcurrent = 3,
  model = 'claude-sonnet-4-6',
): ProjectPoller {
  // 기존 폴러 정리
  stopProjectPoller(projectPath)

  const poller: ProjectPoller = {
    projectPath, token, owner, repo,
    intervalMs, maxConcurrent, model,
    timer: null,
    status: 'running',
    lastCheck: null,
    activeIssues: new Set(),
    sseClients: new Set(),
  }
  projectPollers.set(projectPath, poller)

  // 즉시 첫 폴링 후 주기적 실행
  pollOnce(poller).catch(() => {})
  poller.timer = setInterval(() => {
    if (poller.status === 'running') pollOnce(poller).catch(() => {})
  }, intervalMs)

  console.log(`[poller] started — ${owner}/${repo} every ${intervalMs}ms (max ${maxConcurrent} concurrent)`)
  return poller
}

function stopProjectPoller(projectPath: string): void {
  const poller = projectPollers.get(projectPath)
  if (!poller) return
  if (poller.timer) clearInterval(poller.timer)
  poller.status = 'stopped'
  // 처리 중인 이슈 모두 중단
  for (const number of poller.activeIssues) {
    const issueKey = `${poller.owner}/${poller.repo}#${number}`
    const proc = issueProcessors.get(issueKey)
    proc?.interrupt?.().catch(() => {})
  }
  projectPollers.delete(projectPath)
  console.log(`[poller] stopped — ${poller.owner}/${poller.repo}`)
}

// ── CORS 헤더 (원격 브라우저에서 접근 허용) ───────────────────────
const CORS_HEADERS = {
  'Access-Control-Allow-Origin':  '*',
  'Access-Control-Allow-Methods': 'GET, POST, DELETE, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type',
}

// ── HTTP Request Handler for custom routes ─────────────────────
function handleCustomRoutes(req: import('http').IncomingMessage, res: import('http').ServerResponse): boolean {
  const base = `http://${req.headers.host || 'localhost'}`
  const url = new URL(req.url || '/', base)
  const pathname = url.pathname

  // OPTIONS preflight (CORS)
  if (req.method === 'OPTIONS') {
    res.writeHead(204, CORS_HEADERS)
    res.end()
    return true
  }

  // GET /api/ws/health — Docker healthcheck
  if (pathname === '/api/ws/health' && req.method === 'GET') {
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify({ ok: true }))
    return true
  }

  // GET /api/ws/projects
  if (pathname === '/api/ws/projects' && req.method === 'GET') {
    const projects = discoverProjects()
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify(projects))
    return true
  }

  // POST /api/ws/session/new  { projectId }
  // ~/.claude/projects/{projectId}/{uuid}.jsonl 빈 파일 생성 → 새 세션 ID 반환
  if (pathname === '/api/ws/session/new' && req.method === 'POST') {
    const MAX_BODY = 4 * 1024  // 4 KB
    let body = ''; let bodySize = 0
    req.on('data', (chunk: Buffer) => {
      bodySize += chunk.length
      if (bodySize > MAX_BODY) { req.destroy(); return }
      body += chunk
    })
    req.on('end', () => {
      let projectId: string
      try {
        projectId = (JSON.parse(body) as { projectId?: string }).projectId ?? ''
      } catch {
        res.writeHead(400, CORS_HEADERS); res.end('Invalid JSON'); return
      }

      // 입력 검증: 비어있거나 경로 탈출 문자 포함 금지
      if (!projectId || projectId.includes('/') || projectId.includes('..')) {
        res.writeHead(400, CORS_HEADERS); res.end('Invalid projectId'); return
      }

      const claudeDir = path.resolve(path.join(os.homedir(), '.claude', 'projects'))
      const projectDir = path.join(claudeDir, projectId)

      // 경로 탈출 방지
      if (!path.resolve(projectDir).startsWith(claudeDir + path.sep)) {
        res.writeHead(400, CORS_HEADERS); res.end('Invalid projectId'); return
      }

      try {
        if (!fs.existsSync(projectDir)) fs.mkdirSync(projectDir, { recursive: true })
        const sessionId = crypto.randomUUID()
        fs.writeFileSync(path.join(projectDir, `${sessionId}.jsonl`), '', { encoding: 'utf8' })
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ sessionId }))
      } catch (e) {
        console.error('[session/new] error:', e)
        res.writeHead(500, CORS_HEADERS); res.end('Failed to create session')
      }
    })
    return true
  }

  // DELETE /api/ws/session?projectId={projectId}&sessionId={sessionId}
  if (pathname === '/api/ws/session' && req.method === 'DELETE') {
    const projectId = url.searchParams.get('projectId') ?? ''
    const sessionId = url.searchParams.get('sessionId') ?? ''
    if (!projectId || !sessionId) {
      res.writeHead(400, CORS_HEADERS); res.end('Missing params'); return true
    }
    const claudeDir   = path.join(os.homedir(), '.claude', 'projects')
    const sessionFile = path.join(claudeDir, projectId, `${sessionId}.jsonl`)
    // 경로 탈출 방지
    if (!sessionFile.startsWith(claudeDir)) {
      res.writeHead(400, CORS_HEADERS); res.end('Invalid params'); return true
    }
    try {
      if (fs.existsSync(sessionFile)) fs.rmSync(sessionFile)
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify({ ok: true }))
    } catch {
      res.writeHead(500, CORS_HEADERS); res.end('Delete failed')
    }
    return true
  }

  // DELETE /api/ws/project?id={projectId}
  // ~/.claude/projects/{id} 디렉터리(세션 파일)만 삭제. 실제 소스코드는 삭제하지 않음.
  if (pathname === '/api/ws/project' && req.method === 'DELETE') {
    const projectId = url.searchParams.get('id') ?? ''
    if (!projectId) {
      res.writeHead(400, CORS_HEADERS); res.end('Missing id'); return true
    }
    const claudeDir  = path.join(os.homedir(), '.claude', 'projects')
    const projectDir = path.join(claudeDir, projectId)
    // 경로 탈출 방지
    if (!projectDir.startsWith(claudeDir)) {
      res.writeHead(400, CORS_HEADERS); res.end('Invalid id'); return true
    }
    try {
      if (fs.existsSync(projectDir)) {
        fs.rmSync(projectDir, { recursive: true, force: true })
      }
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify({ ok: true }))
    } catch (e) {
      res.writeHead(500, CORS_HEADERS); res.end('Delete failed')
    }
    return true
  }

  // POST /api/ws/project — 새 프로젝트 경로 등록 (경로 디렉터리가 없어도 가능)
  if (pathname === '/api/ws/project' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })  // slow-client DoS 방지
    const MAX_PROJECT_BODY = 10 * 1024  // 10 KB
    let body = ''
    let bodySize = 0
    req.on('data', (chunk: Buffer) => {
      bodySize += chunk.length
      if (bodySize > MAX_PROJECT_BODY) { req.destroy(); return }
      body += chunk
    })
    req.on('end', () => {
      if (bodySize > MAX_PROJECT_BODY) {
        res.writeHead(413, CORS_HEADERS); res.end(JSON.stringify({ error: 'Request body too large' })); return
      }
      try {
        const { path: projectPath } = JSON.parse(body) as { path?: string }
        if (!projectPath || typeof projectPath !== 'string') {
          res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: 'Missing path' })); return
        }
        const normalized = path.resolve(projectPath.replace(/\/+$/, '').trim())
        if (!normalized.startsWith('/')) {
          res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: '절대 경로를 입력하세요' })); return
        }
        // 시스템 루트 디렉터리 등록 방지
        const BLOCKED_PREFIXES = ['/', '/etc', '/root', '/var', '/sys', '/proc', '/dev', '/boot']
        if (BLOCKED_PREFIXES.some(p => normalized === p || normalized.startsWith(p + '/'))) {
          res.writeHead(403, CORS_HEADERS); res.end(JSON.stringify({ error: '시스템 경로는 등록할 수 없습니다' })); return
        }
        // 파일이면 거부
        if (fs.existsSync(normalized) && !fs.statSync(normalized).isDirectory()) {
          res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: '파일 경로는 등록할 수 없습니다' })); return
        }
        const encoded = normalized.replace(/\//g, '-')
        const claudeDir = path.resolve(path.join(os.homedir(), '.claude', 'projects'))
        const projectDir = path.join(claudeDir, encoded)
        // path.resolve() 기반 엄격한 경로 탈출 방지
        if (!path.resolve(projectDir).startsWith(claudeDir + path.sep)) {
          res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: 'Invalid path' })); return
        }
        if (!fs.existsSync(projectDir)) {
          fs.mkdirSync(projectDir, { recursive: true })
        }
        // 실제 프로젝트 폴더도 없으면 생성
        if (!fs.existsSync(normalized)) {
          fs.mkdirSync(normalized, { recursive: true })
        }
        // 원본 경로 저장 — decodeProjectPath 실패 시 fallback
        fs.writeFileSync(path.join(projectDir, '_webchat_path'), normalized, 'utf8')
        const project = {
          id: encoded,
          name: path.basename(normalized) || encoded,
          path: normalized,
          sessions: [],
        }
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ ok: true, project }))
      } catch {
        res.writeHead(500, CORS_HEADERS); res.end(JSON.stringify({ error: '서버 오류' }))
      }
    })
    return true
  }

  // GET /api/ws/session?projectPath=...&sessionId=...
  if (pathname === '/api/ws/session' && req.method === 'GET') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    const sessionId   = url.searchParams.get('sessionId')   ?? ''
    if (!projectPath || !sessionId) {
      res.writeHead(400); res.end('Missing params'); return true
    }
    const jsonlPath = findSessionJsonl(projectPath, sessionId)
    if (!jsonlPath) {
      console.warn(`[session] jsonl not found — projectPath=${projectPath} sessionId=${sessionId}`)
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify([]))
      return true
    }
    try {
      const messages = parseSessionHistory(jsonlPath)
      console.log(`[session] loaded ${messages.length} messages from ${jsonlPath}`)
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify(messages))
    } catch (e) {
      console.error(`[session] parse error — ${jsonlPath}:`, e)
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify([]))
    }
    return true
  }

  // GET /api/ws/session-lookup?sessionId=...
  // sessionId로 ~/.claude/projects/ 디렉터리를 스캔해 projectPath 반환
  if (pathname === '/api/ws/session-lookup' && req.method === 'GET') {
    const sessionId = url.searchParams.get('sessionId') ?? ''
    if (!sessionId) {
      res.writeHead(400, CORS_HEADERS); res.end('Missing sessionId'); return true
    }
    const claudeDir = path.join(os.homedir(), '.claude', 'projects')
    try {
      for (const entry of fs.readdirSync(claudeDir)) {
        const p = path.join(claudeDir, entry, `${sessionId}.jsonl`)
        if (fs.existsSync(p)) {
          const projectPath = decodeProjectPath(entry)
          res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify({ projectPath: projectPath ?? null }))
          return true
        }
      }
    } catch { /* */ }
    res.writeHead(404, CORS_HEADERS); res.end('Session not found')
    return true
  }

  // GET /api/ws/commands?projectPath=...
  // {projectPath}/.claude/commands/jikime/ 의 .md 파일을 파일명 기준 정렬해서 반환
  if (pathname === '/api/ws/commands' && req.method === 'GET') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    if (!projectPath) {
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify([]))
      return true
    }

    const commandsDir = path.join(projectPath, '.claude', 'commands', 'jikime')

    interface SlashCommandMeta {
      name: string        // 파일명에서 .md 제거
      description: string // frontmatter description
      argumentHint: string // frontmatter argument-hint
      context: string     // frontmatter context
    }

    const commands: SlashCommandMeta[] = []

    try {
      if (fs.existsSync(commandsDir)) {
        const files = fs.readdirSync(commandsDir)
          .filter(f => f.endsWith('.md'))
          .sort()  // 파일명 기준 오름차순

        for (const file of files) {
          const filePath = path.join(commandsDir, file)
          const name = file.replace(/\.md$/, '')
          let description = ''
          let argumentHint = ''
          let context = ''

          try {
            const content = fs.readFileSync(filePath, 'utf8')
            // frontmatter 파싱 (--- 사이의 YAML)
            const fmMatch = content.match(/^---\n([\s\S]*?)\n---/)
            if (fmMatch) {
              const fm = fmMatch[1]
              const descMatch    = fm.match(/^description:\s*["']?(.*?)["']?\s*$/m)
              const argMatch     = fm.match(/^argument-hint:\s*["']?(.*?)["']?\s*$/m)
              const ctxMatch     = fm.match(/^context:\s*["']?(.*?)["']?\s*$/m)
              if (descMatch)    description  = descMatch[1].trim()
              if (argMatch)     argumentHint = argMatch[1].trim()
              if (ctxMatch)     context      = ctxMatch[1].trim()
            }
          } catch { /* frontmatter 없으면 빈 값 */ }

          commands.push({ name, description, argumentHint, context })
        }
      }
    } catch { /* commandsDir 접근 불가 시 빈 배열 */ }

    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify(commands))
    return true
  }

  // GET /api/ws/files?path=...
  if (pathname === '/api/ws/files' && req.method === 'GET') {
    const filePath = url.searchParams.get('path') ?? ''
    if (!filePath) {
      res.writeHead(400, CORS_HEADERS); res.end('Missing path'); return true
    }
    if (!isSafeFilePath(filePath)) {
      res.writeHead(403, CORS_HEADERS); res.end('Forbidden path'); return true
    }
    // 경로가 존재하지 않으면 홈 디렉터리로 폴백
    const effectivePath = fs.existsSync(filePath) ? filePath : os.homedir()
    if (effectivePath !== filePath) {
      console.warn(`[files] path not found: ${filePath} — falling back to ${effectivePath}`)
    } else {
      console.log(`[files] tree request: ${effectivePath}`)
    }
    const tree = getFileTree(effectivePath)
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify({ path: effectivePath, tree }))
    return true
  }

  // GET /api/ws/file?path=... (read single file)
  if (pathname === '/api/ws/file' && req.method === 'GET') {
    const filePath = url.searchParams.get('path') ?? ''
    if (!filePath || !fs.existsSync(filePath)) {
      res.writeHead(404, CORS_HEADERS); res.end('Not found'); return true
    }
    if (!isSafeFilePath(filePath)) {
      res.writeHead(403, CORS_HEADERS); res.end('Forbidden path'); return true
    }
    try {
      const content = fs.readFileSync(filePath, 'utf8')
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify({ content }))
    } catch {
      res.writeHead(500, CORS_HEADERS); res.end('Read error')
    }
    return true
  }

  // POST /api/ws/file (write/save file)
  if (pathname === '/api/ws/file' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })  // slow-client DoS 방지
    const MAX_FILE_WRITE = 10 * 1024 * 1024  // 10 MB
    let body = ''
    let bodySize = 0
    let aborted = false
    req.on('data', (chunk: Buffer) => {
      bodySize += chunk.length
      if (bodySize > MAX_FILE_WRITE) {
        aborted = true
        res.writeHead(413, CORS_HEADERS); res.end('File too large (max 10 MB)')
        req.destroy()
        return
      }
      body += chunk
    })
    req.on('end', () => {
      if (aborted) return
      try {
        const { path: filePath, content } = JSON.parse(body)
        if (!filePath) {
          res.writeHead(400, CORS_HEADERS); res.end('Missing path'); return
        }
        if (!isSafeFilePath(filePath)) {
          res.writeHead(403, CORS_HEADERS); res.end('Forbidden path'); return
        }
        fs.writeFileSync(filePath, content, 'utf8')
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ ok: true }))
      } catch {
        res.writeHead(500, CORS_HEADERS); res.end('Write error')
      }
    })
    return true
  }

  // POST /api/ws/git
  // ── GitHub Issues API ─────────────────────────────────────────

  // GET /api/ws/github/repo?projectPath=...
  // git remote origin 에서 owner/repo 자동 감지
  if (pathname === '/api/ws/github/repo' && req.method === 'GET') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    const result = projectPath ? detectGitHubRepo(projectPath) : null
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify(result ?? { error: 'Could not detect GitHub repository' }))
    return true
  }

  // GET /api/ws/github/issues?projectPath=...
  // 반드시 Authorization: Bearer <PAT> 헤더로 토큰 전달 — URL 쿼리 파라미터는 서버 로그에 노출되므로 완전 차단
  if (pathname === '/api/ws/github/issues' && req.method === 'GET') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    if (url.searchParams.has('token')) {
      res.writeHead(400, CORS_HEADERS)
      res.end(JSON.stringify({ error: 'URL query param token is not allowed. Use Authorization: Bearer <token> header.' }))
      return true
    }
    const authHeader = req.headers['authorization'] ?? ''
    const token = authHeader.startsWith('Bearer ') ? authHeader.slice(7) : ''
    if (!token) {
      res.writeHead(401, CORS_HEADERS); res.end(JSON.stringify({ error: 'GitHub PAT required (Authorization: Bearer <token>)' })); return true
    }
    const ghRepo = detectGitHubRepo(projectPath)
    if (!ghRepo) {
      res.writeHead(404, CORS_HEADERS); res.end(JSON.stringify({ error: 'Cannot detect GitHub repo from git remote' })); return true
    }
    ;(async () => {
      try {
        const issues = await githubApiRequest(
          `/repos/${ghRepo.owner}/${ghRepo.repo}/issues?state=all&per_page=20&sort=created&direction=desc`,
          token,
        ) as Array<Record<string, unknown>>
        // PR 제외
        const filtered = issues
          .filter(i => !i.pull_request)
          .map(i => ({
            number:    i.number,
            title:     i.title,
            state:     i.state,
            body:      (i.body as string | null)?.slice(0, 500) ?? '',
            url:       i.html_url,
            createdAt: i.created_at,
            closedAt:  i.closed_at,
            labels:    (i.labels as Array<{ name: string }>).map(l => l.name),
          }))
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ owner: ghRepo.owner, repo: ghRepo.repo, issues: filtered }))
      } catch (e: unknown) {
        res.writeHead(500, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ error: (e as Error).message }))
      }
    })()
    return true
  }

  // POST /api/ws/github/issues
  // body: { projectPath, title, body? } — token은 Authorization: Bearer 헤더로 전달
  // GitHub 이슈 생성 (jikime-todo 라벨 자동 추가)
  if (pathname === '/api/ws/github/issues' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })  // slow-client DoS 방지
    const MAX_ISSUES_BODY = 100 * 1024  // 100 KB
    const postAuthHeader = req.headers['authorization'] ?? ''
    const postHeaderToken = postAuthHeader.startsWith('Bearer ') ? postAuthHeader.slice(7) : ''
    let body = ''
    let bodySize = 0
    req.on('data', (c: Buffer) => {
      bodySize += c.length
      if (bodySize > MAX_ISSUES_BODY) { req.destroy(); return }
      body += c
    })
    req.on('end', () => {
      if (bodySize > MAX_ISSUES_BODY) {
        res.writeHead(413, CORS_HEADERS); res.end(JSON.stringify({ error: 'Request body too large' })); return
      }
      ;(async () => {
        try {
          const { projectPath, token: bodyToken, title, body: issueBody } = JSON.parse(body) as {
            projectPath: string; token?: string; title: string; body?: string
          }
          // Authorization 헤더 우선, 폴백으로 body token 지원 (하위 호환)
          const token = postHeaderToken || bodyToken || ''
          if (!token || !title) {
            res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: 'token and title required' })); return
          }
          const ghRepo = detectGitHubRepo(projectPath)
          if (!ghRepo) {
            res.writeHead(404, CORS_HEADERS); res.end(JSON.stringify({ error: 'Cannot detect GitHub repo' })); return
          }
          await ensureIssueLabels(ghRepo.owner, ghRepo.repo, token)
          const created = await githubApiRequest(
            `/repos/${ghRepo.owner}/${ghRepo.repo}/issues`,
            token,
            { method: 'POST', body: { title, body: issueBody ?? '', labels: ['jikime-todo'] } },
          ) as Record<string, unknown>
          res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify({ number: created.number, url: created.html_url, title: created.title }))
        } catch (e: unknown) {
          res.writeHead(500, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify({ error: (e as Error).message }))
        }
      })()
    })
    return true
  }

  // POST /api/ws/github/process
  // body: { projectPath, token, issueNumber, issueTitle, issueBody, owner, repo, model? }
  // ADK로 이슈 처리 시작 (백그라운드)
  if (pathname === '/api/ws/github/process' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })  // slow-client DoS 방지
    const MAX_PROCESS_BODY = 100 * 1024  // 100 KB
    let body = ''
    let bodySize = 0
    req.on('data', (c: Buffer) => {
      bodySize += c.length
      if (bodySize > MAX_PROCESS_BODY) { req.destroy(); return }
      body += c
    })
    req.on('end', () => {
      if (bodySize > MAX_PROCESS_BODY) {
        res.writeHead(413, CORS_HEADERS); res.end(JSON.stringify({ error: 'Request body too large' })); return
      }
      try {
        const parsed = JSON.parse(body) as {
          projectPath: string; token: string; issueNumber: number
          issueTitle: string; issueBody: string; owner: string; repo: string; model?: string
        }
        const { projectPath, token, issueNumber, model = 'claude-sonnet-4-6' } = parsed
        const owner = parsed.owner
        const repo  = parsed.repo
        // owner/repo 형식 검증 — GitHub API 경로 삽입 전 필수
        const GITHUB_REPO_RE = /^[a-zA-Z0-9_.-]+$/
        if (!GITHUB_REPO_RE.test(owner) || !GITHUB_REPO_RE.test(repo)) {
          res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: 'Invalid owner or repo format' })); return
        }
        // issueTitle/issueBody 길이 제한 + --- 구분자 치환 (프롬프트 인젝션 방지)
        const issueTitle = (parsed.issueTitle || '').slice(0, 200).replace(/^---/gm, '- - -')
        const issueBody  = (parsed.issueBody  || '').slice(0, 8000).replace(/^---/gm, '- - -')
        const issueKey = `${owner}/${repo}#${issueNumber}`
        if (issueProcessors.has(issueKey) && issueProcessors.get(issueKey)?.status === 'running') {
          res.writeHead(409, CORS_HEADERS); res.end(JSON.stringify({ error: 'Already processing', issueKey })); return
        }
        // Map 크기 제한 500 (신규 키만 체크)
        if (!issueProcessors.has(issueKey) && issueProcessors.size >= 500) {
          res.writeHead(429, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify({ error: 'Too many issue processors (max 500)' })); return
        }
        const processor: IssueProcessor = {
          status: 'running',
          events: [],
          interrupt: null,
          sseClients: new Set(),
        }
        issueProcessors.set(issueKey, processor)
        // 비동기 백그라운드 실행
        processIssueWithADK(issueKey, projectPath, token, issueNumber, issueTitle, issueBody, owner, repo, model)
          .catch(e => console.error('[issue-process]', e))
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ issueKey, status: 'started' }))
      } catch (e: unknown) {
        res.writeHead(500, CORS_HEADERS); res.end(JSON.stringify({ error: (e as Error).message }))
      }
    })
    return true
  }

  // GET /api/ws/github/events?issueKey=owner/repo%23number
  // SSE: 이슈 처리 이벤트 실시간 스트리밍
  if (pathname === '/api/ws/github/events' && req.method === 'GET') {
    const issueKey = url.searchParams.get('issueKey') ?? ''
    // issueKey 형식 검증 — owner/repo#number (로그 인젝션 + 일관성 방지)
    const ISSUE_KEY_RE = /^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+#\d+$/
    if (!ISSUE_KEY_RE.test(issueKey)) {
      res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: 'Invalid issueKey format' })); return true
    }
    res.writeHead(200, {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
      ...CORS_HEADERS,
    })
    const processor = issueProcessors.get(issueKey)
    if (!processor) {
      res.write(`data: ${JSON.stringify({ type: 'error', message: 'No processor found for ' + issueKey })}\n\n`)
      res.end()
      return true
    }
    // 기존 이벤트 재전송
    for (const evt of processor.events) {
      res.write(`data: ${JSON.stringify({ type: 'event', message: evt })}\n\n`)
    }
    if (processor.status !== 'running') {
      res.write(`data: ${JSON.stringify({ type: 'done', status: processor.status })}\n\n`)
      res.end()
      return true
    }
    // 실시간 구독
    processor.sseClients.add(res)
    // 25초 하트비트 — 프록시/방화벽 idle timeout 방지
    const hb = setInterval(() => { try { res.write(': heartbeat\n\n') } catch { clearInterval(hb) } }, 25_000)
    req.on('close', () => { clearInterval(hb); processor.sseClients.delete(res) })
    return true
  }

  // DELETE /api/ws/github/process?issueKey=...
  // 이슈 처리 중단
  if (pathname === '/api/ws/github/process' && req.method === 'DELETE') {
    const issueKey = url.searchParams.get('issueKey') ?? ''
    const ISSUE_KEY_RE = /^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+#\d+$/
    if (!ISSUE_KEY_RE.test(issueKey)) {
      res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: 'Invalid issueKey format' })); return true
    }
    const processor = issueProcessors.get(issueKey)
    if (processor) {
      processor.interrupt?.().catch(() => {})
      processor.status = 'error'
      processor.completedAt = Date.now()
    }
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify({ ok: true }))
    return true
  }

  // ── Poller API ────────────────────────────────────────────────

  // POST /api/ws/github/poller — 자동 폴링 시작
  // body: { projectPath, token, intervalMs?, maxConcurrent?, model? }
  if (pathname === '/api/ws/github/poller' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })  // slow-client DoS 방지
    const MAX_POLLER_BODY = 100 * 1024  // 100 KB
    let body = ''
    let bodySize = 0
    req.on('data', (c: Buffer) => {
      bodySize += c.length
      if (bodySize > MAX_POLLER_BODY) { req.destroy(); return }
      body += c
    })
    req.on('end', () => {
      if (bodySize > MAX_POLLER_BODY) {
        res.writeHead(413, CORS_HEADERS); res.end(JSON.stringify({ error: 'Request body too large' })); return
      }
      try {
        const {
          projectPath, token,
          intervalMs = 15000,
          maxConcurrent = 3,
          model = 'claude-sonnet-4-6',
        } = JSON.parse(body) as {
          projectPath: string; token: string
          intervalMs?: number; maxConcurrent?: number; model?: string
        }
        if (!token || !projectPath) {
          res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: 'projectPath and token required' })); return
        }
        // intervalMs: 5s~1h, maxConcurrent: 1~20 범위 클램핑 — 공격적 폴링 / 과부하 방지
        const safeIntervalMs    = Math.min(Math.max(Number(intervalMs)    || 15000, 5_000), 3_600_000)
        const safeMaxConcurrent = Math.min(Math.max(Number(maxConcurrent) || 3,     1),     20)
        // Map 크기 제한 — 신규 폴러 추가 시 최대 50개 초과 차단 (기존 키 갱신은 허용)
        if (!projectPollers.has(projectPath) && projectPollers.size >= 50) {
          res.writeHead(429, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify({ error: 'Too many pollers (max 50)' })); return
        }
        const ghRepo = detectGitHubRepo(projectPath)
        if (!ghRepo) {
          res.writeHead(404, CORS_HEADERS); res.end(JSON.stringify({ error: 'Cannot detect GitHub repo' })); return
        }
        // 라벨 준비 (비동기, 오류 무시)
        ensureIssueLabels(ghRepo.owner, ghRepo.repo, token).catch(() => {})

        const poller = startProjectPoller(
          projectPath, token, ghRepo.owner, ghRepo.repo,
          safeIntervalMs, safeMaxConcurrent, model,
        )
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({
          status: 'started',
          owner: poller.owner,
          repo: poller.repo,
          intervalMs: poller.intervalMs,
          maxConcurrent: poller.maxConcurrent,
        }))
      } catch (e: unknown) {
        res.writeHead(500, CORS_HEADERS); res.end(JSON.stringify({ error: (e as Error).message }))
      }
    })
    return true
  }

  // DELETE /api/ws/github/poller?projectPath=... — 폴링 중지
  if (pathname === '/api/ws/github/poller' && req.method === 'DELETE') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    stopProjectPoller(projectPath)
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify({ status: 'stopped' }))
    return true
  }

  // GET /api/ws/github/poller?projectPath=... — 폴러 상태 조회
  if (pathname === '/api/ws/github/poller' && req.method === 'GET') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    const poller = projectPollers.get(projectPath)
    if (!poller) {
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify({ status: 'stopped' }))
      return true
    }
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify({
      status: poller.status,
      owner: poller.owner,
      repo: poller.repo,
      intervalMs: poller.intervalMs,
      maxConcurrent: poller.maxConcurrent,
      lastCheck: poller.lastCheck,
      activeCount: poller.activeIssues.size,
      activeIssues: Array.from(poller.activeIssues),
    }))
    return true
  }

  // GET /api/ws/github/poller-events?projectPath=... — SSE 폴러 이벤트 스트리밍
  if (pathname === '/api/ws/github/poller-events' && req.method === 'GET') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    res.writeHead(200, {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
      ...CORS_HEADERS,
    })
    const poller = projectPollers.get(projectPath)
    if (!poller) {
      res.write(`data: ${JSON.stringify({ type: 'tick', status: 'stopped' })}\n\n`)
      res.end()
      return true
    }
    // 현재 상태 즉시 전송
    res.write(`data: ${JSON.stringify({
      type: 'tick',
      lastCheck: poller.lastCheck,
      activeCount: poller.activeIssues.size,
      activeIssues: Array.from(poller.activeIssues),
    })}\n\n`)
    poller.sseClients.add(res)
    // 25초 하트비트 — 프록시/방화벽 idle timeout 방지
    const hbp = setInterval(() => { try { res.write(': heartbeat\n\n') } catch { clearInterval(hbp) } }, 25_000)
    req.on('close', () => { clearInterval(hbp); poller.sseClients.delete(res) })
    return true
  }

  if (pathname === '/api/ws/git' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })  // slow-client DoS 방지
    let body = ''
    let bodySize = 0
    const MAX_GIT_BODY = 1 * 1024 * 1024  // 1 MB — 대형 commit message / files[] OOM 방지
    req.on('data', (chunk: Buffer) => {
      bodySize += chunk.length
      if (bodySize > MAX_GIT_BODY) { req.destroy(); return }
      body += chunk
    })
    req.on('end', () => {
      if (bodySize > MAX_GIT_BODY) {
        res.writeHead(413, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ error: 'Request body too large' })); return
      }
      ;(async () => {
        try {
          const { action, cwd, args, file, files, message, branch, pat } = JSON.parse(body)

          // cwd 기본 검증: 절대경로 + 존재하는 디렉터리 — 경로 탈출 방지
          const resolvedCwd = path.resolve(String(cwd || ''))
          if (!cwd || !path.isAbsolute(resolvedCwd) || !fs.existsSync(resolvedCwd) || !fs.statSync(resolvedCwd).isDirectory()) {
            res.writeHead(400, { 'Content-Type': 'application/json', ...CORS_HEADERS })
            res.end(JSON.stringify({ error: 'Invalid cwd' })); return
          }

          // git 저장소 여부 먼저 확인 — execFileSync: 셸 없이 실행
          try {
            execFileSync('git', ['rev-parse', '--git-dir'], { cwd: resolvedCwd, stdio: 'pipe', timeout: 3000 })
          } catch {
            res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
            res.end(JSON.stringify({ notGit: true, output: '' }))
            return
          }

          let result: unknown
          if (action === 'status') {
            result = { output: runGit(resolvedCwd, ['status', '--short']) }
          } else if (action === 'log') {
            result = { output: runGit(resolvedCwd, ['log', '--oneline', '-20']) }
          } else if (action === 'diff') {
            result = { output: runGit(resolvedCwd, ['diff', '--stat']) }
          } else if (action === 'file_diff') {
            const target = file as string | undefined
            const diffArgs = target ? ['diff', 'HEAD', '--', target] : ['diff', 'HEAD']
            result = { output: runGit(resolvedCwd, diffArgs) }
          } else if (action === 'branch') {
            result = { output: runGit(resolvedCwd, ['branch', '-a']) }
          } else if (action === 'add') {
            const targets: string[] = Array.isArray(files) && files.length > 0 ? files : ['.']
            result = { output: runGit(resolvedCwd, ['add', '--', ...targets]) }
          } else if (action === 'commit') {
            if (!message) { res.writeHead(400, CORS_HEADERS); res.end('Missing message'); return }
            result = { output: runGit(resolvedCwd, ['commit', '-m', message as string]) }
          } else if (action === 'checkout') {
            if (!branch) { res.writeHead(400, CORS_HEADERS); res.end('Missing branch'); return }
            result = { output: runGit(resolvedCwd, ['checkout', branch as string]) }
          } else if (action === 'push' || action === 'pull') {
            if (pat) {
              result = { output: runGitWithPat(resolvedCwd, action, pat as string) }
            } else {
              // 세마포어 적용 비동기 실행 — 이벤트 루프 블로킹 + 동시 프로세스 폭증 방지
              result = { output: await runGitSemaphore(resolvedCwd, [action]) }
            }
          } else if (action === 'custom' && Array.isArray(args)) {
            const subCmd = args[0] as string | undefined
            if (!subCmd || !ALLOWED_GIT_CUSTOM.has(subCmd)) {
              res.writeHead(400, CORS_HEADERS); res.end(JSON.stringify({ error: `Disallowed git subcommand: ${subCmd}` })); return
            }
            result = { output: runGit(resolvedCwd, args) }
          } else {
            res.writeHead(400, CORS_HEADERS); res.end('Unknown action'); return
          }
          res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify(result))
        } catch (e: unknown) {
          res.writeHead(500, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify({ error: (e as Error).message }))
        }
      })()
    })
    return true
  }

  // ── Harness Engineering (WORKFLOW.md 기반 자율 이슈 처리) ────────
  if (handleHarnessRoutes(req, res, pathname, CLAUDE_PATH)) return true

  return false
}

// ── Main ───────────────────────────────────────────────────────
app.prepare().then(() => {
  const httpServer = createServer((req, res) => {
    if (!handleCustomRoutes(req, res)) {
      handle(req, res)
    }
  })

  // WebSocket for Terminal
  const terminalWss = new WebSocketServer({ noServer: true, maxPayload: 5 * 1024 * 1024 })
  // WebSocket for Chat
  const chatWss = new WebSocketServer({ noServer: true, maxPayload: 5 * 1024 * 1024 })

  // Next.js HMR(/_next/webpack-hmr 등) upgrade 요청도 처리할 수 있도록
  const nextUpgradeHandler = app.getUpgradeHandler()

  httpServer.on('upgrade', (req, socket, head) => {
    const pathname = new URL(req.url || '/', 'http://localhost').pathname
    if (pathname === '/ws/terminal') {
      terminalWss.handleUpgrade(req, socket, head, (ws) => terminalWss.emit('connection', ws, req))
    } else if (pathname === '/ws/chat') {
      chatWss.handleUpgrade(req, socket, head, (ws) => chatWss.emit('connection', ws, req))
    } else {
      // Next.js HMR WebSocket 등 나머지는 Next.js에게 위임
      nextUpgradeHandler(req, socket, head)
    }
  })

  // Terminal WebSocket Handler
  terminalWss.on('connection', (ws: WebSocket) => {
    let sessionId = ''

    ws.on('message', (raw) => {
      try {
        const msg = JSON.parse(raw.toString())

        if (msg.type === 'init') {
          sessionId = msg.sessionId || `term-${Date.now()}`
          const cwd = msg.cwd || os.homedir()
          const cols = msg.cols || 120
          const rows = msg.rows || 40

          if (!pty) {
            ws.send(JSON.stringify({
              type: 'output',
              data: '\r\n\x1b[31m[터미널 비활성화]\x1b[0m node-pty 컴파일이 필요합니다.\r\n' +
                    '  sudo dnf install -y python3 make gcc gcc-c++\r\n' +
                    '  pnpm rebuild node-pty\r\n' +
                    '그 후 서버를 재시작하세요.\r\n',
            }))
            return
          }

          const session = getOrCreatePtySession(sessionId, cwd, cols, rows)
          session.clients.add(ws)
          ws.send(JSON.stringify({ type: 'ready', sessionId }))

        } else if (msg.type === 'input' && sessionId) {
          const session = ptySessions.get(sessionId)
          if (session) {
            // 메시지당 4096 bytes 제한 + 10ms rate limit (PTY 플러딩 방지)
            const inputData = String(msg.data ?? '').slice(0, 4096)
            const now = Date.now()
            const elapsed = now - session._lastInputMs
            // 허용/거부 관계없이 매 시도마다 타임스탬프 갱신 — burst 패턴 차단
            session._lastInputMs = now
            if (elapsed >= 10) {
              session.pty.write(inputData)
            }
          }

        } else if (msg.type === 'resize' && sessionId) {
          const session = ptySessions.get(sessionId)
          if (session && msg.cols && msg.rows) {
            session.pty.resize(msg.cols, msg.rows)
            // linked session 은 PTY resize 만으로 충분 — tmux 가 SIGWINCH 를 감지해 갱신함
          }

        } else if (msg.type === 'pause' && sessionId) {
          // ✅ 흐름 제어: 클라이언트 버퍼 포화 시 PTY 출력 일시 중지
          const session = ptySessions.get(sessionId)
          if (session) session.pty.pause()

        } else if (msg.type === 'resume' && sessionId) {
          // ✅ 흐름 제어: 클라이언트 렌더 완료 후 PTY 출력 재개
          const session = ptySessions.get(sessionId)
          if (session) session.pty.resume()
        }
      } catch (err) {
        console.error('Terminal WS error:', err)
      }
    })

    ws.on('close', () => {
      if (sessionId) {
        const session = ptySessions.get(sessionId)
        if (session) {
          session.clients.delete(ws)
          // 마지막 클라이언트 종료 시 PTY 즉시 정리 — fd 누수 방지 (1시간 대기 제거)
          if (session.clients.size === 0) {
            try { session.pty.kill() } catch { /* */ }
            ptySessions.delete(sessionId)
          }
        }
      }
    })
  })

  // Chat WebSocket Handler
  chatWss.on('connection', (ws: WebSocket) => {
    const wsKey = `chat-${Date.now()}-${Math.random()}`
    chatSessions.set(wsKey, { ws, queryInstance: null, claudeSessionId: null })

    ws.on('message', (raw) => {
      // async IIFE — 핸들러 자체를 async로 두면 unhandledRejection 위험
      ;(async () => {
        try {
          const msg = JSON.parse(raw.toString())

          if (msg.type === 'chat') {
            const { sessionId, projectPath, prompt, model, permissionMode, extendedThinking } = msg
            await handleClaudeMessage(
              ws, wsKey,
              sessionId || null,
              projectPath || os.homedir(),
              prompt,
              model,
              permissionMode,
              !!extendedThinking,
            )

          } else if (msg.type === 'abort') {
            const session = chatSessions.get(wsKey)
            if (session?.queryInstance?.interrupt) {
              try { await session.queryInstance.interrupt() } catch { /* */ }
            }

          } else if (msg.type === 'permission_response') {
            // 브라우저에서 도구 실행 허용/거부 응답
            const resolver = pendingPermissions.get(msg.requestId as string)
            if (resolver) resolver({ allow: !!msg.allow, alwaysAllow: !!msg.alwaysAllow })
          }
        } catch (err) {
          console.error('Chat WS error:', err)
          try { ws.send(JSON.stringify({ type: 'error', message: String(err) })) } catch { /* ws already closed */ }
        }
      })()
    })

    ws.on('close', () => {
      const session = chatSessions.get(wsKey)
      if (session?.queryInstance?.interrupt) {
        // interrupt 완료 후 삭제 — 진행 중 조회 가능하도록 순서 보장
        session.queryInstance.interrupt()
          .catch(() => { /* ignore */ })
          .finally(() => { chatSessions.delete(wsKey) })
      } else {
        chatSessions.delete(wsKey)
      }
    })
  })

  // Cleanup stale PTY sessions
  setInterval(() => {
    const now = Date.now()
    for (const [id, session] of ptySessions) {
      if (now - session.lastActivity > 3600000) { // 1 hour
        try { session.pty.kill() } catch { /* */ }
        ptySessions.delete(id)
      }
    }
  }, 5 * 60 * 1000)

  httpServer.listen(port, () => {
    console.log(`> Ready on http://${hostname}:${port}`)
    console.log(`> Terminal WS: ws://${hostname}:${port}/ws/terminal`)
    console.log(`> Chat WS: ws://${hostname}:${port}/ws/chat`)
  })

  // SIGTERM: PTY 세션 정리 후 종료 — fd/메모리 누수 방지
  process.once('SIGTERM', () => {
    console.log('[server] SIGTERM 수신 — PTY 세션 정리 중')
    for (const [, session] of ptySessions) {
      try { session.pty.kill() } catch { /* */ }
    }
    ptySessions.clear()
    httpServer.close(() => process.exit(0))
  })
})
