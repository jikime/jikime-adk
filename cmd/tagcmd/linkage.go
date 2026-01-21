package tagcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"jikime-adk-v2/tag"
)

func newLinkageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "linkage",
		Short: "Manage TAG↔CODE linkage database",
		Long: `Manage the bidirectional TAG↔CODE linkage database.

Commands:
  list      List all TAGs in the database
  orphans   Find TAGs referencing missing SPEC documents
  clear     Clear the linkage database
  rebuild   Rebuild linkage database from source files`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newLinkageListCmd())
	cmd.AddCommand(newLinkageOrphansCmd())
	cmd.AddCommand(newLinkageClearCmd())
	cmd.AddCommand(newLinkageRebuildCmd())

	return cmd
}

func newLinkageListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all TAGs in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			lm, err := getLinkageManager()
			if err != nil {
				return err
			}

			tags := lm.GetAllTags()
			if len(tags) == 0 {
				fmt.Println("No TAGs in linkage database")
				return nil
			}

			color.Cyan("Linkage Database: %d TAG(s)\n", len(tags))
			fmt.Println()

			for _, t := range tags {
				color.Yellow("  @SPEC[%s:%s]", t.SpecID, t.Verb)
				fmt.Printf("  %s:%d\n", t.FilePath, t.Line)
			}

			return nil
		},
	}
}

func newLinkageOrphansCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "orphans",
		Short: "Find TAGs referencing missing SPEC documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			lm, err := getLinkageManager()
			if err != nil {
				return err
			}

			cwd, _ := os.Getwd()
			specsDir := filepath.Join(cwd, ".jikime", "specs")

			orphans := lm.FindOrphanedTags(specsDir)
			if len(orphans) == 0 {
				color.Green("No orphaned TAGs found")
				return nil
			}

			color.Yellow("Found %d orphaned TAG(s):\n", len(orphans))
			fmt.Println()

			for _, t := range orphans {
				color.Red("  @SPEC[%s:%s]", t.SpecID, t.Verb)
				fmt.Printf("  %s:%d\n", t.FilePath, t.Line)
			}

			return nil
		},
	}
}

func newLinkageClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear the linkage database",
		RunE: func(cmd *cobra.Command, args []string) error {
			lm, err := getLinkageManager()
			if err != nil {
				return err
			}

			if err := lm.Clear(); err != nil {
				return err
			}

			color.Green("Linkage database cleared")
			return nil
		},
	}
}

func newLinkageRebuildCmd() *cobra.Command {
	var (
		recursive bool
		pattern   string
	)

	cmd := &cobra.Command{
		Use:   "rebuild [path]",
		Short: "Rebuild linkage database from source files",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			lm, err := getLinkageManager()
			if err != nil {
				return err
			}

			// Clear existing database
			if err := lm.Clear(); err != nil {
				return err
			}

			// Scan for TAGs
			if pattern == "" {
				pattern = "*"
			}

			tags := tag.ExtractTagsFromDirectory(absPath, pattern, recursive)

			// Add TAGs to database
			for _, t := range tags {
				if err := lm.AddTag(t); err != nil {
					color.Yellow("Warning: failed to add TAG: %v", err)
				}
			}

			color.Green("Linkage database rebuilt: %d TAG(s)", len(tags))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", true, "Scan directories recursively")
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "File pattern to match")

	return cmd
}

// getLinkageManager returns a LinkageManager instance.
func getLinkageManager() (*tag.LinkageManager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(cwd, ".jikime", "tag_linkage.json")
	return tag.NewLinkageManager(dbPath)
}
