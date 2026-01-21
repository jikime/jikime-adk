package hookscmd

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// StopAuditCmd represents the stop-audit hook command
var StopAuditCmd = &cobra.Command{
	Use:   "stop-audit",
	Short: "Final audit for console.log in modified files",
	Long: `Stop hook that audits modified files for console.log statements before session ends.

Features:
- Check all git-modified JS/TS files for console.log
- Warn about console.log statements that should be removed
- Help maintain clean production code

Based on everything-claude-code patterns.`,
	RunE: runStopAudit,
}

// consoleLogAuditPattern matches console.log statements
var consoleLogAuditPattern = regexp.MustCompile(`console\.(log|debug|info|warn|error)\s*\(`)

// File extensions to check
var auditExtensions = map[string]bool{
	".ts":  true,
	".tsx": true,
	".js":  true,
	".jsx": true,
	".mjs": true,
	".cjs": true,
}

type stopAuditOutput struct {
	HookSpecificOutput *stopAuditHookOutput `json:"hookSpecificOutput,omitempty"`
	SuppressOutput     bool                 `json:"suppressOutput,omitempty"`
}

type stopAuditHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func runStopAudit(cmd *cobra.Command, args []string) error {
	// Get modified files from git
	modifiedFiles := getGitModifiedFiles()
	if len(modifiedFiles) == 0 {
		output := stopAuditOutput{SuppressOutput: true}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(output)
	}

	// Check each modified file for console.log
	var warnings []string
	for _, file := range modifiedFiles {
		ext := strings.ToLower(filepath.Ext(file))
		if !auditExtensions[ext] {
			continue
		}

		// Check if file exists and contains console.log
		if matches := findConsoleLogInFile(file); len(matches) > 0 {
			warnings = append(warnings, "[Hook] WARNING: console.log found in "+file)
			for _, match := range matches {
				warnings = append(warnings, "  "+match)
			}
		}
	}

	if len(warnings) > 0 {
		warnings = append(warnings, "[Hook] Remove console.log statements before committing")
		output := stopAuditOutput{
			HookSpecificOutput: &stopAuditHookOutput{
				HookEventName:     "Stop",
				AdditionalContext: strings.Join(warnings, "\n"),
			},
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(output)
	}

	// No warnings - suppress output
	output := stopAuditOutput{SuppressOutput: true}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}

func getGitModifiedFiles() []string {
	// Check if we're in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil
	}

	// Get modified files (both staged and unstaged)
	cmd = exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to just unstaged changes
		cmd = exec.Command("git", "status", "--porcelain")
		output, err = cmd.Output()
		if err != nil {
			return nil
		}

		var files []string
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if len(line) > 3 {
				file := strings.TrimSpace(line[3:])
				if file != "" {
					files = append(files, file)
				}
			}
		}
		return files
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func findConsoleLogInFile(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if consoleLogAuditPattern.MatchString(line) {
			matches = append(matches, strings.TrimSpace(line)+" (line "+strconv.Itoa(lineNum)+")")
			if len(matches) >= 5 {
				break // Limit to first 5 matches per file
			}
		}
	}
	return matches
}

func intToString(n int) string {
	return strconv.Itoa(n)
}
