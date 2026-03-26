/**
 * team.ts — jikime team orchestration HTTP API + SSE
 *
 * Provides REST endpoints and Server-Sent Events for the Team Board UI.
 *
 * HTTP API:
 *   GET    /api/team/list                        — list all teams
 *   GET    /api/team/:name                       — team status snapshot
 *   POST   /api/team/create                      — create team (delegates to jikime team create)
 *   DELETE /api/team/:name                       — stop + remove team
 *   GET    /api/team/:name/tasks                 — list tasks (filter: ?status=&agent=)
 *   POST   /api/team/:name/tasks                 — create task
 *   PATCH  /api/team/:name/tasks/:id             — update task
 *   GET    /api/team/:name/agents                — list agents
 *   POST   /api/team/:name/inbox/send            — send message
 *   GET    /api/team/:name/budget                — budget summary
 *   GET    /api/team/:name/events                — SSE stream
 */

import * as fs   from 'fs'
import * as path from 'path'
import * as os   from 'os'
import { execFile } from 'child_process'
import type { IncomingMessage, ServerResponse } from 'http'

// ── Helpers ────────────────────────────────────────────────────────

function dataDir(): string {
  return process.env.JIKIME_DATA_DIR || path.join(os.homedir(), '.jikime')
}

function teamsDir(): string {
  return path.join(dataDir(), 'teams')
}

function teamDir(name: string): string {
  return path.join(teamsDir(), name)
}

function readJSON<T>(filePath: string, fallback: T): T {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8')) as T
  } catch {
    return fallback
  }
}

function corsHeaders(): Record<string, string> {
  return {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET,POST,PATCH,DELETE,OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type',
  }
}

function jsonReply(res: ServerResponse, status: number, body: unknown): void {
  res.writeHead(status, { 'Content-Type': 'application/json', ...corsHeaders() })
  res.end(JSON.stringify(body))
}

const PARSE_BODY_LIMIT = 10 * 1024 * 1024  // 10 MB
const NAME_RE = /^[a-zA-Z0-9_-]{1,80}$/

function parseBody(req: IncomingMessage): Promise<unknown> {
  return new Promise((resolve) => {
    // 슬로우 클라이언트 DoS 방지 — 30초 내 전체 바디 수신 없으면 연결 끊김
    req.setTimeout(30_000, () => { req.destroy(); resolve({}) })
    const chunks: Buffer[] = []
    let totalSize = 0
    req.on('data', (chunk: Buffer) => {
      totalSize += chunk.length
      if (totalSize > PARSE_BODY_LIMIT) { req.destroy(); resolve({}); return }
      chunks.push(chunk)
    })
    req.on('end', () => {
      try {
        resolve(JSON.parse(Buffer.concat(chunks).toString('utf8')))
      } catch {
        resolve({})
      }
    })
  })
}

// ── TeamFileStore ──────────────────────────────────────────────────

export class TeamFileStore {
  constructor(private readonly root: string = teamsDir()) {}

  listTeams(projectPath?: string): string[] {
    if (!fs.existsSync(this.root)) return []
    // withFileTypes: true — 개별 statSync 호출 없이 isDirectory() 판별 가능
    const all = fs.readdirSync(this.root, { withFileTypes: true })
      .filter(e => e.isDirectory()).map(e => e.name)
    if (!projectPath) return all
    let realQuery = projectPath
    try { realQuery = fs.realpathSync(projectPath) } catch { /* use as-is */ }
    // 팀별 realpathSync 결과를 로컬 캐시 — 반복 syscall 방지
    const rpCache = new Map<string, string>()
    const resolve = (p: string): string => {
      if (!rpCache.has(p)) {
        try { rpCache.set(p, fs.realpathSync(p)) } catch { rpCache.set(p, p) }
      }
      return rpCache.get(p)!
    }
    return all.filter((name) => {
      const meta = this.getWebchatMeta(name)
      if (!meta.projectPath) return false
      return resolve(meta.projectPath as string) === realQuery
    })
  }

  getConfig(name: string): Record<string, unknown> {
    return readJSON(path.join(teamDir(name), 'config.json'), {})
  }

  getWebchatMeta(name: string): { projectPath?: string } {
    return readJSON(path.join(teamDir(name), 'webchat.json'), {})
  }

