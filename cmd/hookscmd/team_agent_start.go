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

// TeamAgentStartCmd fires at SessionStart when JIKIME_TEAM_NAME is set.
// It registers the agent in the team registry and starts a background heartbeat.
var TeamAgentStartCmd = &cobra.Command{
	Use:   "team-agent-start",
	Short: "Register agent in team registry on session start",
	Long: `Team-aware SessionStart hook.
Reads JIKIME_TEAM_NAME, JIKIME_AGENT_ID, JIKIME_ROLE, JIKIME_DATA_DIR from env.
Registers the agent and emits a system message with team context.
Only activates when JIKIME_TEAM_NAME is set.`,
	RunE: runTeamAgentStart,
}

func runTeamAgentStart(cmd *cobra.Command, args []string) error {
	teamName := os.Getenv("JIKIME_TEAM_NAME")
	if teamName == "" {
		// Not in a team context — pass through silently.
		return writeResponse(HookResponse{Continue: true})
	}

	agentID := os.Getenv("JIKIME_AGENT_ID")
	role := os.Getenv("JIKIME_ROLE")
	dataDir := os.Getenv("JIKIME_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".jikime")
	}
	if agentID == "" {
		agentID = fmt.Sprintf("agent-%d", os.Getpid())
	}
	if role == "" {
		role = "worker"
	}

	td := filepath.Join(dataDir, "teams", teamName)

	// Read stdin (Claude Code passes session payload)
	var input map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	reg, err := team.NewRegistry(filepath.Join(td, "registry"))
	if err != nil {
		// Non-fatal: log and continue
		fmt.Fprintf(os.Stderr, "[jikime/team] registry open failed: %v\n", err)
		return writeResponse(HookResponse{Continue: true})
	}

	info := &team.AgentInfo{
		ID:            agentID,
		TeamName:      teamName,
		Role:          role,
		Status:        team.AgentStatusActive,
		PID:           os.Getpid(),
		LastHeartbeat: time.Now(),
	}
	if err := reg.Register(info); err != nil {
		fmt.Fprintf(os.Stderr, "[jikime/team] register failed: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "[jikime/team] agent %s registered in team %s (role:%s)\n",
			agentID, teamName, role)
	}

	// Send join message to team inbox so the leader knows we started.
	ti := team.NewTeamInbox(td)
	joinMsg := &team.Message{
		TeamName: teamName,
		Kind:     team.MessageKindDirect,
		From:     agentID,
		To:       "leader",
		Subject:  "agent_joined",
		Body:     fmt.Sprintf("Agent %s (%s) joined team %s at %s", agentID, role, teamName, time.Now().Format(time.RFC3339)),
		SentAt:   time.Now(),
	}
	_ = ti.Send(joinMsg)

	msg := fmt.Sprintf("🤝 Team %s | Agent: %s | Role: %s | PID: %d",
		teamName, agentID, role, os.Getpid())

	return writeResponse(HookResponse{
		Continue:      true,
		SystemMessage: msg,
	})
}
