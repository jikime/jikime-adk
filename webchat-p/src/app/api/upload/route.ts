import { NextRequest } from 'next/server'
import { writeFileSync, mkdirSync } from 'fs'
import { join } from 'path'

export const dynamic = 'force-dynamic'

const UPLOAD_DIR = '/tmp/webchat-p-uploads'

export async function POST(request: NextRequest) {
  try {
    mkdirSync(UPLOAD_DIR, { recursive: true })

    const formData = await request.formData()
    const files = formData.getAll('files') as File[]

    if (!files.length) {
      return Response.json({ error: 'No files' }, { status: 400 })
    }

    const results: { name: string; path: string; size: number; type: string }[] = []

    for (const file of files) {
      const timestamp = Date.now()
      const safeName = file.name.replace(/[^a-zA-Z0-9._-]/g, '_')
      const filename = `${timestamp}_${safeName}`
      const filePath = join(UPLOAD_DIR, filename)

      const buffer = Buffer.from(await file.arrayBuffer())
      writeFileSync(filePath, buffer)

      results.push({
        name: file.name,
        path: filePath,
        size: file.size,
        type: file.type,
      })
    }

    return Response.json(results)
  } catch (err) {
    return Response.json({ error: (err as Error).message }, { status: 500 })
  }
}
