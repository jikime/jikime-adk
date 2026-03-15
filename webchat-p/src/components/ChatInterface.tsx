'use client'

import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { Bot, User, Send, Loader2, FolderOpen, Trash2, RotateCcw, ChevronDown, ChevronRight, Terminal, FileText, Paperclip, X, Image as ImageIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import MarkdownRenderer from '@/components/MarkdownRenderer'
import SessionPicker from '@/components/SessionPicker'
import { loadMessages, saveMessages, clearMessages } from '@/lib/db'
import type { SlashCommand } from '@/app/api/commands/route'

// ─── 타입 ────────────────────────────────────────────────────────────────────

interface AttachedFile {
  id: string
  name: string
  size: number
  type: string
  // 텍스트 파일: content (클라이언트 읽기), 이미지/바이너리: serverPath (업로드 후 경로)
  content?: string
  serverPath?: string
  preview?: string  // 이미지 미리보기 DataURL
}

interface ProgressStep {
  type: 'tool_call' | 'tool_result'
  name?: string
  input?: string
  content?: string
  images?: { data: string; mediaType: string }[]
}

type MsgStatus = 'pending' | 'started' | 'running' | 'done'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'error'
  text: string
  streaming?: boolean
  progress: ProgressStep[]
  status?: MsgStatus
  attachments?: AttachedFile[]
}

// ─── 상태 배지 ────────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status?: MsgStatus }) {
  if (!status || status === 'pending') return (
    <span className="inline-flex items-center gap-1 text-[10px] text-zinc-500">
      <Loader2 className="w-2.5 h-2.5 animate-spin" /> 대기 중
    </span>
  )
  if (status === 'started') return (
    <span className="inline-flex items-center gap-1 text-[10px] text-blue-400">
      <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" /> 시작
    </span>
  )
  if (status === 'running') return (
    <span className="inline-flex items-center gap-1 text-[10px] text-emerald-400">
      <Loader2 className="w-2.5 h-2.5 animate-spin" /> 진행 중
    </span>
  )
  return (
    <span className="inline-flex items-center gap-1 text-[10px] text-zinc-500">
      <span className="w-1.5 h-1.5 rounded-full bg-zinc-500" /> 완료
    </span>
  )
}

// ─── 툴 아이콘 ────────────────────────────────────────────────────────────────

function ToolIcon({ name }: { name: string }) {
  if (name === 'Bash') return <Terminal className="w-3 h-3" />
  return <FileText className="w-3 h-3" />
}

// ─── 진행 상황 패널 ──────────────────────────────────────────────────────────

