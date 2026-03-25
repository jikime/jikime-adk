import { NextRequest, NextResponse } from 'next/server'
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
  const name        = (body['name'] || `team-${Date.now()}`).slice(0, 80)
  const projectPath = body['projectPath'] || undefined

  if (!NAME_RE.test(name)) {
    return NextResponse.json({ error: 'Invalid team name (alphanumeric, _ and - only)' }, { status: 400 })
  }

  const args: string[] = ['team', 'create', name]
  if (body['template']) args.push('--template', body['template'].slice(0, 80))
  if (body['budget'] && BUDGET_RE.test(body['budget'])) args.push('--budget', body['budget'])

  return new Promise<NextResponse>((resolve) => {
    execFile('jikime', args, (err, stdout) => {
      if (err) {
        resolve(NextResponse.json({ error: err.message }, { status: 500 }))
        return
      }
      if (projectPath) {
        try { new TeamFileStore().writeWebchatMeta(name, { projectPath }) } catch { /* */ }
      }
      resolve(NextResponse.json({ ok: true, output: stdout.trim() }, { status: 201 }))
    })
  })
}
