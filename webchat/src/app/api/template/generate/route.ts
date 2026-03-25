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

const MAX_PROMPT_LEN = 2_000
const MAX_YAML_LEN   = 50_000

/** Neutralize prompt-injection attempts by escaping control chars */
function sanitizeInput(s: string): string {
  return s
    .slice(0, MAX_PROMPT_LEN)
    .replace(/[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]/g, '')  // strip control chars except \t \n \r
    .trim()
}

export async function POST(request: NextRequest) {
  const body         = await request.json() as Record<string, string>
  const userPrompt   = sanitizeInput(body['prompt'] || '')
  const existingYaml = (body['existingYaml'] || '').slice(0, MAX_YAML_LEN)
  if (!userPrompt) {
    return new Response(JSON.stringify({ error: 'prompt is required' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  // Wrap user input in clear delimiters to prevent prompt injection
  const userMessage = existingYaml
    ? `Modify this existing template based on the request.\n\nExisting YAML:\n<yaml>\n${existingYaml}\n</yaml>\n\nRequest: <request>\n${userPrompt}\n</request>`
    : `Create a new team template for: <request>\n${userPrompt}\n</request>`

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
