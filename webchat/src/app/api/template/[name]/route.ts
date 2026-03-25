import { NextRequest, NextResponse } from 'next/server'
import * as fs   from 'fs'
import * as path from 'path'
import { templatesDir } from '@/lib/team-store'

const BUILTIN  = ['leader-worker', 'leader-worker-reviewer', 'parallel-workers']
const NAME_RE  = /^[a-zA-Z0-9_-]{1,64}$/

type Params = { params: Promise<{ name: string }> }

export async function DELETE(_req: NextRequest, { params }: Params) {
  const { name } = await params

  // Whitelist validation — prevents path traversal via ../.. segments
  if (!NAME_RE.test(name)) {
    return NextResponse.json({ error: 'Invalid template name' }, { status: 400 })
  }
  if (BUILTIN.includes(name)) {
    return NextResponse.json({ error: '내장 템플릿은 삭제할 수 없습니다' }, { status: 403 })
  }

  const dir      = templatesDir()
  const filePath = path.join(dir, `${name}.yaml`)

  // Boundary check — ensure resolved path stays within templatesDir
  if (!filePath.startsWith(dir + path.sep) && filePath !== path.join(dir, `${name}.yaml`)) {
    return NextResponse.json({ error: 'Invalid template path' }, { status: 400 })
  }

  try {
    if (fs.existsSync(filePath)) fs.unlinkSync(filePath)
    return NextResponse.json({ ok: true })
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 500 })
  }
}
