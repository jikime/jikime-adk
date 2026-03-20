import { NextRequest } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import { teamDir, buildTeamSnapshot } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(request: NextRequest, { params }: Params) {
  const { name } = await params
  const encoder  = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      const send = (data: unknown) => {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch { /* client disconnected */ }
      }

      // SSE retry hint
      controller.enqueue(encoder.encode('retry: 2000\n\n'))

      // Initial snapshot
      send(buildTeamSnapshot(name))

      // Watch team subdirs for changes
      const watchDirs = ['tasks', 'registry', 'inbox', 'costs']
      const watchers: fs.FSWatcher[] = []

      for (const dir of watchDirs) {
        const dirPath = path.join(teamDir(name), dir)
        if (!fs.existsSync(dirPath)) continue
        try {
          const w = fs.watch(dirPath, { recursive: false }, () => {
            send(buildTeamSnapshot(name))
          })
          watchers.push(w)
        } catch { /* dir may not be watchable */ }
      }

      // Cleanup when client disconnects
      request.signal.addEventListener('abort', () => {
        watchers.forEach((w) => { try { w.close() } catch { /* */ } })
        try { controller.close() } catch { /* */ }
      })
    },
  })

  return new Response(stream, {
    headers: {
      'Content-Type':  'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection':    'keep-alive',
    },
  })
}
