package hookscmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/hooks"
)

// SessionEndCleanupCmd represents the session-end-cleanup hook command
var SessionEndCleanupCmd = &cobra.Command{
	Use:   "session-end-cleanup",
	Short: "Cleanup and state saving on session end",
	Long: `SessionEnd hook that performs cleanup and state saving tasks:
- Clean up temporary files and cache
- Warn about uncommitted Git changes
- Generate session summary`,
	RunE: runSessionEndCleanup,
}

type cleanupResult struct {
	Hook                 string          `json:"hook"`
	Success              bool            `json:"success"`
	ExecutionTimeSeconds float64         `json:"execution_time_seconds"`
	CleanupStats         cleanupStats    `json:"cleanup_stats"`
	UncommittedWarning   string          `json:"uncommitted_warning,omitempty"`
	SessionSummary       string          `json:"session_summary"`
	Timestamp            string          `json:"timestamp"`
	Performance          map[string]bool `json:"performance,omitempty"`
}

type cleanupStats struct {
	TempCleaned  int `json:"temp_cleaned"`
	CacheCleaned int `json:"cache_cleaned"`
	TotalCleaned int `json:"total_cleaned"`
}

func runSessionEndCleanup(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	result := cleanupResult{
		Hook:      "session_end__auto_cleanup",
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
		Performance: map[string]bool{
			"git_manager_used": true,
		},
	}

	// P0-1: Clean up temporary files
	stats := cleanupOldFiles()
	result.CleanupStats = stats

	// P0-2: Check for uncommitted Git changes
	uncommittedWarning := checkGitUncommittedChanges()
	if uncommittedWarning != "" {
		result.UncommittedWarning = uncommittedWarning
	}

	// P0-3: Generate session summary
	result.SessionSummary = generateCleanupSummary(stats, uncommittedWarning)

	// Record execution time
	result.ExecutionTimeSeconds = time.Since(startTime).Seconds()

	// Output result
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}

func cleanupOldFiles() cleanupStats {
	stats := cleanupStats{}

	// Find jikime directory
	jikimeDir, err := hooks.FindJikimeDir()
	if err != nil {
		return stats
	}

	// Clean up temp directory (files older than 7 days)
	tempDir := filepath.Join(jikimeDir, "temp")
	if info, err := os.Stat(tempDir); err == nil && info.IsDir() {
		stats.TempCleaned = cleanupDirectory(tempDir, 7)
	}

	// Clean up cache directory (files older than 7 days)
	cacheDir := filepath.Join(jikimeDir, "cache")
	if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
		stats.CacheCleaned = cleanupDirectory(cacheDir, 7)
	}

	stats.TotalCleaned = stats.TempCleaned + stats.CacheCleaned

	return stats
}

func cleanupDirectory(directory string, daysOld int) int {
	cleanedCount := 0
	cutoffTime := time.Now().AddDate(0, 0, -daysOld)

	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		entryPath := filepath.Join(directory, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if older than cutoff
		if info.ModTime().Before(cutoffTime) {
			if entry.IsDir() {
				if err := os.RemoveAll(entryPath); err == nil {
					cleanedCount++
				}
			} else {
				if err := os.Remove(entryPath); err == nil {
					cleanedCount++
				}
			}
		}
	}

	return cleanedCount
}

func checkGitUncommittedChanges() string {
	// Get project root
	projectRoot, err := hooks.FindProjectRoot()
	if err != nil {
		return ""
	}

	// Check git status
	status, err := hooks.RunCommandInDir(projectRoot, "git", "status", "--porcelain")
	if err != nil {
		return ""
	}

	if strings.TrimSpace(status) == "" {
		return ""
	}

	// Count uncommitted files
	lines := strings.Split(strings.TrimSpace(status), "\n")
	lineCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			lineCount++
		}
	}

	if lineCount > 0 {
		return "Warning: " + strconv.Itoa(lineCount) + " uncommitted files detected - Consider committing or stashing changes"
	}

	return ""
}

func generateCleanupSummary(stats cleanupStats, uncommittedWarning string) string {
	var summaryLines []string
	summaryLines = append(summaryLines, "Session Ended")

	// Cleanup information
	if stats.TotalCleaned > 0 {
		summaryLines = append(summaryLines, "  - Cleaned: "+strconv.Itoa(stats.TotalCleaned)+" temp files")
	}

	// Uncommitted warning
	if uncommittedWarning != "" {
		summaryLines = append(summaryLines, "  - "+uncommittedWarning)
	}

	return strings.Join(summaryLines, "\n")
}
