// Package skillcmd provides the search command for skills.
package skillcmd

import (
	"fmt"
	"strings"

	"jikime-adk-v2/skill"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	searchTags      []string
	searchPhases    []string
	searchAgents    []string
	searchLanguages []string
	searchLimit     int
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search skills by text, tags, or triggers",
		Long: `Search skills by text, tags, or trigger conditions.

Examples:
  jikime-adk skill search nextjs           # Search by text
  jikime-adk skill search --tags framework,nextjs  # Filter by tags
  jikime-adk skill search --phases run     # Filter by phases
  jikime-adk skill search --agents expert-frontend  # Filter by agents
  jikime-adk skill search "react components" --limit 5  # Text search with limit`,
		RunE: runSearch,
	}

	cmd.Flags().StringSliceVar(&searchTags, "tags", nil, "Filter by tags (comma-separated)")
	cmd.Flags().StringSliceVar(&searchPhases, "phases", nil, "Filter by phases (comma-separated)")
	cmd.Flags().StringSliceVar(&searchAgents, "agents", nil, "Filter by agents (comma-separated)")
	cmd.Flags().StringSliceVar(&searchLanguages, "languages", nil, "Filter by languages (comma-separated)")
	cmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum number of results")

	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %w", err)
	}

	// Load skills
	registry := skill.NewRegistry()
	if err := registry.LoadFromProjectRoot(projectRoot); err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	// Build search query
	query := skill.SearchQuery{
		Tags:      searchTags,
		Phases:    searchPhases,
		Agents:    searchAgents,
		Languages: searchLanguages,
		Limit:     searchLimit,
	}

	// Add text query if provided
	if len(args) > 0 {
		query.Text = strings.Join(args, " ")
	}

	// Search
	results := registry.Search(query)

	if len(results) == 0 {
		fmt.Println("No skills found matching the criteria.")
		return nil
	}

	// Output results
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	bold.Printf("Search Results (%d found):\n", len(results))
	fmt.Println(strings.Repeat("-", 80))

	for i, result := range results {
		// Skill name and score
		cyan.Printf("%d. %s", i+1, result.Skill.Name)
		green.Printf(" (score: %.1f)\n", result.Score)

		// Description
		fmt.Printf("   %s\n", truncate(result.Skill.Description, 70))

		// Tags
		if len(result.Skill.Tags) > 0 {
			fmt.Print("   Tags: ")
			yellow.Printf("%s\n", strings.Join(result.Skill.Tags, ", "))
		}

		// Triggers
		if len(result.Skill.Triggers.Keywords) > 0 {
			fmt.Printf("   Keywords: %s\n", strings.Join(result.Skill.Triggers.Keywords[:min(3, len(result.Skill.Triggers.Keywords))], ", "))
		}

		fmt.Println()
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
