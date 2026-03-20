import { NextRequest } from 'next/server'

const SYSTEM_PROMPT = `You are an expert at creating agent team YAML templates for JiKiME-ADK.
Generate a valid YAML template following this exact schema:

name: <kebab-case-name>
version: "1.0.0"
description: '<short description>'
default_budget: <number>

agents:
  - id: <agent-id>
    role: leader | worker | reviewer
    auto_spawn: true | false
    description: '<what this agent does>'
    task: |
      Goal: {{goal}}

      You are the <role> for {{team_name}}.

      <detailed step-by-step instructions>

Available placeholders in task field:
- {{goal}} — the team's objective
- {{team_name}} — name of the team
- {{agent_id}} — this agent's ID

Rules:
- Always have exactly one leader agent with auto_spawn: true
- Workers check inbox: jikime team inbox receive {{team_name}}
- Workers report to leader: jikime team inbox send {{team_name}} leader "Done: <summary>"
- Workers update task status: jikime team tasks update {{team_name}} <task-id> --status done
- Leader creates tasks: jikime team tasks create {{team_name}} "Title" --desc "..." --dod "..."
- Output ONLY the raw YAML, no markdown fences, no explanation, no commentary`

export async function POST(request: NextRequest) {
  const body         = await request.json() as Record<string, string>
  const userPrompt   = (body['prompt'] || '').trim()
  const existingYaml = (body['existingYaml'] || '').trim()
  if (!userPrompt) {
    return new Response(JSON.stringify({ error: 'prompt is required' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const userMessage = existingYaml
    ? `Modify this existing template based on the request.\n\nExisting YAML:\n${existingYaml}\n\nRequest: ${userPrompt}`
    : `Create a new team template for: ${userPrompt}`

  const fullPrompt = `${SYSTEM_PROMPT}\n\n${userMessage}`
  const encoder    = new TextEncoder()

  const stream = new ReadableStream({
    async start(controller) {
      try {
        const { query } = await import('@anthropic-ai/claude-agent-sdk')
        const queryInstance = query({
          prompt: fullPrompt,
          options: { model: 'claude-haiku-4-5-20251001', settingSources: ['user'] },
        })
        for await (const event of queryInstance) {
          const e = event as Record<string, unknown>
          if (e['type'] === 'assistant') {
            const msg     = e['message'] as Record<string, unknown>
            const content = (msg?.['content'] ?? []) as Array<Record<string, unknown>>
            for (const block of content) {
              if (block['type'] === 'text' && typeof block['text'] === 'string') {
                controller.enqueue(encoder.encode(`data: ${JSON.stringify({ chunk: block['text'] })}\n\n`))
              }
            }
          }
        }
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err)
        controller.enqueue(encoder.encode(`data: ${JSON.stringify({ error: msg })}\n\n`))
      } finally {
        controller.enqueue(encoder.encode(`data: ${JSON.stringify({ done: true })}\n\n`))
        controller.close()
      }
    },
  })

  return new Response(stream, {
    headers: {
      'Content-Type':  'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection':    'keep-alive',
    },
  })
}
