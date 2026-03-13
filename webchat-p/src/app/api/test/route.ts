import { NextRequest } from 'next/server'

export const dynamic = 'force-dynamic'

export async function GET(_request: NextRequest) {
  const encoder = new TextEncoder()
  const stream = new ReadableStream({
    start(controller) {
      const send = (data: object) => {
        controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
      }
      send({ type: 'text', text: 'SSE 테스트 1' })
      setTimeout(() => send({ type: 'text', text: 'SSE 테스트 2' }), 500)
      setTimeout(() => {
        send({ type: 'done' })
        controller.close()
      }, 1000)
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
