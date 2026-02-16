package memorycmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newSearchCmd() *cobra.Command {
	var (
		projectDir string
		limit      int
		memType    string
		minScore   float64
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search memories using full-text search",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")

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

			results, err := store.Search(memory.SearchQuery{
				ProjectDir: projectDir,
				Query:      query,
				Type:       memType,
				Limit:      limit,
				MinScore:   minScore,
			})
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}

			if len(results) == 0 {
				fmt.Println("No memories found.")
				return nil
			}

			fmt.Printf("Found %d memories:\n\n", len(results))
			for i, r := range results {
				fmt.Printf("[%d] ID: %s\n", i+1, r.Memory.ID)
				fmt.Printf("    Type: %s | Score: %.2f\n", r.Memory.Type, r.Score)
				fmt.Printf("    Session: %s\n", r.Memory.SessionID)
				content := r.Memory.Content
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				fmt.Printf("    Content: %s\n", content)
				fmt.Printf("    Created: %s\n\n", r.Memory.CreatedAt.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum results")
	cmd.Flags().StringVarP(&memType, "type", "t", "", "Filter by memory type")
	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum relevance score (0-1)")

	return cmd
}
