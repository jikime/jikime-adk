// Package webchatcmd provides CLI commands for webchat management.
package webchatcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// webchatDir returns the webchat installation directory (~/.jikime/webchat).
func webchatDir() string {
	if d := os.Getenv("JIKIME_WEBCHAT_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jikime", "webchat")
}

// isInstalled checks if webchat is installed and built.
func isInstalled() bool {
	dir := webchatDir()
	// package.json 존재 확인
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		return false
	}
	// node_modules 존재 확인
	if _, err := os.Stat(filepath.Join(dir, "node_modules")); err != nil {
		return false
	}
	return true
}

// isBuilt checks if webchat has been built (.next directory exists).
func isBuilt() bool {
	dir := webchatDir()
	info, err := os.Stat(filepath.Join(dir, ".next"))
	return err == nil && info.IsDir()
}

// findPnpm returns the path to pnpm binary.
func findPnpm() (string, error) {
	path, err := exec.LookPath("pnpm")
	if err != nil {
		return "", fmt.Errorf("pnpm을 찾을 수 없습니다. 먼저 설치해주세요: npm install -g pnpm")
	}
	return path, nil
}

// findNode returns the path to node binary.
func findNode() (string, error) {
	path, err := exec.LookPath("node")
	if err != nil {
		return "", fmt.Errorf("Node.js를 찾을 수 없습니다. Node.js 22+ 를 설치해주세요")
	}
	return path, nil
}

// findTsx returns the path to tsx binary (local or global).
func findTsx(dir string) string {
	// local node_modules first
	local := filepath.Join(dir, "node_modules", ".bin", "tsx")
	if _, err := os.Stat(local); err == nil {
		return local
	}
	// global
	if p, err := exec.LookPath("tsx"); err == nil {
		return p
	}
	return ""
}

// NewWebchat creates the `jikime webchat` command group.
func NewWebchat() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webchat",
		Aliases: []string{"wc"},
		Short:   "Manage and run the webchat UI",
		Long: `Manage and run the JikiME webchat UI.

The webchat provides a web-based interface for Claude Code
with team management, GitHub issues integration, and harness automation.

Examples:
  jikime webchat start              # Start webchat server (port 4000)
  jikime webchat start --port 3000  # Start on custom port
  jikime webchat install            # Install/update webchat dependencies
  jikime webchat status             # Check installation status
  jikime webchat build              # Build webchat for production`,
	}

	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newBuildCmd())

	return cmd
}
