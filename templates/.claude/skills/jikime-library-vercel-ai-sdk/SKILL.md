---
name: jikime-library-vercel-ai-sdk
description: Vercel AI SDK v5/v6 implementation guide with useChat, tool(), streamText, AI Elements, and ToolLoopAgent patterns. Use when building AI-powered applications with Next.js.
version: 1.0.0
tags: ["library", "ai", "vercel", "llm", "streaming", "useChat", "tools"]
triggers:
  keywords: ["ai-sdk", "useChat", "streamText", "tool()", "LLM", "AI 애플리케이션"]
  phases: ["run"]
  agents: ["backend", "frontend"]
  languages: ["typescript"]
# Progressive Disclosure Configuration
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~3617
user-invocable: false
context: fork
agent: ai-sdk-specialist
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__query-docs
---

# JikiME Vercel AI SDK Skill

Comprehensive guide for building AI-powered applications with Vercel AI SDK v5/v6.

## Overview

Vercel AI SDK provides a unified API for working with large language models (LLMs) in JavaScript/TypeScript applications.

## Quick Reference

| SDK Version | Key Features |
|-------------|--------------|
| **v5** | useChat, tool(), streamText, inputSchema |
| **v6** | ToolLoopAgent, Output patterns, enhanced streaming |

---

## CRITICAL: API Differences from v4

### Tool Definition

```typescript
// ❌ WRONG (v4 legacy)
const weatherTool = {
  parameters: z.object({
    city: z.string()
  }),
  execute: async ({ city }) => { /* ... */ }
}

// ✅ CORRECT (v5+) - MUST use tool() helper with inputSchema
import { tool } from 'ai'

const weatherTool = tool({
  description: 'Get weather for a city',
  inputSchema: z.object({  // NOT "parameters"!
    city: z.string().describe('City name'),
  }),
  execute: async ({ city }) => {
    const weather = await fetchWeather(city)
    return weather
  },
})
```

### useChat Hook

```typescript
// ❌ WRONG (v4 legacy)
const { messages, append } = useChat()
append({ content: input, role: 'user' })

// ✅ CORRECT (v5+)
const { messages, sendMessage } = useChat()
sendMessage({ text: input })  // Use sendMessage, not append!
```

### Message Content

```typescript
// ❌ WRONG (v4 legacy)
{messages.map(m => (
  <div>{m.content}</div>
))}

// ✅ CORRECT (v5+) - Use message.parts
{messages.map(m => (
  <div>
    {m.parts.map((part, i) => {
      if (part.type === 'text') return <span key={i}>{part.text}</span>
      if (part.type === 'tool-call') return <ToolCall key={i} {...part} />
      if (part.type === 'tool-result') return <ToolResult key={i} {...part} />
    })}
  </div>
))}
```

---

## Installation

```bash
# Core package
npm install ai

# Provider packages
npm install @ai-sdk/openai
npm install @ai-sdk/anthropic
npm install @ai-sdk/google

# For AI Elements (UI components)
npm install @anthropic-ai/ai-elements
```

---

## Core Patterns

### 1. Basic Text Generation

```typescript
// app/api/chat/route.ts
import { streamText } from 'ai'
import { openai } from '@ai-sdk/openai'

export async function POST(req: Request) {
  const { messages } = await req.json()

  const result = streamText({
    model: openai('gpt-4o'),  // String format: 'openai/gpt-4o'
    messages,
    system: 'You are a helpful assistant.',
  })

  return result.toDataStreamResponse()
}
```

### 2. Client-Side Chat

```tsx
// components/chat.tsx
'use client'

import { useChat } from '@ai-sdk/react'

export function Chat() {
  const {
    messages,
    input,
    setInput,
    sendMessage,  // NOT append!
    isLoading,
    stop,
    reload,
  } = useChat({
    api: '/api/chat',
  })

  return (
    <div>
      <div className="messages">
        {messages.map(m => (
          <div key={m.id} className={m.role}>
            {m.parts.map((part, i) => {
              if (part.type === 'text') {
                return <p key={i}>{part.text}</p>
              }
              return null
            })}
          </div>
        ))}
      </div>

      <form onSubmit={(e) => {
        e.preventDefault()
        sendMessage({ text: input })
        setInput('')
      }}>
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type a message..."
        />
        <button type="submit" disabled={isLoading}>
          {isLoading ? 'Sending...' : 'Send'}
        </button>
        {isLoading && <button onClick={stop}>Stop</button>}
      </form>
    </div>
  )
}
```

---

## Tool Calling

### Define Tools with inputSchema

