// Package statuslinecmd provides the statusline command for jikime-adk.
// This renders status information for Claude Code's statusline feature.
package statuslinecmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"jikime-adk-v2/version"
)

// SessionContext represents the JSON context from Claude Code
type SessionContext struct {
	Model struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
	} `json:"model"`
	Version      string `json:"version"`
	CWD          string `json:"cwd"`
	OutputStyle  struct {
		Name string `json:"name"`
	} `json:"output_style"`
	ContextWindow struct {
		ContextWindowSize int `json:"context_window_size"`
		TotalInputTokens  int `json:"total_input_tokens"`
		CurrentUsage      struct {
			InputTokens              int `json:"input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"current_usage"`
	} `json:"context_window"`
	Statusline struct {
		Mode string `json:"mode"`
	} `json:"statusline"`
}

// StatuslineData contains all the information for the statusline
type StatuslineData struct {
	Model           string
	ClaudeVersion   string
	Version         string
	Branch          string
	GitStatus       string
	Duration        string
	Directory       string
	ActiveTask      string
	OutputStyle     string
	UpdateAvailable bool
	LatestVersion   string
	ContextWindow   string
	MemoryUsage     string
}

// StatuslineConfig represents the configuration from statusline-config.yaml
type StatuslineConfig struct {
	Statusline struct {
		Enabled           bool   `yaml:"enabled"`
		Mode              string `yaml:"mode"`
		RefreshIntervalMS int    `yaml:"refresh_interval_ms"`
		Display           struct {
			Model           bool `yaml:"model"`
			Version         bool `yaml:"version"`
			ContextWindow   bool `yaml:"context_window"`
			OutputStyle     bool `yaml:"output_style"`
			MemoryUsage     bool `yaml:"memory_usage"`
			TodoCount       bool `yaml:"todo_count"`
			Branch          bool `yaml:"branch"`
			GitStatus       bool `yaml:"git_status"`
			Duration        bool `yaml:"duration"`
			Directory       bool `yaml:"directory"`
			ActiveTask      bool `yaml:"active_task"`
			UpdateIndicator bool `yaml:"update_indicator"`
		} `yaml:"display"`
		Format struct {
			MaxBranchLength int    `yaml:"max_branch_length"`
			TruncateWith    string `yaml:"truncate_with"`
			Separator       string `yaml:"separator"`
			Icons           struct {
				Git           string `yaml:"git"`
				GitStatus     string `yaml:"git_status"`
				Model         string `yaml:"model"`
				ClaudeVersion string `yaml:"claude_version"`
				ContextWindow string `yaml:"context_window"`
				OutputStyle   string `yaml:"output_style"`
				Duration      string `yaml:"duration"`
				Update        string `yaml:"update"`
				Project       string `yaml:"project"`
			} `yaml:"icons"`
		} `yaml:"format"`
		Cache struct {
			GitTTLSeconds    int `yaml:"git_ttl_seconds"`
			UpdateTTLSeconds int `yaml:"update_ttl_seconds"`
		} `yaml:"cache"`
	} `yaml:"statusline"`
}

// Global configuration and cache
var (
	configOnce   sync.Once
	globalConfig *StatuslineConfig

	// Update check cache
	updateCache struct {
		sync.RWMutex
		checked   time.Time
		available bool
		version   string
	}
)

// loadConfig loads the statusline configuration from YAML file
func loadConfig() *StatuslineConfig {
	configOnce.Do(func() {
		globalConfig = &StatuslineConfig{}
		// Set defaults
		globalConfig.Statusline.Enabled = true
		globalConfig.Statusline.Mode = "extended"
		globalConfig.Statusline.Display.Model = true
		globalConfig.Statusline.Display.Branch = true
		globalConfig.Statusline.Display.GitStatus = true
		globalConfig.Statusline.Display.ContextWindow = true
		globalConfig.Statusline.Display.OutputStyle = true
		globalConfig.Statusline.Display.ActiveTask = true
		globalConfig.Statusline.Display.UpdateIndicator = true
		globalConfig.Statusline.Format.Separator = " | "
		globalConfig.Statusline.Format.MaxBranchLength = 30
		globalConfig.Statusline.Format.TruncateWith = "..."
		globalConfig.Statusline.Cache.GitTTLSeconds = 10
		globalConfig.Statusline.Cache.UpdateTTLSeconds = 600

		// Try to load from config file
		configPaths := []string{
			filepath.Join(".jikime", "config", "statusline-config.yaml"),
		}

		// Also check home directory
		if home, err := os.UserHomeDir(); err == nil {
			configPaths = append(configPaths, filepath.Join(home, ".jikime", "config", "statusline-config.yaml"))
		}

		for _, path := range configPaths {
			if data, err := os.ReadFile(path); err == nil {
				yaml.Unmarshal(data, globalConfig)
				break
			}
		}
	})
	return globalConfig
}

