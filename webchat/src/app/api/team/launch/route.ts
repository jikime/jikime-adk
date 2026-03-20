import { NextRequest, NextResponse } from 'next/server'
import * as fs from 'fs'
import * as os from 'os'
import { exec } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

export async function POST(request: NextRequest) {
  const body        = await request.json() as Record<string, string>
  const tmpl        = body['template'] || 'leader-worker'
  const projectPath = body['projectPath'] || undefined

  const args = ['team', 'launch', '--template', tmpl]
  if (body['name'])     args.push('--name',   body['name'])
  if (body['goal'])     args.push('--goal',   JSON.stringify(body['goal']))
  if (body['budget'])   args.push('--budget', body['budget'])
  if (body['worktree']) args.push('--worktree')

  const cwd = (projectPath && fs.existsSync(projectPath)) ? projectPath : os.homedir()

  return new Promise<NextResponse>((resolve) => {
    exec(`jikime ${args.join(' ')}`, { timeout: 120_000, cwd }, (err, stdout, stderr) => {
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
