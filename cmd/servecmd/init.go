package servecmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// workflowTemplateBasic is the WORKFLOW.md template for projects without JiKiME-ADK.
const workflowTemplateBasic = `---
tracker:
  kind: github
  # api_key: $GITHUB_TOKEN   # omit → uses gh auth token automatically
  project_slug: {{.Slug}}
  active_states:
    - {{.Label}}
  terminal_states:
    - jikime-done
    - Done

polling:
  interval_ms: 15000

workspace:
  root: {{.WorkspaceRoot}}

hooks:
  after_create: |
    git clone https://github.com/{{.Slug}}.git .
    echo "[after_create] cloned repo to $(pwd)"

  before_run: |
    echo "[before_run] syncing to latest main..."
    git fetch origin
    git checkout main
    git reset --hard origin/main
    echo "[before_run] ready at $(git rev-parse --short HEAD)"

  after_run: |
    echo "[after_run] done"
    if [ -d "{{.WorkDir}}/.git" ]; then
      cd "{{.WorkDir}}" && git pull --ff-only 2>&1 \
        && echo "[after_run] local repo synced at $(git rev-parse --short HEAD)" \
        || echo "[after_run] git pull skipped (local changes or diverged branch)"
    fi

  timeout_ms: 60000

agent:
  max_concurrent_agents: {{.MaxAgents}}
  max_turns: 5
  max_retry_backoff_ms: 300000

claude:
  command: claude
  turn_timeout_ms: 3600000
  stall_timeout_ms: 180000

server:
  port: {{.Port}}
---

You are an autonomous software engineer working on a GitHub issue.

Repository: https://github.com/{{.Slug}}

## Issue

**{{ "{{" }} issue.identifier {{ "}}" }}**: {{ "{{" }} issue.title {{ "}}" }}

{{ "{{" }} issue.description {{ "}}" }}

## Instructions

1. Read the issue carefully and implement what is requested.
2. Create a feature branch: ` + "`" + `git checkout -b fix/issue-{{ "{{" }} issue.id {{ "}}" }}` + "`" + `
3. Make your changes using the available file tools.
4. Commit: ` + "`" + `git add -A && git commit -m "fix: {{ "{{" }} issue.identifier {{ "}}" }} - {{ "{{" }} issue.title {{ "}}" }}"` + "`" + `
5. Push the branch: ` + "`" + `git push origin fix/issue-{{ "{{" }} issue.id {{ "}}" }}` + "`" + `
6. Create a pull request:
   ` + "`" + `gh pr create --title "fix: {{ "{{" }} issue.title {{ "}}" }}" --body "Closes #{{ "{{" }} issue.id {{ "}}" }}" --base main --head fix/issue-{{ "{{" }} issue.id {{ "}}" }}` + "`" + `
7. Merge the pull request and delete the branch:
   ` + "`" + `gh pr merge --squash --delete-branch --admin` + "`" + `

Work in the current directory. The repository has already been cloned here.
`

// workflowTemplateJikiME is the WORKFLOW.md template for projects with JiKiME-ADK installed.
const workflowTemplateJikiME = `---
tracker:
  kind: github
  # api_key: $GITHUB_TOKEN   # omit → uses gh auth token automatically
  project_slug: {{.Slug}}
  active_states:
    - {{.Label}}
  terminal_states:
    - jikime-done
    - Done

polling:
  interval_ms: 15000

workspace:
  root: {{.WorkspaceRoot}}

hooks:
  after_create: |
    git clone https://github.com/{{.Slug}}.git .
    echo "[after_create] cloned repo to $(pwd)"

  before_run: |
    echo "[before_run] syncing to latest main..."
    git fetch origin
    git checkout main
    git reset --hard origin/main
    echo "[before_run] ready at $(git rev-parse --short HEAD)"

  after_run: |
    echo "[after_run] done"
    if [ -d "{{.WorkDir}}/.git" ]; then
      cd "{{.WorkDir}}" && git pull --ff-only 2>&1 \
        && echo "[after_run] local repo synced at $(git rev-parse --short HEAD)" \
        || echo "[after_run] git pull skipped (local changes or diverged branch)"
    fi

  timeout_ms: 60000

agent:
  max_concurrent_agents: {{.MaxAgents}}
  max_turns: 10
  max_retry_backoff_ms: 300000

claude:
  command: claude
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000

server:
  port: {{.Port}}
---

You are an autonomous software engineer working on a GitHub issue.
This repository has JiKiME-ADK installed (.claude/ directory is present).
CLAUDE.md is automatically loaded — the full J.A.R.V.I.S. agent stack is available.

Repository: https://github.com/{{.Slug}}

## Issue

**{{ "{{" }} issue.identifier {{ "}}" }}**: {{ "{{" }} issue.title {{ "}}" }}

{{ "{{" }} issue.description {{ "}}" }}

## Instructions

Use the ` + "`" + `jarvis` + "`" + ` sub-agent to implement this issue. J.A.R.V.I.S. will orchestrate
specialist agents based on the issue type.

Invoke jarvis with this prompt:
"Implement the following GitHub issue on branch fix/issue-{{ "{{" }} issue.id {{ "}}" }}:

Title: {{ "{{" }} issue.title {{ "}}" }}

{{ "{{" }} issue.description {{ "}}" }}

After implementation, create a PR and merge it:
  gh pr create --title 'fix: {{ "{{" }} issue.title {{ "}}" }}' --body 'Closes #{{ "{{" }} issue.id {{ "}}" }}' --base main --head fix/issue-{{ "{{" }} issue.id {{ "}}" }}
  gh pr merge --squash --delete-branch --admin"
`

