package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// StartLoopCmd represents the start-loop hook command
var StartLoopCmd = &cobra.Command{
	Use:   "start-loop",
	Short: "Initialize Ralph Loop session",
	Long: `Start a new Ralph Loop session with specified parameters.

The Ralph Loop enables iterative code improvement with LSP/AST-grep feedback.
Each iteration collects diagnostics and evaluates completion conditions.

Examples:
  jikime hooks start-loop --task "Fix all TypeScript errors"
  jikime hooks start-loop --task "Remove security vulnerabilities" --max-iterations 5 --zero-security
  jikime hooks start-loop --task "Pass all tests" --tests-pass --max-iterations 10`,
	RunE: runStartLoop,
}

var (
	loopTask            string
	loopMaxIterations   int
	loopZeroErrors      bool
	loopZeroWarnings    bool
	loopZeroSecurity    bool
	loopTestsPass       bool
	loopStagnationLimit int
)

func init() {
	StartLoopCmd.Flags().StringVarP(&loopTask, "task", "t", "", "Task description for the loop")
	StartLoopCmd.Flags().IntVarP(&loopMaxIterations, "max-iterations", "m", 10, "Maximum number of iterations")
	StartLoopCmd.Flags().BoolVar(&loopZeroErrors, "zero-errors", true, "Require zero errors for completion")
	StartLoopCmd.Flags().BoolVar(&loopZeroWarnings, "zero-warnings", false, "Require zero warnings for completion")
	StartLoopCmd.Flags().BoolVar(&loopZeroSecurity, "zero-security", false, "Require zero security issues for completion")
	StartLoopCmd.Flags().BoolVar(&loopTestsPass, "tests-pass", false, "Require all tests to pass for completion")
	StartLoopCmd.Flags().IntVar(&loopStagnationLimit, "stagnation-limit", 3, "Max iterations without improvement before stopping")
}

type startLoopOutput struct {
	Status    string                 `json:"status"`
	SessionID string                 `json:"session_id"`
	Task      string                 `json:"task"`
	MaxIter   int                    `json:"max_iterations"`
	Criteria  map[string]interface{} `json:"criteria"`
	Initial   map[string]int         `json:"initial,omitempty"`
	Message   string                 `json:"message,omitempty"`
}

func runStartLoop(cmd *cobra.Command, args []string) error {
	// Check for existing active loop
	existingState := LoadEnhancedLoopState()
	if existingState.Active {
		output := startLoopOutput{
			Status:    "error",
			SessionID: existingState.SessionID,
			Message:   fmt.Sprintf("Loop already active (iteration %d/%d). Use 'jikime hooks cancel-loop' to cancel.", existingState.Iteration, existingState.MaxIterations),
		}
		return outputJSON(output)
	}

	// Generate session ID
	sessionID := fmt.Sprintf("loop-%d", time.Now().Unix())

	// Create initial state
	state := &LoopState{
		Active:          true,
		SessionID:       sessionID,
		StartedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Iteration:       1,
		MaxIterations:   loopMaxIterations,
		TaskDescription: loopTask,
		Criteria: CompletionCriteria{
			ZeroErrors:      loopZeroErrors,
			ZeroWarnings:    loopZeroWarnings,
			ZeroSecurity:    loopZeroSecurity,
			TestsPass:       loopTestsPass,
			StagnationLimit: loopStagnationLimit,
		},
		Snapshots: []DiagnosticSnapshot{},
	}

	// Collect initial snapshot
	initialSnapshot := collectInitialDiagnostics()
	state.AddSnapshot(initialSnapshot)

	// Save state
	if err := SaveEnhancedLoopState(state); err != nil {
		output := startLoopOutput{
			Status:  "error",
			Message: "Failed to save loop state: " + err.Error(),
		}
		return outputJSON(output)
	}

	// Build criteria map for output
	criteria := map[string]interface{}{
		"zero_errors":      loopZeroErrors,
		"zero_warnings":    loopZeroWarnings,
		"zero_security":    loopZeroSecurity,
		"tests_pass":       loopTestsPass,
		"stagnation_limit": loopStagnationLimit,
	}

	// Output success
	output := startLoopOutput{
		Status:    "started",
		SessionID: sessionID,
		Task:      loopTask,
		MaxIter:   loopMaxIterations,
		Criteria:  criteria,
		Initial: map[string]int{
			"errors":   initialSnapshot.ErrorCount,
			"warnings": initialSnapshot.WarningCount,
			"security": initialSnapshot.SecurityIssues,
		},
		Message: formatStartMessage(state, initialSnapshot),
	}

	return outputJSON(output)
}

func collectInitialDiagnostics() DiagnosticSnapshot {
	snapshot := DiagnosticSnapshot{
		Iteration: 1,
		Timestamp: time.Now(),
	}

	// Try to get current diagnostic counts from environment or cache
	// These would be populated by recent post-tool hooks
	if errCount := os.Getenv("JIKIME_LOOP_ERROR_COUNT"); errCount != "" {
		if n, err := strconv.Atoi(errCount); err == nil {
			snapshot.ErrorCount = n
		}
	}

	if warnCount := os.Getenv("JIKIME_LOOP_WARNING_COUNT"); warnCount != "" {
		if n, err := strconv.Atoi(warnCount); err == nil {
			snapshot.WarningCount = n
		}
	}

	if secCount := os.Getenv("JIKIME_LOOP_SECURITY_COUNT"); secCount != "" {
		if n, err := strconv.Atoi(secCount); err == nil {
			snapshot.SecurityIssues = n
		}
	}

	return snapshot
}

func formatStartMessage(state *LoopState, snapshot DiagnosticSnapshot) string {
	msg := "Ralph Loop started"

	if state.TaskDescription != "" {
		msg += ": " + state.TaskDescription
	}

	msg += "\n"
	msg += fmt.Sprintf("Session: %s\n", state.SessionID)
	msg += fmt.Sprintf("Max iterations: %d\n", state.MaxIterations)

	// Initial state
	total := snapshot.ErrorCount + snapshot.WarningCount + snapshot.SecurityIssues
	if total > 0 {
		msg += fmt.Sprintf("Initial issues: %d error(s), %d warning(s), %d security issue(s)\n",
			snapshot.ErrorCount, snapshot.WarningCount, snapshot.SecurityIssues)
	} else {
		msg += "Initial issues: None detected (will collect during first iteration)\n"
	}

	// Completion criteria
	msg += "Completion criteria: "
	var criteria []string
	if state.Criteria.ZeroErrors {
		criteria = append(criteria, "zero errors")
	}
	if state.Criteria.ZeroWarnings {
		criteria = append(criteria, "zero warnings")
	}
	if state.Criteria.ZeroSecurity {
		criteria = append(criteria, "zero security issues")
	}
	if state.Criteria.TestsPass {
		criteria = append(criteria, "tests pass")
	}
	if len(criteria) > 0 {
		for i, c := range criteria {
			if i > 0 {
				msg += ", "
			}
			msg += c
		}
	} else {
		msg += "none specified"
	}

	return msg
}

func outputJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(v)
}
