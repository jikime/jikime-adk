# Comprehensive Code Review Report

## Summary
- **Verdict**: WARN
- **Files Reviewed**: ~60 files (Go CLI core, webchat API routes, internal packages, templates)
- **Total Findings**: 21 (Critical: 1, High: 4, Medium: 10, Low: 6)

---

## Architecture Review

### MEDIUM-01: Large files exceeding 800-line guideline
- **Files**: `statusline.go` (853 lines), `board.go` (801 lines), `update.go` (748 lines)
- **Impact**: Reduced readability and maintainability
- **Recommendation**:
  - `board.go`: Extract `boardHTML` (240+ lines of inline HTML/CSS/JS React SPA starting at line 563) into a separate embedded file. Extract `collectBoardData()` and `boardServeCmd` into separate files (e.g., `board_data.go`, `board_serve.go`).
  - `statusline.go`: Split rate-limit fetching, session parsing, and UI rendering into separate files.
  - `update.go`: Extract binary update logic (`performBinaryUpdate`, `updateSingleBinary`) into `binary_update.go`.

### MEDIUM-02: Inline HTML SPA in Go source
- **File**: `cmd/teamcmd/board.go:563-801`
- **Impact**: 240 lines of raw HTML/CSS/JS embedded as a Go string constant. Difficult to maintain, no syntax highlighting, no linting.
- **Recommendation**: Use `go:embed` to load the HTML from a separate file (e.g., `board_spa.html`).

### MEDIUM-03: No test files detected
- **Impact**: Zero test coverage across the entire Go codebase and webchat TypeScript. No `*_test.go`, `*.test.ts`, or `*.spec.ts` files found.
- **Risk**: Regressions go undetected. Refactoring is risky without safety net.
- **Recommendation**: Prioritize tests for critical paths: `internal/team/spawner.go` (agent spawning), `cmd/hookscmd/pre_tool_security.go` (security hooks), `cmd/updatecmd/update.go` (self-update), and webchat API routes.

### LOW-01: Inconsistent NAME_RE maximum length across webchat routes
- Some routes use `{1,80}` (launch, create) while others use `{1,64}` (tasks, inbox, events, run).
- **Recommendation**: Define a single shared constant (e.g., `MAX_NAME_LEN = 64`) and import across all routes.

---

## Security Review

### CRITICAL-01: Terminal route auto-executes `claude --dangerously-skip-permissions`
- **File**: `webchat/src/app/api/terminal/route.ts:274`
- **Code**: `proc.stdin.write('claude --dangerously-skip-permissions\r')`
- **Impact**: Every new terminal session automatically launches Claude with ALL permission checks disabled. Combined with the webchat being network-accessible, any user who can reach the webchat API can execute arbitrary commands on the host machine via Claude.
- **Mitigation factors**: The terminal route does have session limits (50 max), idle timeouts (1hr), and cwd blocking for sensitive directories.
- **Recommendation**:
  1. Do NOT auto-inject `--dangerously-skip-permissions` by default. Make it an explicit opt-in via environment variable or config flag.
  2. Add authentication to the webchat API routes (currently none exist).
  3. At minimum, bind the server to `127.0.0.1` only and verify this is enforced.

### HIGH-01: No authentication on webchat API routes
- **Impact**: All webchat API routes (team management, terminal sessions, chat, template generation) have zero authentication. Anyone with network access to the server can:
  - Create/stop teams
  - Execute Claude commands
  - Create terminal sessions
  - Send messages to agents
- **Recommendation**: Add at minimum a bearer token or session-based authentication middleware. For local-only use, enforce `127.0.0.1` binding and add CORS restrictions (see HIGH-03).

