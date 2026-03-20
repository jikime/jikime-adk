package teamcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newTasksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage team tasks",
	}
	cmd.AddCommand(newTaskCreateCmd())
	cmd.AddCommand(newTaskGetCmd())
	cmd.AddCommand(newTaskUpdateCmd())
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskWaitCmd())
	cmd.AddCommand(newTaskClaimCmd())
	cmd.AddCommand(newTaskCompleteCmd())
	return cmd
}

func newTaskCreateCmd() *cobra.Command {
	var (
		desc      string
		dod       string
		dependsOn string
		priority  int
		tags      string
		owner     string
	)

	cmd := &cobra.Command{
		Use:   "create <team-name> <title>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := team.NewStore(filepath.Join(teamDir(args[0]), "tasks"))
			if err != nil {
				return err
			}
			var deps []string
			if dependsOn != "" {
				deps = strings.Split(dependsOn, ",")
			}
			var tagList []string
			if tags != "" {
				tagList = strings.Split(tags, ",")
			}
			t, err := store.Create(args[1], desc, dod, deps, priority, tagList, owner)
			if err != nil {
				return err
			}
			fmt.Printf("✅ Task created: %s\n", t.ID[:8])
			fmt.Printf("   title:  %s\n", t.Title)
			if t.Owner != "" {
				fmt.Printf("   owner:  %s\n", t.Owner)
			}
			fmt.Printf("   status: %s\n", t.Status)
			return nil
		},
	}
	cmd.Flags().StringVarP(&desc, "desc", "d", "", "Task description")
	cmd.Flags().StringVar(&dod, "dod", "", "Definition of Done")
	cmd.Flags().StringVar(&dependsOn, "depends-on", "", "Comma-separated task IDs this depends on")
	cmd.Flags().IntVarP(&priority, "priority", "p", 0, "Priority (higher = more important)")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Pre-assign task to a specific agent ID")
	return cmd
}

func newTaskGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-name> <task-id>",
		Short: "Get task details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := team.NewStore(filepath.Join(teamDir(args[0]), "tasks"))
			if err != nil {
				return err
			}
			t, err := store.Get(args[1])
			if err != nil {
				return err
			}
			if t == nil {
				return fmt.Errorf("task %s not found", args[1])
			}
			fmt.Printf("ID:      %s\n", t.ID)
			fmt.Printf("Title:   %s\n", t.Title)
			fmt.Printf("Status:  %s\n", t.Status)
			fmt.Printf("Agent:   %s\n", t.AgentID)
			if t.Description != "" {
				fmt.Printf("Desc:    %s\n", t.Description)
			}
			if t.DoD != "" {
				fmt.Printf("DoD:     %s\n", t.DoD)
			}
			if len(t.DependsOn) > 0 {
				fmt.Printf("Depends: %s\n", strings.Join(t.DependsOn, ", "))
			}
			return nil
		},
	}
}

func newTaskUpdateCmd() *cobra.Command {
	var (
		title   string
		desc    string
		dod     string
		prio    int
		status  string
		agentID string
		result  string
	)
	cmd := &cobra.Command{
		Use:   "update <team-name> <task-id>",
		Short: "Update task status or metadata",
		Long: `Update a task's status and/or metadata.

Status transitions (--status):
  pending     Release back to the queue (admin override, no ownership check)
  in_progress Claim the task for --agent
  done        Force-complete the task (admin override, no ownership check)
  blocked     Mark the task as blocked
  failed      Mark the task as failed

Examples:
  jikime team tasks update my-team abc123 --status done --agent worker-1
  jikime team tasks update my-team abc123 --status pending
  jikime team tasks update my-team abc123 --title "New title" --priority 2`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName, taskID := args[0], args[1]
			store, err := team.NewStore(filepath.Join(teamDir(teamName), "tasks"))
			if err != nil {
				return err
			}

			// Handle status transitions.
			switch status {
			case "pending":
				// Release back to queue — works even if task has no owner (admin use).
				t, err := store.Get(taskID)
				if err != nil {
					return err
				}
				if t == nil {
					return fmt.Errorf("task %s not found", taskID)
				}
				t.Status = team.TaskStatusPending
				t.AgentID = ""
				t.UpdatedAt = time.Now()
				if err := forceUpdateTask(store, t); err != nil {
					return err
				}
				fmt.Printf("✅ Task %s → pending\n", taskID[:8])
				return nil

			case "in_progress":
				if agentID == "" {
					agentID = os.Getenv("JIKIME_AGENT_ID")
				}
				if agentID == "" {
					return fmt.Errorf("--agent required for --status in_progress")
				}
				if _, err := store.Claim(taskID, agentID); err != nil {
					return err
				}
				fmt.Printf("✅ Task %s → in_progress (agent: %s)\n", taskID[:8], agentID)
				return nil

			case "done":
				// Admin force-complete: bypass ownership check.
				t, err := store.Get(taskID)
				if err != nil {
					return err
				}
				if t == nil {
					return fmt.Errorf("task %s not found", taskID)
				}
				t.Status = team.TaskStatusDone
				t.Result = result
				t.UpdatedAt = time.Now()
				if agentID != "" {
					t.AgentID = agentID
				}
				if err := forceUpdateTask(store, t); err != nil {
					return err
				}
				fmt.Printf("✅ Task %s → done\n", taskID[:8])
				return nil

			case "blocked":
				t, err := store.Get(taskID)
				if err != nil {
					return err
				}
				if t == nil {
					return fmt.Errorf("task %s not found", taskID)
				}
				t.Status = team.TaskStatusBlocked
				t.UpdatedAt = time.Now()
				if err := forceUpdateTask(store, t); err != nil {
					return err
				}
				fmt.Printf("✅ Task %s → blocked\n", taskID[:8])
				return nil

			case "failed":
				t, err := store.Get(taskID)
				if err != nil {
					return err
				}
				if t == nil {
					return fmt.Errorf("task %s not found", taskID)
				}
				t.Status = team.TaskStatusFailed
				t.Result = result
				t.UpdatedAt = time.Now()
				if err := forceUpdateTask(store, t); err != nil {
					return err
				}
				fmt.Printf("✅ Task %s → failed\n", taskID[:8])
				return nil

			case "":
				// No status change — update metadata only.

			default:
				return fmt.Errorf("unknown status %q (pending|in_progress|done|blocked|failed)", status)
			}

			// Metadata-only update.
			t, err := store.Update(taskID, title, desc, dod, prio)
			if err != nil {
				return err
			}
			fmt.Printf("✅ Task %s updated\n", t.ID[:8])
			return nil
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "", "New title")
	cmd.Flags().StringVarP(&desc, "desc", "d", "", "New description")
	cmd.Flags().StringVar(&dod, "dod", "", "New definition of done")
	cmd.Flags().IntVarP(&prio, "priority", "p", 0, "New priority")
	cmd.Flags().StringVarP(&status, "status", "s", "", "New status (pending|in_progress|done|blocked|failed)")
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID (required for in_progress; optional for done)")
	cmd.Flags().StringVarP(&result, "result", "r", "", "Result summary (for done/failed)")
	return cmd
}