// checkForUpdate checks if a new version of jikime-adk is available
func checkForUpdate() (bool, string) {
	config := loadConfig()
	ttl := time.Duration(config.Statusline.Cache.UpdateTTLSeconds) * time.Second
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	updateCache.RLock()
	if time.Since(updateCache.checked) < ttl {
		available, ver := updateCache.available, updateCache.version
		updateCache.RUnlock()
		return available, ver
	}
	updateCache.RUnlock()

	// Check GitHub releases API
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/jikime/jikime-adk/releases/latest", nil)
	if err != nil {
		return false, ""
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "jikime-adk-v2/"+version.String())

	resp, err := client.Do(req)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return false, ""
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := version.String()
	available := compareVersions(currentVersion, latestVersion) < 0

	// Update cache
	updateCache.Lock()
	updateCache.checked = time.Now()
	updateCache.available = available
	updateCache.version = latestVersion
	updateCache.Unlock()

	return available, latestVersion
}

// compareVersions compares two version strings
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1Parts) {
			fmt.Sscanf(v1Parts[i], "%d", &n1)
		}
		if i < len(v2Parts) {
			fmt.Sscanf(v2Parts[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// NewStatusline creates the statusline command.
func NewStatusline() *cobra.Command {
	var mode string
	var demo bool
	var pretty bool

	cmd := &cobra.Command{
		Use:   "statusline",
		Short: "Render statusline for Claude Code",
		Long: `Generate status information for Claude Code's statusline feature.

The statusline displays:
  🤖 Model     - AI model name (e.g., Opus 4.5)
  💰 Context   - Context window usage (e.g., 15K/200K)
  💬 Style     - Output style name
  📁 Directory - Current project directory
  📊 Status    - Git changes (+staged M modified ?untracked)
  🔀 Branch    - Git branch name
  ⏱️  Duration  - Session duration
  🎯 Task      - Active task indicator
  📦 Version   - JikiME-ADK version

Modes:
  extended (default) - Full information with all sections
  compact            - Condensed view (80 chars max)
  minimal            - Essential info only (40 chars max)`,
		Example: `  # Show statusline (extended mode)
  jikime statusline

  # Show compact statusline
  jikime statusline --mode compact

  # Show demo with sample data
  jikime statusline --demo

  # Show in pretty box format
  jikime statusline --pretty`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if demo {
				runDemo()
				return nil
			}
			if pretty {
				runPretty()
				return nil
			}
			return runStatusline(mode)
		},
	}

	cmd.Flags().StringVarP(&mode, "mode", "m", "", "Display mode (compact, extended, minimal)")
	cmd.Flags().BoolVar(&demo, "demo", false, "Show demo statusline with sample data")
	cmd.Flags().BoolVar(&pretty, "pretty", false, "Show pretty box formatted output")

	return cmd
}

func runStatusline(modeOverride string) error {
	// Read session context from stdin
	sessionContext := readSessionContext()

	// Determine display mode
	mode := modeOverride
	if mode == "" {
		mode = sessionContext.Statusline.Mode
	}
	if mode == "" {
		mode = os.Getenv("JIKIME_STATUSLINE_MODE")
	}
	if mode == "" {
		mode = "extended"
	}

	// Build statusline data
	data := buildStatuslineData(sessionContext)

	// Render and output
	statusline := renderStatusline(data, mode)
	if statusline != "" {
		fmt.Print(statusline)
	}

	return nil
}

func readSessionContext() *SessionContext {
	ctx := &SessionContext{}

	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ctx
	}

	// Only read if stdin is a pipe or has data
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return ctx
	}

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var input strings.Builder
	for scanner.Scan() {
		input.WriteString(scanner.Text())
	}

	if input.Len() > 0 {
		json.Unmarshal([]byte(input.String()), ctx)
	}

	return ctx
}

