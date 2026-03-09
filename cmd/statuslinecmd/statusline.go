// Package statuslinecmd provides the statusline command for jikime-adk.
// UI format ported from claude-statusline by Kamran Ahmed.
// Adds: Rate limit (5h window, weekly, extra credits) via Claude OAuth API.
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
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	router "jikime-adk/internal/router"
)

// ── ANSI Colors (RGB, matching claude-statusline) ──────────────────────────
const (
	colorBlue         = "\033[38;2;0;153;255m"
	colorOrange       = "\033[38;2;255;176;85m"
	colorGreen        = "\033[38;2;0;175;80m"
	colorCyan         = "\033[38;2;86;182;194m"
	colorRed          = "\033[38;2;255;85;85m"
	colorYellow       = "\033[38;2;230;200;0m"
	colorWhite        = "\033[38;2;220;220;220m"
	colorMagenta      = "\033[38;2;180;140;255m"
	colorDim          = "\033[2m"
	colorReset        = "\033[0m"
	// Bright variants for rate limit labels
	colorBrightOrange = "\033[38;2;255;165;0m"   // vivid amber-orange (current)
	colorBrightCyan   = "\033[38;2;0;220;255m"   // vivid aqua-cyan   (weekly)
)

// ── SessionContext: JSON from Claude Code stdin ────────────────────────────
type SessionContext struct {
	Model struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
	} `json:"model"`
	Version string `json:"version"`
	CWD     string `json:"cwd"`
	Session struct {
		StartTime string `json:"start_time"`
	} `json:"session"`
	ContextWindow struct {
		ContextWindowSize int `json:"context_window_size"`
		// current_usage: current turn (preferred, matches claude-statusline)
		CurrentUsage struct {
			InputTokens              int `json:"input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"current_usage"`
		// Fallback: session totals
		TotalInputTokens  int `json:"total_input_tokens"`
		TotalOutputTokens int `json:"total_output_tokens"`
	} `json:"context_window"`
}

// ── RateLimitData: Claude OAuth usage API response ─────────────────────────
type RateLimitData struct {
	FiveHour struct {
		Utilization float64 `json:"utilization"`
		ResetsAt    string  `json:"resets_at"`
	} `json:"five_hour"`
	SevenDay struct {
		Utilization float64 `json:"utilization"`
		ResetsAt    string  `json:"resets_at"`
	} `json:"seven_day"`
	ExtraUsage struct {
		IsEnabled    bool    `json:"is_enabled"`
		Utilization  float64 `json:"utilization"`
		UsedCredits  int     `json:"used_credits"`
		MonthlyLimit int     `json:"monthly_limit"`
	} `json:"extra_usage"`
}

// in-memory cache for rate limit data
var rateLimitCache struct {
	sync.RWMutex
	data    *RateLimitData
	fetched time.Time
}

const (
	rateLimitCacheDuration = 60 * time.Second
	rateLimitCacheFile     = "/tmp/claude/statusline-usage-cache.json"
)

// ── Command ────────────────────────────────────────────────────────────────

// NewStatusline creates the statusline command.
func NewStatusline() *cobra.Command {
	var demo bool
	var debug bool

	cmd := &cobra.Command{
		Use:   "statusline",
		Short: "Render statusline for Claude Code",
		Long: `Generate status information for Claude Code's statusline feature.

UI Format (claude-statusline compatible):
  Line 1: Model │ ✍️ Context% │ Directory (branch*) │ ⏱ Session │ ◐/◑ thinking

  current ●●●●●○○○○○  45% ⟳ 2:30pm
  weekly  ○○○○○○○○○○   5% ⟳ mar 15
  extra   ●●○○○○○○○○  $12.50/$50.00
  resets  mar 15`,
		Example: `  # Render statusline (Claude Code pipes JSON via stdin automatically)
  jikime statusline

  # Show demo with sample data
  jikime statusline --demo

  # Debug: show raw JSON from Claude Code
  jikime statusline --debug`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if demo {
				runDemo()
				return nil
			}
			if debug {
				return runDebug()
			}
			return runStatusline()
		},
	}

	cmd.Flags().BoolVar(&demo, "demo", false, "Show demo statusline with sample data")
	cmd.Flags().BoolVar(&debug, "debug", false, "Show raw JSON input from Claude Code")

	return cmd
}

