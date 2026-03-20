import { NextRequest, NextResponse } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import { templatesDir } from '@/lib/team-store'

export async function POST(request: NextRequest) {
  const body = await request.json() as Record<string, string>
  const name = (body['name'] || '').trim().replace(/[^a-zA-Z0-9_-]/g, '-')
  const yaml = body['yaml']
  if (!name || !yaml) return NextResponse.json({ error: 'name and yaml are required' }, { status: 400 })

  const dir = templatesDir()
  try {
    fs.mkdirSync(dir, { recursive: true })
    fs.writeFileSync(path.join(dir, `${name}.yaml`), yaml, 'utf8')
    return NextResponse.json({ ok: true, name })
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 500 })
  }
}
