'use client'

import {
  createContext, useContext, useState, useEffect,
  useCallback, useMemo, ReactNode,
} from 'react'

export interface RemoteServer {
  id: string
  name: string     // 표시 이름 (예: "개발 서버", "프로덕션")
  host: string     // "host:port" (예: "192.168.1.100:3000")
  secure: boolean  // true → wss:// / https://
}

interface ServerContextType {
  servers: RemoteServer[]
  activeServer: RemoteServer | null
  setActiveServerId: (id: string) => void
  addServer: (s: Omit<RemoteServer, 'id'>) => void
  updateServer: (id: string, patch: Partial<Omit<RemoteServer, 'id'>>) => void
  removeServer: (id: string) => void
  getWsUrl: (path: string) => string   // ws(s)://host/ws/terminal 등
  getApiUrl: (path: string) => string  // http(s)://host/api/ws/files 등
}

const ServerContext = createContext<ServerContextType | null>(null)

const STORAGE_KEY = 'webchat_servers'
const LOCAL_ID    = '__local__'
// host:port 형식 — 외부 입력 검증용
const HOST_RE = /^[\w.-]+(:\d{1,5})?$/

function loadFromStorage(): { servers: RemoteServer[]; activeId: string } | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : null
  } catch { return null }
}

function saveToStorage(servers: RemoteServer[], activeId: string) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ servers, activeId }))
  } catch { /* */ }
}

export function ServerProvider({ children }: { children: ReactNode }) {
  const [servers,  setServers]  = useState<RemoteServer[]>([])
  const [activeId, setActiveId] = useState<string>(LOCAL_ID)

  // 클라이언트에서만 초기화 (SSR 안전)
  useEffect(() => {
    const stored = loadFromStorage()
    if (stored && stored.servers.length > 0) {
      setServers(stored.servers)
      setActiveId(stored.activeId)
    } else {
      const local: RemoteServer = {
        id:     LOCAL_ID,
        name:   '로컬 (localhost)',
        host:   window.location.host,
        secure: window.location.protocol === 'https:',
      }
      setServers([local])
      setActiveId(LOCAL_ID)
    }
  }, [])

  // 변경 시 localStorage 동기화
  useEffect(() => {
    if (servers.length > 0) saveToStorage(servers, activeId)
  }, [servers, activeId])

  const activeServer = useMemo(
    () => servers.find(s => s.id === activeId) ?? servers[0] ?? null,
    [servers, activeId],
  )

  const setActiveServerId = useCallback((id: string) => setActiveId(id), [])

  const addServer = useCallback((s: Omit<RemoteServer, 'id'>) => {
    const host = s.host.trim()
    if (!HOST_RE.test(host)) return  // 잘못된 host 형식 무시
    const id = `server-${Date.now()}`
    setServers(prev => [...prev, { ...s, host, id }])
  }, [])

  const updateServer = useCallback((id: string, patch: Partial<Omit<RemoteServer, 'id'>>) => {
    if (patch.host !== undefined && !HOST_RE.test(patch.host.trim())) return  // 잘못된 host 형식 무시
    setServers(prev => prev.map(s => s.id === id ? { ...s, ...patch } : s))
  }, [])

  const removeServer = useCallback((id: string) => {
    if (id === LOCAL_ID) return  // 로컬 서버는 삭제 불가
    setServers(prev => {
      const next = prev.filter(s => s.id !== id)
      // 삭제된 서버가 active면 로컬로 전환
      setActiveId(cur => cur === id ? LOCAL_ID : cur)
      return next
    })
  }, [])

  const getWsUrl = useCallback((path: string): string => {
    if (!activeServer) return path
    const proto = activeServer.secure ? 'wss:' : 'ws:'
    return `${proto}//${activeServer.host}${path}`
  }, [activeServer])

  const getApiUrl = useCallback((path: string): string => {
    if (!activeServer) return path
    const proto = activeServer.secure ? 'https:' : 'http:'
    // 로컬 서버는 상대 경로 그대로 사용 (CORS 불필요)
    if (activeServer.id === LOCAL_ID) return path
    return `${proto}//${activeServer.host}${path}`
  }, [activeServer])

  const value = useMemo(() => ({
    servers, activeServer,
    setActiveServerId, addServer, updateServer, removeServer,
    getWsUrl, getApiUrl,
  }), [servers, activeServer, setActiveServerId, addServer, updateServer, removeServer, getWsUrl, getApiUrl])

  return (
    <ServerContext.Provider value={value}>
      {children}
    </ServerContext.Provider>
  )
}

export function useServer() {
  const ctx = useContext(ServerContext)
  if (!ctx) throw new Error('useServer must be used inside ServerProvider')
  return ctx
}
