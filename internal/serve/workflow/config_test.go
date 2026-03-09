package workflow

import (
	"os"
	"testing"

	"jikime-adk/internal/serve"
)

func makeConfig(yaml string) *Config {
	def, err := Parse([]byte(yaml))
	if err != nil {
		panic(err)
	}
	return NewConfig(def)
}

func TestConfig_Defaults(t *testing.T) {
	cfg := makeConfig("---\n{}\n---\n")

	if cfg.PollIntervalMS() != 30000 {
		t.Errorf("PollIntervalMS = %d, want 30000", cfg.PollIntervalMS())
	}
	if cfg.MaxConcurrentAgents() != 10 {
		t.Errorf("MaxConcurrentAgents = %d, want 10", cfg.MaxConcurrentAgents())
	}
	if cfg.MaxRetryBackoffMS() != 300000 {
		t.Errorf("MaxRetryBackoffMS = %d, want 300000", cfg.MaxRetryBackoffMS())
	}
	if cfg.HookTimeoutMS() != 60000 {
		t.Errorf("HookTimeoutMS = %d, want 60000", cfg.HookTimeoutMS())
	}
	if cfg.ClaudeCommand() != "claude" {
		t.Errorf("ClaudeCommand = %q, want claude", cfg.ClaudeCommand())
	}

	activeStates := cfg.TrackerActiveStates()
	if len(activeStates) != 2 {
		t.Errorf("ActiveStates len = %d, want 2", len(activeStates))
	}
}

func TestConfig_VarResolution(t *testing.T) {
	os.Setenv("TEST_API_KEY", "my-secret-key")
	defer os.Unsetenv("TEST_API_KEY")

	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: $TEST_API_KEY
  project_slug: owner/repo
---
`)
	if cfg.TrackerAPIKey() != "my-secret-key" {
		t.Errorf("TrackerAPIKey = %q, want my-secret-key", cfg.TrackerAPIKey())
	}
}

func TestConfig_APIKey_ExplicitTakesPrecedenceOverGh(t *testing.T) {
	// When api_key is explicitly set, gh auth token must NOT be called.
	os.Setenv("EXPLICIT_KEY", "explicit-wins")
	defer os.Unsetenv("EXPLICIT_KEY")

	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: $EXPLICIT_KEY
  project_slug: owner/repo
---
`)
	got := cfg.TrackerAPIKey()
	if got != "explicit-wins" {
		t.Errorf("TrackerAPIKey = %q, want explicit-wins (should not fall back to gh)", got)
	}
}

func TestConfig_APIKey_GhFallback_DoesNotPanic(t *testing.T) {
	// When api_key is absent, TrackerAPIKey() attempts gh auth token.
	// Whether gh is installed or not, the call must not panic and must return a string.
	cfg := makeConfig(`---
tracker:
  kind: github
  project_slug: owner/repo
---
`)
	key := cfg.TrackerAPIKey() // gh may or may not be installed; must not panic
	_ = key                    // value is environment-dependent
}

func TestConfig_LiteralAPIKey(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: literal-token-123
  project_slug: owner/repo
---
`)
	if cfg.TrackerAPIKey() != "literal-token-123" {
		t.Errorf("TrackerAPIKey = %q", cfg.TrackerAPIKey())
	}
}

func TestConfig_Validate_MissingKind(t *testing.T) {
	cfg := makeConfig("---\n{}\n---\n")
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing tracker.kind")
	}
}

func TestConfig_Validate_UnsupportedKind(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: jira
  api_key: token
  project_slug: proj
---
`)
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for unsupported tracker.kind")
	}
}

func TestConfig_Validate_MissingAPIKey(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  project_slug: owner/repo
---
`)
	// If gh auth token resolves a token, missing api_key in WORKFLOW.md is valid.
	// Skip in environments where gh is logged in.
	if cfg.TrackerAPIKey() != "" {
		t.Skip("gh auth token available; api_key not required in this environment")
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing api_key")
	}
}

func TestConfig_Validate_MissingProjectSlug(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: token
---
`)
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing project_slug")
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: token
  project_slug: owner/repo
---
`)
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error = %v", err)
	}
}

func TestConfig_RenderPrompt_BasicVariables(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: token
  project_slug: owner/repo
---
Fix {{ issue.identifier }}: {{ issue.title }}
State: {{ issue.state }}`)

	issue := &serve.Issue{
		ID:         "42",
		Identifier: "owner/repo#42",
		Title:      "Fix the bug",
		State:      "In Progress",
	}

	rendered, err := cfg.RenderPrompt(issue, nil)
	if err != nil {
		t.Fatalf("RenderPrompt() error = %v", err)
	}

	want := "Fix owner/repo#42: Fix the bug\nState: In Progress"
	if rendered != want {
		t.Errorf("rendered = %q\nwant     = %q", rendered, want)
	}
}

func TestConfig_RenderPrompt_AttemptVariable(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: token
  project_slug: owner/repo
---
Attempt: {{ attempt }}`)

	issue := &serve.Issue{Identifier: "owner/repo#1", Title: "t"}
	attempt := 3

	rendered, err := cfg.RenderPrompt(issue, &attempt)
	if err != nil {
		t.Fatalf("RenderPrompt() error = %v", err)
	}
	if rendered != "Attempt: 3" {
		t.Errorf("rendered = %q", rendered)
	}
}

func TestConfig_RenderPrompt_UnresolvedVariable(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: token
  project_slug: owner/repo
---
Hello {{ unknown.variable }}`)

	issue := &serve.Issue{Identifier: "x#1", Title: "t"}
	_, err := cfg.RenderPrompt(issue, nil)
	if err == nil {
		t.Error("expected error for unresolved template variable")
	}
}

func TestConfig_ActiveStates_List(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: token
  project_slug: owner/repo
  active_states:
    - todo
    - in-progress
    - review
---
`)
	states := cfg.TrackerActiveStates()
	if len(states) != 3 {
		t.Errorf("ActiveStates len = %d, want 3", len(states))
	}
	if states[0] != "todo" {
		t.Errorf("states[0] = %q, want todo", states[0])
	}
}

func TestConfig_ActiveStates_CommaSeparated(t *testing.T) {
	cfg := makeConfig(`---
tracker:
  kind: github
  api_key: token
  project_slug: owner/repo
  active_states: "todo, in-progress"
---
`)
	states := cfg.TrackerActiveStates()
	if len(states) != 2 {
		t.Errorf("ActiveStates len = %d, want 2", len(states))
	}
}
