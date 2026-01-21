// Package rankcmd provides rank-related commands for jikime-adk.
package rankcmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	defaultBaseURL = "https://rank.mo.ai.kr"
	apiVersion     = "v1"
)

// RankCredentials represents user credentials for the rank API
type RankCredentials struct {
	APIKey    string `json:"api_key"`
	Username  string `json:"username"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

// RankInfo represents ranking information for a specific period
type RankInfo struct {
	Position          int     `json:"position"`
	CompositeScore    float64 `json:"compositeScore"`
	TotalParticipants int     `json:"totalParticipants"`
}

// UserRank represents complete user ranking data
type UserRank struct {
	Username      string    `json:"username"`
	Daily         *RankInfo `json:"daily,omitempty"`
	Weekly        *RankInfo `json:"weekly,omitempty"`
	Monthly       *RankInfo `json:"monthly,omitempty"`
	AllTime       *RankInfo `json:"allTime,omitempty"`
	TotalTokens   int       `json:"totalTokens"`
	TotalSessions int       `json:"totalSessions"`
	InputTokens   int       `json:"inputTokens"`
	OutputTokens  int       `json:"outputTokens"`
	LastUpdated   string    `json:"lastUpdated"`
}

// NewRank creates the rank command group.
func NewRank() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rank",
		Short: "Jikime Rank - Token usage leaderboard",
		Long:  "Track your Claude Code token usage and compete on the leaderboard.\nVisit https://rank.mo.ai.kr for the web dashboard.",
	}

	cmd.AddCommand(newRegisterCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newExcludeCmd())
	cmd.AddCommand(newIncludeCmd())
	cmd.AddCommand(newSyncCmd())

	return cmd
}

func newRegisterCmd() *cobra.Command {
	var noSync bool
	var backgroundSync bool

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register with Jikime Rank via GitHub OAuth",
		Long:  "Opens your browser to authorize with GitHub.\nYour API key will be stored securely in ~/.jikime/rank/credentials.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegister(noSync, backgroundSync)
		},
	}

	cmd.Flags().BoolVar(&noSync, "no-sync", false, "Skip syncing existing sessions after registration")
	cmd.Flags().BoolVarP(&backgroundSync, "background-sync", "b", false, "Sync existing sessions in background after registration")

	return cmd
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show your current rank and statistics",
		Long:  "Displays your ranking position across different time periods,\ncumulative token usage statistics, and hook installation status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored Jikime Rank credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout()
		},
	}
}

func newExcludeCmd() *cobra.Command {
	var showList bool

	cmd := &cobra.Command{
		Use:   "exclude [path]",
		Short: "Exclude a project from session tracking",
		Long:  "Adds the specified project path (or current directory) to the exclusion list.\nSessions from excluded projects will not be submitted to Jikime Rank.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if showList {
				return runExcludeList()
			}
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runExclude(path)
		},
	}

	cmd.Flags().BoolVarP(&showList, "list", "l", false, "List all excluded projects")

	return cmd
}

func newIncludeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "include [path]",
		Short: "Re-include a previously excluded project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runInclude(path)
		},
	}
}

func newSyncCmd() *cobra.Command {
	var background bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all existing Claude Code sessions to Jikime Rank",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(background)
		},
	}

	cmd.Flags().BoolVarP(&background, "background", "b", false, "Run sync in background")

	return cmd
}

// Config directory paths
func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jikime", "rank")
}

func getCredentialsPath() string {
	return filepath.Join(getConfigDir(), "credentials.json")
}

func getConfigPath() string {
	return filepath.Join(getConfigDir(), "config.json")
}

func hasCredentials() bool {
	_, err := os.Stat(getCredentialsPath())
	return err == nil
}

func loadCredentials() (*RankCredentials, error) {
	data, err := os.ReadFile(getCredentialsPath())
	if err != nil {
		return nil, err
	}

	var creds RankCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

func saveCredentials(creds *RankCredentials) error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getCredentialsPath(), data, 0600)
}

func deleteCredentials() error {
	return os.Remove(getCredentialsPath())
}

func loadRankConfig() (map[string]any, error) {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return make(map[string]any), nil
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return make(map[string]any), nil
	}

	return config, nil
}

func saveRankConfig(config map[string]any) error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}

// HMAC signature computation
func computeSignature(apiKey, timestamp, body string) string {
	message := fmt.Sprintf("%s:%s", timestamp, body)
	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// API client functions
func makeAPIRequest(method, endpoint string, apiKey string, body []byte) ([]byte, error) {
	baseURL := os.Getenv("JIKIME_RANK_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	url := fmt.Sprintf("%s/api/%s%s", baseURL, apiVersion, endpoint)

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, strings.NewReader(string(body)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "jikime-adk-v2/1.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func getUserRank(apiKey string) (*UserRank, error) {
	respBody, err := makeAPIRequest("GET", "/rank", apiKey, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Username string `json:"username"`
			Rankings struct {
				Daily   *RankInfo `json:"daily"`
				Weekly  *RankInfo `json:"weekly"`
				Monthly *RankInfo `json:"monthly"`
				AllTime *RankInfo `json:"allTime"`
			} `json:"rankings"`
			Stats struct {
				TotalTokens   int `json:"totalTokens"`
				TotalSessions int `json:"totalSessions"`
				InputTokens   int `json:"inputTokens"`
				OutputTokens  int `json:"outputTokens"`
			} `json:"stats"`
			LastUpdated string `json:"lastUpdated"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return &UserRank{
		Username:      response.Data.Username,
		Daily:         response.Data.Rankings.Daily,
		Weekly:        response.Data.Rankings.Weekly,
		Monthly:       response.Data.Rankings.Monthly,
		AllTime:       response.Data.Rankings.AllTime,
		TotalTokens:   response.Data.Stats.TotalTokens,
		TotalSessions: response.Data.Stats.TotalSessions,
		InputTokens:   response.Data.Stats.InputTokens,
		OutputTokens:  response.Data.Stats.OutputTokens,
		LastUpdated:   response.Data.LastUpdated,
	}, nil
}

