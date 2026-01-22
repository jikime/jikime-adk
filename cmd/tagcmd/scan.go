package tagcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"jikime-adk/tag"
)

func newScanCmd() *cobra.Command {
	var (
		recursive bool
		pattern   string
	)

	cmd := &cobra.Command{
		Use:   "scan [path]",
		Short: "Scan files for TAGs",
		Long: `Scan files or directories for @SPEC TAGs.

Examples:
  jikime tag scan .
  jikime tag scan src/ --recursive
  jikime tag scan --pattern "*.py"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			info, err := os.Stat(absPath)
			if err != nil {
				return fmt.Errorf("path not found: %s", path)
			}

			var tags []*tag.TAG

			if info.IsDir() {
				if pattern == "" {
					pattern = "*"
				}
				tags = tag.ExtractTagsFromDirectory(absPath, pattern, recursive)
			} else {
				tags = tag.ExtractTagsFromFile(absPath)
			}

			if len(tags) == 0 {
				fmt.Println("No TAGs found")
				return nil
			}

			// Display results
			color.Cyan("Found %d TAG(s):\n", len(tags))
			fmt.Println()

			for _, t := range tags {
				color.Yellow("  @SPEC[%s:%s]", t.SpecID, t.Verb)
				fmt.Printf("  %s:%d\n\n", t.FilePath, t.Line)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Scan directories recursively")
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "File pattern to match (e.g., *.py)")

	return cmd
}
