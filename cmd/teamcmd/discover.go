package teamcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

// newDiscoverCmd returns the `jikime team discover` command group.
func newDiscoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover running teams and manage membership",
	}
	cmd.AddCommand(newDiscoverListCmd())
	cmd.AddCommand(newDiscoverJoinCmd())
	cmd.AddCommand(newDiscoverApproveCmd())
	cmd.AddCommand(newDiscoverRejectCmd())
	return cmd
}

// newDiscoverListCmd lists all teams known to this machine (file-based + tmux sessions).
func newDiscoverListCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all active teams on this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			type teamInfo struct {
				Name     string   `json:"name"`
				Template string   `json:"template"`
				Agents   int      `json:"agents"`
				Tasks    int      `json:"tasks"`
				Sessions []string `json:"tmux_sessions,omitempty"`
				Active   bool     `json:"active"`
			}

			teamsDir := filepath.Join(dataDir(), "teams")
			entries, err := os.ReadDir(teamsDir)
			if os.IsNotExist(err) {
				if jsonOut {
					fmt.Println("[]")
					return nil
				}
				fmt.Println("No teams found.")
				return nil
			}
			if err != nil {
				return err
			}

			// Collect active tmux sessions for this machine.
			tmuxSessions := map[string][]string{}
			if out, err2 := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output(); err2 == nil {
				for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
					line = strings.TrimSpace(line)
					if !strings.HasPrefix(line, "jikime-") {
						continue
					}
					// session format: jikime-{team}-{agent} or jikime-{team}-board
					parts := strings.SplitN(strings.TrimPrefix(line, "jikime-"), "-", 2)
					if len(parts) == 2 {
						tmuxSessions[parts[0]] = append(tmuxSessions[parts[0]], line)
					}
				}
			}

			var infos []teamInfo
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				name := e.Name()
				td := filepath.Join(teamsDir, name)

				cfg := struct {
					Template string `json:"template"`
				}{}
				if data, _ := os.ReadFile(filepath.Join(td, "config.json")); len(data) > 0 {
					_ = json.Unmarshal(data, &cfg)
				}

				agentFiles, _ := os.ReadDir(filepath.Join(td, "registry"))
				taskFiles, _ := os.ReadDir(filepath.Join(td, "tasks"))

				sanitizedName := boardSanitize(name)
				sessions := tmuxSessions[sanitizedName]
				infos = append(infos, teamInfo{
					Name:     name,
					Template: cfg.Template,
					Agents:   len(agentFiles),
					Tasks:    len(taskFiles),
					Sessions: sessions,
					Active:   len(sessions) > 0,
				})
			}

			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(infos)
			}

			fmt.Printf("%-20s  %-6s  %-8s  %-6s  %s\n", "TEAM", "ACTIVE", "TEMPLATE", "AGENTS", "TASKS")
			for _, ti := range infos {
				active := "  "
				if ti.Active {
					active = "✅"
				}
				fmt.Printf("%-20s  %-6s  %-8s  %-6d  %d\n",
					ti.Name, active, ti.Template, ti.Agents, ti.Tasks)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

// newDiscoverJoinCmd spawns a new agent that sends a join-request to the team leader.
func newDiscoverJoinCmd() *cobra.Command {
	var (
		role            string
		agentID         string
		backend         string
		worktreePath    string
		skipPermissions bool
	)
	cmd := &cobra.Command{
		Use:   "join <team-name>",
		Short: "Join an existing team as a new agent",
		Long: `Spawn a new agent that announces itself to the team leader via inbox
and begins claiming tasks.

Example:
  jikime team discover join my-team --role worker`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName := args[0]
			td := teamDir(teamName)

			// Ensure team exists.
			if _, err := os.Stat(filepath.Join(td, "config.json")); os.IsNotExist(err) {
				return fmt.Errorf("team %q not found (run `jikime team discover list` to see available teams)", teamName)
			}

			if agentID == "" {
				agentID = fmt.Sprintf("%s-%d", role, time.Now().Unix()%10000)
			}

			// Build a join-request prompt for the new agent.
			prompt := buildJoinPrompt(teamName, agentID, role)

			cfg := team.SpawnConfig{
				TeamName:        teamName,
				AgentID:         agentID,
				Role:            role,
				WorktreePath:    worktreePath,
				InitialPrompt:   prompt,
				Backend:         team.SpawnBackend(backend),
				DataDir:         dataDir(),
				SkipPermissions: skipPermissions,
			}

			spawner := team.NewSpawner()
			res, err := spawner.Spawn(cfg)
			if err != nil {
				return fmt.Errorf("spawn: %w", err)
			}

			// Register agent in registry.
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return fmt.Errorf("registry: %w", err)
			}
			info := &team.AgentInfo{
				ID:          res.AgentID,
				TeamName:    teamName,
				Role:        role,
				Status:      team.AgentStatusActive,
				PID:         res.PID,
				TmuxSession: res.TmuxSession,
			}
			_ = reg.Register(info)

			// Send join-request message to leader inbox.
			ti := team.NewTeamInbox(td)
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     agentID,
				To:       "leader",
				Subject:  "join_request",
				Body:     fmt.Sprintf("Agent %s requests to join team %s as %s", agentID, teamName, role),
				SentAt:   time.Now(),
			}
			_ = ti.Send(msg)

			fmt.Printf("✅ Agent %q joined team %q (role: %s)\n", res.AgentID, teamName, role)
			if res.TmuxSession != "" {
				fmt.Printf("   tmux:  %s\n", res.TmuxSession)
				fmt.Printf("   Attach: tmux attach-session -t %s\n", res.TmuxSession)
			}
			fmt.Printf("   Leader notified via inbox.\n")
			return nil
		},
	}
	cmd.Flags().StringVarP(&role, "role", "r", "worker", "Agent role: worker, reviewer")
	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID (auto-generated if empty)")
	cmd.Flags().StringVarP(&backend, "backend", "b", "tmux", "Spawn backend: tmux or subprocess")
	cmd.Flags().StringVar(&worktreePath, "worktree", "", "Git worktree path for this agent")
	cmd.Flags().BoolVar(&skipPermissions, "skip-permissions", true, "Pass --dangerously-skip-permissions to Claude")
	return cmd
}