func formatTokens(tokens int) string {
	if tokens >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	} else if tokens >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1_000)
	}
	return strconv.Itoa(tokens)
}

func getRankMedal(position int) string {
	gold := color.New(color.FgYellow, color.Bold)
	silver := color.New(color.FgWhite)
	bronze := color.New(color.FgRed)

	switch position {
	case 1:
		return gold.Sprint("1st")
	case 2:
		return silver.Sprint("2nd")
	case 3:
		return bronze.Sprint("3rd")
	default:
		return fmt.Sprintf("#%d", position)
	}
}

func createProgressBar(value, total, width int) string {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)

	if total == 0 {
		return dim.Sprint(strings.Repeat("-", width))
	}

	ratio := float64(value) / float64(total)
	if ratio > 1.0 {
		ratio = 1.0
	}
	filled := int(float64(width) * ratio)
	return cyan.Sprint(strings.Repeat("█", filled)) + dim.Sprint(strings.Repeat("░", width-filled))
}

func runRegister(noSync, backgroundSync bool) error {
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	// Check if already registered
	if hasCredentials() {
		creds, _ := loadCredentials()
		if creds != nil {
			yellow.Printf("Already registered as %s\n", color.New(color.Bold).Sprint(creds.Username))
			fmt.Print("Do you want to re-register? (y/N): ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) != "y" {
				return nil
			}
		}
	}

	fmt.Println()
	cyan.Println("╔════════════════════════════════════════╗")
	cyan.Println("║       Jikime Rank Registration         ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("This will open your browser to authorize with GitHub.")
	fmt.Println("After authorization, your API key will be stored securely.")
	fmt.Println()

	// For now, show instructions for manual registration
	dim.Println("OAuth flow not yet implemented in Go version.")
	dim.Println("Please visit https://rank.mo.ai.kr to register manually.")
	dim.Println()
	dim.Println("After registration, you can set your API key with:")
	fmt.Println()
	fmt.Println("  export JIKIME_RANK_API_KEY=your-api-key")
	fmt.Println()
	dim.Println("Or create ~/.jikime/rank/credentials.json manually.")

	return nil
}

func runStatus() error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	if !hasCredentials() {
		yellow.Println("Not registered with Jikime Rank.")
		dim.Println("Run 'jikime-adk rank register' to connect your account.")
		return nil
	}

	creds, err := loadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	fmt.Println()
	dim.Println("Fetching your rank...")

	userRank, err := getUserRank(creds.APIKey)
	if err != nil {
		return fmt.Errorf("failed to fetch rank: %w", err)
	}

	fmt.Println()

	// Header Panel
	headerContent := fmt.Sprintf("%s\n\n", cyan.Sprint(userRank.Username))
	if userRank.Weekly != nil {
		headerContent += fmt.Sprintf("%s  %s %s\n",
			dim.Sprint("Weekly Rank"),
			getRankMedal(userRank.Weekly.Position),
			dim.Sprintf("/ %d", userRank.Weekly.TotalParticipants))
		headerContent += fmt.Sprintf("%s        %s",
			dim.Sprint("Score"),
			color.New(color.Bold).Sprintf("%.0f", userRank.Weekly.CompositeScore))
	} else {
		headerContent += dim.Sprint("No ranking data")
	}

	cyan.Println("╔════════════════════════════════════════╗")
	cyan.Println("║           Jikime Rank                  ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println(headerContent)
	fmt.Println()

	// Rankings Grid
	periods := []struct {
		name  string
		rank  *RankInfo
		color *color.Color
	}{
		{"Daily", userRank.Daily, yellow},
		{"Weekly", userRank.Weekly, cyan},
		{"Monthly", userRank.Monthly, green},
		{"All Time", userRank.AllTime, color.New(color.FgMagenta)},
	}

	for _, p := range periods {
		if p.rank != nil {
			fmt.Printf("  %s: %s %s (%.0f)\n",
				p.color.Sprint(p.name),
				getRankMedal(p.rank.Position),
				dim.Sprintf("/ %d", p.rank.TotalParticipants),
				p.rank.CompositeScore)
		} else {
			fmt.Printf("  %s: %s\n", p.color.Sprint(p.name), dim.Sprint("-"))
		}
	}

	fmt.Println()

	// Token Statistics
	total := userRank.TotalTokens
	var inputPct, outputPct float64
	if total > 0 {
		inputPct = float64(userRank.InputTokens) / float64(total) * 100
		outputPct = float64(userRank.OutputTokens) / float64(total) * 100
	}

	green.Println("Token Usage")
	green.Println("───────────")
	fmt.Printf("  %s %s\n", color.New(color.Bold).Sprint(formatTokens(total)), dim.Sprint("total tokens"))
	fmt.Println()
	fmt.Printf("  %s  %s %s %s\n",
		dim.Sprint("Input"),
		createProgressBar(userRank.InputTokens, total, 15),
		color.New(color.Bold).Sprint(formatTokens(userRank.InputTokens)),
		dim.Sprintf("(%.0f%%)", inputPct))
	fmt.Printf("  %s %s %s %s\n",
		dim.Sprint("Output"),
		createProgressBar(userRank.OutputTokens, total, 15),
		color.New(color.Bold).Sprint(formatTokens(userRank.OutputTokens)),
		dim.Sprintf("(%.0f%%)", outputPct))
	fmt.Println()
	fmt.Printf("  %s %s\n", dim.Sprint("Sessions:"), color.New(color.Bold).Sprint(userRank.TotalSessions))

	fmt.Println()
	fmt.Printf("  %s  %s\n", cyan.Sprint("https://rank.mo.ai.kr"), dim.Sprint("for full leaderboard"))

	return nil
}

func runLogout() error {
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	dim := color.New(color.Faint)

	if !hasCredentials() {
		yellow.Println("No credentials stored.")
		return nil
	}

	creds, _ := loadCredentials()
	username := "unknown"
	if creds != nil {
		username = creds.Username
	}

	fmt.Printf("Remove credentials for %s? (y/N): ", username)
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) == "y" {
		if err := deleteCredentials(); err != nil {
			return fmt.Errorf("failed to delete credentials: %w", err)
		}
		green.Println("Credentials removed successfully.")
	} else {
		dim.Println("Cancelled.")
	}

	return nil
}

