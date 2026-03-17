import { NextRequest } from 'next/server'
import { spawn, execSync, type ChildProcessWithoutNullStreams } from 'child_process'

export const dynamic = 'force-dynamic'

/* ── Session store ── */

type ShellSession = {
  proc: ChildProcessWithoutNullStreams
  buffer: string[]
  created: number
  lastActivity: number
  cwd: string
  alive: boolean
  listeners: Set<(event: TerminalEvent) => void>
  // 입력 후 idle 감지용 타이머
  idleTimer: ReturnType<typeof setTimeout> | null
  waitingResponse: boolean
}

type TerminalEvent =
  | { type: 'output'; text: string }
  | { type: 'status'; alive: boolean }
  | { type: 'snapshot_request' }  // 클라이언트에 xterm 스냅샷 요청

const sessions = new Map<string, ShellSession>()

/* ── Python PTY bridge ── */
const PY_PTY_BRIDGE = `
import os, pty, select, signal, sys, fcntl, termios, struct, errno

RESIZE_PREFIX = b"__RESIZE__:"

shell = os.environ.get("SHELL", "/bin/zsh")
if not os.path.exists(shell):
    for candidate in ["/bin/zsh", "/bin/bash", "/bin/sh"]:
        if os.path.exists(candidate):
            shell = candidate
            break

init_cols = int(sys.argv[1]) if len(sys.argv) > 1 else 220
init_rows = int(sys.argv[2]) if len(sys.argv) > 2 else 50

pid, fd = pty.fork()
if pid == 0:
    os.execvp(shell, [shell, "-l"])
    sys.exit(1)

try:
    winsize = struct.pack("HHHH", init_rows, init_cols, 0, 0)
    fcntl.ioctl(fd, termios.TIOCSWINSZ, winsize)
    os.kill(pid, signal.SIGWINCH)
except Exception:
    pass

def set_size(cols, rows):
    try:
        winsize = struct.pack("HHHH", rows, cols, 0, 0)
        fcntl.ioctl(fd, termios.TIOCSWINSZ, winsize)
        os.kill(pid, signal.SIGWINCH)
    except Exception:
        pass

def _terminate(_signum, _frame):
    try:
        os.kill(pid, signal.SIGTERM)
    except Exception:
        pass
    sys.exit(0)

signal.signal(signal.SIGTERM, _terminate)

stdin_fd = sys.stdin.fileno()
stdout_fd = sys.stdout.fileno()
stdin_buf = b""

while True:
    try:
        r, _, _ = select.select([fd, stdin_fd], [], [], 30.0)
    except (OSError, ValueError):
        break
    except InterruptedError:
        continue

    if not r:
        continue

    if fd in r:
        try:
            data = os.read(fd, 16384)
        except OSError as e:
            if e.errno == errno.EIO:
                break
            raise
        if not data:
            break
        os.write(stdout_fd, data)

    if stdin_fd in r:
        try:
            data = os.read(stdin_fd, 16384)
        except OSError:
            break
        if not data:
            break
        stdin_buf += data
        while RESIZE_PREFIX in stdin_buf:
            idx = stdin_buf.index(RESIZE_PREFIX)
            if idx > 0:
                os.write(fd, stdin_buf[:idx])
            nl = stdin_buf.find(b"\\n", idx)
            if nl == -1:
                stdin_buf = stdin_buf[idx:]
                break
            cmd = stdin_buf[idx + len(RESIZE_PREFIX):nl]
            stdin_buf = stdin_buf[nl + 1:]
            try:
                parts = cmd.split(b":")
                c, rr = int(parts[0]), int(parts[1])
                if 2 <= c <= 500 and 2 <= rr <= 200:
                    set_size(c, rr)
            except Exception:
                pass
        else:
            if stdin_buf:
                os.write(fd, stdin_buf)
                stdin_buf = b""

try:
    os.waitpid(pid, 0)
except Exception:
    pass
`.trim()

/* ── Helpers ── */