// newDiscoverApproveCmd sends an approval message to the joining agent.
func newDiscoverApproveCmd() *cobra.Command {
	var teamName string
	cmd := &cobra.Command{
		Use:   "approve <agent-id>",
		Short: "Approve a pending join request (run as leader)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]
			if teamName == "" {
				teamName = os.Getenv("JIKIME_TEAM_NAME")
			}
			if teamName == "" {
				return fmt.Errorf("--team or JIKIME_TEAM_NAME required")
			}
			td := teamDir(teamName)
			ti := team.NewTeamInbox(td)
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     "leader",
				To:       agentID,
				Subject:  "join_approved",
				Body:     fmt.Sprintf("Welcome to team %s, %s! Begin claiming pending tasks.", teamName, agentID),
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Join approved: %s → %s\n", agentID, teamName)
			return nil
		},
	}
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Team name")
	return cmd
}

// newDiscoverRejectCmd sends a rejection message to the joining agent.
func newDiscoverRejectCmd() *cobra.Command {
	var (
		teamName string
		reason   string
	)
	cmd := &cobra.Command{
		Use:   "reject <agent-id>",
		Short: "Reject a pending join request (run as leader)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]
			if teamName == "" {
				teamName = os.Getenv("JIKIME_TEAM_NAME")
			}
			if teamName == "" {
				return fmt.Errorf("--team or JIKIME_TEAM_NAME required")
			}
			if reason == "" {
				reason = "team capacity reached"
			}
			td := teamDir(teamName)
			ti := team.NewTeamInbox(td)
			msg := &team.Message{
				TeamName: teamName,
				Kind:     team.MessageKindSystem,
				From:     "leader",
				To:       agentID,
				Subject:  "join_rejected",
				Body:     fmt.Sprintf("Join request rejected: %s", reason),
				SentAt:   time.Now(),
			}
			if err := ti.Send(msg); err != nil {
				return err
			}
			fmt.Printf("✅ Join rejected: %s (reason: %s)\n", agentID, reason)
			return nil
		},
	}
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Team name")
	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Rejection reason")
	return cmd
}

// buildJoinPrompt creates the initial prompt for an agent joining an existing team.
func buildJoinPrompt(teamName, agentID, role string) string {
	var sb strings.Builder
	sb.WriteString("## Identity\n\n")
	sb.WriteString(fmt.Sprintf("- Agent: %s\n", agentID))
	sb.WriteString(fmt.Sprintf("- Role: %s\n", role))
	sb.WriteString(fmt.Sprintf("- Team: %s\n", teamName))
	sb.WriteString("\n## Task\n\n")
	sb.WriteString(fmt.Sprintf("You are joining team %q as a %s.\n\n", teamName, role))
	sb.WriteString("Your join request has been sent to the leader. Proceed as follows:\n\n")
	sb.WriteString("1. Check your inbox for approval or rejection:\n")
	sb.WriteString(fmt.Sprintf("   jikime team inbox receive %s\n", teamName))
	sb.WriteString("2. Once approved, begin claiming pending tasks:\n")
	sb.WriteString(fmt.Sprintf("   jikime team tasks list %s --status pending\n", teamName))
	sb.WriteString(fmt.Sprintf("   jikime team tasks claim %s <task-id> --agent %s\n", teamName, agentID))
	sb.WriteString("3. Complete tasks and notify leader:\n")
	sb.WriteString(fmt.Sprintf("   jikime team tasks complete %s <task-id> --result \"summary\"\n", teamName))
	sb.WriteString(fmt.Sprintf("   jikime team inbox send %s leader \"Completed <task-id>: summary\"\n", teamName))
	sb.WriteString("4. Repeat until no tasks remain, then notify leader you are idle.\n")
	return sb.String()
}
