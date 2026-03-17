'use client'

import { useState, useRef, useCallback, useEffect } from 'react'
import {
  Send, Square, Bot, ChevronDown, ChevronRight,
  Wrench, AlertCircle, Brain, Shield, Check, X, RefreshCw,
  Plus, Mic, MicOff, FileText, Image as ImageIcon, Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { useWebSocket } from '@/contexts/WebSocketContext'
import { useProject } from '@/contexts/ProjectContext'
import { MODELS, loadSettings, type ModelId } from '@/components/sidebar/Sidebar'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

// ── 타입 ─────────────────────────────────────────────────────────
interface AttachedFile {
  id: string
  name: string
  size: number
  type: string
  content?: string   // 텍스트 파일
  preview?: string   // 이미지 미리보기 DataURL
}

interface ToolCall {
  name: string
  input: unknown
  result?: unknown
}

type MessageStatus = 'streaming' | 'done' | 'error' | 'aborted'

interface Message {
  id: string
  role: 'user' | 'assistant'
  text: string
  status?: MessageStatus
  toolCalls?: ToolCall[]
  thinking?: string
}

interface TokenBudget {
  used: number
  total: number
}

interface PermissionRequest {
  requestId: string
  toolName: string
  input: unknown
}

// ── 서브 컴포넌트 ─────────────────────────────────────────────────

function ThinkingIndicator() {
  return (
    <>
      <style>{`
        @keyframes wc-spin   { from { transform: rotate(0deg);   } to { transform: rotate(360deg);  } }
        @keyframes wc-spin-r { from { transform: rotate(0deg);   } to { transform: rotate(-360deg); } }
        @keyframes wc-spin-d { from { transform: rotate(-30deg); } to { transform: rotate(330deg);  } }
        @keyframes wc-glow   { 0%,100% { opacity:.75; transform:scale(.93); } 50% { opacity:1; transform:scale(1.07); } }
        @keyframes wc-dot    { 0%,80%,100% { transform:translateY(0);   opacity:.35; }
                               40%          { transform:translateY(-4px); opacity:1;   } }
      `}</style>

      <div className="flex items-center gap-3 px-1 py-2">
        {/* 궤도 애니메이션 */}
        <div className="relative shrink-0" style={{ width: 40, height: 40 }}>

          {/* 궤도 링 1 — 황금색, 정방향 */}
          <svg
            className="absolute inset-0 w-full h-full"
            style={{ animation: 'wc-spin 3s linear infinite' }}
            viewBox="0 0 40 40"
          >
            <ellipse cx="20" cy="20" rx="18" ry="6.5"
              fill="none" stroke="#fbbf24" strokeWidth="1" opacity="0.55" />
            {/* 다이아몬드 스파클 */}
            <polygon points="20,0 21.4,2 20,4 18.6,2"
              fill="#fbbf24" opacity="0.95" />
            <polygon points="20,36 21.4,38 20,40 18.6,38"
              fill="#fbbf24" opacity="0.6" />
          </svg>

          {/* 궤도 링 2 — 은색, 역방향, 기울어짐 */}
          <svg
            className="absolute inset-0 w-full h-full"
            style={{ animation: 'wc-spin-r 2.3s linear infinite', transform: 'rotate(55deg)' }}
            viewBox="0 0 40 40"
          >
            <ellipse cx="20" cy="20" rx="18" ry="6.5"
              fill="none" stroke="#94a3b8" strokeWidth="0.9" opacity="0.4" />
            <polygon points="20,0 21.3,2 20,4 18.7,2"
              fill="#94a3b8" opacity="0.85" />
          </svg>

          {/* 궤도 링 3 — 살짝 기울어진 세 번째 */}
          <svg
            className="absolute inset-0 w-full h-full"
            style={{ animation: 'wc-spin-d 4s linear infinite', transform: 'rotate(-40deg)' }}
            viewBox="0 0 40 40"
          >
            <ellipse cx="20" cy="20" rx="18" ry="6.5"
              fill="none" stroke="#d4a057" strokeWidth="0.7" opacity="0.28" />
          </svg>

          {/* 중심 — 황금빛 구체 */}
          <div className="absolute inset-0 flex items-center justify-center">
            <div
              className="w-5 h-5 rounded-full"
              style={{
                background: 'radial-gradient(circle at 38% 35%, #fde68a, #f59e0b 55%, #b45309)',
                animation: 'wc-glow 1.8s ease-in-out infinite',
                boxShadow: '0 0 12px 4px rgba(251,191,36,0.45)',
              }}
            />
          </div>
        </div>

        {/* 텍스트 + 바운스 점 */}
        <span className="text-xs text-zinc-400 flex items-center gap-1.5">
          생각 중
          {[0, 0.18, 0.36].map((delay, i) => (
            <span
              key={i}
              className="inline-block w-1 h-1 rounded-full bg-amber-400/70"
              style={{ animation: `wc-dot 1.1s ease-in-out ${delay}s infinite` }}
            />
          ))}
        </span>
      </div>
    </>
  )
}

function ToolCallView({ tool }: { tool: ToolCall }) {
  const [open, setOpen] = useState(false)
  const inputStr = typeof tool.input === 'string'
    ? tool.input
    : JSON.stringify(tool.input, null, 2)

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="flex items-center gap-1.5 text-xs text-amber-400 hover:text-amber-300 py-0.5 w-full text-left">
        <Wrench className="w-3 h-3 shrink-0" />
        <span className="font-mono truncate">{tool.name}</span>
        {open ? <ChevronDown className="w-3 h-3 ml-auto shrink-0" /> : <ChevronRight className="w-3 h-3 ml-auto shrink-0" />}
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="mt-1 space-y-1">
          {tool.input !== undefined && (
            <pre className="text-xs bg-zinc-900 rounded p-2 overflow-auto text-zinc-300 border border-zinc-700 max-h-40
              [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar]:h-1
              [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-600
              [&::-webkit-scrollbar-track]:bg-transparent">
              {inputStr}
            </pre>
          )}
          {tool.result !== undefined && (
            <pre className="text-xs bg-emerald-950/30 rounded p-2 overflow-auto text-emerald-300 border border-emerald-900/50 max-h-40
              [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar]:h-1
              [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-emerald-900
              [&::-webkit-scrollbar-track]:bg-transparent">
              {typeof tool.result === 'string' ? tool.result : JSON.stringify(tool.result, null, 2)}
            </pre>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

function ThinkingView({ text }: { text: string }) {
  const [open, setOpen] = useState(false)
  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="flex items-center gap-1.5 text-xs text-purple-400 hover:text-purple-300 py-0.5 w-full text-left">
        <Brain className="w-3 h-3 shrink-0" />
        <span>Thinking...</span>
        {open ? <ChevronDown className="w-3 h-3 ml-auto shrink-0" /> : <ChevronRight className="w-3 h-3 ml-auto shrink-0" />}
      </CollapsibleTrigger>
      <CollapsibleContent>
        <pre className="mt-1 text-xs bg-purple-950/20 rounded p-2 overflow-auto text-purple-200 border border-purple-900/40 whitespace-pre-wrap max-h-48
          [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar]:h-1
          [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-purple-800/60
          [&::-webkit-scrollbar-track]:bg-transparent">
          {text}
        </pre>
      </CollapsibleContent>
    </Collapsible>
  )
}

function MessageBubble({ msg }: { msg: Message }) {
  const isUser = msg.role === 'user'
  return (
    <div className={cn('flex', isUser ? 'justify-end' : 'justify-start')}>
      <div className={cn('space-y-1 min-w-0', isUser ? 'max-w-[82%]' : 'w-full')}>
        {/* Thinking */}
        {msg.thinking && (
          <div className="bg-zinc-900 rounded-lg px-3 py-2 border border-zinc-700">
            <ThinkingView text={msg.thinking} />
          </div>
        )}

        {/* Tool calls */}
        {msg.toolCalls && msg.toolCalls.length > 0 && (
          <div className="bg-zinc-900 rounded-lg px-3 py-2 border border-zinc-700 space-y-1">
            {msg.toolCalls.map((tool, i) => (
              <ToolCallView key={i} tool={tool} />
            ))}
          </div>
        )}

        {/* Message text */}
        {(msg.text || msg.status === 'streaming') && (
          isUser ? (
            <div className="rounded-2xl px-3.5 py-2.5 text-sm leading-relaxed bg-blue-600 text-white rounded-tr-sm">
              <span className="whitespace-pre-wrap break-words">{msg.text}</span>
            </div>
          ) : msg.status === 'streaming' && !msg.text ? (
            /* 텍스트가 아직 없는 스트리밍 초기 — 궤도 애니메이션 */
            <ThinkingIndicator />
          ) : (
            <div className={cn('text-sm leading-relaxed py-1',
              msg.status === 'error' && 'bg-red-950/50 text-red-200 border border-red-800/50 rounded-lg px-3.5 py-2.5'
            )}>
              <div className="prose prose-sm prose-invert max-w-none
                prose-p:my-1 prose-p:leading-relaxed
                prose-headings:text-zinc-100 prose-headings:font-semibold prose-headings:mt-3 prose-headings:mb-1
                prose-h1:text-base prose-h2:text-sm prose-h3:text-xs
                prose-strong:text-zinc-100
                prose-code:text-amber-300 prose-code:bg-zinc-900 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:text-xs prose-code:before:content-none prose-code:after:content-none
                prose-pre:bg-zinc-900 prose-pre:border prose-pre:border-zinc-700 prose-pre:rounded-lg prose-pre:text-xs
                prose-pre:[&::-webkit-scrollbar]:h-1 prose-pre:[&::-webkit-scrollbar-thumb]:bg-zinc-600 prose-pre:[&::-webkit-scrollbar-thumb]:rounded-full prose-pre:[&::-webkit-scrollbar-track]:bg-transparent
                prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline
                prose-blockquote:border-zinc-600 prose-blockquote:text-zinc-400 prose-blockquote:not-italic
                prose-li:my-0.5 prose-ul:my-1 prose-ol:my-1
                prose-hr:border-zinc-700
                prose-table:text-xs prose-th:text-zinc-300 prose-td:text-zinc-400">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {msg.text || ''}
                </ReactMarkdown>
                {msg.status === 'streaming' && (
                  <span className="inline-block w-0.5 h-3.5 ml-0.5 bg-emerald-400 align-middle animate-pulse" />
                )}
              </div>
              {msg.status === 'error' && (
                <div className="flex items-center gap-1 mt-1 text-red-400 text-xs">
                  <AlertCircle className="w-3 h-3" />
                  오류가 발생했어요
                </div>
              )}
              {msg.status === 'aborted' && (
                <div className="flex items-center gap-1 mt-1 text-zinc-500 text-xs">
                  <Square className="w-3 h-3" />
                  중단됨
                </div>
              )}
            </div>
          )
        )}
      </div>
    </div>
  )
}

// 토큰 사용량 바
function TokenBar({ budget }: { budget: TokenBudget }) {
  const pct = Math.min(100, Math.round((budget.used / budget.total) * 100))
  const color = pct >= 90 ? 'bg-red-500' : pct >= 70 ? 'bg-amber-500' : 'bg-emerald-500'
  return (
    <div className="flex items-center gap-1.5 ml-auto">
      <div className="w-20 h-1.5 bg-zinc-700 rounded-full overflow-hidden">
        <div className={cn('h-full rounded-full transition-all', color)} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-xs text-zinc-500 tabular-nums whitespace-nowrap">
        {(budget.used / 1000).toFixed(1)}k / {(budget.total / 1000).toFixed(0)}k
      </span>
    </div>
  )
}

// 권한 요청 배너
function PermissionBanner({
  req,
  onAllow,
  onDeny,
}: {
  req: PermissionRequest
  onAllow: (alwaysAllow: boolean) => void
  onDeny: () => void
}) {
  const inputPreview = typeof req.input === 'string'
    ? req.input.slice(0, 120)
    : JSON.stringify(req.input).slice(0, 120)

  return (
    <div className="mx-3 mb-2 rounded-lg border border-amber-700/50 bg-amber-950/30 px-3 py-2 shrink-0">
      <div className="flex items-start gap-2">
        <Shield className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
        <div className="flex-1 min-w-0">
          <p className="text-xs font-medium text-amber-300">도구 실행 승인 요청</p>
          <p className="text-xs text-amber-400 font-mono mt-0.5">{req.toolName}</p>
          {inputPreview && (
            <p className="text-xs text-zinc-400 mt-0.5 truncate">{inputPreview}</p>
          )}
        </div>
      </div>
      <div className="flex gap-1.5 mt-2">
        <button
          onClick={() => onAllow(false)}
          className="flex items-center gap-1 px-2 py-1 rounded text-xs bg-emerald-700 hover:bg-emerald-600 text-white transition-colors"
        >
          <Check className="w-3 h-3" /> 허용
        </button>
        <button
          onClick={() => onAllow(true)}
          className="flex items-center gap-1 px-2 py-1 rounded text-xs bg-zinc-600 hover:bg-zinc-500 text-zinc-200 transition-colors"
        >
          <Check className="w-3 h-3" /> 항상 허용
        </button>
        <button
          onClick={onDeny}
          className="flex items-center gap-1 px-2 py-1 rounded text-xs bg-red-900 hover:bg-red-800 text-red-200 transition-colors"
        >
          <X className="w-3 h-3" /> 거부
        </button>
      </div>
    </div>
  )
}

// ── 메인 컴포넌트 ─────────────────────────────────────────────────
export default function ChatInterface() {
  const { sendMessage, onMessage } = useWebSocket()
  const { activeProject, activeSessionId, setActiveSessionId } = useProject()

  const [messages, setMessages]               = useState<Message[]>([])
  const [input, setInput]                     = useState('')
  const [isStreaming, setIsStreaming]          = useState(false)
  const [model, setModel]                     = useState<ModelId>(() => loadSettings().model)
  const [tokenBudget, setTokenBudget]         = useState<TokenBudget | null>(null)
  const [permissionReq, setPermissionReq]     = useState<PermissionRequest | null>(null)
  const [showModelMenu, setShowModelMenu]     = useState(false)
  const [showMoreModels, setShowMoreModels]   = useState(false)
  const [extendedThinking, setExtendedThinking] = useState(false)
  const [attachments, setAttachments]         = useState<AttachedFile[]>([])
  const [uploading, setUploading]             = useState(false)
  const [listening, setListening]             = useState(false)
  const [micSupported, setMicSupported]       = useState(false)
  const [historyLoading, setHistoryLoading]   = useState(false)

  const streamingIdRef  = useRef<string | null>(null)
  const scrollRef       = useRef<HTMLDivElement>(null)
  const inputRef        = useRef<HTMLTextAreaElement>(null)
  const fileInputRef    = useRef<HTMLInputElement>(null)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const recognitionRef  = useRef<any>(null)
  // 서버가 새로 할당한 session ID를 저장 — 히스토리 로드를 건너뛸 때 사용
  // boolean이 아닌 ID 값 자체를 보관해야 StrictMode 이중 실행에서도 안전
  const serverAssignedIdRef = useRef<string | null>(null)

  // 마이크 지원 여부 확인
  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const w = window as any
    setMicSupported(!!(w.SpeechRecognition || w.webkitSpeechRecognition))
  }, [])

  // 파일 처리
  const TEXT_EXTS = /\.(md|txt|ts|tsx|js|jsx|json|yaml|yml|toml|xml|csv|sh|py|go|rs|java|c|cpp|h|css|html)$/i
  const isTextFile = (f: File) =>
    f.type.startsWith('text/') || f.type.includes('json') || TEXT_EXTS.test(f.name)

  const handleFiles = useCallback(async (files: FileList | File[]) => {
    const arr = Array.from(files)
    if (!arr.length) return
    setUploading(true)
    const result: AttachedFile[] = []
    for (const file of arr) {
      const id = `att-${Date.now()}-${Math.random().toString(36).slice(2)}`
      if (isTextFile(file)) {
        const content = await file.text()
        result.push({ id, name: file.name, size: file.size, type: file.type, content })
      } else if (file.type.startsWith('image/')) {
        const preview = await new Promise<string>(res => {
          const reader = new FileReader()
          reader.onload = e => res(e.target?.result as string)
          reader.readAsDataURL(file)
        })
        result.push({ id, name: file.name, size: file.size, type: file.type, preview })
      } else {
        result.push({ id, name: file.name, size: file.size, type: file.type })
      }
    }
    setAttachments(prev => [...prev, ...result])
    setUploading(false)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const removeAttachment = useCallback((id: string) => {
    setAttachments(prev => prev.filter(a => a.id !== id))
  }, [])

  // 마이크 토글
  const toggleMic = useCallback(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const w = window as any
    const SR = w.SpeechRecognition || w.webkitSpeechRecognition
    if (!SR) return
    if (listening) {
      recognitionRef.current?.stop()
      setListening(false)
      return
    }
    const rec = new SR()
    rec.lang = 'ko-KR'
    rec.continuous = true
    rec.interimResults = false
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    rec.onresult = (e: any) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const transcript = Array.from(e.results).map((r: any) => r[0].transcript).join('')
      setInput(prev => prev + transcript)
    }
    rec.onend = () => setListening(false)
    recognitionRef.current = rec
    rec.start()
    setListening(true)
  }, [listening])

  // 세션 변경 시 히스토리 로드
  // 서버가 할당한 세션 ID(serverAssignedIdRef)와 일치하면 건너뜀 — 스트리밍 중이므로
  // ref를 소비(초기화)하지 않아야 React StrictMode 이중 실행에서도 안전
  useEffect(() => {
    if (activeSessionId !== null && serverAssignedIdRef.current === activeSessionId) {
      return  // 서버 할당 세션 — 현재 스트리밍 중, 히스토리 로드 불필요
    }
    // 사용자가 직접 세션을 변경했을 때만 도달
    serverAssignedIdRef.current = null  // ref 초기화
    setMessages([])
    setTokenBudget(null)
    setPermissionReq(null)
    if (!activeSessionId || !activeProject?.path) return

    setHistoryLoading(true)
    const params = new URLSearchParams({ projectPath: activeProject.path, sessionId: activeSessionId })
    fetch(`/api/ws/session?${params}`)
      .then(r => r.json())
      .then(data => setMessages(data))
      .catch(() => {/* */})
      .finally(() => setHistoryLoading(false))
  }, [activeSessionId, activeProject?.path])

  // 프로젝트 변경 시 초기화 (세션 없는 새 대화)
  useEffect(() => {
    if (activeSessionId === null) {
      setMessages([])
      setTokenBudget(null)
      setPermissionReq(null)
    }
  }, [activeProject?.id, activeSessionId])

  // Auto-scroll
  useEffect(() => {
    const el = scrollRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [messages])

  // WS 메시지 핸들러
  useEffect(() => {
    const unsub = onMessage((msg) => {
      const id = streamingIdRef.current

      // permission_request 는 streamingId 와 무관하게 처리
      if (msg.type === 'permission_request') {
        setPermissionReq({
          requestId: msg.requestId as string,
          toolName:  msg.toolName  as string,
          input:     msg.input,
        })
        return
      }

      if (msg.type === 'token_budget') {
        setTokenBudget({ used: msg.used as number, total: msg.total as number })
        return
      }

      if (!id) return

      if (msg.type === 'session_id') {
        // 서버가 새로 할당한 session ID를 보관 — 히스토리 로드 건너뜀에 사용
        serverAssignedIdRef.current = msg.sessionId as string
        setActiveSessionId(msg.sessionId as string)
      } else if (msg.type === 'text') {
        setMessages(prev => prev.map(m =>
          m.id === id ? { ...m, text: m.text + (msg.text as string) } : m
        ))
      } else if (msg.type === 'thinking') {
        setMessages(prev => prev.map(m =>
          m.id === id ? { ...m, thinking: (m.thinking || '') + (msg.text as string) } : m
        ))
      } else if (msg.type === 'tool_call') {
        setMessages(prev => prev.map(m =>
          m.id === id ? {
            ...m,
            toolCalls: [...(m.toolCalls || []), { name: msg.name as string, input: msg.input }],
          } : m
        ))
      } else if (msg.type === 'tool_result') {
        setMessages(prev => prev.map(m => {
          if (m.id !== id) return m
          const toolCalls = [...(m.toolCalls || [])]
          const last = toolCalls[toolCalls.length - 1]
          if (last) toolCalls[toolCalls.length - 1] = { ...last, result: msg.content }
          return { ...m, toolCalls }
        }))
      } else if (msg.type === 'done' || msg.type === 'aborted') {
        setMessages(prev => prev.map(m =>
          m.id === id ? { ...m, status: msg.type as MessageStatus } : m
        ))
        setIsStreaming(false)
        streamingIdRef.current = null
        setPermissionReq(null)
      } else if (msg.type === 'error') {
        setMessages(prev => prev.map(m =>
          m.id === id ? { ...m, status: 'error', text: m.text || String(msg.message) } : m
        ))
        setIsStreaming(false)
        streamingIdRef.current = null
        setPermissionReq(null)
      }
    })
    return unsub
  }, [onMessage, setActiveSessionId])

  const submit = useCallback(() => {
    const text = input.trim()
    if ((!text && attachments.length === 0) || isStreaming) return

    // 첨부파일을 prompt에 추가
    const parts: string[] = [text]
    for (const att of attachments) {
      if (att.content !== undefined) {
        const lang = att.name.split('.').pop() ?? ''
        parts.push(`\n\n**첨부: ${att.name}**\n\`\`\`${lang}\n${att.content}\n\`\`\``)
      } else if (att.preview) {
        parts.push(`\n\n**첨부 이미지: ${att.name}**`)
      }
    }
    const prompt = parts.join('').trim()
    const displayText = text || '(파일 첨부)'

    const userMsg: Message      = { id: `user-${Date.now()}`, role: 'user', text: displayText }
    const assistantId           = `assistant-${Date.now()}`
    const assistantMsg: Message = { id: assistantId, role: 'assistant', text: '', status: 'streaming', toolCalls: [] }

    setMessages(prev => [...prev, userMsg, assistantMsg])
    streamingIdRef.current = assistantId
    setIsStreaming(true)
    setInput('')
    setAttachments([])

    sendMessage({
      type:             'chat',
      sessionId:        activeSessionId,
      projectPath:      activeProject?.path || process.env.HOME || '/tmp',
      prompt,
      model,
      extendedThinking,
      permissionMode:   loadSettings().permissionMode,
    })

    if (inputRef.current) inputRef.current.style.height = '52px'
    inputRef.current?.focus()
  }, [input, attachments, isStreaming, sendMessage, activeSessionId, activeProject, model, extendedThinking])

  const abort = useCallback(() => {
    sendMessage({ type: 'abort' })
    if (streamingIdRef.current) {
      setMessages(prev => prev.map(m =>
        m.id === streamingIdRef.current ? { ...m, status: 'aborted' } : m
      ))
    }
    setIsStreaming(false)
    streamingIdRef.current = null
    setPermissionReq(null)
  }, [sendMessage])

  const handlePermission = useCallback((allow: boolean, alwaysAllow = false) => {
    if (!permissionReq) return
    sendMessage({ type: 'permission_response', requestId: permissionReq.requestId, allow, alwaysAllow })
    setPermissionReq(null)
  }, [permissionReq, sendMessage])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); submit() }
  }

  const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`
  }

  const currentModel   = MODELS.find(m => m.id === model) ?? MODELS[0]
  const featuredModels = MODELS.slice(0, 3)   // 상위 3개 기본 노출
  const moreModels     = MODELS.slice(3)       // 나머지는 접어서 보관

  return (
    <div className="flex flex-col h-full bg-zinc-950 rounded-lg overflow-hidden border border-zinc-800">

      {/* Header */}
      <div className="flex items-center gap-2 px-3 py-2 bg-zinc-900 border-b border-zinc-800 shrink-0">
        <Bot className="w-4 h-4 text-blue-400 shrink-0" />
        <span className="text-sm font-medium text-zinc-200 shrink-0">Claude</span>
        {activeProject && (
          <span className="text-xs text-zinc-500 truncate min-w-0">· {activeProject.name}</span>
        )}

        {/* 토큰 사용량 */}
        {tokenBudget && <div className="ml-auto"><TokenBar budget={tokenBudget} /></div>}

        {/* 세션 ID */}
        {activeSessionId && !tokenBudget && (
          <span className="text-xs text-zinc-600 font-mono shrink-0">{activeSessionId.slice(0, 8)}</span>
        )}
      </div>

      {/* 권한 요청 배너 */}
      {permissionReq && (
        <PermissionBanner
          req={permissionReq}
          onAllow={(alwaysAllow) => handlePermission(true, alwaysAllow)}
          onDeny={() => handlePermission(false)}
        />
      )}

      {/* 메시지 목록 */}
      <div
        ref={scrollRef}
        className="flex-1 min-h-0 overflow-y-auto [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar-thumb]:bg-zinc-700 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-track]:bg-transparent"
        onClick={() => setShowModelMenu(false)}
      >
        <div className="flex flex-col gap-4 p-4">
          {historyLoading ? (
            <div className="flex flex-col items-center justify-center h-48 gap-2">
              <RefreshCw className="w-5 h-5 text-zinc-600 animate-spin" />
              <p className="text-xs text-zinc-600">이전 대화 불러오는 중...</p>
            </div>
          ) : messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-48 text-center gap-3">
              <Bot className="w-12 h-12 text-zinc-700" />
              <div>
                <p className="text-sm text-zinc-400">Claude Code에게 작업을 요청하세요</p>
                <p className="text-xs text-zinc-600 mt-1">
                  {activeProject ? `프로젝트: ${activeProject.path}` : '사이드바에서 프로젝트를 선택하세요'}
                </p>
              </div>
            </div>
          ) : null}
          {messages.map(msg => <MessageBubble key={msg.id} msg={msg} />)}
        </div>
      </div>

      {/* 입력창 */}
      <div className="shrink-0 border-t border-zinc-800 bg-zinc-950 px-3 py-3">
        <div
          className="bg-zinc-800 border border-zinc-700 rounded-2xl focus-within:border-zinc-500 transition-colors"
          onDragOver={e => e.preventDefault()}
          onDrop={e => { e.preventDefault(); handleFiles(e.dataTransfer.files) }}
        >
          {/* 첨부 파일 칩 */}
          {(attachments.length > 0 || uploading) && (
            <div className="flex flex-wrap gap-1.5 px-3 pt-2.5">
              {attachments.map(att => (
                <div key={att.id} className="flex items-center gap-1.5 bg-zinc-700 border border-zinc-600 rounded-lg pl-2 pr-1 py-1 text-xs group">
                  {att.preview
                    ? <img src={att.preview} alt={att.name} className="w-5 h-5 rounded object-cover shrink-0" />
                    : att.content !== undefined
                    ? <FileText className="w-3.5 h-3.5 shrink-0 text-zinc-400" />
                    : <ImageIcon className="w-3.5 h-3.5 shrink-0 text-zinc-400" />
                  }
                  <span className="truncate max-w-[120px] text-zinc-300">{att.name}</span>
                  <span className="text-zinc-500 text-[10px]">{(att.size / 1024).toFixed(0)}K</span>
                  <button onClick={() => removeAttachment(att.id)} className="ml-0.5 text-zinc-500 hover:text-red-400 transition-colors">
                    <X className="w-3 h-3" />
                  </button>
                </div>
              ))}
              {uploading && (
                <div className="flex items-center gap-1.5 bg-zinc-700 border border-zinc-600 rounded-lg px-2 py-1 text-xs text-zinc-500">
                  <Loader2 className="w-3 h-3 animate-spin" /> 처리 중...
                </div>
              )}
            </div>
          )}

          {/* 텍스트 입력 */}
          <textarea
            ref={inputRef}
            value={input}
            onChange={handleInput}
            onKeyDown={handleKeyDown}
            placeholder={isStreaming ? '응답 중...' : '작업 내용을 입력하세요...'}
            rows={1}
            disabled={isStreaming}
            className="w-full bg-transparent px-4 pt-3 pb-2 text-sm text-zinc-100 placeholder:text-zinc-500 outline-none resize-none min-h-[52px] max-h-[200px] disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ height: '52px' }}
          />

          {/* 하단 툴바 */}
          <div className="flex items-center gap-2 px-3 pb-3">
            {/* 숨겨진 파일 입력 */}
            <input
              ref={fileInputRef}
              type="file"
              multiple
              accept="image/*,.pdf,.txt,.md,.ts,.tsx,.js,.jsx,.json,.yaml,.yml,.toml,.xml,.csv,.sh,.py,.go,.rs,.java,.c,.cpp,.h,.css,.html"
              className="hidden"
              onChange={e => { if (e.target.files) handleFiles(e.target.files); e.target.value = '' }}
            />
            {/* + 첨부 버튼 */}
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              title="파일 첨부"
              className="w-8 h-8 rounded-full border border-zinc-600 flex items-center justify-center text-zinc-400 hover:text-zinc-200 hover:border-zinc-400 transition-colors"
            >
              <Plus className="w-4 h-4" />
            </button>

            <div className="flex-1" />

            {/* 모델 선택 */}
            <div className="relative">
              <button
                type="button"
                onClick={() => setShowModelMenu(v => !v)}
                className="flex items-center gap-1.5 h-8 px-3 rounded-full bg-zinc-700 hover:bg-zinc-600 text-xs text-zinc-200 transition-colors whitespace-nowrap"
              >
                <span className="font-medium">{currentModel.label}</span>
                {extendedThinking && <span className="text-zinc-400">확장</span>}
                <ChevronDown className="w-3 h-3 text-zinc-400" />
              </button>

              {showModelMenu && (
                <div className="absolute right-0 bottom-full mb-2 z-20 bg-zinc-900 border border-zinc-700 rounded-2xl shadow-2xl w-72 py-2 overflow-hidden">

                  {/* 기본 모델 3개 */}
                  {featuredModels.map(m => (
                    <button
                      key={m.id}
                      onClick={() => { setModel(m.id); setShowModelMenu(false) }}
                      className="flex items-center w-full px-4 py-3 text-left hover:bg-zinc-800/60 transition-colors"
                    >
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-semibold text-zinc-100 leading-snug">{m.label}</p>
                        <p className="text-xs text-zinc-500 mt-0.5 leading-snug">{m.description}</p>
                      </div>
                      <span className="w-6 flex justify-end shrink-0">
                        {model === m.id && <Check className="w-4 h-4 text-blue-400" />}
                      </span>
                    </button>
                  ))}

                  {/* 더 많은 모델 토글 */}
                  {moreModels.length > 0 && (
                    <>
                      <button
                        type="button"
                        onClick={() => setShowMoreModels(v => !v)}
                        className="flex items-center w-full px-4 py-2.5 text-sm text-zinc-400 hover:bg-zinc-800/60 hover:text-zinc-200 transition-colors"
                      >
                        <span className="flex-1 text-left">더 많은 모델</span>
                        <ChevronRight className={cn('w-4 h-4 transition-transform', showMoreModels && 'rotate-90')} />
                      </button>

                      {showMoreModels && moreModels.map(m => (
                        <button
                          key={m.id}
                          onClick={() => { setModel(m.id); setShowModelMenu(false) }}
                          className="flex items-center w-full px-4 py-3 text-left hover:bg-zinc-800/60 transition-colors bg-zinc-800/30"
                        >
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-semibold text-zinc-100 leading-snug">{m.label}</p>
                            <p className="text-xs text-zinc-500 mt-0.5 leading-snug">{m.description}</p>
                          </div>
                          <span className="w-6 flex justify-end shrink-0">
                            {model === m.id && <Check className="w-4 h-4 text-blue-400" />}
                          </span>
                        </button>
                      ))}
                    </>
                  )}

                  {/* 확장 사고 — 둥근 카드 블록 */}
                  <div className="mx-2 mt-1 mb-1 rounded-xl bg-zinc-800 px-3 py-3">
                    <div className="flex items-center gap-3">
                      <div className="flex-1">
                        <p className="text-sm font-medium text-zinc-100">확장 사고</p>
                        <p className="text-xs text-zinc-500 mt-0.5">복잡한 작업을 위해 더 오래 사고</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setExtendedThinking(v => !v)}
                        className={cn(
                          'relative w-11 h-6 rounded-full transition-colors duration-200 shrink-0',
                          extendedThinking ? 'bg-blue-500' : 'bg-zinc-600'
                        )}
                      >
                        <span className={cn(
                          'absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform duration-200',
                          extendedThinking ? 'translate-x-5' : 'translate-x-0'
                        )} />
                      </button>
                    </div>
                  </div>

                </div>
              )}
            </div>

            {/* 마이크 버튼 */}
            {micSupported && (
              <button
                type="button"
                onClick={toggleMic}
                title={listening ? '음성 인식 중지' : '음성으로 입력'}
                className={cn(
                  'w-8 h-8 rounded-full border flex items-center justify-center transition-all',
                  listening
                    ? 'border-red-500 bg-red-500/20 text-red-400 animate-pulse'
                    : 'border-zinc-600 text-zinc-400 hover:text-zinc-200 hover:border-zinc-400'
                )}
              >
                {listening ? <MicOff className="w-4 h-4" /> : <Mic className="w-4 h-4" />}
              </button>
            )}

            {/* 전송 / 중단 버튼 */}
            {isStreaming ? (
              <button
                type="button"
                onClick={abort}
                className="w-9 h-9 rounded-xl bg-red-600 hover:bg-red-500 flex items-center justify-center transition-colors"
              >
                <Square className="w-3.5 h-3.5 text-white" />
              </button>
            ) : (
              <button
                type="button"
                onClick={submit}
                disabled={!input.trim() && attachments.length === 0}
                className="w-9 h-9 rounded-xl bg-orange-500 hover:bg-orange-400 disabled:bg-zinc-700 disabled:cursor-not-allowed flex items-center justify-center transition-colors"
              >
                <Send className="w-4 h-4 text-white" />
              </button>
            )}
          </div>
        </div>
        <p className="text-xs text-zinc-600 mt-1.5 text-center">Enter 전송 · Shift+Enter 줄바꿈</p>
      </div>
    </div>
  )
}
