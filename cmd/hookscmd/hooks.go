package hookscmd

import "github.com/spf13/cobra"

// NewHooks creates the hooks command
func NewHooks() *cobra.Command {
	hooks := &cobra.Command{
		Use:   "hooks",
		Short: "Hook commands for Claude Code integration",
		Long:  `Hook commands that integrate with Claude Code lifecycle events`,
	}

	// Add subcommands - SessionStart/End hooks
	hooks.AddCommand(SessionStartCmd)
	hooks.AddCommand(SessionEndRankCmd)
	hooks.AddCommand(SessionEndCleanupCmd)

	// PreCompact hooks (from everything-claude-code)
	hooks.AddCommand(PreCompactCmd)

	// PreToolUse hooks
	hooks.AddCommand(PreToolSecurityCmd)
	hooks.AddCommand(PreBashCmd)  // from everything-claude-code
	hooks.AddCommand(PreWriteCmd) // from everything-claude-code

	// PostToolUse hooks
	hooks.AddCommand(PostToolFormatterCmd)
	hooks.AddCommand(PostToolLinterCmd)
	hooks.AddCommand(PostToolAstGrepCmd)
	hooks.AddCommand(PostToolLspCmd)
	hooks.AddCommand(PostBashCmd) // from everything-claude-code

	// Stop hooks
	hooks.AddCommand(StopLoopCmd)
	hooks.AddCommand(StopAuditCmd) // from everything-claude-code

	return hooks
}
