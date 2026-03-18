'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import {
  CircleDot, CircleCheck, Loader2, RefreshCw, Plus, X,
  ExternalLink, Play, Square, AlertCircle, Tag,
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
  issue,
  selected,
  onClick,
}: {
  issue: GithubIssue
  selected: boolean
  onClick: () => void
}) {
  const isTodo = issue.labels.includes('jikime-todo')
  const isDone = issue.labels.includes('jikime-done')

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
        {issue.state === 'open'
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
        {isDone && <CircleCheck className="w-3 h-3 mt-0.5 shrink-0 text-emerald-400" />}
      </div>
    </button>
  )
}

// ── 처리 로그 뷰어 ───────────────────────────────────────────────
function ProcessingLog({
  events,
  status,
}: {
  events: string[]
  status: 'running' | 'done' | 'error' | null
}) {
  const bottomRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [events])

  if (events.length === 0 && !status) return null

  return (
    <div className="border-t border-border mt-2">
      <div className="flex items-center gap-1.5 px-3 py-1.5 bg-muted/50">
        {status === 'running' && <Loader2 className="w-3 h-3 animate-spin text-blue-400" />}
        {status === 'done'    && <CircleCheck className="w-3 h-3 text-emerald-400" />}
        {status === 'error'   && <AlertCircle className="w-3 h-3 text-red-400" />}
        <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
          {status === 'running' ? 'Processing...' : status === 'done' ? 'Completed' : status === 'error' ? 'Error' : 'Log'}
        </span>
      </div>
      <ScrollArea className="max-h-48">
        <div className="px-3 py-2 space-y-1">
          {events.map((evt, i) => (
            <p key={i} className="text-[11px] font-mono text-foreground/70 whitespace-pre-wrap break-words">{evt}</p>
          ))}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>
    </div>
  )
}

