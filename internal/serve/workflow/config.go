package workflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"jikime-adk/internal/serve"
)

// Config provides typed access to workflow configuration with defaults and $VAR resolution.
// Implements Symphony SPEC Section 6.
type Config struct {
	def *serve.WorkflowDefinition
}

// NewConfig creates a Config from a WorkflowDefinition.
func NewConfig(def *serve.WorkflowDefinition) *Config {
	return &Config{def: def}
}

// --- Tracker ---

func (c *Config) TrackerKind() string {
	return c.getString("tracker", "kind", "")
}

func (c *Config) TrackerEndpoint() string {
	switch c.TrackerKind() {
	case "linear":
		return c.getString("tracker", "endpoint", "https://api.linear.app/graphql")
	case "github":
		return c.getString("tracker", "endpoint", "https://api.github.com")
	default:
		return c.getString("tracker", "endpoint", "")
	}
}

func (c *Config) TrackerAPIKey() string {
	key := resolveVar(c.getString("tracker", "api_key", ""))
	if key != "" {
		return key
	}
	// Fallback: gh auth token (uses currently logged-in gh CLI session)
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (c *Config) TrackerProjectSlug() string {
	return c.getString("tracker", "project_slug", "")
}

func (c *Config) TrackerActiveStates() []string {
	return c.getStringList("tracker", "active_states", []string{"Todo", "In Progress"})
}

func (c *Config) TrackerTerminalStates() []string {
	return c.getStringList("tracker", "terminal_states",
		[]string{"Closed", "Cancelled", "Canceled", "Duplicate", "Done"})
}

// --- Polling ---

func (c *Config) PollIntervalMS() int {
	return c.getInt("polling", "interval_ms", 30000)
}

// --- Workspace ---

func (c *Config) WorkspaceRoot() string {
	raw := c.getString("workspace", "root", "")
	if raw == "" {
		return filepath.Join(os.TempDir(), "jikime_workspaces")
	}
	return expandPath(resolveVar(raw))
}

// --- Hooks ---

func (c *Config) HookAfterCreate() string  { return c.getString("hooks", "after_create", "") }
func (c *Config) HookBeforeRun() string    { return c.getString("hooks", "before_run", "") }
func (c *Config) HookAfterRun() string     { return c.getString("hooks", "after_run", "") }
func (c *Config) HookBeforeRemove() string { return c.getString("hooks", "before_remove", "") }
func (c *Config) HookTimeoutMS() int       { return c.getInt("hooks", "timeout_ms", 60000) }

// --- Agent ---

func (c *Config) MaxConcurrentAgents() int {
	return c.getInt("agent", "max_concurrent_agents", 10)
}

func (c *Config) MaxRetryBackoffMS() int {
	return c.getInt("agent", "max_retry_backoff_ms", 300000)
}

func (c *Config) MaxTurns() int {
	return c.getInt("agent", "max_turns", 20)
}

// --- Claude (replaces Codex in Symphony) ---

func (c *Config) ClaudeCommand() string {
	return c.getString("claude", "command", "claude")
}

func (c *Config) TurnTimeoutMS() int {
	return c.getInt("claude", "turn_timeout_ms", 3600000)
}

func (c *Config) StallTimeoutMS() int {
	return c.getInt("claude", "stall_timeout_ms", 300000)
}

// --- HTTP Server (optional extension) ---

func (c *Config) ServerPort() int {
	return c.getInt("server", "port", 0)
}

// --- Validation ---

// Validate runs dispatch preflight checks.
// Returns error if critical config is missing.
func (c *Config) Validate() error {
	if c.TrackerKind() == "" {
		return fmt.Errorf("missing tracker.kind")
	}
	switch c.TrackerKind() {
	case "github", "linear":
		// supported
	default:
		return fmt.Errorf("unsupported tracker.kind: %q", c.TrackerKind())
	}
	if c.TrackerAPIKey() == "" {
		return fmt.Errorf("missing tracker.api_key (or $VAR env var is unset)")
	}
	if c.TrackerProjectSlug() == "" {
		return fmt.Errorf("missing tracker.project_slug")
	}
	return nil
}

// --- Prompt Rendering ---

// RenderPrompt renders the prompt template with issue data.
// Uses a Liquid-compatible subset of variable substitution.
func (c *Config) RenderPrompt(issue *serve.Issue, attempt *int) (string, error) {
	tmpl := c.def.PromptTemplate
	if tmpl == "" {
		tmpl = "You are working on issue {{ issue.identifier }}: {{ issue.title }}\n\n{{ issue.description }}"
	}

	attemptStr := ""
	if attempt != nil {
		attemptStr = strconv.Itoa(*attempt)
	}

	replacements := [][2]string{
		{"{{ issue.id }}", issue.ID},
		{"{{ issue.identifier }}", issue.Identifier},
		{"{{ issue.title }}", issue.Title},
		{"{{ issue.description }}", issue.Description},
		{"{{ issue.state }}", issue.State},
		{"{{ issue.url }}", issue.URL},
		{"{{ issue.branch_name }}", issue.BranchName},
		{"{{ attempt }}", attemptStr},
	}

	result := tmpl
	for _, pair := range replacements {
		result = strings.ReplaceAll(result, pair[0], pair[1])
	}

	// Strict mode: fail on unresolved variables
	if strings.Contains(result, "{{") {
		return "", fmt.Errorf("template_render_error: unresolved template variables in prompt")
	}

	return result, nil
}

// --- Helpers ---

func (c *Config) getSection(section string) map[string]any {
	if c.def == nil || c.def.Config == nil {
		return nil
	}
	v, ok := c.def.Config[section]
	if !ok {
		return nil
	}
	m, _ := v.(map[string]any)
	return m
}

func (c *Config) getString(section, key, defaultVal string) string {
	m := c.getSection(section)
	if m == nil {
		return defaultVal
	}
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	s, _ := v.(string)
	if s == "" {
		return defaultVal
	}
	return s
}

func (c *Config) getInt(section, key string, defaultVal int) int {
	m := c.getSection(section)
	if m == nil {
		return defaultVal
	}
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch x := v.(type) {
	case int:
		if x <= 0 {
			return defaultVal
		}
		return x
	case string:
		n, err := strconv.Atoi(x)
		if err != nil || n <= 0 {
			return defaultVal
		}
		return n
	}
	return defaultVal
}

func (c *Config) getStringList(section, key string, defaultVal []string) []string {
	m := c.getSection(section)
	if m == nil {
		return defaultVal
	}
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch x := v.(type) {
	case []any:
		var result []string
		for _, item := range x {
			if s, ok := item.(string); ok {
				result = append(result, strings.TrimSpace(s))
			}
		}
		if len(result) == 0 {
			return defaultVal
		}
		return result
	case string:
		var result []string
		for _, s := range strings.Split(x, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) == 0 {
			return defaultVal
		}
		return result
	}
	return defaultVal
}

// resolveVar resolves $VAR_NAME to environment variable value.
func resolveVar(s string) string {
	if strings.HasPrefix(s, "$") {
		return os.Getenv(s[1:])
	}
	return s
}

// expandPath expands ~ to home directory.
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