func buildStatuslineData(ctx *SessionContext) *StatuslineData {
	config := loadConfig()

	data := &StatuslineData{
		Version: version.String(),
	}

	// Extract model name
	if ctx.Model.DisplayName != "" {
		data.Model = ctx.Model.DisplayName
	} else if ctx.Model.Name != "" {
		data.Model = ctx.Model.Name
	} else {
		data.Model = "Unknown"
	}

	// Claude version
	data.ClaudeVersion = ctx.Version

	// Directory
	if ctx.CWD != "" {
		data.Directory = filepath.Base(ctx.CWD)
		if data.Directory == "" {
			data.Directory = filepath.Base(filepath.Dir(ctx.CWD))
		}
	}
	if data.Directory == "" {
		data.Directory = "project"
	}

	// Output style
	data.OutputStyle = ctx.OutputStyle.Name

	// Context window
	data.ContextWindow = extractContextWindow(ctx)

	// Git info
	data.Branch, data.GitStatus = collectGitInfo()

	// Duration (from metrics tracker - simplified)
	data.Duration = collectDuration()

	// Active task (from Alfred detector - simplified)
	data.ActiveTask = collectActiveTask()

	// Memory usage (if enabled)
	if config.Statusline.Display.MemoryUsage {
		data.MemoryUsage = collectMemoryUsage()
	}

	// Check for updates (if enabled)
	if config.Statusline.Display.UpdateIndicator {
		data.UpdateAvailable, data.LatestVersion = checkForUpdate()
	}

	return data
}

func extractContextWindow(ctx *SessionContext) string {
	if ctx.ContextWindow.ContextWindowSize == 0 {
		return ""
	}

	var currentTokens int
	if ctx.ContextWindow.CurrentUsage.InputTokens > 0 {
		currentTokens = ctx.ContextWindow.CurrentUsage.InputTokens +
			ctx.ContextWindow.CurrentUsage.CacheCreationInputTokens +
			ctx.ContextWindow.CurrentUsage.CacheReadInputTokens
	} else {
		currentTokens = ctx.ContextWindow.TotalInputTokens
	}

	if currentTokens > 0 {
		return fmt.Sprintf("%s/%s",
			formatTokenCount(currentTokens),
			formatTokenCount(ctx.ContextWindow.ContextWindowSize))
	}

	return ""
}

func formatTokenCount(tokens int) string {
	if tokens >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	}
	if tokens >= 1000 {
		return fmt.Sprintf("%dK", tokens/1000)
	}
	return fmt.Sprintf("%d", tokens)
}

func collectGitInfo() (branch, status string) {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return "N/A", ""
	}

	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return "N/A", ""
	}

	// Get branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		branch = "N/A"
	} else {
		branch = strings.TrimSpace(string(branchOutput))
	}

	// Get status counts
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return branch, ""
	}

	var staged, modified, untracked int
	lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		indexStatus := line[0]
		workTreeStatus := line[1]

		if indexStatus == '?' {
			untracked++
		} else {
			if indexStatus != ' ' && indexStatus != '?' {
				staged++
			}
			if workTreeStatus != ' ' && workTreeStatus != '?' {
				modified++
			}
		}
	}

	if staged > 0 || modified > 0 || untracked > 0 {
		status = fmt.Sprintf("+%d M%d ?%d", staged, modified, untracked)
	}

	return branch, status
}

