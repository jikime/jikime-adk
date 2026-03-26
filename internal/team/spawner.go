package team

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// waitForClaudeReady polls the tmux pane until Claude's interactive input
// prompt (❯) appears, or until the timeout is exceeded.
// Returns nil when ready, error on timeout.
func waitForClaudeReady(target string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	// Markers that indicate Claude's interactive UI is fully loaded.
	// "❯" is the primary input cursor; "bypass permissions" appears in the
	// status bar once Claude is interactive.
	readyMarkers := []string{"❯", "bypass permissions", "esc to interrupt"}

	for time.Now().Before(deadline) {
		// 개별 tmux 명령 5초 타임아웃 — 폴링 루프에서 단일 명령 무한 대기 방지
		paneCtx, paneCancel := context.WithTimeout(context.Background(), 5*time.Second)
		out, err := exec.CommandContext(paneCtx, "tmux", "capture-pane", "-p", "-t", target).Output()
		paneCancel()
		if err == nil {
			content := string(out)
			for _, marker := range readyMarkers {
				if strings.Contains(content, marker) {
					return nil
				}
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for Claude to be ready in %s", target)
}

// SpawnBackend selects how agent processes are launched.
type SpawnBackend string

const (
	SpawnBackendTmux       SpawnBackend = "tmux"
	SpawnBackendSubprocess SpawnBackend = "subprocess"
)

// SpawnConfig holds all parameters needed to launch an agent.
type SpawnConfig struct {
	// TeamName is the team this agent belongs to.
	TeamName string

	// AgentID is the unique identifier for the new agent.
	// Auto-generated if empty.
	AgentID string

	// Role is the agent's function: "leader", "worker", "reviewer", …
	Role string

	// WorktreePath is the git worktree directory for this agent.
	// Empty means use the current working directory.
	WorktreePath string

	// InitialPrompt is injected as the agent's first user message.
	InitialPrompt string

	// Backend selects the spawn mechanism.
	Backend SpawnBackend

	// DataDir is the root ~/.jikime directory.
	DataDir string

	// ExtraEnv is additional environment variables to pass to the agent.
	ExtraEnv map[string]string

	// SkipPermissions passes --dangerously-skip-permissions to Claude CLI.
	SkipPermissions bool
}

// SpawnResult is returned by a successful Spawn call.
type SpawnResult struct {
	AgentID     string
	TmuxSession string // non-empty for tmux backend
	PID         int    // non-zero for subprocess backend
}

// Spawner launches agent processes with team identity baked in.
type Spawner struct{}

// NewSpawner returns a Spawner.
func NewSpawner() *Spawner { return &Spawner{} }

// Spawn starts a new agent process according to cfg.
func (s *Spawner) Spawn(cfg SpawnConfig) (*SpawnResult, error) {
	if cfg.AgentID == "" {
		cfg.AgentID = "agent-" + uuid.New().String()[:8]
	}
	switch cfg.Backend {
	case SpawnBackendTmux, "":
		return s.spawnTmux(cfg)
	case SpawnBackendSubprocess:
		return s.spawnSubprocess(cfg)
	default:
		return nil, fmt.Errorf("team/spawner: unknown backend %q", cfg.Backend)
	}
}

// Kill terminates an agent based on its SpawnResult.
func (s *Spawner) Kill(res *SpawnResult) error {
	if res.TmuxSession != "" {
		return exec.Command("tmux", "kill-session", "-t", res.TmuxSession).Run()
	}
	if res.PID > 0 {
		p, err := os.FindProcess(res.PID)
		if err != nil {
			return err
		}
		return p.Kill()
	}
	return nil
}

// --- tmux backend ---
//
// ClawTeam-style interactive spawning:
//  1. Start claude in interactive mode (no -p flag) inside a new tmux session.
//  2. Wait for claude to finish its startup banner.
//  3. Load the prompt into a named tmux buffer.
//  4. Paste the buffer into the running claude session (simulates typing).
//  5. Send Enter to submit.
//
// This keeps the tmux window alive so you can attach and watch the agent work.

func (s *Spawner) spawnTmux(cfg SpawnConfig) (*SpawnResult, error) {
	session := tmuxSessionName(cfg.TeamName, cfg.AgentID)
	// Window name = agent ID (e.g. "worker-1") so target is session:agent
	windowName := sanitize(cfg.AgentID)
	target := session + ":" + windowName
	env := buildEnv(cfg)

	// Write prompt to file (avoids shell escaping issues when pasting).
	promptFile, err := writePromptFile(cfg)
	if err != nil {
		return nil, fmt.Errorf("team/spawner: write prompt: %w", err)
	}

	// Build interactive claude command — no -p flag so the session stays alive.
	// ClawTeam key insight: unset CLAUDECODE* vars so a nested claude (spawned
	// from inside a Claude Code session) doesn't refuse to start.
	claudeCmd := "claude"
	if cfg.SkipPermissions {
		claudeCmd = "claude --dangerously-skip-permissions"
	}

	// Export env vars so they persist across the whole shell session.
	// This ensures the on-exit hook can read JIKIME_* vars after claude exits.
	var exportParts []string
	for k, v := range env {
		exportParts = append(exportParts, "export "+k+"="+shellQuote(v))
	}
	sort.Strings(exportParts) // deterministic ordering
	exportStr := strings.Join(exportParts, "; ")

	// On-exit hook: release tasks and mark agent offline when claude exits.
	// Silently ignored if jikime is not in PATH or not in team context.
	onExitHook := "jikime team lifecycle on-exit 2>/dev/null; true"

	// Full shell command:
	//   unset CLAUDECODE* (allow nested claude) → export env vars → cd worktree → claude → on-exit → exec bash
	// exec bash at the end keeps the tmux pane alive after claude exits so the
	// session remains inspectable (no [exited] tombstone).
	unset := "unset CLAUDECODE CLAUDE_CODE_ENTRYPOINT CLAUDE_CODE_SESSION 2>/dev/null"
	var fullCmd string
	if cfg.WorktreePath != "" {
		fullCmd = fmt.Sprintf("%s; %s; cd %s && %s; %s; exec bash",
			unset, exportStr, shellQuote(cfg.WorktreePath), claudeCmd, onExitHook)
	} else {
		fullCmd = fmt.Sprintf("%s; %s; %s; %s; exec bash", unset, exportStr, claudeCmd, onExitHook)
	}

	// Step 1: Create a new detached tmux session with the named window running claude.
	// -c sets the starting directory at the tmux level so Claude's cwd is correct
	// even before the shell command executes — this ensures ~/.claude/projects/ sessions
	// are saved under the correct project path.
	startDir := cfg.WorktreePath
	if startDir == "" {
		startDir, _ = os.Getwd()
	}
	// tmux 명령 공통 10초 타임아웃 — 명령 무한 블로킹 방지
	tmuxCtx, tmuxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer tmuxCancel()

	args := []string{"new-session", "-d", "-s", session, "-n", windowName, "-c", startDir, fullCmd}
	cmd := exec.CommandContext(tmuxCtx, "tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("team/spawner: tmux new-session %s: %w\n%s", cfg.AgentID, err, out)
	}

	// Step 2: Poll until Claude's interactive prompt (❯) appears.
	// Timeout is generous (60s) to handle slow machines or large Claude updates.
	if err := waitForClaudeReady(target, 60*time.Second); err != nil {
		// Non-fatal: log and continue — paste may still succeed.
		fmt.Printf("  ⚠️  %s: %v — attempting paste anyway\n", cfg.AgentID, err)
	}
	// Brief pause after detecting the prompt to let the UI settle.
	time.Sleep(300 * time.Millisecond)

	// Step 3: Load the prompt file into a named tmux buffer.
	bufName := "jikime-" + sanitize(cfg.AgentID)
	if out, err := exec.CommandContext(tmuxCtx, "tmux", "load-buffer", "-b", bufName, promptFile).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("team/spawner: tmux load-buffer %s: %w\n%s", cfg.AgentID, err, out)
	}

	// Step 4: Paste the buffer into the window (simulates typing the full prompt).
	if out, err := exec.CommandContext(tmuxCtx, "tmux", "paste-buffer", "-b", bufName, "-t", target).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("team/spawner: tmux paste-buffer %s: %w\n%s", cfg.AgentID, err, out)
	}

	// Step 5: Send Enter twice — claude interactive mode needs two Enters after paste.
	//   First Enter: confirms the pasted multi-line text.
	//   Second Enter: submits the message.
	time.Sleep(500 * time.Millisecond)
	if out, err := exec.CommandContext(tmuxCtx, "tmux", "send-keys", "-t", target, "Enter").CombinedOutput(); err != nil {
		return nil, fmt.Errorf("team/spawner: tmux send-keys Enter1 %s: %w\n%s", cfg.AgentID, err, out)
	}
	time.Sleep(300 * time.Millisecond)
	if out, err := exec.CommandContext(tmuxCtx, "tmux", "send-keys", "-t", target, "Enter").CombinedOutput(); err != nil {
		return nil, fmt.Errorf("team/spawner: tmux send-keys Enter2 %s: %w\n%s", cfg.AgentID, err, out)
	}

	// Clean up the named buffer.
	_ = exec.CommandContext(tmuxCtx, "tmux", "delete-buffer", "-b", bufName).Run()

	return &SpawnResult{
		AgentID:     cfg.AgentID,
		TmuxSession: session,
	}, nil
}

// shellQuote wraps a string in single quotes for safe shell embedding.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// --- subprocess backend ---
//
// Subprocess backend keeps the -p (print/non-interactive) approach because
// there is no visual display to watch. Output is piped to a log file.

func (s *Spawner) spawnSubprocess(cfg SpawnConfig) (*SpawnResult, error) {
	env := buildEnv(cfg)

	promptFile, err := writePromptFile(cfg)
	if err != nil {
		return nil, fmt.Errorf("team/spawner: write prompt: %w", err)
	}

	// Ensure log directory exists.
	logDir := logDirPath(cfg.DataDir, cfg.TeamName)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("team/spawner: mkdir logs: %w", err)
	}
	lf, err := os.Create(logPath(cfg.DataDir, cfg.TeamName, cfg.AgentID))
	if err != nil {
		return nil, fmt.Errorf("team/spawner: create log: %w", err)
	}

	args := []string{"--dangerously-skip-permissions", "-p"}
	if !cfg.SkipPermissions {
		args = []string{"-p"}
	}
	cmd := exec.Command("claude", args...)
	cmd.Env = append(os.Environ(), envSlice(env)...)
	if cfg.WorktreePath != "" {
		cmd.Dir = cfg.WorktreePath
	}

	// Feed prompt via stdin.
	f, err := os.Open(promptFile)
	if err != nil {
		lf.Close()
		return nil, fmt.Errorf("team/spawner: open prompt file: %w", err)
	}
	cmd.Stdin = f
	cmd.Stdout = lf
	cmd.Stderr = lf

	if err := cmd.Start(); err != nil {
		f.Close()
		lf.Close()
		return nil, fmt.Errorf("team/spawner: subprocess spawn %s: %w", cfg.AgentID, err)
	}
	f.Close()
	// lf stays open — the subprocess holds it via inheritance.
	return &SpawnResult{
		AgentID: cfg.AgentID,
		PID:     cmd.Process.Pid,
	}, nil
}

