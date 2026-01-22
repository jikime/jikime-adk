package hookscmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// CancelLoopCmd represents the cancel-loop hook command
var CancelLoopCmd = &cobra.Command{
	Use:   "cancel-loop",
	Short: "Cancel active Ralph Loop session",
	Long: `Cancel an active Ralph Loop session.

This command stops the current loop iteration and clears the loop state.
Use this when you want to stop the loop early or start a fresh session.

Examples:
  jikime hooks cancel-loop`,
	RunE: runCancelLoop,
}

type cancelLoopOutput struct {
	Status    string `json:"status"`
	SessionID string `json:"session_id,omitempty"`
	Iteration int    `json:"iteration,omitempty"`
	Message   string `json:"message"`
}

func runCancelLoop(cmd *cobra.Command, args []string) error {
	// Load current state
	state := LoadEnhancedLoopState()

	// Check if loop is active
	if !state.Active {
		output := cancelLoopOutput{
			Status:  "no_loop",
			Message: "No active loop to cancel",
		}
		return outputCancelJSON(output)
	}

	// Get session info before clearing
	sessionID := state.SessionID
	iteration := state.Iteration

	// Update state
	state.Active = false
	state.FinalStatus = "CANCELLED"
	state.CompletionReason = "Cancelled by user"

	// Clear state file
	if err := ClearEnhancedLoopState(); err != nil {
		output := cancelLoopOutput{
			Status:    "error",
			SessionID: sessionID,
			Iteration: iteration,
			Message:   "Failed to cancel loop: " + err.Error(),
		}
		return outputCancelJSON(output)
	}

	// Generate report
	report := generateCancellationReport(state)

	output := cancelLoopOutput{
		Status:    "cancelled",
		SessionID: sessionID,
		Iteration: iteration,
		Message:   report,
	}

	return outputCancelJSON(output)
}

func generateCancellationReport(state *LoopState) string {
	report := "Ralph Loop cancelled\n"
	report += "Session: " + state.SessionID + "\n"
	report += "Iterations completed: " + itoa(state.Iteration) + "/" + itoa(state.MaxIterations) + "\n"

	// Show progress if available
	if len(state.Snapshots) >= 2 {
		rate := state.CalculateImprovementRate()
		report += "Progress: " + formatPercent(rate) + " improvement\n"

		initial := state.Snapshots[0]
		latest := state.GetLatestSnapshot()

		report += "Initial issues: " + itoa(initial.ErrorCount) + " errors, " +
			itoa(initial.WarningCount) + " warnings\n"

		if latest != nil {
			report += "Final issues: " + itoa(latest.ErrorCount) + " errors, " +
				itoa(latest.WarningCount) + " warnings\n"
		}
	}

	return report
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	if negative {
		digits = "-" + digits
	}
	return digits
}

func formatPercent(rate float64) string {
	// Convert rate (0.0-1.0) to percentage string
	percent := int(rate * 100)
	return itoa(percent) + "%"
}

func outputCancelJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(v)
}
