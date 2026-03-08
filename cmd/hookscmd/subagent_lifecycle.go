package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// SubagentStartCmd represents the subagent-start hook command
var SubagentStartCmd = &cobra.Command{
	Use:   "subagent-start",
	Short: "Log subagent startup for session tracking",
	Long: `SubagentStart hook that fires when a sub-agent begins execution.
Logs agent startup with timestamp for performance tracking.`,
	RunE: runSubagentStart,
}

// SubagentStopCmd represents the subagent-stop hook command
var SubagentStopCmd = &cobra.Command{
	Use:   "subagent-stop",
	Short: "Log subagent completion for session tracking",
	Long: `SubagentStop hook that fires when a sub-agent finishes execution.
Logs agent completion with timestamp for performance tracking.`,
	RunE: runSubagentStop,
}

type subagentInput struct {
	SessionID           string `json:"session_id"`
	AgentID             string `json:"agent_id"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
}

func runSubagentStart(cmd *cobra.Command, args []string) error {
	var input subagentInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	fmt.Fprintf(os.Stderr, "[jikime] subagent started | agent: %s | time: %s\n",
		input.AgentID, time.Now().Format("15:04:05"))

	response := HookResponse{Continue: true}
	return writeResponse(response)
}

func runSubagentStop(cmd *cobra.Command, args []string) error {
	var input subagentInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	fmt.Fprintf(os.Stderr, "[jikime] subagent stopped | agent: %s | time: %s\n",
		input.AgentID, time.Now().Format("15:04:05"))

	response := HookResponse{Continue: true}
	return writeResponse(response)
}
