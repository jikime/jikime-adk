# Usage

## Layout

```
+--------------------------------------------------------------------------+
| Header: [=] [Chat] [Terminal] [Files] [Git] [Team]  Project  [Harness] [Settings][Lang][Theme] |
+---------------+--+-------------------------------------------------------+
|  Sidebar      |: |                    Main Panel                          |
|  (resizable)  |  |  [Chat | Terminal | Files | Git | Team]               |
|  Server       |  |                                                       |
|  ----------   |  |  Chat:     Bot icon + "Chat" header + conversation    |
|  Projects     |  |  Terminal: SquareTerminal icon + "Terminal" header     |
|  ----------   |  |  Files:    FolderOpen icon + "Files" + Monaco editor  |
|  Sessions     |  |  Git:      GitBranch icon + "Git" + branch name       |
|  + New Chat   |  |  Team:     Users icon + "Team" + kanban board         |
+---------------+--+-------------------------------------------------------+
                 ^ Drag handle
```

---

## Sidebar Width

Drag the **handle (`:`)** between the sidebar and main panel to resize freely.

| Action | Behavior |
|---|---|
| Drag handle | Adjust sidebar width (160 px ~ 480 px) |
| Header `=` button | Toggle sidebar collapse/expand |

---

## Server Connection Status

| Indicator | Meaning |
|---|---|
| Green dot + `Connected` | WebSocket connection normal |
| Gray dot + `Connecting...` | Connection attempt or disconnected |

Auto-reconnect after 3 seconds on disconnect.

---

## Project and Session Selection

### Projects

Select a project from the sidebar list. The list is read from `~/.claude/projects/` on the server.

### Sessions and URL Routing

Clicking a session changes the URL to `/session/{sessionId}`. Session IDs are `.jsonl` filenames. Browser refresh restores the same session.

### New Chat

Click **+ New Chat** next to a project to create a new session with auto-generated UUID.

---

## Chat (Claude Conversation)

### Sending Messages

Type a message and press `Enter` or click send. Claude responds via streaming.

### Permission Modes

| Mode | Description |
|---|---|
| `bypassPermissions` | Auto-approve all tool use (switches to `acceptEdits` under root) |
| `acceptEdits` | Auto-approve file edits, prompt for other tools |
| `default` | Prompt for all tool use in browser |

### Tool Approval

In `default` mode, a browser popup appears when Claude wants to use tools. Auto-approves after 30 seconds of no response.

### Code Block Copy

Hover over a code block to reveal a copy button. Shows checkmark for 2 seconds after copying.

### Export Conversation

Click the download button in the chat header to export as `.md` file.

### Model Selection

| Model | Characteristics |
|---|---|
| `claude-sonnet-4-6` | Default. Balance of speed and performance |
| `claude-opus-4-6` | Best performance, for complex tasks |
| `claude-haiku-4-5` | Fast responses, for simple tasks |

---

## Files

Browse and edit project files via the **Files** tab. Click a file in the tree to open in Monaco editor. Save with `Ctrl+S` / `Cmd+S`.

---

## Terminal

Access the server shell via the **Terminal** tab. Requires node-pty to be built successfully. Auto-resizes with the window.

---

## Team Tab

Create and run multi-agent teams. Select a team from the dropdown or create new with YAML templates. Agent status is displayed as a kanban board with real-time log streaming.

---

## Harness Badge

When a Harness Engineering workflow is running, a green badge appears in the header showing the active agent count. Polls every 5 seconds.

---

## Git Panel

| Feature | Description |
|---|---|
| Changes | Modified file list, staging, commit, Push/Pull |
| Log | Commit history and diff viewer |
| Branches | Branch list and checkout |
| Issues | GitHub Issues integration panel |
