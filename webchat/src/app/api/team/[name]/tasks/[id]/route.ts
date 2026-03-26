import { NextRequest, NextResponse } from 'next/server'
import { execFile } from 'child_process'

const NAME_RE  = /^[a-zA-Z0-9_-]{1,64}$/
const VALID_STATUSES = new Set(['pending', 'in_progress', 'done', 'failed', 'blocked'])

type Params = { params: Promise<{ name: string; id: string }> }

export async function PATCH(request: NextRequest, { params }: Params) {
  const { name, id } = await params
  if (!NAME_RE.test(name)) return NextResponse.json({ error: 'Invalid team name' }, { status: 400 })
  if (!NAME_RE.test(id))   return NextResponse.json({ error: 'Invalid task id' },   { status: 400 })
  const contentLength = request.headers.get('content-length')
  if (contentLength && parseInt(contentLength) > 10_240) {
    return NextResponse.json({ error: 'Request too large' }, { status: 413 })
  }
  const body         = await request.json() as Record<string, string>
  const args: string[] = ['team', 'tasks', 'update', name, id]
  if (body['status'] && VALID_STATUSES.has(body['status'])) args.push('--status', body['status'])
  if (body['agent_id']) args.push('--agent',  body['agent_id'].slice(0, 80))
  if (body['result'])   args.push('--result',  body['result'].slice(0, 2000))

  return new Promise<NextResponse>((resolve) => {
    execFile('jikime', args, (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
