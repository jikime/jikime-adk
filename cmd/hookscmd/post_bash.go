package hookscmd

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// PostBashCmd represents the post-bash hook command
var PostBashCmd = &cobra.Command{
	Use:   "post-bash",
	Short: "Log useful info after bash commands",
	Long: `PostToolUse hook that logs useful information after bash command execution.

Features:
- Detect PR creation and log PR URL
- Provide review command after PR creation
- Log CI/CD status hints

Based on everything-claude-code patterns.`,
	RunE: runPostBash,
}

var prURLPattern = regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/pull/\d+`)
var ghPrCreatePattern = regexp.MustCompile(`gh pr create`)

type postBashInput struct {
	ToolName   string `json:"tool_name"`
	ToolInput  struct {
		Command string `json:"command"`
	} `json:"tool_input"`
	ToolOutput struct {
		Output   string `json:"output"`
		ExitCode int    `json:"exit_code"`
	} `json:"tool_output"`
}

type postBashOutput struct {
	HookSpecificOutput *postBashHookOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool                `json:"suppressOutput,omitempty"`
}

type postBashHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func runPostBash(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input postBashInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return nil // Invalid JSON - allow by default
	}

	// Only process Bash tool
	if input.ToolName != "Bash" {
		return nil
	}

	command := input.ToolInput.Command
	output := input.ToolOutput.Output

	// Check for PR creation
	if ghPrCreatePattern.MatchString(command) && input.ToolOutput.ExitCode == 0 {
		prURL := prURLPattern.FindString(output)
		if prURL != "" {
			// Extract repo and PR number for review command
			parts := strings.Split(prURL, "/")
			if len(parts) >= 5 {
				repo := parts[3] + "/" + parts[4]
				prNum := parts[len(parts)-1]

				message := strings.Join([]string{
					"[Hook] PR created: " + prURL,
					"[Hook] Check GitHub Actions status for CI results",
					"[Hook] To review PR: gh pr review " + prNum + " --repo " + repo,
				}, "\n")

				hookOutput := postBashOutput{
					HookSpecificOutput: &postBashHookOutput{
						HookEventName:     "PostToolUse",
						AdditionalContext: message,
					},
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetEscapeHTML(false)
				return encoder.Encode(hookOutput)
			}
		}
	}

	// Check for test commands and summarize results
	testPatterns := []string{"npm test", "pnpm test", "yarn test", "pytest", "go test", "cargo test", "vitest"}
	for _, pattern := range testPatterns {
		if strings.Contains(command, pattern) {
			var message string
			if input.ToolOutput.ExitCode == 0 {
				message = "[Hook] Tests passed successfully"
			} else {
				message = "[Hook] Tests failed - review output above"
			}

			hookOutput := postBashOutput{
				HookSpecificOutput: &postBashHookOutput{
					HookEventName:     "PostToolUse",
					AdditionalContext: message,
				},
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetEscapeHTML(false)
			return encoder.Encode(hookOutput)
		}
	}

	// Suppress output for other commands
	hookOutput := postBashOutput{SuppressOutput: true}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(hookOutput)
}
