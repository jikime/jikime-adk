import { createServer } from 'http'
import next from 'next'
import { WebSocketServer, WebSocket } from 'ws'
import * as fs from 'fs'
import * as path from 'path'
import { execSync } from 'child_process'
import * as os from 'os'

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

  // 2. which -a 로 PATH 전체 탐색 (여러 후보 확인)
  try {
    const all = execSync('which -a claude 2>/dev/null || which claude', { encoding: 'utf8', stdio: 'pipe' })
    for (const line of all.split('\n').map(l => l.trim()).filter(Boolean)) {
      const found = check(line)
      if (found) return found
    }
  } catch { /* */ }

  // 3. 흔한 네이티브 설치 경로 직접 확인
  const guesses = [
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
    const ver = execSync(`"${CLAUDE_PATH}" --version 2>&1`, { encoding: 'utf8', timeout: 5000 }).trim()
    console.log(`[server] claude 버전: ${ver}`)
  } catch (e: unknown) {
    const err = e as { stderr?: string; stdout?: string; message?: string }
    console.warn(`[server] claude --version 실패: ${err.stderr || err.stdout || err.message}`)
  }
}

// ── PTY Session Store ──────────────────────────────────────────
interface PtySession {
  pty: import('node-pty').IPty
  clients: Set<WebSocket>
  created: number
  lastActivity: number
}
const ptySessions = new Map<string, PtySession>()

