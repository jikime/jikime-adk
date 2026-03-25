import { NextRequest, NextResponse } from 'next/server'
import * as fs   from 'fs'
import * as os   from 'os'
import * as path from 'path'
import { execFile } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function POST(request: NextRequest, { params }: Params) {
  const { name }    = await params
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
