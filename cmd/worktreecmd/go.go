package worktreecmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newGoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go <spec-id>",
		Short: "Go to a worktree (opens new shell)",
		Long: `Go to a worktree by opening a new shell in that directory.

This command opens your default shell ($SHELL) in the worktree directory,
allowing you to work on the SPEC in an isolated environment.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specID := args[0]

			manager, err := getManager()
			if err != nil {
				return err
			}

			info := manager.Registry.Get(specID, manager.ProjectName)
			if info == nil {
				color.Red("✗ Worktree not found: %s", specID)
				return fmt.Errorf("worktree not found: %s", specID)
			}

			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/bash"
			}

			color.Green("→ Opening new shell in %s", info.Path)
			shellCmd := exec.Command(shell)
			shellCmd.Dir = info.Path
			shellCmd.Stdin = os.Stdin
			shellCmd.Stdout = os.Stdout
			shellCmd.Stderr = os.Stderr

			return shellCmd.Run()
		},
	}

	return cmd
}
