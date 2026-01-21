// Package main provides the jikime-wt CLI entry point.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk-v2/cmd/worktreecmd"
	"jikime-adk-v2/version"
)

func main() {
	// Get the worktree command and make it the root
	wtCmd := worktreecmd.NewWorktree()

	// Reconfigure as root command
	root := &cobra.Command{
		Use:     "jikime-wt",
		Short:   "Git worktree management for parallel SPEC development",
		Version: version.String(),
		Long: `jikime-wt - Git worktree management for parallel SPEC development.

Alias for: jikime-adk worktree

Git worktrees allow you to have multiple working directories attached to the
same repository, enabling parallel development of multiple SPECs without
stashing or switching branches.

Commands:
  new       Create a new worktree for a SPEC
  list      List all active worktrees
  go        Go to a worktree (opens new shell)
  remove    Remove a worktree
  status    Show worktree status
  sync      Sync worktree with base branch
  clean     Remove worktrees for merged branches
  recover   Recover registry from existing directories
  done      Complete worktree: merge and cleanup
  config    Get or set configuration`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	root.SetVersionTemplate("jikime-wt {{.Version}}\n")

	// Copy subcommands from worktree command
	for _, subCmd := range wtCmd.Commands() {
		root.AddCommand(subCmd)
	}

	// Copy persistent flags
	root.PersistentFlags().AddFlagSet(wtCmd.PersistentFlags())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "jikime-wt: %v\n", err)
		os.Exit(1)
	}
}
