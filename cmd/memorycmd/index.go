package memorycmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newIndexCmd() *cobra.Command {
	var (
		projectDir string
		all        bool
		filePath   string
	)

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index memory MD files into SQLite for search",
		Long: `Index memory markdown files into the SQLite chunks table.
By default, indexes all files under .jikime/memory/.
Use --file to index a single file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir, _ = os.Getwd()
			}

			store, err := memory.NewStore(projectDir)
			if err != nil {
				return fmt.Errorf("open memory store: %w", err)
			}
			defer store.Close()

			// Load embedding provider (non-fatal)
			cfg := memory.LoadEmbeddingConfig()
			provider, _ := memory.NewEmbeddingProvider(cfg)

			indexer := memory.NewIndexer(store, provider)
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			if filePath != "" {
				// Index a single file
				fmt.Printf("Indexing %s...\n", filePath)
				if err := indexer.IndexFile(ctx, projectDir, filePath); err != nil {
					return fmt.Errorf("index file: %w", err)
				}
				fmt.Println("Done.")
			} else {
				// Index all files
				fmt.Println("Indexing all memory MD files...")
				if err := indexer.IndexAll(ctx, projectDir); err != nil {
					return fmt.Errorf("index all: %w", err)
				}
			}

			chunkCount, _ := store.ChunkCount()
			fileCount, _ := store.FileCount()
			fmt.Printf("Indexed files: %d, Total chunks: %d\n", fileCount, chunkCount)

			if provider != nil {
				fmt.Printf("Embedding provider: %s (%s)\n", provider.ID(), provider.Model())
			} else {
				fmt.Println("Embedding provider: none (text search only)")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")
	cmd.Flags().BoolVar(&all, "all", false, "Force re-index all files (alias for default behavior)")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Index a single file (relative path)")

	return cmd
}
