package mcpcmd

import "github.com/spf13/cobra"

// NewMCP creates the mcp command group.
func NewMCP() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
		Long:  `MCP (Model Context Protocol) server for Claude Code integration.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newServeCmd())

	return cmd
}
