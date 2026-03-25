import { NextRequest, NextResponse } from 'next/server'
import { execFile } from 'child_process'

const VALID_STATUSES = new Set(['pending', 'in_progress', 'done', 'failed', 'blocked'])

type Params = { params: Promise<{ name: string; id: string }> }

export async function PATCH(request: NextRequest, { params }: Params) {
  const { name, id } = await params
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
