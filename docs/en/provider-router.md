# Provider Router

A proxy system that routes Claude Code's API requests to external LLM providers (OpenAI, Gemini, GLM, Ollama).

## Overview

```
Claude Code ─── ANTHROPIC_BASE_URL ──→ jikime router (localhost:8787/{provider})
                                            │
                                    ┌───────┴───────┐
                                    ▼               ▼
                              OpenAI-Compatible    Gemini API
                              (OpenAI/GLM/Ollama)
```

It intercepts Claude Code's API requests via the `ANTHROPIC_BASE_URL` environment variable and transforms them into the format of the provider specified in the URL path.

### URL-Based Routing

The router identifies the provider from the URL path:

```
http://localhost:8787/{provider}/v1/messages
                       ↑
                       Provider name (openai, gemini, ollama, etc.)
```

This approach allows **multiple Claude Code sessions to use different providers simultaneously**. A single router instance handles all provider requests.

## Supported Providers

| Provider | Connection Method | Description |
|----------|------------------|-------------|
| OpenAI | Via proxy | chat/completions format conversion |
| Gemini | Via proxy | generateContent format conversion |
| GLM | Direct connection | Uses Z.ai Anthropic-compatible endpoint |
| Ollama | Via proxy | Uses OpenAI-compatible mode |

## Configuration

### Configuration File

`~/.jikime/router.yaml`:

```yaml
router:
  port: 8787
  host: "127.0.0.1"

providers:
  openai:
    model: gpt-5.1
    base_url: https://api.openai.com/v1

  gemini:
    model: gemini-2.5-flash
    base_url: https://generativelanguage.googleapis.com

  glm:
    model: glm-4.7
    base_url: https://api.z.ai/api/paas/v4
    anthropic_url: https://api.z.ai/api/anthropic
    region: international

  ollama:
    model: llama3.1
    base_url: http://localhost:11434

# Scenario-based routing (optional)
scenarios:
  default: openai/gpt-5.1
  background: ollama/llama3.1
  think: openai/o1
  long_context: gemini/gemini-2.5-flash
  long_context_threshold: 60000
```

API keys are not stored in `router.yaml`. They are automatically read from environment variables.

### API Key Configuration

Set API keys as environment variables:

```bash
# Add to ~/.zshrc or ~/.bashrc
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="AI..."
export GLM_API_KEY="..."
```

Ollama runs locally, so no API key is required.

## Usage

### Switching Providers

```bash
# Switch to GLM (no proxy needed, direct connection)
jikime router switch glm

# Switch to OpenAI (proxy starts automatically)
jikime router switch openai

# Switch to Gemini (proxy starts automatically)
jikime router switch gemini

# Switch to Ollama (proxy starts automatically)
jikime router switch ollama

# Restore to native Claude
jikime router switch claude
```

The `switch` command automatically updates the current project's `.claude/settings.local.json`.
It must be run inside a project directory (where `.git` or `.claude` exists).
**Claude Code restart is required** after changing.

Different providers can be used for different projects:

```bash
# Project A: Cost savings (GLM)
cd ~/projects/project-a && jikime router switch glm

# Project B: High quality (Native Claude)
cd ~/projects/project-b && jikime router switch claude
```

### Model Selection

You can specify a particular model using the `provider/model` format:

```bash
# Default model (uses the model value from router.yaml)
jikime router switch openai          # → gpt-5.1 (config default)

# Specify a particular model
jikime router switch openai/gpt-4o-mini
jikime router switch openai/o1
jikime router switch gemini/gemini-2.5-pro
jikime router switch glm/glm-4.7
jikime router switch ollama/deepseek-r1
jikime router switch ollama/llama3.1:70b
```

When a model is specified, it will be used instead of the default model in `router.yaml`.

### Router Management

```bash
# Start router (foreground)
jikime router start

# Start router (background daemon)
jikime router start -d

# Specify port
jikime router start -d -p 9090

# Check status
jikime router status

# Send test request (provider specification required)
jikime router test openai
jikime router test gemini

# Stop router
jikime router stop
```

## How It Works

### Proxy Mode (OpenAI, Gemini, Ollama)

1. The `switch` command automatically starts the router (does not restart if already running)
2. Sets `ANTHROPIC_BASE_URL=http://localhost:8787/{provider}` in `.claude/settings.local.json`
3. Claude Code sends requests to the local router (provider included in URL path)
4. Router identifies the provider from the URL path and converts Anthropic format → provider format
5. Converts provider response back to Anthropic format and returns it

**Multi-session support**: Since one router handles all providers, multiple Claude Code sessions can use different providers simultaneously.

### Direct Mode (GLM)

1. The `switch` command sets directly in `.claude/settings.local.json`:
   - `ANTHROPIC_BASE_URL`: Z.ai's Anthropic-compatible endpoint
   - `ANTHROPIC_API_KEY`: GLM API key
   - `ANTHROPIC_DEFAULT_*_MODEL`: GLM model name
2. Claude Code sends requests directly to Z.ai (no proxy needed)

### Streaming

SSE (Server-Sent Events) streaming is supported for all providers:

- OpenAI: `chat/completions` SSE → Anthropic SSE
- Gemini: `streamGenerateContent?alt=sse` → Anthropic SSE
- Ollama: OpenAI-compatible SSE → Anthropic SSE

### Tool Use

Supports Claude Code's tool_use (file read/write, code execution, etc.):

