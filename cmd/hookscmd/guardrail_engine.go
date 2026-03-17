package hookscmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// GuardrailEngineCmd represents the guardrail-engine hook command
var GuardrailEngineCmd = &cobra.Command{
	Use:   "guardrail-engine",
	Short: "Harness Engineering declarative guardrail rule evaluator (R01-R08)",
	Long: `PostToolUse hook that evaluates Harness Engineering guardrail rules on Plans.md changes.

Rules evaluated:
  R01 Plans.md required     — Warn if no Plans.md exists when harness commands detected
  R02 WIP concurrency limit — Error if more than 2 tasks are cc:WIP simultaneously
  R03 Dependency order      — Error if WIP task has incomplete dependencies
  R05 Review before OK      — Error if pm:OK set without prior pm:REVIEW
  R06 Blocked duration      — Warn if any task has been blocked for 5+ days
  R07 Scope creep           — Warn if task count grew more than 30% from first commit
  R08 Phase order           — Error if a later-phase task is DONE before earlier-phase tasks`,
	RunE: runGuardrailEngine,
}

// guardrailResult holds one rule evaluation outcome
type guardrailResult struct {
	rule    string // e.g. "R02"
	level   string // "error" | "warn" | "info"
	message string
}

// blockedPattern matches blocked:<reason> markers
var blockedPattern = regexp.MustCompile(`^blocked:.+$`)

// pmOKPattern matches pm:OK marker
var pmOKPattern = regexp.MustCompile(`^pm:OK$`)

// pmReviewPattern matches pm:REVIEW marker
var pmReviewPattern = regexp.MustCompile(`^pm:REVIEW$`)

// phaseHeaderPattern matches ## Phase N: name
var phaseHeaderPattern = regexp.MustCompile(`^##\s+Phase\s+(\d+):`)

func runGuardrailEngine(cmd *cobra.Command, args []string) error {
	var input postToolInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return nil
	}

	if input.ToolName != "Write" && input.ToolName != "Edit" {
		suppressOutput()
		return nil
	}

	filePathRaw, ok := input.ToolInput["file_path"]
	if !ok {
		suppressOutput()
		return nil
	}
	filePath, ok := filePathRaw.(string)
	if !ok || filepath.Base(filePath) != "Plans.md" {
		suppressOutput()
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		suppressOutput()
		return nil
	}

	tasks, _, err := parsePlansFile(filePath)
	if err != nil || len(tasks) == 0 {
		suppressOutput()
		return nil
	}

	var results []guardrailResult

	results = append(results, evaluateR02WIPConcurrency(tasks)...)
	results = append(results, evaluateR03DependencyOrder(tasks)...)
	results = append(results, evaluateR05ReviewBeforeOK(tasks)...)
	results = append(results, evaluateR06BlockedDuration(tasks, filePath)...)
	results = append(results, evaluateR07ScopeCreep(tasks, filePath)...)
	results = append(results, evaluateR08PhaseOrder(tasks, filePath)...)

	msg := buildGuardrailMessage(results)
	if msg == "" {
		suppressOutput()
		return nil
	}

	out := postToolOutput{
		HookSpecificOutput: &postHookSpecificOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: msg,
		},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

// -----------------------------------------------------------------------------
// R02: No more than 2 simultaneous cc:WIP tasks
// -----------------------------------------------------------------------------

func evaluateR02WIPConcurrency(tasks []plansTask) []guardrailResult {
	var wip []string
	for _, t := range tasks {
		if t.status == "cc:WIP" {
			wip = append(wip, t.id)
		}
	}
	if len(wip) <= 2 {
		return nil
	}
	return []guardrailResult{{
		rule:  "R02",
		level: "error",
		message: fmt.Sprintf(
			"WIP 동시성 초과: %d개 태스크가 cc:WIP 상태입니다 (%s). 최대 2개 권장.\n"+
				"  → 완료되지 않은 태스크를 먼저 처리하거나 blocked 처리하세요.",
			len(wip), strings.Join(wip, ", "),
		),
	}}
}

// -----------------------------------------------------------------------------
// R03: WIP task must have all dependencies cc:DONE or pm:OK
// -----------------------------------------------------------------------------

func evaluateR03DependencyOrder(tasks []plansTask) []guardrailResult {
	// Build status map
	statusOf := make(map[string]string, len(tasks))
	for _, t := range tasks {
		statusOf[t.id] = t.status
	}

	var results []guardrailResult
	for _, t := range tasks {
		if t.status != "cc:WIP" {
			continue
		}
		if t.depends == "-" || t.depends == "" {
			continue
		}
		for dep := range strings.SplitSeq(t.depends, ",") {
			dep = strings.TrimSpace(dep)
			if dep == "-" || dep == "" {
				continue
			}
			depStatus, exists := statusOf[dep]
			if !exists {
				continue
			}
			if depStatus != "cc:DONE" && !strings.HasPrefix(depStatus, "cc:DONE ") &&
				depStatus != "pm:OK" && depStatus != "pm:REVIEW" {
				results = append(results, guardrailResult{
					rule:  "R03",
					level: "error",
					message: fmt.Sprintf(
						"의존성 미완료: Task %s가 cc:WIP이지만, 의존 태스크 %s는 %s 상태입니다.\n"+
							"  → %s를 먼저 완료하세요.",
						t.id, dep, depStatus, dep,
					),
				})
			}
		}
	}
	return results
}

