'use client'

import React from 'react'
import { useTeam, TeamAgent } from '@/contexts/TeamContext'
import { cn } from '@/lib/utils'

function agentStatusColor(status: TeamAgent['status']): string {
  switch (status) {
    case 'active':  return 'bg-green-500'
    case 'idle':    return 'bg-yellow-500'
    case 'offline': return 'bg-slate-400'
    case 'dead':    return 'bg-red-500'
    default:        return 'bg-slate-300'
  }
}

function AgentBadge({ agent }: { agent: TeamAgent }) {
  return (
    <div className="flex items-center gap-1.5 px-2 py-1 bg-slate-100 dark:bg-slate-800 rounded border border-slate-200 dark:border-slate-700 text-base">
      <span className={cn('w-2 h-2 rounded-full flex-shrink-0', agentStatusColor(agent.status))} />
      <span className="font-medium text-slate-700 dark:text-slate-300 max-w-[80px] truncate">
        {agent.id}
      </span>
      <span className="text-slate-400 dark:text-slate-500">
        {agent.role}
      </span>
      {agent.current_task_id && (
        <span className="text-blue-500 font-mono">
          ↳{agent.current_task_id.slice(0, 6)}
        </span>
      )}
    </div>
  )
}

export default function AgentStatusBar() {
  const { agents, activeTeam } = useTeam()

  if (!activeTeam) return null

  const counts = {
    active:  agents.filter((a) => a.status === 'active').length,
    idle:    agents.filter((a) => a.status === 'idle').length,
    offline: agents.filter((a) => a.status === 'offline').length,
  }

  return (
    <div className="flex items-center gap-2 flex-wrap px-3 py-1.5 bg-slate-50 dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 text-base">
      <span className="font-semibold text-slate-500 dark:text-slate-400 flex-shrink-0">
        에이전트 {agents.length}
      </span>
      <span className="text-green-600">●{counts.active}활성</span>
      <span className="text-yellow-600">●{counts.idle}대기</span>
      <span className="text-slate-400">●{counts.offline}오프라인</span>
      <div className="flex gap-1 flex-wrap ml-2">
        {agents.map((a) => <AgentBadge key={a.id} agent={a} />)}
      </div>
    </div>
  )
}
