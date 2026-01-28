package memorycmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newMigrateCmd() *cobra.Command {
	var (
		projectDir string
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate v2 memories to MD files + chunk index (v3)",
		Long: `Migrate existing memories from the SQLite database to markdown files.

This reads all entries from the memories table, writes them as daily log
entries in .jikime/memory/ MD files, and then indexes the generated files
into the chunks table for search.

The original memories table is preserved for backward compatibility.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir, _ = os.Getwd()
			}

			store, err := memory.NewStore(projectDir)
			if err != nil {
				return fmt.Errorf("open memory store: %w", err)
			}
			defer store.Close()

			// Get all memories from the old table (use large limit)
			memories, err := store.ListMemories(projectDir, 10000)
			if err != nil {
				return fmt.Errorf("list memories: %w", err)
			}

			if len(memories) == 0 {
				fmt.Println("No memories to migrate.")
				return nil
			}

			fmt.Printf("Found %d memories to migrate.\n", len(memories))

			if dryRun {
				fmt.Println("[dry-run] Would write the following entries:")
				for _, m := range memories {
					fmt.Printf("  - [%s] %s\n", m.Type, truncateStr(m.Content, 80))
				}
				return nil
			}

			// Write each memory to daily log MD
			written := 0
			var lastRelPath string
			for _, m := range memories {
				meta := m.Metadata
				if meta == "" && m.SessionID != "" {
					// Include session ID as metadata
					metaBytes, _ := json.Marshal(map[string]string{"session_id": m.SessionID})
					meta = string(metaBytes)
				}

				entry := memory.DailyLogEntry{
					Type:     m.Type,
					Content:  m.Content,
					Metadata: meta,
				}

				relPath, err := memory.AppendDailyLog(projectDir, entry)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to write memory %s: %v\n", m.ID, err)
					continue
				}
				lastRelPath = relPath
				written++
			}

			fmt.Printf("Wrote %d/%d memories to MD files.\n", written, len(memories))

			// Index the generated files
			fmt.Println("Indexing generated files...")
			cfg := memory.LoadEmbeddingConfig()
			provider, _ := memory.NewEmbeddingProvider(cfg)

			indexer := memory.NewIndexer(store, provider)
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			if err := indexer.IndexAll(ctx, projectDir); err != nil {
				return fmt.Errorf("index: %w", err)
			}

			chunkCount, _ := store.ChunkCount()
			fileCount, _ := store.FileCount()
			fmt.Printf("Indexed files: %d, Total chunks: %d\n", fileCount, chunkCount)

			if lastRelPath != "" {
				fmt.Printf("Last file written: %s\n", lastRelPath)
			}

			fmt.Println("Migration complete. Original memories table is preserved.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview migration without writing files")

	return cmd
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
