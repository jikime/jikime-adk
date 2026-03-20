package teamcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newLaunchCmd() *cobra.Command {
	var (
		templateName string
		teamName     string
		goal         string
		backend      string
		worktree     bool
		budget       int
	)

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch a full agent team from a template in one command",
		Long: `Create a team, add members, create initial tasks, and spawn all
agents in a single command using a team template.

Example:
  jikime team launch --template leader-worker --goal "implement auth API" --team auth-team`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if templateName == "" {
				return fmt.Errorf("--template is required")
			}
			if teamName == "" {
				teamName = fmt.Sprintf("team-%s", time.Now().Format("0102T1504"))
			}

			// Load template
			ts := team.NewTemplateStore(templateDirs()...)
			def, err := ts.Load(templateName)
			if err != nil {
				return fmt.Errorf("load template: %w", err)
			}
			def = team.Render(def, goal, teamName)

			fmt.Printf("🚀 Launching team %q from template %q\n", teamName, templateName)

			// Step 1: Create team directory structure
			td := teamDir(teamName)
			for _, sub := range []string{"tasks", "inbox", "registry", "costs", "events"} {
				if err := mkdirAll(filepath.Join(td, sub)); err != nil {
					return err
				}
			}
			cfg := team.TeamConfig{
				Name:      teamName,
				Template:  templateName,
				BaseDir:   td,
				Budget:    budget,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if def.DefaultBudget > 0 && budget == 0 {
				cfg.Budget = def.DefaultBudget
			}
			if err := writeJSON(filepath.Join(td, "config.json"), cfg); err != nil {
				return err
			}
			fmt.Printf("  ✅ Team created: %s\n", td)

			// Step 2: Pre-create tasks from template
			if len(def.Tasks) > 0 {
				taskStore, err := team.NewStore(filepath.Join(td, "tasks"))
				if err != nil {
					return fmt.Errorf("task store: %w", err)
				}
				for _, taskDef := range def.Tasks {
					t, err := taskStore.Create(taskDef.Subject, taskDef.Description, taskDef.DoD, nil, 0, nil)
					if err != nil {
						fmt.Printf("  ⚠️  create task %q: %v\n", taskDef.Subject, err)
						continue
					}
					// Pre-assign to owner if specified
					if taskDef.Owner != "" {
						_, _ = taskStore.Claim(t.ID, taskDef.Owner)
					}
					fmt.Printf("  ✅ Task: %s\n", taskDef.Subject)
				}
			}

			// Step 3: Collect worker IDs for leader prompt
			var workerIDs []string
			leaderID := ""
			for _, a := range def.Agents {
				if a.Role == "leader" {
					leaderID = a.ID
				} else if a.Role == "worker" {
					workerIDs = append(workerIDs, a.ID)
				}
			}

			// Step 4: Spawn agents with role-specific prompts
			spawner := team.NewSpawner()
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}

			// Capture the working directory at launch time so agents run in the same dir.
			cwd, _ := os.Getwd()

			// If --worktree, resolve git root and prepare worktree base.
			gitRoot := ""
			if worktree {
				if root, err := gitRepoRoot(cwd); err != nil {
					fmt.Printf("  ⚠️  --worktree requires a git repository: %v\n", err)
					fmt.Printf("     Falling back to current directory.\n")
					worktree = false
				} else {
					gitRoot = root
					fmt.Printf("  📂 Git root: %s\n", gitRoot)
				}
			}

			for _, agentDef := range def.Agents {
				if !agentDef.AutoSpawn {
					continue
				}
				wsPath := cwd // default: launch directory
				if worktree {
					wsPath = workspaceRoot(teamName, agentDef.ID)
					branch := fmt.Sprintf("jikime-%s-%s", teamName, agentDef.ID)
					if err := createWorktree(gitRoot, wsPath, branch); err != nil {
						fmt.Printf("  ⚠️  worktree %s: %v — using cwd\n", agentDef.ID, err)
						wsPath = cwd
					} else {
						fmt.Printf("  🌿 Worktree %-12s → %s\n", agentDef.ID, wsPath)
					}
				}

				// Build role-specific prompt
				pcfg := team.PromptConfig{
					TeamName:     teamName,
					AgentID:      agentDef.ID,
					Role:         agentDef.Role,
					LeaderID:     leaderID,
					Workers:      workerIDs,
					Goal:         goal,
					TaskBody:     agentDef.Task,
					WorktreePath: wsPath,
				}
				// Leader has no separate leaderID field
				if agentDef.Role == "leader" {
					pcfg.LeaderID = ""
				}

				spawnCfg := team.SpawnConfig{
					TeamName:        teamName,
					AgentID:         agentDef.ID,
					Role:            agentDef.Role,
					WorktreePath:    wsPath,
					InitialPrompt:   team.BuildAgentPrompt(pcfg),
					Backend:         team.SpawnBackend(backend),
					DataDir:         dataDir(),
					SkipPermissions: true,
				}

				res, err := spawner.Spawn(spawnCfg)
				if err != nil {
					fmt.Printf("  ⚠️  spawn %s: %v\n", agentDef.ID, err)
					continue
				}
				info := &team.AgentInfo{
					ID:          res.AgentID,
					TeamName:    teamName,
					Role:        agentDef.Role,
					Status:      team.AgentStatusActive,
					PID:         res.PID,
					TmuxSession: res.TmuxSession,
				}
				_ = reg.Register(info)

				if res.TmuxSession != "" {
					fmt.Printf("  ✅ Agent %-12s [%s] → tmux:%s\n", agentDef.ID, agentDef.Role, res.TmuxSession)
				} else {
					fmt.Printf("  ✅ Agent %-12s [%s] → pid:%d\n", agentDef.ID, agentDef.Role, res.PID)
				}
			}

			fmt.Printf("\n✅ Team %q launched with %d agents.\n", teamName, len(def.Agents))
			fmt.Printf("   Monitor: jikime team status %s\n", teamName)
			fmt.Printf("   Board:   jikime team board show %s\n", teamName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&templateName, "template", "t", "", "Template name (required)")
	cmd.Flags().StringVar(&teamName, "name", "", "Team name (auto-generated if empty)")
	cmd.Flags().StringVarP(&goal, "goal", "g", "", "Goal to inject into agent prompts")
	cmd.Flags().StringVarP(&backend, "backend", "b", "tmux", "Spawn backend: tmux or subprocess")
	cmd.Flags().BoolVarP(&worktree, "worktree", "w", false, "Create isolated git worktree per agent")
	cmd.Flags().IntVar(&budget, "budget", 0, "Token budget (overrides template default)")
	_ = cmd.MarkFlagRequired("template")
	return cmd
}