// --- helpers ---

// writePromptFile saves the initial prompt to a file under the team's data dir.
func writePromptFile(cfg SpawnConfig) (string, error) {
	if cfg.InitialPrompt == "" {
		return "", nil
	}
	dir := promptDir(cfg.DataDir, cfg.TeamName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := fmt.Sprintf("%s/%s.txt", dir, sanitize(cfg.AgentID))
	return path, os.WriteFile(path, []byte(cfg.InitialPrompt), 0o644)
}

func promptDir(dataDir, teamName string) string {
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.jikime"
	}
	return fmt.Sprintf("%s/teams/%s/prompts", dataDir, teamName)
}

func logDirPath(dataDir, teamName string) string {
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.jikime"
	}
	return fmt.Sprintf("%s/teams/%s/logs", dataDir, teamName)
}

// logPath returns the path for an agent's output log file (subprocess backend).
func logPath(dataDir, teamName, agentID string) string {
	return fmt.Sprintf("%s/%s.log", logDirPath(dataDir, teamName), sanitize(agentID))
}

func tmuxSessionName(teamName, agentID string) string {
	return fmt.Sprintf("jikime-%s-%s", sanitize(teamName), sanitize(agentID))
}

func sanitize(s string) string {
	return strings.NewReplacer(" ", "-", "/", "-", ":", "-").Replace(s)
}

func buildEnv(cfg SpawnConfig) map[string]string {
	env := map[string]string{
		"JIKIME_AGENT_ID":   cfg.AgentID,
		"JIKIME_TEAM_NAME":  cfg.TeamName,
		"JIKIME_ROLE":       cfg.Role,
		"JIKIME_DATA_DIR":   cfg.DataDir,
		"JIKIME_SPAWN_TIME": time.Now().UTC().Format(time.RFC3339),
	}
	if cfg.WorktreePath != "" {
		env["JIKIME_WORKTREE_PATH"] = cfg.WorktreePath
	}
	for k, v := range cfg.ExtraEnv {
		env[k] = v
	}
	return env
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
