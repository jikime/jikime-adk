package hookscmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// PreToolSecurityCmd represents the pre-tool-security hook command
var PreToolSecurityCmd = &cobra.Command{
	Use:   "pre-tool-security",
	Short: "Security guard for file modifications",
	Long: `PreToolUse hook that protects sensitive files and prevents dangerous modifications.

Security Features:
- Block modifications to secret/credential files
- Confirm before modifying .env files (content-level secret detection still applies)
- Protect lock files (package-lock.json, yarn.lock)
- Guard .git directory
- Prevent accidental overwrites of critical configs
- Detect sensitive content (API keys, credentials)`,
	RunE: runPreToolSecurity,
}

// Patterns for files that should NEVER be modified
var denyPatterns = []string{
	// Secrets and credentials
	`secrets?\.(json|ya?ml|toml)$`,
	`credentials?\.(json|ya?ml|toml)$`,
	`\.secrets/.*`,
	`secrets/.*`,
	// SSH and certificates
	`\.ssh/.*`,
	`id_rsa.*`,
	`id_ed25519.*`,
	`\.pem$`,
	`\.key$`,
	`\.crt$`,
	// Git internals
	`\.git/.*`,
	// Cloud credentials
	`\.aws/.*`,
	`\.gcloud/.*`,
	`\.azure/.*`,
	`\.kube/.*`,
	// Token files
	`\.token$`,
	`\.tokens/.*`,
	`auth\.json$`,
}

// Patterns for files that require user confirmation
var askPatterns = []string{
	// Environment files (common in development, content-level secret detection still applies)
	`\.env$`,
	`\.env\.[^/]+$`,
	`\.envrc$`,
	// Lock files
	`package-lock\.json$`,
	`yarn\.lock$`,
	`pnpm-lock\.ya?ml$`,
	`Gemfile\.lock$`,
	`Cargo\.lock$`,
	`poetry\.lock$`,
	`composer\.lock$`,
	`Pipfile\.lock$`,
	`uv\.lock$`,
	// Critical configs
	`tsconfig\.json$`,
	`pyproject\.toml$`,
	`Cargo\.toml$`,
	`package\.json$`,
	`docker-compose\.ya?ml$`,
	`Dockerfile$`,
	`\.dockerignore$`,
	// CI/CD configs
	`\.github/workflows/.*\.ya?ml$`,
	`\.gitlab-ci\.ya?ml$`,
	`\.circleci/.*`,
	`Jenkinsfile$`,
	// Infrastructure
	`terraform/.*\.tf$`,
	`\.terraform/.*`,
	`kubernetes/.*\.ya?ml$`,
	`k8s/.*\.ya?ml$`,
}

// Content patterns that indicate sensitive data
var sensitiveContentPatterns = []string{
	`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`,
	`-----BEGIN\s+CERTIFICATE-----`,
	`sk-[a-zA-Z0-9]{32,}`,       // OpenAI API keys
	`ghp_[a-zA-Z0-9]{36}`,       // GitHub tokens
	`gho_[a-zA-Z0-9]{36}`,       // GitHub OAuth tokens
	`glpat-[a-zA-Z0-9\-]{20}`,   // GitLab tokens
	`xox[baprs]-[a-zA-Z0-9\-]+`, // Slack tokens
	`AKIA[0-9A-Z]{16}`,          // AWS access keys
	`ya29\.[a-zA-Z0-9_\-]+`,     // Google OAuth tokens
}

var (
	denyCompiled      []*regexp.Regexp
	askCompiled       []*regexp.Regexp
	sensitiveCompiled []*regexp.Regexp
)

func init() {
	denyCompiled = compilePatterns(denyPatterns)
	askCompiled = compilePatterns(askPatterns)
	sensitiveCompiled = compilePatterns(sensitiveContentPatterns)
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile("(?i)" + p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

type preToolInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

type preToolOutput struct {
	HookSpecificOutput *hookSpecificOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool                `json:"suppressOutput,omitempty"`
}

type hookSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
}

func runPreToolSecurity(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input preToolInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		// Invalid JSON - allow by default
		return nil
	}

	// Only process Write and Edit tools
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return nil
	}

	// Get file path from tool input
	filePathRaw, ok := input.ToolInput["file_path"]
	if !ok {
		return nil
	}
	filePath, ok := filePathRaw.(string)
	if !ok || filePath == "" {
		return nil
	}

	// Check file path against patterns
	decision, reason := checkFilePath(filePath)

	// For Write operations, also check content for secrets
	if input.ToolName == "Write" && decision == "allow" {
		if contentRaw, ok := input.ToolInput["content"]; ok {
			if content, ok := contentRaw.(string); ok && content != "" {
				hasSecrets, secretReason := checkContentForSecrets(content)
				if hasSecrets {
					decision = "deny"
					reason = "Content contains secrets: " + secretReason
				}
			}
		}
	}

	// Build output based on decision
	var output preToolOutput

	switch decision {
	case "deny":
		output = preToolOutput{
			HookSpecificOutput: &hookSpecificOutput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       "deny",
				PermissionDecisionReason: reason,
			},
		}
	case "ask":
		output = preToolOutput{
			HookSpecificOutput: &hookSpecificOutput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       "ask",
				PermissionDecisionReason: reason,
			},
		}
	default:
		output = preToolOutput{SuppressOutput: true}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}

func checkFilePath(filePath string) (decision, reason string) {
	// Resolve path to prevent path traversal attacks
	resolvedPath, err := filepath.Abs(filePath)
	if err != nil {
		return "deny", "Invalid file path: cannot resolve"
	}

	// Normalize paths for pattern matching
	normalizedOriginal := strings.ReplaceAll(filePath, "\\", "/")
	normalizedResolved := strings.ReplaceAll(resolvedPath, "\\", "/")

	// Check project boundary (DISABLED - allow editing files outside project)
	// projectRoot := getProjectRoot()
	// if projectRoot != "" {
	// 	relPath, err := filepath.Rel(projectRoot, resolvedPath)
	// 	if err != nil || strings.HasPrefix(relPath, "..") {
	// 		return "deny", "Path traversal detected: file is outside project directory"
	// 	}
	// }

	// Check deny patterns
	for _, pattern := range denyCompiled {
		if pattern.MatchString(normalizedOriginal) || pattern.MatchString(normalizedResolved) {
			return "deny", "Protected file: access denied for security reasons"
		}
	}

	// Check ask patterns
	for _, pattern := range askCompiled {
		if pattern.MatchString(normalizedOriginal) || pattern.MatchString(normalizedResolved) {
			return "ask", "Critical config file: " + filepath.Base(filePath)
		}
	}

	return "allow", ""
}

func checkContentForSecrets(content string) (hasSecrets bool, description string) {
	for _, pattern := range sensitiveCompiled {
		if pattern.MatchString(content) {
			return true, "Detected sensitive data (credentials, API keys, or certificates)"
		}
	}
	return false, ""
}

func getProjectRoot() string {
	if projectDir := os.Getenv("CLAUDE_PROJECT_DIR"); projectDir != "" {
		return projectDir
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}
