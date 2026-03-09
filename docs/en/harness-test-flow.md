# Harness Engineering Test Flow

Step-by-step guide for applying Harness Engineering to a new project for the first time.

---

## Prerequisites

- `gh` CLI installed and authenticated (`gh auth status`)
- Go 1.21+ installed (for source build)
- GitHub account with permission to create public repositories

---

## Step 0: Install the Latest jikime

```bash
cd /path/to/jikime-adk

# Build and install from latest source
go install .

# Verify installation — should show the init subcommand
jikime serve --help
```

**Check**: `Available Commands: init` must appear in the output.

---

## Step 1: Create a GitHub Remote Repository

```bash
cd /path/to/harness-test

# Create a public repo on GitHub
gh repo create <owner>/harness-test --public --description "Harness Engineering Test"

# Initialize local git, create first commit, connect to remote
git init
echo "# harness-test" > README.md
git add .
git commit -m "chore: initial commit"
git branch -M main
git remote add origin https://github.com/<owner>/harness-test.git
git push -u origin main
```

---

## Step 2: Generate WORKFLOW.md (`jikime serve init`)

```bash
cd /path/to/harness-test

jikime serve init
```

Interactive wizard inputs:

| Prompt | Value | Notes |
|--------|-------|-------|
| GitHub repo slug | `<owner>/harness-test` | Auto-detected from git remote |
| Active label | `jikime-todo` | Press Enter (default) |
| Workspace root | `/tmp/jikime-harness-test` | Press Enter (default) |
| HTTP status API port | `8888` | Press Enter (default) |
| Max concurrent agents | `1` | Press Enter (default) |

A `WORKFLOW.md` will be created in the current directory.

> **JiKiME-ADK Mode**: If the project has a `.claude/` directory, the wizard
> automatically generates a JiKiME-ADK mode config (uses the jarvis agent).

---

## Step 3: Create GitHub Labels

Run the commands shown in the `serve init` success output:

```bash
# Active label — agent processes issues with this label
gh label create "jikime-todo" \
  --repo <owner>/harness-test \
  --description "Ready for AI agent" \
  --color "0e8a16"

# Done label — automatically applied by the agent on completion
gh label create "jikime-done" \
  --repo <owner>/harness-test \
  --description "Completed by AI agent" \
  --color "6f42c1"
```

---

## Step 4: Create a Test GitHub Issue

```bash
gh issue create \
  --repo <owner>/harness-test \
  --title "Build a simple app with Next.js 16, Tailwind CSS 4, and shadcn/ui" \
  --body "## Requirements

Implement a simple app using Next.js 16 App Router, Tailwind CSS 4, and shadcn/ui.

### Features
- Main page: hero section + card list
- Card component: title, description, badge, and button
- Dark mode toggle (shadcn/ui ThemeProvider)
- Responsive layout (mobile/desktop)

### Tech Stack
- Next.js 16 (App Router)
- TypeScript
- Tailwind CSS 4
- shadcn/ui (Card, Button, Badge, Switch components)

### Setup Commands
\`\`\`bash
npx create-next-app@latest . --typescript --tailwind --app --yes
npx shadcn@latest init -y
npx shadcn@latest add card button badge switch
\`\`\`

### File Structure
\`\`\`
app/
  page.tsx              # Main page (hero + card list)
  layout.tsx            # Root layout (with ThemeProvider)
  globals.css           # Tailwind CSS 4 config
components/
  theme-provider.tsx    # Dark mode provider
  theme-toggle.tsx      # Dark mode toggle button
  feature-card.tsx      # Reusable card component
\`\`\`

### Acceptance Criteria
- \`npm run build\` passes successfully
- shadcn/ui components render correctly
- Dark mode toggle works
- No TypeScript errors" \
  --label "jikime-todo"
```

Once the `jikime-todo` label is applied to an issue, `jikime serve` will detect it automatically.

---

## Step 5: Start jikime serve

```bash
cd /path/to/harness-test

jikime serve WORKFLOW.md
```

On normal operation, logs appear in this order:

```
[poller]    found issue #1 "Build a simple app with Next.js 16, Tailwind CSS 4, and shadcn/ui" [jikime-todo]
[workspace] creating /tmp/jikime-harness-test/issue-1
[hook]      after_create: git clone https://github.com/<owner>/harness-test.git .
[hook]      before_run: syncing to latest main...
[agent]     starting claude on issue #1 (attempt 1)
...
[agent]     done — created PR #1
[poller]    issue #1 state changed → jikime-done
```

---

## Step 6: Monitor Status (Separate Terminal)

While the service is running, check status from another terminal:

```bash
# Text dashboard (human-readable)
curl http://localhost:8888/

# JSON state API
curl -s http://localhost:8888/api/v1/state | jq .

# Running agents only
curl -s http://localhost:8888/api/v1/state | jq '.running'

# Trigger immediate poll (detect issues faster)
curl -s -X POST http://localhost:8888/api/v1/refresh | jq .
```

---

## Step 7: Verify Results

```bash
# List created PRs
gh pr list --repo <owner>/harness-test

# View PR details
gh pr view 1 --repo <owner>/harness-test

# View file changes
gh pr diff 1 --repo <owner>/harness-test
```

---

## Complete Flow Summary

```
go install .
  ↓
cd harness-test
gh repo create + git push
  ↓
jikime serve init
  → WORKFLOW.md created
  ↓
gh label create jikime-todo
gh label create jikime-done
  ↓
gh issue create --label jikime-todo
  ↓
jikime serve WORKFLOW.md
  ↓
[other terminal] curl localhost:8888/status
  ↓
gh pr list → verify results
```

---

## Common Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| `serve init` command not found | Outdated binary | Re-run `go install .` |
| Issue not detected | Label name mismatch | Ensure label matches `active_states` in WORKFLOW.md |
| Clone fails | GitHub auth issue | Check `gh auth status` |
| Agent stalls | `stall_timeout_ms` exceeded | Increase `claude.stall_timeout_ms` in WORKFLOW.md |
| Port conflict | 8888 already in use | Use `--port 9999` or set `server.port: 0` (disabled) |

---

## Related Docs

- [Harness Engineering Overview](./harness-engineering.md)
- [WORKFLOW.md Reference](./harness-engineering.md#workflowmd-configuration-reference)
- [SPEC.md](../../symphony/SPEC.md) (Symphony original specification)
