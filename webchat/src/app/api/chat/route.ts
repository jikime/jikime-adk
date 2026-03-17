import { NextRequest } from 'next/server'
import { spawn } from 'child_process'

export const dynamic = 'force-dynamic'

/**
 * claude -p "message" --output-format stream-json 을 실행해
 * 응답 텍스트를 SSE 스트림으로 반환합니다.
 *
 * stream-json 이벤트 구조 (Claude Code CLI):
 *   { type: "system", ... }
 *   { type: "assistant", message: { content: [{ type: "text", text: "..." }] }, ... }
 *   { type: "result", result: "...", ... }
 */
export async function POST(request: NextRequest) {
  const body = await request.json()
  const message = (body.message as string)?.trim()
  const cwd = (body.cwd as string) || '/tmp'

  if (!message) {
    return Response.json({ error: 'No message' }, { status: 400 })
  }

  // CLAUDECODE="" 로 설정 → nested session 오류 방지 (delete는 안됨)
  const env = { ...process.env, CLAUDECODE: '' }

  const encoder = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      const send = (data: object) => {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch { /* controller already closed */ }
      }

      const proc = spawn(
        'claude',
        ['-p', message, '--output-format', 'stream-json', '--verbose', '--dangerously-skip-permissions'],
        {
          cwd,
          env: { ...env, FORCE_COLOR: '0', NO_COLOR: '1', TERM: 'dumb' },
        },
      )

      let buf = ''
      let sentText = false

      proc.stdout.on('data', (chunk: Buffer) => {
        buf += chunk.toString()
        const lines = buf.split('\n')
        buf = lines.pop() || ''

        for (const line of lines) {
          if (!line.trim()) continue
          try {
            const event = JSON.parse(line)

            // assistant 이벤트: 실제 응답 텍스트 추출
            if (event.type === 'assistant' && Array.isArray(event.message?.content)) {
              for (const block of event.message.content) {
                if (block.type === 'text' && block.text) {
                  send({ type: 'text', text: block.text })
                  sentText = true
                }
              }
            }

            // result 이벤트: 최종 텍스트 (assistant 이벤트 없을 때 fallback)
            if (event.type === 'result' && event.result && !sentText) {
              send({ type: 'text', text: event.result })
            }
          } catch { /* skip invalid JSON lines */ }
        }
      })

      proc.stderr.on('data', (chunk: Buffer) => {
        const text = chunk.toString().trim()
        if (text) console.error('[claude -p stderr]', text)
      })

      proc.on('close', (code) => {
        if (code !== 0 && !sentText) {
          send({ type: 'error', text: `Claude 실행 오류 (exit code: ${code})` })
        }
        send({ type: 'done' })
        try { controller.close() } catch { /* */ }
      })

      proc.on('error', (err) => {
        send({ type: 'error', text: `claude 명령어를 찾을 수 없어요: ${err.message}` })
        send({ type: 'done' })
        try { controller.close() } catch { /* */ }
      })

      request.signal.addEventListener('abort', () => {
        proc.kill('SIGTERM')
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
