package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk-v2/cmd/banner"
	"jikime-adk-v2/cmd/doctorcmd"
	"jikime-adk-v2/cmd/hookscmd"
	"jikime-adk-v2/cmd/initcmd"
	"jikime-adk-v2/cmd/languagecmd"
	"jikime-adk-v2/cmd/rankcmd"
	"jikime-adk-v2/cmd/statuscmd"
	"jikime-adk-v2/cmd/statuslinecmd"
	"jikime-adk-v2/cmd/tagcmd"
	"jikime-adk-v2/cmd/updatecmd"
	"jikime-adk-v2/cmd/worktreecmd"
	"jikime-adk-v2/version"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:     "jikime-adk",
		Short:   "Jikime ADK — Agentic development toolkit",
		Version: version.String(),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	root.SetVersionTemplate("jikime-adk {{.Version}}\n")

	// Core commands
	root.AddCommand(initcmd.NewInit())
	root.AddCommand(statuscmd.NewStatus())
	root.AddCommand(doctorcmd.NewDoctor())

	// Hook management
	root.AddCommand(hookscmd.NewHooks())

	// Language management
	root.AddCommand(languagecmd.NewLanguage())

	// Rank leaderboard
	root.AddCommand(rankcmd.NewRank())

	// Update command
	root.AddCommand(updatecmd.NewUpdate())

	// Statusline for Claude Code
	root.AddCommand(statuslinecmd.NewStatusline())

	// Git worktree management
	root.AddCommand(worktreecmd.NewWorktree())

	// TAG System v2.0
	root.AddCommand(tagcmd.NewTag())

	// Banner preview (dev tool)
	root.AddCommand(banner.NewBannerPreview())

	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "jikime-adk: %v\n", err)
		os.Exit(1)
	}
}