// ── 메인 컴포넌트 ─────────────────────────────────────────────────
export default function IssuesPanel() {
  const { activeProject } = useProject()
  const { getApiUrl }     = useServer()

  const [ghRepo, setGhRepo]             = useState<GitHubRepo | null>(null)
  const [repoError, setRepoError]       = useState<string | null>(null)
  const [issues, setIssues]             = useState<GithubIssue[]>([])
  const [loading, setLoading]           = useState(false)
  const [selected, setSelected]         = useState<GithubIssue | null>(null)

  // 이슈 생성 폼
  const [showForm, setShowForm]         = useState(false)
  const [newTitle, setNewTitle]         = useState('')
  const [newBody, setNewBody]           = useState('')
  const [creating, setCreating]         = useState(false)

  // ADK 처리
  const [processing, setProcessing]     = useState<string | null>(null) // issueKey
  const [procEvents, setProcEvents]     = useState<string[]>([])
  const [procStatus, setProcStatus]     = useState<'running' | 'done' | 'error' | null>(null)
  const sseRef = useRef<EventSource | null>(null)

  const pat = loadSettings().gitPat ?? ''
  const projectPath = activeProject?.path ?? ''

  // ── repo 자동 감지 ───────────────────────────────────────────
  useEffect(() => {
    if (!projectPath) { setGhRepo(null); setRepoError(null); return }
    fetch(getApiUrl(`/api/ws/github/repo?projectPath=${encodeURIComponent(projectPath)}`))
      .then(r => r.json())
      .then((data: GitHubRepo | { error: string }) => {
        if ('error' in data) { setGhRepo(null); setRepoError(data.error) }
        else { setGhRepo(data); setRepoError(null) }
      })
      .catch(e => { setGhRepo(null); setRepoError(String(e)) })
  }, [projectPath, getApiUrl])

  // ── 이슈 목록 로드 ──────────────────────────────────────────
  const loadIssues = useCallback(async () => {
    if (!projectPath || !pat) return
    setLoading(true)
    try {
      const res = await fetch(
        getApiUrl(`/api/ws/github/issues?projectPath=${encodeURIComponent(projectPath)}&token=${encodeURIComponent(pat)}`),
      )
      const data = await res.json() as { issues?: GithubIssue[]; error?: string }
      if (data.issues) setIssues(data.issues)
      else throw new Error(data.error ?? 'Failed to load issues')
    } catch (e) {
      console.error('[issues]', e)
    } finally {
      setLoading(false)
    }
  }, [projectPath, pat, getApiUrl])

  useEffect(() => { if (ghRepo && pat) loadIssues() }, [ghRepo, pat, loadIssues])

  // ── 이슈 생성 ───────────────────────────────────────────────
  const createIssue = async () => {
    if (!newTitle.trim() || !pat) return
    setCreating(true)
    try {
      const res = await fetch(getApiUrl('/api/ws/github/issues'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ projectPath, token: pat, title: newTitle.trim(), body: newBody.trim() || undefined }),
      })
      const data = await res.json() as { number?: number; error?: string }
      if (!res.ok) throw new Error(data.error ?? 'Failed')
      setNewTitle(''); setNewBody(''); setShowForm(false)
      await loadIssues()
    } catch (e) {
      alert(`이슈 생성 실패: ${(e as Error).message}`)
    } finally {
      setCreating(false)
    }
  }

  // ── ADK로 이슈 처리 ────────────────────────────────────────
  const startProcessing = async (issue: GithubIssue) => {
    if (!ghRepo || !pat) return
    const issueKey = `${ghRepo.owner}/${ghRepo.repo}#${issue.number}`
    setProcEvents([]); setProcStatus('running'); setProcessing(issueKey)

    try {
      const res = await fetch(getApiUrl('/api/ws/github/process'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          projectPath,
          token: pat,
          issueNumber: issue.number,
          issueTitle: issue.title,
          issueBody: issue.body,
          owner: ghRepo.owner,
          repo: ghRepo.repo,
          model: loadSettings().model ?? 'claude-sonnet-4-6',
        }),
      })
      const data = await res.json() as { issueKey?: string; error?: string }
      if (!res.ok) throw new Error(data.error ?? 'Failed to start')

      // SSE 연결
      sseRef.current?.close()
      const sse = new EventSource(getApiUrl(`/api/ws/github/events?issueKey=${encodeURIComponent(issueKey)}`))
      sseRef.current = sse
      sse.onmessage = (e) => {
        const msg = JSON.parse(e.data) as { type: string; message?: string; status?: string }
        if (msg.type === 'event' && msg.message) {
          setProcEvents(prev => [...prev, msg.message!])
        } else if (msg.type === 'done') {
          setProcStatus(msg.status as 'done' | 'error')
          sse.close()
          loadIssues()
        }
      }
      sse.onerror = () => {
        setProcStatus(prev => prev === 'running' ? 'error' : prev)
        sse.close()
      }
    } catch (e) {
      setProcStatus('error')
      setProcEvents(prev => [...prev, `❌ ${(e as Error).message}`])
    }
  }

  const stopProcessing = async () => {
    if (!processing) return
    sseRef.current?.close()
    await fetch(getApiUrl(`/api/ws/github/process?issueKey=${encodeURIComponent(processing)}`), { method: 'DELETE' })
    setProcStatus('error')
    setProcessing(null)
  }

  // ── 가드 ────────────────────────────────────────────────────
  if (!activeProject) {
    return (
      <div className="flex items-center justify-center flex-1 h-full text-xs text-muted-foreground/50">
        프로젝트를 선택하세요
      </div>
    )
  }

  const isProcessingSelected =
    selected !== null && ghRepo !== null &&
    processing === `${ghRepo.owner}/${ghRepo.repo}#${selected.number}`

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* 헤더 */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-border shrink-0">
        <span className="text-xs font-semibold text-foreground/80 flex-1">
          {ghRepo ? `${ghRepo.owner}/${ghRepo.repo}` : 'Issues'}
          {issues.length > 0 && <span className="ml-1 text-muted-foreground">({issues.length})</span>}
        </span>
        <Button variant="ghost" size="icon" className="h-6 w-6" onClick={loadIssues} disabled={loading || !pat}>
          <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
        </Button>
        <Button
          variant="outline" size="sm"
          className="h-6 px-2 text-xs gap-1"
          onClick={() => setShowForm(v => !v)}
          disabled={!ghRepo || !pat}
        >
          <Plus className="w-3 h-3" />
          New
        </Button>
      </div>

      {/* PAT/Repo 경고 */}
      {!pat && (
        <div className="mx-3 mt-2 p-2 rounded-lg bg-amber-950/30 border border-amber-700/40 text-xs text-amber-300">
          GitHub PAT가 설정되지 않았습니다. 설정 → Git PAT에서 입력해 주세요.
        </div>
      )}
      {pat && repoError && (
        <div className="mx-3 mt-2 p-2 rounded-lg bg-red-950/30 border border-red-700/40 text-xs text-red-300">
          GitHub 저장소를 감지할 수 없습니다. git remote origin이 GitHub URL이어야 합니다.
        </div>
      )}

      {/* 이슈 생성 폼 */}
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
            {creating ? '생성 중...' : 'jikime-todo로 생성'}
          </Button>
        </div>
      )}

      {/* 이슈 목록 */}
      <ScrollArea className="flex-1 min-h-0">
        {loading ? (
          <div className="flex items-center justify-center py-8 gap-2 text-xs text-muted-foreground">
            <Loader2 className="w-4 h-4 animate-spin" />
            로딩 중...
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
              onClick={() => setSelected(prev => prev?.number === issue.number ? null : issue)}
            />
          ))
        )}
      </ScrollArea>

      {/* 선택된 이슈 상세 + ADK 처리 버튼 */}
      {selected && (
        <div className="shrink-0 border-t border-border">
          <div className="px-3 py-2 space-y-2">
            {/* 이슈 링크 */}
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

            {/* 처리 버튼 */}
            {selected.state === 'open' && selected.labels.includes('jikime-todo') && (
              isProcessingSelected ? (
                <Button
                  size="sm"
                  variant="outline"
                  className="w-full h-7 text-xs gap-1.5 border-red-500/40 text-red-400 hover:bg-red-950/30"
                  onClick={stopProcessing}
                >
                  <Square className="w-3 h-3 fill-current" />
                  처리 중지
                </Button>
              ) : (
                <Button
                  size="sm"
                  className="w-full h-7 text-xs gap-1.5 bg-emerald-700 hover:bg-emerald-600 text-white"
                  onClick={() => startProcessing(selected)}
                  disabled={procStatus === 'running' && !isProcessingSelected}
                >
                  <Play className="w-3 h-3 fill-current" />
                  ADK로 처리
                </Button>
              )
            )}
          </div>

          {/* 처리 로그 */}
          {isProcessingSelected && (
            <ProcessingLog events={procEvents} status={procStatus} />
          )}
        </div>
      )}
    </div>
  )
}
