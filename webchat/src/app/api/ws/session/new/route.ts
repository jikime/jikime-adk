import { NextRequest, NextResponse } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import * as os from 'os'
import * as crypto from 'crypto'

export async function POST(req: NextRequest) {
  let body: { projectId?: string }
  try {
    body = await req.json() as { projectId?: string }
  } catch {
    return NextResponse.json({ error: 'Invalid JSON' }, { status: 400 })
  }

  const projectId = body.projectId ?? ''

  // 입력 검증: 비어있거나 경로 탈출 문자 포함 금지
  if (!projectId || projectId.includes('/') || projectId.includes('..')) {
    return NextResponse.json({ error: 'Invalid projectId' }, { status: 400 })
  }

  const claudeDir = path.resolve(path.join(os.homedir(), '.claude', 'projects'))
  const projectDir = path.join(claudeDir, projectId)

  // 경로 탈출 방지
  if (!path.resolve(projectDir).startsWith(claudeDir + path.sep)) {
    return NextResponse.json({ error: 'Invalid projectId' }, { status: 400 })
  }

  try {
    if (!fs.existsSync(projectDir)) fs.mkdirSync(projectDir, { recursive: true })
    const sessionId = crypto.randomUUID()
    fs.writeFileSync(path.join(projectDir, `${sessionId}.jsonl`), '', { encoding: 'utf8' })
    return NextResponse.json({ sessionId })
  } catch (e) {
    console.error('[session/new] error:', e)
    return NextResponse.json({ error: 'Failed to create session' }, { status: 500 })
  }
}
