## Phase 1: Dev Server Lifecycle Management

### Overview

Before running any Playwright tests, the command MUST ensure both source and target applications are running and accessible.

### Step 1.1: Read Configuration

```
Read .migrate-config.yaml to extract:
- source_dir: Original project path
- output_dir: Migrated project path
- source_framework: Source framework name
- target_framework: Target framework name
- verification.dev_command: Override dev command (optional)
- verification.source_port: Source port (default: 3000)
- verification.target_port: Target port (default: 3001)
```

### Step 1.2: Framework Dev Command Detection

Auto-detect the dev command based on framework and package.json:

**Detection Priority:**
1. `verification.dev_command` in `.migrate-config.yaml` (explicit override)
2. `package.json` scripts → `dev` or `start` script
3. Framework-based fallback mapping

**Framework Command Mapping:**

| Framework | Dev Command | Port Flag |
|-----------|------------|-----------|
| Next.js | `npx next dev` | `--port {port}` |
| Vite (React/Vue/Svelte) | `npx vite` | `--port {port}` |
| Create React App | `npx react-scripts start` | `PORT={port}` (env) |
| Nuxt | `npx nuxt dev` | `--port {port}` |
| Angular | `npx ng serve` | `--port {port}` |
| Remix | `npx remix dev` | `--port {port}` |
| Gatsby | `npx gatsby develop` | `-p {port}` |
| Astro | `npx astro dev` | `--port {port}` |
| SvelteKit | `npx vite dev` | `--port {port}` |
| Express/Fastify | `node server.js` | `PORT={port}` (env) |
| Django | `python manage.py runserver` | `0.0.0.0:{port}` |
| Flask | `flask run` | `--port {port}` |
| Spring Boot | `./mvnw spring-boot:run` | `-Dserver.port={port}` |
| Rails | `rails server` | `-p {port}` |
| Laravel | `php artisan serve` | `--port={port}` |
| Go (Fiber/Gin) | `go run .` | `PORT={port}` (env) |

**Detection Algorithm:**

```
1. IF verification.dev_command exists in config:
     → Use it directly (user override)

2. ELIF package.json exists:
     → Read scripts.dev or scripts.start
     → Extract base command
     → Append port flag based on framework detection

3. ELIF requirements.txt / manage.py exists:
     → Use Django/Flask command

4. ELIF go.mod exists:
     → Use Go run command

5. ELIF pom.xml / build.gradle exists:
     → Use Spring Boot / Gradle command

6. ELSE:
     → Ask user via AskUserQuestion for dev command
```

### Step 1.3: Install Dependencies

Before starting servers, ensure dependencies are installed:

```
For each project (source_dir, output_dir):
  IF package.json exists AND node_modules/ missing:
    → Run: npm install (or pnpm install / yarn install based on lockfile)
  ELIF requirements.txt exists AND venv/ missing:
    → Run: pip install -r requirements.txt
  ELIF go.mod exists:
    → Run: go mod download
```

**Lockfile Detection:**
- `pnpm-lock.yaml` → `pnpm install`
- `yarn.lock` → `yarn install`
- `package-lock.json` → `npm install`
- `bun.lockb` → `bun install`

### Step 1.4: Start Dev Servers

**Startup Sequence:**

```
Step 1: Start SOURCE server (if --source-url not provided)
  → cd {source_dir}
  → Run dev command with --source-port (background process)
  → Record PID for cleanup

Step 2: Start TARGET server (if --target-url not provided)
  → cd {output_dir}
  → Run dev command with --port (background process)
  → Record PID for cleanup

Step 3: Health check both servers
  → Wait for both to be accessible
  → Abort if either fails
```

**Background Process Management:**

```bash
# Start server in background, capture PID
{dev_command} > /tmp/jikime-verify-{role}.log 2>&1 &
SERVER_PID=$!

# Store PIDs for cleanup
echo $SERVER_PID >> /tmp/jikime-verify-pids.txt
```

### Step 1.5: Health Check

**Health Check Algorithm:**

```
FUNCTION healthCheck(url, maxWaitSeconds=30, intervalMs=1000):
  startTime = now()

  WHILE (now() - startTime) < maxWaitSeconds:
    TRY:
      response = HTTP_GET(url)
      IF response.status < 500:
        RETURN SUCCESS
    CATCH:
      WAIT(intervalMs)

  RETURN TIMEOUT_ERROR
```

**Health Check Targets:**
- Source: `http://localhost:{source_port}` (or `--source-url`)
- Target: `http://localhost:{port}` (or `--target-url`)

**Failure Handling:**
- If source fails to start: Warn user, continue with target-only verification
- If target fails to start: ABORT with error (target is required)
- Log server output to `/tmp/jikime-verify-{role}.log` for debugging

### Step 1.6: Server Cleanup (Post-Verification)

**Cleanup is MANDATORY after all tests complete (success or failure):**

```
FINALLY:
  Read PIDs from /tmp/jikime-verify-pids.txt
  FOR each PID:
    kill -TERM {PID}
    Wait 5s for graceful shutdown
    IF still running: kill -9 {PID}
  Remove /tmp/jikime-verify-*.log
  Remove /tmp/jikime-verify-pids.txt
```

### Step 1.7: URL Resolution

After servers are running, resolve final URLs:

```
source_url = --source-url OR "http://localhost:{source_port}"
target_url = --target-url OR "http://localhost:{port}"
```

These URLs are passed to all subsequent verification phases.

