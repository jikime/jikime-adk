import { NextRequest } from 'next/server'
import { readdir, open, stat } from 'fs/promises'
import { join, basename } from 'path'

export const dynamic = 'force-dynamic'

export interface SessionInfo {
  sessionId: string
  cwd: string
  firstMessage: string
  timestamp: string   // ISO — 첫 메시지 시각
  lastActive: number  // mtime ms — 정렬용
}

const CLAUDE_PROJECTS = join(process.env.HOME ?? '~', '.claude', 'projects')
const MAX_SESSIONS = 50
const READ_BYTES = 4096  // JSONL 앞부분만 읽어서 파싱

/** JSONL 앞부분 raw text → 메타데이터 추출 */
function parseMeta(raw: string, sessionId: string): Omit<SessionInfo, 'lastActive'> | null {
  let cwd = ''
  let firstMessage = ''
  let timestamp = ''

  for (const line of raw.split('\n')) {
    if (!line.trim()) continue
    try {
      const obj = JSON.parse(line) as Record<string, unknown>
      if (!cwd && typeof obj.cwd === 'string') cwd = obj.cwd
      if (!firstMessage && obj.type === 'user') {
        const msg = obj.message as { role?: string; content?: unknown } | undefined
        if (msg?.role === 'user') {
          const c = msg.content
          if (typeof c === 'string') firstMessage = c.slice(0, 80)
          else if (Array.isArray(c)) {
            const t = (c as Array<{ type?: string; text?: string }>).find(b => b.type === 'text')
            if (t?.text) firstMessage = t.text.slice(0, 80)
          }
          if (typeof obj.timestamp === 'string') timestamp = obj.timestamp
        }
      }
    } catch { /* invalid JSON line */ }
    if (cwd && firstMessage) break
  }

  if (!firstMessage) return null
  return { sessionId, cwd: cwd || '/', firstMessage, timestamp }
}

/** JSONL 파일 앞부분만 읽기 */
async function readHead(filePath: string): Promise<string> {
  const fh = await open(filePath, 'r')
  try {
    const buf = Buffer.alloc(READ_BYTES)
    const { bytesRead } = await fh.read(buf, 0, READ_BYTES, 0)
    return buf.subarray(0, bytesRead).toString('utf8')
  } finally {
    await fh.close()
  }
}

/** cwd 경로 → CLAUDE_PROJECTS 하위 디렉터리명 prefix로 변환
 *  e.g. /Users/foo/bar  →  -Users-foo-bar
 */
function cwdToDirName(cwd: string): string {
  return cwd.replace(/\//g, '-')
}

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const filterCwd = searchParams.get('cwd')?.trim() || ''

    const sessions: SessionInfo[] = []
    const projectDirs = await readdir(CLAUDE_PROJECTS).catch(() => [] as string[])

    // filterCwd가 있으면 해당 경로에 해당하는 디렉터리만 탐색
    const targetDir = filterCwd ? cwdToDirName(filterCwd) : ''
    const dirsToScan = filterCwd
      ? projectDirs.filter(d => d === targetDir || d.startsWith(targetDir))
      : projectDirs

    for (const dir of dirsToScan) {
      const dirPath = join(CLAUDE_PROJECTS, dir)
      const dirStat = await stat(dirPath).catch(() => null)
      if (!dirStat?.isDirectory()) continue

      const files = await readdir(dirPath).catch(() => [] as string[])
      for (const file of files) {
        if (!file.endsWith('.jsonl')) continue
        const sessionId = basename(file, '.jsonl')
        if (!/^[0-9a-f-]{36}$/i.test(sessionId)) continue

        const filePath = join(dirPath, file)
        const [fileStat, raw] = await Promise.all([
          stat(filePath).catch(() => null),
          readHead(filePath).catch(() => ''),
        ])
        if (!fileStat || !raw) continue

        const meta = parseMeta(raw, sessionId)
        if (!meta) continue
        sessions.push({ ...meta, lastActive: fileStat.mtimeMs })
      }
    }

    sessions.sort((a, b) => b.lastActive - a.lastActive)
    return Response.json(sessions.slice(0, MAX_SESSIONS))
  } catch (err) {
    console.error('[sessions]', err)
    return Response.json([])
  }
}
