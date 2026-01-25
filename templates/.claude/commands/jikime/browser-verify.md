---
description: "Browser runtime error detection and auto-fix loop. Start dev server, capture console/page errors via Playwright, and fix iteratively."
argument-hint: "[--max N] [--port N] [--url URL] [--headed] [--routes path1,path2] [--skip-fix] [--e2e]"
context: debug
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
---

## Pre-execution Context

!cat package.json | grep -A 5 '"scripts"'

## Essential Files

@package.json

---

# /jikime:browser-verify - Browser Runtime Error Detection & Auto-Fix

## Core Principle: Catch What Static Analysis Misses

Detect runtime browser errors (undefined references, missing modules, DOM errors) that only appear when the application runs in a real browser environment.

```
START: Dev Server Launch
  |
Playwright: Navigate Routes
  |
Capture: Console Errors + Page Errors
  |
Analyze: Stack Traces -> Source Files
  |
Fix: Delegate to Agent
  |
Re-verify: Loop Until Clean
  |
(--e2e) E2E Functional Tests
  |
<jikime>DONE</jikime>
```

## Command Purpose

Autonomously detect and fix runtime browser errors:

1. **Package Manager Detection** (lock file scan)
2. **Dev Server Launch** (background, port detection)
3. **Playwright Navigation** (route discovery + page load)
4. **Error Capture** (console.error, uncaught exceptions, unhandled rejections)
5. **Stack Trace Analysis** (map errors to source files)
6. **Agent-Delegated Fix** (debugger/frontend subagent)
7. **Re-verification Loop** (until zero errors or max iterations)

Arguments: $ARGUMENTS

## Quick Start

```bash
# Default: detect package manager, start dev server, verify all routes
/jikime:browser-verify

# Specify port
/jikime:browser-verify --port 5173

# Specify routes to check
/jikime:browser-verify --routes /,/about,/dashboard

# Maximum iterations
/jikime:browser-verify --max 10

# Show browser window (debug)
/jikime:browser-verify --headed

# Only detect errors (no auto-fix)
/jikime:browser-verify --skip-fix

# Run E2E functional tests after error fixing
/jikime:browser-verify --e2e

# Specify dev server URL directly
/jikime:browser-verify --url http://localhost:3000
```

## Command Options

| Option | Alias | Description | Default |
|--------|-------|-------------|---------|
| `--max N` | --max-iterations | Maximum fix iterations | 10 |
| `--port N` | - | Dev server port | Auto-detect |
| `--url URL` | - | Dev server URL (skip server start) | - |
| `--headed` | - | Show browser window | false (headless) |
| `--routes paths` | - | Comma-separated routes to verify | Auto-discover |
| `--skip-fix` | --report-only | Only report errors, no auto-fix | false |
| `--e2e` | - | Run E2E functional tests after error fixing completes | false |
| `--timeout N` | - | Page load timeout (ms) | 30000 |
| `--stagnation-limit N` | - | Max iterations without improvement | 3 |

## Dev Server Auto Detection

**Pre-start Check**: Before launching a new dev server, automatically detect if one is already running.

### Detection Logic

```bash
# Scan common dev server ports
COMMON_PORTS=(3000 3001 5173 5174 8080 4200 8000 8888)

# Check each port for running server
for port in ${COMMON_PORTS[@]}; do
  if lsof -i :$port -sTCP:LISTEN >/dev/null 2>&1; then
    # Server found on this port
  fi
done
```

### Behavior Based on Detection

| Scenario | Action |
|----------|--------|
| **`--url` provided** | Skip detection, use provided URL |
| **1 server found** | Use automatically, skip server start |
| **Multiple servers found** | Ask user which one to use |
| **No servers found** | Proceed with normal server start |

### Example Interactions

**Single Server Detected:**
```
Detected dev server already running on http://localhost:3000
Skipping server start, using existing server.
```

**Multiple Servers Detected:**
```
Detected multiple dev servers:
1. http://localhost:3000
2. http://localhost:5173

Which server should be used for browser verification?
(Select one, or choose 'Start New' to launch a fresh dev server)
```

**No Server Detected:**
```
No running dev server detected.
Starting new server with: pnpm run dev
```

### Port Detection Priority

