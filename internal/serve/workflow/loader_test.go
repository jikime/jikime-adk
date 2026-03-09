package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"jikime-adk/internal/serve"
)

func TestParse_WithFrontMatter(t *testing.T) {
	input := `---
tracker:
  kind: github
  api_key: $GITHUB_TOKEN
  project_slug: owner/repo
polling:
  interval_ms: 15000
---

You are working on {{ issue.title }}.
`
	def, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if def.PromptTemplate != "You are working on {{ issue.title }}." {
		t.Errorf("PromptTemplate = %q", def.PromptTemplate)
	}
	if def.Config["tracker"] == nil {
		t.Error("expected tracker section in config")
	}
}

func TestParse_WithoutFrontMatter(t *testing.T) {
	input := "Just a plain prompt with no front matter."
	def, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if def.PromptTemplate != input {
		t.Errorf("PromptTemplate = %q, want %q", def.PromptTemplate, input)
	}
	if len(def.Config) != 0 {
		t.Errorf("expected empty config, got %v", def.Config)
	}
}

func TestParse_MissingClosingDelimiter(t *testing.T) {
	input := `---
tracker:
  kind: github
`
	_, err := Parse([]byte(input))
	if err == nil {
		t.Error("expected error for unclosed front matter")
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	input := `---
tracker: [unclosed
---
prompt
`
	_, err := Parse([]byte(input))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParse_EmptyPromptBody(t *testing.T) {
	input := `---
tracker:
  kind: github
---
`
	def, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if def.PromptTemplate != "" {
		t.Errorf("expected empty prompt, got %q", def.PromptTemplate)
	}
}

func TestLoader_MissingFile(t *testing.T) {
	_, err := NewLoader("/nonexistent/WORKFLOW.md", nil)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoader_HotReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")

	initial := `---
tracker:
  kind: github
  api_key: $TOKEN
  project_slug: owner/repo
---
initial prompt`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	reloaded := make(chan *serve.WorkflowDefinition, 1)
	loader, err := NewLoader(path, func(def *serve.WorkflowDefinition) {
		reloaded <- def
	})
	if err != nil {
		t.Fatalf("NewLoader() error = %v", err)
	}
	defer loader.Close()

	if loader.Current().PromptTemplate != "initial prompt" {
		t.Errorf("initial prompt = %q", loader.Current().PromptTemplate)
	}

	updated := `---
tracker:
  kind: github
  api_key: $TOKEN
  project_slug: owner/repo
---
updated prompt`
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case def := <-reloaded:
		if def.PromptTemplate != "updated prompt" {
			t.Errorf("reloaded prompt = %q, want %q", def.PromptTemplate, "updated prompt")
		}
	case <-make(chan struct{}): // non-blocking fallback
	}
}