// forceUpdateTask persists a task bypassing ownership/transition checks.
func forceUpdateTask(store *team.Store, t *team.Task) error {
	return store.ForceStatus(t)
}

func newTaskListCmd() *cobra.Command {
	var (
		status  string
		agentID string
		owner   string
	)
	cmd := &cobra.Command{
		Use:   "list <team-name>",
		Short: "List team tasks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := team.NewStore(filepath.Join(teamDir(args[0]), "tasks"))
			if err != nil {
				return err
			}
			tasks, err := store.List(team.TaskStatus(status), agentID, owner)
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Println("No tasks.")
				return nil
			}
			for _, t := range tasks {
				id := t.ID
				if len(id) > 8 {
					id = id[:8]
				}
				agent := t.AgentID
				if agent == "" {
					agent = "-"
				}
				ownerLabel := ""
				if t.Owner != "" {
					ownerLabel = "  owner:" + t.Owner
				}
				fmt.Printf("  %s  [%-11s]  %-30s  agent:%s%s\n",
					id, t.Status, t.Title, agent, ownerLabel)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status")
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Filter by agent ID")
	cmd.Flags().StringVarP(&owner, "owner", "o", "", "Filter by pre-assigned owner")
	return cmd
}

func newTaskWaitCmd() *cobra.Command {
	var (
		timeout  int
		interval int
	)
	cmd := &cobra.Command{
		Use:   "wait <team-name>",
		Short: "Wait until all tasks are completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err != nil {
				return err
			}
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}

			cb := team.WaiterCallbacks{
				OnProgress: func(r team.WaitResult) {
					fmt.Printf("\r  tasks: %d/%d done | wip:%d pending:%d blocked:%d",
						r.Done, r.Total, r.InProgress, r.Pending, r.Blocked)
				},
				OnAgentDead: func(agentID string, taskIDs []string) {
					fmt.Printf("\n  ⚠️  agent %s died; released tasks: %v\n", agentID, taskIDs)
				},
			}

			dur := time.Duration(interval) * time.Second
			w := team.NewWaiter(store, reg, nil, name, dur, cb)

			ctx := context.Background()
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
				defer cancel()
			}

			result, err := w.Wait(ctx)
			fmt.Println()
			if err != nil {
				return err
			}
			fmt.Printf("\n  status: %s  elapsed: %s\n", result.Status, result.Elapsed.Round(time.Second))
			return nil
		},
	}
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Max wait time in seconds (0 = no limit)")
	cmd.Flags().IntVarP(&interval, "interval", "i", 5, "Poll interval in seconds")
	return cmd
}

func newTaskClaimCmd() *cobra.Command {
	var agentID string
	cmd := &cobra.Command{
		Use:   "claim <team-name> <task-id>",
		Short: "Claim a task for the current agent",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			store, err := team.NewStore(filepath.Join(teamDir(args[0]), "tasks"))
			if err != nil {
				return err
			}
			t, err := store.Claim(args[1], agentID)
			if err != nil {
				return err
			}
			fmt.Printf("✅ Task %s claimed by %s\n", t.ID[:8], agentID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID claiming the task")
	return cmd
}

func newTaskCompleteCmd() *cobra.Command {
	var (
		agentID string
		result  string
	)
	cmd := &cobra.Command{
		Use:   "complete <team-name> <task-id>",
		Short: "Mark a task as completed",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			store, err := team.NewStore(filepath.Join(teamDir(args[0]), "tasks"))
			if err != nil {
				return err
			}
			t, err := store.Complete(args[1], agentID, result)
			if err != nil {
				return err
			}
			fmt.Printf("✅ Task %s completed\n", t.ID[:8])
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID that completed the task")
	cmd.Flags().StringVarP(&result, "result", "r", "", "Result summary")
	return cmd
}
