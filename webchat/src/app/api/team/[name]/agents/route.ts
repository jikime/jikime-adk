import { NextRequest, NextResponse } from 'next/server'
import { execFile } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(_req: NextRequest, { params }: Params) {
  const { name } = await params
  const agents   = new TeamFileStore().listAgents(name)
  return NextResponse.json({ agents })
}

export async function DELETE(request: NextRequest, { params }: Params) {
  const { name }  = await params
  const agentId   = request.nextUrl.searchParams.get('agentId') || ''
  if (!agentId) return NextResponse.json({ error: 'agentId required' }, { status: 400 })

  // Sanitize to prevent any shell metachar injection via execFile args
  const safeName    = name.replace(/[^a-zA-Z0-9_-]/g, '-')
  const safeAgentId = agentId.replace(/[^a-zA-Z0-9_-]/g, '-')
  const sessionName = `jikime-${safeName}-${safeAgentId}`

  return new Promise<NextResponse>((resolve) => {
    execFile('tmux', ['kill-session', '-t', sessionName], (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
