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
	hooks.AddCommand(PlansWatcherCmd)    // Plans.md structure validator
	hooks.AddCommand(GuardrailEngineCmd) // Harness guardrail rules R01-R08

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

	// Lifecycle hooks (v1.0.0+)
	hooks.AddCommand(PostToolFailureCmd)  // PostToolUseFailure: log tool failures
	hooks.AddCommand(NotificationCmd)     // Notification: desktop notifications
	hooks.AddCommand(PermissionRequestCmd) // PermissionRequest: policy-based decisions
	hooks.AddCommand(SubagentStartCmd)    // SubagentStart: log subagent startup
	hooks.AddCommand(SubagentStopCmd)     // SubagentStop: log subagent completion

	// Team orchestration hooks (v1.5.x+, activated by JIKIME_TEAM_NAME)
	hooks.AddCommand(TeamAgentStartCmd)  // SessionStart: register agent in team registry
	hooks.AddCommand(TeamAgentStopCmd)   // SessionEnd: release tasks + mark offline
	hooks.AddCommand(TeamCostTrackCmd)   // PostToolUse: track tokens + enforce budget
	hooks.AddCommand(TeamPlanGateCmd)    // UserPromptSubmit: gate on plan approval

	return hooks
}
