package worktreecmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"jikime-adk/worktree"
)

func newSyncCmd() *cobra.Command {
	var (
		base        string
		rebase      bool
		ffOnly      bool
		squash      bool
		syncAll     bool
		autoResolve bool
	)

	cmd := &cobra.Command{
		Use:   "sync [spec-id]",
		Short: "Sync worktree with base branch",
		Long: `Sync worktree with base branch.

Fetches latest changes from the base branch and merges them into
the worktree. Supports merge, rebase, squash, and fast-forward strategies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !syncAll {
				color.Red("✗ Either SPEC_ID or --all option is required")
				return fmt.Errorf("spec_id or --all required")
			}

			if len(args) > 0 && syncAll {
				color.Red("✗ Cannot use both SPEC_ID and --all option")
				return fmt.Errorf("cannot use both spec_id and --all")
			}

			manager, err := getManager()
			if err != nil {
				return err
			}

			if syncAll {
				worktrees := manager.List()
				if len(worktrees) == 0 {
					color.Yellow("No worktrees found to sync")
					return nil
				}

				color.Cyan("Syncing %d worktrees...", len(worktrees))

				successCount := 0
				conflictCount := 0

				for _, info := range worktrees {
					err := manager.Sync(info.SpecID, base, rebase, ffOnly, squash, autoResolve)
					if err != nil {
						if _, ok := err.(*worktree.MergeConflictError); ok {
							color.Red("✗ %s (conflicts)", info.SpecID)
							conflictCount++
						} else {
							color.Red("✗ %s (failed: %v)", info.SpecID, err)
							conflictCount++
						}
					} else {
						method := "merge"
						if rebase {
							method = "rebase"
						} else if ffOnly {
							method = "fast-forward"
						} else if squash {
							method = "squash"
						}
						color.Green("✓ %s (%s)", info.SpecID, method)
						successCount++
					}
				}

				fmt.Println()
				color.Green("Summary: %d synced, %d failed", successCount, conflictCount)
			} else {
				specID := args[0]
				if err := manager.Sync(specID, base, rebase, ffOnly, squash, autoResolve); err != nil {
					color.Red("✗ %v", err)
					return err
				}

				method := "merge"
				if rebase {
					method = "rebase"
				} else if ffOnly {
					method = "fast-forward"
				} else if squash {
					method = "squash"
				}
				color.Green("✓ Worktree synced: %s (%s)", specID, method)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&base, "base", "main", "Base branch to sync from")
	cmd.Flags().BoolVar(&rebase, "rebase", false, "Use rebase instead of merge")
	cmd.Flags().BoolVar(&ffOnly, "ff-only", false, "Only sync if fast-forward is possible")
	cmd.Flags().BoolVar(&squash, "squash", false, "Squash all commits into a single commit")
	cmd.Flags().BoolVar(&syncAll, "all", false, "Sync all worktrees")
	cmd.Flags().BoolVar(&autoResolve, "auto-resolve", false, "Automatically resolve conflicts")

	return cmd
}
