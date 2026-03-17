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

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const { getWsUrl } = useServer()
  const wsRef          = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const handlersRef    = useRef<Set<(msg: WsMessage) => void>>(new Set())
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const disposedRef    = useRef(false)

  const connect = useCallback(() => {
    if (disposedRef.current) return
    const url = getWsUrl('/ws/chat')
    const ws  = new WebSocket(url)
    wsRef.current = ws

    ws.onopen  = () => setIsConnected(true)
    ws.onclose = () => {
      setIsConnected(false)
      if (!disposedRef.current)
        reconnectTimer.current = setTimeout(connect, 3000)
    }
    ws.onerror = () => ws.close()
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data)
        for (const handler of handlersRef.current) handler(msg)
      } catch { /* */ }
    }
  }, [getWsUrl])

  // 서버(getWsUrl)가 바뀌면 기존 연결 끊고 새 서버로 재연결
  useEffect(() => {
    disposedRef.current = false
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
    handlersRef.current.add(handler)
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
