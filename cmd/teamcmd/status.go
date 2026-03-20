package teamcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newStatusCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "status <team-name>",
		Short: "Show team status (agents, tasks, cost)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			// Load config
			cfgData, err := os.ReadFile(filepath.Join(td, "config.json"))
			if err != nil {
				return fmt.Errorf("team %q not found (run `jikime team create` first)", name)
			}
			var cfg team.TeamConfig
			_ = json.Unmarshal(cfgData, &cfg)

			// Load agents
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}
			agents, _ := reg.List()

			// Load tasks
			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err != nil {
				return err
			}
			tasks, _ := store.List("", "")

			// Load cost summary
			costStore, err := team.NewCostStore(filepath.Join(td, "costs"), cfg.Budget)
			if err != nil {
				return err
			}
			summary, _ := costStore.Summary("")

			if jsonOut {
				out := map[string]interface{}{
					"config":  cfg,
					"agents":  agents,
					"tasks":   tasks,
					"cost":    summary,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			// Human-readable output
			fmt.Printf("Team: %s\n", name)
			fmt.Printf("Dir:  %s\n\n", td)

			fmt.Printf("Agents (%d):\n", len(agents))
			for _, a := range agents {
				alive, _ := reg.IsAlive(a.ID)
				liveness := "✅"
				if !alive {
					liveness = "❌"
				}
				fmt.Printf("  %s %s [%s] task:%s\n", liveness, a.ID, a.Role, a.CurrentTaskID)
			}

			// Task summary
			var todo, wip, done, blocked int
			for _, t := range tasks {
				switch t.Status {
				case team.TaskStatusPending:
					todo++
				case team.TaskStatusInProgress:
					wip++
				case team.TaskStatusDone:
					done++
				case team.TaskStatusBlocked:
					blocked++
				}
			}
			fmt.Printf("\nTasks (%d): todo=%d wip=%d done=%d blocked=%d\n",
				len(tasks), todo, wip, done, blocked)

			if summary != nil {
				fmt.Printf("\nTokens: %d", summary.TotalTokens)
				if cfg.Budget > 0 {
					fmt.Printf(" / %d (%.1f%%)", cfg.Budget, summary.BudgetUsedPercent)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
