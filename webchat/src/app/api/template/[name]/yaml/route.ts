import { NextRequest, NextResponse } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import { templatesDir } from '@/lib/team-store'

const NAME_RE = /^[a-zA-Z0-9_-]{1,64}$/

type Params = { params: Promise<{ name: string }> }

export async function GET(_req: NextRequest, { params }: Params) {
  const { name } = await params

  if (!NAME_RE.test(name)) {
    return NextResponse.json({ error: 'Invalid template name' }, { status: 400 })
  }

  const dir      = templatesDir()
  const filePath = path.join(dir, `${name}.yaml`)

  if (!filePath.startsWith(dir + path.sep)) {
    return NextResponse.json({ error: 'Invalid template path' }, { status: 400 })
  }
  if (!fs.existsSync(filePath)) return NextResponse.json({ error: 'Template not found' }, { status: 404 })

  try {
    const yaml = fs.readFileSync(filePath, 'utf8')
    return NextResponse.json({ yaml })
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 500 })
  }
}
