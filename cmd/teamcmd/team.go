// Package teamcmd provides CLI commands for jikime team orchestration.
package teamcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// dataDir returns the jikime data root (~/.jikime by default).
// Override with JIKIME_DATA_DIR env var.
func dataDir() string {
	if d := os.Getenv("JIKIME_DATA_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jikime")
}

// teamDir returns the directory for a specific team.
func teamDir(teamName string) string {
	return filepath.Join(dataDir(), "teams", teamName)
}

// NewTeam creates the `jikime team` command group.
func NewTeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "team",
		Aliases: []string{"t"},
		Short:   "Manage multi-agent teams",
		Long: `Manage multi-agent teams for parallel task execution.

Each team consists of a leader agent and one or more worker agents.
Agents communicate via file-based inboxes and share a task store.

Examples:
  jikime team create my-team --workers 3
  jikime team spawn  my-team --role worker
  jikime team status my-team
  jikime team tasks  my-team
  jikime team launch --template leader-worker --goal "implement auth"`,
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newSpawnCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newIdentityCmd())
	cmd.AddCommand(newSessionCmd())
	cmd.AddCommand(newPlanCmd())
	cmd.AddCommand(newLifecycleCmd())
	cmd.AddCommand(newTasksCmd())
	cmd.AddCommand(newInboxCmd())
	cmd.AddCommand(newWorkspaceCmd())
	cmd.AddCommand(newTemplateCmd())
	cmd.AddCommand(newLaunchCmd())
	cmd.AddCommand(newBoardCmd())
	cmd.AddCommand(newBudgetCmd())
	cmd.AddCommand(newDiscoverCmd())

	return cmd
}

// printJSON is a minimal JSON pretty-printer for CLI output.
func printJSON(v interface{}) {
	// Uses fmt as a simple fallback; callers use encoding/json directly.
	fmt.Fprintf(os.Stdout, "%v\n", v)
}

// printJSONList encodes a slice as indented JSON to stdout.
func printJSONList(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// mkdirAll creates a directory and all parents, returning a wrapped error.
func mkdirAll(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return nil
}

// writeJSON atomically writes v as indented JSON to path.
// It writes to a .tmp file first, then renames to ensure atomicity.
func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("writeJSON marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writeJSON write: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("writeJSON rename: %w", err)
	}
	return nil
}
