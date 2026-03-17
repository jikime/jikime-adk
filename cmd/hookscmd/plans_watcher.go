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

// PlansWatcherCmd represents the plans-watcher hook command
var PlansWatcherCmd = &cobra.Command{
	Use:   "plans-watcher",
	Short: "Plans.md structure validator after write/edit",
	Long: `PostToolUse hook that validates Plans.md structure whenever it is written or edited.

Checks:
- Valid marker syntax (cc:TODO, cc:WIP, cc:DONE [hash], pm:REVIEW, pm:OK, blocked:<reason>, cc:SKIP)
- DoD column is not empty or vague
- Dependency references point to valid task IDs
- Reports task summary (TODO/WIP/DONE counts)`,
	RunE: runPlansWatcher,
}

// markerPattern matches valid Plans.md status markers
var markerPattern = regexp.MustCompile(
	`^(cc:TODO|cc:WIP|cc:DONE\s+\[[a-f0-9]{6,40}\]|pm:REVIEW|pm:OK|blocked:.+|cc:SKIP)$`,
)

// taskRowPattern matches a Plans.md task table row
// Format: | TaskID | 내용 | DoD | Depends | Status |
var taskRowPattern = regexp.MustCompile(
	`^\|\s*([\d]+\.[\d]+)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|`,
)

// vagueDoD patterns to warn about
var vagueDoDPatterns = []string{
	"잘 작동함", "코드 품질 향상", "개선됨", "완료", "done", "works", "improved",
	"better", "good", "ok", "완료 기준", "yes/no", "기준", "DoD",
}

type plansTask struct {
	id      string
	content string
	dod     string
	depends string
	status  string
	lineNum int
}

func runPlansWatcher(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input postToolInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return nil
	}

	// Only process Write and Edit tools
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return nil
	}

	// Get file path from tool input
	filePathRaw, ok := input.ToolInput["file_path"]
	if !ok {
		return nil
	}
	filePath, ok := filePathRaw.(string)
	if !ok || filePath == "" {
		return nil
	}

	// Only watch Plans.md files
	if filepath.Base(filePath) != "Plans.md" {
		suppressOutput()
		return nil
	}

	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		suppressOutput()
		return nil
	}

	// Parse and validate Plans.md
	tasks, warnings, err := parsePlansFile(filePath)
	if err != nil {
		suppressOutput()
		return nil
	}

	// Build summary
	summary := buildSummary(tasks, warnings)

	if summary == "" {
		suppressOutput()
		return nil
	}

	output := postToolOutput{
		HookSpecificOutput: &postHookSpecificOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: summary,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}

func parsePlansFile(filePath string) ([]plansTask, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var tasks []plansTask
	var warnings []string
	taskIDs := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		m := taskRowPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		task := plansTask{
			id:      strings.TrimSpace(m[1]),
			content: strings.TrimSpace(m[2]),
			dod:     strings.TrimSpace(m[3]),
			depends: strings.TrimSpace(m[4]),
			status:  strings.TrimSpace(m[5]),
			lineNum: lineNum,
		}

		// Skip header rows
		if task.id == "Task" || strings.Contains(task.id, "---") {
			continue
		}

		tasks = append(tasks, task)
		taskIDs[task.id] = true

		// Validate marker
		if !markerPattern.MatchString(task.status) {
			warnings = append(warnings, fmt.Sprintf(
				"Line %d: Task %s has invalid marker `%s`",
				lineNum, task.id, task.status,
			))
		}

		// Validate DoD is not vague
		if isDodVague(task.dod) {
			warnings = append(warnings, fmt.Sprintf(
				"Line %d: Task %s has vague DoD: `%s` — DoD must be Yes/No judgeable",
				lineNum, task.id, task.dod,
			))
		}
	}

	// Validate dependency references
	for _, task := range tasks {
		if task.depends == "-" || task.depends == "" {
			continue
		}
		for dep := range strings.SplitSeq(task.depends, ",") {
			dep = strings.TrimSpace(dep)
			if dep != "-" && dep != "" && !taskIDs[dep] {
				warnings = append(warnings, fmt.Sprintf(
					"Task %s depends on `%s` which does not exist in Plans.md",
					task.id, dep,
				))
			}
		}
	}

	return tasks, warnings, nil
}

func isDodVague(dod string) bool {
	lower := strings.ToLower(dod)
	for _, pattern := range vagueDoDPatterns {
		if strings.EqualFold(dod, pattern) {
			return true
		}
		if lower == strings.ToLower(pattern) {
			return true
		}
	}
	// If it's just a short generic phrase under 5 chars
	if len([]rune(dod)) < 5 {
		return true
	}
	return false
}

func buildSummary(tasks []plansTask, warnings []string) string {
	if len(tasks) == 0 && len(warnings) == 0 {
		return ""
	}

	// Count by status
	counts := map[string]int{
		"TODO": 0, "WIP": 0, "DONE": 0,
		"REVIEW": 0, "OK": 0, "BLOCKED": 0, "SKIP": 0,
	}

	for _, t := range tasks {
		switch {
		case t.status == "cc:TODO":
			counts["TODO"]++
		case t.status == "cc:WIP":
			counts["WIP"]++
		case strings.HasPrefix(t.status, "cc:DONE"):
			counts["DONE"]++
		case t.status == "pm:REVIEW":
			counts["REVIEW"]++
		case t.status == "pm:OK":
			counts["OK"]++
		case strings.HasPrefix(t.status, "blocked:"):
			counts["BLOCKED"]++
		case t.status == "cc:SKIP":
			counts["SKIP"]++
		}
	}

	total := len(tasks)
	if total == 0 && len(warnings) == 0 {
		return ""
	}

	var sb strings.Builder

	// Task summary
	if total > 0 {
		fmt.Fprintf(&sb, "📋 Plans.md: %d tasks", total)

		parts := []string{}
		if counts["TODO"] > 0 {
			parts = append(parts, fmt.Sprintf("TODO:%d", counts["TODO"]))
		}
		if counts["WIP"] > 0 {
			parts = append(parts, fmt.Sprintf("WIP:%d", counts["WIP"]))
		}
		if counts["DONE"] > 0 {
			parts = append(parts, fmt.Sprintf("DONE:%d", counts["DONE"]))
		}
		if counts["REVIEW"] > 0 {
			parts = append(parts, fmt.Sprintf("REVIEW:%d", counts["REVIEW"]))
		}
		if counts["OK"] > 0 {
			parts = append(parts, fmt.Sprintf("OK:%d", counts["OK"]))
		}
		if counts["BLOCKED"] > 0 {
			parts = append(parts, fmt.Sprintf("BLOCKED:%d", counts["BLOCKED"]))
		}
		if counts["SKIP"] > 0 {
			parts = append(parts, fmt.Sprintf("SKIP:%d", counts["SKIP"]))
		}

		if len(parts) > 0 {
			sb.WriteString(" [")
			sb.WriteString(strings.Join(parts, " | "))
			sb.WriteString("]")
		}
		sb.WriteString("\n")
	}

	// Warnings
	if len(warnings) > 0 {
		sb.WriteString("\n⚠️ Plans.md validation warnings:\n")
		maxWarnings := 5
		for i, w := range warnings {
			if i >= maxWarnings {
				fmt.Fprintf(&sb, "  ... and %d more\n", len(warnings)-maxWarnings)
				break
			}
			sb.WriteString("  - " + w + "\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func suppressOutput() {
	output := postToolOutput{SuppressOutput: true}
	encoder := json.NewEncoder(os.Stdout)
	_ = encoder.Encode(output)
}
