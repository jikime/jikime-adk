import { NextRequest, NextResponse } from 'next/server'
import { exec } from 'child_process'

type Params = { params: Promise<{ name: string; id: string }> }

export async function PATCH(request: NextRequest, { params }: Params) {
  const { name, id } = await params
  const body         = await request.json() as Record<string, string>
  const args         = ['team', 'tasks', 'update', name, id]
  if (body['status'])   args.push('--status', body['status'])
  if (body['agent_id']) args.push('--agent',  body['agent_id'])
  if (body['result'])   args.push('--result',  JSON.stringify(body['result']))

  return new Promise<NextResponse>((resolve) => {
    exec(`jikime ${args.join(' ')}`, (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
