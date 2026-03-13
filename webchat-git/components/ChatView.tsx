'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import {
  Send, Bot, User, ExternalLink, Play, Square, Loader2,
  ChevronDown, ChevronUp, Terminal, Wrench, CheckCircle2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import MarkdownRenderer from './MarkdownRenderer'
import IssueList, { type GithubIssue, type LiveOverlay } from './IssueList'

interface Project {
  id: string
  name: string
  repo: string
  cwd: string
  port: number
  pid: number | null
  status: 'running' | 'stopped'
}

interface ChatMessage {
  id: string
  role: 'user' | 'system'
  text: string
  issueUrl?: string
  issueNumber?: number
}

interface ParsedEvent {
  kind: 'text' | 'tool_call' | 'tool_result' | 'raw'
  content: string
  toolName?: string
  _ts?: string  // dedup key: Timestamp from RecentEvents
}

interface LiveIssue {
  identifier: string   // "owner/repo#N"
  turnCount: number
  lastEvent: string
  lastMessage: string  // 중복 감지용 — 새 값이면 events에 추가
  events: ParsedEvent[]
  tokens: number
  startedAt: string
}

interface ServeState {
  generated_at: string
  counts: { running: number; retrying: number }
  running: Array<{
    IssueID: string
    IssueIdentifier: string
    State: string
    TurnCount: number
    LastEvent: string
    LastMessage: string
    StartedAt: string
    LastEventAt: string | null
    Tokens: { InputTokens: number; OutputTokens: number; TotalTokens: number }
    RecentEvents: Array<{ IssueIdentifier: string; Message: string; Timestamp: string }> | null
  }>
  retrying: Array<{
    IssueID: string
    Identifier: string
    Attempt: number
    DueAt: string | null
    Error: string
  }>
  jikime_totals: { InputTokens: number; OutputTokens: number; TotalTokens: number; SecondsRunning: number }
}

interface Props {
  project: Project
  onProjectUpdate: (p: Project) => void
}

// stream-json 한 줄을 파싱해서 표시할 내용 추출
function parseStreamEvent(raw: string): ParsedEvent | null {
  if (!raw?.trim()) return null
  try {
    const ev = JSON.parse(raw)
    // assistant 이벤트 — tool_use / text 블록
    if (ev.type === 'assistant' && Array.isArray(ev.message?.content)) {
      for (const block of ev.message.content) {
        if (block.type === 'tool_use') {
          const input = formatToolInput(block.name, block.input)
          return { kind: 'tool_call', toolName: block.name, content: input }
        }
        if (block.type === 'text' && block.text) {
          return { kind: 'text', content: block.text }
        }
      }
    }
    // result 이벤트 — 최종 답변
    if (ev.type === 'result' && ev.result) {
      return { kind: 'text', content: ev.result }
    }
    // user 이벤트 — tool_result
    if (ev.type === 'user' && Array.isArray(ev.message?.content)) {
      for (const block of ev.message.content) {
        if (block.type === 'tool_result') {
          const text = extractToolResult(block.content)
          if (text) return { kind: 'tool_result', content: text }
        }
      }
    }
  } catch { /* not JSON */ }
  // JSON이 아니거나 알 수 없는 형식 → 그대로 표시
  return { kind: 'raw', content: raw }
}

function formatToolInput(name: string, input: Record<string, unknown>): string {
  if (!input) return ''
  if (name === 'Bash' && input.command) return String(input.command)
  if (input.file_path) return String(input.file_path)
  if (input.path) return String(input.path)
  return JSON.stringify(input, null, 2)
}

function extractToolResult(content: unknown): string {
  if (typeof content === 'string') return content.slice(0, 300)
  if (Array.isArray(content)) {
    return content.map(c => (typeof c === 'string' ? c : (c as { text?: string }).text ?? '')).join('\n').slice(0, 300)
  }
  return ''
}

