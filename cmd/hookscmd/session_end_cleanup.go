package hookscmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	// P0-1.5: Clear orchestrator state (reset to default for next session)
	clearOrchestratorState()

	// P0-2: Check for uncommitted Git changes
	uncommittedWarning := checkGitUncommittedChanges()
	if uncommittedWarning != "" {
		result.UncommittedWarning = uncommittedWarning
	}

	// P0-3: Generate session summary
	result.SessionSummary = generateCleanupSummary(stats, uncommittedWarning)

	// P0-4: Send desktop notification
	sendDesktopNotification("JikiME-ADK Session Ended", result.SessionSummary)

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

// sendDesktopNotification sends a cross-platform desktop notification
// Supports macOS (osascript), Linux (notify-send), and Windows (PowerShell)
func sendDesktopNotification(title, message string) {
	// Skip if JIKIME_NO_NOTIFY is set
	if os.Getenv("JIKIME_NO_NOTIFY") != "" {
		return
	}

	// Escape quotes in message for shell safety
	escapedTitle := strings.ReplaceAll(title, `"`, `\"`)
	escapedMessage := strings.ReplaceAll(message, `"`, `\"`)
	escapedMessage = strings.ReplaceAll(escapedMessage, "\n", " ")

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS: Use osascript for native notifications
		script := `display notification "` + escapedMessage + `" with title "` + escapedTitle + `"`
		cmd = exec.Command("osascript", "-e", script)

	case "linux":
		// Linux: Use notify-send (requires libnotify)
		cmd = exec.Command("notify-send", title, message)

	case "windows":
		// Windows: Use PowerShell toast notification
		psScript := `
		[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
		$template = "<toast><visual><binding template='ToastText02'><text id='1'>` + escapedTitle + `</text><text id='2'>` + escapedMessage + `</text></binding></visual></toast>"
		$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
		$xml.LoadXml($template)
		$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
		[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("JikiME-ADK").Show($toast)
		`
		cmd = exec.Command("powershell", "-Command", psScript)

	default:
		// Unsupported OS - silently skip
		return
	}

	// Run notification in background, ignore errors (non-critical)
	_ = cmd.Start()
}
