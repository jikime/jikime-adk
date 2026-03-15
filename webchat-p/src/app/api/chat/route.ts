import { NextRequest } from 'next/server'
import { spawn, execSync } from 'child_process'
import { readFileSync, existsSync } from 'fs'
import { join } from 'path'
import os from 'os'

export const dynamic = 'force-dynamic'

/**
 * 슬래시 커맨드 감지: /namespace:command args 또는 /command args
 * 반환: { namespace, command, args } 또는 null
 */
function parseSlashCommand(message: string): { namespace: string; command: string; args: string } | null {
  if (!message.startsWith('/')) return null
  const [head, ...rest] = message.slice(1).split(' ')
  const args = rest.join(' ')
  if (head.includes(':')) {
    const [namespace, command] = head.split(':')
    if (namespace && command) return { namespace, command, args }
  }
  return null
}

/**
 * 슬래시 커맨드 MD 파일을 찾아 읽고 프롬프트로 변환
 * 탐색 순서: {cwd}/.claude/commands/{ns}/{cmd}.md → ~/.claude/commands/{ns}/{cmd}.md
 */
function resolveSlashCommandPrompt(
  namespace: string, command: string, args: string, cwd: string
): { prompt: string; found: boolean; filePath: string } {
  const relPath = join('.claude', 'commands', namespace, `${command}.md`)
  const candidates = [
    join(cwd, relPath),
    join(os.homedir(), '.claude', 'commands', namespace, `${command}.md`),
  ]

  for (const filePath of candidates) {
    if (!existsSync(filePath)) continue

    let content = readFileSync(filePath, 'utf8')

    // 1. frontmatter 제거 (--- ... --- 블록)
    content = content.replace(/^---[\s\S]*?---\n?/, '').trim()

    // 2. $ARGUMENTS 치환
    content = content.replace(/\$ARGUMENTS/g, args)

    // 3. @filepath 디렉티브 처리 — 파일이 존재하면 내용 인라인, 없으면 줄 제거
    content = content.replace(/^@(.+)$/gm, (_, filePath: string) => {
      const abs = filePath.startsWith('/')
        ? filePath
        : join(cwd, filePath.trim())
      if (existsSync(abs)) {
        try {
          return `\`\`\`\n${readFileSync(abs, 'utf8').trim()}\n\`\`\``
        } catch { return '' }
      }
      return ''  // 존재하지 않는 파일 참조는 제거
    })

    // 4. !command 디렉티브 처리 — 실행 후 결과 인라인, 실패 시 줄 제거
    content = content.replace(/^!(.+)$/gm, (_, cmd: string) => {
      try {
        const out = execSync(cmd.trim(), { cwd, timeout: 5000, stdio: ['ignore', 'pipe', 'pipe'] }).toString().trim()
        return out ? `\`\`\`\n${out}\n\`\`\`` : ''
      } catch { return '' }
    })

    // 5. 연속된 빈 줄 정리
    content = content.replace(/\n{3,}/g, '\n\n').trim()

    // 6. AskUserQuestion 감지 → headless 모드 지시문 prepend
    //    claude -p 모드에서는 AskUserQuestion이 작동하지 않으므로
    //    자동으로 합리적인 기본값으로 진행하도록 안내
    if (content.includes('AskUserQuestion')) {
      const headlessPrefix = [
        '> **[Headless Mode]** You are running in non-interactive `-p` mode.',
        '> `AskUserQuestion` is NOT available. Do NOT call it.',
        '> Instead: infer reasonable defaults from the codebase context, proceed automatically, and summarize decisions made.',
        '',
      ].join('\n')
      content = headlessPrefix + content
    }

    return { prompt: content, found: true, filePath }
  }

  return { prompt: '', found: false, filePath: '' }
}

function findClaude(): string {
  const candidates = [
    'claude',
    `${process.env.HOME}/.local/bin/claude`,
    '/usr/local/bin/claude',
    '/opt/homebrew/bin/claude',
  ]
  for (const c of candidates) {
    try {
      execSync(`${c} --version`, { stdio: 'pipe', timeout: 3000 })
      return c
    } catch { /* continue */ }
  }
  return 'claude'
}

let cachedClaude: string | null = null
function getClaude(): string {
  if (!cachedClaude) cachedClaude = findClaude()
  return cachedClaude
}

/**
 * SSE 이벤트 타입:
 *   { type: 'session_id', id }
 *   { type: 'tool_call',  name, input }   ← 도구 호출 시작
 *   { type: 'tool_result', content }      ← 도구 실행 결과
 *   { type: 'text', text }                ← 최종 답변 텍스트
 *   { type: 'error', text }
 *   { type: 'done' }
 */
