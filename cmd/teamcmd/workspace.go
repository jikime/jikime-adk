package teamcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage agent git worktree workspaces",
	}
	cmd.AddCommand(newWorkspaceListCmd())
	cmd.AddCommand(newWorkspaceCheckpointCmd())
	cmd.AddCommand(newWorkspaceMergeCmd())
	cmd.AddCommand(newWorkspaceCleanupCmd())
	cmd.AddCommand(newWorkspaceStatusCmd())
	return cmd
}

func workspaceRoot(teamName, agentID string) string {
	return filepath.Join(dataDir(), "worktrees", teamName, agentID)
}

// gitRepoRoot returns the root directory of the git repository containing dir.
// For a worktree, this returns the worktree's own root (not the main repo).
func gitRepoRoot(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repo: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// gitMainRepoRoot returns the root of the MAIN (non-worktree) repository.
// Works correctly whether called from a worktree or the main repo.
func gitMainRepoRoot(dir string) (string, error) {
	// --git-common-dir returns the shared .git directory (e.g. /main-repo/.git)
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repo: %w", err)
	}
	commonDir := strings.TrimSpace(string(out))
	// commonDir is either ".git" (relative, inside main repo) or an absolute path.
	if !filepath.IsAbs(commonDir) {
		// Relative ".git" means dir itself is the main repo.
		return gitRepoRoot(dir)
	}
	// Absolute path — parent of the .git dir is the main repo root.
	return filepath.Dir(commonDir), nil
}

// createWorktree creates a git worktree at wsPath on a new branch named branch.
// If the worktree already exists it is reused. If the branch already exists
// the worktree is created without -b (checkout existing branch).
func createWorktree(gitRoot, wsPath, branch string) error {
	if err := os.MkdirAll(filepath.Dir(wsPath), 0o755); err != nil {
		return err
	}
	// If the worktree directory already exists and has a .git file, reuse it.
	if _, err := os.Stat(filepath.Join(wsPath, ".git")); err == nil {
		return nil
	}

	// Try to create with a new branch first.
	out, err := exec.Command("git", "-C", gitRoot, "worktree", "add", "-b", branch, wsPath).CombinedOutput()
	if err == nil {
		return nil
	}
	// Branch already exists — create worktree without -b.
	if strings.Contains(string(out), "already exists") || strings.Contains(string(out), "fatal: A branch named") {
		out2, err2 := exec.Command("git", "-C", gitRoot, "worktree", "add", wsPath, branch).CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("git worktree add: %s", strings.TrimSpace(string(out2)))
		}
		return nil
	}
	return fmt.Errorf("git worktree add: %s", strings.TrimSpace(string(out)))
}

func newWorkspaceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-name>",
		Short: "List active worktree workspaces for a team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := filepath.Join(dataDir(), "worktrees", args[0])
			entries, err := os.ReadDir(root)
			if os.IsNotExist(err) {
				fmt.Println("No workspaces.")
				return nil
			}
			if err != nil {
				return err
			}
			fmt.Printf("Workspaces for team %q:\n", args[0])
			for _, e := range entries {
				if e.IsDir() {
					fmt.Printf("  %s  %s\n", e.Name(), filepath.Join(root, e.Name()))
				}
			}
			return nil
		},
	}
}

func newWorkspaceCheckpointCmd() *cobra.Command {
	var (
		agentID string
		message string
	)
	cmd := &cobra.Command{
		Use:   "checkpoint <team-name>",
		Short: "Auto-commit current workspace changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			wsDir := workspaceRoot(args[0], agentID)
			if message == "" {
				message = fmt.Sprintf("checkpoint: %s %s", agentID, time.Now().Format("2006-01-02T15:04"))
			}
			for _, c := range [][]string{
				{"git", "-C", wsDir, "add", "-A"},
				{"git", "-C", wsDir, "commit", "-m", message},
			} {
				out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
				if err != nil && !strings.Contains(string(out), "nothing to commit") {
					return fmt.Errorf("git: %w\n%s", err, out)
				}
			}
			fmt.Printf("✅ Checkpoint committed in %s\n", wsDir)
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message")
	return cmd
}

func newWorkspaceMergeCmd() *cobra.Command {
	var (
		agentID string
		target  string
		cleanup bool
	)
	cmd := &cobra.Command{
		Use:   "merge <team-name>",
		Short: "Merge agent workspace branch back to base branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			wsDir := workspaceRoot(args[0], agentID)
			branch := fmt.Sprintf("jikime-%s-%s", args[0], agentID)

			// Find the MAIN repo root (not the worktree root) so the merge lands on main.
			gitRoot, err := gitMainRepoRoot(wsDir)
			if err != nil {
				return fmt.Errorf("find git root: %w", err)
			}

			// Merge the agent branch into target FROM the main repo (not from the worktree).
			mergeMsg := fmt.Sprintf("merge: %s into %s", branch, target)
			out, mergeErr := exec.Command("git", "-C", gitRoot, "merge", "--no-ff", branch, "-m", mergeMsg).CombinedOutput()
			if mergeErr != nil {
				return fmt.Errorf("git merge: %w\n%s", mergeErr, out)
			}

			if cleanup {
				_ = exec.Command("git", "-C", gitRoot, "worktree", "remove", "--force", wsDir).Run()
				_ = os.RemoveAll(wsDir)
			}
			fmt.Printf("✅ Branch %s merged into %s\n", branch, target)
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	cmd.Flags().StringVarP(&target, "target", "t", "main", "Target branch")
	cmd.Flags().BoolVar(&cleanup, "cleanup", false, "Remove worktree after merge")
	return cmd
}

func newWorkspaceCleanupCmd() *cobra.Command {
	var agentID string
	cmd := &cobra.Command{
		Use:   "cleanup <team-name>",
		Short: "Remove workspace(s) for a team or specific agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID != "" {
				wsDir := workspaceRoot(args[0], agentID)
				_ = exec.Command("git", "worktree", "remove", "--force", wsDir).Run()
				_ = os.RemoveAll(wsDir)
				fmt.Printf("✅ Workspace for %s removed\n", agentID)
				return nil
			}
			root := filepath.Join(dataDir(), "worktrees", args[0])
			_ = os.RemoveAll(root)
			fmt.Printf("✅ All workspaces for team %q removed\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Specific agent ID (all if omitted)")
	return cmd
}

func newWorkspaceStatusCmd() *cobra.Command {
	var agentID string
	cmd := &cobra.Command{
		Use:   "status <team-name>",
		Short: "Show git diff stat for an agent's workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				agentID = os.Getenv("JIKIME_AGENT_ID")
			}
			if agentID == "" {
				return fmt.Errorf("--agent or JIKIME_AGENT_ID required")
			}
			wsDir := workspaceRoot(args[0], agentID)
			out, err := exec.Command("git", "-C", wsDir, "diff", "--stat").Output()
			if err != nil {
				return fmt.Errorf("git diff: %w", err)
			}
			if len(out) == 0 {
				fmt.Println("No changes.")
			} else {
				fmt.Print(string(out))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentID, "agent", "a", "", "Agent ID")
	return cmd
}
