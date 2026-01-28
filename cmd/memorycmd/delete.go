package memorycmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newDeleteCmd() *cobra.Command {
	var (
		projectDir string
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a memory entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir, _ = os.Getwd()
			}

			store, err := memory.NewStore(projectDir)
			if err != nil {
				return fmt.Errorf("open memory store: %w", err)
			}
			defer store.Close()

			// Verify existence
			m, err := store.GetMemory(args[0])
			if err != nil {
				return fmt.Errorf("memory not found: %w", err)
			}

			if !force {
				content := m.Content
				if len(content) > 100 {
					content = content[:100] + "..."
				}
				fmt.Printf("Delete memory [%s] (%s)?\n", m.ID, m.Type)
				fmt.Printf("  Content: %s\n", content)
				fmt.Printf("Use --force to skip confirmation.\n")
				return nil
			}

			if err := store.DeleteMemory(m.ID); err != nil {
				return fmt.Errorf("delete: %w", err)
			}

			fmt.Printf("Deleted memory %s\n", m.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
