// Package skillcmd provides the info command for skills.
package skillcmd

import (
	"fmt"
	"strings"

	"jikime-adk/skill"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var infoShowBody bool

func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [skill-name]",
		Short: "Show detailed information about a skill",
		Long: `Show detailed information about a specific skill.

Examples:
  jikime-adk skill info jikime-lang-typescript
  jikime-adk skill info jikime-platform-vercel --body  # Show markdown body`,
		Args: cobra.ExactArgs(1),
		RunE: runInfo,
	}

	cmd.Flags().BoolVar(&infoShowBody, "body", false, "Show full markdown body content")

	return cmd
}

func runInfo(cmd *cobra.Command, args []string) error {
	skillName := args[0]

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

	// Get skill
	s := registry.Get(skillName)
	if s == nil {
		return fmt.Errorf("skill not found: %s", skillName)
	}

	// If body is requested, reload with full content
	if infoShowBody && s.FilePath != "" {
		fullSkill, err := skill.LoadFromFile(s.FilePath)
		if err == nil {
			s = fullSkill
		}
	}

	// Output
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)

	bold.Println("Skill Information")
	fmt.Println(strings.Repeat("=", 80))

	// Basic info
	cyan.Printf("Name: ")
	fmt.Println(s.Name)

	cyan.Printf("Description: ")
	fmt.Println(s.Description)

	if s.FilePath != "" {
		cyan.Printf("File: ")
		fmt.Println(s.FilePath)
	}

	// Tags
	if len(s.Tags) > 0 {
		cyan.Printf("\nTags: ")
		yellow.Println(strings.Join(s.Tags, ", "))
	}

	// Triggers section
	bold.Println("\nTriggers:")
	fmt.Println(strings.Repeat("-", 40))

	if len(s.Triggers.Keywords) > 0 {
		green.Printf("  Keywords: ")
		fmt.Println(strings.Join(s.Triggers.Keywords, ", "))
	}

	if len(s.Triggers.Phases) > 0 {
		green.Printf("  Phases: ")
		fmt.Println(strings.Join(s.Triggers.Phases, ", "))
	}

	if len(s.Triggers.Agents) > 0 {
		green.Printf("  Agents: ")
		fmt.Println(strings.Join(s.Triggers.Agents, ", "))
	}

	if len(s.Triggers.Languages) > 0 {
		green.Printf("  Languages: ")
		fmt.Println(strings.Join(s.Triggers.Languages, ", "))
	}

	// Optional fields
	if s.Type != "" || s.Framework != "" || s.Version != "" {
		bold.Println("\nMetadata:")
		fmt.Println(strings.Repeat("-", 40))

		if s.Type != "" {
			fmt.Printf("  Type: %s\n", s.Type)
		}
		if s.Framework != "" {
			fmt.Printf("  Framework: %s\n", s.Framework)
		}
		if s.Version != "" {
			fmt.Printf("  Version: %s\n", s.Version)
		}
	}

	// Context and agent
	if s.Context != "" || s.Agent != "" {
		bold.Println("\nExecution Context:")
		fmt.Println(strings.Repeat("-", 40))

		if s.Context != "" {
			fmt.Printf("  Context: %s\n", s.Context)
		}
		if s.Agent != "" {
			fmt.Printf("  Agent: %s\n", s.Agent)
		}
	}

	// Allowed tools
	if len(s.AllowedTools) > 0 {
		bold.Println("\nAllowed Tools:")
		fmt.Println(strings.Repeat("-", 40))
		for _, tool := range s.AllowedTools {
			fmt.Printf("  - %s\n", tool)
		}
	}

	// Body content
	if infoShowBody && s.Body != "" {
		bold.Println("\nContent (Markdown Body):")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println(s.Body)
	}

	return nil
}
