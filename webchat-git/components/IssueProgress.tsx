'use client'

import { Loader2, CheckCircle2, AlertCircle, RefreshCw, ExternalLink } from 'lucide-react'
import { cn } from '@/lib/utils'

interface TokenInfo {
  InputTokens: number
  OutputTokens: number
  TotalTokens: number
}

interface RunningEntry {
  IssueID: string
  IssueIdentifier: string
  State: string
  TurnCount: number
  LastEvent: string
  LastMessage: string
  StartedAt: string
  LastEventAt: string | null
  Tokens: TokenInfo
}

interface RetryEntry {
  IssueID: string
  Identifier: string
  Attempt: number
  DueAt: string | null
  Error: string
}

interface ServeState {
  generated_at: string
  counts: { running: number; retrying: number }
  running: RunningEntry[]
  retrying: RetryEntry[]
  jikime_totals: TokenInfo & { SecondsRunning: number }
}

interface Props {
  state: ServeState | null
  connected: boolean
  onRefresh: () => void
}

function relTime(iso: string | undefined): string {
  if (!iso) return '방금 전'
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '방금 전'
  const diff = Date.now() - d.getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}초 전`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}분 전`
  return `${Math.floor(m / 60)}시간 전`
}

export default function IssueProgress({ state, connected, onRefresh }: Props) {
  if (!connected) {
    return (
      <div className="flex items-center gap-2 text-zinc-500 text-xs px-1">
        <Loader2 className="w-3 h-3 animate-spin" />
        jikime serve 연결 중...
      </div>
    )
  }

  if (!state || (state.counts.running === 0 && state.counts.retrying === 0)) {
    return (
      <div className="flex items-center justify-between px-1">
        <span className="text-xs text-zinc-600">처리 중인 이슈 없음</span>
        <button onClick={onRefresh} className="text-zinc-600 hover:text-zinc-400 transition-colors">
          <RefreshCw className="w-3 h-3" />
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between px-1">
        <span className="text-[11px] text-zinc-500 font-mono">
          실행 {state.counts.running} · 재시도 {state.counts.retrying} ·
          총 {(state.jikime_totals?.TotalTokens ?? 0).toLocaleString()} tokens
        </span>
        <button onClick={onRefresh} className="text-zinc-600 hover:text-zinc-400 transition-colors">
          <RefreshCw className="w-3 h-3" />
        </button>
      </div>

      {/* 실행 중 */}
      {state.running.map((entry, i) => (
        <div key={i} className="bg-zinc-800/60 border border-zinc-700/50 rounded-xl px-3 py-2.5 space-y-1.5">
          <div className="flex items-center gap-2">
            <Loader2 className="w-3.5 h-3.5 text-emerald-400 animate-spin shrink-0" />
            <a
              href={`https://github.com/${entry.IssueIdentifier.replace('#', '/issues/')}`}
              target="_blank" rel="noopener noreferrer"
              className="text-xs font-mono text-blue-400 hover:underline flex items-center gap-1"
            >
              {entry.IssueIdentifier}
              <ExternalLink className="w-2.5 h-2.5" />
            </a>
            <span className="ml-auto text-[10px] text-zinc-600">턴 {entry.TurnCount}</span>
          </div>
          {entry.LastMessage && (
            <p className="text-[11px] text-zinc-400 leading-relaxed pl-5 line-clamp-2">
              {entry.LastMessage}
            </p>
          )}
          <div className="flex items-center gap-3 pl-5 text-[10px] text-zinc-600">
            <span>{(entry.Tokens?.TotalTokens ?? 0).toLocaleString()} tokens</span>
            <span>{relTime(entry.LastEventAt ?? entry.StartedAt)}</span>
          </div>
        </div>
      ))}

      {/* 재시도 대기 */}
      {state.retrying.map((entry, i) => (
        <div key={i} className="bg-amber-950/30 border border-amber-800/40 rounded-xl px-3 py-2.5 space-y-1">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-3.5 h-3.5 text-amber-400 shrink-0" />
            <span className="text-xs font-mono text-amber-400">{entry.Identifier}</span>
            <span className="ml-auto text-[10px] text-zinc-600">시도 {entry.Attempt}</span>
          </div>
          <p className="text-[11px] text-red-400 pl-5 truncate">{entry.Error}</p>
          <p className="text-[10px] text-zinc-600 pl-5">
            {entry.DueAt
              ? (() => { const d = new Date(entry.DueAt); return isNaN(d.getTime()) ? '잠시 후' : d.toLocaleTimeString() })()
              : '잠시 후'
            } 재시도
          </p>
        </div>
      ))}
    </div>
  )
}

export function IssueDoneTag({ identifier }: { identifier: string }) {
  return (
    <span className={cn('inline-flex items-center gap-1 text-[11px] text-emerald-400')}>
      <CheckCircle2 className="w-3 h-3" />
      {identifier} 완료
    </span>
  )
}
