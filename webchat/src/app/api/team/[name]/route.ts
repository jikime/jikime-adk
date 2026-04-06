import { NextRequest, NextResponse } from 'next/server'
import { execFile } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'
import { NAME_RE } from '@/lib/validation'

type Params = { params: Promise<{ name: string }> }

export async function GET(_req: NextRequest, { params }: Params) {
  const { name } = await params
  if (!NAME_RE.test(name)) return NextResponse.json({ error: 'Invalid team name' }, { status: 400 })
  const store    = new TeamFileStore()
  const config   = store.getConfig(name)
  const tasks    = store.listTasks(name)
  const agents   = store.listAgents(name)
  const costs    = store.getCosts(name)
  const counts: Record<string, number> = {}
  for (const t of tasks as Array<Record<string, unknown>>) {
    const s = (t['status'] as string) || 'unknown'
    counts[s] = (counts[s] || 0) + 1
  }
  return NextResponse.json({ name, config, taskCounts: counts, agentCount: agents.length, costs })
}

export async function DELETE(_req: NextRequest, { params }: Params) {
  const { name } = await params
  if (!NAME_RE.test(name)) return NextResponse.json({ error: 'Invalid team name' }, { status: 400 })
  return new Promise<NextResponse>((resolve) => {
    execFile('jikime', ['team', 'stop', name, '--force'], (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
