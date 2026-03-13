'use client'

import { ExternalLink, Loader2, CheckCircle2, AlertCircle, Clock } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface GithubIssue {
  number: number
  title: string
  state: 'open' | 'closed'
  url: string
  createdAt: string
  closedAt: string | null
  labels: string[]
}

// jikime serve state의 running/retrying 정보
export interface LiveOverlay {
  // key: issue identifier "owner/repo#N" or issueId string
  running: Map<string, { turnCount: number; tokens: number; lastMessage: string }>
  retrying: Map<string, { attempt: number; error: string }>
}

interface Props {
  projectRepo: string
  issues: GithubIssue[]
  overlay: LiveOverlay
  loading: boolean
  onRefresh: () => void
}

function issueIdentifier(repo: string, number: number) {
  return `${repo}#${number}`
}

function relTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}초 전`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}분 전`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}시간 전`
  return `${Math.floor(h / 24)}일 전`
}

export default function IssueList({ projectRepo, issues, overlay, loading, onRefresh }: Props) {
  if (loading) {
    return (
      <div className="flex items-center gap-2 text-zinc-600 text-xs py-2 px-1">
        <Loader2 className="w-3 h-3 animate-spin" />
        이슈 목록 로딩 중...
      </div>
    )
  }

  if (issues.length === 0) {
    return (
      <p className="text-xs text-zinc-700 py-2 px-1">아직 이슈가 없어요</p>
    )
  }

  return (
    <div className="space-y-1.5">
      {issues.map(issue => {
        const id = issueIdentifier(projectRepo, issue.number)
        const running = overlay.running.get(id)
        const retrying = overlay.retrying.get(id)

        // 상태 결정
        let statusIcon: React.ReactNode
        let statusColor: string
        let statusLabel: string

        if (running) {
          statusIcon = <Loader2 className="w-3 h-3 animate-spin" />
          statusColor = 'text-emerald-400'
          statusLabel = `작업 중 · 턴 ${running.turnCount} · ${running.tokens.toLocaleString()} tokens`
        } else if (retrying) {
          statusIcon = <AlertCircle className="w-3 h-3" />
          statusColor = 'text-amber-400'
          statusLabel = `재시도 #${retrying.attempt}`
        } else if (issue.state === 'closed') {
          statusIcon = <CheckCircle2 className="w-3 h-3 fill-emerald-500 text-emerald-500" />
          statusColor = 'text-emerald-600'
          statusLabel = `완료 · ${issue.closedAt ? relTime(issue.closedAt) : ''}`
        } else {
          statusIcon = <Clock className="w-3 h-3" />
          statusColor = 'text-zinc-600'
          statusLabel = '대기 중'
        }

        return (
          <div
            key={issue.number}
            className={cn(
              'rounded-xl border px-3 py-2 space-y-1',
              running
                ? 'border-emerald-800/40 bg-emerald-950/20'
                : retrying
                ? 'border-amber-800/40 bg-amber-950/10'
                : issue.state === 'closed'
                ? 'border-zinc-800/40 bg-zinc-900/30'
                : 'border-zinc-700/40 bg-zinc-800/30'
            )}
          >
            {/* 헤더 행 */}
            <div className="flex items-center gap-2">
              <span className={cn('shrink-0', statusColor)}>{statusIcon}</span>
              <a
                href={issue.url}
                target="_blank" rel="noopener noreferrer"
                className="text-[11px] font-mono text-blue-400 hover:underline flex items-center gap-0.5 shrink-0"
              >
                #{issue.number} <ExternalLink className="w-2 h-2" />
              </a>
              <p className="text-xs text-zinc-300 truncate flex-1">{issue.title}</p>
            </div>

            {/* 상태 행 */}
            <div className={cn('text-[10px] pl-5', statusColor)}>
              {statusLabel}
              {running?.lastMessage && (
                <p className="text-zinc-500 mt-0.5 line-clamp-1">{running.lastMessage}</p>
              )}
              {retrying?.error && (
                <p className="text-red-500/70 mt-0.5 line-clamp-1">{retrying.error}</p>
              )}
            </div>
          </div>
        )
      })}

      <button
        onClick={onRefresh}
        className="text-[10px] text-zinc-700 hover:text-zinc-500 transition-colors w-full text-right pt-0.5"
      >
        새로고침
      </button>
    </div>
  )
}
