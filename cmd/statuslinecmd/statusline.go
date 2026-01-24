// Package statuslinecmd provides the statusline command for jikime-adk.
// This renders status information for Claude Code's statusline feature.
package statuslinecmd

import (
	"bufio"
	"bytes"
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

	router "jikime-adk/internal/router"
	"jikime-adk/version"
)

// SessionContext represents the JSON context from Claude Code
type SessionContext struct {
	Model struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
	} `json:"model"`
	Version     string `json:"version"`
	CWD         string `json:"cwd"`
	OutputStyle struct {
		Name string `json:"name"`
	} `json:"output_style"` // kept for backward compat, statusline uses state file
	Cost struct {
		TotalCostUSD      float64 `json:"total_cost_usd"`
		TotalDurationMS   int     `json:"total_duration_ms"`
		TotalLinesAdded   int     `json:"total_lines_added"`
		TotalLinesRemoved int     `json:"total_lines_removed"`
	} `json:"cost"`
	ContextWindow struct {
		TotalInputTokens  int `json:"total_input_tokens"`
		TotalOutputTokens int `json:"total_output_tokens"`
		ContextWindowSize int `json:"context_window_size"`
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
	Orchestrator    string
	UpdateAvailable bool
	LatestVersion   string
	ContextWindow   string
	MemoryUsage     string
	TokenCost       string // e.g., "$0.12"
	TokensUsed      int    // raw token count for cost calculation
	ContextPercent  int    // context usage percentage for progress bar
}

// StatuslineConfig represents the configuration from statusline-config.yaml
type StatuslineConfig struct {
	Statusline struct {
		Enabled           bool   `yaml:"enabled"`
		Mode              string `yaml:"mode"`
		RefreshIntervalMS int    `yaml:"refresh_interval_ms"`
		Display struct {
			Model           bool `yaml:"model"`
			Version         bool `yaml:"version"`
			ContextWindow   bool `yaml:"context_window"`
			Orchestrator    bool `yaml:"orchestrator"`
			MemoryUsage     bool `yaml:"memory_usage"`
			TodoCount       bool `yaml:"todo_count"`
			Branch          bool `yaml:"branch"`
			GitStatus       bool `yaml:"git_status"`
			Duration        bool `yaml:"duration"`
			Directory       bool `yaml:"directory"`
			ActiveTask      bool `yaml:"active_task"`
			UpdateIndicator bool `yaml:"update_indicator"`
			TokenCost       bool `yaml:"token_cost"`
			ProgressBar     bool `yaml:"progress_bar"`
		} `yaml:"display"`
		TokenCost struct {
			InputPricePerMTok  float64 `yaml:"input_price_per_mtok"`
			OutputPricePerMTok float64 `yaml:"output_price_per_mtok"`
		} `yaml:"token_cost"`
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
				Orchestrator  string `yaml:"orchestrator"`
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
	// Update check cache
	updateCache struct {
		sync.RWMutex
		checked   time.Time
		available bool
		version   string
	}
)

// loadConfig loads the statusline configuration from YAML file
// projectPath is the project directory (ctx.CWD from Claude Code)
func loadConfig(projectPath string) *StatuslineConfig {
	config := &StatuslineConfig{}
	// Set defaults
	config.Statusline.Enabled = true
	config.Statusline.Mode = "extended"
	config.Statusline.Display.Model = true
	config.Statusline.Display.Branch = true
	config.Statusline.Display.GitStatus = true
	config.Statusline.Display.ContextWindow = true
	config.Statusline.Display.Orchestrator = true
	config.Statusline.Display.ActiveTask = true
	config.Statusline.Display.UpdateIndicator = true
	config.Statusline.Display.TokenCost = true
	config.Statusline.Display.ProgressBar = true
	// Token cost defaults (Claude Opus pricing)
	config.Statusline.TokenCost.InputPricePerMTok = 15.0  // $15 per 1M input tokens
	config.Statusline.TokenCost.OutputPricePerMTok = 75.0 // $75 per 1M output tokens
	config.Statusline.Format.Separator = " ┃ " // Geek style separator
	config.Statusline.Format.MaxBranchLength = 30
	config.Statusline.Format.TruncateWith = "..."
	config.Statusline.Cache.GitTTLSeconds = 10
	config.Statusline.Cache.UpdateTTLSeconds = 600

	// Build config paths - prioritize project directory, then home directory
	var configPaths []string

	// 1. Project directory (from Claude Code's cwd)
	if projectPath != "" {
		configPaths = append(configPaths, filepath.Join(projectPath, ".jikime", "config", "statusline-config.yaml"))
	}

	// 2. Current working directory (fallback)
	configPaths = append(configPaths, filepath.Join(".jikime", "config", "statusline-config.yaml"))

	// 3. Home directory
	if home, err := os.UserHomeDir(); err == nil {
		configPaths = append(configPaths, filepath.Join(home, ".jikime", "config", "statusline-config.yaml"))
	}

	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			yaml.Unmarshal(data, config)
			break
		}
	}

	return config
}

