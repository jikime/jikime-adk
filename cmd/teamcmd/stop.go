package teamcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newStopCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "stop <team-name>",
		Short: "Stop all agents and clean up the team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			if !force {
				fmt.Printf("Stop team %q and kill all agents? [y/N] ", name)
				var confirm string
				fmt.Scan(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}
			agents, _ := reg.List()
			spawner := team.NewSpawner()
			for _, a := range agents {
				res := &team.SpawnResult{
					AgentID:     a.ID,
					TmuxSession: a.TmuxSession,
					PID:         a.PID,
				}
				if err := spawner.Kill(res); err != nil {
					fmt.Fprintf(os.Stderr, "  warn: kill %s: %v\n", a.ID, err)
				}
				_ = reg.MarkDead(a.ID)
			}

			// Remove team directory
			if err := os.RemoveAll(td); err != nil {
				return fmt.Errorf("cleanup: %w", err)
			}
			// Remove session dir
			_ = os.RemoveAll(filepath.Join(dataDir(), "sessions", name))

			fmt.Printf("✅ Team %q stopped and removed.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}