type workflowParams struct {
	Slug          string
	Label         string
	WorkspaceRoot string
	WorkDir       string
	Port          int
	MaxAgents     int
}

// NewServeInit returns the `jikime serve init` sub-command.
func NewServeInit() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a WORKFLOW.md file interactively",
		Long: `Interactive wizard that generates a WORKFLOW.md configuration file
for use with 'jikime serve'.

Automatically detects your GitHub repository and whether JiKiME-ADK
is installed (.claude/ directory), then generates an optimized prompt
template accordingly.

Examples:
  jikime serve init              # creates ./WORKFLOW.md
  jikime serve init -o my.md    # custom output path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServeInit(outputPath)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "WORKFLOW.md", "Output file path")
	return cmd
}

func runServeInit(outputPath string) error {
	printInitHeader()

	// --- Step 1: GitHub repo slug ---
	detectedSlug := detectGitRemoteSlug()
	slug, err := promptInput("GitHub repo slug (owner/repo)", detectedSlug, validateSlug)
	if err != nil {
		return printAbort()
	}

	// --- Step 2: Active label ---
	label, err := promptInput("Active label (issues with this label will be processed)", "jikime-todo", nil)
	if err != nil {
		return printAbort()
	}

	// --- Step 3: Workspace root ---
	repoName := slug
	if parts := strings.SplitN(slug, "/", 2); len(parts) == 2 {
		repoName = parts[1]
	}
	defaultWorkspace := fmt.Sprintf("/tmp/jikime-%s", strings.ReplaceAll(repoName, "/", "-"))
	workspaceRoot, err := promptInput("Workspace root directory", defaultWorkspace, nil)
	if err != nil {
		return printAbort()
	}

	// --- Step 4: HTTP port ---
	port, err := promptSelect("HTTP status API port", []selectOpt{
		{"8888", "8888 (recommended)"},
		{"9999", "9999"},
		{"0", "0 (disabled)"},
	})
	if err != nil {
		return printAbort()
	}
	portInt := 8888
	switch port {
	case "9999":
		portInt = 9999
	case "0":
		portInt = 0
	}

	// --- Step 5: Max concurrent agents ---
	maxAgents, err := promptSelect("Max concurrent agents", []selectOpt{
		{"1", "1 (safe, recommended for new projects)"},
		{"3", "3 (parallel processing)"},
		{"5", "5 (high throughput)"},
	})
	if err != nil {
		return printAbort()
	}
	maxAgentsInt := 1
	switch maxAgents {
	case "3":
		maxAgentsInt = 3
	case "5":
		maxAgentsInt = 5
	}

	// --- Detect JiKiME-ADK mode ---
	cwd, _ := os.Getwd()
	hasJikiME := detectJikiME(cwd)

	// --- Generate WORKFLOW.md ---
	params := workflowParams{
		Slug:          slug,
		Label:         label,
		WorkspaceRoot: workspaceRoot,
		WorkDir:       cwd,
		Port:          portInt,
		MaxAgents:     maxAgentsInt,
	}

	tmplStr := workflowTemplateBasic
	mode := "basic"
	if hasJikiME {
		tmplStr = workflowTemplateJikiME
		mode = "jikime-adk"
	}

	content, err := renderTemplate(tmplStr, params)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	// --- Write output ---
	absOut, _ := filepath.Abs(outputPath)
	if err := os.WriteFile(absOut, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", absOut, err)
	}

	printSuccess(slug, label, workspaceRoot, portInt, maxAgentsInt, mode, absOut)
	return nil
}

// --- Helpers ---

func printInitHeader() {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("  ╔══════════════════════════════════════╗")
	color.New(color.FgCyan, color.Bold).Println("  ║   jikime serve init — Harness Wizard ║")
	color.New(color.FgCyan, color.Bold).Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
}

func printAbort() error {
	fmt.Println()
	color.Yellow("  ⚠ Aborted")
	return nil
}

func printSuccess(slug, label, workspace string, port, maxAgents int, mode, outPath string) {
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("  ✓ WORKFLOW.md created")
	fmt.Println()

	color.Cyan("  Configuration:")
	fmt.Printf("    Repo:       %s\n", color.WhiteString(slug))
	fmt.Printf("    Label:      %s\n", color.WhiteString(label))
	fmt.Printf("    Workspace:  %s\n", color.WhiteString(workspace))
	fmt.Printf("    Port:       %s\n", color.WhiteString("%d", port))
	fmt.Printf("    Agents:     %s\n", color.WhiteString("%d", maxAgents))
	if mode == "jikime-adk" {
		fmt.Printf("    Mode:       %s\n", color.MagentaString("JiKiME-ADK (J.A.R.V.I.S. agent stack)"))
	} else {
		fmt.Printf("    Mode:       %s\n", color.WhiteString("Basic"))
	}
	fmt.Println()

	color.Cyan("  Next steps:")
	fmt.Printf("    1. Create GitHub labels:\n")
	fmt.Printf("       %s\n", color.YellowString("gh label create \"%s\" --repo %s --description \"Ready for AI agent\" --color \"0e8a16\"", label, slug))
	fmt.Printf("       %s\n", color.YellowString("gh label create \"jikime-done\" --repo %s --description \"Completed by AI agent\" --color \"6f42c1\"", slug))
	fmt.Println()
	fmt.Printf("    2. Start the service:\n")
	fmt.Printf("       %s\n", color.GreenString("jikime serve %s", outPath))
	fmt.Println()
}

func detectGitRemoteSlug() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	url := strings.TrimSpace(string(out))
	return parseGitRemoteSlug(url)
}

// parseGitRemoteSlug extracts "owner/repo" from various git remote URL formats.
func parseGitRemoteSlug(url string) string {
	// SSH: git@github.com:owner/repo.git
	sshRe := regexp.MustCompile(`github\.com[:/]([^/]+/[^/]+?)(?:\.git)?$`)
	if m := sshRe.FindStringSubmatch(url); len(m) == 2 {
		return m[1]
	}
	// HTTPS: https://github.com/owner/repo.git
	httpsRe := regexp.MustCompile(`github\.com/([^/]+/[^/]+?)(?:\.git)?$`)
	if m := httpsRe.FindStringSubmatch(url); len(m) == 2 {
		return m[1]
	}
	return ""
}

func detectJikiME(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".claude"))
	return err == nil
}

func validateSlug(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("required")
	}
	if !strings.Contains(s, "/") {
		return fmt.Errorf("must be owner/repo format")
	}
	return nil
}

type selectOpt struct {
	Value   string
	Display string
}

func promptInput(label, defaultVal string, validate func(string) error) (string, error) {
	p := promptui.Prompt{
		Label:   label,
		Default: defaultVal,
	}
	if validate != nil {
		p.Validate = validate
	}
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result), nil
}

func promptSelect(label string, opts []selectOpt) (string, error) {
	displays := make([]string, len(opts))
	for i, o := range opts {
		displays[i] = o.Display
	}
	s := promptui.Select{
		Label:    label,
		Items:    displays,
		HideHelp: true,
	}
	idx, _, err := s.Run()
	if err != nil {
		return "", err
	}
	return opts[idx].Value, nil
}

func renderTemplate(tmplStr string, params workflowParams) (string, error) {
	tmpl, err := template.New("workflow").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", err
	}
	return buf.String(), nil
}
