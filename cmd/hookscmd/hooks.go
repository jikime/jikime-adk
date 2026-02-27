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
	hooks.AddCommand(SessionEndCleanupCmd)

	// UserPromptSubmit hooks
	hooks.AddCommand(UserPromptSubmitCmd)

	// PreCompact hooks
	hooks.AddCommand(PreCompactCmd)

	// PreToolUse hooks
	hooks.AddCommand(PreToolSecurityCmd)
	hooks.AddCommand(PreWriteCmd)

	// PostToolUse hooks
	hooks.AddCommand(PostToolFormatterCmd)
	hooks.AddCommand(PostToolLinterCmd)
	hooks.AddCommand(PostToolAstGrepCmd)
	hooks.AddCommand(PostToolLspCmd)
	hooks.AddCommand(PostBashCmd)

	// Stop hooks
	hooks.AddCommand(StopLoopCmd)
	hooks.AddCommand(StopAuditCmd)

	// Loop control hooks
	hooks.AddCommand(StartLoopCmd)
	hooks.AddCommand(CancelLoopCmd)

	// Orchestrator routing hooks
	hooks.AddCommand(OrchestratorRouteCmd)

	// Agent Teams hooks
	hooks.AddCommand(TaskCompletedCmd)  // TaskCompleted: validate SPEC acceptance criteria
	hooks.AddCommand(TeammateIdleCmd)   // TeammateIdle: validate quality gates

	return hooks
}