func runExcludeList() error {
	dim := color.New(color.Faint)

	config, _ := loadRankConfig()
	exclusions, ok := config["exclude_projects"].([]any)
	if !ok || len(exclusions) == 0 {
		dim.Println("No projects are excluded from tracking.")
		dim.Println("Use 'jikime-adk rank exclude <path>' to exclude a project.")
		return nil
	}

	fmt.Println()
	color.New(color.Bold).Println("Excluded Projects:")
	for _, exc := range exclusions {
		if path, ok := exc.(string); ok {
			fmt.Printf("  %s %s\n", dim.Sprint("•"), path)
		}
	}
	fmt.Println()
	dim.Printf("Total: %d project(s) excluded\n", len(exclusions))

	return nil
}

func runExclude(path string) error {
	green := color.New(color.FgGreen)
	dim := color.New(color.Faint)

	// Use current directory if no path specified
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}

	config, _ := loadRankConfig()
	exclusions, _ := config["exclude_projects"].([]any)

	// Check if already excluded
	for _, exc := range exclusions {
		if exc == path {
			dim.Printf("Project already excluded: %s\n", path)
			return nil
		}
	}

	// Add to exclusions
	exclusions = append(exclusions, path)
	config["exclude_projects"] = exclusions

	if err := saveRankConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	green.Printf("Excluded: %s\n", color.New(color.Bold).Sprint(path))
	dim.Println("Sessions from this project will not be tracked.")

	if len(exclusions) > 1 {
		dim.Printf("\nTotal excluded projects: %d\n", len(exclusions))
	}

	return nil
}

func runInclude(path string) error {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	dim := color.New(color.Faint)

	// Use current directory if no path specified
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}

	config, _ := loadRankConfig()
	exclusions, _ := config["exclude_projects"].([]any)

	// Find and remove from exclusions
	found := false
	newExclusions := make([]any, 0)
	for _, exc := range exclusions {
		if exc == path {
			found = true
		} else {
			newExclusions = append(newExclusions, exc)
		}
	}

	if !found {
		red.Printf("Project not in exclusion list: %s\n", path)
		return nil
	}

	config["exclude_projects"] = newExclusions

	if err := saveRankConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	green.Printf("Included: %s\n", color.New(color.Bold).Sprint(path))
	dim.Println("Sessions from this project will now be tracked.")

	if len(newExclusions) > 0 {
		dim.Printf("\nRemaining excluded projects: %d\n", len(newExclusions))
	}

	return nil
}

func runSync(background bool) error {
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	if !hasCredentials() {
		yellow.Println("Not registered with Jikime Rank.")
		dim.Println("Run 'jikime-adk rank register' first.")
		return nil
	}

	if background {
		// TODO: Implement background sync
		dim.Println("Background sync not yet implemented in Go version.")
		dim.Println("Running foreground sync instead...")
	}

	// TODO: Implement session sync logic
	dim.Println("Session sync not yet fully implemented.")
	dim.Println("Sync functionality will scan ~/.claude/projects/ for session transcripts.")

	return nil
}
