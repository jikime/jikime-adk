package worktreecmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newRecoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover worktree registry from existing directories",
		Long: `Recover worktree registry from existing directories.

Scans the worktree root directory for existing worktrees and
re-registers them in the registry file. Useful when the registry
is lost or corrupted.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := getManager()
			if err != nil {
				return err
			}

			color.Cyan("Scanning %s for worktrees...", manager.WorktreeRoot)

			recovered, err := manager.Registry.RecoverFromDisk()
			if err != nil {
				color.Red("✗ Error recovering worktrees: %v", err)
				return err
			}

			if recovered > 0 {
				color.Green("✓ Recovered %d worktree(s)", recovered)

				// Show recovered worktrees
				worktrees := manager.List()
				for _, info := range worktrees {
					if info.Status == "recovered" {
						fmt.Printf("  - %s (%s)\n", info.SpecID, info.Branch)
					}
				}
			} else {
				color.Yellow("No new worktrees found to recover")
			}

			return nil
		},
	}

	return cmd
}
