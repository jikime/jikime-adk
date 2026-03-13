import { NextRequest } from 'next/server'
import { getProject } from '@/lib/store'
import { isRunning, isPortRunning } from '@/lib/serve'

export const dynamic = 'force-dynamic'

/**
 * GET /api/state?projectId=xxx
 * jikime serve /api/v1/state를 SSE로 프록시
 * 3초마다 폴링 → 브라우저에 스트리밍
 */
export async function GET(request: NextRequest) {
  const projectId = new URL(request.url).searchParams.get('projectId')
  if (!projectId) return Response.json({ error: 'projectId 필요' }, { status: 400 })

  const project = await getProject(projectId)
  if (!project) return Response.json({ error: '없는 프로젝트' }, { status: 404 })

  const encoder = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      const send = (data: object) => {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch { /* closed */ }
      }

      let stopped = false

      const poll = async () => {
        if (stopped) return

        // 프로세스 상태 확인 (pid → 포트 순서로 fallback)
        const running = !!(project.pid && isRunning(project.pid)) || await isPortRunning(project.port)
        if (!running) {
          send({ type: 'status', running: false })
          stopped = true
          try { controller.close() } catch { /* */ }
          return
        }

        try {
          const res = await fetch(`http://127.0.0.1:${project.port}/api/v1/state`, {
            signal: AbortSignal.timeout(5000),
          })
          if (res.ok) {
            const state = await res.json()
            send({ type: 'state', data: state })
          }
        } catch {
          send({ type: 'status', running: true, error: 'jikime serve 연결 실패' })
        }

        if (!stopped) setTimeout(poll, 3000)
      }

      // 즉시 첫 폴링
      poll()

      request.signal.addEventListener('abort', () => {
        stopped = true
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

/** POST /api/state?projectId=xxx → jikime serve /api/v1/refresh 호출 */
export async function POST(request: NextRequest) {
  const projectId = new URL(request.url).searchParams.get('projectId')
  if (!projectId) return Response.json({ error: 'projectId 필요' }, { status: 400 })

  const project = await getProject(projectId)
  if (!project) return Response.json({ error: '없는 프로젝트' }, { status: 404 })

  try {
    await fetch(`http://127.0.0.1:${project.port}/api/v1/refresh`, {
      method: 'POST',
      signal: AbortSignal.timeout(5000),
    })
    return Response.json({ ok: true })
  } catch {
    return Response.json({ error: '연결 실패' }, { status: 503 })
  }
}
