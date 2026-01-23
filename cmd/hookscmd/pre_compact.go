package hookscmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// PreCompactCmd represents the pre-compact hook command
var PreCompactCmd = &cobra.Command{
	Use:   "pre-compact",
	Short: "Save state before context compaction",
	Long: `PreCompact hook that saves important context before Claude Code compacts the conversation.

Features:
- Save current session state
- Preserve important context markers
- Create checkpoint for recovery

Based on JikiME-ADK memory-persistence pattern.`,
	RunE: runPreCompact,
}

type preCompactInput struct {
	SessionID string `json:"session_id,omitempty"`
}

type preCompactOutput struct {
	HookSpecificOutput *preCompactHookOutput `json:"hookSpecificOutput,omitempty"`
}

type preCompactHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

type compactCheckpoint struct {
	Timestamp   string   `json:"timestamp"`
	SessionID   string   `json:"session_id,omitempty"`
	WorkingDir  string   `json:"working_dir,omitempty"`
	GitBranch   string   `json:"git_branch,omitempty"`
	ActiveFiles []string `json:"active_files,omitempty"`
}

func runPreCompact(cmd *cobra.Command, args []string) error {
	// Read JSON input from stdin
	var input preCompactInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input) // Ignore decode errors, use defaults

	// Get project root
	projectRoot := getCompactProjectRoot()
	if projectRoot == "" {
		return nil
	}

	// Create checkpoint directory
	memoryDir := filepath.Join(projectRoot, ".jikime", "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return nil // Fail silently
	}

	// Build checkpoint data
	checkpoint := compactCheckpoint{
		Timestamp:  time.Now().Format(time.RFC3339),
		SessionID:  input.SessionID,
		WorkingDir: projectRoot,
	}

	// Get git branch if available
	if gitDir := filepath.Join(projectRoot, ".git"); dirExists(gitDir) {
		if headRef, err := os.ReadFile(filepath.Join(gitDir, "HEAD")); err == nil {
			checkpoint.GitBranch = string(headRef)
		}
	}

	// Save checkpoint
	checkpointPath := filepath.Join(memoryDir, "pre-compact-checkpoint.json")
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return nil
	}
	if err := os.WriteFile(checkpointPath, data, 0644); err != nil {
		return nil
	}

	// Output hook response with context reminder
	output := preCompactOutput{
		HookSpecificOutput: &preCompactHookOutput{
			HookEventName:     "PreCompact",
			AdditionalContext: "[Pre-Compact] Session state saved. Resume context available at .jikime/memory/",
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}

func getCompactProjectRoot() string {
	if projectDir := os.Getenv("CLAUDE_PROJECT_DIR"); projectDir != "" {
		return projectDir
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