  writeWebchatMeta(name: string, meta: { projectPath?: string }): void {
    const dir = teamDir(name)
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true })
    // 임시 파일 → rename 원자 쓰기 — 크래시/동시 쓰기 시 webchat.json 손상 방지
    const target = path.join(dir, 'webchat.json')
    const tmp    = `${target}.tmp.${process.pid}`
    try {
      fs.writeFileSync(tmp, JSON.stringify(meta), 'utf8')
      fs.renameSync(tmp, target)
    } catch (e) {
      try { fs.unlinkSync(tmp) } catch { /* */ }
      throw e
    }
  }

  listTasks(name: string, status?: string, agent?: string, owner?: string, limit = 1000): unknown[] {
    const dir = path.join(teamDir(name), 'tasks')
    if (!fs.existsSync(dir)) return []
    // withFileTypes — isDirectory() 체크로 빈 디렉터리 엔트리 제거
    const files = fs.readdirSync(dir, { withFileTypes: true })
      .filter(e => !e.isDirectory() && e.name.endsWith('.json'))
      .map(e => e.name)
    // limit 먼저 적용 — 100K 태스크 전체 로딩으로 인한 OOM 방지
    const results: unknown[] = []
    for (const f of files) {
      if (results.length >= limit) break
      const t = readJSON<Record<string, unknown>>(path.join(dir, f), {})
      if (status && t['status']   !== status) continue
      if (agent  && t['agent_id'] !== agent)  continue
      if (owner  && t['owner']    !== owner)   continue
      results.push(t)
    }
    return results
  }

  getTask(name: string, taskID: string): unknown | null {
    const dir = path.join(teamDir(name), 'tasks')
    if (!fs.existsSync(dir)) return null
    const files = fs.readdirSync(dir).filter((f) => f.startsWith(taskID) && f.endsWith('.json'))
    if (files.length === 0) return null
    return readJSON(path.join(dir, files[0]), null)
  }

  listAgents(name: string): unknown[] {
    const dir = path.join(teamDir(name), 'registry')
    if (!fs.existsSync(dir)) return []
    return fs.readdirSync(dir)
      .filter((f) => f.endsWith('.json'))
      .map((f) => readJSON<Record<string, unknown>>(path.join(dir, f), {}))
  }

  getCosts(name: string): { total: number; agents: Record<string, { tokens: number }> } {
    const dir = path.join(teamDir(name), 'costs')
    const result: { total: number; agents: Record<string, { tokens: number }> } = {
      total: 0,
      agents: {},
    }
    if (!fs.existsSync(dir)) return result
    fs.readdirSync(dir)
      .filter((f) => f.endsWith('.json'))
      .forEach((f) => {
        const ev = readJSON<Record<string, unknown>>(path.join(dir, f), {})
        const agentID = ev['agent_id'] as string || 'unknown'
        const tokens  = ((ev['input_tokens'] as number) || 0) + ((ev['output_tokens'] as number) || 0)
        result.total += tokens
        if (!result.agents[agentID]) result.agents[agentID] = { tokens: 0 }
        result.agents[agentID].tokens += tokens
      })
    return result
  }

  getInboxMessages(name: string, agentID: string): unknown[] {
    const dir = path.join(teamDir(name), 'inbox', agentID)
    if (!fs.existsSync(dir)) return []
    return fs.readdirSync(dir)
      .filter((f) => f.endsWith('.json'))
      .sort()
      .map((f) => readJSON<unknown>(path.join(dir, f), {}))
  }
}

// ── SSE Manager ────────────────────────────────────────────────────

interface SSEClient {
  res: ServerResponse
  teamName: string
}

const sseClients: Set<SSEClient> = new Set()
const sseWatchers: Map<string, fs.FSWatcher[]> = new Map()

// 전역 SSE 클라이언트 상한 — N팀 × per-team 500 합산 DoS 방지
const MAX_GLOBAL_SSE_CLIENTS = 5000

// 30초마다 끊어진 SSE 클라이언트 정리 (기존 5분 → 30초로 단축)
setInterval(() => {
  // [...sseClients] 스냅샷 — 순회 중 다른 핸들러의 동시 삭제 충돌 방지
  for (const client of [...sseClients]) {
    if (client.res.destroyed || client.res.writableEnded) {
      sseClients.delete(client)
    }
  }
}, 30 * 1000).unref()

// 팀 이름별 브로드캐스트 디바운스 타이머 — 연속 파일 변경 시 burst 방지
const _broadcastDebounce = new Map<string, ReturnType<typeof setTimeout>>()

function startTeamWatcher(teamName: string): void {
  // 진입 시 즉시 NAME_RE 검증 — 경로 탈출 방지
  if (!NAME_RE.test(teamName)) throw new Error(`Invalid team name: ${teamName}`)
  if (sseWatchers.has(teamName)) return

  const watchDirs = ['tasks', 'registry', 'inbox', 'costs'].map((d) =>
    path.join(teamDir(teamName), d)
  )
  const watchers: fs.FSWatcher[] = []
  for (const dir of watchDirs) {
    if (!fs.existsSync(dir)) continue
    const w = fs.watch(dir, { recursive: false }, () => {
      // 50ms 디바운스 — 연속 이벤트를 하나로 묶음
      const existing = _broadcastDebounce.get(teamName)
      if (existing) clearTimeout(existing)
      _broadcastDebounce.set(teamName, setTimeout(() => {
        _broadcastDebounce.delete(teamName)
        broadcastTeamEvent(teamName)
      }, 50))
    })
    watchers.push(w)
  }
  sseWatchers.set(teamName, watchers)
}

