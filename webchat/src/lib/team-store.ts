/**
 * team-store.ts — server-side business logic for team orchestration
 *
 * Used by App Router Route Handlers (runs on Node.js runtime only).
 */

import * as fs   from 'fs'
import * as path from 'path'
import * as os   from 'os'

// ── Helpers ────────────────────────────────────────────────────────

export function dataDir(): string {
  return process.env.JIKIME_DATA_DIR || path.join(os.homedir(), '.jikime')
}

export function teamsDir(): string {
  return path.join(dataDir(), 'teams')
}

export function teamDir(name: string): string {
  return path.join(teamsDir(), name)
}

export function templatesDir(): string {
  return path.join(dataDir(), 'templates')
}

export function readJSON<T>(filePath: string, fallback: T): T {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8')) as T
  } catch {
    return fallback
  }
}

// ── TeamFileStore ──────────────────────────────────────────────────

export class TeamFileStore {
  constructor(private readonly root: string = teamsDir()) {}

  listTeams(projectPath?: string): string[] {
    if (!fs.existsSync(this.root)) return []
    // Use withFileTypes to avoid a separate statSync per entry
    const all = fs.readdirSync(this.root, { withFileTypes: true })
      .filter((e) => e.isDirectory())
      .map((e) => e.name)
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
        if (status && t['status']   !== status) return false
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
    const dir    = path.join(teamDir(name), 'costs')
    const result: { total: number; agents: Record<string, { tokens: number }> } = { total: 0, agents: {} }
    if (!fs.existsSync(dir)) return result
    fs.readdirSync(dir)
      .filter((f) => f.endsWith('.json'))
      .forEach((f) => {
        const ev      = readJSON<Record<string, unknown>>(path.join(dir, f), {})
        const agentID = (ev['agent_id'] as string) || 'unknown'
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

// ── Team Snapshot cache ─────────────────────────────────────────────
// Prevents redundant full-directory scans when multiple watchers fire in rapid succession.

const SNAPSHOT_TTL_MS = 3_000  // 3 seconds

const snapshotCache = new Map<string, { snapshot: unknown; expiresAt: number }>()

export function invalidateSnapshot(teamName: string): void {
  snapshotCache.delete(teamName)
}

// ── Team Snapshot (SSE payload builder) ────────────────────────────

export function buildTeamSnapshot(teamName: string): unknown {
  const now    = Date.now()
  const cached = snapshotCache.get(teamName)
  if (cached && cached.expiresAt > now) return cached.snapshot

  const store      = new TeamFileStore()
  const tasksList  = store.listTasks(teamName)  as Array<Record<string, unknown>>
  const agentsList = store.listAgents(teamName) as Array<Record<string, unknown>>
  const config     = store.getConfig(teamName)  as Record<string, unknown>

  const taskSummary: Record<string, number> = {
    pending: 0, in_progress: 0, done: 0, failed: 0, blocked: 0,
  }
  for (const t of tasksList) {
    const s = (t['status'] as string) || 'unknown'
    if (s in taskSummary) taskSummary[s]++
  }

  let leaderName = (config['leader_id'] as string) || ''
  if (!leaderName) {
    const leader = agentsList.find((a) => a['role'] === 'leader')
    leaderName = (leader?.['id'] as string) || ''
  }

  const members = agentsList.map((a) => {
    const agentID  = a['id'] as string
    const inboxDir = path.join(teamDir(teamName), 'inbox', agentID)
    let inboxCount = 0
    try { inboxCount = fs.readdirSync(inboxDir).filter((f) => f.endsWith('.json')).length } catch { /* */ }
    return { name: agentID, agentType: (a['role'] as string) || 'worker', inboxCount }
  })

  const eventLogFile = path.join(teamDir(teamName), 'inbox', 'event-log.jsonl')
  const messages: Array<{ from: string; to: string; type: string; timestamp: string; content: string }> = []
  try {
    const lines = fs.readFileSync(eventLogFile, 'utf8').split('\n').filter(Boolean)
    for (const line of lines.slice(-50)) {
      try {
        const m = JSON.parse(line) as Record<string, unknown>
        messages.push({
          from:      (m['from']    as string) || '',
          to:        (m['to']      as string) || '',
          type:      (m['kind']    as string) || 'direct',
          timestamp: (m['sent_at'] as string) || '',
          content:   (m['body']    as string) || '',
        })
      } catch { /* skip malformed line */ }
    }
  } catch { /* log file may not exist yet */ }

  const snapshot = {
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
  }

  snapshotCache.set(teamName, { snapshot, expiresAt: now + SNAPSHOT_TTL_MS })
  return snapshot
}
