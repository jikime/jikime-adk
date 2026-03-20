package teamcmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Save and restore team sessions",
	}
	cmd.AddCommand(newSessionSaveCmd())
	cmd.AddCommand(newSessionShowCmd())
	cmd.AddCommand(newSessionClearCmd())
	return cmd
}

func newSessionSaveCmd() *cobra.Command {
	var desc string

	cmd := &cobra.Command{
		Use:   "save <team-name>",
		Short: "Save current team state as a session snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err != nil {
				return err
			}
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}
			tasks, _ := store.List("", "")
			agents, _ := reg.List()

			ss, err := team.NewSessionStore(filepath.Join(dataDir(), "sessions", name))
			if err != nil {
				return err
			}
			id, err := ss.Save(name, desc, tasks, agents)
			if err != nil {
				return err
			}
			fmt.Printf("✅ Session saved: %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&desc, "desc", "d", "", "Session description")
	return cmd
}

func newSessionShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <team-name>",
		Short: "List saved sessions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ss, err := team.NewSessionStore(filepath.Join(dataDir(), "sessions", args[0]))
			if err != nil {
				return err
			}
			sessions, err := ss.List()
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				fmt.Println("No saved sessions.")
				return nil
			}
			for _, s := range sessions {
				fmt.Printf("  %s  %s  tasks:%d  agents:%d\n",
					s.ID[:8], s.SavedAt.Format("2006-01-02 15:04:05"),
					len(s.Tasks), len(s.Agents))
				if s.Description != "" {
					fmt.Printf("         %s\n", s.Description)
				}
			}
			return nil
		},
	}
}

func newSessionClearCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "clear <team-name>",
		Short: "Delete a saved session (or all if --id not specified)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ss, err := team.NewSessionStore(filepath.Join(dataDir(), "sessions", args[0]))
			if err != nil {
				return err
			}
			if sessionID != "" {
				if err := ss.Delete(sessionID); err != nil {
					return err
				}
				fmt.Printf("✅ Session %s deleted.\n", sessionID)
				return nil
			}
			sessions, _ := ss.List()
			for _, s := range sessions {
				_ = ss.Delete(s.ID)
			}
			fmt.Printf("✅ All sessions for team %q deleted.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&sessionID, "id", "", "Specific session ID to delete")
	return cmd
}
