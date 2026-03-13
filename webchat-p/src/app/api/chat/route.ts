import { NextRequest } from 'next/server'
import { spawn, execSync } from 'child_process'

export const dynamic = 'force-dynamic'

function findClaude(): string {
  const candidates = [
    'claude',
    `${process.env.HOME}/.local/bin/claude`,
    '/usr/local/bin/claude',
    '/opt/homebrew/bin/claude',
  ]
  for (const c of candidates) {
    try {
      execSync(`${c} --version`, { stdio: 'pipe', timeout: 3000 })
      return c
    } catch { /* continue */ }
  }
  return 'claude'
}

let cachedClaude: string | null = null
function getClaude(): string {
  if (!cachedClaude) cachedClaude = findClaude()
  return cachedClaude
}

/**
 * SSE 이벤트 타입:
 *   { type: 'session_id', id }
 *   { type: 'tool_call',  name, input }   ← 도구 호출 시작
 *   { type: 'tool_result', content }      ← 도구 실행 결과
 *   { type: 'text', text }                ← 최종 답변 텍스트
 *   { type: 'error', text }
 *   { type: 'done' }
 */
export async function POST(request: NextRequest) {
  const body = await request.json()
  const message = (body.message as string)?.trim()
  const sessionId = (body.sessionId as string) || ''
  const cwd = (body.cwd as string) || process.env.HOME || '/tmp'

  if (!message) {
    return Response.json({ error: 'No message' }, { status: 400 })
  }

  const claudePath = getClaude()
  const args = [
    '-p', message,
    '--output-format', 'stream-json',
    '--verbose',
    '--dangerously-skip-permissions',
  ]
  if (sessionId) args.push('--resume', sessionId)

  const env: Record<string, string> = {}
  for (const [k, v] of Object.entries(process.env)) {
    if (k !== 'CLAUDECODE' && v !== undefined) env[k] = v
  }
  env.FORCE_COLOR = '0'
  env.NO_COLOR = '1'
  env.TERM = 'dumb'
  env.PATH = (process.env.PATH ?? '') + `:${process.env.HOME}/.local/bin`

  const encoder = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      const send = (data: object) => {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch { /* closed */ }
      }

      let proc: ReturnType<typeof spawn>
      try {
        proc = spawn(claudePath, args, { cwd, env: env as NodeJS.ProcessEnv, stdio: ['ignore', 'pipe', 'pipe'] })
      } catch (err) {
        send({ type: 'error', text: `실행 오류: ${(err as Error).message}` })
        send({ type: 'done' })
        try { controller.close() } catch { /* */ }
        return
      }

      let buf = ''
      let sentText = false
      let extractedSessionId = ''

      proc.stdout?.on('data', (chunk: Buffer) => {
        buf += chunk.toString()
        const lines = buf.split('\n')
        buf = lines.pop() || ''

        for (const line of lines) {
          if (!line.trim()) continue
          try {
            const event = JSON.parse(line)

            // session_id 추출
            if (event.session_id && !extractedSessionId) {
              extractedSessionId = event.session_id
              send({ type: 'session_id', id: extractedSessionId })
            }

            // assistant 이벤트 — thinking / tool_use / text 블록 처리
            if (event.type === 'assistant' && Array.isArray(event.message?.content)) {
              for (const block of event.message.content) {
                if (block.type === 'tool_use') {
                  // 도구 호출 — input을 읽기 좋게 포맷
                  const inputStr = formatToolInput(block.name, block.input)
                  send({ type: 'tool_call', name: block.name as string, input: inputStr })
                }
                if (block.type === 'text' && block.text) {
                  send({ type: 'text', text: block.text as string })
                  sentText = true
                }
                // thinking 블록은 생략 (내부 추론이라 사용자에게 불필요)
              }
            }

            // user 이벤트 — tool_result 블록 처리
            if (event.type === 'user' && Array.isArray(event.message?.content)) {
              for (const block of event.message.content) {
                if (block.type === 'tool_result') {
                  const content = extractToolResult(block.content)
                  send({ type: 'tool_result', content })
                }
              }
            }

            // result fallback (text가 없을 때)
            if (event.type === 'result') {
              if (event.is_error && !sentText) {
                // is_error: true → result가 빈 문자열인 경우도 처리
                const errMsg = (event.result as string) || '실행 중 오류가 발생했어요.'
                send({ type: 'error', text: errMsg })
                sentText = true
              } else if (event.result && !sentText) {
                send({ type: 'text', text: event.result as string })
                sentText = true
              }
            }
          } catch { /* invalid JSON */ }
        }
      })

      proc.stderr?.on('data', (chunk: Buffer) => {
        const text = chunk.toString().trim()
        if (text) console.error('[claude stderr]', text)
      })

      proc.on('close', (code) => {
        if (!sentText) {
          if (code !== 0) {
            send({ type: 'error', text: `오류가 발생했어요 (exit ${code})` })
          } else {
            // exit 0이지만 텍스트가 없는 경우 — 슬래시 커맨드 미지원 등
            send({ type: 'error', text: '응답이 없어요. 슬래시 커맨드(/)는 -p 모드에서 지원되지 않을 수 있어요.' })
          }
        }
        send({ type: 'done' })
        try { controller.close() } catch { /* */ }
      })

      proc.on('error', (err) => {
        send({ type: 'error', text: `claude를 찾을 수 없어요: ${err.message}` })
        send({ type: 'done' })
        try { controller.close() } catch { /* */ }
      })

      request.signal.addEventListener('abort', () => {
        try { proc.kill('SIGTERM') } catch { /* */ }
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

/** 도구 입력을 사람이 읽기 쉬운 문자열로 변환 */
function formatToolInput(toolName: string, input: Record<string, unknown>): string {
  if (!input) return ''
  // Bash: command 필드만 표시
  if (toolName === 'Bash' && input.command) return String(input.command)
  // Read/Write/Edit: file_path 우선
  if (input.file_path) return String(input.file_path)
  if (input.path) return String(input.path)
  // 나머지는 JSON
  return JSON.stringify(input, null, 2)
}

/** tool_result content를 문자열로 추출 */
function extractToolResult(content: unknown): string {
  if (typeof content === 'string') return content
  if (Array.isArray(content)) {
    return content
      .map(c => (typeof c === 'string' ? c : (c as { text?: string }).text ?? JSON.stringify(c)))
      .join('\n')
  }
  return JSON.stringify(content ?? '')
}
