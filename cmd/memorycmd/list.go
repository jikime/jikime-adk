package memorycmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newListCmd() *cobra.Command {
	var (
		projectDir string
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir, _ = os.Getwd()
			}
			// Find project root by searching for .jikime directory upward
			projectDir = memory.FindProjectRoot(projectDir)

			store, err := memory.NewStore(projectDir)
			if err != nil {
				return fmt.Errorf("open memory store: %w", err)
			}
			defer store.Close()

			memories, err := store.ListMemories(projectDir, limit)
			if err != nil {
				return fmt.Errorf("list memories: %w", err)
			}

			if len(memories) == 0 {
				fmt.Println("No memories found.")
				return nil
			}

			fmt.Printf("Memories (%d):\n\n", len(memories))
			for i, m := range memories {
				content := m.Content
				if len(content) > 120 {
					content = content[:120] + "..."
				}
				idDisplay := m.ID
			if len(idDisplay) > 24 {
				idDisplay = idDisplay[:24]
			}
			fmt.Printf("[%d] %s | %s | %s\n", i+1, idDisplay, m.Type, m.CreatedAt.Format("2006-01-02 15:04"))
				fmt.Printf("    %s\n\n", content)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum entries")

	return cmd
}
