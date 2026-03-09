// Package agent wraps the claude CLI for headless execution.
// Replaces Codex app-server (Symphony SPEC Section 10) with claude --no-interactive.
package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"jikime-adk/internal/serve"
)

// streamEvent is a parsed line from --output-format stream-json.
type streamEvent struct {
	Type              string `json:"type"`
	Result            string `json:"result"`
	IsError           bool   `json:"is_error"`
	TotalInputTokens  int    `json:"total_input_tokens"`
	TotalOutputTokens int    `json:"total_output_tokens"`
}

// RunResult is the outcome of a single agent run attempt.
type RunResult struct {
	Success     bool
	ExitCode    int
	Error       string
	LastMessage string
	Tokens      *serve.TokenUsage
	Duration    time.Duration
}

// Runner executes claude in headless mode inside a workspace directory.
type Runner struct {
	claudeCmd     string
	turnTimeoutMS int
	stallTimeoutMS int
	logger        *slog.Logger
	onEvent       func(serve.AgentEvent)
}

// Option configures a Runner.
type Option func(*Runner)

func WithClaudeCommand(cmd string) Option    { return func(r *Runner) { r.claudeCmd = cmd } }
func WithTurnTimeoutMS(ms int) Option        { return func(r *Runner) { r.turnTimeoutMS = ms } }
func WithStallTimeoutMS(ms int) Option       { return func(r *Runner) { r.stallTimeoutMS = ms } }
func WithLogger(l *slog.Logger) Option       { return func(r *Runner) { r.logger = l } }
func WithEventCallback(fn func(serve.AgentEvent)) Option {
	return func(r *Runner) { r.onEvent = fn }
}

