'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { History, ChevronDown, Plus, FolderOpen, Clock, Database } from 'lucide-react'
import { cn } from '@/lib/utils'
import { listSessions, clearMessages } from '@/lib/db'
import type { SessionInfo } from '@/app/api/sessions/route'
import type { SessionMeta } from '@/lib/db'

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
  const [claudeSessions, setClaudeSessions] = useState<SessionInfo[]>([])
  const [dbSessions, setDbSessions] = useState<SessionMeta[]>([])
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<'saved' | 'claude'>('saved')
  const panelRef = useRef<HTMLDivElement>(null)
  const prevCwdRef = useRef('')

  // IndexedDB 저장 세션 로드
  const loadDbSessions = useCallback(() => {
    listSessions().then(setDbSessions).catch(() => setDbSessions([]))
  }, [])

  // Claude 서버 세션 로드
  const loadClaudeSessions = useCallback((targetCwd: string) => {
    setLoading(true)
    const url = targetCwd
      ? `/api/sessions?cwd=${encodeURIComponent(targetCwd)}`
      : '/api/sessions'
    fetch(url)
      .then(r => r.json())
      .then((data: SessionInfo[]) => {
        setClaudeSessions(data)
        // 저장된 채팅도 없고 세션도 없을 때 최신 claude 세션 자동 선택
        if (!sessionId && data.length > 0 && !prevCwdRef.current) {
          onChange(data[0].sessionId, data[0].cwd)
        }
      })
      .catch(() => setClaudeSessions([]))
      .finally(() => setLoading(false))
  }, [sessionId, onChange])

  // cwd 변경 시 재로드
  useEffect(() => {
    const timer = setTimeout(() => {
      if (cwd !== prevCwdRef.current) {
        prevCwdRef.current = cwd
        loadClaudeSessions(cwd)
      }
    }, 400)
    return () => clearTimeout(timer)
  }, [cwd, loadClaudeSessions])

  // 최초 로드
  useEffect(() => {
    loadClaudeSessions(cwd)
    loadDbSessions()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 드롭다운 열릴 때마다 IndexedDB 세션 갱신
  useEffect(() => {
    if (open) loadDbSessions()
  }, [open, loadDbSessions])

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

  // 현재 세션 표시명
  const savedCurrent = dbSessions.find(s => s.sessionId === sessionId)
  const claudeCurrent = claudeSessions.find(s => s.sessionId === sessionId)
  const displayName = savedCurrent?.firstMessage
    ?? claudeCurrent?.firstMessage?.slice(0, 24)
    ?? (sessionId ? sessionId.slice(0, 8) + '…' : null)

  const handleDelete = useCallback((e: React.MouseEvent, sid: string) => {
    e.stopPropagation()
    clearMessages(sid).catch(() => {})
    setDbSessions(prev => prev.filter(s => s.sessionId !== sid))
    if (sid === sessionId) onChange('', '')
  }, [sessionId, onChange])

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
          {loading ? '로딩 중...' : !sessionId ? '새 대화' : (displayName ?? '세션')}
        </span>
        <ChevronDown className={cn('w-3 h-3 text-zinc-500 shrink-0 transition-transform', open && 'rotate-180')} />
      </button>

      {/* 드롭다운 패널 */}
      {open && (
        <div className="absolute right-0 top-full mt-1.5 w-80 z-50 rounded-xl border border-zinc-700 bg-zinc-900 shadow-xl shadow-black/40 flex flex-col max-h-96">

          {/* 새 대화 */}
          <button
            onClick={() => { onChange('', ''); setOpen(false) }}
            className={cn(
              'w-full flex items-center gap-2.5 px-3 py-2.5 text-sm hover:bg-zinc-800 transition-colors shrink-0',
              'border-b border-zinc-800 rounded-t-xl',
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

          {/* 탭 */}
          <div className="flex border-b border-zinc-800 shrink-0">
            <button
              onClick={() => setTab('saved')}
              className={cn(
                'flex-1 flex items-center justify-center gap-1.5 py-2 text-xs font-medium transition-colors',
                tab === 'saved' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-zinc-500 hover:text-zinc-300'
              )}
            >
              <Database className="w-3 h-3" />
              저장된 대화 {dbSessions.length > 0 && `(${dbSessions.length})`}
            </button>
            <button
              onClick={() => setTab('claude')}
              className={cn(
                'flex-1 flex items-center justify-center gap-1.5 py-2 text-xs font-medium transition-colors',
                tab === 'claude' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-zinc-500 hover:text-zinc-300'
              )}
            >
              <History className="w-3 h-3" />
              Claude 세션 {claudeSessions.length > 0 && `(${claudeSessions.length})`}
            </button>
          </div>

          {/* 세션 목록 */}
          <div className="overflow-y-auto flex-1 [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]">

            {/* 탭: 저장된 대화 (IndexedDB) */}
            {tab === 'saved' && (
              <>
                {dbSessions.length === 0 ? (
                  <p className="text-center text-zinc-500 text-xs py-6">저장된 대화가 없어요</p>
                ) : dbSessions.map(s => (
                  <div
                    key={s.sessionId}
                    className={cn(
                      'group flex items-start gap-2.5 px-3 py-2.5 border-b border-zinc-800/60 last:border-0',
                      'hover:bg-zinc-800 transition-colors cursor-pointer',
                      s.sessionId === sessionId && 'bg-zinc-800/70',
                    )}
                    onClick={() => { onChange(s.sessionId, cwd); setOpen(false) }}
                  >
                    <div className={cn(
                      'w-1.5 h-1.5 rounded-full mt-1.5 shrink-0',
                      s.sessionId === sessionId ? 'bg-blue-400' : 'bg-zinc-700',
                    )} />
                    <div className="flex-1 min-w-0 text-left">
                      <p className="text-zinc-200 text-xs font-medium truncate leading-relaxed">
                        {s.firstMessage || '(메시지 없음)'}
                      </p>
                      <div className="flex items-center justify-between mt-0.5">
                        <span className="flex items-center gap-1 text-[10px] text-zinc-600">
                          <Clock className="w-2.5 h-2.5" />
                          {relativeTime(s.savedAt)}
                        </span>
                        <span className="text-[10px] text-zinc-600">{s.messageCount}개 메시지</span>
                      </div>
                    </div>
                    {/* 삭제 버튼 */}
                    <button
                      onMouseDown={e => handleDelete(e, s.sessionId)}
                      className="opacity-0 group-hover:opacity-100 text-[10px] text-zinc-500 hover:text-red-400 transition-all shrink-0 px-1"
                    >
                      삭제
                    </button>
                  </div>
                ))}
              </>
            )}

            {/* 탭: Claude 세션 */}
            {tab === 'claude' && (
              <>
                {claudeSessions.length === 0 && !loading && (
                  <p className="text-center text-zinc-500 text-xs py-6">세션이 없어요</p>
                )}
                {claudeSessions.map(s => (
                  <button
                    key={s.sessionId}
                    onClick={() => { onChange(s.sessionId, s.cwd); setOpen(false) }}
                    className={cn(
                      'w-full flex items-start gap-2.5 px-3 py-2.5 text-sm hover:bg-zinc-800 transition-colors',
                      'border-b border-zinc-800/60 last:border-0',
                      s.sessionId === sessionId && 'bg-zinc-800/70',
                    )}
                  >
                    <div className={cn(
                      'w-1.5 h-1.5 rounded-full mt-1.5 shrink-0',
                      s.sessionId === sessionId ? 'bg-blue-400' : 'bg-zinc-700',
                    )} />
                    <div className="flex-1 min-w-0 text-left">
                      <p className="text-zinc-200 text-xs font-medium truncate leading-relaxed">
                        {s.firstMessage}
                      </p>
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
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
