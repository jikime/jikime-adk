package hookscmd

import (
	"github.com/spf13/cobra"
	"jikime-adk-v2/internal/hooks"
)

// SessionEndRankCmd represents the session-end-rank hook command
var SessionEndRankCmd = &cobra.Command{
	Use:   "session-end-rank",
	Short: "Submit session token usage (placeholder)",
	Long: `SessionEnd hook that submits token usage to rank service.
Currently a placeholder - will be implemented when rank service is available.`,
	RunE: runSessionEndRank,
}

func runSessionEndRank(cmd *cobra.Command, args []string) error {
	// Read input from stdin
	_, _ = hooks.ReadInput()

	// Currently a no-op placeholder
	// When rank service is available, implement token submission here

	response := hooks.SuccessResponse("")
	return hooks.WriteResponse(response)
}