### HIGH-02: `inbox watch --exec` command injection risk (contained)
- **File**: `cmd/teamcmd/inbox.go:207-220`
- **Code**: `exec.Command("sh", "-c", execCmd)` where `execCmd` comes from `--exec` flag, and message content is passed via environment variables.
- **Assessment**: The `execCmd` itself is user-provided via CLI flag (not message content), so this is a self-inflicted risk similar to `bash -c`. Message data is passed through environment variables (JIKIME_MSG_BODY, etc.), not interpolated into the shell command.
- **Risk**: If a user writes `--exec 'echo $JIKIME_MSG_BODY'`, the body content would be expanded by the shell. A malicious message body containing `$(rm -rf /)` or backticks could execute arbitrary code.
- **Recommendation**: Document this risk clearly. Consider offering a structured callback mechanism (e.g., `--exec-json` that pipes JSON to stdin of the command instead of using environment variables expanded by shell).

### HIGH-03: Wildcard CORS `Access-Control-Allow-Origin: *`
- **Files**: `board.go:386,439`, `webchat/team.ts:50`, `webchat/server.ts:1144`, `webchat/harness.ts:101`
- **Impact**: Any website can make requests to the jikime board serve endpoint and webchat server. Combined with no authentication, a malicious website could control teams if the user visits it while jikime is running.
- **Recommendation**: Restrict to `http://localhost:*` or specific configured origins.

### HIGH-04: Unbounded `io.ReadAll` on HTTP response
- **File**: `cmd/routercmd/test.go:83`
- **Code**: `respBody, _ := io.ReadAll(resp.Body)` -- no size limit, error discarded.
- **Impact**: A malicious or misconfigured upstream provider could return a multi-GB response, causing OOM.
- **Recommendation**: Use `io.LimitReader` (as already done in `update.go:250`). Check the error.

### MEDIUM-04: Chat route message not length-limited
- **File**: `webchat/src/app/api/chat/route.ts:19`
- **Code**: `const message = (body.message as string)?.trim()` -- no `.slice()` or length check.
- **Impact**: An attacker could send a very large message payload, causing resource exhaustion when passed to `claude -p`.
- **Recommendation**: Add `message.slice(0, MAX_LEN)` with a reasonable limit (e.g., 100KB).

### MEDIUM-05: `http.ListenAndServe` without timeouts
- **File**: `cmd/teamcmd/board.go:427`
- **Impact**: Default Go HTTP server has no read/write timeouts, making it vulnerable to slowloris attacks.
- **Recommendation**: Use `http.Server{ReadTimeout, WriteTimeout, IdleTimeout}` instead.

---

## Performance Review

### MEDIUM-06: Synchronous `execFile` calls in webchat API routes
- **Files**: Multiple routes in `webchat/src/app/api/team/`
- **Impact**: Each API call spawns a new `jikime` process. Under load, this could exhaust system resources (process limits, file descriptors).
- **Recommendation**: Consider a persistent connection to jikime (e.g., Unix socket or HTTP API) instead of spawning a new process per request for frequently-called endpoints.

### MEDIUM-07: Regex compilation in Go `init()` -- acceptable but rigid
- **File**: `cmd/hookscmd/pre_tool_security.go:111-115`
- **Assessment**: Patterns are compiled once at startup, which is efficient. However, they are hardcoded and cannot be customized per-project.
- **Recommendation**: Allow supplementary patterns via project config (e.g., `.jikime/config/security.yaml`).

### LOW-02: Terminal session cleanup interval
- **File**: `webchat/src/app/api/terminal/route.ts:174`
- The cleanup interval (2 minutes) and session limit (50) are reasonable. However, the `setInterval` registration using `globalThis` could create multiple intervals in development with hot-reload.
- **Recommendation**: Add cleanup on interval creation: clear existing interval before setting new one.

### LOW-03: `board.go` `collectBoardData` creates multiple store instances per call
- Each SSE push re-creates `team.NewRegistry`, `team.NewStore`, `team.NewTeamInbox`.
- **Recommendation**: Cache store instances per team for the duration of a serve session.

---

## Style Review

### MEDIUM-08: Unchecked errors in PID file operations
- **File**: `cmd/routercmd/start.go:137-138`
- **Code**:
  ```go
  os.MkdirAll(filepath.Dir(pidPath), 0o755)
  os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0o644)
  ```
- Both return values are silently ignored. If PID file write fails, the daemon appears to start but the PID tracking is broken, leading to orphan processes.
- **Recommendation**: Check and return or log errors.

