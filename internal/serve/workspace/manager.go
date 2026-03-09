// Package workspace manages per-issue isolated workspaces using git worktrees.
// Implements Symphony SPEC Section 9.
package workspace

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var sanitizeRe = regexp.MustCompile(`[^A-Za-z0-9._-]`)

// Workspace represents an issue workspace.
type Workspace struct {
	Path       string
	Key        string
	CreatedNow bool
}

// Manager creates and manages per-issue workspace directories.
type Manager struct {
	root          string
	hookAfterCreate  string
	hookBeforeRun    string
	hookAfterRun     string
	hookBeforeRemove string
	hookTimeoutMS    int
	logger        *slog.Logger
}

// NewManager creates a workspace manager.
func NewManager(root string, opts ...Option) *Manager {
	m := &Manager{
		root:          root,
		hookTimeoutMS: 60000,
		logger:        slog.Default(),
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// Option configures a Manager.
type Option func(*Manager)

func WithHookAfterCreate(script string) Option  { return func(m *Manager) { m.hookAfterCreate = script } }
func WithHookBeforeRun(script string) Option    { return func(m *Manager) { m.hookBeforeRun = script } }
func WithHookAfterRun(script string) Option     { return func(m *Manager) { m.hookAfterRun = script } }
func WithHookBeforeRemove(script string) Option { return func(m *Manager) { m.hookBeforeRemove = script } }
func WithHookTimeoutMS(ms int) Option           { return func(m *Manager) { m.hookTimeoutMS = ms } }
func WithLogger(l *slog.Logger) Option          { return func(m *Manager) { m.logger = l } }

// CreateForIssue ensures a workspace exists for the given issue identifier.
// Runs after_create hook only on new creation.
// Safety invariant: returned path is always under root.
func (m *Manager) CreateForIssue(identifier string) (*Workspace, error) {
	key := sanitizeKey(identifier)
	wsPath, err := m.safePath(key)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(wsPath)
	createdNow := false

	if os.IsNotExist(err) {
		if err := os.MkdirAll(wsPath, 0o755); err != nil {
			return nil, fmt.Errorf("workspace create %s: %w", wsPath, err)
		}
		createdNow = true
		m.logger.Info("workspace created", "path", wsPath, "key", key)
	} else if err != nil {
		return nil, fmt.Errorf("workspace stat %s: %w", wsPath, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("workspace path exists but is not a directory: %s", wsPath)
	}

	ws := &Workspace{Path: wsPath, Key: key, CreatedNow: createdNow}

	if createdNow && m.hookAfterCreate != "" {
		if err := m.runHook("after_create", m.hookAfterCreate, wsPath); err != nil {
			// Fatal: remove partially created workspace
			os.RemoveAll(wsPath)
			return nil, fmt.Errorf("after_create hook failed: %w", err)
		}
	}

	return ws, nil
}

// BeforeRun runs the before_run hook. Failure aborts the current attempt.
func (m *Manager) BeforeRun(wsPath string) error {
	if m.hookBeforeRun == "" {
		return nil
	}
	return m.runHook("before_run", m.hookBeforeRun, wsPath)
}

// AfterRun runs the after_run hook. Failure is logged and ignored.
func (m *Manager) AfterRun(wsPath string) {
	if m.hookAfterRun == "" {
		return
	}
	if err := m.runHook("after_run", m.hookAfterRun, wsPath); err != nil {
		m.logger.Warn("after_run hook failed (ignored)", "path", wsPath, "error", err)
	}
}

// CleanupForIssue removes the workspace directory for a terminal issue.
func (m *Manager) CleanupForIssue(identifier string) {
	key := sanitizeKey(identifier)
	wsPath, err := m.safePath(key)
	if err != nil {
		m.logger.Warn("cleanup: invalid path", "identifier", identifier, "error", err)
		return
	}

	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		return
	}

	// before_remove hook (failure ignored)
	if m.hookBeforeRemove != "" {
		if err := m.runHook("before_remove", m.hookBeforeRemove, wsPath); err != nil {
			m.logger.Warn("before_remove hook failed (ignored)", "path", wsPath, "error", err)
		}
	}

	if err := os.RemoveAll(wsPath); err != nil {
		m.logger.Warn("workspace cleanup failed", "path", wsPath, "error", err)
		return
	}
	m.logger.Info("workspace cleaned", "path", wsPath)
}

// PathForIssue returns the workspace path for an identifier (without creating).
func (m *Manager) PathForIssue(identifier string) (string, error) {
	return m.safePath(sanitizeKey(identifier))
}

// --- Safety invariants ---

// safePath computes and validates the workspace path.
// Invariant: path must be under root (no path traversal).
func (m *Manager) safePath(key string) (string, error) {
	absRoot, err := filepath.Abs(m.root)
	if err != nil {
		return "", fmt.Errorf("workspace root abs: %w", err)
	}
	wsPath := filepath.Join(absRoot, key)
	absWS, err := filepath.Abs(wsPath)
	if err != nil {
		return "", fmt.Errorf("workspace path abs: %w", err)
	}
	// Enforce root containment
	if !strings.HasPrefix(absWS+string(os.PathSeparator), absRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("workspace path %q escapes root %q", absWS, absRoot)
	}
	// Ensure root exists
	if err := os.MkdirAll(absRoot, 0o755); err != nil {
		return "", fmt.Errorf("workspace root create: %w", err)
	}
	return absWS, nil
}

// --- Hook execution ---

// runHook executes a shell script in the workspace directory with timeout.
func (m *Manager) runHook(name, script, wsPath string) error {
	timeout := time.Duration(m.hookTimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-lc", script)
	cmd.Dir = wsPath
	cmd.Env = os.Environ()

	m.logger.Info("hook start", "hook", name, "path", wsPath)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("hook %s timed out after %v", name, timeout)
	}
	if err != nil {
		// Truncate output in logs
		outStr := string(out)
		if len(outStr) > 500 {
			outStr = outStr[:500] + "...(truncated)"
		}
		return fmt.Errorf("hook %s failed: %w\n%s", name, err, outStr)
	}
	m.logger.Info("hook done", "hook", name)
	return nil
}

// --- Helpers ---

// sanitizeKey replaces non-allowed characters with underscore.
// Allowed: [A-Za-z0-9._-]
func sanitizeKey(identifier string) string {
	return sanitizeRe.ReplaceAllString(identifier, "_")
}
