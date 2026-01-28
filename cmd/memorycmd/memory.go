package memorycmd

import "github.com/spf13/cobra"

// NewMemory creates the memory command group.
func NewMemory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Memory system commands",
		Long:  `Commands for managing the session memory system.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newGCCmd())
	cmd.AddCommand(newIndexCmd())
	cmd.AddCommand(newMigrateCmd())

	return cmd
}
