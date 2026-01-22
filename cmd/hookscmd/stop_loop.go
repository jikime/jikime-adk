package hookscmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"jikime-adk-v2/internal/hooks"
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

// Completion promise markers
var completionMarkers = []string{
	"<promise>DONE</promise>",
	"<promise>COMPLETE</promise>",
	"<ralph:done />",
	"<ralph:complete />",
	"<alfred:complete />",
	"<jikime:done />",
	"<jikime:complete />",
}

type stopLoopInput struct {
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type stopLoopOutput struct {
	HookSpecificOutput *stopLoopHookOutput `json:"hookSpecificOutput,omitempty"`
}

type stopLoopHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func runStopLoop(cmd *cobra.Command, args []string) error {
	// Check if loop controller is disabled
	if disabled := os.Getenv(disableEnvVar); disabled != "" {
		if disabled == "1" || strings.ToLower(disabled) == "true" || strings.ToLower(disabled) == "yes" {
			return nil
		}
	}

	// Load current loop state
	state := LoadEnhancedLoopState()

	// If loop is not active, just exit
	if !state.Active {
		return nil
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

	// PRIORITY CHECK: Completion promise marker
	if checkCompletionPromise(conversationText) {
		state.Active = false
		state.FinalStatus = "COMPLETE"
		state.CompletionReason = "Completion promise detected"
		ClearEnhancedLoopState()

		output := stopLoopOutput{
			HookSpecificOutput: &stopLoopHookOutput{
				HookEventName:     "Stop",
				AdditionalContext: formatFinalReport(state, "COMPLETE - Completion promise detected"),
			},
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.Encode(output)
		return nil
	}

	// Collect current diagnostics and update snapshot
	currentSnapshot := collectCurrentDiagnostics()
	state.AddSnapshot(currentSnapshot)

	// Evaluate completion conditions
	result := state.EvaluateCompletion()

	// Determine action
	var action string
	var exitCode int

	if result.Complete {
		// Loop complete - all conditions met
		state.Active = false
		state.FinalStatus = "COMPLETE"
		state.CompletionReason = result.Reason
		action = "COMPLETE - " + result.Reason
		ClearEnhancedLoopState()
		exitCode = 0
	} else if state.Iteration >= state.MaxIterations {
		// Max iterations reached
		state.Active = false
		state.FinalStatus = "STOPPED"
		state.CompletionReason = "Max iterations reached"
		action = "STOPPED - Max iterations (" + strconv.Itoa(state.MaxIterations) + ") reached"
		ClearEnhancedLoopState()
		exitCode = 0
	} else if state.IsStagnant() {
		// Stagnation detected
		state.Active = false
		state.FinalStatus = "STOPPED"
		state.CompletionReason = "Stagnation - no improvement detected"
		action = "STOPPED - No improvement in last " + strconv.Itoa(state.Criteria.StagnationLimit) + " iterations"
		ClearEnhancedLoopState()
		exitCode = 0
	} else {
		// Continue loop
		state.Iteration++
		action = "CONTINUE"
		SaveEnhancedLoopState(state)
		exitCode = 1
	}

	// Format output with feedback
	context := formatLoopFeedback(state, result, action)

	// Prepare hook output
	output := stopLoopOutput{
		HookSpecificOutput: &stopLoopHookOutput{
			HookEventName:     "Stop",
			AdditionalContext: context,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(output); err != nil {
		return err
	}

	os.Exit(exitCode)
	return nil
}

func checkCompletionPromise(text string) bool {
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
