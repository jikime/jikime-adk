'use client'

import { createContext, useContext, useState, useEffect, useCallback, useRef, useMemo, ReactNode } from 'react'
import { useServer } from '@/contexts/ServerContext'
import { useProject } from '@/contexts/ProjectContext'

// ── Types ──────────────────────────────────────────────────────────

export interface TeamTask {
  id:          string
  title:       string
  description: string
  status:      'pending' | 'in_progress' | 'done' | 'failed' | 'blocked'
  owner?:      string   // pre-assigned agent ID (set at creation; separate from agent_id)
  agent_id:    string   // agent that claimed the task (set on claim)
  priority:    number
  tags:        string[]
  depends_on:  string[]
  created_at:  string
  updated_at:  string
}

export interface TeamAgent {
  id:              string
  team_name:       string
  role:            string
  status:          'active' | 'idle' | 'offline' | 'dead'
  pid:             number
  tmux_session:    string
  current_task_id: string
  last_heartbeat:  string
}

export interface TeamInfo {
  name:        string
  config:      Record<string, unknown>
  taskCounts:  Record<string, number>
  agentCount:  number
  costs:       { total: number; agents: Record<string, { tokens: number }> }
}

export interface TeamSummary {
  name:   string
  config: Record<string, unknown>
}

export interface TeamMessage {
  from:      string
  to:        string
  type:      string
  timestamp: string
  content:   string
}

export interface TeamMember {
  name:       string
  agentType:  string
  inboxCount: number
}

export interface TeamStat {
  pending:     number
  in_progress: number
  done:        number
  failed:      number
  blocked:     number
}

export interface TeamBrief {
  name:        string
  leaderName:  string
  description: string
}

// ── Context ────────────────────────────────────────────────────────

interface TeamContextType {
  teams:          TeamSummary[]
  activeTeam:     string | null
  teamInfo:       TeamInfo | null
  tasks:          TeamTask[]
  agents:         TeamAgent[]
  members:        TeamMember[]
  messages:       TeamMessage[]
  taskSummary:    TeamStat
  teamBrief:      TeamBrief | null
  connected:      boolean
  lastEvent:      string | null
  setActiveTeam:  (name: string | null) => void
  refreshTeams:   () => Promise<void>
  refreshTeam:    () => Promise<void>
  createTask:     (teamName: string, title: string, desc?: string) => Promise<void>
  updateTask:     (teamName: string, taskId: string, patch: Partial<{ status: string; agent_id: string; result: string }>) => Promise<void>
  sendMessage:    (teamName: string, to: string, body: string) => Promise<void>
}

const EMPTY_STAT: TeamStat = { pending: 0, in_progress: 0, done: 0, failed: 0, blocked: 0 }

const TeamContext = createContext<TeamContextType | null>(null)

