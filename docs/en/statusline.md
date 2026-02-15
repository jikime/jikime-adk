# Statusline

A custom renderer that displays session status information at the bottom of the Claude Code terminal.

## Overview

The `jikime statusline` command is an external renderer called when Claude Code uses the statusline feature. It receives session context as JSON from Claude Code and returns a formatted status string.

```
Claude Code → stdin (JSON) → jikime statusline → stdout (status string)
```

## Display Information

| Icon | Item | Description |
|------|------|-------------|
| 🤖 | Model | AI model name (e.g., Opus 4.5) |
| ▰▱ | Progress | Context window usage (Progress Bar) |
| 💵 | Cost | Estimated token cost (e.g., $0.23) |
| 💬 | OutputStyle | Response style/persona (e.g., J.A.R.V.I.S.) |
| 📁 | Directory | Current project directory |
| 🔀 | Branch | Git branch and status |
| 💾 | Memory | Memory usage |
| ⚡ | CPU | System CPU load |
| 💿 | Disk | Disk usage |
| 🌐 | Network | API response latency |
| 🌤️ | Weather | Current weather (optional) |
| ⏱️ | Duration | Session duration |
| 🎯 | Task | Active task display |
| 📦 | Version | JikiME-ADK version |
| 🔄 | Update | Update availability |

## Usage

```bash
# Basic usage (extended mode, includes progress bar)
jikime statusline

# Display in specific mode
jikime statusline --mode compact
jikime statusline --mode minimal
jikime statusline --mode geek

# View demo
jikime statusline --demo

# Display in pretty box format
jikime statusline --pretty
```

## Display Modes

### Minimal

Displays only model and context.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 7%
```

### Compact

Displays core information in a compressed format.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 15K/200K 7% ┃ 💵 $0.23 ┃ 💬 J.A.R.V.I.S. ┃ 🔀 main +0 M5 ?5 ┃ 💾 128MB ┃ ⚡ 45% ┃ ☀️ +12°C
```

### Extended (Default)

Displays balanced information with progress bar.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 15K/200K ┃ 💵 $0.23 ┃ 💬 J.A.R.V.I.S. ┃ 📁 jikime-adk ┃ 🔀 main +0 M5 ?5 ┃ 💾 128MB ┃ ⚡ 45% ┃ ☀️ +12°C ┃ ⏱️ 45m ┃ 🎯 IMPLEMENT ┃ 📦 v2.0.0
```

### Geek (Full Features)

Developer mode with all features included. Displays color-coded progress bar and all system information.

```
🤖 Opus 4.5 ┃ ▰▱▱▱▱▱▱▱▱▱ 15K/200K (7%) ┃ 💵 $0.23 ┃ 💬 J.A.R.V.I.S. ┃ 📁 jikime-adk ┃ 🔀 main +0 M5 ?5 ┃ 💾 128MB ┃ ⚡ 45% ┃ 💿 120GB (65%) ┃ 🌐 120ms ┃ ☀️ +12°C ┃ ⏱️ 45m ┃ 📦 v2.0.0
```

## Progress Bar

Visually displays context usage:

```
▱▱▱▱▱▱▱▱▱▱ = 0%
▰▰▰▰▰▱▱▱▱▱ = 50%
▰▰▰▰▰▰▰▰▰▰ = 100%
```

Color coding (Geek mode):
- 🟢 Green: 0-49% (Normal)
- 🟡 Yellow: 50-79% (Caution)
- 🔴 Red: 80-100% (Warning)

## Claude Code Configuration

Enable statusline in Claude Code's `settings.json`:

```json
{
  "statusline": {
    "enabled": true,
    "command": "jikime statusline"
  }
}
```

## Configuration File

You can customize the statusline with the `.jikime/config/statusline-config.yaml` file:

```yaml
statusline:
  enabled: true
  mode: extended  # extended | compact | minimal | geek
  refresh_interval_ms: 1000

  display:
    model: true
    version: true
    context_window: true
    output_style: true
    memory_usage: true
    branch: true
    git_status: true
    duration: true
    directory: true
    active_task: true
    update_indicator: true
    # New features
    token_cost: true
    cpu_load: true
    disk_usage: false    # opt-in
    network_latency: false # opt-in
    weather: false       # opt-in
    progress_bar: true

  weather:
    enabled: false
    location: ""  # Empty = auto-detect by IP
    unit: "celsius"  # celsius | fahrenheit

  token_cost:
    input_price_per_mtok: 15.0   # $15 per 1M input tokens
    output_price_per_mtok: 75.0  # $75 per 1M output tokens

  format:
    max_branch_length: 30
    truncate_with: "..."
    separator: " ┃ "

  cache:
    git_ttl_seconds: 10
    update_ttl_seconds: 600
```

## New Features

### Token Cost (💵)

Displays the estimated API cost for the current session. Calculated based on Claude Opus pricing:
- Input: $15 / 1M tokens
- Output: $75 / 1M tokens

### System Status

Displays system status information:
- **CPU Load (⚡)**: Measures system load using the `uptime` command
- **Disk Usage (💿)**: Current directory disk usage via the `df` command
- **Network Latency (🌐)**: Anthropic API response latency (60-second cache)

### Weather (🌤️)

Displays current weather (uses wttr.in API, 30-minute cache):
- Auto-detect location or manual configuration
- Celsius/Fahrenheit unit selection available

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `JIKIME_STATUSLINE_MODE` | Display mode | `extended` |

## Related Files

| File | Description |
|------|-------------|
| `cmd/statuslinecmd/statusline.go` | Statusline command implementation |
| `.jikime/config/statusline-config.yaml` | Configuration file |
| `~/.jikime/metrics/session.json` | Session start time (for Duration calculation) |
| `~/.jikime/state/active_task.json` | Active task information |
