package hookscmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// PostToolLinterCmd represents the post-tool-linter hook command
var PostToolLinterCmd = &cobra.Command{
	Use:   "post-tool-linter",
	Short: "Automatic code linting after file modifications",
	Long: `PostToolUse hook that automatically runs linters after Claude writes or edits files.

Supports multiple languages with automatic tool detection:
- Python: ruff check, mypy
- JavaScript/TypeScript: eslint, biome lint
- Go: golangci-lint, go vet
- Rust: clippy
- And more...

Exit code 2 indicates lint issues that need attention.`,
	RunE: runPostToolLinter,
}

// Maximum number of issues to report
const maxIssuesToReport = 5

// Linter definitions by file extension
type linterDef struct {
	command string
	args    []string
}

var linters = map[string][]linterDef{
	".py": {
		{command: "ruff", args: []string{"check", "--output-format", "text"}},
		{command: "mypy", args: []string{"--no-error-summary"}},
	},
	".js": {
		{command: "eslint", args: []string{"--format", "compact"}},
		{command: "biome", args: []string{"lint"}},
	},
	".jsx": {
		{command: "eslint", args: []string{"--format", "compact"}},
		{command: "biome", args: []string{"lint"}},
	},
	".ts": {
		{command: "eslint", args: []string{"--format", "compact"}},
		{command: "biome", args: []string{"lint"}},
	},
	".tsx": {
		{command: "eslint", args: []string{"--format", "compact"}},
		{command: "biome", args: []string{"lint"}},
	},
	".go": {
		{command: "golangci-lint", args: []string{"run", "--out-format", "line-number"}},
		{command: "go", args: []string{"vet"}},
	},
	".rs": {
		{command: "cargo", args: []string{"clippy", "--message-format", "short", "--"}},
	},
}

// File extensions to skip for linting
var skipLintExtensions = map[string]bool{
	".json": true, ".lock": true, ".min.js": true, ".min.css": true,
	".map": true, ".svg": true, ".png": true, ".jpg": true,
	".gif": true, ".ico": true, ".woff": true, ".woff2": true,
	".ttf": true, ".eot": true, ".pdf": true, ".zip": true,
	".md": true, // Markdown usually doesn't need linting
}

// Directories to skip for linting
var skipLintDirectories = map[string]bool{
	"node_modules": true, ".git": true, ".venv": true, "venv": true,
	"__pycache__": true, ".cache": true, "dist": true, "build": true,
	".next": true, ".nuxt": true, "target": true, "vendor": true,
}

type linterInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

type linterOutput struct {
	HookSpecificOutput *linterHookOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool              `json:"suppressOutput,omitempty"`
}

type linterHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func runPostToolLinter(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input linterInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
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

	// Check if we should skip
	shouldSkip, _ := shouldSkipLinting(filePath)
	if shouldSkip {
		output := linterOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		return encoder.Encode(output)
	}

	// Get linters for this file
	ext := strings.ToLower(filepath.Ext(filePath))
	linterList, ok := linters[ext]
	if !ok {
		output := linterOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		return encoder.Encode(output)
	}

	// Try linters until one succeeds
	var lintedWith string
	var issues []string

	for _, linter := range linterList {
		// Check if linter exists
		if _, err := exec.LookPath(linter.command); err != nil {
			continue
		}

		// Run linter
		linterArgs := append(linter.args, filePath)
		lintCmd := exec.Command(linter.command, linterArgs...)
		output, err := lintCmd.CombinedOutput()

		lintedWith = linter.command

		if err != nil {
			// Linter found issues - parse them
			issues = parseLintOutput(string(output))
		}

		break // Only run first available linter
	}

	// Build output
	var output linterOutput

	if len(issues) > 0 {
		// Issues found - report to Claude
		issueCount := len(issues)
		var issuesSummary string

		if issueCount <= 3 {
			issuesSummary = strings.Join(issues, "; ")
		} else {
			issuesSummary = strings.Join(issues[:3], "; ")
			issuesSummary += " (+" + string(rune('0'+issueCount-3)) + " more)"
		}

		// Use stderr and exit code 2 to alert Claude
		os.Stderr.WriteString("Lint issues found: " + issuesSummary + "\n")
		os.Exit(2)
	} else if lintedWith != "" {
		// Linting passed with no issues
		output = linterOutput{SuppressOutput: true}
	} else {
		// No linter available
		output = linterOutput{SuppressOutput: true}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}

func shouldSkipLinting(filePath string) (bool, string) {
	// Check extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if skipLintExtensions[ext] {
		return true, "Skipped: " + ext + " files are not linted"
	}

	// Check for minified files
	if strings.Contains(filepath.Base(filePath), ".min.") {
		return true, "Skipped: minified file"
	}

	// Check if in skip directory
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for _, part := range parts {
		if skipLintDirectories[part] {
			return true, "Skipped: file in " + part + "/ directory"
		}
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true, "Skipped: file does not exist"
	}

	return false, ""
}

func parseLintOutput(output string) []string {
	var issues []string

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip non-issue lines
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "warning:") ||
			strings.Contains(lineLower, "info:") ||
			strings.Contains(lineLower, "running") ||
			strings.Contains(lineLower, "checking") ||
			strings.Contains(lineLower, "finished") ||
			strings.Contains(lineLower, "success") ||
			strings.Contains(line, "✓") ||
			strings.Contains(line, "✔") {
			continue
		}

		// Look for error patterns
		if strings.Contains(lineLower, "error") || strings.Contains(line, ":") {
			// Truncate long messages
			if len(line) > 200 {
				line = line[:200]
			}
			issues = append(issues, line)
		}

		if len(issues) >= maxIssuesToReport {
			break
		}
	}

	return issues
}