function findPython3(): string {
  const candidates = ['python3', '/usr/bin/python3', '/opt/homebrew/bin/python3', '/usr/local/bin/python3']
  for (const p of candidates) {
    try {
      execSync(`${p} --version`, { stdio: 'pipe', timeout: 3000 })
      return p
    } catch { /* continue */ }
  }
  return 'python3'
}

let cachedPython: string | null = null
function getPython(): string {
  if (!cachedPython) cachedPython = findPython3()
  return cachedPython
}

/* ── Cleanup stale sessions every 2 minutes ── */

const SESSION_IDLE_TIMEOUT = 60 * 60 * 1000
const SESSION_MAX_AGE = 4 * 60 * 60 * 1000

if (typeof globalThis !== 'undefined') {
  const cleanup = () => {
    const now = Date.now()
    for (const [id, s] of sessions) {
      if (!s.alive || now - s.lastActivity > SESSION_IDLE_TIMEOUT || now - s.created > SESSION_MAX_AGE) {
        if (s.idleTimer) clearTimeout(s.idleTimer)
        try { s.proc.kill('SIGTERM') } catch { /* */ }
        sessions.delete(id)
      }
    }
  }
  const g = globalThis as unknown as Record<string, unknown>
  if (!g.__webchatTerminalCleanup) {
    g.__webchatTerminalCleanup = setInterval(cleanup, 2 * 60 * 1000)
  }
}

function createSession(cols = 220, rows = 50, cwd = '/tmp'): string {
  const id = crypto.randomUUID().slice(0, 8)
  const python = getPython()
  const env = { ...process.env, CLAUDECODE: '' }

  const proc = spawn(python, ['-u', '-c', PY_PTY_BRIDGE, String(cols), String(rows)], {
    cwd,
    env: {
      ...env,
      TERM: 'xterm-256color',
      COLORTERM: 'truecolor',
      FORCE_COLOR: '3',
      LANG: process.env.LANG || 'en_US.UTF-8',
      HOME: process.env.HOME || '/tmp',
      CLICOLOR: '1',
      CLICOLOR_FORCE: '1',
    },
    stdio: ['pipe', 'pipe', 'pipe'],
  })

  const now = Date.now()
  const session: ShellSession = {
    proc,
    buffer: [],
    created: now,
    lastActivity: now,
    cwd,
    alive: true,
    listeners: new Set(),
    idleTimer: null,
    waitingResponse: false,
  }

  const emitEvent = (event: TerminalEvent) => {
    for (const fn of session.listeners) {
      try { fn(event) } catch { /* */ }
    }
  }

  const pushOutput = (text: string) => {
    session.lastActivity = Date.now()
    session.buffer.push(text)
    let totalLen = 0
    for (const chunk of session.buffer) totalLen += chunk.length
    while (totalLen > 200_000 && session.buffer.length > 1) {
      totalLen -= session.buffer.shift()!.length
    }
    emitEvent({ type: 'output', text })

    // 응답 대기 중이면 idle 타이머 리셋
    if (session.waitingResponse) {
      if (session.idleTimer) clearTimeout(session.idleTimer)
      session.idleTimer = setTimeout(() => {
        if (!session.waitingResponse) return
        session.waitingResponse = false
        session.idleTimer = null
        // 클라이언트에 스냅샷 요청 — xterm.js 렌더링 결과를 그대로 사용
        emitEvent({ type: 'snapshot_request' })
      }, 2000)
    }
  }

  proc.stdout.on('data', (data: Buffer) => pushOutput(data.toString()))
  proc.stderr.on('data', (data: Buffer) => pushOutput(data.toString()))

  proc.on('close', () => {
    session.alive = false
    pushOutput('\r\n\x1b[90m[Session ended]\x1b[0m\r\n')
    emitEvent({ type: 'status', alive: false })
  })

  proc.on('error', (err) => {
    session.alive = false
    pushOutput(`\r\n\x1b[31m[Error: ${err.message}]\x1b[0m\r\n`)
    emitEvent({ type: 'status', alive: false })
  })

  sessions.set(id, session)

  setTimeout(() => {
    if (session.alive) {
      proc.stdin.write('claude --dangerously-skip-permissions\r')
    }
  }, 600)

  return id
}

