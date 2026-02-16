package memorycmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newGCCmd() *cobra.Command {
	var (
		projectDir string
		maxAgeDays int
		maxCount   int
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Garbage collect old memories",
		Long:  `Remove old or excess memories to keep the database size manageable.`,
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

			opts := memory.GCOptions{
				MaxAge:   time.Duration(maxAgeDays) * 24 * time.Hour,
				MaxCount: maxCount,
				DryRun:   dryRun,
			}

			result, err := store.GarbageCollect(projectDir, opts)
			if err != nil {
				return fmt.Errorf("gc: %w", err)
			}

			if dryRun {
				fmt.Println("Dry run — no changes made:")
			} else {
				fmt.Println("Garbage collection complete:")
			}
			fmt.Printf("  Deleted (age > %dd):  %d\n", maxAgeDays, result.DeletedByAge)
			fmt.Printf("  Deleted (excess):     %d\n", result.DeletedByCount)
			fmt.Printf("  Remaining:            %d\n", result.Remaining)

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")
	cmd.Flags().IntVar(&maxAgeDays, "max-age", 90, "Delete memories older than N days")
	cmd.Flags().IntVar(&maxCount, "max-count", 1000, "Keep at most N memories per project")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Report without deleting")

	return cmd
}
