package worktreecmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	var (
		branch    string
		base      string
		force     bool
		llmConfig string
	)

	cmd := &cobra.Command{
		Use:   "new <spec-id>",
		Short: "Create a new worktree for a SPEC",
		Long: `Create a new worktree for a SPEC.

Creates a new Git worktree with an isolated branch for the specified SPEC ID.
The worktree will be created in the worktree root directory under the project namespace.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specID := args[0]

			manager, err := getManager()
			if err != nil {
				return err
			}

			// Determine LLM config path
			var llmConfigPath string
			if llmConfig != "" {
				llmConfigPath = llmConfig
				if _, err := os.Stat(llmConfigPath); os.IsNotExist(err) {
					color.Red("✗ LLM config file not found: %s", llmConfig)
					return err
				}
			}

			info, err := manager.Create(specID, branch, base, force, llmConfigPath)
			if err != nil {
				color.Red("✗ %v", err)
				return err
			}

			color.Green("✓ Worktree created successfully")
			fmt.Printf("  SPEC ID:    %s\n", info.SpecID)
			fmt.Printf("  Path:       %s\n", info.Path)
			fmt.Printf("  Branch:     %s\n", info.Branch)
			fmt.Printf("  Status:     %s\n", info.Status)
			if llmConfigPath != "" {
				fmt.Printf("  LLM Config: %s\n", filepath.Base(llmConfigPath))
			}
			fmt.Println()
			color.Yellow("Next steps:")
			fmt.Printf("  jikime-wt go %s       # Go to this worktree\n", specID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Custom branch name")
	cmd.Flags().StringVar(&base, "base", "main", "Base branch to create from")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation even if worktree exists")
	cmd.Flags().StringVar(&llmConfig, "llm-config", "", "Path to custom LLM config file")

	return cmd
}
