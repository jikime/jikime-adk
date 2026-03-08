package hookscmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// PermissionRequestCmd represents the permission-request hook command
var PermissionRequestCmd = &cobra.Command{
	Use:   "permission-request",
	Short: "Handle permission requests with policy-based decisions",
	Long: `PermissionRequest hook that evaluates tool permission requests.
Returns allow/deny/ask based on configured policy rules.
Default: defer to user settings ("ask").`,
	RunE: runPermissionRequest,
}

type permissionRequestInput struct {
	SessionID string `json:"session_id"`
	ToolName  string `json:"tool_name"`
	ToolInput any    `json:"tool_input"`
}

type permissionRequestOutput struct {
	Continue           bool                   `json:"continue"`
	HookSpecificOutput *permissionSpecific    `json:"hookSpecificOutput,omitempty"`
}

type permissionSpecific struct {
	HookEventName      string `json:"hookEventName"`
	PermissionDecision string `json:"permissionDecision"`
}

func runPermissionRequest(cmd *cobra.Command, args []string) error {
	var input permissionRequestInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// Default policy: defer to user/settings.json configuration
	// Claude Code v2.1.59+: hookSpecificOutput.hookEventName must be "PermissionRequest"
	out := permissionRequestOutput{
		Continue: true,
		HookSpecificOutput: &permissionSpecific{
			HookEventName:      "PermissionRequest",
			PermissionDecision: "ask",
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(out)
}
