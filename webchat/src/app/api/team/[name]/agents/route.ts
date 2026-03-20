import { NextRequest, NextResponse } from 'next/server'
import { exec } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(_req: NextRequest, { params }: Params) {
  const { name } = await params
  const agents   = new TeamFileStore().listAgents(name)
  return NextResponse.json({ agents })
}

// DELETE /api/team/:name/agents/:agentId is handled via a nested dynamic route
// but agent kill via team name is done here with ?agentId=
export async function DELETE(request: NextRequest, { params }: Params) {
  const { name }  = await params
  const agentId   = request.nextUrl.searchParams.get('agentId') || ''
  if (!agentId) return NextResponse.json({ error: 'agentId required' }, { status: 400 })

  const sessionName = `jikime-${name.replace(/[ /:]/g, '-')}-${agentId.replace(/[ /:]/g, '-')}`
  return new Promise<NextResponse>((resolve) => {
    exec(`tmux kill-session -t ${sessionName}`, (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
