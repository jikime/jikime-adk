// Package lspsetupcmd provides the lsp-setup command for jikime-adk.
// It detects LSP server paths from common package managers and updates
// ~/.claude.json env.PATH for Claude Code's LSP integration.
package lspsetupcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewLspSetup creates the lsp-setup command.
func NewLspSetup() *cobra.Command {
	var (
		dryRun  bool
		verbose bool
	)

	cmd := &cobra.Command{
		Use:   "lsp-setup",
		Short: "Configure LSP server paths for Claude Code",
		Long: `Detect LSP server binary paths from package managers (nvm, pyenv, brew, cargo, gem)
and update ~/.claude.json env.PATH so that Claude Code can find language servers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLspSetup(dryRun, verbose)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show detected paths without updating config")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func runLspSetup(dryRun, verbose bool) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	cyan.Println("LSP Environment Setup")
	cyan.Println("─────────────────────")
	fmt.Println()

	// Detect all LSP paths
	paths := detectLSPPaths(verbose)

	if len(paths) == 0 {
		yellow.Println("No additional LSP paths detected.")
		fmt.Println("Your system PATH should already include necessary tools.")
		return nil
	}

	// Display detected paths
	fmt.Printf("Detected %d path(s):\n", len(paths))
	for _, p := range paths {
		fmt.Printf("  %s %s\n", green.Sprint("+"), p)
	}
	fmt.Println()

	if dryRun {
		dim.Println("(dry-run mode: no changes made)")
		return nil
	}

	// Update ~/.claude.json
	added, err := updateClaudeEnv(paths, verbose)
	if err != nil {
		return fmt.Errorf("failed to update ~/.claude.json: %w", err)
	}

	if added == 0 {
		dim.Println("All paths already configured in ~/.claude.json")
	} else {
		green.Printf("Added %d new path(s) to ~/.claude.json env.PATH\n", added)
	}

	return nil
}

// detectLSPPaths aggregates paths from all package manager detectors.
func detectLSPPaths(verbose bool) []string {
	dim := color.New(color.Faint)
	seen := make(map[string]bool)
	var result []string

	detectors := []struct {
		name   string
		detect func() []string
	}{
		{"nvm", getNvmPaths},
		{"pyenv", getPyenvPaths},
		{"brew", getBrewPaths},
		{"cargo", getCargoPaths},
		{"gem", getGemPaths},
	}

	for _, d := range detectors {
		paths := d.detect()
		if verbose && len(paths) > 0 {
			dim.Printf("  [%s] found %d path(s)\n", d.name, len(paths))
		}
		for _, p := range paths {
			if !seen[p] {
				seen[p] = true
				result = append(result, p)
			}
		}
	}

	sort.Strings(result)
	return result
}

// getNvmPaths detects Node.js bin directories from nvm.
func getNvmPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	nvmDir := filepath.Join(home, ".nvm", "versions", "node")
	if _, err := os.Stat(nvmDir); os.IsNotExist(err) {
		return nil
	}

	var paths []string
	entries, err := os.ReadDir(nvmDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "v") {
			binDir := filepath.Join(nvmDir, entry.Name(), "bin")
			if isValidBinDir(binDir) {
				paths = append(paths, binDir)
			}
		}
	}

	return paths
}

// getPyenvPaths detects Python shims from pyenv.
func getPyenvPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var paths []string

	// Check pyenv shims
	shimsDir := filepath.Join(home, ".pyenv", "shims")
	if isValidBinDir(shimsDir) {
		paths = append(paths, shimsDir)
	}

	// Check pyenv versions bin dirs
	versionsDir := filepath.Join(home, ".pyenv", "versions")
	if entries, err := os.ReadDir(versionsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				binDir := filepath.Join(versionsDir, entry.Name(), "bin")
				if isValidBinDir(binDir) {
					paths = append(paths, binDir)
				}
			}
		}
	}

	return paths
}

// getBrewPaths detects Homebrew bin directories.
func getBrewPaths() []string {
	var candidates []string

	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			candidates = append(candidates, "/opt/homebrew/bin")
		default:
			candidates = append(candidates, "/usr/local/bin")
		}
	case "linux":
		home, err := os.UserHomeDir()
		if err == nil {
			candidates = append(candidates, filepath.Join(home, ".linuxbrew", "bin"))
		}
		candidates = append(candidates, "/home/linuxbrew/.linuxbrew/bin")
	}

	var paths []string
	for _, c := range candidates {
		if isValidBinDir(c) {
			paths = append(paths, c)
		}
	}

	return paths
}

// getCargoPaths detects Rust cargo bin directory.
func getCargoPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	cargoDir := filepath.Join(home, ".cargo", "bin")
	if isValidBinDir(cargoDir) {
		return []string{cargoDir}
	}

	return nil
}

// getGemPaths detects Ruby gem bin directories.
func getGemPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var paths []string

	// Check GEM_HOME environment variable
	if gemHome := os.Getenv("GEM_HOME"); gemHome != "" {
		binDir := filepath.Join(gemHome, "bin")
		if isValidBinDir(binDir) {
			paths = append(paths, binDir)
		}
	}

	// Check ~/.gem/ruby/*/bin
	gemRubyDir := filepath.Join(home, ".gem", "ruby")
	if entries, err := os.ReadDir(gemRubyDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				binDir := filepath.Join(gemRubyDir, entry.Name(), "bin")
				if isValidBinDir(binDir) {
					paths = append(paths, binDir)
				}
			}
		}
	}

	return paths
}

// updateClaudeEnv reads ~/.claude.json and adds missing paths to env.PATH.
// Returns the number of newly added paths.
func updateClaudeEnv(paths []string, verbose bool) (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("cannot determine home directory: %w", err)
	}

	claudeJSON := filepath.Join(home, ".claude.json")

	// Read existing config
	var config map[string]any
	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		if os.IsNotExist(err) {
			config = make(map[string]any)
		} else {
			return 0, fmt.Errorf("cannot read %s: %w", claudeJSON, err)
		}
	} else {
		if err := json.Unmarshal(data, &config); err != nil {
			return 0, fmt.Errorf("invalid JSON in %s: %w", claudeJSON, err)
		}
	}

	// Get or create env section
	envSection, ok := config["env"].(map[string]any)
	if !ok {
		envSection = make(map[string]any)
		config["env"] = envSection
	}

	// Get current PATH value
	currentPath, _ := envSection["PATH"].(string)

	// Parse existing entries
	existingEntries := make(map[string]bool)
	if currentPath != "" {
		for entry := range strings.SplitSeq(currentPath, ":") {
			entry = strings.TrimSpace(entry)
			if entry != "" && entry != "$PATH" {
				existingEntries[entry] = true
			}
		}
	}

	// Add new paths
	added := 0
	for _, p := range paths {
		if !existingEntries[p] {
			existingEntries[p] = true
			added++
		}
	}

	if added == 0 {
		return 0, nil
	}

	// Reconstruct PATH: $PATH:sorted_entries
	var entries []string
	for entry := range existingEntries {
		entries = append(entries, entry)
	}
	sort.Strings(entries)

	newPath := "$PATH:" + strings.Join(entries, ":")
	envSection["PATH"] = newPath
	config["env"] = envSection

	// Write back
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(claudeJSON, append(output, '\n'), 0644); err != nil {
		return 0, fmt.Errorf("failed to write %s: %w", claudeJSON, err)
	}

	if verbose {
		dim := color.New(color.Faint)
		dim.Printf("  Updated: %s\n", claudeJSON)
		dim.Printf("  PATH: %s\n", newPath)
	}

	return added, nil
}

// isValidBinDir checks if a directory exists and contains at least one executable.
func isValidBinDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	// Check if at least one file exists (likely an executable)
	for _, entry := range entries {
		if !entry.IsDir() {
			return true
		}
	}

	return false
}