```typescript
// lib/tools.ts
import { tool } from 'ai'
import { z } from 'zod'

export const weatherTool = tool({
  description: 'Get the current weather for a location',
  inputSchema: z.object({
    city: z.string().describe('The city to get weather for'),
    unit: z.enum(['celsius', 'fahrenheit']).default('celsius'),
  }),
  execute: async ({ city, unit }) => {
    const weather = await fetchWeatherAPI(city)
    return {
      city,
      temperature: unit === 'celsius' ? weather.tempC : weather.tempF,
      condition: weather.condition,
    }
  },
})

export const searchTool = tool({
  description: 'Search the web for information',
  inputSchema: z.object({
    query: z.string().describe('Search query'),
    maxResults: z.number().default(5),
  }),
  execute: async ({ query, maxResults }) => {
    const results = await searchAPI(query, maxResults)
    return results
  },
})
```

### Use Tools in API Route

```typescript
// app/api/chat/route.ts
import { streamText } from 'ai'
import { openai } from '@ai-sdk/openai'
import { weatherTool, searchTool } from '@/lib/tools'

export async function POST(req: Request) {
  const { messages } = await req.json()

  const result = streamText({
    model: openai('gpt-4o'),
    messages,
    tools: {
      weather: weatherTool,
      search: searchTool,
    },
    maxSteps: 5,  // Allow multi-step tool use
  })

  return result.toDataStreamResponse()
}
```

### Display Tool Results

```tsx
// components/message.tsx
'use client'

import type { Message } from '@ai-sdk/react'

export function ChatMessage({ message }: { message: Message }) {
  return (
    <div className={`message ${message.role}`}>
      {message.parts.map((part, index) => {
        switch (part.type) {
          case 'text':
            return <p key={index}>{part.text}</p>

          case 'tool-call':
            return (
              <div key={index} className="tool-call">
                <span className="tool-name">{part.toolName}</span>
                <pre>{JSON.stringify(part.args, null, 2)}</pre>
              </div>
            )

          case 'tool-result':
            return (
              <div key={index} className="tool-result">
                <span className="tool-name">{part.toolName}</span>
                <pre>{JSON.stringify(part.result, null, 2)}</pre>
              </div>
            )

          default:
            return null
        }
      })}
    </div>
  )
}
```

---

## AI SDK v6 Features

### ToolLoopAgent

```typescript
// lib/agents/tool-loop-agent.ts
import { streamText, tool } from 'ai'
import { openai } from '@ai-sdk/openai'
import { z } from 'zod'

export async function runToolLoopAgent(userMessage: string) {
  const tools = {
    search: tool({
      description: 'Search for information',
      inputSchema: z.object({ query: z.string() }),
      execute: async ({ query }) => await search(query),
    }),
    calculate: tool({
      description: 'Perform calculations',
      inputSchema: z.object({ expression: z.string() }),
      execute: async ({ expression }) => eval(expression),
    }),
  }

  let messages = [{ role: 'user' as const, content: userMessage }]
  let iteration = 0
  const maxIterations = 10

  while (iteration < maxIterations) {
    const result = await streamText({
      model: openai('gpt-4o'),
      messages,
      tools,
    })

    const response = await result.text
    const toolCalls = await result.toolCalls

    if (toolCalls.length === 0) {
      return response  // No more tool calls, return final response
    }

    // Add assistant message with tool calls
    messages.push({
      role: 'assistant',
      content: response,
      toolCalls,
    })

    // Execute tools and add results
    for (const toolCall of toolCalls) {
      const toolResult = await tools[toolCall.toolName].execute(toolCall.args)
      messages.push({
        role: 'tool',
        toolCallId: toolCall.toolCallId,
        content: JSON.stringify(toolResult),
      })
    }

    iteration++
  }

  throw new Error('Max iterations reached')
}
```

### Output Patterns (v6)

```typescript
// generateObject and streamObject deprecated
// Use Output.object({ schema }) instead

import { streamText, Output } from 'ai'
import { openai } from '@ai-sdk/openai'
import { z } from 'zod'

const productSchema = z.object({
  name: z.string(),
  description: z.string(),
  price: z.number(),
  features: z.array(z.string()),
})

const result = await streamText({
  model: openai('gpt-4o'),
  prompt: 'Generate a product description for a smartwatch',
  output: Output.object({ schema: productSchema }),
})

const product = await result.object  // Typed as Product
```

---

## AI Elements (UI Components)

### Installation

```bash
npm install @anthropic-ai/ai-elements streamdown shiki
```

### Core Components

