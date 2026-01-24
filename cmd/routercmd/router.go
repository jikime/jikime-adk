package routercmd

import "github.com/spf13/cobra"

// NewRouter creates the router parent command.
func NewRouter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "router",
		Short: "LLM router proxy management",
		Long:  "Manage the LLM router proxy that forwards Claude Code requests to external providers (OpenAI, Gemini, GLM, Ollama).",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newSwitchCmd())

	return cmd
}
