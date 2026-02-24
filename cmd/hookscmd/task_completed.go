package hookscmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// TaskCompletedCmd represents the task-completed hook command
var TaskCompletedCmd = &cobra.Command{
	Use:   "task-completed",
	Short: "Validate acceptance criteria before accepting task completion",
	Long: `TaskCompleted hook that checks SPEC acceptance criteria before allowing a task to be marked complete.

Features:
- Extract SPEC ID from task subject (e.g., SPEC-AUTH-001)
- Parse acceptance criteria from spec.md
- Reject if unchecked criteria remain (exit code 2)
- Accept if all criteria checked or no SPEC referenced

Only active in Agent Teams mode (CLAUDE_HOOK_EVENT_TEAM_NAME is set).
Based on moai-adk task_completed hook pattern.`,
	RunE: runTaskCompleted,
}

var specIDPattern = regexp.MustCompile(`SPEC-[A-Z]+-\d+`)

type taskCompletedInput struct {
	TaskSubject     string `json:"task_subject"`
	TaskDescription string `json:"task_description"`
	TaskID          string `json:"task_id"`
}

type taskCompletedOutput struct {
	Continue       bool   `json:"continue"`
	SystemMessage  string `json:"systemMessage,omitempty"`
	SuppressOutput bool   `json:"suppressOutput,omitempty"`
}

func runTaskCompleted(cmd *cobra.Command, args []string) error {
	// Only enforce in team mode
	teamName := os.Getenv("CLAUDE_HOOK_EVENT_TEAM_NAME")
	if teamName == "" {
		return outputTaskCompleted(taskCompletedOutput{
			Continue:       true,
			SuppressOutput: true,
		})
	}

	teammateName := os.Getenv("CLAUDE_HOOK_EVENT_TEAMMATE_NAME")

	// Read stdin JSON for task data
	var input taskCompletedInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		// Cannot determine project root - accept completion
		return outputTaskCompleted(taskCompletedOutput{
			Continue:       true,
			SuppressOutput: true,
		})
	}

	// Extract SPEC ID from task subject
	taskSubject := input.TaskSubject
	if taskSubject == "" {
		// No task subject available - accept completion
		return outputTaskCompleted(taskCompletedOutput{
			Continue:       true,
			SuppressOutput: true,
		})
	}

	specID := specIDPattern.FindString(taskSubject)
	if specID == "" {
		// No SPEC ID in task subject - accept completion
		return outputTaskCompleted(taskCompletedOutput{
			Continue:       true,
			SuppressOutput: true,
		})
	}

	// Check if spec.md exists
	specPath := filepath.Join(projectRoot, ".jikime", "specs", specID, "spec.md")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		// SPEC file not found - reject
		msg := fmt.Sprintf(
			"[Hook] TaskCompleted rejected for %s: SPEC %s referenced in task subject but spec.md not found at %s. "+
				"Create the SPEC document or remove the SPEC reference from the task.",
			teammateName, specID, specPath,
		)
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(2)
	}

	// Parse unchecked acceptance criteria
	unchecked := parseUncheckedCriteria(specPath)
	if len(unchecked) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"[Hook] TaskCompleted rejected for %s: %d unchecked acceptance criteria in %s:\n",
			teammateName, len(unchecked), specID,
		))
		for _, item := range unchecked {
			sb.WriteString("  ")
			sb.WriteString(item)
			sb.WriteString("\n")
		}
		sb.WriteString("Complete all acceptance criteria before marking the task as done.")
		fmt.Fprint(os.Stderr, sb.String())
		os.Exit(2)
	}

	// All acceptance criteria checked - accept completion
	return outputTaskCompleted(taskCompletedOutput{
		Continue:       true,
		SuppressOutput: true,
	})
}

func outputTaskCompleted(out taskCompletedOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(out)
}

// parseUncheckedCriteria reads spec.md and returns unchecked items
// under the "## Acceptance Criteria" section.
func parseUncheckedCriteria(specPath string) []string {
	f, err := os.Open(specPath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var (
		inSection bool
		unchecked []string
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Detect section boundaries
		if strings.HasPrefix(line, "## ") {
			if inSection {
				// Exited acceptance criteria section
				break
			}
			if strings.EqualFold(strings.TrimSpace(line), "## Acceptance Criteria") {
				inSection = true
			}
			continue
		}

		// Collect unchecked items
		if inSection {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- [ ] ") {
				unchecked = append(unchecked, trimmed)
			}
		}
	}

	return unchecked
}
