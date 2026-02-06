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
// NOTE: Keywords must be specific enough to avoid false positives in general development.
// Removed "port" (matches "port 3000", "--port") and "transform" (matches CSS/data transforms).
var migrationKeywords = []string{
	"migrate",
	"migration",
	"convert legacy",
	"legacy migration",
	"legacy code",
	"legacy system",
	"upgrade framework",
	"porting",
	"smart-rebuild",
	"smart rebuild",
	"rebuild site",
	"screenshot migration",
	"/jikime:friday",
	"/jikime:migrate",
	"/jikime:smart-rebuild",
}

// Orchestrator names
const (
	OrchestratorJARVIS = "J.A.R.V.I.S."
	OrchestratorFRIDAY = "F.R.I.D.A.Y."
	StateFileName      = "active-orchestrator"
	StateDirName       = "state"
)

// Migration artifact files that indicate an active migration project
var migrationArtifacts = []string{
	".migrate-config.yaml",
	"progress.yaml",
	"as_is_spec.md",
	"migration_plan.md",
}

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

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		projectRoot, _ = os.Getwd()
	}

	// Priority 1: Input-based detection
	// - Migration keywords → F.R.I.D.A.Y.
	// - All other non-empty input → J.A.R.V.I.S. (HARD rule)
	// - Empty input → keep current state
	orchestrator := detectOrchestrator(userInput)

	if orchestrator != "" {
		// Non-empty input detected - update state based on content
		writeOrchestratorState(projectRoot, orchestrator)
	} else if !stateFileExists(projectRoot) {
		// Priority 2: No input AND no state file - check for migration artifacts
		if hasMigrationArtifacts(projectRoot) {
			writeOrchestratorState(projectRoot, OrchestratorFRIDAY)
		} else {
			writeOrchestratorState(projectRoot, OrchestratorJARVIS)
		}
	}
	// Priority 3: No input + state file exists → keep current state

	// Return success response
	response := HookResponse{
		Continue: true,
	}
	return writeResponse(response)
}

// stateFileExists checks if the orchestrator state file exists
func stateFileExists(projectRoot string) bool {
	statePath := filepath.Join(projectRoot, ".jikime", StateDirName, StateFileName)
	_, err := os.Stat(statePath)
	return err == nil
}

// hasMigrationArtifacts checks if migration-related files exist in the project
func hasMigrationArtifacts(projectRoot string) bool {
	// Check root-level artifacts
	for _, artifact := range migrationArtifacts {
		artifactPath := filepath.Join(projectRoot, artifact)
		if _, err := os.Stat(artifactPath); err == nil {
			return true
		}
	}

	// Check migrations/ directory
	migrationsDir := filepath.Join(projectRoot, "migrations")
	if info, err := os.Stat(migrationsDir); err == nil && info.IsDir() {
		// Check for migration artifacts inside migrations/
		for _, artifact := range migrationArtifacts {
			artifactPath := filepath.Join(migrationsDir, artifact)
			if _, err := os.Stat(artifactPath); err == nil {
				return true
			}
		}
	}

	return false
}

// isActiveMigration checks if a migration is currently in progress (not just artifacts exist).
// Returns true only if progress.yaml exists with status: in_progress or status: executing.
func isActiveMigration(projectRoot string) bool {
	// Check possible progress.yaml locations
	progressPaths := []string{
		filepath.Join(projectRoot, "progress.yaml"),
		filepath.Join(projectRoot, "migrations", "progress.yaml"),
	}

	// Also check .migrate-config.yaml for artifacts_dir
	configPath := filepath.Join(projectRoot, ".migrate-config.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		// Simple extraction of artifacts_dir (avoid full YAML parsing)
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "artifacts_dir:") {
				artifactsDir := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "artifacts_dir:"))
				artifactsDir = strings.Trim(artifactsDir, "\"'")
				if artifactsDir != "" {
					progressPaths = append(progressPaths, filepath.Join(projectRoot, artifactsDir, "progress.yaml"))
				}
			}
		}
	}

	// Check each progress.yaml location
	for _, progressPath := range progressPaths {
		data, err := os.ReadFile(progressPath)
		if err != nil {
			continue
		}

		content := strings.ToLower(string(data))
		// Check for active migration status
		if strings.Contains(content, "status: in_progress") ||
			strings.Contains(content, "status: executing") ||
			strings.Contains(content, "status: \"in_progress\"") ||
			strings.Contains(content, "status: 'in_progress'") {
			return true
		}
	}

	return false
}

// detectOrchestrator checks user input to determine which orchestrator should be active.
// Per HARD rule: "Migration requests activate F.R.I.D.A.Y., all other requests activate J.A.R.V.I.S."
// Returns "" only if input is empty (no signal to process).
func detectOrchestrator(input string) string {
	if input == "" {
		return "" // No input, keep current state
	}

	lower := strings.ToLower(input)

	// Check migration keywords (FRIDAY activation)
	for _, keyword := range migrationKeywords {
		if strings.Contains(lower, keyword) {
			return OrchestratorFRIDAY
		}
	}

	// All non-migration requests activate J.A.R.V.I.S. (HARD rule compliance)
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

// clearOrchestratorState removes the orchestrator state file (resets to default)
func clearOrchestratorState() {
	projectRoot, err := findProjectRoot()
	if err != nil {
		projectRoot, _ = os.Getwd()
	}

	statePath := filepath.Join(projectRoot, ".jikime", StateDirName, StateFileName)
	os.Remove(statePath) // Ignore errors - file may not exist
}