func collectDuration() string {
	// Try to read session start time from metrics file
	home, err := os.UserHomeDir()
	if err != nil {
		return "0m"
	}

	metricsFile := filepath.Join(home, ".jikime", "metrics", "session.json")
	data, err := os.ReadFile(metricsFile)
	if err != nil {
		return "0m"
	}

	var metrics struct {
		StartTime string `json:"start_time"`
	}
	if err := json.Unmarshal(data, &metrics); err != nil {
		return "0m"
	}

	if metrics.StartTime == "" {
		return "0m"
	}

	startTime, err := time.Parse(time.RFC3339, metrics.StartTime)
	if err != nil {
		return "0m"
	}

	duration := time.Since(startTime)
	if duration.Hours() >= 1 {
		return fmt.Sprintf("%.1fh", duration.Hours())
	}
	return fmt.Sprintf("%dm", int(duration.Minutes()))
}

func collectActiveTask() string {
	// Try to read active task from Alfred state file
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	taskFile := filepath.Join(home, ".jikime", "state", "active_task.json")
	data, err := os.ReadFile(taskFile)
	if err != nil {
		return ""
	}

	var task struct {
		Command string `json:"command"`
		Stage   string `json:"stage"`
	}
	if err := json.Unmarshal(data, &task); err != nil {
		return ""
	}

	if task.Command != "" {
		if task.Stage != "" {
			return fmt.Sprintf("[%s-%s]", strings.ToUpper(task.Command), task.Stage)
		}
		return fmt.Sprintf("[%s]", strings.ToUpper(task.Command))
	}

	return ""
}

// collectMemoryUsage collects system memory usage information
func collectMemoryUsage() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Use Sys (total memory obtained from OS) for more accurate representation
	// This includes heap, stack, and other memory
	memMB := float64(memStats.Sys) / (1024 * 1024)

	return formatMemorySize(memMB)
}

// formatMemorySize formats memory size in MB to human-readable string
func formatMemorySize(sizeMB float64) string {
	if sizeMB >= 1024 {
		return fmt.Sprintf("%.1fGB", sizeMB/1024)
	} else if sizeMB >= 100 {
		return fmt.Sprintf("%.0fMB", sizeMB)
	} else if sizeMB >= 10 {
		return fmt.Sprintf("%.1fMB", sizeMB)
	}
	return fmt.Sprintf("%.2fMB", sizeMB)
}

// Pretty icons for statusline
const (
	IconModel       = "🤖"
	IconContext     = "💰"
	IconStyle       = "💬"
	IconDirectory   = "📁"
	IconGitStatus   = "📊"
	IconMemory      = "💾"
	IconBranch      = "🔀"
	IconTask        = "🎯"
	IconUpdate      = "🔄"
	IconVersion     = "📦"
	IconClaude      = "🤖"
	IconTime        = "⏱️"
)

func renderStatusline(data *StatuslineData, mode string) string {
	var parts []string
	separator := " │ "

	switch mode {
	case "minimal":
		// Minimal: 🤖 Model | 💰 Context
		if data.Model != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconModel, data.Model))
		}
		if data.ContextWindow != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconContext, data.ContextWindow))
		}

	case "compact":
		// Compact: 🤖 Model | 💰 Context | 💬 Style | 📁 Dir | 📊 Status | 💾 Memory | 🔀 Branch
		if data.Model != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconModel, data.Model))
		}
		if data.ContextWindow != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconContext, data.ContextWindow))
		}
		if data.OutputStyle != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconStyle, data.OutputStyle))
		}
		if data.Directory != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconDirectory, data.Directory))
		}
		if data.GitStatus != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconGitStatus, data.GitStatus))
		}
		if data.MemoryUsage != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconMemory, data.MemoryUsage))
		}
		if data.Branch != "" && data.Branch != "N/A" {
			parts = append(parts, fmt.Sprintf("%s %s", IconBranch, truncateBranch(data.Branch, 20)))
		}

	default: // extended
		// Extended: full information with icons
		if data.Model != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconModel, data.Model))
		}
		if data.ContextWindow != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconContext, data.ContextWindow))
		}
		if data.OutputStyle != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconStyle, data.OutputStyle))
		}
		if data.Directory != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconDirectory, data.Directory))
		}
		if data.GitStatus != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconGitStatus, data.GitStatus))
		}
		if data.MemoryUsage != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconMemory, data.MemoryUsage))
		}
		if data.Branch != "" && data.Branch != "N/A" {
			parts = append(parts, fmt.Sprintf("%s %s", IconBranch, truncateBranch(data.Branch, 30)))
		}
		if data.Duration != "" && data.Duration != "0m" {
			parts = append(parts, fmt.Sprintf("%s %s", IconTime, data.Duration))
		}
		if data.ActiveTask != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconTask, data.ActiveTask))
		}
		if data.Version != "" {
			parts = append(parts, fmt.Sprintf("%s v%s", IconVersion, data.Version))
		}
		if data.UpdateAvailable && data.LatestVersion != "" {
			parts = append(parts, fmt.Sprintf("%s %s available", IconUpdate, data.LatestVersion))
		}
	}

	return strings.Join(parts, separator)
}

