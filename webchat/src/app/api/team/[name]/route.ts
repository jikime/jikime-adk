import { NextRequest, NextResponse } from 'next/server'
import { exec } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(_req: NextRequest, { params }: Params) {
  const { name } = await params
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
  return new Promise<NextResponse>((resolve) => {
    exec(`jikime team stop ${name} --force`, (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
