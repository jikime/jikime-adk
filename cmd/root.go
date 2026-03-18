package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk/cmd/banner"
	"jikime-adk/cmd/doctorcmd"
	"jikime-adk/cmd/routercmd"
	"jikime-adk/cmd/hookscmd"
	"jikime-adk/cmd/mcpcmd"
	"jikime-adk/cmd/initcmd"
	"jikime-adk/cmd/languagecmd"
	"jikime-adk/cmd/lspsetupcmd"
	"jikime-adk/cmd/skillcmd"
	"jikime-adk/cmd/statuscmd"
	"jikime-adk/cmd/statuslinecmd"
	"jikime-adk/cmd/tagcmd"
	"jikime-adk/cmd/updatecmd"
	"jikime-adk/cmd/worktreecmd"
	"jikime-adk/version"
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

	// LSP environment setup
	root.AddCommand(lspsetupcmd.NewLspSetup())

	// Update command
	root.AddCommand(updatecmd.NewUpdate())

	// Statusline for Claude Code
	root.AddCommand(statuslinecmd.NewStatusline())

	// Git worktree management
	root.AddCommand(worktreecmd.NewWorktree())

	// TAG System v2.0
	root.AddCommand(tagcmd.NewTag())

	// Skill System (tag-based skill discovery)
	root.AddCommand(skillcmd.NewSkill())

	// LLM Router proxy
	root.AddCommand(routercmd.NewRouter())

	// MCP server
	root.AddCommand(mcpcmd.NewMCP())

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
