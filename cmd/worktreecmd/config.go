package worktreecmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [key] [value]",
		Short: "Get or set worktree configuration",
		Long: `Get or set worktree configuration.

Supported configuration keys:
- root: Worktree root directory
- registry: Registry file path
- all: Show all configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := getManager()
			if err != nil {
				return err
			}

			if len(args) == 0 {
				// Show all configuration
				color.Cyan("Configuration:")
				fmt.Printf("  root:      %s\n", manager.WorktreeRoot)
				fmt.Printf("  registry:  %s\n", manager.Registry.Path())
				fmt.Println()
				color.Yellow("Available commands:")
				fmt.Println("  jikime-wt config all           # Show all config")
				fmt.Println("  jikime-wt config root         # Show worktree root")
				fmt.Println("  jikime-wt config registry     # Show registry path")
			} else if len(args) == 1 {
				key := args[0]
				switch key {
				case "root":
					color.Cyan("Worktree root: %s", manager.WorktreeRoot)
				case "registry":
					color.Cyan("Registry path: %s", manager.Registry.Path())
				case "all":
					color.Cyan("Configuration:")
					fmt.Printf("  root:      %s\n", manager.WorktreeRoot)
					fmt.Printf("  registry:  %s\n", manager.Registry.Path())
				default:
					color.Yellow("Unknown config key: %s", key)
					fmt.Println("Available keys: root, registry, all")
				}
			} else {
				key := args[0]
				switch key {
				case "root":
					color.Yellow("Use --worktree-root option to change root directory")
				default:
					color.Yellow("Cannot set configuration key: %s", key)
				}
			}

			return nil
		},
	}

	return cmd
}
