import { NextRequest, NextResponse } from 'next/server'
import { TeamFileStore } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(_req: NextRequest, { params }: Params) {
  const { name } = await params
  const store    = new TeamFileStore()
  const costs    = store.getCosts(name)
  const config   = store.getConfig(name) as Record<string, unknown>
  const budget   = (config['budget'] as number) || 0
  return NextResponse.json({ budget, costs })
}
