import { NextRequest, NextResponse } from 'next/server'
import { TeamFileStore } from '@/lib/team-store'

export async function GET(request: NextRequest) {
  const projectPath = request.nextUrl.searchParams.get('projectPath') || undefined
  const store = new TeamFileStore()
  const teams = store.listTeams(projectPath).map((name) => ({
    name,
    config: store.getConfig(name),
  }))
  return NextResponse.json({ teams })
}
