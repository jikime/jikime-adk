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
  refreshProjects: () => Promise<void>
}

const ProjectContext = createContext<ProjectContextType | null>(null)

// ── URL ↔ Path 변환 ─────────────────────────────────────────────────
// URL 포맷: /project/Users/jikime/Dev/...
// 파일시스템 경로: /Users/jikime/Dev/...
//
// 인코딩: 맨 앞 '/' 제거 → /project/ 뒤에 붙임
// 디코딩: /project/ 이후 세그먼트에 '/' 다시 붙임
// Next.js [...segments] catch-all이 '/'를 포함한 경로를 자연스럽게 처리

const PROJECT_PREFIX = '/project/'

export function pathToUrl(fsPath: string): string {
  // '/Users/foo/bar' → '/project/Users/foo/bar'
  return PROJECT_PREFIX + fsPath.replace(/^\//, '')
}

export function urlToPath(pathname: string): string | null {
  if (!pathname.startsWith(PROJECT_PREFIX)) return null
  // '/project/Users/foo/bar' → '/Users/foo/bar'
  return '/' + pathname.slice(PROJECT_PREFIX.length)
}

// ── Provider ────────────────────────────────────────────────────────

export function ProjectProvider({ children }: { children: ReactNode }) {
  const { getApiUrl, activeServer } = useServer()
  const router   = useRouter()
  const pathname = usePathname()

  const [projects, setProjects]                 = useState<Project[]>([])
  const [activeProject, setActiveProjectState]  = useState<Project | null>(null)
  const [activeSessionId, setActiveSessionId]   = useState<string | null>(null)
  const prevServerIdRef = useRef<string | null>(null)
  const restoredRef     = useRef(false)

  // 프로젝트 선택 + URL 업데이트 (Next.js router)
  const setActiveProject = useCallback((p: Project | null) => {
    setActiveProjectState(p)
    router.push(p ? pathToUrl(p.path) : '/')
  }, [router])

  const refreshProjects = useCallback(async () => {
    try {
      const res = await fetch(getApiUrl('/api/ws/projects'))
      if (res.ok) {
        const data: Project[] = await res.json()
        setProjects(data)
        setActiveProjectState(prev => {
          if (prev !== null) return prev
          // URL에 프로젝트 경로가 있으면 우선 선택, 없으면 첫 번째 프로젝트
          const pathFromUrl = urlToPath(pathname)
          if (pathFromUrl) {
            const found = data.find(p => p.path === pathFromUrl)
            if (found) return found
          }
          return data.length > 0 ? data[0] : null
        })
      }
    } catch { /* */ }
  }, [getApiUrl, pathname])

  // URL이 /project/... 형태로 변경되면 해당 프로젝트 자동 선택 (최초 1회)
  useEffect(() => {
    if (restoredRef.current || projects.length === 0) return
    const pathFromUrl = urlToPath(pathname)
    if (!pathFromUrl) return
    const found = projects.find(p => p.path === pathFromUrl)
    if (found) {
      restoredRef.current = true
      setActiveProjectState(found)
    }
  }, [projects, pathname])

  // 서버가 바뀌면 상태 초기화 후 새 서버에서 로드
  useEffect(() => {
    if (!activeServer) return
    if (prevServerIdRef.current !== activeServer.id) {
      const isInitialMount = prevServerIdRef.current === null
      prevServerIdRef.current = activeServer.id
      restoredRef.current = false
      setProjects([])
      setActiveProjectState(null)
      setActiveSessionId(null)
      // 초기 마운트가 아닌 실제 서버 전환 시에만 루트로 이동
      if (!isInitialMount) router.replace('/')
      refreshProjects()
    }
  }, [activeServer, refreshProjects, router])

  const value = useMemo(() => ({
    projects, activeProject, activeSessionId,
    setActiveProject, setActiveSessionId, refreshProjects,
  }), [projects, activeProject, activeSessionId, setActiveProject, refreshProjects])

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
