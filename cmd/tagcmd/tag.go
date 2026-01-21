// Package tagcmd provides TAG System v2.0 commands.
package tagcmd

import (
	"github.com/spf13/cobra"
)

// NewTag creates the tag command with subcommands.
func NewTag() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "TAG System v2.0 commands",
		Long: `TAG System v2.0 commands for SPEC↔CODE traceability.

Commands:
  validate  Validate TAGs in staged files (pre-commit)
  scan      Scan files for TAGs
  linkage   Manage TAG↔CODE linkage database`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newScanCmd())
	cmd.AddCommand(newLinkageCmd())

	return cmd
}
