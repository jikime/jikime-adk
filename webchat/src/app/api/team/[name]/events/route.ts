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
      let closed = false

      const send = (data: unknown) => {
        if (closed) return
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch { closed = true }
      }

      // SSE retry hint
      controller.enqueue(encoder.encode('retry: 2000\n\n'))

      // Initial snapshot
      send(buildTeamSnapshot(name))

      // Debounce: 500ms after last file change before sending snapshot
      let debounceTimer: NodeJS.Timeout | null = null
      const debouncedSend = () => {
        if (closed) return
        if (debounceTimer) clearTimeout(debounceTimer)
        debounceTimer = setTimeout(() => {
          debounceTimer = null
          send(buildTeamSnapshot(name))
        }, 500)
      }

      // Heartbeat every 30s — keeps connection alive and detects stale clients
      const heartbeat = setInterval(() => {
        if (closed) return
        try {
          controller.enqueue(encoder.encode(': heartbeat\n\n'))
        } catch { closed = true }
      }, 30_000)

      // Watch team subdirs for changes
      const watchDirs = ['tasks', 'registry', 'inbox', 'costs']
      const watchers: fs.FSWatcher[] = []

      for (const dir of watchDirs) {
        const dirPath = path.join(teamDir(name), dir)
        if (!fs.existsSync(dirPath)) continue
        try {
          const w = fs.watch(dirPath, { recursive: false }, debouncedSend)
          watchers.push(w)
        } catch { /* dir may not be watchable */ }
      }

      // Cleanup when client disconnects
      request.signal.addEventListener('abort', () => {
        closed = true
        if (debounceTimer) clearTimeout(debounceTimer)
        clearInterval(heartbeat)
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
