import { NextRequest, NextResponse } from 'next/server'
import { TeamFileStore } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(request: NextRequest, { params }: Params) {
  const { name } = await params
  const agentID  = request.nextUrl.searchParams.get('agent') || ''
  if (!agentID) return NextResponse.json({ error: 'agent param required' }, { status: 400 })
  const messages = new TeamFileStore().getInboxMessages(name, agentID)
  return NextResponse.json({ messages })
}