function stopTeamWatcher(teamName: string): void {
  // 대기 중인 debounce 타이머 정리 — 타이머 객체 누수 방지
  const pending = _broadcastDebounce.get(teamName)
  if (pending) { clearTimeout(pending); _broadcastDebounce.delete(teamName) }
  const watchers = sseWatchers.get(teamName)
  if (watchers) {
    watchers.forEach((w) => w.close())
    sseWatchers.delete(teamName)
  }
  // 팀 캐시 즉시 삭제 — 삭제 후 재생성 시 오염된 캐시 사용 방지
  _teamDataCache.delete(teamName)
  _eventLogCache.delete(`evlog:${teamName}`)
}

// 이벤트 로그 파일 tail 캐시 — 3초 TTL (매 broadcast마다 readFileSync 방지)
const _eventLogCache = new Map<string, { lines: string[]; ts: number }>()
// tasks/agents/config 캐시 — 2초 TTL (listTasks O(n) 반복 호출 방지)
const _teamDataCache = new Map<string, {
  tasks:  Array<Record<string, unknown>>
  agents: Array<Record<string, unknown>>
  config: Record<string, unknown>
  ts:     number
}>()
const EVENT_LOG_TTL = 3000

function broadcastTeamEvent(teamName: string): void {
  // 팀이 삭제된 후 FSWatcher 이벤트가 지연 도착하는 경우 조기 종료
  if (!fs.existsSync(teamDir(teamName))) return

  const store   = new TeamFileStore()
  const nowData = Date.now()
  const dataCached = _teamDataCache.get(teamName)
  let tasksList: Array<Record<string, unknown>>
  let agentsList: Array<Record<string, unknown>>
  let config: Record<string, unknown>
  if (dataCached && nowData - dataCached.ts < 2000) {
    tasksList  = dataCached.tasks
    agentsList = dataCached.agents
    config     = dataCached.config
  } else {
    tasksList  = store.listTasks(teamName)  as Array<Record<string, unknown>>
    agentsList = store.listAgents(teamName) as Array<Record<string, unknown>>
    config     = store.getConfig(teamName)  as Record<string, unknown>
    _teamDataCache.set(teamName, { tasks: tasksList, agents: agentsList, config, ts: nowData })
  }

  // Task summary counts
  const taskSummary: Record<string, number> = {
    pending: 0, in_progress: 0, done: 0, failed: 0, blocked: 0,
  }
  for (const t of tasksList) {
    const s = (t['status'] as string) || 'unknown'
    if (s in taskSummary) taskSummary[s]++
  }

  // Leader name
  let leaderName = (config['leader_id'] as string) || ''
  if (!leaderName) {
    const leader = agentsList.find((a) => a['role'] === 'leader')
    leaderName = (leader?.['id'] as string) || ''
  }

  // Members with inbox count (count unread files in inbox/<agentID>/)
  const members = agentsList.map((a) => {
    const agentID   = a['id'] as string
    const inboxDir  = path.join(teamDir(teamName), 'inbox', agentID)
    let inboxCount  = 0
    try {
      inboxCount = fs.readdirSync(inboxDir).filter((f) => f.endsWith('.json')).length
    } catch { /* dir may not exist yet */ }
    return {
      name:       agentID,
      agentType:  (a['role'] as string) || 'worker',
      inboxCount,
    }
  })

  // Event log messages (last 50 from inbox/event-log.jsonl) — teamName 키 TTL 캐시
  const eventLogFile = path.join(teamDir(teamName), 'inbox', 'event-log.jsonl')
  const messages: Array<{ from: string; to: string; type: string; timestamp: string; content: string }> = []
  try {
    const now = Date.now()
    const cacheKey = `evlog:${teamName}`   // 절대 경로 대신 teamName 기반 키
    const cached = _eventLogCache.get(cacheKey)
    const tailLines = (cached && now - cached.ts < EVENT_LOG_TTL)
      ? cached.lines
      : (() => {
          // 2 MB 초과 시 마지막 500줄로 자동 truncate — 무제한 증가 방지
          const MAX_EVLOG_BYTES = 2 * 1024 * 1024
          try {
            const evStat = fs.statSync(eventLogFile)
            if (evStat.size > MAX_EVLOG_BYTES) {
              const all  = fs.readFileSync(eventLogFile, 'utf8').split('\n').filter(Boolean)
              const kept = all.slice(-500)
              fs.writeFileSync(eventLogFile, kept.join('\n') + '\n', 'utf8')
              console.log(`[team] event-log.jsonl auto-truncated: ${(evStat.size / 1024).toFixed(0)} KB → ${kept.length} lines`)
              const tail = kept.slice(-50)
              _eventLogCache.set(cacheKey, { lines: tail, ts: now })
              return tail
            }
          } catch { /* statSync/truncate 실패 시 무시 — 아래 readFileSync 계속 */ }
          const raw = fs.readFileSync(eventLogFile, 'utf8').split('\n').filter(Boolean)
          const tail = raw.slice(-50)
          _eventLogCache.set(cacheKey, { lines: tail, ts: now })
          return tail
        })()
    for (const line of tailLines) {
      try {
        const m = JSON.parse(line) as Record<string, unknown>
        messages.push({
          from:      (m['from']     as string) || '',
          to:        (m['to']       as string) || '',
          type:      (m['kind']     as string) || 'direct',
          timestamp: (m['sent_at']  as string) || '',
          content:   (m['body']     as string) || '',
        })
      } catch { /* skip malformed line */ }
    }
  } catch { /* log file may not exist yet */ }

  const payload = JSON.stringify({
    type: 'update',
    team: {
      name:        teamName,
      leaderName,
      description: (config['description'] as string) || '',
    },
    time:        new Date().toISOString(),
    tasks:       tasksList,
    agents:      agentsList,
    members,
    taskSummary,
    messages,
  })

  const MAX_SSE_PER_TEAM = 500
  let teamClientCount = 0
  for (const client of sseClients) {
    if (client.teamName !== teamName) continue
    if (++teamClientCount > MAX_SSE_PER_TEAM) {
      // 한도 초과 클라이언트는 종료 처리
      try { client.res.end() } catch { /* */ }
      sseClients.delete(client)
      continue
    }
    // 백프레셔 확인 후 쓰기
    if (client.res.destroyed || client.res.writableEnded) {
      sseClients.delete(client)
      continue
    }
    try {
      // write() 반환 false = 소켓 버퍼 가득 참 → 클라이언트 제거로 메모리 누수 방지
      const ok = client.res.write(`data: ${payload}\n\n`)
      if (!ok) sseClients.delete(client)
    } catch {
      sseClients.delete(client)
    }
  }
}