| Port | Common Framework |
|------|------------------|
| 3000 | Next.js, Create React App |
| 3001 | Next.js (alt) |
| 5173 | Vite |
| 5174 | Vite (alt) |
| 8080 | Vue CLI, generic |
| 4200 | Angular |
| 8000 | Django, generic |
| 8888 | Jupyter, generic |

## Package Manager Detection

Detect package manager from lock files in project root:

| Lock File | Package Manager | Dev Command |
|-----------|----------------|-------------|
| `pnpm-lock.yaml` | pnpm | `pnpm run dev` |
| `yarn.lock` | yarn | `yarn dev` |
| `package-lock.json` | npm | `npm run dev` |
| `bun.lockb` | bun | `bun run dev` |

Priority: pnpm > yarn > npm > bun (check in this order)

## Dev Server Management

### Server Start

```bash
# Background execution with output capture
# JIKIME_MANAGED=1 prefix bypasses pre-bash tmux check for managed dev servers
JIKIME_MANAGED=1 {package_manager} run dev &

# Alternative script names to try (in order):
# 1. "dev"
# 2. "start"
# 3. "serve"
```

### Port Detection Strategy

1. If `--port` specified: Use directly
2. If `--url` specified: Skip server start, use URL directly
3. Parse `package.json` scripts.dev for port hints (e.g., `--port 5173`, `-p 3000`)
4. Monitor stdout for URL pattern: `http://localhost:NNNN`
5. Fallback: Try common ports (3000, 5173, 8080, 4200, 8000)

### Server Ready Detection

Poll the detected URL until HTTP 200 response:

```
Attempt 1: GET http://localhost:{port} -> Connection refused (retry)
Attempt 2: GET http://localhost:{port} -> Connection refused (retry)
...
Attempt N: GET http://localhost:{port} -> 200 OK (ready!)
```

- Max wait: 60 seconds
- Poll interval: 2 seconds
- Fail after timeout with clear error message

### Server Cleanup

Always terminate dev server process on:
- Command completion (success or failure)
- User interruption
- Error that prevents continuation

## Route Discovery

### Strategy 1: From Arguments

If `--routes` is specified, use the provided list directly.

### Strategy 2: From Project Files

Scan for route definitions in common patterns:

```bash
# React Router
grep -r "path=" src/ --include="*.tsx" --include="*.jsx"
grep -r "Route " src/ --include="*.tsx" --include="*.jsx"

# Next.js App Router
find app/ -name "page.tsx" -o -name "page.jsx"

# Next.js Pages Router
find pages/ -name "*.tsx" -o -name "*.jsx" | grep -v "_app\|_document\|api/"

# Vue Router
grep -r "path:" src/router/ --include="*.ts" --include="*.js"

# Angular
grep -r "path:" src/app/ --include="*routing*"
```

### Strategy 3: Fallback

If no routes discovered, use: `["/"]` (root path only)

## Error Capture with Playwright

### Error Types to Capture

| Type | Source | Severity |
|------|--------|----------|
| `console.error` | `page.on('console', msg => msg.type() === 'error')` | HIGH |
| `uncaughtException` | `page.on('pageerror')` | CRITICAL |
| `unhandledRejection` | `page.on('pageerror')` for Promise rejections | CRITICAL |
| `network error` | `page.on('response', r => r.status() >= 400)` | MEDIUM |
| `resource load failure` | `page.on('requestfailed')` | MEDIUM |

### Navigation Flow

For each route:

```
1. Navigate to route: page.goto(baseUrl + route, { waitUntil: 'networkidle' })
2. Wait for hydration: page.waitForTimeout(2000)
3. Collect errors accumulated during load
4. Optional: Click interactive elements to trigger lazy errors
5. Record errors with route context
```

### Error Record Structure

```typescript
interface BrowserError {
  type: 'console_error' | 'page_error' | 'network_error' | 'resource_error'
  message: string
  route: string
  stack?: string           // Stack trace if available
  sourceFile?: string      // Extracted from stack trace
  sourceLine?: number      // Line number from stack trace
  timestamp: number
}
```

## Stack Trace Analysis

Extract source file and line from error stack traces:

```
TypeError: Cannot set properties of undefined (setting 'width')
    at Module.setup (http://localhost:5173/src/components/Canvas.tsx:45:12)
    at renderWithHooks (http://localhost:5173/node_modules/react-dom/...)
```

