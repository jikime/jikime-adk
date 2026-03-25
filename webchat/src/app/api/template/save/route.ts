import { NextRequest, NextResponse } from 'next/server'
import * as fs   from 'fs'
import * as path from 'path'
import { load, dump } from 'js-yaml'
import { templatesDir } from '@/lib/team-store'

const NAME_RE      = /^[a-zA-Z0-9_-]{1,64}$/
const MAX_YAML_LEN = 50_000  // 50KB

export async function POST(request: NextRequest) {
  const body = await request.json() as Record<string, string>
  const name = (body['name'] || '').trim()
  const yaml = body['yaml']

  // Strict name validation — prevents path traversal and injection
  if (!NAME_RE.test(name)) {
    return NextResponse.json(
      { error: 'Invalid template name (alphanumeric, _ and - only, 1–64 chars)' },
      { status: 400 },
    )
  }
  if (!yaml) return NextResponse.json({ error: 'yaml is required' }, { status: 400 })
  if (yaml.length > MAX_YAML_LEN) {
    return NextResponse.json({ error: 'YAML too large (max 50 KB)' }, { status: 413 })
  }

  // Parse YAML to validate syntax, then re-serialize — prevents raw YAML injection
  let safeYaml: string
  try {
    const parsed = load(yaml)
    if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
      return NextResponse.json({ error: 'Invalid YAML: must be a mapping object' }, { status: 400 })
    }
    safeYaml = dump(parsed, { lineWidth: -1 })
  } catch (e) {
    return NextResponse.json({ error: `YAML parse error: ${String(e)}` }, { status: 400 })
  }

  try {
    const dir = templatesDir()
    fs.mkdirSync(dir, { recursive: true })
    fs.writeFileSync(path.join(dir, `${name}.yaml`), safeYaml, 'utf8')
    return NextResponse.json({ ok: true, name })
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 500 })
  }
}
