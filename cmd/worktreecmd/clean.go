package worktreecmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newCleanCmd() *cobra.Command {
	var (
		mergedOnly  bool
		stale       bool
		days        int
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove worktrees for merged branches or stale worktrees",
		Long: `Remove worktrees for merged branches or stale worktrees.

By default, removes all worktrees. Use --merged-only to only remove
worktrees whose branches have been merged to main. Use --stale to
remove worktrees not accessed within the specified days.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := getManager()
			if err != nil {
				return err
			}

			var cleaned []string

			if mergedOnly {
				cleaned = manager.CleanMerged()
			} else if stale {
				// Clean stale worktrees (not accessed within N days)
				threshold := time.Now().AddDate(0, 0, -days)
				worktrees := manager.List()

				if len(worktrees) == 0 {
					color.Yellow("No worktrees found")
					return nil
				}

				var staleWorktrees []string
				for _, info := range worktrees {
					if info.LastAccessed.Before(threshold) {
						staleWorktrees = append(staleWorktrees, info.SpecID)
					}
				}

				if len(staleWorktrees) == 0 {
					color.Yellow("No stale worktrees found (threshold: %d days)", days)
					return nil
				}

				color.Cyan("Found %d stale worktree(s) (not accessed in %d days):", len(staleWorktrees), days)
				for _, specID := range staleWorktrees {
					fmt.Printf("  - %s\n", specID)
				}

				// Remove stale worktrees
				for _, specID := range staleWorktrees {
					if err := manager.Remove(specID, true); err != nil {
						color.Red("✗ Failed to remove %s: %v", specID, err)
					} else {
						cleaned = append(cleaned, specID)
					}
				}
			} else if interactive {
				worktrees := manager.List()
				if len(worktrees) == 0 {
					color.Yellow("No worktrees found to clean")
					return nil
				}

				color.Cyan("Found %d worktrees:", len(worktrees))
				for i, info := range worktrees {
					fmt.Printf("  %d. %s (%s) - %s\n", i+1, info.SpecID, info.Branch, info.Status)
				}

				fmt.Println()
				color.Yellow("Select worktrees to remove (comma-separated numbers, or 'all'):")
				fmt.Print("> ")

				reader := bufio.NewReader(os.Stdin)
				selection, _ := reader.ReadString('\n')
				selection = strings.TrimSpace(selection)

				if strings.ToLower(selection) == "all" {
					for _, info := range worktrees {
						cleaned = append(cleaned, info.SpecID)
					}
				} else {
					parts := strings.Split(selection, ",")
					for _, part := range parts {
						idx, err := strconv.Atoi(strings.TrimSpace(part))
						if err != nil || idx < 1 || idx > len(worktrees) {
							continue
						}
						cleaned = append(cleaned, worktrees[idx-1].SpecID)
					}
				}

				if len(cleaned) == 0 {
					color.Yellow("No worktrees selected for cleanup")
					return nil
				}

				// Confirm
				color.Yellow("About to remove %d worktrees: %s", len(cleaned), strings.Join(cleaned, ", "))
				fmt.Print("Continue? [y/N]: ")
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(strings.ToLower(confirm))

				if confirm != "y" && confirm != "yes" {
					color.Yellow("Cleanup cancelled")
					return nil
				}

				// Remove selected worktrees
				var actualCleaned []string
				for _, specID := range cleaned {
					if err := manager.Remove(specID, true); err != nil {
						color.Red("✗ Failed to remove %s: %v", specID, err)
					} else {
						actualCleaned = append(actualCleaned, specID)
					}
				}
				cleaned = actualCleaned
			} else {
				// Clean all
				worktrees := manager.List()
				if len(worktrees) == 0 {
					color.Yellow("No worktrees found to clean")
					return nil
				}

				color.Yellow("Removing all worktrees. Use --merged-only for merged branches only or --interactive for selective cleanup.")

				for _, info := range worktrees {
					if err := manager.Remove(info.SpecID, true); err != nil {
						color.Red("✗ Failed to remove %s: %v", info.SpecID, err)
					} else {
						cleaned = append(cleaned, info.SpecID)
					}
				}
			}

			if len(cleaned) > 0 {
				color.Green("✓ Cleaned %d worktree(s)", len(cleaned))
				for _, specID := range cleaned {
					fmt.Printf("  - %s\n", specID)
				}
			} else {
				color.Yellow("No worktrees were cleaned")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&mergedOnly, "merged-only", false, "Only remove merged branch worktrees")
	cmd.Flags().BoolVar(&stale, "stale", false, "Remove worktrees not accessed within the specified days")
	cmd.Flags().IntVar(&days, "days", 30, "Stale threshold in days (default: 30)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive cleanup with confirmation prompts")

	return cmd
}
