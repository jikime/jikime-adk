package memorycmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"jikime-adk/internal/memory"
)

func newShowCmd() *cobra.Command {
	var projectDir string

	cmd := &cobra.Command{
		Use:   "show [id]",
		Short: "Show memory details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir, _ = os.Getwd()
			}
			// Find project root by searching for .jikime directory upward
			projectDir = memory.FindProjectRoot(projectDir)

			store, err := memory.NewStore(projectDir)
			if err != nil {
				return fmt.Errorf("open memory store: %w", err)
			}
			defer store.Close()

			m, err := store.GetMemory(args[0])
			if err != nil {
				return fmt.Errorf("memory not found: %w", err)
			}

			fmt.Printf("ID:           %s\n", m.ID)
			fmt.Printf("Type:         %s\n", m.Type)
			fmt.Printf("Session:      %s\n", m.SessionID)
			fmt.Printf("Project:      %s\n", m.ProjectDir)
			fmt.Printf("Content Hash: %s\n", m.ContentHash)
			fmt.Printf("Created:      %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Accessed:     %s\n", m.AccessedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Access Count: %d\n", m.AccessCount)

			if m.Metadata != "" {
				var meta map[string]interface{}
				if json.Unmarshal([]byte(m.Metadata), &meta) == nil {
					fmt.Printf("Metadata:     %v\n", meta)
				}
			}

			fmt.Printf("\n--- Content ---\n%s\n", m.Content)

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current dir)")

	return cmd
}
