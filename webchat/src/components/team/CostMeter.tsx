'use client'

import React from 'react'
import { useTeam } from '@/contexts/TeamContext'

export default function CostMeter() {
  const { teamInfo, activeTeam } = useTeam()

  if (!activeTeam || !teamInfo) return null

  const { costs, config } = teamInfo
  const budget = (config['budget'] as number) || 0
  const total  = costs?.total || 0
  const pct    = budget > 0 ? Math.min((total / budget) * 100, 100) : 0
  const exceeded = budget > 0 && total >= budget

  const barColor = exceeded
    ? 'bg-red-500'
    : pct > 80
      ? 'bg-orange-500'
      : 'bg-blue-500'

  const agentEntries = Object.entries(costs?.agents || {})
    .sort((a, b) => b[1].tokens - a[1].tokens)
    .slice(0, 5)

  return (
    <div className="px-3 py-2 bg-slate-50 dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 text-base">
      <div className="flex items-center gap-2 mb-1">
        <span className="font-semibold text-slate-500 dark:text-slate-400">토큰 예산</span>
        <span className={exceeded ? 'text-red-500 font-bold' : 'text-slate-600 dark:text-slate-300'}>
          {total.toLocaleString()} {budget > 0 ? `/ ${budget.toLocaleString()}` : ''}
        </span>
        {exceeded && <span className="text-red-500 font-bold">⚠ 초과</span>}
      </div>
      {budget > 0 && (
        <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 mb-1">
          <div
            className={`${barColor} h-1.5 rounded-full transition-all`}
            style={{ width: `${pct}%` }}
          />
        </div>
      )}
      {agentEntries.length > 0 && (
        <div className="flex gap-2 flex-wrap">
          {agentEntries.map(([id, { tokens }]) => (
            <span key={id} className="text-slate-500 dark:text-slate-400">
              <span className="text-blue-500">{id}</span>: {tokens.toLocaleString()}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
