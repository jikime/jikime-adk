import { NextRequest, NextResponse } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import * as os from 'os'

export const dynamic = 'force-dynamic'

// POST /api/ws/project — 새 프로젝트 경로 등록
// Claude Code 세션이 없는 경로도 등록 가능
export async function POST(request: NextRequest) {
  try {
    const body = await request.json() as { path?: string }
    const projectPath = body.path

    if (!projectPath || typeof projectPath !== 'string') {
      return NextResponse.json({ error: 'Missing path' }, { status: 400 })
    }

    const normalized = projectPath.replace(/\/+$/, '') // trailing slash 제거

    if (!normalized.startsWith('/')) {
      return NextResponse.json({ error: '절대 경로를 입력하세요' }, { status: 400 })
    }

    // Claude Code 인코딩: 경로의 '/' → '-'
    const encoded = normalized.replace(/\//g, '-')
    const claudeDir = path.join(os.homedir(), '.claude', 'projects')
    const projectDir = path.join(claudeDir, encoded)

    // 경로 탈출 방지
    if (!projectDir.startsWith(claudeDir + path.sep) && projectDir !== claudeDir) {
      return NextResponse.json({ error: 'Invalid path' }, { status: 400 })
    }

    if (!fs.existsSync(projectDir)) {
      fs.mkdirSync(projectDir, { recursive: true })
    }

    // 실제 프로젝트 폴더도 없으면 생성
    if (!fs.existsSync(normalized)) {
      fs.mkdirSync(normalized, { recursive: true })
    }

    // 원본 경로를 파일로 저장 — decodeProjectPath가 실패할 경우 fallback으로 사용
    fs.writeFileSync(
      path.join(projectDir, '_webchat_path'),
      normalized,
      'utf8',
    )

    const project = {
      id: encoded,
      name: path.basename(normalized) || encoded,
      path: normalized,
      sessions: [],
    }

    return NextResponse.json({ ok: true, project })
  } catch {
    return NextResponse.json({ error: '서버 오류' }, { status: 500 })
  }
}