// -----------------------------------------------------------------------------
// R05: pm:OK must be preceded by pm:REVIEW in git history
// -----------------------------------------------------------------------------

func evaluateR05ReviewBeforeOK(tasks []plansTask) []guardrailResult {
	var results []guardrailResult
	for _, t := range tasks {
		if t.status != "pm:OK" {
			continue
		}
		// Check git log: find if pm:REVIEW ever existed for this task row
		// We look for a commit that had "pm:REVIEW" in the context of this task ID
		hadReview := gitLogContainedMarkerForTask(t.id, "pm:REVIEW")
		if !hadReview {
			results = append(results, guardrailResult{
				rule:  "R05",
				level: "error",
				message: fmt.Sprintf(
					"리뷰 미선행: Task %s가 pm:OK이지만 pm:REVIEW 이력이 없습니다.\n"+
						"  → /jikime:harness-review %s 를 먼저 실행하세요.",
					t.id, t.id,
				),
			})
		}
	}
	return results
}

// gitLogContainedMarkerForTask checks if a given marker ever appeared for a task
// in Plans.md git history (best-effort via git log -S)
func gitLogContainedMarkerForTask(taskID, marker string) bool {
	searchStr := fmt.Sprintf("| %s ", taskID)
	cmd := exec.Command("git", "log", "--oneline", "-S", searchStr, "--", "Plans.md")
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		// If no git history, assume OK (new project)
		return true
	}

	// For each commit that touched this task row, check the content
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 0 {
			continue
		}
		hash := parts[0]
		content := gitShowFile(hash, "Plans.md")
		if strings.Contains(content, searchStr) && strings.Contains(content, marker) {
			return true
		}
	}
	return false
}

// gitShowFile returns file content at a given commit
func gitShowFile(hash, filePath string) string {
	cmd := exec.Command("git", "show", hash+":"+filePath)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// -----------------------------------------------------------------------------
// R06: Tasks blocked for 5+ days emit a warning
// -----------------------------------------------------------------------------

const blockedWarningDays = 5

func evaluateR06BlockedDuration(tasks []plansTask, plansPath string) []guardrailResult {
	var blocked []plansTask
	for _, t := range tasks {
		if blockedPattern.MatchString(t.status) {
			blocked = append(blocked, t)
		}
	}
	if len(blocked) == 0 {
		return nil
	}

	// Find when blocked marker was last introduced in git history
	// Approximation: find oldest commit where "blocked:" appears in Plans.md
	oldestBlockedTime := gitOldestIntroductionTime("blocked:", plansPath)
	if oldestBlockedTime.IsZero() {
		// No git history → just warn existence
		var results []guardrailResult
		for _, t := range blocked {
			results = append(results, guardrailResult{
				rule:  "R06",
				level: "warn",
				message: fmt.Sprintf(
					"차단 태스크: Task %s — %s\n  → 차단 원인을 해결하거나 cc:SKIP으로 처리하세요.",
					t.id, t.status,
				),
			})
		}
		return results
	}

	days := int(time.Since(oldestBlockedTime).Hours() / 24)
	if days < blockedWarningDays {
		return nil
	}

	var results []guardrailResult
	for _, t := range blocked {
		results = append(results, guardrailResult{
			rule:  "R06",
			level: "warn",
			message: fmt.Sprintf(
				"장기 차단 (%d일+): Task %s — %s\n"+
					"  → 차단 원인을 해결하거나 cc:SKIP으로 처리하세요.",
				days, t.id, t.status,
			),
		})
	}
	return results
}

// gitOldestIntroductionTime finds the earliest commit timestamp that introduced the searchStr
func gitOldestIntroductionTime(searchStr, filePath string) time.Time {
	cmd := exec.Command("git", "log", "--pretty=format:%at", "-S", searchStr, "--", filePath)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return time.Time{}
	}

	// Each line is a unix timestamp; the last line is the oldest
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	oldest := time.Time{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ts, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			continue
		}
		t := time.Unix(ts, 0)
		if oldest.IsZero() || t.Before(oldest) {
			oldest = t
		}
	}
	return oldest
}

// -----------------------------------------------------------------------------
// R07: Scope creep — task count grew more than 30% from first commit
// -----------------------------------------------------------------------------

const scopeCreepThreshold = 0.30

