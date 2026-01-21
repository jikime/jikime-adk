package worktreecmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show worktree status and sync registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := getManager()
			if err != nil {
				return err
			}

			// Sync registry with Git
			_ = manager.Registry.SyncWithGit(manager.RepoPath)

			worktrees := manager.List()
			if len(worktrees) == 0 {
				color.Yellow("No worktrees found")
				return nil
			}

			color.Cyan("Total worktrees: %d", len(worktrees))
			fmt.Println()

			for _, info := range worktrees {
				statusColor := color.New(color.FgGreen)
				if info.Status != "active" {
					statusColor = color.New(color.FgYellow)
				}

				statusColor.Printf("%s\n", info.SpecID)
				fmt.Printf("  Branch: %s\n", info.Branch)
				fmt.Printf("  Path:   %s\n", info.Path)
				fmt.Printf("  Status: %s\n", info.Status)
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}
