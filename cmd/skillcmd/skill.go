// Package skillcmd provides Skill System commands for jikime-adk.
package skillcmd

import (
	"github.com/spf13/cobra"
)

// NewSkill creates the skill command with subcommands.
func NewSkill() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Skill System commands",
		Long: `Skill System commands for tag-based skill discovery.

Commands:
  list      List all available skills
  search    Search skills by text, tags, or triggers
  related   Find skills related to a given skill
  info      Show detailed information about a skill`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newRelatedCmd())
	cmd.AddCommand(newInfoCmd())

	return cmd
}
