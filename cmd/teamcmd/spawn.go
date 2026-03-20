package teamcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newSpawnCmd() *cobra.Command {
	var (
		role            string
		agentID         string
		backend         string
		worktreePath    string
		prompt          string
		skipPermissions bool
		resume          bool
	)

	cmd := &cobra.Command{
		Use:   "spawn <team-name>",
		Short: "Spawn a new agent in the team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName := args[0]
			td := teamDir(teamName)

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
			if resume {
				cfg.ExtraEnv = map[string]string{"JIKIME_RESUME": "1"}
			}

			spawner := team.NewSpawner()
			res, err := spawner.Spawn(cfg)
			if err != nil {
				return fmt.Errorf("spawn: %w", err)
			}

			// Register agent in registry.
			reg, err := team.NewRegistry(td + "/registry")
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
			if err := reg.Register(info); err != nil {
				return fmt.Errorf("register: %w", err)
			}

			fmt.Printf("✅ Agent %q spawned (role: %s)\n", res.AgentID, role)
			if res.TmuxSession != "" {
				fmt.Printf("   tmux:  %s\n", res.TmuxSession)
			}
			if res.PID > 0 {
				fmt.Printf("   pid:   %d\n", res.PID)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&role, "role", "r", "worker", "Agent role: leader, worker, reviewer")
	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID (auto-generated if empty)")
	cmd.Flags().StringVarP(&backend, "backend", "b", "tmux", "Spawn backend: tmux or subprocess")
	cmd.Flags().StringVar(&worktreePath, "worktree", "", "Git worktree path for this agent")
	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Initial prompt for the agent")
	cmd.Flags().BoolVar(&skipPermissions, "skip-permissions", true, "Pass --dangerously-skip-permissions to Claude")
	cmd.Flags().BoolVar(&resume, "resume", false, "Resume previous session if available")
	return cmd
}