export async function POST(request: NextRequest) {
  const body = await request.json()
  const message = (body.message as string)?.trim()
  const sessionId = (body.sessionId as string) || ''
  const cwd = (body.cwd as string) || process.env.HOME || '/tmp'
  // 첨부 이미지/파일 경로 목록 (클라이언트에서 --add-file로 전달)
  const imagePaths = ((body.imagePaths as string[]) || []).filter(Boolean)

  if (!message && imagePaths.length === 0) {
    return Response.json({ error: 'No message' }, { status: 400 })
  }

  // 슬래시 커맨드 처리: /namespace:command [args] → MD 파일 로드 후 프롬프트 교체
  let finalMessage = message || '첨부된 파일을 확인해주세요.'
  const slashCmd = parseSlashCommand(message)
  if (slashCmd) {
    const { prompt, found, filePath } = resolveSlashCommandPrompt(
      slashCmd.namespace, slashCmd.command, slashCmd.args, cwd
    )
    if (found) {
      console.log(`[slash-cmd] ${message} → ${filePath}`)
      finalMessage = prompt
    } else {
      // MD 파일 없음 → 에러 즉시 반환
      const encoder = new TextEncoder()
      const body = new ReadableStream({
        start(controller) {
          const send = (d: object) => controller.enqueue(encoder.encode(`data: ${JSON.stringify(d)}\n\n`))
          send({ type: 'error', text: `슬래시 커맨드 파일을 찾을 수 없어요: /${slashCmd.namespace}:${slashCmd.command}\n탐색 경로: ${cwd}/.claude/commands/${slashCmd.namespace}/${slashCmd.command}.md` })
          send({ type: 'done' })
          controller.close()
        },
      })
      return new Response(body, {
        headers: { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache', 'X-Accel-Buffering': 'no' },
      })
    }
  }

  const claudePath = getClaude()
  console.log(`[chat] cwd=${cwd} slash=${!!slashCmd} files=${imagePaths.length} promptLen=${finalMessage.length}`)

  /** claude spawn 인수 생성 — withResume: false이면 --resume 없이 새 세션 시작 */
  function buildArgs(withResume: boolean): string[] {
    const prompt = imagePaths.length > 0
      ? finalMessage + '\n' + imagePaths.map(p => `@${p}`).join('\n')
      : finalMessage
    const a = ['-p', prompt, '--output-format', 'stream-json', '--verbose', '--dangerously-skip-permissions']
    if (withResume && sessionId && !slashCmd) a.push('--resume', sessionId)
    return a
  }

  const env: Record<string, string> = {}
  for (const [k, v] of Object.entries(process.env)) {
    if (k !== 'CLAUDECODE' && v !== undefined) env[k] = v
  }
  env.FORCE_COLOR = '0'
  env.NO_COLOR = '1'
  env.TERM = 'dumb'
  env.PATH = (process.env.PATH ?? '') + `:${process.env.HOME}/.local/bin`

  const encoder = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      const send = (data: object) => {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch { /* closed */ }
      }

      let buf = ''
      let sentText = false
      let extractedSessionId = ''
      let stderrBuf = ''
      let retried = false  // "No conversation found" 재시도 방지 플래그
      let proc: ReturnType<typeof spawn>

      function startProc(withResume: boolean) {
        buf = ''; sentText = false; extractedSessionId = ''; stderrBuf = ''

        try {
          proc = spawn(claudePath, buildArgs(withResume), { cwd, env: env as NodeJS.ProcessEnv, stdio: ['ignore', 'pipe', 'pipe'] })
        } catch (err) {
          send({ type: 'error', text: `실행 오류: ${(err as Error).message}` })
          send({ type: 'done' })
          try { controller.close() } catch { /* */ }
          return
        }

        // 재시도 시 이전 프로세스의 close 이벤트가 스트림을 닫지 않도록
        // 현재 프로세스 인스턴스를 캡처해 close 핸들러에서 비교
        const activeProc = proc

        proc.stdout?.on('data', (chunk: Buffer) => {
          buf += chunk.toString()
          const lines = buf.split('\n')
          buf = lines.pop() || ''

          for (const line of lines) {
            if (!line.trim()) continue
            console.log('[claude stdout]', line.slice(0, 300))
            try {
              const event = JSON.parse(line)

              // session_id 추출
              if (event.session_id && !extractedSessionId) {
                extractedSessionId = event.session_id
                send({ type: 'session_id', id: extractedSessionId })
              }

              // assistant 이벤트 — thinking / tool_use / text 블록 처리
              if (event.type === 'assistant' && Array.isArray(event.message?.content)) {
                for (const block of event.message.content) {
                  if (block.type === 'tool_use') {
                    const inputStr = formatToolInput(block.name, block.input)
                    send({ type: 'tool_call', name: block.name as string, input: inputStr })
                  }
                  if (block.type === 'text' && block.text) {
                    send({ type: 'text', text: block.text as string })
                    sentText = true
                  }
                }
              }

              // user 이벤트 — tool_result 블록 처리
              if (event.type === 'user' && Array.isArray(event.message?.content)) {
                for (const block of event.message.content) {
                  if (block.type === 'tool_result') {
                    const { text, images } = extractToolResult(block.content)
                    send({ type: 'tool_result', content: text, ...(images.length > 0 ? { images } : {}) })
                  }
                }
              }

              // result 이벤트
              if (event.type === 'result') {
                if (event.is_error) {
                  console.log('[claude result error]', JSON.stringify(event))

                  // errors 배열에서 "No conversation found" 감지 → --resume 없이 재시도
                  const errors = (event.errors as string[]) || []
                  const isNotFound = errors.some(e => typeof e === 'string' && e.includes('No conversation found'))
                  if (isNotFound && !retried) {
                    retried = true
                    console.log('[chat] session not found, retrying without --resume')
                    try { proc.kill('SIGTERM') } catch { /* */ }
                    startProc(false)
                    return
                  }

                  if (!sentText) {
                    // errors 배열 우선 → result → fallback
                    const errText = errors.join('\n') || (event.result as string) || ''
                    const extra = stderrBuf ? `\n[stderr]\n${stderrBuf}` : ''
                    send({ type: 'error', text: errText + extra || `실행 중 오류가 발생했어요. (subtype=${event.subtype ?? 'unknown'})` })
                    sentText = true
                  }
                } else if (event.result && !sentText) {
                  send({ type: 'text', text: event.result as string })
                  sentText = true
                }
              }
            } catch { /* invalid JSON */ }
          }
        })

        proc.stderr?.on('data', (chunk: Buffer) => {
          const text = chunk.toString()
          stderrBuf += text
          console.error('[claude stderr]', text.trim())
        })

        proc.on('close', (code) => {
          // 재시도로 새 프로세스가 시작됐으면 이 핸들러는 건너뜀
          if (proc !== activeProc) return
          if (!sentText) {
            if (code !== 0) {
              send({ type: 'error', text: `오류가 발생했어요 (exit ${code})` })
            } else {
              send({ type: 'error', text: '응답이 없어요. 슬래시 커맨드(/)는 -p 모드에서 지원되지 않을 수 있어요.' })
            }
          }
          send({ type: 'done' })
          try { controller.close() } catch { /* */ }
        })

        proc.on('error', (err) => {
          if (proc !== activeProc) return
          send({ type: 'error', text: `claude를 찾을 수 없어요: ${err.message}` })
          send({ type: 'done' })
          try { controller.close() } catch { /* */ }
        })
      }

      startProc(true)  // 첫 시도: --resume 포함

      request.signal.addEventListener('abort', () => {
        try { proc.kill('SIGTERM') } catch { /* */ }
      })
    },
  })

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache, no-transform',
      'Connection': 'keep-alive',
      'X-Accel-Buffering': 'no',
    },
  })
}

