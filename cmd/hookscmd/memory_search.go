package hookscmd

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

// MemoryPromptSaveCmd represents the memory-prompt-save hook command (UserPromptSubmit).
var MemoryPromptSaveCmd = &cobra.Command{
	Use:   "memory-prompt-save",
	Short: "Save user prompt to memory DB (UserPromptSubmit hook)",
	Long: `UserPromptSubmit hook that saves the user's prompt to the memory DB.
Memory search is handled by the MCP memory_search tool (on-demand, Clawdbot philosophy).`,
	RunE:         runMemoryPromptSave,
	SilenceUsage: true,
}

type promptSaveInput struct {
	Prompt    string `json:"prompt"`
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
}

func runMemoryPromptSave(cmd *cobra.Command, args []string) error {
	var input promptSaveInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return outputPromptSave()
	}

	if strings.TrimSpace(input.Prompt) == "" {
		return outputPromptSave()
	}

	projectDir := input.CWD
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	if projectDir == "" {
		return outputPromptSave()
	}
	// Find actual project root by searching for .jikime directory upward
	projectDir = memory.FindProjectRoot(projectDir)

	store, err := memory.NewStore(projectDir)
	if err != nil {
		return outputPromptSave()
	}
	defer store.Close()

	// Save user prompt to DB — text only, no embedding.
	// Embedding is deferred to SessionEnd (memory-save) for batch processing.
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}
	store.SaveIfNew(memory.Memory{
		SessionID:  sessionID,
		ProjectDir: projectDir,
		Type:       memory.TypeUserPrompt,
		Content:    input.Prompt,
	})

	// Also save to daily log MD file — text only, no indexing.
	// Indexing with embeddings is deferred to SessionEnd (memory-save).
	_, _ = memory.AppendDailyLog(projectDir, memory.DailyLogEntry{
		Type:    memory.TypeUserPrompt,
		Content: input.Prompt,
	})

	return outputPromptSave()
}

func outputPromptSave() error {
	// Always return empty JSON — no additionalContext injection.
	// Memory search is now handled by MCP memory_search tool (on-demand).
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(struct{}{})
}