// ── Route Handler ──────────────────────────────────────────────────

export function handleTeamRoutes(
  req: IncomingMessage,
  res: ServerResponse,
  pathname: string,
): boolean {
  if (!pathname.startsWith('/api/team') && !pathname.startsWith('/api/template')) return false

  const method = req.method?.toUpperCase() || 'GET'

  // CORS preflight (shared early exit)
  if (method === 'OPTIONS') {
    res.writeHead(204, corsHeaders())
    res.end()
    return true
  }

  // GET /api/template/list
  if (method === 'GET' && pathname === '/api/template/list') {
    const templatesDir = path.join(os.homedir(), '.jikime', 'templates')
    const templates: Array<{ name: string; description: string; defaultBudget: number }> = []
    try {
      const files = fs.readdirSync(templatesDir).filter((f) => f.endsWith('.yaml'))
      for (const f of files) {
        const content = fs.readFileSync(path.join(templatesDir, f), 'utf8')
        const name        = content.match(/^name:\s*(.+)$/m)?.[1]?.trim().replace(/^['"]|['"]$/g, '') ?? f.replace('.yaml', '')
        const description = content.match(/^description:\s*["']?([^"'\n]+)["']?$/m)?.[1]?.trim() ?? ''
        const budgetMatch = content.match(/^default_budget:\s*(\d+)$/m)
        const defaultBudget = budgetMatch ? parseInt(budgetMatch[1]) : 0
        templates.push({ name, description, defaultBudget })
      }
    } catch { /* dir may not exist yet */ }
    jsonReply(res, 200, { templates })
    return true
  }

  // GET /api/template/:name/yaml — 템플릿 YAML 원문 조회
  const tmplYamlMatch = pathname.match(/^\/api\/template\/([^/]+)\/yaml$/)
  if (method === 'GET' && tmplYamlMatch) {
    const tmplName = decodeURIComponent(tmplYamlMatch[1] || '')
    // 디코딩 후 재검증 — %2F 등 URL 인코딩 경로 탈출 방지
    if (!NAME_RE.test(tmplName)) { jsonReply(res, 400, { error: 'Invalid template name' }); return true }
    const filePath = path.join(os.homedir(), '.jikime', 'templates', `${tmplName}.yaml`)
    try {
      if (!fs.existsSync(filePath)) {
        jsonReply(res, 404, { error: 'Template not found' }); return true
      }
      const yaml = fs.readFileSync(filePath, 'utf8')
      jsonReply(res, 200, { yaml })
    } catch (e) {
      jsonReply(res, 500, { error: String(e) })
    }
    return true
  }

  // POST /api/template/save — 새 템플릿 YAML 저장
  if (method === 'POST' && pathname === '/api/template/save') {
    parseBody(req).then((body) => {
      const b = body as Record<string, unknown>
      const name = (b['name'] as string || '').trim().replace(/[^a-zA-Z0-9_-]/g, '-')
      const yaml = b['yaml'] as string
      if (!name || !yaml) {
        jsonReply(res, 400, { error: 'name and yaml are required' }); return
      }
      // 1 MB 상한 — 대용량 YAML로 인한 디스크 소진 방지
      if (yaml.length > 1 * 1024 * 1024) {
        jsonReply(res, 413, { error: 'YAML exceeds 1 MB limit' }); return
      }
      const templatesDir = path.join(os.homedir(), '.jikime', 'templates')
      try {
        fs.mkdirSync(templatesDir, { recursive: true })
        fs.writeFileSync(path.join(templatesDir, `${name}.yaml`), yaml, 'utf8')
        jsonReply(res, 200, { ok: true, name })
      } catch (e) {
        jsonReply(res, 500, { error: String(e) })
      }
    })
    return true
  }

  // POST /api/template/generate — AI로 YAML 템플릿 생성 (SSE 스트리밍)
  if (method === 'POST' && pathname === '/api/template/generate') {
    parseBody(req).then(async (body) => {
      const b            = body as Record<string, string>
      // YAML front-matter 구분자 제거 — 프롬프트가 system prompt YAML을 탈출하는 인젝션 방지
      const userPrompt   = (b['prompt'] || '').trim().replace(/---/g, '- - -')
      const existingYaml = (b['existingYaml'] || '').trim()
      if (!userPrompt) { jsonReply(res, 400, { error: 'prompt is required' }); return }

      const systemPrompt = `You are an expert at creating agent team YAML templates for JiKiME-ADK.
Generate a valid YAML template following this exact schema:

name: <kebab-case-name>
version: "1.0.0"
description: '<short description>'
default_budget: <number>

agents:
  - id: <agent-id>
    role: leader | worker | reviewer
    auto_spawn: true | false
    description: '<what this agent does>'
    task: |
      Goal: {{goal}}

      You are the <role> for {{team_name}}.

      <detailed step-by-step instructions>

Available placeholders in task field:
- {{goal}} — the team's objective
- {{team_name}} — name of the team
- {{agent_id}} — this agent's ID

Rules:
- Always have exactly one leader agent with auto_spawn: true
- Workers check inbox: jikime team inbox receive {{team_name}}
- Workers report to leader: jikime team inbox send {{team_name}} leader "Done: <summary>"
- Workers update task status: jikime team tasks update {{team_name}} <task-id> --status done
- Leader creates tasks: jikime team tasks create {{team_name}} "Title" --desc "..." --dod "..."
- Output ONLY the raw YAML, no markdown fences, no explanation, no commentary`

      const userMessage = existingYaml
        ? `Modify this existing template based on the request.\n\nExisting YAML:\n${existingYaml}\n\nRequest: ${userPrompt}`
        : `Create a new team template for: ${userPrompt}`

      res.writeHead(200, {
        'Content-Type':  'text/event-stream',
        'Cache-Control': 'no-cache',
        'Connection':    'keep-alive',
        ...corsHeaders(),
      })

      const fullPrompt = `${systemPrompt}\n\n${userMessage}`;

      (async () => {
        try {
          const { query } = await import('@anthropic-ai/claude-agent-sdk')
          const queryInstance = query({
            prompt: fullPrompt,
            options: {
              model: 'claude-haiku-4-5-20251001',
              settingSources: ['user'],
            },
          })

          for await (const event of queryInstance) {
            const e = event as Record<string, unknown>
            if (e['type'] === 'assistant') {
              const msg = e['message'] as Record<string, unknown>
              const content = (msg?.['content'] ?? []) as Array<Record<string, unknown>>
              for (const block of content) {
                if (block['type'] === 'text' && typeof block['text'] === 'string') {
                  // writableEnded 체크 — 클라이언트 끊김 후 write 시도로 인한 오류 방지
                  if (!res.writableEnded) res.write(`data: ${JSON.stringify({ chunk: block['text'] })}\n\n`)
                }
              }
            }
          }
        } catch (err) {
          const msg = err instanceof Error ? err.message : String(err)
          if (!res.writableEnded) res.write(`data: ${JSON.stringify({ error: msg })}\n\n`)
        } finally {
          if (!res.writableEnded) res.write(`data: ${JSON.stringify({ done: true })}\n\n`)
          if (!res.destroyed)     res.end()
        }
      })()
    })
    return true
  }

  // DELETE /api/template/:name — 템플릿 삭제 (내장 템플릿은 보호)
  const tmplDeleteMatch = pathname.match(/^\/api\/template\/([^/]+)$/)
  if (method === 'DELETE' && tmplDeleteMatch) {
    const tmplName = decodeURIComponent(tmplDeleteMatch[1] || '')
    if (!NAME_RE.test(tmplName)) { jsonReply(res, 400, { error: 'Invalid template name' }); return true }
    const BUILTIN = ['leader-worker', 'leader-worker-reviewer', 'parallel-workers']
    if (BUILTIN.includes(tmplName)) {
      jsonReply(res, 403, { error: '내장 템플릿은 삭제할 수 없습니다' }); return true
    }
    const filePath = path.join(os.homedir(), '.jikime', 'templates', `${tmplName}.yaml`)
    try {
      if (fs.existsSync(filePath)) fs.unlinkSync(filePath)
      jsonReply(res, 200, { ok: true })
    } catch (e) {
      jsonReply(res, 500, { error: String(e) })
    }
    return true
  }

  // POST /api/team/launch  (team serve: create + spawn all agents from template)
  if (method === 'POST' && pathname === '/api/team/launch') {
    parseBody(req).then((body) => {
      const b           = body as Record<string, unknown>
      const projectPath = (b['projectPath'] as string) || undefined

      // template 검증: NAME_RE 먼저 → 슬래시 포함 경로 탈출 방지
      const ALLOWED_TEMPLATES = ['leader-worker', 'leader-worker-reviewer', 'parallel-workers']
      const rawTmpl = (b['template'] as string || 'leader-worker').slice(0, 80)
      if (!NAME_RE.test(rawTmpl)) { jsonReply(res, 400, { error: 'Invalid template name' }); return }
      if (!ALLOWED_TEMPLATES.includes(rawTmpl) &&
          !(projectPath && fs.existsSync(path.join(dataDir(), 'templates', `${rawTmpl}.yaml`)))) {
        jsonReply(res, 400, { error: `Invalid template: ${rawTmpl}` }); return
      }
      const tmpl = rawTmpl

      // name 검증
      const rawName = (b['name'] as string || '').slice(0, 80)
      if (rawName && !NAME_RE.test(rawName)) { jsonReply(res, 400, { error: 'Invalid team name' }); return }

      // goal 크기 제한 (2000자)
      const rawGoal = (b['goal'] as string || '').slice(0, 2000)

      // worktree 는 명시적 boolean true 만 허용
      const useWorktree = b['worktree'] === true

      const args = ['team', 'launch', '--template', tmpl]
      if (rawName)    args.push('--name',   rawName)
      if (rawGoal)    args.push('--goal',   JSON.stringify(rawGoal))
      if (b['budget']) args.push('--budget', String(b['budget']).slice(0, 20))
      if (useWorktree) args.push('--worktree')
      const effectiveCwd1 = (projectPath && fs.existsSync(projectPath)) ? projectPath : os.homedir()
      execFile('jikime', args, { timeout: 120_000, cwd: effectiveCwd1 }, (err, stdout, stderr) => {
        if (err) { jsonReply(res, 500, { error: err.message, output: ((stdout ?? '') + (stderr ?? '')).trim() }); return }
        // Extract team name from output and save webchat meta
        const launched = stdout.match(/Launching team "([^"]+)"/)?.[1] || rawName || ''
        if (launched && projectPath) {
          try { new TeamFileStore().writeWebchatMeta(launched, { projectPath }) } catch { /* */ }
        }
        jsonReply(res, 201, { ok: true, output: stdout.trim() })
      })
    })
    return true
  }

  if (!pathname.startsWith('/api/team')) return false

  const store = new TeamFileStore()

  // GET /api/team/list
  if (method === 'GET' && pathname === '/api/team/list') {
    const url         = new URL(req.url || '', 'http://localhost')
    const projectPath = url.searchParams.get('projectPath') || undefined
    const teams = store.listTeams(projectPath).map((name) => ({
      name,
      config: store.getConfig(name),
    }))
    jsonReply(res, 200, { teams })
    return true
  }

  // POST /api/team/create
  if (method === 'POST' && pathname === '/api/team/create') {
    parseBody(req).then((body) => {
      const b           = body as Record<string, string>
      // name 검증
      const rawCreateName = (b['name'] || '').trim()
      if (rawCreateName && !NAME_RE.test(rawCreateName)) { jsonReply(res, 400, { error: 'Invalid name' }); return }
      const name        = rawCreateName || `team-${Date.now()}`
      const projectPath = b['projectPath'] || undefined
      // workers: 숫자만 허용
      const createWorkers  = b['workers']  ? String(b['workers']).replace(/\D/g, '')  : ''
      // template: NAME_RE 검증
      const createTemplate = (b['template'] || '').slice(0, 80)
      if (createTemplate && !NAME_RE.test(createTemplate)) { jsonReply(res, 400, { error: 'Invalid template' }); return }
      // budget: 숫자만 허용
      const createBudget   = b['budget']   ? String(b['budget']).replace(/\D/g, '')   : ''
      const args = ['team', 'create', name]
      if (createWorkers)   args.push('--workers',  createWorkers)
      if (createTemplate)  args.push('--template', createTemplate)
      if (createBudget)    args.push('--budget',   createBudget)
      execFile('jikime', args, (err, stdout) => {
        if (err) { jsonReply(res, 500, { error: err.message }); return }
        if (projectPath) {
          try { new TeamFileStore().writeWebchatMeta(name, { projectPath }) } catch { /* */ }
        }
        jsonReply(res, 201, { ok: true, output: stdout.trim() })
      })
    })
    return true
  }

  // Match /api/team/:name[/...]
  const teamMatch = pathname.match(/^\/api\/team\/([^/]+)(\/.*)?$/)
  if (!teamMatch) return false

  // decodeURIComponent 실패 시 400 — 잘못된 퍼센트 인코딩으로 인한 오류 방지
  let teamName: string
  try {
    teamName = decodeURIComponent(teamMatch[1])
  } catch {
    jsonReply(res, 400, { error: 'Invalid URL encoding in team name' }); return true
  }
  if (!NAME_RE.test(teamName)) { jsonReply(res, 400, { error: 'Invalid team name' }); return true }
  const subPath  = teamMatch[2] || ''

  // GET /api/team/:name
  if (method === 'GET' && subPath === '') {
    const config  = store.getConfig(teamName)
    const tasks   = store.listTasks(teamName)
    const agents  = store.listAgents(teamName)
    const costs   = store.getCosts(teamName)
    const counts: Record<string, number> = {}
    for (const t of tasks as Array<Record<string, unknown>>) {
      const s = (t['status'] as string) || 'unknown'
      counts[s] = (counts[s] || 0) + 1
    }
    jsonReply(res, 200, { name: teamName, config, taskCounts: counts, agentCount: agents.length, costs })
    return true
  }

  // DELETE /api/team/:name/agents/:agentId — kill individual agent tmux session
  const agentKillMatch = subPath.match(/^\/agents\/([^/]+)$/)
  if (method === 'DELETE' && agentKillMatch) {
    const agentId = decodeURIComponent(agentKillMatch[1] || '')
    // 디코딩 후 재검증 — tmux 세션명 구성 전 경로 탈출 / 인젝션 방지
    if (!NAME_RE.test(agentId)) { jsonReply(res, 400, { error: 'Invalid agent ID' }); return true }
    const sessionName = `jikime-${teamName.replace(/[ /:]/g, '-')}-${agentId.replace(/[ /:]/g, '-')}`
    execFile('tmux', ['kill-session', '-t', sessionName], (err) => {
      if (err) { jsonReply(res, 500, { error: err.message }); return }
      jsonReply(res, 200, { ok: true })
    })
    return true
  }

  // DELETE /api/team/:name
  if (method === 'DELETE' && subPath === '') {
    execFile('jikime', ['team', 'stop', teamName, '--force'], (err) => {
      if (err) { jsonReply(res, 500, { error: err.message }); return }
      stopTeamWatcher(teamName)
      jsonReply(res, 200, { ok: true })
    })
    return true
  }

  // GET /api/team/:name/tasks
  if (method === 'GET' && subPath === '/tasks') {
    const url   = new URL(req.url || '', 'http://localhost')
    const status = url.searchParams.get('status') || ''
    const agent  = url.searchParams.get('agent')  || ''
    const owner  = url.searchParams.get('owner')  || ''
    // limit: 1~5000 범위, 기본 1000 — 대량 태스크 OOM 방지
    const limitParam = parseInt(url.searchParams.get('limit') || '0')
    const limit = Math.min(Math.max(1, limitParam || 1000), 5000)
    jsonReply(res, 200, { tasks: store.listTasks(teamName, status || undefined, agent || undefined, owner || undefined, limit) })
    return true
  }

  // POST /api/team/:name/tasks
  if (method === 'POST' && subPath === '/tasks') {
    parseBody(req).then((body) => {
      const b = body as Record<string, string>
      const args = ['team', 'tasks', 'create', teamName, JSON.stringify(b['title'] || 'task')]
      if (b['desc'])   args.push('--desc',  JSON.stringify(b['desc']))
      if (b['dod'])    args.push('--dod',   JSON.stringify(b['dod']))
      if (b['owner'])  args.push('--owner', b['owner'])
      execFile('jikime', args, (err, stdout) => {
        if (err) { jsonReply(res, 500, { error: err.message }); return }
        jsonReply(res, 201, { ok: true, output: stdout.trim() })
      })
    })
    return true
  }

  // GET /api/team/:name/agents
  if (method === 'GET' && subPath === '/agents') {
    jsonReply(res, 200, { agents: store.listAgents(teamName) })
    return true
  }

  // GET /api/team/:name/budget
  if (method === 'GET' && subPath === '/budget') {
    const costs  = store.getCosts(teamName)
    const config = store.getConfig(teamName) as Record<string, unknown>
    const budget = (config['budget'] as number) || 0
    jsonReply(res, 200, { budget, costs })
    return true
  }

  // POST /api/team/:name/run  (run existing team with a new goal)
  if (method === 'POST' && subPath === '/run') {
    parseBody(req).then((body) => {
      const b           = body as Record<string, string>
      const config      = store.getConfig(teamName) as Record<string, unknown>
      const template    = (config['template'] as string)
                       || (config['template_name'] as string)
                       || 'leader-worker'
      // projectPath: from request body, fallback to stored webchat.json
      const projectPath = b['projectPath']
                       || store.getWebchatMeta(teamName).projectPath
                       || undefined
      const args = ['team', 'launch', '--template', template, '--name', teamName]
      if (b['goal'])     args.push('--goal', JSON.stringify(b['goal']))
      if (b['worktree']) args.push('--worktree')
      const effectiveCwd2 = (projectPath && fs.existsSync(projectPath)) ? projectPath : os.homedir()
      execFile('jikime', args, { timeout: 120_000, cwd: effectiveCwd2 }, (err, stdout, stderr) => {
        if (err) { jsonReply(res, 500, { error: err.message, output: ((stdout ?? '') + (stderr ?? '')).trim() }); return }
        if (projectPath) {
          try { store.writeWebchatMeta(teamName, { projectPath }) } catch { /* */ }
        }
        jsonReply(res, 201, { ok: true, output: stdout.trim() })
      })
    })
    return true
  }

  // PATCH /api/team/:name/tasks/:id
  const taskPatchMatch = subPath.match(/^\/tasks\/([^/]+)$/)
  if (method === 'PATCH' && taskPatchMatch) {
    const rawTaskID = decodeURIComponent(taskPatchMatch[1] || '')
    if (!NAME_RE.test(rawTaskID)) { jsonReply(res, 400, { error: 'Invalid task ID' }); return true }
    const taskID = rawTaskID
    parseBody(req).then((body) => {
      const b = body as Record<string, string>
      // status 화이트리스트 — 임의 문자열이 CLI 인수로 전달되는 것 방지
      const ALLOWED_STATUSES = ['pending', 'in_progress', 'done', 'failed', 'blocked']
      const patchStatus = (b['status'] || '').trim()
      if (patchStatus && !ALLOWED_STATUSES.includes(patchStatus)) {
        jsonReply(res, 400, { error: `Invalid status: ${patchStatus}` }); return
      }
      // agent_id NAME_RE 검증
      const patchAgent = (b['agent_id'] || '').slice(0, 80)
      if (patchAgent && !NAME_RE.test(patchAgent)) {
        jsonReply(res, 400, { error: 'Invalid agent_id' }); return
      }
      const args = ['team', 'tasks', 'update', teamName, taskID]
      if (patchStatus) args.push('--status', patchStatus)
      if (patchAgent)  args.push('--agent',  patchAgent)
      if (b['result']) args.push('--result',  JSON.stringify((b['result'] || '').slice(0, 2000)))
      execFile('jikime', args, (err) => {
        if (err) { jsonReply(res, 500, { error: err.message }); return }
        jsonReply(res, 200, { ok: true })
      })
    })
    return true
  }

  // GET /api/team/:name/inbox/peek?agent=<id>
  if (method === 'GET' && subPath === '/inbox/peek') {
    const url = new URL(req.url || '', 'http://localhost')
    const agentID = url.searchParams.get('agent') || ''
    if (!agentID) { jsonReply(res, 400, { error: 'agent param required' }); return true }
    jsonReply(res, 200, { messages: store.getInboxMessages(teamName, agentID) })
    return true
  }

  // POST /api/team/:name/inbox/send
  if (method === 'POST' && subPath === '/inbox/send') {
    parseBody(req).then((body) => {
      const b = body as Record<string, string>
      const to   = (b['to']  || '').slice(0, 80).trim()
      const msg  = (b['body'] || b['message'] || '').slice(0, 4000).trim()
      const from = (b['from'] || 'webchat').slice(0, 80).trim()
      // to/from 은 알파뉴메릭+_- 만 허용 — CLI 파서 혼란 방지
      if (!NAME_RE.test(to))   { jsonReply(res, 400, { error: '"to" must match [a-zA-Z0-9_-]{1,80}' }); return }
      if (!msg)                 { jsonReply(res, 400, { error: 'body required' }); return }
      if (!NAME_RE.test(from)) { jsonReply(res, 400, { error: '"from" must match [a-zA-Z0-9_-]{1,80}' }); return }
      execFile('jikime', ['team', 'inbox', 'send', teamName, to, msg, '--from', from], (err) => {
        if (err) { jsonReply(res, 500, { error: err.message }); return }
        jsonReply(res, 200, { ok: true })
      })
    })
    return true
  }

  // GET /api/team/:name/events  (SSE)
  if (method === 'GET' && subPath === '/events') {
    // 전역 SSE 클라이언트 상한 — N팀 × per-team 500 합산 DoS 방지
    if (sseClients.size >= MAX_GLOBAL_SSE_CLIENTS) {
      jsonReply(res, 429, { error: 'Too many SSE connections' }); return true
    }
    res.writeHead(200, {
      'Content-Type':  'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection':    'keep-alive',
      ...corsHeaders(),
    })
    res.write('retry: 2000\n\n')

    const client: SSEClient = { res, teamName }
    sseClients.add(client)
    startTeamWatcher(teamName)

    // Send initial snapshot
    broadcastTeamEvent(teamName)

    req.on('close', () => {
      sseClients.delete(client)
      // Stop watcher if no more clients watching this team
      const hasClients = [...sseClients].some((c) => c.teamName === teamName)
      if (!hasClients) stopTeamWatcher(teamName)
    })
    return true
  }

  return false
}
