package hookscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"jikime-adk/version"
)

// SessionStartCmd represents the session-start hook command
var SessionStartCmd = &cobra.Command{
	Use:   "session-start",
	Short: "Display enhanced project information at session start",
	Long: `SessionStart hook that displays project information including:
- Project name and language settings
- Git branch and status
- Configuration overview`,
	RunE: runSessionStart,
}

// HookResponse represents the JSON response format for Claude Code hooks
type HookResponse struct {
	Continue      bool              `json:"continue"`
	SystemMessage string            `json:"systemMessage,omitempty"`
	Performance   map[string]bool   `json:"performance,omitempty"`
	ErrorDetails  map[string]string `json:"error_details,omitempty"`
}

// ConfigSection represents a section of the configuration
type ConfigSection struct {
	User struct {
		Name              string `yaml:"name"`
		Honorific         string `yaml:"honorific"`
		TonePreset        string `yaml:"tone_preset"`
		OrchestratorStyle string `yaml:"orchestrator_style"`
	} `yaml:"user"`
	Language struct {
		ConversationLanguage     string `yaml:"conversation_language"`
		ConversationLanguageName string `yaml:"conversation_language_name"`
		GitCommitMessages        string `yaml:"git_commit_messages"`
		CodeComments             string `yaml:"code_comments"`
		Documentation            string `yaml:"documentation"`
	} `yaml:"language"`
	Project struct {
		Name string `yaml:"name"`
	} `yaml:"project"`
	GitStrategy struct {
		Mode     string                `yaml:"mode"`
		Manual   GitStrategyModeConfig `yaml:"manual"`
		Personal GitStrategyModeConfig `yaml:"personal"`
		Team     GitStrategyModeConfig `yaml:"team"`
	} `yaml:"git_strategy"`
}

// GitStrategyModeConfig represents mode-specific git strategy configuration
type GitStrategyModeConfig struct {
	BranchCreation struct {
		AutoEnabled bool `yaml:"auto_enabled"`
	} `yaml:"branch_creation"`
}

func runSessionStart(cmd *cobra.Command, args []string) error {
	// Read input from stdin (Claude Code passes session info)
	var input map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	// Find project root once (avoid duplicate findProjectRoot calls)
	projectRoot, err := findProjectRoot()
	if err != nil {
		response := HookResponse{
			Continue:      true,
			SystemMessage: "⚠️ Session start encountered an error - continuing",
			ErrorDetails: map[string]string{
				"error": err.Error(),
			},
		}
		return writeResponse(response)
	}

	// Load config once (avoid duplicate loadConfig calls)
	config, _ := loadConfig(projectRoot)

	// Initialize orchestrator state only if no state file exists (preserve sticky state after compact)
	if !stateFileExists(projectRoot) {
		orchestrator := determineInitialOrchestratorWithConfig(projectRoot, config)
		writeOrchestratorState(projectRoot, orchestrator)
	}

	// Generate session output
	output := formatSessionOutputWithContext(projectRoot, config)

	// Return success response
	response := HookResponse{
		Continue:      true,
		SystemMessage: output,
		Performance: map[string]bool{
			"go_hook": true,
		},
	}
	return writeResponse(response)
}

func writeResponse(response HookResponse) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(response)
}

func formatSessionOutputWithContext(projectRoot string, config *ConfigSection) string {
	if config == nil {
		config = &ConfigSection{}
	}

	// Get Git information (3 commands in parallel)
	gitInfo := getGitInfoParallel(projectRoot)

	// Get version
	appVersion := version.String()

	// Get git strategy info (Github-Flow and Auto Branch)
	gitMode := config.GitStrategy.Mode
	if gitMode == "" {
		gitMode = "manual"
	}

	var autoBranch string
	switch gitMode {
	case "manual":
		autoBranch = "No"
		if config.GitStrategy.Manual.BranchCreation.AutoEnabled {
			autoBranch = "Yes"
		}
	case "personal":
		autoBranch = "Yes"
		if !config.GitStrategy.Personal.BranchCreation.AutoEnabled {
			autoBranch = "No"
		}
	case "team":
		autoBranch = "Yes"
		if !config.GitStrategy.Team.BranchCreation.AutoEnabled {
			autoBranch = "No"
		}
	default:
		autoBranch = "unknown"
	}

	// Get language display name
	languageName := config.Language.ConversationLanguageName
	if languageName == "" {
		languageName = "English"
	}
	conversationLang := config.Language.ConversationLanguage
	if conversationLang == "" {
		conversationLang = "en"
	}

	// Get user name for greeting
	userName := strings.TrimSpace(config.User.Name)

	// Build output message (jikime-adk format)
	var output strings.Builder

	output.WriteString("🚀 JikiME-ADK Session Started\n")
	output.WriteString(fmt.Sprintf("   📦 Version: %s\n", appVersion))
	output.WriteString(fmt.Sprintf("   🔄 Changes: %s\n", gitInfo["changes"]))
	output.WriteString(fmt.Sprintf("   🌿 Branch: %s\n", gitInfo["branch"]))
	output.WriteString(fmt.Sprintf("   🔧 Github-Flow: %s | Auto Branch: %s\n", gitMode, autoBranch))
	output.WriteString(fmt.Sprintf("   🔨 Last Commit: %s\n", gitInfo["last_commit"]))
	output.WriteString(fmt.Sprintf("   🌐 Language: %s (%s)\n", languageName, conversationLang))

	// Add welcome message if user name is set and not a template variable
	if userName != "" && !strings.HasPrefix(userName, "{{") {
		output.WriteString(fmt.Sprintf("   👋 Welcome back, %s!\n", userName))
	}

	return output.String()
}

