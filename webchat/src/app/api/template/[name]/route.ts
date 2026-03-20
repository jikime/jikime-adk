import { NextRequest, NextResponse } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import { templatesDir } from '@/lib/team-store'

const BUILTIN = ['leader-worker', 'leader-worker-reviewer', 'parallel-workers']

type Params = { params: Promise<{ name: string }> }

export async function DELETE(_req: NextRequest, { params }: Params) {
  const { name } = await params
  if (BUILTIN.includes(name)) {
    return NextResponse.json({ error: '내장 템플릿은 삭제할 수 없습니다' }, { status: 403 })
  }
  const filePath = path.join(templatesDir(), `${name}.yaml`)
  try {
    if (fs.existsSync(filePath)) fs.unlinkSync(filePath)
    return NextResponse.json({ ok: true })
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 500 })
  }
}
