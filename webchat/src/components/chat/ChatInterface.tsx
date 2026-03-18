'use client'

import { useState, useRef, useCallback, useEffect, useMemo, memo } from 'react'
import {
  Send, Square, Bot, ChevronDown, ChevronRight,
  Wrench, AlertCircle, Brain, Shield, Check, X, RefreshCw,
  Plus, Mic, MicOff, FileText, Image as ImageIcon, Loader2,
  Trash2, HelpCircle, Terminal, RotateCcw,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { useWebSocket } from '@/contexts/WebSocketContext'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import { MODELS, loadSettings, type ModelId } from '@/components/sidebar/Sidebar'
import { useLocale } from '@/contexts/LocaleContext'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

// ── 슬래시 커맨드 정의 ────────────────────────────────────────────
type SlashCommandType = 'client' | 'passthrough'

interface SlashCommand {
  name: string          // e.g. "clear"
  description: string
  usage: string         // e.g. "/clear"
  type: SlashCommandType
  icon: React.ReactNode
  args?: string         // optional argument hint, e.g. "<topic>"
  context?: string      // optional context hint from frontmatter
}

const SLASH_COMMANDS: SlashCommand[] = [
  {
    name: 'clear',
    description: 'Clear conversation and start a new session',
    usage: '/clear',
    type: 'client',
    icon: <Trash2 className="w-3.5 h-3.5" />,
  },
  {
    name: 'new',
    description: 'Start a new session (alias for /clear)',
    usage: '/new',
    type: 'client',
    icon: <RotateCcw className="w-3.5 h-3.5" />,
  },
  {
    name: 'help',
    description: 'Show all available slash commands',
    usage: '/help',
    type: 'client',
    icon: <HelpCircle className="w-3.5 h-3.5" />,
  },
]

const HELP_TEXT = `**Available Slash Commands**

| Command | Description |
|---------|-------------|
| \`/clear\` | Clear conversation and start a new session |
| \`/new\` | Start a new session (alias for /clear) |
| \`/help\` | Show this help message |

> 💡 **Tip**: Type \`/\` to open the command menu and use ↑↓ to navigate.`

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

const ThinkingIndicator = memo(function ThinkingIndicator() {
  const { t } = useLocale()
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
        <span className="text-xs text-muted-foreground flex items-center gap-1.5">
          {t.chat.thinking}
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
})

const ToolCallView = memo(function ToolCallView({ tool }: { tool: ToolCall }) {
  const [open, setOpen] = useState(false)
  const inputStr = typeof tool.input === 'string'
    ? tool.input
    : JSON.stringify(tool.input, null, 2)

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="flex items-center gap-1.5 text-xs dark:text-amber-400 dark:hover:text-amber-300 text-amber-600 hover:text-amber-700 py-0.5 w-full text-left">
        <Wrench className="w-3 h-3 shrink-0" />
        <span className="font-mono truncate">{tool.name}</span>
        {open ? <ChevronDown className="w-3 h-3 ml-auto shrink-0" /> : <ChevronRight className="w-3 h-3 ml-auto shrink-0" />}
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="mt-1 space-y-1">
          {tool.input !== undefined && (
            <pre className="text-xs bg-background rounded p-2 overflow-auto text-foreground/80 border border-border max-h-40
              [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar]:h-1
              [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-border
              [&::-webkit-scrollbar-track]:bg-transparent">
              {inputStr}
            </pre>
          )}
          {tool.result !== undefined && (
            <pre className="text-xs dark:bg-emerald-950/30 dark:text-emerald-300 dark:border-emerald-900/50 bg-emerald-50 text-emerald-800 border border-emerald-200 rounded p-2 overflow-auto max-h-40
              [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar]:h-1
              [&::-webkit-scrollbar-thumb]:rounded-full dark:[&::-webkit-scrollbar-thumb]:bg-emerald-900 [&::-webkit-scrollbar-thumb]:bg-emerald-200
              [&::-webkit-scrollbar-track]:bg-transparent">
              {typeof tool.result === 'string' ? tool.result : JSON.stringify(tool.result, null, 2)}
            </pre>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
})

const ThinkingView = memo(function ThinkingView({ text }: { text: string }) {
  const [open, setOpen] = useState(false)
  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="flex items-center gap-1.5 text-xs dark:text-purple-400 dark:hover:text-purple-300 text-purple-600 hover:text-purple-700 py-0.5 w-full text-left">
        <Brain className="w-3 h-3 shrink-0" />
        <span>Thinking...</span>
        {open ? <ChevronDown className="w-3 h-3 ml-auto shrink-0" /> : <ChevronRight className="w-3 h-3 ml-auto shrink-0" />}
      </CollapsibleTrigger>
      <CollapsibleContent>
        <pre className="mt-1 text-xs dark:bg-purple-950/20 dark:text-purple-200 dark:border-purple-900/40 bg-purple-50 text-purple-800 border border-purple-200 rounded p-2 overflow-auto whitespace-pre-wrap max-h-48
          [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar]:h-1
          [&::-webkit-scrollbar-thumb]:rounded-full dark:[&::-webkit-scrollbar-thumb]:bg-purple-800/60 [&::-webkit-scrollbar-thumb]:bg-purple-200
          [&::-webkit-scrollbar-track]:bg-transparent">
          {text}
        </pre>
      </CollapsibleContent>
    </Collapsible>
  )
})

const MessageBubble = memo(function MessageBubble({ msg }: { msg: Message }) {
  const isUser = msg.role === 'user'
  return (
    <div className={cn('flex', isUser ? 'justify-end' : 'justify-start')}>
      <div className={cn('space-y-1 min-w-0', isUser ? 'max-w-[82%]' : 'w-full')}>
        {/* Thinking */}
        {msg.thinking && (
          <div className="bg-card rounded-lg px-3 py-2 border border-border">
            <ThinkingView text={msg.thinking} />
          </div>
        )}

        {/* Tool calls */}
        {msg.toolCalls && msg.toolCalls.length > 0 && (
          <div className="bg-card rounded-lg px-3 py-2 border border-border space-y-1">
            {msg.toolCalls.map((tool, i) => (
              <ToolCallView key={i} tool={tool} />
            ))}
          </div>
        )}

        {/* Message text */}
        {(msg.text || msg.status === 'streaming') && (
          isUser ? (
            <div className="rounded-2xl px-4 py-3 text-base leading-relaxed bg-blue-600 text-white rounded-tr-sm">
              <span className="whitespace-pre-wrap break-words">{msg.text}</span>
            </div>
          ) : msg.status === 'streaming' && !msg.text ? (
            /* 텍스트가 아직 없는 스트리밍 초기 — 궤도 애니메이션 */
            <ThinkingIndicator />
          ) : (
            <div className={cn('text-base leading-relaxed py-1',
              msg.status === 'error' && 'dark:bg-red-950/50 dark:text-red-200 dark:border-red-800/50 bg-red-50 text-red-800 border border-red-200 rounded-lg px-3.5 py-2.5'
            )}>
              <div className="prose prose-base dark:prose-invert max-w-none
                prose-p:my-1.5 prose-p:leading-relaxed
                prose-headings:text-foreground prose-headings:font-semibold prose-headings:mt-4 prose-headings:mb-1.5
                prose-h1:text-xl prose-h2:text-lg prose-h3:text-base
                prose-strong:text-foreground
                dark:prose-code:text-amber-300 prose-code:text-amber-700 prose-code:bg-muted prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:text-sm prose-code:before:content-none prose-code:after:content-none
                prose-pre:bg-muted prose-pre:border prose-pre:border-border prose-pre:rounded-lg prose-pre:text-sm
                prose-pre:[&::-webkit-scrollbar]:h-1 prose-pre:[&::-webkit-scrollbar-thumb]:bg-border prose-pre:[&::-webkit-scrollbar-thumb]:rounded-full prose-pre:[&::-webkit-scrollbar-track]:bg-transparent
                dark:prose-a:text-blue-400 prose-a:text-blue-600 prose-a:no-underline hover:prose-a:underline
                prose-blockquote:border-border prose-blockquote:text-muted-foreground prose-blockquote:not-italic
                prose-li:my-0.5 prose-ul:my-1 prose-ol:my-1
                prose-hr:border-border
                prose-table:text-xs prose-th:text-foreground/80 prose-td:text-muted-foreground">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {msg.text || ''}
                </ReactMarkdown>
                {msg.status === 'streaming' && (
                  <span className="inline-block w-0.5 h-3.5 ml-0.5 bg-emerald-400 align-middle animate-pulse" />
                )}
              </div>
              {msg.status === 'error' && (
                <div className="flex items-center gap-1 mt-1 dark:text-red-400 text-red-600 text-xs">
                  <AlertCircle className="w-3 h-3" />
                  오류가 발생했어요
                </div>
              )}
            </div>
          )
        )}
      </div>
    </div>
  )
}, (prev, next) =>
  prev.msg.id === next.msg.id &&
  prev.msg.text === next.msg.text &&
  prev.msg.status === next.msg.status &&
  prev.msg.thinking === next.msg.thinking &&
  prev.msg.toolCalls === next.msg.toolCalls
)

// 토큰 사용량 바
const TokenBar = memo(function TokenBar({ budget }: { budget: TokenBudget }) {
  const pct = Math.min(100, Math.round((budget.used / budget.total) * 100))
  const color = pct >= 90 ? 'bg-red-500' : pct >= 70 ? 'bg-amber-500' : 'bg-emerald-500'
  return (
    <div className="flex items-center gap-1.5 ml-auto">
      <div className="w-20 h-1.5 bg-muted rounded-full overflow-hidden">
        <div className={cn('h-full rounded-full transition-all', color)} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-xs text-muted-foreground tabular-nums whitespace-nowrap">
        {(budget.used / 1000).toFixed(1)}k / {(budget.total / 1000).toFixed(0)}k
      </span>
    </div>
  )
})

// 권한 요청 배너
const PermissionBanner = memo(function PermissionBanner({
  req,
  onAllow,
  onDeny,
}: {
  req: PermissionRequest
  onAllow: (alwaysAllow: boolean) => void
  onDeny: () => void
}) {
  const { t } = useLocale()
  const inputPreview = typeof req.input === 'string'
    ? req.input.slice(0, 120)
    : JSON.stringify(req.input).slice(0, 120)

  return (
    <div className="mx-3 mb-2 rounded-lg border border-amber-700/50 bg-amber-950/30 px-3 py-2 shrink-0">
      <div className="flex items-start gap-2">
        <Shield className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
        <div className="flex-1 min-w-0">
          <p className="text-xs font-medium text-amber-300">{t.chat.permissionRequest}</p>
          <p className="text-xs text-amber-400 font-mono mt-0.5">{req.toolName}</p>
          {inputPreview && (
            <p className="text-xs text-muted-foreground mt-0.5 truncate">{inputPreview}</p>
          )}
        </div>
      </div>
      <div className="flex gap-1.5 mt-2">
        <Button
          size="sm"
          onClick={() => onAllow(false)}
          className="h-6 px-2 text-xs bg-emerald-700 hover:bg-emerald-600 text-white"
        >
          <Check className="w-3 h-3" /> {t.chat.allow}
        </Button>
        <Button
          size="sm" variant="outline"
          onClick={() => onAllow(true)}
          className="h-6 px-2 text-xs"
        >
          <Check className="w-3 h-3" /> {t.chat.alwaysAllow}
        </Button>
        <Button
          size="sm"
          onClick={onDeny}
          className="h-6 px-2 text-xs bg-red-900 hover:bg-red-800 text-red-200 border-0"
        >
          <X className="w-3 h-3" /> {t.chat.deny}
        </Button>
      </div>
    </div>
  )
})

// ── 메인 컴포넌트 ─────────────────────────────────────────────────
export default function ChatInterface() {
  const { t } = useLocale()
  const { sendMessage, onMessage } = useWebSocket()
  const { activeProject, activeSessionId, setActiveSessionId } = useProject()
  const { getApiUrl } = useServer()

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

  // 슬래시 커맨드 드롭다운
  const [slashOpen, setSlashOpen]     = useState(false)
  const [slashFilter, setSlashFilter] = useState('')
  const [slashIndex, setSlashIndex]   = useState(0)
  const [projectCommands, setProjectCommands] = useState<SlashCommand[]>([])
  const [tipPos, setTipPos] = useState<{ x: number; y: number; text: string } | null>(null)

  const streamingIdRef    = useRef<string | null>(null)
  const lastPromptRef     = useRef<string>('')          // 재시도용 마지막 프롬프트
  const scrollRef         = useRef<HTMLDivElement>(null)
  const inputRef          = useRef<HTMLTextAreaElement>(null)
  const fileInputRef      = useRef<HTMLInputElement>(null)
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
  useEffect(() => {
    if (activeSessionId !== null && serverAssignedIdRef.current === activeSessionId) {
      return
    }
    serverAssignedIdRef.current = null
    setMessages([])
    setTokenBudget(null)
    setPermissionReq(null)
    if (!activeSessionId || !activeProject?.path) return

    setHistoryLoading(true)
    const params = new URLSearchParams({ projectPath: activeProject.path, sessionId: activeSessionId })
    fetch(getApiUrl(`/api/ws/session?${params}`))
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

    // 클라이언트 슬래시 커맨드 처리 (Enter로 직접 실행 시)
    const clientCmds = ['clear', 'new', 'help']
    if (text.startsWith('/')) {
      const cmdName = text.slice(1).split(' ')[0].toLowerCase()
      if (clientCmds.includes(cmdName)) {
        runClientCommand(cmdName)
        return
      }
    }

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

    lastPromptRef.current = prompt

    sendMessage({
      type:             'chat',
      sessionId:        activeSessionId,
      projectPath:      activeProject?.path || process.env.HOME || '/tmp',
      prompt,
      model,
      extendedThinking,
      permissionMode:   loadSettings().permissionMode,
    })

    if (inputRef.current) inputRef.current.style.height = '64px'
    inputRef.current?.focus()
  }, [input, attachments, isStreaming, sendMessage, activeSessionId, activeProject, model, extendedThinking])

  const abort = useCallback(() => {
    // updater 함수가 실행될 시점에 ref는 이미 null이 되므로
    // 반드시 먼저 로컬 변수에 캡처해야 함
    const abortingId = streamingIdRef.current
    sendMessage({ type: 'abort' })
    if (abortingId) {
      setMessages(prev => prev.map(m =>
        m.id === abortingId ? { ...m, status: 'aborted' } : m
      ))
    }
    setIsStreaming(false)
    streamingIdRef.current = null
    setPermissionReq(null)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sendMessage])

  // 중단된 응답 재시도 — 마지막 어시스턴트 메시지 제거 후 동일 프롬프트 재전송
  const retry = useCallback(() => {
    const prompt = lastPromptRef.current
    if (!prompt || isStreaming) return

    // 마지막 aborted 어시스턴트 메시지 제거
    setMessages(prev => {
      const lastAbortedIdx = [...prev].reverse().findIndex(m => m.role === 'assistant' && m.status === 'aborted')
      if (lastAbortedIdx === -1) return prev
      const idx = prev.length - 1 - lastAbortedIdx
      return prev.filter((_, i) => i !== idx)
    })

    const assistantId = `assistant-${Date.now()}`
    const assistantMsg: Message = { id: assistantId, role: 'assistant', text: '', status: 'streaming', toolCalls: [] }
    setMessages(prev => [...prev, assistantMsg])
    streamingIdRef.current = assistantId
    setIsStreaming(true)

    sendMessage({
      type:           'chat',
      sessionId:      activeSessionId,
      projectPath:    activeProject?.path || process.env.HOME || '/tmp',
      prompt,
      model,
      extendedThinking,
      permissionMode: loadSettings().permissionMode,
    })
    inputRef.current?.focus()
  }, [sendMessage, isStreaming, activeSessionId, activeProject, model, extendedThinking])

  const handlePermission = useCallback((allow: boolean, alwaysAllow = false) => {
    if (!permissionReq) return
    sendMessage({ type: 'permission_response', requestId: permissionReq.requestId, allow, alwaysAllow })
    setPermissionReq(null)
  }, [permissionReq, sendMessage])

  // 프로젝트 변경 시 .claude/commands/jikime 커맨드 로드
  useEffect(() => {
    if (!activeProject?.path) { setProjectCommands([]); return }
    fetch(getApiUrl(`/api/ws/commands?projectPath=${encodeURIComponent(activeProject.path)}`))
      .then(r => r.json())
      .then((data: Array<{ name: string; description: string; argumentHint: string; context: string }>) => {
        const cmds: SlashCommand[] = data.map(c => ({
          name:        c.name,
          description: c.description || c.name,
          usage:       `/jikime:${c.name}`,
          type:        'passthrough' as const,
          icon:        <Terminal className="w-3.5 h-3.5" />,
          args:        c.argumentHint || undefined,
          context:     c.context || undefined,
        }))
        setProjectCommands(cmds)
      })
      .catch(() => setProjectCommands([]))
  }, [activeProject?.path, getApiUrl])

  // 슬래시 커맨드 필터링: 고정 커맨드 + 프로젝트 커맨드 합산
  const allCommands = useMemo(
    () => [...SLASH_COMMANDS, ...projectCommands],
    [projectCommands],
  )
  const filteredCommands = useMemo(
    () => allCommands.filter(c => c.usage.slice(1).startsWith(slashFilter.toLowerCase())),
    [allCommands, slashFilter],
  )

  const closeSlash = useCallback(() => {
    setSlashOpen(false)
    setSlashFilter('')
    setSlashIndex(0)
  }, [])

  // 클라이언트 커맨드 실행 (서버 전송 없음)
  const runClientCommand = useCallback((name: string) => {
    if (name === 'clear' || name === 'new') {
      setActiveSessionId(null)
      setMessages(prev => {
        void prev
        return [{
          id: `sys-${Date.now()}`,
          role: 'assistant' as const,
          text: '✅ Conversation cleared. Starting a new session.',
          status: 'done' as const,
        }]
      })
    } else if (name === 'help') {
      setMessages(prev => [...prev, {
        id: `sys-${Date.now()}`,
        role: 'assistant' as const,
        text: HELP_TEXT,
        status: 'done' as const,
      }])
    }
    setInput('')
    closeSlash()
    if (inputRef.current) inputRef.current.style.height = '64px'
    inputRef.current?.focus()
  }, [setActiveSessionId, closeSlash])

  // 슬래시 메뉴에서 커맨드 선택
  const selectSlashCommand = useCallback((cmd: SlashCommand) => {
    if (cmd.type === 'client') {
      runClientCommand(cmd.name)
    } else {
      // 패스스루: 전체 usage(/jikime:name)를 입력창에 채워줌 (인자 입력 대기)
      const text = `${cmd.usage}${cmd.args ? ' ' : ''}`
      setInput(text)
      closeSlash()
      inputRef.current?.focus()
      // 커서를 끝으로
      setTimeout(() => {
        if (inputRef.current) {
          inputRef.current.selectionStart = text.length
          inputRef.current.selectionEnd   = text.length
        }
      }, 0)
    }
  }, [runClientCommand, closeSlash])

  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // 슬래시 메뉴가 열려있을 때 키보드 탐색
    if (slashOpen && filteredCommands.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSlashIndex(i => (i + 1) % filteredCommands.length)
        return
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSlashIndex(i => (i - 1 + filteredCommands.length) % filteredCommands.length)
        return
      }
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        selectSlashCommand(filteredCommands[slashIndex])
        return
      }
      if (e.key === 'Escape') {
        e.preventDefault()
        closeSlash()
        return
      }
      if (e.key === 'Tab') {
        e.preventDefault()
        selectSlashCommand(filteredCommands[slashIndex])
        return
      }
    }
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); submit() }
  }, [slashOpen, filteredCommands, slashIndex, selectSlashCommand, closeSlash, submit])

  const handleInput = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value
    setInput(val)
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`

    // 슬래시 커맨드 감지: 입력이 '/'로 시작하고 공백 없는 경우
    if (val.startsWith('/') && !val.includes(' ') && val.length >= 1) {
      const filter = val.slice(1)  // '/' 이후 문자
      setSlashFilter(filter)
      setSlashIndex(0)
      setSlashOpen(true)
    } else {
      closeSlash()
    }
  }, [closeSlash])

  const { currentModel, featuredModels, moreModels } = useMemo(() => ({
    currentModel:   MODELS.find(m => m.id === model) ?? MODELS[0],
    featuredModels: MODELS.slice(0, 3),
    moreModels:     MODELS.slice(3),
  }), [model])

  return (
    <div className="flex flex-col h-full bg-muted dark:bg-background rounded-lg overflow-hidden border border-border">

      {/* Header */}
      <div className="flex items-center gap-3 px-5 py-3 bg-white dark:bg-accent border-b border-border shrink-0">
        <Bot className="w-5 h-5 text-blue-400 shrink-0" />
        <span className="text-base font-semibold text-foreground shrink-0">Claude</span>
        {activeProject && (
          <span className="text-sm text-muted-foreground truncate min-w-0">· {activeProject.name}</span>
        )}

        {tokenBudget && <div className="ml-auto"><TokenBar budget={tokenBudget} /></div>}

        {activeSessionId && !tokenBudget && (
          <span className="text-xs text-muted-foreground/50 font-mono shrink-0">{activeSessionId.slice(0, 8)}</span>
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
        className="flex-1 min-h-0 overflow-y-auto [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar-thumb]:bg-border [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-track]:bg-transparent"
        onClick={() => setShowModelMenu(false)}
      >
        <div className="flex flex-col gap-6 p-6 min-h-full">
          {historyLoading ? (
            <div className="flex flex-col items-center justify-center flex-1 min-h-64 gap-3">
              <RefreshCw className="w-6 h-6 text-muted-foreground/50 animate-spin" />
              <p className="text-sm text-muted-foreground/50">{t.chat.loadingHistory}</p>
            </div>
          ) : messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center flex-1 min-h-64 text-center gap-5">
              <div className="w-20 h-20 rounded-full bg-blue-500/10 border border-blue-500/20 flex items-center justify-center">
                <Bot className="w-10 h-10 text-blue-400/70" />
              </div>
              <div className="space-y-2">
                <p className="text-lg font-semibold text-foreground/80">{t.chat.emptyState}</p>
                <p className="text-sm text-muted-foreground">
                  {activeProject ? t.chat.projectLabel(activeProject.path) : t.chat.selectProject}
                </p>
              </div>
            </div>
          ) : null}
          {messages.map(msg => (
            <div key={msg.id}>
              <MessageBubble msg={msg} />
              {msg.role === 'assistant' && msg.status === 'aborted' && (
                <div className="flex items-center gap-3 mt-2 px-3 py-2.5 rounded-xl border border-border bg-card text-sm text-foreground/70">
                  <AlertCircle className="w-4 h-4 shrink-0 text-muted-foreground" />
                  <span className="flex-1">{t.chat.abortedBanner}</span>
                  <button
                    type="button"
                    onClick={retry}
                    disabled={isStreaming}
                    className="shrink-0 px-3 py-1 rounded-lg border border-border text-sm font-medium text-foreground/80 hover:bg-muted transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                  >
                    {t.chat.retry}
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* 입력창 */}
      <div className="shrink-0 border-t border-border bg-white dark:bg-background px-4 py-4">
        <div className="relative">

        {/* ── 슬래시 커맨드 드롭다운 ─────────────────────────────── */}
        {slashOpen && filteredCommands.length > 0 && (
          <div className="absolute bottom-full mb-2 left-0 z-30 w-80 bg-card border border-border rounded-xl shadow-xl">
            {/* 헤더 */}
            <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border bg-muted/50">
              <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">Slash Commands</span>
              <span className="ml-auto text-[10px] text-muted-foreground">↑↓ · Enter · Esc</span>
            </div>
            {/* 스크롤 영역 */}
            <div className="max-h-[240px] overflow-y-auto overflow-x-hidden
              [&::-webkit-scrollbar]:w-1
              [&::-webkit-scrollbar-track]:bg-transparent
              [&::-webkit-scrollbar-thumb]:bg-border
              [&::-webkit-scrollbar-thumb]:rounded-full
              [&::-webkit-scrollbar-thumb:hover]:bg-muted-foreground/40">
              {filteredCommands.map((cmd, i) => (
                <button
                  key={`${cmd.type}-${cmd.name}`}
                  type="button"
                  onMouseDown={e => { e.preventDefault(); selectSlashCommand(cmd) }}
                  onMouseEnter={() => setSlashIndex(i)}
                  className={cn(
                    'flex items-center gap-2.5 w-full px-3 py-2 text-left transition-colors',
                    i === slashIndex
                      ? 'bg-primary/10 text-foreground'
                      : 'hover:bg-muted text-foreground/80'
                  )}
                >
                  {/* 타입 배지 + 아이콘 */}
                  <span className={cn(
                    'flex items-center justify-center w-5 h-5 rounded shrink-0',
                    cmd.type === 'client'
                      ? 'bg-amber-500/15 text-amber-500 dark:text-amber-400'
                      : 'bg-blue-500/15 text-blue-500 dark:text-blue-400'
                  )}>
                    {cmd.icon}
                  </span>
                  {/* 커맨드명 */}
                  <span className="font-mono text-sm font-medium truncate flex-1">{cmd.usage}</span>
                  {/* args ? 배지 — 툴팁은 fixed로 overflow 탈출 */}
                  {cmd.args && (
                    <span
                      className="shrink-0 flex items-center justify-center w-4 h-4 rounded-full text-[10px] font-bold
                        bg-muted-foreground/15 text-muted-foreground hover:bg-muted-foreground/30 cursor-default"
                      onMouseEnter={e => {
                        const r = e.currentTarget.getBoundingClientRect()
                        setTipPos({ x: r.right + 8, y: r.top + r.height / 2, text: cmd.args! })
                      }}
                      onMouseLeave={() => setTipPos(null)}
                    >
                      ?
                    </span>
                  )}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* args fixed 툴팁 — overflow 클리핑 완전 탈출 */}
        {tipPos && (
          <div
            className="fixed z-[200] px-2 py-1 rounded-md text-xs whitespace-nowrap pointer-events-none
              bg-popover text-popover-foreground border border-border shadow-md"
            style={{ left: tipPos.x, top: tipPos.y, transform: 'translateY(-50%)' }}
          >
            {tipPos.text}
          </div>
        )}

        <div
          className={cn(
            'bg-white dark:bg-muted border rounded-2xl shadow-sm transition-colors',
            isStreaming
              ? 'border-foreground/25 ring-2 ring-foreground/5'
              : 'border-border focus-within:border-ring'
          )}
          onDragOver={e => e.preventDefault()}
          onDrop={e => { e.preventDefault(); handleFiles(e.dataTransfer.files) }}
        >
          {/* 첨부 파일 칩 */}
          {(attachments.length > 0 || uploading) && (
            <div className="flex flex-wrap gap-1.5 px-3 pt-2.5">
              {attachments.map(att => (
                <div key={att.id} className="flex items-center gap-1.5 bg-accent border border-border rounded-lg pl-2 pr-1 py-1 text-xs group">
                  {att.preview
                    ? <img src={att.preview} alt={att.name} className="w-5 h-5 rounded object-cover shrink-0" />
                    : att.content !== undefined
                    ? <FileText className="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
                    : <ImageIcon className="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
                  }
                  <span className="truncate max-w-[120px] text-foreground/80">{att.name}</span>
                  <span className="text-muted-foreground/60 text-[10px]">{(att.size / 1024).toFixed(0)}K</span>
                  <button onClick={() => removeAttachment(att.id)} className="ml-0.5 text-muted-foreground hover:text-red-400 transition-colors">
                    <X className="w-3 h-3" />
                  </button>
                </div>
              ))}
              {uploading && (
                <div className="flex items-center gap-1.5 bg-accent border border-border rounded-lg px-2 py-1 text-xs text-muted-foreground">
                  <Loader2 className="w-3 h-3 animate-spin" /> 처리 중...
                </div>
              )}
            </div>
          )}

          {/* 텍스트 입력 */}
          <Textarea
            ref={inputRef}
            value={input}
            onChange={handleInput}
            onKeyDown={handleKeyDown}
            placeholder={t.chat.placeholder}
            rows={1}
            className="w-full bg-transparent border-0 ring-0 shadow-none px-5 pt-4 pb-2 text-base text-foreground placeholder:text-muted-foreground outline-none resize-none min-h-[64px] max-h-[240px] disabled:opacity-50 disabled:cursor-not-allowed focus-visible:ring-0 focus-visible:border-0"
            style={{ height: '64px', fieldSizing: 'fixed' } as React.CSSProperties}
          />

          {/* 하단 툴바 */}
          <div className="flex items-center gap-2 px-4 pb-4">
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
            <Button
              type="button" variant="outline" size="icon"
              onClick={() => fileInputRef.current?.click()}
              title={t.chat.attachFile}
              className="w-9 h-9 rounded-full"
            >
              <Plus className="w-5 h-5" />
            </Button>

            <div className="flex-1" />

            {/* 모델 선택 */}
            <div className="relative">
              <button
                type="button"
                onClick={() => setShowModelMenu(v => !v)}
                className="flex items-center gap-1.5 h-9 px-4 rounded-full bg-accent hover:bg-accent/80 text-sm text-foreground transition-colors whitespace-nowrap"
              >
                <span className="font-medium">{currentModel.label}</span>
                {extendedThinking && <span className="text-muted-foreground">{t.chat.extendedThinking}</span>}
                <ChevronDown className="w-3 h-3 text-muted-foreground" />
              </button>

              {showModelMenu && (
                <div className="absolute right-0 bottom-full mb-2 z-20 bg-card border border-border rounded-2xl shadow-2xl w-72 py-2 overflow-hidden">

                  {/* 기본 모델 3개 */}
                  {featuredModels.map(m => (
                    <button
                      key={m.id}
                      onClick={() => { setModel(m.id); setShowModelMenu(false) }}
                      className="flex items-center w-full px-4 py-3 text-left hover:bg-muted/60 transition-colors"
                    >
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-semibold text-foreground leading-snug">{m.label}</p>
                        <p className="text-xs text-muted-foreground mt-0.5 leading-snug">{t.sidebar.models[m.descKey]}</p>
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
                        className="flex items-center w-full px-4 py-2.5 text-sm text-muted-foreground hover:bg-muted/60 hover:text-foreground transition-colors"
                      >
                        <span className="flex-1 text-left">{t.chat.moreModels}</span>
                        <ChevronRight className={cn('w-4 h-4 transition-transform', showMoreModels && 'rotate-90')} />
                      </button>

                      {showMoreModels && moreModels.map(m => (
                        <button
                          key={m.id}
                          onClick={() => { setModel(m.id); setShowModelMenu(false) }}
                          className="flex items-center w-full px-4 py-3 text-left hover:bg-muted/60 transition-colors bg-muted/30"
                        >
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-semibold text-foreground leading-snug">{m.label}</p>
                            <p className="text-xs text-muted-foreground mt-0.5 leading-snug">{t.sidebar.models[m.descKey]}</p>
                          </div>
                          <span className="w-6 flex justify-end shrink-0">
                            {model === m.id && <Check className="w-4 h-4 text-blue-400" />}
                          </span>
                        </button>
                      ))}
                    </>
                  )}

                  {/* 확장 사고 — 둥근 카드 블록 */}
                  <div className="mx-2 mt-1 mb-1 rounded-xl bg-muted px-3 py-3">
                    <div className="flex items-center gap-3">
                      <div className="flex-1">
                        <p className="text-sm font-medium text-foreground">{t.chat.extendedThinkingTitle}</p>
                        <p className="text-xs text-muted-foreground mt-0.5">{t.chat.extendedThinkingDesc}</p>
                      </div>
                      <Switch
                        checked={extendedThinking}
                        onCheckedChange={setExtendedThinking}
                        className="shrink-0"
                      />
                    </div>
                  </div>

                </div>
              )}
            </div>

            {/* 마이크 버튼 */}
            {micSupported && (
              <Button
                type="button" variant="outline" size="icon"
                onClick={toggleMic}
                title={listening ? t.chat.voiceStop : t.chat.voiceStart}
                className={cn(
                  'w-9 h-9 rounded-full',
                  listening && 'border-red-500 bg-red-500/20 text-red-400 animate-pulse'
                )}
              >
                {listening ? <MicOff className="w-4 h-4" /> : <Mic className="w-4 h-4" />}
              </Button>
            )}

            {/* 전송 / 중단 버튼 */}
            {isStreaming ? (
              <button
                type="button"
                onClick={abort}
                title={t.chat.stopResponding}
                className="relative flex items-center justify-center w-10 h-10 rounded-xl border border-border bg-background hover:bg-muted transition-colors group"
              >
                <span className="absolute inset-0 rounded-xl bg-foreground/10 animate-ping" />
                <Square className="relative w-3.5 h-3.5 text-foreground/80 group-hover:text-foreground fill-current transition-colors" />
              </button>
            ) : (
              <Button
                type="button" size="icon"
                onClick={submit}
                disabled={!input.trim() && attachments.length === 0}
                className="w-10 h-10 rounded-xl dark:bg-blue-600 dark:hover:bg-blue-500 bg-blue-700 hover:bg-blue-600 disabled:bg-muted disabled:cursor-not-allowed text-white disabled:text-muted-foreground"
              >
                <Send className="w-5 h-5" />
              </Button>
            )}
          </div>
        </div>
        <p className="text-xs text-muted-foreground mt-1.5 text-center">Enter 전송 · Shift+Enter 줄바꿈 · / 슬래시 커맨드</p>
        </div>{/* relative wrapper */}
      </div>
    </div>
  )
}
