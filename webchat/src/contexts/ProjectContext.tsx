'use client'

import { createContext, useContext, useState, useEffect, useCallback, useRef, useMemo, ReactNode } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { useServer } from '@/contexts/ServerContext'

export interface Project {
  id: string
  name: string
  path: string
  sessions: string[]
}

interface ProjectContextType {
  projects: Project[]
  activeProject: Project | null
  activeSessionId: string | null
  setActiveProject: (p: Project | null) => void
  setActiveSessionId: (id: string | null) => void
  navigateToSession: (p: Project, sessionId: string) => void
  refreshProjects: () => Promise<void>
}

const ProjectContext = createContext<ProjectContextType | null>(null)

// /session/{sessionId} URL 포맷
const SESSION_PREFIX = '/session/'

// ── Provider ────────────────────────────────────────────────────────

export function ProjectProvider({ children }: { children: ReactNode }) {
  const { getApiUrl, activeServer } = useServer()
  const router   = useRouter()
  const pathname = usePathname()

  const [projects, setProjects]                 = useState<Project[]>([])
  const [activeProject, setActiveProjectState]  = useState<Project | null>(null)
  const [activeSessionId, setActiveSessionId]   = useState<string | null>(null)
  const prevServerIdRef = useRef<string | null>(null)
  // 마지막으로 복원한 pathname 저장 — 같은 URL 재복원 방지, URL 변경 시 재복원 허용
  const restoredPathnameRef = useRef<string | null>(null)

  // 프로젝트 선택 — URL 변경 없이 상태만 업데이트
  const setActiveProject = useCallback((p: Project | null) => {
    setActiveProjectState(p)
  }, [])

  // 세션 선택 + URL 이동: /session/{sessionId}
  const navigateToSession = useCallback((p: Project, sessionId: string) => {
    setActiveProjectState(p)
    if (sessionId) {
      setActiveSessionId(sessionId)
      router.push(`${SESSION_PREFIX}${sessionId}`)
    } else {
      // 새 채팅: 세션 ID 없음 → 루트로
      setActiveSessionId(null)
      router.push('/')
    }
  }, [router])

  const refreshProjects = useCallback(async () => {
    try {
      const res = await fetch(getApiUrl('/api/ws/projects'))
      if (res.ok) {
        const data: Project[] = await res.json()
        setProjects(data)
        // 자동 선택 없음 — URL 복원 effect가 처리하거나, 사용자가 직접 선택
      }
    } catch { /* */ }
  }, [getApiUrl])

  // URL이 /session/... 형태이면 해당 세션과 프로젝트를 복원
  // pathname이 바뀔 때마다 재실행 — 같은 URL은 ref로 중복 방지
  useEffect(() => {
    if (projects.length === 0) return
    if (!pathname.startsWith(SESSION_PREFIX)) return

    const sessionId = pathname.slice(SESSION_PREFIX.length)
    if (!sessionId) return

    // 이미 이 pathname으로 복원했으면 건너뜀 (URL 이동 없이 프로젝트 재로드 시 중복 방지)
    if (restoredPathnameRef.current === pathname) return

    // 이미 로드된 프로젝트에서 세션 검색
    for (const p of projects) {
      if (p.sessions.includes(sessionId)) {
        restoredPathnameRef.current = pathname
        setActiveProjectState(p)
        setActiveSessionId(sessionId)
        return
      }
    }

    // 세션이 프로젝트 목록에 없으면 서버에 lookup 요청
    ;(async () => {
      try {
        const res = await fetch(getApiUrl(`/api/ws/session-lookup?sessionId=${encodeURIComponent(sessionId)}`))
        if (res.ok) {
          const data = await res.json() as { projectPath?: string }
          if (data.projectPath) {
            const found = projects.find(p => p.path === data.projectPath)
            if (found) {
              restoredPathnameRef.current = pathname
              setActiveProjectState(found)
              setActiveSessionId(sessionId)
            }
          }
        }
      } catch { /* */ }
    })()
  }, [projects, pathname, getApiUrl])

  // 서버가 바뀌면 상태 초기화 후 새 서버에서 로드
  useEffect(() => {
    if (!activeServer) return
    if (prevServerIdRef.current !== activeServer.id) {
      const isInitialMount = prevServerIdRef.current === null
      prevServerIdRef.current = activeServer.id
      restoredPathnameRef.current = null
      setProjects([])
      setActiveProjectState(null)
      setActiveSessionId(null)
      if (!isInitialMount) router.replace('/')
      refreshProjects()
    }
  }, [activeServer, refreshProjects, router])

  const value = useMemo(() => ({
    projects, activeProject, activeSessionId,
    setActiveProject, setActiveSessionId, navigateToSession, refreshProjects,
  }), [projects, activeProject, activeSessionId, setActiveProject, navigateToSession, refreshProjects])

  return (
    <ProjectContext.Provider value={value}>
      {children}
    </ProjectContext.Provider>
  )
}

export function useProject() {
  const ctx = useContext(ProjectContext)
  if (!ctx) throw new Error('useProject must be used inside ProjectProvider')
  return ctx
}
