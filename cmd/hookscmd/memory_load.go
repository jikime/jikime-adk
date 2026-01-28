package hookscmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// MemoryLoadCmd represents the memory-load hook command (SessionStart).
// NOTE: This hook is now a no-op. Project context loading has been moved to
// the memory_load MCP tool, which Claude calls on demand instead of injecting
// context unconditionally at every session start.
var MemoryLoadCmd = &cobra.Command{
	Use:   "memory-load",
	Short: "No-op (context loading moved to memory_load MCP tool)",
	Long: `SessionStart hook — now a no-op.

Project context loading has been moved to the memory_load MCP tool
so that Claude can load context on demand rather than unconditionally.
Use the memory_load MCP tool to load MEMORY.md and daily logs when needed.`,
	RunE:    runMemoryLoad,
	SilenceUsage: true,
}

type memoryLoadInput struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	Source    string `json:"source"` // "startup", "compact", "resume", "clear"
}

type memoryLoadOutput struct {
	HookSpecificOutput *memoryLoadHookOutput `json:"hookSpecificOutput,omitempty"`
}

type memoryLoadHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func runMemoryLoad(cmd *cobra.Command, args []string) error {
	// Consume stdin (required by hook protocol) but do nothing
	var input memoryLoadInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// No-op: context loading moved to memory_load MCP tool
	return outputMemoryLoad("")
}

func outputMemoryLoad(context string) error {
	output := memoryLoadOutput{}
	if context != "" {
		output.HookSpecificOutput = &memoryLoadHookOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: context,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
