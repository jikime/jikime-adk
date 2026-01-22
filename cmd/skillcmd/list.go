// Package skillcmd provides the list command for skills.
package skillcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jikime-adk/skill"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	listTag      string
	listPhase    string
	listAgent    string
	listLanguage string
	listFormat   string
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available skills",
		Long: `List all available skills in the project.

Examples:
  jikime-adk skill list                    # List all skills
  jikime-adk skill list --tag framework    # Filter by tag
  jikime-adk skill list --phase run        # Filter by phase
  jikime-adk skill list --agent expert-frontend  # Filter by agent
  jikime-adk skill list --format json      # Output as JSON`,
		RunE: runList,
	}

	cmd.Flags().StringVar(&listTag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&listPhase, "phase", "", "Filter by phase (plan, run, sync)")
	cmd.Flags().StringVar(&listAgent, "agent", "", "Filter by agent")
	cmd.Flags().StringVar(&listLanguage, "language", "", "Filter by language")
	cmd.Flags().StringVar(&listFormat, "format", "table", "Output format (table, json, compact)")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	// Find project root (where .claude directory exists)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %w", err)
	}

	// Load skills
	registry := skill.NewRegistry()
	if err := registry.LoadFromProjectRoot(projectRoot); err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	// Get filtered skills
	var skills []*skill.Skill

	if listTag != "" {
		skills = registry.GetByTag(listTag)
	} else if listPhase != "" {
		skills = registry.GetByPhase(listPhase)
	} else if listAgent != "" {
		skills = registry.GetByAgent(listAgent)
	} else if listLanguage != "" {
		skills = registry.GetByLanguage(listLanguage)
	} else {
		skills = registry.AllSorted()
	}

	// Output
	switch listFormat {
	case "json":
		return outputJSON(skills)
	case "compact":
		return outputCompact(skills)
	default:
		return outputTable(skills)
	}
}

func outputTable(skills []*skill.Skill) error {
	if len(skills) == 0 {
		fmt.Println("No skills found.")
		return nil
	}

	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	bold.Println("Skills:")
	fmt.Println(strings.Repeat("-", 80))

	for _, s := range skills {
		cyan.Printf("%-40s", s.Name)
		fmt.Printf("  %s\n", truncate(s.Description, 35))

		// Show tags
		if len(s.Tags) > 0 {
			fmt.Printf("  Tags: ")
			yellow.Printf("%s\n", strings.Join(s.Tags, ", "))
		}

		// Show triggers summary
		triggers := []string{}
		if len(s.Triggers.Phases) > 0 {
			triggers = append(triggers, fmt.Sprintf("phases=%s", strings.Join(s.Triggers.Phases, ",")))
		}
		if len(s.Triggers.Agents) > 0 {
			triggers = append(triggers, fmt.Sprintf("agents=%d", len(s.Triggers.Agents)))
		}
		if len(s.Triggers.Languages) > 0 {
			triggers = append(triggers, fmt.Sprintf("langs=%s", strings.Join(s.Triggers.Languages, ",")))
		}
		if len(triggers) > 0 {
			fmt.Printf("  Triggers: %s\n", strings.Join(triggers, " | "))
		}

		fmt.Println()
	}

	fmt.Printf("Total: %d skills\n", len(skills))
	return nil
}

func outputCompact(skills []*skill.Skill) error {
	for _, s := range skills {
		fmt.Printf("%s: %s\n", s.Name, truncate(s.Description, 60))
	}
	return nil
}

func outputJSON(skills []*skill.Skill) error {
	fmt.Println("[")
	for i, s := range skills {
		comma := ","
		if i == len(skills)-1 {
			comma = ""
		}
		fmt.Printf("  {\"name\": %q, \"description\": %q, \"tags\": %v}%s\n",
			s.Name, s.Description, toJSONArray(s.Tags), comma)
	}
	fmt.Println("]")
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func toJSONArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	quoted := make([]string, len(arr))
	for i, s := range arr {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func findProjectRoot() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up looking for .claude or templates/.claude directory
	for {
		// Check for .claude directory (initialized projects)
		claudeDir := filepath.Join(dir, ".claude")
		if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
			return dir, nil
		}

		// Check for templates/.claude directory (development/build environment)
		templatesClaudeDir := filepath.Join(dir, "templates", ".claude")
		if info, err := os.Stat(templatesClaudeDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			return "", fmt.Errorf(".claude directory not found")
		}
		dir = parent
	}
}
