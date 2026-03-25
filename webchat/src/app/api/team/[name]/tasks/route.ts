import { NextRequest, NextResponse } from 'next/server'
import { execFile } from 'child_process'
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
  const title    = (body['title'] || 'task').slice(0, 200)
  const args: string[] = ['team', 'tasks', 'create', name, title]
  if (body['desc'])  args.push('--desc',  body['desc'].slice(0, 1000))
  if (body['dod'])   args.push('--dod',   body['dod'].slice(0, 1000))
  if (body['owner']) args.push('--owner', body['owner'].slice(0, 80))

  return new Promise<NextResponse>((resolve) => {
    execFile('jikime', args, (err, stdout) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true, output: stdout.trim() }, { status: 201 }))
    })
  })
}
