# Installation Guide

## Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| Node.js | 22+ | |
| pnpm | Latest | `corepack enable && corepack prepare pnpm@latest --activate` |
| Claude CLI | Latest | `npm install -g @anthropic-ai/claude-code` |
| JikiME-ADK | 1.8.0+ | Required for Method 1 (`go install` or `install.sh`) |
| Git | 2.x+ | Required for terminal Git panel |

---

## Method 1 — JikiME CLI (Recommended)

Starting from JikiME-ADK 1.8.0, you can install and manage webchat using the `jikime webchat` command.
Install location: `~/.jikime/webchat/`

### 1. Install

```bash
jikime webchat install
```

This command automatically performs:
1. Download webchat source from GitHub Release
2. Run `pnpm install --frozen-lockfile`
3. Run `pnpm build`

> **Note**: When installing JikiME-ADK via `install.sh`, webchat is installed automatically. Use the `--skip-webchat` flag to skip it.

### 2. Start

```bash
jikime webchat start                # Default port 4000
jikime webchat start --port 3000    # Custom port
```

Open `http://localhost:4000` in your browser.

### 3. Check Status

```bash
jikime webchat status
```

Displays install path, version, dependency status, build status, and Node.js/pnpm versions.

### 4. Rebuild

When rebuilding is needed after source changes or updates:

```bash
jikime webchat build
```

### 5. Update

Update to the latest version:

```bash
jikime webchat install                    # Install webchat matching current jikime version
jikime webchat install --version 1.8.0    # Install specific version
```

Existing `node_modules` and `.next` are preserved; only source files are replaced, followed by `pnpm install` + `pnpm build`.

---

## Method 2 — Local Direct Execution

### 1. Install Dependencies

```bash
pnpm install
```

On macOS, `spawn-helper` permissions are automatically granted after install.
For node-pty compilation on Linux (Rocky / RHEL), see [Troubleshooting — node-pty](./troubleshooting.md#node-pty-build-failure-linux)

### 2. Run Development Server

```bash
pnpm dev
```

Open `http://localhost:4000` in your browser.

### 3. Production Build and Run

```bash
pnpm build
NODE_ENV=production pnpm dev
```

### Environment Variables

Copy `.env.example` to `.env` and configure as needed.

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `PORT` | `4000` | Server port |
| `HOSTNAME` | `localhost` | Bind address (set to `0.0.0.0` for external access) |
| `CLAUDE_PATH` | Auto-detect | Claude CLI native binary path (specify manually if auto-detection fails) |

---

## Method 3 — Docker (Recommended for Server Deployment)

### Prerequisites

- Docker Engine 24+
- Docker Compose v2

### 1. Build Image and Start Container

```bash
docker compose up -d --build
```

Build process:
- Install Node.js 22 + build tools (python3, gcc, g++)
- `pnpm install` — including node-pty native compilation
- `next build` — Next.js production build
- Claude CLI installation (`npm install -g @anthropic-ai/claude-code`)

### 2. Claude CLI Authentication

After the container starts, authenticate directly inside the container.

```bash
# Enter container shell
docker exec -it webchat bash

# Authenticate Claude CLI
claude auth login

# Verify authentication
claude --version

# Exit shell
exit
```

Authentication data is stored in the `claude_data` volume (`/root/.claude`) and **persists across container restarts**.

### 3. Access

```
http://<server-ip>:4000
```

### Key Commands

```bash
# View logs
docker compose logs -f

# Restart container
docker compose restart

# Stop (preserve volumes)
docker compose down

# Stop + delete volumes (full reset including auth)
docker compose down -v

# Rebuild image
docker compose up -d --build --force-recreate
```

### Volumes

| Volume Name | Container Path | Contents |
|---|---|---|
| `claude_data` | `/root/.claude` | Claude auth credentials, session history, project list |

### Mounting Project Directories

To access host project directories from inside the container, modify the volumes section in `docker-compose.yml`.

```yaml
volumes:
  - claude_data:/root/.claude
  - /home:/home          # Mount host /home at same path in container
  - /root:/root/projects # Mount host /root to container /root/projects
```

> **Note**: Changing mount paths affects how Claude recognizes project paths. Mount at the same absolute path as the host whenever possible.
