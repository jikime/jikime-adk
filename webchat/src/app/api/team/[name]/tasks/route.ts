import { NextRequest, NextResponse } from 'next/server'
import { exec } from 'child_process'
import { TeamFileStore } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(request: NextRequest, { params }: Params) {
  const { name } = await params
  const sp       = request.nextUrl.searchParams
  const store    = new TeamFileStore()
  const tasks    = store.listTasks(
    name,
    sp.get('status') || undefined,
    sp.get('agent')  || undefined,
    sp.get('owner')  || undefined,
  )
  return NextResponse.json({ tasks })
}

export async function POST(request: NextRequest, { params }: Params) {
  const { name } = await params
  const body     = await request.json() as Record<string, string>
  const args     = ['team', 'tasks', 'create', name, JSON.stringify(body['title'] || 'task')]
  if (body['desc'])  args.push('--desc',  JSON.stringify(body['desc']))
  if (body['dod'])   args.push('--dod',   JSON.stringify(body['dod']))
  if (body['owner']) args.push('--owner', body['owner'])

  return new Promise<NextResponse>((resolve) => {
    exec(`jikime ${args.join(' ')}`, (err, stdout) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true, output: stdout.trim() }, { status: 201 }))
    })
  })
}
