package hookscmd

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// PreBashCmd represents the pre-bash hook command
var PreBashCmd = &cobra.Command{
	Use:   "pre-bash",
	Short: "Guard for dangerous bash commands",
	Long: `PreToolUse hook that blocks or warns about dangerous bash commands.

Features:
- Block dev servers outside tmux (ensures log access)
- Warn about long-running commands without tmux
- Suggest tmux usage for session persistence

Based on everything-claude-code patterns.`,
	RunE: runPreBash,
}

// Patterns that should be blocked (dev servers without tmux)
var blockDevServerPatterns = []string{
	`npm run dev`,
	`pnpm( run)? dev`,
	`yarn dev`,
	`bun run dev`,
	`next dev`,
	`vite`,
}

// Patterns that should warn about using tmux
var suggestTmuxPatterns = []string{
	`npm (install|test|run build)`,
	`pnpm (install|test|build)`,
	`yarn (install|test|build)`,
	`bun (install|test|build)`,
	`cargo (build|test)`,
	`make\b`,
	`docker\b`,
	`pytest`,
	`vitest`,
	`playwright`,
	`go test`,
	`go build`,
}

var (
	blockDevServerCompiled []*regexp.Regexp
	suggestTmuxCompiled    []*regexp.Regexp
)

func init() {
	blockDevServerCompiled = compileBashPatterns(blockDevServerPatterns)
	suggestTmuxCompiled = compileBashPatterns(suggestTmuxPatterns)
}

func compileBashPatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile("(?i)" + p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

type preBashInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

type preBashOutput struct {
	HookSpecificOutput *preBashHookOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool               `json:"suppressOutput,omitempty"`
}

type preBashHookOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
}

func runPreBash(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input preBashInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return nil // Invalid JSON - allow by default
	}

	// Only process Bash tool
	if input.ToolName != "Bash" {
		return nil
	}

	command := input.ToolInput.Command
	if command == "" {
		return nil
	}

	// Check if running in tmux
	inTmux := os.Getenv("TMUX") != ""

	var output preBashOutput

	// Check for dev server commands that should be blocked outside tmux
	if !inTmux {
		for _, pattern := range blockDevServerCompiled {
			if pattern.MatchString(command) {
				output = preBashOutput{
					HookSpecificOutput: &preBashHookOutput{
						HookEventName:      "PreToolUse",
						PermissionDecision: "deny",
						PermissionDecisionReason: strings.Join([]string{
							"[Hook] BLOCKED: Dev server must run in tmux for log access",
							"[Hook] Use: tmux new-session -d -s dev '" + command + "'",
							"[Hook] Then: tmux attach -t dev",
						}, "\n"),
					},
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetEscapeHTML(false)
				return encoder.Encode(output)
			}
		}
	}

	// Check for long-running commands that should suggest tmux
	if !inTmux {
		for _, pattern := range suggestTmuxCompiled {
			if pattern.MatchString(command) {
				output = preBashOutput{
					HookSpecificOutput: &preBashHookOutput{
						HookEventName:      "PreToolUse",
						PermissionDecision: "ask",
						PermissionDecisionReason: strings.Join([]string{
							"[Hook] Consider running in tmux for session persistence",
							"[Hook] tmux new -s dev | tmux attach -t dev",
						}, "\n"),
					},
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetEscapeHTML(false)
				return encoder.Encode(output)
			}
		}
	}

	// Allow by default
	output = preBashOutput{SuppressOutput: true}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
