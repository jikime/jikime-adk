'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import {
  CircleDot, CircleCheck, Loader2, RefreshCw, Plus, X,
  ExternalLink, Play, Square, AlertCircle, Tag, Radio,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import { loadSettings } from '@/components/sidebar/Sidebar'

// ── 타입 ──────────────────────────────────────────────────────────
interface GithubIssue {
  number: number
  title: string
  state: 'open' | 'closed'
  body: string
  url: string
  createdAt: string
  closedAt: string | null
  labels: string[]
}

interface GitHubRepo {
  owner: string
  repo: string
}

interface PollerStatus {
  status: 'running' | 'stopped'
  owner?: string
  repo?: string
  intervalMs?: number
  maxConcurrent?: number
  lastCheck?: string | null
  activeCount?: number
  activeIssues?: number[]
}

interface PollerEvent {
  type: 'tick' | 'issue_found' | 'issue_done' | 'error'
  lastCheck?: string
  activeCount?: number
  activeIssues?: number[]
  issueNumber?: number
  issueTitle?: string
  status?: string
  message?: string
}

// ── 라벨 뱃지 ────────────────────────────────────────────────────
function LabelBadge({ name }: { name: string }) {
  const style =
    name === 'jikime-todo' ? 'bg-blue-500/20 text-blue-400 border-blue-500/30' :
    name === 'jikime-done' ? 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30' :
    'bg-muted text-muted-foreground border-border'
  return (
    <span className={cn('inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[10px] font-medium border', style)}>
      <Tag className="w-2.5 h-2.5" />
      {name}
    </span>
  )
}

// ── 이슈 카드 ────────────────────────────────────────────────────
function IssueCard({
  issue, selected, processing, onClick,
}: {
  issue: GithubIssue
  selected: boolean
  processing: boolean
  onClick: () => void
}) {
  const isTodo = issue.labels.includes('jikime-todo')

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'w-full text-left px-3 py-2.5 border-b border-border/50 last:border-0 transition-colors',
        selected ? 'bg-primary/10' : 'hover:bg-muted/50',
      )}
    >
      <div className="flex items-start gap-2">
        {processing
          ? <Loader2 className="w-3.5 h-3.5 mt-0.5 shrink-0 text-blue-400 animate-spin" />
          : issue.state === 'open'
          ? <CircleDot className={cn('w-3.5 h-3.5 mt-0.5 shrink-0', isTodo ? 'text-blue-400' : 'text-emerald-400')} />
          : <CircleCheck className="w-3.5 h-3.5 mt-0.5 shrink-0 text-muted-foreground/50" />
        }
        <div className="flex-1 min-w-0">
          <p className="text-xs font-medium text-foreground/90 truncate">
            <span className="text-muted-foreground mr-1">#{issue.number}</span>
            {issue.title}
          </p>
          {issue.labels.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-1">
              {issue.labels.map(l => <LabelBadge key={l} name={l} />)}
            </div>
          )}
        </div>
      </div>
    </button>
  )
}