/** 도구 입력을 사람이 읽기 쉬운 문자열로 변환 */
function formatToolInput(toolName: string, input: Record<string, unknown>): string {
  if (!input) return ''
  // Bash: command 필드만 표시
  if (toolName === 'Bash' && input.command) return String(input.command)
  // Read/Write/Edit: file_path 우선
  if (input.file_path) return String(input.file_path)
  if (input.path) return String(input.path)
  // 나머지는 JSON
  return JSON.stringify(input, null, 2)
}

/** tool_result content를 텍스트 + 이미지 배열로 분리 추출 */
function extractToolResult(content: unknown): { text: string; images: { data: string; mediaType: string }[] } {
  const images: { data: string; mediaType: string }[] = []

  if (typeof content === 'string') return { text: content, images }

  if (Array.isArray(content)) {
    const texts: string[] = []
    for (const c of content) {
      if (typeof c === 'string') {
        texts.push(c)
      } else if (c && typeof c === 'object') {
        const obj = c as Record<string, unknown>
        if (obj.type === 'text' && obj.text) {
          texts.push(String(obj.text))
        } else if (obj.type === 'image') {
          const src = obj.source as Record<string, unknown> | undefined
          if (src?.type === 'base64' && src.data && src.media_type) {
            images.push({ data: String(src.data), mediaType: String(src.media_type) })
          }
        } else {
          const text = (obj as { text?: string }).text
          if (text) texts.push(text)
        }
      }
    }
    return { text: texts.join('\n'), images }
  }

  return { text: JSON.stringify(content ?? ''), images }
}
