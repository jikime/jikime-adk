'use client'

import { useState, useMemo } from 'react'
import { useTeam, TeamMessage, TeamMember, TeamStat } from '@/contexts/TeamContext'
import { useProject } from '@/contexts/ProjectContext'
import { useLocale } from '@/contexts/LocaleContext'
import TeamBoard       from './TeamBoard'
import TeamCreateModal from './TeamCreateModal'
import TaskAddModal    from './TaskAddModal'
import TeamServeModal  from './TeamServeModal'
import { Button }      from '@/components/ui/button'
import { Badge }       from '@/components/ui/badge'
import { ScrollArea }  from '@/components/ui/scroll-area'
import { Separator }   from '@/components/ui/separator'
import {
  Tooltip, TooltipTrigger, TooltipContent, TooltipProvider,
} from '@/components/ui/tooltip'
import {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent,
  DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import {
  Plus, RefreshCw, Play, FolderOpen,
  ChevronDown, Users, MessageSquare, Kanban, Inbox,
  CheckCircle2, Clock, XCircle, Ban, Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'

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
    teamBrief, members, taskSummary, messages, connected,
  } = useTeam()
  const { activeProject } = useProject()
  const { t } = useLocale()
  const [showCreateTeam, setShowCreateTeam] = useState(false)
  const [showAddTask,    setShowAddTask]    = useState(false)
  const [showRunTeam,    setShowRunTeam]    = useState(false)

  // teamBrief 사용 — teams 배열 전체가 새 참조로 바뀔 때마다 재연산하던 문제 해결
  // teamBrief 는 SSE 로 현재 팀 정보만 업데이트되므로 불필요한 재렌더링 없음
  const currentTeamDesc = teamBrief?.description ?? null

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

          <span
            className="inline-flex items-center gap-1 text-[11px] font-mono bg-blue-50 dark:bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-200 dark:border-blue-500/30 rounded px-1.5 py-0.5 truncate max-w-[280px]"
            title={activeProject.path}
          >
            <FolderOpen className="w-3 h-3 shrink-0" />
            <span className="truncate">{activeProject.path}</span>
          </span>

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
                  ) : teams.map((team) => {
                    const desc = (team.config as { description?: string }).description
                    return (
                      <DropdownMenuItem
                        key={team.name}
                        onClick={() => setActiveTeam(team.name)}
                        className={cn(
                          'flex flex-col items-start gap-0.5 cursor-pointer',
                          activeTeam === team.name && 'bg-accent',
                        )}
                      >
                        <span className="text-[13px] font-medium">{team.name}</span>
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

        ) : (

          /* ── Dashboard ── */
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