### MEDIUM-09: Unchecked `json.Unmarshal` in `routercmd/test.go:87`
- **Code**: `json.Unmarshal(respBody, &anthropicResp)` -- error discarded
- **Impact**: If the response is malformed, the code proceeds with zero-value struct, potentially printing misleading "Success" with empty content.
- **Recommendation**: Check the error and report it.

### MEDIUM-10: Multiple discarded errors in `board.go`
- Lines ~482-508: `data, _ := os.ReadFile(...)`, `_ = json.Unmarshal(data, &cfg)`, `reg, _ := team.NewRegistry(...)`, `agents, _ := reg.List()`, `store, _ := team.NewStore(...)`
- **Impact**: Silent failures make debugging difficult when team data is missing or corrupted.
- **Recommendation**: At minimum, log warnings for non-trivial errors.

### LOW-04: `fmt.Sscanf` return values unchecked in version comparison
- **File**: `cmd/updatecmd/update.go:277-280`
- **Code**: `fmt.Sscanf(v1Parts[i], "%d", &n1)` -- if parsing fails, n1 remains 0, which could produce incorrect version comparisons for pre-release tags like `1.0.0-rc1`.
- **Recommendation**: Handle parse failures explicitly or use `strconv.Atoi`.

### LOW-05: Emoji usage in Go CLI output
- Multiple files use emoji (e.g., launch.go, board.go) in terminal output. While modern terminals handle this, some CI environments or piped output may not render correctly.
- **Recommendation**: Use an option flag or detect terminal capability.

### LOW-06: Duplicated `NAME_RE` and `BUDGET_RE` constants
- Webchat routes define `NAME_RE` locally in 12+ files independently.
- **Recommendation**: Create a shared validation module (e.g., `@/lib/validation.ts`) and import.

---

## Recommendations (Priority-Ordered)

| Priority | ID | Action |
|----------|----|--------|
| 1 | CRITICAL-01 | Remove auto `--dangerously-skip-permissions` from terminal route; make opt-in |
| 2 | HIGH-01 | Add authentication middleware to webchat API |
| 3 | HIGH-03 | Replace wildcard CORS with localhost-only or configured origins |
| 4 | HIGH-02 | Document `inbox watch --exec` security implications; consider safer callback |
| 5 | HIGH-04 | Add `io.LimitReader` to `routercmd/test.go` response reading |
| 6 | MEDIUM-03 | Add tests for critical paths (security hooks, spawner, update) |
| 7 | MEDIUM-04 | Add message length limit in chat route |
| 8 | MEDIUM-05 | Add HTTP server timeouts to board serve |
| 9 | MEDIUM-08/09/10 | Fix unchecked errors in routercmd and board.go |
| 10 | MEDIUM-01/02 | Decompose large files; extract inline HTML |
| 11 | LOW-01/06 | Centralize shared constants (NAME_RE, validation) |
| 12 | LOW-02/03/04/05 | Minor quality improvements |

---

## Positive Observations

The codebase demonstrates several strong practices:

1. **Input validation**: Webchat routes consistently use `NAME_RE` regex validation, `content-length` checks, and `slice()` truncation on user inputs.
2. **execFile over exec**: Using `execFile` instead of `exec` (shell mode) prevents most command injection in webchat routes.
3. **Security hooks**: `pre_tool_security.go` provides a well-structured deny/ask/allow pattern for file access control with compiled regex patterns.
4. **Path traversal prevention**: Terminal route blocks sensitive directories (`/etc`, `/root`, `/sys`, etc.) and validates cwd is an existing directory.
5. **Atomic binary updates**: `update.go` implements proper backup-and-rollback with SHA256 checksum verification.
6. **Rate limiting awareness**: GitHub API calls include token support to avoid rate limiting.
7. **Resource limits**: Terminal sessions have max count (50), idle timeout (1hr), max age (4hr), and buffer limits (50KB).
8. **shellQuote function**: `spawner.go:244` properly handles single-quote escaping for shell arguments.

---

*Report generated: 2026-04-06*
*Reviewer: J.A.R.V.I.S. Code Review System*
