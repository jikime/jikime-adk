import { NextRequest } from 'next/server'
import { readdirSync, readFileSync, existsSync } from 'fs'
import { join } from 'path'
import os from 'os'

export const dynamic = 'force-dynamic'

export interface SlashCommand {
  namespace: string
  command: string
  description: string
  source: 'project' | 'global'
}

/** frontmatter에서 description 추출 */
function extractDescription(content: string): string {
  const match = content.match(/^---[\s\S]*?^description:\s*["']?(.+?)["']?\s*$/m)
  return match?.[1]?.trim() ?? ''
}

/** 디렉토리 스캔 → SlashCommand 목록 */
function scanCommands(baseDir: string, source: 'project' | 'global'): SlashCommand[] {
  if (!existsSync(baseDir)) return []
  const cmds: SlashCommand[] = []
  try {
    for (const ns of readdirSync(baseDir)) {
      const nsPath = join(baseDir, ns)
      try {
        for (const file of readdirSync(nsPath)) {
          if (!file.endsWith('.md')) continue
          const command = file.slice(0, -3)
          let description = ''
          try {
            description = extractDescription(readFileSync(join(nsPath, file), 'utf8'))
          } catch { /* skip */ }
          cmds.push({ namespace: ns, command, description, source })
        }
      } catch { /* skip */ }
    }
  } catch { /* skip */ }
  return cmds
}

export async function GET(request: NextRequest) {
  const cwd = request.nextUrl.searchParams.get('cwd')?.trim() || ''

  // 프로젝트 레벨 → 글로벌 순서로 스캔
  const projectCmds = cwd
    ? scanCommands(join(cwd, '.claude', 'commands'), 'project')
    : []
  const globalCmds = scanCommands(
    join(os.homedir(), '.claude', 'commands'),
    'global'
  )

  // 프로젝트 커맨드가 글로벌 커맨드보다 우선 (중복 제거)
  const seen = new Set(projectCmds.map(c => `${c.namespace}:${c.command}`))
  const merged = [
    ...projectCmds,
    ...globalCmds.filter(c => !seen.has(`${c.namespace}:${c.command}`)),
  ]

  // namespace 기준 정렬, 같은 namespace 내에서 command 정렬
  merged.sort((a, b) =>
    a.namespace !== b.namespace
      ? a.namespace.localeCompare(b.namespace)
      : a.command.localeCompare(b.command)
  )

  return Response.json(merged)
}
