package mcpcmd

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"jikime-adk/version"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "serve",
		Short:        "Start MCP server (STDIO transport)",
		Long:         `Starts an MCP server over STDIO.`,
		RunE:         runServe,
		SilenceUsage: true,
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "jikime-adk",
			Version: version.String(),
		},
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return server.Run(ctx, &mcp.StdioTransport{})
}
