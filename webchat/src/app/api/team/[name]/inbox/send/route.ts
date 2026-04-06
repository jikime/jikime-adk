import { NextRequest, NextResponse } from 'next/server'
import { execFile } from 'child_process'
import { NAME_RE } from '@/lib/validation'

type Params = { params: Promise<{ name: string }> }

export async function POST(request: NextRequest, { params }: Params) {
  const { name } = await params
  if (!NAME_RE.test(name)) return NextResponse.json({ error: 'Invalid team name' }, { status: 400 })
  const contentLength = request.headers.get('content-length')
  if (contentLength && parseInt(contentLength) > 10_240) {
    return NextResponse.json({ error: 'Request too large' }, { status: 413 })
  }
  const body = await request.json() as Record<string, string>
  const to   = (body['to']   || '').slice(0, 80).trim()
  const msg  = (body['body'] || body['message'] || '').slice(0, 4000)
  const from = (body['from'] || 'webchat').slice(0, 80).trim()
  if (!to || !msg) return NextResponse.json({ error: 'to and body required' }, { status: 400 })
  if (!NAME_RE.test(to))   return NextResponse.json({ error: '"to" must match [a-zA-Z0-9_-]{1,80}' }, { status: 400 })
  if (!NAME_RE.test(from)) return NextResponse.json({ error: '"from" must match [a-zA-Z0-9_-]{1,80}' }, { status: 400 })

  return new Promise<NextResponse>((resolve) => {
    execFile('jikime', ['team', 'inbox', 'send', name, to, msg, '--from', from], (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