export function TeamProvider({ children }: { children: ReactNode }) {
  const { getApiUrl }   = useServer()
  const { activeProject } = useProject()

  const [teams, setTeams]             = useState<TeamSummary[]>([])
  const [activeTeam, setActiveTeam]   = useState<string | null>(null)
  const [teamInfo, setTeamInfo]       = useState<TeamInfo | null>(null)
  const [tasks, setTasks]             = useState<TeamTask[]>([])
  const [agents, setAgents]           = useState<TeamAgent[]>([])
  const [members, setMembers]         = useState<TeamMember[]>([])
  const [messages, setMessages]       = useState<TeamMessage[]>([]  )
  const [taskSummary, setTaskSummary] = useState<TeamStat>(EMPTY_STAT)
  const [teamBrief, setTeamBrief]     = useState<TeamBrief | null>(null)
  const [connected, setConnected]     = useState(false)
  const [lastEvent, setLastEvent]     = useState<string | null>(null)
  const sseRef = useRef<EventSource | null>(null)

  const refreshTeams = useCallback(async () => {
    try {
      const projectPath = activeProject?.path
      const qs  = projectPath ? `?projectPath=${encodeURIComponent(projectPath)}` : ''
      const res = await fetch(getApiUrl(`/api/team/list${qs}`))
      if (res.ok) {
        const data = await res.json() as { teams: TeamSummary[] }
        setTeams(data.teams || [])
      }
    } catch { /* */ }
  }, [getApiUrl, activeProject?.path])

  const refreshTeam = useCallback(async () => {
    if (!activeTeam) return
    try {
      const [infoRes, tasksRes, agentsRes] = await Promise.all([
        fetch(getApiUrl(`/api/team/${activeTeam}`)),
        fetch(getApiUrl(`/api/team/${activeTeam}/tasks`)),
        fetch(getApiUrl(`/api/team/${activeTeam}/agents`)),
      ])
      if (infoRes.ok)   setTeamInfo(await infoRes.json() as TeamInfo)
      if (tasksRes.ok)  setTasks((await tasksRes.json() as { tasks: TeamTask[] }).tasks || [])
      if (agentsRes.ok) setAgents((await agentsRes.json() as { agents: TeamAgent[] }).agents || [])
    } catch { /* */ }
  }, [getApiUrl, activeTeam])

  // activeProject 바뀌면 팀 목록 갱신 + 선택 초기화
  useEffect(() => {
    setActiveTeam(null)
    setTeams([])
    refreshTeams()
  }, [activeProject?.path]) // eslint-disable-line react-hooks/exhaustive-deps

  // Subscribe to SSE events when activeTeam changes
  useEffect(() => {
    if (sseRef.current) {
      sseRef.current.close()
      sseRef.current = null
    }
    setConnected(false)
    if (!activeTeam) return

    refreshTeam()

    const url = getApiUrl(`/api/team/${activeTeam}/events`)
    const sse = new EventSource(url)
    sseRef.current = sse

    sse.onmessage = (e: MessageEvent) => {
      try {
        const payload = JSON.parse(e.data as string) as {
          type?:        string
          tasks?:       TeamTask[]
          agents?:      TeamAgent[]
          members?:     TeamMember[]
          messages?:    TeamMessage[]
          taskSummary?: TeamStat
          team?:        TeamBrief
          time?:        string
        }
        if (payload.tasks)       setTasks(payload.tasks)
        if (payload.agents)      setAgents(payload.agents)
        if (payload.members)     setMembers(payload.members)
        if (payload.messages)    setMessages(payload.messages)
        if (payload.taskSummary) setTaskSummary(payload.taskSummary)
        if (payload.team)        setTeamBrief(payload.team)
        if (payload.time)        setLastEvent(payload.time)
        setConnected(true)
      } catch { /* */ }
    }
    sse.onerror = () => {
      setConnected(false)
      sse.close()
    }

    return () => {
      sse.close()
      sseRef.current = null
    }
  }, [activeTeam, getApiUrl, refreshTeam])

  // Load teams on mount
  useEffect(() => { refreshTeams() }, [refreshTeams])

  const createTask = useCallback(async (teamName: string, title: string, desc?: string) => {
    await fetch(getApiUrl(`/api/team/${teamName}/tasks`), {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({ title, desc }),
    })
    await refreshTeam()
  }, [getApiUrl, refreshTeam])

  const updateTask = useCallback(async (
    teamName: string,
    taskId: string,
    patch: Partial<{ status: string; agent_id: string; result: string }>,
  ) => {
    await fetch(getApiUrl(`/api/team/${teamName}/tasks/${taskId}`), {
      method:  'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify(patch),
    })
  }, [getApiUrl])

  const sendMessage = useCallback(async (teamName: string, to: string, body: string) => {
    await fetch(getApiUrl(`/api/team/${teamName}/inbox/send`), {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({ to, body }),
    })
  }, [getApiUrl])

  // useMemo prevents a new object reference on every render,
  // which would otherwise cause all context consumers to re-render unnecessarily.
  const contextValue = useMemo(() => ({
    teams, activeTeam, teamInfo, tasks, agents,
    members, messages, taskSummary, teamBrief, connected,
    lastEvent,
    setActiveTeam, refreshTeams, refreshTeam,
    createTask, updateTask, sendMessage,
  }), [
    teams, activeTeam, teamInfo, tasks, agents,
    members, messages, taskSummary, teamBrief, connected,
    lastEvent,
    setActiveTeam, refreshTeams, refreshTeam,
    createTask, updateTask, sendMessage,
  ])

  return (
    <TeamContext.Provider value={contextValue}>
      {children}
    </TeamContext.Provider>
  )
}

export function useTeam(): TeamContextType {
  const ctx = useContext(TeamContext)
  if (!ctx) throw new Error('useTeam must be used within TeamProvider')
  return ctx
}