// truncateBranch truncates branch name intelligently
func truncateBranch(branch string, maxLen int) string {
	if len(branch) <= maxLen {
		return branch
	}

	// Try to preserve SPEC ID in feature branches
	if strings.Contains(branch, "SPEC") {
		parts := strings.Split(branch, "-")
		for i, part := range parts {
			if strings.Contains(part, "SPEC") && i+1 < len(parts) {
				specTruncated := strings.Join(parts[:i+2], "-")
				if len(specTruncated) <= maxLen {
					return specTruncated
				}
			}
		}
	}

	// Simple truncation with ellipsis
	if maxLen > 3 {
		return branch[:maxLen-1] + "…"
	}
	return branch[:maxLen]
}

// runDemo shows demo statusline with sample data
func runDemo() {
	demoData := &StatuslineData{
		Model:         "Opus 4.5",
		ClaudeVersion: "2.0.46",
		Version:       version.String(),
		OutputStyle:   "Mr.Alfred",
		Directory:     "jikime-adk",
		Branch:        "main",
		GitStatus:     "+0 M5 ?5",
		ContextWindow: "15K/200K",
		Duration:      "45m",
		ActiveTask:    "IMPLEMENT",
		MemoryUsage:   "128MB",
	}

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  🎨 JikiME-ADK Statusline Demo                                                  │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│                                                                                 │")
	fmt.Printf("│  📐 Compact Mode:                                                               │\n")
	fmt.Printf("│  %s\n", renderStatusline(demoData, "compact"))
	fmt.Println("│                                                                                 │")
	fmt.Printf("│  📏 Extended Mode:                                                              │\n")
	fmt.Printf("│  %s\n", renderStatusline(demoData, "extended"))
	fmt.Println("│                                                                                 │")
	fmt.Printf("│  📍 Minimal Mode:                                                               │\n")
	fmt.Printf("│  %s\n", renderStatusline(demoData, "minimal"))
	fmt.Println("│                                                                                 │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Pretty box
	fmt.Println("🎁 Pretty Box Format:")
	printPrettyBox(renderStatusline(demoData, "compact"))
	fmt.Println()
}

// runPretty shows statusline in pretty box format
func runPretty() {
	sessionContext := readSessionContext()
	data := buildStatuslineData(sessionContext)
	content := renderStatusline(data, "extended")
	printPrettyBox(content)
}

// printPrettyBox prints content in a decorative box
func printPrettyBox(content string) {
	// Strip ANSI codes for length calculation
	visibleLen := len(stripAnsi(content))

	// Calculate box width
	boxWidth := visibleLen + 4
	if boxWidth < 60 {
		boxWidth = 60
	}

	// Build decorative box
	fmt.Println()
	fmt.Print("╭")
	for i := 0; i < boxWidth; i++ {
		fmt.Print("─")
	}
	fmt.Println("╮")

	fmt.Print("│ ")
	fmt.Print(content)
	padding := boxWidth - visibleLen - 2
	for i := 0; i < padding; i++ {
		fmt.Print(" ")
	}
	fmt.Println(" │")

	fmt.Print("╰")
	for i := 0; i < boxWidth; i++ {
		fmt.Print("─")
	}
	fmt.Println("╯")
	fmt.Println()
}

// stripAnsi removes ANSI escape codes from string for length calculation
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}