// ── Main render pipeline ───────────────────────────────────────────────────

func runStatusline() error {
	ctx := readSessionContext()
	output := renderStatusline(ctx)
	if output != "" {
		fmt.Print(output)
	}
	return nil
}

func readSessionContext() *SessionContext {
	ctx := &SessionContext{}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return ctx
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return ctx
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for large contexts
	var input strings.Builder
	for scanner.Scan() {
		input.WriteString(scanner.Text())
	}

	if input.Len() > 0 {
		json.Unmarshal([]byte(input.String()), ctx) //nolint:errcheck
	}

	return ctx
}

// renderStatusline produces the full statusline output.
//
// Output structure:
//
//	Line 1: Claude Sonnet 4.6 │ ✍️ 45% │ jikime-adk (main*) │ ⏱ 2h30m │ ◐ thinking
//	(blank)
//	current ●●●●●○○○○○  45% ⟳ 2:30pm
//	weekly  ○○○○○○○○○○   5% ⟳ mar 15
//	extra   ●●○○○○○○○○  $12.50/$50.00
//	resets  mar 15
func renderStatusline(ctx *SessionContext) string {
	sep := fmt.Sprintf(" %s│%s ", colorDim, colorReset)

	// ── Model ──
	modelName := resolveModelName(ctx)

	// ── Context % ──
	windowSize := ctx.ContextWindow.ContextWindowSize
	if windowSize == 0 {
		windowSize = 200000
	}
	// Prefer current_usage (per-turn), fallback to session totals
	currentTokens := ctx.ContextWindow.CurrentUsage.InputTokens +
		ctx.ContextWindow.CurrentUsage.CacheCreationInputTokens +
		ctx.ContextWindow.CurrentUsage.CacheReadInputTokens
	if currentTokens == 0 {
		currentTokens = ctx.ContextWindow.TotalInputTokens + ctx.ContextWindow.TotalOutputTokens
	}
	pct := 0
	if windowSize > 0 && currentTokens > 0 {
		pct = currentTokens * 100 / windowSize
		if pct > 100 {
			pct = 100
		}
	}

	// ── Directory + Git ──
	cwd := ctx.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	dirName := filepath.Base(cwd)
	gitBranch := ""
	gitDirty := false
	if isGitRepo(cwd) {
		gitBranch = getGitBranch(cwd)
		gitDirty = isGitDirty(cwd)
	}

	// ── Session duration ──
	sessionDuration := ""
	if ctx.Session.StartTime != "" && ctx.Session.StartTime != "null" {
		sessionDuration = calcSessionDuration(ctx.Session.StartTime)
	}

	// ── Thinking status ──
	thinkingOn := isThinkingEnabled()

	// ── Orchestrator ──
	orchestrator := readOrchestrator(cwd)

	// ── Build Line 1 ──
	pctColor := colorForPct(pct)

	var line1 strings.Builder
	line1.WriteString(colorBlue)
	line1.WriteString(modelName)
	line1.WriteString(colorReset)

	line1.WriteString(sep)
	fmt.Fprintf(&line1, "%s💬 %s%s", colorMagenta, orchestrator, colorReset)

	line1.WriteString(sep)
	line1.WriteString("🧠 ")
	line1.WriteString(pctColor)
	fmt.Fprintf(&line1, "%d%%", pct)
	line1.WriteString(colorReset)

	line1.WriteString(sep)
	line1.WriteString(colorCyan)
	line1.WriteString(dirName)
	line1.WriteString(colorReset)
	if gitBranch != "" {
		dirtyMark := ""
		if gitDirty {
			dirtyMark = colorRed + "*"
		}
		fmt.Fprintf(&line1, " %s(%s%s%s)%s",
			colorGreen, gitBranch, dirtyMark, colorGreen, colorReset)
	}

	if sessionDuration != "" {
		line1.WriteString(sep)
		fmt.Fprintf(&line1, "%s⏱%s %s%s%s",
			colorDim, colorReset, colorWhite, sessionDuration, colorReset)
	}

	line1.WriteString(sep)
	if thinkingOn {
		fmt.Fprintf(&line1, "%s◐ thinking%s", colorMagenta, colorReset)
	} else {
		fmt.Fprintf(&line1, "%s◑ thinking%s", colorDim, colorReset)
	}

	// ── Rate limits (async-friendly: uses cached data) ──
	rateData := fetchRateLimits()
	rateLines := renderRateLimitLines(rateData)

	// ── Assemble output ──
	var out strings.Builder
	out.WriteString(line1.String())
	if rateLines != "" {
		out.WriteString("\n\n")
		out.WriteString(rateLines)
	}

	return out.String()
}