// ─── 단일 이벤트 렌더 ─────────────────────────────────────────────────────────

function EventRow({ event }: { event: ParsedEvent }) {
  if (event.kind === 'tool_call') {
    return (
      <div className="mt-2">
        <div className="flex items-center gap-1.5 text-[11px] text-amber-400 mb-1">
          <Wrench className="w-3 h-3" />
          <span className="font-mono">{event.toolName}</span>
        </div>
        <pre className="text-[11px] text-zinc-400 bg-zinc-900 rounded-lg px-3 py-2 overflow-x-auto whitespace-pre-wrap break-all font-mono max-h-32 overflow-y-auto">
          {event.content}
        </pre>
      </div>
    )
  }
  if (event.kind === 'tool_result') {
    return (
      <div className="mt-2">
        <div className="flex items-center gap-1.5 text-[11px] text-zinc-500 mb-1">
          <CheckCircle2 className="w-3 h-3 text-emerald-500" />
          <span>결과</span>
        </div>
        <pre className="text-[11px] text-zinc-500 bg-zinc-900 rounded-lg px-3 py-2 overflow-x-auto whitespace-pre-wrap break-all font-mono max-h-24 overflow-y-auto">
          {event.content}
        </pre>
      </div>
    )
  }
  // text / raw → 마크다운 렌더
  return <div className="mt-2"><MarkdownRenderer className="text-sm">{event.content}</MarkdownRenderer></div>
}

// ─── 라이브 이슈 카드 ─────────────────────────────────────────────────────────

function LiveIssueCard({ live }: { live: LiveIssue }) {
  const [open, setOpen] = useState(true)
  const issueUrl = `https://github.com/${live.identifier.replace('#', '/issues/')}`

  return (
    <div className="flex gap-3">
      <div className="w-7 h-7 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center shrink-0 mt-0.5">
        <Loader2 className="w-3.5 h-3.5 text-emerald-400 animate-spin" />
      </div>
      <div className="flex-1 min-w-0 rounded-2xl rounded-tl-sm border border-zinc-700/50 bg-zinc-800/60 overflow-hidden">
        {/* 헤더 */}
        <button
          onClick={() => setOpen(v => !v)}
          className="w-full flex items-center gap-2 px-4 py-2.5 text-left hover:bg-zinc-700/30 transition-colors"
        >
          <a
            href={issueUrl}
            target="_blank" rel="noopener noreferrer"
            className="text-xs font-mono text-blue-400 hover:underline flex items-center gap-1 shrink-0"
            onClick={e => e.stopPropagation()}
          >
            {live.identifier} <ExternalLink className="w-2.5 h-2.5" />
          </a>
          <span className="flex-1" />
          <span className="text-[10px] text-zinc-500">턴 {live.turnCount}</span>
          <span className="text-[10px] text-zinc-600">{(live.tokens || 0).toLocaleString()} tokens</span>
          {open ? <ChevronUp className="w-3 h-3 text-zinc-600" /> : <ChevronDown className="w-3 h-3 text-zinc-600" />}
        </button>

        {/* 누적된 이벤트 목록 */}
        {open && live.events.length > 0 && (
          <div className="px-4 pb-3 border-t border-zinc-700/40 divide-y divide-zinc-800/50">
            {live.events.map((event, i) => (
              <EventRow key={i} event={event} />
            ))}
          </div>
        )}
        {open && live.events.length === 0 && (
          <div className="px-4 py-2.5 border-t border-zinc-700/40 flex items-center gap-2 text-[11px] text-zinc-600">
            <Loader2 className="w-3 h-3 animate-spin text-emerald-500" />
            작업 준비 중...
          </div>
        )}
      </div>
    </div>
  )
}

// ─── 완료 카드 ───────────────────────────────────────────────────────────────