func evaluateR07ScopeCreep(tasks []plansTask, plansPath string) []guardrailResult {
	originalCount := gitOriginalTaskCount(plansPath)
	if originalCount <= 0 {
		return nil
	}

	currentCount := len(tasks)
	growth := float64(currentCount-originalCount) / float64(originalCount)

	if growth <= scopeCreepThreshold {
		return nil
	}

	return []guardrailResult{{
		rule:  "R07",
		level: "warn",
		message: fmt.Sprintf(
			"스코프 크리프 감지: 최초 %d개 → 현재 %d개 태스크 (+%.0f%%).\n"+
				"  → Phase 분리 또는 범위 재조정을 고려하세요 (기준: 30%% 초과).",
			originalCount, currentCount, growth*100,
		),
	}}
}

// gitOriginalTaskCount counts tasks in Plans.md at the first commit
func gitOriginalTaskCount(plansPath string) int {
	// Get first commit hash for Plans.md
	cmd := exec.Command("git", "log", "--pretty=format:%H", "--", filepath.Base(plansPath))
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return 0
	}
	firstHash := strings.TrimSpace(lines[len(lines)-1])
	if firstHash == "" {
		return 0
	}

	// Get Plans.md content at first commit
	content := gitShowFile(firstHash, filepath.Base(plansPath))
	if content == "" {
		return 0
	}

	// Count task rows
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader([]byte(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if taskRowPattern.MatchString(line) {
			m := taskRowPattern.FindStringSubmatch(line)
			if len(m) > 1 && m[1] != "Task" && !strings.Contains(m[1], "---") {
				count++
			}
		}
	}
	return count
}

// -----------------------------------------------------------------------------
// R08: Phase ordering — a task in Phase N cannot be DONE before Phase N-1 tasks
// -----------------------------------------------------------------------------

func evaluateR08PhaseOrder(tasks []plansTask, plansPath string) []guardrailResult {
	// Build phase → tasks map preserving order
	type phaseGroup struct {
		num   int
		tasks []plansTask
	}

	// Read raw file to get phase ordering
	file, err := os.Open(plansPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var phases []phaseGroup
	currentPhase := 0

	scanner := bufio.NewScanner(file)
	tasksByID := make(map[string]plansTask)
	for _, t := range tasks {
		tasksByID[t.id] = t
	}

	for scanner.Scan() {
		line := scanner.Text()
		if m := phaseHeaderPattern.FindStringSubmatch(line); m != nil {
			num, err := strconv.Atoi(m[1])
			if err == nil {
				currentPhase = num
				phases = append(phases, phaseGroup{num: num})
			}
		}
		if currentPhase == 0 {
			continue
		}
		if taskRowPattern.MatchString(line) {
			m := taskRowPattern.FindStringSubmatch(line)
			if len(m) > 1 {
				id := strings.TrimSpace(m[1])
				if t, ok := tasksByID[id]; ok {
					if len(phases) > 0 {
						phases[len(phases)-1].tasks = append(phases[len(phases)-1].tasks, t)
					}
				}
			}
		}
	}

	if len(phases) < 2 {
		return nil
	}

	var results []guardrailResult

	for i := 1; i < len(phases); i++ {
		laterPhase := phases[i]
		earlierPhase := phases[i-1]

		// Check if any task in laterPhase is DONE/OK while earlier phase has TODO/WIP
		laterHasDone := false
		for _, t := range laterPhase.tasks {
			if strings.HasPrefix(t.status, "cc:DONE") || t.status == "pm:REVIEW" || t.status == "pm:OK" {
				laterHasDone = true
				break
			}
		}
		if !laterHasDone {
			continue
		}

		// Check if earlier phase still has unfinished tasks
		for _, t := range earlierPhase.tasks {
			if t.status == "cc:TODO" || t.status == "cc:WIP" {
				results = append(results, guardrailResult{
					rule:  "R08",
					level: "warn",
					message: fmt.Sprintf(
						"Phase 순서 역전: Phase %d 태스크가 완료됐지만 Phase %d의 Task %s는 %s 상태입니다.\n"+
							"  → Phase %d 태스크를 먼저 완료하는 것을 권장합니다.",
						laterPhase.num, earlierPhase.num, t.id, t.status, earlierPhase.num,
					),
				})
				break // one warning per phase pair is enough
			}
		}
	}

	return results
}

// -----------------------------------------------------------------------------
// Output formatting
// -----------------------------------------------------------------------------

func buildGuardrailMessage(results []guardrailResult) string {
	if len(results) == 0 {
		return ""
	}

	var errors, warns []guardrailResult
	for _, r := range results {
		switch r.level {
		case "error":
			errors = append(errors, r)
		case "warn":
			warns = append(warns, r)
		}
	}

	var sb strings.Builder

	if len(errors) > 0 {
		fmt.Fprintf(&sb, "🚨 Harness Guardrail 위반 (%d건):\n", len(errors))
		for _, r := range errors {
			fmt.Fprintf(&sb, "\n[%s] %s\n", r.rule, r.message)
		}
	}

	if len(warns) > 0 {
		if len(errors) > 0 {
			sb.WriteString("\n")
		}
		fmt.Fprintf(&sb, "⚠️  Harness Guardrail 경고 (%d건):\n", len(warns))
		for _, r := range warns {
			fmt.Fprintf(&sb, "\n[%s] %s\n", r.rule, r.message)
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}
