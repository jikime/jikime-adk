'use client'

import { useEffect, useRef, useState } from 'react'
import { useServer } from '@/contexts/ServerContext'

type Status = 'connecting' | 'ready' | 'dead' | 'error'

const XTERM_STYLES = `
  .xterm .xterm-screen { outline: none !important; }
  .xterm:focus .xterm-screen { outline: none !important; }
  .xterm-screen:focus { outline: none !important; }
  .xterm canvas { outline: none !important; border: none !important; }
`

// 워터마크 기반 흐름 제어 (xterm.js 공식 권장)
const FLOW_HIGH = 100_000  // 100KB — PTY 일시 중지
const FLOW_LOW  =  10_000  //  10KB — PTY 재개

interface Props {
  tmuxSession: string  // e.g. "jikime-lotto-team-leader"
}

export default function AgentTerminalPanel({ tmuxSession }: Props) {
  const { getWsUrl }  = useServer()
  const containerRef  = useRef<HTMLDivElement>(null)
  const wsRef         = useRef<WebSocket | null>(null)
  const termRef       = useRef<import('@xterm/xterm').Terminal | null>(null)
  const fitRef        = useRef<import('@xterm/addon-fit').FitAddon | null>(null)
  const initRef       = useRef(false)
  const resizeTimer   = useRef<number | null>(null)
  const pendingBytes  = useRef(0)
  const [status, setStatus] = useState<Status>('connecting')

  useEffect(() => {
    if (!containerRef.current || initRef.current) return
    initRef.current = true
    let disposed = false

    const styleEl = document.createElement('style')
    styleEl.textContent = XTERM_STYLES
    document.head.appendChild(styleEl)

    ;(async () => {
      const { Terminal }        = await import('@xterm/xterm')
      const { FitAddon }        = await import('@xterm/addon-fit')
      const { WebLinksAddon }   = await import('@xterm/addon-web-links')
      const { Unicode11Addon }  = await import('@xterm/addon-unicode11')

      if (disposed || !containerRef.current) return

      const term = new Terminal({
        cursorBlink:       true,
        cursorStyle:       'bar',
        fontSize:          13,
        fontFamily:        'Menlo, Monaco, "Courier New", monospace',
        lineHeight:        1.35,
        allowProposedApi:  true,
        allowTransparency: false,
        // ✅ tmux는 자체적으로 \r\n을 관리 — convertEol:true 시 이중 개행/화면 깨짐
        convertEol:        false,
        // ✅ scrollback은 tmux가 담당 — xterm 버퍼와 충돌 방지
        scrollback:        0,
        macOptionIsMeta:   true,
        theme: {
          background:          '#0a0a0a',
          foreground:          '#d4d4d8',
          cursor:              '#00ff88',
          selectionBackground: '#264f7840',
          black:               '#09090b', red:           '#ef4444',
          green:               '#22c55e', yellow:        '#eab308',
          blue:                '#3b82f6', magenta:       '#a855f7',
          cyan:                '#06b6d4', white:         '#d4d4d8',
          brightBlack:         '#52525b', brightRed:     '#f87171',
          brightGreen:         '#4ade80', brightYellow:  '#facc15',
          brightBlue:          '#60a5fa', brightMagenta: '#c084fc',
          brightCyan:          '#22d3ee', brightWhite:   '#fafafa',
        },
      })

      // ✅ Unicode11 — tmux 박스 그리기 문자(─│┌┐) 정상 렌더링
      const unicode11 = new Unicode11Addon()
      term.loadAddon(unicode11)
      term.unicode.activeVersion = '11'

      const fitAddon = new FitAddon()
      fitRef.current = fitAddon
      term.loadAddon(fitAddon)
      term.loadAddon(new WebLinksAddon())

      // ✅ WebGL 렌더러 — tmux 이스케이프 시퀀스 고밀도 출력 시 CPU 절감
      try {
        const { WebglAddon } = await import('@xterm/addon-webgl')
        const webgl = new WebglAddon()
        webgl.onContextLoss(() => { webgl.dispose() })
        term.loadAddon(webgl)
      } catch { /* canvas renderer fallback */ }

      term.open(containerRef.current)
      termRef.current = term

      const ws = new WebSocket(getWsUrl('/ws/terminal'))
      wsRef.current = ws

      // ✅ 워터마크 흐름 제어 — 입력 지연 원인인 버퍼 포화 방지
      ws.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data as string)
          if (msg.type === 'output') {
            const data = msg.data as string
            pendingBytes.current += data.length
            // 버퍼 임계치 초과 시 PTY 일시 중지 요청
            if (pendingBytes.current > FLOW_HIGH) {
              ws.send(JSON.stringify({ type: 'pause' }))
            }
            // term.write 콜백: 렌더 완료 후 바이트 카운트 감산 → 임계치 이하면 재개
            term.write(data, () => {
              pendingBytes.current -= data.length
              if (pendingBytes.current < FLOW_LOW) {
                ws.send(JSON.stringify({ type: 'resume' }))
              }
            })
          } else if (msg.type === 'ready') {
            setStatus('ready')
            // linked session 이 열리면 fit → resize → Ctrl+L 순서로 화면 초기화
            window.setTimeout(() => {
              try { fitAddon.fit() } catch { /* */ }
              if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
                // Ctrl+L — tmux 에 현재 팬 전체를 새로 그리도록 요청
                ws.send(JSON.stringify({ type: 'input', data: '\x0c' }))
              }
            }, 400)
          } else if (msg.type === 'exit') {
            setStatus('dead')
          }
        } catch { /* */ }
      }
      ws.onerror = () => setStatus('error')
      ws.onclose = () => { if (!disposed) setStatus('dead') }

      // 키보드 입력 → tmux에 전달
      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN)
          ws.send(JSON.stringify({ type: 'input', data }))
      })

      const sendInit = () => {
        try { fitAddon.fit() } catch { /* */ }
        ws.send(JSON.stringify({
          type:      'init',
          sessionId: `tmux:${tmuxSession}`,
          cwd:       '/',
          cols:      term.cols,
          rows:      term.rows,
        }))
      }

      window.setTimeout(() => {
        if (disposed) return
        if (ws.readyState === WebSocket.OPEN) {
          sendInit()
        } else {
          ws.onopen = sendInit
        }
        term.focus()
      }, 100)

      // ResizeObserver — 50ms 디바운스
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
  }, [tmuxSession, getWsUrl]) // eslint-disable-line react-hooks/exhaustive-deps

  const statusColor =
    status === 'ready'      ? 'text-green-400'  :
    status === 'connecting' ? 'text-yellow-400' :
    status === 'dead'       ? 'text-zinc-500'   : 'text-red-400'

  const statusLabel =
    status === 'ready'      ? 'Live'       :
    status === 'connecting' ? 'Connecting' :
    status === 'dead'       ? 'Exited'     : 'Error'

  return (
    <div className="flex flex-col h-full bg-[#0a0a0a] overflow-hidden">
      <div className="flex items-center justify-between px-3 py-1.5 bg-zinc-900 border-b border-zinc-800 shrink-0">
        <div className="flex items-center gap-2">
          <div className="flex gap-1">
            <div className="w-2.5 h-2.5 rounded-full bg-red-500/70" />
            <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/70" />
            <div className="w-2.5 h-2.5 rounded-full bg-green-500/70" />
          </div>
          <span className="text-[11px] text-zinc-500 font-mono">{tmuxSession}</span>
        </div>
        <div className="flex items-center gap-1.5">
          {status === 'ready' && (
            <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />
          )}
          <span className={`text-[11px] font-medium ${statusColor}`}>{statusLabel}</span>
        </div>
      </div>

      <div className="relative flex-1 overflow-hidden p-[3px]">
        <div
          ref={containerRef}
          className="h-full w-full focus:outline-none"
          style={{ outline: 'none' }}
        />
      </div>
    </div>
  )
}
