'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import {
  CircleDot, CircleCheck, Loader2, RefreshCw, Plus, X,
  ExternalLink, Play, Square, AlertCircle, Tag, Radio, Zap, Settings,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import { useLocale } from '@/contexts/LocaleContext'
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

interface HarnessStatus {
  status: 'running' | 'stopped'
  projectSlug?: string
  intervalMs?: number
  maxConcurrent?: number
  lastCheck?: string | null
  activeCount?: number
}

interface HarnessEvent {
  type: 'tick' | 'issue_found' | 'issue_done' | 'error' | 'worker_event' | 'retrying'
  lastCheck?: string
  activeCount?: number
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

type Gt = import('@/i18n').Messages['git']

// ── 상태 뱃지 ────────────────────────────────────────────────────
function StateBadge({ state }: { state: 'open' | 'closed' }) {
  return state === 'open' ? (
    <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[10px] font-medium border bg-emerald-500/15 text-emerald-400 border-emerald-500/30">
      <CircleDot className="w-2.5 h-2.5" /> open
    </span>
  ) : (
    <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[10px] font-medium border bg-muted text-muted-foreground/60 border-border">
      <CircleCheck className="w-2.5 h-2.5" /> closed
    </span>
  )
}

// ── 이슈 카드 ────────────────────────────────────────────────────
function IssueCard({
  issue, selected, processing, onClick, t,
}: {
  issue: GithubIssue
  selected: boolean
  processing: boolean
  onClick: () => void
  t: Gt
}) {
  const isPollingTarget = issue.state === 'open' && issue.labels.includes('jikime-todo')

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'w-full text-left px-3 py-2.5 border-b border-border/50 last:border-0 transition-colors',
        selected ? 'bg-primary/10' : 'hover:bg-muted/50',
        isPollingTarget && !selected && 'bg-blue-950/10',
      )}
    >
      <div className="flex items-start gap-2">
        {processing
          ? <Loader2 className="w-3.5 h-3.5 mt-0.5 shrink-0 text-blue-400 animate-spin" />
          : isPollingTarget
          ? <Radio className="w-3.5 h-3.5 mt-0.5 shrink-0 text-blue-400" />
          : issue.state === 'open'
          ? <CircleDot className="w-3.5 h-3.5 mt-0.5 shrink-0 text-emerald-400" />
          : <CircleCheck className="w-3.5 h-3.5 mt-0.5 shrink-0 text-muted-foreground/40" />
        }
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5 mb-0.5">
            <StateBadge state={issue.state} />
            {processing && (
              <span className="text-[10px] text-blue-400 font-medium">{t.issuesProcessing}</span>
            )}
            {isPollingTarget && !processing && (
              <span className="text-[10px] text-blue-400/70">{t.issuesPollingLabel}</span>
            )}
          </div>
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
function ProcessingLog({ events, status, t }: { events: string[]; status: 'running' | 'done' | 'error' | null; t: Gt }) {
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
          {status === 'running' ? t.issuesLogRunning : status === 'done' ? t.issuesLogDone : t.issuesLogError}
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

// ── Harness Worker 스트림 뷰어 ──────────────────────────────────────
interface WorkerMsg {
  id: number
  kind: 'tool' | 'status' | 'claude'
  text: string
}

function parseWorkerMsg(raw: string): WorkerMsg['kind'] {
  if (/^🔧\s/.test(raw) && raw.includes(':')) return 'tool'
  if (/^[✅❌⚠️🔍🚀📝🔄💡🏁]/.test(raw)) return 'status'
  return 'claude'
}

// 도구 메시지 파싱: "🔧 ToolName: {json}" → { name, args }
function parseTool(raw: string): { name: string; args: Record<string, unknown> } | null {
  const m = raw.match(/^🔧\s+(\w+):\s*(\{[\s\S]*)$/)
  if (!m) return null
  try { return { name: m[1], args: JSON.parse(m[2]) } }
  catch { return { name: m[1], args: {} } }
}

// 도구별 스타일
const TOOL_STYLES: Record<string, { headerBg: string; border: string; nameColor: string; label: string }> = {
  Bash:      { headerBg: 'bg-zinc-800',       border: 'border-zinc-600/60',    nameColor: 'text-emerald-400', label: '$'  },
  Read:      { headerBg: 'bg-blue-950/60',    border: 'border-blue-700/50',    nameColor: 'text-blue-300',    label: 'R'  },
  Write:     { headerBg: 'bg-amber-950/60',   border: 'border-amber-700/50',   nameColor: 'text-amber-300',   label: 'W'  },
  Edit:      { headerBg: 'bg-amber-950/60',   border: 'border-amber-700/50',   nameColor: 'text-amber-300',   label: 'E'  },
  MultiEdit: { headerBg: 'bg-amber-950/60',   border: 'border-amber-700/50',   nameColor: 'text-amber-300',   label: 'ME' },
  Glob:      { headerBg: 'bg-purple-950/60',  border: 'border-purple-700/50',  nameColor: 'text-purple-300',  label: 'G'  },
  Grep:      { headerBg: 'bg-teal-950/60',    border: 'border-teal-700/50',    nameColor: 'text-teal-300',    label: 'F'  },
  Agent:     { headerBg: 'bg-violet-950/60',  border: 'border-violet-700/50',  nameColor: 'text-violet-300',  label: 'A'  },
  WebSearch: { headerBg: 'bg-sky-950/60',     border: 'border-sky-700/50',     nameColor: 'text-sky-300',     label: '🌐' },
  WebFetch:  { headerBg: 'bg-sky-950/60',     border: 'border-sky-700/50',     nameColor: 'text-sky-300',     label: '🌐' },
  Task:      { headerBg: 'bg-indigo-950/60',  border: 'border-indigo-700/50',  nameColor: 'text-indigo-300',  label: 'T'  },
}
const TOOL_STYLE_DEFAULT = { headerBg: 'bg-muted/80', border: 'border-border', nameColor: 'text-muted-foreground', label: '⚙' }

function ToolBubble({ text }: { text: string }) {
  const parsed = parseTool(text)
  if (!parsed) {
    return (
      <div className="text-[10px] font-mono text-muted-foreground/50 bg-muted/50 rounded px-2 py-1 break-all">
        {text.replace(/^🔧\s*/, '')}
      </div>
    )
  }
  const { name, args } = parsed
  const s = TOOL_STYLES[name] ?? TOOL_STYLE_DEFAULT

  let body: React.ReactNode = null
  switch (name) {
    case 'Bash':
      body = (
        <code className="block text-[10px] font-mono text-emerald-300/90 whitespace-pre-wrap break-all leading-relaxed">
          {String(args.command ?? '')}
        </code>
      )
      break
    case 'Read':
      body = <span className="text-[10px] font-mono text-blue-300/80 break-all">{String(args.file_path ?? '')}</span>
      break
    case 'Write':
    case 'Edit':
    case 'MultiEdit':
      body = <span className="text-[10px] font-mono text-amber-300/80 break-all">{String(args.file_path ?? '')}</span>
      break
    case 'Glob':
      body = (
        <div className="space-y-0.5">
          <div className="text-[10px] font-mono text-purple-300/80">{String(args.pattern ?? '')}</div>
          {args.path ? <div className="text-[9px] font-mono text-muted-foreground/40">{String(args.path)}</div> : null}
        </div>
      )
      break
    case 'Grep':
      body = (
        <div className="space-y-0.5">
          <div className="text-[10px] font-mono text-teal-300/80">{String(args.pattern ?? '')}</div>
          {args.path ? <div className="text-[9px] font-mono text-muted-foreground/40">{String(args.path)}</div> : null}
        </div>
      )
      break
    case 'Agent':
      body = (
        <div className="space-y-1">
          {args.subagent_type ? (
            <span className="inline-block text-[9px] bg-violet-500/25 text-violet-300 rounded px-1.5 py-0.5 font-mono font-semibold uppercase">
              {String(args.subagent_type)}
            </span>
          ) : null}
          {(args.description || args.prompt) ? (
            <div className="text-[10px] text-violet-200/70 leading-relaxed">
              {String(args.description ?? args.prompt ?? '').slice(0, 120)}
            </div>
          ) : null}
        </div>
      )
      break
    case 'Task':
      body = (
        <div className="text-[10px] text-indigo-200/70 leading-relaxed">
          {String(args.description ?? args.prompt ?? Object.values(args)[0] ?? '').slice(0, 120)}
        </div>
      )
      break
    case 'WebSearch':
      body = <span className="text-[10px] text-sky-300/80">{String(args.query ?? '')}</span>
      break
    case 'WebFetch':
      body = <span className="text-[10px] font-mono text-sky-300/80 break-all">{String(args.url ?? '')}</span>
      break
    default: {
      const first = Object.entries(args)[0]
      if (first) body = <span className="text-[10px] font-mono text-muted-foreground/60 break-all">{String(first[1]).slice(0, 120)}</span>
    }
  }

  return (
    <div className={cn('rounded-md border overflow-hidden text-left', s.border)}>
      <div className={cn('flex items-center gap-1.5 px-2 py-0.5', s.headerBg)}>
        <span className={cn('text-[9px] font-bold font-mono w-3.5 text-center', s.nameColor)}>{s.label}</span>
        <span className="text-[9px] font-semibold text-white/50 uppercase tracking-widest">{name}</span>
      </div>
      {body && (
        <div className={cn('px-2 py-1.5', s.headerBg, 'brightness-75')}>
          {body}
        </div>
      )}
    </div>
  )
}

function HarnessWorkerStream({
  projectPath, issueNumber, getApiUrl,
}: {
  projectPath: string
  issueNumber: number
  getApiUrl: (path: string) => string
}) {
  const [msgs, setMsgs] = useState<WorkerMsg[]>([])
  const [streamStatus, setStreamStatus] = useState<'running' | 'done' | 'error'>('running')
  const bottomRef = useRef<HTMLDivElement>(null)
  const sseRef = useRef<EventSource | null>(null)
  const idRef = useRef(0)

  useEffect(() => {
    setMsgs([])
    setStreamStatus('running')
    const url = getApiUrl(
      `/api/harness/worker-events?projectPath=${encodeURIComponent(projectPath)}&issueNumber=${issueNumber}`,
    )
    const sse = new EventSource(url)
    sseRef.current = sse
    sse.onmessage = (e) => {
      const data = JSON.parse(e.data) as { type: string; message?: string; status?: string }
      if (data.type === 'event' && data.message) {
        const text = data.message
        setMsgs(prev => [...prev, { id: idRef.current++, kind: parseWorkerMsg(text), text }])
      } else if (data.type === 'done') {
        setStreamStatus(data.status === 'error' ? 'error' : 'done')
        sse.close()
      }
    }
    sse.onerror = () => { setStreamStatus('error'); sse.close() }
    return () => { sse.close(); sseRef.current = null }
  }, [projectPath, issueNumber, getApiUrl])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [msgs])

  return (
    <div className="mt-2 rounded-lg border border-border overflow-hidden">
      {/* 헤더 */}
      <div className="flex items-center gap-1.5 px-2.5 py-1.5 bg-muted/50 border-b border-border/50">
        {streamStatus === 'running' && <Loader2 className="w-3 h-3 animate-spin text-blue-400 shrink-0" />}
        {streamStatus === 'done'    && <CircleCheck className="w-3 h-3 text-emerald-400 shrink-0" />}
        {streamStatus === 'error'   && <AlertCircle className="w-3 h-3 text-red-400 shrink-0" />}
        <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
          {streamStatus === 'running' ? 'Processing...' : streamStatus === 'done' ? 'Completed' : 'Error'}
        </span>
      </div>
      {/* 메시지 목록 */}
      <ScrollArea className="max-h-64">
        <div className="px-2.5 py-2 space-y-1.5">
          {msgs.length === 0 && streamStatus === 'running' && (
            <p className="text-[10px] text-muted-foreground/50 italic">Waiting for agent...</p>
          )}
          {msgs.map(m => {
            if (m.kind === 'tool') {
              return <ToolBubble key={m.id} text={m.text} />
            }
            if (m.kind === 'status') {
              return (
                <p key={m.id} className="text-[10px] text-muted-foreground/70 leading-relaxed">
                  {m.text}
                </p>
              )
            }
            // claude 말풍선
            return (
              <div key={m.id} className="flex items-start gap-1.5">
                <span className="mt-0.5 shrink-0 w-4 h-4 rounded-full bg-orange-500/80 flex items-center justify-center text-[8px] font-bold text-white">C</span>
                <div className="flex-1 bg-orange-500/8 border border-orange-500/15 rounded-lg px-2.5 py-1.5 min-w-0">
                  <div className="prose prose-xs dark:prose-invert max-w-none
                    prose-p:my-0.5 prose-p:leading-relaxed prose-p:text-[11px]
                    prose-headings:text-foreground/90 prose-headings:font-semibold prose-headings:mt-2 prose-headings:mb-0.5
                    prose-h1:text-sm prose-h2:text-xs prose-h3:text-xs
                    prose-strong:text-foreground/90
                    prose-code:text-amber-600 dark:prose-code:text-amber-300 prose-code:bg-muted prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:text-[10px] prose-code:before:content-none prose-code:after:content-none
                    prose-pre:bg-muted prose-pre:border prose-pre:border-border prose-pre:rounded prose-pre:text-[10px] prose-pre:my-1
                    prose-ul:my-0.5 prose-ol:my-0.5 prose-li:my-0 prose-li:text-[11px]
                    prose-a:text-blue-500 dark:prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline
                    prose-blockquote:border-border prose-blockquote:text-muted-foreground prose-blockquote:not-italic prose-blockquote:text-[11px]
                    prose-hr:border-border prose-hr:my-1
                    text-foreground/80">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                      {m.text}
                    </ReactMarkdown>
                  </div>
                </div>
              </div>
            )
          })}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>
    </div>
  )
}

// ── 하네스 상태 배너 ──────────────────────────────────────────────
function HarnessBanner({
  harnessStatus, onStop, lastEventMsg, t,
}: {
  harnessStatus: HarnessStatus | null
  onStop: () => void
  lastEventMsg: string | null
  t: Gt
}) {
  if (!harnessStatus || harnessStatus.status !== 'running') return null
  const secs = harnessStatus.lastCheck
    ? Math.round((Date.now() - new Date(harnessStatus.lastCheck).getTime()) / 1000)
    : null
  const lastCheckLabel = secs !== null ? t.issuesLastCheckAgo(secs) : t.issuesChecking
  return (
    <div className="mx-3 mt-2 rounded-lg border border-border bg-card px-3 py-2 shrink-0">
      <div className="flex items-center gap-2">
        <Zap className="w-3.5 h-3.5 text-blue-500 animate-pulse shrink-0" />
        <div className="flex-1 min-w-0">
          <p className="text-xs font-semibold text-foreground">
            {t.harnessRunning}
            <span className="ml-1.5 font-mono font-normal text-muted-foreground text-[10px]">
              {(harnessStatus.intervalMs ?? 15000) / 1000}{t.issuesIntervalSuffix}
            </span>
            {(harnessStatus.activeCount ?? 0) > 0 && (
              <span className="ml-1.5 text-orange-500 dark:text-orange-400 font-medium text-[10px]">
                {t.issuesActiveCount(harnessStatus.activeCount ?? 0)}
              </span>
            )}
          </p>
          {lastEventMsg && (
            <p className="text-[10px] text-muted-foreground truncate mt-0.5">{lastEventMsg}</p>
          )}
          <p className="text-[10px] text-muted-foreground/60 mt-0.5">{t.issuesLastCheck} {lastCheckLabel}</p>
        </div>
        <button
          type="button"
          onClick={onStop}
          className="shrink-0 p-1 rounded text-muted-foreground hover:text-red-500 transition-colors"
          title={t.issuesStopPolling}
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
  const { t }               = useLocale()

  const [ghRepo, setGhRepo]         = useState<GitHubRepo | null>(null)
  const [repoError, setRepoError]   = useState<string | null>(null)
  const [issues, setIssues]         = useState<GithubIssue[]>([])
  const [loading, setLoading]       = useState(false)
  const [selected, setSelected]     = useState<GithubIssue | null>(null)

  // 오른쪽 사이드 패널: 'detail' | 'create' | 'harness-setup' | null
  const [sidePanel, setSidePanel]   = useState<'detail' | 'create' | 'harness-setup' | null>(null)

  // 이슈 생성 폼
  const [newTitle, setNewTitle]     = useState('')
  const [newBody, setNewBody]       = useState('')
  const [creating, setCreating]     = useState(false)

  // 개별 이슈 ADK 처리 (수동)
  const [manualProcIssueKey, setManualProcIssueKey]   = useState<string | null>(null)
  const [manualProcEvents, setManualProcEvents]       = useState<string[]>([])
  const [manualProcStatus, setManualProcStatus]       = useState<'running' | 'done' | 'error' | null>(null)
  const manualSseRef = useRef<EventSource | null>(null)

  // WORKFLOW.md / Harness 상태
  const [workflowExists, setWorkflowExists]         = useState<boolean | null>(null)
  const [detectedIsJikiMe, setDetectedIsJikiMe]     = useState(false)
  const [harnessStatus, setHarnessStatus]           = useState<HarnessStatus | null>(null)
  const [harnessStarting, setHarnessStarting]       = useState(false)
  const [activeIssueNums, setActiveIssueNums]       = useState<Set<number>>(new Set())
  const [lastEventMsg, setLastEventMsg]             = useState<string | null>(null)
  const harnessSseRef = useRef<EventSource | null>(null)

  // Harness 설정 폼
  const [hSlug, setHSlug]                           = useState('')
  const [hLabel, setHLabel]                         = useState('jikime-todo')
  const [hWorkspaceRoot, setHWorkspaceRoot]         = useState('~/jikime_workspaces')
  const [hPort, setHPort]                           = useState(0)
  const [hMaxAgents, setHMaxAgents]                 = useState(3)
  const [hMode, setHMode]                           = useState<'basic' | 'jikime'>('basic')
  const [generatingWorkflow, setGeneratingWorkflow] = useState(false)

  const pat         = loadSettings().gitPat ?? ''
  const projectPath = activeProject?.path ?? ''

  // ── WORKFLOW.md 존재 여부 확인 ──────────────────────────────
  const checkWorkflow = useCallback(async () => {
    if (!projectPath) return
    try {
      const r = await fetch(getApiUrl(`/api/harness/check?projectPath=${encodeURIComponent(projectPath)}`))
      const d = await r.json() as { exists: boolean; isRunning: boolean; isJikiMe: boolean; slug: string }
      setWorkflowExists(d.exists)
      setDetectedIsJikiMe(d.isJikiMe)
      if (!d.exists && d.slug) {
        setHSlug(d.slug)
      }
      if (d.isJikiMe && !d.exists) setHMode('jikime')
      if (d.isRunning) {
        setHarnessStatus(prev => ({ ...prev, status: 'running' }))
      }
    } catch { setWorkflowExists(false) }
  }, [projectPath, getApiUrl])

  useEffect(() => { checkWorkflow() }, [checkWorkflow])

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

  // ── Harness 상태 조회 (WORKFLOW.md 존재 시) ────────────────
  useEffect(() => {
    if (!projectPath || workflowExists !== true) return
    fetch(getApiUrl(`/api/harness/status?projectPath=${encodeURIComponent(projectPath)}`))
      .then(r => r.json())
      .then((d: HarnessStatus) => setHarnessStatus(d))
      .catch(() => {})
  }, [projectPath, workflowExists, getApiUrl])

  // ── Harness SSE 연결/해제 ──────────────────────────────────
  const connectHarnessSSE = useCallback(() => {
    if (!projectPath) return
    harnessSseRef.current?.close()
    const sse = new EventSource(
      getApiUrl(`/api/harness/events?projectPath=${encodeURIComponent(projectPath)}`),
    )
    harnessSseRef.current = sse
    sse.onmessage = (e) => {
      const msg = JSON.parse(e.data) as HarnessEvent
      if (msg.type === 'tick') {
        setHarnessStatus(prev => ({
          ...prev,
          status: 'running',
          lastCheck: msg.lastCheck ?? prev?.lastCheck,
          activeCount: msg.activeCount ?? 0,
        }))
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

  // 하네스가 running이면 SSE 자동 연결
  useEffect(() => {
    if (harnessStatus?.status === 'running') {
      connectHarnessSSE()
    } else {
      harnessSseRef.current?.close()
    }
    return () => harnessSseRef.current?.close()
  }, [harnessStatus?.status, connectHarnessSSE])

  // ── 하네스 시작 ──────────────────────────────────────────
  const startHarness = async () => {
    if (!projectPath) return
    setHarnessStarting(true)
    try {
      const r = await fetch(getApiUrl('/api/harness/start'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ projectPath }),
      })
      const d = await r.json() as HarnessStatus & { error?: string }
      if (!r.ok) throw new Error(d.error ?? 'Failed')
      setHarnessStatus({ ...d, status: 'running' })
      connectHarnessSSE()
    } catch (e) {
      alert(`하네스 시작 실패: ${(e as Error).message}`)
    } finally {
      setHarnessStarting(false)
    }
  }

  // ── 하네스 중지 ──────────────────────────────────────────
  const stopHarness = async () => {
    harnessSseRef.current?.close()
    await fetch(getApiUrl(`/api/harness/stop?projectPath=${encodeURIComponent(projectPath)}`), { method: 'DELETE' })
    setHarnessStatus({ status: 'stopped' })
    setActiveIssueNums(new Set())
    setLastEventMsg(null)
  }

  // ── WORKFLOW.md 생성 ─────────────────────────────────────
  const generateWorkflow = async () => {
    if (!hSlug.trim()) return
    setGeneratingWorkflow(true)
    try {
      const r = await fetch(getApiUrl('/api/harness/init'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          projectPath,
          slug:          hSlug.trim(),
          label:         hLabel.trim() || 'jikime-todo',
          workspaceRoot: hWorkspaceRoot.trim() || '~/jikime_workspaces',
          port:          hPort,
          maxAgents:     hMaxAgents,
          mode:          hMode,
        }),
      })
      const d = await r.json() as { success?: boolean; error?: string }
      if (!r.ok) throw new Error(d.error ?? 'Failed')
      setWorkflowExists(true)
      setSidePanel(null)
      // WORKFLOW.md 생성 직후 하네스 자동 시작
      await startHarness()
    } catch (e) {
      alert(`WORKFLOW.md 생성 실패: ${(e as Error).message}`)
    } finally {
      setGeneratingWorkflow(false)
    }
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
      setNewTitle(''); setNewBody(''); setSidePanel(null)
      await loadIssues()
      // 하네스 가동 중이면 즉시 폴 트리거
      if (isHarnessRunning) {
        fetch(getApiUrl(`/api/harness/refresh?projectPath=${encodeURIComponent(projectPath)}`), { method: 'POST' })
          .catch(() => {})
      }
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
        {t.git.issuesSelectProject}
      </div>
    )
  }

  const isHarnessRunning = harnessStatus?.status === 'running'
  const isManualProcessingSelected =
    selected !== null && ghRepo !== null &&
    manualProcIssueKey === `${ghRepo.owner}/${ghRepo.repo}#${selected.number}`
  const isAutoProcessingSelected =
    selected !== null && activeIssueNums.has(selected.number)

  // 패널 열기 헬퍼
  const openDetail = (issue: GithubIssue) => {
    setSelected(issue)
    setSidePanel('detail')
  }
  const openCreate = () => {
    setSelected(null)
    setNewTitle(''); setNewBody('')
    setSidePanel('create')
  }
  const openHarnessSetup = () => {
    setSelected(null)
    setSidePanel('harness-setup')
  }
  const closePanel = () => {
    setSidePanel(null)
    setSelected(null)
  }

  return (
    <div className="flex flex-col h-full overflow-hidden bg-background">

      {/* ── 헤더 ── */}
      <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border/50 shrink-0">
        <span className="text-xs font-semibold text-foreground/80 flex-1 truncate">
          {ghRepo ? `${ghRepo.owner}/${ghRepo.repo}` : t.git.issues}
          {issues.length > 0 && <span className="ml-1 text-muted-foreground">({issues.length})</span>}
        </span>
        <Button variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={loadIssues} disabled={loading || !pat}>
          <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
        </Button>
        <Button
          variant="outline" size="sm"
          className={cn('h-6 px-2 text-xs gap-1 shrink-0', sidePanel === 'create' && 'bg-muted')}
          onClick={sidePanel === 'create' ? closePanel : openCreate}
          disabled={!ghRepo || !pat}
        >
          <Plus className="w-3 h-3" /> New
        </Button>

        {/* Harness 설정 버튼 (WORKFLOW.md 없음) */}
        {workflowExists === false && (
          <button
            type="button"
            onClick={sidePanel === 'harness-setup' ? closePanel : openHarnessSetup}
            className={cn(
              'inline-flex items-center gap-1 h-6 px-2 rounded-md text-xs shrink-0 border transition-colors',
              'border-amber-500/50 text-amber-400 bg-transparent',
              'hover:border-amber-400',
              sidePanel === 'harness-setup' && 'border-amber-400/70',
            )}
          >
            <Settings className="w-3 h-3" />
            {t.git.harnessSetup}
          </button>
        )}

        {/* Harness 시작/중지 버튼 (WORKFLOW.md 있음) */}
        {workflowExists === true && (
          isHarnessRunning ? (
            <button
              type="button"
              onClick={stopHarness}
              className="inline-flex items-center gap-1 h-6 px-2 rounded-md text-xs shrink-0 border transition-colors border-red-500/50 text-red-400 bg-transparent hover:border-red-400"
            >
              <Square className="w-3 h-3 fill-current" />
              {t.git.issuesStop}
            </button>
          ) : (
            <button
              type="button"
              onClick={startHarness}
              disabled={harnessStarting}
              className="inline-flex items-center gap-1 h-6 px-2 rounded-md text-xs shrink-0 border transition-colors border-blue-500/50 text-blue-400 bg-transparent hover:border-blue-400 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {harnessStarting ? <Loader2 className="w-3 h-3 animate-spin" /> : <Zap className="w-3 h-3" />}
              {harnessStarting ? t.git.issuesStarting : t.git.harnessStart}
            </button>
          )
        )}
      </div>

      {/* ── 경고 ── */}
      {!pat && (
        <div className="mx-3 mt-2 p-2 rounded-lg bg-amber-50 border border-amber-200 text-amber-700 dark:bg-amber-950/30 dark:border-amber-700/40 dark:text-amber-300 text-xs shrink-0">
          {t.git.issuesNoPat}
        </div>
      )}
      {pat && repoError && (
        <div className="mx-3 mt-2 p-2 rounded-lg bg-red-50 border border-red-200 text-red-700 dark:bg-red-950/30 dark:border-red-700/40 dark:text-red-300 text-xs shrink-0">
          <p className="font-medium">{t.git.issuesRepoError}</p>
          <p className="mt-0.5 font-mono text-[11px] opacity-70 break-all">{repoError}</p>
        </div>
      )}

      {/* ── WORKFLOW.md 없음 안내 배너 ── */}
      {workflowExists === false && ghRepo && (
        <div className="mx-3 mt-2 flex items-center gap-2 p-2 rounded-lg bg-amber-50 border border-amber-200 text-amber-700 dark:bg-amber-950/20 dark:border-amber-700/30 dark:text-amber-300/80 text-xs shrink-0">
          <Settings className="w-3.5 h-3.5 shrink-0 text-amber-400" />
          <span className="flex-1">{t.git.harnessSetupNoWorkflow}</span>
          <button
            type="button"
            onClick={openHarnessSetup}
            className="text-amber-400 hover:text-amber-300 underline underline-offset-2 whitespace-nowrap"
          >
            {t.git.harnessSetup}
          </button>
        </div>
      )}

      {/* ── 하네스 가동 배너 ── */}
      <HarnessBanner harnessStatus={harnessStatus} onStop={stopHarness} lastEventMsg={lastEventMsg} t={t.git} />

      {/* ── 메인 영역: 목록 + 오른쪽 사이드 패널 ── */}
      <div className="flex flex-1 min-h-0 overflow-hidden">

        {/* ── 이슈 목록 (왼쪽) ── */}
        <ScrollArea className="flex-1 min-w-0 border-r border-border">
          {loading ? (
            <div className="flex items-center justify-center py-8 gap-2 text-xs text-muted-foreground">
              <Loader2 className="w-4 h-4 animate-spin" /> {t.git.loading}
            </div>
          ) : issues.length === 0 ? (
            <p className="text-xs text-muted-foreground/50 text-center py-8">
              {!pat ? t.git.issuesPatHint : t.git.issuesNoIssues}
            </p>
          ) : (() => {
            const todoIssues  = issues.filter(i => i.state === 'open' && i.labels.includes('jikime-todo'))
            const otherIssues = issues.filter(i => !(i.state === 'open' && i.labels.includes('jikime-todo')))
            const renderCard = (issue: GithubIssue) => (
              <IssueCard
                key={issue.number}
                issue={issue}
                t={t.git}
                selected={sidePanel === 'detail' && selected?.number === issue.number}
                processing={activeIssueNums.has(issue.number) || (
                  manualProcIssueKey !== null && ghRepo !== null &&
                  manualProcIssueKey === `${ghRepo.owner}/${ghRepo.repo}#${issue.number}` &&
                  manualProcStatus === 'running'
                )}
                onClick={() => openDetail(issue)}
              />
            )
            return (
              <>
                {todoIssues.length > 0 && (
                  <>
                    <div className="flex items-center gap-1.5 px-3 py-1 bg-blue-50 border-b border-blue-200 dark:bg-blue-950/20 dark:border-blue-500/20">
                      <Radio className="w-3 h-3 text-blue-500 dark:text-blue-400" />
                      <span className="text-[10px] font-semibold text-blue-600 dark:text-blue-400 uppercase tracking-wider">
                        {t.git.issuesPollingTarget} ({todoIssues.length})
                      </span>
                    </div>
                    {todoIssues.map(renderCard)}
                  </>
                )}
                {otherIssues.length > 0 && (
                  <>
                    <div className="flex items-center gap-1.5 px-3 py-1 bg-muted/30 border-b border-border/50">
                      <CircleDot className="w-3 h-3 text-muted-foreground/50" />
                      <span className="text-[10px] font-semibold text-muted-foreground/60 uppercase tracking-wider">
                        {t.git.issuesOther} ({otherIssues.length})
                      </span>
                    </div>
                    {otherIssues.map(renderCard)}
                  </>
                )}
              </>
            )
          })()}
        </ScrollArea>

        {/* ── 오른쪽 사이드 패널 (슬라이드인) ── */}
        <div className={cn(
          'flex flex-col shrink-0 h-full overflow-hidden bg-card transition-all duration-200',
          sidePanel === 'create' ? 'w-[26rem]' : sidePanel ? 'w-96' : 'w-0'
        )}>
          {sidePanel && (
            <div className="flex flex-col h-full">
            {/* ─ 패널 헤더 ─ */}
            <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border/50 shrink-0">
              {sidePanel === 'harness-setup' ? (
                <>
                  <Zap className="w-3 h-3 text-amber-400 shrink-0" />
                  <span className="text-[11px] font-semibold text-foreground/80 flex-1">
                    {t.git.harnessSetupTitle}
                  </span>
                </>
              ) : sidePanel === 'detail' && selected ? (
                <>
                  <StateBadge state={selected.state} />
                  <span className="text-[11px] font-semibold text-foreground/80 flex-1 truncate">
                    #{selected.number}
                  </span>
                  <a
                    href={selected.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-muted-foreground hover:text-foreground transition-colors shrink-0"
                  >
                    <ExternalLink className="w-3 h-3" />
                  </a>
                </>
              ) : (
                <span className="text-[11px] font-semibold text-foreground/80 flex-1">
                  {t.git.issuesNewIssue}
                </span>
              )}
              <button
                type="button"
                onClick={closePanel}
                className="text-muted-foreground hover:text-foreground transition-colors rounded hover:bg-muted shrink-0"
              >
                <X className="w-3 h-3" />
              </button>
            </div>

            {/* ─ 패널 본문 ─ */}
            {sidePanel === 'harness-setup' ? (
              /* ── Harness 설정 폼 ── */
              <ScrollArea className="flex-1 min-h-0">
                <div className="px-2.5 py-3 space-y-3">

                  {/* 모드 선택 */}
                  <div>
                    <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                      {t.git.harnessMode}
                    </label>
                    <div className="flex gap-1.5">
                      <button
                        type="button"
                        onClick={() => setHMode('basic')}
                        className={cn(
                          'flex-1 px-2 py-1.5 rounded text-[11px] border transition-colors',
                          hMode === 'basic'
                            ? 'bg-primary/15 border-primary/40 text-primary'
                            : 'border-border text-muted-foreground hover:bg-muted/50',
                        )}
                      >
                        {t.git.harnessModeBasic}
                      </button>
                      <button
                        type="button"
                        onClick={() => setHMode('jikime')}
                        className={cn(
                          'flex-1 px-2 py-1.5 rounded text-[11px] border transition-colors',
                          hMode === 'jikime'
                            ? 'bg-blue-500/15 border-blue-500/40 text-blue-400'
                            : 'border-border text-muted-foreground hover:bg-muted/50',
                        )}
                      >
                        {t.git.harnessModeJikiMe}
                        {detectedIsJikiMe && (
                          <span className="block text-[9px] text-blue-400/70">{t.git.harnessAutoDetected}</span>
                        )}
                      </button>
                    </div>
                  </div>

                  {/* owner/repo */}
                  <div>
                    <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                      {t.git.harnessSlug}
                    </label>
                    <Input
                      value={hSlug}
                      onChange={e => setHSlug(e.target.value)}
                      placeholder="owner/repo"
                      className="h-7 text-xs"
                    />
                  </div>

                  {/* 라벨 */}
                  <div>
                    <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                      {t.git.harnessLabel}
                    </label>
                    <Input
                      value={hLabel}
                      onChange={e => setHLabel(e.target.value)}
                      placeholder="jikime-todo"
                      className="h-7 text-xs"
                    />
                  </div>

                  {/* 워크스페이스 경로 */}
                  <div>
                    <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                      {t.git.harnessWorkspaceRoot}
                    </label>
                    <Input
                      value={hWorkspaceRoot}
                      onChange={e => setHWorkspaceRoot(e.target.value)}
                      placeholder="~/jikime_workspaces"
                      className="h-7 text-xs"
                    />
                  </div>

                  {/* 포트 / 최대 에이전트 (나란히) */}
                  <div className="flex gap-2">
                    <div className="flex-1">
                      <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                        {t.git.harnessPort}
                      </label>
                      <Input
                        type="number"
                        min={0}
                        max={65535}
                        value={hPort}
                        onChange={e => setHPort(Number(e.target.value))}
                        className="h-7 text-xs"
                      />
                    </div>
                    <div className="flex-1">
                      <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                        {t.git.harnessMaxAgents}
                      </label>
                      <Input
                        type="number"
                        min={1}
                        max={10}
                        value={hMaxAgents}
                        onChange={e => setHMaxAgents(Number(e.target.value))}
                        className="h-7 text-xs"
                      />
                    </div>
                  </div>

                  {/* 생성 버튼 */}
                  <Button
                    size="sm"
                    className="w-full h-7 text-xs gap-1.5 bg-amber-700 hover:bg-amber-600 text-white"
                    onClick={generateWorkflow}
                    disabled={!hSlug.trim() || generatingWorkflow}
                  >
                    {generatingWorkflow
                      ? <><Loader2 className="w-3 h-3 animate-spin" /> {t.git.harnessGenerating}</>
                      : <><Zap className="w-3 h-3" /> {t.git.harnessGenerate}</>
                    }
                  </Button>
                </div>
              </ScrollArea>

            ) : sidePanel === 'detail' && selected ? (
              /* ── 이슈 상세 ── */
              <ScrollArea className="flex-1 min-h-0">
                <div className="px-2.5 py-2 space-y-2">
                  <p className="text-xs font-medium text-foreground/90 leading-relaxed">
                    {selected.title}
                  </p>
                  {selected.labels.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {selected.labels.map(l => <LabelBadge key={l} name={l} />)}
                    </div>
                  )}
                  {selected.body && (
                    <p className="text-[11px] text-muted-foreground leading-relaxed whitespace-pre-wrap break-words border-t border-border/50 pt-2">
                      {selected.body}
                    </p>
                  )}
                  {isAutoProcessingSelected && (
                    <HarnessWorkerStream
                      projectPath={projectPath}
                      issueNumber={selected.number}
                      getApiUrl={getApiUrl}
                    />
                  )}
                  {selected.state === 'open' && selected.labels.includes('jikime-todo') && !isAutoProcessingSelected && (
                    isManualProcessingSelected && manualProcStatus === 'running' ? (
                      <Button
                        size="sm" variant="outline"
                        className="w-full h-7 text-xs gap-1.5 border-red-500/40 text-red-400 hover:bg-red-950/30"
                        onClick={stopManualProcess}
                      >
                        <Square className="w-3 h-3 fill-current" /> {t.git.issuesStopManual}
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        className="w-full h-7 text-xs gap-1.5 bg-emerald-700 hover:bg-emerald-600 text-white"
                        onClick={() => startManualProcess(selected)}
                        disabled={manualProcStatus === 'running'}
                      >
                        <Play className="w-3 h-3 fill-current" /> {t.git.issuesManualProcess}
                      </Button>
                    )
                  )}
                </div>
                {isManualProcessingSelected && (
                  <ProcessingLog events={manualProcEvents} status={manualProcStatus} t={t.git} />
                )}
              </ScrollArea>

            ) : (
              /* ── 이슈 생성 ── */
              <div className="flex flex-col flex-1 min-h-0 px-3 py-3 gap-2.5">
                <Input
                  value={newTitle}
                  onChange={e => setNewTitle(e.target.value)}
                  placeholder={t.git.issuesTitlePlaceholder}
                  className="h-8 text-xs shrink-0"
                  onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) createIssue() }}
                  autoFocus
                />
                <Textarea
                  value={newBody}
                  onChange={e => setNewBody(e.target.value)}
                  placeholder={t.git.issuesBodyPlaceholder}
                  className="text-xs resize-none flex-1 min-h-0"
                />
                <Button
                  size="sm"
                  className="w-full h-7 text-xs bg-blue-700 hover:bg-blue-600 text-white shrink-0"
                  onClick={createIssue}
                  disabled={!newTitle.trim() || creating}
                >
                  {creating ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : <Plus className="w-3 h-3 mr-1" />}
                  {creating ? t.git.issuesCreating : t.git.issuesCreate}
                </Button>
              </div>
            )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