// NewRunner creates a new agent runner.
func NewRunner(opts ...Option) *Runner {
	r := &Runner{
		claudeCmd:      "claude",
		turnTimeoutMS:  3600000,
		stallTimeoutMS: 300000,
		logger:         slog.Default(),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Run executes claude with the given prompt in the workspace directory.
// Returns RunResult. The caller is responsible for retry logic.
func (r *Runner) Run(ctx context.Context, issue *serve.Issue, prompt, workspacePath string) RunResult {
	start := time.Now()
	sessionID := fmt.Sprintf("%s-%d", issue.Identifier, start.UnixMilli())

	r.emit(serve.AgentEvent{
		Type:      serve.AgentEventStarted,
		IssueID:   issue.ID,
		Message:   fmt.Sprintf("session_id=%s workspace=%s", sessionID, workspacePath),
		Timestamp: start,
	})

	// Validate workspace cwd (safety invariant)
	if _, err := os.Stat(workspacePath); err != nil {
		return r.fail(issue, start, fmt.Sprintf("invalid workspace: %v", err))
	}

	result := r.runTurn(ctx, issue, sessionID, prompt, workspacePath, start)
	return result
}

// runTurn executes one claude turn with timeout and stall detection.
func (r *Runner) runTurn(ctx context.Context, issue *serve.Issue, sessionID, prompt, wsPath string, start time.Time) RunResult {
	turnTimeout := time.Duration(r.turnTimeoutMS) * time.Millisecond
	if turnTimeout <= 0 {
		turnTimeout = time.Hour
	}

	turnCtx, cancel := context.WithTimeout(ctx, turnTimeout)
	defer cancel()

	// Build claude command with stream-json output for token tracking.
	// --print: non-interactive mode
	// --output-format stream-json: NDJSON events including token counts in final result
	// --verbose: required by claude CLI when using --output-format stream-json with --print
	// --dangerously-skip-permissions: allow tool use without interactive prompts
	cmd := exec.CommandContext(turnCtx, r.claudeCmd,
		"--print",
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
		prompt,
	)
	cmd.Dir = wsPath

	// Unset CLAUDECODE to allow nested execution inside a Claude Code session
	env := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "CLAUDECODE=") {
			env = append(env, e)
		}
	}
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return r.fail(issue, start, fmt.Sprintf("stdout pipe: %v", err))
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return r.fail(issue, start, fmt.Sprintf("stderr pipe: %v", err))
	}

	if err := cmd.Start(); err != nil {
		return r.fail(issue, start, fmt.Sprintf("claude not found or failed to start: %v", err))
	}

	// Stream stdout, detect stall
	stallTimeout := time.Duration(r.stallTimeoutMS) * time.Millisecond
	lastActivity := time.Now()
	var lastMessage string
	var tokens *serve.TokenUsage

	outputCh := make(chan string, 64)
	doneCh := make(chan struct{})

	// Read stdout in goroutine
	go func() {
		defer close(outputCh)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			outputCh <- line
		}
	}()

	// Drain stderr — log at Warn so errors are visible in default log level
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			r.logger.Warn("claude stderr", "issue_id", issue.ID, "line", scanner.Text())
		}
	}()

	// Stall detection ticker
	stallTicker := time.NewTicker(5 * time.Second)
	defer stallTicker.Stop()

	go func() {
		defer close(doneCh)
		for {
			select {
			case line, ok := <-outputCh:
				if !ok {
					return
				}
				lastActivity = time.Now()

				// Parse stream-json event; extract tokens from final result event.
				msg := line
				var ev streamEvent
				if json.Unmarshal([]byte(line), &ev) == nil {
					switch ev.Type {
					case "result":
						tokens = &serve.TokenUsage{
							InputTokens:  ev.TotalInputTokens,
							OutputTokens: ev.TotalOutputTokens,
							TotalTokens:  ev.TotalInputTokens + ev.TotalOutputTokens,
						}
						if ev.Result != "" {
							msg = ev.Result
						}
					}
				}

				lastMessage = truncate(msg, 200)
				r.emit(serve.AgentEvent{
					Type:      serve.AgentEventMessage,
					IssueID:   issue.ID,
					Message:   lastMessage,
					Timestamp: lastActivity,
				})
			case <-stallTicker.C:
				if stallTimeout > 0 && time.Since(lastActivity) > stallTimeout {
					r.emit(serve.AgentEvent{
						Type:      serve.AgentEventStalled,
						IssueID:   issue.ID,
						Message:   fmt.Sprintf("no activity for %v", stallTimeout),
						Timestamp: time.Now(),
					})
					cancel() // trigger context cancellation
					return
				}
			case <-turnCtx.Done():
				return
			}
		}
	}()

	<-doneCh
	err = cmd.Wait()
	duration := time.Since(start)

	if turnCtx.Err() == context.DeadlineExceeded {
		return RunResult{
			Success:     false,
			Error:       fmt.Sprintf("turn_timeout after %v", turnTimeout),
			LastMessage: lastMessage,
			Duration:    duration,
		}
	}

	if err != nil {
		exitCode := 0
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		// Check for user input request (claude exits with specific message)
		if strings.Contains(lastMessage, "waiting for input") ||
			strings.Contains(lastMessage, "user input required") {
			return RunResult{
				Success:  false,
				ExitCode: exitCode,
				Error:    "turn_input_required",
				Duration: duration,
			}
		}
		r.emit(serve.AgentEvent{
			Type:      serve.AgentEventFailed,
			IssueID:   issue.ID,
			Message:   fmt.Sprintf("exit_code=%d", exitCode),
			Timestamp: time.Now(),
		})
		return RunResult{
			Success:     false,
			ExitCode:    exitCode,
			Error:       fmt.Sprintf("turn_failed: exit code %d", exitCode),
			LastMessage: lastMessage,
			Duration:    duration,
		}
	}

	r.emit(serve.AgentEvent{
		Type:      serve.AgentEventCompleted,
		IssueID:   issue.ID,
		Message:   sessionID,
		Timestamp: time.Now(),
	})

	return RunResult{
		Success:     true,
		LastMessage: lastMessage,
		Tokens:      tokens,
		Duration:    duration,
	}
}

func (r *Runner) fail(issue *serve.Issue, start time.Time, msg string) RunResult {
	r.logger.Error("agent run failed", "issue_id", issue.ID, "error", msg)
	return RunResult{
		Success:  false,
		Error:    msg,
		Duration: time.Since(start),
	}
}

func (r *Runner) emit(e serve.AgentEvent) {
	if r.onEvent != nil {
		r.onEvent(e)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
