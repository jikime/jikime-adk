package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// OrchestratorRouteCmd represents the orchestrator-route hook command
var OrchestratorRouteCmd = &cobra.Command{
	Use:   "orchestrator-route",
	Short: "Route user request to appropriate orchestrator (J.A.R.V.I.S. or F.R.I.D.A.Y.)",
	Long: `UserPromptSubmit hook that detects migration keywords in user input
and updates the active orchestrator state file accordingly.

Migration keywords activate F.R.I.D.A.Y., all other requests activate J.A.R.V.I.S.`,
	RunE: runOrchestratorRoute,
}

// Migration keywords that trigger F.R.I.D.A.Y. activation
var migrationKeywords = []string{
	"migrate",
	"migration",
	"convert",
	"legacy",
	"transform",
	"port",
	"upgrade framework",
	"/jikime:friday",
	"/jikime:migrate",
}

// Orchestrator names
const (
	OrchestratorJARVIS = "J.A.R.V.I.S."
	OrchestratorFRIDAY = "F.R.I.D.A.Y."
	StateFileName      = "active-orchestrator"
	StateDirName       = "state"
)

func runOrchestratorRoute(cmd *cobra.Command, args []string) error {
	// Get user input from environment variable
	userInput := os.Getenv("USER_INPUT")

	// If not in env, try reading from stdin
	if userInput == "" {
		var input struct {
			UserInput string `json:"user_input"`
			Prompt    string `json:"prompt"`
		}
		decoder := json.NewDecoder(os.Stdin)
		if err := decoder.Decode(&input); err == nil {
			if input.UserInput != "" {
				userInput = input.UserInput
			} else if input.Prompt != "" {
				userInput = input.Prompt
			}
		}
	}

	// Determine active orchestrator
	orchestrator := detectOrchestrator(userInput)

	// Find project root and write state
	projectRoot, err := findProjectRoot()
	if err != nil {
		// If no project root found, try CWD
		projectRoot, _ = os.Getwd()
	}

	if err := writeOrchestratorState(projectRoot, orchestrator); err != nil {
		// Non-fatal: return success response even if state write fails
		response := HookResponse{
			Continue: true,
		}
		return writeResponse(response)
	}

	// Return success response
	response := HookResponse{
		Continue: true,
	}
	return writeResponse(response)
}

// detectOrchestrator checks user input for migration keywords
func detectOrchestrator(input string) string {
	if input == "" {
		return OrchestratorJARVIS
	}

	lower := strings.ToLower(input)
	for _, keyword := range migrationKeywords {
		if strings.Contains(lower, keyword) {
			return OrchestratorFRIDAY
		}
	}

	return OrchestratorJARVIS
}

// writeOrchestratorState writes the active orchestrator to the state file
func writeOrchestratorState(projectRoot, orchestrator string) error {
	stateDir := filepath.Join(projectRoot, ".jikime", StateDirName)

	// Create state directory if not exists
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	statePath := filepath.Join(stateDir, StateFileName)
	return os.WriteFile(statePath, []byte(orchestrator), 0644)
}

// ReadOrchestratorState reads the active orchestrator from the state file
func ReadOrchestratorState(projectRoot string) string {
	statePath := filepath.Join(projectRoot, ".jikime", StateDirName, StateFileName)

	data, err := os.ReadFile(statePath)
	if err != nil {
		return OrchestratorJARVIS // Default to J.A.R.V.I.S.
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return OrchestratorJARVIS
	}

	return name
}