// checkForUpdate checks if a new version of jikime-adk is available
func checkForUpdate(projectPath string) (bool, string) {
	config := loadConfig(projectPath)
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
	req.Header.Set("User-Agent", "jikime-adk/"+version.String())

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
	var debug bool

	cmd := &cobra.Command{
		Use:   "statusline",
		Short: "Render statusline for Claude Code",
		Long: `Generate status information for Claude Code's statusline feature.

The statusline displays:
  🤖 Model     - AI model name (e.g., Opus 4.5)
  💰 Context   - Context window usage with progress bar
  💵 Cost      - Estimated token cost (e.g., $0.12)
  📁 Directory - Current project directory
  🔀 Branch    - Git branch and status
  💾 Memory    - Memory usage
  ⏱️  Duration  - Session duration
  📦 Version   - JikiME-ADK version

Modes:
  extended (default) - Balanced view with progress bar
  compact            - Condensed view with essential info
  minimal            - Model and context only
  geek               - Full developer mode with all features`,
		Example: `  # Show statusline (extended mode with progress bar)
  jikime statusline

  # Show compact statusline
  jikime statusline --mode compact

  # Show full geek mode with all features
  jikime statusline --mode geek

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
			if debug {
				return runDebug()
			}
			return runStatusline(mode)
		},
	}

	cmd.Flags().StringVarP(&mode, "mode", "m", "", "Display mode (compact, extended, minimal)")
	cmd.Flags().BoolVar(&demo, "demo", false, "Show demo statusline with sample data")
	cmd.Flags().BoolVar(&pretty, "pretty", false, "Show pretty box formatted output")
	cmd.Flags().BoolVar(&debug, "debug", false, "Show raw JSON input from Claude Code")

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
	config := loadConfig(ctx.CWD)

	data := &StatuslineData{
		Version: version.String(),
	}

	// Extract model name — check router state first
	if state := router.LoadState(); state != nil && state.Active {
		data.Model = fmt.Sprintf("%s/%s", state.Provider, state.Model)
	} else if ctx.Model.DisplayName != "" {
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

	// Orchestrator (read from state file, dynamic switching)
	data.Orchestrator = readOrchestratorFromState(ctx.CWD)

	// Context window (tokens)
	data.ContextWindow = extractContextWindow(ctx)

	// Token cost - use actual cost from Claude Code directly
	if config.Statusline.Display.TokenCost && ctx.Cost.TotalCostUSD > 0 {
		data.TokenCost = formatCost(ctx.Cost.TotalCostUSD)
	}

	// Token usage from Claude Code (total_input_tokens + total_output_tokens)
	totalTokens := ctx.ContextWindow.TotalInputTokens + ctx.ContextWindow.TotalOutputTokens
	data.TokensUsed = totalTokens

	// Context percentage for progress bar
	if ctx.ContextWindow.ContextWindowSize > 0 && totalTokens > 0 {
		data.ContextPercent = (totalTokens * 100) / ctx.ContextWindow.ContextWindowSize
	}

	// Duration - use actual duration from Claude Code
	if ctx.Cost.TotalDurationMS > 0 {
		data.Duration = formatDuration(ctx.Cost.TotalDurationMS)
	}

	// Git info (only if in a git repo)
	data.Branch, data.GitStatus = collectGitInfo(ctx.CWD)

	// Memory usage (if enabled)
	if config.Statusline.Display.MemoryUsage {
		data.MemoryUsage = collectMemoryUsage()
	}

	// Check for updates (if enabled)
	if config.Statusline.Display.UpdateIndicator {
		data.UpdateAvailable, data.LatestVersion = checkForUpdate(ctx.CWD)
	}

	return data
}

// readOrchestratorFromState reads the active orchestrator from .jikime/state/active-orchestrator
func readOrchestratorFromState(cwd string) string {
	if cwd == "" {
		return "J.A.R.V.I.S."
	}

	// Walk up directories to find .jikime
	dir := cwd
	for {
		statePath := filepath.Join(dir, ".jikime", "state", "active-orchestrator")
		data, err := os.ReadFile(statePath)
		if err == nil {
			name := strings.TrimSpace(string(data))
			if name != "" {
				return name
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "J.A.R.V.I.S." // Default
}

func extractContextWindow(ctx *SessionContext) string {
	if ctx.ContextWindow.ContextWindowSize == 0 {
		return ""
	}

	// Use total_input_tokens + total_output_tokens from Claude Code
	currentTokens := ctx.ContextWindow.TotalInputTokens + ctx.ContextWindow.TotalOutputTokens
	if currentTokens > 0 {
		return fmt.Sprintf("%s/%s",
			formatTokenCount(currentTokens),
			formatTokenCount(ctx.ContextWindow.ContextWindowSize))
	}

	return ""
}

// formatTokenCount formats token count with K/M units for display
func formatTokenCount(tokens int) string {
	if tokens >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	}
	if tokens >= 1000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	}
	return fmt.Sprintf("%d", tokens)
}

// formatCost formats the cost in USD
func formatCost(cost float64) string {
	if cost < 0.001 {
		return fmt.Sprintf("$%.4f", cost)
	} else if cost < 0.01 {
		return fmt.Sprintf("$%.3f", cost)
	} else if cost < 1.0 {
		return fmt.Sprintf("$%.2f", cost)
	}
	return fmt.Sprintf("$%.1f", cost)
}

// formatDuration formats duration in milliseconds to human readable
func formatDuration(ms int) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	seconds := ms / 1000
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	remainingMin := minutes % 60
	if remainingMin > 0 {
		return fmt.Sprintf("%dh%dm", hours, remainingMin)
	}
	return fmt.Sprintf("%dh", hours)
}


func collectGitInfo(projectPath string) (branch, status string) {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return "", ""
	}

	// Use project path if provided, otherwise use current directory
	workDir := projectPath
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return "", ""
		}
	}

	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		return "", ""
	}

	// Get branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = workDir
	branchOutput, err := branchCmd.Output()
	if err != nil {
		branch = ""
	} else {
		branch = strings.TrimSpace(string(branchOutput))
	}

	// Get status counts
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = workDir
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


// renderProgressBar renders a progress bar for context usage
func renderProgressBar(percent int, width int) string {
	if width <= 0 {
		width = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := (percent * width) / 100
	// Show at least 1 filled block if percent > 0
	if percent > 0 && filled == 0 {
		filled = 1
	}
	empty := width - filled

	var bar strings.Builder
	for i := 0; i < filled; i++ {
		bar.WriteString(ProgressFilled)
	}
	for i := 0; i < empty; i++ {
		bar.WriteString(ProgressEmpty)
	}

	return bar.String()
}

// Pretty icons for statusline
const (
	IconModel     = "🤖"
	IconContext   = "💰"
	IconOrchestrator = "💬"
	IconDirectory = "📁"
	IconGitStatus = "📊"
	IconMemory    = "💾"
	IconBranch    = "🔀"
	IconTask      = "🎯"
	IconUpdate    = "🔄"
	IconVersion   = "📦"
	IconClaude    = "🤖"
	IconTime      = "⏱️"
	// New icons
	IconCost = "💵"
)

// Progress bar characters
const (
	ProgressFilled = "▰"
	ProgressEmpty  = "▱"
)

func renderStatusline(data *StatuslineData, mode string) string {
	var parts []string
	separator := " ┃ " // Geek style separator

	switch mode {
	case "minimal":
		// Minimal: 🤖 Model | Progress Bar
		if data.Model != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconModel, data.Model))
		}
		if data.ContextPercent > 0 {
			parts = append(parts, fmt.Sprintf("%s %d%%", renderProgressBar(data.ContextPercent, 10), data.ContextPercent))
		} else if data.ContextWindow != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconContext, data.ContextWindow))
		}

	case "compact":
		// Compact with progress bar style
		if data.Model != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconModel, data.Model))
		}
		// Context with progress bar
		if data.ContextPercent > 0 {
			parts = append(parts, fmt.Sprintf("%s %s %d%%", renderProgressBar(data.ContextPercent, 10), data.ContextWindow, data.ContextPercent))
		} else if data.ContextWindow != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconContext, data.ContextWindow))
		}
		// Token cost
		if data.TokenCost != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconCost, data.TokenCost))
		}
		// Orchestrator
		if data.Orchestrator != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconOrchestrator, data.Orchestrator))
		}
		// Git info
		if data.Branch != "" && data.Branch != "N/A" {
			branchInfo := fmt.Sprintf("%s %s", IconBranch, truncateBranch(data.Branch, 15))
			if data.GitStatus != "" {
				branchInfo = fmt.Sprintf("%s %s", branchInfo, data.GitStatus)
			}
			parts = append(parts, branchInfo)
		}
		// Memory usage
		if data.MemoryUsage != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconMemory, data.MemoryUsage))
		}
		// Update indicator
		if data.UpdateAvailable && data.LatestVersion != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconUpdate, data.LatestVersion))
		}

	case "geek":
		// Full geek mode with all bells and whistles
		if data.Model != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconModel, data.Model))
		}
		// Context with fancy progress bar
		if data.ContextPercent > 0 {
			bar := renderProgressBar(data.ContextPercent, 10)
			contextColor := getContextColor(data.ContextPercent)
			parts = append(parts, fmt.Sprintf("%s%s %s (%d%%)%s", contextColor, bar, data.ContextWindow, data.ContextPercent, "\033[0m"))
		} else if data.ContextWindow != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconContext, data.ContextWindow))
		}
		// Token cost
		if data.TokenCost != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconCost, data.TokenCost))
		}
		// Orchestrator
		if data.Orchestrator != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconOrchestrator, data.Orchestrator))
		}
		// Directory
		if data.Directory != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconDirectory, data.Directory))
		}
		// Git info combined
		if data.Branch != "" && data.Branch != "N/A" {
			branchInfo := fmt.Sprintf("%s %s", IconBranch, truncateBranch(data.Branch, 20))
			if data.GitStatus != "" {
				branchInfo = fmt.Sprintf("%s %s", branchInfo, data.GitStatus)
			}
			parts = append(parts, branchInfo)
		}
		// Memory usage
		if data.MemoryUsage != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconMemory, data.MemoryUsage))
		}
		// Duration
		if data.Duration != "" && data.Duration != "0m" {
			parts = append(parts, fmt.Sprintf("%s %s", IconTime, data.Duration))
		}
		// Version with update indicator
		if data.Version != "" {
			versionStr := fmt.Sprintf("%s v%s", IconVersion, data.Version)
			if data.UpdateAvailable && data.LatestVersion != "" {
				versionStr = fmt.Sprintf("%s %s→%s", versionStr, IconUpdate, data.LatestVersion)
			}
			parts = append(parts, versionStr)
		}

	default: // extended (default) - balanced geek style
		if data.Model != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconModel, data.Model))
		}
		// Context with progress bar
		if data.ContextPercent > 0 {
			parts = append(parts, fmt.Sprintf("%s %s", renderProgressBar(data.ContextPercent, 10), data.ContextWindow))
		} else if data.ContextWindow != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconContext, data.ContextWindow))
		}
		// Token cost
		if data.TokenCost != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconCost, data.TokenCost))
		}
		// Orchestrator
		if data.Orchestrator != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconOrchestrator, data.Orchestrator))
		}
		// Directory
		if data.Directory != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconDirectory, data.Directory))
		}
		// Git info
		if data.Branch != "" && data.Branch != "N/A" {
			branchInfo := fmt.Sprintf("%s %s", IconBranch, truncateBranch(data.Branch, 25))
			if data.GitStatus != "" {
				branchInfo = fmt.Sprintf("%s %s", branchInfo, data.GitStatus)
			}
			parts = append(parts, branchInfo)
		}
		// Memory usage
		if data.MemoryUsage != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconMemory, data.MemoryUsage))
		}
		// Duration
		if data.Duration != "" && data.Duration != "0m" {
			parts = append(parts, fmt.Sprintf("%s %s", IconTime, data.Duration))
		}
		// Active task
		if data.ActiveTask != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconTask, data.ActiveTask))
		}
		// Version
		if data.Version != "" {
			parts = append(parts, fmt.Sprintf("%s v%s", IconVersion, data.Version))
		}
		// Update indicator
		if data.UpdateAvailable && data.LatestVersion != "" {
			parts = append(parts, fmt.Sprintf("%s %s", IconUpdate, data.LatestVersion))
		}
	}

	return strings.Join(parts, separator)
}

// getContextColor returns ANSI color code based on context usage percentage
func getContextColor(percent int) string {
	if percent >= 80 {
		return "\033[31m" // Red
	} else if percent >= 50 {
		return "\033[33m" // Yellow
	}
	return "\033[32m" // Green
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
		Model:          "Opus 4.5",
		ClaudeVersion:  "2.0.46",
		Version:        version.String(),
		Orchestrator:   "J.A.R.V.I.S.",
		Directory:      "jikime-adk",
		Branch:         "main",
		GitStatus:      "+0 M5 ?5",
		ContextWindow:  "15K/200K",
		ContextPercent: 7,
		Duration:       "45m",
		ActiveTask:     "IMPLEMENT",
		MemoryUsage:    "128MB",
		TokenCost:      "$0.23",
	}

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  🎨 JikiME-ADK Statusline Demo - Developer Geek Edition                                                   ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                                                           ║")
	fmt.Println("║  📍 Minimal Mode:                                                                                         ║")
	fmt.Printf("║  %s\n", renderStatusline(demoData, "minimal"))
	fmt.Println("║                                                                                                           ║")
	fmt.Println("║  📐 Compact Mode:                                                                                         ║")
	fmt.Printf("║  %s\n", renderStatusline(demoData, "compact"))
	fmt.Println("║                                                                                                           ║")
	fmt.Println("║  📏 Extended Mode (Default):                                                                              ║")
	fmt.Printf("║  %s\n", renderStatusline(demoData, "extended"))
	fmt.Println("║                                                                                                           ║")
	fmt.Println("║  🔥 Geek Mode (Full Features):                                                                            ║")
	fmt.Printf("║  %s\n", renderStatusline(demoData, "geek"))
	fmt.Println("║                                                                                                           ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Progress bar legend
	fmt.Println("📊 Progress Bar Legend:")
	fmt.Printf("   Context Usage: %s = 0%% | %s = 50%% | %s = 100%%\n",
		renderProgressBar(0, 10), renderProgressBar(50, 10), renderProgressBar(100, 10))
	fmt.Println()

	// Pretty box
	fmt.Println("🎁 Pretty Box Format:")
	printPrettyBox(renderStatusline(demoData, "extended"))
	fmt.Println()
}

// runDebug shows raw JSON input from Claude Code for debugging
func runDebug() error {
	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Error checking stdin:", err)
		return err
	}

	// Only read if stdin is a pipe or has data
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("No stdin data (not piped)")
		fmt.Println()
		fmt.Println("Usage: echo '{...}' | jikime statusline --debug")
		fmt.Println()
		fmt.Println("In Claude Code, the JSON is automatically piped to statusline command.")
		return nil
	}

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var input strings.Builder
	for scanner.Scan() {
		input.WriteString(scanner.Text())
	}

	rawJSON := input.String()
	if rawJSON == "" {
		fmt.Println("Empty stdin input")
		return nil
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  🔍 Claude Code Statusline Debug                               ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                ║")
	fmt.Println("║  Raw JSON from Claude Code:                                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Pretty print JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(rawJSON), "", "  "); err != nil {
		fmt.Println("Raw (not valid JSON):")
		fmt.Println(rawJSON)
	} else {
		fmt.Println(prettyJSON.String())
	}

	fmt.Println()

	// Parse and show what we extracted
	ctx := &SessionContext{}
	if err := json.Unmarshal([]byte(rawJSON), ctx); err != nil {
		fmt.Println("Parse error:", err)
		return nil
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  📊 Extracted Values                                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Model.DisplayName:     %q\n", ctx.Model.DisplayName)
	fmt.Printf("  Model.Name:            %q\n", ctx.Model.Name)
	fmt.Printf("  Version:               %q\n", ctx.Version)
	fmt.Printf("  CWD:                   %q\n", ctx.CWD)
	fmt.Printf("  OutputStyle.Name:      %q\n", ctx.OutputStyle.Name)
	fmt.Printf("  Cost.TotalCostUSD:     $%.6f\n", ctx.Cost.TotalCostUSD)
	fmt.Printf("  Cost.TotalDurationMS:  %dms\n", ctx.Cost.TotalDurationMS)
	fmt.Printf("  ContextWindow.Size:    %d\n", ctx.ContextWindow.ContextWindowSize)
	fmt.Printf("  ContextWindow.Input:   %d\n", ctx.ContextWindow.TotalInputTokens)
	fmt.Printf("  ContextWindow.Output:  %d\n", ctx.ContextWindow.TotalOutputTokens)
	fmt.Printf("  Statusline.Mode:       %q\n", ctx.Statusline.Mode)
	fmt.Println()

	return nil
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
