import { NextRequest, NextResponse } from 'next/server'
import { exec } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

export async function POST(request: NextRequest) {
  const body        = await request.json() as Record<string, string>
  const name        = body['name'] || `team-${Date.now()}`
  const projectPath = body['projectPath'] || undefined

  const args = ['team', 'create', name]
  if (body['template']) args.push('--template', body['template'])
  if (body['budget'])   args.push('--budget',   body['budget'])

  return new Promise<NextResponse>((resolve) => {
    exec(`jikime ${args.join(' ')}`, (err, stdout) => {
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