// ── 처리 로그 뷰어 ───────────────────────────────────────────────
function ProcessingLog({ events, status }: { events: string[]; status: 'running' | 'done' | 'error' | null }) {
  const bottomRef = useRef<HTMLDivElement>(null)
  useEffect(() => { bottomRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [events])
  if (events.length === 0) return null
  return (
    <div className="border-t border-border">
      <div className="flex items-center gap-1.5 px-3 py-1.5 bg-muted/50">
        {status === 'running' && <Loader2 className="w-3 h-3 animate-spin text-blue-400" />}
        {status === 'done'    && <CircleCheck className="w-3 h-3 text-emerald-400" />}
        {status === 'error'   && <AlertCircle className="w-3 h-3 text-red-400" />}
        <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
          {status === 'running' ? '처리 중...' : status === 'done' ? '완료' : '오류'}
        </span>
      </div>
      <ScrollArea className="max-h-36">
        <div className="px-3 py-2 space-y-0.5">
          {events.map((evt, i) => (
            <p key={i} className="text-[11px] font-mono text-foreground/70 whitespace-pre-wrap break-words">{evt}</p>
          ))}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>
    </div>
  )
}

// ── 폴러 상태 배너 ────────────────────────────────────────────────
function PollerBanner({
  pollerStatus, onStop, lastEventMsg,
}: {
  pollerStatus: PollerStatus | null
  onStop: () => void
  lastEventMsg: string | null
}) {
  if (!pollerStatus || pollerStatus.status !== 'running') return null
  const lastCheckLabel = pollerStatus.lastCheck
    ? `${Math.round((Date.now() - new Date(pollerStatus.lastCheck).getTime()) / 1000)}초 전`
    : '확인 중...'
  return (
    <div className="mx-3 mt-2 rounded-lg border border-blue-500/30 bg-blue-950/20 px-3 py-2 shrink-0">
      <div className="flex items-center gap-2">
        <Radio className="w-3.5 h-3.5 text-blue-400 animate-pulse shrink-0" />
        <div className="flex-1 min-w-0">
          <p className="text-xs font-medium text-blue-300">
            자동 폴링 중
            <span className="ml-1.5 text-blue-400/70 font-mono text-[10px]">
              {(pollerStatus.intervalMs ?? 15000) / 1000}s 주기
            </span>
            {(pollerStatus.activeCount ?? 0) > 0 && (
              <span className="ml-1.5 text-amber-400 text-[10px]">
                {pollerStatus.activeCount}개 처리 중
              </span>
            )}
          </p>
          {lastEventMsg && (
            <p className="text-[10px] text-blue-400/60 truncate mt-0.5">{lastEventMsg}</p>
          )}
          <p className="text-[10px] text-blue-400/50 mt-0.5">마지막 확인: {lastCheckLabel}</p>
        </div>
        <button
          type="button"
          onClick={onStop}
          className="shrink-0 p-1 rounded hover:bg-red-950/40 text-red-400 hover:text-red-300 transition-colors"
          title="폴링 중지"
        >
          <Square className="w-3 h-3 fill-current" />
        </button>
      </div>
    </div>
  )
}

// ── 메인 컴포넌트 ─────────────────────────────────────────────────
export default function IssuesPanel() {
  const { activeProject }   = useProject()
  const { getApiUrl }       = useServer()

  const [ghRepo, setGhRepo]         = useState<GitHubRepo | null>(null)
  const [repoError, setRepoError]   = useState<string | null>(null)
  const [issues, setIssues]         = useState<GithubIssue[]>([])
  const [loading, setLoading]       = useState(false)
  const [selected, setSelected]     = useState<GithubIssue | null>(null)

  // 이슈 생성 폼
  const [showForm, setShowForm]     = useState(false)
  const [newTitle, setNewTitle]     = useState('')
  const [newBody, setNewBody]       = useState('')
  const [creating, setCreating]     = useState(false)

  // 개별 이슈 ADK 처리 (수동)
  const [manualProcIssueKey, setManualProcIssueKey]     = useState<string | null>(null)
  const [manualProcEvents, setManualProcEvents]         = useState<string[]>([])
  const [manualProcStatus, setManualProcStatus]         = useState<'running' | 'done' | 'error' | null>(null)
  const manualSseRef = useRef<EventSource | null>(null)

  // 자동 폴러
  const [pollerStatus, setPollerStatus]   = useState<PollerStatus | null>(null)
  const [pollerStarting, setPollerStarting] = useState(false)
  const [activeIssueNums, setActiveIssueNums] = useState<Set<number>>(new Set())
  const [lastEventMsg, setLastEventMsg]   = useState<string | null>(null)
  const pollerSseRef = useRef<EventSource | null>(null)

  const pat         = loadSettings().gitPat ?? ''
  const projectPath = activeProject?.path ?? ''

  // ── repo 감지 ──────────────────────────────────────────────
  useEffect(() => {
    if (!projectPath) { setGhRepo(null); setRepoError(null); return }
    fetch(getApiUrl(`/api/ws/github/repo?projectPath=${encodeURIComponent(projectPath)}`))
      .then(r => r.json())
      .then((d: GitHubRepo | { error: string }) => {
        if ('error' in d) { setGhRepo(null); setRepoError(d.error) }
        else { setGhRepo(d); setRepoError(null) }
      })
      .catch(e => { setGhRepo(null); setRepoError(String(e)) })
  }, [projectPath, getApiUrl])

  // ── 이슈 목록 로드 ──────────────────────────────────────────
  const loadIssues = useCallback(async () => {
    if (!projectPath || !pat) return
    setLoading(true)
    try {
      const r = await fetch(
        getApiUrl(`/api/ws/github/issues?projectPath=${encodeURIComponent(projectPath)}&token=${encodeURIComponent(pat)}`),
      )
      const d = await r.json() as { issues?: GithubIssue[]; error?: string }
      if (d.issues) setIssues(d.issues)
    } finally { setLoading(false) }
  }, [projectPath, pat, getApiUrl])

  useEffect(() => { if (ghRepo && pat) loadIssues() }, [ghRepo, pat, loadIssues])

  // ── 폴러 상태 조회 (마운트 시) ────────────────────────────
  useEffect(() => {
    if (!projectPath) return
    fetch(getApiUrl(`/api/ws/github/poller?projectPath=${encodeURIComponent(projectPath)}`))
      .then(r => r.json())
      .then((d: PollerStatus) => setPollerStatus(d))
      .catch(() => {})
  }, [projectPath, getApiUrl])

  // ── 폴러 SSE 연결/해제 ────────────────────────────────────
  const connectPollerSSE = useCallback(() => {
    if (!projectPath) return
    pollerSseRef.current?.close()
    const sse = new EventSource(
      getApiUrl(`/api/ws/github/poller-events?projectPath=${encodeURIComponent(projectPath)}`),
    )
    pollerSseRef.current = sse
    sse.onmessage = (e) => {
      const msg = JSON.parse(e.data) as PollerEvent
      if (msg.type === 'tick') {
        setPollerStatus(prev => ({
          ...prev,
          status: 'running',
          lastCheck: msg.lastCheck ?? prev?.lastCheck,
          activeCount: msg.activeCount ?? 0,
          activeIssues: msg.activeIssues ?? [],
        }))
        if (msg.activeIssues) setActiveIssueNums(new Set(msg.activeIssues))
      } else if (msg.type === 'issue_found') {
        setLastEventMsg(`🔍 이슈 #${msg.issueNumber} 발견: ${msg.issueTitle}`)
        if (msg.issueNumber) setActiveIssueNums(prev => new Set([...prev, msg.issueNumber!]))
        loadIssues()
      } else if (msg.type === 'issue_done') {
        setLastEventMsg(
          msg.status === 'done'
            ? `✅ 이슈 #${msg.issueNumber} 완료`
            : `❌ 이슈 #${msg.issueNumber} 오류`,
        )
        if (msg.issueNumber) setActiveIssueNums(prev => { const s = new Set(prev); s.delete(msg.issueNumber!); return s })
        loadIssues()
      } else if (msg.type === 'error') {
        setLastEventMsg(`⚠️ ${msg.message}`)
      }
    }
    sse.onerror = () => sse.close()
  }, [projectPath, getApiUrl, loadIssues])

  // 폴러가 running이면 SSE 자동 연결
  useEffect(() => {
    if (pollerStatus?.status === 'running') {
      connectPollerSSE()
    } else {
      pollerSseRef.current?.close()
    }
    return () => pollerSseRef.current?.close()
  }, [pollerStatus?.status, connectPollerSSE])

  // ── 폴러 시작 ────────────────────────────────────────────
  const startPoller = async () => {
    if (!pat || !projectPath) return
    setPollerStarting(true)
    try {
      const r = await fetch(getApiUrl('/api/ws/github/poller'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          projectPath,
          token: pat,
          intervalMs: 15000,
          maxConcurrent: 3,
          model: loadSettings().model ?? 'claude-sonnet-4-6',
        }),
      })
      const d = await r.json() as PollerStatus & { error?: string }
      if (!r.ok) throw new Error(d.error ?? 'Failed')
      setPollerStatus({ ...d, status: 'running' })
      connectPollerSSE()
    } catch (e) {
      alert(`폴링 시작 실패: ${(e as Error).message}`)
    } finally {
      setPollerStarting(false)
    }
  }

  // ── 폴러 중지 ────────────────────────────────────────────
  const stopPoller = async () => {
    pollerSseRef.current?.close()
    await fetch(getApiUrl(`/api/ws/github/poller?projectPath=${encodeURIComponent(projectPath)}`), { method: 'DELETE' })
    setPollerStatus({ status: 'stopped' })
    setActiveIssueNums(new Set())
    setLastEventMsg(null)
  }

  // ── 이슈 생성 ────────────────────────────────────────────
  const createIssue = async () => {
    if (!newTitle.trim() || !pat) return
    setCreating(true)
    try {
      const r = await fetch(getApiUrl('/api/ws/github/issues'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ projectPath, token: pat, title: newTitle.trim(), body: newBody.trim() || undefined }),
      })
      const d = await r.json() as { number?: number; error?: string }
      if (!r.ok) throw new Error(d.error ?? 'Failed')
      setNewTitle(''); setNewBody(''); setShowForm(false)
      await loadIssues()
    } catch (e) {
      alert(`이슈 생성 실패: ${(e as Error).message}`)
    } finally {
      setCreating(false)
    }
  }

  // ── 수동 ADK 처리 (개별 이슈) ────────────────────────────
  const startManualProcess = async (issue: GithubIssue) => {
    if (!ghRepo || !pat) return
    const issueKey = `${ghRepo.owner}/${ghRepo.repo}#${issue.number}`
    setManualProcEvents([]); setManualProcStatus('running'); setManualProcIssueKey(issueKey)

    try {
      const r = await fetch(getApiUrl('/api/ws/github/process'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          projectPath, token: pat,
          issueNumber: issue.number,
          issueTitle: issue.title,
          issueBody: issue.body,
          owner: ghRepo.owner,
          repo: ghRepo.repo,
          model: loadSettings().model ?? 'claude-sonnet-4-6',
        }),
      })
      if (!r.ok) throw new Error((await r.json() as { error?: string }).error ?? 'Failed')

      manualSseRef.current?.close()
      const sse = new EventSource(getApiUrl(`/api/ws/github/events?issueKey=${encodeURIComponent(issueKey)}`))
      manualSseRef.current = sse
      sse.onmessage = (e) => {
        const msg = JSON.parse(e.data) as { type: string; message?: string; status?: string }
        if (msg.type === 'event' && msg.message) setManualProcEvents(prev => [...prev, msg.message!])
        else if (msg.type === 'done') {
          setManualProcStatus(msg.status as 'done' | 'error')
          sse.close(); loadIssues()
        }
      }
      sse.onerror = () => { setManualProcStatus(p => p === 'running' ? 'error' : p); sse.close() }
    } catch (e) {
      setManualProcStatus('error')
      setManualProcEvents(p => [...p, `❌ ${(e as Error).message}`])
    }
  }

  const stopManualProcess = async () => {
    if (!manualProcIssueKey) return
    manualSseRef.current?.close()
    await fetch(getApiUrl(`/api/ws/github/process?issueKey=${encodeURIComponent(manualProcIssueKey)}`), { method: 'DELETE' })
    setManualProcStatus('error'); setManualProcIssueKey(null)
  }

  // ── 가드 ────────────────────────────────────────────────
  if (!activeProject) {
    return (
      <div className="flex items-center justify-center flex-1 h-full text-xs text-muted-foreground/50">
        프로젝트를 선택하세요
      </div>
    )
  }

  const isPolling = pollerStatus?.status === 'running'
  const isManualProcessingSelected =
    selected !== null && ghRepo !== null &&
    manualProcIssueKey === `${ghRepo.owner}/${ghRepo.repo}#${selected.number}`
  const isAutoProcessingSelected =
    selected !== null && activeIssueNums.has(selected.number)

  return (
    <div className="flex flex-col h-full overflow-hidden">

      {/* ── 헤더 ── */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-border shrink-0">
        <span className="text-xs font-semibold text-foreground/80 flex-1 truncate">
          {ghRepo ? `${ghRepo.owner}/${ghRepo.repo}` : 'Issues'}
          {issues.length > 0 && <span className="ml-1 text-muted-foreground">({issues.length})</span>}
        </span>
        <Button variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={loadIssues} disabled={loading || !pat}>
          <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
        </Button>
        <Button
          variant="outline" size="sm"
          className="h-6 px-2 text-xs gap-1 shrink-0"
          onClick={() => setShowForm(v => !v)}
          disabled={!ghRepo || !pat}
        >
          <Plus className="w-3 h-3" /> New
        </Button>
      </div>

      {/* ── 경고 ── */}
      {!pat && (
        <div className="mx-3 mt-2 p-2 rounded-lg bg-amber-950/30 border border-amber-700/40 text-xs text-amber-300 shrink-0">
          GitHub PAT가 설정되지 않았습니다. 설정 → Git PAT에서 입력해 주세요.
        </div>
      )}
      {pat && repoError && (
        <div className="mx-3 mt-2 p-2 rounded-lg bg-red-950/30 border border-red-700/40 text-xs text-red-300 shrink-0">
          GitHub 저장소 감지 불가. git remote origin이 GitHub URL이어야 합니다.
        </div>
      )}

      {/* ── 자동 폴링 배너 ── */}
      <PollerBanner pollerStatus={pollerStatus} onStop={stopPoller} lastEventMsg={lastEventMsg} />

      {/* ── 폴링 시작 버튼 (미실행 시) ── */}
      {!isPolling && ghRepo && pat && (
        <div className="mx-3 mt-2 shrink-0">
          <Button
            size="sm"
            className="w-full h-8 text-xs gap-2 bg-blue-700 hover:bg-blue-600 text-white"
            onClick={startPoller}
            disabled={pollerStarting}
          >
            {pollerStarting
              ? <Loader2 className="w-3.5 h-3.5 animate-spin" />
              : <Radio className="w-3.5 h-3.5" />
            }
            {pollerStarting ? '시작 중...' : '자동 폴링 시작 (jikime-todo)'}
          </Button>
        </div>
      )}

      {/* ── 이슈 생성 폼 ── */}
      {showForm && (
        <div className="mx-3 mt-2 p-3 rounded-lg border border-border bg-card space-y-2 shrink-0">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-foreground/80">새 이슈</span>
            <button onClick={() => setShowForm(false)} className="text-muted-foreground hover:text-foreground">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
          <Input
            value={newTitle}
            onChange={e => setNewTitle(e.target.value)}
            placeholder="이슈 제목..."
            className="h-7 text-xs"
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) createIssue() }}
          />
          <textarea
            value={newBody}
            onChange={e => setNewBody(e.target.value)}
            placeholder="설명 (선택)..."
            rows={3}
            className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-xs resize-none focus:outline-none focus:ring-1 focus:ring-ring"
          />
          <Button
            size="sm"
            className="w-full h-7 text-xs bg-blue-700 hover:bg-blue-600 text-white"
            onClick={createIssue}
            disabled={!newTitle.trim() || creating}
          >
            {creating ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : <Plus className="w-3 h-3 mr-1" />}
            {creating ? '생성 중...' : 'jikime-todo 이슈 생성'}
          </Button>
        </div>
      )}

      {/* ── 이슈 목록 ── */}
      <ScrollArea className="flex-1 min-h-0 mt-2">
        {loading ? (
          <div className="flex items-center justify-center py-8 gap-2 text-xs text-muted-foreground">
            <Loader2 className="w-4 h-4 animate-spin" /> 로딩 중...
          </div>
        ) : issues.length === 0 ? (
          <p className="text-xs text-muted-foreground/50 text-center py-8">
            {!pat ? 'PAT를 설정하면 이슈를 불러옵니다' : '이슈 없음'}
          </p>
        ) : (
          issues.map(issue => (
            <IssueCard
              key={issue.number}
              issue={issue}
              selected={selected?.number === issue.number}
              processing={activeIssueNums.has(issue.number) || (manualProcIssueKey !== null &&
                ghRepo !== null &&
                manualProcIssueKey === `${ghRepo.owner}/${ghRepo.repo}#${issue.number}` &&
                manualProcStatus === 'running')}
              onClick={() => setSelected(prev => prev?.number === issue.number ? null : issue)}
            />
          ))
        )}
      </ScrollArea>

      {/* ── 선택된 이슈 상세 ── */}
      {selected && (
        <div className="shrink-0 border-t border-border">
          <div className="px-3 py-2 space-y-1.5">
            <div className="flex items-center gap-1.5">
              <span className="text-xs font-semibold text-foreground/80 truncate flex-1">
                #{selected.number} {selected.title}
              </span>
              <a
                href={selected.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-foreground transition-colors shrink-0"
              >
                <ExternalLink className="w-3.5 h-3.5" />
              </a>
            </div>

            {/* 자동 폴링으로 처리 중 표시 */}
            {isAutoProcessingSelected && (
              <div className="flex items-center gap-1.5 text-[11px] text-blue-400">
                <Loader2 className="w-3 h-3 animate-spin" />
                자동 폴링으로 처리 중...
              </div>
            )}

            {/* 수동 처리 버튼 (jikime-todo 이슈만, 폴링 미처리 중) */}
            {selected.state === 'open' && selected.labels.includes('jikime-todo') && !isAutoProcessingSelected && (
              isManualProcessingSelected && manualProcStatus === 'running' ? (
                <Button
                  size="sm" variant="outline"
                  className="w-full h-7 text-xs gap-1.5 border-red-500/40 text-red-400 hover:bg-red-950/30"
                  onClick={stopManualProcess}
                >
                  <Square className="w-3 h-3 fill-current" /> 수동 처리 중지
                </Button>
              ) : (
                <Button
                  size="sm"
                  className="w-full h-7 text-xs gap-1.5 bg-emerald-700 hover:bg-emerald-600 text-white"
                  onClick={() => startManualProcess(selected)}
                  disabled={manualProcStatus === 'running'}
                >
                  <Play className="w-3 h-3 fill-current" /> ADK로 수동 처리
                </Button>
              )
            )}
          </div>

          {/* 수동 처리 로그 */}
          {isManualProcessingSelected && (
            <ProcessingLog events={manualProcEvents} status={manualProcStatus} />
          )}
        </div>
      )}
    </div>
  )
}
