package hookscmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

// MemoryFlushCmd represents the memory-flush hook command (PreCompact).
var MemoryFlushCmd = &cobra.Command{
	Use:   "memory-flush",
	Short: "Flush memories before context compaction (PreCompact hook)",
	Long: `PreCompact hook that parses the current transcript and writes
important content to daily log MD files before Claude Code
compacts the conversation context.

This is the critical "pre-compaction flush" — content saved here
will be reloaded via SessionStart(source="compact") after compaction.`,
	RunE:    runMemoryFlush,
	SilenceUsage: true,
}

type memoryFlushInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	Trigger        string `json:"trigger"` // "manual" or "auto"
}

type memoryFlushOutput struct {
	HookSpecificOutput *memoryFlushHookOutput `json:"hookSpecificOutput,omitempty"`
}

type memoryFlushHookOutput struct {
	HookEventName string `json:"hookEventName"`
}

func runMemoryFlush(cmd *cobra.Command, args []string) error {
	// Read input from stdin
	var input memoryFlushInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// Determine project dir
	projectDir := input.CWD
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	if projectDir == "" {
		return outputMemoryFlush()
	}
	// Find actual project root by searching for .jikime directory upward
	projectDir = memory.FindProjectRoot(projectDir)

	// Validate transcript path
	if input.TranscriptPath == "" {
		return outputMemoryFlush()
	}
	if _, err := os.Stat(input.TranscriptPath); err != nil {
		return outputMemoryFlush()
	}

	// Open memory store (needed for indexing)
	store, err := memory.NewStore(projectDir)
	if err != nil {
		return outputMemoryFlush()
	}
	defer store.Close()

	// Parse transcript
	transcript, err := memory.ParseTranscript(input.TranscriptPath)
	if err != nil {
		return outputMemoryFlush()
	}

	// Use session ID from input or transcript
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = transcript.SessionID
	}
	_ = sessionID // retained for future use

	// Extract important content
	extracted := memory.Extract(transcript, memory.ExtractOptions{
		SessionID:  sessionID,
		ProjectDir: projectDir,
		Trigger:    input.Trigger,
	})

	// Write each extracted memory to daily log MD file
	var relPath string
	for _, m := range extracted {
		rp, err := memory.AppendDailyLog(projectDir, memory.DailyLogEntry{
			Type:     m.Type,
			Content:  m.Content,
			Metadata: m.Metadata,
		})
		if err == nil {
			relPath = rp
		}
	}

	// Index the daily log file once (with embedding provider)
	if relPath != "" {
		cfg := memory.LoadEmbeddingConfig()
		provider, _ := memory.NewEmbeddingProvider(cfg)
		indexer := memory.NewIndexer(store, provider)
		ctx := context.Background()
		_ = indexer.IndexFile(ctx, projectDir, relPath)
	}

	// Also flush PostToolUse track buffer (safety net)
	memory.FlushTrack(projectDir)

	return outputMemoryFlush()
}

func outputMemoryFlush() error {
	output := memoryFlushOutput{
		HookSpecificOutput: &memoryFlushHookOutput{
			HookEventName: "PreCompact",
		},
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