/* ── GET: SSE stream ── */

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const action = searchParams.get('action') || 'stream'
  const sessionId = searchParams.get('session') || ''

  if (action === 'list') {
    const list = [...sessions.entries()].map(([id, s]) => ({
      id, alive: s.alive, created: s.created,
      age: Math.round((Date.now() - s.created) / 1000),
    }))
    return Response.json({ sessions: list })
  }

  if (!sessionId || !sessions.has(sessionId)) {
    return Response.json({ error: 'Invalid session' }, { status: 404 })
  }

  const session = sessions.get(sessionId)!
  const encoder = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      if (session.buffer.length > 0) {
        const replay = session.buffer.join('')
        controller.enqueue(
          encoder.encode(`data: ${JSON.stringify({ type: 'output', text: replay })}\n\n`),
        )
      }
      controller.enqueue(
        encoder.encode(`data: ${JSON.stringify({ type: 'status', alive: session.alive })}\n\n`),
      )

      const listener = (event: TerminalEvent) => {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(event)}\n\n`))
        } catch {
          session.listeners.delete(listener)
        }
      }
      session.listeners.add(listener)

      const heartbeat = setInterval(() => {
        try {
          controller.enqueue(encoder.encode(`: heartbeat\n\n`))
        } catch {
          clearInterval(heartbeat)
        }
      }, 15000)

      request.signal.addEventListener('abort', () => {
        session.listeners.delete(listener)
        clearInterval(heartbeat)
        try { controller.close() } catch { /* */ }
      })
    },
  })

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache, no-transform',
      'Connection': 'keep-alive',
      'X-Accel-Buffering': 'no',
    },
  })
}

/* ── POST ── */

export async function POST(request: NextRequest) {
  const body = await request.json()
  const action = body.action as string

  switch (action) {
    case 'create': {
      const cols = Number(body.cols) || 220
      const rows = Number(body.rows) || 50
      const cwd = (body.cwd as string) || '/tmp'
      const id = createSession(cols, rows, cwd)
      return Response.json({ ok: true, session: id })
    }

    case 'input': {
      const sessionId = body.session as string
      const data = body.data as string
      const session = sessions.get(sessionId)
      if (!session || !session.alive) {
        return Response.json({ error: 'Session not found or dead' }, { status: 404 })
      }
      session.lastActivity = Date.now()
      session.proc.stdin.write(data)

      // Enter 키 → 응답 대기 시작 (에코가 지나간 뒤)
      if (data.includes('\r')) {
        setTimeout(() => {
          if (session.idleTimer) clearTimeout(session.idleTimer)
          session.waitingResponse = true
          session.idleTimer = null
        }, 150)
      }

      return Response.json({ ok: true })
    }

    case 'resize': {
      const sessionId = body.session as string
      const cols = Number(body.cols)
      const rows = Number(body.rows)
      const session = sessions.get(sessionId)
      if (!session || !session.alive) {
        return Response.json({ error: 'Session not found or dead' }, { status: 404 })
      }
      if (!Number.isFinite(cols) || !Number.isFinite(rows) || cols < 2 || rows < 2) {
        return Response.json({ error: 'Invalid cols/rows' }, { status: 400 })
      }
      session.proc.stdin.write(`__RESIZE__:${cols}:${rows}\n`)
      return Response.json({ ok: true })
    }

    case 'kill': {
      const sessionId = body.session as string
      const session = sessions.get(sessionId)
      if (session) {
        if (session.idleTimer) clearTimeout(session.idleTimer)
        try { session.proc.kill('SIGTERM') } catch { /* */ }
        sessions.delete(sessionId)
      }
      return Response.json({ ok: true })
    }

    default:
      return Response.json({ error: `Unknown action: ${action}` }, { status: 400 })
  }
}
