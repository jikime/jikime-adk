package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// PostToolFailureCmd represents the post-tool-failure hook command
var PostToolFailureCmd = &cobra.Command{
	Use:   "post-tool-failure",
	Short: "Log tool execution failures for diagnostics",
	Long: `PostToolUseFailure hook that logs tool execution failures.
Provides diagnostic information when a tool call fails during execution.`,
	RunE: runPostToolFailure,
}

type postToolFailureInput struct {
	SessionID   string `json:"session_id"`
	ToolName    string `json:"tool_name"`
	ToolUseID   string `json:"tool_use_id"`
	Error       string `json:"error"`
	IsInterrupt bool   `json:"is_interrupt"`
}

func runPostToolFailure(cmd *cobra.Command, args []string) error {
	var input postToolFailureInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// Log to stderr for diagnostics (not visible to Claude, but useful for debugging)
	if input.Error != "" {
		fmt.Fprintf(os.Stderr, "[jikime] tool failure: %s | error: %s | session: %s\n",
			input.ToolName, input.Error, input.SessionID)
	}

	// Return continue:true — let Claude Code handle failure recovery
	response := HookResponse{Continue: true}
	return writeResponse(response)
}