```tsx
// components/ai-chat.tsx
'use client'

import {
  Conversation,
  Message,
  PromptInput,
  Reasoning,
  Sources,
  Tool,
} from '@anthropic-ai/ai-elements'
import { useChat } from '@ai-sdk/react'

export function AIChat() {
  const { messages, input, setInput, sendMessage, isLoading } = useChat()

  return (
    <div className="flex flex-col h-screen">
      <Conversation className="flex-1 overflow-auto">
        {messages.map((m) => (
          <Message
            key={m.id}
            role={m.role}
            className={m.role === 'user' ? 'bg-blue-100' : 'bg-gray-100'}
          >
            {m.parts.map((part, i) => {
              if (part.type === 'text') {
                return <span key={i}>{part.text}</span>
              }
              if (part.type === 'tool-call') {
                return (
                  <Tool key={i} name={part.toolName} status="running">
                    <pre>{JSON.stringify(part.args, null, 2)}</pre>
                  </Tool>
                )
              }
              if (part.type === 'tool-result') {
                return (
                  <Tool key={i} name={part.toolName} status="complete">
                    <pre>{JSON.stringify(part.result, null, 2)}</pre>
                  </Tool>
                )
              }
              if (part.type === 'reasoning') {
                return (
                  <Reasoning key={i} collapsed>
                    {part.text}
                  </Reasoning>
                )
              }
              return null
            })}
          </Message>
        ))}
      </Conversation>

      <PromptInput
        value={input}
        onChange={setInput}
        onSubmit={() => sendMessage({ text: input })}
        loading={isLoading}
        placeholder="Ask me anything..."
      />
    </div>
  )
}
```

### Sources Component

```tsx
import { Sources, Source } from '@anthropic-ai/ai-elements'

function SearchResults({ results }) {
  return (
    <Sources>
      {results.map((r, i) => (
        <Source
          key={i}
          title={r.title}
          url={r.url}
          snippet={r.snippet}
        />
      ))}
    </Sources>
  )
}
```

---

## Full-Stack AI App Structure

```
src/
├── app/
│   ├── layout.tsx
│   ├── page.tsx
│   └── api/
│       └── chat/
│           └── route.ts
├── components/
│   ├── chat/
│   │   ├── chat.tsx
│   │   ├── message.tsx
│   │   └── input.tsx
│   └── ui/
│       └── (shadcn components)
├── lib/
│   ├── ai/
│   │   ├── tools.ts
│   │   ├── agents.ts
│   │   └── prompts.ts
│   └── utils.ts
└── types/
    └── ai.ts
```

---

## Model Configuration

### Provider Setup

```typescript
// lib/ai/providers.ts
import { createOpenAI } from '@ai-sdk/openai'
import { createAnthropic } from '@ai-sdk/anthropic'
import { createGoogleGenerativeAI } from '@ai-sdk/google'

export const openai = createOpenAI({
  apiKey: process.env.OPENAI_API_KEY,
})

export const anthropic = createAnthropic({
  apiKey: process.env.ANTHROPIC_API_KEY,
})

export const google = createGoogleGenerativeAI({
  apiKey: process.env.GOOGLE_API_KEY,
})
```

### Model Selection

```typescript
// Recommended string format for model specification
const models = {
  openai: 'openai/gpt-4o',
  anthropic: 'anthropic/claude-3-5-sonnet-20241022',
  google: 'google/gemini-1.5-pro',
}

// Or use provider instances
import { openai } from '@ai-sdk/openai'
const model = openai('gpt-4o')
```

---

## Error Handling

```typescript
import { streamText, AISDKError } from 'ai'

try {
  const result = await streamText({
    model: openai('gpt-4o'),
    messages,
  })
  return result.toDataStreamResponse()
} catch (error) {
  if (error instanceof AISDKError) {
    console.error('AI SDK Error:', error.message, error.cause)
    return new Response(error.message, { status: 500 })
  }
  throw error
}
```

---

## Best Practices

### 1. Always use tool() helper

```typescript
// ✅ CORRECT
import { tool } from 'ai'
const myTool = tool({
  description: '...',
  inputSchema: z.object({ /* ... */ }),
  execute: async (args) => { /* ... */ },
})

// ❌ WRONG - plain object
const myTool = {
  parameters: z.object({ /* ... */ }),
  execute: async (args) => { /* ... */ },
}
```

### 2. Use inputSchema, not parameters

```typescript
// ✅ CORRECT
inputSchema: z.object({ city: z.string() })

// ❌ WRONG
parameters: z.object({ city: z.string() })
```

### 3. Use sendMessage, not append

```typescript
// ✅ CORRECT
sendMessage({ text: input })

// ❌ WRONG
append({ content: input, role: 'user' })
```

### 4. Use message.parts, not message.content

```typescript
// ✅ CORRECT
message.parts.map(part => {
  if (part.type === 'text') return part.text
})

// ❌ WRONG
message.content
```

### 5. Set maxSteps for tool loops

```typescript
// ✅ CORRECT - explicit limit
streamText({
  model,
  messages,
  tools,
  maxSteps: 5,  // Prevent infinite loops
})
```

---

## Migration from v4

| v4 Pattern | v5/v6 Pattern |
|------------|---------------|
| `parameters` | `inputSchema` |
| `append()` | `sendMessage()` |
| `message.content` | `message.parts` |
| `generateObject()` | `Output.object({ schema })` |
| Plain tool object | `tool()` helper |

---

Version: 1.0.0
Last Updated: 2026-01-22
Sources: wsimmonds/vercel-ai-sdk, laguagu/ai-sdk-6, laguagu/ai-elements
