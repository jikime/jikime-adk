package teamcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Manage agent identity environment variables",
	}
	cmd.AddCommand(newIdentityShowCmd())
	cmd.AddCommand(newIdentitySetCmd())
	return cmd
}

func newIdentityShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current agent identity from environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := []string{
				"JIKIME_AGENT_ID",
				"JIKIME_TEAM_NAME",
				"JIKIME_ROLE",
				"JIKIME_DATA_DIR",
				"JIKIME_WORKTREE_PATH",
				"JIKIME_SPAWN_TIME",
				"JIKIME_RESUME",
			}
			fmt.Println("Agent Identity:")
			for _, k := range vars {
				v := os.Getenv(k)
				if v == "" {
					v = "(not set)"
				}
				fmt.Printf("  %-24s %s\n", k+":", v)
			}
			return nil
		},
	}
}

func newIdentitySetCmd() *cobra.Command {
	var agentID, teamName, role string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Print shell export commands to set agent identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID != "" {
				fmt.Printf("export JIKIME_AGENT_ID=%q\n", agentID)
			}
			if teamName != "" {
				fmt.Printf("export JIKIME_TEAM_NAME=%q\n", teamName)
			}
			if role != "" {
				fmt.Printf("export JIKIME_ROLE=%q\n", role)
			}
			if agentID != "" || teamName != "" || role != "" {
				fmt.Printf("export JIKIME_DATA_DIR=%q\n", dataDir())
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&teamName, "team", "", "Team name")
	cmd.Flags().StringVar(&role, "role", "", "Agent role")
	return cmd
}
