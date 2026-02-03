package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

// MemorySaveCmd represents the memory-save hook command (SessionEnd).
var MemorySaveCmd = &cobra.Command{
	Use:   "memory-save",
	Short: "Trigger batch embedding on session end (SessionEnd hook)",
	Long: `SessionEnd hook that spawns a background process to generate embeddings
for all un-embedded memories from the current session. Text data is already
saved by UserPromptSubmit and Stop hooks — this hook only handles embedding.`,
	RunE:         runMemorySave,
	SilenceUsage: true,
}

type memorySaveInput struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
}

func runMemorySave(cmd *cobra.Command, args []string) error {
	var input memorySaveInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	projectDir := input.CWD
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	if projectDir == "" {
		return outputMemorySave("no project directory")
	}
	// Find actual project root by searching for .jikime directory upward
	projectDir = memory.FindProjectRoot(projectDir)

	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}

	// Spawn background process for batch embedding.
	// Text data (user_prompt, assistant_response, tool_usage) is already saved
	// by UserPromptSubmit and Stop hooks. This hook only triggers embedding
	// generation as a detached process to avoid blocking session exit.
	spawnEmbedBackfill(projectDir, sessionID)

	return outputMemorySave("Session ended")
}

// spawnEmbedBackfill starts a detached background process to generate embeddings.
// The process runs independently — if it fails or times out, embeddings can be
// retried later via memory_reindex or the next session end.
func spawnEmbedBackfill(projectDir, sessionID string) {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "embed-backfill: cannot find executable: %v\n", err)
		return
	}

	cmd := exec.Command(exe, "hooks", "embed-backfill",
		"--project-dir", projectDir,
		"--session-id", sessionID,
	)
	// Detach: no stdin, stderr only for logging, no stdout
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "embed-backfill: spawn failed: %v\n", err)
		return
	}

	// Release the child process — don't wait for it
	_ = cmd.Process.Release()
}

func outputMemorySave(msg string) error {
	response := struct {
		Continue      bool            `json:"continue"`
		SystemMessage string          `json:"systemMessage,omitempty"`
		Performance   map[string]bool `json:"performance,omitempty"`
	}{
		Continue: true,
		Performance: map[string]bool{
			"go_hook": true,
		},
	}

	if msg != "" {
		response.SystemMessage = msg
	} else {
		response.SystemMessage = "Session ended"
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(response)
}
