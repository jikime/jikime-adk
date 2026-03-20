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
import { exec } from 'child_process'
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

function parseBody(req: IncomingMessage): Promise<unknown> {
  return new Promise((resolve) => {
    const chunks: Buffer[] = []
    req.on('data', (chunk: Buffer) => chunks.push(chunk))
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
    const all = fs.readdirSync(this.root).filter((e) =>
      fs.statSync(path.join(this.root, e)).isDirectory()
    )
    if (!projectPath) return all
    const realQuery = (() => { try { return fs.realpathSync(projectPath) } catch { return projectPath } })()
    return all.filter((name) => {
      const meta = this.getWebchatMeta(name)
      if (!meta.projectPath) return false
      const realMeta = (() => { try { return fs.realpathSync(meta.projectPath) } catch { return meta.projectPath } })()
      return realMeta === realQuery
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
    fs.writeFileSync(path.join(dir, 'webchat.json'), JSON.stringify(meta))
  }

  listTasks(name: string, status?: string, agent?: string, owner?: string): unknown[] {
    const dir = path.join(teamDir(name), 'tasks')
    if (!fs.existsSync(dir)) return []
    return fs.readdirSync(dir)
      .filter((f) => f.endsWith('.json'))
      .map((f) => readJSON<Record<string, unknown>>(path.join(dir, f), {}))
      .filter((t) => {
        if (status && t['status'] !== status) return false
        if (agent  && t['agent_id'] !== agent)  return false
        if (owner  && t['owner']    !== owner)   return false
        return true
      })
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

function startTeamWatcher(teamName: string): void {
  if (sseWatchers.has(teamName)) return

  const watchDirs = ['tasks', 'registry', 'inbox', 'costs'].map((d) =>
    path.join(teamDir(teamName), d)
  )
  const watchers: fs.FSWatcher[] = []
  for (const dir of watchDirs) {
    if (!fs.existsSync(dir)) continue
    const w = fs.watch(dir, { recursive: false }, () => {
      broadcastTeamEvent(teamName)
    })
    watchers.push(w)
  }
  sseWatchers.set(teamName, watchers)
}

function stopTeamWatcher(teamName: string): void {
  const watchers = sseWatchers.get(teamName)
  if (watchers) {
    watchers.forEach((w) => w.close())
    sseWatchers.delete(teamName)
  }
}

function broadcastTeamEvent(teamName: string): void {
  const store      = new TeamFileStore()
  const tasksList  = store.listTasks(teamName)  as Array<Record<string, unknown>>
  const agentsList = store.listAgents(teamName) as Array<Record<string, unknown>>
  const config     = store.getConfig(teamName)  as Record<string, unknown>

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

  // Event log messages (last 50 from inbox/event-log.jsonl)
  const eventLogFile = path.join(teamDir(teamName), 'inbox', 'event-log.jsonl')
  const messages: Array<{ from: string; to: string; type: string; timestamp: string; content: string }> = []
  try {
    const lines = fs.readFileSync(eventLogFile, 'utf8').split('\n').filter(Boolean)
    for (const line of lines.slice(-50)) {
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

  for (const client of sseClients) {
    if (client.teamName !== teamName) continue
    try {
      client.res.write(`data: ${payload}\n\n`)
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
      const userPrompt   = (b['prompt'] || '').trim()
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
                  res.write(`data: ${JSON.stringify({ chunk: block['text'] })}\n\n`)
                }
              }
            }
          }
        } catch (err) {
          const msg = err instanceof Error ? err.message : String(err)
          res.write(`data: ${JSON.stringify({ error: msg })}\n\n`)
        } finally {
          res.write(`data: ${JSON.stringify({ done: true })}\n\n`)
          res.end()
        }
      })()
    })
    return true
  }

  // DELETE /api/template/:name — 템플릿 삭제 (내장 템플릿은 보호)
  const tmplDeleteMatch = pathname.match(/^\/api\/template\/([^/]+)$/)
  if (method === 'DELETE' && tmplDeleteMatch) {
    const tmplName = decodeURIComponent(tmplDeleteMatch[1] || '')
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
      const b           = body as Record<string, string>
      const tmpl        = b['template'] || 'leader-worker'
      const projectPath = b['projectPath'] || undefined
      const args = ['team', 'launch', '--template', tmpl]
      if (b['name'])     args.push('--name',   b['name'])
      if (b['goal'])     args.push('--goal',   JSON.stringify(b['goal']))
      if (b['budget'])   args.push('--budget', b['budget'])
      if (b['worktree']) args.push('--worktree')
      const effectiveCwd1 = (projectPath && fs.existsSync(projectPath)) ? projectPath : os.homedir()
      const execOpts: { timeout: number; cwd: string } = { timeout: 120_000, cwd: effectiveCwd1 }
      exec(`jikime ${args.join(' ')}`, execOpts, (err, stdout, stderr) => {
        if (err) { jsonReply(res, 500, { error: err.message, output: (stdout + stderr).trim() }); return }
        // Extract team name from output and save webchat meta
        const launched = stdout.match(/Launching team "([^"]+)"/)?.[1] || b['name'] || ''
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
      const name        = b['name'] || `team-${Date.now()}`
      const projectPath = b['projectPath'] || undefined
      const args = ['team', 'create', name]
      if (b['workers'])  args.push('--workers',  b['workers'])
      if (b['template']) args.push('--template', b['template'])
      if (b['budget'])   args.push('--budget',   b['budget'])
      exec(`jikime ${args.join(' ')}`, (err, stdout) => {
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

  const teamName = teamMatch[1]
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
    const sessionName = `jikime-${teamName.replace(/[ /:]/g, '-')}-${agentId.replace(/[ /:]/g, '-')}`
    exec(`tmux kill-session -t ${sessionName}`, (err) => {
      if (err) { jsonReply(res, 500, { error: err.message }); return }
      jsonReply(res, 200, { ok: true })
    })
    return true
  }

  // DELETE /api/team/:name
  if (method === 'DELETE' && subPath === '') {
    exec(`jikime team stop ${teamName} --force`, (err) => {
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
    jsonReply(res, 200, { tasks: store.listTasks(teamName, status || undefined, agent || undefined, owner || undefined) })
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
      exec(`jikime ${args.join(' ')}`, (err, stdout) => {
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
      const execOpts: { timeout: number; cwd: string } = { timeout: 120_000, cwd: effectiveCwd2 }
      exec(`jikime ${args.join(' ')}`, execOpts, (err, stdout, stderr) => {
        if (err) { jsonReply(res, 500, { error: err.message, output: (stdout + stderr).trim() }); return }
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
    const taskID = taskPatchMatch[1]
    parseBody(req).then((body) => {
      const b = body as Record<string, string>
      const args = ['team', 'tasks', 'update', teamName, taskID]
      if (b['status'])   args.push('--status', b['status'])
      if (b['agent_id']) args.push('--agent',  b['agent_id'])
      if (b['result'])   args.push('--result',  JSON.stringify(b['result']))
      exec(`jikime ${args.join(' ')}`, (err) => {
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
      const to  = b['to'] || ''
      const msg = b['body'] || b['message'] || ''
      const from = b['from'] || 'webchat'
      if (!to || !msg) { jsonReply(res, 400, { error: 'to and body required' }); return }
      exec(`jikime team inbox send ${teamName} ${to} ${JSON.stringify(msg)} --from ${from}`, (err) => {
        if (err) { jsonReply(res, 500, { error: err.message }); return }
        jsonReply(res, 200, { ok: true })
      })
    })
    return true
  }

  // GET /api/team/:name/events  (SSE)
  if (method === 'GET' && subPath === '/events') {
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
