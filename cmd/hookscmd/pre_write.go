package hookscmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// PreWriteCmd represents the pre-write hook command
var PreWriteCmd = &cobra.Command{
	Use:   "pre-write",
	Short: "Block unnecessary documentation file creation",
	Long: `PreToolUse hook that blocks creation of unnecessary documentation files.

Features:
- Block creation of random .md/.txt files
- Allow essential docs (README.md, CLAUDE.md, AGENTS.md, CONTRIBUTING.md)
- Keep documentation consolidated

Based on everything-claude-code patterns.`,
	RunE: runPreWrite,
}

// Patterns for allowed documentation files
var allowedDocPatterns = []string{
	`README\.md$`,
	`CLAUDE\.md$`,
	`AGENTS\.md$`,
	`CONTRIBUTING\.md$`,
	`CHANGELOG\.md$`,
	`LICENSE(\.md)?$`,
	`CODE_OF_CONDUCT\.md$`,
	`SECURITY\.md$`,
	`SKILL\.md$`,
	`docs/.*\.md$`,
	`\.jikime/.*\.md$`,
	`\.claude/.*\.md$`,
}

var allowedDocCompiled []*regexp.Regexp

func init() {
	allowedDocCompiled = compileWritePatterns(allowedDocPatterns)
}

func compileWritePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile("(?i)" + p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

type preWriteInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	} `json:"tool_input"`
}

type preWriteOutput struct {
	HookSpecificOutput *preWriteHookOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool                `json:"suppressOutput,omitempty"`
}

type preWriteHookOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
}

func runPreWrite(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input preWriteInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return nil // Invalid JSON - allow by default
	}

	// Only process Write tool
	if input.ToolName != "Write" {
		return nil
	}

	filePath := input.ToolInput.FilePath
	if filePath == "" {
		return nil
	}

	// Normalize path
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
	ext := strings.ToLower(filepath.Ext(filePath))

	// Only check .md and .txt files
	if ext != ".md" && ext != ".txt" {
		output := preWriteOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(output)
	}

	// Check if it's an allowed documentation file
	for _, pattern := range allowedDocCompiled {
		if pattern.MatchString(normalizedPath) {
			output := preWriteOutput{SuppressOutput: true}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetEscapeHTML(false)
			return encoder.Encode(output)
		}
	}

	// Block creation of random documentation files
	output := preWriteOutput{
		HookSpecificOutput: &preWriteHookOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "deny",
			PermissionDecisionReason: strings.Join([]string{
				"[Hook] BLOCKED: Unnecessary documentation file creation",
				"[Hook] File: " + filePath,
				"[Hook] Use README.md for documentation instead",
				"[Hook] Allowed patterns: README.md, CLAUDE.md, CHANGELOG.md, docs/*.md, .jikime/*.md",
			}, "\n"),
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
