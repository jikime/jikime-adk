'use client'

import { useEffect, useRef, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import { useLocale } from '@/contexts/LocaleContext'

type Status = 'connecting' | 'ready' | 'dead' | 'error'
const STATUS_VARIANT: Record<Status, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  connecting: 'secondary', ready: 'default', dead: 'outline', error: 'destructive',
}

// claudecodeui 와 동일한 xterm focus outline 제거
const XTERM_STYLES = `
  .xterm .xterm-screen { outline: none !important; }
  .xterm:focus .xterm-screen { outline: none !important; }
  .xterm-screen:focus { outline: none !important; }
  .xterm canvas { outline: none !important; border: none !important; }
`

export default function ShellPanel() {
  const { t } = useLocale()
  const { activeProject } = useProject()
  const { getWsUrl }      = useServer()
  const containerRef   = useRef<HTMLDivElement>(null)
  const wsRef          = useRef<WebSocket | null>(null)
  const termRef        = useRef<import('@xterm/xterm').Terminal | null>(null)
  const fitRef         = useRef<import('@xterm/addon-fit').FitAddon | null>(null)
  const sessionIdRef   = useRef<string>('')
  const initRef        = useRef(false)
  const resizeTimer    = useRef<number | null>(null)
  const [status, setStatus] = useState<Status>('connecting')

  useEffect(() => {
    if (!containerRef.current || initRef.current) return
    initRef.current = true
    let disposed = false

    // xterm CSS 스타일 주입 (claudecodeui 방식)
    const styleEl = document.createElement('style')
    styleEl.textContent = XTERM_STYLES
    document.head.appendChild(styleEl)

    ;(async () => {
      const { Terminal }     = await import('@xterm/xterm')
      const { FitAddon }     = await import('@xterm/addon-fit')
      const { WebLinksAddon } = await import('@xterm/addon-web-links')

      if (disposed || !containerRef.current) return

      const term = new Terminal({
        cursorBlink:                 true,
        cursorStyle:                 'bar',
        fontSize:                    13,
        fontFamily:                  'Menlo, Monaco, "Courier New", monospace',
        lineHeight:                  1.35,
        allowProposedApi:            true,
        allowTransparency:           false,
        convertEol:                  true,
        scrollback:                  10000,
        macOptionIsMeta:             true,
        theme: {
          background:        '#0a0a0a',
          foreground:        '#d4d4d8',
          cursor:            '#00ff88',
          selectionBackground: '#264f7840',
          black:             '#09090b', red:           '#ef4444',
          green:             '#22c55e', yellow:        '#eab308',
          blue:              '#3b82f6', magenta:       '#a855f7',
          cyan:              '#06b6d4', white:         '#d4d4d8',
          brightBlack:       '#52525b', brightRed:     '#f87171',
          brightGreen:       '#4ade80', brightYellow:  '#facc15',
          brightBlue:        '#60a5fa', brightMagenta: '#c084fc',
          brightCyan:        '#22d3ee', brightWhite:   '#fafafa',
        },
      })

      const fitAddon = new FitAddon()
      fitRef.current = fitAddon
      term.loadAddon(fitAddon)
      term.loadAddon(new WebLinksAddon())

      // WebGL 시도 → 실패 시 canvas 렌더러로 폴백 (claudecodeui 방식)
      try {
        const { WebglAddon } = await import('@xterm/addon-webgl')
        const webgl = new WebglAddon()
        webgl.onContextLoss(() => { webgl.dispose() })
        term.loadAddon(webgl)
      } catch { /* canvas renderer fallback */ }

      term.open(containerRef.current)
      termRef.current = term

      // ── WebSocket 연결 ──────────────────────────────────────────
      const ws = new WebSocket(getWsUrl('/ws/terminal'))
      wsRef.current = ws

      ws.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data)
          if      (msg.type === 'output') term.write(msg.data)
          else if (msg.type === 'ready')  setStatus('ready')
          else if (msg.type === 'exit')   setStatus('dead')
        } catch { /* */ }
      }
      ws.onerror  = () => setStatus('error')
      ws.onclose  = () => { if (!disposed) setStatus('dead') }

      // ── 키보드 입력 ─────────────────────────────────────────────
      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN)
          ws.send(JSON.stringify({ type: 'input', data }))
      })

      // ── 초기 fit → 정확한 크기로 init 전송 (claudecodeui 핵심) ──
      // 100ms 뒤 fit() 완료 후 cols/rows 확정, 그 후 WS init 전송
      window.setTimeout(() => {
        if (disposed) return
        try { fitAddon.fit() } catch { /* */ }

        const sendInit = () => {
          const sessionId = `term-${Date.now()}`
          sessionIdRef.current = sessionId
          ws.send(JSON.stringify({
            type:    'init',
            sessionId,
            cwd:     activeProject?.path || process.env.HOME || '/tmp',
            cols:    term.cols,
            rows:    term.rows,
          }))
        }

        if (ws.readyState === WebSocket.OPEN) {
          sendInit()
        } else {
          // WS가 아직 열리지 않은 경우 onopen에서 전송
          ws.onopen = sendInit
        }

        term.focus()
      }, 100)

      // ── ResizeObserver: 50ms 디바운스 (claudecodeui 방식) ───────
      const observer = new ResizeObserver(() => {
        if (resizeTimer.current !== null) clearTimeout(resizeTimer.current)
        resizeTimer.current = window.setTimeout(() => {
          if (disposed) return
          try {
            fitAddon.fit()
            if (ws.readyState === WebSocket.OPEN) {
              ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
            }
          } catch { /* */ }
        }, 50)
      })
      observer.observe(containerRef.current)

      // ── cleanup 등록 ────────────────────────────────────────────
      ;(containerRef.current as unknown as { __cleanup?: () => void }).__cleanup = () => {
        disposed = true
        if (resizeTimer.current !== null) clearTimeout(resizeTimer.current)
        observer.disconnect()
        ws.close()
        term.dispose()
        styleEl.remove()
      }
    })()

    return () => {
      disposed = true
      wsRef.current?.close()
      const el = containerRef.current as unknown as { __cleanup?: () => void } | null
      el?.__cleanup?.()
      initRef.current = false
    }
  }, [activeProject?.path, getWsUrl]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="flex flex-col h-full bg-[#0a0a0a] rounded-lg overflow-hidden border border-border">
      {/* 헤더 */}
      <div className="flex items-center justify-between px-4 py-2 bg-card border-b border-border shrink-0">
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-3 h-3 rounded-full bg-red-500/80" />
            <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
            <div className="w-3 h-3 rounded-full bg-green-500/80" />
          </div>
          <span className="text-xs text-muted-foreground ml-1 font-mono">Claude Code Terminal</span>
        </div>
        <Badge variant={STATUS_VARIANT[status]} className="text-xs">
          {status === 'connecting' ? t.shell.connecting : status === 'ready' ? t.shell.ready : status === 'dead' ? t.shell.dead : t.shell.error}
        </Badge>
      </div>

      {/* xterm 컨테이너: h-full w-full, outline 없음 (claudecodeui 동일) */}
      <div className="relative flex-1 overflow-hidden p-[4px]">
        <div
          ref={containerRef}
          className="h-full w-full focus:outline-none"
          style={{ outline: 'none' }}
        />
      </div>
    </div>
  )
}
