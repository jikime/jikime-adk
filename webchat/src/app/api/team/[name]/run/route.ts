import { NextRequest, NextResponse } from 'next/server'
import * as fs   from 'fs'
import * as os   from 'os'
import * as path from 'path'
import { execFile } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

const NAME_RE = /^[a-zA-Z0-9_-]{1,64}$/

type Params = { params: Promise<{ name: string }> }

export async function POST(request: NextRequest, { params }: Params) {
  const { name }    = await params
  if (!NAME_RE.test(name)) return NextResponse.json({ error: 'Invalid team name' }, { status: 400 })
  const contentLength = request.headers.get('content-length')
  if (contentLength && parseInt(contentLength) > 10_240) {
    return NextResponse.json({ error: 'Request too large' }, { status: 413 })
  }
  const body        = await request.json() as Record<string, string>
  const store       = new TeamFileStore()
  const config      = store.getConfig(name) as Record<string, unknown>
  const template    = (config['template'] as string) || (config['template_name'] as string) || 'leader-worker'
  const projectPath = body['projectPath'] || store.getWebchatMeta(name).projectPath || undefined

  const args: string[] = ['team', 'launch', '--template', template, '--name', name]
  if (body['goal'])     args.push('--goal', body['goal'].slice(0, 2000))
  if (body['worktree']) args.push('--worktree')

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
      if (projectPath) {
        try { store.writeWebchatMeta(name, { projectPath }) } catch { /* */ }
      }
      resolve(NextResponse.json({ ok: true, output: stdout.trim() }, { status: 201 }))
    })
  })
}
