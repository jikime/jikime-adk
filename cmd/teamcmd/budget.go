package teamcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newBudgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage team token budget",
	}
	cmd.AddCommand(newBudgetShowCmd())
	cmd.AddCommand(newBudgetSetCmd())
	cmd.AddCommand(newBudgetReportCmd())
	return cmd
}

func newBudgetShowCmd() *cobra.Command {
	var agentID string
	cmd := &cobra.Command{
		Use:   "show <team-name>",
		Short: "Show current token usage and budget",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			// Load config for budget
			cfg := struct {
				Budget int `json:"budget"`
			}{}
			data, _ := os.ReadFile(filepath.Join(td, "config.json"))
			_ = json.Unmarshal(data, &cfg)

			costStore, err := team.NewCostStore(filepath.Join(td, "costs"), cfg.Budget)
			if err != nil {
				return fmt.Errorf("cost store: %w", err)
			}

			if agentID != "" {
				// Show single agent
				summary, err := costStore.Summary(agentID)
				if err != nil {
					return err
				}
				printBudgetSummary(name, agentID, summary, cfg.Budget)
				return nil
			}

			// Show all agents
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}
			agents, _ := reg.List()

			var totalTokens int
			fmt.Printf("Budget for team %q (limit: %d tokens)\n\n", name, cfg.Budget)
			fmt.Printf("  %-14s  %-10s  %-10s  %-10s  %s\n",
				"AGENT", "INPUT", "OUTPUT", "TOTAL", "BUDGET%")

			for _, a := range agents {
				summary, _ := costStore.Summary(a.ID)
				if summary == nil {
					fmt.Printf("  %-14s  %-10d  %-10d  %-10d  %.1f%%\n",
						a.ID, 0, 0, 0, 0.0)
					continue
				}
				totalTokens += summary.TotalTokens
				pct := 0.0
				if cfg.Budget > 0 {
					pct = float64(summary.TotalTokens) / float64(cfg.Budget) * 100
				}
				fmt.Printf("  %-14s  %-10d  %-10d  %-10d  %.1f%%\n",
					a.ID, summary.TotalInputTokens, summary.TotalOutputTokens,
					summary.TotalTokens, pct)
			}

			fmt.Printf("\n  %-14s  %s\n", "TOTAL", fmt.Sprintf("%d tokens", totalTokens))
			if cfg.Budget > 0 {
				pct := float64(totalTokens) / float64(cfg.Budget) * 100
				exceeded := ""
				if exceeded2, _ := costStore.BudgetExceeded(); exceeded2 {
					exceeded = "  ⚠️  BUDGET EXCEEDED"
				}
				fmt.Printf("  Budget used: %.1f%% (%d / %d)%s\n",
					pct, totalTokens, cfg.Budget, exceeded)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Show budget for specific agent only")
	return cmd
}

func newBudgetSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <team-name> <tokens>",
		Short: "Set token budget for a team",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)
			cfgPath := filepath.Join(td, "config.json")

			// Read existing config
			data, err := os.ReadFile(cfgPath)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}
			var cfg map[string]interface{}
			if err := json.Unmarshal(data, &cfg); err != nil {
				return fmt.Errorf("parse config: %w", err)
			}

			var budget int
			if _, err := fmt.Sscanf(args[1], "%d", &budget); err != nil {
				return fmt.Errorf("invalid budget value: %s", args[1])
			}
			cfg["budget"] = budget

			if err := writeJSON(cfgPath, cfg); err != nil {
				return err
			}
			fmt.Printf("✅ Budget for team %q set to %d tokens\n", name, budget)
			return nil
		},
	}
}

func newBudgetReportCmd() *cobra.Command {
	var (
		agentID      string
		taskID       string
		model        string
		inputTokens  int
		outputTokens int
	)
	cmd := &cobra.Command{
		Use:   "report <team-name>",
		Short: "Report token usage for an agent (call from agent hooks)",
		Long: `Report token usage to the team cost store.
Agents call this after each interaction to track cumulative token consumption.

Example (from a Claude Code hook):
  jikime team budget report $JIKIME_TEAM_NAME \
    --agent $JIKIME_AGENT_ID \
    --input-tokens 1234 --output-tokens 567 \
    --model claude-sonnet-4-6`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}

			name := args[0]
			td := teamDir(name)

			cfg := struct {
				Budget int `json:"budget"`
			}{}
			data, _ := os.ReadFile(filepath.Join(td, "config.json"))
			_ = json.Unmarshal(data, &cfg)

			costStore, err := team.NewCostStore(filepath.Join(td, "costs"), cfg.Budget)
			if err != nil {
				return fmt.Errorf("cost store: %w", err)
			}

			ev, err := costStore.Record(agentID, taskID, "", model, inputTokens, outputTokens)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Reported: agent=%s in=%d out=%d total=%d (id: %s)\n",
				agentID, ev.InputTokens, ev.OutputTokens, ev.TotalTokens, ev.ID[:8])

			if cfg.Budget > 0 {
				if exceeded, _ := costStore.BudgetExceeded(); exceeded {
					fmt.Printf("⚠️  Budget exceeded! Notify leader.\n")
				} else {
					s, _ := costStore.Summary("")
					fmt.Printf("   Budget used: %.1f%% (%d / %d)\n",
						s.BudgetUsedPercent, s.TotalTokens, cfg.Budget)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID (default: JIKIME_AGENT_ID)")
	cmd.Flags().StringVar(&taskID, "task", "", "Task ID this cost belongs to")
	cmd.Flags().StringVar(&model, "model", "", "Model name (e.g. claude-sonnet-4-6)")
	cmd.Flags().IntVar(&inputTokens, "input-tokens", 0, "Input tokens consumed")
	cmd.Flags().IntVar(&outputTokens, "output-tokens", 0, "Output tokens consumed")
	return cmd
}

func printBudgetSummary(teamName, agentID string, summary *team.CostSummary, budget int) {
	fmt.Printf("Budget summary — team:%s  agent:%s\n\n", teamName, agentID)
	fmt.Printf("  Input tokens:   %d\n", summary.TotalInputTokens)
	fmt.Printf("  Output tokens:  %d\n", summary.TotalOutputTokens)
	fmt.Printf("  Total tokens:   %d\n", summary.TotalTokens)
	if budget > 0 {
		fmt.Printf("  Budget limit:   %d\n", budget)
		fmt.Printf("  Budget used:    %.1f%%\n", summary.BudgetUsedPercent)
	}
	fmt.Printf("  Events:         %d\n", summary.EventCount)
}
