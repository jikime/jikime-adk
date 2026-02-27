package hookscmd

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// UserPromptSubmitCmd represents the user-prompt-submit hook command
var UserPromptSubmitCmd = &cobra.Command{
	Use:   "user-prompt-submit",
	Short: "Analyze user prompt and provide agent hints",
	Long: `UserPromptSubmit hook that analyzes user input before execution:
- Suggests appropriate agents based on prompt keywords
- Warns about dangerous command patterns
- Provides helpful hints for better results`,
	RunE: runUserPromptSubmit,
}

// Agent hint patterns - keyword to agent mapping
var agentHintPatterns = []struct {
	keywords []string
	agent    string
	hint     string
}{
	{
		keywords: []string{"security", "vulnerability", "audit", "owasp", "cve", "exploit", "injection", "xss", "csrf"},
		agent:    "security-auditor",
		hint:     "Security analysis detected - security-auditor agent recommended",
	},
	{
		keywords: []string{"debug", "error", "bug", "fix", "crash", "exception", "stacktrace", "traceback"},
		agent:    "debugger",
		hint:     "Debugging task detected - debugger agent recommended",
	},
	{
		keywords: []string{"refactor", "cleanup", "simplify", "optimize code", "code smell", "technical debt"},
		agent:    "refactorer",
		hint:     "Refactoring task detected - refactorer agent recommended",
	},
	{
		keywords: []string{"test", "coverage", "unit test", "integration test", "e2e", "spec", "jest", "vitest", "pytest"},
		agent:    "test-guide",
		hint:     "Testing task detected - test-guide agent recommended",
	},
	{
		keywords: []string{"architecture", "design", "system design", "scalability", "microservice", "monolith"},
		agent:    "architect",
		hint:     "Architecture task detected - architect agent recommended",
	},
	{
		keywords: []string{"document", "readme", "api doc", "jsdoc", "docstring", "changelog"},
		agent:    "documenter",
		hint:     "Documentation task detected - documenter agent recommended",
	},
	{
		keywords: []string{"review", "code review", "pr review", "pull request"},
		agent:    "reviewer",
		hint:     "Code review detected - reviewer agent recommended",
	},
	{
		keywords: []string{"build", "compile", "webpack", "vite", "rollup", "esbuild", "bundle"},
		agent:    "build-fixer",
		hint:     "Build task detected - build-fixer agent may help with errors",
	},
	{
		keywords: []string{"deploy", "ci/cd", "pipeline", "github actions", "docker", "kubernetes", "k8s"},
		agent:    "devops",
		hint:     "DevOps task detected - devops agent recommended",
	},
	{
		keywords: []string{"performance", "slow", "optimize", "profiling", "bottleneck", "memory leak", "latency"},
		agent:    "optimizer",
		hint:     "Performance task detected - optimizer agent recommended",
	},
	{
		keywords: []string{"react", "vue", "angular", "component", "ui", "frontend", "css", "tailwind", "styled"},
		agent:    "frontend",
		hint:     "Frontend task detected - frontend agent recommended",
	},
	{
		keywords: []string{"api", "endpoint", "rest", "graphql", "database", "sql", "backend", "server"},
		agent:    "backend",
		hint:     "Backend task detected - backend agent recommended",
	},
}

// Dangerous patterns that should trigger warnings
var dangerousPatterns = []struct {
	pattern *regexp.Regexp
	warning string
}{
	{
		pattern: regexp.MustCompile(`(?i)\brm\s+(-rf?|--recursive)\s`),
		warning: "WARNING: Destructive 'rm -rf' command detected - use with caution",
	},
	{
		pattern: regexp.MustCompile(`(?i)git\s+push\s+(--force|-f)\b`),
		warning: "WARNING: 'git push --force' detected - this can overwrite remote history",
	},
	{
		pattern: regexp.MustCompile(`(?i)git\s+reset\s+--hard\b`),
		warning: "WARNING: 'git reset --hard' detected - this discards all local changes",
	},
	{
		pattern: regexp.MustCompile(`(?i)\bdrop\s+(database|table)\b`),
		warning: "WARNING: DROP DATABASE/TABLE detected - this is destructive and irreversible",
	},
	{
		pattern: regexp.MustCompile(`(?i)\btruncate\s+table\b`),
		warning: "WARNING: TRUNCATE TABLE detected - this deletes all data",
	},
	{
		pattern: regexp.MustCompile(`(?i)\bsudo\s+rm\b`),
		warning: "WARNING: 'sudo rm' detected - elevated privileges for deletion is risky",
	},
	{
		pattern: regexp.MustCompile(`(?i)chmod\s+777\b`),
		warning: "WARNING: 'chmod 777' detected - this makes files world-writable (security risk)",
	},
	{
		pattern: regexp.MustCompile(`(?i)\bformat\s+[a-z]:\b`),
		warning: "WARNING: Drive format command detected - this erases all data",
	},
	{
		pattern: regexp.MustCompile(`(?i)>\s*/dev/sd[a-z]\b`),
		warning: "WARNING: Direct write to disk device detected - this can destroy data",
	},
	{
		pattern: regexp.MustCompile(`(?i)\b:(){ :|:& };:\b`),
		warning: "WARNING: Fork bomb detected - this will crash the system",
	},
}

type userPromptInput struct {
	Prompt string `json:"prompt"`
}

type userPromptOutput struct {
	Continue      bool              `json:"continue"`
	SystemMessage string            `json:"systemMessage,omitempty"`
	Performance   map[string]bool   `json:"performance,omitempty"`
}

func runUserPromptSubmit(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input userPromptInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		// Invalid JSON - continue without hints
		return outputPromptResult("", true)
	}

	prompt := strings.ToLower(input.Prompt)
	var messages []string

	// Check for dangerous patterns first (highest priority)
	for _, dp := range dangerousPatterns {
		if dp.pattern.MatchString(input.Prompt) {
			messages = append(messages, dp.warning)
		}
	}

	// Check for agent hints
	agentSuggested := make(map[string]bool)
	for _, hint := range agentHintPatterns {
		for _, keyword := range hint.keywords {
			if strings.Contains(prompt, keyword) {
				if !agentSuggested[hint.agent] {
					messages = append(messages, "💡 "+hint.hint)
					agentSuggested[hint.agent] = true
				}
				break
			}
		}
	}

	// Build system message
	systemMessage := ""
	if len(messages) > 0 {
		systemMessage = strings.Join(messages, "\n")
	}

	return outputPromptResult(systemMessage, true)
}

func outputPromptResult(systemMessage string, continueExec bool) error {
	output := userPromptOutput{
		Continue:      continueExec,
		SystemMessage: systemMessage,
		Performance: map[string]bool{
			"user_prompt_hook": true,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