Extraction rules:
1. Skip `node_modules` frames
2. Match project source patterns: `/src/`, relative paths
3. Strip URL prefix to get relative file path
4. Extract line:column numbers

## Auto-Fix Loop

### Fix Delegation

[HARD] ALL fixes MUST be delegated to specialized agents:

```
Error Type → Agent Selection:
- Component/UI errors       → frontend subagent
- Module resolution         → debugger subagent
- Type/reference errors     → debugger subagent
- API/network errors        → backend subagent
- Build/bundler errors      → build-fixer subagent
```

### Loop Flow

```
Iteration 1:
  Errors Found: 5
  → Delegate fixes to agents
  → Re-navigate all routes

Iteration 2:
  Errors Found: 2 (3 fixed)
  → Delegate remaining fixes
  → Re-navigate all routes

Iteration 3:
  Errors Found: 0
  → SUCCESS: All runtime errors resolved
  → <jikime>DONE</jikime>
```

### Stagnation Detection

If error count doesn't decrease for N consecutive iterations (default: 3):
- Stop loop
- Report remaining errors with analysis
- Suggest manual intervention

## Output Format

### Running

```markdown
## Browser Verify: Iteration 2/10

### Dev Server
- Package Manager: pnpm
- Command: pnpm run dev
- URL: http://localhost:5173
- Status: Running

### Routes Scanned: 5/8
- [x] / (0 errors)
- [x] /about (0 errors)
- [x] /dashboard (2 errors)
- [ ] /settings ← scanning
- [ ] /profile

### Errors Found: 3
1. CRITICAL: TypeError at src/components/Canvas.tsx:45
   "Cannot set properties of undefined (setting 'width')"
2. HIGH: ReferenceError at src/hooks/useAuth.ts:23
   "authContext is not defined"
3. MEDIUM: 404 at /api/health (network error)

### TODO
1. [x] src/components/Canvas.tsx:45 - undefined property access
2. [in_progress] src/hooks/useAuth.ts:23 - missing context
3. [ ] /api/health - endpoint not found

Fixing...
```

### Complete

```markdown
## Browser Verify: COMPLETE

### Summary
- Iterations: 3
- Routes Verified: 8
- Errors Found: 5
- Errors Fixed: 5
- Remaining: 0

### Fixed Issues
1. src/components/Canvas.tsx:45 - Added null check for canvas ref
2. src/hooks/useAuth.ts:23 - Wrapped with AuthProvider
3. src/lib/chart.ts:12 - Fixed import path for chart library
4. src/pages/Dashboard.tsx:67 - Added loading state guard
5. src/utils/format.ts:8 - Fixed undefined object access

### E2E Results (--e2e)
- Total: 12 tests
- Passed: 12 (100%)
- Failed: 0
- Duration: 8.2s

### Dev Server
- Clean shutdown: Yes
- Final state: All routes load without errors

<jikime>DONE</jikime>
```

### Stagnation

```markdown
## Browser Verify: STAGNATION (no improvement in 3 iterations)

### Analysis
- Iteration 5: 2 errors
- Iteration 6: 2 errors (no change)
- Iteration 7: 2 errors (no change)

### Remaining Issues
1. src/vendor/legacy-lib.js:1203 - Third-party library internal error
   (Cannot fix without library update)
2. src/components/Map.tsx:89 - WebGL context not available in headless
   (Environment-specific, not a real bug)

### Recommendation
- Issue #1: Update library or find alternative
- Issue #2: Skip in headless mode, verify manually with --headed
```

## Related Commands

- `/jikime:loop` - General iterative fix loop (LSP, tests, coverage)
- `/jikime:e2e` - E2E test generation and execution
- `/jikime:build-fix` - Build error fixing
- `/jikime:migrate-4-verify` - Migration verification (includes browser checks)

---

## EXECUTION DIRECTIVE

1. Parse $ARGUMENTS (extract --max, --port, --url, --headed, --routes, --skip-fix, --e2e, --timeout, --stagnation-limit flags)

2. Detect package manager:
   - Check for lock files in order: pnpm-lock.yaml, yarn.lock, package-lock.json, bun.lockb
   - Read package.json to confirm scripts.dev exists
   - If no dev script found, try: start, serve (in order)
   - If none found: Error with "No dev script found in package.json"