function DoneCard({ identifier }: { identifier: string }) {
  const issueUrl = `https://github.com/${identifier.replace('#', '/issues/')}`
  return (
    <div className="flex gap-3">
      <div className="w-7 h-7 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center shrink-0 mt-0.5">
        <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
      </div>
      <div className="rounded-2xl rounded-tl-sm border border-emerald-800/40 bg-emerald-950/20 px-4 py-2.5 text-sm text-emerald-400 flex items-center gap-2">
        <a href={issueUrl} target="_blank" rel="noopener noreferrer"
          className="font-mono text-xs hover:underline">{identifier}</a>
        작업 완료
      </div>
    </div>
  )
}

// ─── 메인 컴포넌트 ────────────────────────────────────────────────────────────

export default function ChatView({ project, onProjectUpdate }: Props) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [toggling, setToggling] = useState(false)
  const [serveState, setServeState] = useState<ServeState | null>(null)
  const [sseConnected, setSseConnected] = useState(false)

  // 라이브 이슈 상태
  const [liveIssues, setLiveIssues] = useState<Map<string, LiveIssue>>(new Map())
  const [doneIssues, setDoneIssues] = useState<string[]>([])

  // GitHub 이슈 목록
  const [githubIssues, setGithubIssues] = useState<GithubIssue[]>([])
  const [issuesLoading, setIssuesLoading] = useState(false)

  // 로그 패널
  const [logLines, setLogLines] = useState<string[]>([])
  const [logOpen, setLogOpen] = useState(false)
  const logEndRef = useRef<HTMLDivElement>(null)
  const logSseRef = useRef<EventSource | null>(null)

  const prevRunningRef = useRef<Set<string>>(new Set())
  const scrollRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const sseRef = useRef<EventSource | null>(null)

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages, liveIssues, doneIssues])

  // ─── GitHub 이슈 목록 ────────────────────────────────────────────────────
  const fetchGithubIssues = useCallback(async () => {
    setIssuesLoading(true)
    try {
      const res = await fetch(`/api/issue?projectId=${project.id}`)
      if (res.ok) setGithubIssues(await res.json() as GithubIssue[])
    } finally {
      setIssuesLoading(false)
    }
  }, [project.id])

  useEffect(() => { fetchGithubIssues() }, [fetchGithubIssues])

  useEffect(() => {
    if (logOpen) logEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logLines, logOpen])

  // ─── State SSE ──────────────────────────────────────────────────────────
  const connectSSE = useCallback(() => {
    if (sseRef.current) sseRef.current.close()
    if (project.status !== 'running') return

    const es = new EventSource(`/api/state?projectId=${project.id}`)
    sseRef.current = es
    es.onopen = () => setSseConnected(true)

    es.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data) as { type: string; data?: ServeState; running?: boolean }
        if (msg.type === 'state' && msg.data) {
          const state = msg.data
          setServeState(state)

          // 라이브 이슈 업데이트 (RecentEvents 기반 누적)
          setLiveIssues(prev => {
            const next = new Map(prev)
            state.running.forEach(r => {
              const existing = next.get(r.IssueIdentifier)
              const seenTs = new Set(existing?.events.map(e => e._ts) ?? [])

              // RecentEvents가 있으면 서버 누적 이벤트를 merge
              const newFromServer: ParsedEvent[] = []
              if (r.RecentEvents && r.RecentEvents.length > 0) {
                for (const ev of r.RecentEvents) {
                  if (!ev.Message || seenTs.has(ev.Timestamp)) continue
                  // 백엔드 summarizeEvent()가 이미 요약한 문자열 — 직접 ParsedEvent 생성
                  const msg = ev.Message
                  let parsed: ParsedEvent
                  if (msg.startsWith('🔧 ')) {
                    const colonIdx = msg.indexOf(': ')
                    const toolName = colonIdx > 0 ? msg.slice(2, colonIdx) : msg.slice(2)
                    const content = colonIdx > 0 ? msg.slice(colonIdx + 2) : ''
                    parsed = { kind: 'tool_call', toolName, content, _ts: ev.Timestamp }
                  } else if (msg === '✓ tool result received') {
                    parsed = { kind: 'tool_result', content: msg, _ts: ev.Timestamp }
                  } else {
                    parsed = { kind: 'text', content: msg, _ts: ev.Timestamp }
                  }
                  newFromServer.push(parsed)
                }
              } else if (r.LastMessage && r.LastMessage !== existing?.lastMessage) {
                // fallback: RecentEvents 없으면 기존 방식
                const parsed = parseStreamEvent(r.LastMessage)
                if (parsed) newFromServer.push({ ...parsed, _ts: r.LastEventAt ?? r.StartedAt })
              }

              const mergedEvents = [...(existing?.events ?? []), ...newFromServer]

              next.set(r.IssueIdentifier, {
                identifier: r.IssueIdentifier,
                turnCount: r.TurnCount,
                lastEvent: r.LastEvent,
                lastMessage: r.LastMessage,
                events: mergedEvents,
                tokens: r.Tokens?.TotalTokens ?? 0,
                startedAt: r.StartedAt,
              })
            })
            return next
          })

          // 완료 감지: 이전 running에서 사라진 것
          const currentIds = new Set([
            ...state.running.map(r => r.IssueIdentifier),
            ...state.retrying.map(r => r.Identifier),
          ])
          prevRunningRef.current.forEach(id => {
            if (!currentIds.has(id)) {
              setLiveIssues(prev => { const n = new Map(prev); n.delete(id); return n })
              setDoneIssues(prev => prev.includes(id) ? prev : [...prev, id])
            }
          })
          prevRunningRef.current = currentIds
        }
        if (msg.type === 'status' && msg.running === false) {
          setSseConnected(false)
          es.close()
        }
      } catch { /* */ }
    }
    es.onerror = () => setSseConnected(false)
  }, [project.id, project.status])

  useEffect(() => {
    connectSSE()
    return () => sseRef.current?.close()
  }, [connectSSE])

  // ─── Log SSE ────────────────────────────────────────────────────────────
  const connectLogSSE = useCallback(() => {
    if (logSseRef.current) logSseRef.current.close()
    if (project.status !== 'running') return
    setLogLines([])
    const es = new EventSource(`/api/logs?projectId=${project.id}`)
    logSseRef.current = es
    es.onmessage = (e) => {
      try {
        const { line } = JSON.parse(e.data) as { line: string }

        // agent event 로그 라인 → LiveIssueCard 실시간 업데이트
        // 형식: level=INFO msg="agent event" type=notification issue_identifier=owner/repo#N message="..."
        if (line.includes('msg="agent event"') && line.includes('type=notification')) {
          const identMatch = line.match(/issue_identifier=(\S+)/)
          const msgMatch = line.match(/message="((?:[^"\\]|\\.)*)"/)
          if (identMatch && msgMatch) {
            const identifier = identMatch[1]
            const rawMsg = msgMatch[1].replace(/\\"/g, '"')
            const ts = new Date().toISOString()
            let parsed: ParsedEvent
            if (rawMsg.startsWith('🔧 ')) {
              const colonIdx = rawMsg.indexOf(': ')
              parsed = {
                kind: 'tool_call',
                toolName: colonIdx > 0 ? rawMsg.slice(2, colonIdx) : rawMsg.slice(2),
                content: colonIdx > 0 ? rawMsg.slice(colonIdx + 2) : '',
                _ts: ts,
              }
            } else if (rawMsg === '✓ tool result received') {
              parsed = { kind: 'tool_result', content: rawMsg, _ts: ts }
            } else {
              parsed = { kind: 'text', content: rawMsg, _ts: ts }
            }
            setLiveIssues(prev => {
              const next = new Map(prev)
              const existing = next.get(identifier)
              next.set(identifier, {
                identifier,
                turnCount: existing?.turnCount ?? 0,
                lastEvent: 'notification',
                lastMessage: rawMsg,
                events: [...(existing?.events ?? []), parsed],
                tokens: existing?.tokens ?? 0,
                startedAt: existing?.startedAt ?? ts,
              })
              return next
            })
          }
        }

        const formatted = formatLogLine(line)
        if (formatted) setLogLines(prev => [...prev, formatted])
      } catch { /* */ }
    }
    es.onerror = () => es.close()
  }, [project.id, project.status])

  useEffect(() => {
    connectLogSSE()
    return () => logSseRef.current?.close()
  }, [connectLogSSE])

  // ─── Serve 시작/중지 ────────────────────────────────────────────────────
  const toggleServe = async () => {
    setToggling(true)
    try {
      const action = project.status === 'running' ? 'stop' : 'start'
      const res = await fetch('/api/projects', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: project.id, action }),
      })
      const data = await res.json() as { ok?: boolean; alreadyRunning?: boolean; error?: string }

      // 이미 실행 중 → running 상태로 전환 (정상 처리)
      if (res.status === 409 && data.alreadyRunning) {
        onProjectUpdate({ ...project, status: 'running' })
        setTimeout(() => { connectSSE(); connectLogSSE() }, 500)
        addSystem(`jikime serve 이미 실행 중 — 연결됨 (포트 ${project.port})`)
        return
      }

      if (res.ok) {
        const updated = { ...project, status: action === 'start' ? 'running' : 'stopped' } as Project
        onProjectUpdate(updated)
        if (action === 'start') {
          setLogLines([])
          setLiveIssues(new Map())
          setDoneIssues([])
          setTimeout(() => { connectSSE(); connectLogSSE() }, 1500)
          addSystem(`jikime serve 시작됨 (포트 ${project.port})`)
        } else {
          setServeState(null)
          setSseConnected(false)
          setLiveIssues(new Map())
          addSystem('jikime serve 중지됨')
        }
      }
    } finally {
      setToggling(false)
    }
  }

  const addSystem = (text: string) => {
    setMessages(prev => [...prev, { id: `sys-${Date.now()}`, role: 'system', text }])
  }

  // ─── 메시지 전송 ─────────────────────────────────────────────────────────
  const submit = useCallback(async () => {
    const trimmed = input.trim()
    if (!trimmed || sending) return

    setInput('')
    setSending(true)
    if (inputRef.current) inputRef.current.style.height = '42px'

    setMessages(prev => [...prev, { id: `user-${Date.now()}`, role: 'user', text: trimmed }])

    try {
      const title = trimmed.length > 72 ? trimmed.slice(0, 72).trimEnd() + '...' : trimmed
      const res = await fetch('/api/issue', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ projectId: project.id, title, body: trimmed }),
      })
      const data = await res.json() as { number?: number; url?: string; error?: string }

      if (!res.ok) {
        addSystem(`❌ 이슈 생성 실패: ${data.error}`)
      } else {
        setMessages(prev => [...prev, {
          id: `sys-${Date.now()}`,
          role: 'system',
          text: `GitHub 이슈 #${data.number} 생성됨 — jikime serve가 작업을 시작해요`,
          issueUrl: data.url,
          issueNumber: data.number,
        }])
        fetch(`/api/state?projectId=${project.id}`, { method: 'POST' }).catch(() => {})
        fetchGithubIssues()
      }
    } catch {
      addSystem('❌ 네트워크 오류')
    } finally {
      setSending(false)
    }
  }, [input, sending, project.id, fetchGithubIssues])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); submit() }
  }

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`
  }

  const [issueListOpen, setIssueListOpen] = useState(false)

  const overlay: LiveOverlay = {
    running: new Map(
      (serveState?.running ?? []).map(r => [
        r.IssueIdentifier,
        { turnCount: r.TurnCount, tokens: r.Tokens?.TotalTokens ?? 0, lastMessage: r.LastMessage },
      ])
    ),
    retrying: new Map(
      (serveState?.retrying ?? []).map(r => [
        r.Identifier,
        { attempt: r.Attempt, error: r.Error },
      ])
    ),
  }

  return (
    <div className="flex flex-col h-full">
      {/* 헤더 */}
      <div className="shrink-0 px-4 py-2.5 bg-zinc-900 border-b border-zinc-800 flex items-center gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <p className="text-sm font-semibold text-zinc-100 truncate">{project.name}</p>
            {/* SSE 연결 상태 인디케이터 */}
            {project.status === 'running' && (
              <span className={cn(
                'w-1.5 h-1.5 rounded-full shrink-0',
                sseConnected ? 'bg-emerald-500' : 'bg-zinc-600'
              )} title={sseConnected ? '연결됨' : '연결 중...'} />
            )}
            {/* 작업 중 이슈 수 */}
            {(serveState?.counts?.running ?? 0) > 0 && (
              <span className="text-[10px] font-mono text-emerald-500 bg-emerald-950/50 px-1.5 py-0.5 rounded">
                {serveState!.counts.running} 작업 중
              </span>
            )}
            {(serveState?.counts?.retrying ?? 0) > 0 && (
              <span className="text-[10px] font-mono text-amber-500 bg-amber-950/50 px-1.5 py-0.5 rounded">
                {serveState!.counts.retrying} 재시도
              </span>
            )}
          </div>
          <a
            href={`https://github.com/${project.repo}`}
            target="_blank" rel="noopener noreferrer"
            className="text-xs text-zinc-500 hover:text-blue-400 flex items-center gap-1 w-fit"
          >
            {project.repo} <ExternalLink className="w-2.5 h-2.5" />
          </a>
        </div>

        {/* 이슈 목록 토글 버튼 */}
        <button
          onClick={() => setIssueListOpen(v => !v)}
          className="flex items-center gap-1 px-2 py-1 rounded text-[11px] text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800 transition-colors"
        >
          이슈
          {githubIssues.length > 0 && (
            <span className="bg-zinc-700 text-zinc-300 rounded px-1 text-[10px]">{githubIssues.length}</span>
          )}
          {issueListOpen ? <ChevronUp className="w-3 h-3" /> : <ChevronDown className="w-3 h-3" />}
        </button>

        <span className="text-[11px] font-mono text-zinc-700 hidden sm:block">:{project.port}</span>
        <button
          onClick={toggleServe}
          disabled={toggling}
          className={cn(
            'flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium transition-colors',
            project.status === 'running'
              ? 'bg-red-900/40 text-red-400 hover:bg-red-900/60 border border-red-800/50'
              : 'bg-emerald-900/40 text-emerald-400 hover:bg-emerald-900/60 border border-emerald-800/50'
          )}
        >
          {toggling ? <Loader2 className="w-3 h-3 animate-spin" /> :
           project.status === 'running'
            ? <><Square className="w-3 h-3" /> 중지</>
            : <><Play className="w-3 h-3" /> 시작</>
          }
        </button>
      </div>

      {/* 이슈 목록 패널 (접기/펼치기) */}
      {issueListOpen && (
        <div className="shrink-0 px-4 py-2.5 bg-zinc-950 border-b border-zinc-800 max-h-64 overflow-y-auto">
          <IssueList
            projectRepo={project.repo}
            issues={githubIssues}
            overlay={overlay}
            loading={issuesLoading}
            onRefresh={() => {
              fetchGithubIssues()
              fetch(`/api/state?projectId=${project.id}`, { method: 'POST' }).catch(() => {})
            }}
          />
        </div>
      )}

      {/* 로그 패널 (접기/펼치기) */}
      {project.status === 'running' && (
        <div className="shrink-0 border-b border-zinc-800">
          <button
            onClick={() => setLogOpen(v => !v)}
            className="w-full flex items-center gap-2 px-4 py-1 bg-zinc-900 hover:bg-zinc-800 transition-colors text-left"
          >
            <Terminal className="w-3 h-3 text-zinc-600" />
            <span className="text-[11px] text-zinc-600 font-mono flex-1">
              로그 {logLines.length > 0 ? `(${logLines.length})` : ''}
            </span>
            {logOpen ? <ChevronUp className="w-3 h-3 text-zinc-700" /> : <ChevronDown className="w-3 h-3 text-zinc-700" />}
          </button>
          {logOpen && (
            <div className="bg-zinc-950 max-h-36 overflow-y-auto px-4 py-2 font-mono">
              {logLines.length === 0
                ? <p className="text-[11px] text-zinc-700 py-1">로그 대기 중...</p>
                : logLines.map((line, i) => (
                  <p key={i} className={cn(
                    'text-[11px] leading-relaxed whitespace-pre-wrap break-all',
                    line.startsWith('❌') ? 'text-red-400' :
                    line.startsWith('⚠️') ? 'text-amber-400' :
                    line.startsWith('  [') ? 'text-zinc-400' :
                    'text-zinc-600'
                  )}>{line}</p>
                ))
              }
              <div ref={logEndRef} />
            </div>
          )}
        </div>
      )}

      {/* 메시지 + 라이브 카드 영역 — 주 화면 */}
      <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto">
        <div className="max-w-3xl mx-auto px-4 py-6 flex flex-col gap-4">
          {messages.length === 0 && liveIssues.size === 0 && doneIssues.length === 0 && (
            <div className="flex flex-col items-center justify-center h-64 gap-4 text-center">
              <Bot className="w-14 h-14 text-zinc-700" />
              <div>
                <p className="text-zinc-300 font-medium">작업을 입력하세요</p>
                <p className="text-sm text-zinc-600 mt-1">
                  메시지를 보내면 GitHub 이슈가 생성되고<br />
                  jikime serve가 자동으로 작업을 처리해요
                </p>
              </div>
            </div>
          )}

          {/* 채팅 메시지 */}
          {messages.map(msg => (
            <div key={msg.id} className={cn('flex gap-3', msg.role === 'user' ? 'flex-row-reverse' : 'flex-row')}>
              <div className={cn(
                'w-7 h-7 rounded-full flex items-center justify-center shrink-0 mt-0.5',
                msg.role === 'user' ? 'bg-blue-600' : 'bg-zinc-800 border border-zinc-700'
              )}>
                {msg.role === 'user'
                  ? <User className="w-3.5 h-3.5 text-white" />
                  : <Bot className="w-3.5 h-3.5 text-zinc-400" />
                }
              </div>
              <div className={cn(
                'max-w-[80%] rounded-2xl px-4 py-2.5 text-sm',
                msg.role === 'user'
                  ? 'bg-blue-600 text-white rounded-tr-sm'
                  : 'bg-zinc-800/60 text-zinc-300 rounded-tl-sm border border-zinc-700/50'
              )}>
                <p className="whitespace-pre-wrap break-words leading-relaxed">{msg.text}</p>
                {msg.issueUrl && (
                  <a href={msg.issueUrl} target="_blank" rel="noopener noreferrer"
                    className="mt-1.5 flex items-center gap-1 text-xs text-blue-400 hover:underline">
                    <ExternalLink className="w-3 h-3" />
                    GitHub Issue #{msg.issueNumber} 보기
                  </a>
                )}
              </div>
            </div>
          ))}

          {/* 라이브 이슈 카드 (작업 중) — 채팅 영역의 핵심 */}
          {Array.from(liveIssues.values()).map(live => (
            <LiveIssueCard key={live.identifier} live={live} />
          ))}

          {/* 완료된 이슈 카드 */}
          {doneIssues.map(id => (
            <DoneCard key={id} identifier={id} />
          ))}
        </div>
      </div>

      {/* 입력창 */}
      <div className="shrink-0 border-t border-zinc-800 bg-zinc-900 px-4 py-3">
        {project.status !== 'running' && (
          <div className="max-w-3xl mx-auto mb-2 flex items-center gap-2 text-xs text-amber-500">
            <Play className="w-3 h-3" />
            jikime serve를 먼저 시작해야 이슈가 처리돼요
          </div>
        )}
        <div className="max-w-3xl mx-auto flex items-end gap-3">
          <textarea
            ref={inputRef}
            value={input}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            placeholder="작업 내용을 입력하세요 (GitHub 이슈로 생성됨)..."
            rows={1}
            className="flex-1 min-h-[42px] max-h-[160px] resize-none rounded-xl bg-zinc-800 border border-zinc-700 px-4 py-2.5 text-sm text-zinc-100 placeholder:text-zinc-500 outline-none focus:ring-1 focus:ring-blue-500 transition-colors"
            style={{ height: '42px' }}
          />
          <button
            onClick={submit}
            disabled={!input.trim() || sending}
            className="bg-blue-600 hover:bg-blue-500 disabled:opacity-40 text-white rounded-xl h-[42px] w-[42px] flex items-center justify-center shrink-0 transition-colors"
          >
            {sending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
          </button>
        </div>
        <p className="max-w-3xl mx-auto text-xs text-zinc-600 mt-1.5 px-1">
          Enter 전송 · Shift+Enter 줄바꿈 · 메시지 = GitHub Issue 생성
        </p>
      </div>
    </div>
  )
}

