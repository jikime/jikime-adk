'use client'

import { useState, useMemo } from 'react'
import dynamic from 'next/dynamic'
import { useTeam, TeamMessage, TeamMember, TeamStat } from '@/contexts/TeamContext'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import { useLocale } from '@/contexts/LocaleContext'
import TeamBoard       from './TeamBoard'
import TeamCreateModal from './TeamCreateModal'
import TaskAddModal    from './TaskAddModal'
import TeamServeModal  from './TeamServeModal'
import { Button }      from '@/components/ui/button'
import { Badge }       from '@/components/ui/badge'
import { ScrollArea }  from '@/components/ui/scroll-area'
import { Separator }   from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Tooltip, TooltipTrigger, TooltipContent, TooltipProvider,
} from '@/components/ui/tooltip'
import {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent,
  DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import {
  Plus, RefreshCw, Play, FolderOpen, LayoutDashboard, MonitorPlay,
  ChevronDown, Users, MessageSquare, Kanban, Inbox,
  CheckCircle2, Clock, XCircle, Ban, Loader2, X,
} from 'lucide-react'
import { cn } from '@/lib/utils'

const AgentTerminalPanel = dynamic(() => import('./AgentTerminalPanel'), { ssr: false })

type BoardTab = 'board' | 'agents'

// ── Summary cards ───────────────────────────────────────────────────

function SummaryCards({ summary }: { summary: TeamStat }) {
  const { t } = useLocale()
  const cols: Array<{ key: keyof TeamStat; label: string; icon: React.ReactNode; badge: string }> = [
    { key: 'pending',     label: t.team.statusPending,    icon: <Clock        className="w-4 h-4" />, badge: 'bg-orange-500/10 text-orange-400 border-orange-500/20' },
    { key: 'in_progress', label: t.team.statusInProgress, icon: <Loader2      className="w-4 h-4 animate-spin" />, badge: 'bg-blue-500/10 text-blue-400 border-blue-500/20' },
    { key: 'done',        label: t.team.statusDone,       icon: <CheckCircle2 className="w-4 h-4" />, badge: 'bg-green-500/10 text-green-400 border-green-500/20' },
    { key: 'failed',      label: t.team.statusFailed,     icon: <XCircle      className="w-4 h-4" />, badge: 'bg-red-500/10 text-red-400 border-red-500/20' },
    { key: 'blocked',     label: t.team.statusBlocked,    icon: <Ban          className="w-4 h-4" />, badge: 'bg-purple-500/10 text-purple-400 border-purple-500/20' },
  ]
  return (
    <div className="grid grid-cols-5 gap-2 mb-4">
      {cols.map((col) => (
        <div key={col.key} className={cn('text-center rounded-lg border px-3 py-3 flex flex-col items-center gap-1', col.badge)}>
          <div className="opacity-70">{col.icon}</div>
          <div className="text-3xl font-bold leading-none tracking-tight">
            {summary[col.key]}
          </div>
          <div className="text-[10px] font-medium uppercase tracking-wider opacity-70">
            {col.label}
          </div>
        </div>
      ))}
    </div>
  )
}

// ── Members panel ───────────────────────────────────────────────────

function MembersPanel({ members }: { members: TeamMember[] }) {
  const { t } = useLocale()
  return (
    <div className="mb-4 rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <div className="flex items-center gap-2 text-[13px] font-semibold">
          <Users className="w-3.5 h-3.5 text-muted-foreground" />
          {t.team.sectionMembers}
        </div>
        <Badge variant="secondary" className="text-[11px] h-5">{members.length}</Badge>
      </div>
      <div>
        {members.length === 0 ? (
          <div className="py-8 text-center text-[13px] text-muted-foreground">
            {t.team.noMembers}
          </div>
        ) : (
          <div
            className="grid divide-x divide-y divide-border/60"
            style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))' }}
          >
            {members.map((m) => (
              <div key={m.name} className="p-4 flex flex-col gap-1.5">
                <div className="flex items-center gap-2.5">
                  <div className="w-7 h-7 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
                    <span className="text-[11px] font-bold text-primary/80">
                      {m.name.slice(0, 2).toUpperCase()}
                    </span>
                  </div>
                  <div>
                    <div className="text-[13px] font-semibold text-foreground leading-tight">{m.name}</div>
                    <div className="text-[11px] text-muted-foreground">{m.agentType}</div>
                  </div>
                </div>
                <Badge
                  variant="outline"
                  className={cn(
                    'w-fit text-[10px] h-4 px-1.5',
                    m.inboxCount > 0
                      ? 'bg-red-500/10 text-red-400 border-red-500/20'
                      : 'text-muted-foreground',
                  )}
                >
                  {m.inboxCount > 0 ? (
                    <><Inbox className="w-2.5 h-2.5" /> {t.team.inboxMsg(m.inboxCount)}</>
                  ) : (
                    t.team.inboxEmpty
                  )}
                </Badge>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// ── Messages panel ──────────────────────────────────────────────────

const MSG_TYPE_COLOR: Record<string, string> = {
  task:      'bg-blue-500/10 text-blue-400 border-blue-500/20',
  broadcast: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
  update:    'bg-green-500/10 text-green-400 border-green-500/20',
  error:     'bg-red-500/10 text-red-400 border-red-500/20',
}

function MessagesPanel({ messages }: { messages: TeamMessage[] }) {
  const { t } = useLocale()
  const sorted = useMemo(() => [...messages].reverse(), [messages])
  return (
    <div className="mb-4 rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <div className="flex items-center gap-2 text-[13px] font-semibold">
          <MessageSquare className="w-3.5 h-3.5 text-muted-foreground" />
          {t.team.sectionMessages}
        </div>
        <Badge variant="secondary" className="text-[11px] h-5">{messages.length}</Badge>
      </div>
      <ScrollArea className="h-64">
        {sorted.length === 0 ? (
          <div className="py-8 text-center text-[13px] text-muted-foreground">
            {t.team.noMessages}
          </div>
        ) : sorted.map((m, i) => {
          const ts = (m.timestamp || '').slice(11, 19)
          const dt = (m.timestamp || '').slice(5, 10)
          const typeColor = MSG_TYPE_COLOR[m.type] ?? 'bg-muted text-muted-foreground border-border'
          return (
            <div key={i} className="px-4 py-3 border-b border-border/40 last:border-0">
              <div className="flex items-center gap-1.5 mb-1.5 flex-wrap">
                <Badge variant="outline" className={cn('text-[10px] h-4 px-1.5', typeColor)}>
                  {m.type || 'msg'}
                </Badge>
                <span className="text-[12px] font-medium text-foreground">{m.from}</span>
                <span className="text-[11px] text-muted-foreground">→ {m.to || 'all'}</span>
                <span className="text-[11px] text-muted-foreground/60 ml-auto">{dt} {ts}</span>
              </div>
              <p className="text-[12px] text-muted-foreground leading-relaxed whitespace-pre-wrap break-words">
                {m.content || ''}
              </p>
            </div>
          )
        })}
      </ScrollArea>
    </div>
  )
}

// ── Kanban wrapper ──────────────────────────────────────────────────

function KanbanSection() {
  const { t } = useLocale()
  return (
    <div className="rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center gap-2 text-[13px] font-semibold">
        <Kanban className="w-3.5 h-3.5 text-muted-foreground" />
        {t.team.sectionTasks}
      </div>
      <div style={{ height: 380 }}>
        <TeamBoard />
      </div>
    </div>
  )
}

// ── BoardPanel (main export) ────────────────────────────────────────

export default function BoardPanel() {
  const {
    teams, activeTeam, setActiveTeam, refreshTeam,
    teamBrief, members, taskSummary, messages, connected, agents,
  } = useTeam()
  const { activeProject } = useProject()
  const { getApiUrl } = useServer()
  const { t } = useLocale()
  const [showCreateTeam, setShowCreateTeam] = useState(false)
  const [showAddTask,    setShowAddTask]    = useState(false)
  const [showRunTeam,    setShowRunTeam]    = useState(false)
  const [boardTab,       setBoardTab]       = useState<BoardTab>('board')
  const [activeAgent,    setActiveAgent]    = useState<string>('')
  const [killingAgent,   setKillingAgent]   = useState<string>('')

  const liveAgents = useMemo(
    () => agents.filter(a => a.tmux_session),
    [agents]
  )

  async function killAgent(agentId: string) {
    if (!activeTeam || killingAgent) return
    setKillingAgent(agentId)
    try {
      await fetch(getApiUrl(`/api/team/${activeTeam}/agents/${encodeURIComponent(agentId)}`), {
        method: 'DELETE',
      })
      await refreshTeam()
      const killed = liveAgents.find(a => a.id === agentId)
      if (killed && activeAgent === killed.tmux_session) {
        const remaining = liveAgents.filter(a => a.id !== agentId)
        setActiveAgent(remaining[0]?.tmux_session ?? '')
      }
    } finally {
      setKillingAgent('')
    }
  }

  const currentTeamDesc = useMemo(() => {
    if (!activeTeam) return null
    const t = teams.find(t => t.name === activeTeam)
    return (t?.config as { description?: string })?.description ?? null
  }, [activeTeam, teams])

  // 프로젝트 미선택
  if (!activeProject) {
    return (
      <div className="flex flex-col h-full items-center justify-center gap-4 bg-background rounded-lg border border-border text-center px-8">
        <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center">
          <FolderOpen className="w-6 h-6 text-muted-foreground/50" />
        </div>
        <div>
          <p className="text-[14px] font-semibold text-foreground mb-1">{t.team.selectProject}</p>
          <p className="text-[12px] text-muted-foreground leading-relaxed">
            {t.team.selectProjectHint}
          </p>
        </div>
      </div>
    )
  }

  return (
    <TooltipProvider>
      <div className="flex flex-col h-full overflow-hidden bg-background rounded-lg border border-border">

        {/* ── Toolbar ──────────────────────────────────────────── */}
        <div className="flex items-center gap-2 px-3 py-2 border-b border-border shrink-0 bg-background/95 backdrop-blur-sm">

          {/* Board / Agents 탭 */}
          {activeTeam && (
            <Tabs
              value={boardTab}
              onValueChange={(v) => {
                setBoardTab(v as BoardTab)
                if (v === 'agents' && !activeAgent && liveAgents.length > 0)
                  setActiveAgent(liveAgents[0].tmux_session)
              }}
            >
              <TabsList variant="default" className="h-7">
                <TabsTrigger value="board" className="text-[12px] h-6 px-2.5 gap-1.5">
                  <LayoutDashboard className="w-3 h-3" /> {t.team.tabBoard}
                </TabsTrigger>
                <TabsTrigger value="agents" className="text-[12px] h-6 px-2.5 gap-1.5">
                  <MonitorPlay className="w-3 h-3" /> {t.team.tabAgents}
                  {liveAgents.length > 0 && (
                    <Badge className="h-4 px-1 text-[10px] bg-green-500/20 text-green-400 border-green-500/30 ml-0.5">
                      {liveAgents.length}
                    </Badge>
                  )}
                </TabsTrigger>
              </TabsList>
            </Tabs>
          )}

          {/* 우측 영역 */}
          <div className="ml-auto flex items-center gap-1.5">

            {/* 팀 선택 드롭다운 */}
            <DropdownMenu>
              <DropdownMenuTrigger className="inline-flex items-center gap-1.5 h-7 px-2.5 text-[12px] rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground max-w-[180px] font-medium transition-colors">
                <span className="truncate">
                  {activeTeam ?? t.team.selectTeam}
                </span>
                <ChevronDown className="w-3 h-3 shrink-0 opacity-50" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-52">
                <DropdownMenuGroup>
                  <DropdownMenuLabel className="text-[11px] text-muted-foreground font-normal">
                    {activeProject.name}
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {teams.length === 0 ? (
                    <div className="px-2 py-3 text-center text-[12px] text-muted-foreground">
                      {t.team.noTeams}
                    </div>
                  ) : teams.map((t) => {
                    const desc = (t.config as { description?: string }).description
                    return (
                      <DropdownMenuItem
                        key={t.name}
                        onClick={() => { setActiveTeam(t.name); setBoardTab('board') }}
                        className={cn(
                          'flex flex-col items-start gap-0.5 cursor-pointer',
                          activeTeam === t.name && 'bg-accent',
                        )}
                      >
                        <span className="text-[13px] font-medium">{t.name}</span>
                        {desc && <span className="text-[11px] text-muted-foreground">{desc}</span>}
                      </DropdownMenuItem>
                    )
                  })}
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>

            <Separator orientation="vertical" className="h-4" />

            {/* 팀 선택 시 액션 버튼 */}
            {activeTeam && (
              <>
                <Tooltip>
                  <TooltipTrigger
                    className="inline-flex items-center justify-center h-7 w-7 rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                    onClick={() => setShowAddTask(true)}
                  >
                    <Plus className="w-3.5 h-3.5" />
                  </TooltipTrigger>
                  <TooltipContent side="bottom">{t.team.addTask}</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger
                    className="inline-flex items-center justify-center h-7 w-7 rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                    onClick={refreshTeam}
                  >
                    <RefreshCw className="w-3.5 h-3.5" />
                  </TooltipTrigger>
                  <TooltipContent side="bottom">{t.team.refresh}</TooltipContent>
                </Tooltip>

                <Button
                  size="sm" variant="default"
                  className="h-7 text-[12px] gap-1.5 bg-green-600 hover:bg-green-500 text-white"
                  onClick={() => setShowRunTeam(true)}
                >
                  <Play className="w-3 h-3" /> {t.team.runTeam}
                </Button>
              </>
            )}

            <Button
              size="sm" variant="outline"
              className="h-7 text-[12px]"
              onClick={() => setShowCreateTeam(true)}
            >
              <Plus className="w-3 h-3" /> {t.team.newTeam}
            </Button>

            {/* 연결 상태 */}
            <div className="flex items-center gap-1.5 pl-1">
              <span className={cn(
                'w-1.5 h-1.5 rounded-full',
                connected && activeTeam ? 'bg-green-400 shadow-[0_0_6px_#4ade80]' : 'bg-muted-foreground/30',
              )} />
              <span className="text-[11px] text-muted-foreground hidden sm:block">
                {connected && activeTeam ? t.team.live : t.team.offline}
              </span>
            </div>
          </div>
        </div>

        {/* ── Content ──────────────────────────────────────────── */}
        {!activeTeam ? (

          /* 팀 미선택 */
          <div className="flex-1 flex flex-col items-center justify-center gap-3 text-center px-8">
            <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center">
              <Users className="w-5 h-5 text-muted-foreground/50" />
            </div>
            <div>
              <p className="text-[13px] font-medium text-foreground mb-1">
                {t.team.selectTeamHint(activeProject.name)}
              </p>
              <p className="text-[12px] text-muted-foreground">
                {t.team.selectTeamDesc}
              </p>
            </div>
          </div>

        ) : boardTab === 'agents' ? (

          /* ── Agents 탭 ── */
          <div className="flex flex-col flex-1 min-h-0">
            {liveAgents.length === 0 ? (
              <div className="flex-1 flex flex-col items-center justify-center gap-3 text-center px-8">
                <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center">
                  <MonitorPlay className="w-5 h-5 text-muted-foreground/40" />
                </div>
                <div>
                  <p className="text-[13px] font-medium text-foreground mb-1">{t.team.noAgents}</p>
                  <p className="text-[12px] text-muted-foreground">{t.team.noAgentsHint}</p>
                </div>
              </div>
            ) : (
              <>
                {/* 에이전트 탭 바 */}
                <div className="flex items-center gap-1 px-3 py-1.5 border-b border-border bg-muted/30 shrink-0 overflow-x-auto">
                  {liveAgents.map((a) => (
                    <div
                      key={a.tmux_session}
                      className={cn(
                        'flex items-center gap-1 rounded-md text-[12px] font-medium whitespace-nowrap transition-all shrink-0',
                        activeAgent === a.tmux_session
                          ? 'bg-background text-foreground border border-border'
                          : 'text-muted-foreground hover:text-foreground hover:bg-background/50',
                      )}
                    >
                      <button
                        onClick={() => setActiveAgent(a.tmux_session)}
                        className="flex items-center gap-1.5 pl-3 pr-1 py-1"
                      >
                        <span className={cn(
                          'w-1.5 h-1.5 rounded-full shrink-0 transition-colors',
                          a.status === 'active'
                            ? 'bg-green-400 shadow-[0_0_4px_#4ade80]'
                            : 'bg-muted-foreground/40',
                        )} />
                        {a.id}
                        <Badge variant="outline" className="text-[10px] h-4 px-1 ml-0.5">
                          {a.role}
                        </Badge>
                      </button>
                      <Tooltip>
                        <TooltipTrigger render={<span />}>
                          <button
                            onClick={(e) => { e.stopPropagation(); killAgent(a.id) }}
                            disabled={killingAgent === a.id}
                            className="pr-1.5 py-1 text-muted-foreground/50 hover:text-red-400 transition-colors disabled:opacity-40"
                          >
                            {killingAgent === a.id
                              ? <Loader2 className="w-3 h-3 animate-spin" />
                              : <X className="w-3 h-3" />
                            }
                          </button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom" className="text-[11px]">세션 종료</TooltipContent>
                      </Tooltip>
                    </div>
                  ))}
                </div>

                {/* 터미널 */}
                <div className="flex-1 min-h-0">
                  {liveAgents.map((a) => (
                    <div
                      key={a.tmux_session}
                      className={cn('h-full', activeAgent === a.tmux_session ? 'block' : 'hidden')}
                    >
                      <AgentTerminalPanel tmuxSession={a.tmux_session} />
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>

        ) : (

          /* ── Board 탭 ── */
          <ScrollArea className="flex-1">
            <div className="px-4 py-4">

              {/* 팀 헤더 */}
              {teamBrief && (
                <div className="flex items-center gap-3 mb-4">
                  <div className="w-9 h-9 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
                    <span className="text-[13px] font-bold text-primary/80">
                      {teamBrief.name.slice(0, 2).toUpperCase()}
                    </span>
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <h2 className="text-[16px] font-bold text-foreground tracking-tight truncate">
                        {teamBrief.name}
                      </h2>
                      {members.length > 0 && (
                        <Badge variant="secondary" className="text-[10px] h-4 shrink-0">
                          {t.team.members(members.length)}
                        </Badge>
                      )}
                    </div>
                    <p className="text-[12px] text-muted-foreground truncate">
                      {t.team.ledBy(teamBrief.leaderName || '?')}
                      {currentTeamDesc ? ` — ${currentTeamDesc}` : ''}
                    </p>
                  </div>
                </div>
              )}

              <SummaryCards summary={taskSummary} />
              <MembersPanel members={members} />
              <MessagesPanel messages={messages} />
              <KanbanSection />

            </div>
          </ScrollArea>
        )}

        {/* Modals */}
        <TeamServeModal  open={showRunTeam}    onClose={() => setShowRunTeam(false)} />
        <TeamCreateModal open={showCreateTeam} onClose={() => setShowCreateTeam(false)} />
        <TaskAddModal    open={showAddTask}    onClose={() => setShowAddTask(false)} />
      </div>
    </TooltipProvider>
  )
}
