'use client'

import React, { useMemo } from 'react'
import { useTeam, TeamTask } from '@/contexts/TeamContext'
import { useLocale } from '@/contexts/LocaleContext'
import { cn } from '@/lib/utils'
import { CheckCircle2, Circle, Loader2, XCircle, AlertCircle } from 'lucide-react'

function StatusIcon({ status }: { status: TeamTask['status'] }) {
  switch (status) {
    case 'done':        return <CheckCircle2 className="w-3 h-3 text-green-500" />
    case 'in_progress': return <Loader2     className="w-3 h-3 text-blue-500 animate-spin" />
    case 'failed':      return <XCircle     className="w-3 h-3 text-red-500" />
    case 'blocked':     return <AlertCircle className="w-3 h-3 text-orange-500" />
    default:            return <Circle      className="w-3 h-3 text-slate-400" />
  }
}

function TaskCard({ task }: { task: TeamTask }) {
  return (
    <div className="bg-card rounded border border-border p-2 mb-2 text-xs shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-center gap-1 mb-1">
        <StatusIcon status={task.status} />
        <span className="font-mono text-[11px] text-muted-foreground">{task.id.slice(0, 7)}</span>
      </div>
      <div className="text-xs font-medium text-foreground leading-tight line-clamp-2">
        {task.title}
      </div>
      {task.owner && !task.agent_id && (
        <div className="mt-1 text-[11px] text-muted-foreground">
          owner: <span className="text-purple-500">{task.owner}</span>
        </div>
      )}
      {task.agent_id && (
        <div className="mt-1 text-[11px] text-muted-foreground">
          agent: <span className="text-blue-500">{task.agent_id}</span>
        </div>
      )}
      {task.priority > 0 && (
        <div className="mt-0.5 text-[11px] text-amber-500">
          ★ {task.priority}
        </div>
      )}
    </div>
  )
}

export default function TeamBoard() {
  const { tasks, activeTeam } = useTeam()
  const { t } = useLocale()

  const STATUS_COLS: Array<{ key: TeamTask['status']; label: string; color: string }> = [
    { key: 'pending',     label: t.team.statusPending,    color: 'border-slate-400' },
    { key: 'in_progress', label: t.team.statusInProgress, color: 'border-blue-500'  },
    { key: 'blocked',     label: t.team.statusBlocked,    color: 'border-orange-500' },
    { key: 'done',        label: t.team.statusDone,       color: 'border-green-500' },
    { key: 'failed',      label: t.team.statusFailed,     color: 'border-red-500'   },
  ]

  const columns = useMemo(() => {
    const map: Record<string, TeamTask[]> = {}
    for (const col of STATUS_COLS) map[col.key] = []
    for (const t of tasks) {
      if (map[t.status]) map[t.status].push(t)
    }
    return map
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tasks, t])

  if (!activeTeam) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        {t.team.selectTeam}
      </div>
    )
  }

  return (
    <div className="flex gap-3 h-full overflow-x-auto p-3">
      {STATUS_COLS.map((col) => (
        <div
          key={col.key}
          className={cn('flex-1 min-w-0 flex flex-col rounded-lg border-t-2 bg-muted/50', col.color)}
        >
          <div className="p-2 font-semibold text-xs text-muted-foreground flex items-center justify-between">
            <span>{col.label}</span>
            <span className="bg-muted rounded-full px-1.5 text-[10px]">
              {columns[col.key]?.length ?? 0}
            </span>
          </div>
          <div className="flex-1 overflow-y-auto p-2">
            {(columns[col.key] || []).map((t) => (
              <TaskCard key={t.id} task={t} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
