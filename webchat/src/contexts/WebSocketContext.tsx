'use client'

import { createContext, useContext, useEffect, useRef, useState, useCallback, useMemo, ReactNode } from 'react'
import { useServer } from '@/contexts/ServerContext'

type WsMessage = Record<string, unknown>

interface WebSocketContextType {
  isConnected: boolean
  sendMessage: (msg: WsMessage) => void
  onMessage: (handler: (msg: WsMessage) => void) => () => void
}

const WebSocketContext = createContext<WebSocketContextType | null>(null)

const MAX_RECONNECT_ATTEMPTS = 10
const BASE_RECONNECT_DELAY_MS = 1_000
const MAX_RECONNECT_DELAY_MS  = 30_000

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const { getWsUrl } = useServer()
  const wsRef              = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const handlersRef        = useRef<Set<(msg: WsMessage) => void>>(new Set())
  const reconnectTimer     = useRef<ReturnType<typeof setTimeout> | null>(null)
  const disposedRef        = useRef(false)
  const reconnectAttempts  = useRef(0)
  // 항상 최신 connect를 가리키는 ref — stale closure 방지
  const connectRef         = useRef<() => void>(() => {})

  const connect = useCallback(() => {
    if (disposedRef.current) return
    const url = getWsUrl('/ws/chat')
    const ws  = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      reconnectAttempts.current = 0  // 연결 성공 시 재시도 카운터 초기화
      setIsConnected(true)
    }
    ws.onclose = () => {
      setIsConnected(false)
      if (!disposedRef.current && reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
        // Exponential backoff: 1s, 2s, 4s, 8s, ... (max 30s)
        const delay = Math.min(
          BASE_RECONNECT_DELAY_MS * Math.pow(2, reconnectAttempts.current),
          MAX_RECONNECT_DELAY_MS,
        )
        reconnectAttempts.current++
        reconnectTimer.current = setTimeout(() => connectRef.current(), delay)
      }
    }
    ws.onerror = () => ws.close()
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data)
        for (const handler of handlersRef.current) handler(msg)
      } catch { /* */ }
    }
  }, [getWsUrl])

  // connect가 바뀔 때마다 ref 동기화
  useEffect(() => {
    connectRef.current = connect
  }, [connect])

  // 서버(getWsUrl)가 바뀌면 기존 연결 끊고 새 서버로 재연결
  useEffect(() => {
    disposedRef.current     = false
    reconnectAttempts.current = 0  // 서버 전환 시 재시도 카운터 초기화
    if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
    wsRef.current?.close()
    connect()
    return () => {
      disposedRef.current = true
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  const sendMessage = useCallback((msg: WsMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN)
      wsRef.current.send(JSON.stringify(msg))
  }, [])

  const onMessage = useCallback((handler: (msg: WsMessage) => void) => {
    // 핸들러 하드 캡 100 — 무제한 누적으로 인한 메모리/CPU 고갈 방지
    if (handlersRef.current.size >= 100) {
      console.error(`[WebSocket] handler cap (100) reached — rejecting new handler; check useEffect cleanup`)
      return () => { /* no-op: handler was not registered */ }
    }
    handlersRef.current.add(handler)
    if (handlersRef.current.size > 50) {
      console.warn(`[WebSocket] handler count (${handlersRef.current.size}) exceeds 50 — possible leak; ensure cleanup function is called on unmount`)
    }
    return () => { handlersRef.current.delete(handler) }
  }, [])

  const value = useMemo(
    () => ({ isConnected, sendMessage, onMessage }),
    [isConnected, sendMessage, onMessage],
  )

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocket() {
  const ctx = useContext(WebSocketContext)
  if (!ctx) throw new Error('useWebSocket must be used inside WebSocketProvider')
  return ctx
}
