'use client'

import { createContext, useContext, useState, useEffect, useCallback, useRef, useMemo, ReactNode } from 'react'
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

export function ProjectProvider({ children }: { children: ReactNode }) {
  const { getApiUrl, activeServer } = useServer()
  const [projects, setProjects]     = useState<Project[]>([])
  const [activeProject, setActiveProject]   = useState<Project | null>(null)
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null)
  const prevServerIdRef = useRef<string | null>(null)

  const refreshProjects = useCallback(async () => {
    try {
      const res = await fetch(getApiUrl('/api/ws/projects'))
      if (res.ok) {
        const data: Project[] = await res.json()
        setProjects(data)
        setActiveProject(prev => (prev === null && data.length > 0) ? data[0] : prev)
      }
    } catch { /* */ }
  }, [getApiUrl])

  // 서버가 바뀌면 프로젝트 목록 초기화 후 새 서버에서 로드
  useEffect(() => {
    if (!activeServer) return
    if (prevServerIdRef.current !== activeServer.id) {
      prevServerIdRef.current = activeServer.id
      setProjects([])
      setActiveProject(null)
      setActiveSessionId(null)
      refreshProjects()
    }
  }, [activeServer, refreshProjects])

  const value = useMemo(() => ({
    projects, activeProject, activeSessionId,
    setActiveProject, setActiveSessionId, refreshProjects,
  }), [projects, activeProject, activeSessionId, refreshProjects])

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
