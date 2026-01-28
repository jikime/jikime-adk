package hookscmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

// EmbedBackfillCmd generates embeddings for un-embedded memories in the background.
// Spawned by memory-save hook as a detached process to avoid blocking session exit.
var EmbedBackfillCmd = &cobra.Command{
	Use:    "embed-backfill",
	Short:  "Generate embeddings for un-embedded memories (background process)",
	Hidden: true, // Internal use only — spawned by memory-save
	RunE:   runEmbedBackfill,
}

func init() {
	EmbedBackfillCmd.Flags().String("project-dir", "", "Project directory")
	EmbedBackfillCmd.Flags().String("session-id", "", "Session ID to scope embedding")
}

func runEmbedBackfill(cmd *cobra.Command, args []string) error {
	projectDir, _ := cmd.Flags().GetString("project-dir")
	sessionID, _ := cmd.Flags().GetString("session-id")

	if projectDir == "" {
		return fmt.Errorf("--project-dir required")
	}

	store, err := memory.NewStore(projectDir)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	cfg := memory.LoadEmbeddingConfig()
	provider, err := memory.NewEmbeddingProvider(cfg)
	if err != nil || provider == nil {
		return nil // No embedding provider available — nothing to do
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := store.BackfillMemoryEmbeddings(ctx, provider, projectDir, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "embed-backfill: %v\n", err)
		return nil // Non-fatal — embeddings can be retried later
	}

	if count > 0 {
		fmt.Fprintf(os.Stderr, "embed-backfill: %d embeddings generated\n", count)
	}

	return nil
}
