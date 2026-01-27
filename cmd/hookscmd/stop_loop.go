package hookscmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"jikime-adk/internal/hooks"
)

// StopLoopCmd represents the stop-loop hook command
var StopLoopCmd = &cobra.Command{
	Use:   "stop-loop",
	Short: "Loop controller for feedback loop",
	Long: `Stop hook that checks completion conditions after Claude response and controls feedback loop.

Features:
- Check if loop is active
- Collect diagnostic snapshots from LSP/AST-grep
- Evaluate completion conditions (zero errors, tests pass, stagnation)
- Calculate improvement rate
- Either continue loop or signal completion

Exit codes:
- 0: Loop complete or disabled
- 1: Continue loop (more work needed)`,
	RunE: runStopLoop,
}

const (
	disableEnvVar = "JIKIME_DISABLE_LOOP_CONTROLLER"
)

// Completion markers
var completionMarkers = []string{
	"<jikime>DONE</jikime>",
	"<jikime>COMPLETE</jikime>",
	"<jikime:done />",
	"<jikime:complete />",
}

type stopLoopInput struct {
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

// stopLoopOutput follows Claude Code's expected Stop hook schema
// Stop hooks should NOT use hookSpecificOutput - it's only for PreToolUse/PostToolUse/UserPromptSubmit
type stopLoopOutput struct {
	Continue       bool   `json:"continue"`
	SystemMessage  string `json:"systemMessage,omitempty"`
	SuppressOutput bool   `json:"suppressOutput,omitempty"`
}

func runStopLoop(cmd *cobra.Command, args []string) error {
	// Check if loop controller is disabled
	if disabled := os.Getenv(disableEnvVar); disabled != "" {
		if disabled == "1" || strings.ToLower(disabled) == "true" || strings.ToLower(disabled) == "yes" {
			// Loop disabled - output proper JSON response
			output := stopLoopOutput{
				Continue:       true,
				SuppressOutput: true,
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetEscapeHTML(false)
			return encoder.Encode(output)
		}
	}

	// Read input from stdin (contains conversation context)
	var conversationText string
	var input stopLoopInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err == nil {
		// Extract recent assistant messages for completion marker detection
		for i := len(input.Messages) - 1; i >= 0 && i >= len(input.Messages)-3; i-- {
			if input.Messages[i].Role == "assistant" {
				conversationText += " " + input.Messages[i].Content
			}
		}
	}

	// PRIORITY CHECK: Completion marker - always check first
	if checkCompletionMarker(conversationText) {
		// Clear any active loop state
		ClearEnhancedLoopState()

		output := stopLoopOutput{
			Continue:      true,
			SystemMessage: "Ralph Loop: COMPLETE - Completion marker detected",
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.Encode(output)
		return nil // exit 0 - complete
	}

	// Load current loop state FIRST (lightweight operation)
	state := LoadEnhancedLoopState()

	// If loop is not active, skip expensive diagnostics and exit immediately
	if !state.Active {
		output := stopLoopOutput{
			Continue:       true,
			SuppressOutput: true,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(output)
	}

	// Loop is explicitly active - collect diagnostics (expensive: ruff, tsc, tests)
	currentSnapshot := collectCurrentDiagnostics()

	// Use full loop logic
	state.AddSnapshot(currentSnapshot)

	// Evaluate completion conditions
	result := state.EvaluateCompletion()

	// Determine action
	var action string

	if result.Complete {
		// Loop complete - all conditions met
		state.Active = false
		state.FinalStatus = "COMPLETE"
		state.CompletionReason = result.Reason
		action = "COMPLETE - " + result.Reason
		ClearEnhancedLoopState()
	} else if state.Iteration >= state.MaxIterations {
		// Max iterations reached
		state.Active = false
		state.FinalStatus = "STOPPED"
		state.CompletionReason = "Max iterations reached"
		action = "STOPPED - Max iterations (" + strconv.Itoa(state.MaxIterations) + ") reached"
		ClearEnhancedLoopState()
	} else if state.IsStagnant() {
		// Stagnation detected
		state.Active = false
		state.FinalStatus = "STOPPED"
		state.CompletionReason = "Stagnation - no improvement detected"
		action = "STOPPED - No improvement in last " + strconv.Itoa(state.Criteria.StagnationLimit) + " iterations"
		ClearEnhancedLoopState()
	} else {
		// Continue loop
		state.Iteration++
		action = "CONTINUE"
		SaveEnhancedLoopState(state)
	}

	// Format output with feedback
	context := formatLoopFeedback(state, result, action)

	// Prepare hook output
	output := stopLoopOutput{
		Continue:      true,
		SystemMessage: context,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(output); err != nil {
		return err
	}

	return nil // Always exit 0 - Continue field controls behavior
}

func checkCompletionMarker(text string) bool {
	if text == "" {
		return false
	}

	textLower := strings.ToLower(text)
	for _, marker := range completionMarkers {
		if strings.Contains(textLower, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

func collectCurrentDiagnostics() DiagnosticSnapshot {
	snapshot := DiagnosticSnapshot{}

	// Run ruff for Python diagnostics
	if _, err := exec.LookPath("ruff"); err == nil {
		projectRoot, err := hooks.FindProjectRoot()
		if err == nil {
			cmd := exec.Command("ruff", "check", "--output-format=json", ".")
			cmd.Dir = projectRoot
			output, _ := cmd.CombinedOutput()

			if len(output) > 0 {
				var issues []map[string]interface{}
				if err := json.Unmarshal(output, &issues); err == nil {
					for _, issue := range issues {
						code, _ := issue["code"].(string)
						if strings.HasPrefix(code, "E") || strings.HasPrefix(code, "F") {
							snapshot.ErrorCount++
						} else {
							snapshot.WarningCount++
						}
					}
				}
			}
		}
	}

	// Check TypeScript if applicable
	if _, err := exec.LookPath("tsc"); err == nil {
		projectRoot, err := hooks.FindProjectRoot()
		if err == nil {
			if hooks.FileExists(filepath.Join(projectRoot, "tsconfig.json")) {
				cmd := exec.Command("tsc", "--noEmit", "--pretty", "false")
				cmd.Dir = projectRoot
				output, _ := cmd.CombinedOutput()

				if len(output) > 0 {
					lines := strings.Split(string(output), "\n")
					for _, line := range lines {
						if strings.Contains(strings.ToLower(line), "error") {
							snapshot.ErrorCount++
						}
					}
				}
			}
		}
	}

	// Check tests
	snapshot.TestsPassed, _ = checkTests()

	return snapshot
}

func checkTests() (bool, string) {
	projectRoot, err := hooks.FindProjectRoot()
	if err != nil {
		return true, "No project root found"
	}

	// Check for Python project
	if hooks.FileExists(filepath.Join(projectRoot, "pyproject.toml")) ||
		hooks.FileExists(filepath.Join(projectRoot, "pytest.ini")) {
		if _, err := exec.LookPath("pytest"); err == nil {
			cmd := exec.Command("pytest", "--tb=no", "-q", "--no-header")
			cmd.Dir = projectRoot
			output, err := cmd.CombinedOutput()
			if err != nil {
				return false, string(output)
			}
			return true, "Tests passed"
		}
	}

	// Check for JavaScript project
	if hooks.FileExists(filepath.Join(projectRoot, "package.json")) {
		if _, err := exec.LookPath("npm"); err == nil {
			cmd := exec.Command("npm", "test", "--", "--passWithNoTests")
			cmd.Dir = projectRoot
			output, err := cmd.CombinedOutput()
			if err != nil {
				return false, string(output)
			}
			return true, "Tests passed"
		}
	}

	return true, "No test framework detected"
}

func formatLoopFeedback(state *LoopState, result CompletionResult, action string) string {
	var parts []string

	// Header
	parts = append(parts, "Ralph Loop: "+action)

	// Iteration info
	parts = append(parts, "Iteration: "+strconv.Itoa(state.Iteration)+"/"+strconv.Itoa(state.MaxIterations))

	// Current snapshot info
	if latest := state.GetLatestSnapshot(); latest != nil {
		if latest.ErrorCount > 0 || latest.WarningCount > 0 || latest.SecurityIssues > 0 {
			parts = append(parts, "Current: "+strconv.Itoa(latest.ErrorCount)+" error(s), "+
				strconv.Itoa(latest.WarningCount)+" warning(s), "+
				strconv.Itoa(latest.SecurityIssues)+" security issue(s)")
		} else {
			parts = append(parts, "Current: No issues detected")
		}

		if !latest.TestsPassed && state.Criteria.TestsPass {
			parts = append(parts, "Tests: FAILING")
		}
	}

	// Progress info
	if len(state.Snapshots) >= 2 {
		rate := state.CalculateImprovementRate()
		parts = append(parts, "Progress: "+formatProgressPercent(rate)+" improvement")
	}

	// Guidance for continuation
	if result.Guidance != "" && !result.Complete {
		parts = append(parts, "Next: "+result.Guidance)
	}

	return strings.Join(parts, " | ")
}

func formatFinalReport(state *LoopState, action string) string {
	var parts []string

	parts = append(parts, "Ralph Loop: "+action)
	parts = append(parts, "Session: "+state.SessionID)
	parts = append(parts, "Iterations: "+strconv.Itoa(state.Iteration))

	if len(state.Snapshots) >= 2 {
		rate := state.CalculateImprovementRate()
		parts = append(parts, "Total improvement: "+formatProgressPercent(rate))

		initial := state.Snapshots[0]
		latest := state.GetLatestSnapshot()

		parts = append(parts, "Initial: "+strconv.Itoa(initial.ErrorCount)+" errors, "+
			strconv.Itoa(initial.WarningCount)+" warnings")

		if latest != nil {
			parts = append(parts, "Final: "+strconv.Itoa(latest.ErrorCount)+" errors, "+
				strconv.Itoa(latest.WarningCount)+" warnings")
		}
	}

	return strings.Join(parts, " | ")
}

func formatProgressPercent(rate float64) string {
	percent := int(rate * 100)
	return strconv.Itoa(percent) + "%"
}

// formatAutoLoopFeedback creates feedback message for automatic loop continuation
func formatAutoLoopFeedback(snapshot DiagnosticSnapshot) string {
	var parts []string

	parts = append(parts, "Ralph Loop: AUTO-CONTINUE")
	parts = append(parts, "Issues detected - continuing automatically")

	// Current issues
	if snapshot.ErrorCount > 0 {
		parts = append(parts, strconv.Itoa(snapshot.ErrorCount)+" error(s) remaining")
	}
	if snapshot.SecurityIssues > 0 {
		parts = append(parts, strconv.Itoa(snapshot.SecurityIssues)+" security issue(s) remaining")
	}
	if snapshot.WarningCount > 0 {
		parts = append(parts, strconv.Itoa(snapshot.WarningCount)+" warning(s)")
	}

	// Guidance
	if snapshot.ErrorCount > 0 {
		parts = append(parts, "Next: Fix the remaining errors")
	} else if snapshot.SecurityIssues > 0 {
		parts = append(parts, "Next: Address security issues")
	}

	// Completion instruction
	parts = append(parts, "Output <jikime:done /> when complete")

	return strings.Join(parts, " | ")
}
