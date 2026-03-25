import { NextRequest, NextResponse } from 'next/server'
import * as fs   from 'fs'
import * as os   from 'os'
import * as path from 'path'
import { execFile } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

const NAME_RE   = /^[a-zA-Z0-9_-]{1,80}$/
const BUDGET_RE = /^\d{1,10}$/

export async function POST(request: NextRequest) {
  const contentLength = request.headers.get('content-length')
  if (contentLength && parseInt(contentLength) > 10_240) {
    return NextResponse.json({ error: 'Request too large' }, { status: 413 })
  }

  const body        = await request.json() as Record<string, string>
  const tmpl        = (body['template'] || 'leader-worker').slice(0, 80)
  const projectPath = body['projectPath'] || undefined

  const args: string[] = ['team', 'launch', '--template', tmpl]
  if (body['name'] && NAME_RE.test(body['name'])) args.push('--name', body['name'])
  if (body['goal'])   args.push('--goal',   body['goal'].slice(0, 2000))
  if (body['budget'] && BUDGET_RE.test(body['budget'])) args.push('--budget', body['budget'])
  if (body['worktree']) args.push('--worktree')

  // Validate and resolve cwd — only allow existing directories
  let cwd = os.homedir()
  if (projectPath) {
    const resolved = path.resolve(projectPath)
    try {
      if (fs.existsSync(resolved) && fs.statSync(resolved).isDirectory()) cwd = resolved
    } catch { /* use homedir */ }
  }

  return new Promise<NextResponse>((resolve) => {
    execFile('jikime', args, { timeout: 120_000, cwd }, (err, stdout, stderr) => {
      if (err) {
        resolve(NextResponse.json({ error: err.message, output: (stdout + stderr).trim() }, { status: 500 }))
        return
      }
      const launched = stdout.match(/Launching team "([^"]+)"/)?.[1] || body['name'] || ''
      if (launched && projectPath) {
        try { new TeamFileStore().writeWebchatMeta(launched, { projectPath }) } catch { /* */ }
      }
      resolve(NextResponse.json({ ok: true, output: stdout.trim() }, { status: 201 }))
    })
  })
}
