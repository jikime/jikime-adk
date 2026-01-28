package memorycmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newStatsCmd() *cobra.Command {
	var projectDir string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show memory statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir, _ = os.Getwd()
			}

			store, err := memory.NewStore(projectDir)
			if err != nil {
				return fmt.Errorf("open memory store: %w", err)
			}
			defer store.Close()

			stats, err := store.GetStats(projectDir)
			if err != nil {
				return fmt.Errorf("get stats: %w", err)
			}

			fmt.Println("Memory System Statistics")
			fmt.Println("========================")
			fmt.Printf("Total Memories:  %d\n", stats.TotalMemories)
			fmt.Printf("Total Sessions:  %d\n", stats.TotalSessions)
			fmt.Printf("Total Knowledge: %d\n", stats.TotalKnowledge)
			fmt.Printf("DB Size:         %s\n", formatBytes(stats.DBSizeBytes))

			if stats.OldestMemory != "" {
				fmt.Printf("Oldest Memory:   %s\n", stats.OldestMemory)
			}
			if stats.NewestMemory != "" {
				fmt.Printf("Newest Memory:   %s\n", stats.NewestMemory)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")

	return cmd
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
