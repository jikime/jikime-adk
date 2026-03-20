import { NextRequest, NextResponse } from 'next/server'
import { exec } from 'child_process'

type Params = { params: Promise<{ name: string }> }

export async function POST(request: NextRequest, { params }: Params) {
  const { name } = await params
  const body     = await request.json() as Record<string, string>
  const to       = body['to'] || ''
  const msg      = body['body'] || body['message'] || ''
  const from     = body['from'] || 'webchat'
  if (!to || !msg) return NextResponse.json({ error: 'to and body required' }, { status: 400 })

  return new Promise<NextResponse>((resolve) => {
    exec(`jikime team inbox send ${name} ${to} ${JSON.stringify(msg)} --from ${from}`, (err) => {
      if (err) resolve(NextResponse.json({ error: err.message }, { status: 500 }))
      else     resolve(NextResponse.json({ ok: true }))
    })
  })
}