3. IF --url flag specified: Skip to step 6 (use provided URL directly)

4. Dev Server Auto Detection (NEW):
   - Scan common ports for running servers: 3000, 3001, 5173, 5174, 8080, 4200, 8000, 8888
   - Use `lsof -i :PORT -sTCP:LISTEN` or `netstat` to detect listening processes
   - IF exactly 1 server found:
     - Log: "Detected dev server already running on http://localhost:{port}"
     - Use detected URL, skip to step 6
   - IF multiple servers found:
     - Use AskUserQuestion to let user select which server to use
     - Options: List detected servers + "Start New Server" option
     - IF user selects existing server: Use that URL, skip to step 6
     - IF user selects "Start New": Continue to step 5
   - IF no servers found: Continue to step 5

5. Start dev server:
   - Run detected command in background using Bash with run_in_background
   - [HARD] Prefix with `JIKIME_MANAGED=1` to bypass pre-bash tmux hook
   - Example: `JIKIME_MANAGED=1 pnpm run dev` or `JIKIME_MANAGED=1 npm run dev`

5. Wait for server ready:
   - IF --port specified: Use that port
   - ELSE: Parse package.json dev script for port hints, or monitor background task output for localhost URL
   - Poll with curl/wget until HTTP response received (max 60s, interval 2s)
   - Record the base URL (e.g., http://localhost:5173)

6. Discover routes:
   - IF --routes specified: Parse comma-separated list
   - ELSE: Scan project files for route definitions (React Router, Next.js, Vue Router, Angular)
   - Fallback to ["/"] if none found

7. Initialize Playwright via MCP:
   - Connect to browser (headless by default, --headed for visible)
   - Set up error listeners on page object:
     - page.on('console') for console.error messages
     - page.on('pageerror') for uncaught exceptions
     - page.on('requestfailed') for failed resource loads

8. Initialize iteration counter to 0

9. LOOP START (while iteration < max):

   9a. Clear error collection for this iteration

   9b. For each route:
       - Navigate: page.goto(baseUrl + route, { waitUntil: 'networkidle', timeout: --timeout })
       - Wait for hydration: page.waitForTimeout(2000)
       - Collect all captured errors with route context
       - Extract source file and line from stack traces

   9c. Aggregate errors:
       - Deduplicate by message + sourceFile + sourceLine
       - Sort by severity (CRITICAL > HIGH > MEDIUM)
       - Count unique errors

   9d. Check completion:
       - IF zero errors: Add completion marker and exit loop
       - Display current error summary

   9e. IF --skip-fix: Display error report and exit (no fixing)

   9f. Check stagnation:
       - Compare error count with previous N iterations (stagnation-limit)
       - IF no improvement: Exit with stagnation report

   9g. [HARD] Call TodoWrite tool to add discovered errors with pending status

   9h. [HARD] Before each fix, call TodoWrite to change item to in_progress

   9i. [HARD] AGENT DELEGATION MANDATE for Fix Execution:
       - ALL fix tasks MUST be delegated to specialized agents
       - NEVER execute fixes directly
       - Agent Selection by Error Type:
         - Component/rendering errors: Use frontend subagent
         - Module/import resolution: Use debugger subagent
         - Type/reference errors: Use debugger subagent
         - API/network errors: Use backend subagent
         - Build/bundler errors: Use build-fixer subagent

   9j. [HARD] After each fix completion, call TodoWrite to change item to completed

   9k. Increment iteration counter

10. LOOP END

11. IF --e2e flag AND errors resolved successfully:
    - Dev server is still running, reuse for E2E tests
    - Execute `/jikime:e2e --run` to run existing E2E tests
    - If no E2E tests exist, generate tests for discovered routes using `/jikime:e2e`
    - Report E2E results (pass/fail count)
    - If E2E failures found: delegate fixes to agents, re-run E2E (max 3 attempts)

12. Cleanup:
    - Close Playwright browser connection
    - Terminate dev server background process (kill by PID)

13. IF max iterations reached without completion: Display remaining issues and options

14. Report final summary with evidence (include E2E results if --e2e was used)

---

Version: 1.1.0
Last Updated: 2026-01-25
Core: Browser Runtime Error Detection & Auto-Fix Loop (with Dev Server Auto Detection)
