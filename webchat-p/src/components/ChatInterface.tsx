'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { Bot, User, Send, Loader2, FolderOpen, Trash2, RotateCcw, ChevronDown, ChevronRight, Terminal, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import MarkdownRenderer from '@/components/MarkdownRenderer'
import SessionPicker from '@/components/SessionPicker'
import { loadMessages, saveMessages, clearMessages } from '@/lib/db'

// ─── 타입 ────────────────────────────────────────────────────────────────────

interface ProgressStep {
  type: 'tool_call' | 'tool_result'
  name?: string    // tool_call: 도구 이름
  input?: string   // tool_call: 입력
  content?: string // tool_result: 결과
}

interface Message {
  id: string
  role: 'user' | 'assistant' | 'error'
  text: string
  streaming?: boolean
  progress: ProgressStep[]
}

// ─── 상수 ────────────────────────────────────────────────────────────────────

const CHARS_PER_TICK = 4
const TICK_MS = 16

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
    <div className="mb-3 rounded-lg border border-zinc-700/60 bg-zinc-900/60 overflow-hidden text-xs">
      {/* 헤더 */}
      <button
        onClick={() => setExpanded(v => !v)}
        className="w-full flex items-center gap-2 px-3 py-2 text-zinc-400 hover:text-zinc-300 hover:bg-zinc-800/40 transition-colors"
      >
        {expanded ? <ChevronDown className="w-3 h-3 shrink-0" /> : <ChevronRight className="w-3 h-3 shrink-0" />}
        <span className="font-mono font-medium">작업 내역</span>
        <span className="text-zinc-600">
          {steps.filter(s => s.type === 'tool_call').length}개 도구 사용
        </span>
        {streaming && (
          <span className="ml-auto flex items-center gap-1 text-emerald-500">
            <Loader2 className="w-2.5 h-2.5 animate-spin" />
            실행 중
          </span>
        )}
      </button>

      {/* 스텝 목록 */}
      {expanded && (
        <div className="border-t border-zinc-700/60 divide-y divide-zinc-800/60">
          {steps.map((step, i) => (
            <ProgressStepRow key={i} step={step} />
          ))}
          {streaming && steps.length === 0 && (
            <div className="px-3 py-2 text-zinc-500 flex items-center gap-2">
              <Loader2 className="w-3 h-3 animate-spin text-emerald-500" />
              준비 중...
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
            <button
              onClick={() => setExpanded(v => !v)}
              className="ml-2 text-zinc-500 hover:text-zinc-300 underline"
            >
              {expanded ? '접기' : '더 보기'}
            </button>
          )}
        </div>
      </div>
    )
  }

  // tool_result
  const lines = (step.content ?? '').split('\n')
  const preview = lines.slice(0, 4).join('\n')
  const hasMore = lines.length > 4

  return (
    <div className="px-3 py-2 bg-zinc-800/30">
      <div className="text-zinc-500 mb-1">결과</div>
      <pre className="font-mono text-zinc-400 whitespace-pre-wrap break-all text-[11px] leading-relaxed">
        {expanded ? step.content : preview}
      </pre>
      {hasMore && (
        <button
          onClick={() => setExpanded(v => !v)}
          className="mt-1 text-zinc-500 hover:text-zinc-300 underline text-[11px]"
        >
          {expanded ? '접기' : `+${lines.length - 4}줄 더 보기`}
        </button>
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
  const [sessionId, setSessionId] = useState('')

  const scrollRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const abortRef = useRef<AbortController | null>(null)

  // 타이프라이터 refs
  const twQueueRef = useRef('')
  const twIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const twDoneRef = useRef(false)
  const twAssistantIdRef = useRef('')

  // cwd 변경 시 → localStorage에 저장
  useEffect(() => {
    localStorage.setItem('webchat_cwd', cwd)
  }, [cwd])

  // 세션 변경 시 → DB에서 메시지 로드
  useEffect(() => {
    loadMessages(sessionId)
      .then(stored => setMessages(stored))
      .catch(() => setMessages([]))
  }, [sessionId])

  // 메시지 변경 시 → DB에 저장 (스트리밍 중 제외)
  useEffect(() => {
    if (messages.length === 0) return
    if (messages.some(m => m.streaming)) return
    saveMessages(sessionId, messages).catch(() => {})
  }, [messages, sessionId])

  // 스크롤 자동 하단
  useEffect(() => {
    const el = scrollRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [messages])

  const stopTypewriter = useCallback(() => {
    if (twIntervalRef.current) {
      clearInterval(twIntervalRef.current)
      twIntervalRef.current = null
    }
    twQueueRef.current = ''
    twDoneRef.current = false
  }, [])

  const startTypewriter = useCallback((assistantId: string) => {
    stopTypewriter()
    twAssistantIdRef.current = assistantId
    twDoneRef.current = false

    twIntervalRef.current = setInterval(() => {
      const queue = twQueueRef.current
      if (queue.length === 0) {
        if (twDoneRef.current) {
          clearInterval(twIntervalRef.current!)
          twIntervalRef.current = null
          setMessages(prev => prev.map(m =>
            m.id === twAssistantIdRef.current ? { ...m, streaming: false } : m
          ))
          setIsProcessing(false)
        }
        return
      }
      const chunk = queue.slice(0, CHARS_PER_TICK)
      twQueueRef.current = queue.slice(CHARS_PER_TICK)
      setMessages(prev => prev.map(m =>
        m.id === twAssistantIdRef.current ? { ...m, text: m.text + chunk } : m
      ))
    }, TICK_MS)
  }, [stopTypewriter])

  const submit = useCallback(async () => {
    const trimmed = input.trim()
    if (!trimmed || isProcessing) return

    setInput('')
    setIsProcessing(true)
    if (inputRef.current) inputRef.current.style.height = '42px'

    const userId = `user-${Date.now()}`
    const assistantId = `assistant-${Date.now() + 1}`

    setMessages(prev => [
      ...prev,
      { id: userId, role: 'user', text: trimmed, progress: [] },
      { id: assistantId, role: 'assistant', text: '', streaming: true, progress: [] },
    ])

    startTypewriter(assistantId)

    const controller = new AbortController()
    abortRef.current = controller

    try {
      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: trimmed,
          sessionId: sessionId || undefined,
          cwd: cwd || undefined,
        }),
        signal: controller.signal,
      })

      if (!res.ok || !res.body) {
        stopTypewriter()
        setMessages(prev => prev.map(m =>
          m.id === assistantId
            ? { ...m, role: 'error' as const, text: `HTTP ${res.status} 오류`, streaming: false }
            : m
        ))
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
          try {
            const event = JSON.parse(line.slice(6))

            if (event.type === 'session_id') {
              setSessionId(event.id)

            } else if (event.type === 'tool_call') {
              // 진행 상황 패널에 도구 호출 추가
              setMessages(prev => prev.map(m =>
                m.id === assistantId
                  ? { ...m, progress: [...m.progress, { type: 'tool_call', name: event.name, input: event.input }] }
                  : m
              ))

            } else if (event.type === 'tool_result') {
              // 진행 상황 패널에 결과 추가
              setMessages(prev => prev.map(m =>
                m.id === assistantId
                  ? { ...m, progress: [...m.progress, { type: 'tool_result', content: event.content }] }
                  : m
              ))

            } else if (event.type === 'text') {
              // 타이프라이터 큐에 추가
              twQueueRef.current += event.text

            } else if (event.type === 'error') {
              stopTypewriter()
              setMessages(prev => prev.map(m =>
                m.id === assistantId
                  ? { ...m, role: 'error' as const, text: event.text, streaming: false }
                  : m
              ))
              setIsProcessing(false)
              return

            } else if (event.type === 'done') {
              twDoneRef.current = true
            }
          } catch { /* skip */ }
        }
      }
    } catch (err) {
      if ((err as Error).name !== 'AbortError') {
        stopTypewriter()
        setMessages(prev => prev.map(m =>
          m.id === assistantId && m.streaming
            ? { ...m, role: 'error' as const, text: `오류: ${(err as Error).message}`, streaming: false }
            : m
        ))
        setIsProcessing(false)
      }
    } finally {
      abortRef.current = null
    }
  }, [input, isProcessing, sessionId, cwd, startTypewriter, stopTypewriter])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
  }

  const handleTextareaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`
  }

  const handleStop = () => {
    abortRef.current?.abort()
    stopTypewriter()
    setIsProcessing(false)
    setMessages(prev => prev.map(m => m.streaming ? { ...m, streaming: false } : m))
  }

  return (
    <div className="flex flex-col h-screen bg-zinc-950 text-zinc-100">
      {/* 헤더 */}
      <header className="flex items-center gap-3 px-5 py-3 bg-zinc-900 border-b border-zinc-800 shrink-0">
        {/* 로고 */}
        <div className="flex items-center gap-2 shrink-0">
          <Bot className="w-5 h-5 text-blue-400" />
          <span className="text-base font-bold text-blue-400">JiKiME</span>
          <span className="text-xs text-zinc-500 font-mono">claude -p</span>
        </div>

        {/* 작업 경로 */}
        <div className="flex items-center gap-1.5 flex-1 min-w-0 max-w-xs">
          <FolderOpen className="w-3.5 h-3.5 text-zinc-500 shrink-0" />
          <Input
            value={cwd}
            onChange={e => setCwd(e.target.value)}
            placeholder="작업 경로 (비워두면 HOME)"
            className="h-7 text-xs bg-zinc-800 border-zinc-700 text-zinc-400 font-mono"
          />
        </div>

        {/* 세션 피커 + 액션 버튼 */}
        <div className="ml-auto flex items-center gap-2 shrink-0">
          <SessionPicker
            sessionId={sessionId}
            cwd={cwd}
            onChange={(sid, scwd) => {
              setSessionId(sid)
              if (scwd) setCwd(scwd)
              setMessages([])
            }}
          />
          <Button
            variant="ghost" size="icon"
            onClick={() => { setMessages([]); setSessionId('') }}
            className="h-7 w-7 text-zinc-400 hover:text-zinc-100"
            title="새 대화"
          >
            <RotateCcw className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="ghost" size="icon"
            onClick={() => {
              clearMessages(sessionId).catch(() => {})
              setMessages([])
            }}
            className="h-7 w-7 text-zinc-400 hover:text-red-400"
            title="대화 삭제"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      </header>

      {/* 메시지 영역 */}
      <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto">
        <div className="max-w-3xl mx-auto px-4 py-6 flex flex-col gap-5">

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
            <div
              key={msg.id}
              className={cn('flex gap-3', msg.role === 'user' ? 'flex-row-reverse' : 'flex-row')}
            >
              {/* 아바타 */}
              <div className={cn(
                'flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center mt-0.5',
                msg.role === 'user' ? 'bg-blue-600'
                  : msg.role === 'error' ? 'bg-red-800'
                  : 'bg-emerald-700'
              )}>
                {msg.role === 'user'
                  ? <User className="w-4 h-4 text-white" />
                  : <Bot className="w-4 h-4 text-zinc-200" />
                }
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
                  <p className="whitespace-pre-wrap break-words">{msg.text}</p>
                ) : msg.role === 'error' ? (
                  <p className="whitespace-pre-wrap break-words">{msg.text || '...'}</p>
                ) : (
                  <>
                    {/* 진행 상황 패널 */}
                    <ProgressPanel steps={msg.progress} streaming={msg.streaming && msg.progress.length === 0 && !msg.text} />

                    {/* 최종 답변 + 타이프라이터 커서 */}
                    {msg.text
                      ? <MarkdownRenderer>{msg.text}</MarkdownRenderer>
                      : !msg.streaming
                      ? <span className="text-zinc-500">...</span>
                      : null
                    }
                    {msg.streaming && (
                      <span className="inline-block w-0.5 h-4 ml-0.5 bg-emerald-400 align-middle animate-pulse" />
                    )}
                  </>
                )}
              </div>
            </div>
          ))}

          {/* 초기 로딩 (스트리밍 메시지 없음) */}
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
        <div className="max-w-3xl mx-auto flex items-end gap-3">
          <textarea
            ref={inputRef}
            value={input}
            onChange={handleTextareaChange}
            onKeyDown={handleKeyDown}
            placeholder={isProcessing ? '응답 대기 중...' : 'Claude Code에게 메시지를 보내세요...'}
            rows={1}
            disabled={isProcessing}
            className="flex-1 min-h-[42px] max-h-[160px] resize-none rounded-xl bg-zinc-800 border border-zinc-700 px-4 py-2.5 text-sm text-zinc-100 placeholder:text-zinc-500 outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ height: '42px' }}
          />
          {isProcessing ? (
            <Button
              onClick={handleStop}
              size="icon"
              className="bg-red-600 hover:bg-red-500 shrink-0 self-end h-[42px] w-[42px] rounded-xl"
            >
              <span className="w-3 h-3 bg-white rounded-sm block" />
            </Button>
          ) : (
            <Button
              onClick={submit}
              disabled={!input.trim()}
              size="icon"
              className="bg-blue-600 hover:bg-blue-500 shrink-0 self-end h-[42px] w-[42px] rounded-xl disabled:opacity-40"
            >
              <Send className="w-4 h-4" />
            </Button>
          )}
        </div>
        <p className="max-w-3xl mx-auto text-xs text-zinc-600 mt-1.5 px-1">
          Enter 전송 · Shift+Enter 줄바꿈 · 대화가 자동으로 이어집니다
        </p>
      </div>
    </div>
  )
}
