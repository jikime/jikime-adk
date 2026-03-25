import { NextRequest, NextResponse } from 'next/server'
import { execFile } from 'child_process'

type Params = { params: Promise<{ name: string }> }

export async function POST(request: NextRequest, { params }: Params) {
  const { name } = await params
  const body     = await request.json() as Record<string, string>
  const to       = (body['to']   || '').slice(0, 80)
  const msg      = (body['body'] || body['message'] || '').slice(0, 4000)
  const from     = (body['from'] || 'webchat').slice(0, 80)
  if (!to || !msg) return NextResponse.json({ error: 'to and body required' }, { status: 400 })

  return new Promise<NextResponse>((resolve) => {
    execFile('jikime', ['team', 'inbox', 'send', name, to, msg, '--from', from], (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