- OpenAI/Ollama: `tool_calls` ↔ `tool_use` content block conversion
- Gemini: `functionCall`/`functionResponse` ↔ `tool_use`/`tool_result` conversion

### API Parameter Conversion

Automatically handles API compatibility for each provider:

#### OpenAI

- **max_tokens handling**: Latest models (gpt-5.x, o1, o3, o4 series) require the `max_completion_tokens` parameter, but these models return 400 errors when the limit is too low, so output token limits are not sent and model defaults are used. Legacy models (gpt-4o, gpt-4, etc.) send `max_tokens` normally.
- **tool_choice conversion**: `any` → `required`, `tool` → `function` format

#### Gemini

- **JSON Schema cleanup**: Recursively removes schema fields not supported by Gemini (`exclusiveMinimum`, `additionalProperties`, `propertyNames`, `$schema`, `exclusiveMaximum`). Handles nested `properties`, `items`, and `allOf`/`anyOf`/`oneOf` contents.
- **system_instruction**: Converts system messages to Gemini's `system_instruction` format

## Statusline Integration

Router status is automatically reflected in Claude Code's statusline.

### Display Format

```
# When router is active
🤖 openai/gpt-5.1

# When model is specified
🤖 gemini/gemini-2.5-pro
🤖 glm/glm-4.7
🤖 ollama/deepseek-r1

# When using native Claude
🤖 Claude Opus 4.5
```

### How It Works

The `switch` command records the current state in `~/.jikime/router-state.json`:

```json
{
  "provider": "openai",
  "model": "gpt-5.1",
  "mode": "proxy",
  "active": true
}
```

The statusline reads this file and displays in `provider/model` format.
When `switch claude` is executed, the state file is deleted and the native model name is displayed.

## File Structure

```
cmd/routercmd/
├── router.go          # Parent command
├── start.go           # jikime router start
├── stop.go            # jikime router stop
├── status.go          # jikime router status
├── switch.go          # jikime router switch
└── test.go            # jikime router test

internal/router/
├── config.go          # Config loader
├── server.go          # HTTP proxy server
├── handler.go         # /v1/messages handler
├── stream.go          # SSE utilities
├── types/
│   └── types.go       # Anthropic API types
└── provider/
    ├── provider.go    # Provider interface
    ├── openai.go      # OpenAI provider
    ├── gemini.go      # Gemini provider
    ├── glm.go         # GLM provider (OpenAI wrapper)
    └── ollama.go      # Ollama provider (OpenAI wrapper)
```

## Debugging

To check router requests/responses, run in foreground mode:

```bash
# Stop existing router
jikime router stop

# Run in foreground (logs output to terminal)
jikime router start
```

Log example:
```
[router] 2026/01/25 01:30:00 Starting on 127.0.0.1:8787 (providers: openai, gemini, glm, ollama)
[router] 2026/01/25 01:30:05 -> openai/gpt-5.1 (stream=true)
[router] 2026/01/25 01:30:10 <- Provider error (400): {...}
```

## Troubleshooting

### API Key Error

```
Error: API key environment variable not set for 'openai'
```

Check if the environment variable for the provider is set:

```bash
echo $OPENAI_API_KEY
```

### Router Connection Failed

```
Error: router is not running
```

Start the router first or use `switch` (starts automatically):

```bash
jikime router switch openai
```

### Not Applied to Claude Code

After `switch`, you must restart Claude Code.
`.claude/settings.local.json` is only read at session start.

### Gemini 400 Error (Schema Related)

```
Unsupported value for 'properties.exclusiveMinimum'
```

This occurs when there's a tool using JSON Schema fields not supported by Gemini.
The router automatically removes these, so update to the latest binary:

```bash
jikime-adk update
```

### Ollama Connection Error

Check if Ollama is running locally:

```bash
ollama list
# or
curl http://localhost:11434/api/tags
```

### GLM Region Configuration

When using in China, change the region in `router.yaml`:

```yaml
glm:
  region: china  # base_url switches to open.bigmodel.cn
```

## Technical Details

### API Key Management

API keys are never stored in `router.yaml` (`yaml:"-"` tag applied).
They are read only from environment variables and used only in memory:

| Provider | Environment Variable |
|----------|---------------------|
| OpenAI | `OPENAI_API_KEY` |
| Gemini | `GEMINI_API_KEY` |
| GLM | `GLM_API_KEY` |
| Ollama | Not required (local) |

### Token Handling by Supported Models

| Model Group | max_tokens Handling | Reason |
|-------------|---------------------|--------|
| gpt-4o, gpt-4, gpt-3.5 | Sends `max_tokens` | Legacy API compatibility |
| gpt-5.x | Not sent (model default) | Returns 400 error when limit is low |
| o1, o3, o4 series | Not sent (model default) | `max_completion_tokens` required but no limit needed |

### URL-Based Routing Behavior

When the `switch` command is executed:
1. If the router is not running, it starts in the background
2. Sets `ANTHROPIC_BASE_URL=http://localhost:8787/{provider}` in `.claude/settings.local.json`

**Multi-session support**: Since the router identifies the provider from the URL path, a single started router handles all provider requests. There's no need to restart the router when switching providers.

Example:
- Session A: `ANTHROPIC_BASE_URL=http://localhost:8787/openai` → Uses OpenAI
- Session B: `ANTHROPIC_BASE_URL=http://localhost:8787/gemini` → Uses Gemini
- The same router instance handles both sessions
