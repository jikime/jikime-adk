package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

// TeamCostTrackCmd fires on PostToolUse when JIKIME_TEAM_NAME is set.
// It records token usage and stops the agent if budget is exceeded.
var TeamCostTrackCmd = &cobra.Command{
	Use:   "team-cost-track",
	Short: "Track token costs per PostToolUse and enforce budget",
	Long: `Team-aware PostToolUse hook that accumulates token usage into the cost store.
If JIKIME_TEAM_BUDGET is set and total tokens exceed it, the hook returns
Continue:false to stop the agent gracefully.
Only activates when JIKIME_TEAM_NAME is set.`,
	RunE: runTeamCostTrack,
}

// postToolUseInput is the payload Claude Code sends to PostToolUse hooks.
type postToolUseInput struct {
	SessionID  string `json:"session_id"`
	ToolName   string `json:"tool_name"`
	ToolInput  any    `json:"tool_input"`
	ToolResult any    `json:"tool_result"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func runTeamCostTrack(cmd *cobra.Command, args []string) error {
	teamName := os.Getenv("JIKIME_TEAM_NAME")
	if teamName == "" {
		return writeResponse(HookResponse{Continue: true})
	}

	agentID := os.Getenv("JIKIME_AGENT_ID")
	dataDir := os.Getenv("JIKIME_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".jikime")
	}
	if agentID == "" {
		return writeResponse(HookResponse{Continue: true})
	}

	var input postToolUseInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	inputTokens := input.Usage.InputTokens
	outputTokens := input.Usage.OutputTokens
	if inputTokens == 0 && outputTokens == 0 {
		// No token info in this event — skip silently.
		return writeResponse(HookResponse{Continue: true})
	}

	td := filepath.Join(dataDir, "teams", teamName)

	// Determine budget from team config.
	budget := 0
	cfgData, _ := os.ReadFile(filepath.Join(td, "config.json"))
	var cfg struct {
		Budget int `json:"budget"`
	}
	_ = json.Unmarshal(cfgData, &cfg)
	budget = cfg.Budget

	costStore, err := team.NewCostStore(filepath.Join(td, "costs"), budget)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[jikime/team] cost store: %v\n", err)
		return writeResponse(HookResponse{Continue: true})
	}

	taskID := os.Getenv("JIKIME_TASK_ID") // optional: current task ID
	_, err = costStore.Record(agentID, taskID, input.ToolName, "", inputTokens, outputTokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[jikime/team] record cost: %v\n", err)
	}

	// Enforce budget.
	if budget > 0 {
		exceeded, _ := costStore.BudgetExceeded()
		if exceeded {
			summary, _ := costStore.Summary(agentID)
			total := 0
			if summary != nil {
				total = summary.TotalTokens
			}
			msg := fmt.Sprintf("⛔ Budget exceeded: %d / %d tokens used by agent %s in team %s. Stopping agent.",
				total, budget, agentID, teamName)
			fmt.Fprintf(os.Stderr, "[jikime/team] %s\n", msg)
			return writeResponse(HookResponse{
				Continue:      false,
				SystemMessage: msg,
			})
		}
	}

	return writeResponse(HookResponse{Continue: true})
}
