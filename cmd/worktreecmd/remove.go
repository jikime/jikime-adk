package worktreecmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "remove <spec-id>",
		Aliases: []string{"rm"},
		Short:   "Remove a worktree",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specID := args[0]

			manager, err := getManager()
			if err != nil {
				return err
			}

			if err := manager.Remove(specID, force); err != nil {
				color.Red("✗ %v", err)
				return err
			}

			color.Green("✓ Worktree removed: %s", specID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force remove with uncommitted changes")

	return cmd
}
