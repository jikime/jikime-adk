package teamcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

// boardSanitize replaces chars that are invalid in tmux session/window names.
func boardSanitize(s string) string {
	return strings.NewReplacer(" ", "-", "/", "-", ":", "-").Replace(s)
}

func newBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "View team activity board",
	}
	cmd.AddCommand(newBoardShowCmd())
	cmd.AddCommand(newBoardLiveCmd())
	cmd.AddCommand(newBoardOverviewCmd())
	cmd.AddCommand(newBoardAttachCmd())
	cmd.AddCommand(newBoardServeCmd())
	return cmd
}

func newBoardShowCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show <team-name>",
		Short: "Show current board snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err != nil {
				return err
			}
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}

			tasks, _ := store.List("", "")
			agents, _ := reg.List()

			if jsonOut {
				out := map[string]interface{}{
					"team":   name,
					"agents": agents,
					"tasks":  tasks,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("╔══════════════════════════════════════════════════════╗\n")
			fmt.Printf("║  Team Board: %-39s║\n", name)
			fmt.Printf("╚══════════════════════════════════════════════════════╝\n\n")

			// Agents section
			fmt.Printf("Agents (%d):\n", len(agents))
			for _, a := range agents {
				alive := "❌"
				if ok, _ := reg.IsAlive(a.ID); ok {
					alive = "✅"
				}
				task := "-"
				if a.CurrentTaskID != "" {
					task = a.CurrentTaskID[:8]
				}
				fmt.Printf("  %s %-14s [%-8s]  role:%-12s task:%s\n",
					alive, a.ID, a.Status, a.Role, task)
			}

			// Tasks section
			counts := map[team.TaskStatus]int{}
			for _, t := range tasks {
				counts[t.Status]++
			}
			fmt.Printf("\nTasks (%d total):\n", len(tasks))
			fmt.Printf("  pending:%-4d  in_progress:%-4d  done:%-4d  failed:%-4d  blocked:%-4d\n",
				counts[team.TaskStatusPending],
				counts[team.TaskStatusInProgress],
				counts[team.TaskStatusDone],
				counts[team.TaskStatusFailed],
				counts[team.TaskStatusBlocked],
			)

			fmt.Printf("\nRecent tasks:\n")
			shown := 0
			for _, t := range tasks {
				if shown >= 10 {
					break
				}
				id := t.ID
				if len(id) > 8 {
					id = id[:8]
				}
				agent := t.AgentID
				if agent == "" {
					agent = "-"
				}
				fmt.Printf("  %s  [%-11s]  %-30s  agent:%s\n", id, t.Status, t.Title, agent)
				shown++
			}
			if len(tasks) > 10 {
				fmt.Printf("  ... and %d more\n", len(tasks)-10)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newBoardLiveCmd() *cobra.Command {
	var interval int
	cmd := &cobra.Command{
		Use:   "live <team-name>",
		Short: "Live-refresh board every N seconds (Ctrl+C to stop)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err != nil {
				return err
			}
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}

			tick := time.NewTicker(time.Duration(interval) * time.Second)
			defer tick.Stop()

			printBoard := func() {
				// Clear screen (ANSI)
				fmt.Print("\033[2J\033[H")
				fmt.Printf("Team Board: %s  [%s]  (Ctrl+C to stop)\n\n",
					name, time.Now().Format("15:04:05"))

				tasks, _ := store.List("", "")
				agents, _ := reg.List()

				counts := map[team.TaskStatus]int{}
				for _, t := range tasks {
					counts[t.Status]++
				}
				fmt.Printf("Agents: %d  |  Tasks: pending:%d  wip:%d  done:%d  failed:%d\n\n",
					len(agents),
					counts[team.TaskStatusPending],
					counts[team.TaskStatusInProgress],
					counts[team.TaskStatusDone],
					counts[team.TaskStatusFailed],
				)

				for _, a := range agents {
					alive := "❌"
					if ok, _ := reg.IsAlive(a.ID); ok {
						alive = "✅"
					}
					fmt.Printf("  %s %-14s [%-8s]  role:%s\n", alive, a.ID, a.Status, a.Role)
				}
			}

			printBoard()
			for range tick.C {
				printBoard()
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&interval, "interval", "i", 3, "Refresh interval in seconds")
	return cmd
}

// newBoardAttachCmd creates a dashboard tmux session that links all agent windows
// for a team, giving a unified view without disrupting running agents.
func newBoardAttachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attach <team-name>",
		Short: "Open a tmux dashboard linking all agent windows for the team",
		Long: `Creates a board tmux session and links each agent's window into it.
Use Ctrl-b n/p to navigate between agents. The board session is read-only;
agents continue running in their own sessions unaffected.

Example:
  jikime team board attach my-team`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			prefix := "jikime-" + boardSanitize(name) + "-"
			boardSession := "jikime-" + boardSanitize(name) + "-board"

			// tmux 명령 공통 10초 context — 무한 대기 방지
			tCtx, tCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer tCancel()

			// List all tmux sessions to find agents for this team.
			out, err := exec.CommandContext(tCtx, "tmux", "list-sessions", "-F", "#{session_name}").Output()
			if err != nil {
				return fmt.Errorf("tmux list-sessions: %w (is tmux running?)", err)
			}

			var agentSessions []string
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || line == boardSession {
					continue
				}
				if strings.HasPrefix(line, prefix) {
					agentSessions = append(agentSessions, line)
				}
			}

			if len(agentSessions) == 0 {
				return fmt.Errorf("no active tmux sessions found for team %q (prefix: %s)", name, prefix)
			}

			// Kill stale board session if it exists.
			_ = exec.CommandContext(tCtx, "tmux", "kill-session", "-t", boardSession).Run()

			// Create a fresh board session (starts with a temporary shell window).
			if out, err := exec.CommandContext(tCtx, "tmux", "new-session", "-d", "-s", boardSession, "-n", "_board_").CombinedOutput(); err != nil {
				return fmt.Errorf("create board session: %w\n%s", err, out)
			}

			// Link each agent's window into the board session.
			// -t must point to an existing session only (no window name);
			// tmux appends the linked window automatically at the next index.
			linked := 0
			for _, sess := range agentSessions {
				agentName := strings.TrimPrefix(sess, prefix)
				srcWin := sess + ":" + boardSanitize(agentName)

				if err := exec.CommandContext(tCtx, "tmux", "link-window", "-s", srcWin, "-t", boardSession).Run(); err != nil {
					// Fallback: try window index 0 if named window not found.
					srcWin0 := sess + ":0"
					if err2 := exec.CommandContext(tCtx, "tmux", "link-window", "-s", srcWin0, "-t", boardSession).Run(); err2 != nil {
						fmt.Printf("  ⚠️  could not link %s: %v\n", sess, err2)
						continue
					}
				}
				linked++
			}

			// Remove the initial placeholder window.
			_ = exec.CommandContext(tCtx, "tmux", "kill-window", "-t", boardSession+":_board_").Run()

			if linked == 0 {
				_ = exec.CommandContext(tCtx, "tmux", "kill-session", "-t", boardSession).Run()
				return fmt.Errorf("failed to link any agent windows into board session")
			}

			// Move to first window.
			_ = exec.CommandContext(tCtx, "tmux", "select-window", "-t", boardSession+":0").Run()

			fmt.Printf("📺 Board session: %s\n", boardSession)
			fmt.Printf("   %d agent windows linked (Ctrl-b n/p to switch, Ctrl-b d to detach)\n\n", linked)

			// Replace current process with tmux so it gets full TTY control.
			tmuxBin, err := exec.LookPath("tmux")
			if err != nil {
				return fmt.Errorf("tmux not found: %w", err)
			}
			if os.Getenv("TMUX") != "" {
				// Already inside tmux — switch the current client to the board session.
				return syscall.Exec(tmuxBin, []string{"tmux", "switch-client", "-t", boardSession}, os.Environ())
			}
			return syscall.Exec(tmuxBin, []string{"tmux", "attach-session", "-t", boardSession}, os.Environ())
		},
	}
}

func newBoardOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show overview of all teams",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			teamsDir := filepath.Join(dataDir(), "teams")
			entries, err := os.ReadDir(teamsDir)
			if os.IsNotExist(err) {
				fmt.Printf("No teams found in %s\n", teamsDir)
				_ = home
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Printf("%-20s  %-8s  %-6s  %s\n", "TEAM", "AGENTS", "TASKS", "TEMPLATE")
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				teamName := e.Name()
				td := filepath.Join(teamsDir, teamName)

				// Load config
				cfg := struct {
					Template string `json:"template"`
				}{}
				data, _ := os.ReadFile(filepath.Join(td, "config.json"))
				_ = json.Unmarshal(data, &cfg)

				// Count agents
				regDir := filepath.Join(td, "registry")
				agentFiles, _ := os.ReadDir(regDir)
				agentCount := len(agentFiles)

				// Count tasks
				taskFiles, _ := os.ReadDir(filepath.Join(td, "tasks"))
				taskCount := len(taskFiles)

				fmt.Printf("%-20s  %-8d  %-6d  %s\n",
					teamName, agentCount, taskCount, cfg.Template)
			}
			return nil
		},
	}
}
