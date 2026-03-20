import { NextRequest, NextResponse } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import { templatesDir } from '@/lib/team-store'

type Params = { params: Promise<{ name: string }> }

export async function GET(_req: NextRequest, { params }: Params) {
  const { name } = await params
  const filePath = path.join(templatesDir(), `${name}.yaml`)
  if (!fs.existsSync(filePath)) return NextResponse.json({ error: 'Template not found' }, { status: 404 })
  try {
    const yaml = fs.readFileSync(filePath, 'utf8')
    return NextResponse.json({ yaml })
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 500 })
  }
}
