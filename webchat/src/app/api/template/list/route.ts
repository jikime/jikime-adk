import { NextResponse } from 'next/server'
import * as fs from 'fs'
import * as path from 'path'
import { templatesDir } from '@/lib/team-store'

export async function GET() {
  const dir = templatesDir()
  const templates: Array<{ name: string; description: string; defaultBudget: number }> = []
  try {
    const files = fs.readdirSync(dir).filter((f) => f.endsWith('.yaml'))
    for (const f of files) {
      const content       = fs.readFileSync(path.join(dir, f), 'utf8')
      const name          = content.match(/^name:\s*(.+)$/m)?.[1]?.trim().replace(/^['"]|['"]$/g, '') ?? f.replace('.yaml', '')
      const description   = content.match(/^description:\s*["']?([^"'\n]+)["']?$/m)?.[1]?.trim() ?? ''
      const budgetMatch   = content.match(/^default_budget:\s*(\d+)$/m)
      const defaultBudget = budgetMatch ? parseInt(budgetMatch[1]) : 0
      templates.push({ name, description, defaultBudget })
    }
  } catch { /* dir may not exist yet */ }
  return NextResponse.json({ templates })
}
