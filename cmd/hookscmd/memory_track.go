package hookscmd

import (
	"encoding/json"
	"os"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

// MemoryTrackCmd represents the memory-track hook command (PostToolUse).
var MemoryTrackCmd = &cobra.Command{
	Use:   "memory-track",
	Short: "Track file modifications from Edit/Write tools (PostToolUse hook)",
	Long: `PostToolUse hook that records file modifications to a JSONL buffer.
Only tracks Edit and Write tool usage. Buffer is flushed at session end (Stop hook).`,
	RunE:         runMemoryTrack,
	SilenceUsage: true,
}

type memoryTrackInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

type memoryTrackOutput struct {
	SuppressOutput bool `json:"suppressOutput"`
}

func runMemoryTrack(cmd *cobra.Command, args []string) error {
	var input memoryTrackInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return outputMemoryTrack()
	}

	// Only track Edit/Write tools
	if input.ToolName != "Edit" && input.ToolName != "Write" {
		return outputMemoryTrack()
	}

	// Get file path
	filePathRaw, ok := input.ToolInput["file_path"]
	if !ok {
		return outputMemoryTrack()
	}
	filePath, ok := filePathRaw.(string)
	if !ok || filePath == "" {
		return outputMemoryTrack()
	}

	projectDir, _ := os.Getwd()
	if projectDir == "" {
		return outputMemoryTrack()
	}
	// Find actual project root by searching for .jikime directory upward
	projectDir = memory.FindProjectRoot(projectDir)

	// Append to track buffer (best-effort)
	record := memory.FileTrackRecord{
		SessionID: os.Getenv("CLAUDE_SESSION_ID"),
		FilePath:  filePath,
		ToolName:  input.ToolName,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_ = memory.AppendTrack(projectDir, record)

	return outputMemoryTrack()
}

func outputMemoryTrack() error {
	output := memoryTrackOutput{SuppressOutput: true}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
