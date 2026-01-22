// Package skillcmd provides the related command for skills.
package skillcmd

import (
	"fmt"
	"strings"

	"jikime-adk-v2/skill"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var relatedLimit int

func newRelatedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "related [skill-name]",
		Short: "Find skills related to a given skill",
		Long: `Find skills that are related to a given skill.

Related skills are determined by shared tags, phases, agents, and languages.

Examples:
  jikime-adk skill related jikime-lang-typescript
  jikime-adk skill related jikime-platform-vercel --limit 5`,
		Args: cobra.ExactArgs(1),
		RunE: runRelated,
	}

	cmd.Flags().IntVar(&relatedLimit, "limit", 10, "Maximum number of results")

	return cmd
}

func runRelated(cmd *cobra.Command, args []string) error {
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

	// Check if skill exists
	targetSkill := registry.Get(skillName)
	if targetSkill == nil {
		return fmt.Errorf("skill not found: %s", skillName)
	}

	// Get related skills
	related := registry.GetRelated(skillName, relatedLimit)

	if len(related) == 0 {
		fmt.Printf("No related skills found for '%s'.\n", skillName)
		return nil
	}

	// Output
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	bold.Printf("Skills related to '%s':\n", skillName)
	fmt.Println(strings.Repeat("-", 80))

	// Show target skill info
	fmt.Printf("Target: %s\n", targetSkill.Name)
	fmt.Printf("Tags: %s\n", strings.Join(targetSkill.Tags, ", "))
	if len(targetSkill.Triggers.Phases) > 0 {
		fmt.Printf("Phases: %s\n", strings.Join(targetSkill.Triggers.Phases, ", "))
	}
	fmt.Println()

	bold.Println("Related Skills:")
	for i, s := range related {
		cyan.Printf("%d. %s\n", i+1, s.Name)
		fmt.Printf("   %s\n", truncate(s.Description, 70))

		// Show what makes them related
		sharedTags := findSharedTags(targetSkill, s)
		if len(sharedTags) > 0 {
			fmt.Print("   Shared tags: ")
			yellow.Printf("%s\n", strings.Join(sharedTags, ", "))
		}

		sharedPhases := findSharedPhases(targetSkill, s)
		if len(sharedPhases) > 0 {
			fmt.Printf("   Shared phases: %s\n", strings.Join(sharedPhases, ", "))
		}

		sharedAgents := findSharedAgents(targetSkill, s)
		if len(sharedAgents) > 0 {
			fmt.Printf("   Shared agents: %s\n", strings.Join(sharedAgents, ", "))
		}

		fmt.Println()
	}

	return nil
}

func findSharedTags(a, b *skill.Skill) []string {
	var shared []string
	tagSet := make(map[string]bool)
	for _, t := range a.Tags {
		tagSet[t] = true
	}
	for _, t := range b.Tags {
		if tagSet[t] {
			shared = append(shared, t)
		}
	}
	return shared
}

func findSharedPhases(a, b *skill.Skill) []string {
	var shared []string
	phaseSet := make(map[string]bool)
	for _, p := range a.Triggers.Phases {
		phaseSet[p] = true
	}
	for _, p := range b.Triggers.Phases {
		if phaseSet[p] {
			shared = append(shared, p)
		}
	}
	return shared
}

func findSharedAgents(a, b *skill.Skill) []string {
	var shared []string
	agentSet := make(map[string]bool)
	for _, ag := range a.Triggers.Agents {
		agentSet[ag] = true
	}
	for _, ag := range b.Triggers.Agents {
		if agentSet[ag] {
			shared = append(shared, ag)
		}
	}
	return shared
}
