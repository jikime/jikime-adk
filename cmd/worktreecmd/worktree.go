// Package worktreecmd provides CLI commands for Git worktree management.
package worktreecmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"jikime-adk-v2/worktree"
)

var (
	repoPath      string
	worktreeRoot  string
)

// NewWorktree creates the worktree command group.
func NewWorktree() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wt"},
		Short:   "Manage Git worktrees for parallel SPEC development",
		Long: `Manage Git worktrees for parallel SPEC development.

Alias: jikime-wt

Git worktrees allow you to have multiple working directories attached to the
same repository, enabling parallel development of multiple SPECs without
stashing or switching branches.`,
	}

	// Persistent flags for all subcommands
	cmd.PersistentFlags().StringVar(&repoPath, "repo", "", "Repository path (default: current directory)")
	cmd.PersistentFlags().StringVar(&worktreeRoot, "worktree-root", "", "Worktree root directory (default: auto-detect)")

	// Add subcommands
	cmd.AddCommand(newNewCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGoCmd())
	cmd.AddCommand(newRemoveCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newSyncCmd())
	cmd.AddCommand(newCleanCmd())
	cmd.AddCommand(newRecoverCmd())
	cmd.AddCommand(newDoneCmd())
	cmd.AddCommand(newConfigCmd())

	return cmd
}

// getManager creates a WorktreeManager with the appropriate paths.
func getManager() (*worktree.Manager, error) {
	// Resolve repository path
	repo := repoPath
	if repo == "" {
		var err error
		repo, err = findGitRepo()
		if err != nil {
			repo, _ = os.Getwd()
		}
	}

	// Resolve worktree root
	wtRoot := worktreeRoot
	if wtRoot == "" {
		wtRoot = detectWorktreeRoot(repo)
	}

	// Get project name from repo path
	projectName := filepath.Base(repo)

	return worktree.NewManager(repo, wtRoot, projectName), nil
}

// findGitRepo walks up the directory tree to find a Git repository.
func findGitRepo() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", os.ErrNotExist
		}
		current = parent
	}
}

// detectWorktreeRoot determines the best location for worktrees.
func detectWorktreeRoot(repoPath string) string {
	home, _ := os.UserHomeDir()

	// Priority: ~/jikime/worktrees > ~/worktrees
	candidates := []string{
		filepath.Join(home, "jikime", "worktrees"),
		filepath.Join(home, "worktrees"),
	}

	// Check if registry exists in any candidate
	for _, root := range candidates {
		registryPath := filepath.Join(root, ".jikime-worktree-registry.json")
		if _, err := os.Stat(registryPath); err == nil {
			return root
		}
	}

	// Check for existing worktrees
	for _, root := range candidates {
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			entries, err := os.ReadDir(root)
			if err == nil && len(entries) > 0 {
				return root
			}
		}
	}

	// Default to ~/jikime/worktrees
	return filepath.Join(home, "jikime", "worktrees")
}
