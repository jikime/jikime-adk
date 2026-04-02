# Troubleshooting

---

## node-pty Build Failure (Linux)

### Symptom

Terminal tab is disabled with message: `node-pty load failed — terminal feature disabled`

### Fix

**1. Install build tools**

```bash
# Rocky Linux / RHEL
sudo dnf install -y python3 make gcc gcc-c++

# Ubuntu / Debian
sudo apt-get install -y python3 make gcc g++ build-essential
```

**2. Recompile node-pty**

```bash
pnpm fix-pty
# or: bash scripts/fix-pty-linux.sh
```

**3. Restart server**

Docker handles this automatically during image build.

---

## Claude CLI Not Found

### Symptom

`claude native binary not found` — chat returns `Claude Code executable not found`

### Fix

```bash
# Verify installation
which claude && claude --version

# Set path manually
CLAUDE_PATH=/usr/local/bin/claude jikime webchat start
# or add to ~/.jikime/webchat/.env
```

---

## Chat Error: exit code 1

### Causes & Solutions

1. **Auth not completed**: Run `claude auth login`
2. **Root environment**: Switch permission mode to `default` or `acceptEdits`
3. **API key expired**: Run `claude auth status`

---

## Remote Connection Failure

### Checklist

1. **Server running?** `curl http://localhost:4000/api/ws/health` should return `{"ok":true}`
2. **Firewall open?** Check with `firewall-cmd --list-ports` or `ufw status`
3. **Host format?** Enter as `IP:port` without `ws://` or `http://` prefix

---

## Project Path Displayed Incorrectly

Claude encodes `/` as `-` in project directory names. Directories with `-` in their name (e.g., `jikime-adk`) use a filesystem-based longest-match algorithm for correct path restoration.

---

## pnpm node-pty Build Error

Ensure `package.json` contains:

```json
"pnpm": {
  "allowedBuilds": ["node-pty"]
}
```

Then run `pnpm install` again.