// ── Orchestrator resolution ────────────────────────────────────────────────

// readOrchestrator reads the active orchestrator from .jikime/state/active-orchestrator,
// walking up from cwd. Defaults to "J.A.R.V.I.S." if not found.
func readOrchestrator(cwd string) string {
	if cwd == "" {
		return "J.A.R.V.I.S."
	}
	dir := cwd
	for {
		data, err := os.ReadFile(filepath.Join(dir, ".jikime", "state", "active-orchestrator"))
		if err == nil {
			if name := strings.TrimSpace(string(data)); name != "" {
				return name
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "J.A.R.V.I.S."
}

// ── Model resolution ───────────────────────────────────────────────────────

func resolveModelName(ctx *SessionContext) string {
	// Check jikime router state first (dynamic model switching)
	if state := router.LoadState(); state != nil && state.Active {
		return fmt.Sprintf("%s/%s", state.Provider, state.Model)
	}
	if ctx.Model.DisplayName != "" {
		return ctx.Model.DisplayName
	}
	if ctx.Model.Name != "" {
		return ctx.Model.Name
	}
	return "Claude"
}

// ── Color helpers ──────────────────────────────────────────────────────────

// colorForPct returns ANSI color based on usage percentage.
// green <50%, orange 50-69%, yellow 70-89%, red 90%+
func colorForPct(pct int) string {
	switch {
	case pct >= 90:
		return colorRed
	case pct >= 70:
		return colorYellow
	case pct >= 50:
		return colorOrange
	default:
		return colorGreen
	}
}

// ── Bar chart ──────────────────────────────────────────────────────────────

// buildBar renders a ●●●●●○○○○○ bar using usage-based color (green→red).
func buildBar(pct, width int) string {
	return buildBarColor(pct, width, colorForPct(pct))
}

// buildBarColor renders a ██████░░░░ bar with an explicit fill color.
// Uses block elements (█/░) instead of circles to avoid terminal overlap issues.
func buildBarColor(pct, width int, fillColor string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := pct * width / 100
	empty := width - filled

	var bar strings.Builder
	bar.WriteString(fillColor)
	for i := 0; i < filled; i++ {
		bar.WriteString("█")
	}
	bar.WriteString(colorDim)
	for i := 0; i < empty; i++ {
		bar.WriteString("░")
	}
	bar.WriteString(colorReset)
	return bar.String()
}

// utilToPct safely converts API utilization value to integer percentage.
// Handles both 0.0-1.0 (fraction) and 0-100 (percentage) formats.
func utilToPct(u float64) int {
	if u > 1.0 {
		return int(u)
	}
	return int(u * 100)
}

// ── Git helpers ────────────────────────────────────────────────────────────

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	return cmd.Run() == nil
}

func getGitBranch(dir string) string {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func isGitDirty(dir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// ── Session duration ───────────────────────────────────────────────────────

// calcSessionDuration converts an ISO 8601 start time to human-readable elapsed time.
func calcSessionDuration(startTimeISO string) string {
	var t time.Time
	var err error

	for _, format := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	} {
		t, err = time.Parse(format, startTimeISO)
		if err == nil {
			break
		}
	}
	if err != nil {
		return ""
	}

	elapsed := time.Since(t)
	seconds := int(elapsed.Seconds())
	if seconds < 0 {
		seconds = 0
	}

	switch {
	case seconds >= 3600:
		h := seconds / 3600
		m := (seconds % 3600) / 60
		if m > 0 {
			return fmt.Sprintf("%dh%dm", h, m)
		}
		return fmt.Sprintf("%dh", h)
	case seconds >= 60:
		return fmt.Sprintf("%dm", seconds/60)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// ── Thinking status ────────────────────────────────────────────────────────

func isThinkingEnabled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		return false
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return false
	}
	if val, ok := settings["alwaysThinkingEnabled"]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// ── OAuth token resolution ─────────────────────────────────────────────────

// getOAuthToken resolves the Claude OAuth token from multiple sources:
// 1. CLAUDE_CODE_OAUTH_TOKEN env var
// 2. macOS Keychain (security find-generic-password)
// 3. ~/.claude/.credentials.json
// 4. Linux secret-tool
func getOAuthToken() string {
	// 1. Environment variable
	if tok := os.Getenv("CLAUDE_CODE_OAUTH_TOKEN"); tok != "" {
		return tok
	}

	// 2. macOS Keychain
	if _, err := exec.LookPath("security"); err == nil {
		cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
		if blob, err := cmd.Output(); err == nil && len(blob) > 0 {
			if tok := extractAccessToken(bytes.TrimSpace(blob)); tok != "" {
				return tok
			}
		}
	}

	// 3. Credentials file (~/.claude/.credentials.json)
	if home, err := os.UserHomeDir(); err == nil {
		credsPath := filepath.Join(home, ".claude", ".credentials.json")
		if data, err := os.ReadFile(credsPath); err == nil {
			if tok := extractAccessToken(data); tok != "" {
				return tok
			}
		}
	}

	// 4. Linux secret-tool
	if _, err := exec.LookPath("secret-tool"); err == nil {
		cmd := exec.Command("secret-tool", "lookup", "service", "Claude Code-credentials")
		if blob, err := cmd.Output(); err == nil && len(blob) > 0 {
			if tok := extractAccessToken(bytes.TrimSpace(blob)); tok != "" {
				return tok
			}
		}
	}

	return ""
}

// extractAccessToken extracts .claudeAiOauth.accessToken from JSON blob.
func extractAccessToken(jsonBlob []byte) string {
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBlob, &parsed); err != nil {
		return ""
	}
	oauthData, ok := parsed["claudeAiOauth"].(map[string]interface{})
	if !ok {
		return ""
	}
	tok, _ := oauthData["accessToken"].(string)
	return tok
}

// ── Rate limit fetching ────────────────────────────────────────────────────

// fetchRateLimits returns rate limit data with 60s caching (memory + file).
func fetchRateLimits() *RateLimitData {
	// 1. In-memory cache
	rateLimitCache.RLock()
	if rateLimitCache.data != nil && time.Since(rateLimitCache.fetched) < rateLimitCacheDuration {
		data := rateLimitCache.data
		rateLimitCache.RUnlock()
		return data
	}
	rateLimitCache.RUnlock()

	// 2. File cache (60s TTL)
	if info, err := os.Stat(rateLimitCacheFile); err == nil {
		if time.Since(info.ModTime()) < rateLimitCacheDuration {
			if raw, err := os.ReadFile(rateLimitCacheFile); err == nil {
				var data RateLimitData
				if json.Unmarshal(raw, &data) == nil {
					updateMemoryCache(&data)
					return &data
				}
			}
		}
	}

	// 3. Fetch from API
	token := getOAuthToken()
	if token == "" {
		return loadStaleCache()
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.anthropic.com/api/oauth/usage", nil)
	if err != nil {
		return loadStaleCache()
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("User-Agent", "claude-code/2.1.34")

	resp, err := client.Do(req)
	if err != nil {
		return loadStaleCache()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return loadStaleCache()
	}

	// Validate response has expected fields
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return loadStaleCache()
	}
	if _, ok := raw["five_hour"]; !ok {
		return loadStaleCache()
	}

	var data RateLimitData
	if err := json.Unmarshal(body, &data); err != nil {
		return loadStaleCache()
	}

	// Save to file cache
	os.MkdirAll("/tmp/claude", 0o755)                 //nolint:errcheck
	os.WriteFile(rateLimitCacheFile, body, 0o644)     //nolint:errcheck
	updateMemoryCache(&data)

	return &data
}

func updateMemoryCache(data *RateLimitData) {
	rateLimitCache.Lock()
	rateLimitCache.data = data
	rateLimitCache.fetched = time.Now()
	rateLimitCache.Unlock()
}

func loadStaleCache() *RateLimitData {
	raw, err := os.ReadFile(rateLimitCacheFile)
	if err != nil {
		return nil
	}
	var data RateLimitData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}
	return &data
}

// ── Rate limit rendering ───────────────────────────────────────────────────

// renderRateLimitLines produces the rate limit display lines.
//
//	current ●●●●●○○○○○  45% ⟳ 2:30pm   (orange — 5-hour window)
//	weekly  ○○○○○○○○○○   5% ⟳ mar 15   (cyan   — 7-day window)
//	extra   ●●○○○○○○○○  $12.50/$50.00
//	resets  mar 15
func renderRateLimitLines(data *RateLimitData) string {
	if data == nil {
		return ""
	}

	const barWidth = 10
	var lines []string

	// ── current (5-hour window) — bright orange theme ──
	fivePct := utilToPct(data.FiveHour.Utilization)
	fiveBar := buildBarColor(fivePct, barWidth, colorBrightOrange)
	fiveReset := formatResetTime(data.FiveHour.ResetsAt, "time")
	fiveResetStr := ""
	if fiveReset != "" {
		fiveResetStr = fmt.Sprintf(" %s⟳%s %s%s%s", colorDim, colorReset, colorBrightOrange, fiveReset, colorReset)
	}
	lines = append(lines, fmt.Sprintf("%scurrent%s %s %s%3d%%%s%s",
		colorBrightOrange, colorReset,
		fiveBar,
		colorBrightOrange, fivePct, colorReset,
		fiveResetStr))

	// ── weekly (7-day window) — bright cyan theme ──
	sevenPct := utilToPct(data.SevenDay.Utilization)
	sevenBar := buildBarColor(sevenPct, barWidth, colorBrightCyan)
	sevenReset := formatResetTime(data.SevenDay.ResetsAt, "datetime")
	sevenResetStr := ""
	if sevenReset != "" {
		sevenResetStr = fmt.Sprintf(" %s⟳%s %s%s%s", colorDim, colorReset, colorBrightCyan, sevenReset, colorReset)
	}
	lines = append(lines, fmt.Sprintf("%sweekly%s  %s %s%3d%%%s%s",
		colorBrightCyan, colorReset,
		sevenBar,
		colorBrightCyan, sevenPct, colorReset,
		sevenResetStr))

	// ── extra credits (if enabled) ──
	if data.ExtraUsage.IsEnabled {
		extraPct := utilToPct(data.ExtraUsage.Utilization)
		extraBar := buildBar(extraPct, barWidth)
		usedDollars := float64(data.ExtraUsage.UsedCredits) / 100.0
		limitDollars := float64(data.ExtraUsage.MonthlyLimit) / 100.0

		lines = append(lines, fmt.Sprintf("%sextra%s   %s %s$%.2f%s%s/%s%s$%.2f%s",
			colorWhite, colorReset,
			extraBar,
			colorForPct(extraPct), usedDollars, colorReset,
			colorDim, colorReset,
			colorWhite, limitDollars, colorReset))

		// resets: first of next month
		now := time.Now()
		firstOfNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		extraReset := strings.ToLower(firstOfNextMonth.Format("Jan 2"))
		lines = append(lines, fmt.Sprintf("%sresets%s  %s%s%s",
			colorDim, colorReset, colorWhite, extraReset, colorReset))
	}

	return strings.Join(lines, "\n")
}

// formatResetTime formats an ISO 8601 timestamp for display.
//
//	style "time"     → "2:30pm"
//	style "datetime" → "mar 15, 2:30pm"
//	style ""         → "mar 15"
func formatResetTime(isoStr, style string) string {
	if isoStr == "" || isoStr == "null" {
		return ""
	}

	var t time.Time
	var err error
	for _, format := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	} {
		t, err = time.Parse(format, isoStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return ""
	}

	t = t.Local()
	switch style {
	case "time":
		return strings.ToLower(t.Format("3:04pm"))
	case "datetime":
		return strings.ToLower(t.Format("Jan 2, 3:04pm"))
	default:
		return strings.ToLower(t.Format("Jan 2"))
	}
}

// ── Demo ───────────────────────────────────────────────────────────────────

func runDemo() {
	sep := fmt.Sprintf(" %s│%s ", colorDim, colorReset)

	// Demo line 1
	line1 := fmt.Sprintf(
		"%sClaude Sonnet 4.6%s%s%s💬 J.A.R.V.I.S.%s%s🧠 %s45%%%s%s%sjikime-adk%s %s(main%s%s)%s%s%s⏱%s %s2h30m%s%s%s◐ thinking%s",
		colorBlue, colorReset, sep,
		colorMagenta, colorReset, sep,
		colorGreen, colorReset, sep,
		colorCyan, colorReset,
		colorGreen, colorRed+"*", colorGreen, colorReset, sep,
		colorDim, colorReset, colorWhite, colorReset, sep,
		colorMagenta, colorReset,
	)

	// Demo rate limit data
	now := time.Now().UTC()
	demoData := &RateLimitData{}
	demoData.FiveHour.Utilization = 0.45
	demoData.FiveHour.ResetsAt = now.Add(2 * time.Hour).Format(time.RFC3339)
	demoData.SevenDay.Utilization = 0.07
	demoData.SevenDay.ResetsAt = now.AddDate(0, 0, 5).Format(time.RFC3339)
	demoData.ExtraUsage.IsEnabled = true
	demoData.ExtraUsage.Utilization = 0.25
	demoData.ExtraUsage.UsedCredits = 1250
	demoData.ExtraUsage.MonthlyLimit = 5000

	rateLines := renderRateLimitLines(demoData)

	fmt.Println()
	fmt.Println(line1)
	fmt.Println()
	fmt.Println(rateLines)
	fmt.Println()
}

// ── Debug ──────────────────────────────────────────────────────────────────

func runDebug() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Error checking stdin:", err)
		return err
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("No stdin data (not piped)")
		fmt.Println()
		fmt.Println("Usage: echo '{...}' | jikime statusline --debug")
		fmt.Println("In Claude Code, JSON is automatically piped to the statusline command.")
		return nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
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
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(rawJSON), "", "  "); err != nil {
		fmt.Println("Raw (not valid JSON):")
		fmt.Println(rawJSON)
	} else {
		fmt.Println(prettyJSON.String())
	}

	// Show parsed values
	ctx := &SessionContext{}
	if err := json.Unmarshal([]byte(rawJSON), ctx); err == nil {
		fmt.Println("╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║  📊 Extracted Values                                           ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("  Model.DisplayName:     %q\n", ctx.Model.DisplayName)
		fmt.Printf("  Model.Name:            %q\n", ctx.Model.Name)
		fmt.Printf("  CWD:                   %q\n", ctx.CWD)
		fmt.Printf("  Session.StartTime:     %q\n", ctx.Session.StartTime)
		fmt.Printf("  ContextWindow.Size:    %d\n", ctx.ContextWindow.ContextWindowSize)
		fmt.Printf("  CurrentUsage.Input:    %d\n", ctx.ContextWindow.CurrentUsage.InputTokens)
		fmt.Printf("  CurrentUsage.Cache+:   %d\n", ctx.ContextWindow.CurrentUsage.CacheCreationInputTokens)
		fmt.Printf("  CurrentUsage.Cache~:   %d\n", ctx.ContextWindow.CurrentUsage.CacheReadInputTokens)
		fmt.Printf("  Total.InputTokens:     %d\n", ctx.ContextWindow.TotalInputTokens)
		fmt.Printf("  Total.OutputTokens:    %d\n", ctx.ContextWindow.TotalOutputTokens)
		fmt.Println()
	}

	return nil
}
