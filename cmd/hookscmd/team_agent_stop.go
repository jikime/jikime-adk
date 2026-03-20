package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

// TeamAgentStopCmd fires at SessionEnd when JIKIME_TEAM_NAME is set.
// It releases any in-progress tasks and marks the agent offline in the registry.
var TeamAgentStopCmd = &cobra.Command{
	Use:   "team-agent-stop",
	Short: "Release agent tasks and mark offline on session end",
	Long: `Team-aware SessionEnd hook.
Releases claimed tasks back to pending, marks agent as offline in the registry,
and emits a departure message so the leader can reassign work.
Only activates when JIKIME_TEAM_NAME is set.`,
	RunE: runTeamAgentStop,
}

func runTeamAgentStop(cmd *cobra.Command, args []string) error {
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

	td := filepath.Join(dataDir, "teams", teamName)

	// Decode stdin (optional)
	var input map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// Release all in-progress tasks claimed by this agent.
	store, err := team.NewStore(filepath.Join(td, "tasks"))
	if err == nil {
		tasks, _ := store.List(team.TaskStatusInProgress, agentID)
		for _, t := range tasks {
			if t.AgentID == agentID {
				if _, releaseErr := store.Release(t.ID); releaseErr != nil {
					fmt.Fprintf(os.Stderr, "[jikime/team] release task %s: %v\n", t.ID[:8], releaseErr)
				} else {
					fmt.Fprintf(os.Stderr, "[jikime/team] released task %s\n", t.ID[:8])
				}
			}
		}
	}

	// Mark agent offline in registry.
	reg, err := team.NewRegistry(filepath.Join(td, "registry"))
	if err == nil {
		_ = reg.SetStatus(agentID, team.AgentStatusOffline)
		fmt.Fprintf(os.Stderr, "[jikime/team] agent %s marked offline\n", agentID)
	}

	// Notify leader.
	ti := team.NewTeamInbox(td)
	leaveMsg := &team.Message{
		TeamName: teamName,
		Kind:     team.MessageKindDirect,
		From:     agentID,
		To:       "leader",
		Subject:  "agent_left",
		Body:     fmt.Sprintf("Agent %s left team %s at %s", agentID, teamName, time.Now().Format(time.RFC3339)),
		SentAt:   time.Now(),
	}
	_ = ti.Send(leaveMsg)

	return writeResponse(HookResponse{Continue: true})
}
