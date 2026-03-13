'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { History, ChevronDown, Plus, FolderOpen, Clock } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { SessionInfo } from '@/app/api/sessions/route'

interface Props {
  sessionId: string
  cwd: string
  onChange: (sessionId: string, cwd: string) => void
}

function relativeTime(ms: number): string {
  const diff = Date.now() - ms
  const m = Math.floor(diff / 60000)
  if (m < 1) return '방금 전'
  if (m < 60) return `${m}분 전`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}시간 전`
  const d = Math.floor(h / 24)
  return `${d}일 전`
}

export default function SessionPicker({ sessionId, cwd, onChange }: Props) {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const panelRef = useRef<HTMLDivElement>(null)
  const prevCwdRef = useRef('')

  // cwd가 바뀔 때마다 세션 목록 재로드
  const loadSessions = useCallback((targetCwd: string) => {
    setLoading(true)
    const url = targetCwd
      ? `/api/sessions?cwd=${encodeURIComponent(targetCwd)}`
      : '/api/sessions'
    fetch(url)
      .then(r => r.json())
      .then((data: SessionInfo[]) => {
        setSessions(data)
        // 기본값: 가장 최신 세션 자동 선택 (최초 1회)
        if (!sessionId && data.length > 0 && !prevCwdRef.current) {
          onChange(data[0].sessionId, data[0].cwd)
        }
      })
      .catch(() => setSessions([]))
      .finally(() => setLoading(false))
  }, [sessionId, onChange])

  useEffect(() => {
    // cwd가 실제로 바뀔 때만 재로드 (빈 문자열 입력 중 타이핑 방지용 디바운스)
    const timer = setTimeout(() => {
      if (cwd !== prevCwdRef.current) {
        prevCwdRef.current = cwd
        loadSessions(cwd)
      }
    }, 400)
    return () => clearTimeout(timer)
  }, [cwd, loadSessions])

  // 최초 로드
  useEffect(() => {
    loadSessions(cwd)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 외부 클릭 닫기
  const handleOutside = useCallback((e: MouseEvent) => {
    if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
      setOpen(false)
    }
  }, [])

  useEffect(() => {
    if (open) document.addEventListener('mousedown', handleOutside)
    else document.removeEventListener('mousedown', handleOutside)
    return () => document.removeEventListener('mousedown', handleOutside)
  }, [open, handleOutside])

  const current = sessions.find(s => s.sessionId === sessionId)

  return (
    <div ref={panelRef} className="relative">
      {/* 트리거 버튼 */}
      <button
        onClick={() => setOpen(v => !v)}
        className={cn(
          'flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs transition-colors',
          'border border-zinc-700 bg-zinc-800 hover:bg-zinc-700 hover:border-zinc-600',
          open && 'border-blue-500/50 bg-zinc-700',
        )}
      >
        <History className="w-3 h-3 text-zinc-400 shrink-0" />
        <span className="max-w-[180px] truncate text-zinc-300 font-mono">
          {loading ? '로딩 중...'
            : !sessionId ? '새 대화'
            : current
            ? current.firstMessage.slice(0, 24) + (current.firstMessage.length > 24 ? '…' : '')
            : sessionId.slice(0, 8) + '…'
          }
        </span>
        <ChevronDown className={cn('w-3 h-3 text-zinc-500 shrink-0 transition-transform', open && 'rotate-180')} />
      </button>

      {/* 드롭다운 패널 */}
      {open && (
        <div className="absolute right-0 top-full mt-1.5 w-80 max-h-96 overflow-y-auto z-50 rounded-xl border border-zinc-700 bg-zinc-900 shadow-xl shadow-black/40">
          {/* 새 대화 */}
          <button
            onClick={() => { onChange('', ''); setOpen(false) }}
            className={cn(
              'w-full flex items-center gap-2.5 px-3 py-2.5 text-sm hover:bg-zinc-800 transition-colors',
              'border-b border-zinc-800',
              !sessionId && 'bg-zinc-800 text-blue-400',
            )}
          >
            <div className="w-7 h-7 rounded-full bg-blue-600/20 flex items-center justify-center shrink-0">
              <Plus className="w-3.5 h-3.5 text-blue-400" />
            </div>
            <div className="text-left">
              <p className="font-medium text-zinc-200">새 대화</p>
              <p className="text-xs text-zinc-500">새로운 세션 시작</p>
            </div>
          </button>

          {/* 세션 목록 */}
          {sessions.length === 0 && !loading && (
            <p className="text-center text-zinc-500 text-xs py-6">세션이 없어요</p>
          )}
          {sessions.map(s => (
            <button
              key={s.sessionId}
              onClick={() => { onChange(s.sessionId, s.cwd); setOpen(false) }}
              className={cn(
                'w-full flex items-start gap-2.5 px-3 py-2.5 text-sm hover:bg-zinc-800 transition-colors',
                'border-b border-zinc-800/60 last:border-0',
                s.sessionId === sessionId && 'bg-zinc-800/70',
              )}
            >
              {/* 선택 인디케이터 */}
              <div className={cn(
                'w-1.5 h-1.5 rounded-full mt-1.5 shrink-0',
                s.sessionId === sessionId ? 'bg-blue-400' : 'bg-zinc-700',
              )} />

              <div className="flex-1 min-w-0 text-left">
                {/* 첫 메시지 */}
                <p className="text-zinc-200 text-xs font-medium truncate leading-relaxed">
                  {s.firstMessage}
                </p>
                {/* cwd + 시간 */}
                <div className="flex items-center gap-2 mt-0.5">
                  <span className="flex items-center gap-1 text-[10px] text-zinc-500 font-mono truncate max-w-[150px]">
                    <FolderOpen className="w-2.5 h-2.5 shrink-0" />
                    {s.cwd.replace(process.env.HOME ?? '', '~')}
                  </span>
                  <span className="flex items-center gap-1 text-[10px] text-zinc-600 shrink-0">
                    <Clock className="w-2.5 h-2.5" />
                    {relativeTime(s.lastActive)}
                  </span>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
