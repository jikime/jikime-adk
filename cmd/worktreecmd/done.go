package worktreecmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newDoneCmd() *cobra.Command {
	var (
		base  string
		push  bool
		force bool
	)

	cmd := &cobra.Command{
		Use:   "done <spec-id>",
		Short: "Complete worktree: merge to main and cleanup",
		Long: `Complete worktree workflow: merge to main and cleanup.

This command performs the full completion workflow:
1. Checkout base branch (main)
2. Merge worktree branch into base
3. Remove worktree
4. Delete feature branch`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specID := args[0]

			manager, err := getManager()
			if err != nil {
				return err
			}

			// Get worktree info for display
			info := manager.Registry.Get(specID, manager.ProjectName)
			if info == nil {
				color.Red("✗ Worktree not found: %s", specID)
				return fmt.Errorf("worktree not found: %s", specID)
			}

			color.Cyan("Completing worktree: %s", specID)
			fmt.Printf("  Branch: %s\n", info.Branch)
			fmt.Printf("  Merging into: %s\n", base)
			fmt.Println()

			result, err := manager.Done(specID, base, push, force)
			if err != nil {
				color.Red("✗ %v", err)
				return err
			}

			color.Green("✓ Worktree completed successfully")
			fmt.Printf("  Merged: %s → %s\n", result.MergedBranch, result.BaseBranch)
			if result.Pushed {
				fmt.Printf("  Pushed: origin/%s\n", result.BaseBranch)
			}
			fmt.Println()
			color.Yellow("Branch cleanup:")
			fmt.Printf("  - Worktree removed: %s\n", specID)
			fmt.Printf("  - Branch deleted: %s\n", result.MergedBranch)

			return nil
		},
	}

	cmd.Flags().StringVar(&base, "base", "main", "Base branch to merge into")
	cmd.Flags().BoolVar(&push, "push", false, "Push to remote after merge")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force remove with uncommitted changes")

	return cmd
}
