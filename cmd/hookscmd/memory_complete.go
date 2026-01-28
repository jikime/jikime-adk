package hookscmd

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

// MemoryCompleteCmd represents the memory-complete hook command (Stop).
var MemoryCompleteCmd = &cobra.Command{
	Use:   "memory-complete",
	Short: "Flush file tracking buffer and save assistant response to daily log (Stop hook)",
	Long: `Stop hook that:
1. Extracts last assistant response from transcript (like jikime-mem session-end)
2. Saves assistant response to daily log MD file
3. Flushes the PostToolUse file tracking buffer
4. Writes tool_usage entry to daily log MD file
5. Indexes the daily log file with embeddings`,
	RunE:         runMemoryComplete,
	SilenceUsage: true,
}

type memoryCompleteInput struct {
	StopHookActive bool   `json:"stop_hook_active"`
	TranscriptPath string `json:"transcript_path"`
	SessionID      string `json:"session_id"`
	CWD            string `json:"cwd"`
}

type memoryCompleteOutput struct {
	Continue       bool   `json:"continue"`
	SuppressOutput bool   `json:"suppressOutput,omitempty"`
	SystemMessage  string `json:"systemMessage,omitempty"`
}

func runMemoryComplete(cmd *cobra.Command, args []string) error {
	var input memoryCompleteInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	projectDir := input.CWD
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	if projectDir == "" {
		return outputMemoryComplete()
	}

	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}

	// Open store for saving to memories table (text only, no embedding).
	// Embedding is deferred to SessionEnd (memory-save) for batch processing.
	store, err := memory.NewStore(projectDir)
	if err != nil {
		return outputMemoryComplete()
	}
	defer store.Close()

	// 1. Extract last assistant response from transcript
	if input.TranscriptPath != "" {
		if _, err := os.Stat(input.TranscriptPath); err == nil {
			assistantMsg, err := memory.ExtractLastAssistantMessage(input.TranscriptPath)
			if err == nil && assistantMsg != "" {
				content := assistantMsg
				if len(content) > 3000 {
					content = content[:3000] + "..."
				}
				// Save to memories table (for session-scoped batch embedding later)
				store.SaveIfNew(memory.Memory{
					SessionID:  sessionID,
					ProjectDir: projectDir,
					Type:       memory.TypeAssistantResponse,
					Content:    content,
				})
				// Save to daily log MD (human-readable log)
				_, _ = memory.AppendDailyLog(projectDir, memory.DailyLogEntry{
					Type:    memory.TypeAssistantResponse,
					Content: content,
				})
			}
		}
	}

	// 2. Flush track buffer
	records, _ := memory.FlushTrack(projectDir)

	// 3. Write tool_usage entry if files were modified
	if len(records) > 0 {
		fileSet := make(map[string]bool)
		for _, r := range records {
			fileSet[r.FilePath] = true
		}
		var files []string
		for f := range fileSet {
			files = append(files, f)
		}
		content := "Files modified: " + strings.Join(files, ", ")
		// Save to memories table
		store.SaveIfNew(memory.Memory{
			SessionID:  sessionID,
			ProjectDir: projectDir,
			Type:       "tool_usage",
			Content:    content,
		})
		// Save to daily log MD (human-readable log)
		_, _ = memory.AppendDailyLog(projectDir, memory.DailyLogEntry{
			Type:    "tool_usage",
			Content: content,
		})
	}

	return outputMemoryComplete()
}

func outputMemoryComplete() error {
	output := memoryCompleteOutput{
		Continue:       true,
		SuppressOutput: true,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
