package teamcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

func newCreateCmd() *cobra.Command {
	var (
		workers    int
		backend    string
		budget     int
		timeout    int
		maxAgents  int
		tmplName   string
	)

	cmd := &cobra.Command{
		Use:   "create <team-name>",
		Short: "Create a new team workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			// Create directory structure.
			dirs := []string{
				filepath.Join(td, "tasks"),
				filepath.Join(td, "inbox"),
				filepath.Join(td, "registry"),
				filepath.Join(td, "costs"),
				filepath.Join(td, "events"),
			}
			for _, d := range dirs {
				if err := os.MkdirAll(d, 0o755); err != nil {
					return fmt.Errorf("create dir %s: %w", d, err)
				}
			}

			// Also create session and plan dirs.
			sessDir := filepath.Join(dataDir(), "sessions", name)
			planDir := filepath.Join(dataDir(), "plans")
			for _, d := range []string{sessDir, planDir} {
				if err := os.MkdirAll(d, 0o755); err != nil {
					return fmt.Errorf("create dir %s: %w", d, err)
				}
			}

			cfg := team.TeamConfig{
				Name:                  name,
				Template:              tmplName,
				BaseDir:               td,
				MaxAgents:             maxAgents,
				Budget:                budget,
				TimeoutSeconds:        timeout,
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			}
			if workers > 0 {
				cfg.MaxAgents = workers + 1 // +1 for leader
			}

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			cfgPath := filepath.Join(td, "config.json")
			if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Printf("✅ Team %q created\n", name)
			fmt.Printf("   dir:      %s\n", td)
			fmt.Printf("   backend:  %s\n", backend)
			if budget > 0 {
				fmt.Printf("   budget:   %d tokens\n", budget)
			}
			if workers > 0 {
				fmt.Printf("   workers:  %d\n", workers)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&workers, "workers", "w", 0, "Number of worker agents (0 = unlimited)")
	cmd.Flags().StringVarP(&backend, "backend", "b", "tmux", "Spawn backend: tmux or subprocess")
	cmd.Flags().IntVar(&budget, "budget", 0, "Token budget limit (0 = no limit)")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "Execution timeout in seconds (0 = no timeout)")
	cmd.Flags().IntVar(&maxAgents, "max-agents", 0, "Max concurrent agents (0 = unlimited)")
	cmd.Flags().StringVarP(&tmplName, "template", "t", "", "Template name to use")
	return cmd
}
