'use client'

import { useEffect, useRef } from 'react'
import { useTeam } from '@/contexts/TeamContext'

interface LogEntry {
  time:    string
  message: string
}

export default function TeamEventLog() {
  const { lastEvent, activeTeam, tasks, agents } = useTeam()
  const logsRef  = useRef<LogEntry[]>([])
  const scrollRef = useRef<HTMLDivElement>(null)

  // Append a log entry whenever lastEvent changes
  useEffect(() => {
    if (!lastEvent) return
    const doneTasks = tasks.filter((t) => t.status === 'done').length
    const wip       = tasks.filter((t) => t.status === 'in_progress').length
    const active    = agents.filter((a) => a.status === 'active').length
    logsRef.current = [
      ...logsRef.current.slice(-99),
      {
        time: new Date(lastEvent).toLocaleTimeString('ko-KR'),
        message: `tasks done:${doneTasks} wip:${wip} | agents active:${active}`,
      },
    ]
    // Force re-render by a tiny workaround: we track via a stable ref and manual scroll
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [lastEvent, tasks, agents])

  if (!activeTeam) return null

  return (
    <div
      ref={scrollRef}
      className="h-28 overflow-y-auto bg-slate-900 text-green-400 font-mono text-[11px] p-2 border-t border-slate-700"
    >
      {logsRef.current.length === 0 ? (
        <div className="text-slate-500">이벤트 대기 중…</div>
      ) : (
        logsRef.current.map((entry, i) => (
          <div key={i} className="whitespace-nowrap">
            <span className="text-slate-500">[{entry.time}]</span>{' '}
            {entry.message}
          </div>
        ))
      )}
    </div>
  )
}