func findProjectRoot() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for .jikime directory
	for {
		jikimePath := filepath.Join(dir, ".jikime")
		if info, err := os.Stat(jikimePath); err == nil && info.IsDir() {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .jikime
			return "", fmt.Errorf(".jikime directory not found")
		}
		dir = parent
	}
}

func loadConfig(projectRoot string) (*ConfigSection, error) {
	config := &ConfigSection{}

	sectionsDir := filepath.Join(projectRoot, ".jikime", "config")

	// Load section files
	sectionFiles := []string{
		"user.yaml",
		"language.yaml",
		"project.yaml",
		"git-strategy.yaml",
	}

	for _, filename := range sectionFiles {
		path := filepath.Join(sectionsDir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			// Skip missing files
			continue
		}

		// Merge into config
		if err := yaml.Unmarshal(data, config); err != nil {
			// Skip files that fail to parse
			continue
		}
	}

	return config, nil
}

func getGitInfo(projectRoot string) map[string]string {
	info := make(map[string]string)

	// Get current branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectRoot
	if output, err := cmd.Output(); err == nil {
		info["branch"] = strings.TrimSpace(string(output))
	}

	// Get status summary
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectRoot
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			info["changes"] = fmt.Sprintf("%d file(s) modified", len(lines))
		} else {
			info["changes"] = "No changes"
		}
	}

	// Get last commit
	cmd = exec.Command("git", "log", "-1", "--pretty=format:%h - %s (%ar)")
	cmd.Dir = projectRoot
	if output, err := cmd.Output(); err == nil {
		info["last_commit"] = strings.TrimSpace(string(output))
	}

	return info
}

// getGitInfoParallel runs 3 git commands concurrently for faster startup.
func getGitInfoParallel(projectRoot string) map[string]string {
	info := make(map[string]string)

	type gitResult struct {
		key   string
		value string
	}

	ch := make(chan gitResult, 3)

	// Branch (goroutine 1)
	go func() {
		cmd := exec.Command("git", "branch", "--show-current")
		cmd.Dir = projectRoot
		if output, err := cmd.Output(); err == nil {
			ch <- gitResult{"branch", strings.TrimSpace(string(output))}
		} else {
			ch <- gitResult{"branch", ""}
		}
	}()

	// Status (goroutine 2)
	go func() {
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = projectRoot
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) > 0 && lines[0] != "" {
				ch <- gitResult{"changes", fmt.Sprintf("%d file(s) modified", len(lines))}
			} else {
				ch <- gitResult{"changes", "No changes"}
			}
		} else {
			ch <- gitResult{"changes", "unknown"}
		}
	}()

	// Last commit (goroutine 3)
	go func() {
		cmd := exec.Command("git", "log", "-1", "--pretty=format:%h - %s (%ar)")
		cmd.Dir = projectRoot
		if output, err := cmd.Output(); err == nil {
			ch <- gitResult{"last_commit", strings.TrimSpace(string(output))}
		} else {
			ch <- gitResult{"last_commit", ""}
		}
	}()

	for i := 0; i < 3; i++ {
		r := <-ch
		info[r.key] = r.value
	}

	return info
}

// determineInitialOrchestrator decides the initial orchestrator using improved logic:
// Priority 1: Active migration detection (progress.yaml with in_progress status)
// Priority 2: User's orchestrator_style setting (fallback)
// Priority 3: Default to J.A.R.V.I.S.
func determineInitialOrchestrator(projectRoot string) string {
	config, _ := loadConfig(projectRoot)
	return determineInitialOrchestratorWithConfig(projectRoot, config)
}

// determineInitialOrchestratorWithConfig uses a pre-loaded config to avoid duplicate loadConfig calls.
func determineInitialOrchestratorWithConfig(projectRoot string, config *ConfigSection) string {
	// Priority 1: Check for ACTIVE migration (not just artifacts)
	if isActiveMigration(projectRoot) {
		return OrchestratorFRIDAY
	}

	// Priority 2: Check user's orchestrator_style setting
	if config != nil && config.User.OrchestratorStyle != "" {
		switch config.User.OrchestratorStyle {
		case "jarvis":
			return OrchestratorJARVIS
		case "friday":
			return OrchestratorFRIDAY
		case "auto":
			// Fall through to artifact-based detection
		}
	}

	// Priority 3 (auto mode): Artifact-based detection (legacy behavior)
	if hasMigrationArtifacts(projectRoot) {
		return OrchestratorJARVIS
	}

	// Default: J.A.R.V.I.S. for development work
	return OrchestratorJARVIS
}