function getOrCreatePtySession(sessionId: string, cwd: string, cols: number, rows: number): PtySession {
  if (!pty) throw new Error('node-pty unavailable')
  if (ptySessions.has(sessionId)) {
    return ptySessions.get(sessionId)!
  }
  const shell = process.env.SHELL || '/bin/zsh'
  const ptyProcess = pty.spawn(shell, [], {
    name: 'xterm-256color',
    cols,
    rows,
    cwd: fs.existsSync(cwd) ? cwd : os.homedir(),
    env: { ...process.env, TERM: 'xterm-256color', COLORTERM: 'truecolor' },
  })
  const session: PtySession = {
    pty: ptyProcess,
    clients: new Set(),
    created: Date.now(),
    lastActivity: Date.now(),
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
  // Auto-start claude
  setTimeout(() => {
    if (ptySessions.has(sessionId)) {
      ptyProcess.write('claude --dangerously-skip-permissions\r')
    }
  }, 800)
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
  model = 'claude-sonnet-4-5',
  permissionMode = 'bypassPermissions',
) {
  console.log(`[chat] start — model=${model} mode=${permissionMode} cwd=${projectPath}`)
  const { query } = await import('@anthropic-ai/claude-agent-sdk')

  let capturedSessionId = claudeSessionId

  const options: Record<string, unknown> = {
    cwd: projectPath,
    permissionMode,
    model,
    ...(permissionMode === 'bypassPermissions' && { allowDangerouslySkipPermissions: true }),
    ...(CLAUDE_PATH && { pathToClaudeCodeExecutable: CLAUDE_PATH }),
  }
  if (claudeSessionId) options.resume = claudeSessionId

  // canUseTool — 권한 확인 모드일 때 브라우저에 승인 요청
  if (permissionMode !== 'bypassPermissions') {
    options.canUseTool = async (toolName: string, input: unknown) => {
      const requestId = `perm-${Date.now()}-${Math.random().toString(36).slice(2)}`
      ws.send(JSON.stringify({ type: 'permission_request', requestId, toolName, input }))

      return new Promise<{ behavior: string; updatedInput?: unknown; message?: string }>((resolve) => {
        // 30초 타임아웃 → 허용으로 처리
        const timer = setTimeout(() => {
          pendingPermissions.delete(requestId)
          resolve({ behavior: 'allow', updatedInput: input })
        }, 30000)

        pendingPermissions.set(requestId, ({ allow }) => {
          clearTimeout(timer)
          pendingPermissions.delete(requestId)
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
      // 실제 오류 확인을 위해 직접 실행
      let detail = ''
      try {
        execSync(`"${CLAUDE_PATH ?? 'claude'}" --dangerously-skip-permissions --output-format stream-json -p "ping" 2>&1`, {
          encoding: 'utf8', timeout: 10000,
        })
      } catch (e: unknown) {
        const ce = e as { stderr?: string; stdout?: string; message?: string }
        detail = (ce.stderr || ce.stdout || ce.message || '').trim()
      }
      const hint = [
        `claude 종료 코드 1 (경로: ${CLAUDE_PATH ?? 'claude'})`,
        detail ? `오류: ${detail}` : '원격 서버에서 직접 확인: claude --dangerously-skip-permissions --output-format stream-json -p "hello"',
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

// JSONL 세션 파일에서 실제 cwd 추출 (경로에 '-'가 포함된 경우 디렉터리명 디코딩이 부정확하므로)
function extractCwdFromJsonl(jsonlPath: string): string | null {
  try {
    const lines = fs.readFileSync(jsonlPath, 'utf8').split('\n')
    for (const line of lines.slice(0, 30)) {
      if (!line.trim()) continue
      try {
        const event = JSON.parse(line)
        // Claude Code가 저장하는 다양한 cwd 필드 형식 탐색
        if (typeof event.cwd === 'string' && event.cwd.startsWith('/')) return event.cwd
        if (typeof event.workdir === 'string' && event.workdir.startsWith('/')) return event.workdir
        if (event.type === 'system' && typeof event.path === 'string') return event.path
        // summary 이벤트에 포함된 경우
        if (event.summary?.cwd) return event.summary.cwd
      } catch { /* */ }
    }
  } catch { /* */ }
  return null
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

      // 세션 파일 목록 수집
      const sessions: string[] = []
      try {
        const files = fs.readdirSync(fullPath).filter(f => f.endsWith('.jsonl'))
        for (const file of files) sessions.push(file.replace('.jsonl', ''))
      } catch { /* */ }

      // 실제 경로: JSONL에서 cwd 읽기 → 없으면 디렉터리명 디코딩 폴백
      let actualPath: string | null = null
      for (const sessionId of sessions.slice(0, 3)) {
        actualPath = extractCwdFromJsonl(path.join(fullPath, `${sessionId}.jsonl`))
        if (actualPath) break
      }
      // 폴백: 디렉터리명에서 추정 (entry 앞의 '-'는 leading '/')
      if (!actualPath) {
        actualPath = entry.startsWith('-') ? entry.replace(/^-/, '/').replace(/-/g, '/') : entry.replace(/-/g, '/')
      }

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

function parseSessionHistory(jsonlPath: string): HistoryMessage[] {
  const lines = fs.readFileSync(jsonlPath, 'utf8').split('\n').filter(Boolean)
  const messages: HistoryMessage[] = []

  for (const line of lines) {
    let event: Record<string, unknown>
    try { event = JSON.parse(line) } catch { continue }
    if (event.type !== 'user' && event.type !== 'assistant') continue

    const msg = event.message as { role?: string; content?: Record<string, unknown>[] } | undefined
    if (!msg?.content) continue
    const content = msg.content

    if (msg.role === 'user') {
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
        if (text) {
          messages.push({ id: `h-u-${messages.length}`, role: 'user', text, status: 'done' })
        }
      }
    } else if (msg.role === 'assistant') {
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
  return messages
}

// ── Git Operations ─────────────────────────────────────────────
function runGit(cwd: string, args: string[]): string {
  try {
    return execSync(`git ${args.join(' ')}`, { cwd, encoding: 'utf8', timeout: 10000 })
  } catch (e: unknown) {
    throw new Error((e as { stderr?: string; message?: string }).stderr || (e as Error).message)
  }
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

  // GET /api/ws/projects
  if (pathname === '/api/ws/projects' && req.method === 'GET') {
    const projects = discoverProjects()
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify(projects))
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

  // GET /api/ws/session?projectPath=...&sessionId=...
  if (pathname === '/api/ws/session' && req.method === 'GET') {
    const projectPath = url.searchParams.get('projectPath') ?? ''
    const sessionId   = url.searchParams.get('sessionId')   ?? ''
    if (!projectPath || !sessionId) {
      res.writeHead(400); res.end('Missing params'); return true
    }
    const jsonlPath = findSessionJsonl(projectPath, sessionId)
    if (!jsonlPath) {
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify([]))
      return true
    }
    try {
      const messages = parseSessionHistory(jsonlPath)
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
      res.end(JSON.stringify(messages))
    } catch (e) {
      res.writeHead(500, CORS_HEADERS); res.end(String(e))
    }
    return true
  }

  // GET /api/ws/files?path=...
  if (pathname === '/api/ws/files' && req.method === 'GET') {
    const filePath = url.searchParams.get('path') ?? ''
    if (!filePath) {
      res.writeHead(400, CORS_HEADERS); res.end('Missing path'); return true
    }
    const tree = getFileTree(filePath)
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
    res.end(JSON.stringify(tree))
    return true
  }

  // GET /api/ws/file?path=... (read single file)
  if (pathname === '/api/ws/file' && req.method === 'GET') {
    const filePath = url.searchParams.get('path') ?? ''
    if (!filePath || !fs.existsSync(filePath)) {
      res.writeHead(404, CORS_HEADERS); res.end('Not found'); return true
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
    let body = ''
    req.on('data', (chunk) => { body += chunk })
    req.on('end', () => {
      try {
        const { path: filePath, content } = JSON.parse(body)
        if (!filePath) {
          res.writeHead(400, CORS_HEADERS); res.end('Missing path'); return
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
  if (pathname === '/api/ws/git' && req.method === 'POST') {
    let body = ''
    req.on('data', (chunk) => { body += chunk })
    req.on('end', () => {
      try {
        const { action, cwd, args, file, files, message, branch } = JSON.parse(body)

        // git 저장소 여부 먼저 확인
        try {
          execSync('git rev-parse --git-dir', { cwd, stdio: 'pipe', timeout: 3000 })
        } catch {
          res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
          res.end(JSON.stringify({ notGit: true, output: '' }))
          return
        }

        let result: unknown
        if (action === 'status') {
          result = { output: runGit(cwd, ['status', '--short']) }
        } else if (action === 'log') {
          result = { output: runGit(cwd, ['log', '--oneline', '-20']) }
        } else if (action === 'diff') {
          result = { output: runGit(cwd, ['diff', '--stat']) }
        } else if (action === 'file_diff') {
          const target = file as string | undefined
          const diffArgs = target ? ['diff', 'HEAD', '--', target] : ['diff', 'HEAD']
          result = { output: runGit(cwd, diffArgs) }
        } else if (action === 'branch') {
          result = { output: runGit(cwd, ['branch', '-a']) }
        } else if (action === 'add') {
          const targets: string[] = Array.isArray(files) && files.length > 0 ? files : ['.']
          result = { output: runGit(cwd, ['add', '--', ...targets]) }
        } else if (action === 'commit') {
          if (!message) { res.writeHead(400, CORS_HEADERS); res.end('Missing message'); return }
          result = { output: runGit(cwd, ['commit', '-m', message as string]) }
        } else if (action === 'checkout') {
          if (!branch) { res.writeHead(400, CORS_HEADERS); res.end('Missing branch'); return }
          result = { output: runGit(cwd, ['checkout', branch as string]) }
        } else if (action === 'custom' && Array.isArray(args)) {
          result = { output: runGit(cwd, args) }
        } else {
          res.writeHead(400, CORS_HEADERS); res.end('Unknown action'); return
        }
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify(result))
      } catch (e: unknown) {
        res.writeHead(500, { 'Content-Type': 'application/json', ...CORS_HEADERS })
        res.end(JSON.stringify({ error: (e as Error).message }))
      }
    })
    return true
  }

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
  const terminalWss = new WebSocketServer({ noServer: true })
  // WebSocket for Chat
  const chatWss = new WebSocketServer({ noServer: true })

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
          if (session) session.pty.write(msg.data)

        } else if (msg.type === 'resize' && sessionId) {
          const session = ptySessions.get(sessionId)
          if (session && msg.cols && msg.rows) {
            session.pty.resize(msg.cols, msg.rows)
          }
        }
      } catch (err) {
        console.error('Terminal WS error:', err)
      }
    })

    ws.on('close', () => {
      if (sessionId) {
        const session = ptySessions.get(sessionId)
        if (session) session.clients.delete(ws)
      }
    })
  })

  // Chat WebSocket Handler
  chatWss.on('connection', (ws: WebSocket) => {
    const wsKey = `chat-${Date.now()}-${Math.random()}`
    chatSessions.set(wsKey, { ws, queryInstance: null, claudeSessionId: null })

    ws.on('message', async (raw) => {
      try {
        const msg = JSON.parse(raw.toString())

        if (msg.type === 'chat') {
          const { sessionId, projectPath, prompt, model, permissionMode } = msg
          await handleClaudeMessage(
            ws, wsKey,
            sessionId || null,
            projectPath || os.homedir(),
            prompt,
            model,
            permissionMode,
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
        ws.send(JSON.stringify({ type: 'error', message: String(err) }))
      }
    })

    ws.on('close', () => {
      const session = chatSessions.get(wsKey)
      if (session?.queryInstance?.interrupt) {
        session.queryInstance.interrupt().catch(() => {})
      }
      chatSessions.delete(wsKey)
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
})
