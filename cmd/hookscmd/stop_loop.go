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
- Check completion conditions (zero errors, tests pass)
- Either continue loop or signal completion

Exit codes:
- 0: Loop complete or disabled
- 1: Continue loop (more work needed)`,
	RunE: runStopLoop,
}

const (
	disableEnvVar       = "JIKIME_DISABLE_LOOP_CONTROLLER"
	loopActiveEnvVar    = "JIKIME_LOOP_ACTIVE"
	loopIterationEnvVar = "JIKIME_LOOP_ITERATION"
	maxIterations       = 10
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

type loopState struct {
	Active           bool     `json:"active"`
	Iteration        int      `json:"iteration"`
	MaxIterations    int      `json:"max_iterations"`
	LastErrorCount   int      `json:"last_error_count"`
	LastWarningCount int      `json:"last_warning_count"`
	FilesModified    []string `json:"files_modified,omitempty"`
	CompletionReason string   `json:"completion_reason,omitempty"`
}

type completionStatus struct {
	ZeroErrors       bool
	ZeroWarnings     bool
	TestsPass        bool
	AllConditionsMet bool
	ErrorCount       int
	WarningCount     int
	TestDetails      string
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
	state := loadLoopState()

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
		state.CompletionReason = "Completion promise detected"
		clearLoopState()

		output := stopLoopOutput{
			HookSpecificOutput: &stopLoopHookOutput{
				HookEventName:     "Stop",
				AdditionalContext: "Ralph Loop: COMPLETE - <promise>DONE</promise> detected",
			},
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.Encode(output)
		return nil
	}

	// Check completion conditions
	status := checkCompletionConditions()

	// Update state
	state.LastErrorCount = status.ErrorCount
	state.LastWarningCount = status.WarningCount

	// Determine action
	var action string
	var exitCode int

	if status.AllConditionsMet {
		// Loop complete
		state.Active = false
		state.CompletionReason = "All conditions met"
		action = "COMPLETE - All conditions satisfied"
		clearLoopState()
		exitCode = 0
	} else if state.Iteration >= state.MaxIterations {
		// Max iterations reached
		state.Active = false
		state.CompletionReason = "Max iterations reached"
		action = "STOPPED - Max iterations (" + strconv.Itoa(state.MaxIterations) + ") reached"
		clearLoopState()
		exitCode = 0
	} else {
		// Continue loop
		state.Iteration++
		action = "CONTINUE - Issues remain"
		saveLoopState(state)
		exitCode = 1
	}

	// Format output
	context := formatLoopOutput(state, status, action)

	// Build guidance for Claude
	guidance := ""
	if exitCode == 1 {
		var issues []string
		if status.ErrorCount > 0 {
			issues = append(issues, "Fix "+strconv.Itoa(status.ErrorCount)+" error(s)")
		}
		if status.WarningCount > 0 && !status.ZeroWarnings {
			issues = append(issues, "Address "+strconv.Itoa(status.WarningCount)+" warning(s)")
		}
		if !status.TestsPass {
			issues = append(issues, "Fix failing tests")
		}
		if len(issues) > 0 {
			guidance = "\nNext actions: " + strings.Join(issues, ", ")
		}
	}

	// Prepare hook output
	output := stopLoopOutput{
		HookSpecificOutput: &stopLoopHookOutput{
			HookEventName:     "Stop",
			AdditionalContext: context + guidance,
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

func loadLoopState() loopState {
	// First check environment variables
	if active := os.Getenv(loopActiveEnvVar); active != "" {
		if active == "1" || strings.ToLower(active) == "true" || strings.ToLower(active) == "yes" {
			iteration := 0
			if iterStr := os.Getenv(loopIterationEnvVar); iterStr != "" {
				if iter, err := strconv.Atoi(iterStr); err == nil {
					iteration = iter
				}
			}
			return loopState{Active: true, Iteration: iteration, MaxIterations: maxIterations}
		}
	}

	// Then check state file
	statePath := getLoopStatePath()
	data, err := os.ReadFile(statePath)
	if err != nil {
		return loopState{MaxIterations: maxIterations}
	}

	var state loopState
	if err := json.Unmarshal(data, &state); err != nil {
		return loopState{MaxIterations: maxIterations}
	}

	if state.MaxIterations == 0 {
		state.MaxIterations = maxIterations
	}

	return state
}

func saveLoopState(state loopState) {
	statePath := getLoopStatePath()

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(statePath), 0755)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(statePath, data, 0644)
}

func clearLoopState() {
	statePath := getLoopStatePath()
	os.Remove(statePath)
}

func getLoopStatePath() string {
	jikimeDir, err := hooks.FindJikimeDir()
	if err != nil {
		// Fallback to current directory
		return ".jikime_loop_state.json"
	}
	return filepath.Join(jikimeDir, "cache", ".jikime_loop_state.json")
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

func checkCompletionConditions() completionStatus {
	status := completionStatus{
		ZeroErrors:   true,
		ZeroWarnings: true,
		TestsPass:    true,
	}

	// Check for errors using ruff (Python)
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
							status.ErrorCount++
						} else {
							status.WarningCount++
						}
					}
				}
			}
		}
	}

	status.ZeroErrors = status.ErrorCount == 0
	status.ZeroWarnings = status.WarningCount == 0

	// Check tests
	status.TestsPass, status.TestDetails = checkLoopTests()

	// All conditions met if no errors and tests pass
	status.AllConditionsMet = status.ZeroErrors && status.TestsPass

	return status
}

func checkLoopTests() (bool, string) {
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

func formatLoopOutput(state loopState, status completionStatus, action string) string {
	parts := []string{"Ralph Loop: " + action}

	if state.Active || state.CompletionReason != "" {
		parts = append(parts, "Iteration: "+strconv.Itoa(state.Iteration)+"/"+strconv.Itoa(state.MaxIterations))
	}

	if status.ErrorCount > 0 || status.WarningCount > 0 {
		parts = append(parts, "Errors: "+strconv.Itoa(status.ErrorCount))
		parts = append(parts, "Warnings: "+strconv.Itoa(status.WarningCount))
	}

	if status.TestDetails != "" && status.TestDetails != "No test framework detected" {
		if status.TestsPass {
			parts = append(parts, "Tests: PASS")
		} else {
			parts = append(parts, "Tests: FAIL")
		}
	}

	if status.AllConditionsMet {
		parts = append(parts, "Status: COMPLETE")
	} else if state.Active {
		parts = append(parts, "Status: CONTINUE")
	}

	return strings.Join(parts, " | ")
}
