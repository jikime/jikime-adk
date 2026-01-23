package hookscmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// PreWriteCmd represents the pre-write hook command
var PreWriteCmd = &cobra.Command{
	Use:   "pre-write",
	Short: "Allow all documentation file creation",
	Long:  `PreToolUse hook that allows all file writes (no restrictions).`,
	RunE:  runPreWrite,
}

type preWriteOutput struct {
	SuppressOutput bool `json:"suppressOutput,omitempty"`
}

func runPreWrite(cmd *cobra.Command, args []string) error {
	// Allow all writes - no restrictions
	output := preWriteOutput{SuppressOutput: true}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