function ProgressPanel({ steps, streaming }: { steps: ProgressStep[]; streaming?: boolean }) {
  const [expanded, setExpanded] = useState(true)
  if (steps.length === 0 && !streaming) return null

  return (
    <div className="mt-2 rounded-lg border border-zinc-700/60 bg-zinc-900/60 overflow-hidden text-xs">
      <button
        onClick={() => setExpanded(v => !v)}
        className="w-full flex items-center gap-2 px-3 py-2 text-zinc-400 hover:text-zinc-300 hover:bg-zinc-800/40 transition-colors"
      >
        {expanded ? <ChevronDown className="w-3 h-3 shrink-0" /> : <ChevronRight className="w-3 h-3 shrink-0" />}
        <span className="font-mono font-medium">작업 내역</span>
        <span className="text-zinc-600">{steps.filter(s => s.type === 'tool_call').length}개 도구 사용</span>
        {streaming && (
          <span className="ml-auto flex items-center gap-1 text-emerald-500">
            <Loader2 className="w-2.5 h-2.5 animate-spin" /> 실행 중
          </span>
        )}
      </button>
      {expanded && (
        <div className="border-t border-zinc-700/60 divide-y divide-zinc-800/60">
          {steps.map((step, i) => <ProgressStepRow key={i} step={step} />)}
          {streaming && steps.length === 0 && (
            <div className="px-3 py-2 text-zinc-500 flex items-center gap-2">
              <Loader2 className="w-3 h-3 animate-spin text-emerald-500" /> 준비 중...
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function ProgressStepRow({ step }: { step: ProgressStep }) {
  const [expanded, setExpanded] = useState(false)

  if (step.type === 'tool_call') {
    const preview = (step.input ?? '').split('\n')[0].slice(0, 80)
    const hasMore = (step.input ?? '').length > preview.length || (step.input ?? '').includes('\n')
    return (
      <div className="px-3 py-2">
        <div className="flex items-center gap-2 text-emerald-400">
          <ToolIcon name={step.name ?? ''} />
          <span className="font-mono font-semibold">{step.name}</span>
        </div>
        <div className="mt-1 font-mono text-zinc-300 whitespace-pre-wrap break-all">
          {expanded ? step.input : preview}
          {hasMore && (
            <button onClick={() => setExpanded(v => !v)} className="ml-2 text-zinc-500 hover:text-zinc-300 underline">
              {expanded ? '접기' : '더 보기'}
            </button>
          )}
        </div>
      </div>
    )
  }

  const lines = (step.content ?? '').split('\n').filter(Boolean)
  const preview = lines.slice(0, 4).join('\n')
  const hasMore = lines.length > 4
  return (
    <div className="px-3 py-2 bg-zinc-800/30">
      <div className="text-zinc-500 mb-1">결과</div>
      {lines.length > 0 && (
        <>
          <pre className="font-mono text-zinc-400 whitespace-pre-wrap break-all text-[11px] leading-relaxed">
            {expanded ? step.content : preview}
          </pre>
          {hasMore && (
            <button onClick={() => setExpanded(v => !v)} className="mt-1 text-zinc-500 hover:text-zinc-300 underline text-[11px]">
              {expanded ? '접기' : `+${lines.length - 4}줄 더 보기`}
            </button>
          )}
        </>
      )}
      {step.images && step.images.length > 0 && (
        <div className="flex flex-wrap gap-2 mt-2">
          {step.images.map((img, i) => (
            <a
              key={i}
              href={`data:${img.mediaType};base64,${img.data}`}
              target="_blank"
              rel="noopener noreferrer"
              title="클릭하면 원본 크기로 열려요"
            >
              <img
                src={`data:${img.mediaType};base64,${img.data}`}
                alt={`결과 이미지 ${i + 1}`}
                className="max-w-sm max-h-64 rounded-lg border border-zinc-700 object-contain hover:border-blue-500 transition-colors cursor-zoom-in"
              />
            </a>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── 메인 컴포넌트 ────────────────────────────────────────────────────────────

export default function ChatInterface() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [cwd, setCwd] = useState(() => {
    if (typeof window === 'undefined') return ''
    return localStorage.getItem('webchat_cwd') ?? ''
  })
  // sessionId를 localStorage에서 초기화 — 새로고침 후에도 마지막 세션 복원
  const [sessionId, setSessionId] = useState(() => {
    if (typeof window === 'undefined') return ''
    return localStorage.getItem('webchat_session') ?? ''
  })

  const [attachments, setAttachments] = useState<AttachedFile[]>([])
  const [uploading, setUploading] = useState(false)

  const scrollRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const abortRef = useRef<AbortController | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // ─── 슬래시 커맨드 자동완성 ───────────────────────────────────────────────
  const [commands, setCommands] = useState<SlashCommand[]>([])
  const [slashOpen, setSlashOpen] = useState(false)
  const [slashIdx, setSlashIdx] = useState(0)
  const slashListRef = useRef<HTMLDivElement>(null)

  useEffect(() => { localStorage.setItem('webchat_cwd', cwd) }, [cwd])

  // sessionId 변경 시 localStorage에 저장
  useEffect(() => { localStorage.setItem('webchat_session', sessionId) }, [sessionId])

  // cwd 변경 시 커맨드 목록 갱신
  useEffect(() => {
    const params = cwd ? `?cwd=${encodeURIComponent(cwd)}` : ''
    fetch(`/api/commands${params}`)
      .then(r => r.json())
      .then(setCommands)
      .catch(() => {})
  }, [cwd])

  // 입력값 기반 필터링
  const slashFiltered = useMemo(() => {
    if (!input.startsWith('/')) return []
    const q = input.slice(1).toLowerCase()
    if (!q) return commands
    if (q.includes(':')) {
      const [ns, cmd] = q.split(':')
      return commands.filter(c =>
        c.namespace.toLowerCase().startsWith(ns) &&
        c.command.toLowerCase().includes(cmd)
      )
    }
    return commands.filter(c =>
      c.namespace.toLowerCase().includes(q) ||
      `${c.namespace}:${c.command}`.toLowerCase().includes(q)
    )
  }, [input, commands])

  // 드롭다운 열기/닫기 + 인덱스 리셋
  useEffect(() => {
    if (input.startsWith('/') && slashFiltered.length > 0) {
      setSlashOpen(true)
      setSlashIdx(0)
    } else {
      setSlashOpen(false)
    }
  }, [input, slashFiltered.length])

  // 선택된 항목이 드롭다운 내에서 보이도록 스크롤
  useEffect(() => {
    const el = slashListRef.current?.children[slashIdx] as HTMLElement | undefined
    el?.scrollIntoView({ block: 'nearest' })
  }, [slashIdx])

  const selectSlash = useCallback((cmd: SlashCommand) => {
    setInput(`/${cmd.namespace}:${cmd.command} `)
    setSlashOpen(false)
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [])

  // ─── 파일 첨부 ────────────────────────────────────────────────────────────
  const TEXT_TYPES = ['text/', 'application/json', 'application/xml', 'application/javascript', 'application/typescript']
  const isTextFile = (f: File) => TEXT_TYPES.some(t => f.type.startsWith(t)) || /\.(md|txt|ts|tsx|js|jsx|json|yaml|yml|toml|xml|csv|sh|py|go|rs|java|c|cpp|h|css|html|env|ini|cfg)$/i.test(f.name)

  const handleFiles = useCallback(async (files: FileList | File[]) => {
    const fileArr = Array.from(files)
    if (!fileArr.length) return
    setUploading(true)

    const newAttachments: AttachedFile[] = []
    const toUpload: File[] = []

    for (const file of fileArr) {
      const id = `attach-${Date.now()}-${Math.random().toString(36).slice(2)}`

      if (isTextFile(file)) {
        // 텍스트 파일: 클라이언트에서 직접 읽기
        const content = await file.text()
        newAttachments.push({ id, name: file.name, size: file.size, type: file.type || 'text/plain', content })
      } else if (file.type.startsWith('image/')) {
        // 이미지: 미리보기 생성 + 서버 업로드
        const preview = await new Promise<string>(resolve => {
          const reader = new FileReader()
          reader.onload = e => resolve(e.target?.result as string)
          reader.readAsDataURL(file)
        })
        newAttachments.push({ id, name: file.name, size: file.size, type: file.type, preview, serverPath: '' })
        toUpload.push(file)
      } else {
        // 기타 바이너리: 서버 업로드
        newAttachments.push({ id, name: file.name, size: file.size, type: file.type, serverPath: '' })
        toUpload.push(file)
      }
    }

    // 이미지/바이너리 업로드
    if (toUpload.length > 0) {
      try {
        const formData = new FormData()
        toUpload.forEach(f => formData.append('files', f))
        const res = await fetch('/api/upload', { method: 'POST', body: formData })
        if (res.ok) {
          const uploaded: { name: string; path: string; size: number; type: string }[] = await res.json()
          // 업로드 결과를 ID 순서대로 매핑
          let upIdx = 0
          for (const att of newAttachments) {
            if (att.serverPath === '' && !att.content) {
              att.serverPath = uploaded[upIdx]?.path ?? ''
              upIdx++
            }
          }
        }
      } catch { /* 업로드 실패 시 path 빈 상태 유지 */ }
    }

    setAttachments(prev => [...prev, ...newAttachments])
    setUploading(false)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const removeAttachment = useCallback((id: string) => {
    setAttachments(prev => prev.filter(a => a.id !== id))
  }, [])

  // 첨부파일을 포함한 최종 메시지 조합
  // 텍스트/코드 파일만 인라인 — 이미지/바이너리는 imagePaths 배열로 분리 전달
  const buildMessageWithAttachments = useCallback((text: string, atts: AttachedFile[]): string => {
    if (!atts.length) return text
    const parts: string[] = [text]
    for (const att of atts) {
      if (att.content !== undefined) {
        // 텍스트/코드 파일: 내용 인라인
        const lang = att.name.split('.').pop() ?? ''
        parts.push(`\n\n[첨부 파일: ${att.name}]\n\`\`\`${lang}\n${att.content}\n\`\`\``)
      }
      // 이미지/바이너리는 텍스트에 포함하지 않음 → imagePaths 배열로 --add-file 전달
    }
    return parts.join('')
  }, [])

  useEffect(() => {
    if (!sessionId) return  // 새 대화 시작 — DB 로드 없이 빈 상태 유지
    loadMessages(sessionId)
      .then(stored => setMessages(prev =>
        // DB에 저장된 메시지가 있으면 교체 (세션 피커로 이전 대화 로드)
        // 없으면 현재 메모리 메시지 유지 (SSE로 새 session_id 할당된 경우 등)
        stored.length > 0 ? stored : prev
      ))
      .catch(() => {})
  }, [sessionId])

  useEffect(() => {
    if (messages.length === 0) return
    if (messages.some(m => m.streaming)) return
    saveMessages(sessionId, messages).catch(() => {})
  }, [messages, sessionId])

  // 새 메시지 도착 시 자동 스크롤
  useEffect(() => {
    const el = scrollRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [messages])

  const submit = useCallback(async () => {
    const trimmed = input.trim()
    if ((!trimmed && attachments.length === 0) || isProcessing) return

    const currentAttachments = [...attachments]
    setInput('')
    setAttachments([])
    setIsProcessing(true)
    if (inputRef.current) inputRef.current.style.height = '42px'

    const userId = `user-${Date.now()}`
    const assistantId = `assistant-${Date.now() + 1}`

    // 사용자 메시지에 첨부파일 목록 표시용 접두사
    const displayText = trimmed || '(파일 첨부)'
    // 텍스트/코드 파일만 인라인 — 이미지/바이너리는 imagePaths로 --add-file 전달
    const finalMessage = buildMessageWithAttachments(trimmed || '첨부된 파일을 분석해주세요.', currentAttachments)
    // 이미지/바이너리 파일 경로만 추출 (텍스트 파일은 이미 인라인됨)
    const imagePaths = currentAttachments
      .filter(a => a.serverPath && a.content === undefined)
      .map(a => a.serverPath as string)
      .filter(Boolean)

    // 사용자 메시지 + 빈 어시스턴트 메시지 즉시 추가
    setMessages(prev => [
      ...prev,
      { id: userId, role: 'user', text: displayText, progress: [], attachments: currentAttachments },
      { id: assistantId, role: 'assistant', text: '', streaming: true, progress: [], status: 'pending' },
    ])

    const controller = new AbortController()
    abortRef.current = controller

    // 메시지 업데이트 헬퍼 — 항상 최신 상태 기반으로 업데이트
    const update = (patch: Partial<Message>) =>
      setMessages(prev => prev.map(m => m.id === assistantId ? { ...m, ...patch } : m))

    // SSE에서 수신한 세션 ID — 스트리밍 완료 후 finally에서 적용
    // (중간에 setSessionId하면 useEffect가 loadMessages를 호출해 메시지를 덮어씀)
    let receivedSessionId = ''

    try {
      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: finalMessage,
          sessionId: sessionId || undefined,
          cwd: cwd || undefined,
          imagePaths: imagePaths.length > 0 ? imagePaths : undefined,
        }),
        signal: controller.signal,
      })

      if (!res.ok || !res.body) {
        update({ role: 'error', text: `HTTP ${res.status} 오류`, streaming: false })
        setIsProcessing(false)
        return
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let partial = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        partial += decoder.decode(value, { stream: true })
        const lines = partial.split('\n')
        partial = lines.pop() || ''

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          let event: Record<string, unknown>
          try { event = JSON.parse(line.slice(6)) } catch { continue }

          switch (event.type) {
            case 'session_id':
              // 즉시 setSessionId 하지 않고 저장만 — 중간에 setSessionId 하면
              // useEffect가 loadMessages를 호출해 메시지 배열을 덮어써 버림
              receivedSessionId = event.id as string
              update({ status: 'started' })
              break

            case 'text':
              // 받는 즉시 text에 누적 — 타이프라이터 없이 직접 반영
              setMessages(prev => prev.map(m =>
                m.id === assistantId
                  ? { ...m, text: m.text + (event.text as string), status: 'running' }
                  : m
              ))
              break

            case 'tool_call':
              setMessages(prev => prev.map(m =>
                m.id === assistantId
                  ? { ...m, status: 'running', progress: [...m.progress, { type: 'tool_call', name: event.name as string, input: event.input as string }] }
                  : m
              ))
              break

            case 'tool_result':
              setMessages(prev => prev.map(m =>
                m.id === assistantId
                  ? { ...m, status: 'running', progress: [...m.progress, {
                      type: 'tool_result',
                      content: event.content as string,
                      images: event.images as { data: string; mediaType: string }[] | undefined,
                    }] }
                  : m
              ))
              break

            case 'error':
              update({ role: 'error', text: event.text as string, streaming: false })
              setIsProcessing(false)
              return

            case 'done':
              update({ streaming: false, status: 'done' })
              setIsProcessing(false)
              break
          }
        }
      }
    } catch (err) {
      if ((err as Error).name !== 'AbortError') {
        update({ role: 'error', text: `오류: ${(err as Error).message}`, streaming: false })
        setIsProcessing(false)
      }
    } finally {
      // 스트리밍 완료 후 세션 ID 적용 — 여기서 해야 useEffect가 메시지를 덮어쓰지 않음
      if (receivedSessionId) setSessionId(receivedSessionId)
      abortRef.current = null
    }
  }, [input, attachments, isProcessing, sessionId, cwd, buildMessageWithAttachments])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (slashOpen) {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSlashIdx(i => Math.min(i + 1, slashFiltered.length - 1))
        return
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSlashIdx(i => Math.max(i - 1, 0))
        return
      }
      if (e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault()
        if (slashFiltered[slashIdx]) selectSlash(slashFiltered[slashIdx])
        return
      }
      if (e.key === 'Escape') {
        e.preventDefault()
        setSlashOpen(false)
        return
      }
    }
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); submit() }
  }

  const handleTextareaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`
  }

  const handleStop = () => {
    abortRef.current?.abort()
    setIsProcessing(false)
    setMessages(prev => prev.map(m => m.streaming ? { ...m, streaming: false, status: 'done' } : m))
  }

  return (
    <div className="flex flex-col h-screen bg-zinc-950 text-zinc-100">
      {/* 헤더 */}
      <header className="flex items-center gap-3 px-5 py-3 bg-zinc-900 border-b border-zinc-800 shrink-0">
        <div className="flex items-center gap-2 shrink-0">
          <Bot className="w-5 h-5 text-blue-400" />
          <span className="text-base font-bold text-blue-400">JiKiME</span>
          <span className="text-xs text-zinc-500 font-mono">claude -p</span>
        </div>
        <div className="flex items-center gap-1.5 flex-1 min-w-0 max-w-xs">
          <FolderOpen className="w-3.5 h-3.5 text-zinc-500 shrink-0" />
          <Input
            value={cwd}
            onChange={e => setCwd(e.target.value)}
            placeholder="작업 경로 (비워두면 HOME)"
            className="h-7 text-xs bg-zinc-800 border-zinc-700 text-zinc-400 font-mono"
          />
        </div>
        <div className="ml-auto flex items-center gap-2 shrink-0">
          <SessionPicker
            sessionId={sessionId}
            cwd={cwd}
            onChange={(sid, scwd) => { setSessionId(sid); if (scwd) setCwd(scwd); setMessages([]) }}
          />
          <Button variant="ghost" size="icon" onClick={() => { setMessages([]); setSessionId('') }}
            className="h-7 w-7 text-zinc-400 hover:text-zinc-100" title="새 대화">
            <RotateCcw className="w-3.5 h-3.5" />
          </Button>
          <Button variant="ghost" size="icon"
            onClick={() => { clearMessages(sessionId).catch(() => {}); setMessages([]) }}
            className="h-7 w-7 text-zinc-400 hover:text-red-400" title="대화 삭제">
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      </header>

      {/* 메시지 영역 */}
      <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto">
        <div className="max-w-5xl mx-auto px-4 py-6 flex flex-col gap-5">
          {messages.length === 0 && (
            <div className="flex flex-col items-center justify-center h-64 gap-4 text-center">
              <Bot className="w-14 h-14 text-zinc-700" />
              <div>
                <p className="text-zinc-300 font-medium">Claude Code에게 무엇이든 물어보세요</p>
                <p className="text-sm text-zinc-600 mt-1">대화가 자동으로 이어져요 (멀티턴)</p>
              </div>
            </div>
          )}

          {messages.map(msg => (
            <div key={msg.id} className={cn('flex gap-3', msg.role === 'user' ? 'flex-row-reverse' : 'flex-row')}>
              {/* 아바타 */}
              <div className={cn(
                'flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center mt-0.5',
                msg.role === 'user' ? 'bg-blue-600' : msg.role === 'error' ? 'bg-red-800' : 'bg-emerald-700'
              )}>
                {msg.role === 'user' ? <User className="w-4 h-4 text-white" /> : <Bot className="w-4 h-4 text-zinc-200" />}
              </div>

              {/* 말풍선 */}
              <div className={cn(
                'rounded-2xl px-4 py-3 text-sm leading-relaxed',
                msg.role === 'user'
                  ? 'max-w-[85%] bg-blue-600 text-white rounded-tr-sm'
                  : msg.role === 'error'
                  ? 'max-w-[85%] bg-red-900/50 text-red-200 rounded-tl-sm border border-red-800/50'
                  : 'flex-1 min-w-0 bg-zinc-800 text-zinc-100 rounded-tl-sm border border-zinc-700/50'
              )}>
                {msg.role === 'user' ? (
                  <>
                    <p className="whitespace-pre-wrap break-words">{msg.text}</p>
                    {msg.attachments && msg.attachments.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 mt-2">
                        {msg.attachments.map(att => (
                          <div key={att.id} className="flex items-center gap-1.5 bg-blue-500/30 border border-blue-400/30 rounded-lg px-2 py-1 text-xs">
                            {att.preview
                              ? <img src={att.preview} alt={att.name} className="w-6 h-6 rounded object-cover shrink-0" />
                              : att.content !== undefined
                              ? <FileText className="w-3.5 h-3.5 shrink-0 text-blue-200" />
                              : <ImageIcon className="w-3.5 h-3.5 shrink-0 text-blue-200" />
                            }
                            <span className="truncate max-w-[120px] text-blue-100">{att.name}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </>
                ) : msg.role === 'error' ? (
                  <p className="whitespace-pre-wrap break-words">{msg.text || '...'}</p>
                ) : (
                  <>
                    {/* 상태 배지 — 스트리밍 중에만 표시 */}
                    {msg.streaming && (
                      <div className="mb-2"><StatusBadge status={msg.status} /></div>
                    )}

                    {/* 응답 텍스트 — data.text 수신 즉시 누적 표시 */}
                    {msg.text
                      ? <MarkdownRenderer>{msg.text}</MarkdownRenderer>
                      : !msg.streaming
                      ? <span className="text-zinc-500">...</span>
                      : null
                    }
                    {msg.streaming && (
                      <span className="inline-block w-0.5 h-4 ml-0.5 bg-emerald-400 align-middle animate-pulse" />
                    )}

                    {/* 툴 호출 내역 — 텍스트 아래 배치 */}
                    <ProgressPanel steps={msg.progress} streaming={msg.streaming} />
                  </>
                )}
              </div>
            </div>
          ))}

          {/* 대기 중 (아직 어시스턴트 메시지 없을 때) */}
          {isProcessing && !messages.some(m => m.streaming) && (
            <div className="flex gap-3">
              <div className="flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center bg-emerald-700">
                <Bot className="w-4 h-4 text-zinc-200" />
              </div>
              <div className="flex items-center gap-2 bg-zinc-800 border border-zinc-700/50 rounded-2xl rounded-tl-sm px-4 py-3">
                <Loader2 className="w-4 h-4 text-emerald-400 animate-spin" />
                <span className="text-sm text-zinc-400">생각 중...</span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* 입력창 */}
      <div className="shrink-0 border-t border-zinc-800 bg-zinc-900 px-4 py-3">
        <div className="max-w-5xl mx-auto relative">

          {/* 슬래시 커맨드 드롭다운 */}
          {slashOpen && slashFiltered.length > 0 && (
            <div
              ref={slashListRef}
              className="absolute bottom-full left-0 right-0 mb-2 bg-zinc-900 border border-zinc-700 rounded-xl shadow-2xl overflow-hidden max-h-72 overflow-y-auto z-50 [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]"
            >
              <div className="px-3 py-1.5 text-[10px] text-zinc-500 border-b border-zinc-800 font-mono">
                슬래시 커맨드 — ↑↓ 탐색 · Enter/Tab 선택 · Esc 닫기
              </div>
              {slashFiltered.map((cmd, i) => (
                <button
                  key={`${cmd.namespace}:${cmd.command}`}
                  onMouseDown={e => { e.preventDefault(); selectSlash(cmd) }}
                  className={cn(
                    'w-full flex items-start gap-3 px-3 py-2 text-left transition-colors',
                    i === slashIdx ? 'bg-zinc-700' : 'hover:bg-zinc-800'
                  )}
                >
                  <span className="font-mono text-sm text-emerald-400 shrink-0 mt-0.5">
                    /{cmd.namespace}:{cmd.command}
                  </span>
                  <div className="min-w-0">
                    {cmd.description && (
                      <p className="text-xs text-zinc-400 truncate">{cmd.description}</p>
                    )}
                    {cmd.source === 'project' && (
                      <span className="text-[10px] text-blue-400">project</span>
                    )}
                  </div>
                </button>
              ))}
            </div>
          )}

          {/* 첨부 파일 칩 */}
          {attachments.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mb-2">
              {attachments.map(att => (
                <div key={att.id} className="flex items-center gap-1.5 bg-zinc-800 border border-zinc-700 rounded-lg pl-2 pr-1 py-1 text-xs group">
                  {att.preview
                    ? <img src={att.preview} alt={att.name} className="w-5 h-5 rounded object-cover shrink-0" />
                    : att.content !== undefined
                    ? <FileText className="w-3.5 h-3.5 shrink-0 text-zinc-400" />
                    : <ImageIcon className="w-3.5 h-3.5 shrink-0 text-zinc-400" />
                  }
                  <span className="truncate max-w-[140px] text-zinc-300">{att.name}</span>
                  <span className="text-zinc-600 text-[10px]">{(att.size / 1024).toFixed(0)}K</span>
                  <button
                    onClick={() => removeAttachment(att.id)}
                    className="ml-0.5 text-zinc-600 hover:text-red-400 transition-colors"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </div>
              ))}
              {uploading && (
                <div className="flex items-center gap-1.5 bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1 text-xs text-zinc-500">
                  <Loader2 className="w-3 h-3 animate-spin" /> 업로드 중...
                </div>
              )}
            </div>
          )}

          {/* 텍스트 입력 + 버튼 행 */}
          <div className="flex items-end gap-2">
            {/* 숨겨진 파일 입력 */}
            <input
              ref={fileInputRef}
              type="file"
              multiple
              accept="image/*,.pdf,.txt,.md,.ts,.tsx,.js,.jsx,.json,.yaml,.yml,.toml,.xml,.csv,.sh,.py,.go,.rs,.java,.c,.cpp,.h,.css,.html"
              className="hidden"
              onChange={e => { if (e.target.files) handleFiles(e.target.files); e.target.value = '' }}
            />

            {/* 첨부 버튼 */}
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => fileInputRef.current?.click()}
              disabled={isProcessing}
              title="파일 첨부"
              className="shrink-0 self-end h-[42px] w-[42px] rounded-xl text-zinc-500 hover:text-zinc-200 hover:bg-zinc-800 border border-zinc-700"
            >
              <Paperclip className="w-4 h-4" />
            </Button>

            <textarea
              ref={inputRef}
              value={input}
              onChange={handleTextareaChange}
              onKeyDown={handleKeyDown}
              onDragOver={e => e.preventDefault()}
              onDrop={e => { e.preventDefault(); handleFiles(e.dataTransfer.files) }}
              placeholder={isProcessing ? '응답 대기 중...' : 'Claude Code에게 메시지를 보내세요... (파일을 드래그하여 첨부)'}
              rows={1}
              disabled={isProcessing}
              className="flex-1 min-h-[42px] max-h-[160px] resize-none rounded-xl bg-zinc-800 border border-zinc-700 px-4 py-2.5 text-sm text-zinc-100 placeholder:text-zinc-500 outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              style={{ height: '42px' }}
            />
            {isProcessing ? (
              <Button onClick={handleStop} size="icon"
                className="bg-red-600 hover:bg-red-500 shrink-0 self-end h-[42px] w-[42px] rounded-xl">
                <span className="w-3 h-3 bg-white rounded-sm block" />
              </Button>
            ) : (
              <Button onClick={submit} disabled={!input.trim() && attachments.length === 0} size="icon"
                className="bg-blue-600 hover:bg-blue-500 shrink-0 self-end h-[42px] w-[42px] rounded-xl disabled:opacity-40">
                <Send className="w-4 h-4" />
              </Button>
            )}
          </div>
        </div>
        <p className="max-w-5xl mx-auto text-xs text-zinc-600 mt-1.5 px-1">
          Enter 전송 · Shift+Enter 줄바꿈 · 파일 드래그 또는 📎 버튼으로 첨부 · 대화가 자동으로 이어집니다
        </p>
      </div>
    </div>
  )
}