// slog 텍스트 형식에서 key=value 또는 key="value" 추출
function slogAttr(raw: string, key: string): string | null {
  // key="value with spaces"
  const quoted = raw.match(new RegExp(`${key}="((?:[^"\\\\]|\\\\.)*)"` ))
  if (quoted) return quoted[1].replace(/\\"/g, '"')
  // key=value_without_spaces
  const plain = raw.match(new RegExp(`${key}=(\\S+)`))
  if (plain) return plain[1]
  return null
}

function formatLogLine(raw: string): string | null {
  if (!raw.trim()) return null

  // slog 구조화 로그
  const msg = slogAttr(raw, 'msg')
  if (msg) {
    const level = slogAttr(raw, 'level') ?? 'INFO'
    const prefix = level === 'ERROR' ? '❌' : level === 'WARN' ? '⚠️' : '•'

    // agent event: 채팅 LiveIssueCard에서 표시 — 로그 패널에서는 제외
    if (msg === 'agent event') return null

    // claude stderr: 실제 stderr 내용 표시
    if (msg === 'claude stderr') {
      const line = slogAttr(raw, 'line')
      return line ? `⚠️ stderr: ${line}` : null
    }

    // 그 외 slog 라인: msg + 주요 속성
    const parts: string[] = []
    const identifier = slogAttr(raw, 'issue_identifier')
    const hook = slogAttr(raw, 'hook')
    const error = slogAttr(raw, 'error')
    const attempt = slogAttr(raw, 'attempt')
    const path = slogAttr(raw, 'path') ?? slogAttr(raw, 'addr')
    if (identifier) parts.push(identifier)
    if (hook) parts.push(`hook:${hook}`)
    if (attempt && attempt !== 'first') parts.push(`시도 #${attempt}`)
    if (error) parts.push(`→ ${error}`)
    if (path) parts.push(path)
    return `${prefix} ${msg}${parts.length ? ' — ' + parts.join(' ') : ''}`
  }

  // 훅 스크립트 출력 ([after_create], [before_run] 등)
  if (raw.trim().startsWith('[')) return `  ${raw.trim()}`

  // 배너, config 출력 등 그 외 모든 라인
  const trimmed = raw.trim()
  if (trimmed) return `  ${trimmed}`
  return null
}
