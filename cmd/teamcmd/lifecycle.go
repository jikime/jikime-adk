package teamcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newLifecycleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lifecycle",
		Short: "Agent lifecycle control (shutdown, idle)",
	}
	cmd.AddCommand(newLifecycleIdleCmd())
	cmd.AddCommand(newLifecycleOnExitCmd())
	cmd.AddCommand(newLifecycleShutdownCmd())
	cmd.AddCommand(newLifecycleRequestShutdownCmd())
	cmd.AddCommand(newLifecycleApproveShutdownCmd())
	cmd.AddCommand(newLifecycleRejectShutdownCmd())
	return cmd
}

func newLifecycleIdleCmd() *cobra.Command {
	var (
		agentID  string
		teamName string
		lastTask string
	)

	cmd := &cobra.Command{
		Use:   "idle",
		Short: "Send idle notification to leader",
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if teamName == "" {
				teamName = os.Getenv("JIKIME_TEAM_NAME")
			}
			if agentID == "" || teamName == "" {
				return fmt.Errorf("--agent and --team required (or set JIKIME_AGENT_ID / JIKIME_TEAM_NAME)")
			}

			ti := team.NewTeamInbox(teamDir(teamName))
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     agentID,
				To:       "leader",
				Subject:  "idle",
				Body:     fmt.Sprintf("agent %s is idle; last_task=%s", agentID, lastTask),
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Idle notification sent (last_task: %s)\n", lastTask)
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Team name")
	cmd.Flags().StringVar(&lastTask, "last-task", "", "Last completed task ID")
	return cmd
}

func newLifecycleOnExitCmd() *cobra.Command {
	var (
		agentID  string
		teamName string
	)

	cmd := &cobra.Command{
		Use:   "on-exit",
		Short: "Handle agent process exit (release tasks, mark offline)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if teamName == "" {
				teamName = os.Getenv("JIKIME_TEAM_NAME")
			}
			if agentID == "" || teamName == "" {
				return nil // Silently exit if not in team context
			}

			td := teamDir(teamName)

			// Release any in_progress tasks held by this agent.
			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err == nil {
				tasks, _ := store.List(team.TaskStatusInProgress, agentID)
				for _, t := range tasks {
					_, _ = store.Release(t.ID)
				}
			}

			// Mark agent offline.
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err == nil {
				_ = reg.MarkDead(agentID)
			}

			fmt.Printf("on-exit: agent %s cleaned up\n", agentID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Team name")
	return cmd
}

func newLifecycleShutdownCmd() *cobra.Command {
	var (
		agentID  string
		teamName string
		reason   string
	)

	cmd := &cobra.Command{
		Use:   "shutdown",
		Short: "Request graceful shutdown of an agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if teamName == "" {
				teamName = os.Getenv("JIKIME_TEAM_NAME")
			}
			if agentID == "" || teamName == "" {
				return fmt.Errorf("--agent and --team required")
			}
			ti := team.NewTeamInbox(teamDir(teamName))
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     "orchestrator",
				To:       agentID,
				Subject:  "shutdown_request",
				Body:     reason,
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Shutdown request sent to %s\n", agentID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Target agent ID")
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Team name")
	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Shutdown reason")
	return cmd
}

// newLifecycleRequestShutdownCmd sends a shutdown request with a unique requestId.
// The target agent should respond with approve-shutdown or reject-shutdown.
func newLifecycleRequestShutdownCmd() *cobra.Command {
	var (
		fromAgent string
		toAgent   string
		teamName  string
		reason    string
	)
	cmd := &cobra.Command{
		Use:   "request-shutdown <team-name> <from-agent> <to-agent>",
		Short: "Send a shutdown request to an agent (3-step protocol)",
		Long: `Send a shutdown request that the target agent must explicitly approve or reject.
The returned requestId must be passed to approve-shutdown or reject-shutdown.

Example:
  jikime team lifecycle request-shutdown my-team leader worker-1 --reason "all tasks done"`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName = args[0]
			fromAgent = args[1]
			toAgent = args[2]

			requestID := uuid.New().String()[:8]
			ti := team.NewTeamInbox(teamDir(teamName))
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     fromAgent,
				To:       toAgent,
				Subject:  "shutdown_request",
				Body:     fmt.Sprintf("requestId=%s reason=%s", requestID, reason),
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Shutdown request sent to %s (requestId: %s)\n", toAgent, requestID)
			fmt.Printf("   Use: jikime team lifecycle approve-shutdown %s %s %s\n", teamName, requestID, toAgent)
			return nil
		},
	}
	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Shutdown reason")
	return cmd
}

// newLifecycleApproveShutdownCmd notifies the requester that the agent agrees to shut down.
func newLifecycleApproveShutdownCmd() *cobra.Command {
	var (
		teamName  string
		requestID string
		agentName string
		feedback  string
	)
	cmd := &cobra.Command{
		Use:   "approve-shutdown <team-name> <request-id> <agent>",
		Short: "Approve a shutdown request (agent agrees to shut down)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName = args[0]
			requestID = args[1]
			agentName = args[2]

			if teamName == "" {
				teamName = os.Getenv("JIKIME_TEAM_NAME")
			}

			ti := team.NewTeamInbox(teamDir(teamName))
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     agentName,
				To:       "leader",
				Subject:  "shutdown_approved",
				Body:     fmt.Sprintf("requestId=%s agent=%s feedback=%s", requestID, agentName, feedback),
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Shutdown approved: agent=%s requestId=%s\n", agentName, requestID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&feedback, "feedback", "f", "", "Optional feedback")
	return cmd
}

// newLifecycleRejectShutdownCmd notifies the requester that the agent refuses to shut down.
func newLifecycleRejectShutdownCmd() *cobra.Command {
	var (
		teamName  string
		requestID string
		agentName string
		reason    string
	)
	cmd := &cobra.Command{
		Use:   "reject-shutdown <team-name> <request-id> <agent>",
		Short: "Reject a shutdown request (agent refuses to shut down)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName = args[0]
			requestID = args[1]
			agentName = args[2]

			if teamName == "" {
				teamName = os.Getenv("JIKIME_TEAM_NAME")
			}

			ti := team.NewTeamInbox(teamDir(teamName))
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     agentName,
				To:       "leader",
				Subject:  "shutdown_rejected",
				Body:     fmt.Sprintf("requestId=%s agent=%s reason=%s", requestID, agentName, reason),
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Shutdown rejected: agent=%s requestId=%s reason=%s\n", agentName, requestID, reason)
			return nil
		},
	}
	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Rejection reason")
	return cmd
}
