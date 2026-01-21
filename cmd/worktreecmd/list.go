package worktreecmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all active worktrees",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := getManager()
			if err != nil {
				return err
			}

			worktrees := manager.List()
			if len(worktrees) == 0 {
				color.Yellow("No worktrees found")
				return nil
			}

			if format == "json" {
				var data []map[string]any
				for _, wt := range worktrees {
					data = append(data, wt.ToMap())
				}
				output, _ := json.MarshalIndent(data, "", "  ")
				fmt.Println(string(output))
			} else {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "SPEC ID\tBRANCH\tPATH\tSTATUS\tCREATED")
				fmt.Fprintln(w, "-------\t------\t----\t------\t-------")
				for _, info := range worktrees {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
						info.SpecID,
						info.Branch,
						info.Path,
						info.Status,
						info.CreatedAt.Format("2006-01-02 15:04:05"),
					)
				}
				w.Flush()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}
